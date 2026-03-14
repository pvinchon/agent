package assistant

import (
	"flag"
	"os"
	"testing"
)

func TestFlagSet_claude(t *testing.T) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	mustAssistant := FlagSet(fs)
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
	mustAssistant := FlagSet(fs)
	fs.Parse([]string{"--assistant=copilot"})

	a := mustAssistant()
	if a == nil {
		t.Fatal("expected non-nil assistant")
	}
	if _, ok := a.(*Copilot); !ok {
		t.Errorf("expected Copilot assistant, got %T", a)
	}
}

func TestFlagSet_withModel(t *testing.T) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	mustAssistant := FlagSet(fs)
	fs.Parse([]string{"--assistant=claude", "--model=claude-sonnet-4-5"})

	a := mustAssistant()
	c, ok := a.(*Claude)
	if !ok {
		t.Fatalf("expected *Claude, got %T", a)
	}
	if c.Model != "claude-sonnet-4-5" {
		t.Errorf("expected Model=%q, got %q", "claude-sonnet-4-5", c.Model)
	}
}

func TestFlagSet_modelRegistered(t *testing.T) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	FlagSet(fs)
	if fs.Lookup("model") == nil {
		t.Error("expected --model flag to be registered")
	}
}
