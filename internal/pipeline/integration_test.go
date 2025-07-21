package pipeline_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/orchestre-dev/ccproxy/internal/config"
	"github.com/orchestre-dev/ccproxy/internal/pipeline"
	"github.com/orchestre-dev/ccproxy/internal/providers"
	"github.com/orchestre-dev/ccproxy/internal/router"
	"github.com/orchestre-dev/ccproxy/internal/transformer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipelineIntegration(t *testing.T) {
	// Create test server to simulate provider
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request details
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.Header.Get("Content-Type"), "application/json")
		
		// Check provider-specific auth
		if strings.Contains(r.URL.Path, "/v1/messages") {
			// Anthropic
			assert.Equal(t, "test-anthropic-key", r.Header.Get("X-API-Key"))
			assert.Equal(t, "2023-06-01", r.Header.Get("anthropic-version"))
		} else if strings.Contains(r.URL.Path, "/v1/chat/completions") {
			// OpenAI
			assert.Equal(t, "Bearer test-openai-key", r.Header.Get("Authorization"))
		}
		
		// Read and validate request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		
		var reqBody map[string]interface{}
		err = json.Unmarshal(body, &reqBody)
		require.NoError(t, err)
		
		// Check that transformers have been applied
		// MaxToken transformer should have added max_tokens
		assert.Contains(t, reqBody, "max_tokens")
		
		// Parameters transformer should have validated temperature
		if temp, ok := reqBody["temperature"].(float64); ok {
			assert.True(t, temp >= 0 && temp <= 1)
		}
		
		// Send response
		response := map[string]interface{}{
			"id": "test-response",
			"type": "message",
			"role": "assistant",
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": "Hello from test provider!",
				},
			},
			"model": reqBody["model"],
			"usage": map[string]interface{}{
				"input_tokens": 10,
				"output_tokens": 5,
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create config
	cfg := &config.Config{
		Providers: []config.Provider{
			{
				Name:       "anthropic",
				APIBaseURL: server.URL,
				APIKey:     "test-anthropic-key",
				Enabled:    true,
				Models:     []string{"claude-3-opus"},
			},
			{
				Name:       "openai",
				APIBaseURL: server.URL,
				APIKey:     "test-openai-key",
				Enabled:    true,
				Models:     []string{"gpt-4"},
			},
		},
		Routes: map[string]config.Route{
			"claude-3-opus": {
				Provider: "anthropic",
				Model:    "claude-3-opus",
			},
			"gpt-4": {
				Provider: "openai",
				Model:    "gpt-4",
			},
		},
	}

	// Create services
	configService := config.NewService()
	configService.SetConfig(cfg)
	providerService := providers.NewService(configService)
	providerService.Initialize() // Initialize providers
	transformerService := transformer.GetRegistry()
	routerService := router.New(cfg)
	
	// Create pipeline
	pipelineService := pipeline.NewPipeline(cfg, providerService, transformerService, routerService)
	
	// Test cases
	tests := []struct {
		name         string
		request      map[string]interface{}
		expectedPath string
		expectedAuth string
	}{
		{
			name: "anthropic request",
			request: map[string]interface{}{
				"model": "claude-3-opus",
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
				},
				"temperature": 0.7,
			},
			expectedPath: "/v1/messages",
			expectedAuth: "X-API-Key",
		},
		{
			name: "openai request",
			request: map[string]interface{}{
				"model": "gpt-4",
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
				},
				"temperature": 0.8,
			},
			expectedPath: "/v1/chat/completions",
			expectedAuth: "Authorization",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request context
			reqCtx := &pipeline.RequestContext{
				Body:        tt.request,
				Headers:     make(map[string]string),
				IsStreaming: false,
			}
			
			// Process request
			ctx := context.Background()
			respCtx, err := pipelineService.ProcessRequest(ctx, reqCtx)
			require.NoError(t, err)
			
			// Verify response
			assert.NotNil(t, respCtx)
			assert.NotNil(t, respCtx.Response)
			assert.Equal(t, 200, respCtx.Response.StatusCode)
			
			// Read response body
			body, err := io.ReadAll(respCtx.Response.Body)
			require.NoError(t, err)
			
			var respBody map[string]interface{}
			err = json.Unmarshal(body, &respBody)
			require.NoError(t, err)
			
			// Verify response content
			assert.Equal(t, "test-response", respBody["id"])
			assert.Equal(t, tt.request["model"], respBody["model"])
			
			// Verify usage was preserved/enhanced by MaxToken transformer
			usage, ok := respBody["usage"].(map[string]interface{})
			assert.True(t, ok)
			assert.Contains(t, usage, "total_tokens")
		})
	}
}

