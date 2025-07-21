package testing

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// TestIsolation provides test isolation to prevent interference between parallel tests
type TestIsolation struct {
	t          *testing.T
	testID     string
	tempDir    string
	originalEnv map[string]string
	mu         sync.Mutex
}

// generateTestID creates a unique test ID for parallel execution
func generateTestID(t *testing.T) string {
	// Use test name as base
	testName := t.Name()
	
	// Add random suffix to ensure uniqueness in parallel runs
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	randomSuffix := hex.EncodeToString(randomBytes)
	
	// Clean test name for filesystem use
	cleanName := filepath.Clean(testName)
	cleanName = filepath.Base(cleanName)
	
	return fmt.Sprintf("%s_%s", cleanName, randomSuffix)
}

// SetupIsolatedTest creates an isolated test environment
func SetupIsolatedTest(t *testing.T) *TestIsolation {
	isolation := &TestIsolation{
		t:           t,
		testID:      generateTestID(t),
		originalEnv: make(map[string]string),
	}
	
	// Create isolated temp directory
	baseTempDir := t.TempDir()
	isolation.tempDir = filepath.Join(baseTempDir, isolation.testID)
	if err := os.MkdirAll(isolation.tempDir, 0755); err != nil {
		t.Fatalf("Failed to create isolated temp dir: %v", err)
	}
	
	// Set up isolated HOME directory
	isolation.setupIsolatedHome()
	
	// Set up isolated config directory
	isolation.setupIsolatedConfig()
	
	// Register cleanup
	t.Cleanup(func() {
		isolation.cleanup()
	})
	
	return isolation
}

// setupIsolatedHome creates an isolated HOME directory
func (i *TestIsolation) setupIsolatedHome() {
	i.mu.Lock()
	defer i.mu.Unlock()
	
	// Save original HOME
	if original, exists := os.LookupEnv("HOME"); exists {
		i.originalEnv["HOME"] = original
	}
	
	// Create isolated HOME
	homeDir := filepath.Join(i.tempDir, "home")
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		i.t.Fatalf("Failed to create isolated HOME: %v", err)
	}
	
	// Set new HOME
	os.Setenv("HOME", homeDir)
}

// setupIsolatedConfig creates isolated config directories
func (i *TestIsolation) setupIsolatedConfig() {
	i.mu.Lock()
	defer i.mu.Unlock()
	
	// Create .ccproxy directory in isolated HOME
	homeDir := os.Getenv("HOME")
	ccproxyDir := filepath.Join(homeDir, ".ccproxy")
	if err := os.MkdirAll(ccproxyDir, 0755); err != nil {
		i.t.Fatalf("Failed to create .ccproxy directory: %v", err)
	}
	
	// Save and override XDG_CONFIG_HOME if set
	if original, exists := os.LookupEnv("XDG_CONFIG_HOME"); exists {
		i.originalEnv["XDG_CONFIG_HOME"] = original
	}
	configDir := filepath.Join(i.tempDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		i.t.Fatalf("Failed to create config directory: %v", err)
	}
	os.Setenv("XDG_CONFIG_HOME", configDir)
	
	// Save and override XDG_DATA_HOME if set
	if original, exists := os.LookupEnv("XDG_DATA_HOME"); exists {
		i.originalEnv["XDG_DATA_HOME"] = original
	}
	dataDir := filepath.Join(i.tempDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		i.t.Fatalf("Failed to create data directory: %v", err)
	}
	os.Setenv("XDG_DATA_HOME", dataDir)
}

// GetTestID returns the unique test ID
func (i *TestIsolation) GetTestID() string {
	return i.testID
}

// GetTempDir returns the isolated temp directory
func (i *TestIsolation) GetTempDir() string {
	return i.tempDir
}

// GetHomeDir returns the isolated HOME directory
func (i *TestIsolation) GetHomeDir() string {
	return os.Getenv("HOME")
}

// GetConfigDir returns the isolated config directory
func (i *TestIsolation) GetConfigDir() string {
	return filepath.Join(i.GetHomeDir(), ".ccproxy")
}

// CreateTempFile creates a file in the isolated temp directory
func (i *TestIsolation) CreateTempFile(name, content string) string {
	path := filepath.Join(i.tempDir, name)
	dir := filepath.Dir(path)
	
	if err := os.MkdirAll(dir, 0755); err != nil {
		i.t.Fatalf("Failed to create directory: %v", err)
	}
	
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		i.t.Fatalf("Failed to write file: %v", err)
	}
	
	return path
}

// GetPIDFilePath returns the PID file path for this isolated test
func (i *TestIsolation) GetPIDFilePath() string {
	pidDir := filepath.Join(i.GetHomeDir(), ".ccproxy")
	return filepath.Join(pidDir, "ccproxy.pid")
}

// SetEnv sets an environment variable and tracks it for cleanup
func (i *TestIsolation) SetEnv(key, value string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	
	// Save original value if not already saved
	if _, saved := i.originalEnv[key]; !saved {
		if original, exists := os.LookupEnv(key); exists {
			i.originalEnv[key] = original
		} else {
			i.originalEnv[key] = "" // Mark as unset
		}
	}
	
	os.Setenv(key, value)
}

// cleanup restores the original environment
func (i *TestIsolation) cleanup() {
	i.mu.Lock()
	defer i.mu.Unlock()
	
	// Restore original environment variables
	for key, value := range i.originalEnv {
		if value == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, value)
		}
	}
}

// IsolatedTest provides a helper for running isolated tests
type IsolatedTest struct {
	*testing.T
	isolation *TestIsolation
}

// NewIsolatedTest creates a new isolated test
func NewIsolatedTest(t *testing.T) *IsolatedTest {
	return &IsolatedTest{
		T:         t,
		isolation: SetupIsolatedTest(t),
	}
}

// Isolation returns the test isolation instance
func (it *IsolatedTest) Isolation() *TestIsolation {
	return it.isolation
}

// TempDir returns the isolated temp directory
func (it *IsolatedTest) TempDir() string {
	return it.isolation.GetTempDir()
}

// HomeDir returns the isolated HOME directory
func (it *IsolatedTest) HomeDir() string {
	return it.isolation.GetHomeDir()
}

// RunParallel runs a test function in parallel with proper isolation
func RunParallel(t *testing.T, testFunc func(t *testing.T)) {
	t.Run(t.Name(), func(t *testing.T) {
		t.Parallel()
		isolation := SetupIsolatedTest(t)
		_ = isolation // Ensure isolation is set up
		testFunc(t)
	})
}

// RunIsolated runs a test function with isolation (non-parallel)
func RunIsolated(t *testing.T, testFunc func(t *testing.T)) {
	isolation := SetupIsolatedTest(t)
	_ = isolation // Ensure isolation is set up
	testFunc(t)
}