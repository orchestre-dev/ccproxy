package transformer

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestNewOpenRouterTransformer(t *testing.T) {
	trans := NewOpenRouterTransformer()
	
	if trans == nil {
		t.Fatal("Expected non-nil transformer")
	}
	
	// Check name
	if trans.GetName() != "openrouter" {
		t.Errorf("Expected name 'openrouter', got %s", trans.GetName())
	}
	
	// Check endpoint
	if trans.GetEndpoint() != "/api/v1/chat/completions" {
		t.Errorf("Expected endpoint '/api/v1/chat/completions', got %s", trans.GetEndpoint())
	}
}

func TestOpenRouterTransformer_TransformResponseOut(t *testing.T) {
	trans := NewOpenRouterTransformer()
	ctx := context.Background()
	
	// Test non-streaming response (should pass through)
	nonStreamResp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: io.NopCloser(strings.NewReader(`{"result": "test"}`)),
	}
	
	result, err := trans.TransformResponseOut(ctx, nonStreamResp)
	if err != nil {
		t.Errorf("Unexpected error for non-streaming response: %v", err)
	}
	if result != nonStreamResp {
		t.Error("Expected non-streaming response to pass through unchanged")
	}
	
	// Test streaming response with reasoning content
	streamResp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{
			"Content-Type": []string{"text/event-stream"},
		},
		Body: io.NopCloser(strings.NewReader(
			`data: {"choices":[{"delta":{"reasoning_content":"Let me think"}}]}` + "\n\n" +
			`data: {"choices":[{"delta":{"reasoning_content":" about this."}}]}` + "\n\n" +
			`data: {"choices":[{"delta":{"content":"The answer is 42."}}]}` + "\n\n" +
			`data: {"choices":[{"finish_reason":"stop"}]}` + "\n\n" +
			`data: [DONE]` + "\n\n",
		)),
	}
	
	result, err = trans.TransformResponseOut(ctx, streamResp)
	if err != nil {
		t.Fatalf("Failed to transform streaming response: %v", err)
	}
	
	// Small delay to allow the goroutine to process
	time.Sleep(10 * time.Millisecond)
	
	// Read transformed response
	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("Failed to read transformed body: %v", err)
	}
	
	responseText := string(body)
	
	// Debug: print the actual response
	// t.Logf("Response: %s", responseText)
	
	// Should have transformed reasoning_content to thinking
	// The thinking content is in a nested structure: delta.thinking.content
	if !strings.Contains(responseText, `"thinking"`) {
		t.Error("Expected reasoning_content to be transformed to thinking")
	}
	
	// Should have the thinking content in the proper structure
	// Looking for the thinking content in the delta
	if !strings.Contains(responseText, "Let me think") {
		t.Error("Expected thinking content to be present")
	}
	
	// Should have regular content
	if !strings.Contains(responseText, "The answer is 42.") {
		t.Error("Expected regular content to be preserved")
	}
}

func TestOpenRouterTransformer_transformStreamData(t *testing.T) {
	trans := NewOpenRouterTransformer()
	
	tests := []struct {
		name           string
		data           string
		state          *openrouterStreamState
		expectedCount  int
		checkContent   []string
	}{
		{
			name: "reasoning content only",
			data: `{"choices":[{"delta":{"reasoning_content":"Thinking..."}}]}`,
			state: &openrouterStreamState{
				reasoningContent:    "",
				isReasoningComplete: false,
				contentIndex:        0,
			},
			expectedCount: 1,
			checkContent: []string{`"thinking"`},
		},
		{
			name: "transition from reasoning to content",
			data: `{"choices":[{"delta":{"content":"Answer"}}]}`,
			state: &openrouterStreamState{
				reasoningContent:    "Previous thinking",
				isReasoningComplete: false,
				contentIndex:        0,
			},
			expectedCount: 2, // thinking block + content
			checkContent: []string{"Previous thinking", "Answer"},
		},
		{
			name: "tool calls",
			data: `{"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"name":"test"}}]}}]}`,
			state: &openrouterStreamState{
				reasoningContent:    "Thinking",
				isReasoningComplete: false,
				contentIndex:        0,
			},
			expectedCount: 2, // thinking block + tool call
			checkContent: []string{"test"},
		},
		{
			name: "finish reason with incomplete reasoning",
			data: `{"choices":[{"index":0,"finish_reason":"stop"}]}`,
			state: &openrouterStreamState{
				reasoningContent:    "Incomplete thinking",
				isReasoningComplete: false,
				contentIndex:        0,
			},
			expectedCount: 2, // thinking block + finish
			checkContent: []string{"Incomplete thinking", "stop"},
		},
		{
			name: "invalid json",
			data: `invalid json`,
			state: &openrouterStreamState{},
			expectedCount: 1,
			checkContent: []string{"invalid json"},
		},
		{
			name: "no choices",
			data: `{"other":"data"}`,
			state: &openrouterStreamState{},
			expectedCount: 1,
			checkContent: []string{`"other"`},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := trans.transformStreamData(tt.data, tt.state)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if len(results) != tt.expectedCount {
				t.Errorf("Expected %d results, got %d", tt.expectedCount, len(results))
			}
			
			// Check that expected content is present
			allContent := strings.Join(results, " ")
			for _, check := range tt.checkContent {
				if !strings.Contains(allContent, check) {
					t.Errorf("Expected content '%s' not found in results", check)
				}
			}
		})
	}
}

func TestOpenRouterTransformer_createThinkingBlockChunk(t *testing.T) {
	trans := NewOpenRouterTransformer()
	
	baseChunk := map[string]interface{}{
		"id": "test-id",
		"created": 1234567890,
	}
	
	content := "This is my thinking"
	index := 1
	
	result := trans.createThinkingBlockChunk(baseChunk, content, index)
	
	// Check basic structure
	if result["id"] != "test-id" {
		t.Error("Expected base chunk properties to be preserved")
	}
	
	// Check choices structure
	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) != 1 {
		t.Fatal("Expected choices array with one element")
	}
	
	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		t.Fatal("Expected choice to be a map")
	}
	
	// Check index
	if choice["index"] != index {
		t.Errorf("Expected index %d, got %v", index, choice["index"])
	}
	
	// Check delta
	delta, ok := choice["delta"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected delta to be a map")
	}
	
	// Check content structure
	contentMap, ok := delta["content"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected content to be a map")
	}
	
	if contentMap["content"] != content {
		t.Errorf("Expected content '%s', got %v", content, contentMap["content"])
	}
	
	// Should have signature
	if _, hasSignature := contentMap["signature"]; !hasSignature {
		t.Error("Expected signature in content map")
	}
}

// Test error handling in streaming
func TestOpenRouterTransformer_StreamingErrorHandling(t *testing.T) {
	trans := NewOpenRouterTransformer()
	ctx := context.Background()
	
	// Test with reader that errors
	errorReader := &errorReader{err: io.ErrUnexpectedEOF}
	streamResp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{
			"Content-Type": []string{"text/event-stream"},
		},
		Body: io.NopCloser(errorReader),
	}
	
	result, err := trans.TransformResponseOut(ctx, streamResp)
	if err != nil {
		t.Errorf("TransformResponseOut should not error immediately: %v", err)
	}
	
	// Error should occur when reading the body
	_, readErr := io.ReadAll(result.Body)
	if readErr != nil {
		// This is expected - the error occurs during streaming
		t.Logf("Expected error during streaming: %v", readErr)
	}
}

// Helper error reader for testing
type errorReader struct {
	err error
}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, e.err
}