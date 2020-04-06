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

// FzEngine : An Engine for FzMovies
type FzEngine struct {
	Props
}

// NewFzEngine : A Movie Engine Constructor for FzEngine
func NewFzEngine() *FzEngine {
	base := "https://www.fzmovies.net/"
	baseURL, err := url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}
	// Search URL
	searchURL, err := url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}
	searchURL.Path = "/csearch.php"

	// List URL
	listURL, err := url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}
	listURL.Path = "/movieslist.php"

	fzEngine := FzEngine{}
	fzEngine.Name = "FzMovies"
	fzEngine.BaseURL = baseURL
	fzEngine.Description = `FzMovies is a site where you can find Bollywood, Hollywood and DHollywood Movies.`
	fzEngine.SearchURL = searchURL
	fzEngine.ListURL = listURL
	return &fzEngine
}

// Engine Interface Methods

func (engine *FzEngine) String() string {
	st := fmt.Sprintf("%s (%s)", engine.Name, engine.BaseURL)
	return st
}

func (engine *FzEngine) getParseAttrs() (string, string, error) {
	return "body", "div.mainbox", nil
}

func (engine *FzEngine) parseSingleMovie(el *colly.HTMLElement, movieIndex int) (Movie, error) {
	movie := Movie{
		Index:    movieIndex,
		IsSeries: false,
		Source:   engine.Name,
	}
	cover, err := url.Parse(el.Request.AbsoluteURL(el.ChildAttr("img", "src")))
	if err != nil {
		log.Fatal(err)
	}
	movie.CoverPhotoLink = cover.String()
	// Remove all Video: or Movie: Prefixes
	movie.UploadDate = strings.TrimSpace(el.ChildTexts("small")[1])
	// Update Year
	year, err := strconv.Atoi(strings.TrimSpace(el.ChildTexts("small")[1]))
	if err == nil {
		movie.Year = year
	}
	movie.Title = strings.TrimSuffix(strings.TrimSpace(el.ChildText("b")), "<more>")
	movie.Description = strings.TrimSpace(el.ChildTexts("small")[3])
	downloadLink, err := url.Parse(el.Request.AbsoluteURL(el.ChildAttr("a", "href")))

	if err != nil {
		log.Fatal(err)
	}
	// download link is current link path + /download
	downloadLink.Path = path.Join(engine.BaseURL.Path, downloadLink.Path)

	movie.DownloadLink = downloadLink
	return movie, nil
}

func (engine *FzEngine) updateDownloadProps(downloadCollector *colly.Collector, scrapedMovies *scraped) {
	// Update movie download link if ul.downloadlinks on page
	downloadCollector.OnHTML("ul.moviesfiles", func(e *colly.HTMLElement) {
		movie := getMovieFromMovies(e.Request, scrapedMovies)
		link := strings.Replace(e.ChildAttr("a", "href"), "download1.php", "download.php", 1)
		downloadLink, err := url.Parse(e.Request.AbsoluteURL(link + "&pt=jRGarGzOo2"))
		//    downloadLink, err := url.Parse(e.ChildAttr("a", "href") + "&pt=jRGarGzOo2")
		if err != nil {
			log.Fatal(err)
		}

		scrapedMovies.Lock()
		defer scrapedMovies.Unlock()
		movie.DownloadLink = downloadLink
		re := regexp.MustCompile(`(.* MB)`)
		dl := strings.TrimPrefix(re.FindStringSubmatch(e.ChildText("dcounter"))[0], "(")
		movie.Size = dl
		downloadCollector.Visit(downloadLink.String())
	})

	// Update Download Link if "Download" HTML on page
	downloadCollector.OnHTML("input[name=download1]", func(e *colly.HTMLElement) {
		if strings.HasSuffix(strings.TrimSpace(e.Attr("value")), "mp4") {
			downloadLink, err := url.Parse(e.Request.AbsoluteURL(e.Attr("value")))
			if err != nil {
				log.Fatal(err)
			}
			movie := getMovieFromMovies(e.Request, scrapedMovies)
			scrapedMovies.Lock()
			defer scrapedMovies.Unlock()
			movie.DownloadLink = downloadLink
		}
	})
}

// List : list all the movies on a page
func (engine *FzEngine) List(page int) SearchResult {
	engine.mode = ListMode
	result := SearchResult{
		Query: "List of Recent Uploads - Page " + strconv.Itoa(page),
	}
	q := engine.ListURL.Query()
	q.Set("catID", "2")
	q.Set("by", "date")
	q.Set("pg", strconv.Itoa(page))
	engine.ListURL.RawQuery = q.Encode()
	movies, err := Scrape(engine)
	if err != nil {
		log.Fatal(err)
	}
	result.Movies = movies
	return result
}

// Search : Searches netnaija for a particular query and return an array of movies
func (engine *FzEngine) Search(query string) SearchResult {
	engine.mode = SearchMode
	result := SearchResult{
		Query: query,
	}
	q := engine.SearchURL.Query()
	q.Set("searchname", query)
	engine.SearchURL.RawQuery = q.Encode()
	movies, err := Scrape(engine)
	if err != nil {
		log.Fatal(err)
	}
	result.Movies = movies
	return result
}
