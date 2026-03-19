package reviewer

import (
	"flag"
	"fmt"
	"os"

	"github.com/pvinchon/agent/internal/prompt"
)

const defaultReviewerTemplateURL = "https://raw.githubusercontent.com/pvinchon/agent/main/internal/reviewer/data/prompt_template.md"

// FlagSet registers --reviewers and --reviewer-template flags on fs and returns resolvers.
// Both resolvers call os.Exit(2) on failure.
func FlagSet(fs *flag.FlagSet) (mustReviewers func() []Reviewer, mustTemplate func() prompt.Prompt) {
	paths := fs.String("reviewers", "", "Comma-separated list of reviewer prompt paths (local files or URLs)")
	templatePath := fs.String("reviewer-template", defaultReviewerTemplateURL, "Path or URL to the reviewer prompt template")

	mustTemplate = func() prompt.Prompt {
		tmpl, err := prompt.Load(*templatePath)
		if err != nil {
			fs.Usage()
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(2)
		}
		return tmpl
	}

	mustReviewers = func() []Reviewer {
		if *paths == "" {
			fs.Usage()
			fmt.Fprintln(os.Stderr, "error: --reviewers is required")
			os.Exit(2)
		}
		tmpl := mustTemplate()
		r, err := resolve(*paths, tmpl)
		if err != nil {
			fs.Usage()
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(2)
		}
		return r
	}

	return mustReviewers, mustTemplate
}
