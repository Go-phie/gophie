/*
Copyright © 2020 Bisoncorps

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"strconv"
	"strings"

	"github.com/go-phie/gophie/downloader"
	"github.com/go-phie/gophie/engine"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "search for a movie",
	Long: `Search
			gophie search The Longests Nights
	
	Search returns a list of movies which can be selected using arrowkeys on the keyboard
	`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Engine is set from root.go
		page := strconv.Itoa(pageNum)
		query := strings.Join(args, " ")
		selectedEngine, err := engine.GetEngine(viper.GetString("engine"))
		// only run pagination for
		if strings.ToLower(viper.GetString("engine")) == "tvseries" {
			searchPager(query, page)
		} else {
			selectedMovie := processSearch(selectedEngine, query)
			log.Debugf("Movie: %v\n", selectedMovie)
			// Start Movie Download
			if err = downloader.DownloadMovie(&selectedMovie, viper.GetString("output-dir")); err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	searchCmd.Flags().IntVarP(&pageNum, "page", "p", 1, "Page Number to search and return from")
	rootCmd.AddCommand(searchCmd)
}

func searchPager(param ...string) {
	selectedEngine, err := engine.GetEngine(viper.GetString("engine"))
	if err != nil {
		log.Fatal(err)
	}
	selectedMovie := processSearch(selectedEngine, param...)
	log.Debugf("Movie: %v\n", selectedMovie)
	// Start Movie Download
	if err = downloader.DownloadMovie(&selectedMovie, viper.GetString("output-dir")); err != nil {
		log.Fatal(err)
	}
}

func processSearch(e engine.Engine, param ...string) engine.Movie {
	// Initialize process and show loader on terminal and store result in result
	var choice string
	var choiceIndex int
	var result engine.SearchResult
	query := param[0]
	if len(param) > 1 {
		pageNum, _ = strconv.Atoi(param[1])
		result = ProcessFetchTask(func() engine.SearchResult { return e.Search(param...) })
		var items []string
		items = append(result.Titles(), []string{">>> Next Page"}...)
		if pageNum != 1 {
			items = append([]string{"<<< Previous Page"}, items...)
		}
		choiceIndex, choice = SelectOpts(result.Query, items)

		if choiceIndex != len(items)-1 {
			if choiceIndex == 0 && pageNum != 1 {
				searchPager(query, strconv.Itoa(pageNum-1))
			}
		} else {
			searchPager(query, strconv.Itoa(pageNum+1))
		}
	} else {
		result = ProcessFetchTask(func() engine.SearchResult { return e.Search(query) })
		_, choice = SelectOpts(result.Query, result.Titles())
	}
	selectedMovie, err := result.GetMovieByTitle(choice)
	if err != nil {
		log.Fatal(err)
	}
	return selectedMovie
}
