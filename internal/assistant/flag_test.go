package assistant

import (
	"flag"
	"os"
	"testing"
)

func TestFlagSet_claude(t *testing.T) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	mustAssistant := FlagSet(fs, "")
	fs.Parse([]string{"--assistant=claude"})

	a := mustAssistant()
	if a == nil {
		t.Fatal("expected non-nil assistant")
	}
	if _, ok := a.(*Claude); !ok {
		t.Errorf("expected Claude assistant, got %T", a)
	}
}

func TestFlagSet_explicit(t *testing.T) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	mustAssistant := FlagSet(fs, "")
	fs.Parse([]string{"--assistant=copilot"})

	a := mustAssistant()
	if a == nil {
		t.Fatal("expected non-nil assistant")
	}
	if _, ok := a.(*Copilot); !ok {
		t.Errorf("expected Copilot assistant, got %T", a)
	}
}

func TestFlagSet_suffix(t *testing.T) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	FlagSet(fs, "review")
	FlagSet(fs, "fix")

	if fs.Lookup("assistant-for-review") == nil {
		t.Error("expected --assistant-for-review flag to be registered")
	}
	if fs.Lookup("assistant-for-fix") == nil {
		t.Error("expected --assistant-for-fix flag to be registered")
	}
	if fs.Lookup("assistant") != nil {
		t.Error("--assistant should not be registered when a suffix is given")
	}
}

func TestFlagSet_modelFlag(t *testing.T) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	FlagSet(fs, "")

	if fs.Lookup("model") == nil {
		t.Error("expected --model flag to be registered")
	}
}

func TestFlagSet_modelFlagWithSuffix(t *testing.T) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	FlagSet(fs, "review")
	FlagSet(fs, "fix")

	if fs.Lookup("model-for-review") == nil {
		t.Error("expected --model-for-review flag to be registered")
	}
	if fs.Lookup("model-for-fix") == nil {
		t.Error("expected --model-for-fix flag to be registered")
	}
	if fs.Lookup("model") != nil {
		t.Error("--model should not be registered when a suffix is given")
	}
}

func TestFlagSet_withValidModel(t *testing.T) {
	makeFakeCLIWithModels(t, "claude", claudeTestModels, "")
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	mustAssistant := FlagSet(fs, "")
	fs.Parse([]string{"--assistant=claude", "--model=test-claude-sonnet"})

	a := mustAssistant()
	c, ok := a.(*Claude)
	if !ok {
		t.Fatalf("expected *Claude, got %T", a)
	}
	if c.Model != "test-claude-sonnet" {
		t.Errorf("Model = %q, want %q", c.Model, "test-claude-sonnet")
	}
}
