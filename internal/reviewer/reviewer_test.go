package reviewer

import (
	"os/exec"
	"strings"
	"testing"
)

func TestNewReviewer(t *testing.T) {
	r, err := New("security")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Name != "security" {
		t.Errorf("got name %q, want %q", r.Name, "security")
	}
	if r.Prompt == "" {
		t.Error("prompt is empty")
	}
}

func TestNew_unknown(t *testing.T) {
	_, err := New("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown reviewer")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error should mention the unknown name, got: %v", err)
	}
}

func TestBuildPrompt(t *testing.T) {
	r, err := New("security")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	diff := `--- a/main.go
+++ b/main.go
@@ -1 +1 @@
-fmt.Println("hello")
+fmt.Println(userInput)`

	prompt := r.buildPrompt(diff)

	if !strings.Contains(prompt, "You are a senior reviewer") {
		t.Error("prompt does not contain base template text")
	}
	if !strings.Contains(prompt, r.Prompt) {
		t.Error("prompt does not contain reviewer-specific prompt")
	}
	if !strings.Contains(prompt, diff) {
		t.Error("prompt does not contain the diff")
	}
}

func TestBuildPrompt_order(t *testing.T) {
	r, _ := New("security")
	diff := "some diff"
	prompt := r.buildPrompt(diff)

	baseIdx := strings.Index(prompt, "You are a senior reviewer")
	reviewerIdx := strings.Index(prompt, r.Prompt)
	diffIdx := strings.Index(prompt, diff)

	if !(baseIdx < reviewerIdx && reviewerIdx < diffIdx) {
		t.Error("expected order: base template text, reviewer prompt, diff")
	}
}

func TestReview_issues(t *testing.T) {
	r := Reviewer{Name: "security", Prompt: "check security"}
	a := &fakeAssistant{fn: func(string) *exec.Cmd {
		return echoCmd(`[{"severity":"HIGH","title":"SQL injection","location":"db.go:10","description":"User input in query."}]`)
	}}

	issues, err := r.review("diff", a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("got %d issues, want 1", len(issues))
	}
	if issues[0].Reviewer != "security" {
		t.Errorf("got reviewer %q, want %q", issues[0].Reviewer, "security")
	}
	if issues[0].Severity != "HIGH" {
		t.Errorf("got severity %q, want %q", issues[0].Severity, "HIGH")
	}
	if issues[0].Title != "SQL injection" {
		t.Errorf("got title %q, want %q", issues[0].Title, "SQL injection")
	}
}

func TestReview_empty(t *testing.T) {
	r := Reviewer{Name: "security", Prompt: "check security"}
	a := &fakeAssistant{fn: func(string) *exec.Cmd { return echoCmd("[]") }}

	issues, err := r.review("diff", a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("got %d issues, want 0", len(issues))
	}
}

func TestReview_invalidJSON(t *testing.T) {
	r := Reviewer{Name: "security", Prompt: "check security"}
	a := &fakeAssistant{fn: func(string) *exec.Cmd { return echoCmd("not json") }}

	_, err := r.review("diff", a)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "unmarshal issues") {
		t.Errorf("error should mention unmarshal issues, got: %v", err)
	}
}

// fakeAssistant implements assistant.Assistant for tests.
type fakeAssistant struct {
	fn func(string) *exec.Cmd
}

func (f *fakeAssistant) Command(prompt string) *exec.Cmd {
	return f.fn(prompt)
}

func echoCmd(output string) *exec.Cmd { return exec.Command("echo", output) }
func failCmd() *exec.Cmd              { return exec.Command("false") }

// singleFileDiff is a minimal valid git diff for one file.
const singleFileDiff = `diff --git a/main.go b/main.go
index abc1234..def5678 100644
--- a/main.go
+++ b/main.go
@@ -1 +1 @@
-fmt.Println("hello")
+fmt.Println(userInput)`

func TestReview(t *testing.T) {
	sec, _ := New("security")
	tests, _ := New("tests")

	a := &fakeAssistant{fn: func(string) *exec.Cmd {
		return echoCmd(`[{"severity":"HIGH","title":"issue","location":"f.go:1","description":"bad"}]`)
	}}

	issues, errs := Review([]Reviewer{sec, tests}, singleFileDiff, a)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(issues) != 2 {
		t.Fatalf("got %d issues, want 2", len(issues))
	}

	reviewerNames := map[string]bool{}
	for _, f := range issues {
		reviewerNames[f.Reviewer] = true
	}
	if !reviewerNames["security"] || !reviewerNames["tests"] {
		t.Errorf("expected issues from both reviewers, got: %v", reviewerNames)
	}
}

func TestReview_promptError(t *testing.T) {
	sec, _ := New("security")
	a := &fakeAssistant{fn: func(string) *exec.Cmd { return failCmd() }}

	issues, errs := Review([]Reviewer{sec}, singleFileDiff, a)
	if len(errs) != 1 {
		t.Fatalf("got %d errors, want 1", len(errs))
	}
	if len(issues) != 0 {
		t.Errorf("got %d issues, want 0 on error", len(issues))
	}
}

