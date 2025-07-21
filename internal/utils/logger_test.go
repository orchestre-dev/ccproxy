package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
	"github.com/sirupsen/logrus"
)

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name     string
		config   *LogConfig
		wantErr  bool
		validate func(*testing.T, *LogConfig)
	}{
		{
			name: "default text logger",
			config: &LogConfig{
				Enabled: false,
				Level:   "info",
				Format:  "text",
			},
			wantErr: false,
			validate: func(t *testing.T, config *LogConfig) {
				logger := GetLogger()
				testutil.AssertNotEqual(t, nil, logger, "Logger should not be nil")
				testutil.AssertEqual(t, logrus.InfoLevel, logger.GetLevel(), "Log level should be info")

				// Check formatter type
				_, isTextFormatter := logger.Formatter.(*logrus.TextFormatter)
				testutil.AssertTrue(t, isTextFormatter, "Should use text formatter")
			},
		},
		{
			name: "json logger with debug level",
			config: &LogConfig{
				Enabled: false,
				Level:   "debug",
				Format:  "json",
			},
			wantErr: false,
			validate: func(t *testing.T, config *LogConfig) {
				logger := GetLogger()
				testutil.AssertEqual(t, logrus.DebugLevel, logger.GetLevel(), "Log level should be debug")

				// Check formatter type
				_, isJSONFormatter := logger.Formatter.(*logrus.JSONFormatter)
				testutil.AssertTrue(t, isJSONFormatter, "Should use JSON formatter")
			},
		},
		{
			name: "file logger",
			config: &LogConfig{
				Enabled:  true,
				FilePath: "test.log",
				Level:    "warn",
				Format:   "text",
			},
			wantErr: false,
			validate: func(t *testing.T, config *LogConfig) {
				logger := GetLogger()
				testutil.AssertEqual(t, logrus.WarnLevel, logger.GetLevel(), "Log level should be warn")

				// Clean up log file
				defer func() {
					CloseLogger()
					resolvedPath, _ := ResolvePath(config.FilePath)
					os.Remove(resolvedPath)
				}()
			},
		},
		{
			name: "invalid log level defaults to info",
			config: &LogConfig{
				Enabled: false,
				Level:   "invalid",
				Format:  "text",
			},
			wantErr: false,
			validate: func(t *testing.T, config *LogConfig) {
				logger := GetLogger()
				testutil.AssertEqual(t, logrus.InfoLevel, logger.GetLevel(), "Invalid level should default to info")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset logger for each test
			logger = nil
			loggerOnce = *new(sync.Once)

			err := InitLogger(tt.config)
			if tt.wantErr {
				testutil.AssertError(t, err, "Expected error during initialization")
			} else {
				testutil.AssertNoError(t, err, "Should not error during initialization")
				if tt.validate != nil {
					tt.validate(t, tt.config)
				}
			}
		})
	}
}

func TestInitLoggerWithFile(t *testing.T) {
	tempConfig := testutil.SetupTest(t)
	logPath := filepath.Join(tempConfig.TempDir, "test.log")

	config := &LogConfig{
		Enabled:  true,
		FilePath: logPath,
		Level:    "info",
		Format:   "text",
	}

	// Reset logger
	logger = nil
	loggerOnce = *new(sync.Once)

	err := InitLogger(config)
	testutil.AssertNoError(t, err, "Should initialize logger with file")

	// Verify file was created
	testutil.AssertTrue(t, FileExists(logPath), "Log file should exist")

	// Test logging to file
	log := GetLogger()
	log.Info("Test message")

	// Close logger to flush
	CloseLogger()

	// Read and verify file content
	content, err := os.ReadFile(logPath)
	testutil.AssertNoError(t, err, "Should read log file")
	testutil.AssertContains(t, string(content), "Test message", "Log file should contain test message")
}

func TestGetLogger(t *testing.T) {
	// Reset logger
	logger = nil
	loggerOnce = *new(sync.Once)

	// Should initialize with defaults if not initialized
	log1 := GetLogger()
	testutil.AssertNotEqual(t, nil, log1, "Should return non-nil logger")

	// Should return same instance
	log2 := GetLogger()
	testutil.AssertEqual(t, log1, log2, "Should return same logger instance")
}

func TestCloseLogger(t *testing.T) {
	tempConfig := testutil.SetupTest(t)
	logPath := filepath.Join(tempConfig.TempDir, "test.log")

	// Reset logger
	logger = nil
	loggerOnce = *new(sync.Once)

	config := &LogConfig{
		Enabled:  true,
		FilePath: logPath,
		Level:    "info",
		Format:   "text",
	}

	err := InitLogger(config)
	testutil.AssertNoError(t, err, "Should initialize logger")

	// Close logger
	err = CloseLogger()
	testutil.AssertNoError(t, err, "Should close logger without error")

	// Closing again should not error
	err = CloseLogger()
	testutil.AssertNoError(t, err, "Should handle multiple close calls")
}

