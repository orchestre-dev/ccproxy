package transformer

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestDeepSeekTransformer_TransformRequestIn(t *testing.T) {
	transformer := NewDeepSeekTransformer()
	ctx := context.Background()

	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "max_tokens under limit",
			input: map[string]interface{}{
				"model":      "deepseek-chat",
				"max_tokens": 4096.0,
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
				},
			},
			expected: map[string]interface{}{
				"model":      "deepseek-chat",
				"max_tokens": 4096.0,
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
				},
			},
		},
		{
			name: "max_tokens over limit",
			input: map[string]interface{}{
				"model":      "deepseek-chat",
				"max_tokens": 10000.0,
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
				},
			},
			expected: map[string]interface{}{
				"model":      "deepseek-chat",
				"max_tokens": 8192,
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
				},
			},
		},
		{
			name: "no max_tokens",
			input: map[string]interface{}{
				"model": "deepseek-chat",
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
				},
			},
			expected: map[string]interface{}{
				"model": "deepseek-chat",
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := transformer.TransformRequestIn(ctx, tt.input, "deepseek")
			if err != nil {
				t.Fatalf("TransformRequestIn() error = %v", err)
			}

			// Compare JSON representations
			expectedJSON, _ := json.MarshalIndent(tt.expected, "", "  ")
			resultJSON, _ := json.MarshalIndent(result, "", "  ")

			if string(expectedJSON) != string(resultJSON) {
				t.Errorf("TransformRequestIn() mismatch:\nExpected:\n%s\nGot:\n%s",
					string(expectedJSON), string(resultJSON))
			}
		})
	}
}

