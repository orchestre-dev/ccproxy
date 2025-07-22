//go:build !windows
// +build !windows

package process

import (
	"fmt"
	"syscall"
)

// IsProcessRunning checks if a process with the given PID is running (Unix-specific)
func (pm *PIDManager) IsProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}

	// Try to send signal 0 to check if process exists
	err := syscall.Kill(pid, 0)
	return err == nil
}

// stopProcessByPID attempts to stop a process by PID (Unix-specific)
func (pm *PIDManager) stopProcessByPID(pid int) error {
	if pid <= 0 {
		return fmt.Errorf("invalid PID: %d", pid)
	}

	// Send SIGTERM for graceful shutdown
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM: %w", err)
	}

	return nil
}

// forceStopProcessByPID forcefully stops a process by PID (Unix-specific)
func (pm *PIDManager) forceStopProcessByPID(pid int) error {
	if pid <= 0 {
		return fmt.Errorf("invalid PID: %d", pid)
	}

	// Send SIGKILL for forceful termination
	if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
		return fmt.Errorf("failed to send SIGKILL: %w", err)
	}

	return nil
}
