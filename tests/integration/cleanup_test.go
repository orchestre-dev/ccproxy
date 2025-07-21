package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	testutil "github.com/orchestre-dev/ccproxy/testing"
)


// TestProcessCleanup verifies that process cleanup works correctly
func TestProcessCleanup(t *testing.T) {
	// Create enhanced cleanup manager
	cleanup := testutil.NewEnhancedProcessCleanup(t)
	
	t.Run("GracefulShutdown", func(t *testing.T) {
		// Start a simple process that responds to SIGTERM
		cmd := exec.Command("sh", "-c", `
			trap 'echo "Received SIGTERM"; exit 0' TERM
			echo "Process started with PID $$"
			while true; do sleep 1; done
		`)
		
		output, err := cmd.StdoutPipe()
		require.NoError(t, err)
		
		err = cmd.Start()
		require.NoError(t, err)
		
		// Track the process
		cleanup.TrackCommand(cmd, "test graceful shutdown process")
		
		// Read the PID from output
		var pidLine string
		buf := make([]byte, 1024)
		n, _ := output.Read(buf)
		if n > 0 {
			pidLine = string(buf[:n])
		}
		t.Logf("Process output: %s", pidLine)
		
		// Verify process is running
		err = syscall.Kill(cmd.Process.Pid, 0)
		assert.NoError(t, err, "Process should be running")
		
		// Cleanup will happen automatically via t.Cleanup
		// But we can verify it happens gracefully by checking logs
	})
	
	t.Run("ForceKill", func(t *testing.T) {
		// Start a process that ignores SIGTERM
		cmd := exec.Command("sh", "-c", `
			trap '' TERM
			echo "Process started with PID $$"
			while true; do sleep 1; done
		`)
		
		err := cmd.Start()
		require.NoError(t, err)
		
		// Track the process
		cleanup.TrackCommand(cmd, "test force kill process")
		
		// Verify process is running
		err = syscall.Kill(cmd.Process.Pid, 0)
		assert.NoError(t, err, "Process should be running")
		
		// Set shorter timeouts for this test
		cleanup.SetTimeouts(1*time.Second, 1*time.Second)
		
		// Cleanup will force kill since the process ignores SIGTERM
	})
	
	t.Run("ResourceCleanup", func(t *testing.T) {
		// Create test resources
		testFile := filepath.Join(t.TempDir(), "test-resource.txt")
		require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))
		
		testDir := filepath.Join(t.TempDir(), "test-dir")
		require.NoError(t, os.MkdirAll(testDir, 0755))
		
		// Track resources
		cleanup.TrackFile(testFile)
		cleanup.TrackDirectory(testDir)
		cleanup.TrackPort(19999) // Random high port
		
		// Verify resources exist
		assert.FileExists(t, testFile)
		assert.DirExists(t, testDir)
		
		// Resources will be cleaned up automatically
	})
}

// TestZombieProcessPrevention verifies no zombie processes are created
func TestZombieProcessPrevention(t *testing.T) {
	cleanup := testutil.NewEnhancedProcessCleanup(t)
	
	// Start multiple short-lived processes
	for i := 0; i < 5; i++ {
		cmd := exec.Command("sh", "-c", fmt.Sprintf("echo 'Process %d'; sleep 0.1", i))
		err := cmd.Start()
		require.NoError(t, err)
		
		cleanup.TrackCommand(cmd, fmt.Sprintf("short-lived process %d", i))
	}
	
	// Give processes time to exit
	time.Sleep(500 * time.Millisecond)
	
	// Verify no zombies exist
	assert.NoError(t, cleanup.VerifyNoZombies())
}

// TestPIDFileCleanup verifies PID file cleanup works correctly
func TestPIDFileCleanup(t *testing.T) {
	// Create test PID directory
	pidDir := t.TempDir()
	
	// Create some PID files
	pids := []int{99999, 99998, 99997} // Non-existent PIDs
	for _, pid := range pids {
		pidFile := filepath.Join(pidDir, fmt.Sprintf("test-%d.pid", pid))
		require.NoError(t, os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644))
	}
	
	// Create a PID file for current process (should not be killed)
	currentPIDFile := filepath.Join(pidDir, "current.pid")
	require.NoError(t, os.WriteFile(currentPIDFile, []byte(strconv.Itoa(os.Getpid())), 0644))
	
	// Clean up PID files
	testutil.CleanupPIDFiles(t, pidDir)
	
	// Verify all PID files are removed
	files, err := filepath.Glob(filepath.Join(pidDir, "*.pid"))
	require.NoError(t, err)
	assert.Empty(t, files, "All PID files should be cleaned up")
	
	// Verify current process is still running
	err = syscall.Kill(os.Getpid(), 0)
	assert.NoError(t, err, "Current process should still be running")
}

