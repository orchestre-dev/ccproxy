package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// HomeDir represents the CCProxy home directory structure
type HomeDir struct {
	Root       string
	ConfigPath string
	LogPath    string
	PIDPath    string
	PluginsDir string
	TempDir    string
}

// GetHomeDir returns the CCProxy home directory path
func GetHomeDir() (string, error) {
	// Get user home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Return ~/.ccproxy path
	return filepath.Join(home, ".ccproxy"), nil
}

// InitializeHomeDir creates and initializes the CCProxy home directory structure
func InitializeHomeDir() (*HomeDir, error) {
	// Get home directory path
	rootDir, err := GetHomeDir()
	if err != nil {
		return nil, err
	}

	// Create home directory structure
	homeDir := &HomeDir{
		Root:       rootDir,
		ConfigPath: filepath.Join(rootDir, "config.json"),
		LogPath:    filepath.Join(rootDir, "ccproxy.log"),
		PIDPath:    filepath.Join(rootDir, ".ccproxy.pid"),
		PluginsDir: filepath.Join(rootDir, "plugins"),
		TempDir:    filepath.Join(rootDir, "tmp"),
	}

	// Create directories with appropriate permissions
	directories := []string{
		homeDir.Root,
		homeDir.PluginsDir,
		homeDir.TempDir,
	}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return homeDir, nil
}

// ResolvePath resolves a path, handling both absolute and relative paths
func ResolvePath(path string) (string, error) {
	// If path is empty, return error
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	// If path is already absolute, clean and return it
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}

	// For relative paths, resolve relative to current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	return filepath.Clean(filepath.Join(cwd, path)), nil
}

// ResolveConfigPath resolves a configuration file path
func ResolveConfigPath(path string) (string, error) {
	// If path is empty, use default config path
	if path == "" {
		homeDir, err := InitializeHomeDir()
		if err != nil {
			return "", err
		}
		return homeDir.ConfigPath, nil
	}

	// Resolve the provided path
	return ResolvePath(path)
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// EnsureDir ensures a directory exists, creating it if necessary
func EnsureDir(path string) error {
	if DirExists(path) {
		return nil
	}
	return os.MkdirAll(path, 0755)
}

// GetTempFile returns a path for a temporary file in the CCProxy temp directory
func GetTempFile(prefix string) (string, error) {
	homeDir, err := InitializeHomeDir()
	if err != nil {
		return "", err
	}

	// Create temp file path
	filename := fmt.Sprintf("%s_%d", prefix, os.Getpid())
	return filepath.Join(homeDir.TempDir, filename), nil
}

// CleanupTempFiles removes old temporary files from the temp directory
func CleanupTempFiles() error {
	homeDir, err := InitializeHomeDir()
	if err != nil {
		return err
	}

	// Read temp directory
	entries, err := os.ReadDir(homeDir.TempDir)
	if err != nil {
		// If directory doesn't exist, that's fine
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read temp directory: %w", err)
	}

	// Remove all files in temp directory
	for _, entry := range entries {
		if !entry.IsDir() {
			path := filepath.Join(homeDir.TempDir, entry.Name())
			if err := os.Remove(path); err != nil {
				// Log but don't fail on individual file errors
				continue
			}
		}
	}

	return nil
}

// GetExecutablePath returns the path to the current executable
func GetExecutablePath() (string, error) {
	if runtime.GOOS == "windows" {
		return os.Executable()
	}

	// For Unix-like systems, use readlink on /proc/self/exe if available
	if runtime.GOOS == "linux" {
		if path, err := os.Readlink("/proc/self/exe"); err == nil {
			return path, nil
		}
	}

	// Fallback to os.Executable
	return os.Executable()
}

// GetReferenceCountPath returns the path to the reference count file
func GetReferenceCountPath() (string, error) {
	// Use system temp directory for reference counting
	tempDir := os.TempDir()
	return filepath.Join(tempDir, "ccproxy_ref_count"), nil
}

// WriteFileAtomic writes data to a file atomically
func WriteFileAtomic(path string, data []byte, perm os.FileMode) error {
	// Create temp file in same directory
	dir := filepath.Dir(path)
	tempFile, err := os.CreateTemp(dir, ".tmp_")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()

	// Clean up on error
	defer func() {
		if tempFile != nil {
			_ = tempFile.Close()    // Best effort cleanup
			_ = os.Remove(tempPath) // Best effort cleanup
		}
	}()

	// Write data
	if _, err := tempFile.Write(data); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	// Sync to disk
	if err := tempFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	// Close file
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}
	tempFile = nil

	// Set permissions
	if err := os.Chmod(tempPath, perm); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}
