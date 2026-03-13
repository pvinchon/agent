package reviewer

import (
	"slices"

	"github.com/pvinchon/agent/internal/assistant"
	"github.com/pvinchon/agent/internal/x/sync"
	"github.com/pvinchon/agent/internal/x/ux"
)

// Issue represents a single issue found by a reviewer.
type Issue struct {
	Reviewer    string `json:"reviewer"`
	Severity    string `json:"severity"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Location    string `json:"location"`
}

// Review runs all reviewers in parallel against the provided diff using the
// given Assistant, and returns the aggregated issues. If reviewTemplate is
// non-empty it is used as the scene-setting context; the output specification
// is always appended from the built-in template.
func Review(reviewers []Reviewer, diff, reviewTemplate string, a assistant.Assistant) ([]Issue, []error) {
	defer ux.Spinner()()
	groups, errs := sync.Parallel(reviewers, func(r Reviewer) ([]Issue, error) {
		return r.review(diff, reviewTemplate, a)
	})
	return slices.Concat(groups...), errs
}
