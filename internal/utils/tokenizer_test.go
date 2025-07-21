package utils

import (
	"testing"
)

func TestInitTokenizer(t *testing.T) {
	err := InitTokenizer()
	if err != nil {
		t.Fatalf("Failed to initialize tokenizer: %v", err)
	}
	
	// Test that encoder is initialized
	enc, err := GetEncoder()
	if err != nil {
		t.Fatalf("Failed to get encoder: %v", err)
	}
	if enc == nil {
		t.Fatal("Encoder is nil after initialization")
	}
}

func TestCountTokens(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		minCount int // Minimum expected tokens (approximate)
	}{
		{
			name:     "simple text",
			text:     "Hello, world!",
			minCount: 2,
		},
		{
			name:     "longer text",
			text:     "This is a longer text with multiple words and sentences. It should have more tokens.",
			minCount: 10,
		},
		{
			name:     "empty string",
			text:     "",
			minCount: 0,
		},
		{
			name:     "special characters",
			text:     "Special chars: !@#$%^&*()",
			minCount: 5,
		},
		{
			name:     "unicode text",
			text:     "Unicode: 你好世界 🌍",
			minCount: 3,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := CountTokens(tt.text)
			if err != nil {
				t.Fatalf("Failed to count tokens: %v", err)
			}
			if count < tt.minCount {
				t.Errorf("Expected at least %d tokens, got %d", tt.minCount, count)
			}
		})
	}
}

func TestCountMessageTokens(t *testing.T) {
	tests := []struct {
		name     string
		params   *MessageCreateParams
		minCount int
	}{
		{
			name: "simple message",
			params: &MessageCreateParams{
				Model: "test-model",
				Messages: []Message{
					{Role: "user", Content: "Hello, how are you?"},
				},
			},
			minCount: 3,
		},
		{
			name: "message with system prompt",
			params: &MessageCreateParams{
				Model: "test-model",
				Messages: []Message{
					{Role: "user", Content: "What's the weather?"},
				},
				System: "You are a helpful weather assistant.",
			},
			minCount: 10,
		},
		{
			name: "message with array content",
			params: &MessageCreateParams{
				Model: "test-model",
				Messages: []Message{
					{
						Role: "user",
						Content: []interface{}{
							map[string]interface{}{
								"type": "text",
								"text": "Here is some text content",
							},
						},
					},
				},
			},
			minCount: 5,
		},
		{
			name: "message with tool use",
			params: &MessageCreateParams{
				Model: "test-model",
				Messages: []Message{
					{
						Role: "assistant",
						Content: []interface{}{
							map[string]interface{}{
								"type": "tool_use",
								"id":   "tool_123",
								"name": "calculator",
								"input": map[string]interface{}{
									"operation": "add",
									"a":         5,
									"b":         3,
								},
							},
						},
					},
				},
			},
			minCount: 10,
		},
		{
			name: "message with tool result",
			params: &MessageCreateParams{
				Model: "test-model",
				Messages: []Message{
					{
						Role: "user",
						Content: []interface{}{
							map[string]interface{}{
								"type":        "tool_result",
								"tool_use_id": "tool_123",
								"content":     "Result: 8",
							},
						},
					},
				},
			},
			minCount: 2,
		},
		{
			name: "message with tools definition",
			params: &MessageCreateParams{
				Model: "test-model",
				Messages: []Message{
					{Role: "user", Content: "Calculate 2+2"},
				},
				Tools: []Tool{
					{
						Name:        "calculator",
						Description: "A simple calculator that can add, subtract, multiply, and divide",
						InputSchema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"operation": map[string]interface{}{
									"type": "string",
									"enum": []string{"add", "subtract", "multiply", "divide"},
								},
								"a": map[string]interface{}{"type": "number"},
								"b": map[string]interface{}{"type": "number"},
							},
							"required": []string{"operation", "a", "b"},
						},
					},
				},
			},
			minCount: 50,
		},
		{
			name: "complex system prompt array",
			params: &MessageCreateParams{
				Model: "test-model",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
				System: []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "You are an AI assistant.",
					},
					map[string]interface{}{
						"type": "text",
						"text": []interface{}{
							"You should be helpful",
							"and friendly",
						},
					},
				},
			},
			minCount: 10,
		},
		{
			name: "tool result with object content",
			params: &MessageCreateParams{
				Model: "test-model",
				Messages: []Message{
					{
						Role: "user",
						Content: []interface{}{
							map[string]interface{}{
								"type":        "tool_result",
								"tool_use_id": "tool_456",
								"content": map[string]interface{}{
									"status": "success",
									"data": map[string]interface{}{
										"temperature": 72,
										"humidity":    45,
									},
								},
							},
						},
					},
				},
			},
			minCount: 10,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := CountMessageTokens(tt.params)
			if err != nil {
				t.Fatalf("Failed to count message tokens: %v", err)
			}
			if count < tt.minCount {
				t.Errorf("Expected at least %d tokens, got %d", tt.minCount, count)
			}
		})
	}
}

func TestCountMessageTokens_Errors(t *testing.T) {
	// Test with invalid content that should handle errors gracefully
	params := &MessageCreateParams{
		Model: "test-model",
		Messages: []Message{
			{
				Role: "user",
				Content: []interface{}{
					"invalid content type", // This should be skipped
					map[string]interface{}{
						"type": "text",
						"text": "Valid text",
					},
				},
			},
		},
	}
	
	count, err := CountMessageTokens(params)
	if err != nil {
		t.Fatalf("Should handle invalid content gracefully: %v", err)
	}
	if count == 0 {
		t.Error("Should have counted tokens from valid content")
	}
}

func BenchmarkCountTokens(b *testing.B) {
	text := "This is a sample text for benchmarking token counting performance. It contains multiple sentences and should give us a good idea of how fast the tokenizer performs."
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := CountTokens(text)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCountMessageTokens(b *testing.B) {
	params := &MessageCreateParams{
		Model: "test-model",
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "What is the meaning of life?"},
			{Role: "assistant", Content: "The meaning of life is a philosophical question that has been pondered throughout human history."},
			{Role: "user", Content: "Can you elaborate?"},
		},
		Tools: []Tool{
			{
				Name:        "search",
				Description: "Search for information",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{"type": "string"},
					},
				},
			},
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := CountMessageTokens(params)
		if err != nil {
			b.Fatal(err)
		}
	}
}