func TestReview_partialFailure(t *testing.T) {
	ok := Reviewer{Name: "ok", Prompt: "ok"}
	bad := Reviewer{Name: "bad", Prompt: "bad"}

	a := &fakeAssistant{fn: func(prompt string) *exec.Cmd {
		if strings.Contains(prompt, "bad") {
			return failCmd()
		}
		return echoCmd(`[{"severity":"LOW","title":"minor","location":"f.go:1","description":"ok"}]`)
	}}

	issues, errs := Review([]Reviewer{ok, bad}, "diff", a)
	if len(errs) != 1 {
		t.Fatalf("got %d errors, want 1", len(errs))
	}
	if len(issues) != 1 {
		t.Fatalf("got %d issues, want 1", len(issues))
	}
	if issues[0].Reviewer != "ok" {
		t.Errorf("got reviewer %q, want %q", issues[0].Reviewer, "ok")
	}
}

func TestResolve(t *testing.T) {
	reviewers, err := resolve("security,tests")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reviewers) != 2 {
		t.Fatalf("got %d reviewers, want 2", len(reviewers))
	}
	if reviewers[0].Name != "security" {
		t.Errorf("got %q, want %q", reviewers[0].Name, "security")
	}
	if reviewers[1].Name != "tests" {
		t.Errorf("got %q, want %q", reviewers[1].Name, "tests")
	}
}

func TestResolve_empty(t *testing.T) {
	reviewers, err := resolve("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reviewers != nil {
		t.Errorf("expected nil for empty input, got %v", reviewers)
	}
}

func TestResolve_unknown(t *testing.T) {
	_, err := resolve("security,bogus")
	if err == nil {
		t.Fatal("expected error for unknown reviewer")
	}
	if !strings.Contains(err.Error(), "bogus") {
		t.Errorf("error should mention the unknown name, got: %v", err)
	}
}

func TestResolve_whitespace(t *testing.T) {
	reviewers, err := resolve("security, tests")
	if err != nil {
		t.Fatalf("unexpected error with whitespace: %v", err)
	}
	if len(reviewers) != 2 {
		t.Fatalf("got %d reviewers, want 2", len(reviewers))
	}
}

func TestParseFrontmatter(t *testing.T) {
	content := "---\nname: Security\ndescription: Checks for vulnerabilities\nscope: file\n---\n## Focus\n\nBody text."
	meta, body := parseFrontmatter(content)

	if meta["name"] != "Security" {
		t.Errorf("got name %q, want %q", meta["name"], "Security")
	}
	if meta["description"] != "Checks for vulnerabilities" {
		t.Errorf("got description %q, want %q", meta["description"], "Checks for vulnerabilities")
	}
	if meta["scope"] != "file" {
		t.Errorf("got scope %q, want %q", meta["scope"], "file")
	}
	if body != "## Focus\n\nBody text." {
		t.Errorf("got body %q, want %q", body, "## Focus\n\nBody text.")
	}
}

func TestParseFrontmatter_noFrontmatter(t *testing.T) {
	content := "## Focus\n\nBody text."
	meta, body := parseFrontmatter(content)

	if meta != nil {
		t.Errorf("expected nil meta for content without frontmatter, got %v", meta)
	}
	if body != content {
		t.Errorf("got body %q, want original content", body)
	}
}

func TestParseFrontmatter_unclosed(t *testing.T) {
	content := "---\nname: Security\n## Focus\n\nBody text."
	meta, body := parseFrontmatter(content)

	if meta != nil {
		t.Errorf("expected nil meta for unclosed frontmatter, got %v", meta)
	}
	if body != content {
		t.Errorf("got body %q, want original content", body)
	}
}

func TestReviewer_scope(t *testing.T) {
	sec, _ := New("security")
	if sec.Scope != ScopeFile {
		t.Errorf("security scope = %q, want %q", sec.Scope, ScopeFile)
	}

	arch, _ := New("architecture")
	if arch.Scope != ScopeProject {
		t.Errorf("architecture scope = %q, want %q", arch.Scope, ScopeProject)
	}
}

func TestReviewer_description(t *testing.T) {
	sec, _ := New("security")
	if sec.Description == "" {
		t.Error("security description is empty")
	}
}

// twoFileDiff is a minimal valid git diff for two files in different directories.
const twoFileDiff = `diff --git a/pkg/foo/foo.go b/pkg/foo/foo.go
index abc1234..def5678 100644
--- a/pkg/foo/foo.go
+++ b/pkg/foo/foo.go
@@ -1 +1 @@
-old
+new
diff --git a/pkg/bar/bar.go b/pkg/bar/bar.go
index abc1234..def5678 100644
--- a/pkg/bar/bar.go
+++ b/pkg/bar/bar.go
@@ -1 +1 @@
-old
+new`

