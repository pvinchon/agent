package reviewer

import (
	"embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"

	"github.com/pvinchon/agent/internal/assistant"
)

//go:embed data/prompt_template.md
var promptTemplate string

//go:embed data/prompts
var prompts embed.FS

// Reviewer focuses on a specific aspect of code quality, defined by its prompt.
type Reviewer struct {
	Slug        string
	Name        string
	Description string
	Prompt      string
}

var reviewersBySlug = func() map[string]Reviewer {
	entries, err := prompts.ReadDir("data/prompts")
	if err != nil {
		panic(err)
	}
	m := make(map[string]Reviewer, len(entries))
	for _, e := range entries {
		data, err := prompts.ReadFile("data/prompts/" + e.Name())
		if err != nil {
			panic(err)
		}
		fm, body, err := parseFrontmatter(string(data))
		if err != nil {
			panic(fmt.Sprintf("prompt %s: %v", e.Name(), err))
		}
		m[fm.Slug] = Reviewer{
			Slug:        fm.Slug,
			Name:        fm.Name,
			Description: fm.Description,
			Prompt:      body,
		}
	}
	return m
}()

var reviewerSlugs = strings.Join(slices.Sorted(maps.Keys(reviewersBySlug)), ", ")

// All returns all registered reviewers sorted by slug.
func All() []Reviewer {
	return slices.SortedFunc(maps.Values(reviewersBySlug), func(a, b Reviewer) int {
		return strings.Compare(a.Slug, b.Slug)
	})
}

// resolve parses a comma-separated list of reviewer slugs and returns the
// corresponding Reviewers. Returns an error if any slug is unknown.
func resolve(slugs string) ([]Reviewer, error) {
	if slugs == "" {
		return nil, nil
	}
	var result []Reviewer
	for slug := range strings.SplitSeq(slugs, ",") {
		r, err := New(strings.TrimSpace(slug))
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

// New returns the Reviewer registered under slug.
func New(slug string) (Reviewer, error) {
	r, ok := reviewersBySlug[slug]
	if !ok {
		return Reviewer{}, fmt.Errorf("unknown reviewer %q: choose from %s", slug, reviewerSlugs)
	}
	return r, nil
}

// buildPrompt assembles the full prompt for this reviewer against the provided diff.
func (r Reviewer) buildPrompt(diff string) string {
	return strings.NewReplacer("{{prompt}}", r.Prompt, "{{diff}}", diff).Replace(promptTemplate)
}

// review runs this reviewer against diff and returns the parsed issues.
func (r Reviewer) review(diff string, a assistant.Assistant) ([]Issue, error) {
	prompt := r.buildPrompt(diff)

	slog.Debug("reviewer", "slug", r.Slug, "prompt", prompt)

	response, err := assistant.Prompt(a, prompt)

	if err != nil {
		return nil, fmt.Errorf("reviewer %q: %w", r.Slug, err)
	}

	var issues []Issue
	if err := json.Unmarshal([]byte(response), &issues); err != nil {
		slog.Debug("reviewer", "slug", r.Slug, "prompt", prompt, "response", response)
		return nil, fmt.Errorf("reviewer %q: unmarshal issues: %w", r.Slug, err)
	}
	for i := range issues {
		issues[i].Reviewer = r.Slug
	}
	return issues, nil
}
