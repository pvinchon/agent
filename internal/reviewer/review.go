package reviewer

import (
	"path"
	"slices"
	"strings"

	"github.com/pvinchon/agent/internal/assistant"
	"github.com/pvinchon/agent/internal/x/sync"
	"github.com/pvinchon/agent/internal/x/ux"
)

// Issue represents a single issue found by a reviewer.
type Issue struct {
	Reviewer    string `json:"reviewer"`
	Severity    string `json:"severity"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Location    string `json:"location"`
}

// fileDiff pairs a file path with its portion of a git diff.
type fileDiff struct {
	path string
	diff string
}

// splitDiffByFile splits a unified git diff into per-file diffs.
func splitDiffByFile(diff string) []fileDiff {
	if diff == "" {
		return nil
	}
	var result []fileDiff
	var currentPath string
	var currentLines []string
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "diff --git ") {
			if currentPath != "" {
				result = append(result, fileDiff{path: currentPath, diff: strings.Join(currentLines, "\n")})
			}
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				currentPath = strings.TrimPrefix(fields[3], "b/")
			}
			currentLines = []string{line}
		} else if currentPath != "" {
			currentLines = append(currentLines, line)
		}
	}
	if currentPath != "" {
		result = append(result, fileDiff{path: currentPath, diff: strings.Join(currentLines, "\n")})
	}
	return result
}

// groupByFolder groups per-file diffs by their parent directory,
// preserving insertion order.
func groupByFolder(files []fileDiff) []fileDiff {
	if len(files) == 0 {
		return nil
	}
	folderDiffs := make(map[string][]string)
	var order []string
	for _, f := range files {
		folder := path.Dir(f.path)
		if _, exists := folderDiffs[folder]; !exists {
			order = append(order, folder)
		}
		folderDiffs[folder] = append(folderDiffs[folder], f.diff)
	}
	result := make([]fileDiff, 0, len(order))
	for _, folder := range order {
		result = append(result, fileDiff{path: folder, diff: strings.Join(folderDiffs[folder], "\n")})
	}
	return result
}

// reviewTask pairs a reviewer with the diff segment it should evaluate.
type reviewTask struct {
	r    Reviewer
	diff string
}

// Review runs all reviewers in parallel against the provided diff using the
// given Assistant. Each reviewer is applied according to its Scope: once for
// the whole diff (project), once per directory (folder), or once per file
// (file). Returns the aggregated issues and any errors.
func Review(reviewers []Reviewer, diff string, a assistant.Assistant) ([]Issue, []error) {
	defer ux.Spinner()()

	var byFile []fileDiff
	var byFolder []fileDiff

	var tasks []reviewTask
	for _, r := range reviewers {
		switch r.Scope {
		case ScopeFile:
			if byFile == nil {
				byFile = splitDiffByFile(diff)
			}
			for _, f := range byFile {
				tasks = append(tasks, reviewTask{r, f.diff})
			}
		case ScopeFolder:
			if byFile == nil {
				byFile = splitDiffByFile(diff)
			}
			if byFolder == nil {
				byFolder = groupByFolder(byFile)
			}
			for _, f := range byFolder {
				tasks = append(tasks, reviewTask{r, f.diff})
			}
		default: // ScopeProject
			tasks = append(tasks, reviewTask{r, diff})
		}
	}

	groups, errs := sync.Parallel(tasks, func(t reviewTask) ([]Issue, error) {
		return t.r.review(t.diff, a)
	})
	return slices.Concat(groups...), errs
}
