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

// Nkiri : An Engine for  Nkiri
type NkiriEngine struct {
	Props
}

// NewNkiriEngine : A Movie Engine Constructor for Nkiri
func NewNkiriEngine() *NkiriEngine {
	base := "https://nkiri.com/"
	baseURL, err := url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}
	// Search URL
	searchURL, err := url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}
	searchURL.Path = ""
	// List URL
	listURL, err := url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}
	listURL.Path = "/category/"
	nkiriEngine := NkiriEngine{}
	nkiriEngine.Name = "Nkiri"
	nkiriEngine.BaseURL = baseURL
	nkiriEngine.Description = `Nkiri is an entertainment website where you can download Hollywood, Korean, Chinese and other movies, TV Series and Dramas freely and easily.`
	nkiriEngine.SearchURL = searchURL
	nkiriEngine.ListURL = listURL
	return &nkiriEngine
}

// Engine Interface Methods
func (engine *NkiriEngine) String() string {
	return fmt.Sprintf("%s (%s)", engine.Name, engine.BaseURL)
}

func (engine *NkiriEngine) getParseAttrs() (string, string, error) {
	var (
		article string
		main    string
	)
	// When in search mode, results are in <article class="result">
	switch engine.mode {
	case SearchMode:
		article = "article"
		main = "div.site-content"
	case ListMode:
		main = "div.entries"
		article = "article.blog-entry"
	default:
		return "", "", fmt.Errorf("Invalid mode %v", engine.mode)
	}
	return main, article, nil
}

func (engine *NkiriEngine) parseSingleMovie(el *colly.HTMLElement, index int) (Movie, error) {
	// movie title identifier
	re := regexp.MustCompile(`\((.*)\)`)
	movie := Movie{
		Index:    index,
		IsSeries: false,
		Source:   engine.Name,
	}
	movie.CoverPhotoLink = el.ChildAttr("img", "src")
	// Split with '|': Title at Index 0
	titleSplit := strings.Split(el.ChildText("h2"), " | ")
	if len(titleSplit) >= 1 {
		movie.Title = strings.TrimSpace(
			strings.ReplaceAll(
				titleSplit[0], ":", "_"))
		movie.Category = strings.TrimPrefix(titleSplit[1], "Download") //TrimSuffix Movies
	}
	//Fetch UploadDate for ListMode Items
	if engine.mode == ListMode {
		movie.UploadDate = strings.TrimSpace(el.ChildText("div.blog-entry-date"))
		//movie.Category = strings.TrimSpace(el.ChildText("div.blog-entry-category"))
	}
	//Fetch DownloadLink
	downloadLink, err := url.Parse(el.ChildAttr("a", "href"))
	if err != nil {
		log.Fatal(err)
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

func (engine *NkiriEngine) updateDownloadProps(downloadCollector *colly.Collector, movies *[]Movie) {
	sizeRe := regexp.MustCompile(`(\d.*)`)
	//Fetch DownloadLink for series
	downloadCollector.OnHTML("div[data-elementor-type=wp-post]", func(e *colly.HTMLElement) {
		movieIndex := getMovieIndexFromCtx(e.Request)
		movie := &((*movies)[movieIndex])
		movie.IsSeries = true
		seriesMap := map[string]*url.URL{}
		e.ForEach("section.elementor-section", func(n int, inner *colly.HTMLElement) {
			if strings.HasPrefix(inner.ChildText("a"), "Download Episode") {
				downloadLink, err := url.Parse(inner.Attr("href"))
				if err != nil {
					log.Fatal(err)
				}
				seriesMap[strconv.Itoa(n)] = downloadLink
			}
		})
		movie.SDownloadLink = seriesMap
	})
	// Fetch Movie details from movie details / downloadlink
	downloadCollector.OnHTML("div.elementor-section-wrap", func(e *colly.HTMLElement) {
		movieIndex := getMovieIndexFromCtx(e.Request)
		movie := &((*movies)[movieIndex])
		e.ForEach("section.elementor-section", func(n int, inner *colly.HTMLElement) {
			//Update movie size
			if strings.HasPrefix(inner.ChildText("span.elementor-alert-title"), "Download Size") {
				sizeMatch := sizeRe.FindStringSubmatch(
					strings.TrimSpace(
						inner.ChildText("span.elementor-alert-description")))
				if len(sizeMatch) >= 1 {
					movie.Size = sizeMatch[0]
				}
			}
			//Fetch Movie Description
			if inner.Attr("div.nkiri-content_9") != "" && inner.Attr("div.nkiri-content_4") != "" {
				movie.Description = inner.ChildText("p")
			}
			//Fetch DownloadLink For Movies
			if strings.HasPrefix(inner.ChildText("a"), "Download Movie") {
				downloadLink, err := url.Parse(inner.ChildAttr("a", "href"))
				if err != nil {
					log.Fatal(err)
				}
				movie.DownloadLink = downloadLink
			}
		})
	})
}

// List : list all the movies on a page
func (engine *NkiriEngine) List(page int) SearchResult {
	engine.mode = ListMode
	result := SearchResult{
		Query: "List of Recent Uploads - Page " + strconv.Itoa(page),
	}
	pageParam := fmt.Sprintf("page/%v", strconv.Itoa(page))
	categories := [5]string{
		"international",
		"african",
		"asian-movies/download-bollywood-movies",
		"asian-movies/download-korean-movies",
		"asian-movies/download-philippine-movies"}
	movies := []Movie{}
	listPath := engine.ListURL.Path
	for _, category := range categories {
		engine.ListURL.Path = path.Join(listPath, category, pageParam)
		listResult, err := Scrape(engine)
		if err != nil {
			log.Fatal(err)
		}
		movies = append(movies, listResult...)
	}
	result.Movies = movies
	return result
}

// Search : Searches nkiri for a particular query and return an array of movies
func (engine *NkiriEngine) Search(param ...string) SearchResult {
	query := param[0]
	engine.mode = SearchMode
	result := SearchResult{
		Query: query,
	}
	q := engine.SearchURL.Query()
	q.Set("s", query)
	q.Set("post_type", "post")
	engine.SearchURL.RawQuery = q.Encode()
	movies, err := Scrape(engine)
	if err != nil {
		log.Fatal(err)
	}
	result.Movies = movies
	return result
}
