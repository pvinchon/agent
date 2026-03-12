package fixer

import (
	_ "embed"
	"fmt"
	"log/slog"
	"strings"

	"github.com/pvinchon/agent/internal/assistant"
	"github.com/pvinchon/agent/internal/reviewer"
	"github.com/pvinchon/agent/internal/x/ux"
)

//go:embed data/prompt_template.md
var promptTemplate string

// Fix asks the assistant to fix all provided issues by editing files directly.
func Fix(issues []reviewer.Issue, diff string, a assistant.Assistant) error {
	defer ux.Spinner()()
	prompt := buildPrompt(issues, diff)

	slog.Debug("fixer", "prompt", prompt)

	response, err := assistant.Prompt(a, prompt)
	if err != nil {
		return fmt.Errorf("fixer: %w", err)
	}

	slog.Debug("fixer", "prompt", prompt, "response", response)

	return nil
}

func buildPrompt(issues []reviewer.Issue, diff string) string {
	return strings.NewReplacer(
		"{{issues}}", formatIssues(issues),
		"{{diff}}", diff,
	).Replace(promptTemplate)
}

func formatIssues(issues []reviewer.Issue) string {
	var sb strings.Builder
	for i, issue := range issues {
		fmt.Fprintf(&sb, "%d. [%s] %s\n   Location: %s\n   %s\n",
			i+1, issue.Severity, issue.Title, issue.Location, issue.Description)
	}
	return strings.TrimRight(sb.String(), "\n")
}
