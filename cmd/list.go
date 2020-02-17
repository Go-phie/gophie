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

	"github.com/bisoncorps/gophie/engine"
	"github.com/spf13/cobra"
)

var pageNum *int

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "lists the recent movies by page number",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		Engine := engine.GetEngine("NetNaija")
		result := Engine.List(*pageNum)
		fmt.Println(result)
	},
}

func init() {
	pageNum = listCmd.Flags().IntP("page", "p", 1, "Page Number to search and return from")
	rootCmd.AddCommand(listCmd)
}
