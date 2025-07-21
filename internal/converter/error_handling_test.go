package converter

import (
	"encoding/json"
	"testing"

	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
)

func TestMessageConverter_ErrorHandling(t *testing.T) {
	converter := NewMessageConverter()

	t.Run("Request conversion errors", func(t *testing.T) {
		tests := []struct {
			name        string
			from        MessageFormat
			to          MessageFormat
			input       string
			expectedErr string
		}{
			{
				name:        "Invalid JSON in request",
				from:        FormatAnthropic,
				to:          FormatOpenAI,
				input:       `{"model": "test", "messages": invalid_json}`,
				expectedErr: "failed to convert to generic format",
			},
			{
				name:        "Empty JSON in request OpenAI",
				from:        FormatOpenAI,
				to:          FormatAnthropic,
				input:       `{}`,
				expectedErr: "", // Empty structure, should not fail
			},
			{
				name:        "Malformed JSON in request",
				from:        FormatGoogle,
				to:          FormatAWS,
				input:       `{"contents": [`,
				expectedErr: "failed to convert to generic format",
			},
			{
				name:        "Null JSON in request AWS",
				from:        FormatAWS,
				to:          FormatGoogle,
				input:       `null`,
				expectedErr: "", // Null gets parsed as zero value, should not fail
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := converter.ConvertRequest(json.RawMessage(tt.input), tt.from, tt.to)
				if tt.expectedErr == "" {
					testutil.AssertNoError(t, err)
				} else {
					testutil.AssertError(t, err)
					if tt.expectedErr != "" {
						testutil.AssertContains(t, err.Error(), tt.expectedErr)
					}
				}
			})
		}
	})

	t.Run("Response conversion errors", func(t *testing.T) {
		tests := []struct {
			name        string
			from        MessageFormat
			to          MessageFormat
			input       string
			expectedErr string
		}{
			{
				name:        "Invalid JSON in response",
				from:        FormatAnthropic,
				to:          FormatOpenAI,
				input:       `{"id": "test", "content": invalid_json}`,
				expectedErr: "failed to convert to generic format",
			},
			{
				name:        "OpenAI response with no choices",
				from:        FormatOpenAI,
				to:          FormatAnthropic,
				input:       `{"id": "test", "choices": []}`,
				expectedErr: "no choices in OpenAI response",
			},
			{
				name:        "Google response with no candidates",
				from:        FormatGoogle,
				to:          FormatOpenAI,
				input:       `{"candidates": []}`,
				expectedErr: "no candidates in Google response",
			},
			{
				name:        "Malformed JSON in response",
				from:        FormatAWS,
				to:          FormatGoogle,
				input:       `{"id": "test", "content": [`,
				expectedErr: "failed to convert to generic format",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := converter.ConvertResponse(json.RawMessage(tt.input), tt.from, tt.to)
				testutil.AssertError(t, err)
				testutil.AssertContains(t, err.Error(), tt.expectedErr)
			})
		}
	})

	t.Run("Stream event conversion errors", func(t *testing.T) {
		tests := []struct {
			name        string
			from        MessageFormat
			to          MessageFormat
			input       []byte
			expectedErr string
		}{
			{
				name:        "Unsupported source format for stream",
				from:        MessageFormat("unsupported"),
				to:          FormatAnthropic,
				input:       []byte("data: test"),
				expectedErr: "unsupported source format",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := converter.ConvertStreamEvent(tt.input, tt.from, tt.to)
				testutil.AssertError(t, err)
				testutil.AssertContains(t, err.Error(), tt.expectedErr)
			})
		}
	})
}

func TestConverters_ContentParsingErrors(t *testing.T) {
	t.Run("Anthropic converter content parsing errors", func(t *testing.T) {
		converter := NewAnthropicConverter()

		// Test content parsing error in FromGeneric
		invalidContentInput := `{
			"model": "claude-3-haiku-20240307",
			"messages": [
				{
					"role": "user",
					"content": invalid_json
				}
			],
			"max_tokens": 100
		}`

		_, err := converter.FromGeneric(json.RawMessage(invalidContentInput), true)
		testutil.AssertError(t, err)
		// The error will actually be a JSON unmarshaling error, not a content parsing error
		testutil.AssertContains(t, err.Error(), "failed to unmarshal generic request")
	})

	t.Run("OpenAI converter content parsing errors", func(t *testing.T) {
		converter := NewOpenAIConverter()

		// Test malformed content in FromGeneric request
		invalidRequestInput := `{
			"model": "gpt-4",
			"messages": [
				{
					"role": "user",
					"content": invalid_json
				}
			]
		}`

		_, err := converter.FromGeneric(json.RawMessage(invalidRequestInput), true)
		testutil.AssertError(t, err)

		// Test malformed content in FromGeneric response
		invalidResponseInput := `{
			"id": "msg_123",
			"content": invalid_json,
			"role": "assistant"
		}`

		_, err = converter.FromGeneric(json.RawMessage(invalidResponseInput), false)
		testutil.AssertError(t, err)
		// The error will actually be a JSON unmarshaling error
		testutil.AssertContains(t, err.Error(), "failed to unmarshal generic response")
	})

	t.Run("Google converter content parsing errors", func(t *testing.T) {
		converter := NewGoogleConverter()

		// Test object content parsing error in FromGeneric
		invalidInput := `{
			"messages": [
				{
					"role": "user",
					"content": invalid_json
				}
			]
		}`

		_, err := converter.FromGeneric(json.RawMessage(invalidInput), true)
		testutil.AssertError(t, err)
	})
}

