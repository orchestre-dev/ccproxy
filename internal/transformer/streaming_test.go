package transformer

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestNewSSEReader(t *testing.T) {
	data := "data: test\n\n"
	reader := NewSSEReader(io.NopCloser(strings.NewReader(data)))
	
	if reader == nil {
		t.Error("Expected non-nil SSEReader")
	}
	
	if reader.closed {
		t.Error("Expected reader to not be closed initially")
	}
}

func TestSSEReader_ReadEvent(t *testing.T) {
	t.Run("SimpleEvent", func(t *testing.T) {
		data := "data: hello world\n\n"
		reader := NewSSEReader(io.NopCloser(strings.NewReader(data)))
		
		event, err := reader.ReadEvent()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if event.Data != "hello world" {
			t.Errorf("Expected data 'hello world', got %q", event.Data)
		}
	})

	t.Run("EventWithType", func(t *testing.T) {
		data := "event: message\ndata: test data\n\n"
		reader := NewSSEReader(io.NopCloser(strings.NewReader(data)))
		
		event, err := reader.ReadEvent()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if event.Event != "message" {
			t.Errorf("Expected event 'message', got %q", event.Event)
		}
		
		if event.Data != "test data" {
			t.Errorf("Expected data 'test data', got %q", event.Data)
		}
	})

	t.Run("EventWithID", func(t *testing.T) {
		data := "id: 123\ndata: test\n\n"
		reader := NewSSEReader(io.NopCloser(strings.NewReader(data)))
		
		event, err := reader.ReadEvent()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if event.ID != "123" {
			t.Errorf("Expected ID '123', got %q", event.ID)
		}
	})

	t.Run("MultilineData", func(t *testing.T) {
		data := "data: line 1\ndata: line 2\ndata: line 3\n\n"
		reader := NewSSEReader(io.NopCloser(strings.NewReader(data)))
		
		event, err := reader.ReadEvent()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		expected := "line 1\nline 2\nline 3"
		if event.Data != expected {
			t.Errorf("Expected data %q, got %q", expected, event.Data)
		}
	})

	t.Run("MultipleEvents", func(t *testing.T) {
		data := "data: event1\n\ndata: event2\n\n"
		reader := NewSSEReader(io.NopCloser(strings.NewReader(data)))
		
		// Read first event
		event1, err := reader.ReadEvent()
		if err != nil {
			t.Fatalf("Unexpected error reading first event: %v", err)
		}
		if event1.Data != "event1" {
			t.Errorf("Expected first event data 'event1', got %q", event1.Data)
		}
		
		// Read second event
		event2, err := reader.ReadEvent()
		if err != nil {
			t.Fatalf("Unexpected error reading second event: %v", err)
		}
		if event2.Data != "event2" {
			t.Errorf("Expected second event data 'event2', got %q", event2.Data)
		}
		
		// Should get EOF on third read
		_, err = reader.ReadEvent()
		if err != io.EOF {
			t.Errorf("Expected EOF, got %v", err)
		}
	})

	t.Run("EmptyData", func(t *testing.T) {
		data := ""
		reader := NewSSEReader(io.NopCloser(strings.NewReader(data)))
		
		_, err := reader.ReadEvent()
		if err != io.EOF {
			t.Errorf("Expected EOF for empty data, got %v", err)
		}
	})

	t.Run("ReadAfterClose", func(t *testing.T) {
		data := "data: test\n\n"
		reader := NewSSEReader(io.NopCloser(strings.NewReader(data)))
		
		err := reader.Close()
		if err != nil {
			t.Fatalf("Unexpected error closing reader: %v", err)
		}
		
		_, err = reader.ReadEvent()
		if err != io.EOF {
			t.Errorf("Expected EOF after close, got %v", err)
		}
	})
}

func TestSSEReader_Close(t *testing.T) {
	data := "data: test\n\n"
	reader := NewSSEReader(io.NopCloser(strings.NewReader(data)))
	
	err := reader.Close()
	if err != nil {
		t.Errorf("Unexpected error closing reader: %v", err)
	}
	
	if !reader.closed {
		t.Error("Expected reader to be marked as closed")
	}
	
	// Closing again should not error
	err = reader.Close()
	if err != nil {
		t.Errorf("Unexpected error closing reader again: %v", err)
	}
}

