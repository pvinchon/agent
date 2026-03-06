package git

import (
	"fmt"
	"os"
	"os/exec"
)

// Clone performs a shallow clone of a single branch from a remote repository.
// It clones into dir with depth=1, fetching only the specified branch.
func Clone(url, branch, dir string) error {
	cmd := exec.Command("git", "clone", "--depth", "1", "--branch", branch, "--single-branch", url, dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone: %w", err)
	}
	return nil
}

// CommitAll stages all changes and commits them with the provided message.
func CommitAll(dir, message string) error {
	add := exec.Command("git", "add", "--all")
	add.Dir = dir
	add.Stdout = os.Stdout
	add.Stderr = os.Stderr
	if err := add.Run(); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	commit := exec.Command("git", "commit", "-m", message)
	commit.Dir = dir
	commit.Stdout = os.Stdout
	commit.Stderr = os.Stderr
	if err := commit.Run(); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	return nil
}

// Push pushes committed changes to the remote.
func Push(dir string) error {
	cmd := exec.Command("git", "push")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git push: %w", err)
	}
	return nil
}
