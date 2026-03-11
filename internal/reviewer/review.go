package reviewer

import (
	"slices"

	"github.com/pvinchon/agent/internal/assistant"
	"github.com/pvinchon/agent/internal/syncx"
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
// given Assistant, and returns the aggregated issues.
func Review(reviewers []Reviewer, diff string, a assistant.Assistant) ([]Issue, []error) {
	groups, errs := syncx.Parallel(reviewers, func(r Reviewer) ([]Issue, error) {
		return r.review(diff, a)
	})
	return slices.Concat(groups...), errs
}
