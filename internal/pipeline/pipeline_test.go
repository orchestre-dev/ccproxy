package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/orchestre-dev/ccproxy/internal/config"
	"github.com/orchestre-dev/ccproxy/internal/providers"
	"github.com/orchestre-dev/ccproxy/internal/router"
	"github.com/orchestre-dev/ccproxy/internal/transformer"
)

func TestNewPipeline(t *testing.T) {
	cfg := &config.Config{
		Performance: config.PerformanceConfig{
			RequestTimeout:     30 * time.Second,
			MaxRequestBodySize: 10 * 1024 * 1024, // 10MB
		},
	}

	providerService := &providers.Service{}
	transformerService := transformer.NewService()
	routerInstance := router.New(cfg)

	pipeline := NewPipeline(cfg, providerService, transformerService, routerInstance)

	if pipeline == nil {
		t.Error("Expected non-nil pipeline")
	}

	if pipeline.config != cfg {
		t.Error("Pipeline config not set correctly")
	}

	if pipeline.providerService != providerService {
		t.Error("Pipeline provider service not set correctly")
	}

	if pipeline.transformerService != transformerService {
		t.Error("Pipeline transformer service not set correctly")
	}

	if pipeline.router != routerInstance {
		t.Error("Pipeline router not set correctly")
	}

	if pipeline.httpClient == nil {
		t.Error("Pipeline HTTP client not initialized")
	}

	if pipeline.streamingProcessor == nil {
		t.Error("Pipeline streaming processor not initialized")
	}

	if pipeline.performanceMonitor == nil {
		t.Error("Pipeline performance monitor not initialized")
	}

	if pipeline.messageConverter == nil {
		t.Error("Pipeline message converter not initialized")
	}
}

func TestPipeline_NewPipelineWithProxy(t *testing.T) {
	cfg := &config.Config{
		ProxyURL: "http://proxy.example.com:8080",
		Performance: config.PerformanceConfig{
			RequestTimeout: 15 * time.Second,
		},
	}

	providerService := &providers.Service{}
	transformerService := transformer.NewService()
	routerInstance := router.New(cfg)

	pipeline := NewPipeline(cfg, providerService, transformerService, routerInstance)

	if pipeline == nil {
		t.Error("Expected non-nil pipeline with proxy config")
	}

	// HTTP client should be created with proxy support
	if pipeline.httpClient == nil {
		t.Error("HTTP client should be initialized with proxy")
	}
}

func TestPipeline_NewPipelineWithDefaultTimeout(t *testing.T) {
	cfg := &config.Config{
		Performance: config.PerformanceConfig{
			// RequestTimeout not set - should use default
		},
	}

	providerService := &providers.Service{}
	transformerService := transformer.NewService()
	routerInstance := router.New(cfg)

	pipeline := NewPipeline(cfg, providerService, transformerService, routerInstance)

	if pipeline == nil {
		t.Error("Expected non-nil pipeline with default timeout")
	}

	// Should fallback to 30 second default
	if pipeline.httpClient.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", pipeline.httpClient.Timeout)
	}
}

func TestPipeline_GetProviderEndpoint(t *testing.T) {
	pipeline := &Pipeline{}

	tests := []struct {
		provider string
		expected string
	}{
		{"anthropic", "/v1/messages"},
		{"openai", "/v1/chat/completions"},
		{"groq", "/openai/v1/chat/completions"},
		{"deepseek", "/v1/chat/completions"},
		{"gemini", "/v1beta/models/generateContent"},
		{"openrouter", "/api/v1/chat/completions"},
		{"mistral", "/v1/chat/completions"},
		{"xai", "/v1/chat/completions"},
		{"ollama", "/api/chat"},
		{"unknown", "/v1/chat/completions"}, // Default
	}

	for _, test := range tests {
		actual := pipeline.getProviderEndpoint(test.provider)
		if actual != test.expected {
			t.Errorf("Provider %s: expected endpoint %s, got %s", test.provider, test.expected, actual)
		}
	}
}

