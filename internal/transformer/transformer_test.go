package transformer

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
)

func TestBaseTransformer(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	// Test NewBaseTransformer
	t.Run("NewBaseTransformer", func(t *testing.T) {
		transformer := NewBaseTransformer("test", "/test/endpoint")

		testutil.AssertEqual(t, "test", transformer.GetName())
		testutil.AssertEqual(t, "/test/endpoint", transformer.GetEndpoint())
	})

	// Test default implementations
	t.Run("DefaultImplementations", func(t *testing.T) {
		transformer := NewBaseTransformer("test", "/test")
		ctx := context.Background()

		// Test TransformRequestIn (should pass through)
		request := map[string]interface{}{"test": "value"}
		result, err := transformer.TransformRequestIn(ctx, request, "provider")
		testutil.AssertNoError(t, err)
		// Check the result is a map with the same content
		resultMap, ok := result.(map[string]interface{})
		testutil.AssertEqual(t, true, ok)
		testutil.AssertEqual(t, "value", resultMap["test"])

		// Test TransformRequestOut (should pass through)
		result, err = transformer.TransformRequestOut(ctx, request)
		testutil.AssertNoError(t, err)
		resultMap, ok = result.(map[string]interface{})
		testutil.AssertEqual(t, true, ok)
		testutil.AssertEqual(t, "value", resultMap["test"])

		// Test TransformResponseIn (should pass through)
		resp := &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"test": "response"}`)),
		}
		resultResp, err := transformer.TransformResponseIn(ctx, resp)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 200, resultResp.StatusCode)

		// Test TransformResponseOut (should pass through)
		resp.Body = io.NopCloser(strings.NewReader(`{"test": "response"}`))
		resultResp, err = transformer.TransformResponseOut(ctx, resp)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 200, resultResp.StatusCode)
	})
}

func TestSSEEvent(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	t.Run("SSEEventCreation", func(t *testing.T) {
		event := &SSEEvent{
			Event: "message",
			Data:  "test data",
			ID:    "123",
			Retry: 5000,
		}

		testutil.AssertEqual(t, "message", event.Event)
		testutil.AssertEqual(t, "test data", event.Data)
		testutil.AssertEqual(t, "123", event.ID)
		testutil.AssertEqual(t, 5000, event.Retry)
	})
}

func TestRequestConfig(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	t.Run("RequestConfigCreation", func(t *testing.T) {
		reqConfig := &RequestConfig{
			Body:    map[string]interface{}{"message": "test"},
			URL:     "https://api.example.com/v1/chat",
			Headers: map[string]string{"Authorization": "Bearer token"},
			Method:  "POST",
			Timeout: 30,
		}

		testutil.AssertEqual(t, "https://api.example.com/v1/chat", reqConfig.URL)
		testutil.AssertEqual(t, "POST", reqConfig.Method)
		testutil.AssertEqual(t, 30, reqConfig.Timeout)
		testutil.AssertEqual(t, "Bearer token", reqConfig.Headers["Authorization"])
	})
}

func TestTransformerChain(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	// Create mock transformers
	transformer1 := &mockTransformer{name: "transformer1"}
	transformer2 := &mockTransformer{name: "transformer2"}

	t.Run("NewTransformerChain", func(t *testing.T) {
		chain := NewTransformerChain(transformer1, transformer2)
		testutil.AssertEqual(t, 2, len(chain.transformers))
	})

	t.Run("Add", func(t *testing.T) {
		chain := NewTransformerChain()
		chain.Add(transformer1)
		chain.Add(transformer2)
		testutil.AssertEqual(t, 2, len(chain.transformers))
	})

	t.Run("TransformRequestIn", func(t *testing.T) {
		chain := NewTransformerChain(transformer1, transformer2)
		ctx := context.Background()

		request := map[string]interface{}{"test": "value"}
		result, err := chain.TransformRequestIn(ctx, request, "provider")
		testutil.AssertNoError(t, err)

		// Should contain transformations from both transformers
		resultMap, ok := result.(map[string]interface{})
		testutil.AssertEqual(t, true, ok)
		testutil.AssertEqual(t, "value", resultMap["test"])
		testutil.AssertEqual(t, "transformer1", resultMap["transformed_by_transformer1"])
		testutil.AssertEqual(t, "transformer2", resultMap["transformed_by_transformer2"])
	})

	t.Run("TransformRequestInWithError", func(t *testing.T) {
		errorTransformer := &mockTransformer{name: "error", shouldError: true}
		chain := NewTransformerChain(transformer1, errorTransformer)
		ctx := context.Background()

		request := map[string]interface{}{"test": "value"}
		_, err := chain.TransformRequestIn(ctx, request, "provider")
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "mock error")
	})

	t.Run("TransformResponseOut", func(t *testing.T) {
		chain := NewTransformerChain(transformer1, transformer2)
		ctx := context.Background()

		// Create test response
		body := `{"message": "test"}`
		resp := &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
		}

		result, err := chain.TransformResponseOut(ctx, resp)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 200, result.StatusCode)

		// Verify body was processed (transformed in reverse order)
		resultBody, _ := io.ReadAll(result.Body)
		var resultData map[string]interface{}
		json.Unmarshal(resultBody, &resultData)
		testutil.AssertEqual(t, "test", resultData["message"])
		testutil.AssertEqual(t, "transformer2", resultData["response_transformed_by_transformer2"])
		testutil.AssertEqual(t, "transformer1", resultData["response_transformed_by_transformer1"])
	})

	t.Run("TransformResponseOutWithError", func(t *testing.T) {
		errorTransformer := &mockTransformer{name: "error", shouldError: true}
		chain := NewTransformerChain(transformer1, errorTransformer)
		ctx := context.Background()

		body := `{"message": "test"}`
		resp := &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
		}

		_, err := chain.TransformResponseOut(ctx, resp)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "mock error")
	})

	t.Run("TransformSSEEvent", func(t *testing.T) {
		chain := NewTransformerChain(transformer1, transformer2)
		ctx := context.Background()

		event := &SSEEvent{
			Data: `{"type": "message", "content": "test"}`,
		}

		result, err := chain.TransformSSEEvent(ctx, event, "provider")
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, event.Data, result.Data) // Should pass through unchanged
	})
}

func TestTransformerChainEdgeCases(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	t.Run("EmptyChain", func(t *testing.T) {
		chain := NewTransformerChain()
		ctx := context.Background()

		request := map[string]interface{}{"test": "value"}
		result, err := chain.TransformRequestIn(ctx, request, "provider")
		testutil.AssertNoError(t, err)
		resultMap, ok := result.(map[string]interface{})
		testutil.AssertEqual(t, true, ok)
		testutil.AssertEqual(t, "value", resultMap["test"])

		resp := &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"test": "response"}`)),
		}
		resultResp, err := chain.TransformResponseOut(ctx, resp)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 200, resultResp.StatusCode)
	})

	t.Run("SingleTransformerChain", func(t *testing.T) {
		transformer := &mockTransformer{name: "single"}
		chain := NewTransformerChain(transformer)
		ctx := context.Background()

		request := map[string]interface{}{"test": "value"}
		result, err := chain.TransformRequestIn(ctx, request, "provider")
		testutil.AssertNoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		testutil.AssertEqual(t, true, ok)
		testutil.AssertEqual(t, "value", resultMap["test"])
		testutil.AssertEqual(t, "single", resultMap["transformed_by_single"])
	})
}

