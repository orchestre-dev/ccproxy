package testing

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestConfig holds common test configuration
type TestConfig struct {
	TempDir     string
	CleanupFunc func()
}

// SetupTest provides common test setup functionality
func SetupTest(t *testing.T) *TestConfig {
	t.Helper()

	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "ccproxy-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Setup cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	// Register cleanup to run after test
	t.Cleanup(cleanup)

	return &TestConfig{
		TempDir:     tempDir,
		CleanupFunc: cleanup,
	}
}

// CreateTempFile creates a temporary file with the given content
func CreateTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	filePath := filepath.Join(dir, name)
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file %s: %v", filePath, err)
	}

	return filePath
}

// AssertNoError fails the test if err is not nil
func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err != nil {
		if len(msgAndArgs) > 0 {
			t.Fatalf("Expected no error, but got: %v - %v", err, msgAndArgs[0])
		} else {
			t.Fatalf("Expected no error, but got: %v", err)
		}
	}
}

// AssertError fails the test if err is nil
func AssertError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err == nil {
		if len(msgAndArgs) > 0 {
			t.Fatalf("Expected an error, but got nil - %v", msgAndArgs[0])
		} else {
			t.Fatal("Expected an error, but got nil")
		}
	}
}

// AssertEqual fails the test if expected != actual
func AssertEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if expected != actual {
		if len(msgAndArgs) > 0 {
			t.Fatalf("Expected %v, but got %v - %v", expected, actual, msgAndArgs[0])
		} else {
			t.Fatalf("Expected %v, but got %v", expected, actual)
		}
	}
}

// AssertNotEqual fails the test if expected == actual
func AssertNotEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if expected == actual {
		if len(msgAndArgs) > 0 {
			t.Fatalf("Expected %v to not equal %v - %v", expected, actual, msgAndArgs[0])
		} else {
			t.Fatalf("Expected %v to not equal %v", expected, actual)
		}
	}
}

// AssertTrue fails the test if value is not true
func AssertTrue(t *testing.T, value bool, msgAndArgs ...interface{}) {
	t.Helper()
	if !value {
		if len(msgAndArgs) > 0 {
			t.Fatalf("Expected true, but got false - %v", msgAndArgs[0])
		} else {
			t.Fatal("Expected true, but got false")
		}
	}
}

// AssertFalse fails the test if value is not false
func AssertFalse(t *testing.T, value bool, msgAndArgs ...interface{}) {
	t.Helper()
	if value {
		if len(msgAndArgs) > 0 {
			t.Fatalf("Expected false, but got true - %v", msgAndArgs[0])
		} else {
			t.Fatal("Expected false, but got true")
		}
	}
}

// AssertContains fails the test if container does not contain item
func AssertContains(t *testing.T, container, item string, msgAndArgs ...interface{}) {
	t.Helper()
	if !ContainsString(container, item) {
		if len(msgAndArgs) > 0 {
			t.Fatalf("Expected '%s' to contain '%s' - %v", container, item, msgAndArgs[0])
		} else {
			t.Fatalf("Expected '%s' to contain '%s'", container, item)
		}
	}
}

// AssertNotContains fails the test if container contains item
func AssertNotContains(t *testing.T, container, item string, msgAndArgs ...interface{}) {
	t.Helper()
	if ContainsString(container, item) {
		if len(msgAndArgs) > 0 {
			t.Fatalf("Expected '%s' to not contain '%s' - %v", container, item, msgAndArgs[0])
		} else {
			t.Fatalf("Expected '%s' to not contain '%s'", container, item)
		}
	}
}

// ContainsString checks if a string contains a substring
func ContainsString(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Eventually runs the given function repeatedly until it returns true or times out
func Eventually(t *testing.T, condition func() bool, timeout time.Duration, interval time.Duration, msgAndArgs ...interface{}) {
	t.Helper()

	start := time.Now()
	for {
		if condition() {
			return
		}

		if time.Since(start) > timeout {
			if len(msgAndArgs) > 0 {
				t.Fatalf("Condition was not met within %v - %v", timeout, msgAndArgs[0])
			} else {
				t.Fatalf("Condition was not met within %v", timeout)
			}
		}

		time.Sleep(interval)
	}
}

// WithTimeout runs a function with a timeout
func WithTimeout(t *testing.T, timeout time.Duration, fn func()) {
	t.Helper()

	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()

	select {
	case <-done:
		// Function completed successfully
	case <-time.After(timeout):
		t.Fatalf("Function did not complete within %v", timeout)
	}
}
