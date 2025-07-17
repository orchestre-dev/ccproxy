// Package common provides shared utilities for providers
package common

import (
	"context"
	"net/http"
	"time"
)

// Default health check configuration values
const (
	DefaultHealthCheckTimeout = 5 * time.Second
	DefaultHealthCheckRetries = 3
	DefaultHTTPErrorThreshold = 400
)

// HealthChecker defines the interface for provider health checks
type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}

// HealthCheckConfig holds configuration for health checks
type HealthCheckConfig struct {
	Timeout    time.Duration
	Retries    int
	RetryDelay time.Duration
}

// DefaultHealthCheckConfig returns default health check configuration
func DefaultHealthCheckConfig() HealthCheckConfig {
	return HealthCheckConfig{
		Timeout:    DefaultHealthCheckTimeout,
		Retries:    DefaultHealthCheckRetries,
		RetryDelay: 1 * time.Second,
	}
}

// BasicHealthCheck performs a basic HTTP health check
func BasicHealthCheck(ctx context.Context, client *http.Client, url string, provider string) error {
	ctx, cancel := context.WithTimeout(ctx, DefaultHealthCheckTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return NewProviderError(provider, "failed to create health check request", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return NewProviderError(provider, "health check request failed", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log error but don't return it as it's not critical
			// In a real implementation, you might want to log this error
			_ = closeErr // Acknowledge the error to satisfy staticcheck
		}
	}()

	if resp.StatusCode >= DefaultHTTPErrorThreshold {
		return NewHTTPError(provider, resp, nil)
	}

	return nil
}

// PerformHealthCheckWithRetry performs health check with retry logic
func PerformHealthCheckWithRetry(ctx context.Context, checker HealthChecker, config HealthCheckConfig) error {
	var lastErr error

	for i := 0; i < config.Retries; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(config.RetryDelay):
			}
		}

		checkCtx, cancel := context.WithTimeout(ctx, config.Timeout)
		err := checker.HealthCheck(checkCtx)
		cancel()

		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on non-retryable errors
		if providerErr, ok := err.(*ProviderError); ok && !providerErr.IsRetryable() {
			return err
		}
	}

	return lastErr
}
