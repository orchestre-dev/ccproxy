package converter

import (
	"encoding/json"
	"testing"

	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
)

func TestNewAWSConverter(t *testing.T) {
	converter := NewAWSConverter()
	testutil.AssertNotEqual(t, nil, converter)
}

func TestAWSConverter_ToGeneric_Request(t *testing.T) {
	converter := NewAWSConverter()

	tests := []struct {
		name             string
		input            string
		expectedMessages int
		expectedSystem   string
		expectError      bool
	}{
		{
			name: "basic request",
			input: `{
				"anthropic_version": "bedrock-2023-05-31",
				"messages": [
					{
						"role": "user",
						"content": [{"type": "text", "text": "Hello world"}]
					}
				],
				"max_tokens": 100,
				"temperature": 0.7
			}`,
			expectedMessages: 1,
			expectedSystem:   "",
			expectError:      false,
		},
		{
			name: "request with system message",
			input: `{
				"anthropic_version": "bedrock-2023-05-31",
				"messages": [
					{
						"role": "user",
						"content": [{"type": "text", "text": "Hello"}]
					}
				],
				"system": "You are a helpful assistant",
				"max_tokens": 200,
				"temperature": 0.5
			}`,
			expectedMessages: 1,
			expectedSystem:   "You are a helpful assistant",
			expectError:      false,
		},
		{
			name: "request with multiple messages",
			input: `{
				"anthropic_version": "bedrock-2023-05-31",
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
				"max_tokens": 150,
				"temperature": 0.8
			}`,
			expectedMessages: 3,
			expectedSystem:   "",
			expectError:      false,
		},
		{
			name: "request with additional parameters",
			input: `{
				"anthropic_version": "bedrock-2023-05-31",
				"messages": [
					{
						"role": "user",
						"content": [{"type": "text", "text": "Test"}]
					}
				],
				"max_tokens": 300,
				"temperature": 0.9,
				"top_p": 0.95,
				"top_k": 40,
				"stop_sequences": ["END", "STOP"]
			}`,
			expectedMessages: 1,
			expectedSystem:   "",
			expectError:      false,
		},
		{
			name: "request with raw message content",
			input: `{
				"anthropic_version": "bedrock-2023-05-31",
				"messages": [
					{
						"role": "user",
						"content": {"type": "text", "text": "Raw content"}
					}
				],
				"max_tokens": 100
			}`,
			expectedMessages: 1,
			expectedSystem:   "",
			expectError:      false,
		},
		{
			name:        "malformed JSON",
			input:       `{"anthropic_version": "bedrock-2023-05-31", "messages": [`,
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
			testutil.AssertEqual(t, tt.expectedMessages, len(genericReq.Messages))
			testutil.AssertEqual(t, tt.expectedSystem, genericReq.System)
		})
	}
}

