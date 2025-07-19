package pipeline

import (
	"context"
	"fmt"
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

func TestPipeline_StreamingIntegration(t *testing.T) {
	// Create a mock provider server that streams SSE
	providerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-key" {
			t.Errorf("Expected Authorization header, got %s", auth)
		}

		// Send SSE response
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
		
		// Stream some events
		events := []string{
			`data: {"id":"msg_123","type":"content_block_start","content_block":{"type":"text","text":""}}`,
			`data: {"id":"msg_123","type":"content_block_delta","delta":{"type":"text_delta","text":"Hello"}}`,
			`data: {"id":"msg_123","type":"content_block_delta","delta":{"type":"text_delta","text":" streaming"}}`,
			`data: {"id":"msg_123","type":"content_block_delta","delta":{"type":"text_delta","text":" world!"}}`,
			`data: {"id":"msg_123","type":"content_block_stop"}`,
			`data: {"id":"msg_123","type":"message_stop"}`,
			`data: [DONE]`,
		}
		
		flusher := w.(http.Flusher)
		for _, event := range events {
			fmt.Fprintf(w, "%s\n\n", event)
			flusher.Flush()
			time.Sleep(10 * time.Millisecond) // Simulate streaming delay
		}
	}))
	defer providerServer.Close()

	// Create test configuration
	cfg := &config.Config{
		Providers: []config.Provider{
			{
				Name:       "test-provider",
				APIBaseURL: providerServer.URL,
				APIKey:     "test-key",
				Enabled:    true,
				Models:     []string{"test-model"},
			},
		},
		Routes: map[string]config.Route{
			"default": {
				Provider: "test-provider",
				Model:    "test-model",
			},
		},
	}

	// Create services
	configService := config.NewService()
	configService.SetConfig(cfg)
	
	providerService := providers.NewService(configService)
	providerService.Initialize()
	
	transformerService := transformer.GetRegistry()
	routerService := router.New(cfg)

	// Create pipeline
	pipeline := NewPipeline(cfg, providerService, transformerService, routerService)

	// Create request
	reqCtx := &RequestContext{
		Body: map[string]interface{}{
			"model": "test-model",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Hello",
				},
			},
			"stream": true,
		},
		Headers:     map[string]string{},
		IsStreaming: true,
	}

	// Process request
	ctx := context.Background()
	respCtx, err := pipeline.ProcessRequest(ctx, reqCtx)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	// Verify response context
	if respCtx.Provider != "test-provider" {
		t.Errorf("Expected provider test-provider, got %s", respCtx.Provider)
	}
	if respCtx.Model != "test-model" {
		t.Errorf("Expected model test-model, got %s", respCtx.Model)
	}

	// Create response writer
	w := httptest.NewRecorder()

	// Stream the response
	err = pipeline.StreamResponse(ctx, w, respCtx)
	if err != nil {
		t.Fatalf("StreamResponse failed: %v", err)
	}

	// Verify streamed response
	body := w.Body.String()
	expectedContent := []string{
		"Hello",
		"streaming",
		"world!",
		"[DONE]",
	}
	
	for _, expected := range expectedContent {
		if !strings.Contains(body, expected) {
			t.Errorf("Expected response to contain %q, got:\n%s", expected, body)
		}
	}

	// Verify SSE format
	if !strings.Contains(body, "data:") {
		t.Error("Expected SSE format with 'data:' prefix")
	}
}

func TestPipeline_StreamingWithTransformation(t *testing.T) {
	// This test would verify that transformations are applied during streaming
	// For now, we'll skip it as we need more complex mocking
	t.Skip("Skipping transformation test - requires more complex mocking")
}