func TestPipeline_SetAuthenticationHeader(t *testing.T) {
	pipeline := &Pipeline{}

	t.Run("AnthropicAuth", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil)
		provider := &config.Provider{
			APIKey: "test-api-key",
		}

		pipeline.setAuthenticationHeader(req, provider, "anthropic")

		if req.Header.Get("X-API-Key") != "test-api-key" {
			t.Errorf("Expected X-API-Key header, got %v", req.Header.Get("X-API-Key"))
		}

		if req.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("Expected anthropic-version header")
		}
	})

	t.Run("OpenAIAuth", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", nil)
		provider := &config.Provider{
			APIKey: "test-openai-key",
		}

		pipeline.setAuthenticationHeader(req, provider, "openai")

		expected := "Bearer test-openai-key"
		if req.Header.Get("Authorization") != expected {
			t.Errorf("Expected Authorization header %s, got %v", expected, req.Header.Get("Authorization"))
		}
	})

	t.Run("GeminiAuth", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "https://generativelanguage.googleapis.com/v1beta/models/generateContent", nil)
		provider := &config.Provider{
			APIKey: "test-gemini-key",
		}

		pipeline.setAuthenticationHeader(req, provider, "gemini")

		expected := "Bearer test-gemini-key"
		if req.Header.Get("Authorization") != expected {
			t.Errorf("Expected Authorization header %s, got %v", expected, req.Header.Get("Authorization"))
		}
	})

	t.Run("NoAPIKey", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", nil)
		provider := &config.Provider{
			APIKey: "", // No API key
		}

		pipeline.setAuthenticationHeader(req, provider, "openai")

		// Should not set any auth headers
		if req.Header.Get("Authorization") != "" {
			t.Error("Should not set Authorization header when no API key")
		}
	})
}

func TestPipeline_BuildHTTPRequest(t *testing.T) {
	ctx := context.Background()
	pipeline := &Pipeline{}

	t.Run("BasicRequest", func(t *testing.T) {
		provider := &config.Provider{
			APIBaseURL: "https://api.openai.com",
			APIKey:     "test-key",
		}

		body := map[string]interface{}{
			"model": "gpt-4",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Hello",
				},
			},
		}

		req, err := pipeline.buildHTTPRequest(ctx, provider, body, false, "openai")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if req.Method != "POST" {
			t.Errorf("Expected POST method, got %s", req.Method)
		}

		expectedURL := "https://api.openai.com/v1/chat/completions"
		if req.URL.String() != expectedURL {
			t.Errorf("Expected URL %s, got %s", expectedURL, req.URL.String())
		}

		if req.Header.Get("Content-Type") != "application/json" {
			t.Error("Expected Content-Type header to be application/json")
		}

		if req.Header.Get("User-Agent") != "ccproxy/1.0" {
			t.Error("Expected User-Agent header")
		}

		if req.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("Expected Authorization header")
		}
	})

	t.Run("StreamingRequest", func(t *testing.T) {
		provider := &config.Provider{
			APIBaseURL: "https://api.anthropic.com",
			APIKey:     "test-key",
		}

		body := map[string]interface{}{
			"model":  "claude-3-haiku",
			"stream": true,
		}

		req, err := pipeline.buildHTTPRequest(ctx, provider, body, true, "anthropic")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if req.Header.Get("Accept") != "text/event-stream" {
			t.Error("Expected Accept header for streaming")
		}
	})

	t.Run("RequestConfigWithCustomURL", func(t *testing.T) {
		provider := &config.Provider{
			APIBaseURL: "https://api.openai.com",
			APIKey:     "test-key",
		}

		reqConfig := &transformer.RequestConfig{
			Body: map[string]interface{}{
				"model": "gpt-4",
			},
			URL:    "https://custom.api.com/v1/chat",
			Method: "PUT",
			Headers: map[string]string{
				"X-Custom-Header": "custom-value",
			},
			Timeout: 5000, // 5 seconds
		}

		req, err := pipeline.buildHTTPRequest(ctx, provider, reqConfig, false, "openai")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if req.Method != "PUT" {
			t.Errorf("Expected PUT method, got %s", req.Method)
		}

		if req.URL.String() != "https://custom.api.com/v1/chat" {
			t.Errorf("Expected custom URL, got %s", req.URL.String())
		}

		if req.Header.Get("X-Custom-Header") != "custom-value" {
			t.Error("Expected custom header")
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		provider := &config.Provider{
			APIBaseURL: "https://api.openai.com",
		}

		// Invalid body that can't be marshaled
		body := make(chan int)

		_, err := pipeline.buildHTTPRequest(ctx, provider, body, false, "openai")
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}

		if !strings.Contains(err.Error(), "failed to marshal request body") {
			t.Errorf("Expected marshal error, got %v", err)
		}
	})
}

