package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ccproxy/internal/config"
	"ccproxy/internal/handlers"
	"ccproxy/internal/models"
	"ccproxy/internal/provider"
	"ccproxy/pkg/logger"

	"github.com/gin-gonic/gin"
)

func TestIntegrationAnthropicProxy(t *testing.T) {
	// Create a mock provider server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock OpenAI-style response
		response := models.ChatCompletionResponse{
			ID:      "chatcmpl-test123",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "test-model",
			Choices: []models.ChatCompletionChoice{
				{
					Index: 0,
					Message: models.ChatMessage{
						Role:    "assistant",
						Content: "Hello! I'm a test response from the mock provider.",
					},
					FinishReason: "stop",
				},
			},
			Usage: models.ChatCompletionUsage{
				PromptTokens:     15,
				CompletionTokens: 12,
				TotalTokens:      27,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	// Create test configuration
	cfg := &config.Config{
		Provider: "groq",
		Providers: config.ProvidersConfig{
			Groq: config.GroqConfig{
				APIKey:    "test-key",
				BaseURL:   mockServer.URL,
				Model:     "test-model",
				MaxTokens: 1000,
				Timeout:   30 * time.Second,
			},
		},
	}

	// Create logger and provider factory
	testLogger := logger.New(config.LoggingConfig{
		Level:  "info",
		Format: "json",
	})
	factory := provider.NewFactory(cfg, testLogger)
	providerInstance, err := factory.CreateProvider()
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Create handler
	handler := handlers.NewHandler(providerInstance, testLogger)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/v1/messages", handler.ProxyMessages)

	// Create test Anthropic request
	anthropicRequest := models.MessagesRequest{
		Model: "claude-3-sonnet",
		Messages: []models.Message{
			{
				Role:    "user",
				Content: "Hello, how are you today?",
			},
		},
		MaxTokens:   intPtr(100),
		Temperature: floatPtr(0.7),
	}

	// Marshal request
	requestBody, err := json.Marshal(anthropicRequest)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "/v1/messages", bytes.NewReader(requestBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		return
	}

	// Parse response
	var response models.MessagesResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Validate response structure
	if response.Type != "message" {
		t.Errorf("Expected type 'message', got '%s'", response.Type)
	}
	if response.Role != "assistant" {
		t.Errorf("Expected role 'assistant', got '%s'", response.Role)
	}
	if response.StopReason != "end_turn" {
		t.Errorf("Expected stop_reason 'end_turn', got '%s'", response.StopReason)
	}
	if len(response.Content) != 1 {
		t.Errorf("Expected 1 content block, got %d", len(response.Content))
	}
	if response.Content[0].Type != "text" {
		t.Errorf("Expected content type 'text', got '%s'", response.Content[0].Type)
	}
	if response.Content[0].Text != "Hello! I'm a test response from the mock provider." {
		t.Errorf("Unexpected content text: %s", response.Content[0].Text)
	}

	// Validate usage
	if response.Usage.InputTokens != 15 {
		t.Errorf("Expected input_tokens 15, got %d", response.Usage.InputTokens)
	}
	if response.Usage.OutputTokens != 12 {
		t.Errorf("Expected output_tokens 12, got %d", response.Usage.OutputTokens)
	}
}

func TestIntegrationAnthropicProxyWithTools(t *testing.T) {
	// Create a mock provider server that returns tool calls
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock OpenAI-style response with tool calls
		response := models.ChatCompletionResponse{
			ID:      "chatcmpl-test456",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "test-model",
			Choices: []models.ChatCompletionChoice{
				{
					Index: 0,
					Message: models.ChatMessage{
						Role: "assistant",
						ToolCalls: []models.ToolCall{
							{
								ID:   "call_test123",
								Type: "function",
								Function: models.FunctionCall{
									Name:      "get_weather",
									Arguments: "{\"location\":\"San Francisco\",\"unit\":\"celsius\"}",
								},
							},
						},
					},
					FinishReason: "tool_calls",
				},
			},
			Usage: models.ChatCompletionUsage{
				PromptTokens:     20,
				CompletionTokens: 8,
				TotalTokens:      28,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	// Create test configuration
	cfg := &config.Config{
		Provider: "groq",
		Providers: config.ProvidersConfig{
			Groq: config.GroqConfig{
				APIKey:    "test-key",
				BaseURL:   mockServer.URL,
				Model:     "test-model",
				MaxTokens: 1000,
				Timeout:   30 * time.Second,
			},
		},
	}

	// Create logger and provider factory
	testLogger := logger.New(config.LoggingConfig{
		Level:  "info",
		Format: "json",
	})
	factory := provider.NewFactory(cfg, testLogger)
	providerInstance, err := factory.CreateProvider()
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Create handler
	handler := handlers.NewHandler(providerInstance, testLogger)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/v1/messages", handler.ProxyMessages)

	// Create test Anthropic request with tools
	anthropicRequest := models.MessagesRequest{
		Model: "claude-3-sonnet",
		Messages: []models.Message{
			{
				Role:    "user",
				Content: "What's the weather in San Francisco?",
			},
		},
		Tools: []models.Tool{
			{
				Name:        "get_weather",
				Description: stringPtr("Get current weather information"),
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]interface{}{
							"type":        "string",
							"description": "The city name",
						},
						"unit": map[string]interface{}{
							"type":        "string",
							"description": "Temperature unit",
							"enum":        []string{"celsius", "fahrenheit"},
						},
					},
					"required": []string{"location"},
				},
			},
		},
		MaxTokens:   intPtr(100),
		Temperature: floatPtr(0.1),
	}

	// Marshal request
	requestBody, err := json.Marshal(anthropicRequest)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "/v1/messages", bytes.NewReader(requestBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		return
	}

	// Parse response
	var response models.MessagesResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Validate response structure for tool call
	if response.Type != "message" {
		t.Errorf("Expected type 'message', got '%s'", response.Type)
	}
	if response.Role != "assistant" {
		t.Errorf("Expected role 'assistant', got '%s'", response.Role)
	}
	if response.StopReason != "tool_use" {
		t.Errorf("Expected stop_reason 'tool_use', got '%s'", response.StopReason)
	}
	if len(response.Content) != 1 {
		t.Errorf("Expected 1 content block, got %d", len(response.Content))
	}
	if response.Content[0].Type != "tool_use" {
		t.Errorf("Expected content type 'tool_use', got '%s'", response.Content[0].Type)
	}
	if response.Content[0].ID != "call_test123" {
		t.Errorf("Expected tool call ID 'call_test123', got '%s'", response.Content[0].ID)
	}
	if response.Content[0].Name != "get_weather" {
		t.Errorf("Expected tool name 'get_weather', got '%s'", response.Content[0].Name)
	}

	// Validate tool input
	expectedLocation := "San Francisco"
	expectedUnit := "celsius"
	if location, ok := response.Content[0].Input["location"].(string); !ok || location != expectedLocation {
		t.Errorf("Expected location '%s', got '%v'", expectedLocation, response.Content[0].Input["location"])
	}
	if unit, ok := response.Content[0].Input["unit"].(string); !ok || unit != expectedUnit {
		t.Errorf("Expected unit '%s', got '%v'", expectedUnit, response.Content[0].Input["unit"])
	}
}

func TestProviderFactoryAllProviders(t *testing.T) {
	testLogger := logger.New(config.LoggingConfig{
		Level:  "info",
		Format: "json",
	})

	providers := []struct {
		name     string
		provider string
		config   func() *config.Config
	}{
		{
			name:     "Groq",
			provider: "groq",
			config: func() *config.Config {
				return &config.Config{
					Provider: "groq",
					Providers: config.ProvidersConfig{
						Groq: config.GroqConfig{
							APIKey:    "test-key",
							BaseURL:   "https://api.groq.com/openai/v1",
							Model:     "test-model",
							MaxTokens: 1000,
							Timeout:   30 * time.Second,
						},
					},
				}
			},
		},
		{
			name:     "OpenAI",
			provider: "openai",
			config: func() *config.Config {
				return &config.Config{
					Provider: "openai",
					Providers: config.ProvidersConfig{
						OpenAI: config.OpenAIConfig{
							APIKey:    "test-key",
							BaseURL:   "https://api.openai.com/v1",
							Model:     "gpt-4o",
							MaxTokens: 4096,
							Timeout:   30 * time.Second,
						},
					},
				}
			},
		},
		{
			name:     "Mistral",
			provider: "mistral",
			config: func() *config.Config {
				return &config.Config{
					Provider: "mistral",
					Providers: config.ProvidersConfig{
						Mistral: config.MistralConfig{
							APIKey:    "test-key",
							BaseURL:   "https://api.mistral.ai/v1",
							Model:     "mistral-large-latest",
							MaxTokens: 32768,
							Timeout:   30 * time.Second,
						},
					},
				}
			},
		},
		{
			name:     "Ollama",
			provider: "ollama",
			config: func() *config.Config {
				return &config.Config{
					Provider: "ollama",
					Providers: config.ProvidersConfig{
						Ollama: config.OllamaConfig{
							APIKey:    "ollama",
							BaseURL:   "http://localhost:11434",
							Model:     "llama3.2",
							MaxTokens: 4096,
							Timeout:   30 * time.Second,
						},
					},
				}
			},
		},
		{
			name:     "Gemini",
			provider: "gemini",
			config: func() *config.Config {
				return &config.Config{
					Provider: "gemini",
					Providers: config.ProvidersConfig{
						Gemini: config.GeminiConfig{
							APIKey:    "test-key",
							BaseURL:   "https://generativelanguage.googleapis.com",
							Model:     "gemini-2.0-flash",
							MaxTokens: 32768,
							Timeout:   30 * time.Second,
						},
					},
				}
			},
		},
	}

	for _, tt := range providers {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.config()
			factory := provider.NewFactory(cfg, testLogger)

			// Test provider creation
			providerInstance, err := factory.CreateProvider()
			if err != nil {
				t.Fatalf("Failed to create %s provider: %v", tt.name, err)
			}

			// Test provider interface methods
			if providerInstance.GetName() != tt.provider {
				t.Errorf("Expected provider name '%s', got '%s'", tt.provider, providerInstance.GetName())
			}

			if providerInstance.GetModel() == "" {
				t.Errorf("Provider model should not be empty")
			}

			if providerInstance.GetMaxTokens() <= 0 {
				t.Errorf("Provider max tokens should be positive, got %d", providerInstance.GetMaxTokens())
			}

			if providerInstance.GetBaseURL() == "" {
				t.Errorf("Provider base URL should not be empty")
			}

			// Test config validation
			if err := providerInstance.ValidateConfig(); err != nil {
				t.Errorf("Provider config validation failed: %v", err)
			}
		})
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}

func stringPtr(s string) *string {
	return &s
}
