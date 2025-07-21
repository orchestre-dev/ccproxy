package transformer

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestAnthropicTransformer_TransformRequestIn(t *testing.T) {
	transformer := NewAnthropicTransformer()
	ctx := context.Background()

	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
		wantErr  bool
	}{
		{
			name: "basic message transformation",
			input: map[string]interface{}{
				"model":      "claude-3-opus-20240229",
				"max_tokens": 1000,
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
				},
			},
			expected: map[string]interface{}{
				"model":      "claude-3-opus-20240229",
				"max_tokens": 1000,
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
				},
			},
		},
		{
			name: "system message extraction",
			input: map[string]interface{}{
				"model": "claude-3-opus-20240229",
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "system",
						"content": "You are a helpful assistant",
					},
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
				},
			},
			expected: map[string]interface{}{
				"model":  "claude-3-opus-20240229",
				"system": "You are a helpful assistant",
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
				},
			},
		},
		{
			name: "tool transformation",
			input: map[string]interface{}{
				"model": "claude-3-opus-20240229",
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "What's the weather?",
					},
				},
				"tools": []interface{}{
					map[string]interface{}{
						"type": "function",
						"function": map[string]interface{}{
							"name":        "get_weather",
							"description": "Get weather info",
							"parameters": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"location": map[string]interface{}{
										"type": "string",
									},
								},
								"required": []interface{}{"location"},
							},
						},
					},
				},
				"tool_choice": "auto",
			},
			expected: map[string]interface{}{
				"model": "claude-3-opus-20240229",
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "What's the weather?",
					},
				},
				"tools": []interface{}{
					map[string]interface{}{
						"name":        "get_weather",
						"description": "Get weather info",
						"input_schema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"location": map[string]interface{}{
									"type": "string",
								},
							},
							"required": []interface{}{"location"},
						},
					},
				},
				"tool_choice": map[string]interface{}{"type": "any"},
			},
		},
		{
			name: "tool call transformation",
			input: map[string]interface{}{
				"model": "claude-3-opus-20240229",
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "What's the weather in SF?",
					},
					map[string]interface{}{
						"role":    "assistant",
						"content": "I'll check the weather for you.",
						"tool_calls": []interface{}{
							map[string]interface{}{
								"id":   "call_123",
								"type": "function",
								"function": map[string]interface{}{
									"name":      "get_weather",
									"arguments": `{"location": "San Francisco"}`,
								},
							},
						},
					},
					map[string]interface{}{
						"role":         "tool",
						"content":      "72°F, sunny",
						"tool_call_id": "call_123",
					},
				},
			},
			expected: map[string]interface{}{
				"model": "claude-3-opus-20240229",
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "What's the weather in SF?",
					},
					map[string]interface{}{
						"role": "assistant",
						"content": []interface{}{
							map[string]interface{}{
								"type": "text",
								"text": "I'll check the weather for you.",
							},
							map[string]interface{}{
								"type": "tool_use",
								"id":   "call_123",
								"name": "get_weather",
								"input": map[string]interface{}{
									"location": "San Francisco",
								},
							},
						},
					},
					map[string]interface{}{
						"role": "user",
						"content": []interface{}{
							map[string]interface{}{
								"type":        "tool_result",
								"tool_use_id": "call_123",
								"content":     "72°F, sunny",
							},
						},
					},
				},
			},
		},
		{
			name: "thinking parameter",
			input: map[string]interface{}{
				"model":    "claude-3-opus-20240229",
				"thinking": true,
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Solve this complex problem",
					},
				},
			},
			expected: map[string]interface{}{
				"model": "claude-3-opus-20240229",
				"thinking": map[string]interface{}{
					"type":         "enabled",
					"budget_tokens": 16000,
				},
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Solve this complex problem",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := transformer.TransformRequestIn(ctx, tt.input, "anthropic")
			if (err != nil) != tt.wantErr {
				t.Errorf("TransformRequestIn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Compare JSON representations for better error messages
				expectedJSON, _ := json.MarshalIndent(tt.expected, "", "  ")
				resultJSON, _ := json.MarshalIndent(result, "", "  ")
				
				if string(expectedJSON) != string(resultJSON) {
					t.Errorf("TransformRequestIn() mismatch:\nExpected:\n%s\nGot:\n%s", 
						string(expectedJSON), string(resultJSON))
				}
			}
		})
	}
}

