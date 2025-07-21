package converter

import (
	"encoding/json"
	"testing"

	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
)

func TestNewOpenAIConverter(t *testing.T) {
	converter := NewOpenAIConverter()
	testutil.AssertNotEqual(t, nil, converter)
}

func TestOpenAIConverter_ToGeneric_Request(t *testing.T) {
	converter := NewOpenAIConverter()

	tests := []struct {
		name                string
		input               string
		expectedModel       string
		expectedSystem      string
		expectedMessageCount int
		expectError         bool
	}{
		{
			name: "basic request",
			input: `{
				"model": "gpt-4",
				"messages": [
					{
						"role": "user",
						"content": "Hello"
					}
				],
				"max_tokens": 100,
				"temperature": 0.7
			}`,
			expectedModel:       "gpt-4",
			expectedSystem:      "",
			expectedMessageCount: 1,
			expectError:         false,
		},
		{
			name: "request with system message",
			input: `{
				"model": "gpt-3.5-turbo",
				"messages": [
					{
						"role": "system",
						"content": "You are a helpful assistant"
					},
					{
						"role": "user",
						"content": "Hello"
					}
				],
				"max_tokens": 200,
				"temperature": 0.5,
				"stream": true
			}`,
			expectedModel:       "gpt-3.5-turbo",
			expectedSystem:      "You are a helpful assistant",
			expectedMessageCount: 1, // System message extracted
			expectError:         false,
		},
		{
			name: "request with multiple system messages",
			input: `{
				"model": "gpt-4-turbo",
				"messages": [
					{
						"role": "system",
						"content": "First system message"
					},
					{
						"role": "system", 
						"content": "Second system message"
					},
					{
						"role": "user",
						"content": "Hello"
					}
				],
				"max_tokens": 150
			}`,
			expectedModel:       "gpt-4-turbo",
			expectedSystem:      "Second system message", // Last system message wins
			expectedMessageCount: 1,
			expectError:         false,
		},
		{
			name: "request with message name",
			input: `{
				"model": "gpt-4",
				"messages": [
					{
						"role": "user",
						"content": "Hello",
						"name": "john"
					}
				],
				"max_tokens": 100
			}`,
			expectedModel:       "gpt-4",
			expectedSystem:      "",
			expectedMessageCount: 1,
			expectError:         false,
		},
		{
			name: "request with conversation",
			input: `{
				"model": "gpt-3.5-turbo",
				"messages": [
					{
						"role": "user",
						"content": "Hello"
					},
					{
						"role": "assistant",
						"content": "Hi there!"
					},
					{
						"role": "user",
						"content": "How are you?"
					}
				],
				"max_tokens": 100
			}`,
			expectedModel:       "gpt-3.5-turbo",
			expectedSystem:      "",
			expectedMessageCount: 3,
			expectError:         false,
		},
		{
			name: "request with additional parameters",
			input: `{
				"model": "gpt-4",
				"messages": [
					{
						"role": "user",
						"content": "Hello"
					}
				],
				"max_tokens": 100,
				"temperature": 0.8,
				"n": 1,
				"stop": ["###", "END"]
			}`,
			expectedModel:       "gpt-4",
			expectedSystem:      "",
			expectedMessageCount: 1,
			expectError:         false,
		},
		{
			name:        "malformed JSON",
			input:       `{"model": "gpt-4", "messages": [`,
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
			testutil.AssertEqual(t, tt.expectedMessageCount, len(genericReq.Messages))
		})
	}
}

