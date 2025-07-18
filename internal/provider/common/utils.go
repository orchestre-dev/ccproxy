// Package common provides shared utilities for providers
package common

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ccproxy/pkg/logger"
)

// TokenLimitConfig holds token limit configuration for providers
type TokenLimitConfig struct {
	ProviderName  string
	MaxTokens     int
	ProviderLimit int
}

// ApplyTokenLimit applies token limits to a request
func ApplyTokenLimit(req interface{}, config TokenLimitConfig, logger *logger.Logger) {
	// This is a generic implementation - specific providers would need
	// to implement their own token limiting logic
	logger.Debugf("Applying token limit for %s provider", config.ProviderName)
}

// SafeCloseResponse safely closes HTTP response body with error logging
func SafeCloseResponse(resp *http.Response, logger *logger.Logger) {
	if resp != nil && resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			logger.WithError(err).Warn("Failed to close response body")
		}
	}
}

// MarshalJSON efficiently marshals JSON without additional allocations
func MarshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// LogRequest logs HTTP request details
func LogRequest(url, method string, duration time.Duration, logger *logger.Logger) {
	logger.Debugf("HTTP %s %s completed in %v", method, url, duration)
}

// LogResponse logs HTTP response details
func LogResponse(statusCode int, duration time.Duration, logger *logger.Logger) {
	logger.Debugf("HTTP response %d completed in %v", statusCode, duration)
}

// CreateStandardHeaders creates standard HTTP headers for API requests
func CreateStandardHeaders(apiKey, userAgent string) map[string]string {
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["User-Agent"] = userAgent

	if apiKey != "" {
		headers["Authorization"] = "Bearer " + apiKey
	}

	return headers
}

// ValidateHTTPResponse validates HTTP response and returns error for non-2xx status
func ValidateHTTPResponse(resp *http.Response) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}
	return nil
}
