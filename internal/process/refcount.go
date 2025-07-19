package process

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/musistudio/ccproxy/internal/utils"
)

// ReferenceCounter manages reference counting for Claude Code instances
type ReferenceCounter struct {
	path  string
	mutex sync.Mutex
}

// NewReferenceCounter creates a new reference counter
func NewReferenceCounter() (*ReferenceCounter, error) {
	path, err := utils.GetReferenceCountPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get reference count path: %w", err)
	}
	
	return &ReferenceCounter{
		path: path,
	}, nil
}

// Increment atomically increments the reference count
func (rc *ReferenceCounter) Increment() error {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	
	// Read current count
	count := rc.readCount()
	
	// Increment
	count++
	
	// Write back
	return rc.writeCount(count)
}

// Decrement atomically decrements the reference count
func (rc *ReferenceCounter) Decrement() (bool, error) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	
	// Read current count
	count := rc.readCount()
	
	// Decrement with Math.max(0, count - 1) logic
	if count > 0 {
		count--
	}
	
	// Write back
	if err := rc.writeCount(count); err != nil {
		return false, err
	}
	
	// Return true if count reached zero
	return count == 0, nil
}

// GetCount returns the current reference count
func (rc *ReferenceCounter) GetCount() int {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	
	return rc.readCount()
}

// Reset resets the reference count to zero
func (rc *ReferenceCounter) Reset() error {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	
	return rc.writeCount(0)
}

// readCount reads the count from file (assumes mutex is held)
func (rc *ReferenceCounter) readCount() int {
	data, err := os.ReadFile(rc.path)
	if err != nil {
		// File doesn't exist or error reading - assume 0
		return 0
	}
	
	// Parse count
	countStr := string(data)
	count, err := strconv.Atoi(countStr)
	if err != nil {
		// Corrupted file - return 0
		return 0
	}
	
	// Ensure non-negative
	if count < 0 {
		return 0
	}
	
	return count
}

// writeCount writes the count to file (assumes mutex is held)
func (rc *ReferenceCounter) writeCount(count int) error {
	// Ensure non-negative
	if count < 0 {
		count = 0
	}
	
	// Convert to string
	data := []byte(strconv.Itoa(count))
	
	// Write atomically
	return utils.WriteFileAtomic(rc.path, data, 0644)
}

// Cleanup removes the reference count file
func (rc *ReferenceCounter) Cleanup() error {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	
	if err := os.Remove(rc.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove reference count file: %w", err)
	}
	return nil
}

// IncrementAndCheck increments the count and returns the new value
func (rc *ReferenceCounter) IncrementAndCheck() (int, error) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	
	// Read current count
	count := rc.readCount()
	
	// Increment
	count++
	
	// Write back
	if err := rc.writeCount(count); err != nil {
		return 0, err
	}
	
	return count, nil
}

// DecrementAndCheck decrements the count and returns whether it reached zero
func (rc *ReferenceCounter) DecrementAndCheck() (bool, int, error) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	
	// Read current count
	count := rc.readCount()
	
	// Decrement with Math.max(0, count - 1) logic
	if count > 0 {
		count--
	}
	
	// Write back
	if err := rc.writeCount(count); err != nil {
		return false, 0, err
	}
	
	// Return whether count reached zero and the new count
	return count == 0, count, nil
}