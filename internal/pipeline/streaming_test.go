package pipeline

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/orchestre-dev/ccproxy/internal/transformer"
)

func TestNewStreamingProcessor(t *testing.T) {
	transformerService := transformer.NewService()
	processor := NewStreamingProcessor(transformerService)

	if processor == nil {
		t.Error("Expected non-nil streaming processor")
	}

	if processor.transformerService != transformerService {
		t.Error("Transformer service not set correctly")
	}
}

func TestStreamingProcessor_ProcessStreamingResponse(t *testing.T) {
	transformerService := transformer.NewService()
	processor := NewStreamingProcessor(transformerService)

	t.Run("ValidSSEStream", func(t *testing.T) {
		// Create a mock SSE response
		sseData := "data: {\"type\": \"message_start\"}\n\ndata: {\"type\": \"content_block_start\"}\n\ndata: [DONE]\n\n"
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(sseData)),
		}

		w := httptest.NewRecorder()
		ctx := context.Background()

		err := processor.ProcessStreamingResponse(ctx, w, resp, "anthropic")
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

		if w.Header().Get("Connection") != "keep-alive" {
			t.Error("Expected Connection header")
		}

		if w.Header().Get("X-Accel-Buffering") != "no" {
			t.Error("Expected X-Accel-Buffering header")
		}
	})

	t.Run("EmptyStream", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader("")),
		}

		w := httptest.NewRecorder()
		ctx := context.Background()

		err := processor.ProcessStreamingResponse(ctx, w, resp, "openai")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	})

	t.Run("StreamWithDoneMarker", func(t *testing.T) {
		sseData := "data: {\"chunk\": \"hello\"}\n\ndata: [DONE]\n\n"
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(sseData)),
		}

		w := httptest.NewRecorder()
		ctx := context.Background()

		err := processor.ProcessStreamingResponse(ctx, w, resp, "openai")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Should process until [DONE] marker
		body := w.Body.String()
		if !strings.Contains(body, "[DONE]") {
			t.Error("Expected [DONE] marker in output")
		}
	})

	t.Run("CanceledContext", func(t *testing.T) {
		sseData := "data: {\"chunk\": \"hello\"}\n\ndata: {\"chunk\": \"world\"}\n\n"
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(sseData)),
		}

		recorder := httptest.NewRecorder()
		w := &safeTestWriter{ResponseWriter: recorder}
		ctx, cancel := context.WithCancel(context.Background())

		// Cancel context immediately
		cancel()

		err := processor.ProcessStreamingResponse(ctx, w, resp, "openai")
		// Should handle cancellation gracefully
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	})

	t.Run("NonFlushableWriter", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader("data: test\n\n")),
		}

		// Create a non-flushable writer
		w := &nonFlushableWriter{}
		ctx := context.Background()

		err := processor.ProcessStreamingResponse(ctx, w, resp, "openai")
		if err == nil {
			t.Error("Expected error for non-flushable writer")
		}

		if !strings.Contains(err.Error(), "does not support flushing") {
			t.Errorf("Expected flushing error, got %v", err)
		}
	})
}

// nonFlushableWriter implements http.ResponseWriter but not http.Flusher
type nonFlushableWriter struct {
	header http.Header
	body   strings.Builder
	status int
}

func (w *nonFlushableWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *nonFlushableWriter) Write(data []byte) (int, error) {
	return w.body.Write(data)
}

func (w *nonFlushableWriter) WriteHeader(statusCode int) {
	w.status = statusCode
}

