package assistant

import (
	"os"
	"path/filepath"
	"slices"
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

func TestNew_WithModel(t *testing.T) {
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

func TestClaudeCommand_noModel(t *testing.T) {
	c := &Claude{}
	cmd := c.Command("test prompt")
	want := []string{"claude", "--dangerously-skip-permissions", "--print", "test prompt"}
	if !slices.Equal(cmd.Args, want) {
		t.Errorf("Command() args = %v, want %v", cmd.Args, want)
	}
}

func TestClaudeCommand_withModel(t *testing.T) {
	c := &Claude{Model: "claude-opus"}
	cmd := c.Command("test prompt")
	want := []string{"claude", "--dangerously-skip-permissions", "--print", "--model", "claude-opus", "test prompt"}
	if !slices.Equal(cmd.Args, want) {
		t.Errorf("Command() args = %v, want %v", cmd.Args, want)
	}
}

func TestCopilotCommand_noModel(t *testing.T) {
	c := &Copilot{}
	cmd := c.Command("test prompt")
	want := []string{"copilot", "--silent", "--allow-all", "--autopilot", "--prompt", "test prompt"}
	if !slices.Equal(cmd.Args, want) {
		t.Errorf("Command() args = %v, want %v", cmd.Args, want)
	}
}

func TestCopilotCommand_withModel(t *testing.T) {
	c := &Copilot{Model: "gpt-4o"}
	cmd := c.Command("test prompt")
	want := []string{"copilot", "--silent", "--allow-all", "--autopilot", "--model", "gpt-4o", "--prompt", "test prompt"}
	if !slices.Equal(cmd.Args, want) {
		t.Errorf("Command() args = %v, want %v", cmd.Args, want)
	}
}
