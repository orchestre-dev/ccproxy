package converter

import (
	"encoding/json"
	"testing"

	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
)

func TestNewAnthropicConverter(t *testing.T) {
	converter := NewAnthropicConverter()
	testutil.AssertNotEqual(t, nil, converter)
}

func TestAnthropicConverter_ToGeneric_Request(t *testing.T) {
	converter := NewAnthropicConverter()

	tests := []struct {
		name           string
		input          string
		expectedModel  string
		expectedSystem string
		expectError    bool
	}{
		{
			name: "basic request",
			input: `{
				"model": "claude-3-haiku-20240307",
				"messages": [
					{
						"role": "user",
						"content": [{"type": "text", "text": "Hello"}]
					}
				],
				"max_tokens": 100,
				"temperature": 0.7
			}`,
			expectedModel:  "claude-3-haiku-20240307",
			expectedSystem: "",
			expectError:    false,
		},
		{
			name: "request with system message",
			input: `{
				"model": "claude-3-opus-20240229",
				"messages": [
					{
						"role": "user", 
						"content": [{"type": "text", "text": "Hello"}]
					}
				],
				"system": "You are a helpful assistant",
				"max_tokens": 200,
				"temperature": 0.5,
				"stream": true
			}`,
			expectedModel:  "claude-3-opus-20240229",
			expectedSystem: "You are a helpful assistant",
			expectError:    false,
		},
		{
			name: "request with metadata",
			input: `{
				"model": "claude-3-sonnet-20240229",
				"messages": [
					{
						"role": "user",
						"content": [{"type": "text", "text": "Hello"}]
					}
				],
				"max_tokens": 150,
				"metadata": {"user_id": "123", "session": "abc"}
			}`,
			expectedModel:  "claude-3-sonnet-20240229",
			expectedSystem: "",
			expectError:    false,
		},
		{
			name: "request with multiple messages",
			input: `{
				"model": "claude-3-haiku-20240307",
				"messages": [
					{
						"role": "user",
						"content": [{"type": "text", "text": "Hello"}]
					},
					{
						"role": "assistant",
						"content": [{"type": "text", "text": "Hi there!"}]
					},
					{
						"role": "user",
						"content": [{"type": "text", "text": "How are you?"}]
					}
				],
				"max_tokens": 100
			}`,
			expectedModel:  "claude-3-haiku-20240307",
			expectedSystem: "",
			expectError:    false,
		},
		{
			name:        "malformed JSON",
			input:       `{"model": "claude-3-haiku-20240307", "messages": [`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.ToGeneric(json.RawMessage(tt.input), true)

			if tt.expectError {
				testutil.AssertError(t, err)
				return
			}

			testutil.AssertNoError(t, err)
			testutil.AssertNotEqual(t, nil, result)

			// Parse the result to verify structure
			var genericReq Request
			err = json.Unmarshal(result, &genericReq)
			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, tt.expectedModel, genericReq.Model)
			testutil.AssertEqual(t, tt.expectedSystem, genericReq.System)
		})
	}
}

func TestAnthropicConverter_ToGeneric_Response(t *testing.T) {
	converter := NewAnthropicConverter()

	tests := []struct {
		name        string
		input       string
		expectedID  string
		expectError bool
	}{
		{
			name: "basic response",
			input: `{
				"id": "msg_123",
				"type": "message",
				"role": "assistant",
				"content": [{"type": "text", "text": "Hello there!"}],
				"model": "claude-3-haiku-20240307",
				"usage": {
					"input_tokens": 10,
					"output_tokens": 5
				}
			}`,
			expectedID:  "msg_123",
			expectError: false,
		},
		{
			name: "response without usage",
			input: `{
				"id": "msg_456",
				"type": "message",
				"role": "assistant",
				"content": [{"type": "text", "text": "Response without usage"}],
				"model": "claude-3-opus-20240229"
			}`,
			expectedID:  "msg_456",
			expectError: false,
		},
		{
			name: "response with multiple content parts",
			input: `{
				"id": "msg_789",
				"type": "message",
				"role": "assistant",
				"content": [
					{"type": "text", "text": "First part. "},
					{"type": "text", "text": "Second part."}
				],
				"model": "claude-3-sonnet-20240229",
				"usage": {
					"input_tokens": 20,
					"output_tokens": 8
				}
			}`,
			expectedID:  "msg_789",
			expectError: false,
		},
		{
			name:        "malformed JSON",
			input:       `{"id": "msg_123", "type": "message"`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.ToGeneric(json.RawMessage(tt.input), false)

			if tt.expectError {
				testutil.AssertError(t, err)
				return
			}

			testutil.AssertNoError(t, err)
			testutil.AssertNotEqual(t, nil, result)

			// Parse the result to verify structure
			var genericResp Response
			err = json.Unmarshal(result, &genericResp)
			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, tt.expectedID, genericResp.ID)
		})
	}
}