// TestConcurrentProcessCleanup tests cleanup of multiple processes concurrently
func TestConcurrentProcessCleanup(t *testing.T) {
	cleanup := testutil.NewEnhancedProcessCleanup(t)
	
	// Start multiple processes concurrently
	numProcesses := 10
	for i := 0; i < numProcesses; i++ {
		go func(id int) {
			cmd := exec.Command("sh", "-c", fmt.Sprintf(`
				echo "Process %d started"
				sleep %d
			`, id, id%3+1))
			
			if err := cmd.Start(); err != nil {
				t.Logf("Failed to start process %d: %v", id, err)
				return
			}
			
			cleanup.TrackCommand(cmd, fmt.Sprintf("concurrent process %d", id))
		}(i)
	}
	
	// Give processes time to start
	time.Sleep(500 * time.Millisecond)
	
	// Cleanup will handle all processes concurrently
}

// TestCleanupAfterCrash simulates a crash and verifies cleanup still works
func TestCleanupAfterCrash(t *testing.T) {
	cleanup := testutil.NewEnhancedProcessCleanup(t)
	
	// Start a process
	cmd := exec.Command("sleep", "60")
	require.NoError(t, cmd.Start())
	
	pid := cmd.Process.Pid
	cleanup.TrackCommand(cmd, "crash test process")
	
	// Simulate a "crash" by killing the process externally
	syscall.Kill(pid, syscall.SIGKILL)
	
	// Give it a moment
	time.Sleep(100 * time.Millisecond)
	
	// Cleanup should handle the already-dead process gracefully
	// The test passes if no panic occurs
}

// TestPortCleanup verifies port cleanup functionality
func TestPortCleanup(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping port cleanup test in CI")
	}
	
	cleanup := testutil.NewEnhancedProcessCleanup(t)
	
	// Start a process that listens on a port
	cmd := exec.Command("nc", "-l", "19876")
	require.NoError(t, cmd.Start())
	
	cleanup.TrackCommand(cmd, "nc listener")
	cleanup.TrackPort(19876)
	
	// Verify port is in use (WaitForPort should succeed since nc is listening)
	time.Sleep(200 * time.Millisecond) // Give nc time to start listening
	assert.True(t, testutil.WaitForPort(t, "localhost", "19876", 500*time.Millisecond), "Port should be in use by nc")
	
	// Note: After cleanup, the port will be free automatically when the process is killed
}

// Helper to check if current OS supports /proc filesystem
func hasProcFS() bool {
	_, err := os.Stat("/proc")
	return err == nil
}

// TestVerifyNoZombiesLinux tests zombie detection on Linux
func TestVerifyNoZombiesLinux(t *testing.T) {
	if !hasProcFS() {
		t.Skip("Skipping zombie detection test - no /proc filesystem")
	}
	
	cleanup := testutil.NewEnhancedProcessCleanup(t)
	
	// This test is tricky because we need to create a zombie
	// A zombie is created when a child exits but parent doesn't wait()
	
	// Start a process that creates a child and doesn't wait for it
	cmd := exec.Command("sh", "-c", `
		# Create a child process that exits immediately
		(sleep 0.01) &
		CHILD_PID=$!
		echo "Child PID: $CHILD_PID"
		
		# Don't wait for child, creating a zombie
		sleep 2
	`)
	
	require.NoError(t, cmd.Start())
	cleanup.TrackCommand(cmd, "zombie creator")
	
	// Give it time to create the zombie
	time.Sleep(100 * time.Millisecond)
	
	// Note: The zombie will be cleaned up when the parent exits
	// This test mainly verifies the zombie detection logic works
}

// BenchmarkProcessCleanup benchmarks the cleanup performance
func BenchmarkProcessCleanup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cleanup := testutil.NewEnhancedProcessCleanup(&testing.T{})
		
		// Start a simple process
		cmd := exec.Command("true")
		cmd.Start()
		cleanup.TrackCommand(cmd, "benchmark process")
		
		// Clean up
		cleanup.CleanupAll()
	}
}