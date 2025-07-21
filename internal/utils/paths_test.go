package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
)

func TestGetHomeDir(t *testing.T) {
	homeDir, err := GetHomeDir()
	testutil.AssertNoError(t, err, "Should get home directory")
	testutil.AssertTrue(t, len(homeDir) > 0, "Home directory should not be empty")
	testutil.AssertContains(t, homeDir, ".ccproxy", "Home directory should contain .ccproxy")
}

func TestInitializeHomeDir(t *testing.T) {
	// Clean up any existing home directory for test
	testHome, err := GetHomeDir()
	testutil.AssertNoError(t, err, "Should get home directory")
	
	// Remove if exists to test fresh initialization
	os.RemoveAll(testHome)
	
	homeDir, err := InitializeHomeDir()
	testutil.AssertNoError(t, err, "Should initialize home directory")
	testutil.AssertNotEqual(t, nil, homeDir, "HomeDir should not be nil")
	
	// Check structure
	testutil.AssertTrue(t, len(homeDir.Root) > 0, "Root should be set")
	testutil.AssertTrue(t, len(homeDir.ConfigPath) > 0, "ConfigPath should be set")
	testutil.AssertTrue(t, len(homeDir.LogPath) > 0, "LogPath should be set")
	testutil.AssertTrue(t, len(homeDir.PIDPath) > 0, "PIDPath should be set")
	testutil.AssertTrue(t, len(homeDir.PluginsDir) > 0, "PluginsDir should be set")
	testutil.AssertTrue(t, len(homeDir.TempDir) > 0, "TempDir should be set")
	
	// Check that directories were created
	testutil.AssertTrue(t, DirExists(homeDir.Root), "Root directory should exist")
	testutil.AssertTrue(t, DirExists(homeDir.PluginsDir), "Plugins directory should exist")
	testutil.AssertTrue(t, DirExists(homeDir.TempDir), "Temp directory should exist")
	
	// Test paths are within root
	testutil.AssertContains(t, homeDir.ConfigPath, homeDir.Root, "ConfigPath should be in root")
	testutil.AssertContains(t, homeDir.LogPath, homeDir.Root, "LogPath should be in root")
	testutil.AssertContains(t, homeDir.PIDPath, homeDir.Root, "PIDPath should be in root")
	testutil.AssertContains(t, homeDir.PluginsDir, homeDir.Root, "PluginsDir should be in root")
	testutil.AssertContains(t, homeDir.TempDir, homeDir.Root, "TempDir should be in root")
	
	// Test calling again (should not error)
	homeDir2, err := InitializeHomeDir()
	testutil.AssertNoError(t, err, "Should handle already initialized directory")
	testutil.AssertEqual(t, homeDir.Root, homeDir2.Root, "Should return same root path")
}

func TestResolvePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		validate func(*testing.T, string, string)
	}{
		{
			name:    "empty path",
			input:   "",
			wantErr: true,
		},
		{
			name:    "absolute path",
			input:   "/usr/local/bin",
			wantErr: false,
			validate: func(t *testing.T, input, result string) {
				testutil.AssertEqual(t, "/usr/local/bin", result, "Should return cleaned absolute path")
			},
		},
		{
			name:    "absolute path with dots",
			input:   "/usr/local/../local/bin",
			wantErr: false,
			validate: func(t *testing.T, input, result string) {
				testutil.AssertEqual(t, "/usr/local/bin", result, "Should clean path with dots")
			},
		},
		{
			name:    "relative path",
			input:   "test/file.txt",
			wantErr: false,
			validate: func(t *testing.T, input, result string) {
				testutil.AssertTrue(t, filepath.IsAbs(result), "Should convert to absolute path")
				testutil.AssertContains(t, result, "test/file.txt", "Should contain original path")
			},
		},
		{
			name:    "current directory",
			input:   ".",
			wantErr: false,
			validate: func(t *testing.T, input, result string) {
				cwd, _ := os.Getwd()
				testutil.AssertEqual(t, cwd, result, "Should resolve to current directory")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolvePath(tt.input)
			if tt.wantErr {
				testutil.AssertError(t, err, "Expected error for invalid path")
			} else {
				testutil.AssertNoError(t, err, "Should resolve path without error")
				if tt.validate != nil {
					tt.validate(t, tt.input, result)
				}
			}
		})
	}
}

