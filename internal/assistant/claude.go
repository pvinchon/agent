package assistant

import (
	"os/exec"
)

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

// ModelsCommand returns the command that lists available Claude models (one per line).
func (c *Claude) ModelsCommand() *exec.Cmd {
	return exec.Command("claude", "models")
}
