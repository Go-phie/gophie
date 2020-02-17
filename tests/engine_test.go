package main

import (
	"github.com/bisoncorps/gophie/engine"
	"strings"
	"testing"
)

func TestNetNaija(t *testing.T) {
	scrapehandler := new(engine.NetNaijaEngine)
	scrapehandler.Search("Guns")

	if len(scrapehandler.Movies) < 1 {
		t.Errorf("No movies returned")
	} else {
		for _, movie := range scrapehandler.Movies {
			if movie.Series == false {
				if !(strings.HasSuffix(movie.DownloadLink, "1") || strings.HasSuffix(movie.DownloadLink, ".mp4")) {
					t.Errorf("Could not obtain link for single movie")
				}
			}
		}
	}

}
