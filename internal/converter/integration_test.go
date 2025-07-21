package converter

import (
	"encoding/json"
	"testing"

	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
)

func TestMessageConverter_CrossFormatConversions(t *testing.T) {
	converter := NewMessageConverter()

	// Test data for cross-format conversions
	anthropicRequest := `{
		"model": "claude-3-haiku-20240307",
		"messages": [
			{
				"role": "user",
				"content": [{"type": "text", "text": "Hello world"}]
			}
		],
		"system": "You are a helpful assistant",
		"max_tokens": 100,
		"temperature": 0.7
	}`

	openAIRequest := `{
		"model": "gpt-4",
		"messages": [
			{
				"role": "system",
				"content": "You are a helpful assistant"
			},
			{
				"role": "user",
				"content": "Hello world"
			}
		],
		"max_tokens": 100,
		"temperature": 0.7
	}`

	googleRequest := `{
		"contents": [
			{
				"role": "user",
				"parts": [{"text": "You are a helpful assistant"}]
			},
			{
				"role": "user",
				"parts": [{"text": "Hello world"}]
			}
		],
		"generationConfig": {
			"maxOutputTokens": 100,
			"temperature": 0.7
		}
	}`

	awsRequest := `{
		"anthropic_version": "bedrock-2023-05-31",
		"messages": [
			{
				"role": "user",
				"content": [{"type": "text", "text": "Hello world"}]
			}
		],
		"system": "You are a helpful assistant",
		"max_tokens": 100,
		"temperature": 0.7
	}`

	testCases := []struct {
		name   string
		from   MessageFormat
		to     MessageFormat
		input  string
		verify func(t *testing.T, result json.RawMessage, to MessageFormat)
	}{
		{
			name:  "Anthropic to OpenAI",
			from:  FormatAnthropic,
			to:    FormatOpenAI,
			input: anthropicRequest,
			verify: func(t *testing.T, result json.RawMessage, to MessageFormat) {
				var req OpenAIRequest
				err := json.Unmarshal(result, &req)
				testutil.AssertNoError(t, err)
				testutil.AssertEqual(t, "claude-3-haiku-20240307", req.Model)
				testutil.AssertEqual(t, 2, len(req.Messages)) // System + user
				testutil.AssertEqual(t, "system", req.Messages[0].Role)
				testutil.AssertEqual(t, "user", req.Messages[1].Role)
			},
		},
		{
			name:  "Anthropic to Google",
			from:  FormatAnthropic,
			to:    FormatGoogle,
			input: anthropicRequest,
			verify: func(t *testing.T, result json.RawMessage, to MessageFormat) {
				var req GoogleRequest
				err := json.Unmarshal(result, &req)
				testutil.AssertNoError(t, err)
				testutil.AssertEqual(t, 2, len(req.Contents)) // System + user
				testutil.AssertNotEqual(t, nil, req.GenerationConfig)
				testutil.AssertEqual(t, 100, req.GenerationConfig.MaxOutputTokens)
			},
		},
		{
			name:  "Anthropic to AWS",
			from:  FormatAnthropic,
			to:    FormatAWS,
			input: anthropicRequest,
			verify: func(t *testing.T, result json.RawMessage, to MessageFormat) {
				var req AWSRequest
				err := json.Unmarshal(result, &req)
				testutil.AssertNoError(t, err)
				testutil.AssertEqual(t, "bedrock-2023-05-31", req.AnthropicVersion)
				testutil.AssertEqual(t, 1, len(req.Messages))
				testutil.AssertEqual(t, "You are a helpful assistant", req.System)
			},
		},
		{
			name:  "OpenAI to Anthropic",
			from:  FormatOpenAI,
			to:    FormatAnthropic,
			input: openAIRequest,
			verify: func(t *testing.T, result json.RawMessage, to MessageFormat) {
				var req AnthropicRequest
				err := json.Unmarshal(result, &req)
				testutil.AssertNoError(t, err)
				testutil.AssertEqual(t, "gpt-4", req.Model)
				testutil.AssertEqual(t, 1, len(req.Messages)) // User only, system extracted
				testutil.AssertEqual(t, "You are a helpful assistant", req.System)
			},
		},
		{
			name:  "OpenAI to Google",
			from:  FormatOpenAI,
			to:    FormatGoogle,
			input: openAIRequest,
			verify: func(t *testing.T, result json.RawMessage, to MessageFormat) {
				var req GoogleRequest
				err := json.Unmarshal(result, &req)
				testutil.AssertNoError(t, err)
				testutil.AssertEqual(t, 2, len(req.Contents)) // System as user + user
			},
		},
		{
			name:  "Google to OpenAI",
			from:  FormatGoogle,
			to:    FormatOpenAI,
			input: googleRequest,
			verify: func(t *testing.T, result json.RawMessage, to MessageFormat) {
				var req OpenAIRequest
				err := json.Unmarshal(result, &req)
				testutil.AssertNoError(t, err)
				testutil.AssertEqual(t, 2, len(req.Messages))
			},
		},
		{
			name:  "AWS to OpenAI",
			from:  FormatAWS,
			to:    FormatOpenAI,
			input: awsRequest,
			verify: func(t *testing.T, result json.RawMessage, to MessageFormat) {
				var req OpenAIRequest
				err := json.Unmarshal(result, &req)
				testutil.AssertNoError(t, err)
				testutil.AssertEqual(t, 2, len(req.Messages)) // System + user
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := converter.ConvertRequest(json.RawMessage(tc.input), tc.from, tc.to)
			testutil.AssertNoError(t, err)
			testutil.AssertNotEqual(t, nil, result)
			tc.verify(t, result, tc.to)
		})
	}
}

