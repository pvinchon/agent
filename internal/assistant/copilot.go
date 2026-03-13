package assistant

import (
	"os/exec"
)

// copilotModels lists the models supported by the Copilot assistant.
var copilotModels = []string{"claude-3.5-sonnet", "gemini-2.0-flash", "gpt-4-turbo", "gpt-4o", "o1", "o3-mini"}

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
