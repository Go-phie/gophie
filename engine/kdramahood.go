package engine

import (
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	log "github.com/sirupsen/logrus"
)

// KDramaHood : An Engine for KDramaHood
type KDramaHood struct {
	Props
}

// NewKDramaHoodEngine : create a new engine for scraping latest korean drama
func NewKDramaHoodEngine() *KDramaHood {
	base := "https://kdramahood.com"
	baseURL, err := url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}
	// Search URL
	searchURL, err := url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}
	searchURL.Path = "/"

	// List URL
	listURL, err := url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}
	listURL.Path = "/home2/"

	dramaFeverEngine := KDramaHood{}
	dramaFeverEngine.Name = "KDramaHood"
	dramaFeverEngine.BaseURL = baseURL
	dramaFeverEngine.Description = `Watch your favourite korean movie all in one place`
	dramaFeverEngine.SearchURL = searchURL
	dramaFeverEngine.ListURL = listURL
	return &dramaFeverEngine
}

// Engine Interface Methods

func (engine *KDramaHood) String() string {
	st := fmt.Sprintf("%s (%s)", engine.Name, engine.BaseURL)
	return st
}

func (engine *KDramaHood) getParseAttrs() (string, string, error) {
	return "div.items", "div.item", nil
}

func (engine *KDramaHood) parseSingleMovie(el *colly.HTMLElement, index int) (Movie, error) {
	movie := Movie{
		Index:    index,
		IsSeries: true,
		Source:   engine.Name,
		Size:     "---MB",
	}
	switch engine.mode {
	case SearchMode:
		movie.Title = strings.TrimSpace(el.ChildText("span.tt"))
		movie.Description = strings.TrimSpace(el.ChildText("span.ttx"))
	case ListMode:
		movie.Title = strings.TrimSpace(el.ChildAttr("img", "alt"))
		movie.Description = strings.TrimSpace(el.ChildText("div.contenido"))
	}
	movie.CoverPhotoLink = el.ChildAttr("img", "src")
	link := el.Request.AbsoluteURL(el.ChildAttr("a", "href"))
	downloadLink, err := url.Parse(link)

	if err != nil {
		log.Fatal(err)
	}
	movie.DownloadLink = downloadLink
	movie.Category = "kdrama"
	return movie, nil
}

func (engine *KDramaHood) updateDownloadProps(downloadCollector *colly.Collector, movies *[]Movie) {
	innerCollector := downloadCollector.Clone()
	episodeMap := map[string]*url.URL{}
	subtitleMap := map[string]*url.URL{}
	downloadCollector.OnHTML("ul.episodios", func(e *colly.HTMLElement) {
		// clear existing map
		for k := range episodeMap {
			delete(episodeMap, k)
		}
		for k := range subtitleMap {
			delete(subtitleMap, k)
		}
		// create local targets
		targetepisode := make(map[string]*url.URL)
		targetsub := make(map[string]*url.URL)
		movie := &(*movies)[getMovieIndexFromCtx(e.Request)]
		e.ForEach("li", func(_ int, inn *colly.HTMLElement) {
			innerCollector.Visit(inn.ChildAttr("a", "href"))
		})

		// deepcopy to localtargets
		for k, v := range episodeMap {
			targetepisode[k] = v
		}
		for k, v := range subtitleMap {
			targetsub[k] = v
		}

		movie.SDownloadLink = targetepisode
		movie.SubtitleLinks = targetsub
	})

	innerCollector.OnHTML("div.linkstv", func(e *colly.HTMLElement) {
		name := e.ChildAttr("a", "download")
		links := e.ChildAttrs("a", "href")
		if len(links) > 1 {
			// select first link
			movieLink, _ := url.Parse(links[0])
			// subtitle is always the last link
			subLink, _ := url.Parse(links[len(links)-1])
			episodeMap[name] = movieLink
			subtitleMap[name] = subLink
		}
	})
}

// List : list all the movies on a page
func (engine *KDramaHood) List(page int) SearchResult {
	engine.mode = ListMode
	result := SearchResult{
		Query: "List of Recent Uploads - Page " + strconv.Itoa(page),
	}
	pageParam := fmt.Sprintf("page/%v", strconv.Itoa(page))
	engine.ListURL.Path = path.Join(engine.ListURL.Path, pageParam)
	movies, err := Scrape(engine)
	if err != nil {
		log.Fatal(err)
	}
	result.Movies = movies
	return result
}

// Search : Searches fzmovies for a particular query and return an array of movies
func (engine *KDramaHood) Search(param ...string) SearchResult {
	query := param[0]
	engine.mode = SearchMode
	result := SearchResult{
		Query: query,
	}
	q := engine.SearchURL.Query()
	q.Set("s", query)
	engine.SearchURL.RawQuery = q.Encode()
	movies, err := Scrape(engine)
	if err != nil {
		log.Fatal(err)
	}
	result.Movies = movies
	return result
}
