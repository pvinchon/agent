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

var validAssistants = map[string]bool{
	"claude":  true,
	"copilot": true,
}

// modelsByAssistant lists the supported models for each assistant.
// If the user provides no model, the assistant's own default is used.
var modelsByAssistant = map[string][]string{
	"claude":  {"claude-haiku-3-5", "claude-opus-4-5", "claude-sonnet-4-5"},
	"copilot": {"claude-3.5-sonnet", "gemini-2.0-flash", "gpt-4-turbo", "gpt-4o", "o1", "o3-mini"},
}

var assistantNames = strings.Join(slices.Sorted(maps.Keys(validAssistants)), ", ")

// New returns the Assistant registered under name, optionally configured with model.
func New(name, model string) (Assistant, error) {
	if !validAssistants[name] {
		return nil, fmt.Errorf("unknown assistant %q: choose %s", name, assistantNames)
	}
	if model != "" {
		valid := modelsByAssistant[name]
		if !slices.Contains(valid, model) {
			return nil, fmt.Errorf("model %q is not available for assistant %q: choose %s",
				model, name, strings.Join(valid, ", "))
		}
	}
	switch name {
	case "claude":
		return &Claude{Model: model}, nil
	default: // "copilot"
		return &Copilot{Model: model}, nil
	}
}