func TestNewSSEWriter(t *testing.T) {
	var buf bytes.Buffer
	writer := NewSSEWriter(&buf)
	
	if writer == nil {
		t.Error("Expected non-nil SSEWriter")
	}
	
	if writer.closed {
		t.Error("Expected writer to not be closed initially")
	}
}

func TestSSEWriter_WriteEvent(t *testing.T) {
	t.Run("SimpleEvent", func(t *testing.T) {
		var buf bytes.Buffer
		writer := NewSSEWriter(&buf)
		
		event := &SSEEvent{
			Data: "hello world",
		}
		
		err := writer.WriteEvent(event)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		expected := "data: hello world\n\n"
		if buf.String() != expected {
			t.Errorf("Expected %q, got %q", expected, buf.String())
		}
	})

	t.Run("EventWithType", func(t *testing.T) {
		var buf bytes.Buffer
		writer := NewSSEWriter(&buf)
		
		event := &SSEEvent{
			Event: "message",
			Data:  "test data",
		}
		
		err := writer.WriteEvent(event)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		expected := "event: message\ndata: test data\n\n"
		if buf.String() != expected {
			t.Errorf("Expected %q, got %q", expected, buf.String())
		}
	})

	t.Run("EventWithID", func(t *testing.T) {
		var buf bytes.Buffer
		writer := NewSSEWriter(&buf)
		
		event := &SSEEvent{
			ID:   "123",
			Data: "test",
		}
		
		err := writer.WriteEvent(event)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		expected := "id: 123\ndata: test\n\n"
		if buf.String() != expected {
			t.Errorf("Expected %q, got %q", expected, buf.String())
		}
	})

	t.Run("EventWithRetry", func(t *testing.T) {
		var buf bytes.Buffer
		writer := NewSSEWriter(&buf)
		
		event := &SSEEvent{
			Data:  "test",
			Retry: 5000,
		}
		
		err := writer.WriteEvent(event)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		expected := "retry: 5000\ndata: test\n\n"
		if buf.String() != expected {
			t.Errorf("Expected %q, got %q", expected, buf.String())
		}
	})

	t.Run("EventWithMultilineData", func(t *testing.T) {
		var buf bytes.Buffer
		writer := NewSSEWriter(&buf)
		
		event := &SSEEvent{
			Data: "line 1\nline 2\nline 3",
		}
		
		err := writer.WriteEvent(event)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		expected := "data: line 1\ndata: line 2\ndata: line 3\n\n"
		if buf.String() != expected {
			t.Errorf("Expected %q, got %q", expected, buf.String())
		}
	})

	t.Run("WriteAfterClose", func(t *testing.T) {
		var buf bytes.Buffer
		writer := NewSSEWriter(&buf)
		
		err := writer.Close()
		if err != nil {
			t.Fatalf("Unexpected error closing writer: %v", err)
		}
		
		event := &SSEEvent{Data: "test"}
		err = writer.WriteEvent(event)
		if err == nil {
			t.Error("Expected error when writing after close")
		}
		
		if !strings.Contains(err.Error(), "writer is closed") {
			t.Errorf("Expected 'writer is closed' error, got %v", err)
		}
	})
}

func TestSSEWriter_Flush(t *testing.T) {
	var buf bytes.Buffer
	writer := NewSSEWriter(&buf)
	
	// Flush should not error even without a flusher
	err := writer.Flush()
	if err != nil {
		t.Errorf("Unexpected error flushing: %v", err)
	}
}

func TestSSEWriter_Close(t *testing.T) {
	var buf bytes.Buffer
	writer := NewSSEWriter(&buf)
	
	err := writer.Close()
	if err != nil {
		t.Errorf("Unexpected error closing writer: %v", err)
	}
	
	if !writer.closed {
		t.Error("Expected writer to be marked as closed")
	}
	
	// Closing again should not error
	err = writer.Close()
	if err != nil {
		t.Errorf("Unexpected error closing writer again: %v", err)
	}
}

func TestSSEWriter_WithFlusher(t *testing.T) {
	// Create a mock flusher
	mockFlusher := &mockResponseWriter{}
	writer := NewSSEWriter(mockFlusher)
	
	if writer.flusher == nil {
		t.Error("Expected writer to detect flusher capability")
	}
	
	event := &SSEEvent{Data: "test"}
	err := writer.WriteEvent(event)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !mockFlusher.flushed {
		t.Error("Expected flush to be called")
	}
}

// mockResponseWriter implements http.ResponseWriter and http.Flusher
type mockResponseWriter struct {
	buf     bytes.Buffer
	flushed bool
}

