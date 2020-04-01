/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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

	"github.com/bisoncorps/gophie/engine"
	"github.com/bisoncorps/mplayer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var selectedPlayer string

const DEFAULT_PLAYER = "browser"

// streamCmd represents the stream command
var streamCmd = &cobra.Command{
	Use:   "stream",
	Short: "Stream a video from gophie",
	Long: `Stream a video from Gophie
	
Gophie stream allows streaming videos using browser player. You can pass a query argument to the stream call to search for a specific query or left blank to retrieve latest uploaded movies on the specified engine. Full list of supported players can be found at https://github.com/bisoncorps/mplayer.  
example:
  gophie stream Jumanji --player vlc (stream Jumanji using VLC Media Player)
  gophie stream -e fzmovies (check for latest movies on fzmovies for streaming)
	`,
	Run: func(cmd *cobra.Command, args []string) {
		selectedEngine, err := engine.GetEngine(viper.GetString("engine"))
		if err != nil {
			log.Fatal(err)
		}
		query := strings.Join(args, " ")
		var movie engine.Movie
		if query == "" {
			movie = processList(1, selectedEngine)
		} else {
			movie = processSearch(query, selectedEngine)
		}
		p, err := mplayer.GetPlayer(selectedPlayer)
		if err != nil {
			log.Fatal(err)
		}
		p.SetURL(movie.DownloadLink.String())
		p.SetTitle(movie.Title)
		p.Play()
	},
}

func init() {
	streamCmd.Flags().StringVarP(
		&selectedPlayer, "player", "p", DEFAULT_PLAYER, "Player to use for streaming")
	rootCmd.AddCommand(streamCmd)
}