func TestAWSConverter_ToGeneric_Response(t *testing.T) {
	converter := NewAWSConverter()

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
				"model": "anthropic.claude-3-haiku-20240307-v1:0",
				"type": "message",
				"role": "assistant",
				"content": [{"type": "text", "text": "Hello there!"}],
				"stop_reason": "end_turn",
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
				"model": "anthropic.claude-3-opus-20240229-v1:0",
				"type": "message",
				"role": "assistant",
				"content": [{"type": "text", "text": "Response without usage"}],
				"stop_reason": "end_turn"
			}`,
			expectedID:  "msg_456",
			expectError: false,
		},
		{
			name: "response with stop sequence",
			input: `{
				"id": "msg_789",
				"model": "anthropic.claude-3-sonnet-20240229-v1:0",
				"type": "message",
				"role": "assistant",
				"content": [{"type": "text", "text": "Response with stop"}],
				"stop_reason": "stop_sequence",
				"stop_sequence": "END",
				"usage": {
					"input_tokens": 20,
					"output_tokens": 8
				}
			}`,
			expectedID:  "msg_789",
			expectError: false,
		},
		{
			name: "response with complex content",
			input: `{
				"id": "msg_complex",
				"model": "anthropic.claude-3-haiku-20240307-v1:0",
				"type": "message",
				"role": "assistant",
				"content": [
					{"type": "text", "text": "First part. "},
					{"type": "text", "text": "Second part."}
				],
				"stop_reason": "max_tokens",
				"usage": {
					"input_tokens": 15,
					"output_tokens": 12
				}
			}`,
			expectedID:  "msg_complex",
			expectError: false,
		},
		{
			name:        "malformed JSON",
			input:       `{"id": "msg_123", "model": "test"`,
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

func TestAWSConverter_FromGeneric_Request(t *testing.T) {
	converter := NewAWSConverter()

	tests := []struct {
		name             string
		input            string
		expectedMessages int
		expectedMaxTokens int
		expectedSystem   string
		expectError      bool
	}{
		{
			name: "basic request",
			input: `{
				"model": "claude-3-haiku-20240307",
				"messages": [
					{
						"role": "user",
						"content": [{"type": "text", "text": "Hello world"}]
					}
				],
				"max_tokens": 100,
				"temperature": 0.7
			}`,
			expectedMessages: 1,
			expectedMaxTokens: 100,
			expectedSystem:   "",
			expectError:      false,
		},
		{
			name: "request with system",
			input: `{
				"model": "claude-3-opus-20240229",
				"messages": [
					{
						"role": "user",
						"content": [{"type": "text", "text": "Hello"}]
					}
				],
				"system": "You are helpful",
				"max_tokens": 200,
				"temperature": 0.5
			}`,
			expectedMessages: 1,
			expectedMaxTokens: 200,
			expectedSystem:   "You are helpful",
			expectError:      false,
		},
		{
			name: "request without max_tokens",
			input: `{
				"model": "claude-3-sonnet-20240229",
				"messages": [
					{
						"role": "user",
						"content": [{"type": "text", "text": "Test"}]
					}
				],
				"temperature": 0.8
			}`,
			expectedMessages: 1,
			expectedMaxTokens: 4096, // Default value
			expectedSystem:   "",
			expectError:      false,
		},
		{
			name: "request with zero max_tokens",
			input: `{
				"model": "claude-3-haiku-20240307",
				"messages": [
					{
						"role": "user",
						"content": [{"type": "text", "text": "Test"}]
					}
				],
				"max_tokens": 0,
				"temperature": 0.6
			}`,
			expectedMessages: 1,
			expectedMaxTokens: 4096, // Default applied
			expectedSystem:   "",
			expectError:      false,
		},
		{
			name: "request with multiple messages",
			input: `{
				"model": "claude-3-opus-20240229",
				"messages": [
					{
						"role": "user",
						"content": [{"type": "text", "text": "Hello"}]
					},
					{
						"role": "assistant",
						"content": [{"type": "text", "text": "Hi!"}]
					}
				],
				"max_tokens": 300
			}`,
			expectedMessages: 2,
			expectedMaxTokens: 300,
			expectedSystem:   "",
			expectError:      false,
		},
		{
			name:        "malformed JSON",
			input:       `{"messages": [`,
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
			var awsReq AWSRequest
			err = json.Unmarshal(result, &awsReq)
			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, "bedrock-2023-05-31", awsReq.AnthropicVersion)
			testutil.AssertEqual(t, tt.expectedMessages, len(awsReq.Messages))
			testutil.AssertEqual(t, tt.expectedMaxTokens, awsReq.MaxTokens)
			testutil.AssertEqual(t, tt.expectedSystem, awsReq.System)
		})
	}
}

func TestAWSConverter_FromGeneric_Response(t *testing.T) {
	converter := NewAWSConverter()

	tests := []struct {
		name        string
		input       string
		expectedID  string
		expectError bool
	}{
		{
			name: "basic response",
			input: `{
				"id": "resp_123",
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
			expectedID:  "resp_123",
			expectError: false,
		},
		{
			name: "response without usage",
			input: `{
				"id": "resp_456",
				"type": "message",
				"role": "assistant",
				"content": [{"type": "text", "text": "Response"}],
				"model": "claude-3-opus-20240229"
			}`,
			expectedID:  "resp_456",
			expectError: false,
		},
		{
			name: "response with complex content",
			input: `{
				"id": "resp_789",
				"type": "message",
				"role": "assistant",
				"content": [
					{"type": "text", "text": "Multi-part "},
					{"type": "text", "text": "response"}
				],
				"model": "claude-3-sonnet-20240229",
				"usage": {
					"input_tokens": 20,
					"output_tokens": 8,
					"total_tokens": 28
				}
			}`,
			expectedID:  "resp_789",
			expectError: false,
		},
		{
			name:        "malformed JSON",
			input:       `{"id": "resp_123"`,
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
			var awsResp AWSResponse
			err = json.Unmarshal(result, &awsResp)
			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, tt.expectedID, awsResp.ID)
			testutil.AssertEqual(t, "stop_sequence", awsResp.StopReason) // Default stop reason
		})
	}
}

