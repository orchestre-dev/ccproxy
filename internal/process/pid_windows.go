//go:build windows
// +build windows

package process

import (
	"fmt"
	"os"
	"syscall"
)

// IsProcessRunning checks if a process with the given PID is running (Windows-specific)
func (pm *PIDManager) IsProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}

	// On Windows, we try to open the process
	handle, err := syscall.OpenProcess(syscall.PROCESS_QUERY_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(handle)

	// Check if process is still alive
	var exitCode uint32
	err = syscall.GetExitCodeProcess(handle, &exitCode)
	if err != nil {
		return false
	}

	// STILL_ACTIVE = 259
	return exitCode == 259
}

// stopProcessByPID attempts to stop a process by PID (Windows-specific)
func (pm *PIDManager) stopProcessByPID(pid int) error {
	if pid <= 0 {
		return fmt.Errorf("invalid PID: %d", pid)
	}

	// Find the process
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	// Send termination signal
	if err := process.Signal(os.Kill); err != nil {
		return fmt.Errorf("failed to terminate process: %w", err)
	}

	return nil
}

// forceStopProcessByPID forcefully stops a process by PID (Windows-specific)
func (pm *PIDManager) forceStopProcessByPID(pid int) error {
	// On Windows, regular stop is already forceful
	return pm.stopProcessByPID(pid)
}

