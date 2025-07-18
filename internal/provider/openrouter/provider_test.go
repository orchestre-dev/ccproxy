package openrouter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ccproxy/internal/config"
	"ccproxy/internal/models"
	"ccproxy/pkg/logger"
)

func TestNewProvider(t *testing.T) {
	cfg := &config.OpenRouterConfig{
		APIKey:    "test-key",
		Model:     "openai/gpt-4",
		BaseURL:   "https://openrouter.ai/api/v1",
		MaxTokens: 1000,
		Timeout:   30 * time.Second,
	}
	log := logger.New(config.LoggingConfig{Level: "info", Format: "json"})

	provider, err := NewProvider(cfg, log)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if provider == nil {
		t.Fatal("Provider should not be nil")
	}

	if provider.GetName() != "openrouter" {
		t.Errorf("Expected provider name 'openrouter', got %s", provider.GetName())
	}

	if provider.GetModel() != "openai/gpt-4" {
		t.Errorf("Expected model 'openai/gpt-4', got %s", provider.GetModel())
	}
}

func TestNewProvider_NilConfig(t *testing.T) {
	log := logger.New(config.LoggingConfig{Level: "info", Format: "json"})

	provider, err := NewProvider(nil, log)
	if err == nil {
		t.Fatal("Expected error for nil config, got none")
	}

	if provider != nil {
		t.Fatal("Expected nil provider for nil config")
	}
}

func TestProvider_CreateChatCompletion_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/chat/completions" {
			t.Errorf("Expected path /chat/completions, got %s", r.URL.Path)
		}

		// Verify headers
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Expected Authorization header 'Bearer test-key', got %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("HTTP-Referer") != "https://github.com/your-org/ccproxy" {
			t.Errorf("Expected HTTP-Referer header, got %s", r.Header.Get("HTTP-Referer"))
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		//nolint:errcheck
		w.Write([]byte(`{
			"id": "chatcmpl-test123",
			"object": "chat.completion",
			"created": 1677652288,
			"model": "openai/gpt-4",
			"choices": [{
				"index": 0,
				"message": {
					"role": "assistant",
					"content": "Hello! How can I help you today?"
				},
				"finish_reason": "stop"
			}],
			"usage": {
				"prompt_tokens": 12,
				"completion_tokens": 8,
				"total_tokens": 20
			}
		}`))
	}))
	defer server.Close()

	cfg := &config.OpenRouterConfig{
		APIKey:    "test-key",
		Model:     "openai/gpt-4",
		BaseURL:   server.URL,
		MaxTokens: 1000,
		Timeout:   30 * time.Second,
		SiteURL:   "https://github.com/your-org/ccproxy",
	}
	log := logger.New(config.LoggingConfig{Level: "info", Format: "json"})
	provider, _ := NewProvider(cfg, log) //nolint:errcheck

	maxTokens := 100
	req := &models.ChatCompletionRequest{
		Model: "openai/gpt-4",
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

	if resp.ID != "chatcmpl-test123" {
		t.Errorf("Expected ID 'chatcmpl-test123', got %s", resp.ID)
	}

	if len(resp.Choices) != 1 {
		t.Errorf("Expected 1 choice, got %d", len(resp.Choices))
	}

	if resp.Choices[0].Message.Content != "Hello! How can I help you today?" {
		t.Errorf("Expected specific content, got %s", resp.Choices[0].Message.Content)
	}

	if resp.Usage.PromptTokens != 12 {
		t.Errorf("Expected 12 prompt tokens, got %d", resp.Usage.PromptTokens)
	}
}

func TestProvider_CreateChatCompletion_WithTools(t *testing.T) {
	// Create mock server that expects tools in request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		//nolint:errcheck
		w.Write([]byte(`{
			"id": "chatcmpl-test123",
			"object": "chat.completion",
			"created": 1677652288,
			"model": "openai/gpt-4",
			"choices": [{
				"index": 0,
				"message": {
					"role": "assistant",
					"content": null,
					"tool_calls": [{
						"id": "call_123",
						"type": "function",
						"function": {
							"name": "get_weather",
							"arguments": "{\"location\":\"San Francisco\"}"
						}
					}]
				},
				"finish_reason": "tool_calls"
			}],
			"usage": {
				"prompt_tokens": 25,
				"completion_tokens": 15,
				"total_tokens": 40
			}
		}`))
	}))
	defer server.Close()

	cfg := &config.OpenRouterConfig{
		APIKey:    "test-key",
		Model:     "openai/gpt-4",
		BaseURL:   server.URL,
		MaxTokens: 1000,
		Timeout:   30 * time.Second,
	}
	log := logger.New(config.LoggingConfig{Level: "info", Format: "json"})
	provider, _ := NewProvider(cfg, log) //nolint:errcheck

	maxTokens := 100
	req := &models.ChatCompletionRequest{
		Model: "openai/gpt-4",
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

	if resp.Choices[0].FinishReason != "tool_calls" {
		t.Errorf("Expected finish reason 'tool_calls', got %s", resp.Choices[0].FinishReason)
	}

	if len(resp.Choices[0].Message.ToolCalls) != 1 {
		t.Errorf("Expected 1 tool call, got %d", len(resp.Choices[0].Message.ToolCalls))
	}

	toolCall := resp.Choices[0].Message.ToolCalls[0]
	if toolCall.Function.Name != "get_weather" {
		t.Errorf("Expected function name 'get_weather', got %s", toolCall.Function.Name)
	}
}