func TestRotateLogFile(t *testing.T) {
	tempConfig := testutil.SetupTest(t)
	logPath := filepath.Join(tempConfig.TempDir, "test.log")

	// Reset logger
	logger = nil
	loggerOnce = *new(sync.Once)

	config := &LogConfig{
		Enabled:  true,
		FilePath: logPath,
		Level:    "info",
		Format:   "text",
	}

	err := InitLogger(config)
	testutil.AssertNoError(t, err, "Should initialize logger")

	// Write some data to log file
	log := GetLogger()
	for i := 0; i < 100; i++ {
		log.Infof("Test log entry %d with some content to make it longer", i)
	}

	// Test rotation with small max size (should trigger rotation)
	err = RotateLogFile(10) // 10 bytes - very small to force rotation
	testutil.AssertNoError(t, err, "Should rotate log file")

	// Original file should exist and be small
	testutil.AssertTrue(t, FileExists(logPath), "New log file should exist")

	// Test rotation when no rotation needed
	err = RotateLogFile(10000) // Large size - no rotation needed
	testutil.AssertNoError(t, err, "Should handle no rotation case")

	CloseLogger()
}

func TestRotateLogFileWithoutLogger(t *testing.T) {
	// Reset logger
	logger = nil
	loggerOnce = *new(sync.Once)
	logFile = nil

	// Should not error when no log file is open
	err := RotateLogFile(1000)
	testutil.AssertNoError(t, err, "Should handle rotation without open log file")
}

func TestLogRequest(t *testing.T) {
	// Reset logger
	logger = nil
	loggerOnce = *new(sync.Once)

	// Initialize with default config
	config := &LogConfig{
		Enabled: false,
		Level:   "debug",
		Format:  "text",
	}

	err := InitLogger(config)
	testutil.AssertNoError(t, err, "Should initialize logger")

	// Test basic request logging
	LogRequest("GET", "/api/test", nil)

	// Test request logging with fields
	fields := map[string]interface{}{
		"user_id": "123",
		"ip":      "192.168.1.1",
	}
	LogRequest("POST", "/api/users", fields)

	// No errors expected - this is primarily testing that functions don't panic
}

func TestLogResponse(t *testing.T) {
	// Reset logger
	logger = nil
	loggerOnce = *new(sync.Once)

	config := &LogConfig{
		Enabled: false,
		Level:   "debug",
		Format:  "text",
	}

	err := InitLogger(config)
	testutil.AssertNoError(t, err, "Should initialize logger")

	// Test success response
	LogResponse(200, 0.123, map[string]interface{}{
		"bytes": 1024,
	})

	// Test error response
	LogResponse(500, 1.234, map[string]interface{}{
		"error": "internal server error",
	})

	// Test with nil fields
	LogResponse(404, 0.045, nil)
}

func TestLogRouting(t *testing.T) {
	// Reset logger
	logger = nil
	loggerOnce = *new(sync.Once)

	config := &LogConfig{
		Enabled: false,
		Level:   "debug",
		Format:  "text",
	}

	err := InitLogger(config)
	testutil.AssertNoError(t, err, "Should initialize logger")

	LogRouting("gpt-4", "openai", "fastest response time")
	LogRouting("claude-3", "anthropic", "load balancing")
}

func TestLogTransformer(t *testing.T) {
	// Reset logger
	logger = nil
	loggerOnce = *new(sync.Once)

	config := &LogConfig{
		Enabled: false,
		Level:   "debug",
		Format:  "text",
	}

	err := InitLogger(config)
	testutil.AssertNoError(t, err, "Should initialize logger")

	// Test successful transformation
	LogTransformer("anthropic-to-openai", "request", true, nil)

	// Test failed transformation
	LogTransformer("openai-to-anthropic", "response", false,
		fmt.Errorf("transformation failed"))

	// Test with error but marked as success (edge case)
	LogTransformer("test-transformer", "request", true,
		fmt.Errorf("warning error"))
}

func TestLogStartup(t *testing.T) {
	// Reset logger
	logger = nil
	loggerOnce = *new(sync.Once)

	config := &LogConfig{
		Enabled: false,
		Level:   "info",
		Format:  "text",
	}

	err := InitLogger(config)
	testutil.AssertNoError(t, err, "Should initialize logger")

	LogStartup(8080, "1.0.0")
	LogStartup(3456, "dev")
}

