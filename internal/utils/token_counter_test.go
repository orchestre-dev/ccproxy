package utils

import (
	"testing"

	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
)

func TestCountRequestTokens(t *testing.T) {
	tests := []struct {
		name     string
		bodyMap  map[string]interface{}
		minCount int // Minimum expected count
		maxCount int // Maximum expected count (approximate since it's an estimation)
	}{
		{
			name:     "empty request",
			bodyMap:  map[string]interface{}{},
			minCount: 50,  // Base tokens
			maxCount: 50,
		},
		{
			name: "simple message",
			bodyMap: map[string]interface{}{
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello, how are you?", // 19 chars ~= 4-5 tokens + base
					},
				},
			},
			minCount: 50, // Base + ~4 tokens
			maxCount: 60,
		},
		{
			name: "multiple messages",
			bodyMap: map[string]interface{}{
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello", // 5 chars ~= 1 token
					},
					map[string]interface{}{
						"role":    "assistant",
						"content": "Hi there!", // 9 chars ~= 2 tokens
					},
				},
			},
			minCount: 50, // Base + ~3 tokens
			maxCount: 60,
		},
		{
			name: "with system message",
			bodyMap: map[string]interface{}{
				"system": "You are a helpful assistant", // 28 chars ~= 7 tokens
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello", // 5 chars ~= 1 token
					},
				},
			},
			minCount: 50, // Base + ~8 tokens
			maxCount: 70,
		},
		{
			name: "long message",
			bodyMap: map[string]interface{}{
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "This is a very long message that contains many words and should result in a higher token count estimation based on the character length calculation", // ~160 chars = 40 tokens
					},
				},
			},
			minCount: 80, // Base + ~40 tokens
			maxCount: 100,
		},
		{
			name: "non-string content",
			bodyMap: map[string]interface{}{
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": 12345, // Not a string, should be ignored
					},
				},
			},
			minCount: 50, // Only base tokens
			maxCount: 50,
		},
		{
			name: "malformed messages",
			bodyMap: map[string]interface{}{
				"messages": "not an array", // Invalid format
			},
			minCount: 50, // Only base tokens
			maxCount: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := CountRequestTokens(tt.bodyMap)
			testutil.AssertTrue(t, count >= tt.minCount, 
				"Token count should be at least %d, got %d", tt.minCount, count)
			testutil.AssertTrue(t, count <= tt.maxCount, 
				"Token count should be at most %d, got %d", tt.maxCount, count)
		})
	}
}

func TestCountResponseTokens(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		minCount int
		maxCount int
	}{
		{
			name:     "empty response",
			content:  "",
			minCount: 0,
			maxCount: 0,
		},
		{
			name:     "whitespace only",
			content:  "   \n\t  ",
			minCount: 0,
			maxCount: 0,
		},
		{
			name:     "simple response",
			content:  "Hello",
			minCount: 1,
			maxCount: 2,
		},
		{
			name:     "medium response",
			content:  "This is a medium length response", // 33 chars ~= 8 tokens
			minCount: 8,
			maxCount: 9,
		},
		{
			name:     "long response",
			content:  "This is a very long response that contains many words and should be counted accurately based on the character length estimation algorithm", // ~140 chars = 35 tokens
			minCount: 34,
			maxCount: 36,
		},
		{
			name:     "response with formatting",
			content:  "  \n\nHello, world!\n\n  ", // Should be trimmed to "Hello, world!" = 13 chars ~= 3 tokens
			minCount: 3,
			maxCount: 4,
		},
		{
			name:     "response with unicode",
			content:  "Hello 🌍 世界", // Unicode characters, 11 chars including emoji
			minCount: 2,
			maxCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := CountResponseTokens(tt.content)
			testutil.AssertTrue(t, count >= tt.minCount, 
				"Token count should be at least %d, got %d", tt.minCount, count)
			testutil.AssertTrue(t, count <= tt.maxCount, 
				"Token count should be at most %d, got %d", tt.maxCount, count)
		})
	}
}

func TestTokenCountingConsistency(t *testing.T) {
	// Test that the same input always produces the same output
	bodyMap := map[string]interface{}{
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": "Consistent test message",
			},
		},
	}

	count1 := CountRequestTokens(bodyMap)
	count2 := CountRequestTokens(bodyMap)
	count3 := CountRequestTokens(bodyMap)

	testutil.AssertEqual(t, count1, count2, "Token count should be consistent")
	testutil.AssertEqual(t, count2, count3, "Token count should be consistent")

	// Test response counting consistency
	content := "Consistent response content"
	respCount1 := CountResponseTokens(content)
	respCount2 := CountResponseTokens(content)

	testutil.AssertEqual(t, respCount1, respCount2, "Response token count should be consistent")
}

func TestTokenCountingEdgeCases(t *testing.T) {
	t.Run("deeply nested messages", func(t *testing.T) {
		bodyMap := map[string]interface{}{
			"messages": []interface{}{
				map[string]interface{}{
					"role": "user",
					"content": map[string]interface{}{ // Content is not a string
						"nested": "value",
					},
				},
			},
		}

		count := CountRequestTokens(bodyMap)
		testutil.AssertEqual(t, 50, count, "Should return base count for non-string content")
	})

	t.Run("nil messages array", func(t *testing.T) {
		bodyMap := map[string]interface{}{
			"messages": nil,
		}

		count := CountRequestTokens(bodyMap)
		testutil.AssertEqual(t, 50, count, "Should handle nil messages array")
	})

	t.Run("very long content", func(t *testing.T) {
		// Create a very long string
		longContent := ""
		for i := 0; i < 10000; i++ {
			longContent += "word "
		}

		count := CountResponseTokens(longContent)
		expectedCount := len(longContent) / 4
		// Allow for small rounding differences
		testutil.AssertTrue(t, count >= expectedCount-1 && count <= expectedCount+1, 
			"Should handle very long content, expected ~%d, got %d", expectedCount, count)
	})

	t.Run("system message types", func(t *testing.T) {
		// Test with non-string system message
		bodyMap := map[string]interface{}{
			"system": 12345, // Not a string
		}

		count := CountRequestTokens(bodyMap)
		testutil.AssertEqual(t, 50, count, "Should ignore non-string system message")
	})
}

// Benchmark tests for performance verification
func BenchmarkCountRequestTokens(b *testing.B) {
	bodyMap := map[string]interface{}{
		"system": "You are a helpful assistant",
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": "Hello, how are you doing today? I have a question about programming.",
			},
			map[string]interface{}{
				"role":    "assistant",
				"content": "I'm doing well, thank you! I'd be happy to help with your programming question.",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CountRequestTokens(bodyMap)
	}
}

func BenchmarkCountResponseTokens(b *testing.B) {
	content := "This is a sample response content that might be returned from an AI model. It contains multiple sentences and should be representative of typical response lengths in real usage scenarios."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CountResponseTokens(content)
	}
}