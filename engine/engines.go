package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-phie/gophie/transport"
	"github.com/gocolly/colly/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Mode : The mode of operation for scraping
type Mode int

// Dynamically assign Selenium URL
var ()

const (
	// SearchMode : in this mode a query is searched for
	SearchMode Mode = iota
	// ListMode : in this mode a page is looked up
	ListMode
)

func (m Mode) String() string {
	return [...]string{"Search", "List"}[m]
}

// Engine : interface for all engines
type Engine interface {
	getName() string
	getParseURL() *url.URL
	Search(param ...string) SearchResult
	List(page int) SearchResult
	String() string

	// parseSingleMovie: parses the result of a colly HTMLElement and returns a movie
	// The input el is usually the block of code from the article specified in getParseAttrs
	parseSingleMovie(el *colly.HTMLElement, index int) (Movie, error)

	// getParseAttrs : get the attributes to use to parse a returned soup
	// the first return string is the part of the html to be parsed e.g `body`, `main`
	// the second return string is the attributes to be used in parsing the element specified
	// by the first return. For example returning main, article.sr-one results in parsing the main
	// tag and parsing all article.sr-one within the main html
	getParseAttrs() (string, string, error)

	// parseSingleMovie: parses the result of a colly HTMLElement and returns a movie
	updateDownloadProps(downloadCollector *colly.Collector, movies *[]Movie)
}

// Scrape : Parse queries a url and return results
func Scrape(engine Engine) ([]Movie, error) {
	// Config Vars
	//  seleniumURL := fmt.Sprintf("%s/wd/hub", viper.GetString("selenium-url"))
	cacheDir := viper.GetString("cache-dir")
	ignoreCache := viper.GetBool("ignore-cache")
	var (
		t   *transport.ChromeDpTransport
		err error
		c   *colly.Collector
	)

	if ignoreCache {
		c = colly.NewCollector()

	} else {
		c = colly.NewCollector(
			// Cache responses to prevent multiple download of pages
			// even if the collector is restarted
			colly.CacheDir(cacheDir),
		)
	}

	// Add Cloud Flare scraper bypasser
	if engine.getName() == "NetNaija" {
		log.Debug("Switching to ChromeDpTransport")
		t, err = transport.NewChromeDpTransport(http.DefaultTransport)
		if err != nil {
			log.Fatal(err)
		}

		c.WithTransport(t)
	}
	// Close the WebDriver Instance
	defer func() {
		if engine.getName() == "NetNaija" {
			t.RemoteAllocCancel()
			t.Cancel()
		}
	}()

	// Another collector for download Links
	downloadLinkCollector := c.Clone()

	movieIndex := 0
	var movies []Movie

	// Any Extras setup for downloads using can be specified in the function
	engine.updateDownloadProps(downloadLinkCollector, &movies)

	main, article, err := engine.getParseAttrs()
	if err != nil {
		log.Fatal(err)
	}

	//  c.OnHTML("div", func(e *colly.HTMLElement) {
	//    log.Debugf("%#v", e)
	//  })

	c.OnHTML(main, func(e *colly.HTMLElement) {
		e.ForEach(article, func(_ int, el *colly.HTMLElement) {
			movie, err := engine.parseSingleMovie(el, movieIndex)
			if err != nil {
				log.Errorf("%v could not be parsed", movie)
			} else {
				movies = append(movies, movie)
				downloadLinkCollector.Visit(movie.DownloadLink.String())
				movieIndex++
			}
		})
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html")
		log.Debugf("Visiting %v", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		log.Debugf("Done %v", r.Request.URL.String())
	})

	// Attach Movie Index to Context before making visits
	// Adding Movie Index to context ensures we can fetch a reference to the
	// movie details when we need it
	downloadLinkCollector.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml")
		for i, movie := range movies {
			if movie.DownloadLink.String() == r.URL.String() {
				log.Debugf("Retrieving Download Link %v\n", movie.DownloadLink)
				r.Ctx.Put("movieIndex", strconv.Itoa(i))
			}
		}
	})

	// If Response Content Type is not Text, Abort the Request to prevent fully downloading the
	// body in case of other types like mp4
	downloadLinkCollector.OnResponseHeaders(func(r *colly.Response) {
		if !strings.Contains(r.Headers.Get("Content-Type"), "text") {
			r.Request.Abort()
			log.Debugf("Response %s is not text/html. Aborting request", r.Request.URL)
		}
	})

	downloadLinkCollector.OnResponse(func(r *colly.Response) {
		movie := &movies[getMovieIndexFromCtx(r.Request)]
		log.Debugf("Retrieved Download Link %v\n", movie.DownloadLink)
	})
	c.Visit(engine.getParseURL().String())
	return movies, nil
}

