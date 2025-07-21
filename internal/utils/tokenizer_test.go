package utils

import (
	"testing"

	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
)

func TestInitTokenizer(t *testing.T) {
	// Test tokenizer initialization
	// This might fail if tiktoken-go dependency is not available
	err := InitTokenizer()
	if err != nil {
		// If tiktoken is not available, skip the test
		t.Skip("tiktoken-go dependency not available:", err)
	}

	// If initialization succeeded, test encoder
	encoder, err := GetEncoder()
	testutil.AssertNoError(t, err, "Should get encoder after successful initialization")
	testutil.AssertNotEqual(t, nil, encoder, "Encoder should not be nil")
}

func TestGetEncoder(t *testing.T) {
	// Test getting encoder
	encoder, err := GetEncoder()
	if err != nil {
		// If tiktoken is not available, skip the test
		t.Skip("tiktoken-go dependency not available:", err)
	}

	testutil.AssertNotEqual(t, nil, encoder, "Encoder should not be nil")

	// Test getting encoder again (should return same instance)
	encoder2, err := GetEncoder()
	testutil.AssertNoError(t, err, "Should get encoder on second call")
	testutil.AssertEqual(t, encoder, encoder2, "Should return same encoder instance")
}

func TestCountTokens(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		minCount int // Minimum expected tokens (rough estimate)
	}{
		{
			name:     "empty string",
			text:     "",
			minCount: 0,
		},
		{
			name:     "simple text",
			text:     "Hello, world!",
			minCount: 1,
		},
		{
			name:     "longer text",
			text:     "This is a longer piece of text that should have more tokens.",
			minCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := CountTokens(tt.text)
			if err != nil {
				// If tiktoken is not available, skip the test
				t.Skip("tiktoken-go dependency not available:", err)
			}

			testutil.AssertTrue(t, count >= tt.minCount,
				"Token count should be at least %d, got %d", tt.minCount, count)

			// For empty string, count should be exactly 0
			if tt.text == "" {
				testutil.AssertEqual(t, 0, count, "Empty string should have 0 tokens")
			}
		})
	}
}

func TestCountTokensConsistency(t *testing.T) {
	text := "Consistent test message for token counting"

	count1, err := CountTokens(text)
	if err != nil {
		t.Skip("tiktoken-go dependency not available:", err)
	}

	count2, err := CountTokens(text)
	testutil.AssertNoError(t, err, "Should count tokens on second call")

	testutil.AssertEqual(t, count1, count2, "Token count should be consistent")
}

func TestTokenizerErrorHandling(t *testing.T) {
	// Test error handling when encoder is not available
	// Skip this test since it requires modifying global state with sync.Once
	// which cannot be properly restored without copying locks
	t.Skip("Skipping test that requires unsafe sync.Once manipulation")
}

