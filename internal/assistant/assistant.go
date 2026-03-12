package assistant

import (
	"fmt"
	"log/slog"
	"maps"
	"os"
	"os/exec"
	"slices"
	"strings"

	xlog "github.com/pvinchon/agent/internal/x/log"
	xstrings "github.com/pvinchon/agent/internal/x/strings"
)

// Assistant is a generic interface for AI CLI assistants.
type Assistant interface {
	Command(prompt string) *exec.Cmd
}

// Prompt runs the given prompt through the assistant and returns the trimmed output.
func Prompt(a Assistant, prompt string) (string, error) {
	slog.Debug("assistant", "prompt", prompt)

	cmd := a.Command(prompt)
	if xlog.IsLevelDebug() {
		cmd.Stderr = os.Stderr
	}

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	result := xstrings.StripMarkdownFence(strings.TrimSpace(buf.String()))
	return result, nil
}

var assistantByName = map[string]Assistant{
	"claude":  &Claude{},
	"copilot": &Copilot{},
}

var assistantNames = strings.Join(slices.Collect(maps.Keys(assistantByName)), ", ")

// New returns the Assistant registered under name.
func New(name string) (Assistant, error) {
	a, ok := assistantByName[name]
	if !ok {
		return nil, fmt.Errorf("unknown assistant %q: choose %s", name, assistantNames)
	}
	return a, nil
}
