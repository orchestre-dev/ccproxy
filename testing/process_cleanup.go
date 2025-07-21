package testing

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"testing"
	"time"
)

// ProcessResource represents a resource that needs cleanup
type ProcessResource interface {
	Cleanup() error
}

// PortResource represents a port that needs to be freed
type PortResource struct {
	Port int
}

func (p *PortResource) Cleanup() error {
	// Check if port is still in use
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", p.Port))
	if err == nil {
		// Port is free, close the test listener
		listener.Close()
		return nil
	}
	// Port is still in use, can't do much about it here
	// The process cleanup should handle this
	return nil
}

// FileResource represents a file that needs cleanup
type FileResource struct {
	Path string
}

func (f *FileResource) Cleanup() error {
	if err := os.Remove(f.Path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove file %s: %w", f.Path, err)
	}
	return nil
}

// DirectoryResource represents a directory that needs cleanup
type DirectoryResource struct {
	Path string
}

func (d *DirectoryResource) Cleanup() error {
	if err := os.RemoveAll(d.Path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove directory %s: %w", d.Path, err)
	}
	return nil
}

// EnhancedProcessCleanup provides comprehensive process and resource cleanup
type EnhancedProcessCleanup struct {
	t               *testing.T
	mu              sync.Mutex
	processes       map[int]*ProcessInfo
	resources       []ProcessResource
	gracefulTimeout time.Duration
	forceTimeout    time.Duration
}

// ProcessInfo holds information about a tracked process
type ProcessInfo struct {
	Cmd         *exec.Cmd
	StartTime   time.Time
	Description string
	Resources   []ProcessResource
}

// NewEnhancedProcessCleanup creates a new enhanced cleanup manager
func NewEnhancedProcessCleanup(t *testing.T) *EnhancedProcessCleanup {
	epc := &EnhancedProcessCleanup{
		t:               t,
		processes:       make(map[int]*ProcessInfo),
		resources:       make([]ProcessResource, 0),
		gracefulTimeout: 10 * time.Second,
		forceTimeout:    5 * time.Second,
	}
	
	// Register cleanup function
	t.Cleanup(func() {
		epc.CleanupAll()
	})
	
	return epc
}

// SetTimeouts configures the graceful and force shutdown timeouts
func (epc *EnhancedProcessCleanup) SetTimeouts(graceful, force time.Duration) {
	epc.mu.Lock()
	defer epc.mu.Unlock()
	
	epc.gracefulTimeout = graceful
	epc.forceTimeout = force
}

// TrackCommand tracks a command for cleanup with optional description
func (epc *EnhancedProcessCleanup) TrackCommand(cmd *exec.Cmd, description string) {
	epc.mu.Lock()
	defer epc.mu.Unlock()
	
	if cmd.Process != nil {
		epc.processes[cmd.Process.Pid] = &ProcessInfo{
			Cmd:         cmd,
			StartTime:   time.Now(),
			Description: description,
			Resources:   make([]ProcessResource, 0),
		}
	}
}

// TrackResource adds a resource for cleanup
func (epc *EnhancedProcessCleanup) TrackResource(resource ProcessResource) {
	epc.mu.Lock()
	defer epc.mu.Unlock()
	
	epc.resources = append(epc.resources, resource)
}

// TrackPort tracks a port for cleanup
func (epc *EnhancedProcessCleanup) TrackPort(port int) {
	epc.TrackResource(&PortResource{Port: port})
}

// TrackFile tracks a file for cleanup
func (epc *EnhancedProcessCleanup) TrackFile(path string) {
	epc.TrackResource(&FileResource{Path: path})
}

// TrackDirectory tracks a directory for cleanup
func (epc *EnhancedProcessCleanup) TrackDirectory(path string) {
	epc.TrackResource(&DirectoryResource{Path: path})
}

// StartCommand starts a command and tracks it for cleanup
func (epc *EnhancedProcessCleanup) StartCommand(name string, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(name, args...)
	
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	
	epc.TrackCommand(cmd, fmt.Sprintf("%s %v", name, args))
	
	return cmd, nil
}

// gracefulShutdown attempts to gracefully shutdown a process
func (epc *EnhancedProcessCleanup) gracefulShutdown(info *ProcessInfo) error {
	if info.Cmd.Process == nil {
		return nil
	}
	
	pid := info.Cmd.Process.Pid
	epc.t.Logf("Attempting graceful shutdown of process %d (%s)", pid, info.Description)
	
	// Try SIGTERM first (or os.Interrupt on Windows)
	var sig os.Signal
	if syscall.SIGTERM == syscall.Signal(0xf) {
		sig = syscall.SIGTERM
	} else {
		sig = os.Interrupt
	}
	
	if err := info.Cmd.Process.Signal(sig); err != nil {
		if err == os.ErrProcessDone {
			epc.t.Logf("Process %d already terminated", pid)
			return nil
		}
		return fmt.Errorf("failed to send signal to process %d: %w", pid, err)
	}
	
	// Wait for graceful shutdown
	done := make(chan error, 1)
	go func() {
		done <- info.Cmd.Wait()
	}()
	
	select {
	case err := <-done:
		if err != nil && err.Error() != "signal: terminated" {
			epc.t.Logf("Process %d exited with error: %v", pid, err)
		} else {
			epc.t.Logf("Process %d terminated gracefully", pid)
		}
		return nil
	case <-time.After(epc.gracefulTimeout):
		return fmt.Errorf("process %d did not terminate within %v", pid, epc.gracefulTimeout)
	}
}

