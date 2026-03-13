package reviewer

import (
"flag"
"os"
"path/filepath"
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
if reviewers[0].Name != "security" {
t.Errorf("got %q, want %q", reviewers[0].Name, "security")
}
if reviewers[1].Name != "tests" {
t.Errorf("got %q, want %q", reviewers[1].Name, "tests")
}
}

func TestFlagSet_promptsDir(t *testing.T) {
dir := t.TempDir()
if err := os.WriteFile(filepath.Join(dir, "python.md"), []byte("review python"), 0o600); err != nil {
t.Fatal(err)
}

fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
mustReviewers := FlagSet(fs)
fs.Parse([]string{"--reviewers=security,python", "--prompts-dir=" + dir})

reviewers := mustReviewers()
if len(reviewers) != 2 {
t.Fatalf("got %d reviewers, want 2", len(reviewers))
}
names := map[string]bool{}
for _, r := range reviewers {
names[r.Name] = true
}
if !names["security"] {
t.Error("expected built-in 'security' reviewer")
}
if !names["python"] {
t.Error("expected custom 'python' reviewer from prompts-dir")
}
}

func TestFlagSet_promptsDir_override(t *testing.T) {
dir := t.TempDir()
customPrompt := "my custom security prompt"
if err := os.WriteFile(filepath.Join(dir, "security.md"), []byte(customPrompt), 0o600); err != nil {
t.Fatal(err)
}

fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
mustReviewers := FlagSet(fs)
fs.Parse([]string{"--reviewers=security", "--prompts-dir=" + dir})

reviewers := mustReviewers()
if len(reviewers) != 1 {
t.Fatalf("got %d reviewers, want 1", len(reviewers))
}
if reviewers[0].Prompt != customPrompt {
t.Errorf("got prompt %q, want %q", reviewers[0].Prompt, customPrompt)
}
}

func TestTemplateFlagSet_empty(t *testing.T) {
fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
reviewTemplate := TemplateFlagSet(fs)
fs.Parse([]string{})

got := reviewTemplate()
if got != "" {
t.Errorf("expected empty string when no --review-template flag, got %q", got)
}
}

func TestTemplateFlagSet_file(t *testing.T) {
dir := t.TempDir()
content := "Custom template: {{prompt}} and {{diff}}"
path := filepath.Join(dir, "template.md")
if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
t.Fatal(err)
}

fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
reviewTemplate := TemplateFlagSet(fs)
fs.Parse([]string{"--review-template=" + path})

got := reviewTemplate()
if got != content {
t.Errorf("got %q, want %q", got, content)
}
}