func (m *mockResponseWriter) Header() http.Header {
	return make(http.Header)
}

func (m *mockResponseWriter) Write(data []byte) (int, error) {
	return m.buf.Write(data)
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	// No-op for testing
}

func (m *mockResponseWriter) Flush() {
	m.flushed = true
}

func TestStreamPipe(t *testing.T) {
	t.Run("BasicPipe", func(t *testing.T) {
		// Create test events
		events := []SSEEvent{
			{Data: "event1"},
			{Data: "event2"},
			{Data: "event3"},
		}
		
		reader := &mockStreamReader{events: events}
		writer := &mockStreamWriter{}
		
		err := StreamPipe(reader, writer, nil)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if len(writer.events) != 3 {
			t.Errorf("Expected 3 events, got %d", len(writer.events))
		}
		
		for i, event := range writer.events {
			expected := events[i].Data
			if event.Data != expected {
				t.Errorf("Event %d: expected data %q, got %q", i, expected, event.Data)
			}
		}
	})

	t.Run("PipeWithTransform", func(t *testing.T) {
		events := []SSEEvent{
			{Data: "hello"},
			{Data: "world"},
		}
		
		reader := &mockStreamReader{events: events}
		writer := &mockStreamWriter{}
		
		// Transform function that adds prefix
		transform := func(event *SSEEvent) (*SSEEvent, error) {
			return &SSEEvent{
				Event: event.Event,
				Data:  "transformed: " + event.Data,
				ID:    event.ID,
				Retry: event.Retry,
			}, nil
		}
		
		err := StreamPipe(reader, writer, transform)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if len(writer.events) != 2 {
			t.Errorf("Expected 2 events, got %d", len(writer.events))
		}
		
		if writer.events[0].Data != "transformed: hello" {
			t.Errorf("Expected 'transformed: hello', got %q", writer.events[0].Data)
		}
		
		if writer.events[1].Data != "transformed: world" {
			t.Errorf("Expected 'transformed: world', got %q", writer.events[1].Data)
		}
	})

	t.Run("PipeWithFilterTransform", func(t *testing.T) {
		events := []SSEEvent{
			{Data: "keep"},
			{Data: "filter"},
			{Data: "keep"},
		}
		
		reader := &mockStreamReader{events: events}
		writer := &mockStreamWriter{}
		
		// Transform function that filters out events with "filter" data
		transform := func(event *SSEEvent) (*SSEEvent, error) {
			if event.Data == "filter" {
				return nil, nil // Filter out this event
			}
			return event, nil
		}
		
		err := StreamPipe(reader, writer, transform)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if len(writer.events) != 2 {
			t.Errorf("Expected 2 events after filtering, got %d", len(writer.events))
		}
		
		for _, event := range writer.events {
			if event.Data != "keep" {
				t.Errorf("Expected only 'keep' events, got %q", event.Data)
			}
		}
	})

	t.Run("PipeWithTransformError", func(t *testing.T) {
		events := []SSEEvent{
			{Data: "test"},
		}
		
		reader := &mockStreamReader{events: events}
		writer := &mockStreamWriter{}
		
		// Transform function that always errors
		transform := func(event *SSEEvent) (*SSEEvent, error) {
			return nil, &mockError{msg: "transform error"}
		}
		
		err := StreamPipe(reader, writer, transform)
		if err == nil {
			t.Error("Expected error from transform function")
		}
		
		if !strings.Contains(err.Error(), "transform error") {
			t.Errorf("Expected 'transform error', got %v", err)
		}
	})

	t.Run("PipeWithReaderError", func(t *testing.T) {
		reader := &mockStreamReader{shouldError: true}
		writer := &mockStreamWriter{}
		
		err := StreamPipe(reader, writer, nil)
		if err == nil {
			t.Error("Expected error from reader")
		}
		
		if !strings.Contains(err.Error(), "mock error") {
			t.Errorf("Expected 'mock error', got %v", err)
		}
	})

	t.Run("PipeWithWriterError", func(t *testing.T) {
		events := []SSEEvent{
			{Data: "test"},
		}
		
		reader := &mockStreamReader{events: events}
		writer := &mockStreamWriter{shouldError: true}
		
		err := StreamPipe(reader, writer, nil)
		if err == nil {
			t.Error("Expected error from writer")
		}
		
		if !strings.Contains(err.Error(), "mock error") {
			t.Errorf("Expected 'mock error', got %v", err)
		}
	})
}
