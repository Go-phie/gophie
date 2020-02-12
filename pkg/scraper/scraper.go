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
	Index         int
	Title         string
	PictureLink   string
	Description   string
	DownloadLink  string
	SDownloadLink []string
	Size          string
	Series        bool
}

func (movie *NetNaijaMovie) getdownloadlink(c *colly.Collector) {
	formerdownloadlink := movie.DownloadLink
	if strings.HasPrefix(formerdownloadlink, "https://www.thenetnaija.com/videos/series/") {
		movie.Series = true
	}
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

	//for series or parts
	c.OnHTML("div.video-series-latest-episodes", func(inn *colly.HTMLElement) {
		movie.Series = true
		inn.ForEach("a", func(_ int, e *colly.HTMLElement) {
			movie.SDownloadLink = append(movie.SDownloadLink, e.Attr("href")+"/download")
		})
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
				Index:        0,
				Title:        "",
				PictureLink:  "",
				Description:  "",
				DownloadLink: "",
				Size:         "",
				Series:       false,
			}

			movie.PictureLink = el.ChildAttr("img", "src")
			movie.Title = strings.TrimSpace(strings.TrimPrefix(el.ChildText("h3.result-title"), "Movie:"))
			movie.Description = strings.TrimSpace(el.ChildText("p.result-desc"))
			href := el.ChildAttr("a", "href")
			innerlink := href + "/download"

			movie.DownloadLink = innerlink
			if movie.Title != "" {
				movie.Index = movieIndex
				site.Movies = append(site.Movies, movie)
				c.Visit(movie.DownloadLink)
				movieIndex++
			}
		})

	})

	//  c.OnRequest(func(r *colly.Request) {
	//    fmt.Println("Visiting", r.URL.String())
	//  })

	c.OnResponse(func(r *colly.Response) {
		if len(site.Movies) > 0 {
			lastMovie := &site.Movies[len(site.Movies)-1]
			lastMovie.DownloadLink = r.Request.URL.String()
			lastMovie.getdownloadlink(c)
		}
	})

	c.Visit(url)
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
