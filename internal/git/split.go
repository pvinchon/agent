package git

import (
	"path"
	"strings"
)

// SplitByFile splits a git diff string into per-file diffs.
// Returns a map from file path to the diff chunk for that file.
// The file path is the new path (b/ side) of each changed file.
func SplitByFile(diff string) map[string]string {
	result := make(map[string]string)
	if diff == "" {
		return result
	}

	lines := strings.Split(diff, "\n")

	var currentFile string
	var currentLines []string

	flush := func() {
		if currentFile == "" {
			return
		}
		result[currentFile] = strings.TrimRight(strings.Join(currentLines, "\n"), "\n")
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git ") {
			flush()
			currentFile = fileFromDiffHeader(line)
			currentLines = []string{line}
		} else if currentFile != "" {
			currentLines = append(currentLines, line)
		}
	}
	flush()

	return result
}

// SplitByFolder splits a git diff string into per-folder diffs.
// Returns a map from folder path to the combined diff for all files in that folder.
// Files at the repository root are grouped under ".".
func SplitByFolder(diff string) map[string]string {
	byFile := SplitByFile(diff)
	result := make(map[string]string)
	for filename, fileDiff := range byFile {
		dir := path.Dir(filename)
		if existing, ok := result[dir]; ok {
			result[dir] = existing + "\n" + fileDiff
		} else {
			result[dir] = fileDiff
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