// forceKill forcefully kills a process
func (epc *EnhancedProcessCleanup) forceKill(info *ProcessInfo) error {
	if info.Cmd.Process == nil {
		return nil
	}
	
	pid := info.Cmd.Process.Pid
	epc.t.Logf("Force killing process %d (%s)", pid, info.Description)
	
	if err := info.Cmd.Process.Kill(); err != nil {
		if err == os.ErrProcessDone {
			epc.t.Logf("Process %d already terminated", pid)
			return nil
		}
		return fmt.Errorf("failed to kill process %d: %w", pid, err)
	}
	
	// Wait for process to die
	done := make(chan error, 1)
	go func() {
		done <- info.Cmd.Wait()
	}()
	
	select {
	case <-done:
		epc.t.Logf("Process %d killed successfully", pid)
		return nil
	case <-time.After(epc.forceTimeout):
		return fmt.Errorf("process %d did not die within %v after SIGKILL", pid, epc.forceTimeout)
	}
}

// CleanupProcess cleans up a single process
func (epc *EnhancedProcessCleanup) CleanupProcess(pid int) error {
	epc.mu.Lock()
	info, exists := epc.processes[pid]
	if !exists {
		epc.mu.Unlock()
		return fmt.Errorf("process %d not tracked", pid)
	}
	epc.mu.Unlock()
	
	// Try graceful shutdown first
	if err := epc.gracefulShutdown(info); err != nil {
		epc.t.Logf("Graceful shutdown failed: %v", err)
		
		// Force kill if graceful shutdown failed
		if err := epc.forceKill(info); err != nil {
			return fmt.Errorf("failed to cleanup process %d: %w", pid, err)
		}
	}
	
	// Clean up process-specific resources
	for _, resource := range info.Resources {
		if err := resource.Cleanup(); err != nil {
			epc.t.Logf("Failed to cleanup resource: %v", err)
		}
	}
	
	// Remove from tracking
	epc.mu.Lock()
	delete(epc.processes, pid)
	epc.mu.Unlock()
	
	return nil
}

// CleanupAll cleans up all tracked processes and resources
func (epc *EnhancedProcessCleanup) CleanupAll() {
	// Add panic recovery
	defer func() {
		if r := recover(); r != nil {
			epc.t.Logf("Recovered from panic during cleanup: %v", r)
		}
	}()
	
	epc.t.Logf("Starting comprehensive cleanup")
	
	// Clean up all processes
	epc.mu.Lock()
	pids := make([]int, 0, len(epc.processes))
	for pid := range epc.processes {
		pids = append(pids, pid)
	}
	epc.mu.Unlock()
	
	// Clean up processes in parallel for efficiency
	var wg sync.WaitGroup
	for _, pid := range pids {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			if err := epc.CleanupProcess(p); err != nil {
				epc.t.Logf("Failed to cleanup process %d: %v", p, err)
			}
		}(pid)
	}
	wg.Wait()
	
	// Clean up global resources
	epc.mu.Lock()
	resources := epc.resources
	epc.mu.Unlock()
	
	for _, resource := range resources {
		if err := resource.Cleanup(); err != nil {
			epc.t.Logf("Failed to cleanup resource: %v", err)
		}
	}
	
	// Clear all tracking
	epc.mu.Lock()
	epc.processes = make(map[int]*ProcessInfo)
	epc.resources = make([]ProcessResource, 0)
	epc.mu.Unlock()
	
	epc.t.Logf("Cleanup completed")
}

// VerifyNoZombies checks for zombie processes
func (epc *EnhancedProcessCleanup) VerifyNoZombies() error {
	epc.mu.Lock()
	defer epc.mu.Unlock()
	
	for pid, info := range epc.processes {
		if info.Cmd.Process == nil {
			continue
		}
		
		// Check if process is a zombie
		statPath := fmt.Sprintf("/proc/%d/stat", pid)
		data, err := os.ReadFile(statPath)
		if err != nil {
			// Process doesn't exist, that's fine
			continue
		}
		
		// Parse stat file to check process state
		// Format: pid (comm) state ...
		// State Z indicates zombie
		statStr := string(data)
		if len(statStr) > 0 {
			// Find the last ) and check the character after it
			lastParen := len(statStr) - 1
			for i := len(statStr) - 1; i >= 0; i-- {
				if statStr[i] == ')' {
					lastParen = i
					break
				}
			}
			
			if lastParen < len(statStr)-2 && statStr[lastParen+2] == 'Z' {
				return fmt.Errorf("process %d (%s) is a zombie", pid, info.Description)
			}
		}
	}
	
	return nil
}

// CleanupPIDFiles cleans up PID files in a directory
func CleanupPIDFiles(t *testing.T, pidDir string) {
	t.Helper()
	
	// List all PID files
	files, err := filepath.Glob(filepath.Join(pidDir, "*.pid"))
	if err != nil {
		t.Logf("Failed to list PID files: %v", err)
		return
	}
	
	for _, pidFile := range files {
		// Read PID from file
		data, err := os.ReadFile(pidFile)
		if err != nil {
			// Remove the file anyway
			os.Remove(pidFile)
			continue
		}
		
		// Parse PID
		pidStr := string(data)
		if pid, err := strconv.Atoi(pidStr); err == nil && pid > 0 {
			// Don't kill the current process
			if pid == os.Getpid() {
				t.Logf("Skipping current process PID %d", pid)
				// Still remove the PID file
				os.Remove(pidFile)
				continue
			}
			
			// Try to kill the process
			if proc, err := os.FindProcess(pid); err == nil {
				// Check if process exists before trying to kill it
				if err := proc.Signal(syscall.Signal(0)); err == nil {
					t.Logf("Killing process %d from PID file %s", pid, pidFile)
					proc.Signal(syscall.SIGTERM)
					time.Sleep(100 * time.Millisecond)
					proc.Kill()
				}
			}
		}
		
		// Remove PID file
		os.Remove(pidFile)
	}
}