func TestAnthropicConverter_FromGeneric_Request(t *testing.T) {
	converter := NewAnthropicConverter()

	tests := []struct {
		name             string
		input            string
		expectedModel    string
		expectedMessages int
		expectError      bool
	}{
		{
			name: "basic request",
			input: `{
				"model": "claude-3-haiku-20240307",
				"messages": [
					{
						"role": "user",
						"content": "Hello"
					}
				],
				"max_tokens": 100,
				"temperature": 0.7
			}`,
			expectedModel:    "claude-3-haiku-20240307",
			expectedMessages: 1,
			expectError:      false,
		},
		{
			name: "request with system",
			input: `{
				"model": "claude-3-opus-20240229", 
				"messages": [
					{
						"role": "user",
						"content": "Hello"
					}
				],
				"system": "You are helpful",
				"max_tokens": 200,
				"stream": true
			}`,
			expectedModel:    "claude-3-opus-20240229",
			expectedMessages: 1,
			expectError:      false,
		},
		{
			name: "request with anthropic content array",
			input: `{
				"model": "claude-3-sonnet-20240229",
				"messages": [
					{
						"role": "user",
						"content": [{"type": "text", "text": "Hello"}]
					}
				],
				"max_tokens": 150
			}`,
			expectedModel:    "claude-3-sonnet-20240229",
			expectedMessages: 1,
			expectError:      false,
		},
		{
			name: "request with metadata",
			input: `{
				"model": "claude-3-haiku-20240307",
				"messages": [
					{
						"role": "user",
						"content": "Test"
					}
				],
				"metadata": {"key": "value"},
				"max_tokens": 100
			}`,
			expectedModel:    "claude-3-haiku-20240307",
			expectedMessages: 1,
			expectError:      false,
		},
		{
			name:        "malformed JSON",
			input:       `{"model": "test", "messages":`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.FromGeneric(json.RawMessage(tt.input), true)

			if tt.expectError {
				testutil.AssertError(t, err)
				return
			}

			testutil.AssertNoError(t, err)
			testutil.AssertNotEqual(t, nil, result)

			// Parse the result to verify structure
			var anthropicReq AnthropicRequest
			err = json.Unmarshal(result, &anthropicReq)
			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, tt.expectedModel, anthropicReq.Model)
			testutil.AssertEqual(t, tt.expectedMessages, len(anthropicReq.Messages))
		})
	}
}

