package engine

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	log "github.com/sirupsen/logrus"
)

// TakanimeList : An Engine for TakanimeList
type TakanimeList struct {
	Props
}

// NewTakanimeListEngine : create a new engine for scraping latest anime from chia-anime
func NewTakanimeListEngine() *TakanimeList {
	base := "https://takanimelist.live"
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
	listURL.Path = "/"

	takanimeListEngine := TakanimeList{}
	takanimeListEngine.Name = "TakanimeList"
	takanimeListEngine.BaseURL = baseURL
	takanimeListEngine.Description = `Anime in 480p, 720p and 1080p format`
	takanimeListEngine.SearchURL = searchURL
	takanimeListEngine.ListURL = listURL
	return &takanimeListEngine
}

// Engine Interface Methods

func (engine *TakanimeList) String() string {
	st := fmt.Sprintf("%s (%s)", engine.Name, engine.BaseURL)
	return st
}

func (engine *TakanimeList) getParseAttrs() (string, string, error) {
	var main, section string
	switch engine.mode {
	case SearchMode:
		main = "main.site-main"
		section = "article.post"
	case ListMode:
		main = "div.grid-plus-inner"
		//    section = "div.grid-items"
		section = ".grid-post-item"
	}
	return main, section, nil
}

func (engine *TakanimeList) parseSingleMovie(el *colly.HTMLElement, index int) (Movie, error) {
	movie := Movie{
		Index:    index,
		IsSeries: true,
		Source:   engine.Name,
		Size:     "---MB",
	}
	if engine.mode == ListMode {
		movie.Title = strings.TrimSpace(el.ChildText("div.title"))
		movie.CoverPhotoLink = el.ChildAttr("div.thumbnail-image", "data-img")
		movie.Description = strings.TrimSpace(el.ChildText("div.excerpt"))
	} else {
		movie.CoverPhotoLink = el.ChildAttr("img", "src")
		movie.Title = strings.TrimSpace(el.ChildAttr("img", "alt"))
		movie.Description = strings.TrimSpace(el.ChildText("span.entry-excerpt"))
	}
	downloadLink, err := url.Parse(el.Request.AbsoluteURL(el.ChildAttr("a", "href")))

	if err != nil {
		log.Fatal(err)
	}
	log.Debug(movie.Title)
	movie.DownloadLink = downloadLink
	return movie, nil
}

func retrieveSingle(downloadCollector *colly.Collector, initialLink string) string {
	var finalLink string

	downloadCollector.OnHTML(`script[language="Javascript"]`, func(e *colly.HTMLElement) {

		re := regexp.MustCompile(`window.open\(\.+\)`)
		stringsub := re.FindStringSubmatch(e.Text)
		log.Debug(stringsub)
		finalLink = stringsub[0]
	})
	if !strings.HasSuffix(strings.ToLower(initialLink), ".mkv") && !strings.HasSuffix(strings.ToLower(initialLink), ".mp4") {
		downloadCollector.Visit(initialLink)
	} else {
		finalLink = initialLink
	}
	return finalLink
}

func (engine *TakanimeList) updateDownloadProps(downloadCollector *colly.Collector, movies *[]Movie) {
	internaldownloadCollector := downloadCollector.Clone()
	downloadCollector.OnHTML("div.entry-content", func(e *colly.HTMLElement) {
		movie := &(*movies)[getMovieIndexFromCtx(e.Request)]
		episodeMap := map[string]*url.URL{}
		linkArray := e.ChildAttrs("a", "href")
		titleArray := e.ChildTexts("a")

		for i := range linkArray {
			movie.DownloadLink, _ = url.Parse(linkArray[i])
			finalLink := retrieveSingle(internaldownloadCollector, linkArray[i])
			downloadLink, _ := url.Parse(finalLink)
			if strconv.Itoa(i) != "" && downloadLink.String() != "" {
				episodeMap[titleArray[i]] = downloadLink
			}
		}
		movie.SDownloadLink = episodeMap
	})
}

// List : list all the movies on a page
func (engine *TakanimeList) List(page int) SearchResult {
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

// Search : Searches takanimelist for a particular query and return an array of movies
func (engine *TakanimeList) Search(param ...string) SearchResult {
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
