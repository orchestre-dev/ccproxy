package gemini

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"ccproxy/internal/config"
	"ccproxy/internal/models"
	"ccproxy/pkg/logger"
)

func TestNewProvider(t *testing.T) {
	cfg := &config.GeminiConfig{
		APIKey:    "test-key",
		Model:     "gemini-2.0-flash",
		BaseURL:   "https://generativelanguage.googleapis.com",
		MaxTokens: 1000,
	}
	log := logger.New(config.LoggingConfig{Level: "info", Format: "json"})

	provider := NewProvider(cfg, log)

	if provider == nil {
		t.Fatal("Provider should not be nil")
	}

	if provider.GetName() != "gemini" {
		t.Errorf("Expected provider name 'gemini', got %s", provider.GetName())
	}

	if provider.GetModel() != "gemini-2.0-flash" {
		t.Errorf("Expected model 'gemini-2.0-flash', got %s", provider.GetModel())
	}
}

func TestProvider_CreateChatCompletion_Success(t *testing.T) {
	// Create mock server that expects Gemini API format
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path structure
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Gemini API path should include model and API key
		expectedPath := "/v1beta/models/gemini-2.0-flash:generateContent"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Verify API key in query parameters
		if r.URL.Query().Get("key") != "test-key" {
			t.Errorf("Expected API key 'test-key' in query, got %s", r.URL.Query().Get("key"))
		}

		// Return mock Gemini response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"candidates": [{
				"content": {
					"parts": [{
						"text": "Hello! How can I assist you today?"
					}],
					"role": "model"
				},
				"finishReason": "STOP",
				"index": 0
			}],
			"usageMetadata": {
				"promptTokenCount": 12,
				"candidatesTokenCount": 8,
				"totalTokenCount": 20
			}
		}`))
	}))
	defer server.Close()

	cfg := &config.GeminiConfig{
		APIKey:    "test-key",
		Model:     "gemini-2.0-flash",
		BaseURL:   server.URL,
		MaxTokens: 1000,
	}
	log := logger.New(config.LoggingConfig{Level: "info", Format: "json"})
	provider := NewProvider(cfg, log)

	maxTokens := 100
	req := &models.ChatCompletionRequest{
		Model: "gemini-2.0-flash",
		Messages: []models.ChatMessage{
			{Role: "user", Content: "Hello!"},
		},
		MaxTokens: &maxTokens,
	}

	ctx := context.Background()
	resp, err := provider.CreateChatCompletion(ctx, req)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("Response should not be nil")
	}

	if len(resp.Choices) != 1 {
		t.Errorf("Expected 1 choice, got %d", len(resp.Choices))
	}

	if resp.Choices[0].Message.Content != "Hello! How can I assist you today?" {
		t.Errorf("Expected specific content, got %s", resp.Choices[0].Message.Content)
	}

	if resp.Usage.PromptTokens != 12 {
		t.Errorf("Expected 12 prompt tokens, got %d", resp.Usage.PromptTokens)
	}

	if resp.Usage.CompletionTokens != 8 {
		t.Errorf("Expected 8 completion tokens, got %d", resp.Usage.CompletionTokens)
	}
}

func TestProvider_CreateChatCompletion_WithTools(t *testing.T) {
	// Create mock server that expects Gemini function calling format
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"candidates": [{
				"content": {
					"parts": [{
						"functionCall": {
							"name": "get_weather",
							"args": {
								"location": "San Francisco"
							}
						}
					}],
					"role": "model"
				},
				"finishReason": "STOP",
				"index": 0
			}],
			"usageMetadata": {
				"promptTokenCount": 25,
				"candidatesTokenCount": 15,
				"totalTokenCount": 40
			}
		}`))
	}))
	defer server.Close()

	cfg := &config.GeminiConfig{
		APIKey:    "test-key",
		Model:     "gemini-2.0-flash",
		BaseURL:   server.URL,
		MaxTokens: 1000,
	}
	log := logger.New(config.LoggingConfig{Level: "info", Format: "json"})
	provider := NewProvider(cfg, log)

	maxTokens := 100
	req := &models.ChatCompletionRequest{
		Model: "gemini-2.0-flash",
		Messages: []models.ChatMessage{
			{Role: "user", Content: "What's the weather in San Francisco?"},
		},
		Tools: []models.ChatCompletionTool{
			{
				Type: "function",
				Function: models.ChatCompletionToolFunction{
					Name:        "get_weather",
					Description: "Get weather for a location",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"location": map[string]interface{}{
								"type":        "string",
								"description": "The location to get weather for",
							},
						},
						"required": []string{"location"},
					},
				},
			},
		},
		MaxTokens: &maxTokens,
	}

	ctx := context.Background()
	resp, err := provider.CreateChatCompletion(ctx, req)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("Response should not be nil")
	}

	if len(resp.Choices) != 1 {
		t.Errorf("Expected 1 choice, got %d", len(resp.Choices))
	}

	// Verify tool call conversion
	if len(resp.Choices[0].Message.ToolCalls) != 1 {
		t.Errorf("Expected 1 tool call, got %d", len(resp.Choices[0].Message.ToolCalls))
	}

	toolCall := resp.Choices[0].Message.ToolCalls[0]
	if toolCall.Function.Name != "get_weather" {
		t.Errorf("Expected function name 'get_weather', got %s", toolCall.Function.Name)
	}
}

