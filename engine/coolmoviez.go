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

// CoolMoviez : An Engine for CoolMoviez
type CoolMoviez struct {
	Props
}

// NewCoolMoviezEngine : create a new engine for scraping mynewcoolmovies
func NewCoolMoviezEngine() *CoolMoviez {
	base := "https://coolmoviez.buzz"
	baseURL, err := url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}
	// Search URL
	searchURL, err := url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}
	searchURL.Path = "/mobile/search"

	// List URL
	listURL, err := url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}
	listURL.Path = "/movielist/13/Hollywood_movies/default"

	coolMoviesEngine := CoolMoviez{}
	coolMoviesEngine.Name = "CoolMoviez"
	coolMoviesEngine.BaseURL = baseURL
	coolMoviesEngine.Description = `Self reported best download site for mobile, tablets and pc`
	coolMoviesEngine.SearchURL = searchURL
	coolMoviesEngine.ListURL = listURL
	return &coolMoviesEngine
}

// Engine Interface Methods

func (engine *CoolMoviez) String() string {
	st := fmt.Sprintf("%s (%s)", engine.Name, engine.BaseURL)
	return st
}

func (engine *CoolMoviez) getParseAttrs() (string, string, error) {
	return "div.list", "div.fl", nil
}

func (engine *CoolMoviez) parseSingleMovie(el *colly.HTMLElement, index int) (Movie, error) {
	movie := Movie{
		Index:    index,
		IsSeries: false,
		Source:   engine.Name,
	}
	movie.Title = strings.TrimSpace(el.ChildText("a.fileName"))
	appendage := el.ChildText("span")
	if strings.HasSuffix(movie.Title, appendage) {
		movie.Title = strings.TrimSuffix(movie.Title, appendage)
	}
	re := regexp.MustCompile(`(\d+)`)
	stringsub := re.FindStringSubmatch(movie.Title)
	if len(stringsub) > 0 {
		year, _ := strconv.Atoi(stringsub[0])
		movie.Year = year
	}
	downloadLink, err := url.Parse(el.Request.AbsoluteURL(el.ChildAttr("a", "href")))
	movie.CoverPhotoLink = el.ChildAttr("img", "src")

	if err != nil {
		log.Fatal(err)
	}

	movie.DownloadLink = downloadLink
	return movie, nil
}

func (engine *CoolMoviez) updateDownloadProps(downloadCollector *colly.Collector, movies *[]Movie) {

	downloadCollector.OnHTML("div.M1,div.M2", func(e *colly.HTMLElement) {
		reArray := []string{"Quality", "Genre", "Description", "Starcast"}
		movie := &(*movies)[getMovieIndexFromCtx(e.Request)]
		for _, reString := range reArray {
			re := regexp.MustCompile(reString + `:\s+(.*)`)
			stringsub := re.FindStringSubmatch(e.Text)
			if len(stringsub) > 1 {
				value := stringsub[1]
				switch reString {
				case "Quality":
					movie.Quality = value
				case "Description":
					movie.Description = value
				case "Genre":
					movie.Category = value
				case "Starcast":
					movie.Cast = value
				}
			}
		}
	})

	downloadCollector.OnHTML("a.fileName", func(e *colly.HTMLElement) {
		movie := &(*movies)[getMovieIndexFromCtx(e.Request)]
		initialLink := e.Attr("href")
		re := regexp.MustCompile(`Size:\s+(.*)`)
		stringsub := re.FindStringSubmatch(e.Text)
		if len(stringsub) > 1 {
			movie.Size = stringsub[1]
		}

		replacePrefix := "https://www.coolmoviez.buzz/file"
		if strings.HasPrefix(initialLink, replacePrefix) {
			downloadLink, err := url.Parse("https://www.coolmoviez.buzz/server" + strings.TrimPrefix(initialLink, replacePrefix))
			if err == nil {
				movie.DownloadLink = downloadLink
				downloadCollector.Visit(downloadLink.String())
			}
		}
	})

	downloadCollector.OnHTML("a.dwnLink", func(e *colly.HTMLElement) {
		movie := &(*movies)[getMovieIndexFromCtx(e.Request)]
		downloadLink, err := url.Parse(e.Attr("href"))
		if err == nil {
			movie.DownloadLink = downloadLink
		}
	})
}

// List : list all the movies on a page
func (engine *CoolMoviez) List(page int) SearchResult {
	engine.mode = ListMode
	result := SearchResult{
		Query: "List of Recent Uploads - Page " + strconv.Itoa(page),
	}
	pageParam := fmt.Sprintf("%v.html", strconv.Itoa(page))
	engine.ListURL.Path = path.Join(engine.ListURL.Path, pageParam) + "/"
	movies, err := Scrape(engine)
	if err != nil {
		log.Fatal(err)
	}
	result.Movies = movies
	return result
}

// Search : Searches fzmovies for a particular query and return an array of movies
func (engine *CoolMoviez) Search(param ...string) SearchResult {
	query := param[0]
	engine.mode = SearchMode
	result := SearchResult{
		Query: query,
	}
	q := engine.SearchURL.Query()
	q.Set("find", query)
	q.Set("per_page", "1")
	engine.SearchURL.RawQuery = q.Encode()
	movies, err := Scrape(engine)
	if err != nil {
		log.Fatal(err)
	}
	result.Movies = movies
	return result
}
