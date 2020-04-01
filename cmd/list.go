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
	"github.com/bisoncorps/gophie/downloader"
	"github.com/bisoncorps/gophie/engine"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var pageNum int

func listPager(pageNum int) {
	selectedEngine, err := engine.GetEngine(viper.GetString("engine"))
	if err != nil {
		log.Fatal(err)
	}
	selectedMovie := processList(pageNum, selectedEngine)
	log.Debugf("Movie: %v\n", selectedMovie)
	// Start Movie Download
	if err = downloader.DownloadMovie(&selectedMovie, viper.GetString("output-dir")); err != nil {
		log.Fatal(err)
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
func processList(pageNum int, e engine.Engine) engine.Movie {
	// Initialize process and show loader on terminal and store result in result
	result := ProcessFetchTask(func() engine.SearchResult { return e.List(pageNum) })
	var items []string
	items = append(result.Titles(), []string{">>> Next Page"}...)
	if pageNum != 1 {
		items = append([]string{"<<< Previous Page"}, items...)
	}
	choiceIndex, choice := SelectOpts(result.Query, items)

	if choiceIndex != len(items)-1 {
		if choiceIndex == 0 && pageNum != 1 {
			listPager(pageNum - 1)
		}
	} else {
		listPager(pageNum + 1)
	}

	selectedMovie, err := result.GetMovieByTitle(choice)
	if err != nil {
		log.Fatal(err)
	}
	return selectedMovie
}
