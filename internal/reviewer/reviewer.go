package reviewer

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"

	"github.com/pvinchon/agent/internal/assistant"
	"github.com/pvinchon/agent/internal/git"
	xsync "github.com/pvinchon/agent/internal/x/sync"
)

//go:embed data/prompt_template.md
var promptTemplate string

//go:embed data/prompts
var prompts embed.FS

// Scope values control how the reviewer is applied to the diff.
const (
	ScopeAll    = "all"    // review all changes at once (default)
	ScopeFolder = "folder" // review changes once per changed folder
	ScopeFile   = "file"   // review changes once per changed file
)

// Reviewer focuses on a specific aspect of code quality, defined by its prompt.
type Reviewer struct {
	Name        string
	Description string
	Scope       string
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
		r := Reviewer{Name: name, Prompt: strings.TrimSpace(body)}
		if desc, ok := meta["description"]; ok {
			r.Description = desc
		}
		if scope, ok := meta["scope"]; ok {
			r.Scope = scope
		}
		m[name] = r
	}
	return m
}()

// parseFrontmatter parses the YAML-like frontmatter block from the beginning of
// content. The block must be delimited by "---" lines. It returns the parsed
// key-value pairs and the remaining body.
func parseFrontmatter(content string) (map[string]string, string) {
	const delimiter = "---\n"
	if !strings.HasPrefix(content, delimiter) {
		return nil, content
	}
	rest := content[len(delimiter):]
	end := strings.Index(rest, delimiter)
	if end == -1 {
		return nil, content
	}
	frontmatter := rest[:end]
	body := rest[end+len(delimiter):]

	meta := make(map[string]string)
	for line := range strings.SplitSeq(frontmatter, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		key, value, ok := strings.Cut(line, ": ")
		if ok {
			meta[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
	}
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

// reviewWithScope runs this reviewer against diff, splitting by file or folder
// when the reviewer's Scope requests it.
func (r Reviewer) reviewWithScope(diff string, a assistant.Assistant) ([]Issue, error) {
	switch r.Scope {
	case ScopeFile:
		return r.reviewSegments(git.SplitByFile(diff), a)
	case ScopeFolder:
		return r.reviewSegments(git.SplitByFolder(diff), a)
	default: // ScopeAll or unset
		return r.review(diff, a)
	}
}

// reviewSegments runs this reviewer in parallel against each diff segment and
// aggregates the results. Errors from individual segments are joined.
func (r Reviewer) reviewSegments(segments map[string]string, a assistant.Assistant) ([]Issue, error) {
	type entry struct {
		diff string
	}
	entries := make([]entry, 0, len(segments))
	for _, d := range segments {
		entries = append(entries, entry{d})
	}

	groups, errs := xsync.Parallel(entries, func(e entry) ([]Issue, error) {
		return r.review(e.diff, a)
	})
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return slices.Concat(groups...), nil
}
