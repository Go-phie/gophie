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
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Defaults
const (
	defaultEngine = "netnaija"
)

var (
	// Engine : The Engine to use for downloads
	engineFlag string
	// Verbose : Should display verbose logs
	verbose bool
	// OutputPath :
	outputPath string
	// Selenium URL: The URL of the selenium instance to connect to
	seleniumURL string
	// CacheDir: The Directory to store all colly files
	cacheDir string
	// Should Cache requests or not
	ignoreCache bool
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
		fmt.Println("\n\nGophie - Bisoncorp (2020) (https://github.com/go-phie/gophie)")
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

	rootCmd.PersistentFlags().StringVarP(
		&cacheDir, "cache-dir", "c", "", "The directory to store/lookup cache")
	rootCmd.PersistentFlags().StringVarP(
		&seleniumURL, "selenium-url", "s", "", "The URL of selenium instance to use")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Display Verbose logs")
	rootCmd.PersistentFlags().StringVarP(&outputPath, "output-dir", "o", "", "Path to download files to")
	rootCmd.PersistentFlags().BoolVar(&ignoreCache, "ignore-cache", false, "Ignore Cache and makes new requests")

	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("engine", rootCmd.PersistentFlags().Lookup("engine"))
	viper.BindPFlag("selenium-url", rootCmd.PersistentFlags().Lookup("selenium-url"))
	viper.BindPFlag("output-dir", rootCmd.PersistentFlags().Lookup("output-dir"))
	viper.BindPFlag("cache-dir", rootCmd.PersistentFlags().Lookup("cache-dir"))
	viper.BindPFlag("ignore-cache", rootCmd.PersistentFlags().Lookup("ignore-cache"))

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
	viper.SetDefault("selenium-url", "http://localhost:4444")
	viper.SetDefault("output-dir", path.Join(home, "Downloads", "Gophie"))
	viper.Set("cache-dir", path.Join(home, ".gophie_cache"))
	if err := os.MkdirAll(viper.GetString("cache-dir"), os.ModePerm); err != nil {
		log.Fatal(err)
	}

	// Configs From Env
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("gophie") // will be uppercased automatically
	viper.AutomaticEnv()         // read in environment variables that match
}
