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

// NetNaijaEngine : An Engine for  NetNaija
type NetNaijaEngine struct {
	Props
}

// NewNetNaijaEngine : A Movie Engine Constructor for NetNaija
func NewNetNaijaEngine() *NetNaijaEngine {
	base := "https://www.thenetnaija.com/"
	baseURL, err := url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}
	// Search URL
	searchURL, err := url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}
	searchURL.Path = "/search"

	// List URL
	listURL, err := url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}
	listURL.Path = "/videos/movies/"

	netNaijaEngine := NetNaijaEngine{}
	netNaijaEngine.Name = "NetNaija"
	netNaijaEngine.BaseURL = baseURL
	netNaijaEngine.Description = `
			Nigerian forum and media download center. 
			Developed and owned by Analike Emmanuel Bridge`
	netNaijaEngine.SearchURL = searchURL
	netNaijaEngine.ListURL = listURL
	return &netNaijaEngine
}

// Engine Interface Methods

func (engine *NetNaijaEngine) String() string {
	return fmt.Sprintf("%s (%s)", engine.Name, engine.BaseURL)
}

func (engine *NetNaijaEngine) getParseAttrs() (string, string, error) {
	var article string
	// When in search mode, results are in <article class="result">
	switch engine.mode {
	case SearchMode:
		article = "article.result"
	case ListMode:
		article = "article.a-file"
	default:
		return "", "", fmt.Errorf("Invalid mode %v", engine.mode)
	}
	return "main", article, nil
}

func (engine *NetNaijaEngine) parseSingleMovie(el *colly.HTMLElement) (Movie, error) {
	// movie title identifier
	var title string
	if title = "h3.file-name"; engine.mode == SearchMode {
		title = "h3.result-title"
	}

	re := regexp.MustCompile(`\((.*)\)`)
	movie := Movie{
		IsSeries: false,
		Source:   engine.Name,
	}

	movie.CoverPhotoLink = el.ChildAttr("img", "src")
	// Remove all Video: or Movie: Prefixes
	movie.Title = strings.TrimSpace(
		strings.TrimPrefix(
			strings.TrimPrefix(el.ChildText(title), "Movie:"),
			"Video:"))
	movie.UploadDate = strings.TrimSpace(el.ChildText("span.fa-clock-o"))
	movie.Description = strings.TrimSpace(el.ChildText("p.result-desc"))
	downloadLink, err := url.Parse(el.ChildAttr("a", "href"))

	if err != nil {
		log.Fatal(err)
	}
	// download link is current link path + /download
	downloadLink.Path = path.Join(downloadLink.Path, "download")

	if strings.HasPrefix(downloadLink.Path, "/videos/series") {
		movie.IsSeries = true
	}
	movie.DownloadLink = downloadLink
	if movie.Title != "" {
		year := re.FindStringSubmatch(movie.Title)
		if len(year) > 1 {
			intYear, err := strconv.Atoi(year[1])
			if err == nil {
				movie.Year = intYear
			}
		}
	}
	return movie, nil
}

func (engine *NetNaijaEngine) updateDownloadProps(downloadCollector *colly.Collector, movies map[string]*Movie) {
	// Update movie size
	downloadCollector.OnHTML("button[id=download-button]", func(e *colly.HTMLElement) {
		movie := getMovieFromMovies(e.Request.URL.String(), movies)
		movie.Size = strings.TrimSpace(e.ChildText("span.size"))
	})

	downloadCollector.OnHTML("h3.file-name", func(e *colly.HTMLElement) {
		downloadLink, err := url.Parse(path.Join(strings.TrimSpace(e.ChildAttr("a", "href")), "download"))
		if err != nil {
			log.Fatal(err)
		}
		movie := getMovieFromMovies(e.Request.URL.String(), movies)
		movie.DownloadLink = downloadLink
		downloadCollector.Visit(downloadLink.String())
	})

	// Update movie download link if a[id=download] on page
	downloadCollector.OnHTML("a[id=download]", func(e *colly.HTMLElement) {
		movie := getMovieFromMovies(e.Request.URL.String(), movies)
		movie.Size = strings.TrimSpace(e.ChildText("span[id=download-size]"))
		downloadLink, err := url.Parse(e.Attr("href"))
		if err != nil {
			log.Fatal(err)
		}
		movie.DownloadLink = downloadLink
	})

	// Update Download Link if "Direct Download" HTML on page
	downloadCollector.OnHTML("div.row", func(e *colly.HTMLElement) {
		if strings.TrimSpace(e.ChildText("label")) == "Direct Download" {
			downloadLink, err := url.Parse(e.ChildAttr("input", "value"))
			if err != nil {
				log.Fatal(err)
			}
			movie := getMovieFromMovies(e.Request.URL.String(), movies)
			movie.DownloadLink = downloadLink
		}
	})

	//for series or parts
	downloadCollector.OnHTML("div.video-series-latest-episodes", func(inn *colly.HTMLElement) {
		movie := getMovieFromMovies(inn.Request.URL.String(), movies)
		movie.IsSeries = true
		inn.ForEach("a", func(_ int, e *colly.HTMLElement) {
			downloadLink, err := url.Parse(e.Attr("href"))
			if err != nil {
				log.Fatal(err)
			}
			downloadLink.Path = path.Join(downloadLink.Path, "download")
			movie.SDownloadLink = append(movie.SDownloadLink, downloadLink)
		})
	})
}

// List : list all the movies on a page
func (engine *NetNaijaEngine) List(page int) SearchResult {
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

// Search : Searches netnaija for a particular query and return an array of movies
func (engine *NetNaijaEngine) Search(query string) SearchResult {
	engine.mode = SearchMode
	result := SearchResult{
		Query: query,
	}
	q := engine.SearchURL.Query()
	q.Set("t", query)
	q.Set("folder", "videos")
	engine.SearchURL.RawQuery = q.Encode()
	movies, err := Scrape(engine)
	if err != nil {
		log.Fatal(err)
	}
	result.Movies = movies
	return result
}