func TestSplitDiffByFile(t *testing.T) {
	files := splitDiffByFile(twoFileDiff)
	if len(files) != 2 {
		t.Fatalf("got %d files, want 2", len(files))
	}
	if files[0].path != "pkg/foo/foo.go" {
		t.Errorf("got path %q, want %q", files[0].path, "pkg/foo/foo.go")
	}
	if files[1].path != "pkg/bar/bar.go" {
		t.Errorf("got path %q, want %q", files[1].path, "pkg/bar/bar.go")
	}
	if !strings.Contains(files[0].diff, "diff --git") {
		t.Error("file diff does not contain 'diff --git'")
	}
}

func TestSplitDiffByFile_empty(t *testing.T) {
	if splitDiffByFile("") != nil {
		t.Error("expected nil for empty diff")
	}
}

func TestGroupByFolder(t *testing.T) {
	files := splitDiffByFile(twoFileDiff)
	folders := groupByFolder(files)
	if len(folders) != 2 {
		t.Fatalf("got %d folders, want 2", len(folders))
	}
	if folders[0].path != "pkg/foo" {
		t.Errorf("got folder %q, want %q", folders[0].path, "pkg/foo")
	}
	if folders[1].path != "pkg/bar" {
		t.Errorf("got folder %q, want %q", folders[1].path, "pkg/bar")
	}
}

func TestGroupByFolder_sameFolder(t *testing.T) {
	sameFolderDiff := `diff --git a/pkg/foo/a.go b/pkg/foo/a.go
index abc1234..def5678 100644
--- a/pkg/foo/a.go
+++ b/pkg/foo/a.go
@@ -1 +1 @@
-old
+new
diff --git a/pkg/foo/b.go b/pkg/foo/b.go
index abc1234..def5678 100644
--- a/pkg/foo/b.go
+++ b/pkg/foo/b.go
@@ -1 +1 @@
-old
+new`
	files := splitDiffByFile(sameFolderDiff)
	folders := groupByFolder(files)
	if len(folders) != 1 {
		t.Fatalf("got %d folders, want 1", len(folders))
	}
	if folders[0].path != "pkg/foo" {
		t.Errorf("got folder %q, want %q", folders[0].path, "pkg/foo")
	}
	// Both file diffs should be combined
	if !strings.Contains(folders[0].diff, "a.go") || !strings.Contains(folders[0].diff, "b.go") {
		t.Error("folder diff should contain both file diffs")
	}
}

func TestGroupByFolder_empty(t *testing.T) {
	if groupByFolder(nil) != nil {
		t.Error("expected nil for empty input")
	}
}

func TestReview_fileScope(t *testing.T) {
	r := Reviewer{Name: "test", Prompt: "check", Scope: ScopeFile}
	callCount := 0
	a := &fakeAssistant{fn: func(prompt string) *exec.Cmd {
		callCount++
		return echoCmd(`[{"severity":"LOW","title":"issue","location":"f.go:1","description":"ok"}]`)
	}}

	// twoFileDiff has 2 files → reviewer runs twice
	issues, errs := Review([]Reviewer{r}, twoFileDiff, a)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(issues) != 2 {
		t.Fatalf("got %d issues, want 2 (one per file)", len(issues))
	}
	if callCount != 2 {
		t.Errorf("assistant called %d times, want 2 (once per file)", callCount)
	}
}

func TestReview_folderScope(t *testing.T) {
	r := Reviewer{Name: "test", Prompt: "check", Scope: ScopeFolder}
	callCount := 0
	a := &fakeAssistant{fn: func(prompt string) *exec.Cmd {
		callCount++
		return echoCmd(`[{"severity":"LOW","title":"issue","location":"f.go:1","description":"ok"}]`)
	}}

	// twoFileDiff has 2 files in 2 different folders → reviewer runs twice
	issues, errs := Review([]Reviewer{r}, twoFileDiff, a)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(issues) != 2 {
		t.Fatalf("got %d issues, want 2 (one per folder)", len(issues))
	}
	if callCount != 2 {
		t.Errorf("assistant called %d times, want 2 (once per folder)", callCount)
	}
}

func TestReview_projectScope(t *testing.T) {
	r := Reviewer{Name: "test", Prompt: "check", Scope: ScopeProject}
	callCount := 0
	a := &fakeAssistant{fn: func(prompt string) *exec.Cmd {
		callCount++
		return echoCmd(`[{"severity":"LOW","title":"issue","location":"f.go:1","description":"ok"}]`)
	}}

	// project scope → reviewer runs once regardless of file count
	issues, errs := Review([]Reviewer{r}, twoFileDiff, a)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(issues) != 1 {
		t.Fatalf("got %d issues, want 1 (once for whole diff)", len(issues))
	}
	if callCount != 1 {
		t.Errorf("assistant called %d times, want 1 (once for whole diff)", callCount)
	}
}
