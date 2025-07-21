package testing

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestIsolation(t *testing.T) {
	t.Run("creates unique test IDs", func(t *testing.T) {
		// Create multiple isolated tests
		ids := make(map[string]bool)
		var mu sync.Mutex
		
		// Run multiple tests in parallel to ensure unique IDs
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				
				subtest := &testing.T{}
				// Mock the test name
				isolation := &TestIsolation{
					t:           subtest,
					testID:      generateTestID(t),
					originalEnv: make(map[string]string),
				}
				
				mu.Lock()
				if ids[isolation.testID] {
					t.Errorf("Duplicate test ID generated: %s", isolation.testID)
				}
				ids[isolation.testID] = true
				mu.Unlock()
			}(i)
		}
		wg.Wait()
		
		assert.Len(t, ids, 10, "Should have 10 unique test IDs")
	})
	
	t.Run("isolates HOME directory", func(t *testing.T) {
		// Save original HOME
		originalHome := os.Getenv("HOME")
		
		// Create isolation
		isolation := SetupIsolatedTest(t)
		
		// Check that HOME was changed
		newHome := os.Getenv("HOME")
		assert.NotEqual(t, originalHome, newHome)
		assert.Contains(t, newHome, isolation.GetTestID())
		
		// Verify HOME directory exists
		_, err := os.Stat(newHome)
		require.NoError(t, err)
	})
	
	t.Run("isolates config directories", func(t *testing.T) {
		// Create isolation
		isolation := SetupIsolatedTest(t)
		
		// Check config directories were set
		newXDGConfig := os.Getenv("XDG_CONFIG_HOME")
		newXDGData := os.Getenv("XDG_DATA_HOME")
		
		assert.NotEmpty(t, newXDGConfig)
		assert.NotEmpty(t, newXDGData)
		assert.Contains(t, newXDGConfig, isolation.GetTestID())
		assert.Contains(t, newXDGData, isolation.GetTestID())
		
		// Verify directories exist
		_, err := os.Stat(newXDGConfig)
		require.NoError(t, err)
		_, err = os.Stat(newXDGData)
		require.NoError(t, err)
		
		// Check .ccproxy directory was created
		ccproxyDir := isolation.GetConfigDir()
		_, err = os.Stat(ccproxyDir)
		require.NoError(t, err)
	})
	
	t.Run("creates temp files in isolated directory", func(t *testing.T) {
		isolation := SetupIsolatedTest(t)
		
		// Create a temp file
		content := "test content"
		filePath := isolation.CreateTempFile("test.txt", content)
		
		// Verify file exists and has correct content
		assert.True(t, strings.HasPrefix(filePath, isolation.GetTempDir()))
		
		data, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Equal(t, content, string(data))
		
		// Create file in subdirectory
		subFilePath := isolation.CreateTempFile("subdir/test2.txt", "more content")
		assert.True(t, strings.Contains(subFilePath, "subdir"))
		
		_, err = os.Stat(subFilePath)
		require.NoError(t, err)
	})
	
	t.Run("tracks and restores environment variables", func(t *testing.T) {
		// Set up some test env vars
		os.Setenv("TEST_VAR_1", "original1")
		os.Setenv("TEST_VAR_2", "original2")
		defer os.Unsetenv("TEST_VAR_1")
		defer os.Unsetenv("TEST_VAR_2")
		
		isolation := SetupIsolatedTest(t)
		
		// Modify env vars
		isolation.SetEnv("TEST_VAR_1", "modified1")
		isolation.SetEnv("TEST_VAR_2", "modified2")
		isolation.SetEnv("TEST_VAR_3", "new3")
		
		// Check modifications
		assert.Equal(t, "modified1", os.Getenv("TEST_VAR_1"))
		assert.Equal(t, "modified2", os.Getenv("TEST_VAR_2"))
		assert.Equal(t, "new3", os.Getenv("TEST_VAR_3"))
		
		// After cleanup, check restoration
		isolation.cleanup()
		
		assert.Equal(t, "original1", os.Getenv("TEST_VAR_1"))
		assert.Equal(t, "original2", os.Getenv("TEST_VAR_2"))
		assert.Empty(t, os.Getenv("TEST_VAR_3"))
	})
}

func TestRunParallel(t *testing.T) {
	// Track home directories used by parallel tests
	homes := make(map[string]bool)
	var mu sync.Mutex
	
	// Run multiple parallel tests
	for i := 0; i < 5; i++ {
		t.Run(fmt.Sprintf("parallel_test_%d", i), func(t *testing.T) {
			RunParallel(t, func(t *testing.T) {
				home := os.Getenv("HOME")
				
				mu.Lock()
				// Each parallel test should have its own HOME
				if homes[home] {
					t.Errorf("HOME directory %s is already in use by another test", home)
				}
				homes[home] = true
				mu.Unlock()
				
				// Verify we can write to our isolated HOME
				testFile := filepath.Join(home, "test.txt")
				err := os.WriteFile(testFile, []byte("test"), 0644)
				require.NoError(t, err)
			})
		})
	}
}

func TestIsolatedTest(t *testing.T) {
	t.Run("provides convenient access to isolation", func(t *testing.T) {
		it := NewIsolatedTest(t)
		
		// Check that isolation is set up
		assert.NotNil(t, it.Isolation())
		assert.NotEmpty(t, it.TempDir())
		assert.NotEmpty(t, it.HomeDir())
		
		// Verify directories exist
		_, err := os.Stat(it.TempDir())
		require.NoError(t, err)
		
		_, err = os.Stat(it.HomeDir())
		require.NoError(t, err)
	})
}

func TestRunIsolated(t *testing.T) {
	originalHome := os.Getenv("HOME")
	
	RunIsolated(t, func(t *testing.T) {
		// Inside isolated test, HOME should be different
		isolatedHome := os.Getenv("HOME")
		assert.NotEqual(t, originalHome, isolatedHome)
		
		// Should be able to write to isolated HOME
		testFile := filepath.Join(isolatedHome, "isolated_test.txt")
		err := os.WriteFile(testFile, []byte("isolated"), 0644)
		require.NoError(t, err)
	})
	
	// After test, HOME should be restored
	assert.Equal(t, originalHome, os.Getenv("HOME"))
}

func TestEnvironmentRestoration(t *testing.T) {
	// This test specifically verifies environment restoration
	originalHome := os.Getenv("HOME")
	originalCustom := os.Getenv("CUSTOM_TEST_VAR")
	
	// Set a custom var
	os.Setenv("CUSTOM_TEST_VAR", "original_value")
	
	// Run a sub-test with isolation
	t.Run("isolated_subtest", func(t *testing.T) {
		isolation := SetupIsolatedTest(t)
		
		// Verify HOME changed
		assert.NotEqual(t, originalHome, os.Getenv("HOME"))
		
		// Modify custom var
		isolation.SetEnv("CUSTOM_TEST_VAR", "modified_value")
		assert.Equal(t, "modified_value", os.Getenv("CUSTOM_TEST_VAR"))
	})
	
	// After subtest completes, verify restoration
	assert.Equal(t, originalHome, os.Getenv("HOME"), "HOME should be restored after subtest")
	assert.Equal(t, "original_value", os.Getenv("CUSTOM_TEST_VAR"), "Custom var should be restored")
	
	// Clean up
	if originalCustom == "" {
		os.Unsetenv("CUSTOM_TEST_VAR")
	} else {
		os.Setenv("CUSTOM_TEST_VAR", originalCustom)
	}
}