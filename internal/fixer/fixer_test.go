package fixer

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/pvinchon/agent/internal/reviewer"
)

var testIssues = []reviewer.Issue{
	{Reviewer: "security", Severity: "HIGH", Title: "SQL injection", Location: "db.go:10", Description: "User input used directly in query."},
	{Reviewer: "go", Severity: "LOW", Title: "Unused variable", Location: "main.go:5", Description: "Variable x is never used."},
}

func TestBuildPrompt(t *testing.T) {
	diff := "some diff content"
	prompt := buildPrompt(testIssues, diff)

	if !strings.Contains(prompt, "You are a senior engineer") {
		t.Error("prompt does not contain base template text")
	}
	if !strings.Contains(prompt, diff) {
		t.Error("prompt does not contain the diff")
	}
	if !strings.Contains(prompt, "SQL injection") {
		t.Error("prompt does not contain issue title")
	}
	if !strings.Contains(prompt, "db.go:10") {
		t.Error("prompt does not contain issue location")
	}
}

func TestBuildPrompt_order(t *testing.T) {
	diff := "some diff"
	prompt := buildPrompt(testIssues, diff)

	issuesIdx := strings.Index(prompt, "SQL injection")
	diffIdx := strings.Index(prompt, diff)

	if !(issuesIdx < diffIdx) {
		t.Error("expected issues before diff in prompt")
	}
}

func TestFormatIssues(t *testing.T) {
	got := formatIssues(testIssues)

	if !strings.Contains(got, "1. [HIGH] SQL injection") {
		t.Errorf("missing first issue, got:\n%s", got)
	}
	if !strings.Contains(got, "2. [LOW] Unused variable") {
		t.Errorf("missing second issue, got:\n%s", got)
	}
	if !strings.Contains(got, "Location: db.go:10") {
		t.Errorf("missing location, got:\n%s", got)
	}
	if !strings.Contains(got, "User input used directly in query.") {
		t.Errorf("missing description, got:\n%s", got)
	}
}

func TestFormatIssues_empty(t *testing.T) {
	got := formatIssues(nil)
	if got != "" {
		t.Errorf("expected empty string for no issues, got %q", got)
	}
}

func TestFix_success(t *testing.T) {
	a := &fakeAssistant{fn: func(string) *exec.Cmd { return echoCmd("ok") }}
	if err := Fix(testIssues, "diff", a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFix_promptContainsIssuesAndDiff(t *testing.T) {
	var capturedPrompt string
	a := &fakeAssistant{fn: func(prompt string) *exec.Cmd {
		capturedPrompt = prompt
		return echoCmd("")
	}}

	diff := "my diff"
	if err := Fix(testIssues, diff, a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(capturedPrompt, "SQL injection") {
		t.Error("prompt sent to assistant does not contain issue title")
	}
	if !strings.Contains(capturedPrompt, diff) {
		t.Error("prompt sent to assistant does not contain diff")
	}
}

func TestFix_error(t *testing.T) {
	a := &fakeAssistant{fn: func(string) *exec.Cmd { return failCmd() }}
	err := Fix(testIssues, "diff", a)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "fixer") {
		t.Errorf("error should mention fixer, got: %v", err)
	}
}

type fakeAssistant struct {
	fn func(string) *exec.Cmd
}

func (f *fakeAssistant) Command(prompt string) *exec.Cmd { return f.fn(prompt) }

func echoCmd(output string) *exec.Cmd { return exec.Command("echo", output) }
func failCmd() *exec.Cmd              { return exec.Command("false") }
