package git

import (
	"os/exec"
	"strings"
)

func BranchCurrent() string {
	out, err := exec.Command("git", "symbolic-ref", "--short", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func BranchDefault() string {
	out, err := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD").Output()
	if err == nil {
		return strings.TrimPrefix(strings.TrimSpace(string(out)), "refs/remotes/origin/")
	}
	return BranchCurrent()
}
