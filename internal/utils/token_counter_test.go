package utils

import (
	"testing"
)

func TestCountRequestTokens(t *testing.T) {
	tests := []struct {
		name     string
		bodyMap  map[string]interface{}
		minCount int
	}{
		{
			name:     "empty request",
			bodyMap:  map[string]interface{}{},
			minCount: 50, // Base tokens
		},
		{
			name: "simple message",
			bodyMap: map[string]interface{}{
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello, how are you?", // 19 chars = ~4 tokens
					},
				},
			},
			minCount: 54,
		},
		{
			name: "multiple messages",
			bodyMap: map[string]interface{}{
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello, how are you?", // 19 chars = ~4 tokens
					},
					map[string]interface{}{
						"role":    "assistant",
						"content": "I'm doing great, thank you!", // 27 chars = ~6 tokens
					},
					map[string]interface{}{
						"role":    "user",
						"content": "What's the weather like?", // 24 chars = ~6 tokens
					},
				},
			},
			minCount: 66, // 50 + 4 + 6 + 6
		},
		{
			name: "with system message",
			bodyMap: map[string]interface{}{
				"system": "You are a helpful assistant that provides weather information.", // 63 chars = ~15 tokens
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "What's the weather?", // 19 chars = ~4 tokens
					},
				},
			},
			minCount: 69, // 50 + 15 + 4
		},
		{
			name: "messages with non-string content",
			bodyMap: map[string]interface{}{
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": 123, // Non-string content
					},
					"invalid message", // Invalid message format
				},
			},
			minCount: 50, // Only base tokens
		},
		{
			name: "empty messages array",
			bodyMap: map[string]interface{}{
				"messages": []interface{}{},
			},
			minCount: 50,
		},
		{
			name: "non-array messages",
			bodyMap: map[string]interface{}{
				"messages": "not an array",
			},
			minCount: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := CountRequestTokens(tt.bodyMap)
			if count < tt.minCount {
				t.Errorf("expected at least %d tokens, got %d", tt.minCount, count)
			}
		})
	}
}

func TestCountResponseTokens(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "empty content",
			content:  "",
			expected: 0,
		},
		{
			name:     "simple text",
			content:  "Hello world", // 11 chars = ~2 tokens
			expected: 2,
		},
		{
			name:     "longer text",
			content:  "The weather today is sunny with a high of 75 degrees.", // 54 chars = ~13 tokens
			expected: 13,
		},
		{
			name:     "text with whitespace",
			content:  "  \n\tHello world\n\t  ", // Will be trimmed to 11 chars
			expected: 2,
		},
		{
			name:     "unicode text",
			content:  "Hello 世界 👋", // Mixed content - 17 bytes
			expected: 4, // Approximation based on byte length / 4
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := CountResponseTokens(tt.content)
			if count != tt.expected {
				t.Errorf("expected %d tokens, got %d", tt.expected, count)
			}
		})
	}
}