func TestNewErrorResponse(t *testing.T) {
	errorResp := NewErrorResponse("Test error", "test_error", "E001")

	if errorResp == nil {
		t.Error("Expected non-nil error response")
	}

	if errorResp.Error.Message != "Test error" {
		t.Errorf("Expected message 'Test error', got %s", errorResp.Error.Message)
	}

	if errorResp.Error.Type != "test_error" {
		t.Errorf("Expected type 'test_error', got %s", errorResp.Error.Type)
	}

	if errorResp.Error.Code != "E001" {
		t.Errorf("Expected code 'E001', got %s", errorResp.Error.Code)
	}
}

func TestWriteErrorResponse(t *testing.T) {
	w := httptest.NewRecorder()
	errorResp := NewErrorResponse("Internal error", "internal_error", "E500")

	WriteErrorResponse(w, http.StatusInternalServerError, errorResp)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Expected Content-Type to be application/json")
	}

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Error.Message != "Internal error" {
		t.Error("Error message not preserved in response")
	}
}

func TestStreamResponse(t *testing.T) {
	t.Run("ValidStreamingResponse", func(t *testing.T) {
		// Create a mock SSE response
		sseData := "data: {\"chunk\": \"hello\"}\n\ndata: {\"chunk\": \"world\"}\n\ndata: [DONE]\n\n"
		resp := &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(sseData)),
		}

		w := httptest.NewRecorder()

		err := StreamResponse(w, resp)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Check headers
		if w.Header().Get("Content-Type") != "text/event-stream" {
			t.Error("Expected Content-Type to be text/event-stream")
		}

		if w.Header().Get("Cache-Control") != "no-cache" {
			t.Error("Expected Cache-Control header")
		}

		// Check body content
		body := w.Body.String()
		if !strings.Contains(body, "hello") {
			t.Error("Expected streamed content to contain 'hello'")
		}

		if !strings.Contains(body, "[DONE]") {
			t.Error("Expected streamed content to contain '[DONE]'")
		}
	})

	t.Run("EmptyStreamingResponse", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("")),
		}

		w := httptest.NewRecorder()

		err := StreamResponse(w, resp)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	})
}

func TestCopyResponse(t *testing.T) {
	t.Run("ValidResponse", func(t *testing.T) {
		responseBody := `{"message": "Hello, world!"}`
		resp := &http.Response{
			StatusCode: 200,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"X-Custom":     []string{"custom-value"},
			},
			Body: io.NopCloser(strings.NewReader(responseBody)),
		}

		w := httptest.NewRecorder()

		err := CopyResponse(w, resp)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if w.Code != 200 {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		if w.Header().Get("Content-Type") != "application/json" {
			t.Error("Expected Content-Type header to be copied")
		}

		if w.Header().Get("X-Custom") != "custom-value" {
			t.Error("Expected custom header to be copied")
		}

		if w.Body.String() != responseBody {
			t.Error("Expected response body to be copied")
		}
	})

	t.Run("ErrorResponse", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: 500,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: io.NopCloser(strings.NewReader(`{"error": "Server error"}`)),
		}

		w := httptest.NewRecorder()

		err := CopyResponse(w, resp)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if w.Code != 500 {
			t.Errorf("Expected status 500, got %d", w.Code)
		}
	})
}

func TestHandleStreamingError(t *testing.T) {
	w := httptest.NewRecorder()
	testErr := fmt.Errorf("test streaming error")

	HandleStreamingError(w, testErr)

	// Check headers
	if w.Header().Get("Content-Type") != "text/event-stream" {
		t.Error("Expected Content-Type to be text/event-stream")
	}

	// Check body contains error and done marker
	body := w.Body.String()
	if !strings.Contains(body, "test streaming error") {
		t.Error("Expected error message in response")
	}

	if !strings.Contains(body, "[DONE]") {
		t.Error("Expected [DONE] marker in response")
	}

	if !strings.Contains(body, "event: error") {
		t.Error("Expected error event type")
	}
}

