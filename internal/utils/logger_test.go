package utils

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestInitLogger(t *testing.T) {
	// Reset logger for testing
	logger = nil
	loggerOnce = sync.Once{}
	
	tests := []struct {
		name      string
		config    *LogConfig
		wantError bool
	}{
		{
			name: "default config",
			config: &LogConfig{
				Enabled: false,
				Level:   "info",
				Format:  "text",
			},
			wantError: false,
		},
		{
			name: "json format",
			config: &LogConfig{
				Enabled: false,
				Level:   "debug",
				Format:  "json",
			},
			wantError: false,
		},
		{
			name: "invalid log level",
			config: &LogConfig{
				Enabled: false,
				Level:   "invalid",
				Format:  "text",
			},
			wantError: false, // Should default to info
		},
		{
			name: "file logging",
			config: &LogConfig{
				Enabled:  true,
				FilePath: filepath.Join(os.TempDir(), "test.log"),
				Level:    "info",
				Format:   "text",
			},
			wantError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset logger for each test
			logger = nil
			loggerOnce = sync.Once{}
			
			err := InitLogger(tt.config)
			if (err != nil) != tt.wantError {
				t.Errorf("InitLogger() error = %v, wantError %v", err, tt.wantError)
			}
			
			// Clean up log file if created
			if tt.config.Enabled && tt.config.FilePath != "" {
				os.Remove(tt.config.FilePath)
			}
			
			// Close logger after test
			CloseLogger()
		})
	}
}

func TestGetLogger(t *testing.T) {
	// Reset logger
	logger = nil
	loggerOnce = sync.Once{}
	
	// Test getting logger without initialization
	log := GetLogger()
	if log == nil {
		t.Error("Expected non-nil logger")
	}
	
	// Test getting logger again (should return same instance)
	log2 := GetLogger()
	if log != log2 {
		t.Error("Expected same logger instance")
	}
}

func TestCloseLogger(t *testing.T) {
	// Test closing without log file
	logger = nil
	loggerOnce = sync.Once{}
	logFile = nil
	
	err := CloseLogger()
	if err != nil {
		t.Errorf("CloseLogger() unexpected error: %v", err)
	}
	
	// Test closing with log file
	tmpFile := filepath.Join(os.TempDir(), "test_close.log")
	InitLogger(&LogConfig{
		Enabled:  true,
		FilePath: tmpFile,
		Level:    "info",
		Format:   "text",
	})
	
	err = CloseLogger()
	if err != nil {
		t.Errorf("CloseLogger() unexpected error: %v", err)
	}
	
	// Clean up
	os.Remove(tmpFile)
}

func TestRotateLogFile(t *testing.T) {
	// Test without log file
	logger = nil
	loggerOnce = sync.Once{}
	logFile = nil
	
	err := RotateLogFile(1024)
	if err != nil {
		t.Errorf("RotateLogFile() without file should not error: %v", err)
	}
	
	// Test with log file
	tmpFile := filepath.Join(os.TempDir(), "test_rotate.log")
	InitLogger(&LogConfig{
		Enabled:  true,
		FilePath: tmpFile,
		Level:    "info",
		Format:   "text",
	})
	
	// Write some data to exceed size limit
	log := GetLogger()
	for i := 0; i < 100; i++ {
		log.Info("Test log entry for rotation")
	}
	
	// Rotate with small size to trigger rotation
	err = RotateLogFile(100)
	if err != nil {
		t.Errorf("RotateLogFile() unexpected error: %v", err)
	}
	
	// Check if backup file was created
	files, _ := filepath.Glob(tmpFile + ".*")
	if len(files) == 0 {
		t.Error("Expected backup file to be created")
	}
	
	// Clean up
	CloseLogger()
	os.Remove(tmpFile)
	for _, f := range files {
		os.Remove(f)
	}
}

func TestLogRequest(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	logger = logrus.New()
	logger.SetOutput(&buf)
	logger.SetLevel(logrus.InfoLevel)
	
	LogRequest("GET", "/test", map[string]interface{}{
		"user": "test",
	})
	
	output := buf.String()
	if !strings.Contains(output, "HTTP request received") {
		t.Error("Expected log message not found")
	}
	if !strings.Contains(output, "method=GET") {
		t.Error("Expected method field not found")
	}
	if !strings.Contains(output, "path=/test") {
		t.Error("Expected path field not found")
	}
}

