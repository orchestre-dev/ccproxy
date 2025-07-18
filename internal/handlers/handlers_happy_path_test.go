// Package handlers provides happy path tests for handlers
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ccproxy/internal/config"
	"ccproxy/internal/models"
	"ccproxy/pkg/logger"
)

// MockHappyProvider implements provider.Provider for happy path testing
type MockHappyProvider struct {
	name      string
	model     string
	maxTokens int
	baseURL   string
}

func (m *MockHappyProvider) GetName() string       { return m.name }
func (m *MockHappyProvider) GetModel() string      { return m.model }
func (m *MockHappyProvider) GetMaxTokens() int     { return m.maxTokens }
func (m *MockHappyProvider) GetBaseURL() string    { return m.baseURL }
func (m *MockHappyProvider) ValidateConfig() error { return nil }

func (m *MockHappyProvider) CreateChatCompletion(ctx context.Context, req *models.ChatCompletionRequest) (*models.ChatCompletionResponse, error) {
	// Simulate successful response
	return &models.ChatCompletionResponse{
		ID:      "chatcmpl-" + m.name + "-123",
		Object:  "chat.completion",
		Model:   m.model,
		Created: time.Now().Unix(),
		Choices: []models.ChatCompletionChoice{
			{
				Index: 0,
				Message: models.ChatMessage{
					Role:    "assistant",
					Content: "I'm happy to help! The answer to your question is 42.",
				},
				FinishReason: "stop",
			},
		},
		Usage: models.ChatCompletionUsage{
			PromptTokens:     10,
			CompletionTokens: 15,
			TotalTokens:      25,
		},
	}, nil
}

func (m *MockHappyProvider) HealthCheck(ctx context.Context) error {
	// Always healthy
	return nil
}

func TestHealthCheck_HappyPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockLogger := logger.New(config.LoggingConfig{Level: "info", Format: "json"})
	provider := &MockHappyProvider{
		name:      "happy-provider",
		model:     "happy-model",
		maxTokens: 1000,
		baseURL:   "https://api.happy.com",
	}

	handler := NewHandler(provider, mockLogger)
	router := gin.New()
	router.GET("/", handler.HealthCheck)

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "CCProxy Multi-Provider Anthropic API is alive 💡", response["message"])
	assert.Equal(t, "happy-provider", response["provider"])
}

func TestDetailedHealthCheck_HappyPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockLogger := logger.New(config.LoggingConfig{Level: "info", Format: "json"})
	provider := &MockHappyProvider{
		name:      "happy-provider",
		model:     "happy-model",
		maxTokens: 1000,
		baseURL:   "https://api.happy.com",
	}

	handler := NewHandler(provider, mockLogger)
	router := gin.New()
	router.GET("/health", handler.DetailedHealthCheck)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "ccproxy", response["service"])
	assert.Equal(t, "1.0.0", response["version"])
}

func TestProviderStatus_HappyPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockLogger := logger.New(config.LoggingConfig{Level: "info", Format: "json"})
	provider := &MockHappyProvider{
		name:      "happy-provider",
		model:     "happy-model",
		maxTokens: 1000,
		baseURL:   "https://api.happy.com",
	}

	handler := NewHandler(provider, mockLogger)
	router := gin.New()
	router.GET("/status", handler.ProviderStatus)

	req, _ := http.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "happy-provider", response["provider"])
	assert.Equal(t, "happy-model", response["model"])
	assert.Equal(t, "https://api.happy.com", response["base_url"])
	assert.Equal(t, float64(1000), response["max_tokens"])
	assert.Equal(t, "active", response["status"])
}

func TestProxyMessages_SimpleRequest_HappyPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockLogger := logger.New(config.LoggingConfig{Level: "info", Format: "json"})
	provider := &MockHappyProvider{
		name:      "happy-provider",
		model:     "happy-model",
		maxTokens: 1000,
		baseURL:   "https://api.happy.com",
	}

	handler := NewHandler(provider, mockLogger)
	router := gin.New()
	router.POST("/v1/messages", handler.ProxyMessages)

	requestBody := models.MessagesRequest{
		Model: "happy-model",
		Messages: []models.Message{
			{Role: "user", Content: "What is the meaning of life?"},
		},
		MaxTokens: &[]int{100}[0],
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.MessagesResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure
	assert.NotEmpty(t, response.ID)
	assert.Equal(t, "message", response.Type)
	assert.Equal(t, "assistant", response.Role)
	assert.Equal(t, "happy-provider/happy-model", response.Model)
	assert.Equal(t, "end_turn", response.StopReason)
	assert.Greater(t, response.Usage.InputTokens, 0)
	assert.Greater(t, response.Usage.OutputTokens, 0)

	// Verify content
	require.Len(t, response.Content, 1)
	assert.Equal(t, "text", response.Content[0].Type)
	assert.Contains(t, response.Content[0].Text, "42")
}

