package testing

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

// ProcessCleanup provides utilities for cleaning up processes in tests
type ProcessCleanup struct {
	t         *testing.T
	processes []*exec.Cmd
}

// NewProcessCleanup creates a new process cleanup manager
func NewProcessCleanup(t *testing.T) *ProcessCleanup {
	pc := &ProcessCleanup{
		t:         t,
		processes: make([]*exec.Cmd, 0),
	}
	
	// Register cleanup function
	t.Cleanup(func() {
		pc.CleanupAll()
	})
	
	return pc
}

// TrackCommand tracks a command for cleanup
func (pc *ProcessCleanup) TrackCommand(cmd *exec.Cmd) {
	pc.processes = append(pc.processes, cmd)
}

// CleanupAll cleans up all tracked processes
func (pc *ProcessCleanup) CleanupAll() {
	// Add panic recovery
	defer func() {
		if r := recover(); r != nil {
			pc.t.Logf("Recovered from panic during process cleanup: %v", r)
		}
	}()
	
	for _, cmd := range pc.processes {
		if cmd.Process != nil {
			// Try graceful shutdown first
			if err := cmd.Process.Signal(os.Interrupt); err == nil {
				// Give it a moment to exit gracefully
				done := make(chan bool, 1)
				go func() {
					cmd.Wait()
					done <- true
				}()
				
				select {
				case <-done:
					// Process exited gracefully
					continue
				case <-time.After(2 * time.Second):
					// Force kill if not exited
				}
			}
			
			// Force kill
			if err := cmd.Process.Kill(); err != nil {
				pc.t.Logf("Failed to kill process %d: %v", cmd.Process.Pid, err)
			}
			cmd.Wait() // Reap the process
		}
	}
}

// StartCommand starts a command and tracks it for cleanup
func (pc *ProcessCleanup) StartCommand(name string, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(name, args...)
	pc.TrackCommand(cmd)
	
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	
	return cmd, nil
}

// WaitForPort waits for a port to become available with timeout
func WaitForPort(t *testing.T, host string, port string, timeout time.Duration) bool {
	t.Helper()
	
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		cmd := exec.Command("nc", "-z", host, port)
		
		// Run with timeout to prevent hanging
		done := make(chan error, 1)
		go func() {
			done <- cmd.Run()
		}()
		
		select {
		case err := <-done:
			if err == nil {
				return true
			}
		case <-time.After(1 * time.Second):
			// Kill nc if it's hanging
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
		}
		
		time.Sleep(100 * time.Millisecond)
	}
	
	return false
}

// CleanupPIDFile removes a PID file and kills the associated process if running
func CleanupPIDFile(t *testing.T, pidPath string) {
	t.Helper()
	
	// Add panic recovery
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Recovered from panic during PID cleanup: %v", r)
		}
	}()
	
	// Read PID file
	data, err := os.ReadFile(pidPath)
	if err != nil {
		// No PID file, nothing to clean
		return
	}
	
	// Parse PID
	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err == nil && pid > 0 {
		// Try to kill the process
		if proc, err := os.FindProcess(pid); err == nil {
			proc.Kill()
		}
	}
	
	// Remove PID file
	os.Remove(pidPath)
}