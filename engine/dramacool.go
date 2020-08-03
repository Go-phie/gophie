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

// DramaCool : An Engine for DramaCool
type DramaCool struct {
	Props
}

// NewDramaCoolEngine : create a new engine for scraping latest korean drama
func NewDramaCoolEngine() *DramaCool {
	base := "https://www.dramacool9.co"
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
	listURL.Path = "/category/latest-asian-drama-releases/"

	dramaCoolEngine := DramaCool{}
	dramaCoolEngine.Name = "DramaCool"
	dramaCoolEngine.BaseURL = baseURL
	dramaCoolEngine.Description = `Search from over 1000's of encoded anime available`
	dramaCoolEngine.SearchURL = searchURL
	dramaCoolEngine.ListURL = listURL
	return &dramaCoolEngine
}

// Engine Interface Methods

func (engine *DramaCool) String() string {
	st := fmt.Sprintf("%s (%s)", engine.Name, engine.BaseURL)
	return st
}

func (engine *DramaCool) getParseAttrs() (string, string, error) {
	var (
		article string
		main    string
	)
	switch engine.mode {
	case SearchMode:
		main = "ul.list-thumb"
		article = "li"
	case ListMode:
		main = "div#drama"
		article = "li"
	default:
		return "", "", fmt.Errorf("Invalid mode %v", engine.mode)
	}
	return main, article, nil
}

func (engine *DramaCool) parseSingleMovie(el *colly.HTMLElement, index int) (Movie, error) {
	movie := Movie{
		Index:    index,
		IsSeries: true,
		Source:   engine.Name,
		Size:     "---MB",
	}
	movie.Title = strings.TrimSpace(el.ChildAttr("a", "title"))
	movie.CoverPhotoLink = el.ChildAttr("img", "src")
	re := regexp.MustCompile(`(-episode-.+)`)
	link := el.Request.AbsoluteURL(el.ChildAttr("a", "href"))
	downloadLink, err := url.Parse(link)
	remove := re.FindStringSubmatch(link)

	if len(remove) > 0 {
		downloadLink, err = url.Parse(strings.TrimSuffix(link, remove[0]))
		if err != nil {
			log.Fatal(err)
		}
	}
	movie.DownloadLink = downloadLink
	return movie, nil
}

func (engine *DramaCool) updateDownloadProps(downloadCollector *colly.Collector, movies *[]Movie) {
	downloadCollector.OnHTML("div#all-episodes", func(e *colly.HTMLElement) {
		log.Debug("got into all episodes")
		//    movie := &(*movies)[getMovieIndexFromCtx(e.Request)]
		//    episodeMap := map[string]*url.URL{}
	})

	downloadCollector.OnHTML("div.drama-details", func(e *colly.HTMLElement) {
		log.Debug("got into all episodes")
		//    episodeMap := map[string]*url.URL{}
		movie := &(*movies)[getMovieIndexFromCtx(e.Request)]
		description := e.ChildText("div.synopsis")
		re := regexp.MustCompile(`Genres: (.+)`)
		genres := re.FindStringSubmatch(e.ChildText("p.genres"))
		if len(genres) > 1 {
			movie.Category = genres[1]
		}
		re = regexp.MustCompile(`Release Year: (.+)`)
		year := re.FindStringSubmatch(e.ChildText("p.release-year"))
		if len(year) > 1 {
			year, err := strconv.Atoi(year[1])
			if err == nil {
				movie.Year = year
			}
		}

		re = regexp.MustCompile(`(?s)Synopsis:(.+)Also known as:`)
		desc := re.FindStringSubmatch(description)
		if len(desc) > 1 {
			movie.Description = strings.TrimSpace(desc[1])
		}
		log.Debug(movie.Description)
	})
}

// List : list all the movies on a page
func (engine *DramaCool) List(page int) SearchResult {
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
func (engine *DramaCool) Search(param ...string) SearchResult {
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
