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
	"reflect"
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
		// only run pagination for
		if strings.ToLower(viper.GetString("engine")) == "tvseries" {
			searchPager(query, page)
		} else {
			searchPager(query)
		}
	},
}

func init() {
	searchCmd.Flags().IntVarP(&pageNum, "page", "p", 1, "Page Number to search and return from")
	rootCmd.AddCommand(searchCmd)
}

func searchPager(params ...string) {
	selectedEngine, err := engine.GetEngine(viper.GetString("engine"))
	if err != nil {
		log.Fatal(err)
	}
	selectedMovie := processSearch(selectedEngine, compResult, params...)
	// Start Movie Download
	if len(selectedMovie.SDownloadLink) < 1 {
		if err = downloader.DownloadMovie(&selectedMovie, viper.GetString("output-dir")); err != nil {
			log.Fatal(err)
		}
	} else {
		var movieArray []engine.Movie
		index := 0
		for key, val := range selectedMovie.SDownloadLink {
			movieArray = append(movieArray,
				engine.Movie{
					Index:          index,
					Title:          key,
					IsSeries:       false,
					Source:         selectedMovie.Source,
					DownloadLink:   val,
					CoverPhotoLink: selectedMovie.CoverPhotoLink,
					Description:    selectedMovie.Description,
				})
			index++
		}
		searchResult := engine.SearchResult{
			Query:  selectedMovie.Title + " EPISODES",
			Movies: movieArray,
		}
		selectedMovie = processSearch(selectedEngine, searchResult, params...)
		if err = downloader.DownloadMovie(&selectedMovie, viper.GetString("output-dir")); err != nil {
			log.Fatal(err)
		}
	}
}

func processSearch(e engine.Engine, retrievedResult engine.SearchResult, params ...string) engine.Movie {
	// Initialize process and show loader on terminal and store result in result
	var (
		choice      string
		choiceIndex int
		result      engine.SearchResult
		items       []string
	)
	query := params[0]
	if len(params) > 1 {
		pageNum, _ = strconv.Atoi(params[1])
		result = ProcessFetchTask(func() engine.SearchResult { return e.Search(params...) })
		items = append(result.Titles(), []string{">>> Next Page"}...)
		log.Debug(result)
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
		if reflect.DeepEqual(retrievedResult, compResult) {
			result = ProcessFetchTask(func() engine.SearchResult { return e.Search(query) })
			_, choice = SelectOpts(result.Query, result.Titles())
		} else {
			result = retrievedResult
			items = append([]string{"<<< MAIN PAGE"}, result.Titles()...)
			choiceIndex, choice = SelectOpts(result.Query, items)
			if choiceIndex == 0 {
				searchPager(query, strconv.Itoa(pageNum))
			}
		}
	}
	selectedMovie, err := result.GetMovieByTitle(choice)
	if err != nil {
		log.Fatal(err)
	}
	return selectedMovie
}
