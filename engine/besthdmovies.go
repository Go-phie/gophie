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

// BestHDEngine : An Engine for BestHDMovies
type BestHDEngine struct {
	Props
}

// NewBestHDEngine : A Movie Engine Constructor for BestHDEngine
func NewBestHDEngine() *BestHDEngine {
	base := "https://www.besthdmovies.top/"
	baseURL, err := url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}
	// Search URL
	searchURL, err := url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}

	// List URL
	listURL, err := url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}
	listURL.Path = "/new-hd-movies/"

	bestEngine := BestHDEngine{}
	bestEngine.Name = "BestHDMovies"
	bestEngine.BaseURL = baseURL
	bestEngine.Description = `BestHDMovies is a site where you can find high quality Hollywood and Bollywood mkv movies`
	bestEngine.SearchURL = searchURL
	bestEngine.ListURL = listURL
	return &bestEngine
}

// Engine Interface Methods

func (engine *BestHDEngine) String() string {
	st := fmt.Sprintf("%s (%s)", engine.Name, engine.BaseURL)
	return st
}

func (engine *BestHDEngine) getParseAttrs() (string, string, error) {
	return "body", "article.latestPost", nil
}

func (engine *BestHDEngine) parseSingleMovie(el *colly.HTMLElement, movieIndex int) (Movie, error) {
	movie := Movie{
		Index:    movieIndex,
		IsSeries: false,
		Source:   engine.Name,
	}
	cover, err := url.Parse(el.Request.AbsoluteURL(el.ChildAttr("img", "src")))
	if err != nil {
		log.Fatal(err)
	}
	re := regexp.MustCompile(`\d+`)
	movieYear := re.FindStringSubmatch(el.ChildText("div.categories"))
	if len(movieYear) > 0 {
		movie.Year, _ = strconv.Atoi(movieYear[0])
	}
	movie.CoverPhotoLink = cover.String()
	// Remove all Video: or Movie: Prefixes
	movie.UploadDate = strings.TrimSpace(el.ChildTexts("span.thetime")[0])
	movie.Title = strings.TrimSpace(el.ChildAttr("a", "title"))
	movie.Description = ""
	downloadLink, err := url.Parse(el.ChildAttr("a", "href"))

	if err != nil {
		log.Fatal(err)
	}
	// download link is current link path + /download
	downloadLink.Path = path.Join(engine.BaseURL.Path, downloadLink.Path)

	movie.DownloadLink = downloadLink
	return movie, nil
}

