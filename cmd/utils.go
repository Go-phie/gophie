package cmd

import (
	"os"
	"time"

	"github.com/bisoncorps/gophie/engine"
	"github.com/briandowns/spinner"
)

// fetchFunc : A function that performs initiates the fetching process of the
// scrapers. It could be the `Search` or `List` function of the engine
type fetchFunc func() engine.SearchResult

// ProcessFetchTask : Process a task in the Terminal and show processing
func ProcessFetchTask(fn fetchFunc) engine.SearchResult {
	var result engine.SearchResult
	if !Verbose {
		s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
		s.Suffix = " Fetching Data..."
		s.Writer = os.Stderr
		s.Start()
		defer s.Stop()
		result = fn()
	} else {
		result = fn()
	}
	return result
}
