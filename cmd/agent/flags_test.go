package main

import (
	"flag"
	"os"
	"testing"
)

func TestReviewFlags(t *testing.T) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	mustReviewers, mustAssistant, resolveLog := reviewFlags(fs)
	fs.Parse([]string{"--reviewers=security", "--assistant=claude"})

	if mustReviewers == nil {
		t.Fatal("expected non-nil mustReviewers")
	}
	if mustAssistant == nil {
		t.Fatal("expected non-nil mustAssistant")
	}
	if resolveLog == nil {
		t.Fatal("expected non-nil resolveLog")
	}
	if fs.Lookup("reviewers") == nil {
		t.Error("expected --reviewers flag to be registered")
	}
	if fs.Lookup("assistant") == nil {
		t.Error("expected --assistant flag to be registered")
	}
	if fs.Lookup("verbose") == nil {
		t.Error("expected --verbose flag to be registered")
	}
}

func TestFixFlags(t *testing.T) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	mustAssistant, resolveLog := fixFlags(fs)
	fs.Parse([]string{"--assistant=copilot"})

	if mustAssistant == nil {
		t.Fatal("expected non-nil mustAssistant")
	}
	if resolveLog == nil {
		t.Fatal("expected non-nil resolveLog")
	}
	if fs.Lookup("reviewers") != nil {
		t.Error("--reviewers should not be registered for fix")
	}
	if fs.Lookup("assistant") == nil {
		t.Error("expected --assistant flag to be registered")
	}
	if fs.Lookup("verbose") == nil {
		t.Error("expected --verbose flag to be registered")
	}
}

func TestLoopFlags(t *testing.T) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	mustReviewers, mustReviewAssistant, mustFixAssistant, resolveLog, maxAttempts := loopFlags(fs)
	fs.Parse([]string{"--reviewers=security", "--assistant-for-review=claude", "--assistant-for-fix=copilot", "--max-attempts=3"})

	if mustReviewers == nil {
		t.Fatal("expected non-nil mustReviewers")
	}
	if mustReviewAssistant == nil {
		t.Fatal("expected non-nil mustReviewAssistant")
	}
	if mustFixAssistant == nil {
		t.Fatal("expected non-nil mustFixAssistant")
	}
	if resolveLog == nil {
		t.Fatal("expected non-nil resolveLog")
	}
	if *maxAttempts != 3 {
		t.Errorf("got max-attempts=%d, want 3", *maxAttempts)
	}
	if fs.Lookup("max-attempts") == nil {
		t.Error("expected --max-attempts flag to be registered")
	}
	if fs.Lookup("assistant-for-review") == nil {
		t.Error("expected --assistant-for-review flag to be registered")
	}
	if fs.Lookup("assistant-for-fix") == nil {
		t.Error("expected --assistant-for-fix flag to be registered")
	}
	if fs.Lookup("assistant") != nil {
		t.Error("--assistant should not be registered for loop")
	}
}