// Movie : the structure of all downloadable movies
type Movie struct {
	Index          int
	Title          string
	CoverPhotoLink string
	Description    string
	Size           string
	DownloadLink   *url.URL
	Year           int
	IsSeries       bool
	SDownloadLink  map[string]*url.URL // Other links for downloads if movies is series
	Quality        string
	Category       string // csv of categories
	Cast           string // csv of actors in movie
	UploadDate     string
	Source         string              // The Engine From which it is gotten from
	SubtitleLink   *url.URL            // single subtitle link
	SubtitleLinks  map[string]*url.URL // Subtitle links for a series
	ImdbLink       string              // imdb link if available
	Tags           string              // csv of words that are linked to the movie if available
}

// MovieJSON : JSON structure of all downloadable movies
type MovieJSON struct {
	Movie
	DownloadLink  string
	SDownloadLink map[string]string
	SubtitleLinks map[string]string
}

func (m *Movie) String() string {
	return fmt.Sprintf("%s (%v)", m.Title, m.Year)
}

// MarshalJSON Json structure to return from api
func (m *Movie) MarshalJSON() ([]byte, error) {
	sDownloadLink := make(map[string]string)
	for key, val := range m.SDownloadLink {
		sDownloadLink[key] = val.String()
	}
	subtitleLinks := make(map[string]string)
	for key, val := range m.SubtitleLinks {
		subtitleLinks[key] = val.String()
	}

	movie := MovieJSON{
		Movie:         *m,
		DownloadLink:  m.DownloadLink.String(),
		SDownloadLink: sDownloadLink,
		SubtitleLinks: subtitleLinks,
	}

	return json.Marshal(movie)

}

// SearchResult : the results of search from engine
type SearchResult struct {
	Query  string
	Movies []Movie
}

// Titles : Get a slice of the titles of movies
func (s *SearchResult) Titles() []string {
	var titles []string
	for _, movie := range s.Movies {
		titles = append(titles, movie.Title)
	}
	return titles
}

// GetMovieByTitle : Return a movie object from title passed
func (s *SearchResult) GetMovieByTitle(title string) (Movie, error) {
	for _, movie := range s.Movies {
		if movie.Title == title {
			return movie, nil
		}
	}
	return Movie{}, errors.New("Movie not Found")
}

// GetIndexFromTitle : return movieIndex from title
func (s *SearchResult) GetIndexFromTitle(title string) (int, error) {
	for index, movie := range s.Movies {
		if movie.Title == title {
			return index, nil
		}
	}
	return 0, errors.New("Movie not Found")
}

// GetEngines : Returns all the usable engines in the application
func GetEngines() map[string]Engine {
	engines := make(map[string]Engine)
	engines["netnaija"] = NewNetNaijaEngine()
	engines["fzmovies"] = NewFzEngine()
	engines["besthdmovies"] = NewBestHDEngine()
	engines["tvseries"] = NewTvSeriesEngine()
	engines["mycoolmoviez"] = NewMyCoolMoviezEngine()
	engines["coolmoviez"] = NewCoolMoviezEngine()
	engines["animeout"] = NewAnimeOutEngine()
	engines["takanimelist"] = NewTakanimeListEngine()
	engines["kdramahood"] = NewKDramaHoodEngine()
	engines["nkiri"] = NewNkiriEngine()
	return engines
}

// GetEngine : Return an engine
func GetEngine(engine string) (Engine, error) {
	e := GetEngines()[strings.ToLower(engine)]
	if e == nil {
		return nil, fmt.Errorf("Engine %s Does not exist", engine)
	}
	return e, nil
}

// Get the movie index context stored in Request
func getMovieIndexFromCtx(r *colly.Request) int {
	movieIndex, err := strconv.Atoi(r.Ctx.Get("movieIndex"))
	if err != nil {
		log.Fatal(err)
	}
	return movieIndex
}

// Get all form details into a neat map
func getFormDetails(element *colly.HTMLElement) map[string]string {
	submission := make(map[string]string)
	inputNames := element.ChildAttrs("input", "name")
	inputValues := element.ChildAttrs("input", "value")

	for index := range inputNames {
		submission[inputNames[index]] = inputValues[index]
	}
	return submission
}
