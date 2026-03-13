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

// Scope defines how broadly a reviewer is applied to the diff.
type Scope string

const (
	// ScopeProject runs the reviewer once against the entire diff.
	ScopeProject Scope = "project"
	// ScopeFolder runs the reviewer once per changed directory.
	ScopeFolder Scope = "folder"
	// ScopeFile runs the reviewer once per changed file.
	ScopeFile Scope = "file"
)

// Reviewer focuses on a specific aspect of code quality, defined by its prompt.
type Reviewer struct {
	Name        string
	Description string
	Scope       Scope
	Prompt      string
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
		meta, body := parseFrontmatter(string(data))
		r := Reviewer{
			Name:   name,
			Scope:  ScopeProject,
			Prompt: body,
		}
		if v, ok := meta["description"]; ok {
			r.Description = v
		}
		if v, ok := meta["scope"]; ok {
			r.Scope = Scope(v)
		}
		m[name] = r
	}
	return m
}()

// parseFrontmatter parses YAML-style frontmatter from a markdown string.
// Frontmatter is delimited by lines containing only "---".
// Returns the key-value pairs and the remaining body content.
func parseFrontmatter(content string) (map[string]string, string) {
	const marker = "---"
	if !strings.HasPrefix(content, marker+"\n") {
		return nil, content
	}
	rest := content[len(marker)+1:] // skip "---\n"
	end := strings.Index(rest, "\n"+marker)
	if end == -1 {
		return nil, content
	}
	meta := make(map[string]string)
	for _, line := range strings.Split(rest[:end], "\n") {
		if k, v, ok := strings.Cut(line, ": "); ok {
			meta[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	body := strings.TrimPrefix(rest[end+len("\n"+marker):], "\n")
	return meta, body
}

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
