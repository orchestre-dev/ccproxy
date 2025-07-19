package process

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"testing"
	"time"
)

func setupTestPIDManager(t *testing.T) *PIDManager {
	// Create temp directory for testing
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		oldHome = os.Getenv("USERPROFILE")
		os.Setenv("USERPROFILE", tempDir)
	} else {
		os.Setenv("HOME", tempDir)
	}
	t.Cleanup(func() {
		if runtime.GOOS == "windows" {
			os.Setenv("USERPROFILE", oldHome)
		} else {
			os.Setenv("HOME", oldHome)
		}
	})
	
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
	defer cmd.Process.Kill()
	
	// Write child PID to file
	childPID := cmd.Process.Pid
	if err := os.WriteFile(pm.pidPath, []byte(strconv.Itoa(childPID)), 0644); err != nil {
		t.Fatalf("Failed to write child PID: %v", err)
	}
	
	// Stop the process
	if err := pm.StopProcess(); err != nil {
		t.Fatalf("StopProcess failed: %v", err)
	}
	
	// Wait for process to exit (with timeout)
	done := make(chan bool)
	go func() {
		cmd.Wait()
		done <- true
	}()
	
	select {
	case <-done:
		// Process exited
	case <-time.After(2 * time.Second):
		t.Error("Process did not exit within timeout")
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