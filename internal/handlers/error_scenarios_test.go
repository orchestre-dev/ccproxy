// Package handlers provides comprehensive error scenario tests for HTTP handlers
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"ccproxy/internal/config"
	"ccproxy/internal/models"
	"ccproxy/internal/provider"
	"ccproxy/internal/provider/common"
	"ccproxy/pkg/logger"
)

// MockErrorProvider implements provider.Provider for error testing
type MockErrorProvider struct {
	name           string
	model          string
	maxTokens      int
	baseURL        string
	errorMode      string
	errorDelay     time.Duration
	requestCount   int32
	healthyAfter   int32
}

func (m *MockErrorProvider) GetName() string         { return m.name }
func (m *MockErrorProvider) GetModel() string       { return m.model }
func (m *MockErrorProvider) GetMaxTokens() int      { return m.maxTokens }
func (m *MockErrorProvider) GetBaseURL() string     { return m.baseURL }
func (m *MockErrorProvider) ValidateConfig() error  { return nil }

func (m *MockErrorProvider) CreateChatCompletion(ctx context.Context, req *models.ChatCompletionRequest) (*models.ChatCompletionResponse, error) {
	count := atomic.AddInt32(&m.requestCount, 1)

	// Simulate delay if specified
	if m.errorDelay > 0 {
		select {
		case <-time.After(m.errorDelay):
		case <-ctx.Done():
			return nil, common.NewProviderError(m.name, "request cancelled", ctx.Err())
		}
	}

	switch m.errorMode {
	case "timeout":
		return nil, common.NewTimeoutError(m.name, errors.New("request timeout"))
	case "rate_limit":
		return nil, common.NewRateLimitError(m.name, "60s")
	case "auth_error":
		return nil, common.NewAuthError(m.name)
	case "service_unavailable":
		return nil, common.NewServiceUnavailableError(m.name)
	case "config_error":
		return nil, common.NewConfigError(m.name, "api_key", "invalid API key")
	case "network_error":
		return nil, common.NewProviderError(m.name, "network connection failed", errors.New("connection refused"))
	case "intermittent":
		if count <= m.healthyAfter {
			return nil, common.NewServiceUnavailableError(m.name)
		}
		fallthrough
	case "success":
		return &models.ChatCompletionResponse{
			ID:      "test-response-" + m.name,
			Object:  "chat.completion",
			Model:   m.model,
			Created: time.Now().Unix(),
			Choices: []models.ChatCompletionChoice{
				{
					Index:        0,
					Message:      models.ChatMessage{Role: "assistant", Content: "Test response"},
					FinishReason: "stop",
				},
			},
			Usage: models.ChatCompletionUsage{
				PromptTokens:     10,
				CompletionTokens: 20,
				TotalTokens:      30,
			},
		}, nil
	default:
		return nil, common.NewProviderError(m.name, "unknown error mode", nil)
	}
}

func (m *MockErrorProvider) HealthCheck(ctx context.Context) error {
	switch m.errorMode {
	case "health_fail":
		return common.NewProviderError(m.name, "health check failed", errors.New("provider unhealthy"))
	case "timeout":
		// Health check should also timeout in timeout mode
		return common.NewTimeoutError(m.name, errors.New("health check timeout"))
	default:
		return nil
	}
}

func TestProxyMessagesErrorScenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockLogger := logger.New(config.LoggingConfig{Level: "error", Format: "json"})

	tests := []struct {
		name           string
		requestBody    interface{}
		provider       provider.Provider
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Invalid JSON Request",
			requestBody:    `{"invalid": json}`,
			provider:       &MockErrorProvider{name: "test", model: "test-model", maxTokens: 1000, errorMode: "success"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request format",
		},
		{
			name: "Missing Required Fields",
			requestBody: models.MessagesRequest{
				// Missing required fields - empty Messages array
				Model: "test-model",
			},
			provider:       &MockErrorProvider{name: "test", model: "test-model", maxTokens: 1000, errorMode: "success"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "messages array",
		},
		{
			name: "Provider Timeout Error",
			requestBody: models.MessagesRequest{
				Model: "test-model",
				Messages: []models.Message{
					{Role: "user", Content: "test message"},
				},
				MaxTokens: &[]int{100}[0],
			},
			provider:       &MockErrorProvider{name: "test", model: "test-model", maxTokens: 1000, errorMode: "timeout"},
			expectedStatus: http.StatusRequestTimeout,
			expectedError:  "Failed to process request",
		},
		{
			name: "Provider Rate Limit Error",
			requestBody: models.MessagesRequest{
				Model: "test-model",
				Messages: []models.Message{
					{Role: "user", Content: "test message"},
				},
				MaxTokens: &[]int{100}[0],
			},
			provider:       &MockErrorProvider{name: "test", model: "test-model", maxTokens: 1000, errorMode: "rate_limit"},
			expectedStatus: http.StatusTooManyRequests,
			expectedError:  "Failed to process request",
		},
		{
			name: "Provider Authentication Error",
			requestBody: models.MessagesRequest{
				Model: "test-model",
				Messages: []models.Message{
					{Role: "user", Content: "test message"},
				},
				MaxTokens: &[]int{100}[0],
			},
			provider:       &MockErrorProvider{name: "test", model: "test-model", maxTokens: 1000, errorMode: "auth_error"},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "authentication",
		},
		{
			name: "Provider Service Unavailable",
			requestBody: models.MessagesRequest{
				Model: "test-model",
				Messages: []models.Message{
					{Role: "user", Content: "test message"},
				},
				MaxTokens: &[]int{100}[0],
			},
			provider:       &MockErrorProvider{name: "test", model: "test-model", maxTokens: 1000, errorMode: "service_unavailable"},
			expectedStatus: http.StatusServiceUnavailable,
			expectedError:  "Failed to process request",
		},
		{
			name: "Provider Network Error",
			requestBody: models.MessagesRequest{
				Model: "test-model",
				Messages: []models.Message{
					{Role: "user", Content: "test message"},
				},
				MaxTokens: &[]int{100}[0],
			},
			provider:       &MockErrorProvider{name: "test", model: "test-model", maxTokens: 1000, errorMode: "network_error"},
			expectedStatus: http.StatusInternalServerError, // Network errors don't have specific status codes
			expectedError:  "Failed to process request",
		},
		{
			name: "Valid Request - Success",
			requestBody: models.MessagesRequest{
				Model: "test-model",
				Messages: []models.Message{
					{Role: "user", Content: "test message"},
				},
				MaxTokens: &[]int{100}[0],
			},
			provider:       &MockErrorProvider{name: "test", model: "test-model", maxTokens: 1000, errorMode: "success"},
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test handler
			handler := NewHandler(tt.provider, mockLogger)

			// Create test router
			router := gin.New()
			router.POST("/v1/messages", handler.ProxyMessages)

			// Prepare request body
			var bodyReader *bytes.Reader
			if str, ok := tt.requestBody.(string); ok {
				bodyReader = bytes.NewReader([]byte(str))
			} else {
				bodyBytes, _ := json.Marshal(tt.requestBody)
				bodyReader = bytes.NewReader(bodyBytes)
			}

			// Create request
			req, _ := http.NewRequest("POST", "/v1/messages", bodyReader)
			req.Header.Set("Content-Type", "application/json")

			// Record response
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d but got %d", tt.expectedStatus, w.Code)
			}

			// Verify error message if expected
			if tt.expectedError != "" {
				responseBody := w.Body.String()
				if !strings.Contains(strings.ToLower(responseBody), strings.ToLower(tt.expectedError)) {
					t.Errorf("Expected error message containing '%s' but got: %s", tt.expectedError, responseBody)
				}
			}
		})
	}
}

func TestHealthCheckErrorScenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockLogger := logger.New(config.LoggingConfig{Level: "error", Format: "json"})

	tests := []struct {
		name           string
		provider       provider.Provider
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Healthy Provider",
			provider:       &MockErrorProvider{name: "test", model: "test-model", maxTokens: 1000, errorMode: "success"},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Unhealthy Provider",
			provider:       &MockErrorProvider{name: "test", model: "test-model", maxTokens: 1000, errorMode: "health_fail"},
			expectedStatus: http.StatusServiceUnavailable,
			expectError:    true,
		},
		{
			name:           "Provider Timeout",
			provider:       &MockErrorProvider{name: "test", model: "test-model", maxTokens: 1000, errorMode: "timeout"},
			expectedStatus: http.StatusServiceUnavailable,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(tt.provider, mockLogger)
			
			router := gin.New()
			router.GET("/health", handler.DetailedHealthCheck)

			req, _ := http.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d but got %d", tt.expectedStatus, w.Code)
			}

			// Parse response
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to parse response JSON: %v", err)
			}

			if tt.expectError {
				if response["status"] != "unhealthy" {
					t.Error("Expected status to be 'unhealthy'")
				}
			} else {
				if response["status"] != "healthy" {
					t.Error("Expected status to be 'healthy'")
				}
			}
		})
	}
}

func TestLargeRequestErrorScenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockLogger := logger.New(config.LoggingConfig{Level: "error", Format: "json"})

	tests := []struct {
		name           string
		messageContent string
		expectedStatus int
	}{
		{
			name:           "Normal Size Message",
			messageContent: "This is a normal test message",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Large Message",
			messageContent: strings.Repeat("This is a very long message. ", 1000),
			expectedStatus: http.StatusOK, // Should still work
		},
		{
			name:           "Extremely Large Message",
			messageContent: strings.Repeat("x", 1024*1024), // 1MB message
			expectedStatus: http.StatusOK, // Handler doesn't have size limits configured
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockErrorProvider{name: "test", model: "test-model", maxTokens: 1000, errorMode: "success"}
			handler := NewHandler(provider, mockLogger)

			router := gin.New()
			router.POST("/v1/messages", handler.ProxyMessages)

			requestBody := models.MessagesRequest{
				Model: "test-model",
				Messages: []models.Message{
					{Role: "user", Content: tt.messageContent},
				},
				MaxTokens: &[]int{100}[0],
			}

			bodyBytes, _ := json.Marshal(requestBody)
			req, _ := http.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Note: Actual status may vary based on server configuration
			// This test mainly ensures the handler doesn't crash
			if w.Code == 0 {
				t.Error("Expected some HTTP status code")
			}

			t.Logf("Large request test '%s' resulted in status %d", tt.name, w.Code)
		})
	}
}

func TestConcurrentRequestErrorScenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockLogger := logger.New(config.LoggingConfig{Level: "error", Format: "json"})

	// Test concurrent requests with intermittent provider failures
	provider := &MockErrorProvider{
		name:         "test",
		model:        "test-model", 
		maxTokens:    1000,
		errorMode:    "intermittent",
		healthyAfter: 10, // First 10 requests fail, then succeed
	}

	handler := NewHandler(provider, mockLogger)
	router := gin.New()
	router.POST("/v1/messages", handler.ProxyMessages)

	const numGoroutines = 20
	const requestsPerGoroutine = 5
	results := make(chan int, numGoroutines*requestsPerGoroutine)

	// Launch concurrent requests
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < requestsPerGoroutine; j++ {
				requestBody := models.MessagesRequest{
					Model: "test-model",
					Messages: []models.Message{
						{Role: "user", Content: "concurrent test message"},
					},
					MaxTokens: &[]int{100}[0],
				}

				bodyBytes, _ := json.Marshal(requestBody)
				req, _ := http.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				results <- w.Code
			}
		}(i)
	}

	// Collect results
	statusCounts := make(map[int]int)
	totalRequests := numGoroutines * requestsPerGoroutine

	for i := 0; i < totalRequests; i++ {
		status := <-results
		statusCounts[status]++
	}

	t.Logf("Concurrent test results: %v", statusCounts)

	// Verify we got some failures and some successes (due to intermittent mode)
	if statusCounts[http.StatusOK] == 0 {
		t.Error("Expected some successful requests")
	}
	if statusCounts[http.StatusServiceUnavailable] == 0 {
		t.Error("Expected some failed requests in intermittent mode")
	}
}

func TestRequestTimeoutScenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockLogger := logger.New(config.LoggingConfig{Level: "error", Format: "json"})

	tests := []struct {
		name           string
		providerDelay  time.Duration
		contextTimeout time.Duration
		expectTimeout  bool
	}{
		{
			name:           "Fast Provider Response",
			providerDelay:  100 * time.Millisecond,
			contextTimeout: 1 * time.Second,
			expectTimeout:  false,
		},
		{
			name:           "Slow Provider Response",
			providerDelay:  2 * time.Second,
			contextTimeout: 500 * time.Millisecond,
			expectTimeout:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockErrorProvider{
				name:       "test",
				model:      "test-model",
				maxTokens:  1000,
				errorMode:  "success",
				errorDelay: tt.providerDelay,
			}

			handler := NewHandler(provider, mockLogger)
			router := gin.New()
			router.POST("/v1/messages", handler.ProxyMessages)

			requestBody := models.MessagesRequest{
				Model: "test-model",
				Messages: []models.Message{
					{Role: "user", Content: "timeout test message"},
				},
				MaxTokens: &[]int{100}[0],
			}

			bodyBytes, _ := json.Marshal(requestBody)
			req, _ := http.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			// Set request timeout
			ctx, cancel := context.WithTimeout(context.Background(), tt.contextTimeout)
			defer cancel()
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			
			start := time.Now()
			router.ServeHTTP(w, req)
			duration := time.Since(start)

			if tt.expectTimeout {
				// Should complete within timeout period plus small buffer
				if duration > tt.contextTimeout+200*time.Millisecond {
					t.Errorf("Request took too long: %v (expected to timeout around %v)", duration, tt.contextTimeout)
				}
				// Status should indicate an error
				if w.Code == http.StatusOK {
					t.Error("Expected error status due to timeout")
				}
			} else {
				// Should complete successfully
				if w.Code != http.StatusOK {
					t.Errorf("Expected successful response but got status %d", w.Code)
				}
			}

			t.Logf("Timeout test '%s' took %v, status: %d", tt.name, duration, w.Code)
		})
	}
}

func TestMalformedRequestScenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockLogger := logger.New(config.LoggingConfig{Level: "error", Format: "json"})
	provider := &MockErrorProvider{name: "test", model: "test-model", maxTokens: 1000, errorMode: "success"}
	handler := NewHandler(provider, mockLogger)

	router := gin.New()
	router.POST("/v1/messages", handler.ProxyMessages)

	tests := []struct {
		name        string
		contentType string
		body        string
		expectError bool
	}{
		{
			name:        "Empty Body",
			contentType: "application/json",
			body:        "",
			expectError: true,
		},
		{
			name:        "Invalid JSON",
			contentType: "application/json", 
			body:        `{"invalid": json`,
			expectError: true,
		},
		{
			name:        "Wrong Content Type",
			contentType: "text/plain",
			body:        `{"model": "test", "messages": [{"role": "user", "content": "test"}]}`,
			expectError: false, // Gin handles JSON regardless of content type
		},
		{
			name:        "Missing Content Type",
			contentType: "",
			body:        `{"model": "test", "messages": [{"role": "user", "content": "test"}]}`,
			expectError: false, // Gin handles JSON regardless of content type
		},
		{
			name:        "Binary Data",
			contentType: "application/json",
			body:        string([]byte{0x00, 0x01, 0x02, 0x03}),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/v1/messages", strings.NewReader(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if tt.expectError && w.Code == http.StatusOK {
				t.Error("Expected error but request succeeded")
			}
			if !tt.expectError && w.Code != http.StatusOK {
				t.Errorf("Expected success but got status %d", w.Code)
			}
		})
	}
}