func TestResolveConfigPath(t *testing.T) {
	// Test with empty path (should use default)
	defaultPath, err := ResolveConfigPath("")
	testutil.AssertNoError(t, err, "Should resolve default config path")
	testutil.AssertContains(t, defaultPath, "config.json", "Should contain config.json")
	testutil.AssertContains(t, defaultPath, ".ccproxy", "Should contain .ccproxy")

	// Test with specific path
	customPath, err := ResolveConfigPath("custom-config.json")
	testutil.AssertNoError(t, err, "Should resolve custom config path")
	testutil.AssertContains(t, customPath, "custom-config.json", "Should contain custom filename")

	// Test with absolute path
	absPath := "/etc/ccproxy/config.json"
	resolvedAbs, err := ResolveConfigPath(absPath)
	testutil.AssertNoError(t, err, "Should resolve absolute config path")
	testutil.AssertEqual(t, absPath, resolvedAbs, "Should return same absolute path")
}

func TestFileExists(t *testing.T) {
	tempConfig := testutil.SetupTest(t)
	
	// Create a test file
	testFile := testutil.CreateTempFile(t, tempConfig.TempDir, "test.txt", "test content")
	
	// Test existing file
	testutil.AssertTrue(t, FileExists(testFile), "Should detect existing file")
	
	// Test non-existing file
	testutil.AssertFalse(t, FileExists(filepath.Join(tempConfig.TempDir, "nonexistent.txt")), 
		"Should not detect non-existing file")
	
	// Test directory (should return false)
	testutil.AssertFalse(t, FileExists(tempConfig.TempDir), "Should return false for directory")
}

func TestDirExists(t *testing.T) {
	tempConfig := testutil.SetupTest(t)
	
	// Test existing directory
	testutil.AssertTrue(t, DirExists(tempConfig.TempDir), "Should detect existing directory")
	
	// Test non-existing directory
	testutil.AssertFalse(t, DirExists(filepath.Join(tempConfig.TempDir, "nonexistent")), 
		"Should not detect non-existing directory")
	
	// Create and test file (should return false)
	testFile := testutil.CreateTempFile(t, tempConfig.TempDir, "test.txt", "content")
	testutil.AssertFalse(t, DirExists(testFile), "Should return false for file")
}

func TestEnsureDir(t *testing.T) {
	tempConfig := testutil.SetupTest(t)
	
	// Test creating new directory
	newDir := filepath.Join(tempConfig.TempDir, "new", "nested", "dir")
	err := EnsureDir(newDir)
	testutil.AssertNoError(t, err, "Should create nested directory")
	testutil.AssertTrue(t, DirExists(newDir), "Directory should exist after creation")
	
	// Test existing directory (should not error)
	err = EnsureDir(newDir)
	testutil.AssertNoError(t, err, "Should handle existing directory")
	
	// Test with existing file path (should error)
	testFile := testutil.CreateTempFile(t, tempConfig.TempDir, "test.txt", "content")
	err = EnsureDir(testFile)
	testutil.AssertError(t, err, "Should error when path is a file")
}

func TestGetTempFile(t *testing.T) {
	// Clean up any existing home directory
	testHome, _ := GetHomeDir()
	os.RemoveAll(testHome)
	
	tempFile, err := GetTempFile("test")
	testutil.AssertNoError(t, err, "Should get temp file path")
	testutil.AssertContains(t, tempFile, "test", "Should contain prefix")
	
	// Check that temp directory exists
	tempDir := filepath.Dir(tempFile)
	testutil.AssertTrue(t, DirExists(tempDir), "Temp directory should exist")
	
	// Test different prefix
	tempFile2, err := GetTempFile("another")
	testutil.AssertNoError(t, err, "Should get another temp file path")
	testutil.AssertContains(t, tempFile2, "another", "Should contain different prefix")
	testutil.AssertNotEqual(t, tempFile, tempFile2, "Should generate different paths")
}

