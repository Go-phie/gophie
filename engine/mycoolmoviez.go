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

// MyCoolMoviez : An Engine for MyCoolMoviez
type MyCoolMoviez struct {
	Props
}

// NewMyCoolMoviezEngine : create a new engine for scraping mynewcoolmovies
func NewMyCoolMoviezEngine() *MyCoolMoviez {
	base := "https://mycoolmoviez.website"
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
	listURL.Path = "/hollywood_movies/page"

	coolMoviesEngine := MyCoolMoviez{}
	coolMoviesEngine.Name = "MyCoolMoviez"
	coolMoviesEngine.BaseURL = baseURL
	coolMoviesEngine.Description = `MyCoolMoviez is a site that collects movies from across the web in believed to be in a public domain`
	coolMoviesEngine.SearchURL = searchURL
	coolMoviesEngine.ListURL = listURL
	return &coolMoviesEngine
}

// Engine Interface Methods

func (engine *MyCoolMoviez) String() string {
	st := fmt.Sprintf("%s (%s)", engine.Name, engine.BaseURL)
	return st
}

func (engine *MyCoolMoviez) getParseAttrs() (string, string, error) {
	return "ul.cat_ul", "li", nil
}

func (engine *MyCoolMoviez) parseSingleMovie(el *colly.HTMLElement, index int) (Movie, error) {
	movie := Movie{
		Index:    index,
		IsSeries: false,
		Source:   engine.Name,
	}
	movie.Title = strings.TrimSpace(el.ChildText("a"))
	re := regexp.MustCompile(`(\d+)`)
	stringsub := re.FindStringSubmatch(movie.Title)
	if len(stringsub) > 0 {
		year, _ := strconv.Atoi(stringsub[0])
		movie.Year = year
	}
	downloadLink, err := url.Parse(el.Request.AbsoluteURL(el.ChildAttr("a", "href")))

	if err != nil {
		log.Fatal(err)
	}

	movie.DownloadLink = downloadLink
	return movie, nil
}

func (engine *MyCoolMoviez) updateDownloadProps(downloadCollector *colly.Collector, movies *[]Movie) {
	downloadCollector.OnHTML("img.movie-poster", func(e *colly.HTMLElement) {
		movie := &(*movies)[getMovieIndexFromCtx(e.Request)]
		coverphotolink, err := url.Parse(e.Attr("src"))
		if err != nil {
			log.Fatal(err)
		}
		movie.CoverPhotoLink = coverphotolink.String()
	})

	downloadCollector.OnHTML("div.panel-body", func(e *colly.HTMLElement) {
		var genre string
		movie := &(*movies)[getMovieIndexFromCtx(e.Request)]
		listTexts := e.ChildTexts("li")
		for _, text := range listTexts {
			if strings.HasPrefix(text, "Description :") {
				movie.Description = strings.TrimSpace(strings.TrimPrefix(text, "Description :"))
			}
			if strings.HasPrefix(text, "Genre :") {
				genre = strings.TrimSpace(strings.TrimPrefix(text, "Genre :"))
			}
		}
		movie.Description = movie.Description + `\n` + genre
	})

	downloadCollector.OnHTML("div.download", func(e *colly.HTMLElement) {
		movie := &(*movies)[getMovieIndexFromCtx(e.Request)]
		listHrefs := e.ChildAttrs("a", "href")
		for _, link := range listHrefs {
			if strings.HasPrefix(link, "https://") {
				downloadlink, _ := url.Parse(link)
				movie.DownloadLink = downloadlink
			}
		}
		re := regexp.MustCompile(`(\d+)[\s]?MB`)
		stringsub := re.FindStringSubmatch(e.ChildText("span"))
		if len(stringsub) > 0 {
			movie.Size = strings.Replace(stringsub[0], " ", "", -1)
		}
		log.Debug(movie.Size)
		downloadCollector.Visit(movie.DownloadLink.String())
	})

	downloadCollector.OnHTML(`a[rel="nofollow"]`, func(e *colly.HTMLElement) {
		movie := &(*movies)[getMovieIndexFromCtx(e.Request)]
		if strings.HasPrefix(e.Attr("title"), "Download from") {
			downloadLink, _ := url.Parse(e.Attr("href"))
			movie.DownloadLink = downloadLink
		}
	})
}

// List : list all the movies on a page
func (engine *MyCoolMoviez) List(page int) SearchResult {
	engine.mode = ListMode
	result := SearchResult{
		Query: "List of Recent Uploads - Page " + strconv.Itoa(page),
	}
	pageParam := fmt.Sprintf("%v/", strconv.Itoa(page-1))
	engine.ListURL.Path = path.Join(engine.ListURL.Path, pageParam) + "/"
	movies, err := Scrape(engine)
	if err != nil {
		log.Fatal(err)
	}
	result.Movies = movies
	return result
}

// Search : Searches fzmovies for a particular query and return an array of movies
func (engine *MyCoolMoviez) Search(param ...string) SearchResult {
	query := param[0]
	engine.mode = SearchMode
	result := SearchResult{
		Query: query,
	}
	q := engine.SearchURL.Query()
	q.Set("movie", query)
	engine.SearchURL.RawQuery = q.Encode()
	movies, err := Scrape(engine)
	if err != nil {
		log.Fatal(err)
	}
	result.Movies = movies
	return result
}
