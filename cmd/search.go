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
	"log"
	"os"
	"strings"
	"time"

	//  "github.com/bisoncorps/gophie/downloader"
	"github.com/bisoncorps/gophie/engine"
	"github.com/briandowns/spinner"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
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
		selectedEngine := engine.GetEngine(Engine)
		query := strings.Join(args, " ")
		var result engine.SearchResult
		// Start Spinner
		if !Verbose {
			s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
			s.Suffix = " Fetching Data..."
			s.Writer = os.Stderr
			s.Start()
			result = selectedEngine.Search(query)
			s.Stop()
		} else {
			result = selectedEngine.Search(query)
		}
		prompt := promptui.Select{
			Label: result.Query,
			Items: result.Titles(),
		}
		_, choice, err := prompt.Run()
		if err != nil {
			log.Fatalf("Prompt failed %v\n", err)
		}

		selectedMovie, err := result.GetMovieByTitle(choice)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%v\n", selectedMovie)
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
