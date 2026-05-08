//go:build !windows

package mcp

import (
	"os/exec"
	"syscall"
)

// setProcessGroup puts the child in its own process group so a kill on the
// group also reaps grandchildren. Without it, `npx <pkg>` exits but the Node
// process it spawned hangs around as a zombie.
func setProcessGroup(c *exec.Cmd) {
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
