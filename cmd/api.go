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
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/go-phie/gophie/engine"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	port string
	// WhiteListedHosts a list of hosts that can access api
	WhiteListedHosts []string
)

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func authIP(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)

		originURL, _ := url.Parse(r.Header.Get("Origin"))
		if err == nil && (contains(WhiteListedHosts, ip) || contains(WhiteListedHosts, "*") || contains(WhiteListedHosts, originURL.Host)) {
			handler.ServeHTTP(w, r)
		} else {
		log.Debug("Rejecting Request: Host not whitelisted")
		accessDeniedHandler(w, r)
           }
	}
}

func getDefaultsMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		w.Header().Add("Content-Type", "application/json")

		// Set Default Engine to NetNaija
		engine := r.URL.Query().Get("engine")
		if engine == "" {
			q := r.URL.Query()
			q.Add("engine", "netnaija")
			r.URL.RawQuery = q.Encode()
		}
		handler.ServeHTTP(w, r)
	}
}

func accessDeniedHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Unauthorized Access", http.StatusUnauthorized)
	return
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
	}
}

// ListHandler : handles List Requests
func ListHandler(w http.ResponseWriter, r *http.Request) {
	eng := r.URL.Query().Get("engine")
	site, err := engine.GetEngine(eng)
	if site == nil {
		http.Error(w, "Invalid Engine Param", http.StatusBadRequest)
	}
	pageNum, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		http.Error(w, "Page must be a number", http.StatusBadRequest)
	}

	log.Debug("listing page ", pageNum)
	result := site.List(pageNum)
	b, err := json.Marshal(result.Movies)
	if err != nil {
		log.Fatal("failed to serialize response: ", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(b)
	log.Debug("Completed query for ", pageNum)
}

// SearchHandler : handles search requests
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	var (
		result engine.SearchResult
		site   engine.Engine
	)
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Query param must be added to url", http.StatusBadRequest)
	}

	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}

	pageNum, eRR := strconv.Atoi(page)
	if eRR != nil {
		http.Error(w, "Page must be a number", http.StatusBadRequest)
	}

	eng := r.URL.Query().Get("engine")
	site, err := engine.GetEngine(eng)
	if err != nil {
		http.Error(w, "Invalid Engine Param", http.StatusBadRequest)
	}
	log.Debug("Using Engine ", site)
	log.Debug("Searching for ", query)
	log.Debug("Returning search for page")
	result = site.Search(query, strconv.Itoa(pageNum))

	// dump results
	b, err := json.Marshal(result.Movies)
	if err != nil {
		log.Error("failed to serialize response: ", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
		}
		response, err = json.Marshal(site)
		if err != nil {
			log.Error("failed to serialize response: ", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	} else {
		response, err = json.Marshal(engine.GetEngines())
		if err != nil {
			log.Error("failed to serialize response: ", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
		http.HandleFunc("/search", authIP(getDefaultsMiddleware(SearchHandler)))
		http.HandleFunc("/list", authIP(getDefaultsMiddleware(ListHandler)))
		http.HandleFunc("/engine", EngineHandler)
		http.HandleFunc("/", DocHandler)

		log.Info("listening on ", port)
		_, err := strconv.Atoi(port)
		if err != nil {
			log.Fatal(err)
		}
		log.Fatal(http.ListenAndServe(":"+port, nil))
	},
}

func init() {
	if os.Getenv("WHITE_LISTED_HOSTS") != "" {
		WhiteListedHosts = strings.Split(os.Getenv("WHITE_LISTED_HOSTS"), ",")
	}
	apiCmd.Flags().StringVarP(&port, "port", "p", "3000", "Port to run application server on")
	rootCmd.AddCommand(apiCmd)
}
