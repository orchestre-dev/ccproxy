package logger

import (
	"ccproxy/internal/config"

	"github.com/sirupsen/logrus"
)

// Logger wraps logrus.Logger with additional functionality
type Logger struct {
	*logrus.Logger
}

// New creates a new logger instance
func New(config config.LoggingConfig) *Logger {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		logger.Warnf("Invalid log level '%s', using 'info'", config.Level)
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set formatter
	switch config.Format {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	default:
		logger.Warnf("Invalid log format '%s', using 'json'", config.Format)
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	}

	return &Logger{Logger: logger}
}

// WithRequestID adds a request ID to the logger context
func (l *Logger) WithRequestID(requestID string) *logrus.Entry {
	return l.WithField("request_id", requestID)
}

// WithComponent adds a component name to the logger context
func (l *Logger) WithComponent(component string) *logrus.Entry {
	return l.WithField("component", component)
}

// WithError adds an error to the logger context
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithError(err)
}

// HTTPLog logs HTTP request/response information
func (l *Logger) HTTPLog(method, path string, statusCode int, duration int64, requestID string) {
	l.WithFields(logrus.Fields{
		"method":     method,
		"path":       path,
		"status":     statusCode,
		"duration":   duration,
		"request_id": requestID,
		"type":       "http_request",
	}).Info("HTTP request completed")
}

// APILog logs API-specific information
func (l *Logger) APILog(action string, details map[string]interface{}, requestID string) {
	fields := logrus.Fields{
		"action":     action,
		"request_id": requestID,
		"type":       "api_action",
	}
	
	for k, v := range details {
		fields[k] = v
	}
	
	l.WithFields(fields).Info("API action")
}