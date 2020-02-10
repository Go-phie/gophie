package main

import (
	"encoding/json"
	"gophie/pkg/scraper"
	"log"
	"net/http"
	"strings"

	"github.com/abiosoft/ishell"
	"github.com/gocolly/colly/v2"
)

type pageInfo struct {
	StatusCode int
	Links      map[string]int
}

func handler(w http.ResponseWriter, r *http.Request) {
	URL := r.URL.Query().Get("url")
	if URL == "" {
		log.Println("missing URL argument")
		return
	}
	log.Println("visiting", URL)

	c := colly.NewCollector()

	p := &pageInfo{Links: make(map[string]int)}

	// count links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if link != "" {
			p.Links[link]++
		}
	})

	// extract status code
	c.OnResponse(func(r *colly.Response) {
		log.Println("response received", r.StatusCode)
		p.StatusCode = r.StatusCode
	})
	c.OnError(func(r *colly.Response, err error) {
		log.Println("error:", r.StatusCode, err)
		p.StatusCode = r.StatusCode
	})

	c.Visit(URL)

	// dump results
	b, err := json.Marshal(p)
	if err != nil {
		log.Println("failed to serialize response:", err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(b)
}

func main() {
	//  site := new(scraper.NetNaija)
	//  site.Search("Good boys")
	shell := ishell.New()

	// display welcome info.
	shell.Println("Gophie Shell")

	// register a function for "greet" command.
	shell.AddCmd(&ishell.Cmd{
		Name: "search",
		Help: "search for movie",
		Func: func(c *ishell.Context) {
			site := new(scraper.NetNaija)
			site.Search(strings.Join(c.Args, " "))
			choices := []string{}
			for _, i := range site.Movies {
				choices = append(choices, i.Title)
			}
			choice := c.MultiChoice(choices, "Which do you want to download?")

			c.Println(site.Movies[choice].DownloadLink)
		},
	})

	// run shell
	shell.Run()
}
