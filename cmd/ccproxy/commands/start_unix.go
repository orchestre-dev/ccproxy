//go:build !windows
// +build !windows

package commands

import (
	"os/exec"
	"syscall"
)

// setPlatformSpecificAttrs sets platform-specific attributes for the command
func setPlatformSpecificAttrs(cmd *exec.Cmd) {
	// Set process group to ensure child processes are cleaned up
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}
