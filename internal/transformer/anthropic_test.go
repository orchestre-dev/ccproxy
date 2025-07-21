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

func TestNewAnthropicTransformer(t *testing.T) {
	transformer := NewAnthropicTransformer()
	
	if transformer.GetName() != "anthropic" {
		t.Errorf("Expected name 'anthropic', got %s", transformer.GetName())
	}
	
	if transformer.GetEndpoint() != "/v1/messages" {
		t.Errorf("Expected endpoint '/v1/messages', got %s", transformer.GetEndpoint())
	}
}

func TestAnthropicTransformer_TransformRequestIn(t *testing.T) {
	transformer := NewAnthropicTransformer()
	ctx := context.Background()

	t.Run("BasicMessageTransformation", func(t *testing.T) {
		request := map[string]interface{}{
			"model": "claude-3-haiku",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Hello, world!",
				},
			},
			"max_tokens":  1000,
			"temperature": 0.7,
		}
		
		result, err := transformer.TransformRequestIn(ctx, request, "anthropic")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("Result is not a map")
		}
		
		if resultMap["model"] != "claude-3-haiku" {
			t.Errorf("Expected model 'claude-3-haiku', got %v", resultMap["model"])
		}
		
		if resultMap["max_tokens"] != 1000 {
			t.Errorf("Expected max_tokens 1000, got %v", resultMap["max_tokens"])
		}
		
		if resultMap["temperature"] != 0.7 {
			t.Errorf("Expected temperature 0.7, got %v", resultMap["temperature"])
		}
	})

	t.Run("SystemMessageExtraction", func(t *testing.T) {
		request := map[string]interface{}{
			"model": "claude-3-haiku",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "system",
					"content": "You are a helpful assistant.",
				},
				map[string]interface{}{
					"role":    "user",
					"content": "Hello!",
				},
			},
		}
		
		result, err := transformer.TransformRequestIn(ctx, request, "anthropic")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("Result is not a map")
		}
		
		if resultMap["system"] != "You are a helpful assistant." {
			t.Errorf("Expected system message to be extracted, got %v", resultMap["system"])
		}
		
		messages := resultMap["messages"].([]interface{})
		if len(messages) != 1 {
			t.Errorf("Expected 1 message after system extraction, got %d", len(messages))
		}
	})

	t.Run("ModelWithProviderPrefix", func(t *testing.T) {
		request := map[string]interface{}{
			"model": "anthropic,claude-3-sonnet",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Test",
				},
			},
		}
		
		result, err := transformer.TransformRequestIn(ctx, request, "anthropic")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		resultMap := result.(map[string]interface{})
		if resultMap["model"] != "claude-3-sonnet" {
			t.Errorf("Expected model 'claude-3-sonnet', got %v", resultMap["model"])
		}
	})

	t.Run("ToolsTransformation", func(t *testing.T) {
		request := map[string]interface{}{
			"model": "claude-3-haiku",
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
						"description": "Get weather information",
						"parameters": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"location": map[string]interface{}{
									"type": "string",
									"description": "The location",
								},
							},
						},
					},
				},
			},
			"tool_choice": "auto",
		}
		
		result, err := transformer.TransformRequestIn(ctx, request, "anthropic")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		resultMap := result.(map[string]interface{})
		
		tools, ok := resultMap["tools"].([]interface{})
		if !ok || len(tools) != 1 {
			t.Errorf("Expected 1 tool, got %v", tools)
		}
		
		tool := tools[0].(map[string]interface{})
		if tool["name"] != "get_weather" {
			t.Errorf("Expected tool name 'get_weather', got %v", tool["name"])
		}
		
		toolChoice := resultMap["tool_choice"].(map[string]interface{})
		if toolChoice["type"] != "any" {
			t.Errorf("Expected tool_choice type 'any', got %v", toolChoice["type"])
		}
	})

	t.Run("ThinkingParameter", func(t *testing.T) {
		request := map[string]interface{}{
			"model": "claude-3-haiku",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Think step by step",
				},
			},
			"thinking": true,
		}
		
		result, err := transformer.TransformRequestIn(ctx, request, "anthropic")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		resultMap := result.(map[string]interface{})
		
		thinking, ok := resultMap["thinking"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected thinking parameter to be transformed")
		}
		
		if thinking["type"] != "enabled" {
			t.Errorf("Expected thinking type 'enabled', got %v", thinking["type"])
		}
		
		if thinking["budget_tokens"] != 16000 {
			t.Errorf("Expected budget_tokens 16000, got %v", thinking["budget_tokens"])
		}
	})

	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := transformer.TransformRequestIn(ctx, "invalid", "anthropic")
		if err == nil {
			t.Error("Expected error for invalid request format")
		}
	})

	t.Run("MissingMessages", func(t *testing.T) {
		request := map[string]interface{}{
			"model": "claude-3-haiku",
		}
		
		_, err := transformer.TransformRequestIn(ctx, request, "anthropic")
		if err == nil {
			t.Error("Expected error for missing messages")
		}
	})
}

