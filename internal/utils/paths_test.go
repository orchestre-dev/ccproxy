package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetHomeDir(t *testing.T) {
	homeDir, err := GetHomeDir()
	if err != nil {
		t.Fatalf("GetHomeDir failed: %v", err)
	}
	
	// Should end with .ccproxy
	if !strings.HasSuffix(homeDir, ".ccproxy") {
		t.Errorf("Expected home dir to end with .ccproxy, got: %s", homeDir)
	}
	
	// Should be an absolute path
	if !filepath.IsAbs(homeDir) {
		t.Errorf("Expected absolute path, got: %s", homeDir)
	}
}

func TestInitializeHomeDir(t *testing.T) {
	// Create temp directory for testing
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		oldHome = os.Getenv("USERPROFILE")
		os.Setenv("USERPROFILE", tempDir)
	} else {
		os.Setenv("HOME", tempDir)
	}
	defer func() {
		if runtime.GOOS == "windows" {
			os.Setenv("USERPROFILE", oldHome)
		} else {
			os.Setenv("HOME", oldHome)
		}
	}()
	
	homeDir, err := InitializeHomeDir()
	if err != nil {
		t.Fatalf("InitializeHomeDir failed: %v", err)
	}
	
	// Check all paths are set correctly
	expectedRoot := filepath.Join(tempDir, ".ccproxy")
	if homeDir.Root != expectedRoot {
		t.Errorf("Expected root %s, got %s", expectedRoot, homeDir.Root)
	}
	
	if homeDir.ConfigPath != filepath.Join(expectedRoot, "config.json") {
		t.Errorf("Unexpected config path: %s", homeDir.ConfigPath)
	}
	
	if homeDir.LogPath != filepath.Join(expectedRoot, "ccproxy.log") {
		t.Errorf("Unexpected log path: %s", homeDir.LogPath)
	}
	
	if homeDir.PIDPath != filepath.Join(expectedRoot, ".ccproxy.pid") {
		t.Errorf("Unexpected PID path: %s", homeDir.PIDPath)
	}
	
	if homeDir.PluginsDir != filepath.Join(expectedRoot, "plugins") {
		t.Errorf("Unexpected plugins dir: %s", homeDir.PluginsDir)
	}
	
	// Check directories were created
	if !DirExists(homeDir.Root) {
		t.Error("Root directory was not created")
	}
	
	if !DirExists(homeDir.PluginsDir) {
		t.Error("Plugins directory was not created")
	}
	
	if !DirExists(homeDir.TempDir) {
		t.Error("Temp directory was not created")
	}
}

func TestResolvePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		checkAbs bool
	}{
		{
			name:     "empty path",
			input:    "",
			wantErr:  true,
		},
		{
			name:     "absolute path",
			input:    "/tmp/test.json",
			checkAbs: true,
		},
		{
			name:     "relative path",
			input:    "config.json",
			checkAbs: true,
		},
		{
			name:     "relative path with dots",
			input:    "../config.json",
			checkAbs: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := ResolvePath(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolvePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.checkAbs && err == nil && !filepath.IsAbs(resolved) {
				t.Errorf("Expected absolute path, got: %s", resolved)
			}
		})
	}
}

func TestFileAndDirExists(t *testing.T) {
	// Create temp directory and file
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.txt")
	
	// Test non-existent file
	if FileExists(tempFile) {
		t.Error("FileExists returned true for non-existent file")
	}
	
	// Create file
	if err := os.WriteFile(tempFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Test existing file
	if !FileExists(tempFile) {
		t.Error("FileExists returned false for existing file")
	}
	
	// Test directory with FileExists (should return false)
	if FileExists(tempDir) {
		t.Error("FileExists returned true for directory")
	}
	
	// Test directory with DirExists
	if !DirExists(tempDir) {
		t.Error("DirExists returned false for existing directory")
	}
	
	// Test file with DirExists (should return false)
	if DirExists(tempFile) {
		t.Error("DirExists returned true for file")
	}
}

func TestEnsureDir(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "test", "nested", "dir")
	
	// Directory shouldn't exist yet
	if DirExists(testDir) {
		t.Error("Test directory already exists")
	}
	
	// Create directory
	if err := EnsureDir(testDir); err != nil {
		t.Fatalf("EnsureDir failed: %v", err)
	}
	
	// Check it was created
	if !DirExists(testDir) {
		t.Error("Directory was not created")
	}
	
	// Call again (should not error)
	if err := EnsureDir(testDir); err != nil {
		t.Fatalf("EnsureDir failed on existing directory: %v", err)
	}
}

func TestGetTempFile(t *testing.T) {
	// Setup test home directory
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		oldHome = os.Getenv("USERPROFILE")
		os.Setenv("USERPROFILE", tempDir)
	} else {
		os.Setenv("HOME", tempDir)
	}
	defer func() {
		if runtime.GOOS == "windows" {
			os.Setenv("USERPROFILE", oldHome)
		} else {
			os.Setenv("HOME", oldHome)
		}
	}()
	
	tempFile, err := GetTempFile("test")
	if err != nil {
		t.Fatalf("GetTempFile failed: %v", err)
	}
	
	// Should be in the temp directory
	if !strings.Contains(tempFile, filepath.Join(".ccproxy", "tmp")) {
		t.Errorf("Temp file not in expected directory: %s", tempFile)
	}
	
	// Should contain the prefix
	if !strings.Contains(filepath.Base(tempFile), "test_") {
		t.Errorf("Temp file doesn't contain prefix: %s", tempFile)
	}
}

