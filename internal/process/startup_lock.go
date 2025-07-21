package process

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
	"github.com/orchestre-dev/ccproxy/internal/utils"
)

// StartupLock provides process-wide locking for startup operations
type StartupLock struct {
	lockPath string
	flock    *flock.Flock
}

// NewStartupLock creates a new startup lock
func NewStartupLock() (*StartupLock, error) {
	homeDir, err := utils.InitializeHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize home directory: %w", err)
	}

	lockPath := filepath.Join(homeDir.Root, ".startup.lock")
	return &StartupLock{
		lockPath: lockPath,
		flock:    flock.New(lockPath),
	}, nil
}

// TryLock attempts to acquire an exclusive lock for startup
// Returns true if lock was acquired, false if another process holds the lock
func (sl *StartupLock) TryLock() (bool, error) {
	// Try to acquire exclusive lock with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	locked, err := sl.flock.TryLockContext(ctx, 10*time.Millisecond)
	if err != nil {
		return false, fmt.Errorf("failed to acquire startup lock: %w", err)
	}

	return locked, nil
}

// Unlock releases the startup lock
func (sl *StartupLock) Unlock() error {
	return sl.flock.Unlock()
}

// WithLock executes a function while holding the startup lock
func (sl *StartupLock) WithLock(fn func() error) error {
	locked, err := sl.TryLock()
	if err != nil {
		return err
	}
	if !locked {
		return fmt.Errorf("another ccproxy startup is already in progress")
	}
	defer sl.Unlock()

	return fn()
}

// Cleanup removes the lock file
func (sl *StartupLock) Cleanup() error {
	// Unlock first if we hold the lock
	sl.flock.Unlock()

	// Remove the lock file
	if err := os.Remove(sl.lockPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove startup lock file: %w", err)
	}
	return nil
}