func TestLogResponse(t *testing.T) {
	// Test success response
	var buf bytes.Buffer
	logger = logrus.New()
	logger.SetOutput(&buf)
	logger.SetLevel(logrus.InfoLevel)
	
	LogResponse(200, 1.23, map[string]interface{}{
		"bytes": 1024,
	})
	
	output := buf.String()
	if !strings.Contains(output, "HTTP response sent") {
		t.Error("Expected success log message not found")
	}
	
	// Test error response
	buf.Reset()
	LogResponse(500, 0.5, nil)
	
	output = buf.String()
	if !strings.Contains(output, "HTTP response error") {
		t.Error("Expected error log message not found")
	}
}

func TestLogRouting(t *testing.T) {
	var buf bytes.Buffer
	logger = logrus.New()
	logger.SetOutput(&buf)
	logger.SetLevel(logrus.InfoLevel)
	
	LogRouting("gpt-4", "openai", "token limit")
	
	output := buf.String()
	if !strings.Contains(output, "Routing decision made") {
		t.Error("Expected log message not found")
	}
	if !strings.Contains(output, "model=gpt-4") {
		t.Error("Expected model field not found")
	}
}

func TestLogTransformer(t *testing.T) {
	// Test success
	var buf bytes.Buffer
	logger = logrus.New()
	logger.SetOutput(&buf)
	logger.SetLevel(logrus.DebugLevel)
	
	LogTransformer("test", "request", true, nil)
	
	output := buf.String()
	if !strings.Contains(output, "Transformer applied") {
		t.Error("Expected success log message not found")
	}
	
	// Test error
	buf.Reset()
	logger.SetLevel(logrus.ErrorLevel)
	LogTransformer("test", "response", false, io.EOF)
	
	output = buf.String()
	if !strings.Contains(output, "Transformer error") {
		t.Error("Expected error log message not found")
	}
}

func TestLogStartup(t *testing.T) {
	var buf bytes.Buffer
	logger = logrus.New()
	logger.SetOutput(&buf)
	logger.SetLevel(logrus.InfoLevel)
	
	LogStartup(8080, "1.0.0")
	
	output := buf.String()
	if !strings.Contains(output, "CCProxy started") {
		t.Error("Expected startup log message not found")
	}
	if !strings.Contains(output, "port=8080") {
		t.Error("Expected port field not found")
	}
	if !strings.Contains(output, "version=1.0.0") {
		t.Error("Expected version field not found")
	}
}

func TestLogShutdown(t *testing.T) {
	var buf bytes.Buffer
	logger = logrus.New()
	logger.SetOutput(&buf)
	logger.SetLevel(logrus.InfoLevel)
	
	LogShutdown("signal received")
	
	output := buf.String()
	if !strings.Contains(output, "CCProxy shutting down") {
		t.Error("Expected shutdown log message not found")
	}
	if !strings.Contains(output, "reason=") {
		t.Error("Expected reason field not found")
	}
}

// Test concurrent access to logger
func TestLoggerConcurrency(t *testing.T) {
	// Reset logger
	logger = nil
	loggerOnce = sync.Once{}
	
	var wg sync.WaitGroup
	errors := make(chan error, 10)
	
	// Start multiple goroutines accessing logger
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			log := GetLogger()
			if log == nil {
				errors <- io.EOF
				return
			}
			
			// Perform some logging
			LogRequest("GET", "/test", nil)
			LogResponse(200, 1.0, nil)
			LogRouting("model", "provider", "reason")
		}(i)
	}
	
	wg.Wait()
	close(errors)
	
	// Check for errors
	for err := range errors {
		if err != nil {
			t.Errorf("Concurrent access error: %v", err)
		}
	}
}

// Test log rotation with concurrent writes
func TestRotateLogFileConcurrent(t *testing.T) {
	tmpFile := filepath.Join(os.TempDir(), "test_concurrent_rotate.log")
	
	// Initialize logger with file
	logger = nil
	loggerOnce = sync.Once{}
	InitLogger(&LogConfig{
		Enabled:  true,
		FilePath: tmpFile,
		Level:    "info",
		Format:   "text",
	})
	
	var wg sync.WaitGroup
	
	// Start writing logs
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			LogRequest("GET", "/test", nil)
			time.Sleep(time.Millisecond)
		}
	}()
	
	// Start rotation attempts
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			RotateLogFile(1024)
			time.Sleep(10 * time.Millisecond)
		}
	}()
	
	wg.Wait()
	
	// Clean up
	CloseLogger()
	os.Remove(tmpFile)
	files, _ := filepath.Glob(tmpFile + ".*")
	for _, f := range files {
		os.Remove(f)
	}
}