func TestCleanupTempFiles(t *testing.T) {
	// Setup test home directory
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		oldHome = os.Getenv("USERPROFILE")
		os.Setenv("USERPROFILE", tempDir)
	} else {
		os.Setenv("HOME", tempDir)
	}
	defer func() {
		if runtime.GOOS == "windows" {
			os.Setenv("USERPROFILE", oldHome)
		} else {
			os.Setenv("HOME", oldHome)
		}
	}()
	
	// Initialize home dir
	homeDir, err := InitializeHomeDir()
	if err != nil {
		t.Fatalf("InitializeHomeDir failed: %v", err)
	}
	
	// Create some temp files
	for i := 0; i < 3; i++ {
		tempFile := filepath.Join(homeDir.TempDir, strings.ReplaceAll("temp_file_%d", "%d", string(rune('0'+i))))
		if err := os.WriteFile(tempFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
	}
	
	// Create a subdirectory (should not be removed)
	subDir := filepath.Join(homeDir.TempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	
	// Run cleanup
	if err := CleanupTempFiles(); err != nil {
		t.Fatalf("CleanupTempFiles failed: %v", err)
	}
	
	// Check temp files are gone
	entries, err := os.ReadDir(homeDir.TempDir)
	if err != nil {
		t.Fatalf("Failed to read temp dir: %v", err)
	}
	
	// Should only have the subdirectory left
	if len(entries) != 1 || !entries[0].IsDir() {
		t.Errorf("Expected only subdirectory to remain, got %d entries", len(entries))
	}
}

func TestWriteFileAtomic(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "atomic.txt")
	testData := []byte("test data")
	
	// Write file atomically
	if err := WriteFileAtomic(targetFile, testData, 0644); err != nil {
		t.Fatalf("WriteFileAtomic failed: %v", err)
	}
	
	// Read and verify
	data, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	
	if string(data) != string(testData) {
		t.Errorf("Expected %s, got %s", testData, data)
	}
	
	// Check permissions
	info, err := os.Stat(targetFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	
	if runtime.GOOS != "windows" {
		// Windows doesn't support Unix permissions
		perm := info.Mode().Perm()
		if perm != 0644 {
			t.Errorf("Expected permissions 0644, got %o", perm)
		}
	}
}

func TestGetReferenceCountPath(t *testing.T) {
	path, err := GetReferenceCountPath()
	if err != nil {
		t.Fatalf("GetReferenceCountPath failed: %v", err)
	}
	
	// Should be in temp directory
	if !strings.HasPrefix(path, os.TempDir()) {
		t.Errorf("Reference count path not in temp directory: %s", path)
	}
	
	// Should contain ccproxy_ref_count
	if !strings.Contains(path, "ccproxy_ref_count") {
		t.Errorf("Reference count path doesn't contain expected name: %s", path)
	}
}

func TestResolveConfigPath(t *testing.T) {
	// Test with empty path (should use default)
	path, err := ResolveConfigPath("")
	if err != nil {
		t.Fatalf("ResolveConfigPath() with empty path failed: %v", err)
	}
	
	// Should return config.json in home directory
	if !strings.Contains(path, "config.json") {
		t.Errorf("Expected path to contain config.json, got %s", path)
	}
	
	// Test with specific path
	testPath := "/tmp/test-config.json"
	resolvedPath, err := ResolveConfigPath(testPath)
	if err != nil {
		t.Fatalf("ResolveConfigPath() failed: %v", err)
	}
	
	if resolvedPath != testPath {
		t.Errorf("Expected %s, got %s", testPath, resolvedPath)
	}
	
	// Test with home directory path
	userHome, _ := os.UserHomeDir()
	homePath := filepath.Join(userHome, "test-config.json")
	resolvedPath, err = ResolveConfigPath(homePath)
	if err != nil {
		t.Fatalf("ResolveConfigPath() with home path failed: %v", err)
	}
	
	if resolvedPath != homePath {
		t.Errorf("Expected %s, got %s", homePath, resolvedPath)
	}
}

func TestGetExecutablePath(t *testing.T) {
	path, err := GetExecutablePath()
	if err != nil {
		t.Fatalf("GetExecutablePath() failed: %v", err)
	}
	
	// Path should not be empty
	if path == "" {
		t.Error("Expected non-empty executable path")
	}
	
	// Path should be absolute
	if !filepath.IsAbs(path) {
		t.Errorf("Expected absolute path, got %s", path)
	}
	
	// For test environment, the path should exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Executable path does not exist: %s", path)
	}
}

// Additional test for GetTempFile with error case
func TestGetTempFile_Error(t *testing.T) {
	// Save original home
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		oldHome = os.Getenv("USERPROFILE")
	}
	
	// Create a file where the .ccproxy directory should be
	invalidHome := filepath.Join(tempDir, "invalid")
	os.WriteFile(invalidHome, []byte("test"), 0644) // Create a file, not a directory
	
	// Set invalid home
	if runtime.GOOS == "windows" {
		os.Setenv("USERPROFILE", invalidHome)
	} else {
		os.Setenv("HOME", invalidHome)
	}
	
	defer func() {
		if runtime.GOOS == "windows" {
			os.Setenv("USERPROFILE", oldHome)
		} else {
			os.Setenv("HOME", oldHome)
		}
	}()
	
	// This should handle the error gracefully
	_, err := GetTempFile("test")
	
	// Should return an error
	if err == nil {
		t.Error("Expected error when home directory setup fails")
	}
}