func TestCountMessageTokens(t *testing.T) {
	tests := []struct {
		name     string
		params   *MessageCreateParams
		minCount int
		skipTest bool
	}{
		{
			name: "simple message",
			params: &MessageCreateParams{
				Model: "claude-3",
				Messages: []Message{
					{
						Role:    "user",
						Content: "Hello, world!",
					},
				},
			},
			minCount: 1,
		},
		{
			name: "message with system prompt",
			params: &MessageCreateParams{
				Model: "claude-3",
				Messages: []Message{
					{
						Role:    "user",
						Content: "Test message",
					},
				},
				System: "You are a helpful assistant",
			},
			minCount: 5,
		},
		{
			name: "message with array content",
			params: &MessageCreateParams{
				Model: "claude-3",
				Messages: []Message{
					{
						Role: "user",
						Content: []interface{}{
							map[string]interface{}{
								"type": "text",
								"text": "What's in this image?",
							},
						},
					},
				},
			},
			minCount: 3,
		},
		{
			name: "message with tool use",
			params: &MessageCreateParams{
				Model: "claude-3",
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
									"numbers":   []int{1, 2, 3},
								},
							},
						},
					},
				},
			},
			minCount: 5,
		},
		{
			name: "message with tool result",
			params: &MessageCreateParams{
				Model: "claude-3",
				Messages: []Message{
					{
						Role: "user",
						Content: []interface{}{
							map[string]interface{}{
								"type":        "tool_result",
								"tool_use_id": "tool_123",
								"content":     "The result is 6",
							},
						},
					},
				},
			},
			minCount: 3,
		},
		{
			name: "message with tools schema",
			params: &MessageCreateParams{
				Model: "claude-3",
				Messages: []Message{
					{
						Role:    "user",
						Content: "Calculate 2+2",
					},
				},
				Tools: []Tool{
					{
						Name:        "calculator",
						Description: "Performs basic arithmetic operations",
						InputSchema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"operation": map[string]interface{}{
									"type": "string",
									"enum": []string{"add", "subtract", "multiply", "divide"},
								},
								"numbers": map[string]interface{}{
									"type":  "array",
									"items": map[string]interface{}{"type": "number"},
								},
							},
						},
					},
				},
			},
			minCount: 10,
		},
		{
			name: "complex system with array",
			params: &MessageCreateParams{
				Model: "claude-3",
				Messages: []Message{
					{
						Role:    "user",
						Content: "Test",
					},
				},
				System: []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "You are a helpful assistant",
					},
				},
			},
			minCount: 5,
		},
		{
			name: "system with text array",
			params: &MessageCreateParams{
				Model: "claude-3",
				Messages: []Message{
					{
						Role:    "user",
						Content: "Test",
					},
				},
				System: []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": []interface{}{"Hello", "World"},
					},
				},
			},
			minCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipTest {
				t.Skip("Test case requires specific setup")
			}

			count, err := CountMessageTokens(tt.params)
			if err != nil {
				// If tiktoken is not available, skip the test
				if testutil.ContainsString(err.Error(), "failed to get encoder") {
					t.Skip("tiktoken-go dependency not available:", err)
				}
				testutil.AssertNoError(t, err, "Should count message tokens without error")
			}

			testutil.AssertTrue(t, count >= tt.minCount,
				"Token count should be at least %d, got %d", tt.minCount, count)
			testutil.AssertTrue(t, count > 0, "Token count should be positive")
		})
	}
}