func TestStreamingProcessor_PassThrough(t *testing.T) {
	transformerService := transformer.NewService()
	processor := NewStreamingProcessor(transformerService)

	t.Run("ValidPassThrough", func(t *testing.T) {
		sseData := "data: hello\n\ndata: world\n\ndata: [DONE]\n\n"
		reader := transformer.NewSSEReader(io.NopCloser(strings.NewReader(sseData)))

		w := httptest.NewRecorder()
		writer := transformer.NewSSEWriter(w)
		flusher := w // httptest.ResponseRecorder implements http.Flusher

		err := processor.passThrough(reader, writer, flusher)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		body := w.Body.String()
		if !strings.Contains(body, "hello") {
			t.Error("Expected 'hello' in output")
		}

		if !strings.Contains(body, "world") {
			t.Error("Expected 'world' in output")
		}

		if !strings.Contains(body, "[DONE]") {
			t.Error("Expected '[DONE]' in output")
		}
	})

	t.Run("EmptyPassThrough", func(t *testing.T) {
		reader := transformer.NewSSEReader(io.NopCloser(strings.NewReader("")))

		w := httptest.NewRecorder()
		writer := transformer.NewSSEWriter(w)
		flusher := w

		err := processor.passThrough(reader, writer, flusher)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	})

	t.Run("ReaderError", func(t *testing.T) {
		// Create a reader that will return an error
		errorReader := &errorReader{}
		reader := transformer.NewSSEReader(errorReader)

		w := httptest.NewRecorder()
		writer := transformer.NewSSEWriter(w)
		flusher := w

		err := processor.passThrough(reader, writer, flusher)
		if err == nil {
			t.Error("Expected error from reader")
		}
	})
}

// errorReader always returns an error when reading
type errorReader struct{}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func (r *errorReader) Close() error {
	return nil
}

func TestStreamingProcessor_ErrorHandling(t *testing.T) {
	transformerService := transformer.NewService()
	processor := NewStreamingProcessor(transformerService)

	t.Run("MalformedSSE", func(t *testing.T) {
		// Create SSE data with some malformed events
		sseData := "data: {\"valid\": \"event\"}\n\nmalformed line without proper format\ndata: {\"another\": \"event\"}\n\ndata: [DONE]\n\n"
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(sseData)),
		}

		w := httptest.NewRecorder()
		ctx := context.Background()

		// Should handle malformed events gracefully
		err := processor.ProcessStreamingResponse(ctx, w, resp, "anthropic")
		if err != nil {
			t.Fatalf("Should handle malformed events gracefully: %v", err)
		}
	})

	t.Run("EmptyEvents", func(t *testing.T) {
		// Test with empty events that should be skipped
		sseData := "data: \n\ndata: {\"valid\": \"event\"}\n\nevent: ping\ndata: \n\ndata: [DONE]\n\n"
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(sseData)),
		}

		w := httptest.NewRecorder()
		ctx := context.Background()

		err := processor.ProcessStreamingResponse(ctx, w, resp, "anthropic")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Should only contain the valid event and [DONE]
		body := w.Body.String()
		if !strings.Contains(body, "valid") {
			t.Error("Expected valid event in output")
		}

		if !strings.Contains(body, "[DONE]") {
			t.Error("Expected [DONE] marker in output")
		}
	})
}

func TestStreamingProcessor_Integration(t *testing.T) {
	// Test with a more realistic streaming scenario
	transformerService := transformer.NewService()
	processor := NewStreamingProcessor(transformerService)

	t.Run("AnthropicStyleStream", func(t *testing.T) {
		// Simulate Anthropic-style streaming response
		sseData := `event: message_start
data: {"type": "message_start", "message": {"id": "msg_123", "model": "claude-3-haiku"}}

event: content_block_start
data: {"type": "content_block_start", "index": 0, "content_block": {"type": "text", "text": ""}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "Hello"}}

event: content_block_delta
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": " world!"}}

event: content_block_stop
data: {"type": "content_block_stop", "index": 0}

event: message_stop
data: {"type": "message_stop"}

data: [DONE]

`

		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(sseData)),
		}

		w := httptest.NewRecorder()
		ctx := context.Background()

		err := processor.ProcessStreamingResponse(ctx, w, resp, "anthropic")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		body := w.Body.String()

		// Check that all event types are preserved
		if !strings.Contains(body, "message_start") {
			t.Error("Expected message_start event")
		}

		if !strings.Contains(body, "content_block_delta") {
			t.Error("Expected content_block_delta events")
		}

		if !strings.Contains(body, "Hello") {
			t.Error("Expected 'Hello' in stream")
		}

		if !strings.Contains(body, "world!") {
			t.Error("Expected 'world!' in stream")
		}

		if !strings.Contains(body, "[DONE]") {
			t.Error("Expected [DONE] marker")
		}
	})
}

