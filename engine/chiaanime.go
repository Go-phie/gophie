package engine

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	log "github.com/sirupsen/logrus"
)

// ChiaAnime : An Engine for ChiaAnime
type ChiaAnime struct {
	Props
}

// NewChiaAnimeEngine : create a new engine for scraping latest anime from chia-anime
func NewChiaAnimeEngine() *ChiaAnime {
	base := "https://m.chia-anime.me"
	baseURL, err := url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}
	// Search URL
	searchURL, err := url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}
	searchURL.Path = "/catlist.php"

	// List URL
	listURL, err := url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}
	listURL.Path = "/"

	chiaAnimeEngine := ChiaAnime{}
	chiaAnimeEngine.Name = "ChiaAnime"
	chiaAnimeEngine.BaseURL = baseURL
	chiaAnimeEngine.Description = `Founded by a group of college students that love to watch anime`
	chiaAnimeEngine.SearchURL = searchURL
	chiaAnimeEngine.ListURL = listURL
	return &chiaAnimeEngine
}

// Engine Interface Methods

func (engine *ChiaAnime) String() string {
	st := fmt.Sprintf("%s (%s)", engine.Name, engine.BaseURL)
	return st
}

func (engine *ChiaAnime) getParseAttrs() (string, string, error) {
	return "table.items", "tr.episode", nil
}

func (engine *ChiaAnime) parseSingleMovie(el *colly.HTMLElement, index int) (Movie, error) {
	movie := Movie{
		Index:    index,
		IsSeries: false,
		Source:   engine.Name,
	}
	movie.Title = strings.TrimSpace(el.ChildTexts("a")[1])
	re := regexp.MustCompile(`background: url \('.*'\);`)
	stringsub := re.FindStringSubmatch(el.ChildAttr("span", "style"))
	coverphotolink := stringsub[0]
	movie.CoverPhotoLink = coverphotolink
	downloadLink, err := url.Parse(el.Request.AbsoluteURL(el.ChildAttr("a", "href")))

	if err != nil {
		log.Fatal(err)
	}

	movie.DownloadLink = downloadLink
	return movie, nil
}

func (engine *ChiaAnime) updateDownloadProps(downloadCollector *colly.Collector, movies *[]Movie) {
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
		re := regexp.MustCompile(`[ Size : (\d+) MB ]`)
		stringsub := re.FindStringSubmatch(e.ChildText("span"))
		movie.Size = stringsub[0]
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
func (engine *ChiaAnime) List(page int) SearchResult {
	engine.mode = ListMode
	result := SearchResult{
		Query: "List of Recent Uploads - Page " + strconv.Itoa(page),
	}
	q := engine.ListURL.Query()
	q.Set("paged", strconv.Itoa(page))
	movies, err := Scrape(engine)
	if err != nil {
		log.Fatal(err)
	}
	result.Movies = movies
	return result
}

// Search : Searches fzmovies for a particular query and return an array of movies
func (engine *ChiaAnime) Search(param ...string) SearchResult {
	query := param[0]
	engine.mode = SearchMode
	result := SearchResult{
		Query: query,
	}
	q := engine.SearchURL.Query()
	q.Set("tags", query)
	engine.SearchURL.RawQuery = q.Encode()
	movies, err := Scrape(engine)
	if err != nil {
		log.Fatal(err)
	}
	result.Movies = movies
	return result
}