func TestProvider_CreateChatCompletion_Error(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		//nolint:errcheck
		w.Write([]byte(`{
			"error": {
				"message": "Invalid API key",
				"type": "invalid_request_error",
				"code": "invalid_api_key"
			}
		}`))
	}))
	defer server.Close()

	cfg := &config.OpenRouterConfig{
		APIKey:    "invalid-key",
		Model:     "openai/gpt-4",
		BaseURL:   server.URL,
		MaxTokens: 1000,
		Timeout:   30 * time.Second,
	}
	log := logger.New(config.LoggingConfig{Level: "info", Format: "json"})
	provider, _ := NewProvider(cfg, log) //nolint:errcheck

	maxTokens := 100
	req := &models.ChatCompletionRequest{
		Model: "openai/gpt-4",
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

	// Check that error contains meaningful information
	if err.Error() == "" {
		t.Error("Error message should not be empty")
	}
}

func TestProvider_CreateChatCompletion_Timeout(t *testing.T) {
	// Create mock server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond) // Delay longer than context timeout
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		//nolint:errcheck
		w.Write([]byte(`{"id":"test","choices":[{"message":{"role":"assistant","content":"delayed"}}]}`))
	}))
	defer server.Close()

	cfg := &config.OpenRouterConfig{
		APIKey:    "test-key",
		Model:     "openai/gpt-4",
		BaseURL:   server.URL,
		MaxTokens: 1000,
		Timeout:   30 * time.Second,
	}
	log := logger.New(config.LoggingConfig{Level: "info", Format: "json"})
	provider, _ := NewProvider(cfg, log) //nolint:errcheck

	maxTokens := 100
	req := &models.ChatCompletionRequest{
		Model: "openai/gpt-4",
		Messages: []models.ChatMessage{
			{Role: "user", Content: "Hello!"},
		},
		MaxTokens: &maxTokens,
	}

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	resp, err := provider.CreateChatCompletion(ctx, req)

	if err == nil {
		t.Fatal("Expected timeout error, got none")
	}

	if resp != nil {
		t.Fatal("Expected nil response on timeout")
	}

	// Verify it's a context timeout error
	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("Expected context deadline exceeded, got %v", ctx.Err())
	}
}

func TestProvider_CreateChatCompletion_MaxTokensCapping(t *testing.T) {
	// Create mock server that returns successful response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		//nolint:errcheck
		w.Write([]byte(`{
			"id": "chatcmpl-test",
			"object": "chat.completion",
			"created": 1677652288,
			"model": "openai/gpt-4",
			"choices": [{
				"index": 0,
				"message": {
					"role": "assistant",
					"content": "Response"
				},
				"finish_reason": "stop"
			}],
			"usage": {
				"prompt_tokens": 10,
				"completion_tokens": 5,
				"total_tokens": 15
			}
		}`))
	}))
	defer server.Close()

	cfg := &config.OpenRouterConfig{
		APIKey:    "test-key",
		Model:     "openai/gpt-4",
		BaseURL:   server.URL,
		MaxTokens: 500, // Provider max limit
		Timeout:   30 * time.Second,
	}
	log := logger.New(config.LoggingConfig{Level: "info", Format: "json"})
	provider, _ := NewProvider(cfg, log) //nolint:errcheck

	// Request with tokens exceeding provider limit
	maxTokens := 1000
	req := &models.ChatCompletionRequest{
		Model: "openai/gpt-4",
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

	// Verify max tokens was capped to provider limit
	if *req.MaxTokens != 500 {
		t.Errorf("Expected max tokens to be capped to 500, got %d", *req.MaxTokens)
	}
}

