// Package common provides shared utilities for providers
package common

import (
	"context"
	"net/http"
	"time"

	"ccproxy/internal/constants"
)

// Default HTTP client configuration values
const (
	DefaultMaxIdleConns        = 100
	DefaultMaxIdleConnsPerHost = 10
	DefaultMaxConnsPerHost     = 10
	DefaultIdleConnTimeout     = 90 * time.Second
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
			MaxIdleConns:        DefaultMaxIdleConns,
			MaxIdleConnsPerHost: DefaultMaxIdleConnsPerHost,
			MaxConnsPerHost:     DefaultMaxConnsPerHost,
			IdleConnTimeout:     DefaultIdleConnTimeout,
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
func GetRequestID(ctx context.Context) string {
	if requestID := ctx.Value(constants.RequestIDKey); requestID != nil {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return constants.DefaultRequestID
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