func TestPipelineStreamingIntegration(t *testing.T) {
	// Create test SSE server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		
		// Send SSE events
		events := []string{
			`data: {"id":"test","type":"message_start"}`,
			`data: {"type":"content_block_delta","delta":{"text":"Hello"}}`,
			`data: {"type":"content_block_delta","delta":{"text":" world!"}}`,
			`data: {"type":"message_stop"}`,
			`data: [DONE]`,
		}
		
		for _, event := range events {
			w.Write([]byte(event + "\n\n"))
			w.(http.Flusher).Flush()
		}
	}))
	defer server.Close()

	// Create minimal config
	cfg := &config.Config{
		Providers: []config.Provider{
			{
				Name:       "anthropic",
				APIBaseURL: server.URL,
				APIKey:     "test-key",
				Enabled:    true,
				Models:     []string{"claude-3-opus"},
			},
		},
		Routes: map[string]config.Route{
			"claude-3-opus": {
				Provider: "anthropic",
				Model:    "claude-3-opus",
			},
		},
	}

	// Create services
	configService := config.NewService()
	configService.SetConfig(cfg)
	providerService := providers.NewService(configService)
	providerService.Initialize() // Initialize providers
	transformerService := transformer.GetRegistry()
	routerService := router.New(cfg)

	// Create pipeline
	pipelineService := pipeline.NewPipeline(cfg, providerService, transformerService, routerService)

	// Create streaming request
	reqCtx := &pipeline.RequestContext{
		Body: map[string]interface{}{
			"model": "claude-3-opus",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Hello",
				},
			},
			"stream": true,
		},
		Headers:     make(map[string]string),
		IsStreaming: true,
	}

	// Process request
	ctx := context.Background()
	respCtx, err := pipelineService.ProcessRequest(ctx, reqCtx)
	require.NoError(t, err)

	// Verify streaming response
	assert.NotNil(t, respCtx)
	assert.NotNil(t, respCtx.Response)
	assert.Equal(t, "text/event-stream", respCtx.Response.Header.Get("Content-Type"))

	// Create response writer to capture stream
	recorder := httptest.NewRecorder()

	// Stream response
	err = pipelineService.StreamResponse(ctx, recorder, respCtx)
	require.NoError(t, err)

	// Verify streamed content
	result := recorder.Body.String()
	assert.Contains(t, result, "data:")
	assert.Contains(t, result, "[DONE]")
}

func TestPipelineErrorHandling(t *testing.T) {
	// Create config with invalid provider
	cfg := &config.Config{
		Providers: []config.Provider{},
		Routes: map[string]config.Route{
			"test-model": {
				Provider: "non-existent",
				Model:    "test-model",
			},
		},
	}
	
	// Create services
	configService := config.NewService()
	configService.SetConfig(cfg)
	providerService := providers.NewService(configService)
	providerService.Initialize() // Initialize providers
	transformerService := transformer.GetRegistry()
	routerService := router.New(cfg)
	
	// Create pipeline
	pipelineService := pipeline.NewPipeline(cfg, providerService, transformerService, routerService)
	
	// Create request
	reqCtx := &pipeline.RequestContext{
		Body: map[string]interface{}{
			"model": "test-model",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Hello",
				},
			},
		},
		Headers:     make(map[string]string),
		IsStreaming: false,
	}
	
	// Process request - should fail
	ctx := context.Background()
	_, err := pipelineService.ProcessRequest(ctx, reqCtx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider not found")
}