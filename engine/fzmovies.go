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
	fzEngine.Description = ""
	fzEngine.SearchURL = searchURL
	fzEngine.ListURL = listURL
	return &fzEngine
}

// Engine Interface Methods

func (engine *FzEngine) String() string {
	st := fmt.Sprintf("%s (%s)", engine.Name, engine.BaseURL)
	return st
}

// Scrape : does the initial scraping, passed from List or Search
func (engine *FzEngine) Scrape(mode string) ([]Movie, error) {
	var (
		url *url.URL
	)
	// When in search mode, results are in <article class="result">
	switch mode {
	case "search":
		url = engine.SearchURL
	case "list":
		url = engine.ListURL
	default:
		return []Movie{}, fmt.Errorf("Invalid mode %v", mode)
	}

	movieIndex := 0
	c := colly.NewCollector(
	// Cache responses to prevent multiple download of pages
	// even if the collector is restarted
	//    colly.CacheDir("./gophie_cache"),
	)
	// Another collector for download Links
	downloadLinkCollector := c.Clone()

	var movies []Movie

	c.OnHTML("body", func(e *colly.HTMLElement) {
		e.ForEach("div.mainbox", func(_ int, el *colly.HTMLElement) {
			movie := Movie{
				Index:    movieIndex,
				IsSeries: false,
				Source:   engine.Name,
			}
			cover, _ := url.Parse(el.ChildAttr("img", "src"))
			movie.CoverPhotoLink = cover.String()
			// Remove all Video: or Movie: Prefixes
			movie.UploadDate = strings.TrimSpace(el.ChildTexts("small")[1])
			movie.Title = strings.TrimSuffix(strings.TrimSpace(el.ChildText("b")), "<more>")
			movie.Description = strings.TrimSpace(el.ChildTexts("small")[3])
			downloadLink, err := url.Parse(el.ChildAttr("a", "href"))

			if err != nil {
				log.Fatal(err)
			}
			// download link is current link path + /download
			downloadLink.Path = path.Join(engine.BaseURL.Path, downloadLink.Path)

			movie.DownloadLink = downloadLink
			if movie.Title != "" {
				movies = append(movies, movie)
				// FIXME
				// Colly does not visit the first movies for some weird reason
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

	downloadLinkCollector.OnResponse(func(r *colly.Response) {
		movie := &movies[getMovieIndexFromCtx(r.Request)]
		log.Debugf("Retrieved Download Link %v\n", movie.DownloadLink)
	})

	// Update movie download link if ul.downloadlinks on page
	downloadLinkCollector.OnHTML("ul.moviesfiles", func(e *colly.HTMLElement) {
		movie := &movies[getMovieIndexFromCtx(e.Request)]
		link := strings.Replace(e.ChildAttr("a", "href"), "download1.php", "download.php", 1)
		downloadLink, err := url.Parse(link + "&pt=jRGarGzOo2")
		//    downloadLink, err := url.Parse(e.ChildAttr("a", "href") + "&pt=jRGarGzOo2")
		if err != nil {
			log.Fatal(err)
		}
		movie.DownloadLink = downloadLink
		re := regexp.MustCompile(`(.* MB)`)
		dl := strings.TrimPrefix(re.FindStringSubmatch(e.ChildText("dcounter"))[0], "(")
		movie.Size = dl
		downloadLinkCollector.Visit(downloadLink.String())
	})

	// Update Download Link if "Download" HTML on page
	downloadLinkCollector.OnHTML("p", func(e *colly.HTMLElement) {
		if strings.HasSuffix(strings.TrimSpace(e.ChildAttr("input", "value")), "mp4") {
			downloadLink, err := url.Parse(e.ChildAttr("input", "value"))
			if err != nil {
				log.Fatal(err)
			}
			movies[getMovieIndexFromCtx(e.Request)].DownloadLink = downloadLink
		}
	})
	c.Visit(url.String())

	return movies, nil
}

// List : list all the movies on a page
func (engine *FzEngine) List(page int) SearchResult {
	result := SearchResult{
		Query: "List of Recent Uploads - Page " + strconv.Itoa(page),
	}
	q := engine.ListURL.Query()
	q.Set("catID", "2")
	q.Set("by", "date")
	q.Set("pg", strconv.Itoa(page))
	engine.ListURL.RawQuery = q.Encode()
	movies, err := engine.Scrape("list")
	if err != nil {
		log.Fatal(err)
	}
	result.Movies = movies
	return result
}

// Search : Searches netnaija for a particular query and return an array of movies
func (engine *FzEngine) Search(query string) SearchResult {
	result := SearchResult{
		Query: query,
	}
	q := engine.SearchURL.Query()
	q.Set("searchname", query)
	engine.SearchURL.RawQuery = q.Encode()
	movies, err := engine.Scrape("search")
	if err != nil {
		log.Fatal(err)
	}
	result.Movies = movies
	return result
}
