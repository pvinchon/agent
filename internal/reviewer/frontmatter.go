package reviewer

import (
	"fmt"
	"strings"
)

// Frontmatter holds the structured metadata from a prompt file's frontmatter block.
type Frontmatter struct {
	Slug        string
	Name        string
	Description string
}

// parseFrontmatter parses the YAML-style frontmatter block from data.
// It returns the typed metadata and the prompt body (content after the closing ---).
// The file must begin with a --- line. All three fields (slug, name, description) are required.
// Slug must not contain commas.
func parseFrontmatter(data string) (Frontmatter, string, error) {
	const delim = "---"

	if !strings.HasPrefix(data, delim+"\n") {
		return Frontmatter{}, "", fmt.Errorf("frontmatter: missing opening ---")
	}

	rest := data[len(delim)+1:]
	idx := strings.Index(rest, "\n"+delim+"\n")
	if idx == -1 {
		return Frontmatter{}, "", fmt.Errorf("frontmatter: missing closing ---")
	}

	block := rest[:idx]
	body := rest[idx+len("\n"+delim+"\n"):]

	var fm Frontmatter
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			return Frontmatter{}, "", fmt.Errorf("frontmatter: invalid line %q", line)
		}
		switch strings.TrimSpace(k) {
		case "slug":
			fm.Slug = strings.TrimSpace(v)
			if strings.Contains(fm.Slug, ",") {
				return Frontmatter{}, "", fmt.Errorf("frontmatter: slug must not contain commas")
			}
		case "name":
			fm.Name = strings.TrimSpace(v)
		case "description":
			fm.Description = strings.TrimSpace(v)
		}
	}

	if fm.Slug == "" {
		return Frontmatter{}, "", fmt.Errorf("frontmatter: missing required field: slug")
	}
	if fm.Name == "" {
		return Frontmatter{}, "", fmt.Errorf("frontmatter: missing required field: name")
	}
	if fm.Description == "" {
		return Frontmatter{}, "", fmt.Errorf("frontmatter: missing required field: description")
	}

	return fm, body, nil
}
