package main

import (
	"encoding/json"
	"github.com/abiosoft/ishell"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"gophie/lib/downloader"
	"gophie/pkg/scraper"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var red = color.New(color.FgRed).SprintFunc()
var yellow = color.New(color.FgYellow).SprintFunc()

func Handler(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	list := r.URL.Query().Get("list")
	if search == "" && list == "" {
		log.Println("missing search and list argument")
		http.Error(w, "search and list argument is missing in url", http.StatusForbidden)
		return
	}
	site := new(scraper.NetNaija)
	if search != "" {
		log.Println("searching for", search)
		query := strings.ReplaceAll(search, "+", " ")
		site.Search(query)
	} else if list != "" {
		log.Println("listing page ", list)
		pagenum, _ := strconv.Atoi(list)
		site.List(pagenum)
	}

	// dump results
	b, err := json.Marshal(site.Movies)
	if err != nil {
		log.Println("failed to serialize response:", err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(b)
	log.Println("Completed search for", search)
}

func searchOrList(c *ishell.Context, site *scraper.NetNaija) {
	choices := []string{}
	for _, i := range site.Movies {
		if i.Title != "" {
			choices = append(choices, strconv.Itoa(1+i.Index)+" "+i.Title)
		}
	}
	c.ProgressBar().Stop()
	if len(choices) > 0 {
		choice := c.MultiChoice(choices, yellow(site.Title))
		if site.Movies[choice].Series && strings.HasSuffix(site.Movies[choice].DownloadLink, "download") {
			c.Println("This series could not be parsed")
			c.Println(site.Movies[choice].SDownloadLink)
		} else {
			url := site.Movies[choice].DownloadLink
			downloadhandler := &downloader.FileDownloader{
				URL: url,
				Mb:  0.0,
			}
			if file := downloadhandler.Filesize(); file != 0.0 {
				c.Println("Starting Download ==> Size: ", downloadhandler.Mb, "MB")
				err := downloadhandler.DownloadFile(c)
				if err != nil {
					c.Println(red(err))
				}
			}
		}
	} else {
		c.Println(red("Could not find any match"))
	}
}

func main() {
	shell := ishell.New()

	// display welcome info.
	shell.Println("Gophie Movie Downloader Shell")

	// register "list" command
	shell.AddCmd(&ishell.Cmd{
		Name: "list",
		Help: "lists the recent movies by page number",
		Func: func(c *ishell.Context) {

			display := ishell.ProgressDisplayCharSet(spinner.CharSets[35])
			c.ProgressBar().Display(display)
			c.ProgressBar().Start()
			site := new(scraper.NetNaija)
			pagenum, err := strconv.Atoi(strings.Join(c.Args, ""))
			if err != nil {
				c.Println(red("Please enter a number"))
				return
			}
			site.List(pagenum)
			searchOrList(c, site)
		},
	})

	// register a function for "search" command.
	shell.AddCmd(&ishell.Cmd{
		Name: "search",
		Help: "search for movie",
		Func: func(c *ishell.Context) {
			display := ishell.ProgressDisplayCharSet(spinner.CharSets[35])
			c.ProgressBar().Display(display)
			c.ProgressBar().Start()
			site := new(scraper.NetNaija)
			site.Search(strings.Join(c.Args, " "))
			searchOrList(c, site)
		},
	})

	// register a function for the API command
	shell.AddCmd(&ishell.Cmd{
		Name: "api",
		Help: "host gophie as an API on a PORT env variable, fallback to set argument",
		Func: func(c *ishell.Context) {
			port := strings.Join(c.Args, "")
			portByEnv := os.Getenv("PORT")
			if len(portByEnv) > 0 {
				c.Println(yellow("Found PORT env, overriding api argument"))
				port = ":" + portByEnv
			} else {
				if port == "" {
					c.Println(red("PORT must be set or an argument passed to api subcommand"))
					return
				}
				port = ":" + port
			}

			http.HandleFunc("/", Handler)
			log.Println("listening on", port)
			log.Fatal(http.ListenAndServe(port, nil))
		},
	})

	// run shell
	if len(os.Args) > 1 && os.Args[1] == "exit" {
		shell.Process(os.Args[2:]...)
	} else {
		shell.Run()
	}
}
