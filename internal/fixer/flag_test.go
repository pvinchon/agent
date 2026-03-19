package fixer

import (
	"flag"
	"os"
	"testing"
)

func writeTempTemplate(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "fixer-tmpl-*.md")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(f.Name()) })
	f.WriteString(content)
	f.Close()
	return f.Name()
}

func TestFlagSet(t *testing.T) {
	templatePath := writeTempTemplate(t, "FIXER {{issues}} {{diff}}")

	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	mustTemplate := FlagSet(fs)
	fs.Parse([]string{"--fixer-template=" + templatePath})

	tmpl := mustTemplate()
	if tmpl.String() != "FIXER {{issues}} {{diff}}" {
		t.Errorf("unexpected template content: %q", tmpl.String())
	}
}

func TestFlagSet_registeredFlag(t *testing.T) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	FlagSet(fs)

	if fs.Lookup("fixer-template") == nil {
		t.Error("expected --fixer-template flag to be registered")
	}
}