// mockTransformer is a test implementation of the Transformer interface
type mockTransformer struct {
	name        string
	shouldError bool
}

func (m *mockTransformer) GetName() string {
	return m.name
}

func (m *mockTransformer) GetEndpoint() string {
	return "/mock/endpoint"
}

func (m *mockTransformer) TransformRequestIn(ctx context.Context, request interface{}, provider string) (interface{}, error) {
	if m.shouldError {
		return nil, &mockError{msg: "mock error from " + m.name}
	}

	// Add a transformation marker to the request
	if reqMap, ok := request.(map[string]interface{}); ok {
		reqMap["transformed_by_"+m.name] = m.name
		return reqMap, nil
	}
	return request, nil
}

func (m *mockTransformer) TransformRequestOut(ctx context.Context, request interface{}) (interface{}, error) {
	if m.shouldError {
		return nil, &mockError{msg: "mock error from " + m.name}
	}
	return request, nil
}

func (m *mockTransformer) TransformResponseIn(ctx context.Context, response *http.Response) (*http.Response, error) {
	if m.shouldError {
		return nil, &mockError{msg: "mock error from " + m.name}
	}
	return response, nil
}

func (m *mockTransformer) TransformResponseOut(ctx context.Context, response *http.Response) (*http.Response, error) {
	if m.shouldError {
		return nil, &mockError{msg: "mock error from " + m.name}
	}

	// Read response body and add transformation marker
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return response, nil
	}
	response.Body.Close()

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		// Not JSON, return as-is
		response.Body = io.NopCloser(bytes.NewReader(body))
		return response, nil
	}

	// Add transformation marker
	data["response_transformed_by_"+m.name] = m.name

	// Re-encode
	newBody, _ := json.Marshal(data)
	response.Body = io.NopCloser(bytes.NewReader(newBody))
	response.ContentLength = int64(len(newBody))

	return response, nil
}

// mockError implements error interface for testing
type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

