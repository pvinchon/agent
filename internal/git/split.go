package git

import (
	"path"
	"strings"
)

// FileDiff holds the parsed diff for a single file.
// Path is the new file path (b/ side) from the diff header.
// Content is the complete diff text for that file.
type FileDiff struct {
	Path    string
	Content string
}

// ParseDiff parses a raw git diff string into a slice of FileDiff, one entry
// per changed file.
func ParseDiff(diff string) []FileDiff {
	var result []FileDiff
	if diff == "" {
		return result
	}

	lines := strings.Split(diff, "\n")

	var current *FileDiff
	var currentLines []string

	flush := func() {
		if current == nil {
			return
		}
		current.Content = strings.TrimRight(strings.Join(currentLines, "\n"), "\n")
		result = append(result, *current)
		current = nil
		currentLines = nil
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git ") {
			flush()
			current = &FileDiff{Path: fileFromDiffHeader(line)}
			currentLines = []string{line}
		} else if current != nil {
			currentLines = append(currentLines, line)
		}
	}
	flush()

	return result
}

// SplitByFile returns a map from file path to diff content for each FileDiff.
func SplitByFile(diffs []FileDiff) map[string]string {
	result := make(map[string]string, len(diffs))
	for _, d := range diffs {
		result[d.Path] = d.Content
	}
	return result
}

// SplitByFolder returns a map from folder path to the combined diff for all
// files in that folder. Files at the repository root are grouped under ".".
func SplitByFolder(diffs []FileDiff) map[string]string {
	result := make(map[string]string)
	for _, d := range diffs {
		dir := path.Dir(d.Path)
		if existing, ok := result[dir]; ok {
			result[dir] = existing + "\n" + d.Content
		} else {
			result[dir] = d.Content
		}
	}
	return result
}

// fileFromDiffHeader extracts the new file path from a "diff --git a/X b/X" header line.
func fileFromDiffHeader(line string) string {
	// Format: "diff --git a/<path> b/<path>"
	fields := strings.Fields(line)
	if len(fields) >= 4 {
		return strings.TrimPrefix(fields[3], "b/")
	}
	return ""
}
