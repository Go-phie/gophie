package scraper

import (
	_ "encoding/json"
	"fmt"
	_ "os"
	"strings"

	"github.com/gocolly/colly/v2"
)

// Movie : describing a single movie scraper
type Movie struct {
	Index         int
	Name          string
	PermaLink     string
	SafeTxtLink   string
	PictureLink   string
	Description   string
	DownloadLinks []string
}

// Tfpdl : Scraper structure for TFPDL
type Tfpdl struct {
	Query     string
	PageTitle string
	Movies    []Movie
}

// SafetxtlinkSearch : Searches tfpdl for the movie primarily to return each SafeTxtLink
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
			movie := Movie{
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