func TestProxyMessages_MultiTurnConversation_HappyPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockLogger := logger.New(config.LoggingConfig{Level: "info", Format: "json"})
	provider := &MockHappyProvider{
		name:      "happy-provider",
		model:     "happy-model",
		maxTokens: 1000,
		baseURL:   "https://api.happy.com",
	}

	handler := NewHandler(provider, mockLogger)
	router := gin.New()
	router.POST("/v1/messages", handler.ProxyMessages)

	requestBody := models.MessagesRequest{
		Model: "happy-model",
		Messages: []models.Message{
			{Role: "user", Content: "Hello, I'm learning about AI."},
			{Role: "assistant", Content: "That's great! I'm here to help you learn about AI."},
			{Role: "user", Content: "What should I start with?"},
		},
		MaxTokens: &[]int{200}[0],
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.MessagesResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify conversation context was handled
	assert.NotEmpty(t, response.ID)
	assert.Equal(t, "assistant", response.Role)
	assert.Len(t, response.Content, 1)
}

func TestProxyMessages_WithSystemPrompt_HappyPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockLogger := logger.New(config.LoggingConfig{Level: "info", Format: "json"})
	provider := &MockHappyProvider{
		name:      "happy-provider",
		model:     "happy-model",
		maxTokens: 1000,
		baseURL:   "https://api.happy.com",
	}

	handler := NewHandler(provider, mockLogger)
	router := gin.New()
	router.POST("/v1/messages", handler.ProxyMessages)

	// System prompt should be added as first message with system role
	requestBody := models.MessagesRequest{
		Model: "happy-model",
		Messages: []models.Message{
			{Role: "system", Content: "You are a helpful AI assistant specializing in Python programming."},
			{Role: "user", Content: "How do I create a list in Python?"},
		},
		MaxTokens: &[]int{150}[0],
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.MessagesResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.ID)
	assert.Equal(t, "assistant", response.Role)
}

func TestProxyMessages_WithTemperature_HappyPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockLogger := logger.New(config.LoggingConfig{Level: "info", Format: "json"})
	provider := &MockHappyProvider{
		name:      "happy-provider",
		model:     "happy-model",
		maxTokens: 1000,
		baseURL:   "https://api.happy.com",
	}

	handler := NewHandler(provider, mockLogger)
	router := gin.New()
	router.POST("/v1/messages", handler.ProxyMessages)

	requestBody := models.MessagesRequest{
		Model: "happy-model",
		Messages: []models.Message{
			{Role: "user", Content: "Write a creative story about a robot."},
		},
		MaxTokens:   &[]int{200}[0],
		Temperature: float64Ptr(0.9),
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.MessagesResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.ID)
	assert.Len(t, response.Content, 1)
}

func TestProxyMessages_LargeRequest_HappyPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockLogger := logger.New(config.LoggingConfig{Level: "info", Format: "json"})
	provider := &MockHappyProvider{
		name:      "happy-provider",
		model:     "happy-model",
		maxTokens: 1000,
		baseURL:   "https://api.happy.com",
	}

	handler := NewHandler(provider, mockLogger)
	router := gin.New()
	router.POST("/v1/messages", handler.ProxyMessages)

	// Create a large request with multiple messages
	var messages []models.Message
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			messages = append(messages, models.Message{
				Role:    "user",
				Content: "This is message number " + string(rune(i)) + " with some content to make it longer.",
			})
		} else {
			messages = append(messages, models.Message{
				Role:    "assistant",
				Content: "This is response number " + string(rune(i)) + " with helpful information.",
			})
		}
	}

	requestBody := models.MessagesRequest{
		Model:     "happy-model",
		Messages:  messages,
		MaxTokens: &[]int{500}[0],
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.MessagesResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.ID)
	assert.Equal(t, "end_turn", response.StopReason)
}

func TestProxyMessages_ConcurrentRequests_HappyPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockLogger := logger.New(config.LoggingConfig{Level: "info", Format: "json"})
	provider := &MockHappyProvider{
		name:      "happy-provider",
		model:     "happy-model",
		maxTokens: 1000,
		baseURL:   "https://api.happy.com",
	}

	handler := NewHandler(provider, mockLogger)
	router := gin.New()
	router.POST("/v1/messages", handler.ProxyMessages)

	// Launch concurrent requests
	numRequests := 10
	results := make(chan int, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(requestNum int) {
			requestBody := models.MessagesRequest{
				Model: "happy-model",
				Messages: []models.Message{
					{Role: "user", Content: "Request number " + string(rune(requestNum))},
				},
				MaxTokens: &[]int{50}[0],
			}

			bodyBytes, _ := json.Marshal(requestBody)
			req, _ := http.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			results <- w.Code
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < numRequests; i++ {
		if statusCode := <-results; statusCode == http.StatusOK {
			successCount++
		}
	}

	assert.Equal(t, numRequests, successCount, "All concurrent requests should succeed")
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}