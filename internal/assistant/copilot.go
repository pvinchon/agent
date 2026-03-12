package assistant

import (
	"os/exec"
)

// Copilot invokes the `copilot` CLI.
type Copilot struct{}

func (c *Copilot) Command(prompt string) *exec.Cmd {
	return exec.Command("copilot", "--silent", "--allow-all", "--autopilot", "--prompt", prompt)
}
