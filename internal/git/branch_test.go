package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func makeRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "-b", "main"},
		{"commit", "--allow-empty", "-m", "init"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return dir
}

func TestBranchCurrent(t *testing.T) {
	dir := makeRepo(t)
	t.Chdir(dir)

	got := BranchCurrent()
	if got != "main" {
		t.Errorf("BranchCurrent() = %q, want %q", got, "main")
	}
}

func TestBranchCurrent_NotARepo(t *testing.T) {
	t.Chdir(t.TempDir())

	got := BranchCurrent()
	if got != "" {
		t.Errorf("BranchCurrent() = %q, want empty string", got)
	}
}

func TestBranchDefault_FallsBackToBranchCurrent(t *testing.T) {
	dir := makeRepo(t)
	t.Chdir(dir)

	got := BranchDefault()
	if got != "main" {
		t.Errorf("BranchDefault() = %q, want %q", got, "main")
	}
}

func TestBranchDefault_WithRemote(t *testing.T) {
	// Create a bare repo to act as origin
	origin := t.TempDir()
	cmd := exec.Command("git", "init", "--bare", "-b", "trunk")
	cmd.Dir = origin
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init --bare: %v\n%s", err, out)
	}

	// Clone from origin so refs/remotes/origin/HEAD is set
	cloneDir := filepath.Join(t.TempDir(), "clone")
	if out, err := exec.Command("git", "clone", origin, cloneDir).CombinedOutput(); err != nil {
		t.Fatalf("git clone: %v\n%s", err, out)
	}

	// Commit something so the clone is not empty
	for _, args := range [][]string{
		{"-C", cloneDir, "commit", "--allow-empty", "-m", "init"},
		{"-C", cloneDir, "push"},
	} {
		cmd := exec.Command("git", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	// Set remote HEAD explicitly so symbolic-ref resolves
	if out, err := exec.Command("git", "-C", cloneDir, "remote", "set-head", "origin", "trunk").CombinedOutput(); err != nil {
		t.Fatalf("git remote set-head: %v\n%s", err, out)
	}

	t.Chdir(cloneDir)

	got := BranchDefault()
	if got != "trunk" {
		t.Errorf("BranchDefault() = %q, want %q", got, "trunk")
	}
}

func TestBranchDefault_NotARepo(t *testing.T) {
	t.Chdir(t.TempDir())

	got := BranchDefault()
	if got != "" {
		t.Errorf("BranchDefault() = %q, want empty string", got)
	}
}

func init() {
	// Ensure a clean git environment for all tests
	for _, key := range []string{"GIT_DIR", "GIT_WORK_TREE"} {
		os.Unsetenv(key)
	}
}