func TestAWSConverter_ConvertStreamEvent(t *testing.T) {
	converter := NewAWSConverter()

	tests := []struct {
		name     string
		input    []byte
		toFormat MessageFormat
	}{
		{
			name:     "basic stream event",
			input:    []byte("data: {\"type\": \"message_start\", \"message\": {\"id\": \"msg_123\"}}"),
			toFormat: FormatAnthropic,
		},
		{
			name:     "empty event",
			input:    []byte(""),
			toFormat: FormatOpenAI,
		},
		{
			name:     "complex event",
			input:    []byte("event: content_block_delta\ndata: {\"delta\": {\"text\": \"Hello\"}}"),
			toFormat: FormatGoogle,
		},
		{
			name:     "usage event",
			input:    []byte("data: {\"type\": \"message_delta\", \"usage\": {\"output_tokens\": 25}}"),
			toFormat: FormatOpenAI,
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

func TestAWSConverter_DefaultValues(t *testing.T) {
	converter := NewAWSConverter()

	t.Run("default anthropic version", func(t *testing.T) {
		input := `{
			"messages": [
				{
					"role": "user",
					"content": [{"type": "text", "text": "Hello"}]
				}
			],
			"max_tokens": 100
		}`

		result, err := converter.FromGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)

		var awsReq AWSRequest
		err = json.Unmarshal(result, &awsReq)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, "bedrock-2023-05-31", awsReq.AnthropicVersion)
	})

	t.Run("default max tokens", func(t *testing.T) {
		input := `{
			"messages": [
				{
					"role": "user",
					"content": [{"type": "text", "text": "Hello"}]
				}
			]
		}`

		result, err := converter.FromGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)

		var awsReq AWSRequest
		err = json.Unmarshal(result, &awsReq)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 4096, awsReq.MaxTokens)
	})

	t.Run("default stop reason in response", func(t *testing.T) {
		input := `{
			"id": "resp_123",
			"type": "message",
			"role": "assistant",
			"content": [{"type": "text", "text": "Hello"}],
			"model": "claude-3-haiku-20240307"
		}`

		result, err := converter.FromGeneric(json.RawMessage(input), false)
		testutil.AssertNoError(t, err)

		var awsResp AWSResponse
		err = json.Unmarshal(result, &awsResp)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, "stop_sequence", awsResp.StopReason)
	})

	t.Run("preserve non-zero max tokens", func(t *testing.T) {
		input := `{
			"messages": [
				{
					"role": "user",
					"content": [{"type": "text", "text": "Hello"}]
				}
			],
			"max_tokens": 500
		}`

		result, err := converter.FromGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)

		var awsReq AWSRequest
		err = json.Unmarshal(result, &awsReq)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 500, awsReq.MaxTokens)
	})
}

