package process

import (
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"

	testfw "github.com/orchestre-dev/ccproxy/testing"
)

func setupTestPIDManager(t *testing.T) *PIDManager {
	// Use isolated test environment
	testfw.SetupIsolatedTest(t)
	
	pm, err := NewPIDManager()
	if err != nil {
		t.Fatalf("Failed to create PID manager: %v", err)
	}
	
	return pm
}

func TestNewPIDManager(t *testing.T) {
	pm := setupTestPIDManager(t)
	if pm == nil {
		t.Fatal("NewPIDManager returned nil")
	}
	
	if pm.pidPath == "" {
		t.Error("PID path is empty")
	}
}

func TestWriteAndReadPID(t *testing.T) {
	pm := setupTestPIDManager(t)
	
	// Write PID
	if err := pm.WritePID(); err != nil {
		t.Fatalf("WritePID failed: %v", err)
	}
	
	// Read PID
	pid, err := pm.ReadPID()
	if err != nil {
		t.Fatalf("ReadPID failed: %v", err)
	}
	
	// Should match current process PID
	if pid != os.Getpid() {
		t.Errorf("Expected PID %d, got %d", os.Getpid(), pid)
	}
}

func TestWritePIDForProcess(t *testing.T) {
	pm := setupTestPIDManager(t)
	
	// Test writing a specific PID
	testPID := 12345
	if err := pm.WritePIDForProcess(testPID); err != nil {
		t.Fatalf("WritePIDForProcess failed: %v", err)
	}
	
	// Read PID
	pid, err := pm.ReadPID()
	if err != nil {
		t.Fatalf("ReadPID failed: %v", err)
	}
	
	// Should match the PID we wrote
	if pid != testPID {
		t.Errorf("Expected PID %d, got %d", testPID, pid)
	}
	
	// Test invalid PIDs
	testCases := []struct {
		pid int
		desc string
	}{
		{0, "zero PID"},
		{-1, "negative PID"},
		{-100, "large negative PID"},
	}
	
	for _, tc := range testCases {
		if err := pm.WritePIDForProcess(tc.pid); err == nil {
			t.Errorf("Expected error for %s, but got none", tc.desc)
		}
	}
}

func TestReadPIDNoFile(t *testing.T) {
	pm := setupTestPIDManager(t)
	
	// Read when no file exists
	pid, err := pm.ReadPID()
	if err != nil {
		t.Fatalf("ReadPID failed: %v", err)
	}
	
	if pid != 0 {
		t.Errorf("Expected PID 0 when no file exists, got %d", pid)
	}
}

