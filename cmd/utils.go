package cmd

import (
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/go-phie/gophie/engine"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// fetchFunc : A function that performs initiates the fetching process of the
// scrapers. It could be the `Search` or `List` function of the engine
type fetchFunc func() engine.SearchResult

// ProcessFetchTask : Process a task in the Terminal and show processing
func ProcessFetchTask(fn fetchFunc) engine.SearchResult {
	var result engine.SearchResult
	if !viper.GetBool("verbose") {
		s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
		s.Suffix = " Fetching Data..."
		s.Writer = os.Stderr
		s.Start()
		result = fn()
		s.Stop()
	} else {
		result = fn()
	}
	if len(result.Movies) <= 0 {
		log.Info("No Results Found")
		os.Exit(0)
	}
	return result
}

// SelectOpts : use promptui to select amongst options
func SelectOpts(title string, options []string) (int, string) {
	prompt := promptui.Select{
		Label: "",
		Items: options,
		Size:  10,
	}

	index, result, err := prompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed %v\n", err)
	}
	return index, result
}
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
