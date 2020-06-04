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
	"github.com/go-phie/gophie/downloader"
	"github.com/go-phie/gophie/engine"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"reflect"
)

var pageNum int
var compResult = engine.SearchResult{
	Query:  "",
	Movies: []engine.Movie{},
}

func listPager(pageNum int) {
	selectedEngine, err := engine.GetEngine(viper.GetString("engine"))
	if err != nil {
		log.Fatal(err)
	}
	selectedMovie := processList(pageNum, selectedEngine, compResult)
	log.Debugf("Movie: %v\n", selectedMovie)
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
			Query:  selectedMovie.Title + "EPISODES",
			Movies: movieArray,
		}
		selectedMovie = processList(pageNum, selectedEngine, searchResult)
		if err = downloader.DownloadMovie(&selectedMovie, viper.GetString("output-dir")); err != nil {
			log.Fatal(err)
		}
	}
}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "lists the recent movies by page number",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		listPager(pageNum)
	},
}

func init() {
	listCmd.Flags().IntVarP(&pageNum, "page", "p", 1, "Page Number to search and return from")
	rootCmd.AddCommand(listCmd)
}

// Just abstract away the listing process so that it can be reused in other commands
func processList(pageNum int, e engine.Engine, retrievedResult engine.SearchResult) engine.Movie {
	// Initialize process and show loader on terminal and store result in result
	var result engine.SearchResult
	var choice string
	var choiceIndex int
	var items []string
	if reflect.DeepEqual(retrievedResult, compResult) {
		result = ProcessFetchTask(func() engine.SearchResult { return e.List(pageNum) })
		items = append(result.Titles(), []string{">>> Next Page"}...)
		if pageNum != 1 {
			items = append([]string{"<<< Previous Page"}, items...)
		}
		choiceIndex, choice = SelectOpts(result.Query, items)

		if choiceIndex != len(items)-1 {
			if choiceIndex == 0 && pageNum != 1 {
				listPager(pageNum - 1)
			}
		} else {
			listPager(pageNum + 1)
		}
	} else {
		result = retrievedResult
		items = append([]string{"<<< MAIN PAGE"}, result.Titles()...)
		choiceIndex, choice = SelectOpts(result.Query, items)
		if choiceIndex == 0 {
			listPager(pageNum)
		}
	}

	selectedMovie, err := result.GetMovieByTitle(choice)
	if err != nil {
		log.Fatal(err)
	}
	return selectedMovie
}
