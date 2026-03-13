package main

import (
"flag"
"os"
"testing"
)

func TestReviewFlags(t *testing.T) {
fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
mustReviewers, reviewTemplate, mustAssistant, resolveLog := reviewFlags(fs)
fs.Parse([]string{"--reviewers=security", "--assistant=claude"})

if mustReviewers == nil {
t.Fatal("expected non-nil mustReviewers")
}
if reviewTemplate == nil {
t.Fatal("expected non-nil reviewTemplate")
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
if fs.Lookup("review-template") == nil {
t.Error("expected --review-template flag to be registered")
}
if fs.Lookup("prompts-dir") == nil {
t.Error("expected --prompts-dir flag to be registered")
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
mustAssistant, resolveLog, fixTemplate := fixFlags(fs)
fs.Parse([]string{"--assistant=copilot"})

if mustAssistant == nil {
t.Fatal("expected non-nil mustAssistant")
}
if resolveLog == nil {
t.Fatal("expected non-nil resolveLog")
}
if fixTemplate == nil {
t.Fatal("expected non-nil fixTemplate")
}
if fs.Lookup("reviewers") != nil {
t.Error("--reviewers should not be registered for fix")
}
if fs.Lookup("assistant") == nil {
t.Error("expected --assistant flag to be registered")
}
if fs.Lookup("fix-template") == nil {
t.Error("expected --fix-template flag to be registered")
}
if fs.Lookup("verbose") == nil {
t.Error("expected --verbose flag to be registered")
}
}

func TestLoopFlags(t *testing.T) {
fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
mustReviewers, mustAssistant, resolveLog, maxAttempts, reviewTemplate, fixTemplate := loopFlags(fs)
fs.Parse([]string{"--reviewers=security", "--assistant=claude", "--max-attempts=3"})

if mustReviewers == nil {
t.Fatal("expected non-nil mustReviewers")
}
if mustAssistant == nil {
t.Fatal("expected non-nil mustAssistant")
}
if resolveLog == nil {
t.Fatal("expected non-nil resolveLog")
}
if reviewTemplate == nil {
t.Fatal("expected non-nil reviewTemplate")
}
if fixTemplate == nil {
t.Fatal("expected non-nil fixTemplate")
}
if *maxAttempts != 3 {
t.Errorf("got max-attempts=%d, want 3", *maxAttempts)
}
if fs.Lookup("max-attempts") == nil {
t.Error("expected --max-attempts flag to be registered")
}
if fs.Lookup("review-template") == nil {
t.Error("expected --review-template flag to be registered")
}
if fs.Lookup("fix-template") == nil {
t.Error("expected --fix-template flag to be registered")
}
if fs.Lookup("prompts-dir") == nil {
t.Error("expected --prompts-dir flag to be registered")
}
}
