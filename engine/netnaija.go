package engine

import (
	"fmt"
	"net/url"
	"path"
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
	netNaijaEngine.Description = ""
	netNaijaEngine.SearchURL = searchURL
	netNaijaEngine.ListURL = listURL
	return &netNaijaEngine
}

// Engine Interface Methods

func (engine *NetNaijaEngine) String() string {
	st := fmt.Sprintf("%s (%s)", engine.Name, engine.BaseURL)
	return st
}

// Scrape : does the initial scraping, passed from List or Search
func (engine *NetNaijaEngine) Scrape(mode string) ([]Movie, error) {
	var (
		article string
		title   string
		url     *url.URL
	)
	// When in search mode, results are in <article class="result">
	switch mode {
	case "search":
		article = "article.result"
		title = "h3.result-title"
		url = engine.SearchURL
	case "list":
		article = "article.a-file"
		title = "h3.file-name"
		url = engine.ListURL
	default:
		return []Movie{}, fmt.Errorf("Invalid mode %v", mode)
	}

	movieIndex := 0
	c := colly.NewCollector(
		// Cache responses to prevent multiple download of pages
		// even if the collector is restarted
		colly.CacheDir("./gophie_cache"),
	)
	// Another collector for download Links
	downloadLinkCollector := c.Clone()

	var movies []Movie

	c.OnHTML("main", func(e *colly.HTMLElement) {
		e.ForEach(article, func(_ int, el *colly.HTMLElement) {
			movie := Movie{
				Index:    movieIndex,
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
		//    movie := &movies[getMovieIndexFromCtx(r.Request)]
		if !strings.Contains(r.Headers.Get("Content-Type"), "text") {
			r.Request.Abort()
			log.Debugf("Response %s is not text/html. Aborting request", r.Request.URL)
		}
	})

	downloadLinkCollector.OnResponse(func(r *colly.Response) {
		movie := &movies[getMovieIndexFromCtx(r.Request)]
		log.Debugf("Retrieved Download Link %v\n", movie.DownloadLink)
	})

	// Update movie size
	downloadLinkCollector.OnHTML("button[id=download-button]", func(e *colly.HTMLElement) {
		movies[getMovieIndexFromCtx(e.Request)].Size = strings.TrimSpace(e.ChildText("span.size"))
	})

	downloadLinkCollector.OnHTML("h3.file-name", func(e *colly.HTMLElement) {
		downloadlink, _ := url.Parse(path.Join(strings.TrimSpace(e.ChildAttr("a", "href")), "download"))
		movies[getMovieIndexFromCtx(e.Request)].DownloadLink = downloadlink
		downloadLinkCollector.Visit(downloadlink.String())
	})

	// Update movie download link if a[id=download] on page
	downloadLinkCollector.OnHTML("a[id=download]", func(e *colly.HTMLElement) {
		movie := &movies[getMovieIndexFromCtx(e.Request)]
		movie.Size = strings.TrimSpace(e.ChildText("span[id=download-size]"))
		downloadLink, err := url.Parse(e.Attr("href"))
		if err != nil {
			log.Fatal(err)
		}
		movie.DownloadLink = downloadLink
	})

	// Update Download Link if "Direct Download" HTML on page
	downloadLinkCollector.OnHTML("div.row", func(e *colly.HTMLElement) {
		if strings.TrimSpace(e.ChildText("label")) == "Direct Download" {
			downloadLink, err := url.Parse(e.ChildAttr("input", "value"))
			if err != nil {
				log.Fatal(err)
			}
			movies[getMovieIndexFromCtx(e.Request)].DownloadLink = downloadLink
		}
	})

	//for series or parts
	downloadLinkCollector.OnHTML("div.video-series-latest-episodes", func(inn *colly.HTMLElement) {
		movie := &movies[getMovieIndexFromCtx(inn.Request)]
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

	c.Visit(url.String())

	return movies, nil
}

// List : list all the movies on a page
func (engine *NetNaijaEngine) List(page int) SearchResult {
	result := SearchResult{
		Query: "List of Recent Uploads - Page " + strconv.Itoa(page),
	}
	pageParam := fmt.Sprintf("page/%v", strconv.Itoa(page))
	engine.ListURL.Path = path.Join(engine.ListURL.Path, pageParam)
	movies, err := engine.Scrape("list")
	if err != nil {
		log.Fatal(err)
	}
	result.Movies = movies
	return result
}

// Search : Searches netnaija for a particular query and return an array of movies
func (engine *NetNaijaEngine) Search(query string) SearchResult {
	result := SearchResult{
		Query: query,
	}
	q := engine.SearchURL.Query()
	q.Set("t", query)
	q.Set("folder", "videos")
	engine.SearchURL.RawQuery = q.Encode()
	movies, err := engine.Scrape("search")
	if err != nil {
		log.Fatal(err)
	}
	result.Movies = movies
	return result
}
