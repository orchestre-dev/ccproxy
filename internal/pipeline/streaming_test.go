package pipeline

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/musistudio/ccproxy/internal/transformer"
)

func TestStreamingProcessor_ProcessStreamingResponse(t *testing.T) {
	// Create test SSE data
	sseData := `data: {"id":"msg_123","type":"content_block_delta","delta":{"type":"text_delta","text":"Hello"}}

data: {"id":"msg_123","type":"content_block_delta","delta":{"type":"text_delta","text":" world"}}

data: {"id":"msg_123","type":"message_stop"}

data: [DONE]

`

	tests := []struct {
		name         string
		sseData      string
		provider     string
		wantEvents   int
		wantContains []string
	}{
		{
			name:     "basic SSE streaming",
			sseData:  sseData,
			provider: "test-provider",
			wantEvents: 4,
			wantContains: []string{
				"Hello",
				"world",
				"message_stop",
				"[DONE]",
			},
		},
		{
			name: "error event handling",
			sseData: `event: error
data: {"error": {"type": "rate_limit", "message": "Rate limit exceeded"}}

data: [DONE]

`,
			provider:   "test-provider",
			wantEvents: 2,
			wantContains: []string{
				"error",
				"rate_limit",
				"[DONE]",
			},
		},
		{
			name: "empty events filtered",
			sseData: `data: {"test": "data1"}


data: {"test": "data2"}

data: [DONE]

`,
			provider:   "test-provider",
			wantEvents: 3,
			wantContains: []string{
				"data1",
				"data2",
				"[DONE]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create response with SSE data
			resp := &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(tt.sseData)),
				Header:     make(http.Header),
			}
			resp.Header.Set("Content-Type", "text/event-stream")

			// Create response writer
			w := httptest.NewRecorder()

			// Create streaming processor
			processor := NewStreamingProcessor(transformer.GetRegistry())

			// Process streaming response
			ctx := context.Background()
			err := processor.ProcessStreamingResponse(ctx, w, resp, tt.provider)
			if err != nil {
				t.Fatalf("ProcessStreamingResponse failed: %v", err)
			}

			// Verify headers
			if ct := w.Header().Get("Content-Type"); ct != "text/event-stream" {
				t.Errorf("Expected Content-Type text/event-stream, got %s", ct)
			}

			// Verify response contains expected data
			body := w.Body.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(body, want) {
					t.Errorf("Expected response to contain %q, got:\n%s", want, body)
				}
			}

			// Count events (simple check for "data:" lines)
			eventCount := strings.Count(body, "data:")
			if eventCount != tt.wantEvents {
				t.Errorf("Expected %d events, got %d", tt.wantEvents, eventCount)
			}
		})
	}
}

func TestStreamingProcessor_ContextCancellation(t *testing.T) {
	// Create a slow SSE stream
	slowStream := make(chan string)
	go func() {
		defer close(slowStream)
		for i := 0; i < 10; i++ {
			slowStream <- fmt.Sprintf("data: {\"index\": %d}\n\n", i)
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// Create response with channel reader
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		for data := range slowStream {
			pw.Write([]byte(data))
		}
	}()

	resp := &http.Response{
		StatusCode: 200,
		Body:       pr,
		Header:     make(http.Header),
	}

	// Create response writer
	w := httptest.NewRecorder()

	// Create context with cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after 250ms (should get ~2 events)
	go func() {
		time.Sleep(250 * time.Millisecond)
		cancel()
	}()

	// Process streaming response
	processor := NewStreamingProcessor(transformer.GetRegistry())
	err := processor.ProcessStreamingResponse(ctx, w, resp, "test-provider")

	// Should complete without error (context cancellation is normal)
	if err != nil {
		t.Errorf("Expected no error on context cancellation, got: %v", err)
	}

	// Should have received some events but not all
	body := w.Body.String()
	eventCount := strings.Count(body, "data:")
	if eventCount == 0 {
		t.Error("Expected to receive some events before cancellation")
	}
	if eventCount >= 10 {
		t.Error("Expected cancellation to stop processing before all events")
	}
}

func TestHandleStreamingError(t *testing.T) {
	// Create response writer
	w := httptest.NewRecorder()

	// Send streaming error
	err := fmt.Errorf("test streaming error")
	HandleStreamingError(w, err)

	// Verify headers
	if ct := w.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("Expected Content-Type text/event-stream, got %s", ct)
	}

	// Verify error event
	body := w.Body.String()
	if !strings.Contains(body, "event: error") {
		t.Error("Expected error event")
	}
	if !strings.Contains(body, "test streaming error") {
		t.Error("Expected error message in response")
	}
	if !strings.Contains(body, "[DONE]") {
		t.Error("Expected [DONE] marker")
	}
}