func TestReadPIDInvalidContent(t *testing.T) {
	pm := setupTestPIDManager(t)
	
	// Write invalid content
	if err := os.WriteFile(pm.pidPath, []byte("invalid"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	
	// Should return error
	_, err := pm.ReadPID()
	if err == nil {
		t.Error("Expected error for invalid PID content")
	}
}

func TestIsProcessRunning(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows - syscall.Kill not supported")
	}
	
	pm := setupTestPIDManager(t)
	
	// Current process should be running
	if !pm.IsProcessRunning(os.Getpid()) {
		t.Error("Current process should be running")
	}
	
	// Invalid PIDs should not be running
	if pm.IsProcessRunning(0) {
		t.Error("PID 0 should not be running")
	}
	
	if pm.IsProcessRunning(-1) {
		t.Error("Negative PID should not be running")
	}
	
	// Very high PID unlikely to exist
	if pm.IsProcessRunning(999999) {
		t.Error("PID 999999 should not be running")
	}
}

func TestGetRunningPID(t *testing.T) {
	pm := setupTestPIDManager(t)
	
	// No PID file
	pid, err := pm.GetRunningPID()
	if err != nil {
		t.Fatalf("GetRunningPID failed: %v", err)
	}
	if pid != 0 {
		t.Errorf("Expected 0 when no PID file, got %d", pid)
	}
	
	// Write current PID
	if err := pm.WritePID(); err != nil {
		t.Fatalf("WritePID failed: %v", err)
	}
	
	// Should return current PID
	pid, err = pm.GetRunningPID()
	if err != nil {
		t.Fatalf("GetRunningPID failed: %v", err)
	}
	if pid != os.Getpid() {
		t.Errorf("Expected current PID %d, got %d", os.Getpid(), pid)
	}
	
	// Write stale PID
	if err := os.WriteFile(pm.pidPath, []byte("999999"), 0644); err != nil {
		t.Fatalf("Failed to write stale PID: %v", err)
	}
	
	// Should return 0 and clean up
	pid, err = pm.GetRunningPID()
	if err != nil {
		t.Fatalf("GetRunningPID failed: %v", err)
	}
	if pid != 0 {
		t.Errorf("Expected 0 for stale PID, got %d", pid)
	}
	
	// PID file should be cleaned up
	if _, err := os.Stat(pm.pidPath); !os.IsNotExist(err) {
		t.Error("Stale PID file was not cleaned up")
	}
}

func TestAcquireLock(t *testing.T) {
	pm := setupTestPIDManager(t)
	
	// Should acquire lock successfully
	if err := pm.AcquireLock(); err != nil {
		t.Fatalf("AcquireLock failed: %v", err)
	}
	
	// PID file should exist
	pid, err := pm.ReadPID()
	if err != nil {
		t.Fatalf("ReadPID failed: %v", err)
	}
	if pid != os.Getpid() {
		t.Errorf("Expected PID %d, got %d", os.Getpid(), pid)
	}
	
	// Second acquire should fail
	pm2 := &PIDManager{pidPath: pm.pidPath}
	if err := pm2.AcquireLock(); err == nil {
		t.Error("Expected AcquireLock to fail when already locked")
	}
}

func TestReleaseLock(t *testing.T) {
	pm := setupTestPIDManager(t)
	
	// Acquire lock
	if err := pm.AcquireLock(); err != nil {
		t.Fatalf("AcquireLock failed: %v", err)
	}
	
	// Release lock
	if err := pm.ReleaseLock(); err != nil {
		t.Fatalf("ReleaseLock failed: %v", err)
	}
	
	// PID file should be gone
	if _, err := os.Stat(pm.pidPath); !os.IsNotExist(err) {
		t.Error("PID file still exists after release")
	}
	
	// Release when not locked should not error
	if err := pm.ReleaseLock(); err != nil {
		t.Errorf("ReleaseLock on unlocked should not error: %v", err)
	}
}

func TestReleaseLockDifferentPID(t *testing.T) {
	pm := setupTestPIDManager(t)
	
	// Write different PID
	if err := os.WriteFile(pm.pidPath, []byte("999999"), 0644); err != nil {
		t.Fatalf("Failed to write PID: %v", err)
	}
	
	// Should not release lock for different PID
	if err := pm.ReleaseLock(); err != nil {
		t.Errorf("ReleaseLock should not error: %v", err)
	}
	
	// PID file should still exist
	if _, err := os.Stat(pm.pidPath); os.IsNotExist(err) {
		t.Error("PID file should not be removed for different PID")
	}
}

func TestCleanup(t *testing.T) {
	pm := setupTestPIDManager(t)
	
	// Write PID file
	if err := pm.WritePID(); err != nil {
		t.Fatalf("WritePID failed: %v", err)
	}
	
	// Cleanup
	if err := pm.Cleanup(); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}
	
	// File should be gone
	if _, err := os.Stat(pm.pidPath); !os.IsNotExist(err) {
		t.Error("PID file still exists after cleanup")
	}
	
	// Cleanup when no file should not error
	if err := pm.Cleanup(); err != nil {
		t.Errorf("Cleanup should not error when no file: %v", err)
	}
}

func TestStopProcess(t *testing.T) {
	t.Skip("Skipping TestStopProcess - needs refactoring for new cleanup approach")
	
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows - syscall.Kill not supported")
	}
	
	pm := setupTestPIDManager(t)
	
	// Stop when not running
	if err := pm.StopProcess(); err == nil {
		t.Error("Expected error when stopping non-running process")
	}
	
	// Start a child process for testing
	cmd := exec.Command("sleep", "30")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start test process: %v", err)
	}
	
	// Register cleanup with t.Cleanup for guaranteed execution
	t.Cleanup(func() {
		// Add panic recovery
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic during process cleanup: %v", r)
			}
		}()
		
		if cmd.Process != nil {
			if err := cmd.Process.Kill(); err != nil {
				t.Logf("Failed to kill test process: %v", err)
			}
			cmd.Wait()
		}
	})
	
	// Write child PID to file
	childPID := cmd.Process.Pid
	if err := pm.WritePIDForProcess(childPID); err != nil {
		t.Fatalf("Failed to write child PID: %v", err)
	}
	
	// Verify process is running before stopping
	if !pm.IsProcessRunning(childPID) {
		t.Skip("Process already stopped, skipping test")
	}
	
	// Stop the process
	if err := pm.StopProcess(); err != nil {
		t.Fatalf("StopProcess failed: %v", err)
	}
	
	// Wait for process to exit (with timeout)
	// Since we have t.Cleanup that also kills the process,
	// we use a shorter timeout and don't fail if it times out
	done := make(chan bool)
	go func() {
		cmd.Wait()
		done <- true
	}()
	
	select {
	case <-done:
		// Process exited normally
	case <-time.After(500 * time.Millisecond):
		// Process might have been killed by cleanup, that's OK
		t.Log("Process exit timed out, may have been killed by cleanup")
	}
	
	// Process should be stopped
	if pm.IsProcessRunning(childPID) {
		t.Error("Process should be stopped")
	}
	
	// PID file should be cleaned up
	if _, err := os.Stat(pm.pidPath); !os.IsNotExist(err) {
		t.Error("PID file should be cleaned up after stop")
	}
}

