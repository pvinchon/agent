package fixer

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

func TestTemplateFlagSet_empty(t *testing.T) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fixTemplate := TemplateFlagSet(fs)
	fs.Parse([]string{})

	got := fixTemplate()
	if got != "" {
		t.Errorf("expected empty string when no --fix-template flag, got %q", got)
	}
}

func TestTemplateFlagSet_file(t *testing.T) {
	dir := t.TempDir()
	content := "Custom fixer: {{issues}} and {{diff}}"
	path := filepath.Join(dir, "template.md")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fixTemplate := TemplateFlagSet(fs)
	fs.Parse([]string{"--fix-template=" + path})

	got := fixTemplate()
	if got != content {
		t.Errorf("got %q, want %q", got, content)
	}
}
