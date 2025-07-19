package process

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/musistudio/ccproxy/internal/utils"
)

// PIDManager handles PID file operations
type PIDManager struct {
	pidPath string
}

// NewPIDManager creates a new PID manager
func NewPIDManager() (*PIDManager, error) {
	homeDir, err := utils.InitializeHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize home directory: %w", err)
	}
	
	return &PIDManager{
		pidPath: homeDir.PIDPath,
	}, nil
}

// WritePID writes the current process PID to file
func (pm *PIDManager) WritePID() error {
	pid := os.Getpid()
	data := []byte(strconv.Itoa(pid))
	
	// Write PID atomically
	if err := utils.WriteFileAtomic(pm.pidPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}
	
	return nil
}

// ReadPID reads the PID from file
func (pm *PIDManager) ReadPID() (int, error) {
	data, err := os.ReadFile(pm.pidPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil // No PID file means not running
		}
		return 0, fmt.Errorf("failed to read PID file: %w", err)
	}
	
	// Parse PID
	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0, fmt.Errorf("invalid PID in file: %s", pidStr)
	}
	
	return pid, nil
}

// IsProcessRunning checks if a process with given PID is running
func (pm *PIDManager) IsProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	
	// Try to send signal 0 to check if process exists
	err := syscall.Kill(pid, 0)
	return err == nil
}

// GetRunningPID returns the PID if the service is running, 0 otherwise
func (pm *PIDManager) GetRunningPID() (int, error) {
	pid, err := pm.ReadPID()
	if err != nil {
		return 0, err
	}
	
	if pid > 0 && pm.IsProcessRunning(pid) {
		return pid, nil
	}
	
	// Process not running, clean up stale PID file
	if pid > 0 {
		pm.Cleanup()
	}
	
	return 0, nil
}

// Cleanup removes the PID file
func (pm *PIDManager) Cleanup() error {
	if err := os.Remove(pm.pidPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove PID file: %w", err)
	}
	return nil
}

// AcquireLock attempts to acquire an exclusive lock by creating PID file
func (pm *PIDManager) AcquireLock() error {
	// Check if already running
	runningPID, err := pm.GetRunningPID()
	if err != nil {
		return fmt.Errorf("failed to check running process: %w", err)
	}
	
	if runningPID > 0 {
		return fmt.Errorf("service already running with PID %d", runningPID)
	}
	
	// Write our PID
	if err := pm.WritePID(); err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	
	return nil
}

// ReleaseLock releases the lock by removing PID file
func (pm *PIDManager) ReleaseLock() error {
	// Only remove if it's our PID
	pid, err := pm.ReadPID()
	if err != nil {
		return err
	}
	
	if pid == os.Getpid() {
		return pm.Cleanup()
	}
	
	return nil
}

// StopProcess stops the running process
func (pm *PIDManager) StopProcess() error {
	pid, err := pm.GetRunningPID()
	if err != nil {
		return fmt.Errorf("failed to get running PID: %w", err)
	}
	
	if pid == 0 {
		return fmt.Errorf("service is not running")
	}
	
	// Send SIGTERM
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		// Process might have already exited
		if err == syscall.ESRCH {
			// Clean up PID file
			pm.Cleanup()
			return nil
		}
		return fmt.Errorf("failed to stop process: %w", err)
	}
	
	// Always try to clean up PID file after stopping
	// Even if the process doesn't clean up properly
	defer pm.Cleanup()
	
	return nil
}