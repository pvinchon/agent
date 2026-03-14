package assistant

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"os"
	"os/exec"
	"slices"
	"strings"

	xio "github.com/pvinchon/agent/internal/x/io"
	xlog "github.com/pvinchon/agent/internal/x/log"
	xstrings "github.com/pvinchon/agent/internal/x/strings"
)

var assistantByName = map[string]func(string) Assistant{
	"claude":  func(model string) Assistant { return &Claude{Model: model} },
	"copilot": func(model string) Assistant { return &Copilot{Model: model} },
}

// Assistant is a generic interface for AI CLI assistants.
type Assistant interface {
	Command(prompt string) *exec.Cmd
}

// Prompt runs the given prompt through the assistant and returns the trimmed output.
func Prompt(a Assistant, prompt string) (string, error) {
	slog.Debug("assistant", "prompt", prompt)

	cmd := a.Command(prompt)

	var buf bytes.Buffer
	cmd.Stdout = &buf
	if xlog.IsLevelDebug() {
		cmd.Stderr = os.Stderr
		cmd.Stdout = io.MultiWriter(&buf, xio.PrefixWriter(os.Stdout, "assistant> "))
	}

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("assistant: %w", err)
	}
	result := xstrings.StripMarkdownFence(strings.TrimSpace(buf.String()))
	return result, nil
}

var assistantNames = strings.Join(slices.Sorted(maps.Keys(assistantByName)), ", ")

// New returns the Assistant registered under name, optionally configured with model.
func New(name, model string) (Assistant, error) {
	factory, ok := assistantByName[name]
	if !ok {
		return nil, fmt.Errorf("unknown assistant %q: choose %s", name, assistantNames)
	}
	return factory(model), nil
}
