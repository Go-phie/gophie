package main

import (
	"encoding/json"
	"github.com/abiosoft/ishell"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"gophie/pkg/scraper"
	"log"
	"net/http"
	"os"
	"strings"
)

type pageInfo struct {
	StatusCode int
	Links      map[string]int
}

func handler(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	if search == "" {
		log.Println("missing search argument")
		return
	}
	log.Println("searching for", search)
	query := strings.ReplaceAll(search, "+", " ")
	site := new(scraper.NetNaija)
	site.Search(query)

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

func main() {
	//  site := new(scraper.NetNaija)
	//  site.Search("Good boys")
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	shell := ishell.New()

	// display welcome info.
	shell.Println("Gophie Shell")

	// register a function for "search" command.
	shell.AddCmd(&ishell.Cmd{
		Name: "search",
		Help: "search for movie",
		Func: func(c *ishell.Context) {
			display := ishell.ProgressDisplayCharSet(spinner.CharSets[39])
			c.ProgressBar().Display(display)
			c.ProgressBar().Start()
			site := new(scraper.NetNaija)
			site.Search(strings.Join(c.Args, " "))
			choices := []string{}
			for _, i := range site.Movies {
				choices = append(choices, i.Title)
			}
			c.ProgressBar().Stop()
			if len(choices) > 0 {
				choice := c.MultiChoice(choices, yellow("Which do you want to download?"))
				c.Println(site.Movies[choice].DownloadLink)
			} else {
				c.Println(red("Could not find any match"))
			}

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

			http.HandleFunc("/", handler)
			c.Println("listening on", port)
			log.Fatal(http.ListenAndServe(port, nil))
		},
	})

	// run shell
	shell.Run()
}