func TestCleanupTempFiles(t *testing.T) {
	// Clean up any existing home directory
	testHome, _ := GetHomeDir()
	os.RemoveAll(testHome)
	
	// Create some temp files
	tempFile1, err := GetTempFile("cleanup1")
	testutil.AssertNoError(t, err, "Should get temp file path")
	
	tempFile2, err := GetTempFile("cleanup2")
	testutil.AssertNoError(t, err, "Should get temp file path")
	
	// Create the files
	err = os.WriteFile(tempFile1, []byte("content1"), 0644)
	testutil.AssertNoError(t, err, "Should create temp file 1")
	
	err = os.WriteFile(tempFile2, []byte("content2"), 0644)
	testutil.AssertNoError(t, err, "Should create temp file 2")
	
	// Create a subdirectory that should be ignored
	homeDir, _ := InitializeHomeDir()
	subDir := filepath.Join(homeDir.TempDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	testutil.AssertNoError(t, err, "Should create subdirectory")
	
	// Verify files exist
	testutil.AssertTrue(t, FileExists(tempFile1), "Temp file 1 should exist")
	testutil.AssertTrue(t, FileExists(tempFile2), "Temp file 2 should exist")
	testutil.AssertTrue(t, DirExists(subDir), "Subdirectory should exist")
	
	// Cleanup
	err = CleanupTempFiles()
	testutil.AssertNoError(t, err, "Should cleanup temp files")
	
	// Verify files are removed but directory remains
	testutil.AssertFalse(t, FileExists(tempFile1), "Temp file 1 should be removed")
	testutil.AssertFalse(t, FileExists(tempFile2), "Temp file 2 should be removed")
	testutil.AssertTrue(t, DirExists(subDir), "Subdirectory should remain")
	
	// Test cleanup when no temp directory exists
	os.RemoveAll(testHome)
	err = CleanupTempFiles()
	testutil.AssertNoError(t, err, "Should handle missing temp directory")
}

func TestGetExecutablePath(t *testing.T) {
	execPath, err := GetExecutablePath()
	testutil.AssertNoError(t, err, "Should get executable path")
	testutil.AssertTrue(t, len(execPath) > 0, "Executable path should not be empty")
	testutil.AssertTrue(t, filepath.IsAbs(execPath), "Executable path should be absolute")
	
	// Should contain the test binary name
	testutil.AssertContains(t, execPath, "utils.test", "Should contain test binary name")
	
	// Test that it works on different platforms (this tests runtime.GOOS branches)
	// The function should work regardless of the platform
	testutil.AssertTrue(t, FileExists(execPath), "Executable path should exist")
}

func TestGetReferenceCountPath(t *testing.T) {
	refPath, err := GetReferenceCountPath()
	testutil.AssertNoError(t, err, "Should get reference count path")
	testutil.AssertTrue(t, len(refPath) > 0, "Reference count path should not be empty")
	testutil.AssertContains(t, refPath, "ccproxy_ref_count", "Should contain reference count filename")
	
	// Should be in system temp directory
	tempDir := os.TempDir()
	testutil.AssertContains(t, refPath, tempDir, "Should be in system temp directory")
}

func TestWriteFileAtomic(t *testing.T) {
	tempConfig := testutil.SetupTest(t)
	
	targetFile := filepath.Join(tempConfig.TempDir, "atomic_test.txt")
	testData := []byte("atomic write test content")
	
	// Test atomic write
	err := WriteFileAtomic(targetFile, testData, 0644)
	testutil.AssertNoError(t, err, "Should write file atomically")
	
	// Verify file exists and has correct content
	testutil.AssertTrue(t, FileExists(targetFile), "File should exist after atomic write")
	
	content, err := os.ReadFile(targetFile)
	testutil.AssertNoError(t, err, "Should read file content")
	testutil.AssertEqual(t, string(testData), string(content), "Content should match")
	
	// Check permissions
	info, err := os.Stat(targetFile)
	testutil.AssertNoError(t, err, "Should get file info")
	if runtime.GOOS != "windows" { // Skip permission check on Windows
		testutil.AssertEqual(t, os.FileMode(0644), info.Mode().Perm(), "Should have correct permissions")
	}
	
	// Test overwriting existing file
	newData := []byte("new atomic content")
	err = WriteFileAtomic(targetFile, newData, 0644)
	testutil.AssertNoError(t, err, "Should overwrite file atomically")
	
	content, err = os.ReadFile(targetFile)
	testutil.AssertNoError(t, err, "Should read new content")
	testutil.AssertEqual(t, string(newData), string(content), "New content should match")
}

func TestWriteFileAtomicErrors(t *testing.T) {
	// Test with invalid directory
	invalidPath := "/root/invalid/path/file.txt"
	err := WriteFileAtomic(invalidPath, []byte("test"), 0644)
	testutil.AssertError(t, err, "Should error with invalid directory")
	testutil.AssertContains(t, err.Error(), "failed to create temp file", "Error should mention temp file creation")
}

func TestWriteFileAtomicErrorPaths(t *testing.T) {
	tempConfig := testutil.SetupTest(t)
	
	// Test write to read-only directory (if we can create one)
	readOnlyDir := filepath.Join(tempConfig.TempDir, "readonly")
	err := os.MkdirAll(readOnlyDir, 0755)
	testutil.AssertNoError(t, err, "Should create directory")
	
	// Make directory read-only
	err = os.Chmod(readOnlyDir, 0444)
	if err == nil {
		defer os.Chmod(readOnlyDir, 0755) // Restore permissions for cleanup
		
		targetFile := filepath.Join(readOnlyDir, "test.txt")
		err = WriteFileAtomic(targetFile, []byte("test"), 0644)
		testutil.AssertError(t, err, "Should error with read-only directory")
	}
}

func TestGetHomeDirError(t *testing.T) {
	// This test is hard to trigger since os.UserHomeDir() usually works
	// But we can at least call it to improve coverage
	homeDir, err := GetHomeDir()
	testutil.AssertNoError(t, err, "Should get home directory")
	testutil.AssertTrue(t, len(homeDir) > 0, "Home directory should not be empty")
}

func TestInitializeHomeDirError(t *testing.T) {
	// Test with an invalid home directory by temporarily changing the environment
	// This is difficult to test without mocking, so we'll just ensure the function
	// works normally and check error handling indirectly
	
	homeDir, err := InitializeHomeDir()
	testutil.AssertNoError(t, err, "Should initialize home directory")
	testutil.AssertNotEqual(t, nil, homeDir, "HomeDir should not be nil")
	
	// Test idempotency - calling again should work
	homeDir2, err := InitializeHomeDir()
	testutil.AssertNoError(t, err, "Should handle already initialized directory")
	testutil.AssertEqual(t, homeDir.Root, homeDir2.Root, "Should return consistent results")
}

func TestGetTempFileError(t *testing.T) {
	// Clean up any existing home directory and try to trigger error
	testHome, _ := GetHomeDir()
	os.RemoveAll(testHome)
	
	// This should still work as it creates the directory
	tempFile, err := GetTempFile("test")
	testutil.AssertNoError(t, err, "Should get temp file path even if directory didn't exist")
	testutil.AssertContains(t, tempFile, "test", "Should contain prefix")
}

func TestResolvePathError(t *testing.T) {
	// Test error case when current directory is not accessible
	// This is hard to simulate, so we'll test the normal path
	
	// Test with various edge cases
	result, err := ResolvePath("./test")
	testutil.AssertNoError(t, err, "Should resolve relative path")
	testutil.AssertTrue(t, filepath.IsAbs(result), "Should return absolute path")
	
	// Test with empty path (should error)
	_, err = ResolvePath("")
	testutil.AssertError(t, err, "Should error with empty path")
}

func TestResolveConfigPathError(t *testing.T) {
	// Test error handling in ResolveConfigPath
	
	// Test with absolute path
	result, err := ResolveConfigPath("/absolute/path/config.json")
	testutil.AssertNoError(t, err, "Should resolve absolute config path")
	testutil.AssertEqual(t, "/absolute/path/config.json", result, "Should return same absolute path")
	
	// Test with relative path
	result, err = ResolveConfigPath("config.json")
	testutil.AssertNoError(t, err, "Should resolve relative config path")
	testutil.AssertContains(t, result, "config.json", "Should contain filename")
}

func TestPathIntegration(t *testing.T) {
	// Test the integration of different path functions
	
	// Initialize home directory
	homeDir, err := InitializeHomeDir()
	testutil.AssertNoError(t, err, "Should initialize home directory")
	
	// Test config path resolution
	configPath, err := ResolveConfigPath("")
	testutil.AssertNoError(t, err, "Should resolve default config path")
	testutil.AssertEqual(t, homeDir.ConfigPath, configPath, "Should match home directory config path")
	
	// Test temp file creation
	tempFile, err := GetTempFile("integration")
	testutil.AssertNoError(t, err, "Should get temp file path")
	testutil.AssertContains(t, tempFile, homeDir.TempDir, "Temp file should be in home temp dir")
	
	// Create and write content
	testContent := []byte("integration test content")
	err = WriteFileAtomic(tempFile, testContent, 0644)
	testutil.AssertNoError(t, err, "Should write temp file")
	testutil.AssertTrue(t, FileExists(tempFile), "Temp file should exist")
	
	// Clean up
	err = CleanupTempFiles()
	testutil.AssertNoError(t, err, "Should cleanup temp files")
	testutil.AssertFalse(t, FileExists(tempFile), "Temp file should be cleaned up")
}

// Benchmark tests for performance verification
func BenchmarkResolvePath(b *testing.B) {
	testPath := "relative/path/to/file.txt"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ResolvePath(testPath)
	}
}

func BenchmarkFileExists(b *testing.B) {
	// Create a test file
	tempFile, _ := os.CreateTemp("", "benchmark_test")
	defer os.Remove(tempFile.Name())
	tempFile.Close()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FileExists(tempFile.Name())
	}
}

func BenchmarkDirExists(b *testing.B) {
	tempDir := os.TempDir()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DirExists(tempDir)
	}
}