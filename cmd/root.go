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
	"fmt"
	"os"
	"path"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const defaultEngine = "NetNaija"

var (
	// Engine : The Engine to use for downloads
	engineFlag string
	// Verbose : Should display verbose logs
	verbose bool
	// OutputPath :
	outputPath string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gophie",
	Short: "A CLI for downloading movies from different sources",
	Long:  `Gophie`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Log only errors except in Verbose mode
		if verbose {
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(log.InfoLevel)
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		fmt.Println("\n\nGophie - Bisoncorp (2020) (https://github.com/bisoncorps/gophie)")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Get Engine to use for search
	// Default is set to NetNaija at the Moment
	rootCmd.PersistentFlags().StringVarP(
		&engineFlag, "engine", "e", defaultEngine, "The Engine to use for querying and downloading")

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Display Verbose logs")
	rootCmd.PersistentFlags().StringVarP(
		&outputPath, "output-dir", "o", "", "Path to download files to")
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("engine", rootCmd.PersistentFlags().Lookup("engine"))
	viper.BindPFlag("output-dir", rootCmd.PersistentFlags().Lookup("output-dir"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Defaults for Gophie Configs
	viper.SetDefault("engine", defaultEngine)
	viper.SetDefault("verbose", false)
	viper.SetDefault("output-dir", path.Join(home, "Downloads", "Gophie"))
	viper.Set("gophie_cache", path.Join(home, ".gophie_cache"))
	if err := os.MkdirAll(viper.GetString("gophie_cache"), os.ModePerm); err != nil {
		log.Fatal(err)
	}

	// Configs From Env
	viper.SetEnvPrefix("gophie") // will be uppercased automatically
	viper.AutomaticEnv()         // read in environment variables that match

}
