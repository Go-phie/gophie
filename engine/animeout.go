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

// AnimeOut : An Engine for AnimeOut
type AnimeOut struct {
	Props
}

// NewAnimeOutEngine : create a new engine for scraping latest anime from chia-anime
func NewAnimeOutEngine() *AnimeOut {
	base := "https://animeout.xyz"
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
	listURL.Path = "/all-releases/"

	animeOutEngine := AnimeOut{}
	animeOutEngine.Name = "AnimeOut"
	animeOutEngine.BaseURL = baseURL
	animeOutEngine.Description = `Search from over 1000's of encoded anime available`
	animeOutEngine.SearchURL = searchURL
	animeOutEngine.ListURL = listURL
	return &animeOutEngine
}

// Engine Interface Methods

func (engine *AnimeOut) String() string {
	st := fmt.Sprintf("%s (%s)", engine.Name, engine.BaseURL)
	return st
}

func (engine *AnimeOut) getParseAttrs() (string, string, error) {
	return "div.container", "article.post-item", nil
}

func (engine *AnimeOut) parseSingleMovie(el *colly.HTMLElement, index int) (Movie, error) {
	movie := Movie{
		Index:    index,
		IsSeries: true,
		Source:   engine.Name,
		Size:     "---MB",
	}
	movie.Title = strings.TrimSpace(el.ChildText("h3.post-title"))
	movie.CoverPhotoLink = el.ChildAttr("img", "src")
	downloadLink, err := url.Parse(el.Request.AbsoluteURL(el.ChildAttr("a", "href")))

	if err != nil {
		log.Fatal(err)
	}
	movie.DownloadLink = downloadLink
	return movie, nil
}

func (engine *AnimeOut) updateDownloadProps(downloadCollector *colly.Collector, movies *[]Movie) {
	downloadCollector.OnHTML("div.article-content", func(e *colly.HTMLElement) {
		movie := &(*movies)[getMovieIndexFromCtx(e.Request)]
		description := e.ChildText("div.spaceit")
		episodeMap := map[string]*url.URL{}
		if description == "" {
			if len(e.ChildTexts("p")) > 2 {
				description = e.ChildTexts("p")[1]
			}
		}
		episodeLinks := e.ChildAttrs("a", "href")[1:]
		index := 0
		for _, link := range episodeLinks {
			if (link != "#") && (link != "") {
				_, filename := path.Split(link)
				file := strings.TrimSuffix(filename, path.Ext(filename))
				file, _ = url.QueryUnescape(file)
				if strings.HasPrefix(link, "http://download.animeout.com/") {
					link = strings.TrimPrefix(link, "http://download.animeout.com/")
					link = "http://public.animeout.xyz/sv1.animeout.com/" + link
				} else {
					link = strings.TrimPrefix(link, "http://")
					link = strings.TrimPrefix(link, "https://")
					link = "http://public.animeout.xyz/" + link
				}
				downloadlink, _ := url.Parse(link)
				if !strings.HasPrefix(file, "#") {
					episodeMap[file] = downloadlink
				}
				if index == 0 {
					movie.DownloadLink = downloadlink
				}
			}
			index++
		}

		movie.Description = description
		movie.SDownloadLink = episodeMap
	})
}

// List : list all the movies on a page
func (engine *AnimeOut) List(page int) SearchResult {
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
func (engine *AnimeOut) Search(param ...string) SearchResult {
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