func TestAnthropicTransformer_TransformResponseOut(t *testing.T) {
	transformer := NewAnthropicTransformer()
	ctx := context.Background()

	t.Run("NonStreamingResponse", func(t *testing.T) {
		anthropicResp := map[string]interface{}{
			"id":    "msg_123",
			"model": "claude-3-haiku",
			"content": []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "Hello, how can I help?",
				},
			},
			"stop_reason": "end_turn",
			"usage": map[string]interface{}{
				"input_tokens":  10,
				"output_tokens": 15,
			},
		}
		
		body, _ := json.Marshal(anthropicResp)
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(bytes.NewReader(body)),
		}
		
		result, err := transformer.TransformResponseOut(ctx, resp)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		resultBody, _ := io.ReadAll(result.Body)
		var openaiResp map[string]interface{}
		json.Unmarshal(resultBody, &openaiResp)
		
		if openaiResp["object"] != "chat.completion" {
			t.Errorf("Expected object 'chat.completion', got %v", openaiResp["object"])
		}
		
		if openaiResp["model"] != "claude-3-haiku" {
			t.Errorf("Expected model 'claude-3-haiku', got %v", openaiResp["model"])
		}
		
		choices := openaiResp["choices"].([]interface{})
		if len(choices) != 1 {
			t.Errorf("Expected 1 choice, got %d", len(choices))
		}
		
		choice := choices[0].(map[string]interface{})
		message := choice["message"].(map[string]interface{})
		if message["content"] != "Hello, how can I help?" {
			t.Errorf("Expected content 'Hello, how can I help?', got %v", message["content"])
		}
		
		if choice["finish_reason"] != "stop" {
			t.Errorf("Expected finish_reason 'stop', got %v", choice["finish_reason"])
		}
		
		usage := openaiResp["usage"].(map[string]interface{})
		if usage["prompt_tokens"] != float64(10) {
			t.Errorf("Expected prompt_tokens 10, got %v", usage["prompt_tokens"])
		}
		if usage["completion_tokens"] != float64(15) {
			t.Errorf("Expected completion_tokens 15, got %v", usage["completion_tokens"])
		}
		totalTokensValue := usage["total_tokens"]
		if totalTokensValue != float64(25) {
			t.Errorf("Expected total_tokens 25, got %v (type: %T)", totalTokensValue, totalTokensValue)
		}
	})

	t.Run("ResponseWithToolUse", func(t *testing.T) {
		anthropicResp := map[string]interface{}{
			"id":    "msg_456",
			"model": "claude-3-haiku",
			"content": []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "I'll check the weather for you.",
				},
				map[string]interface{}{
					"type": "tool_use",
					"id":   "tool_123",
					"name": "get_weather",
					"input": map[string]interface{}{
						"location": "San Francisco",
					},
				},
			},
			"stop_reason": "tool_use",
		}
		
		body, _ := json.Marshal(anthropicResp)
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(bytes.NewReader(body)),
		}
		
		result, err := transformer.TransformResponseOut(ctx, resp)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		resultBody, _ := io.ReadAll(result.Body)
		var openaiResp map[string]interface{}
		json.Unmarshal(resultBody, &openaiResp)
		
		choices := openaiResp["choices"].([]interface{})
		choice := choices[0].(map[string]interface{})
		message := choice["message"].(map[string]interface{})
		
		if message["content"] != "I'll check the weather for you." {
			t.Errorf("Expected text content, got %v", message["content"])
		}
		
		toolCalls := message["tool_calls"].([]interface{})
		if len(toolCalls) != 1 {
			t.Errorf("Expected 1 tool call, got %d", len(toolCalls))
		}
		
		toolCall := toolCalls[0].(map[string]interface{})
		if toolCall["id"] != "tool_123" {
			t.Errorf("Expected tool call id 'tool_123', got %v", toolCall["id"])
		}
		
		function := toolCall["function"].(map[string]interface{})
		if function["name"] != "get_weather" {
			t.Errorf("Expected function name 'get_weather', got %v", function["name"])
		}
		
		if choice["finish_reason"] != "tool_calls" {
			t.Errorf("Expected finish_reason 'tool_calls', got %v", choice["finish_reason"])
		}
	})

	t.Run("StreamingResponse", func(t *testing.T) {
		// Test streaming response handling
		sseData := "data: {\"type\": \"message_start\", \"message\": {\"id\": \"msg_123\", \"model\": \"claude-3-haiku\"}}\n\n"
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(sseData)),
		}
		
		result, err := transformer.TransformResponseOut(ctx, resp)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if !strings.Contains(result.Header.Get("Content-Type"), "text/event-stream") {
			t.Error("Expected streaming response to maintain event-stream content type")
		}
	})

	t.Run("InvalidJSONResponse", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader("invalid json")),
		}
		
		result, err := transformer.TransformResponseOut(ctx, resp)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		// Should return original response when JSON parsing fails
		resultBody, _ := io.ReadAll(result.Body)
		if string(resultBody) != "invalid json" {
			t.Error("Expected original response body when JSON parsing fails")
		}
	})
}

func TestAnthropicTransformer_ConvertStopReason(t *testing.T) {
	transformer := NewAnthropicTransformer()
	
	tests := []struct {
		anthropicReason string
		expectedOpenAI  string
	}{
		{"end_turn", "stop"},
		{"max_tokens", "length"},
		{"tool_use", "tool_calls"},
		{"unknown", "stop"},
		{"", "stop"},
	}
	
	for _, test := range tests {
		result := transformer.convertStopReason(test.anthropicReason)
		if result != test.expectedOpenAI {
			t.Errorf("For reason %q, expected %q, got %q", test.anthropicReason, test.expectedOpenAI, result)
		}
	}
}

func TestAnthropicTransformer_SumTokens(t *testing.T) {
	transformer := NewAnthropicTransformer()
	
	tests := []struct {
		input    interface{}
		output   interface{}
		expected int
	}{
		{float64(10), float64(15), 25},
		{float64(0), float64(5), 5},
		{nil, float64(5), 5},
		{float64(5), nil, 5},
		{nil, nil, 0},
	}
	
	for _, test := range tests {
		result := transformer.sumTokens(test.input, test.output)
		if result != test.expected {
			t.Errorf("For input %v, output %v, expected %d, got %d", test.input, test.output, test.expected, result)
		}
	}
}
