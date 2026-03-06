package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initBareRepo creates a bare repository that acts as a fake remote.
func initBareRepo(t *testing.T) string {
	t.Helper()
	bare := t.TempDir()
	run(t, bare, "git", "init", "--bare")
	return bare
}

// seedRepo creates a non-bare repo, adds a commit on the given branch, and
// pushes it to the bare remote so we have something to clone.
func seedRepo(t *testing.T, remote, branch string) {
	t.Helper()
	work := t.TempDir()
	run(t, work, "git", "clone", remote, ".")
	run(t, work, "git", "checkout", "-b", branch)
	if err := os.WriteFile(filepath.Join(work, "README.md"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, work, "git", "add", "--all")
	run(t, work, "git", "-c", "user.name=test", "-c", "user.email=test@test.com", "commit", "-m", "seed")
	run(t, work, "git", "push", "origin", branch)
}

// run executes a command in dir and fails the test on error.
func run(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, out)
	}
}

func TestClone(t *testing.T) {
	bare := initBareRepo(t)
	seedRepo(t, bare, "main")

	dst := filepath.Join(t.TempDir(), "repo")
	if err := Clone(bare, "main", dst); err != nil {
		t.Fatalf("Clone: %v", err)
	}

	// The cloned repo should contain the seeded file.
	if _, err := os.Stat(filepath.Join(dst, "README.md")); err != nil {
		t.Fatalf("expected README.md in cloned repo: %v", err)
	}
}

func TestClone_InvalidURL(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "repo")
	if err := Clone("https://invalid.invalid/no/repo.git", "main", dst); err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestCommitAll(t *testing.T) {
	bare := initBareRepo(t)
	seedRepo(t, bare, "main")

	dst := filepath.Join(t.TempDir(), "repo")
	if err := Clone(bare, "main", dst); err != nil {
		t.Fatalf("Clone: %v", err)
	}

	// Configure git user for the cloned repo so commit works.
	run(t, dst, "git", "config", "user.name", "test")
	run(t, dst, "git", "config", "user.email", "test@test.com")

	// Create a new file and commit it.
	if err := os.WriteFile(filepath.Join(dst, "new.txt"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CommitAll(dst, "add new file"); err != nil {
		t.Fatalf("CommitAll: %v", err)
	}

	// Verify the commit message is in the log.
	cmd := exec.Command("git", "log", "--oneline", "-1")
	cmd.Dir = dst
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git log: %v", err)
	}
	if !strings.Contains(string(out), "add new file") {
		t.Fatalf("expected commit message in log, got: %s", out)
	}
}

func TestPush(t *testing.T) {
	bare := initBareRepo(t)
	seedRepo(t, bare, "main")

	dst := filepath.Join(t.TempDir(), "repo")
	if err := Clone(bare, "main", dst); err != nil {
		t.Fatalf("Clone: %v", err)
	}

	run(t, dst, "git", "config", "user.name", "test")
	run(t, dst, "git", "config", "user.email", "test@test.com")

	if err := os.WriteFile(filepath.Join(dst, "pushed.txt"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CommitAll(dst, "to push"); err != nil {
		t.Fatalf("CommitAll: %v", err)
	}
	if err := Push(dst); err != nil {
		t.Fatalf("Push: %v", err)
	}

	// Verify the commit landed in the bare remote.
	cmd := exec.Command("git", "log", "--oneline", "-1", "main")
	cmd.Dir = bare
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git log on bare: %v", err)
	}
	if !strings.Contains(string(out), "to push") {
		t.Fatalf("expected pushed commit in bare remote, got: %s", out)
	}
}
