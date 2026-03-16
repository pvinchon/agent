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
	// ModelsCommand returns the command that prints the list of available models,
	// one model name per line.
	ModelsCommand() *exec.Cmd
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

// Models queries the assistant CLI for its available models and returns them as
// a slice. The CLI is expected to print one model name per line.
func Models(a Assistant) ([]string, error) {
	out, err := a.ModelsCommand().Output()
	if err != nil {
		return nil, fmt.Errorf("could not list models: %w", err)
	}
	var models []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if m := strings.TrimSpace(line); m != "" {
			models = append(models, m)
		}
	}
	return models, nil
}

// assistantFactories maps each assistant name to a constructor that accepts an
// optional model name (empty string = use the CLI's default model).
var assistantFactories = map[string]func(model string) Assistant{
	"claude":  func(model string) Assistant { return &Claude{Model: model} },
	"copilot": func(model string) Assistant { return &Copilot{Model: model} },
}

var assistantNames = strings.Join(slices.Sorted(maps.Keys(assistantFactories)), ", ")

// New returns an Assistant for the given name and optional model.
// If model is empty, the assistant uses its CLI's default model.
// If model is non-empty, the available models are fetched from the CLI and the
// requested model is validated against that live list.
func New(name, model string) (Assistant, error) {
	factory, ok := assistantFactories[name]
	if !ok {
		return nil, fmt.Errorf("unknown assistant %q: choose %s", name, assistantNames)
	}
	if model != "" {
		models, err := Models(factory(""))
		if err != nil {
			return nil, fmt.Errorf("could not fetch models for %q: %w", name, err)
		}
		if !slices.Contains(models, model) {
			return nil, fmt.Errorf("model %q is not supported by assistant %q: choose %s",
				model, name, strings.Join(models, ", "))
		}
	}
	return factory(model), nil
}