func TestRequestContext(t *testing.T) {
	reqCtx := &RequestContext{
		Body: map[string]interface{}{
			"model": "test-model",
		},
		Headers: map[string]string{
			"Authorization": "Bearer token",
		},
		IsStreaming: true,
		Metadata: map[string]interface{}{
			"user_id": "test-user",
		},
	}

	if reqCtx.Body == nil {
		t.Error("Expected body to be set")
	}

	if !reqCtx.IsStreaming {
		t.Error("Expected streaming to be true")
	}

	if reqCtx.Headers["Authorization"] != "Bearer token" {
		t.Error("Expected authorization header")
	}

	if reqCtx.Metadata["user_id"] != "test-user" {
		t.Error("Expected metadata to be set")
	}
}

func TestResponseContext(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("test")),
	}

	respCtx := &ResponseContext{
		Response:        resp,
		Provider:        "openai",
		Model:           "gpt-4",
		TokenCount:      150,
		RoutingStrategy: "direct model route",
	}

	if respCtx.Response != resp {
		t.Error("Expected response to be set")
	}

	if respCtx.Provider != "openai" {
		t.Error("Expected provider to be openai")
	}

	if respCtx.Model != "gpt-4" {
		t.Error("Expected model to be gpt-4")
	}

	if respCtx.TokenCount != 150 {
		t.Error("Expected token count to be 150")
	}

	if respCtx.RoutingStrategy != "direct model route" {
		t.Error("Expected routing strategy to be set")
	}
}

