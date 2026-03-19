package reviewer

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"path"
	"strings"

	"github.com/pvinchon/agent/internal/assistant"
	"github.com/pvinchon/agent/internal/prompt"
)

// Reviewer focuses on a specific aspect of code quality, defined by its prompt.
type Reviewer struct {
	Name     string
	Prompt   prompt.Prompt
	Template prompt.Prompt
}

// resolve parses a comma-separated list of prompt paths and returns the corresponding
// Reviewers, each using tmpl as their template.
func resolve(paths string, tmpl prompt.Prompt) ([]Reviewer, error) {
	if paths == "" {
		return nil, fmt.Errorf("--reviewers is required")
	}
	var result []Reviewer
	for token := range strings.SplitSeq(paths, ",") {
		token = strings.TrimSpace(token)
		if token == "" {
			return nil, fmt.Errorf("empty reviewer path in list")
		}
		p, err := prompt.Load(token)
		if err != nil {
			return nil, err
		}
		base := path.Base(token)
		name := strings.TrimSuffix(base, path.Ext(base))
		result = append(result, Reviewer{Name: name, Prompt: p, Template: tmpl})
	}
	return result, nil
}

// buildPrompt assembles the full prompt for this reviewer against the provided diff.
func (r Reviewer) buildPrompt(diff string) string {
	return strings.NewReplacer("{{prompt}}", r.Prompt.String(), "{{diff}}", diff).Replace(r.Template.String())
}

// review runs this reviewer against diff and returns the parsed issues.
func (r Reviewer) review(diff string, a assistant.Assistant) ([]Issue, error) {
	p := r.buildPrompt(diff)

	slog.Debug("reviewer", "name", r.Name, "prompt", p)

	response, err := assistant.Prompt(a, p)
	if err != nil {
		return nil, fmt.Errorf("reviewer %q: %w", r.Name, err)
	}

	var issues []Issue
	if err := json.Unmarshal([]byte(response), &issues); err != nil {
		slog.Debug("reviewer", "name", r.Name, "prompt", p, "response", response)
		return nil, fmt.Errorf("reviewer %q: unmarshal issues: %w", r.Name, err)
	}
	for i := range issues {
		issues[i].Reviewer = r.Name
	}
	return issues, nil
}
