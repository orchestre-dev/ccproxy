//go:build windows
// +build windows

package commands

import (
	"os/exec"
	"syscall"
)

// setPlatformSpecificAttrs sets platform-specific attributes for the command
func setPlatformSpecificAttrs(cmd *exec.Cmd) {
	// On Windows, we use CREATE_NEW_PROCESS_GROUP to ensure child processes are cleaned up
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
