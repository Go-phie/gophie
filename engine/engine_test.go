package engine

import (
	"fmt"
	"strings"
	"testing"
)

func testResults(t *testing.T, engine Engine) {
	counter := map[string]int{}
	var result SearchResult
	var searchTerm string
	fmt.Println(engine.String())
	// different search terms on engines
	switch {
	case strings.HasPrefix(engine.String(), "TvSeries"):
		searchTerm = "devs"
	case strings.HasPrefix(engine.String(), "TakanimeList"),
		strings.HasPrefix(engine.String(), "AnimeOut"):
		searchTerm = "attack on titans"
	case strings.HasPrefix(engine.String(), "KDramaHood"):
		searchTerm = "flower of evil"
	default:
		searchTerm = "jumanji"
	}
	result = engine.Search(searchTerm)

	if len(result.Movies) < 1 {
		t.Errorf("No movies returned from %v", engine.String())
	} else {
		for _, movie := range result.Movies {
			if _, ok := counter[movie.DownloadLink.String()]; ok {
				if !movie.IsSeries {
					t.Errorf("Duplicated Link %s", movie.DownloadLink.String())
				}
			} else {
				counter[movie.DownloadLink.String()] = 1
			}
			if movie.IsSeries == false {
				downloadlink := strings.ToLower(movie.DownloadLink.String())
				if !(strings.HasSuffix(downloadlink, "1") || strings.HasSuffix(downloadlink, ".mp4") || strings.Contains(downloadlink, ".mkv") || strings.Contains(downloadlink, ".avi") || strings.Contains(downloadlink, ".webm") || strings.Contains(downloadlink, "freeload") || strings.Contains(downloadlink, "download_token=") || strings.Contains(downloadlink, "mycoolmoviez") || strings.Contains(downloadlink, "server") || strings.Contains(downloadlink, "kdramahood")) {
					t.Errorf("Could not obtain link for single movie, linked returned is %v", downloadlink)
				}
			}
		}
	}
}

func TestEngines(t *testing.T) {
	engines := GetEngines()
	for _, engine := range engines {
		if engine.String() == "coolmoviez" || engine.String() == "mycoolmoviez" {
			testResults(t, engine)
		}
	}
}