func TestLogShutdown(t *testing.T) {
	// Reset logger
	logger = nil
	loggerOnce = *new(sync.Once)

	config := &LogConfig{
		Enabled: false,
		Level:   "info",
		Format:  "text",
	}

	err := InitLogger(config)
	testutil.AssertNoError(t, err, "Should initialize logger")

	LogShutdown("SIGTERM received")
	LogShutdown("user requested")
}

func TestLoggerConcurrency(t *testing.T) {
	// Reset logger
	logger = nil
	loggerOnce = *new(sync.Once)

	config := &LogConfig{
		Enabled: false,
		Level:   "debug",
		Format:  "text",
	}

	err := InitLogger(config)
	testutil.AssertNoError(t, err, "Should initialize logger")

	// Test concurrent access to logger
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			log := GetLogger()
			log.WithField("goroutine", id).Info("Concurrent log message")

			LogRequest("GET", fmt.Sprintf("/api/test/%d", id),
				map[string]interface{}{"goroutine": id})

			LogResponse(200, 0.1,
				map[string]interface{}{"goroutine": id})
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent logging")
		}
	}
}

func TestLoggerWithInvalidPaths(t *testing.T) {
	// Reset logger
	logger = nil
	loggerOnce = *new(sync.Once)

	// Test with invalid directory
	config := &LogConfig{
		Enabled:  true,
		FilePath: "/root/invalid/path/test.log", // Should fail on most systems
		Level:    "info",
		Format:   "text",
	}

	err := InitLogger(config)
	// May or may not error depending on system permissions
	// Just ensure it doesn't panic
	if err != nil {
		testutil.AssertContains(t, err.Error(), "failed", "Error should mention failure")
	}
}

func TestInitLoggerErrorCases(t *testing.T) {
	tests := []struct {
		name    string
		config  *LogConfig
		wantErr bool
	}{
		{
			name: "empty file path falls back gracefully",
			config: &LogConfig{
				Enabled:  true,
				FilePath: "", // Empty path should fallback to stdout only (no error)
				Level:    "info",
				Format:   "text",
			},
			wantErr: false,
		},
		{
			name: "invalid file creation",
			config: &LogConfig{
				Enabled:  true,
				FilePath: "/proc/sys/kernel/hostname/invalid.log", // Should fail to create
				Level:    "info",
				Format:   "text",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset logger
			logger = nil
			loggerOnce = *new(sync.Once)

			err := InitLogger(tt.config)
			if tt.wantErr {
				testutil.AssertError(t, err, "Expected error for invalid config")
			} else {
				testutil.AssertNoError(t, err, "Should not error for valid config")
			}
		})
	}
}

func TestRotateLogFileErrorCases(t *testing.T) {
	tempConfig := testutil.SetupTest(t)
	logPath := filepath.Join(tempConfig.TempDir, "rotate_test.log")

	// Reset logger and create with file
	logger = nil
	loggerOnce = *new(sync.Once)

	config := &LogConfig{
		Enabled:  true,
		FilePath: logPath,
		Level:    "info",
		Format:   "text",
	}

	err := InitLogger(config)
	testutil.AssertNoError(t, err, "Should initialize logger")

	// Write some data
	log := GetLogger()
	for i := 0; i < 10; i++ {
		log.Infof("Test entry %d", i)
	}

	// Test rotation with error during rename
	// This is hard to test without extensive mocking, but we can test other paths

	// Test rotation when current file is small (no rotation needed)
	err = RotateLogFile(1000000) // Large size
	testutil.AssertNoError(t, err, "Should not rotate when file is small")

	// Force rotation by using small max size
	err = RotateLogFile(1)
	testutil.AssertNoError(t, err, "Should rotate when file exceeds size")

	// Verify backup file was created
	files, err := filepath.Glob(logPath + ".*")
	testutil.AssertNoError(t, err, "Should list backup files")
	testutil.AssertTrue(t, len(files) >= 1, "Should have at least one backup file")

	CloseLogger()
}

func TestLoggerInitializationOnce(t *testing.T) {
	// Test that InitLogger only initializes once using sync.Once

	// Reset logger
	logger = nil
	loggerOnce = *new(sync.Once)

	config := &LogConfig{
		Enabled: false,
		Level:   "debug",
		Format:  "text",
	}

	// Initialize multiple times concurrently
	done := make(chan error, 5)

	for i := 0; i < 5; i++ {
		go func() {
			err := InitLogger(config)
			done <- err
		}()
	}

	// Wait for all initializations
	for i := 0; i < 5; i++ {
		select {
		case err := <-done:
			testutil.AssertNoError(t, err, "Should not error during concurrent initialization")
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for initialization")
		}
	}

	// Verify logger was initialized
	log := GetLogger()
	testutil.AssertNotEqual(t, nil, log, "Logger should be initialized")
	testutil.AssertEqual(t, logrus.DebugLevel, log.GetLevel(), "Should have debug level")
}

