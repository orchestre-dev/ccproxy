package tools

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestProcessToolStream(t *testing.T) {
	sh := NewStreamingHandler()
	ctx := context.Background()

	tests := []struct {
		name     string
		input    string
		wantErr  bool
		validate func(t *testing.T, output string)
	}{
		{
			name: "simple text event",
			input: `data: {"type": "text", "text": "Hello"}

`,
			validate: func(t *testing.T, output string) {
				if !strings.Contains(output, "Hello") {
					t.Error("Expected output to contain 'Hello'")
				}
			},
		},
		{
			name: "tool use event",
			input: `data: {"delta": {"type": "tool_use", "name": "calculator", "id": "calc_001"}}

`,
			validate: func(t *testing.T, output string) {
				if !strings.Contains(output, "tool_use") {
					t.Error("Expected output to contain 'tool_use'")
				}
				if !strings.Contains(output, "_processed") {
					t.Error("Expected tool use to be marked as processed")
				}
				if !strings.Contains(output, "has_tool_use") {
					t.Error("Expected metadata to indicate tool use")
				}
			},
		},
		{
			name: "multiple events",
			input: `event: message_start
data: {"type": "message_start"}

data: {"delta": {"type": "text", "text": "Let me calculate that."}}

data: {"delta": {"type": "tool_use", "name": "calculator"}}

data: [DONE]

`,
			validate: func(t *testing.T, output string) {
				if !strings.Contains(output, "message_start") {
					t.Error("Expected event type to be preserved")
				}
				if !strings.Contains(output, "Let me calculate that") {
					t.Error("Expected text content to be preserved")
				}
				if !strings.Contains(output, "_processed") {
					t.Error("Expected tool use to be processed")
				}
				if !strings.Contains(output, "[DONE]") {
					t.Error("Expected [DONE] to be preserved")
				}
			},
		},
		{
			name: "multiline data",
			input: `data: {
data:   "type": "text",
data:   "text": "Multi\nline"
data: }

`,
			validate: func(t *testing.T, output string) {
				// Should preserve multiline structure
				if !strings.Contains(output, "data: {") {
					t.Error("Expected multiline data to be preserved")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			var writer bytes.Buffer

			err := sh.ProcessToolStream(ctx, reader, &writer)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessToolStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			output := writer.String()
			if tt.validate != nil {
				tt.validate(t, output)
			}
		})
	}
}

func TestProcessSSEEvent(t *testing.T) {
	sh := NewStreamingHandler()
	ctx := context.Background()

	tests := []struct {
		name  string
		event SSEEvent
		want  func(t *testing.T, result SSEEvent)
	}{
		{
			name: "non-JSON data",
			event: SSEEvent{
				Data: "plain text",
			},
			want: func(t *testing.T, result SSEEvent) {
				if result.Data != "plain text" {
					t.Error("Non-JSON data should remain unchanged")
				}
			},
		},
		{
			name: "JSON without tool use",
			event: SSEEvent{
				Data: `{"type": "text", "text": "Hello"}`,
			},
			want: func(t *testing.T, result SSEEvent) {
				if result.Data != `{"type": "text", "text": "Hello"}` {
					t.Error("JSON without tool use should remain unchanged")
				}
			},
		},
		{
			name: "delta with tool use",
			event: SSEEvent{
				Event: "content_block_delta",
				Data:  `{"delta": {"type": "tool_use", "name": "search"}}`,
			},
			want: func(t *testing.T, result SSEEvent) {
				if !strings.Contains(result.Data, "_processed") {
					t.Error("Tool use delta should be marked as processed")
				}
				if !strings.Contains(result.Data, "has_tool_use") {
					t.Error("Metadata should indicate tool use")
				}
				if result.Event != "content_block_delta" {
					t.Error("Event type should be preserved")
				}
			},
		},
		{
			name: "[DONE] event",
			event: SSEEvent{
				Data: "[DONE]",
			},
			want: func(t *testing.T, result SSEEvent) {
				if result.Data != "[DONE]" {
					t.Error("[DONE] should remain unchanged")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sh.processSSEEvent(ctx, tt.event)
			if err != nil {
				t.Fatalf("processSSEEvent() error = %v", err)
			}
			if tt.want != nil {
				tt.want(t, result)
			}
		})
	}
}

func TestWriteSSEEvent(t *testing.T) {
	sh := NewStreamingHandler()

	tests := []struct {
		name  string
		event SSEEvent
		want  string
	}{
		{
			name: "simple data event",
			event: SSEEvent{
				Data: "Hello, world!",
			},
			want: "data: Hello, world!\n\n",
		},
		{
			name: "event with all fields",
			event: SSEEvent{
				ID:    "123",
				Event: "message",
				Data:  "Test data",
				Retry: "5000",
			},
			want: "id: 123\nevent: message\nretry: 5000\ndata: Test data\n\n",
		},
		{
			name: "multiline data",
			event: SSEEvent{
				Data: "Line 1\nLine 2\nLine 3",
			},
			want: "data: Line 1\ndata: Line 2\ndata: Line 3\n\n",
		},
		{
			name: "empty event",
			event: SSEEvent{},
			want: "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := sh.writeSSEEvent(&buf, tt.event)
			if err != nil {
				t.Fatalf("writeSSEEvent() error = %v", err)
			}
			
			got := buf.String()
			if got != tt.want {
				t.Errorf("writeSSEEvent() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractToolUseFromStream(t *testing.T) {
	sh := NewStreamingHandler()

	tests := []struct {
		name    string
		event   SSEEvent
		want    *ToolUse
		wantNil bool
	}{
		{
			name: "delta with tool use",
			event: SSEEvent{
				Data: `{"delta": {"type": "tool_use", "id": "tool_123", "name": "calc", "input": {"a": 1}}}`,
			},
			want: &ToolUse{
				Type: "tool_use",
				ID:   "tool_123",
				Name: "calc",
			},
		},
		{
			name: "content with tool use",
			event: SSEEvent{
				Data: `{"content": [{"type": "tool_use", "id": "search_001", "name": "search"}]}`,
			},
			want: &ToolUse{
				Type: "tool_use",
				ID:   "search_001",
				Name: "search",
			},
		},
		{
			name: "no tool use",
			event: SSEEvent{
				Data: `{"type": "text", "text": "Hello"}`,
			},
			wantNil: true,
		},
		{
			name: "[DONE]",
			event: SSEEvent{
				Data: "[DONE]",
			},
			wantNil: true,
		},
		{
			name: "invalid JSON",
			event: SSEEvent{
				Data: "not json",
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sh.ExtractToolUseFromStream(tt.event)
			if err != nil {
				t.Fatalf("ExtractToolUseFromStream() error = %v", err)
			}
			
			if tt.wantNil {
				if got != nil {
					t.Errorf("Expected nil, got %+v", got)
				}
				return
			}
			
			if got == nil {
				t.Fatal("Expected tool use, got nil")
			}
			
			if got.Type != tt.want.Type || got.ID != tt.want.ID || got.Name != tt.want.Name {
				t.Errorf("ExtractToolUseFromStream() = %+v, want %+v", got, tt.want)
			}
		})
	}
}