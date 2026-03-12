package assistant

import (
	"flag"
	"os"
	"testing"
)

func TestFlagSet_default(t *testing.T) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	mustAssistant := FlagSet(fs)
	fs.Parse([]string{})

	a := mustAssistant()
	if a == nil {
		t.Fatal("expected non-nil assistant")
	}
	if _, ok := a.(*Claude); !ok {
		t.Errorf("expected default assistant to be Claude, got %T", a)
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