func TestMessageConverter_ResponseConversions(t *testing.T) {
	converter := NewMessageConverter()

	anthropicResponse := `{
		"id": "msg_123",
		"type": "message",
		"role": "assistant",
		"content": [{"type": "text", "text": "Hello there!"}],
		"model": "claude-3-haiku-20240307",
		"usage": {
			"input_tokens": 10,
			"output_tokens": 5
		}
	}`

	openAIResponse := `{
		"id": "chatcmpl-123",
		"object": "chat.completion",
		"created": 1677652288,
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
		],
		"usage": {
			"prompt_tokens": 10,
			"completion_tokens": 5,
			"total_tokens": 15
		}
	}`

	googleResponse := `{
		"candidates": [
			{
				"content": {
					"role": "model",
					"parts": [{"text": "Hello there!"}]
				},
				"finishReason": "STOP"
			}
		],
		"usageMetadata": {
			"promptTokenCount": 10,
			"candidatesTokenCount": 5,
			"totalTokenCount": 15
		}
	}`

	testCases := []struct {
		name   string
		from   MessageFormat
		to     MessageFormat
		input  string
		verify func(t *testing.T, result json.RawMessage, to MessageFormat)
	}{
		{
			name:  "Anthropic to OpenAI Response",
			from:  FormatAnthropic,
			to:    FormatOpenAI,
			input: anthropicResponse,
			verify: func(t *testing.T, result json.RawMessage, to MessageFormat) {
				var resp OpenAIResponse
				err := json.Unmarshal(result, &resp)
				testutil.AssertNoError(t, err)
				testutil.AssertEqual(t, "msg_123", resp.ID)
				testutil.AssertEqual(t, "chat.completion", resp.Object)
				testutil.AssertEqual(t, 1, len(resp.Choices))
				testutil.AssertEqual(t, "Hello there!", resp.Choices[0].Message.Content)
			},
		},
		{
			name:  "OpenAI to Anthropic Response",
			from:  FormatOpenAI,
			to:    FormatAnthropic,
			input: openAIResponse,
			verify: func(t *testing.T, result json.RawMessage, to MessageFormat) {
				var resp AnthropicResponse
				err := json.Unmarshal(result, &resp)
				testutil.AssertNoError(t, err)
				testutil.AssertEqual(t, "chatcmpl-123", resp.ID)
				testutil.AssertEqual(t, 1, len(resp.Content))
				testutil.AssertEqual(t, "text", resp.Content[0].Type)
			},
		},
		{
			name:  "Google to OpenAI Response",
			from:  FormatGoogle,
			to:    FormatOpenAI,
			input: googleResponse,
			verify: func(t *testing.T, result json.RawMessage, to MessageFormat) {
				var resp OpenAIResponse
				err := json.Unmarshal(result, &resp)
				testutil.AssertNoError(t, err)
				testutil.AssertEqual(t, 1, len(resp.Choices))
				testutil.AssertEqual(t, "assistant", resp.Choices[0].Message.Role) // model -> assistant
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := converter.ConvertResponse(json.RawMessage(tc.input), tc.from, tc.to)
			testutil.AssertNoError(t, err)
			testutil.AssertNotEqual(t, nil, result)
			tc.verify(t, result, tc.to)
		})
	}
}

