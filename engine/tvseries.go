package engine

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	log "github.com/sirupsen/logrus"
)

// TvSeriesEngine : An Engine for TvSeries
type TvSeriesEngine struct {
	Props
}

// NewTvSeriesEngine : A Movie Engine Constructor for TvSeriesEngine
func NewTvSeriesEngine() *TvSeriesEngine {
	base := "https://tvseries.in/"
	baseURL, err := url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}
	// Search URL
	searchURL, err := url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}
	searchURL.Path = "/search.php"

	// List URL
	listURL, err := url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}
	listURL.Path = "/tv.php"

	TvSeriesEngine := TvSeriesEngine{}
	TvSeriesEngine.Name = "TvSeries"
	TvSeriesEngine.BaseURL = baseURL
	TvSeriesEngine.Description = `TvSeries is a site owned by the fzmovies group where shows are available`
	TvSeriesEngine.SearchURL = searchURL
	TvSeriesEngine.ListURL = listURL
	return &TvSeriesEngine
}

// Engine Interface Methods

func (engine *TvSeriesEngine) String() string {
	st := fmt.Sprintf("%s (%s)", engine.Name, engine.BaseURL)
	return st
}

func (engine *TvSeriesEngine) getParseAttrs() (string, string, error) {
	return "body", "div.mainbox", nil
}

func (engine *TvSeriesEngine) parseSingleMovie(el *colly.HTMLElement, index int) (Movie, error) {
	movie := Movie{
		Index:    index,
		IsSeries: false,
		Source:   engine.Name,
	}
	cover, err := url.Parse(el.Request.AbsoluteURL(el.ChildAttr("img", "src")))
	if err != nil {
		log.Fatal(err)
	}
	movie.CoverPhotoLink = cover.String()
	titleAndDescription := el.ChildTexts("small")

	if len(titleAndDescription) > 1 {
		movie.Title = strings.TrimSpace(titleAndDescription[0])
		movie.Description = strings.TrimSpace(titleAndDescription[1])
		lastSmall := titleAndDescription[len(titleAndDescription) -1]
		if len(lastSmall) > len(movie.Description) {
			movie.Description = lastSmall
		}
	}
	link := el.Request.AbsoluteURL(el.ChildAttr("a", "href"))
	downloadLink, err := url.Parse(link + "&ftype=2")

	if err != nil {
		log.Fatal(err)
	}

	movie.DownloadLink = downloadLink
	return movie, nil
}

func (engine *TvSeriesEngine) updateDownloadProps(downloadCollector *colly.Collector, movies *[]Movie) {
	// For listing movies and retrieving the most recently updated episode
	downloadCollector.OnHTML("div[itemprop=episode]", func(e *colly.HTMLElement) {
		movie := &(*movies)[getMovieIndexFromCtx(e.Request)]
		if len(e.ChildTexts("b")) > 1 {
			movie.Title = e.ChildTexts("b")[0]
		}
		titleAndDescription := e.ChildTexts("small")
		if len(titleAndDescription) > 1 {
			movie.Description = titleAndDescription[len(titleAndDescription) - 1]
		}
		if e.Request.AbsoluteURL(e.ChildAttr("a", "href")) != e.Request.URL.String(){
			link := e.Request.AbsoluteURL(e.ChildAttr("a", "href")) + "&ftype=2" 	
			downloadLink, err := url.Parse(link)
			if err != nil {
				log.Fatal(err)
			} else {
				movie.DownloadLink = downloadLink
			}
			downloadCollector.Visit(downloadLink.String())
		}
	})

	// // Update movie download link if ul.downloadlinks on page
	// downloadCollector.OnHTML("a[id=dlink2]", func(e *colly.HTMLElement) {
	// 	movie := &(*movies)[getMovieIndexFromCtx(e.Request)]
	// 	link := e.Request.AbsoluteURL(e.Attr("href")) 	
	// 	downloadLink, err := url.Parse(link)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	} else {
	// 		movie.DownloadLink = downloadLink
	// 	}
	// 	downloadCollector.Visit(downloadLink.String())
	// })

	for _, iden := range [...]string{ "a[id=dlink3]",  "a[id=dlink4]", "a[id=dlink2]"} {
		// Update movie download link if ul.downloadlinks on page
		downloadCollector.OnHTML(iden, func(e *colly.HTMLElement) {
			movie := &(*movies)[getMovieIndexFromCtx(e.Request)]
			link := e.Request.AbsoluteURL(e.Attr("href")) 	
			downloadLink, err := url.Parse(link)
			if err != nil {
				log.Fatal(err)
			} else {
				movie.DownloadLink = downloadLink
			}
			if !(strings.HasSuffix(movie.DownloadLink.String(), "mp4") || strings.HasSuffix(movie.DownloadLink.String(), "mp4")){
				downloadCollector.Visit(downloadLink.String())
			}
		})
	}

	// Update Download Link if "Download" HTML on page
	downloadCollector.OnHTML("div.filedownload", func(e *colly.HTMLElement) {
		movie := &(*movies)[getMovieIndexFromCtx(e.Request)]
		re := regexp.MustCompile(`(.* MB)`)
		size := re.FindStringSubmatch(e.ChildText("textcolor2"))[0]
		if e.ChildAttr("a[id=flink1]", "href") !=  "" {
			downloadLink, err := url.Parse(e.ChildAttr("a[id=flink1]", "href"))
			if err == nil && downloadLink.String() != "" {
				movie.DownloadLink = downloadLink
			}
		} else {
			downloadLink, err := url.Parse(e.ChildAttr("input[name=filelink]", "value"))
			if err == nil && downloadLink.String() != "" {
				movie.DownloadLink = downloadLink
			}
		}
		movie.Size = size
	})
}

// List : list all the movies on a page
func (engine *TvSeriesEngine) List(page int) SearchResult {
	engine.mode = ListMode
	result := SearchResult{
		Query: "Series From A to Z latest episode each - Page " + strconv.Itoa(page),
	}
	q := engine.ListURL.Query()
	q.Set("alpha", "AtoZ")
	q.Set("pg", strconv.Itoa(page))
	engine.ListURL.RawQuery = q.Encode()
	movies, err := Scrape(engine)
	if err != nil {
		log.Fatal(err)
	}
	result.Movies = movies
	return result
}

// Search : Searches tvseries for a particular query and return an array of movies
func (engine *TvSeriesEngine) Search(param ...string) SearchResult {
	query := param[0]
	engine.mode = SearchMode
	result := SearchResult{
		Query: query,
	}
	q := engine.SearchURL.Query()
	q.Set("search", query)
	q.Set("beginsearch", "Search")
	q.Set("vsearch", "")
	q.Set("by", "episodes")
	if len(param) > 1 {
		q.Set("pg", param[1])
	}
	engine.SearchURL.RawQuery = q.Encode()
	movies, err := Scrape(engine)
	if err != nil {
		log.Fatal(err)
	}
	result.Movies = movies
	return result
}
