package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"

	"ccproxy/internal/config"
	"ccproxy/internal/models"
	"ccproxy/pkg/logger"
)

// BenchmarkProvider implements a minimal provider for benchmarking
type BenchmarkProvider struct {
	name      string
	model     string
	maxTokens int
}

func (p *BenchmarkProvider) CreateChatCompletion(ctx context.Context, req *models.ChatCompletionRequest) (*models.ChatCompletionResponse, error) {
	// Simulate minimal processing
	return &models.ChatCompletionResponse{
		ID:      "bench-123",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   p.model,
		Choices: []models.ChatCompletionChoice{
			{
				Index: 0,
				Message: models.ChatMessage{
					Role:    "assistant",
					Content: "Benchmark response",
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

func (p *BenchmarkProvider) GetName() string        { return p.name }
func (p *BenchmarkProvider) GetModel() string       { return p.model }
func (p *BenchmarkProvider) GetMaxTokens() int      { return p.maxTokens }
func (p *BenchmarkProvider) ValidateConfig() error  { return nil }
func (p *BenchmarkProvider) GetBaseURL() string     { return "http://benchmark" }
func (p *BenchmarkProvider) HealthCheck(ctx context.Context) error { return nil }

// setupBenchmarkHandler creates a handler for benchmarking
func setupBenchmarkHandler() *Handler {
	gin.SetMode(gin.ReleaseMode)
	log := logger.New(config.LoggingConfig{Level: "error", Format: "json"})
	provider := &BenchmarkProvider{name: "benchmark", model: "bench-model", maxTokens: 1000}
	return &Handler{
		provider: provider,
		logger:   log,
	}
}

// BenchmarkProxyMessages_SimpleRequest benchmarks a simple message request
func BenchmarkProxyMessages_SimpleRequest(b *testing.B) {
	handler := setupBenchmarkHandler()
	router := gin.New()
	router.POST("/v1/messages", handler.ProxyMessages)

	maxTokens := 100
	requestBody := models.MessagesRequest{
		Model: "bench-model",
		Messages: []models.Message{
			{Role: "user", Content: "Hello!"},
		},
		MaxTokens: &maxTokens,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", w.Code)
		}
	}
}

// BenchmarkProxyMessages_LargePayload benchmarks handling of large messages
func BenchmarkProxyMessages_LargePayload(b *testing.B) {
	handler := setupBenchmarkHandler()
	router := gin.New()
	router.POST("/v1/messages", handler.ProxyMessages)

	// Create a large message payload
	largeContent := strings.Repeat("This is a test message. ", 1000) // ~24KB
	maxTokens := 1000
	requestBody := models.MessagesRequest{
		Model: "bench-model",
		Messages: []models.Message{
			{Role: "user", Content: largeContent},
		},
		MaxTokens: &maxTokens,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", w.Code)
		}
	}
}

// BenchmarkProxyMessages_MultiTurn benchmarks multi-turn conversations
func BenchmarkProxyMessages_MultiTurn(b *testing.B) {
	handler := setupBenchmarkHandler()
	router := gin.New()
	router.POST("/v1/messages", handler.ProxyMessages)

	// Create a multi-turn conversation
	messages := make([]models.Message, 10)
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			messages[i] = models.Message{
				Role:    "user",
				Content: fmt.Sprintf("User message %d", i),
			}
		} else {
			messages[i] = models.Message{
				Role:    "assistant",
				Content: fmt.Sprintf("Assistant response %d", i),
			}
		}
	}

	maxTokens := 500
	requestBody := models.MessagesRequest{
		Model:     "bench-model",
		Messages:  messages,
		MaxTokens: &maxTokens,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", w.Code)
		}
	}
}

// BenchmarkProxyMessages_Concurrent benchmarks concurrent request handling
func BenchmarkProxyMessages_Concurrent(b *testing.B) {
	handler := setupBenchmarkHandler()
	router := gin.New()
	router.POST("/v1/messages", handler.ProxyMessages)

	maxTokens := 100
	requestBody := models.MessagesRequest{
		Model: "bench-model",
		Messages: []models.Message{
			{Role: "user", Content: "Concurrent test"},
		},
		MaxTokens: &maxTokens,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-API-Key", "test-key")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			if w.Code != http.StatusOK {
				b.Fatalf("Expected status 200, got %d", w.Code)
			}
		}
	})
}

// BenchmarkHealthCheck benchmarks health check endpoint
func BenchmarkHealthCheck(b *testing.B) {
	handler := setupBenchmarkHandler()
	router := gin.New()
	router.GET("/health", handler.HealthCheck)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", w.Code)
		}
	}
}

// BenchmarkJSONMarshaling benchmarks JSON encoding/decoding performance
func BenchmarkJSONMarshaling(b *testing.B) {
	b.Run("Marshal_AnthropicRequest", func(b *testing.B) {
		maxTokens := 100
		req := models.MessagesRequest{
			Model: "test-model",
			Messages: []models.Message{
				{Role: "user", Content: "Test message"},
				{Role: "assistant", Content: "Test response"},
			},
			MaxTokens:   &maxTokens,
			Temperature: floatPtr(0.7),
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := json.Marshal(req)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Unmarshal_AnthropicRequest", func(b *testing.B) {
		data := []byte(`{
			"model": "test-model",
			"messages": [
				{"role": "user", "content": "Test message"},
				{"role": "assistant", "content": "Test response"}
			],
			"max_tokens": 100,
			"temperature": 0.7
		}`)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var req models.MessagesRequest
			err := json.Unmarshal(data, &req)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkErrorHandling benchmarks error response generation
func BenchmarkErrorHandling(b *testing.B) {
	handler := setupBenchmarkHandler()
	router := gin.New()
	router.POST("/v1/messages", handler.ProxyMessages)

	// Invalid JSON to trigger error handling
	invalidBody := []byte(`{"invalid json"`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(invalidBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusBadRequest {
			b.Fatalf("Expected status 400, got %d", w.Code)
		}
	}
}

// BenchmarkMemoryAllocation tracks memory allocations
func BenchmarkMemoryAllocation(b *testing.B) {
	handler := setupBenchmarkHandler()
	router := gin.New()
	router.POST("/v1/messages", handler.ProxyMessages)

	maxTokens := 100
	requestBody := models.MessagesRequest{
		Model: "bench-model",
		Messages: []models.Message{
			{Role: "user", Content: "Memory test"},
		},
		MaxTokens: &maxTokens,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkHighConcurrency tests performance under high concurrent load
func BenchmarkHighConcurrency(b *testing.B) {
	handler := setupBenchmarkHandler()
	router := gin.New()
	router.POST("/v1/messages", handler.ProxyMessages)

	maxTokens := 50
	requestBody := models.MessagesRequest{
		Model: "bench-model",
		Messages: []models.Message{
			{Role: "user", Content: "High concurrency test"},
		},
		MaxTokens: &maxTokens,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	// Test with 100 concurrent goroutines
	concurrency := 100
	b.ResetTimer()
	
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)
	
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			
			req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-API-Key", "test-key")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}()
	}
	
	wg.Wait()
}

func floatPtr(f float64) *float64 {
	return &f
}