func TestMessageConverter_RoundTripConversions(t *testing.T) {
	converter := NewMessageConverter()

	formats := []MessageFormat{
		FormatAnthropic,
		FormatOpenAI,
		FormatGoogle,
		FormatAWS,
	}

	// Test round-trip conversions: Format A -> Generic -> Format A
	for _, format := range formats {
		t.Run("RoundTrip_"+string(format), func(t *testing.T) {
			// Convert basicRequest (generic format) to target format, then back
			// Since we can't convert FROM FormatGeneric, we'll convert from Anthropic to format to Generic
			anthropicBasic := `{
				"model": "test-model",
				"messages": [
					{
						"role": "user",
						"content": [{"type": "text", "text": "Hello world"}]
					}
				],
				"max_tokens": 100,
				"temperature": 0.5
			}`

			// Convert from Anthropic to target format
			targetData, err := converter.ConvertRequest(json.RawMessage(anthropicBasic), FormatAnthropic, format)
			testutil.AssertNoError(t, err)

			// Then convert back to Anthropic
			backToAnthropic, err := converter.ConvertRequest(targetData, format, FormatAnthropic)
			testutil.AssertNoError(t, err)

			// Basic assertions - just verify no errors and data exists
			testutil.AssertNotEqual(t, nil, backToAnthropic)
			testutil.AssertTrue(t, len(backToAnthropic) > 0, "Result should not be empty")
		})
	}
}

func TestMessageConverter_ChainConversions(t *testing.T) {
	converter := NewMessageConverter()

	// Test converting through multiple formats: Anthropic -> OpenAI -> Google -> AWS
	anthropicRequest := `{
		"model": "claude-3-haiku-20240307",
		"messages": [
			{
				"role": "user",
				"content": [{"type": "text", "text": "Chain conversion test"}]
			}
		],
		"max_tokens": 200,
		"temperature": 0.8
	}`

	// Anthropic -> OpenAI
	openAIData, err := converter.ConvertRequest(json.RawMessage(anthropicRequest), FormatAnthropic, FormatOpenAI)
	testutil.AssertNoError(t, err)

	var openAIReq OpenAIRequest
	err = json.Unmarshal(openAIData, &openAIReq)
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, "claude-3-haiku-20240307", openAIReq.Model)

	// OpenAI -> Google
	googleData, err := converter.ConvertRequest(openAIData, FormatOpenAI, FormatGoogle)
	testutil.AssertNoError(t, err)

	var googleReq GoogleRequest
	err = json.Unmarshal(googleData, &googleReq)
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, 1, len(googleReq.Contents))

	// Google -> AWS
	awsData, err := converter.ConvertRequest(googleData, FormatGoogle, FormatAWS)
	testutil.AssertNoError(t, err)

	var awsReq AWSRequest
	err = json.Unmarshal(awsData, &awsReq)
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, "bedrock-2023-05-31", awsReq.AnthropicVersion)
	testutil.AssertEqual(t, 200, awsReq.MaxTokens)
}

func TestMessageConverter_ToGenericConversions(t *testing.T) {
	converter := NewMessageConverter()

	testCases := []struct {
		name   string
		format MessageFormat
		input  string
		verify func(t *testing.T, result json.RawMessage)
	}{
		{
			name:   "Anthropic to Generic",
			format: FormatAnthropic,
			input: `{
				"model": "claude-3-haiku-20240307",
				"messages": [{"role": "user", "content": [{"type": "text", "text": "Hello"}]}],
				"max_tokens": 100
			}`,
			verify: func(t *testing.T, result json.RawMessage) {
				var req Request
				err := json.Unmarshal(result, &req)
				testutil.AssertNoError(t, err)
				testutil.AssertEqual(t, "claude-3-haiku-20240307", req.Model)
			},
		},
		{
			name:   "OpenAI to Generic",
			format: FormatOpenAI,
			input: `{
				"model": "gpt-4",
				"messages": [{"role": "user", "content": "Hello"}],
				"max_tokens": 100
			}`,
			verify: func(t *testing.T, result json.RawMessage) {
				var req Request
				err := json.Unmarshal(result, &req)
				testutil.AssertNoError(t, err)
				testutil.AssertEqual(t, "gpt-4", req.Model)
			},
		},
		{
			name:   "Google to Generic",
			format: FormatGoogle,
			input: `{
				"contents": [{"role": "user", "parts": [{"text": "Hello"}]}],
				"generationConfig": {"maxOutputTokens": 100}
			}`,
			verify: func(t *testing.T, result json.RawMessage) {
				var req Request
				err := json.Unmarshal(result, &req)
				testutil.AssertNoError(t, err)
				testutil.AssertEqual(t, 100, req.MaxTokens)
			},
		},
		{
			name:   "AWS to Generic",
			format: FormatAWS,
			input: `{
				"anthropic_version": "bedrock-2023-05-31",
				"messages": [{"role": "user", "content": [{"type": "text", "text": "Hello"}]}],
				"max_tokens": 100
			}`,
			verify: func(t *testing.T, result json.RawMessage) {
				var req Request
				err := json.Unmarshal(result, &req)
				testutil.AssertNoError(t, err)
				testutil.AssertEqual(t, 100, req.MaxTokens)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := converter.ConvertRequest(json.RawMessage(tc.input), tc.format, FormatGeneric)
			testutil.AssertNoError(t, err)
			testutil.AssertNotEqual(t, nil, result)
			tc.verify(t, result)
		})
	}
}

