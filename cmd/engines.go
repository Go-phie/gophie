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
	"encoding/json"
	"fmt"

	"github.com/bisoncorps/gophie/engine"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var showEngine string

// engineCmd represents the engine command
var engineCmd = &cobra.Command{
	Use:   "engines",
	Short: "Show summary and list of available engines",
	Long: `Lots of engines have been used to implement gophie. 

		gophie engine list (All available engines)
		gophie engine show (Details about a particular engine)
	`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(`Engines Summaries and List

	gophie engine list - All available engines
	gophie engine show - Details about a particular engine`)
	},
}

// listEngineCmd represents the engine command
var listEngineCmd = &cobra.Command{
	Use:   "list",
	Short: "lists all available engines",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available Engines")
		for k, v := range engine.GetEngines() {
			fmt.Printf("\t%s: %s\n", k, v)
		}
	},
}

// showEngineCmd represents the engine command
var showEngineCmd = &cobra.Command{
	Use:   "show",
	Short: "Show summary of engine ",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		e, err := engine.GetEngine(args[0])
		if err != nil {
			log.Fatal(err)
		}
		b, err := json.MarshalIndent(e, "", "  ")
		if err != nil {
			fmt.Println("error:", err)
		}
		fmt.Print(string(b))
	},
}

func init() {
	engineCmd.AddCommand(showEngineCmd)
	engineCmd.AddCommand(listEngineCmd)
	rootCmd.AddCommand(engineCmd)
}
