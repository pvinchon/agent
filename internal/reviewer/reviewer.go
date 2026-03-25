package reviewer

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/pvinchon/agent/internal/assistant"
)

//go:embed data/prompt_template.md
var promptTemplate string

// Reviewer focuses on a specific aspect of code quality, defined by its prompt.
type Reviewer struct {
	Path   string
	Prompt string
}

// resolve parses a comma-separated list of prompt sources and returns the
// corresponding Reviewers. Each source can be a relative path, absolute path,
// or remote URL (https://). Returns an error if any prompt cannot be loaded.
func resolve(sources string) ([]Reviewer, error) {
	if sources == "" {
		return nil, fmt.Errorf("no reviewer sources provided")
	}
	var result []Reviewer
	for s := range strings.SplitSeq(sources, ",") {
		r, err := New(strings.TrimSpace(s))
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

// New loads a Reviewer from the given source. The source can be a local file
// (relative or absolute) or a remote URL (https://).
func New(source string) (Reviewer, error) {
	prompt, err := readPrompt(source)
	if err != nil {
		return Reviewer{}, fmt.Errorf("prompt %q: %w", source, err)
	}
	return Reviewer{Path: source, Prompt: prompt}, nil
}

// buildPrompt assembles the full prompt for this reviewer against the provided diff.
func (r Reviewer) buildPrompt(diff string) string {
	return strings.NewReplacer("{{prompt}}", r.Prompt, "{{diff}}", diff).Replace(promptTemplate)
}

// review runs this reviewer against diff and returns the parsed issues.
func (r Reviewer) review(diff string, a assistant.Assistant) ([]Issue, error) {
	prompt := r.buildPrompt(diff)

	slog.Debug("reviewer", "path", r.Path, "prompt", prompt)

	response, err := assistant.Prompt(a, prompt)

	if err != nil {
		return nil, fmt.Errorf("reviewer %q: %w", r.Path, err)
	}

	var issues []Issue
	if err := json.Unmarshal([]byte(response), &issues); err != nil {
		slog.Debug("reviewer", "path", r.Path, "prompt", prompt, "response", response)
		return nil, fmt.Errorf("reviewer %q: unmarshal issues: %w", r.Path, err)
	}
	for i := range issues {
		issues[i].Reviewer = r.Path
	}
	return issues, nil
}