func TestAnthropicTransformer_TransformNonStreamingResponse(t *testing.T) {
	transformer := NewAnthropicTransformer()
	ctx := context.Background()

	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "basic text response",
			input: map[string]interface{}{
				"id":      "msg_123",
				"type":    "message",
				"role":    "assistant",
				"model":   "claude-3-opus-20240229",
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "Hello! How can I help you?",
					},
				},
				"stop_reason": "end_turn",
				"usage": map[string]interface{}{
					"input_tokens":  10.0,
					"output_tokens": 20.0,
				},
			},
			expected: map[string]interface{}{
				"id":     "msg_123",
				"object": "chat.completion",
				"model":  "claude-3-opus-20240229",
				"choices": []interface{}{
					map[string]interface{}{
						"index": 0,
						"message": map[string]interface{}{
							"role":    "assistant",
							"content": "Hello! How can I help you?",
						},
						"finish_reason": "stop",
					},
				},
				"usage": map[string]interface{}{
					"prompt_tokens":     10.0,
					"completion_tokens": 20.0,
					"total_tokens":      30,
				},
			},
		},
		{
			name: "tool use response",
			input: map[string]interface{}{
				"id":    "msg_456",
				"model": "claude-3-opus-20240229",
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "I'll check the weather for you.",
					},
					map[string]interface{}{
						"type": "tool_use",
						"id":   "toolu_123",
						"name": "get_weather",
						"input": map[string]interface{}{
							"location": "San Francisco",
						},
					},
				},
				"stop_reason": "tool_use",
				"usage": map[string]interface{}{
					"input_tokens":  15.0,
					"output_tokens": 25.0,
				},
			},
			expected: map[string]interface{}{
				"id":     "msg_456",
				"object": "chat.completion",
				"model":  "claude-3-opus-20240229",
				"choices": []interface{}{
					map[string]interface{}{
						"index": 0,
						"message": map[string]interface{}{
							"role":    "assistant",
							"content": "I'll check the weather for you.",
							"tool_calls": []interface{}{
								map[string]interface{}{
									"id":   "toolu_123",
									"type": "function",
									"function": map[string]interface{}{
										"name":      "get_weather",
										"arguments": `{"location":"San Francisco"}`,
									},
								},
							},
						},
						"finish_reason": "tool_calls",
					},
				},
				"usage": map[string]interface{}{
					"prompt_tokens":     15.0,
					"completion_tokens": 25.0,
					"total_tokens":      40,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create response body
			body, _ := json.Marshal(tt.input)
			resp := &http.Response{
				StatusCode: 200,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewReader(body)),
			}
			resp.Header.Set("Content-Type", "application/json")

			// Transform response
			transformed, err := transformer.transformNonStreamingResponse(ctx, resp)
			if err != nil {
				t.Fatalf("transformNonStreamingResponse() error = %v", err)
			}

			// Read transformed body
			transformedBody, err := io.ReadAll(transformed.Body)
			if err != nil {
				t.Fatalf("Failed to read transformed body: %v", err)
			}

			// Parse result
			var result map[string]interface{}
			if err := json.Unmarshal(transformedBody, &result); err != nil {
				t.Fatalf("Failed to parse transformed response: %v", err)
			}

			// Remove created timestamp for comparison
			delete(result, "created")
			delete(tt.expected, "created")

			// Compare
			expectedJSON, _ := json.MarshalIndent(tt.expected, "", "  ")
			resultJSON, _ := json.MarshalIndent(result, "", "  ")
			
			if string(expectedJSON) != string(resultJSON) {
				t.Errorf("Transform mismatch:\nExpected:\n%s\nGot:\n%s", 
					string(expectedJSON), string(resultJSON))
			}
		})
	}
}

func TestAnthropicTransformer_StreamingTransformation(t *testing.T) {
	transformer := NewAnthropicTransformer()
	ctx := context.Background()

	// Create a streaming response with Anthropic events
	events := []string{
		`event: message_start
data: {"type":"message_start","message":{"id":"msg_123","type":"message","role":"assistant","model":"claude-3-opus-20240229","content":[]}}`,
		`event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
		`event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}`,
		`event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" world!"}}`,
		`event: content_block_stop
