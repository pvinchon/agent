package reviewer

import (
	"flag"
	"fmt"
)

// Flag registers a --reviewers flag (comma-separated names) and returns a
// function that resolves the chosen reviewers after flag.Parse() has been called.
func Flag() func() ([]Reviewer, error) {
	names := flag.String("reviewers", "", "Comma-separated list of reviewers to use: "+reviewerNames)
	return func() ([]Reviewer, error) {
		if *names == "" {
			return nil, fmt.Errorf("--reviewers is required")
		}
		return resolve(*names)
	}
}