func TestAWSConverter_UsageHandling(t *testing.T) {
	converter := NewAWSConverter()

	t.Run("usage mapping in ToGeneric", func(t *testing.T) {
		input := `{
			"id": "msg_123",
			"model": "anthropic.claude-3-haiku-20240307-v1:0",
			"type": "message",
			"role": "assistant",
			"content": [{"type": "text", "text": "Hello"}],
			"usage": {
				"input_tokens": 25,
				"output_tokens": 15
			}
		}`

		result, err := converter.ToGeneric(json.RawMessage(input), false)
		testutil.AssertNoError(t, err)

		var genericResp Response
		err = json.Unmarshal(result, &genericResp)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, genericResp.Usage)
		testutil.AssertEqual(t, 25, genericResp.Usage.InputTokens)
		testutil.AssertEqual(t, 15, genericResp.Usage.OutputTokens)
		testutil.AssertEqual(t, 40, genericResp.Usage.TotalTokens) // Calculated
	})

	t.Run("usage mapping in FromGeneric", func(t *testing.T) {
		input := `{
			"id": "resp_123",
			"type": "message",
			"role": "assistant",
			"content": [{"type": "text", "text": "Hello"}],
			"model": "claude-3-haiku-20240307",
			"usage": {
				"input_tokens": 30,
				"output_tokens": 20,
				"total_tokens": 50
			}
		}`

		result, err := converter.FromGeneric(json.RawMessage(input), false)
		testutil.AssertNoError(t, err)

		var awsResp AWSResponse
		err = json.Unmarshal(result, &awsResp)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, awsResp.Usage)
		testutil.AssertEqual(t, 30, awsResp.Usage.InputTokens)
		testutil.AssertEqual(t, 20, awsResp.Usage.OutputTokens)
		// Note: AWS format doesn't include total_tokens in the struct
	})

	t.Run("no usage in response", func(t *testing.T) {
		input := `{
			"id": "msg_123",
			"model": "anthropic.claude-3-haiku-20240307-v1:0",
			"type": "message",
			"role": "assistant",
			"content": [{"type": "text", "text": "Hello"}]
		}`

		result, err := converter.ToGeneric(json.RawMessage(input), false)
		testutil.AssertNoError(t, err)

		var genericResp Response
		err = json.Unmarshal(result, &genericResp)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, (*Usage)(nil), genericResp.Usage)
	})
}

func TestAWSConverter_ContentHandling(t *testing.T) {
	converter := NewAWSConverter()

	t.Run("raw message content preserved", func(t *testing.T) {
		input := `{
			"anthropic_version": "bedrock-2023-05-31",
			"messages": [
				{
					"role": "user",
					"content": [{"type": "text", "text": "Hello world"}]
				}
			],
			"max_tokens": 100
		}`

		result, err := converter.ToGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)

		var genericReq Request
		err = json.Unmarshal(result, &genericReq)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 1, len(genericReq.Messages))
		
		// Content should be preserved as raw JSON
		testutil.AssertTrue(t, len(genericReq.Messages[0].Content) > 0)
	})

	t.Run("content passed through in FromGeneric", func(t *testing.T) {
		input := `{
			"messages": [
				{
					"role": "user",
					"content": [{"type": "text", "text": "Test content"}]
				}
			],
			"max_tokens": 100
		}`

		result, err := converter.FromGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)

		var awsReq AWSRequest
		err = json.Unmarshal(result, &awsReq)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 1, len(awsReq.Messages))
		
		// Content should be preserved as raw JSON
		testutil.AssertTrue(t, len(awsReq.Messages[0].Content) > 0)
	})

	t.Run("response content preserved", func(t *testing.T) {
		input := `{
			"id": "msg_123",
			"model": "anthropic.claude-3-haiku-20240307-v1:0",
			"type": "message", 
			"role": "assistant",
			"content": [
				{"type": "text", "text": "Multi"},
				{"type": "text", "text": "part"}
			]
		}`

		result, err := converter.ToGeneric(json.RawMessage(input), false)
		testutil.AssertNoError(t, err)

		var genericResp Response
		err = json.Unmarshal(result, &genericResp)
		testutil.AssertNoError(t, err)
		
		// Content should be preserved as raw JSON
		testutil.AssertTrue(t, len(genericResp.Content) > 0)
	})
}

