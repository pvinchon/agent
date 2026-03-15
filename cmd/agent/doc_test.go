package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot returns the repository root directory, verified by the presence of go.mod.
// Go tests run with CWD set to the package directory, so root is two levels up from cmd/agent/.
func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	root := filepath.Join(wd, "..", "..")
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("repo root not found at %s: %v", root, err)
	}
	return root
}

// TestAgentsMDIsSymlink verifies that AGENTS.md is a symlink pointing to CLAUDE.md.
func TestAgentsMDIsSymlink(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "AGENTS.md")

	info, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("lstat AGENTS.md: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("AGENTS.md is not a symlink; it must point to CLAUDE.md")
	}
	target, err := os.Readlink(path)
	if err != nil {
		t.Fatalf("readlink AGENTS.md: %v", err)
	}
	if target != "CLAUDE.md" {
		t.Errorf("AGENTS.md symlink target = %q, want %q", target, "CLAUDE.md")
	}
}

// TestCopilotInstructionsIsSymlink verifies that .github/copilot-instructions.md is a symlink pointing to ../CLAUDE.md.
func TestCopilotInstructionsIsSymlink(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, ".github", "copilot-instructions.md")

	info, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("lstat .github/copilot-instructions.md: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal(".github/copilot-instructions.md is not a symlink; it must point to ../CLAUDE.md")
	}
	target, err := os.Readlink(path)
	if err != nil {
		t.Fatalf("readlink .github/copilot-instructions.md: %v", err)
	}
	if target != "../CLAUDE.md" {
		t.Errorf(".github/copilot-instructions.md symlink target = %q, want %q", target, "../CLAUDE.md")
	}
}

// TestCLAUDEMDMentionsReviewers verifies that CLAUDE.md names every reviewer
// whose prompt lives in internal/reviewer/data/prompts/.
func TestCLAUDEMDMentionsReviewers(t *testing.T) {
	root := repoRoot(t)

	raw, err := os.ReadFile(filepath.Join(root, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	content := string(raw)

	promptsDir := filepath.Join(root, "internal", "reviewer", "data", "prompts")
	entries, err := os.ReadDir(promptsDir)
	if err != nil {
		t.Fatalf("read prompts dir: %v", err)
	}
	for _, e := range entries {
		name := strings.TrimSuffix(e.Name(), ".md")
		// Match the reviewer name as a backtick-quoted code span, matching
		// the Reviewers table format and avoiding false positives (e.g. "go").
		if !strings.Contains(content, "`"+name+"`") {
			t.Errorf("CLAUDE.md does not list reviewer %q in the Reviewers table", name)
		}
	}
}

// TestCLAUDEMDMentionsPackages verifies that CLAUDE.md names every top-level
// package directory under internal/.
func TestCLAUDEMDMentionsPackages(t *testing.T) {
	root := repoRoot(t)

	raw, err := os.ReadFile(filepath.Join(root, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	content := string(raw)

	entries, err := os.ReadDir(filepath.Join(root, "internal"))
	if err != nil {
		t.Fatalf("read internal dir: %v", err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		// Match the package directory with a trailing slash to avoid false
		// positives for short names like "x" or common words like "log".
		if !strings.Contains(content, e.Name()+"/") {
			t.Errorf("CLAUDE.md does not mention internal package %q", e.Name())
		}
	}
}

// TestCLAUDEMDMentionsCommands verifies that CLAUDE.md describes all CLI commands.
func TestCLAUDEMDMentionsCommands(t *testing.T) {
	root := repoRoot(t)

	raw, err := os.ReadFile(filepath.Join(root, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	content := string(raw)

	for _, cmd := range []string{"review", "fix", "loop", "help"} {
		if !strings.Contains(content, cmd) {
			t.Errorf("CLAUDE.md does not mention command %q", cmd)
		}
	}
}