func TestTransformerInterfaces(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	t.Run("TransformerInterface", func(t *testing.T) {
		var transformer Transformer = &mockTransformer{name: "test"}
		testutil.AssertEqual(t, "test", transformer.GetName())
		testutil.AssertEqual(t, "/mock/endpoint", transformer.GetEndpoint())
	})

	t.Run("StreamTransformerInterface", func(t *testing.T) {
		// Test that our base transformer can be extended to implement StreamTransformer
		streamTransformer := &mockStreamTransformer{
			mockTransformer: mockTransformer{name: "stream"},
		}

		var transformer Transformer = streamTransformer
		testutil.AssertEqual(t, "stream", transformer.GetName())

		var streamTrans StreamTransformer = streamTransformer
		testutil.AssertEqual(t, "stream", streamTrans.GetName())
	})
}

// mockStreamTransformer implements StreamTransformer for testing
type mockStreamTransformer struct {
	mockTransformer
}

func (m *mockStreamTransformer) TransformStream(ctx context.Context, reader StreamReader, writer StreamWriter) error {
	// Mock implementation - just pass through
	for {
		event, err := reader.ReadEvent()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if err := writer.WriteEvent(event); err != nil {
			return err
		}
	}
}

// mockStreamReader implements StreamReader for testing
type mockStreamReader struct {
	events      []SSEEvent
	index       int
	closed      bool
	shouldError bool
}

func (r *mockStreamReader) ReadEvent() (*SSEEvent, error) {
	if r.shouldError {
		return nil, &mockError{msg: "mock error"}
	}

	if r.closed {
		return nil, io.EOF
	}

	if r.index >= len(r.events) {
		return nil, io.EOF
	}

	event := r.events[r.index]
	r.index++
	return &event, nil
}

func (r *mockStreamReader) Close() error {
	r.closed = true
	return nil
}

// mockStreamWriter implements StreamWriter for testing
type mockStreamWriter struct {
	events      []SSEEvent
	closed      bool
	shouldError bool
}

func (w *mockStreamWriter) WriteEvent(event *SSEEvent) error {
	if w.shouldError {
		return &mockError{msg: "mock error"}
	}

	if w.closed {
		return &mockError{msg: "writer is closed"}
	}

	w.events = append(w.events, *event)
	return nil
}

func (w *mockStreamWriter) Flush() error {
	return nil
}

func (w *mockStreamWriter) Close() error {
	w.closed = true
	return nil
}

func TestStreamInterfaces(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	t.Run("StreamReader", func(t *testing.T) {
		events := []SSEEvent{
			{Data: "event1"},
			{Data: "event2"},
		}

		reader := &mockStreamReader{events: events}

		// Read first event
		event, err := reader.ReadEvent()
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, "event1", event.Data)

		// Read second event
		event, err = reader.ReadEvent()
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, "event2", event.Data)

		// EOF after all events
		_, err = reader.ReadEvent()
		testutil.AssertError(t, err)
		testutil.AssertEqual(t, io.EOF, err)

		// Close reader
		err = reader.Close()
		testutil.AssertNoError(t, err)

		// Reading after close should return EOF
		_, err = reader.ReadEvent()
		testutil.AssertError(t, err)
		testutil.AssertEqual(t, io.EOF, err)
	})

	t.Run("StreamWriter", func(t *testing.T) {
		writer := &mockStreamWriter{}

		// Write events
		event1 := &SSEEvent{Data: "test1"}
		err := writer.WriteEvent(event1)
		testutil.AssertNoError(t, err)

		event2 := &SSEEvent{Data: "test2"}
		err = writer.WriteEvent(event2)
		testutil.AssertNoError(t, err)

		// Verify events were written
		testutil.AssertEqual(t, 2, len(writer.events))
		testutil.AssertEqual(t, "test1", writer.events[0].Data)
		testutil.AssertEqual(t, "test2", writer.events[1].Data)

		// Flush should work
		err = writer.Flush()
		testutil.AssertNoError(t, err)

		// Close writer
		err = writer.Close()
		testutil.AssertNoError(t, err)

		// Writing after close should error
		err = writer.WriteEvent(&SSEEvent{Data: "test3"})
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "writer is closed")
	})

	t.Run("StreamTransformer", func(t *testing.T) {
		reader := &mockStreamReader{
			events: []SSEEvent{
				{Data: "event1"},
				{Data: "event2"},
			},
		}
		writer := &mockStreamWriter{}
		transformer := &mockStreamTransformer{}

		// Transform stream
		err := transformer.TransformStream(context.Background(), reader, writer)
		testutil.AssertNoError(t, err)

		// Verify all events were passed through
		testutil.AssertEqual(t, 2, len(writer.events))
		testutil.AssertEqual(t, "event1", writer.events[0].Data)
		testutil.AssertEqual(t, "event2", writer.events[1].Data)
	})
}
