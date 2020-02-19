package tests

import (
	"fmt"
	"github.com/bisoncorps/gophie/engine"
	"strings"
	"testing"
)

func TestNetNaija(t *testing.T) {
	scrapehandler := engine.GetEngine("NetNaija")
	result := scrapehandler.Search("avenge")

	if len(result.Movies) < 1 {
		t.Errorf("No movies returned")
	} else {
		for _, movie := range result.Movies {
			if movie.IsSeries == false {
				if !(strings.HasSuffix(movie.DownloadLink.String(), "1") || strings.HasSuffix(movie.DownloadLink.String(), ".mp4")) {
					fmt.Println(movie.DownloadLink.String())
					t.Errorf("Could not obtain link for single movie")
				}
			}
		}
	}

}
