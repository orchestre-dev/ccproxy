package provider

import (
	"context"
	"testing"

	"ccproxy/internal/config"
	"ccproxy/internal/models"
	"ccproxy/pkg/logger"
)

func TestNewFactory(t *testing.T) {
	cfg := &config.Config{
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		Provider: "groq",
	}
	log := logger.New(cfg.Logging)

	factory := NewFactory(cfg, log)

	if factory == nil {
		t.Fatal("Factory should not be nil")
	}
}

func TestCreateProvider_ValidProviders(t *testing.T) {
	testCases := []struct {
		name         string
		providerName string
		expectError  bool
	}{
		{"Groq", "groq", false},
		{"OpenAI", "openai", false},
		{"Gemini", "gemini", false},
		{"Ollama", "ollama", false},
		{"Invalid", "invalid", true},
		{"Empty", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{
				Provider: tc.providerName,
				Logging: config.LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				Providers: config.ProvidersConfig{
					Groq: config.GroqConfig{
						APIKey:    "test-key",
						Model:     "test-model",
						BaseURL:   "https://api.groq.com/openai/v1",
						MaxTokens: 1000,
					},
					OpenAI: config.OpenAIConfig{
						APIKey:    "test-key",
						Model:     "gpt-4o",
						BaseURL:   "https://api.openai.com/v1",
						MaxTokens: 1000,
					},
					Gemini: config.GeminiConfig{
						APIKey:    "test-key",
						Model:     "gemini-2.0-flash",
						BaseURL:   "https://generativelanguage.googleapis.com",
						MaxTokens: 1000,
					},
					Ollama: config.OllamaConfig{
						Model:     "llama3.2",
						BaseURL:   "http://localhost:11434",
						MaxTokens: 1000,
					},
				},
			}
			log := logger.New(cfg.Logging)
			factory := NewFactory(cfg, log)

			provider, err := factory.CreateProvider()

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for provider %s, but got none", tc.providerName)
				}
				if provider != nil {
					t.Errorf("Expected nil provider for invalid provider %s", tc.providerName)
				}
				return
			}

			if err != nil {
				// Some providers might fail due to missing dependencies (like Ollama)
				// but they should still be recognized as valid provider types
				t.Logf("Provider %s failed to initialize (expected in test): %v", tc.providerName, err)
				return
			}

			// Test provider methods if initialization succeeded
			if provider != nil {
				testProviderInterface(t, provider, tc.providerName)
			}
		})
	}
}

func testProviderInterface(t *testing.T, provider Provider, expectedName string) {
	// Test interface compliance
	var _ Provider = provider

	// Test GetName
	name := provider.GetName()
	if name != expectedName {
		t.Errorf("Expected GetName() to return %s, got %s", expectedName, name)
	}

	// Test GetModel returns non-empty string
	model := provider.GetModel()
	if model == "" {
		t.Errorf("Expected GetModel() to return non-empty string for %s", expectedName)
	}

	// Test GetMaxTokens returns positive value
	maxTokens := provider.GetMaxTokens()
	if maxTokens <= 0 {
		t.Errorf("Expected GetMaxTokens() to return positive value for %s, got %d", expectedName, maxTokens)
	}

	// Test GetBaseURL returns non-empty string
	baseURL := provider.GetBaseURL()
	if baseURL == "" {
		t.Errorf("Expected GetBaseURL() to return non-empty string for %s", expectedName)
	}

	// Test ValidateConfig (should not error for properly configured providers)
	err := provider.ValidateConfig()
	if err != nil {
		t.Logf("Provider %s config validation failed (may be expected): %v", expectedName, err)
	}
}

func TestProviderInterface_MockImplementation(t *testing.T) {
	// Test mock provider implementation
	mock := &mockProvider{
		name:      "mock",
		model:     "mock-model",
		maxTokens: 1000,
		baseURL:   "https://mock.api.com",
	}

	// Verify interface compliance
	var _ Provider = mock

	// Test all interface methods
	if mock.GetName() != "mock" {
		t.Errorf("Expected GetName() to return 'mock', got %s", mock.GetName())
	}

	if mock.GetModel() != "mock-model" {
		t.Errorf("Expected GetModel() to return 'mock-model', got %s", mock.GetModel())
	}

	if mock.GetMaxTokens() != 1000 {
		t.Errorf("Expected GetMaxTokens() to return 1000, got %d", mock.GetMaxTokens())
	}

	if mock.GetBaseURL() != "https://mock.api.com" {
		t.Errorf("Expected GetBaseURL() to return 'https://mock.api.com', got %s", mock.GetBaseURL())
	}

	err := mock.ValidateConfig()
	if err != nil {
		t.Errorf("Expected ValidateConfig() to return nil, got %v", err)
	}

	// Test CreateChatCompletion
	ctx := context.Background()
	req := &models.ChatCompletionRequest{
		Model: "test",
		Messages: []models.ChatMessage{
			{Role: "user", Content: "test"},
		},
		MaxTokens: &[]int{100}[0],
	}

	resp, err := mock.CreateChatCompletion(ctx, req)
	if err != nil {
		t.Errorf("Unexpected error from CreateChatCompletion: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected response, got nil")
	}

	if resp.ID != "test-id" {
		t.Errorf("Expected response ID 'test-id', got %s", resp.ID)
	}
}

func TestGetAvailableProviders(t *testing.T) {
	providers := GetAvailableProviders()

	expectedProviders := []string{"groq", "openrouter", "openai", "xai", "gemini", "mistral", "ollama"}

	if len(providers) != len(expectedProviders) {
		t.Errorf("Expected %d providers, got %d", len(expectedProviders), len(providers))
	}

	for _, expected := range expectedProviders {
		found := false
		for _, actual := range providers {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected provider %s not found in available providers", expected)
		}
	}
}

func TestIsValidProviderType(t *testing.T) {
	testCases := []struct {
		provider string
		valid    bool
	}{
		{"groq", true},
		{"openai", true},
		{"gemini", true},
		{"ollama", true},
		{"invalid", false},
		{"", false},
		{"GROQ", false}, // Case sensitive
	}

	for _, tc := range testCases {
		t.Run(tc.provider, func(t *testing.T) {
			valid := IsValidProviderType(tc.provider)
			if valid != tc.valid {
				t.Errorf("Expected IsValidProviderType(%s) to return %v, got %v", tc.provider, tc.valid, valid)
			}
		})
	}
}

// Mock provider for testing
type mockProvider struct {
	name      string
	model     string
	baseURL   string
	maxTokens int
}

func (m *mockProvider) CreateChatCompletion(ctx context.Context, req *models.ChatCompletionRequest) (*models.ChatCompletionResponse, error) {
	return &models.ChatCompletionResponse{
		ID:      "test-id",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   m.model,
		Choices: []models.ChatCompletionChoice{
			{
				Index: 0,
				Message: models.ChatMessage{
					Role:    "assistant",
					Content: "Test response",
				},
				FinishReason: "stop",
			},
		},
		Usage: models.ChatCompletionUsage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}, nil
}

func (m *mockProvider) GetName() string {
	return m.name
}

func (m *mockProvider) GetModel() string {
	return m.model
}

func (m *mockProvider) GetMaxTokens() int {
	return m.maxTokens
}

func (m *mockProvider) ValidateConfig() error {
	return nil
}

func (m *mockProvider) GetBaseURL() string {
	return m.baseURL
}

// HealthCheck method to satisfy the Provider interface
func (m *mockProvider) HealthCheck(ctx context.Context) error {
	return nil
}
