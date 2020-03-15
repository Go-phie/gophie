package engine

import (
	"fmt"
	"strings"
	"testing"
)

func TestNetNaija(t *testing.T) {
	counter := map[string]int{}
	scrapehandler, _ := GetEngine("NetNaija")
	result := scrapehandler.Search("avenge")

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
				if !(strings.HasSuffix(movie.DownloadLink.String(), "1") || strings.HasSuffix(movie.DownloadLink.String(), ".mp4")) {
					fmt.Println(movie.DownloadLink.String())
					t.Errorf("Could not obtain link for single movie")
				}
			}
		}
	}

}
