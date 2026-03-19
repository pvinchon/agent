package reviewer

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/pvinchon/agent/internal/prompt"
)

const syntheticTemplate = "TEMPLATE {{prompt}} MIDDLE {{diff}} END"

func makeReviewer(t *testing.T, name, focus string) Reviewer {
	t.Helper()
	return Reviewer{
		Name:     name,
		Prompt:   prompt.New(focus),
		Template: prompt.New(syntheticTemplate),
	}
}

// fakeAssistant implements assistant.Assistant for tests.
type fakeAssistant struct {
	fn func(string) *exec.Cmd
}

func (f *fakeAssistant) Command(p string) *exec.Cmd { return f.fn(p) }

func echoCmd(output string) *exec.Cmd { return exec.Command("echo", output) }
func failCmd() *exec.Cmd              { return exec.Command("false") }

func TestBuildPrompt(t *testing.T) {
	r := makeReviewer(t, "security", "check security issues")
	diff := "some diff content"

	result := r.buildPrompt(diff)

	if !strings.Contains(result, "check security issues") {
		t.Error("prompt does not contain reviewer focus")
	}
	if !strings.Contains(result, diff) {
		t.Error("prompt does not contain the diff")
	}
	if !strings.Contains(result, "TEMPLATE") {
		t.Error("prompt does not contain template text")
	}
}

func TestBuildPrompt_order(t *testing.T) {
	r := makeReviewer(t, "security", "FOCUS")
	diff := "DIFF"

	result := r.buildPrompt(diff)

	templateIdx := strings.Index(result, "TEMPLATE")
	focusIdx := strings.Index(result, "FOCUS")
	diffIdx := strings.Index(result, "DIFF")

	if !(templateIdx < focusIdx && focusIdx < diffIdx) {
		t.Errorf("expected order: template text, focus, diff; got indices %d %d %d", templateIdx, focusIdx, diffIdx)
	}
}

func TestReview_issues(t *testing.T) {
	r := makeReviewer(t, "security", "check security")
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
}

func TestReview_empty(t *testing.T) {
	r := makeReviewer(t, "security", "check security")
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
	r := makeReviewer(t, "security", "check security")
	a := &fakeAssistant{fn: func(string) *exec.Cmd { return echoCmd("not json") }}

	_, err := r.review("diff", a)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "unmarshal issues") {
		t.Errorf("error should mention unmarshal issues, got: %v", err)
	}
}

func TestReview(t *testing.T) {
	sec := makeReviewer(t, "security", "check security")
	tst := makeReviewer(t, "tests", "check tests")

	a := &fakeAssistant{fn: func(string) *exec.Cmd {
		return echoCmd(`[{"severity":"HIGH","title":"issue","location":"f.go:1","description":"bad"}]`)
	}}

	issues, errs := Review([]Reviewer{sec, tst}, "some diff", a)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(issues) != 2 {
		t.Fatalf("got %d issues, want 2", len(issues))
	}

	names := map[string]bool{}
	for _, f := range issues {
		names[f.Reviewer] = true
	}
	if !names["security"] || !names["tests"] {
		t.Errorf("expected issues from both reviewers, got: %v", names)
	}
}

func TestReview_promptError(t *testing.T) {
	sec := makeReviewer(t, "security", "check security")
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
	ok := makeReviewer(t, "ok", "ok focus")
	bad := makeReviewer(t, "bad", "bad focus")

	a := &fakeAssistant{fn: func(p string) *exec.Cmd {
		if strings.Contains(p, "bad focus") {
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

func writeTempPrompt(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "reviewer-*.md")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(f.Name()) })
	f.WriteString(content)
	f.Close()
	return f.Name()
}

func TestResolve(t *testing.T) {
	p1 := writeTempPrompt(t, "focus 1")
	p2 := writeTempPrompt(t, "focus 2")
	tmpl := prompt.New(syntheticTemplate)

	reviewers, err := resolve(p1+","+p2, tmpl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reviewers) != 2 {
		t.Fatalf("got %d reviewers, want 2", len(reviewers))
	}
	if reviewers[0].Prompt.String() != "focus 1" {
		t.Errorf("got prompt %q, want %q", reviewers[0].Prompt.String(), "focus 1")
	}
	if reviewers[1].Prompt.String() != "focus 2" {
		t.Errorf("got prompt %q, want %q", reviewers[1].Prompt.String(), "focus 2")
	}
}

func TestResolve_empty(t *testing.T) {
	_, err := resolve("", prompt.New("tmpl"))
	if err == nil {
		t.Fatal("expected error for empty paths")
	}
}

func TestResolve_unknown(t *testing.T) {
	_, err := resolve("bogus-nonexistent-file.md", prompt.New("tmpl"))
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
	if !strings.Contains(err.Error(), "bogus-nonexistent-file.md") {
		t.Errorf("error should include the path, got: %v", err)
	}
}

func TestResolve_whitespace(t *testing.T) {
	p1 := writeTempPrompt(t, "focus 1")
	p2 := writeTempPrompt(t, "focus 2")
	tmpl := prompt.New(syntheticTemplate)

	reviewers, err := resolve(p1+",  "+p2, tmpl)
	if err != nil {
		t.Fatalf("unexpected error with whitespace: %v", err)
	}
	if len(reviewers) != 2 {
		t.Fatalf("got %d reviewers, want 2", len(reviewers))
	}
}

func TestResolve_emptyToken(t *testing.T) {
	_, err := resolve(",", prompt.New("tmpl"))
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestResolve_nameFromPath(t *testing.T) {
	p := writeTempPrompt(t, "focus")
	tmpl := prompt.New(syntheticTemplate)

	reviewers, err := resolve(p, tmpl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.HasSuffix(reviewers[0].Name, ".md") {
		t.Errorf("name should not include extension, got %q", reviewers[0].Name)
	}
}
