package assistant

import (
	"os/exec"
)

// claudeModels lists the models supported by the Claude assistant.
var claudeModels = []string{"claude-haiku-3-5", "claude-opus-4-5", "claude-sonnet-4-5"}

// Claude invokes the `claude` CLI.
type Claude struct {
	Model string
}

func (c *Claude) Command(prompt string) *exec.Cmd {
	args := []string{"--dangerously-skip-permissions", "--print"}
	if c.Model != "" {
		args = append(args, "--model", c.Model)
	}
	args = append(args, prompt)
	return exec.Command("claude", args...)
}
