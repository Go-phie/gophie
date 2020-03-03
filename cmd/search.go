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
	"strings"

	"github.com/bisoncorps/gophie/downloader"
	"github.com/bisoncorps/gophie/engine"
	"github.com/manifoldco/promptui"
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
		selectedEngine := engine.GetEngine(viper.GetString("engine"))
		query := strings.Join(args, " ")
		var result engine.SearchResult
		// Initialize process and show loader on terminal and store result in result
		result = ProcessFetchTask(func() engine.SearchResult { return selectedEngine.Search(query) })
		prompt := promptui.Select{
			Label: result.Query,
			Items: result.Titles(),
			Size:  10,
		}
		_, choice, err := prompt.Run()
		if err != nil {
			log.Fatalf("Prompt failed %v\n", err)
		}

		selectedMovie, err := result.GetMovieByTitle(choice)
		if err != nil {
			log.Fatal(err)
		}
		log.Debugf("Movie: %v\n", selectedMovie)
		// Start Movie Download
		if err = downloader.DownloadMovie(&selectedMovie, viper.GetString("output-dir")); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