func TestOpenAIConverter_ToGeneric_Response(t *testing.T) {
	converter := NewOpenAIConverter()

	tests := []struct {
		name        string
		input       string
		expectedID  string
		expectError bool
	}{
		{
			name: "basic response",
			input: `{
				"id": "chatcmpl-123",
				"object": "chat.completion",
				"created": 1677652288,
				"model": "gpt-3.5-turbo-0613",
				"choices": [
					{
						"index": 0,
						"message": {
							"role": "assistant",
							"content": "Hello there!"
						},
						"finish_reason": "stop"
					}
				],
				"usage": {
					"prompt_tokens": 10,
					"completion_tokens": 5,
					"total_tokens": 15
				}
			}`,
			expectedID:  "chatcmpl-123",
			expectError: false,
		},
		{
			name: "response without usage",
			input: `{
				"id": "chatcmpl-456",
				"object": "chat.completion",
				"created": 1677652288,
				"model": "gpt-4",
				"choices": [
					{
						"index": 0,
						"message": {
							"role": "assistant",
							"content": "Response without usage"
						},
						"finish_reason": "stop"
					}
				]
			}`,
			expectedID:  "chatcmpl-456",
			expectError: false,
		},
		{
			name: "response with multiple choices",
			input: `{
				"id": "chatcmpl-789",
				"object": "chat.completion",
				"created": 1677652288,
				"model": "gpt-4",
				"choices": [
					{
						"index": 0,
						"message": {
							"role": "assistant",
							"content": "First choice"
						},
						"finish_reason": "stop"
					},
					{
						"index": 1,
						"message": {
							"role": "assistant",
							"content": "Second choice"
						},
						"finish_reason": "stop"
					}
				],
				"usage": {
					"prompt_tokens": 15,
					"completion_tokens": 8,
					"total_tokens": 23
				}
			}`,
			expectedID:  "chatcmpl-789",
			expectError: false, // Should use first choice
		},
		{
			name: "response with no choices",
			input: `{
				"id": "chatcmpl-error",
				"object": "chat.completion",
				"created": 1677652288,
				"model": "gpt-4",
				"choices": []
			}`,
			expectError: true,
		},
		{
			name:        "malformed JSON",
			input:       `{"id": "chatcmpl-123", "object": "chat.completion"`,
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

func TestOpenAIConverter_FromGeneric_Request(t *testing.T) {
	converter := NewOpenAIConverter()

	tests := []struct {
		name                string
		input               string
		expectedModel       string
		expectedMessageCount int
		expectSystemMessage bool
		expectError         bool
	}{
		{
			name: "basic request",
			input: `{
				"model": "gpt-4",
				"messages": [
					{
						"role": "user",
						"content": "Hello"
					}
				],
				"max_tokens": 100,
				"temperature": 0.7
			}`,
			expectedModel:       "gpt-4",
			expectedMessageCount: 1,
			expectSystemMessage: false,
			expectError:         false,
		},
		{
			name: "request with system",
			input: `{
				"model": "gpt-3.5-turbo", 
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
			expectedModel:       "gpt-3.5-turbo",
			expectedMessageCount: 2, // System + user message
			expectSystemMessage: true,
			expectError:         false,
		},
		{
			name: "request with anthropic content array",
			input: `{
				"model": "gpt-4",
				"messages": [
					{
						"role": "user",
						"content": [{"type": "text", "text": "Hello world"}]
					}
				],
				"max_tokens": 150
			}`,
			expectedModel:       "gpt-4",
			expectedMessageCount: 1,
			expectSystemMessage: false,
			expectError:         false,
		},
		{
			name: "request with message name",
			input: `{
				"model": "gpt-4",
				"messages": [
					{
						"role": "user",
						"content": "Test",
						"name": "john"
					}
				],
				"max_tokens": 100
			}`,
			expectedModel:       "gpt-4",
			expectedMessageCount: 1,
			expectSystemMessage: false,
			expectError:         false,
		},
		{
			name: "request with complex content",
			input: `{
				"model": "gpt-4",
				"messages": [
					{
						"role": "user",
						"content": [
							{"type": "text", "text": "First part "},
							{"type": "text", "text": "Second part"}
						]
					}
				],
				"max_tokens": 100
			}`,
			expectedModel:       "gpt-4",
			expectedMessageCount: 1,
			expectSystemMessage: false,
			expectError:         false,
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
			var openAIReq OpenAIRequest
			err = json.Unmarshal(result, &openAIReq)
			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, tt.expectedModel, openAIReq.Model)
			testutil.AssertEqual(t, tt.expectedMessageCount, len(openAIReq.Messages))

			// Check for system message
			if tt.expectSystemMessage {
				testutil.AssertEqual(t, "system", openAIReq.Messages[0].Role)
			}
		})
	}
}

func TestOpenAIConverter_FromGeneric_Response(t *testing.T) {
	converter := NewOpenAIConverter()

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
				"model": "gpt-4",
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
				"model": "gpt-3.5-turbo"
			}`,
			expectedID:  "msg_456",
			expectError: false,
		},
		{
			name: "response with string content",
			input: `{
				"id": "msg_789",
				"type": "message",
				"role": "assistant",
				"content": "Simple string response",
				"model": "gpt-4"
			}`,
			expectedID:  "msg_789",
			expectError: false,
		},
		{
			name:        "malformed JSON",
			input:       `{"id": "msg_123"`,
			expectError: true,
		},
		{
			name:        "malformed content",
			input:       `{"id": "msg_123", "content": invalid_json}`,
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
			var openAIResp OpenAIResponse
			err = json.Unmarshal(result, &openAIResp)
			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, tt.expectedID, openAIResp.ID)
			testutil.AssertEqual(t, "chat.completion", openAIResp.Object)
			testutil.AssertEqual(t, 1, len(openAIResp.Choices))
		})
	}
}

func TestOpenAIConverter_ConvertStreamEvent(t *testing.T) {
	converter := NewOpenAIConverter()

	tests := []struct {
		name     string
		input    []byte
		toFormat MessageFormat
	}{
		{
			name:     "basic stream event",
			input:    []byte("data: {\"id\": \"chatcmpl-123\", \"object\": \"chat.completion.chunk\"}"),
			toFormat: FormatAnthropic,
		},
		{
			name:     "empty event",
			input:    []byte(""),
			toFormat: FormatGoogle,
		},
		{
			name:     "complex event",
			input:    []byte("event: completion\ndata: {\"choices\": [{\"delta\": {\"content\": \"Hello\"}}]}"),
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

func TestOpenAIConverter_ContentParsing(t *testing.T) {
	converter := NewOpenAIConverter()

	t.Run("anthropic content array parsing", func(t *testing.T) {
		input := `{
			"model": "gpt-4",
			"messages": [
				{
					"role": "user",
					"content": [
						{"type": "text", "text": "First part "},
						{"type": "text", "text": "Second part"}
					]
				}
			],
			"max_tokens": 100
		}`

		result, err := converter.FromGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)

		var openAIReq OpenAIRequest
		err = json.Unmarshal(result, &openAIReq)
		testutil.AssertNoError(t, err)
		testutil.AssertContains(t, openAIReq.Messages[0].Content, "First part")
		testutil.AssertContains(t, openAIReq.Messages[0].Content, "Second part")
	})

	t.Run("generic object content parsing", func(t *testing.T) {
		input := `{
			"model": "gpt-4",
			"messages": [
				{
					"role": "user",
					"content": {"some": "object", "data": 123}
				}
			],
			"max_tokens": 100
		}`

		result, err := converter.FromGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)

		var openAIReq OpenAIRequest
		err = json.Unmarshal(result, &openAIReq)
		testutil.AssertNoError(t, err)
		testutil.AssertContains(t, openAIReq.Messages[0].Content, "some")
	})

	t.Run("malformed content in FromGeneric", func(t *testing.T) {
		input := `{
			"model": "gpt-4",
			"messages": [
				{
					"role": "user",
					"content": invalid_json
				}
			],
			"max_tokens": 100
		}`

		_, err := converter.FromGeneric(json.RawMessage(input), true)
		testutil.AssertError(t, err)
	})

	t.Run("response content parsing - string", func(t *testing.T) {
		input := `{
			"id": "msg_123",
			"type": "message",
			"role": "assistant",
			"content": "Simple string content",
			"model": "gpt-4"
		}`

		result, err := converter.FromGeneric(json.RawMessage(input), false)
		testutil.AssertNoError(t, err)

		var openAIResp OpenAIResponse
		err = json.Unmarshal(result, &openAIResp)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, "Simple string content", openAIResp.Choices[0].Message.Content)
	})

	t.Run("response content parsing - anthropic array", func(t *testing.T) {
		input := `{
			"id": "msg_123",
			"type": "message",
			"role": "assistant",
			"content": [
				{"type": "text", "text": "Part one "},
				{"type": "text", "text": "Part two"}
			],
			"model": "gpt-4"
		}`

		result, err := converter.FromGeneric(json.RawMessage(input), false)
		testutil.AssertNoError(t, err)

		var openAIResp OpenAIResponse
		err = json.Unmarshal(result, &openAIResp)
		testutil.AssertNoError(t, err)
		testutil.AssertContains(t, openAIResp.Choices[0].Message.Content, "Part one")
		testutil.AssertContains(t, openAIResp.Choices[0].Message.Content, "Part two")
	})

	t.Run("response content parsing - malformed", func(t *testing.T) {
		input := `{
			"id": "msg_123",
			"type": "message",
			"role": "assistant",
			"content": invalid_json,
			"model": "gpt-4"
		}`

		_, err := converter.FromGeneric(json.RawMessage(input), false)
		testutil.AssertError(t, err)
		// Error will be JSON unmarshaling error, not content parsing
		testutil.AssertContains(t, err.Error(), "failed to unmarshal generic response")
	})
}

func TestOpenAIConverter_ErrorHandling(t *testing.T) {
	converter := NewOpenAIConverter()

	t.Run("invalid request JSON", func(t *testing.T) {
		input := `{"model": "test", "messages": invalid}`
		_, err := converter.ToGeneric(json.RawMessage(input), true)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "failed to unmarshal OpenAI request")
	})

	t.Run("invalid response JSON", func(t *testing.T) {
		input := `{"id": "test", "choices": invalid}`
		_, err := converter.ToGeneric(json.RawMessage(input), false)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "failed to unmarshal OpenAI response")
	})

	t.Run("content marshal error in ToGeneric", func(t *testing.T) {
		// Test scenario where content marshaling might fail
		input := `{
			"model": "gpt-4",
			"messages": [
				{
					"role": "user",
					"content": "Hello world"
				}
			],
			"max_tokens": 100
		}`
		
		_, err := converter.ToGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)
	})

	t.Run("response content marshal error in ToGeneric", func(t *testing.T) {
		input := `{
			"id": "chatcmpl-123",
			"object": "chat.completion",
			"model": "gpt-4",
			"choices": [
				{
					"index": 0,
					"message": {
						"role": "assistant",
						"content": "Hello there!"
					},
					"finish_reason": "stop"
				}
			]
		}`
		
		_, err := converter.ToGeneric(json.RawMessage(input), false)
		testutil.AssertNoError(t, err)
	})
}

