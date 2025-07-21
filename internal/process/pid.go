package process

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gofrs/flock"
	"github.com/orchestre-dev/ccproxy/internal/utils"
)

// DefaultShutdownTimeout is the default timeout for graceful shutdown
const DefaultShutdownTimeout = 30 * time.Second

// PIDManager handles PID file operations with proper locking
type PIDManager struct {
	pidPath  string
	lockPath string
	flock    *flock.Flock
	mu       sync.Mutex
}

// NewPIDManager creates a new PID manager
func NewPIDManager() (*PIDManager, error) {
	homeDir, err := utils.InitializeHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize home directory: %w", err)
	}

	lockPath := homeDir.PIDPath + ".lock"
	return &PIDManager{
		pidPath:  homeDir.PIDPath,
		lockPath: lockPath,
		flock:    flock.New(lockPath),
	}, nil
}

// WritePID writes the current process PID to file
func (pm *PIDManager) WritePID() error {
	return pm.WritePIDForProcess(os.Getpid())
}

// WritePIDForProcess writes a specific process PID to file with locking
func (pm *PIDManager) WritePIDForProcess(pid int) error {
	if pid <= 0 {
		return fmt.Errorf("invalid PID: %d", pid)
	}

	// Check if flock is initialized - never allow fallback
	if pm.flock == nil {
		return fmt.Errorf("PID file locking not available - cannot safely create PID file")
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Try to acquire exclusive lock with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	locked, err := pm.flock.TryLockContext(ctx, time.Millisecond*100)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	if !locked {
		return fmt.Errorf("could not acquire lock on PID file")
	}
	defer pm.flock.Unlock()

	// Check if there's already a running process while holding the lock
	if data, err := os.ReadFile(pm.pidPath); err == nil {
		pidStr := strings.TrimSpace(string(data))
		if existingPID, err := strconv.Atoi(pidStr); err == nil && existingPID > 0 {
			// Check if that process is actually running
			if pm.IsProcessRunning(existingPID) {
				return fmt.Errorf("service already running with PID %d", existingPID)
			}
		}
	}

	data := []byte(strconv.Itoa(pid))

	// Write PID atomically
	if err := utils.WriteFileAtomic(pm.pidPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	return nil
}

// PIDPath returns the path to the PID file
func (pm *PIDManager) PIDPath() string {
	return pm.pidPath
}

// ReadPID reads the PID from file with shared locking
func (pm *PIDManager) ReadPID() (int, error) {
	// Check if flock is initialized
	if pm.flock == nil {
		// Fallback to simple read without locking for backward compatibility
		return pm.readPIDWithoutLock()
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Try to acquire shared lock with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	locked, err := pm.flock.TryRLockContext(ctx, time.Millisecond*100)
	if err != nil {
		return 0, fmt.Errorf("failed to acquire read lock: %w", err)
	}
	if !locked {
		return 0, fmt.Errorf("could not acquire read lock on PID file")
	}
	defer pm.flock.Unlock()

	return pm.readPIDWithoutLock()
}

// readPIDWithoutLock reads the PID from file without locking
func (pm *PIDManager) readPIDWithoutLock() (int, error) {
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
	if pid != 0 {
		// Clean up synchronously to ensure it's done
		_ = pm.Cleanup()
	}

	return 0, nil
}

// Cleanup removes the PID file with locking
func (pm *PIDManager) Cleanup() error {
	// Check if flock is initialized - never allow unsafe cleanup
	if pm.flock == nil {
		return fmt.Errorf("PID file locking not available - cannot safely remove PID file")
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Try to acquire exclusive lock with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	locked, err := pm.flock.TryLockContext(ctx, time.Millisecond*100)
	if err != nil {
		return fmt.Errorf("failed to acquire lock for cleanup: %w", err)
	}
	if !locked {
		return fmt.Errorf("could not acquire lock for cleanup")
	}
	defer pm.flock.Unlock()

	if err := os.Remove(pm.pidPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove PID file: %w", err)
	}
	return nil
}

// AcquireLock attempts to acquire an exclusive lock by creating PID file
func (pm *PIDManager) AcquireLock() error {
	// Check if flock is initialized - never allow unsafe locking
	if pm.flock == nil {
		return fmt.Errorf("PID file locking not available - cannot safely acquire lock")
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Try to acquire exclusive lock first
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	locked, err := pm.flock.TryLockContext(ctx, time.Millisecond*100)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	if !locked {
		return fmt.Errorf("could not acquire exclusive lock")
	}
	// Don't unlock here - keep the lock until ReleaseLock is called

	// Check if already running while holding the lock
	data, err := os.ReadFile(pm.pidPath)
	if err == nil {
		// PID file exists, check if process is running
		pidStr := strings.TrimSpace(string(data))
		if pid, err := strconv.Atoi(pidStr); err == nil && pm.IsProcessRunning(pid) {
			pm.flock.Unlock()
			return fmt.Errorf("service already running with PID %d", pid)
		}
	}

	// Write our PID while holding the lock
	pid := os.Getpid()
	data = []byte(strconv.Itoa(pid))

	if err := utils.WriteFileAtomic(pm.pidPath, data, 0644); err != nil {
		pm.flock.Unlock()
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	// Keep the lock - it will be released by ReleaseLock
	return nil
}

// ReleaseLock releases the lock by removing PID file
func (pm *PIDManager) ReleaseLock() error {
	// Check if flock is initialized - never allow unsafe release
	if pm.flock == nil {
		return fmt.Errorf("PID file locking not available - cannot safely release lock")
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Read PID file to verify it's ours (without acquiring lock as we already have it)
	data, err := os.ReadFile(pm.pidPath)
	if err == nil {
		pidStr := strings.TrimSpace(string(data))
		if pid, err := strconv.Atoi(pidStr); err == nil && pid == os.Getpid() {
			// It's our PID, remove the file
			os.Remove(pm.pidPath)
		}
	}

	// Always unlock the file lock
	pm.flock.Unlock()

	return nil
}

// StopProcess stops the running process with graceful shutdown
func (pm *PIDManager) StopProcess() error {
	return pm.StopProcessWithTimeout(DefaultShutdownTimeout)
}

// StopProcessWithTimeout stops the running process with configurable timeout
func (pm *PIDManager) StopProcessWithTimeout(timeout time.Duration) error {
	pid, err := pm.GetRunningPID()
	if err != nil {
		return fmt.Errorf("failed to get running PID: %w", err)
	}

	if pid == 0 {
		return fmt.Errorf("service is not running")
	}

	// Always try to clean up PID file after stopping
	// Even if the process doesn't clean up properly
	defer pm.Cleanup()

	// First, try graceful shutdown with SIGTERM
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		// Process might have already exited
		if err == syscall.ESRCH {
			return nil
		}
		return fmt.Errorf("failed to send SIGTERM to process: %w", err)
	}

	// Wait for process to terminate gracefully
	if pm.waitForProcessTermination(pid, timeout) {
		return nil
	}

	// Process didn't terminate gracefully, force kill with SIGKILL
	utils.GetLogger().Warn("Process did not terminate gracefully, sending SIGKILL")
	if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
		// Process might have already exited
		if err == syscall.ESRCH {
			return nil
		}
		return fmt.Errorf("failed to send SIGKILL to process: %w", err)
	}

	// Wait a bit more for forceful termination
	if pm.waitForProcessTermination(pid, 5*time.Second) {
		return nil
	}

	return fmt.Errorf("failed to stop process after SIGKILL")
}

// waitForProcessTermination waits for a process to terminate
func (pm *PIDManager) waitForProcessTermination(pid int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	checkInterval := 100 * time.Millisecond

	for time.Now().Before(deadline) {
		if !pm.IsProcessRunning(pid) {
			return true
		}
		time.Sleep(checkInterval)
	}

	return false
}
