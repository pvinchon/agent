package reviewer

import (
"os"
"os/exec"
"path/filepath"
"strings"
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

prompt := r.buildPrompt(diff, "")

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
prompt := r.buildPrompt(diff, "")

baseIdx := strings.Index(prompt, "You are a senior reviewer")
reviewerIdx := strings.Index(prompt, r.Prompt)
diffIdx := strings.Index(prompt, diff)

if !(baseIdx < reviewerIdx && reviewerIdx < diffIdx) {
t.Error("expected order: base template text, reviewer prompt, diff")
}
}

func TestBuildPrompt_customTemplate(t *testing.T) {
r, _ := New("security")
diff := "my diff"
custom := "Custom context: {{prompt}}\nDiff: {{diff}}"

prompt := r.buildPrompt(diff, custom)

if !strings.Contains(prompt, r.Prompt) {
t.Error("prompt does not contain reviewer-specific prompt")
}
if !strings.Contains(prompt, diff) {
t.Error("prompt does not contain the diff")
}
// Output spec must always be appended
if !strings.Contains(prompt, "Return valid JSON only") {
t.Error("prompt does not contain output specification")
}
// Custom template replaces the built-in scene-setting; must not contain
// the built-in preamble.
if strings.Contains(prompt, "You are a senior reviewer") {
t.Error("prompt should not contain built-in preamble when custom template is used")
}
}

func TestBuildPrompt_customTemplate_outputAppended(t *testing.T) {
r, _ := New("security")
diff := "d"
custom := "Scene: {{prompt}} Diff: {{diff}}"

prompt := r.buildPrompt(diff, custom)

customIdx := strings.Index(prompt, "Scene:")
outputIdx := strings.Index(prompt, "Return valid JSON only")

if customIdx < 0 || outputIdx < 0 {
t.Fatalf("missing expected sections in prompt")
}
if customIdx > outputIdx {
t.Error("expected custom context before output spec")
}
}

func TestReview_issues(t *testing.T) {
r := Reviewer{Name: "security", Prompt: "check security"}
a := &fakeAssistant{fn: func(string) *exec.Cmd {
return echoCmd(`[{"severity":"HIGH","title":"SQL injection","location":"db.go:10","description":"User input in query."}]`)
}}

issues, err := r.review("diff", "", a)
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

issues, err := r.review("diff", "", a)
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

_, err := r.review("diff", "", a)
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

func (f *fakeAssistant) Command(prompt string) *exec.Cmd {
return f.fn(prompt)
}

func echoCmd(output string) *exec.Cmd { return exec.Command("echo", output) }
func failCmd() *exec.Cmd              { return exec.Command("false") }

func TestReview(t *testing.T) {
sec, _ := New("security")
tests, _ := New("tests")

a := &fakeAssistant{fn: func(string) *exec.Cmd {
return echoCmd(`[{"severity":"HIGH","title":"issue","location":"f.go:1","description":"bad"}]`)
}}

issues, errs := Review([]Reviewer{sec, tests}, "some diff", "", a)
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
sec, _ := New("security")
a := &fakeAssistant{fn: func(string) *exec.Cmd { return failCmd() }}

issues, errs := Review([]Reviewer{sec}, "diff", "", a)
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

issues, errs := Review([]Reviewer{ok, bad}, "diff", "", a)
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

func TestLoadPromptsDir(t *testing.T) {
dir := t.TempDir()
if err := os.WriteFile(filepath.Join(dir, "python.md"), []byte("review python code"), 0o600); err != nil {
t.Fatal(err)
}

m, err := LoadPromptsDir(dir)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if len(m) != 1 {
t.Fatalf("got %d reviewers, want 1", len(m))
}
r, ok := m["python"]
if !ok {
t.Fatal("expected reviewer named 'python'")
}
if r.Prompt != "review python code" {
t.Errorf("got prompt %q, want %q", r.Prompt, "review python code")
}
}

func TestLoadPromptsDir_ignoresNonMD(t *testing.T) {
dir := t.TempDir()
os.WriteFile(filepath.Join(dir, "python.md"), []byte("python"), 0o600)
os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("ignore me"), 0o600)

m, err := LoadPromptsDir(dir)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if len(m) != 1 {
t.Errorf("got %d reviewers, want 1 (non-.md files should be ignored)", len(m))
}
}

func TestLoadPromptsDir_notFound(t *testing.T) {
_, err := LoadPromptsDir("/nonexistent/dir")
if err == nil {
t.Fatal("expected error for nonexistent directory")
}
}

func TestResolveFrom_customRegistry(t *testing.T) {
registry := map[string]Reviewer{
"custom": {Name: "custom", Prompt: "custom prompt"},
}
reviewers, err := resolveFrom(registry, "custom")
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if len(reviewers) != 1 {
t.Fatalf("got %d reviewers, want 1", len(reviewers))
}
if reviewers[0].Name != "custom" {
t.Errorf("got name %q, want %q", reviewers[0].Name, "custom")
}
}

func TestResolveFrom_unknownInCustomRegistry(t *testing.T) {
registry := map[string]Reviewer{
"custom": {Name: "custom", Prompt: "custom"},
}
_, err := resolveFrom(registry, "custom,bogus")
if err == nil {
t.Fatal("expected error for unknown reviewer")
}
if !strings.Contains(err.Error(), "bogus") {
t.Errorf("error should mention unknown name, got: %v", err)
}
}
