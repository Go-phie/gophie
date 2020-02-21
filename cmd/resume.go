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
	"encoding/json"
	"github.com/bisoncorps/gophie/downloader"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

// ResumeCmd represents the resume command
var ResumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "resume downloads for previously stopped movies",
	Long: `Resume
			gophie makes use of the annie downloader to resume previous downloads
	
	It stores the list of the downloads and gives you the option to remove or resume
	download
	`,
	Args: cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		var resume []downloader.Downloader
		var titles []string
		file, err := os.Open("gophie_cache/resume")
		if err != nil {
			log.Debug(err)
			return
		}
		err = json.NewDecoder(file).Decode(&resume)
		if err != nil {
			log.Fatal(err)
			return
		}
		file.Close()
		for _, i := range resume {
			titles = append(titles, i.Name)
		}

		prompt := promptui.Select{
			Label: "Resume List",
			Items: titles,
		}
		choiceIndex, _, err := prompt.Run()
		if err != nil {
			log.Fatalf("Prompt failed: %v\n", err)
		}
		selectedDownloader := resume[choiceIndex]
		selectedDownloader.DownloadFile()
	},
}

func init() {

	//  ResumeCmd.PersistentFlags().StringVarP(
	//    &outputPath, "output-dir", "o", "", "Path to download files to")
	rootCmd.AddCommand(ResumeCmd)
}
