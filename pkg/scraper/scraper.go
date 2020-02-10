package scraper

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly/v2"
)

//TfpdlMovie : describing a single tfpdl movie scraper
type TfpdlMovie struct {
	Index         int
	Name          string
	PermaLink     string
	SafeTxtLink   string
	PictureLink   string
	Description   string
	DownloadLinks []string
}

//NetNaijaMovie : describing a single netnaija movie scraper
type NetNaijaMovie struct {
	Index        int
	Title        string
	PictureLink  string
	Description  string
	DownloadLink string
	Size         string
}

func (movie *NetNaijaMovie) getdownloadlink() {
	c := colly.NewCollector()
	formerdownloadlink := movie.DownloadLink
	c.OnHTML("button[id=download-button]", func(inn *colly.HTMLElement) {
		movie.Size = strings.TrimSpace(inn.ChildText("span.size"))
	})

	c.OnHTML("a[id=download]", func(inn *colly.HTMLElement) {
		movie.Size = strings.TrimSpace(inn.ChildText("span[id=download-size]"))
		movie.DownloadLink = inn.Attr("href")
	})
	c.OnHTML("div.row", func(inner *colly.HTMLElement) {
		if strings.TrimSpace(inner.ChildText("label")) == "Direct Download" {
			movie.DownloadLink = inner.ChildAttr("input", "value")
		}
	})
	c.Visit(formerdownloadlink)
	if formerdownloadlink == movie.DownloadLink {
		movie.Size = "Unknown"
	}
}

// Tfpdl : Scraper structure for TFPDL
type Tfpdl struct {
	Query     string
	PageTitle string
	Movies    []TfpdlMovie
}

//NetNaija : Scraper structure for NetNaija
type NetNaija struct {
	Title  string
	Movies []NetNaijaMovie
}

// Search : Searches netnaija for a particular query
func (site *NetNaija) Search(Query string) {
	query := strings.ReplaceAll(Query, " ", "+")
	site.Title = "Search Results for " + Query
	url := "https://thenetnaija.com/search?t=Movie%3A" + query + "&folder=videos"
	movieIndex := 0

	c := colly.NewCollector()

	c.OnHTML("main", func(e *colly.HTMLElement) {
		e.ForEach("article", func(_ int, el *colly.HTMLElement) {
			movie := NetNaijaMovie{
				Index:        movieIndex,
				Title:        "",
				PictureLink:  "",
				Description:  "",
				DownloadLink: "",
				Size:         "",
			}

			movie.PictureLink = el.ChildAttr("img", "src")
			movie.Title = strings.TrimSpace(strings.TrimPrefix(el.ChildText("h3.result-title"), "Movie:"))
			movie.Description = strings.TrimSpace(el.ChildText("p.result-desc"))
			innerlink := el.ChildAttr("a", "href") + "/download"
			movie.DownloadLink = innerlink
			site.Movies = append(site.Movies, movie)
			movieIndex++
		})

	})

	c.Visit(url)

	for _, movie := range site.Movies {
		movie.getdownloadlink()
	}

}

// SafetextlinkSearch : Searches tfpdl for the movie primarily to return each SafeTxtLink
func (site *Tfpdl) SafetextlinkSearch() {
	url := "https://tfp.is/page/1/?s=%s"
	query := strings.ReplaceAll(site.Query, " ", "+")
	movieindex := 0
	url = fmt.Sprintf(url, query)

	c := colly.NewCollector()

	c.OnHTML("div.page-title", func(e *colly.HTMLElement) {
		site.PageTitle = strings.TrimSpace(e.Text)
	})

	c.OnHTML("div.post-listing", func(e *colly.HTMLElement) {
		innercol := colly.NewCollector()
		e.ForEach("article.item-list", func(_ int, el *colly.HTMLElement) {
			movie := TfpdlMovie{
				Index:       movieindex,
				Name:        "",
				PictureLink: "",
				PermaLink:   "",
			}
			movie.Name = strings.TrimSpace(el.ChildText("h2"))
			movie.PictureLink = strings.TrimSpace(el.ChildAttr("img", "src"))
			movie.PermaLink = strings.TrimSpace(el.ChildAttr("a", "href"))

			innercol.OnHTML(".entry", func(permapage *colly.HTMLElement) {
				movie.Description = strings.TrimSpace(permapage.ChildText("p"))
				movie.SafeTxtLink = strings.TrimSpace(permapage.ChildAttr("a.button", "href"))
			})
			innercol.Visit(movie.PermaLink)
			site.Movies = append(site.Movies, movie)
			movieindex++
		})
	})

	c.Visit(url)
	c.Wait()
}

// GetDownloadLinks : Specify the index of the movie to be downloaded
func (site *Tfpdl) GetDownloadLinks(index int) {
	selected := site.Movies[index]
	fmt.Println(selected)
	c := colly.NewCollector()

	c.OnHTML("div", func(e *colly.HTMLElement) {
		fmt.Printf("%+v", e)
		err := c.Post(selected.SafeTxtLink, map[string]string{"Pass1": "tfpdl"})
		fmt.Println(err)
	})

	c.Visit(selected.SafeTxtLink)
}