func (engine *BestHDEngine) updateDownloadProps(downloadCollector *colly.Collector, scrapedMovies *scraped) {
	// Update movie download link if div.post-single-content  on page
	downloadCollector.OnHTML("div.post-single-content", func(e *colly.HTMLElement) {
		movie := getMovieFromMovies(e.Request, scrapedMovies)
		log.Debug(movie.Index)
		scrapedMovies.Lock()
		defer scrapedMovies.Unlock()
		ptags := e.ChildTexts("p")
		if ptags[len(ptags)-3] >= ptags[len(ptags)-2] {
			movie.Description = strings.TrimSpace(ptags[len(ptags)-3])
		} else {
			movie.Description = strings.TrimSpace(ptags[len(ptags)-2])
		}
		for index, content := range ptags {
			if strings.HasPrefix(content, "File Size: ") {
				movie.Size = strings.TrimPrefix(content, "File Size: ")
			}
			if index == 7 {
				movie.Description = content
			}
		}
		links := e.ChildAttrs("a", "href")
		for _, link := range links {
			if strings.HasPrefix(link, "https://freeload") {
				downloadlink, err := url.Parse(link)
				if err == nil {
					movie.DownloadLink = downloadlink
					ctx := colly.NewContext()
					ctx.Put("movieIndex", strconv.Itoa(movie.Index))
					downloadCollector.Request("GET", movie.DownloadLink.String(), nil, ctx, nil)
					// downloadCollector.Visit(downloadlink.String())
				} else {
					log.Fatal(err)
				}
			}
		}
	})

	downloadCollector.OnHTML("div.content-area", func(e *colly.HTMLElement) {
		movie := getMovieFromMovies(e.Request, scrapedMovies)
		links := e.ChildAttrs("a", "href")
		for _, link := range links {
			if strings.HasPrefix(link, "https://zeefiles") || strings.HasPrefix(link, "http://zeefiles") {
				// change all http to https
				if strings.HasPrefix(link, "http://") {
					link = "https://" + strings.TrimPrefix(link, "http://")
				}
				downloadlink, err := url.Parse(link)
				if err == nil {
					scrapedMovies.Lock()
					defer scrapedMovies.Unlock()
					movie.DownloadLink = downloadlink
					ctx := colly.NewContext()
					ctx.Put("movieIndex", strconv.Itoa(movie.Index))
					downloadCollector.Request("GET", downloadlink.String(), nil, ctx, nil)
				} else {
					log.Fatal(err)
				}
			}
		}
	})

	downloadCollector.OnHTML("div.freeDownload", func(e *colly.HTMLElement) {
		movie := getMovieFromMovies(e.Request, scrapedMovies)
		zeesubmission := make(map[string]string)
		if e.ChildAttr("a.link_button", "href") != "" {
			downloadlink, err := url.Parse(e.ChildAttr("a.link_button", "href"))
			if err == nil {
				scrapedMovies.Lock()
				defer scrapedMovies.Unlock()
				movie.DownloadLink = downloadlink
			}
		} else {

			inputNames := e.ChildAttrs("input", "name")
			inputValues := e.ChildAttrs("input", "value")

			for index := range inputNames {
				zeesubmission[inputNames[index]] = inputValues[index]
			}
			ctx := colly.NewContext()
			ctx.Put("movieIndex", strconv.Itoa(movie.Index))
			err := downloadCollector.Request("POST", movie.DownloadLink.String(), createFormReader(zeesubmission), ctx, nil)
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	downloadCollector.OnHTML("form[method=post]", func(e *colly.HTMLElement) {
		var err error
		submissionDetails := make(map[string]string)
		movie := getMovieFromMovies(e.Request, scrapedMovies)
		downloadlink := movie.DownloadLink
		inputNames := e.ChildAttrs("input", "name")
		inputValues := e.ChildAttrs("input", "value")

		for index := range inputNames {
			submissionDetails[inputNames[index]] = inputValues[index]
		}
		requestlink := e.Request.URL.String()
		if !(strings.HasPrefix(requestlink, "https://zeefiles") || strings.HasPrefix(requestlink, "http://zeefiles")) {
			downloadlink, err = url.Parse("https://udown.me/watchonline/?movieIndex=" + strconv.Itoa(movie.Index))
			// log.Debug("inside scraper ====>", downloadlink.String())
			if err == nil {
				movie.DownloadLink = downloadlink
			}
			ctx := colly.NewContext()
			ctx.Put("movieIndex", strconv.Itoa(movie.Index))
			err = downloadCollector.Request("POST", downloadlink.String(), createFormReader(submissionDetails), ctx, nil)
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	downloadCollector.OnHTML("video", func(e *colly.HTMLElement) {
		downloadlink := e.ChildAttr("source", "src")
		movie := getMovieFromMovies(e.Request, scrapedMovies)
		scrapedMovies.Lock()
		defer scrapedMovies.Unlock()
		movie.DownloadLink, _ = url.Parse(downloadlink)
	})
}

// List : list all the movies on a page
func (engine *BestHDEngine) List(page int) SearchResult {
	engine.mode = ListMode
	result := SearchResult{
		Query: "List of Recent Uploads - Page " + strconv.Itoa(page),
	}

	pageParam := fmt.Sprintf("page/%v", strconv.Itoa(page))
	engine.ListURL.Path = path.Join(engine.ListURL.Path, pageParam)
	movies, err := Scrape(engine)
	//  log.Debug(movies)
	if err != nil {
		log.Fatal(err)
	}
	result.Movies = movies
	return result
}

// Search : Searches netnaija for a particular query and return an array of movies
func (engine *BestHDEngine) Search(param ...string) SearchResult {
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
