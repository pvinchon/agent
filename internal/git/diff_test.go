package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// makeRepoWithRemote creates a bare origin and a clone with one empty commit
// pushed to it. Returns the clone directory.
func makeRepoWithRemote(t *testing.T) string {
	t.Helper()

	origin := t.TempDir()
	cmd := exec.Command("git", "init", "--bare", "-b", "main")
	cmd.Dir = origin
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init --bare: %v\n%s", err, out)
	}

	cloneDir := filepath.Join(t.TempDir(), "clone")
	if out, err := exec.Command("git", "clone", origin, cloneDir).CombinedOutput(); err != nil {
		t.Fatalf("git clone: %v\n%s", err, out)
	}

	for _, args := range [][]string{
		{"-C", cloneDir, "commit", "--allow-empty", "-m", "init"},
		{"-C", cloneDir, "push"},
	} {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	return cloneDir
}

func TestDiffWithDefault_NoChanges(t *testing.T) {
	cloneDir := makeRepoWithRemote(t)
	t.Chdir(cloneDir)

	got, err := DiffWithDefault()
	if err != nil {
		t.Fatalf("DiffWithDefault() error: %v", err)
	}
	if got != "" {
		t.Errorf("DiffWithDefault() = %q, want empty string", got)
	}
}

func TestDiffWithDefault_WithChanges(t *testing.T) {
	cloneDir := makeRepoWithRemote(t)

	// Add a new file and commit it locally (not pushed).
	newFile := filepath.Join(cloneDir, "hello.txt")
	if err := os.WriteFile(newFile, []byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"-C", cloneDir, "add", "hello.txt"},
		{"-C", cloneDir, "commit", "-m", "add hello.txt"},
	} {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	t.Chdir(cloneDir)

	got, err := DiffWithDefault()
	if err != nil {
		t.Fatalf("DiffWithDefault() error: %v", err)
	}
	if !strings.Contains(got, "diff --git") {
		t.Errorf("DiffWithDefault() output missing 'diff --git':\n%s", got)
	}
	if !strings.Contains(got, "hello.txt") {
		t.Errorf("DiffWithDefault() output missing 'hello.txt':\n%s", got)
	}
}

func TestDiffWithDefault_NoRemote(t *testing.T) {
	dir := makeRepo(t)
	t.Chdir(dir)

	// No remote exists; should fall back to local default branch without error.
	_, err := DiffWithDefault()
	if err != nil {
		t.Fatalf("DiffWithDefault() error: %v", err)
	}
}
