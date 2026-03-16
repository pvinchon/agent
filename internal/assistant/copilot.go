package assistant

import (
	"os/exec"
)

// copilotModels lists models supported by the copilot CLI.
var copilotModels = []string{
	"gpt-4o",
	"gpt-4.1",
	"o4-mini",
	"claude-3.7-sonnet",
}

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
