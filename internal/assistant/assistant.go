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

var assistantNames = strings.Join(slices.Sorted(maps.Keys(modelsByAssistant)), ", ")

// modelsByAssistant maps each assistant name to its supported models.
var modelsByAssistant = map[string][]string{
	"claude":  claudeModels,
	"copilot": copilotModels,
}

// New returns an Assistant for the given name and optional model.
// If model is empty, the assistant uses its default model.
// Returns an error if the name is unknown or the model is not supported by the assistant.
func New(name, model string) (Assistant, error) {
	models, ok := modelsByAssistant[name]
	if !ok {
		return nil, fmt.Errorf("unknown assistant %q: choose %s", name, assistantNames)
	}
	if model != "" && !slices.Contains(models, model) {
		modelNames := strings.Join(models, ", ")
		return nil, fmt.Errorf("model %q is not supported by assistant %q: choose %s", model, name, modelNames)
	}
	switch name {
	case "claude":
		return &Claude{Model: model}, nil
	default:
		return &Copilot{Model: model}, nil
	}
}