func TestProvider_HealthCheck(t *testing.T) {
	// Create mock server for health check
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/chat/completions" {
			t.Errorf("Expected path /chat/completions, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		//nolint:errcheck
		w.Write([]byte(`{
			"id": "health-check",
			"object": "chat.completion",
			"created": 1677652288,
			"model": "openai/gpt-4",
			"choices": [{
				"index": 0,
				"message": {
					"role": "assistant",
					"content": "OK"
				},
				"finish_reason": "stop"
			}],
			"usage": {
				"prompt_tokens": 2,
				"completion_tokens": 1,
				"total_tokens": 3
			}
		}`))
	}))
	defer server.Close()

	cfg := &config.OpenRouterConfig{
		APIKey:    "test-key",
		Model:     "openai/gpt-4",
		BaseURL:   server.URL,
		MaxTokens: 1000,
		Timeout:   30 * time.Second,
	}
	log := logger.New(config.LoggingConfig{Level: "info", Format: "json"})
	provider, _ := NewProvider(cfg, log) //nolint:errcheck

	ctx := context.Background()
	err := provider.HealthCheck(ctx)
	if err != nil {
		t.Errorf("Expected health check to pass, got error: %v", err)
	}
}

func TestProvider_HealthCheck_Error(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		//nolint:errcheck
		w.Write([]byte(`{"error": {"message": "Invalid API key"}}`))
	}))
	defer server.Close()

	cfg := &config.OpenRouterConfig{
		APIKey:    "invalid-key",
		Model:     "openai/gpt-4",
		BaseURL:   server.URL,
		MaxTokens: 1000,
		Timeout:   30 * time.Second,
	}
	log := logger.New(config.LoggingConfig{Level: "info", Format: "json"})
	provider, _ := NewProvider(cfg, log) //nolint:errcheck

	ctx := context.Background()
	err := provider.HealthCheck(ctx)
	if err == nil {
		t.Error("Expected health check to fail with invalid API key")
	}
}

func TestProvider_InterfaceCompliance(t *testing.T) {
	cfg := &config.OpenRouterConfig{
		APIKey:    "test-key",
		Model:     "openai/gpt-4",
		BaseURL:   "https://openrouter.ai/api/v1",
		MaxTokens: 1000,
		Timeout:   30 * time.Second,
	}
	log := logger.New(config.LoggingConfig{Level: "info", Format: "json"})
	provider, _ := NewProvider(cfg, log) //nolint:errcheck

	// Test interface methods
	if provider.GetName() != "openrouter" {
		t.Errorf("Expected GetName() to return 'openrouter', got %s", provider.GetName())
	}

	if provider.GetModel() != "openai/gpt-4" {
		t.Errorf("Expected GetModel() to return 'openai/gpt-4', got %s", provider.GetModel())
	}

	if provider.GetMaxTokens() != 1000 {
		t.Errorf("Expected GetMaxTokens() to return 1000, got %d", provider.GetMaxTokens())
	}

	if provider.GetBaseURL() != "https://openrouter.ai/api/v1" {
		t.Errorf("Expected GetBaseURL() to return base URL, got %s", provider.GetBaseURL())
	}

	// Test ValidateConfig (should pass with valid config)
	err := provider.ValidateConfig()
	if err != nil {
		t.Errorf("Expected ValidateConfig() to pass, got error: %v", err)
	}
}

func TestProvider_ValidateConfig_InvalidConfig(t *testing.T) {
	testCases := []struct {
		name   string
		config *config.OpenRouterConfig
	}{
		{
			name: "empty API key",
			config: &config.OpenRouterConfig{
				APIKey:    "",
				Model:     "openai/gpt-4",
				BaseURL:   "https://openrouter.ai/api/v1",
				MaxTokens: 1000,
				Timeout:   30 * time.Second,
			},
		},
		{
			name: "empty model",
			config: &config.OpenRouterConfig{
				APIKey:    "test-key",
				Model:     "",
				BaseURL:   "https://openrouter.ai/api/v1",
				MaxTokens: 1000,
				Timeout:   30 * time.Second,
			},
		},
		{
			name: "empty base URL",
			config: &config.OpenRouterConfig{
				APIKey:    "test-key",
				Model:     "openai/gpt-4",
				BaseURL:   "",
				MaxTokens: 1000,
				Timeout:   30 * time.Second,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			log := logger.New(config.LoggingConfig{Level: "info", Format: "json"})
			provider, _ := NewProvider(tc.config, log) //nolint:errcheck

			err := provider.ValidateConfig()
			if err == nil {
				t.Errorf("Expected ValidateConfig() to fail for %s", tc.name)
			}
		})
	}
}