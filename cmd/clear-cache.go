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
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// clearCacheCmd represents the clearCache command
var clearCacheCmd = &cobra.Command{
	Use:   "clear-cache",
	Short: "Clears the Gophie Cache",
	Run: func(cmd *cobra.Command, args []string) {
		err := os.RemoveAll(viper.GetString("cache-dir"))
		if err != nil {
			log.Fatal(err)
		}
		log.Info("Cache Cleared Successfully")
	},
}

func init() {
	rootCmd.AddCommand(clearCacheCmd)
}
