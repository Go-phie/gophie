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
	fzEngine.Description = `FzMovies is a site where you can find Bollywood, Hollywood and DHollywood Movies.`
	fzEngine.SearchURL = searchURL
	fzEngine.ListURL = listURL
	return &fzEngine
}

// Engine Interface Methods

func (engine *FzEngine) String() string {
	st := fmt.Sprintf("%s (%s)", engine.Name, engine.BaseURL)
	return st
}

func (engine *FzEngine) getParseAttrs() (string, string, error) {
	return "body", "div.mainbox", nil
}

func (engine *FzEngine) parseSingleMovie(el *colly.HTMLElement, index int) (Movie, error) {
	movie := Movie{
		Index:    index,
		IsSeries: false,
		Source:   engine.Name,
	}
	cover, err := url.Parse(el.Request.AbsoluteURL(el.ChildAttr("img", "src")))
	if err != nil {
		log.Fatal(err)
	}
	movie.CoverPhotoLink = cover.String()
	// Remove all Video: or Movie: Prefixes
	if len(el.ChildTexts("small")) >= 2 {
		movie.UploadDate = strings.TrimSpace(el.ChildTexts("small")[1])
	}
	movie.Title = strings.TrimSuffix(strings.TrimSpace(el.ChildText("b")), "<more>")
	if len(el.ChildTexts("small")) > 3 {
		description := strings.TrimSpace(el.ChildTexts("small")[3])
		tagsRe := regexp.MustCompile(`Tags\s+: (.*)\.\.\..*`)
		descAndOthers := tagsRe.Split(description, -1)

		if len(descAndOthers) > 0 {
			movie.Description = strings.TrimSuffix(descAndOthers[0], "<more>")
			tag := tagsRe.FindStringSubmatch(description)

			if len(tag) > 1 {
				movie.Tags = strings.ReplaceAll(tag[1], "|", ",")
			}
		} else {
			movie.Description = description
		}

	}
	downloadLink, err := url.Parse(el.Request.AbsoluteURL(el.ChildAttr("a", "href")))

	if err != nil {
		log.Fatal(err)
	}
	downloadLink.Path = path.Join(engine.BaseURL.Path, downloadLink.Path)

	movie.DownloadLink = downloadLink
	return movie, nil
}

func (engine *FzEngine) updateDownloadProps(downloadCollector *colly.Collector, movies *[]Movie) {
	// Update movie download link if ul.downloadlinks on page
	downloadCollector.OnHTML("ul.ptype", func(e *colly.HTMLElement) {
		movie := &(*movies)[getMovieIndexFromCtx(e.Request)]
		link := strings.Replace(e.ChildAttr("a", "href"), "download1.php", "download.php", 1)
		downloadLink, err := url.Parse(e.Request.AbsoluteURL(link + "&pt=jRGarGzOo2"))
		if err != nil {
			log.Fatal(err)
		}
		movie.DownloadLink = downloadLink
		re := regexp.MustCompile(`(.* MB)`)
		stringsub := re.FindStringSubmatch(e.ChildText("dcounter"))
		if len(stringsub) > 0 {
			dl := strings.TrimPrefix(stringsub[0], "(")
			dlRe := regexp.MustCompile(`(\d+ MB)`)
			dlSize := dlRe.FindStringSubmatch(dl)
			if len(dlSize) > 1 {
				movie.Size = dlSize[1]
			} else {
				movie.Size = dl
			}
		}
		if strings.HasSuffix(movie.Title, "Tags") {
			movie.Title = strings.TrimSuffix(movie.Title, "Tags")
		}
		downloadCollector.Visit(downloadLink.String())
	})

	downloadCollector.OnHTML("ul.downloadlinks", func(e *colly.HTMLElement) {
		movie := &(*movies)[getMovieIndexFromCtx(e.Request)]
		links := e.ChildAttrs("a", "href")
		if len(links) > 1 {
			downloadLink, err := url.Parse(e.Request.AbsoluteURL(links[len(links)-1]))
			if err != nil {
				log.Fatal(err)
			}
			movie.DownloadLink = downloadLink
			downloadCollector.Visit(downloadLink.String())
		}
	})

	// Update Download Link if "Download" HTML on page
	downloadCollector.OnHTML("input[name=download1]", func(e *colly.HTMLElement) {
		trimmedValue := strings.TrimSpace(e.Attr("value"))
		if strings.HasSuffix(trimmedValue, "mp4") || strings.HasSuffix(trimmedValue, "mp4?fromwebsite") {
			downloadLink, err := url.Parse(e.Request.AbsoluteURL(e.Attr("value")))
			if err != nil {
				log.Fatal(err)
			}
			(*movies)[getMovieIndexFromCtx(e.Request)].DownloadLink = downloadLink
		}
	})
}

// List : list all the movies on a page
func (engine *FzEngine) List(page int) SearchResult {
	engine.mode = ListMode
	result := SearchResult{
		Query: "List of Recent Uploads - Page " + strconv.Itoa(page),
	}
	q := engine.ListURL.Query()
	q.Set("catID", "2")
	q.Set("by", "date")
	q.Set("pg", strconv.Itoa(page))
	engine.ListURL.RawQuery = q.Encode()
	movies, err := Scrape(engine)
	if err != nil {
		log.Fatal(err)
	}
	result.Movies = movies
	return result
}

// Search : Searches fzmovies for a particular query and return an array of movies
func (engine *FzEngine) Search(param ...string) SearchResult {
	query := param[0]
	engine.mode = SearchMode
	result := SearchResult{
		Query: query,
	}
	q := engine.SearchURL.Query()
	q.Set("searchname", query)
	engine.SearchURL.RawQuery = q.Encode()
	movies, err := Scrape(engine)
	if err != nil {
		log.Fatal(err)
	}
	result.Movies = movies
	return result
}
