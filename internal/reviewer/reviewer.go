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
	Name   string
	Prompt string
}

var reviewersByName = func() map[string]Reviewer {
	entries, err := prompts.ReadDir("data/prompts")
	if err != nil {
		panic(err)
	}
	m := make(map[string]Reviewer, len(entries))
	for _, e := range entries {
		name := strings.TrimSuffix(e.Name(), ".md")
		data, err := prompts.ReadFile("data/prompts/" + e.Name())
		if err != nil {
			panic(err)
		}
		m[name] = Reviewer{Name: name, Prompt: string(data)}
	}
	return m
}()

var reviewerNames = strings.Join(slices.Sorted(maps.Keys(reviewersByName)), ", ")

// resolve parses a comma-separated list of reviewer names and returns the
// corresponding Reviewers. Returns an error if any name is unknown.
func resolve(names string) ([]Reviewer, error) {
	if names == "" {
		return nil, nil
	}
	var result []Reviewer
	for name := range strings.SplitSeq(names, ",") {
		r, err := New(strings.TrimSpace(name))
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

// New returns the Reviewer registered under name.
func New(name string) (Reviewer, error) {
	r, ok := reviewersByName[name]
	if !ok {
		return Reviewer{}, fmt.Errorf("unknown reviewer %q: choose from %s", name, reviewerNames)
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

	slog.Debug("reviewer", "name", r.Name, "prompt", prompt)

	response, err := assistant.Prompt(a, prompt)

	if err != nil {
		return nil, fmt.Errorf("reviewer %q: %w", r.Name, err)
	}

	var issues []Issue
	if err := json.Unmarshal([]byte(response), &issues); err != nil {
		slog.Debug("reviewer", "name", r.Name, "prompt", prompt, "response", response)
		return nil, fmt.Errorf("reviewer %q: unmarshal issues: %w", r.Name, err)
	}
	for i := range issues {
		issues[i].Reviewer = r.Name
	}
	return issues, nil
}