func TestConverters_JSONMarshallingErrors(t *testing.T) {
	// These tests verify error handling paths, though in practice
	// they're hard to trigger with valid Go data structures

	t.Run("Test error paths exist", func(t *testing.T) {
		// Verify that error handling code paths exist in converters
		// by testing with various edge cases

		converter := NewMessageConverter()

		// Test with extremely large JSON that might cause marshaling issues
		hugeContent := make([]byte, 1024*1024) // 1MB of data
		for i := range hugeContent {
			hugeContent[i] = 'A'
		}

		validButLargeJSON := `{
			"model": "test-model",
			"messages": [
				{
					"role": "user",
					"content": [{"type": "text", "text": "` + string(hugeContent) + `"}]
				}
			],
			"max_tokens": 100
		}`

		// This should succeed but tests the marshaling path
		_, err := converter.ConvertRequest(json.RawMessage(validButLargeJSON), FormatAnthropic, FormatOpenAI)
		testutil.AssertNoError(t, err, "Large but valid JSON should not fail")
	})
}

func TestConverters_StructuralErrors(t *testing.T) {
	t.Run("Missing required fields", func(t *testing.T) {
		converter := NewMessageConverter()

		// Test OpenAI response without required fields
		incompleteResponse := `{
			"id": "test"
		}`

		// This should work but produce a response with empty choices
		// which would then fail validation
		_, err := converter.ConvertResponse(json.RawMessage(incompleteResponse), FormatOpenAI, FormatAnthropic)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "no choices")
	})

	t.Run("Type mismatches", func(t *testing.T) {
		converter := NewMessageConverter()

		// Test with wrong data types
		wrongTypeInput := `{
			"model": 123,
			"messages": "not_an_array",
			"max_tokens": "not_a_number"
		}`

		_, err := converter.ConvertRequest(json.RawMessage(wrongTypeInput), FormatGeneric, FormatAnthropic)
		testutil.AssertError(t, err)
	})
}

func TestConverters_EdgeCaseErrorHandling(t *testing.T) {
	t.Run("Null and undefined values", func(t *testing.T) {
		converter := NewMessageConverter()

		nullFieldsInput := `{
			"model": null,
			"messages": null,
			"max_tokens": null,
			"temperature": null
		}`

		// Should handle null values gracefully
		_, err := converter.ConvertRequest(json.RawMessage(nullFieldsInput), FormatGeneric, FormatAnthropic)
		testutil.AssertError(t, err) // Likely to fail due to null messages array
	})

	t.Run("Empty strings and zero values", func(t *testing.T) {
		converter := NewMessageConverter()

		emptyValuesInput := `{
			"model": "",
			"messages": [],
			"max_tokens": 0,
			"temperature": 0.0,
			"system": ""
		}`

		// Should handle empty values gracefully - but need to use supported source format
		result, err := converter.ConvertRequest(json.RawMessage(emptyValuesInput), FormatAnthropic, FormatOpenAI)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, result)
	})

	t.Run("Very deeply nested content", func(t *testing.T) {
		converter := NewMessageConverter()

		// Create deeply nested JSON content
		deeplyNested := `{
			"model": "test",
			"messages": [
				{
					"role": "user",
					"content": [
						{
							"type": "text",
							"text": "Hello",
							"nested": {
								"level1": {
									"level2": {
										"level3": {
											"data": "deep"
										}
									}
								}
							}
						}
					]
				}
			],
			"max_tokens": 100
		}`

		// Should handle complex nested structures
		_, err := converter.ConvertRequest(json.RawMessage(deeplyNested), FormatAnthropic, FormatOpenAI)
		testutil.AssertNoError(t, err)
	})
}

func TestConverters_RecoveryFromErrors(t *testing.T) {
	converter := NewMessageConverter()

	t.Run("Partial success scenarios", func(t *testing.T) {
		// Test scenarios where some parts of the conversion succeed
		// but others might have issues

		partiallyValidInput := `{
			"model": "test-model",
			"messages": [
				{
					"role": "user",
					"content": [{"type": "text", "text": "Valid message"}]
				},
				{
					"role": "assistant",
					"content": [{"type": "text", "text": "Valid response"}]
				}
			],
			"max_tokens": 100
		}`

		// Should handle properly formatted content
		result, err := converter.ConvertRequest(json.RawMessage(partiallyValidInput), FormatAnthropic, FormatOpenAI)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, result)
	})
}

func TestConverters_ErrorMessageQuality(t *testing.T) {
	t.Run("Error messages are informative", func(t *testing.T) {
		converter := NewMessageConverter()

		// Test that error messages provide useful information
		invalidJSON := `{"model": "test", "messages": invalid}`

		_, err := converter.ConvertRequest(json.RawMessage(invalidJSON), FormatAnthropic, FormatOpenAI)
		testutil.AssertError(t, err)

		// Error should indicate what failed and where
		errorMsg := err.Error()
		testutil.AssertContains(t, errorMsg, "failed to convert")
		// Should not contain sensitive information or implementation details
		testutil.AssertNotContains(t, errorMsg, "panic")
		testutil.AssertNotContains(t, errorMsg, "nil pointer")
	})

	t.Run("Consistent error handling across converters", func(t *testing.T) {
		converter := NewMessageConverter()
		invalidJSON := `{"invalid": "json", "structure": incomplete`

		// All converters should handle invalid JSON similarly
		formats := []MessageFormat{FormatAnthropic, FormatOpenAI, FormatGoogle, FormatAWS}

		for _, format := range formats {
			_, err := converter.ConvertRequest(json.RawMessage(invalidJSON), format, FormatGeneric)
			testutil.AssertError(t, err, "Format "+string(format)+" should handle invalid JSON with error")
		}
	})
}
