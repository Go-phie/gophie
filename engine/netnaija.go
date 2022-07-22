package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	var (
		article string
		main    string
	)
	// When in search mode, results are in <article class="result">
	switch engine.mode {
	case SearchMode:
		article = "article.sr-one"
		main = "main"
	case ListMode:
		main = "div.video-files"
		article = "article.file-one"
	default:
		return "", "", fmt.Errorf("Invalid mode %v", engine.mode)
	}
	return main, article, nil
}

func (engine *NetNaijaEngine) parseSingleMovie(el *colly.HTMLElement, index int) (Movie, error) {
	// movie title identifier
	var title string
	if title = "h2"; engine.mode == SearchMode {
		title = "h3"
	}

	re := regexp.MustCompile(`\((.*)\)`)
	movie := Movie{
		Index:    index,
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

// SabiShare Token are usually in the URL in the form https://www.sabishare.com/file/<TOKEN>-<remaining-url>
func (engine *NetNaijaEngine) getDownloadToken(sabiShareURL string) string {
	cleanURL, _ := url.Parse(sabiShareURL)
	return strings.Split(strings.Split(cleanURL.Path, "-")[0], "/")[2]
}

func (engine *NetNaijaEngine) updateDownloadProps(downloadCollector *colly.Collector, movies *[]Movie) {

	sabiShareAPI := "https://api.sabishare.com/token/download/"
	sabiShareURL := ""

	downloadCollector.OnHTML(`meta[property="og:url"]`, func(e *colly.HTMLElement) {
		if sabiShareURL == "" && strings.Contains(e.Attr("href"), "sabishare") {
			sabiShareURL = e.Attr("content")
		}
	})

	downloadCollector.OnHTML("link[rel=canonical]", func(e *colly.HTMLElement) {
		if sabiShareURL == "" && strings.Contains(e.Attr("href"), "sabishare") {
			sabiShareURL = e.Attr("href")
		}
	})

	downloadCollector.OnScraped(func(r *colly.Response) {
		// Do this operation only when we are on the download page.
		if strings.HasSuffix(r.Request.URL.Path, "download") {
			movieIndex := getMovieIndexFromCtx(r.Request)
			movie := &((*movies)[movieIndex])
			// Start by setting the default downloadURL to the sabiShare URL
			downloadURL, _ := url.Parse(sabiShareURL)
			movie.DownloadLink = downloadURL

			token := engine.getDownloadToken(sabiShareURL)
			resp, tokenErr := http.Get(fmt.Sprintf("%s%s", sabiShareAPI, token))
			type DownloadResponse struct {
				Status int `json:"status"`
				Data   struct {
					URL string `json:"url"`
				} `json:"data"`
			}

			var downloadResp DownloadResponse

			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			err = json.Unmarshal(body, &downloadResp)
			if tokenErr == nil {
				if err == nil && downloadResp.Status == 200 {
					downloadURL, _ := url.Parse(downloadResp.Data.URL)
					movie.DownloadLink = downloadURL
					sabiShareURL = ""
				}
			}
		}
	})

	// Update movie size
	downloadCollector.OnHTML("div.file-size", func(e *colly.HTMLElement) {
		(*movies)[getMovieIndexFromCtx(e.Request)].Size = strings.TrimSpace(e.ChildText("span.size-number"))
	})

	// Fetch Movie details from movie detail page
	downloadCollector.OnHTML("article.post-body", func(e *colly.HTMLElement) {
		movieIndex := getMovieIndexFromCtx(e.Request)
		movie := &((*movies)[movieIndex])
		description := e.ChildText("p")
		if description != "" {
			extraRe := regexp.MustCompile(`Genre: `)
			descAndOthers := extraRe.Split(description, -1)

			if len(descAndOthers) > 0 {
				movie.Description = descAndOthers[0]
			}

			if len(descAndOthers) > 1 {
				others := strings.ReplaceAll(descAndOthers[1], "\n", "")
				log.Info(others)
				categoryRe := regexp.MustCompile(`^(.*)Release Date:`)
				releaseDateRe := regexp.MustCompile(`Release Date:(.*)Stars`)
				starsRe := regexp.MustCompile(`Stars:(.*)Source:`)
				imdbRe := regexp.MustCompile(`.*(https:\/\/www\.imdb.*)`)
				categories := categoryRe.FindStringSubmatch(others)
				releaseDate := releaseDateRe.FindStringSubmatch(others)
				stars := starsRe.FindStringSubmatch(others)
				imdb := imdbRe.FindStringSubmatch(others)
				if len(categories) > 1 {
					movie.Category = categories[1]
				}
				if len(releaseDate) > 1 {
					movie.UploadDate = releaseDate[1]
				}
				if len(stars) > 1 {
					movie.Cast = stars[1]
				}
				if len(imdb) > 1 {
					movie.ImdbLink = imdb[1]
				}
			}
		}

		// Proceed to download movie
		// If download suffix already exists, skip. It means this link has already been visited
		if !strings.HasSuffix(movie.DownloadLink.Path, "download") {
			// download link is current link path + /download
			movie.DownloadLink.Path = path.Join(movie.DownloadLink.Path, "download")
			ctx := colly.NewContext()
			ctx.Put("movieIndex", strconv.Itoa(movieIndex))
			downloadCollector.Request("GET", movie.DownloadLink.String(), nil, ctx, nil)
		}
	})

	//for series or parts
	downloadCollector.OnHTML("div.video-series-latest-episodes", func(inn *colly.HTMLElement) {
		movie := &((*movies)[getMovieIndexFromCtx(inn.Request)])
		movie.IsSeries = true
		video_map := map[string]*url.URL{}
		inn.ForEach("a", func(num int, e *colly.HTMLElement) {
			downloadLink, err := url.Parse(e.Attr("href"))
			if err != nil {
				log.Fatal(err)
			}
			downloadLink.Path = path.Join(downloadLink.Path, "download")
			video_map[strconv.Itoa(num)] = downloadLink
		})
		movie.SDownloadLink = video_map
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
func (engine *NetNaijaEngine) Search(param ...string) SearchResult {
	query := param[0]
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