func TestProvider_CreateChatCompletion_Error(t *testing.T) {
	// Create mock server that returns Gemini API error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{
			"error": {
				"code": 400,
				"message": "API key not valid",
				"status": "INVALID_ARGUMENT"
			}
		}`))
	}))
	defer server.Close()

	cfg := &config.GeminiConfig{
		APIKey:    "invalid-key",
		Model:     "gemini-2.0-flash",
		BaseURL:   server.URL,
		MaxTokens: 1000,
	}
	log := logger.New(config.LoggingConfig{Level: "info", Format: "json"})
	provider := NewProvider(cfg, log)

	maxTokens := 100
	req := &models.ChatCompletionRequest{
		Model: "gemini-2.0-flash",
		Messages: []models.ChatMessage{
			{Role: "user", Content: "Hello!"},
		},
		MaxTokens: &maxTokens,
	}

	ctx := context.Background()
	resp, err := provider.CreateChatCompletion(ctx, req)

	if err == nil {
		t.Fatal("Expected error, got none")
	}

	if resp != nil {
		t.Fatal("Expected nil response on error")
	}
}

func TestProvider_InterfaceCompliance(t *testing.T) {
	cfg := &config.GeminiConfig{
		APIKey:    "test-key",
		Model:     "gemini-2.0-flash",
		BaseURL:   "https://generativelanguage.googleapis.com",
		MaxTokens: 1000,
	}
	log := logger.New(config.LoggingConfig{Level: "info", Format: "json"})
	provider := NewProvider(cfg, log)

	// Test interface methods
	if provider.GetName() != "gemini" {
		t.Errorf("Expected GetName() to return 'gemini', got %s", provider.GetName())
	}

	if provider.GetModel() != "gemini-2.0-flash" {
		t.Errorf("Expected GetModel() to return 'gemini-2.0-flash', got %s", provider.GetModel())
	}

	if provider.GetMaxTokens() != 1000 {
		t.Errorf("Expected GetMaxTokens() to return 1000, got %d", provider.GetMaxTokens())
	}

	if provider.GetBaseURL() != "https://generativelanguage.googleapis.com" {
		t.Errorf("Expected GetBaseURL() to return base URL, got %s", provider.GetBaseURL())
	}

	// Test ValidateConfig
	err := provider.ValidateConfig()
	if err != nil {
		t.Errorf("Expected ValidateConfig() to pass, got error: %v", err)
	}
}