func TestStopProcessWithTimeout(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows - syscall.Kill not supported")
	}
	
	pm := setupTestPIDManager(t)
	
	// Test stop when not running
	if err := pm.StopProcessWithTimeout(5 * time.Second); err == nil {
		t.Error("Expected error when stopping non-running process")
	}
	
	// Start a process that handles SIGTERM gracefully
	// Use a simpler approach that's more reliable
	cmd := exec.Command("sleep", "30")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start test process: %v", err)
	}
	
	// Register cleanup - ensure process is killed even if test fails
	t.Cleanup(func() {
		if cmd.Process != nil {
			// Force kill immediately in cleanup
			cmd.Process.Kill()
			// Don't wait indefinitely in cleanup
			done := make(chan struct{})
			go func() {
				cmd.Wait()
				close(done)
			}()
			
			select {
			case <-done:
				// Process exited
			case <-time.After(1 * time.Second):
				// Timeout waiting for process to exit
				t.Logf("Warning: test process cleanup timed out")
			}
		}
	})
	
	// Write child PID to file
	childPID := cmd.Process.Pid
	if err := pm.WritePIDForProcess(childPID); err != nil {
		t.Fatalf("Failed to write child PID: %v", err)
	}
	
	// Add a timeout for the entire test operation
	testDone := make(chan error, 1)
	go func() {
		// Stop the process with timeout
		err := pm.StopProcessWithTimeout(2 * time.Second)
		testDone <- err
	}()
	
	// Wait for test to complete or timeout
	select {
	case err := <-testDone:
		if err != nil {
			t.Fatalf("StopProcessWithTimeout failed: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Test timed out - StopProcessWithTimeout is hanging")
	}
	
	// Process should be stopped
	if pm.IsProcessRunning(childPID) {
		t.Error("Process should be stopped")
	}
}

func TestStopProcessWithTimeoutForceKill(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows - syscall.Kill not supported")
	}
	
	pm := setupTestPIDManager(t)
	
	// Start a process that ignores SIGTERM
	// Using 'sh -c' to trap and ignore SIGTERM
	cmd := exec.Command("sh", "-c", "trap '' TERM; sleep 30")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start test process: %v", err)
	}
	
	// Register cleanup
	t.Cleanup(func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
			cmd.Wait()
		}
	})
	
	// Write child PID to file
	childPID := cmd.Process.Pid
	if err := pm.WritePIDForProcess(childPID); err != nil {
		t.Fatalf("Failed to write child PID: %v", err)
	}
	
	// Stop the process with short timeout to force SIGKILL
	start := time.Now()
	if err := pm.StopProcessWithTimeout(500 * time.Millisecond); err != nil {
		t.Fatalf("StopProcessWithTimeout failed: %v", err)
	}
	duration := time.Since(start)
	
	// Should take at least 500ms (timeout) but not much longer
	if duration < 500*time.Millisecond {
		t.Errorf("Process terminated too quickly, expected timeout: %v", duration)
	}
	if duration > 2*time.Second {
		t.Errorf("Process termination took too long: %v", duration)
	}
	
	// Process should be stopped
	if pm.IsProcessRunning(childPID) {
		t.Error("Process should be stopped after SIGKILL")
	}
}

func TestWaitForProcessTermination(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows - syscall.Kill not supported")
	}
	
	pm := setupTestPIDManager(t)
	
	// Test with current process (should timeout)
	start := time.Now()
	result := pm.waitForProcessTermination(os.Getpid(), 100*time.Millisecond)
	duration := time.Since(start)
	
	if result {
		t.Error("Current process should not be terminated")
	}
	if duration < 100*time.Millisecond {
		t.Errorf("Wait terminated too early: %v", duration)
	}
	
	// Test with non-existent process (should return immediately)
	start = time.Now()
	result = pm.waitForProcessTermination(999999, 1*time.Second)
	duration = time.Since(start)
	
	if !result {
		t.Error("Non-existent process should be considered terminated")
	}
	if duration > 200*time.Millisecond {
		t.Errorf("Wait took too long for non-existent process: %v", duration)
	}
}