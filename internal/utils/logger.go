package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	logger     *logrus.Logger
	loggerOnce sync.Once
	logFile    *os.File
	logMutex   sync.Mutex
)

// LogConfig represents logging configuration
type LogConfig struct {
	Enabled  bool
	FilePath string
	Level    string
	Format   string // "json" or "text"
}

// InitLogger initializes the global logger
func InitLogger(config *LogConfig) error {
	var err error
	
	loggerOnce.Do(func() {
		logger = logrus.New()
		
		// Set log level
		level, parseErr := logrus.ParseLevel(config.Level)
		if parseErr != nil {
			level = logrus.InfoLevel
		}
		logger.SetLevel(level)
		
		// Set formatter
		if config.Format == "json" {
			logger.SetFormatter(&logrus.JSONFormatter{
				TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
			})
		} else {
			logger.SetFormatter(&logrus.TextFormatter{
				TimestampFormat: "2006-01-02 15:04:05",
				FullTimestamp:   true,
			})
		}
		
		// Configure output
		if config.Enabled && config.FilePath != "" {
			// Resolve log file path
			logPath, resolveErr := ResolvePath(config.FilePath)
			if resolveErr != nil {
				err = fmt.Errorf("failed to resolve log path: %w", resolveErr)
				return
			}
			
			// Ensure log directory exists
			logDir := filepath.Dir(logPath)
			if mkdirErr := os.MkdirAll(logDir, 0755); mkdirErr != nil {
				err = fmt.Errorf("failed to create log directory: %w", mkdirErr)
				return
			}
			
			// Open log file
			logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				err = fmt.Errorf("failed to open log file: %w", err)
				return
			}
			
			// Set multi-writer for both file and stdout
			multiWriter := io.MultiWriter(os.Stdout, logFile)
			logger.SetOutput(multiWriter)
		} else {
			// Log to stdout only
			logger.SetOutput(os.Stdout)
		}
	})
	
	return err
}

// GetLogger returns the global logger instance
func GetLogger() *logrus.Logger {
	if logger == nil {
		// Initialize with defaults if not initialized
		InitLogger(&LogConfig{
			Enabled: false,
			Level:   "info",
			Format:  "text",
		})
	}
	return logger
}

// CloseLogger closes the log file if open
func CloseLogger() error {
	logMutex.Lock()
	defer logMutex.Unlock()
	
	if logFile != nil {
		err := logFile.Close()
		logFile = nil
		return err
	}
	return nil
}

// RotateLogFile rotates the log file
func RotateLogFile(maxSize int64) error {
	logMutex.Lock()
	defer logMutex.Unlock()
	
	if logFile == nil {
		return nil
	}
	
	// Get file info
	info, err := logFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat log file: %w", err)
	}
	
	// Check if rotation needed
	if info.Size() < maxSize {
		return nil
	}
	
	// Get file path
	logPath := logFile.Name()
	
	// Close current file
	if err := logFile.Close(); err != nil {
		return fmt.Errorf("failed to close log file: %w", err)
	}
	
	// Rename to backup
	backupPath := fmt.Sprintf("%s.%d", logPath, info.ModTime().Unix())
	if err := os.Rename(logPath, backupPath); err != nil {
		return fmt.Errorf("failed to rename log file: %w", err)
	}
	
	// Open new file
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open new log file: %w", err)
	}
	
	// Update logger output
	if logger != nil {
		multiWriter := io.MultiWriter(os.Stdout, logFile)
		logger.SetOutput(multiWriter)
	}
	
	return nil
}

// Helper functions for common logging patterns

// LogRequest logs an HTTP request
func LogRequest(method, path string, fields map[string]interface{}) {
	log := GetLogger().WithFields(logrus.Fields{
		"method": method,
		"path":   path,
		"type":   "request",
	})
	
	for k, v := range fields {
		log = log.WithField(k, v)
	}
	
	log.Info("HTTP request received")
}

// LogResponse logs an HTTP response
func LogResponse(statusCode int, duration float64, fields map[string]interface{}) {
	log := GetLogger().WithFields(logrus.Fields{
		"status":   statusCode,
		"duration": duration,
		"type":     "response",
	})
	
	for k, v := range fields {
		log = log.WithField(k, v)
	}
	
	if statusCode >= 400 {
		log.Error("HTTP response error")
	} else {
		log.Info("HTTP response sent")
	}
}

// LogRouting logs routing decisions
func LogRouting(model, provider, reason string) {
	GetLogger().WithFields(logrus.Fields{
		"model":    model,
		"provider": provider,
		"reason":   reason,
		"type":     "routing",
	}).Info("Routing decision made")
}

// LogTransformer logs transformer operations
func LogTransformer(name, direction string, success bool, err error) {
	log := GetLogger().WithFields(logrus.Fields{
		"transformer": name,
		"direction":   direction,
		"success":     success,
		"type":        "transformer",
	})
	
	if err != nil {
		log.WithError(err).Error("Transformer error")
	} else if success {
		log.Debug("Transformer applied")
	}
}

// LogStartup logs service startup
func LogStartup(port int, version string) {
	GetLogger().WithFields(logrus.Fields{
		"port":    port,
		"version": version,
		"type":    "startup",
	}).Info("CCProxy started")
}

// LogShutdown logs service shutdown
func LogShutdown(reason string) {
	GetLogger().WithFields(logrus.Fields{
		"reason": reason,
		"type":   "shutdown",
	}).Info("CCProxy shutting down")
}