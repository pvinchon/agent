package reviewer

import (
	"embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pvinchon/agent/internal/assistant"
)

//go:embed data/prompt_template.md
var defaultPromptTemplate string

//go:embed data/prompt_template_output.md
var outputTemplate string

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

// resolve parses a comma-separated list of reviewer names from the built-in
// registry and returns the corresponding Reviewers. Returns an error if any
// name is unknown.
func resolve(names string) ([]Reviewer, error) {
	return resolveFrom(reviewersByName, names)
}

// resolveFrom parses a comma-separated list of reviewer names from registry
// and returns the corresponding Reviewers. Returns an error if any name is
// unknown.
func resolveFrom(registry map[string]Reviewer, names string) ([]Reviewer, error) {
	if names == "" {
		return nil, nil
	}
	allNames := strings.Join(slices.Sorted(maps.Keys(registry)), ", ")
	var result []Reviewer
	for name := range strings.SplitSeq(names, ",") {
		name = strings.TrimSpace(name)
		r, ok := registry[name]
		if !ok {
			return nil, fmt.Errorf("unknown reviewer %q: choose from %s", name, allNames)
		}
		result = append(result, r)
	}
	return result, nil
}

// LoadPromptsDir loads all .md files from dir as reviewer prompts. The file
// name (without the .md extension) becomes the reviewer name.
func LoadPromptsDir(dir string) (map[string]Reviewer, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("prompts-dir: %w", err)
	}
	m := make(map[string]Reviewer)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".md")
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("prompts-dir: %w", err)
		}
		m[name] = Reviewer{Name: name, Prompt: string(data)}
	}
	return m, nil
}

// New returns the Reviewer registered under name.
func New(name string) (Reviewer, error) {
	r, ok := reviewersByName[name]
	if !ok {
		return Reviewer{}, fmt.Errorf("unknown reviewer %q: choose from %s", name, reviewerNames)
	}
	return r, nil
}

// buildPrompt assembles the full prompt for this reviewer against the provided
// diff. If contextTemplate is non-empty it is used as the scene-setting part of
// the prompt; the output specification is always appended from our built-in
// template to guarantee a consistent JSON response. When contextTemplate is
// empty the full built-in prompt template is used unchanged.
func (r Reviewer) buildPrompt(diff, contextTemplate string) string {
	if contextTemplate == "" {
		return strings.NewReplacer("{{prompt}}", r.Prompt, "{{diff}}", diff).Replace(defaultPromptTemplate)
	}
	context := strings.NewReplacer("{{prompt}}", r.Prompt, "{{diff}}", diff).Replace(contextTemplate)
	return context + "\n\n" + outputTemplate
}

// review runs this reviewer against diff and returns the parsed issues.
func (r Reviewer) review(diff, reviewTemplate string, a assistant.Assistant) ([]Issue, error) {
	prompt := r.buildPrompt(diff, reviewTemplate)

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
