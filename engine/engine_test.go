package engine

import (
	"fmt"
	"strings"
	"testing"
)

func testResults(t *testing.T, engine Engine) {
	counter := map[string]int{}
	result := engine.Search("jumanji")

	if len(result.Movies) < 1 {
		t.Errorf("No movies returned")
	} else {
		for _, movie := range result.Movies {
			if _, ok := counter[movie.DownloadLink.String()]; ok {
				t.Errorf("Duplicated Link")
			} else {
				counter[movie.DownloadLink.String()] = 1
			}
			if movie.IsSeries == false {
				downloadlink := movie.DownloadLink.String()
				if !(strings.HasSuffix(downloadlink, "1") || strings.HasSuffix(downloadlink, ".mp4") || strings.Contains(downloadlink, ".mkv") || strings.Contains(downloadlink, "freeload") || strings.Contains(downloadlink, "download_token=")) {
					fmt.Println(downloadlink)
					t.Errorf("Could not obtain link for single movie")
				}
			}
		}
	}
}

func TestEngines(t *testing.T) {
	engines := GetEngines()
	for _, engine := range engines {
		testResults(t, engine)
	}
}
