package common

import (
	"fmt"
	"net/http"
)

// Use standard HTTP status codes from the http package

// ProviderError represents a structured error from a provider
type ProviderError struct {
	Original  error
	Provider  string
	Message   string
	RequestID string
	Code      int
	Retryable bool
}

// Error implements the error interface
func (e *ProviderError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("%s provider error (request: %s): %s", e.Provider, e.RequestID, e.Message)
	}
	return fmt.Sprintf("%s provider error: %s", e.Provider, e.Message)
}

// Unwrap returns the original error
func (e *ProviderError) Unwrap() error {
	return e.Original
}

// IsRetryable returns whether the error is retryable
func (e *ProviderError) IsRetryable() bool {
	return e.Retryable
}

// NewProviderError creates a new provider error
func NewProviderError(provider, message string, original error) *ProviderError {
	return &ProviderError{
		Provider:  provider,
		Message:   message,
		Original:  original,
		Retryable: false,
	}
}

// NewHTTPError creates a provider error from HTTP response
func NewHTTPError(provider string, resp *http.Response, original error) *ProviderError {
	if resp == nil {
		return &ProviderError{
			Provider:  provider,
			Message:   "nil response",
			Original:  original,
			Retryable: false,
		}
	}

	retryable := isRetryableHTTPError(resp.StatusCode)

	return &ProviderError{
		Provider:  provider,
		Code:      resp.StatusCode,
		Message:   fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status),
		Original:  original,
		Retryable: retryable,
	}
}

// isRetryableHTTPError determines if an HTTP error is retryable
func isRetryableHTTPError(statusCode int) bool {
	switch statusCode {
	case http.StatusRequestTimeout, http.StatusTooManyRequests, http.StatusInternalServerError,
		http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

// NewConfigError creates a configuration error
func NewConfigError(provider, field, message string) *ProviderError {
	return &ProviderError{
		Provider:  provider,
		Message:   fmt.Sprintf("config error for %s: %s", field, message),
		Retryable: false,
	}
}

// NewAuthError creates an authentication error
func NewAuthError(provider string) *ProviderError {
	return &ProviderError{
		Provider:  provider,
		Code:      http.StatusUnauthorized,
		Message:   "authentication failed - check API key",
		Retryable: false,
	}
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(provider string, retryAfter string) *ProviderError {
	message := "rate limit exceeded"
	if retryAfter != "" {
		message += fmt.Sprintf(" - retry after %s", retryAfter)
	}

	return &ProviderError{
		Provider:  provider,
		Code:      http.StatusTooManyRequests,
		Message:   message,
		Retryable: true,
	}
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(provider string, original error) *ProviderError {
	return &ProviderError{
		Provider:  provider,
		Code:      http.StatusRequestTimeout,
		Message:   "request timeout",
		Original:  original,
		Retryable: true,
	}
}

// NewServiceUnavailableError creates a service unavailable error
func NewServiceUnavailableError(provider string) *ProviderError {
	return &ProviderError{
		Provider:  provider,
		Code:      http.StatusServiceUnavailable,
		Message:   "service temporarily unavailable",
		Retryable: true,
	}
}