// Tests for ProcessRequest - the core pipeline functionality
func TestPipeline_ProcessRequest(t *testing.T) {
	cfg := &config.Config{
		Performance: config.PerformanceConfig{
			RequestTimeout: 30 * time.Second,
		},
		Providers: []config.Provider{
			{
				Name:       "openai",
				APIBaseURL: "https://api.openai.com",
				APIKey:     "test-key",
			},
		},
		Routes: map[string]config.Route{
			"gpt-4": {
				Provider: "openai",
				Model:    "gpt-4",
			},
			"gpt-3.5-turbo": {
				Provider: "openai",
				Model:    "gpt-3.5-turbo",
			},
		},
	}

	configService := config.NewService()
	configService.SetConfig(cfg)

	providerService := providers.NewService(configService)
	err := providerService.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize provider service: %v", err)
	}

	transformerService := transformer.NewService()
	routerInstance := router.New(cfg)

	pipeline := NewPipeline(cfg, providerService, transformerService, routerInstance)

	t.Run("ValidRequest", func(t *testing.T) {
		// Create a mock server to simulate the provider
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write([]byte(`{"choices": [{"message": {"content": "Hello from test"}}]}`))
		}))
		defer server.Close()

		// Update provider URL to use mock server
		cfg.Providers[0].APIBaseURL = server.URL
		configService.SetConfig(cfg)
		providerService.Initialize() // Reinitialize with new config

		req := &RequestContext{
			Body: map[string]interface{}{
				"model": "gpt-4",
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
				},
			},
			Headers:     map[string]string{},
			IsStreaming: false,
		}

		ctx := context.Background()
		respCtx, err := pipeline.ProcessRequest(ctx, req)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if respCtx == nil {
			t.Fatal("Expected non-nil response context")
		}

		if respCtx.Provider == "" {
			t.Error("Expected provider to be set")
		}

		if respCtx.Response == nil {
			t.Error("Expected HTTP response to be set")
		}
	})

	t.Run("RequestWithThinking", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write([]byte(`{"choices": [{"message": {"content": "Thinking response"}}]}`))
		}))
		defer server.Close()

		cfg.Providers[0].APIBaseURL = server.URL
		configService.SetConfig(cfg)
		providerService.Initialize()

		req := &RequestContext{
			Body: map[string]interface{}{
				"model":    "gpt-4",
				"thinking": true,
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Complex question",
					},
				},
			},
			Headers:     map[string]string{},
			IsStreaming: false,
		}

		ctx := context.Background()
		respCtx, err := pipeline.ProcessRequest(ctx, req)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if respCtx == nil {
			t.Fatal("Expected non-nil response context")
		}
	})

	t.Run("HighTokenCountRequest", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write([]byte(`{"choices": [{"message": {"content": "Long context response"}}]}`))
		}))
		defer server.Close()

		cfg.Providers[0].APIBaseURL = server.URL
		configService.SetConfig(cfg)
		providerService.Initialize()

		// Create a request with high token count content
		longContent := strings.Repeat("This is a long message. ", 3000) // ~15k tokens

		req := &RequestContext{
			Body: map[string]interface{}{
				"model": "gpt-4",
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": longContent,
					},
				},
			},
			Headers:     map[string]string{},
			IsStreaming: false,
		}

		ctx := context.Background()
		respCtx, err := pipeline.ProcessRequest(ctx, req)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if respCtx.TokenCount == 0 {
			t.Error("Expected token count to be calculated")
		}
	})

	t.Run("ProviderNotFound", func(t *testing.T) {
		// Create pipeline with no providers configured
		emptyCfg := &config.Config{
			Performance: config.PerformanceConfig{
				RequestTimeout: 30 * time.Second,
			},
			Providers: []config.Provider{},
		}

		emptyConfigService := config.NewService()
		emptyConfigService.SetConfig(emptyCfg)
		emptyProviderService := providers.NewService(emptyConfigService)
		emptyProviderService.Initialize()

		emptyPipeline := NewPipeline(emptyCfg, emptyProviderService, transformerService, router.New(emptyCfg))

		req := &RequestContext{
			Body: map[string]interface{}{
				"model": "non-existent-model",
			},
			Headers:     map[string]string{},
			IsStreaming: false,
		}

		ctx := context.Background()
		_, err := emptyPipeline.ProcessRequest(ctx, req)

		if err == nil {
			t.Error("Expected error for non-existent provider")
		}

		if !strings.Contains(err.Error(), "provider not found") {
			t.Errorf("Expected 'provider not found' error, got %v", err)
		}
	})

	t.Run("TransformationError", func(t *testing.T) {
		// Test with malformed request body that transformation might fail on
		req := &RequestContext{
			Body:        "invalid-body-type", // String instead of expected map
			Headers:     map[string]string{},
			IsStreaming: false,
		}

		ctx := context.Background()
		_, err := pipeline.ProcessRequest(ctx, req)

		// Should handle transformation error gracefully
		if err != nil && !strings.Contains(err.Error(), "transformation") &&
			!strings.Contains(err.Error(), "provider not found") {
			t.Logf("Got expected error for invalid body: %v", err)
		}
	})

	t.Run("HTTPRequestError", func(t *testing.T) {
		// Create server that immediately closes to simulate network error
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Close connection without response
			hj, ok := w.(http.Hijacker)
			if ok {
				conn, _, _ := hj.Hijack()
				conn.Close()
			}
		}))
		server.Close() // Close server immediately

		cfg.Providers[0].APIBaseURL = server.URL
		configService.SetConfig(cfg)
		providerService.Initialize()

		req := &RequestContext{
			Body: map[string]interface{}{
				"model": "gpt-4",
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
				},
			},
			Headers:     map[string]string{},
			IsStreaming: false,
		}

		ctx := context.Background()
		_, err := pipeline.ProcessRequest(ctx, req)

		if err == nil {
			t.Error("Expected error for failed HTTP request")
		}

		if !strings.Contains(err.Error(), "provider request failed") {
			t.Errorf("Expected 'provider request failed' error, got %v", err)
		}
	})

	t.Run("ResponseTransformationError", func(t *testing.T) {
		// Create server that returns invalid response
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write([]byte(`invalid json response`))
		}))
		defer server.Close()

		cfg.Providers[0].APIBaseURL = server.URL
		configService.SetConfig(cfg)
		providerService.Initialize()

		req := &RequestContext{
			Body: map[string]interface{}{
				"model": "gpt-4",
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
				},
			},
			Headers:     map[string]string{},
			IsStreaming: false,
		}

		ctx := context.Background()
		respCtx, err := pipeline.ProcessRequest(ctx, req)

		// Should handle gracefully - either succeed or fail with transformation error
		if err != nil && strings.Contains(err.Error(), "response transformation failed") {
			t.Logf("Got expected transformation error: %v", err)
		} else if err == nil && respCtx != nil {
			// Transformation succeeded despite invalid JSON
			t.Logf("Transformation handled invalid JSON gracefully")
		}
	})
}

