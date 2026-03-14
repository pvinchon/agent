package assistant

import (
	"os/exec"
)

// Copilot invokes the `copilot` CLI.
type Copilot struct {
	Model string
}

func (c *Copilot) Command(prompt string) *exec.Cmd {
	args := []string{"--silent", "--allow-all", "--autopilot"}
	if c.Model != "" {
		args = append(args, "--model", c.Model)
	}
	args = append(args, "--prompt", prompt)
	return exec.Command("copilot", args...)
}