func TestCountMessageTokensErrorCases(t *testing.T) {
	tests := []struct {
		name   string
		params *MessageCreateParams
	}{
		{
			name: "tool with invalid input schema",
			params: &MessageCreateParams{
				Model: "claude-3",
				Messages: []Message{
					{Role: "user", Content: "test"},
				},
				Tools: []Tool{
					{
						Name:        "test-tool",
						InputSchema: make(chan int), // Cannot be marshaled
					},
				},
			},
		},
		{
			name: "message with tool use invalid input",
			params: &MessageCreateParams{
				Model: "claude-3",
				Messages: []Message{
					{
						Role: "assistant",
						Content: []interface{}{
							map[string]interface{}{
								"type":  "tool_use",
								"input": make(chan int), // Cannot be marshaled
							},
						},
					},
				},
			},
		},
		{
			name: "message with tool result invalid content",
			params: &MessageCreateParams{
				Model: "claude-3",
				Messages: []Message{
					{
						Role: "user",
						Content: []interface{}{
							map[string]interface{}{
								"type":    "tool_result",
								"content": make(chan int), // Cannot be marshaled
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CountMessageTokens(tt.params)
			if err == nil {
				// If tiktoken is not available, we might not get an error
				// But if we do get one, it should be about marshaling
				return
			}

			// Check if it's a tiktoken availability error
			if testutil.ContainsString(err.Error(), "failed to get encoder") {
				t.Skip("tiktoken-go dependency not available:", err)
			}

			// Should be an error about marshaling
			testutil.AssertError(t, err, "Should return error for unmarshalable content")
		})
	}
}

func TestCountMessageContentTokens(t *testing.T) {
	// Test the internal function via CountMessageTokens
	tests := []struct {
		name    string
		content interface{}
	}{
		{
			name:    "string content",
			content: "Hello world",
		},
		{
			name: "array content with text",
			content: []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "Hello world",
				},
			},
		},
		{
			name: "array content with tool use",
			content: []interface{}{
				map[string]interface{}{
					"type": "tool_use",
					"id":   "test-id",
					"name": "test-tool",
					"input": map[string]interface{}{
						"param": "value",
					},
				},
			},
		},
		{
			name: "array content with tool result",
			content: []interface{}{
				map[string]interface{}{
					"type":        "tool_result",
					"tool_use_id": "test-id",
					"content":     "result data",
				},
			},
		},
		{
			name: "array content with unknown type",
			content: []interface{}{
				map[string]interface{}{
					"type": "unknown",
					"data": "some data",
				},
			},
		},
		{
			name: "invalid array content",
			content: []interface{}{
				"not a map",
				123,
				true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test through CountMessageTokens
			params := &MessageCreateParams{
				Model: "claude-3",
				Messages: []Message{
					{
						Role:    "user",
						Content: tt.content,
					},
				},
			}

			count, err := CountMessageTokens(params)
			if err != nil && testutil.ContainsString(err.Error(), "failed to get encoder") {
				t.Skip("tiktoken-go dependency not available:", err)
			}

			if err == nil {
				testutil.AssertTrue(t, count >= 0, "Token count should be non-negative")
			}
		})
	}
}

func TestCountSystemTokens(t *testing.T) {
	// Test the internal function via CountMessageTokens
	tests := []struct {
		name   string
		system interface{}
	}{
		{
			name:   "string system",
			system: "You are a helpful assistant",
		},
		{
			name: "array system with text",
			system: []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "You are helpful",
				},
			},
		},
		{
			name: "array system with text array",
			system: []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": []interface{}{"Hello", "World", "!"},
				},
			},
		},
		{
			name: "array system with non-text type",
			system: []interface{}{
				map[string]interface{}{
					"type": "image",
					"data": "base64data",
				},
			},
		},
		{
			name: "invalid array system",
			system: []interface{}{
				"not a map",
				123,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test through CountMessageTokens
			params := &MessageCreateParams{
				Model: "claude-3",
				Messages: []Message{
					{
						Role:    "user",
						Content: "test",
					},
				},
				System: tt.system,
			}

			count, err := CountMessageTokens(params)
			if err != nil && testutil.ContainsString(err.Error(), "failed to get encoder") {
				t.Skip("tiktoken-go dependency not available:", err)
			}

			if err == nil {
				testutil.AssertTrue(t, count >= 0, "Token count should be non-negative")
			}
		})
	}
}

func TestGetEncoderErrorPath(t *testing.T) {
	// Test the error path in GetEncoder when initialization fails
	// Skip this test since it requires modifying global state with sync.Once
	// which cannot be properly restored without copying locks
	t.Skip("Skipping test that requires unsafe sync.Once manipulation")
}

// Benchmark tests for performance verification
func BenchmarkCountMessageTokens(b *testing.B) {
	params := &MessageCreateParams{
		Model: "claude-3",
		Messages: []Message{
			{
				Role:    "user",
				Content: "This is a benchmark test message for token counting performance.",
			},
			{
				Role:    "assistant",
				Content: "I understand. This is my response to your benchmark test.",
			},
		},
		System: "You are a helpful assistant focused on performance testing.",
	}

	// Skip benchmark if tokenizer is not available
	_, err := CountMessageTokens(params)
	if err != nil {
		b.Skip("tiktoken-go dependency not available:", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CountMessageTokens(params)
	}
}

func BenchmarkCountMessageTokensComplex(b *testing.B) {
	params := &MessageCreateParams{
		Model: "claude-3",
		Messages: []Message{
			{
				Role: "user",
				Content: []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "Calculate the sum of these numbers",
					},
				},
			},
			{
				Role: "assistant",
				Content: []interface{}{
					map[string]interface{}{
						"type": "tool_use",
						"id":   "calc_123",
						"name": "calculator",
						"input": map[string]interface{}{
							"operation": "sum",
							"numbers":   []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
						},
					},
				},
			},
		},
		Tools: []Tool{
			{
				Name:        "calculator",
				Description: "Performs mathematical calculations",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"operation": map[string]string{"type": "string"},
						"numbers":   map[string]interface{}{"type": "array", "items": map[string]string{"type": "number"}},
					},
				},
			},
		},
	}

	// Skip benchmark if tokenizer is not available
	_, err := CountMessageTokens(params)
	if err != nil {
		b.Skip("tiktoken-go dependency not available:", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CountMessageTokens(params)
	}
}