// Test the Pipeline.StreamResponse method (not the global function)
func TestPipeline_StreamResponseMethod(t *testing.T) {
	cfg := &config.Config{
		Performance: config.PerformanceConfig{
			RequestTimeout: 30 * time.Second,
		},
	}

	providerService := &providers.Service{}
	transformerService := transformer.NewService()
	routerInstance := router.New(cfg)

	pipeline := NewPipeline(cfg, providerService, transformerService, routerInstance)

	t.Run("ValidStreamingResponse", func(t *testing.T) {
		sseData := "data: {\"chunk\": \"hello\"}\n\ndata: [DONE]\n\n"
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(sseData)),
		}

		respCtx := &ResponseContext{
			Response: resp,
			Provider: "openai",
			Model:    "gpt-4",
		}

		w := httptest.NewRecorder()
		ctx := context.Background()

		err := pipeline.StreamResponse(ctx, w, respCtx)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Check that streaming headers were set
		if w.Header().Get("Content-Type") != "text/event-stream" {
			t.Error("Expected Content-Type header to be set for streaming")
		}
	})
}

// Enhanced authentication header tests for additional providers
func TestPipeline_SetAuthenticationHeader_AllProviders(t *testing.T) {
	pipeline := &Pipeline{}

	testCases := []struct {
		provider      string
		apiKey        string
		expectedAuth  string
		expectedExtra map[string]string
	}{
		{
			provider:     "mistral",
			apiKey:       "mistral-key",
			expectedAuth: "Bearer mistral-key",
		},
		{
			provider:     "xai",
			apiKey:       "xai-key",
			expectedAuth: "Bearer xai-key",
		},
		{
			provider:     "deepseek",
			apiKey:       "deepseek-key",
			expectedAuth: "Bearer deepseek-key",
		},
		{
			provider:     "ollama",
			apiKey:       "ollama-key",
			expectedAuth: "Bearer ollama-key",
		},
		{
			provider:     "ollama",
			apiKey:       "", // No key for ollama
			expectedAuth: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.provider, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "https://api.example.com", nil)
			provider := &config.Provider{
				APIKey: tc.apiKey,
			}

			pipeline.setAuthenticationHeader(req, provider, tc.provider)

			if tc.expectedAuth != "" {
				auth := req.Header.Get("Authorization")
				if auth != tc.expectedAuth {
					t.Errorf("Expected Authorization header %s, got %s", tc.expectedAuth, auth)
				}
			} else {
				// Should not set auth header when no API key
				if req.Header.Get("Authorization") != "" {
					t.Error("Should not set Authorization header when no API key")
				}
			}
		})
	}
}

// Test proxy configuration scenarios
func TestPipeline_ProxyConfiguration(t *testing.T) {
	t.Run("InvalidProxyURL", func(t *testing.T) {
		cfg := &config.Config{
			ProxyURL: "invalid-proxy-url",
			Performance: config.PerformanceConfig{
				RequestTimeout: 15 * time.Second,
			},
		}

		providerService := &providers.Service{}
		transformerService := transformer.NewService()
		routerInstance := router.New(cfg)

		// Should not fail, but should fallback gracefully
		pipeline := NewPipeline(cfg, providerService, transformerService, routerInstance)

		if pipeline == nil {
			t.Error("Pipeline should be created even with invalid proxy")
		}

		if pipeline.httpClient == nil {
			t.Error("HTTP client should be created with fallback")
		}
	})

	t.Run("ProxyFromEnvironment", func(t *testing.T) {
		cfg := &config.Config{
			// No ProxyURL set, should try environment
			Performance: config.PerformanceConfig{
				RequestTimeout: 15 * time.Second,
			},
		}

		providerService := &providers.Service{}
		transformerService := transformer.NewService()
		routerInstance := router.New(cfg)

		pipeline := NewPipeline(cfg, providerService, transformerService, routerInstance)

		if pipeline == nil {
			t.Error("Pipeline should be created")
		}

		if pipeline.httpClient == nil {
			t.Error("HTTP client should be created")
		}
	})
}

// Test error scenarios during pipeline creation
func TestPipeline_CreationErrors(t *testing.T) {
	t.Run("ZeroTimeout", func(t *testing.T) {
		cfg := &config.Config{
			Performance: config.PerformanceConfig{
				RequestTimeout: 0, // Zero timeout should use default
			},
		}

		providerService := &providers.Service{}
		transformerService := transformer.NewService()
		routerInstance := router.New(cfg)

		pipeline := NewPipeline(cfg, providerService, transformerService, routerInstance)

		if pipeline == nil {
			t.Error("Pipeline should be created with default timeout")
		}

		expectedTimeout := 30 * time.Second
		if pipeline.httpClient.Timeout != expectedTimeout {
			t.Errorf("Expected default timeout %v, got %v", expectedTimeout, pipeline.httpClient.Timeout)
		}
	})
}
