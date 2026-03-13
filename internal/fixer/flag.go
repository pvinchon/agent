package fixer

import (
	"flag"
	"fmt"
	"os"
)

// TemplateFlagSet registers a --fix-template flag on fs and returns a function
// that returns the template content after fs.Parse() has been called. When no
// path is provided the returned string is empty, which causes the built-in
// template to be used.
func TemplateFlagSet(fs *flag.FlagSet) func() string {
	path := fs.String("fix-template", "", "Path to a custom fix prompt template file (must contain {{issues}} and {{diff}} placeholders)")
	return func() string {
		if *path == "" {
			return ""
		}
		data, err := os.ReadFile(*path)
		if err != nil {
			fs.Usage()
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(2)
		}
		return string(data)
	}
}
