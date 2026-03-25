package reviewer

import (
	"flag"
	"fmt"
	"os"
)

// FlagSet registers a --reviewers flag on fs and returns a function that
// resolves the chosen reviewers after fs.Parse() has been called.
func FlagSet(fs *flag.FlagSet) func() []Reviewer {
	sources := fs.String("reviewers", "", "Comma-separated reviewer prompt sources (file path or https:// URL)")
	return func() []Reviewer {
		if *sources == "" {
			fs.Usage()
			fmt.Fprintln(os.Stderr, "error: --reviewers is required")
			os.Exit(2)
		}
		r, err := resolve(*sources)
		if err != nil {
			fs.Usage()
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(2)
		}
		return r
	}
}
