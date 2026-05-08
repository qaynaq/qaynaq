//go:build windows

package mcp

import "os/exec"

func setProcessGroup(c *exec.Cmd) {}
