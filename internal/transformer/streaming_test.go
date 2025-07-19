package transformer

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestSSEReader_ReadEvent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []SSEEvent
	}{
		{
			name:  "simple event",
			input: "data: hello world\n\n",
			expected: []SSEEvent{
				{Data: "hello world"},
			},
		},
		{
			name:  "event with type",
			input: "event: message\ndata: hello\n\n",
			expected: []SSEEvent{
				{Event: "message", Data: "hello"},
			},
		},
		{
			name:  "multiline data",
			input: "data: line1\ndata: line2\ndata: line3\n\n",
			expected: []SSEEvent{
				{Data: "line1\nline2\nline3"},
			},
		},
		{
			name:  "multiple events",
			input: "event: start\ndata: starting\n\nevent: end\ndata: ending\n\n",
			expected: []SSEEvent{
				{Event: "start", Data: "starting"},
				{Event: "end", Data: "ending"},
			},
		},
		{
			name:  "event with id",
			input: "id: 123\nevent: message\ndata: test\n\n",
			expected: []SSEEvent{
				{ID: "123", Event: "message", Data: "test"},
			},
		},
		{
			name:  "empty lines between fields",
			input: "event: test\n\ndata: data\n\n",
			expected: []SSEEvent{
				{Event: "test"},
				{Data: "data"},
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewSSEReader(io.NopCloser(strings.NewReader(tt.input)))
			defer reader.Close()
			
			var events []SSEEvent
			for {
				event, err := reader.ReadEvent()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatalf("Failed to read event: %v", err)
				}
				events = append(events, *event)
			}
			
			if len(events) != len(tt.expected) {
				t.Errorf("Expected %d events, got %d", len(tt.expected), len(events))
				return
			}
			
			for i, event := range events {
				expected := tt.expected[i]
				if event.Event != expected.Event {
					t.Errorf("Event %d: expected event '%s', got '%s'", 
						i, expected.Event, event.Event)
				}
				if event.Data != expected.Data {
					t.Errorf("Event %d: expected data '%s', got '%s'", 
						i, expected.Data, event.Data)
				}
				if event.ID != expected.ID {
					t.Errorf("Event %d: expected ID '%s', got '%s'", 
						i, expected.ID, event.ID)
				}
			}
		})
	}
}

func TestSSEWriter_WriteEvent(t *testing.T) {
	tests := []struct {
		name     string
		event    SSEEvent
		expected string
	}{
		{
			name:     "simple data",
			event:    SSEEvent{Data: "hello world"},
			expected: "data: hello world\n\n",
		},
		{
			name:     "event with type",
			event:    SSEEvent{Event: "message", Data: "hello"},
			expected: "event: message\ndata: hello\n\n",
		},
		{
			name:     "multiline data",
			event:    SSEEvent{Data: "line1\nline2\nline3"},
			expected: "data: line1\ndata: line2\ndata: line3\n\n",
		},
		{
			name:     "full event",
			event:    SSEEvent{ID: "123", Event: "test", Data: "data", Retry: 5000},
			expected: "event: test\nid: 123\nretry: 5000\ndata: data\n\n",
		},
		{
			name:     "empty data",
			event:    SSEEvent{Event: "ping"},
			expected: "event: ping\n\n",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := NewSSEWriter(&buf)
			
			err := writer.WriteEvent(&tt.event)
			if err != nil {
				t.Fatalf("Failed to write event: %v", err)
			}
			
			result := buf.String()
			if result != tt.expected {
				t.Errorf("Expected:\n%q\nGot:\n%q", tt.expected, result)
			}
		})
	}
}

func TestStreamPipe(t *testing.T) {
	// Create input events
	input := "event: start\ndata: hello\n\nevent: end\ndata: world\n\n"
	reader := NewSSEReader(io.NopCloser(strings.NewReader(input)))
	
	// Create output buffer
	var output bytes.Buffer
	writer := NewSSEWriter(&output)
	
	// Test without transformation
	err := StreamPipe(reader, writer, nil)
	if err != nil {
		t.Fatalf("StreamPipe failed: %v", err)
	}
	
	// Should be the same as input
	if output.String() != input {
		t.Errorf("Expected output to match input")
	}
	
	// Test with transformation
	input2 := "data: test1\n\ndata: test2\n\ndata: test3\n\n"
	reader2 := NewSSEReader(io.NopCloser(strings.NewReader(input2)))
	
	var output2 bytes.Buffer
	writer2 := NewSSEWriter(&output2)
	
	// Transform: uppercase data and filter out "test2"
	transform := func(event *SSEEvent) (*SSEEvent, error) {
		if event.Data == "test2" {
			return nil, nil // Filter out
		}
		event.Data = strings.ToUpper(event.Data)
		return event, nil
	}
	
	err = StreamPipe(reader2, writer2, transform)
	if err != nil {
		t.Fatalf("StreamPipe with transform failed: %v", err)
	}
	
	expected := "data: TEST1\n\ndata: TEST3\n\n"
	if output2.String() != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, output2.String())
	}
}

func TestSSEReader_Close(t *testing.T) {
	// Test closing reader
	reader := NewSSEReader(io.NopCloser(strings.NewReader("data: test\n\n")))
	
	// Should be able to read
	event, err := reader.ReadEvent()
	if err != nil {
		t.Errorf("Failed to read before close: %v", err)
	}
	if event.Data != "test" {
		t.Errorf("Expected data 'test', got '%s'", event.Data)
	}
	
	// Close reader
	err = reader.Close()
	if err != nil {
		t.Errorf("Failed to close reader: %v", err)
	}
	
	// Should get EOF after close
	_, err = reader.ReadEvent()
	if err != io.EOF {
		t.Errorf("Expected EOF after close, got %v", err)
	}
	
	// Should be able to close again without error
	err = reader.Close()
	if err != nil {
		t.Errorf("Second close failed: %v", err)
	}
}

func TestSSEWriter_Close(t *testing.T) {
	var buf bytes.Buffer
	writer := NewSSEWriter(&buf)
	
	// Write event
	err := writer.WriteEvent(&SSEEvent{Data: "test"})
	if err != nil {
		t.Errorf("Failed to write: %v", err)
	}
	
	// Close writer
	err = writer.Close()
	if err != nil {
		t.Errorf("Failed to close writer: %v", err)
	}
	
	// Should error on write after close
	err = writer.WriteEvent(&SSEEvent{Data: "after close"})
	if err == nil {
		t.Error("Expected error writing after close")
	}
}