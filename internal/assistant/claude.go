package assistant

import (
	"os/exec"
)

// Claude invokes the `claude` CLI.
type Claude struct{}

func (c *Claude) Command(prompt string) *exec.Cmd {
	return exec.Command("claude", "--dangerously-skip-permissions", "--print", prompt)
}
