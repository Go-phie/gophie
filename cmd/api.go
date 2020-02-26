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
	"net/http"
	"strconv"
	"strings"

	"github.com/bisoncorps/gophie/engine"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var port string

// Handler : handles serving gophie
func Handler(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	list := r.URL.Query().Get("list")
	eng := r.URL.Query().Get("engine")
	var (
		result engine.SearchResult
		site   engine.Engine
	)
	if search == "" && list == "" {
		log.Debug("missing search and list argument")
		http.Error(w, "search and list argument is missing in url", http.StatusForbidden)
		return
	}
	if eng == "" {
		// Use NetNaija as the default engine
		site = engine.NewNetNaijaEngine()
	} else {
		site = engine.GetEngine(eng)
	}
	log.Info("Using Engine ", site)
	if search != "" {
		log.Debug("Searching for ", search)
		query := strings.ReplaceAll(search, "+", " ")
		result = site.Search(query)
	} else if list != "" {
		log.Debug("listing page ", list)
		pagenum, _ := strconv.Atoi(list)
		result = site.List(pagenum)
	}

	// dump results
	b, err := json.Marshal(result.Movies)
	if err != nil {
		log.Fatal("failed to serialize response: ", err)
	}
	enableCors(&w)
	w.Header().Add("Content-Type", "application/json")
	w.Write(b)
	log.Debug("Completed search for", search, list)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

// apiCmd represents the api command
var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "host gophie as an API on a PORT env variable, fallback to set argument",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		http.HandleFunc("/", Handler)
		log.Info("listening on ", port)
		_, err := strconv.Atoi(port)
		if err != nil {
			log.Fatal(err)
		}
		log.Fatal(http.ListenAndServe(":"+port, nil))
	},
}

func init() {
	apiCmd.Flags().StringVarP(&port, "port", "p", "3000", "Port to run application server on")
	rootCmd.AddCommand(apiCmd)
}