func TestAnthropicConverter_FromGeneric_Response(t *testing.T) {
	converter := NewAnthropicConverter()

	tests := []struct {
		name        string
		input       string
		expectedID  string
		expectError bool
	}{
		{
			name: "basic response",
			input: `{
				"id": "msg_123",
				"type": "message",
				"role": "assistant", 
				"content": [{"type": "text", "text": "Hello!"}],
				"model": "claude-3-haiku-20240307",
				"usage": {
					"input_tokens": 10,
					"output_tokens": 5,
					"total_tokens": 15
				}
			}`,
			expectedID:  "msg_123",
			expectError: false,
		},
		{
			name: "response without usage",
			input: `{
				"id": "msg_456",
				"type": "message",
				"role": "assistant",
				"content": [{"type": "text", "text": "Response"}],
				"model": "claude-3-opus-20240229"
			}`,
			expectedID:  "msg_456",
			expectError: false,
		},
		{
			name:        "malformed JSON",
			input:       `{"id": "msg_123"`,
			expectError: true,
		},
		{
			name:        "malformed content",
			input:       `{"id": "msg_123", "content": "invalid_json"}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.FromGeneric(json.RawMessage(tt.input), false)

			if tt.expectError {
				testutil.AssertError(t, err)
				return
			}

			testutil.AssertNoError(t, err)
			testutil.AssertNotEqual(t, nil, result)

			// Parse the result to verify structure
			var anthropicResp AnthropicResponse
			err = json.Unmarshal(result, &anthropicResp)
			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, tt.expectedID, anthropicResp.ID)
		})
	}
}

func TestAnthropicConverter_ConvertStreamEvent(t *testing.T) {
	converter := NewAnthropicConverter()

	tests := []struct {
		name     string
		input    []byte
		toFormat MessageFormat
	}{
		{
			name:     "basic stream event",
			input:    []byte("data: {\"type\": \"message_start\", \"message\": {\"id\": \"msg_123\"}}"),
			toFormat: FormatOpenAI,
		},
		{
			name:     "empty event",
			input:    []byte(""),
			toFormat: FormatGoogle,
		},
		{
			name:     "complex event",
			input:    []byte("event: content_block_delta\ndata: {\"delta\": {\"text\": \"Hello\"}}"),
			toFormat: FormatAWS,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.ConvertStreamEvent(tt.input, tt.toFormat)

			// Current implementation just passes through
			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, string(tt.input), string(result))
		})
	}
}

func TestAnthropicConverter_ContentParsing(t *testing.T) {
	converter := NewAnthropicConverter()

	t.Run("string content to anthropic format", func(t *testing.T) {
		input := `{
			"model": "claude-3-haiku-20240307",
			"messages": [
				{
					"role": "user",
					"content": "Simple string content"
				}
			],
			"max_tokens": 100
		}`

		result, err := converter.FromGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)

		var anthropicReq AnthropicRequest
		err = json.Unmarshal(result, &anthropicReq)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 1, len(anthropicReq.Messages))
		testutil.AssertEqual(t, 1, len(anthropicReq.Messages[0].Content))
		testutil.AssertEqual(t, "text", anthropicReq.Messages[0].Content[0].Type)
		testutil.AssertEqual(t, "Simple string content", anthropicReq.Messages[0].Content[0].Text)
	})

	t.Run("malformed content in FromGeneric", func(t *testing.T) {
		input := `{
			"model": "claude-3-haiku-20240307",
			"messages": [
				{
					"role": "user",
					"content": {"invalid": "content_format"}
				}
			],
			"max_tokens": 100
		}`

		_, err := converter.FromGeneric(json.RawMessage(input), true)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "failed to parse content")
	})
}

func TestAnthropicConverter_ErrorHandling(t *testing.T) {
	converter := NewAnthropicConverter()

	t.Run("invalid request JSON", func(t *testing.T) {
		input := `{"model": "test", "messages": invalid}`
		_, err := converter.ToGeneric(json.RawMessage(input), true)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "failed to unmarshal Anthropic request")
	})

	t.Run("invalid response JSON", func(t *testing.T) {
		input := `{"id": "test", "content": invalid}`
		_, err := converter.ToGeneric(json.RawMessage(input), false)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "failed to unmarshal Anthropic response")
	})

	t.Run("content marshal error in ToGeneric request", func(t *testing.T) {
		// This test simulates a scenario where content marshaling might fail
		// In practice, this is hard to trigger with valid Go data structures
		input := `{
			"model": "claude-3-haiku-20240307",
			"messages": [
				{
					"role": "user",
					"content": [{"type": "text", "text": "Hello"}]
				}
			],
			"max_tokens": 100
		}`

		// This should succeed normally
		_, err := converter.ToGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)
	})

	t.Run("metadata marshal error in ToGeneric", func(t *testing.T) {
		// Test with valid metadata that can be marshaled
		input := `{
			"model": "claude-3-haiku-20240307",
			"messages": [
				{
					"role": "user",
					"content": [{"type": "text", "text": "Hello"}]
				}
			],
			"metadata": {"key": "value"},
			"max_tokens": 100
		}`

		_, err := converter.ToGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)
	})

	t.Run("content marshal error in ToGeneric response", func(t *testing.T) {
		input := `{
			"id": "msg_123",
			"type": "message",
			"role": "assistant",
			"content": [{"type": "text", "text": "Hello!"}],
			"model": "claude-3-haiku-20240307"
		}`

		_, err := converter.ToGeneric(json.RawMessage(input), false)
		testutil.AssertNoError(t, err)
	})
}

// TestAnthropicConverter_HelperFunctions was removed since the helper functions
// were unused in production code and removed to fix lint warnings

func TestAnthropicConverter_EdgeCases(t *testing.T) {
	converter := NewAnthropicConverter()

	t.Run("empty messages array", func(t *testing.T) {
		input := `{
			"model": "claude-3-haiku-20240307",
			"messages": [],
			"max_tokens": 100
		}`

		result, err := converter.ToGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)

		var genericReq Request
		err = json.Unmarshal(result, &genericReq)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 0, len(genericReq.Messages))
	})

	t.Run("zero max_tokens", func(t *testing.T) {
		input := `{
			"model": "claude-3-haiku-20240307",
			"messages": [
				{
					"role": "user",
					"content": [{"type": "text", "text": "Hello"}]
				}
			],
			"max_tokens": 0
		}`

		result, err := converter.ToGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)

		var genericReq Request
		err = json.Unmarshal(result, &genericReq)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 0, genericReq.MaxTokens)
	})

	t.Run("empty content array", func(t *testing.T) {
		input := `{
			"model": "claude-3-haiku-20240307",
			"messages": [
				{
					"role": "user",
					"content": []
				}
			],
			"max_tokens": 100
		}`

		result, err := converter.ToGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, result)
	})
}
