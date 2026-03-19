package reviewer

import (
	"flag"
	"os"
	"testing"
)

func writeTempFlagPrompt(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "flag-prompt-*.md")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(f.Name()) })
	f.WriteString(content)
	f.Close()
	return f.Name()
}

func TestFlagSet(t *testing.T) {
	reviewerPath := writeTempFlagPrompt(t, "focus content")
	templatePath := writeTempFlagPrompt(t, "TEMPLATE {{prompt}} {{diff}} END")

	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	mustReviewers, mustTemplate := FlagSet(fs)
	fs.Parse([]string{
		"--reviewers=" + reviewerPath,
		"--reviewer-template=" + templatePath,
	})

	tmpl := mustTemplate()
	if tmpl.String() != "TEMPLATE {{prompt}} {{diff}} END" {
		t.Errorf("unexpected template content: %q", tmpl.String())
	}

	reviewers := mustReviewers()
	if len(reviewers) != 1 {
		t.Fatalf("got %d reviewers, want 1", len(reviewers))
	}
	if reviewers[0].Prompt.String() != "focus content" {
		t.Errorf("unexpected reviewer prompt: %q", reviewers[0].Prompt.String())
	}
	if reviewers[0].Template.String() != "TEMPLATE {{prompt}} {{diff}} END" {
		t.Errorf("unexpected reviewer template: %q", reviewers[0].Template.String())
	}
}

func TestFlagSet_registeredFlags(t *testing.T) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	FlagSet(fs)

	if fs.Lookup("reviewers") == nil {
		t.Error("expected --reviewers flag to be registered")
	}
	if fs.Lookup("reviewer-template") == nil {
		t.Error("expected --reviewer-template flag to be registered")
	}
}
