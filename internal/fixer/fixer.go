package fixer

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/pvinchon/agent/internal/assistant"
	"github.com/pvinchon/agent/internal/prompt"
	"github.com/pvinchon/agent/internal/reviewer"
	"github.com/pvinchon/agent/internal/x/ux"
)

// Fix asks the assistant to fix all provided issues by editing files directly.
func Fix(issues []reviewer.Issue, diff string, a assistant.Assistant, tmpl prompt.Prompt) error {
	defer ux.Spinner()()
	p := buildPrompt(issues, diff, tmpl)

	slog.Debug("fixer", "prompt", p)

	response, err := assistant.Prompt(a, p)
	if err != nil {
		return fmt.Errorf("fixer: %w", err)
	}

	slog.Debug("fixer", "prompt", p, "response", response)

	return nil
}

func buildPrompt(issues []reviewer.Issue, diff string, tmpl prompt.Prompt) string {
	return strings.NewReplacer(
		"{{issues}}", formatIssues(issues),
		"{{diff}}", diff,
	).Replace(tmpl.String())
}

func formatIssues(issues []reviewer.Issue) string {
	var sb strings.Builder
	for i, issue := range issues {
		fmt.Fprintf(&sb, "%d. [%s] %s\n   Location: %s\n   %s\n",
			i+1, issue.Severity, issue.Title, issue.Location, issue.Description)
	}
	return strings.TrimRight(sb.String(), "\n")
}
