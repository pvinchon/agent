package main

import (
	"flag"
	"os"
	"testing"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "flags-test-*.md")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(f.Name()) })
	f.WriteString(content)
	f.Close()
	return f.Name()
}

func TestReviewFlags(t *testing.T) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	mustReviewers, mustTemplate, mustAssistant, resolveLog := reviewFlags(fs)

	if mustReviewers == nil {
		t.Fatal("expected non-nil mustReviewers")
	}
	if mustTemplate == nil {
		t.Fatal("expected non-nil mustTemplate")
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
	if fs.Lookup("reviewer-template") == nil {
		t.Error("expected --reviewer-template flag to be registered")
	}
	if fs.Lookup("assistant") == nil {
		t.Error("expected --assistant flag to be registered")
	}
	if fs.Lookup("model") == nil {
		t.Error("expected --model flag to be registered")
	}
	if fs.Lookup("verbose") == nil {
		t.Error("expected --verbose flag to be registered")
	}
}

func TestFixFlags(t *testing.T) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	mustAssistant, mustTemplate, resolveLog := fixFlags(fs)

	if mustAssistant == nil {
		t.Fatal("expected non-nil mustAssistant")
	}
	if mustTemplate == nil {
		t.Fatal("expected non-nil mustTemplate")
	}
	if resolveLog == nil {
		t.Fatal("expected non-nil resolveLog")
	}
	if fs.Lookup("reviewers") != nil {
		t.Error("--reviewers should not be registered for fix")
	}
	if fs.Lookup("fixer-template") == nil {
		t.Error("expected --fixer-template flag to be registered")
	}
	if fs.Lookup("assistant") == nil {
		t.Error("expected --assistant flag to be registered")
	}
	if fs.Lookup("model") == nil {
		t.Error("expected --model flag to be registered")
	}
	if fs.Lookup("verbose") == nil {
		t.Error("expected --verbose flag to be registered")
	}
}

func TestLoopFlags(t *testing.T) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	mustReviewers, mustReviewerTemplate, mustReviewAssistant, mustFixAssistant, mustFixerTemplate, resolveLog, maxAttempts := loopFlags(fs)

	if mustReviewers == nil {
		t.Fatal("expected non-nil mustReviewers")
	}
	if mustReviewerTemplate == nil {
		t.Fatal("expected non-nil mustReviewerTemplate")
	}
	if mustReviewAssistant == nil {
		t.Fatal("expected non-nil mustReviewAssistant")
	}
	if mustFixAssistant == nil {
		t.Fatal("expected non-nil mustFixAssistant")
	}
	if mustFixerTemplate == nil {
		t.Fatal("expected non-nil mustFixerTemplate")
	}
	if resolveLog == nil {
		t.Fatal("expected non-nil resolveLog")
	}
	if maxAttempts == nil {
		t.Fatal("expected non-nil maxAttempts")
	}

	fs.Parse([]string{"--max-attempts=3"})
	if *maxAttempts != 3 {
		t.Errorf("got max-attempts=%d, want 3", *maxAttempts)
	}

	if fs.Lookup("max-attempts") == nil {
		t.Error("expected --max-attempts flag to be registered")
	}
	if fs.Lookup("reviewer-template") == nil {
		t.Error("expected --reviewer-template flag to be registered")
	}
	if fs.Lookup("fixer-template") == nil {
		t.Error("expected --fixer-template flag to be registered")
	}
	if fs.Lookup("assistant-for-review") == nil {
		t.Error("expected --assistant-for-review flag to be registered")
	}
	if fs.Lookup("assistant-for-fix") == nil {
		t.Error("expected --assistant-for-fix flag to be registered")
	}
	if fs.Lookup("model-for-review") == nil {
		t.Error("expected --model-for-review flag to be registered")
	}
	if fs.Lookup("model-for-fix") == nil {
		t.Error("expected --model-for-fix flag to be registered")
	}
	if fs.Lookup("assistant") != nil {
		t.Error("--assistant should not be registered for loop")
	}
}