func TestGetLoggerDefaultInitialization(t *testing.T) {
	// Reset logger to test default initialization
	logger = nil
	loggerOnce = *new(sync.Once)

	// GetLogger should initialize with defaults if not initialized
	log := GetLogger()
	testutil.AssertNotEqual(t, nil, log, "Should return initialized logger")

	// Should have default settings
	testutil.AssertEqual(t, logrus.InfoLevel, log.GetLevel(), "Should have default info level")

	// Formatter should be text
	_, isTextFormatter := log.Formatter.(*logrus.TextFormatter)
	testutil.AssertTrue(t, isTextFormatter, "Should use default text formatter")
}

func TestCloseLoggerMultipleTimes(t *testing.T) {
	tempConfig := testutil.SetupTest(t)
	logPath := filepath.Join(tempConfig.TempDir, "close_test.log")

	// Reset logger
	logger = nil
	loggerOnce = *new(sync.Once)

	config := &LogConfig{
		Enabled:  true,
		FilePath: logPath,
		Level:    "info",
		Format:   "text",
	}

	err := InitLogger(config)
	testutil.AssertNoError(t, err, "Should initialize logger")

	// Close multiple times
	err = CloseLogger()
	testutil.AssertNoError(t, err, "Should close logger first time")

	err = CloseLogger()
	testutil.AssertNoError(t, err, "Should handle second close call")

	err = CloseLogger()
	testutil.AssertNoError(t, err, "Should handle third close call")
}

func TestRotateLogFileDetailedCoverage(t *testing.T) {
	tempConfig := testutil.SetupTest(t)
	logPath := filepath.Join(tempConfig.TempDir, "detailed_rotate.log")

	// Reset logger
	logger = nil
	loggerOnce = *new(sync.Once)
	logFile = nil

	// Test rotation with no logger/file
	err := RotateLogFile(100)
	testutil.AssertNoError(t, err, "Should handle no log file gracefully")

	// Initialize logger with file
	config := &LogConfig{
		Enabled:  true,
		FilePath: logPath,
		Level:    "info",
		Format:   "text",
	}

	err = InitLogger(config)
	testutil.AssertNoError(t, err, "Should initialize logger")

	// Write substantial content to ensure file size > 0
	log := GetLogger()
	content := "This is a long log message with substantial content to ensure the file has a meaningful size for rotation testing. "
	for i := 0; i < 50; i++ {
		log.Info(content + fmt.Sprintf("Entry %d", i))
	}

	// Verify the file has content
	info, err := os.Stat(logPath)
	testutil.AssertNoError(t, err, "Should get file info")
	testutil.AssertTrue(t, info.Size() > 100, "File should have substantial content")

	// Test rotation with size less than current file size
	err = RotateLogFile(50) // Smaller than current file
	testutil.AssertNoError(t, err, "Should rotate file when size exceeded")

	// Verify backup file was created
	files, err := filepath.Glob(logPath + ".*")
	testutil.AssertNoError(t, err, "Should list files")
	testutil.AssertTrue(t, len(files) > 0, "Should have backup file")

	// Verify new file is smaller
	newInfo, err := os.Stat(logPath)
	testutil.AssertNoError(t, err, "Should get new file info")
	testutil.AssertTrue(t, newInfo.Size() < info.Size(), "New file should be smaller")

	CloseLogger()
}

// Benchmark tests for performance verification
func BenchmarkLogRequest(b *testing.B) {
	// Reset logger
	logger = nil
	loggerOnce = *new(sync.Once)

	config := &LogConfig{
		Enabled: false,
		Level:   "info",
		Format:  "text",
	}

	InitLogger(config)

	fields := map[string]interface{}{
		"user_id": "12345",
		"ip":      "192.168.1.1",
		"method":  "POST",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LogRequest("POST", "/api/benchmark", fields)
	}
}

func BenchmarkLogResponse(b *testing.B) {
	// Reset logger
	logger = nil
	loggerOnce = *new(sync.Once)

	config := &LogConfig{
		Enabled: false,
		Level:   "info",
		Format:  "text",
	}

	InitLogger(config)

	fields := map[string]interface{}{
		"bytes":    1024,
		"encoding": "gzip",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LogResponse(200, 0.123, fields)
	}
}
