package assistant

import (
	"fmt"
	"maps"
	"os/exec"
	"slices"
	"strings"
)

// Assistant is a generic interface for AI CLI assistants.
type Assistant interface {
	Command(prompt string) *exec.Cmd
}

// Prompt runs the given prompt through the assistant and returns the trimmed output.
func Prompt(a Assistant, prompt string) (string, error) {
	out, err := a.Command(prompt).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
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