data: {"type":"content_block_stop","index":0}`,
		`event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":10}}`,
		`event: message_stop
data: {"type":"message_stop"}`,
	}

	sseData := strings.Join(events, "\n\n") + "\n\n"
	
	resp := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(sseData)),
	}
	resp.Header.Set("Content-Type", "text/event-stream")

	// Transform response
	transformed, err := transformer.TransformResponseOut(ctx, resp)
	if err != nil {
		t.Fatalf("TransformResponseOut() error = %v", err)
	}

	// Read transformed events
	reader := NewSSEReader(transformed.Body)
	var transformedEvents []string
	
	for {
		event, err := reader.ReadEvent()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Failed to read transformed event: %v", err)
		}
		transformedEvents = append(transformedEvents, event.Data)
	}

	// Verify we have events
	if len(transformedEvents) == 0 {
		t.Error("No transformed events received")
	}

	// Check for [DONE] event
	lastEvent := transformedEvents[len(transformedEvents)-1]
	if lastEvent != "[DONE]" {
		t.Errorf("Expected last event to be [DONE], got: %s", lastEvent)
	}

	// Verify text content was streamed
	foundHello := false
	foundWorld := false
	
	for _, eventData := range transformedEvents {
		if strings.Contains(eventData, `"content":"Hello"`) {
			foundHello = true
		}
		if strings.Contains(eventData, `"content":" world!"`) {
			foundWorld = true
		}
	}
	
	if !foundHello || !foundWorld {
		t.Error("Text content not properly streamed")
	}
}

func TestAnthropicTransformer_ToolStreamingTransformation(t *testing.T) {
	transformer := NewAnthropicTransformer()
	ctx := context.Background()

	// Create a streaming response with tool use
	events := []string{
		`event: message_start
data: {"type":"message_start","message":{"id":"msg_456","type":"message","role":"assistant","model":"claude-3-opus-20240229"}}`,
		`event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_789","name":"get_weather","input":{}}}`,
		`event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"location\":"}}`,
		`event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":" \"San Francisco\"}"}}`,
		`event: content_block_stop
data: {"type":"content_block_stop","index":0}`,
		`event: message_stop
data: {"type":"message_stop"}`,
	}

	sseData := strings.Join(events, "\n\n") + "\n\n"
	
	resp := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(sseData)),
	}
	resp.Header.Set("Content-Type", "text/event-stream")

	// Transform response
	transformed, err := transformer.TransformResponseOut(ctx, resp)
	if err != nil {
		t.Fatalf("TransformResponseOut() error = %v", err)
	}

	// Read transformed events
	reader := NewSSEReader(transformed.Body)
	var foundToolCall bool
	var foundArguments bool
	
	for {
		event, err := reader.ReadEvent()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Failed to read transformed event: %v", err)
		}
		
		if strings.Contains(event.Data, `"name":"get_weather"`) {
			foundToolCall = true
		}
		if strings.Contains(event.Data, `"arguments":`) && 
		   (strings.Contains(event.Data, `"{\"location\":`) || 
		    strings.Contains(event.Data, `" \"San Francisco\"}"`)) {
			foundArguments = true
		}
	}

	if !foundToolCall {
		t.Error("Tool call not found in transformed stream")
	}
	if !foundArguments {
		t.Error("Tool arguments not properly streamed")
	}
}

func TestAnthropicTransformer_ConvertStopReason(t *testing.T) {
	transformer := NewAnthropicTransformer()

	tests := []struct {
		input    string
		expected string
	}{
		{"end_turn", "stop"},
		{"max_tokens", "length"},
		{"tool_use", "tool_calls"},
		{"unknown", "stop"},
		{"", "stop"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := transformer.convertStopReason(tt.input)
			if result != tt.expected {
				t.Errorf("convertStopReason(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestAnthropicTransformer_ErrorHandling(t *testing.T) {
	transformer := NewAnthropicTransformer()
	ctx := context.Background()

	// Test invalid request format
	_, err := transformer.TransformRequestIn(ctx, "invalid", "anthropic")
	if err == nil {
		t.Error("Expected error for invalid request format")
	}

	// Test missing messages
	_, err = transformer.TransformRequestIn(ctx, map[string]interface{}{
		"model": "claude-3-opus-20240229",
	}, "anthropic")
	if err == nil {
		t.Error("Expected error for missing messages")
	}
}