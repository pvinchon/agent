package reviewer

import (
	"os/exec"
	"strings"
	"sync/atomic"
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

func (f *fakeAssistant) Command(prompt string) *exec.Cmd { return f.fn(prompt) }

func echoCmd(output string) *exec.Cmd { return exec.Command("echo", output) }
func failCmd() *exec.Cmd              { return exec.Command("false") }

func TestReview(t *testing.T) {
	sec := Reviewer{Name: "security", Scope: ScopeAll, Prompt: "check security"}
	tests := Reviewer{Name: "tests", Scope: ScopeAll, Prompt: "check tests"}

	a := &fakeAssistant{fn: func(string) *exec.Cmd {
		return echoCmd(`[{"severity":"HIGH","title":"issue","location":"f.go:1","description":"bad"}]`)
	}}

	issues, errs := Review([]Reviewer{sec, tests}, "some diff", a)
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
	sec := Reviewer{Name: "security", Scope: ScopeAll, Prompt: "check security"}
	a := &fakeAssistant{fn: func(string) *exec.Cmd { return failCmd() }}

	issues, errs := Review([]Reviewer{sec}, "diff", a)
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

func TestNewReviewer_metadata(t *testing.T) {
	r, err := New("security")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Description == "" {
		t.Error("description should be set from frontmatter")
	}
	if r.Scope == "" {
		t.Error("scope should be set from frontmatter")
	}
	if r.Scope != ScopeFile {
		t.Errorf("got scope %q for security reviewer, want %q", r.Scope, ScopeFile)
	}
	// Prompt should not contain frontmatter markers
	if strings.Contains(r.Prompt, "---") {
		t.Error("prompt should not contain frontmatter delimiters")
	}
}

func TestParseFrontmatter(t *testing.T) {
	content := "---\nname: Test\ndescription: A test reviewer\nscope: file\n---\n\n## Body\n"
	meta, body := parseFrontmatter(content)

	if meta["name"] != "Test" {
		t.Errorf("got name %q, want %q", meta["name"], "Test")
	}
	if meta["description"] != "A test reviewer" {
		t.Errorf("got description %q, want %q", meta["description"], "A test reviewer")
	}
	if meta["scope"] != "file" {
		t.Errorf("got scope %q, want %q", meta["scope"], "file")
	}
	if !strings.Contains(body, "## Body") {
		t.Errorf("body should contain content after frontmatter, got: %q", body)
	}
	if strings.Contains(body, "---") {
		t.Error("body should not contain frontmatter delimiters")
	}
}

func TestParseFrontmatter_noFrontmatter(t *testing.T) {
	content := "## Just a body\nno frontmatter here\n"
	meta, body := parseFrontmatter(content)

	if meta != nil {
		t.Errorf("expected nil meta for content without frontmatter, got %v", meta)
	}
	if body != content {
		t.Errorf("body should equal original content when no frontmatter, got %q", body)
	}
}

func TestReviewWithScope_file(t *testing.T) {
	r := Reviewer{Name: "sec", Scope: ScopeFile, Prompt: "check security"}
	diff := `diff --git a/foo.go b/foo.go
--- a/foo.go
+++ b/foo.go
@@ -1 +1 @@
-old
+new
diff --git a/bar.go b/bar.go
--- a/bar.go
+++ b/bar.go
@@ -1 +1 @@
-old
+new`

	var promptCount atomic.Int32
	a := &fakeAssistant{fn: func(string) *exec.Cmd {
		promptCount.Add(1)
		return echoCmd(`[{"severity":"LOW","title":"t","location":"f.go:1","description":"d"}]`)
	}}

	issues, err := r.reviewWithScope(diff, a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := promptCount.Load(); got != 2 {
		t.Errorf("expected 2 prompts for 2 files, got %d", got)
	}
	if len(issues) != 2 {
		t.Errorf("expected 2 issues (one per file), got %d", len(issues))
	}
}

func TestReviewWithScope_folder(t *testing.T) {
	r := Reviewer{Name: "arch", Scope: ScopeFolder, Prompt: "check architecture"}
	diff := `diff --git a/internal/foo/foo.go b/internal/foo/foo.go
--- a/internal/foo/foo.go
+++ b/internal/foo/foo.go
@@ -1 +1 @@
-old
+new
diff --git a/internal/bar/bar.go b/internal/bar/bar.go
--- a/internal/bar/bar.go
+++ b/internal/bar/bar.go
@@ -1 +1 @@
-old
+new`

	var promptCount atomic.Int32
	a := &fakeAssistant{fn: func(string) *exec.Cmd {
		promptCount.Add(1)
		return echoCmd(`[]`)
	}}

	_, err := r.reviewWithScope(diff, a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := promptCount.Load(); got != 2 {
		t.Errorf("expected 2 prompts for 2 folders, got %d", got)
	}
}

func TestReviewWithScope_all(t *testing.T) {
	r := Reviewer{Name: "dup", Scope: ScopeAll, Prompt: "check duplication"}

	var promptCount int
	a := &fakeAssistant{fn: func(string) *exec.Cmd {
		promptCount++
		return echoCmd(`[]`)
	}}

	_, err := r.reviewWithScope("some diff", a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if promptCount != 1 {
		t.Errorf("expected 1 prompt for scope=all, got %d", promptCount)
	}
}

func TestReviewWithScope_defaultIsAll(t *testing.T) {
	r := Reviewer{Name: "x", Prompt: "check stuff"} // Scope is zero value

	var promptCount int
	a := &fakeAssistant{fn: func(string) *exec.Cmd {
		promptCount++
		return echoCmd(`[]`)
	}}

	_, err := r.reviewWithScope("some diff", a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if promptCount != 1 {
		t.Errorf("expected 1 prompt when scope is unset, got %d", promptCount)
	}
}
