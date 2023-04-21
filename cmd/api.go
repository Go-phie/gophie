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
	"html/template"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/handlers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/go-phie/gophie/engine"
)

var (
	port string
)

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func getDefaultsMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		w.Header().Add("Content-Type", "application/json")

		// Set Default Engine to fzmovies
		engine := r.URL.Query().Get("engine")
		if engine == "" {
			q := r.URL.Query()
			q.Add("engine", "fzmovies")
			r.URL.RawQuery = q.Encode()
		}
		handler.ServeHTTP(w, r)
	}
}

// DocHandler : renders iframe pointing to hosted docs
func DocHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	w.Header().Add("Content-Type", "text/html")
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, "None"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ListHandler : handles List Requests
func ListHandler(w http.ResponseWriter, r *http.Request) {
	var (
		pageNum int
		err     error
	)
	eng := r.URL.Query().Get("engine")
	site, err := engine.GetEngine(eng)
	if site == nil {
		http.Error(w, "Invalid Engine Param", http.StatusBadRequest)
		return
	}
	if r.URL.Query().Get("page") == "" {
		pageNum = 1
	} else {
		pageNum, err = strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			http.Error(w, "Page must be a number", http.StatusBadRequest)
			return
		}
	}

	result := site.List(pageNum)
	b, err := json.Marshal(result.Movies)
	if err != nil {
		log.Fatal("failed to serialize response: ", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(b)
}

// SearchHandler : handles search requests
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	var (
		result  engine.SearchResult
		site    engine.Engine
		err     error
		pageNum int
	)
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Query param must be added to url", http.StatusBadRequest)
		return
	}

	if r.URL.Query().Get("page") == "" {
		pageNum = 1
	} else {
		pageNum, err = strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			http.Error(w, "Page must be a number", http.StatusBadRequest)
			return
		}
	}

	site, err = engine.GetEngine(r.URL.Query().Get("engine"))
	if err != nil {
		http.Error(w, "Invalid Engine Param", http.StatusBadRequest)
		return
	}
	log.Infof("Processing search Request for engine=%s and query=%s", site, query)
	result = site.Search(query, strconv.Itoa(pageNum))

	// dump results
	b, err := json.Marshal(result.Movies)
	if err != nil {
		log.Error("failed to serialize response: ", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Write(b)
	log.Debug("Completed search for ", query)
}

// EngineHandler : handles Engine Listing
func EngineHandler(w http.ResponseWriter, r *http.Request) {
	eng := r.URL.Query().Get("engine")
	enableCors(&w)
	w.Header().Add("Content-Type", "application/json")
	var (
		response []byte
		err      error
	)
	if eng != "" {
		site, err := engine.GetEngine(eng)
		if err != nil {
			http.Error(w, "Invalid Engine Param", http.StatusBadRequest)
			return
		}
		response, err = json.Marshal(site)
		if err != nil {
			log.Error("failed to serialize response: ", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		response, err = json.Marshal(engine.GetEngines())
		if err != nil {
			log.Error("failed to serialize response: ", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	w.Write(response)
}

// apiCmd represents the api command
var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "host gophie as an API on a PORT env variable, fallback to set argument",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		r := http.NewServeMux()
		r.HandleFunc("/search", getDefaultsMiddleware(SearchHandler))
		r.HandleFunc("/list", getDefaultsMiddleware(ListHandler))
		r.HandleFunc("/engine", EngineHandler)
		r.HandleFunc("/", DocHandler)

		log.Info("listening on ", port)
		_, err := strconv.Atoi(port)
		if err != nil {
			log.Fatal(err)
		}
		loggedRouter := handlers.LoggingHandler(os.Stdout, r)
		log.Fatal(http.ListenAndServe(":"+port, loggedRouter))
	},
}

func init() {
	apiCmd.Flags().StringVarP(&port, "port", "p", "3000", "Port to run application server on")
	rootCmd.AddCommand(apiCmd)
}
