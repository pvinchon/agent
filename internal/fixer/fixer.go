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
var defaultPromptTemplate string

// Fix asks the assistant to fix all provided issues by editing files directly.
// If fixTemplate is non-empty it replaces the built-in fixer prompt template.
func Fix(issues []reviewer.Issue, diff, fixTemplate string, a assistant.Assistant) error {
defer ux.Spinner()()
prompt := buildPrompt(issues, diff, fixTemplate)

slog.Debug("fixer", "prompt", prompt)

response, err := assistant.Prompt(a, prompt)
if err != nil {
return fmt.Errorf("fixer: %w", err)
}

slog.Debug("fixer", "prompt", prompt, "response", response)

return nil
}

func buildPrompt(issues []reviewer.Issue, diff, fixTemplate string) string {
tmpl := defaultPromptTemplate
if fixTemplate != "" {
tmpl = fixTemplate
}
return strings.NewReplacer(
"{{issues}}", formatIssues(issues),
"{{diff}}", diff,
).Replace(tmpl)
}

func formatIssues(issues []reviewer.Issue) string {
var sb strings.Builder
for i, issue := range issues {
fmt.Fprintf(&sb, "%d. [%s] %s\n   Location: %s\n   %s\n",
i+1, issue.Severity, issue.Title, issue.Location, issue.Description)
}
return strings.TrimRight(sb.String(), "\n")
}
