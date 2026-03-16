package assistant

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeFakeCLI creates a fake executable named name that prints output and
// prepends its directory to PATH for the duration of the test.
func makeFakeCLI(t *testing.T, name, output string) {
	t.Helper()
	dir := t.TempDir()
	script := filepath.Join(dir, name)
	src := "#!/bin/sh\nprintf '%s' '" + output + "'\n"
	if err := os.WriteFile(script, []byte(src), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// makeFakeCLIWithModels creates a fake executable that:
//   - responds to the "models" subcommand by printing the given models, one per line
//   - prints promptOutput for any other invocation
func makeFakeCLIWithModels(t *testing.T, name string, models []string, promptOutput string) {
	t.Helper()
	dir := t.TempDir()
	script := filepath.Join(dir, name)

	// Build the models output as a series of echo calls to avoid quoting issues.
	var modelLines strings.Builder
	for _, m := range models {
		fmt.Fprintf(&modelLines, "echo %q\n", m)
	}

	src := "#!/bin/sh\n" +
		"if [ \"$1\" = \"models\" ]; then\n" +
		modelLines.String() +
		"else\n" +
		fmt.Sprintf("printf '%%s' %q\n", promptOutput) +
		"fi\n"

	if err := os.WriteFile(script, []byte(src), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestNew_ValidNames(t *testing.T) {
	for _, name := range []string{"claude", "copilot"} {
		a, err := New(name, "")
		if err != nil {
			t.Errorf("New(%q, \"\") unexpected error: %v", name, err)
		}
		if a == nil {
			t.Errorf("New(%q, \"\") returned nil", name)
		}
	}
}

func TestNew_InvalidName(t *testing.T) {
	_, err := New("unknown", "")
	if err == nil {
		t.Error("New(\"unknown\", \"\") expected error, got nil")
	}
}

func TestNew_ValidModel(t *testing.T) {
	makeFakeCLIWithModels(t, "claude", claudeTestModels, "")
	a, err := New("claude", "test-claude-sonnet")
	if err != nil {
		t.Fatalf("New(\"claude\", \"test-claude-sonnet\") unexpected error: %v", err)
	}
	c, ok := a.(*Claude)
	if !ok {
		t.Fatalf("expected *Claude, got %T", a)
	}
	if c.Model != "test-claude-sonnet" {
		t.Errorf("Model = %q, want %q", c.Model, "test-claude-sonnet")
	}
}

func TestNew_InvalidModel(t *testing.T) {
	makeFakeCLIWithModels(t, "claude", claudeTestModels, "")
	_, err := New("claude", "test-gpt-large")
	if err == nil {
		t.Error("New(\"claude\", \"test-gpt-large\") expected error, got nil")
	}
}

func TestNew_ModelNotAvailableForAssistant(t *testing.T) {
	makeFakeCLIWithModels(t, "copilot", copilotTestModels, "")
	_, err := New("copilot", "test-claude-opus")
	if err == nil {
		t.Error("New(\"copilot\", \"test-claude-opus\") expected error: test-claude-opus is a Claude-only model")
	}
}

func TestModels(t *testing.T) {
	makeFakeCLIWithModels(t, "claude", claudeTestModels, "")
	got, err := Models(&Claude{})
	if err != nil {
		t.Fatalf("Models() error: %v", err)
	}
	if len(got) != len(claudeTestModels) {
		t.Fatalf("Models() returned %d models, want %d", len(got), len(claudeTestModels))
	}
	for i, want := range claudeTestModels {
		if got[i] != want {
			t.Errorf("Models()[%d] = %q, want %q", i, got[i], want)
		}
	}
}

func TestModels_CommandNotFound(t *testing.T) {
	t.Setenv("PATH", "")
	_, err := Models(&Claude{})
	if err == nil {
		t.Error("Models() expected error when binary not in PATH, got nil")
	}
}

func TestPrompt(t *testing.T) {
	makeFakeCLI(t, "claude", "hello from claude")
	got, err := Prompt(&Claude{}, "test input")
	if err != nil {
		t.Fatalf("Prompt() error: %v", err)
	}
	if got != "hello from claude" {
		t.Errorf("Prompt() = %q, want %q", got, "hello from claude")
	}
}

func TestPrompt_TrimsWhitespace(t *testing.T) {
	makeFakeCLI(t, "claude", "  trimmed  ")
	got, err := Prompt(&Claude{}, "test input")
	if err != nil {
		t.Fatalf("Prompt() error: %v", err)
	}
	if got != "trimmed" {
		t.Errorf("Prompt() = %q, want %q", got, "trimmed")
	}
}

func TestPrompt_CommandNotFound(t *testing.T) {
	t.Setenv("PATH", "")
	_, err := Prompt(&Claude{}, "test input")
	if err == nil {
		t.Error("Prompt() expected error when binary not in PATH, got nil")
	}
}

// claudeTestModels and copilotTestModels are small fixed sets used only in tests
// to stand in for what the real CLIs would return.
var claudeTestModels = []string{"test-claude-haiku", "test-claude-sonnet", "test-claude-opus"}
var copilotTestModels = []string{"test-gpt-small", "test-gpt-large", "test-o-mini"}