func TestOpenAIConverter_EdgeCases(t *testing.T) {
	converter := NewOpenAIConverter()

	t.Run("empty messages array", func(t *testing.T) {
		input := `{
			"model": "gpt-4",
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

	t.Run("only system messages", func(t *testing.T) {
		input := `{
			"model": "gpt-4",
			"messages": [
				{
					"role": "system",
					"content": "You are helpful"
				}
			],
			"max_tokens": 100
		}`

		result, err := converter.ToGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)

		var genericReq Request
		err = json.Unmarshal(result, &genericReq)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, "You are helpful", genericReq.System)
		testutil.AssertEqual(t, 0, len(genericReq.Messages))
	})

	t.Run("zero max_tokens", func(t *testing.T) {
		input := `{
			"model": "gpt-4",
			"messages": [
				{
					"role": "user",
					"content": "Hello"
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

	t.Run("empty system in FromGeneric", func(t *testing.T) {
		input := `{
			"model": "gpt-4",
			"messages": [
				{
					"role": "user",
					"content": "Hello"
				}
			],
			"system": "",
			"max_tokens": 100
		}`

		result, err := converter.FromGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)

		var openAIReq OpenAIRequest
		err = json.Unmarshal(result, &openAIReq)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 1, len(openAIReq.Messages)) // Only user message
		testutil.AssertEqual(t, "user", openAIReq.Messages[0].Role)
	})

	t.Run("empty content array response parsing", func(t *testing.T) {
		input := `{
			"id": "msg_123",
			"type": "message",
			"role": "assistant",
			"content": [],
			"model": "gpt-4"
		}`

		result, err := converter.FromGeneric(json.RawMessage(input), false)
		testutil.AssertNoError(t, err)

		var openAIResp OpenAIResponse
		err = json.Unmarshal(result, &openAIResp)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, "", openAIResp.Choices[0].Message.Content)
	})

	t.Run("non-text content parts ignored", func(t *testing.T) {
		input := `{
			"model": "gpt-4",
			"messages": [
				{
					"role": "user",
					"content": [
						{"type": "image", "text": "should be ignored"},
						{"type": "text", "text": "Hello world"}
					]
				}
			],
			"max_tokens": 100
		}`

		result, err := converter.FromGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)

		var openAIReq OpenAIRequest
		err = json.Unmarshal(result, &openAIReq)
		testutil.AssertNoError(t, err)
		testutil.AssertContains(t, openAIReq.Messages[0].Content, "Hello world")
	})
}