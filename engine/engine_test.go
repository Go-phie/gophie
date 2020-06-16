package engine

import (
	"fmt"
	"strings"
	"testing"
	// _ "github.com/go-phie/gophie/cmd"
)

func testResults(t *testing.T, engine Engine) {
	counter := map[string]int{}
	var result SearchResult
	var searchTerm string
	fmt.Println(engine.String())
	if !strings.HasPrefix(engine.String(), "TvSeries") {
		if strings.HasPrefix(engine.String(), "Anime") {
			searchTerm = "attack on titan"
		} else {
			searchTerm = "jumanji"
		}
	} else {
		// search for the flash for movie series
		searchTerm = "devs"
	}
	result = engine.Search(searchTerm)

	if len(result.Movies) < 1 {
		t.Errorf("No movies returned from %v", engine.String())
	} else {
		for _, movie := range result.Movies {
			if _, ok := counter[movie.DownloadLink.String()]; ok {
				t.Errorf("Duplicated Link")
			} else {
				counter[movie.DownloadLink.String()] = 1
			}
			if movie.IsSeries == false {
				downloadlink := movie.DownloadLink.String()
				if !(strings.HasSuffix(downloadlink, "1") || strings.HasSuffix(downloadlink, ".mp4") || strings.Contains(downloadlink, ".mkv") || strings.Contains(downloadlink, ".avi") || strings.Contains(downloadlink, ".webm") || strings.Contains(downloadlink, "freeload") || strings.Contains(downloadlink, "download_token=") || strings.Contains(downloadlink, "mycoolmoviez") || strings.Contains(downloadlink, "server")) {
					t.Errorf("Could not obtain link for single movie, linked returned is %v", downloadlink)
				}
			}
		}
	}
}

func TestEngines(t *testing.T) {
	engines := GetEngines()
	for _, engine := range engines {
		if !(strings.HasPrefix(engine.String(), "NetNaija") || strings.HasPrefix(engine.String(), "MyCoolMoviez")) {
			testResults(t, engine)
		}
	}
}