func TestMessageConverter_EdgeCaseConversions(t *testing.T) {
	converter := NewMessageConverter()

	t.Run("Empty messages conversion", func(t *testing.T) {
		// Convert from Anthropic to all other formats
		anthropicInput := `{
			"model": "test-model",
			"messages": [],
			"max_tokens": 100
		}`
		formats := []MessageFormat{FormatOpenAI, FormatGoogle, FormatAWS}
		for _, format := range formats {
			result, err := converter.ConvertRequest(json.RawMessage(anthropicInput), FormatAnthropic, format)
			testutil.AssertNoError(t, err, "Failed to convert empty messages to "+string(format))
			testutil.AssertNotEqual(t, nil, result)
		}
	})

	t.Run("Zero values conversion", func(t *testing.T) {
		// Test converting to AWS (has default max_tokens) - use supported source format
		anthropicInput := `{
			"model": "test-model",
			"messages": [{"role": "user", "content": [{"type": "text", "text": "Test"}]}],
			"max_tokens": 0,
			"temperature": 0.0
		}`
		result, err := converter.ConvertRequest(json.RawMessage(anthropicInput), FormatAnthropic, FormatAWS)
		testutil.AssertNoError(t, err)

		var awsReq AWSRequest
		err = json.Unmarshal(result, &awsReq)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 4096, awsReq.MaxTokens) // Default applied
	})

	t.Run("Complex content conversion", func(t *testing.T) {
		// Anthropic-style complex content
		input := `{
			"model": "claude-3-haiku-20240307",
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

		// Convert to OpenAI (should concatenate text parts)
		result, err := converter.ConvertRequest(json.RawMessage(input), FormatAnthropic, FormatOpenAI)
		testutil.AssertNoError(t, err)

		var openAIReq OpenAIRequest
		err = json.Unmarshal(result, &openAIReq)
		testutil.AssertNoError(t, err)
		testutil.AssertContains(t, openAIReq.Messages[0].Content, "First part")
		testutil.AssertContains(t, openAIReq.Messages[0].Content, "Second part")
	})
}

func TestMessageConverter_ErrorPropagation(t *testing.T) {
	converter := NewMessageConverter()

	t.Run("Invalid JSON propagates error", func(t *testing.T) {
		input := `{"model": "test", "messages": invalid_json}`

		_, err := converter.ConvertRequest(json.RawMessage(input), FormatAnthropic, FormatOpenAI)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "failed to convert to generic format")
	})

	t.Run("Unsupported target format", func(t *testing.T) {
		input := `{"model": "test", "messages": []}`

		_, err := converter.ConvertRequest(json.RawMessage(input), FormatAnthropic, MessageFormat("unsupported"))
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "unsupported target format")
	})

	t.Run("From format conversion error", func(t *testing.T) {
		input := `{"model": "test", "messages": [{"role": "user", "content": [{"type": "text", "text": "test"}]}]}`

		_, err := converter.ConvertRequest(json.RawMessage(input), FormatAnthropic, MessageFormat("unsupported"))
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "unsupported target format")
	})
}

func TestMessageConverter_UsagePreservation(t *testing.T) {
	converter := NewMessageConverter()

	// Test that usage information is preserved across conversions
	anthropicResponse := `{
		"id": "msg_123",
		"type": "message",
		"role": "assistant",
		"content": [{"type": "text", "text": "Hello!"}],
		"model": "claude-3-haiku-20240307",
		"usage": {
			"input_tokens": 25,
			"output_tokens": 15
		}
	}`

	// Convert to OpenAI format
	openAIData, err := converter.ConvertResponse(json.RawMessage(anthropicResponse), FormatAnthropic, FormatOpenAI)
	testutil.AssertNoError(t, err)

	var openAIResp OpenAIResponse
	err = json.Unmarshal(openAIData, &openAIResp)
	testutil.AssertNoError(t, err)
	testutil.AssertNotEqual(t, nil, openAIResp.Usage)
	testutil.AssertEqual(t, 25, openAIResp.Usage.PromptTokens)
	testutil.AssertEqual(t, 15, openAIResp.Usage.CompletionTokens)
	testutil.AssertEqual(t, 40, openAIResp.Usage.TotalTokens) // Should be calculated

	// Convert back to Anthropic
	backToAnthropic, err := converter.ConvertResponse(openAIData, FormatOpenAI, FormatAnthropic)
	testutil.AssertNoError(t, err)

	var anthropicResp AnthropicResponse
	err = json.Unmarshal(backToAnthropic, &anthropicResp)
	testutil.AssertNoError(t, err)
	testutil.AssertNotEqual(t, nil, anthropicResp.Usage)
	testutil.AssertEqual(t, 25, anthropicResp.Usage.InputTokens)
	testutil.AssertEqual(t, 15, anthropicResp.Usage.OutputTokens)
}
