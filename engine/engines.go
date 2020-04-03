package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	//  "github.com/gocolly/colly/v2/debug"
	log "github.com/sirupsen/logrus"
)

// Mode : The mode of operation for scraping
type Mode int

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
	getParseURL() *url.URL
	Search(query string) SearchResult
	List(page int) SearchResult
	String() string
	// parseSingleMovie: parses the result of a colly HTMLElement and returns a movie
	parseSingleMovie(el *colly.HTMLElement) (Movie, error)

	// getParseAttrs : get the attributes to use to parse a returned soup
	// the first return string is the part of the html to be parsed e.g `body`, `main`
	// the second return string is the attributes to be used in parsing the element specified
	// by the first return
	getParseAttrs() (string, string, error)

	// parseSingleMovie: parses the result of a colly HTMLElement and returns a movie
	updateDownloadProps(downloadCollector *colly.Collector, movies map[string]*Movie)
}

// Scrape : Parse queries a url and return results
func Scrape(engine Engine) ([]Movie, error) {
	c := colly.NewCollector(
		// Cache responses to prevent multiple download of pages
		// even if the collector is restarted
		colly.CacheDir("./gophie_cache"),
		// colly.Async(true),
		//    colly.Debugger(&debug.LogDebugger{}),
	)
	// Another collector for download Links
	downloadLinkCollector := colly.NewCollector(
		colly.CacheDir("./gophie-cache"),
		colly.Async(true)
	)

	var movies = make(map[string]*Movie)

	// Any Extras setup for downloads using can be specified in the function
	engine.updateDownloadProps(downloadLinkCollector, movies)

	main, article, err := engine.getParseAttrs()
	if err != nil {
		log.Fatal(err)
	}
	c.OnHTML(main, func(e *colly.HTMLElement) {
		e.ForEach(article, func(_ int, el *colly.HTMLElement) {
			movie, err := engine.parseSingleMovie(el)
			if err != nil {
				log.Errorf("%v could not be parsed", movie)
			} else {
				// Using DownloadLink as key to movie makes it unique
				movies[movie.DownloadLink.String()] = &movie
				// downloadLinkCollector.Visit(movie.DownloadLink.String())
			}
		})

		for _, movie := range movies{
			downloadLinkCollector.Visit(movie.DownloadLink.String())
		}
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
		if movie, ok := movies[r.URL.String()]; ok {
			log.Debugf("Retrieving Download Link %v\n", movie.DownloadLink)
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
		log.Debug(r.Request.URL.String)
		//    movie := movies[r.Request.URL.String()]
		//    log.Infof("%s %v %s", r.Request.URL.String(), movie.DownloadLink, movie.Title)
		//    log.Debugf("Retrieved Download Link %v\n", movie.DownloadLink)
	})

	c.Visit(engine.getParseURL().String())
	c.Wait()
	downloadLinkCollector.Wait()

	// Create a List of Movies
	v := make([]Movie, 0, len(movies))

	for _, value := range movies {
		v = append(v, *value)
	}
	prettyPrint(v)

	return v, nil
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
	SDownloadLink  []*url.URL // Other links for downloads if movies is series
	UploadDate     string
	Source         string // The Engine From which it is gotten from
}

// MovieJSON : JSON structure of all downloadable movies
type MovieJSON struct {
	Movie
	DownloadLink  string
	SDownloadLink []string
}

func (m *Movie) String() string {
	return fmt.Sprintf("%s (%v)", m.Title, m.Year)
}

// MarshalJSON Json structure to return from api
func (m *Movie) MarshalJSON() ([]byte, error) {
	var sDownloadLink []string
	for _, link := range m.SDownloadLink {
		sDownloadLink = append(sDownloadLink, link.String())
	}

	movie := MovieJSON{
		Movie:         *m,
		DownloadLink:  m.DownloadLink.String(),
		SDownloadLink: sDownloadLink,
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

// Get Movie from a URL
func getMovieFromMovies(url string, movies map[string]*Movie) *Movie {
	if _, ok := movies[url]; ok {
		return movies[url]
	}
	for _, movie := range movies {
		if (*movie).DownloadLink.String() == url {
			return movie
		}
	}
	return &Movie{}
}

func prettyPrint(s []Movie) {
	b, err := json.MarshalIndent(s, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
}