// Test enhanced streaming error scenarios
func TestStreamingProcessor_AdvancedErrorHandling(t *testing.T) {
	transformerService := transformer.NewService()
	processor := NewStreamingProcessor(transformerService)

	t.Run("TooManyReadErrors", func(t *testing.T) {
		// Create a reader that always returns errors to trigger the error limit
		errorReader := &intermittentErrorReader{errorCount: 0, maxErrors: 15}
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       errorReader,
		}

		w := httptest.NewRecorder()
		ctx := context.Background()

		err := processor.ProcessStreamingResponse(ctx, w, resp, "openai")
		if err == nil {
			t.Error("Expected error for too many read errors")
		}

		if !strings.Contains(err.Error(), "too many errors") {
			t.Errorf("Expected 'too many errors' error, got %v", err)
		}
	})

	t.Run("ClientDisconnectionDuringStream", func(t *testing.T) {
		sseData := "data: {\"chunk\": \"hello\"}\n\ndata: {\"chunk\": \"world\"}\n\ndata: [DONE]\n\n"
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(sseData)),
		}

		// Create a writer that simulates broken pipe error
		w := &brokenPipeWriter{}
		ctx := context.Background()

		// Should handle client disconnection gracefully
		err := processor.ProcessStreamingResponse(ctx, w, resp, "openai")
		if err != nil {
			t.Logf("Got expected disconnection error: %v", err)
		}
	})

	t.Run("TransformationErrors", func(t *testing.T) {
		// Test streaming with transformation errors
		sseData := "data: {\"malformed\": \"json\",}\n\ndata: {\"valid\": \"json\"}\n\ndata: [DONE]\n\n"
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(sseData)),
		}

		w := httptest.NewRecorder()
		ctx := context.Background()

		// Should handle transformation errors gracefully
		err := processor.ProcessStreamingResponse(ctx, w, resp, "openai")
		if err != nil {
			t.Fatalf("Should handle transformation errors gracefully: %v", err)
		}

		// Should still contain the valid events and [DONE]
		body := w.Body.String()
		if !strings.Contains(body, "[DONE]") {
			t.Error("Expected [DONE] marker despite transformation errors")
		}
	})

	t.Run("NoTransformerChain", func(t *testing.T) {
		// Test with a provider that has no transformer chain (should use pass-through)
		sseData := "data: {\"raw\": \"event\"}\n\ndata: [DONE]\n\n"
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(sseData)),
		}

		w := httptest.NewRecorder()
		ctx := context.Background()

		err := processor.ProcessStreamingResponse(ctx, w, resp, "unknown-provider")
		if err != nil {
			t.Fatalf("Unexpected error for unknown provider: %v", err)
		}

		// Should pass through events unchanged
		body := w.Body.String()
		if !strings.Contains(body, "raw") {
			t.Error("Expected raw event to be passed through")
		}
	})

	t.Run("ContextCancellationDuringProcessing", func(t *testing.T) {
		// Create a slow reader to allow time for context cancellation
		slowReader := &slowReader{data: "data: slow event\n\ndata: [DONE]\n\n", delay: 50 * time.Millisecond}
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       slowReader,
		}

		// Use a custom writer that's safe for concurrent access
		w := &safeTestWriter{ResponseWriter: httptest.NewRecorder()}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		// Should handle context cancellation gracefully
		err := processor.ProcessStreamingResponse(ctx, w, resp, "openai")
		if err != nil {
			t.Logf("Context cancellation handled: %v", err)
		}

		// Wait a bit to ensure goroutines complete
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("WriterCloseErrors", func(t *testing.T) {
		sseData := "data: test\n\ndata: [DONE]\n\n"
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(sseData)),
		}

		// Writer that returns "writer is closed" error
		w := &closedWriter{}
		ctx := context.Background()

		// Should handle writer closure gracefully
		err := processor.ProcessStreamingResponse(ctx, w, resp, "openai")
		if err != nil {
			t.Logf("Writer closure handled: %v", err)
		}
	})
}

