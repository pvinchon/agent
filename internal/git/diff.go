package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// DiffWithDefault returns the raw git diff between the current state and the
// remote default branch (three-dot diff from the common ancestor). Falls back
// to the local default branch when no remote is available.
func DiffWithDefault() (string, error) {
	def := BranchDefault()

	ref := fmt.Sprintf("origin/%s", def)
	out, err := exec.Command("git", "diff", ref+"...HEAD").Output()
	if err != nil {
		ref = def
		out, err = exec.Command("git", "diff", ref+"...HEAD").Output()
		if err != nil {
			return "", err
		}
	}

	return strings.TrimRight(string(out), "\n"), nil
}
