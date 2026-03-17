package reviewer

import (
	"flag"
	"os"
	"testing"
)

func TestFlagSet(t *testing.T) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	mustReviewers := FlagSet(fs)
	fs.Parse([]string{"--reviewers=security,tests"})

	reviewers := mustReviewers()
	if len(reviewers) != 2 {
		t.Fatalf("got %d reviewers, want 2", len(reviewers))
	}
	if reviewers[0].Slug != "security" {
		t.Errorf("got %q, want %q", reviewers[0].Slug, "security")
	}
	if reviewers[1].Slug != "tests" {
		t.Errorf("got %q, want %q", reviewers[1].Slug, "tests")
	}
}
