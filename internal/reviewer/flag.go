package reviewer

import (
	"flag"
	"fmt"
	"os"
)

// FlagSet registers a --reviewers flag on fs and returns a function that
// resolves the chosen reviewers after fs.Parse() has been called.
func FlagSet(fs *flag.FlagSet) func() []Reviewer {
	slugs := fs.String("reviewers", "", "Comma-separated list of reviewers to use")
	return func() []Reviewer {
		if *slugs == "" {
			fs.Usage()
			fmt.Fprintln(os.Stderr, "error: --reviewers is required")
			os.Exit(2)
		}
		r, err := resolve(*slugs)
		if err != nil {
			fs.Usage()
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(2)
		}
		return r
	}
}