func TestDeepSeekTransformer_StreamingWithReasoning(t *testing.T) {
	transformer := NewDeepSeekTransformer()
	ctx := context.Background()

	// Create a streaming response with reasoning content
	events := []string{
		`data: {"id":"1","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"Let me think about this"},"finish_reason":null}]}`,
		`data: {"id":"1","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"reasoning_content":" step by step."},"finish_reason":null}]}`,
		`data: {"id":"1","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"Here's my answer:"},"finish_reason":null}]}`,
		`data: {"id":"1","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":" Hello!"},"finish_reason":null}]}`,
		`data: {"id":"1","object":"chat.completion.chunk","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`,
		`data: [DONE]`,
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

	// Verify reasoning was transformed to thinking
	foundThinkingDelta := false
	foundThinkingBlock := false
	foundRegularContent := false
	incrementedIndex := false

	for _, eventData := range transformedEvents {
		if eventData == "[DONE]" {
			continue
		}

		var chunk map[string]interface{}
		if err := json.Unmarshal([]byte(eventData), &chunk); err != nil {
			continue
		}

		choices, _ := chunk["choices"].([]interface{})
		if len(choices) > 0 {
			choice := choices[0].(map[string]interface{})
			delta, _ := choice["delta"].(map[string]interface{})
			
			// Check for thinking delta
			if thinking, ok := delta["thinking"].(map[string]interface{}); ok {
				if content, ok := thinking["content"].(string); ok && content != "" {
					foundThinkingDelta = true
				}
			}

			// Check for thinking block with signature
			if content, ok := delta["content"].(map[string]interface{}); ok {
				if _, hasContent := content["content"]; hasContent {
					if _, hasSignature := content["signature"]; hasSignature {
						foundThinkingBlock = true
					}
				}
			}

			// Check for regular content
			if content, ok := delta["content"].(string); ok && content != "" {
				foundRegularContent = true
				// Check if index was incremented
				if index, ok := choice["index"].(float64); ok && index > 0 {
					incrementedIndex = true
				}
			}

			// Ensure no reasoning_content in output
			if _, hasReasoning := delta["reasoning_content"]; hasReasoning {
				t.Error("reasoning_content should not be in transformed output")
			}
		}
	}

	if !foundThinkingDelta {
		t.Error("Expected to find thinking delta in transformed stream")
	}
	if !foundThinkingBlock {
		t.Error("Expected to find thinking block with signature")
	}
	if !foundRegularContent {
		t.Error("Expected to find regular content")
	}
	if !incrementedIndex {
		t.Error("Expected index to be incremented for content after reasoning")
	}
}

func TestDeepSeekTransformer_StreamingWithoutReasoning(t *testing.T) {
	transformer := NewDeepSeekTransformer()
	ctx := context.Background()

	// Create a streaming response without reasoning content
	events := []string{
		`data: {"id":"2","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"role":"assistant","content":"Hello"},"finish_reason":null}]}`,
		`data: {"id":"2","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":" world!"},"finish_reason":null}]}`,
		`data: {"id":"2","object":"chat.completion.chunk","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`,
		`data: [DONE]`,
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
	var contentParts []string

	for {
		event, err := reader.ReadEvent()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Failed to read transformed event: %v", err)
		}

		if event.Data == "[DONE]" {
			continue
		}

		var chunk map[string]interface{}
		if err := json.Unmarshal([]byte(event.Data), &chunk); err == nil {
			choices, _ := chunk["choices"].([]interface{})
			if len(choices) > 0 {
				choice := choices[0].(map[string]interface{})
				delta, _ := choice["delta"].(map[string]interface{})
				
				// Collect content
				if content, ok := delta["content"].(string); ok {
					contentParts = append(contentParts, content)
				}

				// Ensure no thinking blocks
				if _, hasThinking := delta["thinking"]; hasThinking {
					t.Error("Should not have thinking blocks without reasoning")
				}
			}
		}
	}

	// Verify content
	fullContent := strings.Join(contentParts, "")
	if fullContent != "Hello world!" {
		t.Errorf("Expected 'Hello world!', got '%s'", fullContent)
	}
}

func TestDeepSeekTransformer_NonStreamingPassthrough(t *testing.T) {
	transformer := NewDeepSeekTransformer()
	ctx := context.Background()

	// Create non-streaming response
	responseBody := `{"id":"3","object":"chat.completion","choices":[{"index":0,"message":{"role":"assistant","content":"Hello"},"finish_reason":"stop"}]}`
	
	resp := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(responseBody)),
	}
	resp.Header.Set("Content-Type", "application/json")

	// Transform response
	transformed, err := transformer.TransformResponseOut(ctx, resp)
	if err != nil {
		t.Fatalf("TransformResponseOut() error = %v", err)
	}

	// Read body
	body, err := io.ReadAll(transformed.Body)
	if err != nil {
		t.Fatalf("Failed to read body: %v", err)
	}

	// Should be unchanged
	if string(body) != responseBody {
		t.Errorf("Non-streaming response should pass through unchanged")
	}
}

func TestDeepSeekTransformer_ErrorHandling(t *testing.T) {
	transformer := NewDeepSeekTransformer()
	ctx := context.Background()

	// Test invalid request format
	_, err := transformer.TransformRequestIn(ctx, "invalid", "deepseek")
	if err == nil {
		t.Error("Expected error for invalid request format")
	}

	// Test malformed streaming data
	events := []string{
		`data: {"invalid json`,
		`data: {"id":"4","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"Still works"},"finish_reason":null}]}`,
		`data: [DONE]`,
	}

	sseData := strings.Join(events, "\n\n") + "\n\n"

	resp := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(sseData)),
	}
	resp.Header.Set("Content-Type", "text/event-stream")

	// Should not panic and should pass through valid data
	transformed, err := transformer.TransformResponseOut(ctx, resp)
	if err != nil {
		t.Fatalf("TransformResponseOut() should not fail on malformed data: %v", err)
	}

	// Verify we can still read valid events
	reader := NewSSEReader(transformed.Body)
	foundValidContent := false

	for {
		event, err := reader.ReadEvent()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		if strings.Contains(event.Data, "Still works") {
			foundValidContent = true
		}
	}

	if !foundValidContent {
		t.Error("Should pass through valid events even with malformed data")
	}
}