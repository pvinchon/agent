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

func TestFlagSet_prefix(t *testing.T) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	FlagSet(fs, "review")
	FlagSet(fs, "fix")

	if fs.Lookup("review-assistant") == nil {
		t.Error("expected --review-assistant flag to be registered")
	}
	if fs.Lookup("fix-assistant") == nil {
		t.Error("expected --fix-assistant flag to be registered")
	}
	if fs.Lookup("assistant") != nil {
		t.Error("--assistant should not be registered when a prefix is given")
	}
}
