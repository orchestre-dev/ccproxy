package common

import (
	"net/http"
	"time"
)

// HTTPClientConfig holds configuration for HTTP client
type HTTPClientConfig struct {
	Timeout             time.Duration
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	MaxConnsPerHost     int
	IdleConnTimeout     time.Duration
}

// NewConfiguredHTTPClient creates a properly configured HTTP client
func NewConfiguredHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			MaxConnsPerHost:     10,
			IdleConnTimeout:     90 * time.Second,
			DisableKeepAlives:   false,
		},
	}
}

// NewConfiguredHTTPClientWithConfig creates an HTTP client with custom config
func NewConfiguredHTTPClientWithConfig(config HTTPClientConfig) *http.Client {
	return &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        config.MaxIdleConns,
			MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
			MaxConnsPerHost:     config.MaxConnsPerHost,
			IdleConnTimeout:     config.IdleConnTimeout,
			DisableKeepAlives:   false,
		},
	}
}

// GetRequestID extracts request ID from context or generates one
func GetRequestID(ctx interface{}) string {
	// This is a simple implementation - in production you might want
	// to use a proper request ID from context
	return "request-id-placeholder"
}

// GetFinishReason determines finish reason from response
func GetFinishReason(reason string) string {
	switch reason {
	case "stop", "end_turn":
		return "end_turn"
	case "length", "max_tokens":
		return "max_tokens"
	case "tool_use", "function_call":
		return "tool_use"
	default:
		return "stop"
	}
}