package reviewer

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func tempPrompt(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func swapHTTPClient(t *testing.T, c *http.Client) {
	t.Helper()
	old := httpClient
	httpClient = c
	t.Cleanup(func() { httpClient = old })
}

func TestNewReviewer(t *testing.T) {
	path := tempPrompt(t, "security.md", "check for vulnerabilities")
	r, err := New(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Path != path {
		t.Errorf("got path %q, want %q", r.Path, path)
	}
	if r.Prompt == "" {
		t.Error("prompt is empty")
	}
}

func TestNew_missing(t *testing.T) {
	_, err := New("/nonexistent/path/to/prompt.md")
	if err == nil {
		t.Fatal("expected error for missing prompt file")
	}
}

func TestBuildPrompt(t *testing.T) {
	path := tempPrompt(t, "security.md", "check for vulnerabilities")
	r, err := New(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	diff := `--- a/main.go
+++ b/main.go
@@ -1 +1 @@
-fmt.Println("hello")
+fmt.Println(userInput)`

	prompt := r.buildPrompt(diff)

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
	path := tempPrompt(t, "security.md", "check for vulnerabilities")
	r, _ := New(path)
	diff := "some diff"
	prompt := r.buildPrompt(diff)

	baseIdx := strings.Index(prompt, "You are a senior reviewer")
	reviewerIdx := strings.Index(prompt, r.Prompt)
	diffIdx := strings.Index(prompt, diff)

	if !(baseIdx < reviewerIdx && reviewerIdx < diffIdx) {
		t.Error("expected order: base template text, reviewer prompt, diff")
	}
}

func TestReview_issues(t *testing.T) {
	r := Reviewer{Path: "security", Prompt: "check security"}
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
	if issues[0].Title != "SQL injection" {
		t.Errorf("got title %q, want %q", issues[0].Title, "SQL injection")
	}
}

func TestReview_empty(t *testing.T) {
	r := Reviewer{Path: "security", Prompt: "check security"}
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
	r := Reviewer{Path: "security", Prompt: "check security"}
	a := &fakeAssistant{fn: func(string) *exec.Cmd { return echoCmd("not json") }}

	_, err := r.review("diff", a)
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

func (f *fakeAssistant) Command(prompt string) *exec.Cmd { return f.fn(prompt) }

func echoCmd(output string) *exec.Cmd { return exec.Command("echo", output) }
func failCmd() *exec.Cmd              { return exec.Command("false") }

func TestReview(t *testing.T) {
	sec := Reviewer{Path: "security", Prompt: "check security"}
	tests := Reviewer{Path: "tests", Prompt: "check tests"}

	a := &fakeAssistant{fn: func(string) *exec.Cmd {
		return echoCmd(`[{"severity":"HIGH","title":"issue","location":"f.go:1","description":"bad"}]`)
	}}

	issues, errs := Review([]Reviewer{sec, tests}, "some diff", a)
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
	sec := Reviewer{Path: "security", Prompt: "check security"}
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
	ok := Reviewer{Path: "ok", Prompt: "ok"}
	bad := Reviewer{Path: "bad", Prompt: "bad"}

	a := &fakeAssistant{fn: func(prompt string) *exec.Cmd {
		if strings.Contains(prompt, "bad") {
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

func TestResolve(t *testing.T) {
	secPath := tempPrompt(t, "security.md", "sec")
	testsPath := tempPrompt(t, "tests.md", "tests")

	reviewers, err := resolve(secPath + "," + testsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
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

func TestResolve_empty(t *testing.T) {
	_, err := resolve("")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestResolve_missing(t *testing.T) {
	path := tempPrompt(t, "security.md", "sec")
	_, err := resolve(path + ",/nonexistent/bogus.md")
	if err == nil {
		t.Fatal("expected error for missing prompt file")
	}
	if !strings.Contains(err.Error(), "bogus") {
		t.Errorf("error should mention the missing path, got: %v", err)
	}
}

func TestResolve_whitespace(t *testing.T) {
	secPath := tempPrompt(t, "security.md", "sec")
	testsPath := tempPrompt(t, "tests.md", "tests")

	reviewers, err := resolve(secPath + ", " + testsPath)
	if err != nil {
		t.Fatalf("unexpected error with whitespace: %v", err)
	}
	if len(reviewers) != 2 {
		t.Fatalf("got %d reviewers, want 2", len(reviewers))
	}
}

func TestFetchPrompt(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("remote prompt content"))
	}))
	defer srv.Close()

	swapHTTPClient(t, srv.Client())

	content, err := fetchPrompt(srv.URL + "/prompt/security.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "remote prompt content" {
		t.Errorf("got %q, want %q", content, "remote prompt content")
	}
}

func TestFetchPrompt_notFound(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	swapHTTPClient(t, srv.Client())

	_, err := fetchPrompt(srv.URL + "/missing.md")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error should mention status code, got: %v", err)
	}
}

func TestNew_remoteURL(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("remote prompt"))
	}))
	defer srv.Close()

	swapHTTPClient(t, srv.Client())

	url := srv.URL + "/prompt/security.md"
	r, err := New(url)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Path != url {
		t.Errorf("got path %q, want %q", r.Path, url)
	}
	if r.Prompt != "remote prompt" {
		t.Errorf("got prompt %q, want %q", r.Prompt, "remote prompt")
	}
}

func TestReadPrompt_local(t *testing.T) {
	path := tempPrompt(t, "test.md", "local content")
	content, err := readPrompt(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "local content" {
		t.Errorf("got %q, want %q", content, "local content")
	}
}