func TestAWSConverter_ErrorHandling(t *testing.T) {
	converter := NewAWSConverter()

	t.Run("invalid request JSON", func(t *testing.T) {
		input := `{"messages": invalid}`
		_, err := converter.ToGeneric(json.RawMessage(input), true)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "failed to unmarshal AWS request")
	})

	t.Run("invalid response JSON", func(t *testing.T) {
		input := `{"id": "test", "content": invalid}`
		_, err := converter.ToGeneric(json.RawMessage(input), false)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "failed to unmarshal AWS response")
	})

	t.Run("invalid generic request JSON", func(t *testing.T) {
		input := `{"messages": invalid}`
		_, err := converter.FromGeneric(json.RawMessage(input), true)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "failed to unmarshal generic request")
	})

	t.Run("invalid generic response JSON", func(t *testing.T) {
		input := `{"id": "test", "content": invalid}`
		_, err := converter.FromGeneric(json.RawMessage(input), false)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "failed to unmarshal generic response")
	})
}

func TestAWSConverter_EdgeCases(t *testing.T) {
	converter := NewAWSConverter()

	t.Run("empty messages array", func(t *testing.T) {
		input := `{
			"anthropic_version": "bedrock-2023-05-31",
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

	t.Run("zero temperature preserved", func(t *testing.T) {
		input := `{
			"messages": [
				{
					"role": "user",
					"content": [{"type": "text", "text": "Hello"}]
				}
			],
			"max_tokens": 100,
			"temperature": 0.0
		}`

		result, err := converter.FromGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)

		var awsReq AWSRequest
		err = json.Unmarshal(result, &awsReq)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 0.0, awsReq.Temperature)
	})

	t.Run("empty system message", func(t *testing.T) {
		input := `{
			"messages": [
				{
					"role": "user",
					"content": [{"type": "text", "text": "Hello"}]
				}
			],
			"system": "",
			"max_tokens": 100
		}`

		result, err := converter.FromGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)

		var awsReq AWSRequest
		err = json.Unmarshal(result, &awsReq)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, "", awsReq.System)
	})

	t.Run("missing required fields handled gracefully", func(t *testing.T) {
		input := `{
			"messages": [
				{
					"role": "user",
					"content": [{"type": "text", "text": "Hello"}]
				}
			]
		}`

		result, err := converter.FromGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)

		var awsReq AWSRequest
		err = json.Unmarshal(result, &awsReq)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, "bedrock-2023-05-31", awsReq.AnthropicVersion)
		testutil.AssertEqual(t, 4096, awsReq.MaxTokens) // Default applied
	})

	t.Run("all optional parameters", func(t *testing.T) {
		input := `{
			"anthropic_version": "bedrock-2023-05-31",
			"messages": [
				{
					"role": "user",
					"content": [{"type": "text", "text": "Test"}]
				}
			],
			"system": "Test system",
			"max_tokens": 200,
			"temperature": 0.7,
			"top_p": 0.95,
			"top_k": 40,
			"stop_sequences": ["END"]
		}`

		result, err := converter.ToGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)

		var genericReq Request
		err = json.Unmarshal(result, &genericReq)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, "Test system", genericReq.System)
		testutil.AssertEqual(t, 200, genericReq.MaxTokens)
		testutil.AssertEqual(t, 0.7, genericReq.Temperature)
	})
}