// Test helpers for streaming error scenarios

// intermittentErrorReader returns errors up to a limit, then returns data
type intermittentErrorReader struct {
	errorCount int
	maxErrors  int
}

func (r *intermittentErrorReader) Read(p []byte) (n int, err error) {
	if r.errorCount < r.maxErrors {
		r.errorCount++
		return 0, io.ErrUnexpectedEOF
	}
	return 0, io.EOF
}

func (r *intermittentErrorReader) Close() error {
	return nil
}

// brokenPipeWriter simulates a client disconnection
type brokenPipeWriter struct {
	header http.Header
	calls  int
}

func (w *brokenPipeWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *brokenPipeWriter) Write(data []byte) (int, error) {
	w.calls++
	if w.calls > 2 {
		// Simulate broken pipe after a few writes
		return 0, fmt.Errorf("write: broken pipe")
	}
	return len(data), nil
}

func (w *brokenPipeWriter) WriteHeader(statusCode int) {}

func (w *brokenPipeWriter) Flush() {}

// closedWriter simulates a closed writer
type closedWriter struct {
	header http.Header
}

func (w *closedWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *closedWriter) Write(data []byte) (int, error) {
	return 0, fmt.Errorf("writer is closed")
}

func (w *closedWriter) WriteHeader(statusCode int) {}

func (w *closedWriter) Flush() {}

// slowReader simulates slow network reads
type slowReader struct {
	data  string
	pos   int
	delay time.Duration
}

func (r *slowReader) Read(p []byte) (n int, err error) {
	if r.delay > 0 {
		time.Sleep(r.delay)
	}

	if r.pos >= len(r.data) {
		return 0, io.EOF
	}

	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func (r *slowReader) Close() error {
	return nil
}

// Test pass-through streaming errors
func TestStreamingProcessor_PassThroughErrors(t *testing.T) {
	transformerService := transformer.NewService()
	processor := NewStreamingProcessor(transformerService)

	t.Run("PassThroughWriterClose", func(t *testing.T) {
		sseData := "data: test1\n\ndata: test2\n\ndata: [DONE]\n\n"
		reader := transformer.NewSSEReader(io.NopCloser(strings.NewReader(sseData)))

		w := &closedWriter{}
		writer := transformer.NewSSEWriter(w)

		// Should handle writer close error gracefully
		err := processor.passThrough(reader, writer, w)
		if err != nil {
			t.Logf("Pass-through writer close handled: %v", err)
		}
	})
}

// Test streaming processor with various event types
func TestStreamingProcessor_EventTypes(t *testing.T) {
	transformerService := transformer.NewService()
	processor := NewStreamingProcessor(transformerService)

	t.Run("MixedEventTypes", func(t *testing.T) {
		sseData := `event: custom
data: {"type": "custom"}

: comment line (should be ignored)

data: {"type": "data-only"}

event: another
data: 

data: [DONE]

`

		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(sseData)),
		}

		w := httptest.NewRecorder()
		ctx := context.Background()

		err := processor.ProcessStreamingResponse(ctx, w, resp, "openai")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		body := w.Body.String()
		if !strings.Contains(body, "custom") {
			t.Error("Expected custom event to be processed")
		}

		if !strings.Contains(body, "data-only") {
			t.Error("Expected data-only event to be processed")
		}

		if !strings.Contains(body, "[DONE]") {
			t.Error("Expected [DONE] marker")
		}
	})
}

// safeTestWriter wraps httptest.ResponseRecorder to be safe for concurrent access
type safeTestWriter struct {
	http.ResponseWriter
	mu sync.Mutex
}

func (w *safeTestWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.ResponseWriter.Write(data)
}

func (w *safeTestWriter) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
