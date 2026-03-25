package reviewer

import (
	"flag"
	"os"
	"testing"
)

func TestFlagSet(t *testing.T) {
	secPath := tempPrompt(t, "security.md", "security prompt")
	testsPath := tempPrompt(t, "tests.md", "tests prompt")

	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	mustReviewers := FlagSet(fs)
	fs.Parse([]string{"--reviewers=" + secPath + "," + testsPath})

	reviewers := mustReviewers()
	if len(reviewers) != 2 {
		t.Fatalf("got %d reviewers, want 2", len(reviewers))
	}
	if reviewers[0].Path != secPath {
		t.Errorf("got %q, want %q", reviewers[0].Path, secPath)
	}
	if reviewers[1].Path != testsPath {
		t.Errorf("got %q, want %q", reviewers[1].Path, testsPath)
	}
}
