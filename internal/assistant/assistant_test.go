package assistant

import (
	"os"
	"path/filepath"
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
	a, err := New("claude", "claude-sonnet-4-5")
	if err != nil {
		t.Fatalf("New(\"claude\", \"claude-sonnet-4-5\") unexpected error: %v", err)
	}
	c, ok := a.(*Claude)
	if !ok {
		t.Fatalf("expected *Claude, got %T", a)
	}
	if c.Model != "claude-sonnet-4-5" {
		t.Errorf("Model = %q, want %q", c.Model, "claude-sonnet-4-5")
	}
}

func TestNew_InvalidModel(t *testing.T) {
	_, err := New("claude", "gpt-4o")
	if err == nil {
		t.Error("New(\"claude\", \"gpt-4o\") expected error, got nil")
	}
}

func TestNew_ModelNotAvailableForAssistant(t *testing.T) {
	_, err := New("copilot", "claude-opus-4-5")
	if err == nil {
		t.Error("New(\"copilot\", \"claude-opus-4-5\") expected error: claude-opus-4-5 is a Claude-only model")
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
