package fixer

import (
	"flag"
	"fmt"
	"os"

	"github.com/pvinchon/agent/internal/prompt"
)

const defaultFixerTemplateURL = "https://raw.githubusercontent.com/pvinchon/agent/main/internal/fixer/data/prompt_template.md"

// FlagSet registers --fixer-template on fs and returns a resolver.
// The resolver calls os.Exit(2) on load failure.
func FlagSet(fs *flag.FlagSet) func() prompt.Prompt {
	templatePath := fs.String("fixer-template", defaultFixerTemplateURL, "Path or URL to the fixer prompt template")
	return func() prompt.Prompt {
		tmpl, err := prompt.Load(*templatePath)
		if err != nil {
			fs.Usage()
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(2)
		}
		return tmpl
	}
}
