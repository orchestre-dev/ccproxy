package converter

import (
	"testing"

	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
)

func TestMessageConverter_StreamEventConversion(t *testing.T) {
	converter := NewMessageConverter()

	t.Run("Stream event pass-through behavior", func(t *testing.T) {
		// Current implementation passes through stream events
		// This test verifies that behavior and will need updating
		// when stream event conversion is implemented

		testCases := []struct {
			name     string
			from     MessageFormat
			to       MessageFormat
			input    []byte
			expected []byte
		}{
			{
				name:     "Anthropic to OpenAI stream event",
				from:     FormatAnthropic,
				to:       FormatOpenAI,
				input:    []byte("data: {\"type\": \"message_start\", \"message\": {\"id\": \"msg_123\"}}"),
				expected: []byte("data: {\"type\": \"message_start\", \"message\": {\"id\": \"msg_123\"}}"),
			},
			{
				name:     "OpenAI to Anthropic stream event",
				from:     FormatOpenAI,
				to:       FormatAnthropic,
				input:    []byte("data: {\"id\": \"chatcmpl-123\", \"object\": \"chat.completion.chunk\"}"),
				expected: []byte("data: {\"id\": \"chatcmpl-123\", \"object\": \"chat.completion.chunk\"}"),
			},
			{
				name:     "Google to AWS stream event",
				from:     FormatGoogle,
				to:       FormatAWS,
				input:    []byte("data: {\"candidates\": [{\"content\": {\"parts\": [{\"text\": \"Hello\"}]}}]}"),
				expected: []byte("data: {\"candidates\": [{\"content\": {\"parts\": [{\"text\": \"Hello\"}]}}]}"),
			},
			{
				name:     "Empty stream event",
				from:     FormatAnthropic,
				to:       FormatOpenAI,
				input:    []byte(""),
				expected: []byte(""),
			},
			{
				name:     "Complex multiline stream event",
				from:     FormatOpenAI,
				to:       FormatGoogle,
				input:    []byte("event: completion\ndata: {\"choices\": [{\"delta\": {\"content\": \"Hello world\"}}]}\nid: 123\n"),
				expected: []byte("event: completion\ndata: {\"choices\": [{\"delta\": {\"content\": \"Hello world\"}}]}\nid: 123\n"),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := converter.ConvertStreamEvent(tc.input, tc.from, tc.to)
				testutil.AssertNoError(t, err)
				testutil.AssertEqual(t, string(tc.expected), string(result))
			})
		}
	})

	t.Run("Same format stream events", func(t *testing.T) {
		// Test that same-format conversions work correctly
		testData := []byte("data: {\"test\": \"stream_event\"}")

		formats := []MessageFormat{
			FormatAnthropic,
			FormatOpenAI,
			FormatGoogle,
			FormatAWS,
		}

		for _, format := range formats {
			t.Run("Same_format_"+string(format), func(t *testing.T) {
				result, err := converter.ConvertStreamEvent(testData, format, format)
				testutil.AssertNoError(t, err)
				testutil.AssertEqual(t, string(testData), string(result))
			})
		}
	})
}

func TestAnthropicConverter_StreamEvents(t *testing.T) {
	converter := NewAnthropicConverter()

	testCases := []struct {
		name     string
		input    []byte
		toFormat MessageFormat
		expected []byte
	}{
		{
			name:     "Message start event",
			input:    []byte("data: {\"type\": \"message_start\", \"message\": {\"id\": \"msg_123\", \"type\": \"message\", \"role\": \"assistant\"}}"),
			toFormat: FormatOpenAI,
			expected: []byte("data: {\"type\": \"message_start\", \"message\": {\"id\": \"msg_123\", \"type\": \"message\", \"role\": \"assistant\"}}"),
		},
		{
			name:     "Content block start",
			input:    []byte("data: {\"type\": \"content_block_start\", \"index\": 0, \"content_block\": {\"type\": \"text\", \"text\": \"\"}}"),
			toFormat: FormatGoogle,
			expected: []byte("data: {\"type\": \"content_block_start\", \"index\": 0, \"content_block\": {\"type\": \"text\", \"text\": \"\"}}"),
		},
		{
			name:     "Content block delta",
			input:    []byte("data: {\"type\": \"content_block_delta\", \"index\": 0, \"delta\": {\"type\": \"text_delta\", \"text\": \"Hello\"}}"),
			toFormat: FormatAWS,
			expected: []byte("data: {\"type\": \"content_block_delta\", \"index\": 0, \"delta\": {\"type\": \"text_delta\", \"text\": \"Hello\"}}"),
		},
		{
			name:     "Content block stop",
			input:    []byte("data: {\"type\": \"content_block_stop\", \"index\": 0}"),
			toFormat: FormatOpenAI,
			expected: []byte("data: {\"type\": \"content_block_stop\", \"index\": 0}"),
		},
		{
			name:     "Message delta",
			input:    []byte("data: {\"type\": \"message_delta\", \"delta\": {\"stop_reason\": \"end_turn\", \"stop_sequence\": null}, \"usage\": {\"output_tokens\": 15}}"),
			toFormat: FormatGoogle,
			expected: []byte("data: {\"type\": \"message_delta\", \"delta\": {\"stop_reason\": \"end_turn\", \"stop_sequence\": null}, \"usage\": {\"output_tokens\": 15}}"),
		},
		{
			name:     "Message stop",
			input:    []byte("data: {\"type\": \"message_stop\"}"),
			toFormat: FormatAWS,
			expected: []byte("data: {\"type\": \"message_stop\"}"),
		},
		{
			name:     "Error event",
			input:    []byte("data: {\"type\": \"error\", \"error\": {\"type\": \"invalid_request_error\", \"message\": \"Invalid request\"}}"),
			toFormat: FormatOpenAI,
			expected: []byte("data: {\"type\": \"error\", \"error\": {\"type\": \"invalid_request_error\", \"message\": \"Invalid request\"}}"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := converter.ConvertStreamEvent(tc.input, tc.toFormat)
			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, string(tc.expected), string(result))
		})
	}
}

func TestOpenAIConverter_StreamEvents(t *testing.T) {
	converter := NewOpenAIConverter()

	testCases := []struct {
		name     string
		input    []byte
		toFormat MessageFormat
		expected []byte
	}{
		{
			name:     "Chat completion chunk",
			input:    []byte("data: {\"id\": \"chatcmpl-123\", \"object\": \"chat.completion.chunk\", \"created\": 1677652288, \"model\": \"gpt-4\", \"choices\": [{\"index\": 0, \"delta\": {\"content\": \"Hello\"}, \"finish_reason\": null}]}"),
			toFormat: FormatAnthropic,
			expected: []byte("data: {\"id\": \"chatcmpl-123\", \"object\": \"chat.completion.chunk\", \"created\": 1677652288, \"model\": \"gpt-4\", \"choices\": [{\"index\": 0, \"delta\": {\"content\": \"Hello\"}, \"finish_reason\": null}]}"),
		},
		{
			name:     "Role delta",
			input:    []byte("data: {\"id\": \"chatcmpl-123\", \"object\": \"chat.completion.chunk\", \"choices\": [{\"index\": 0, \"delta\": {\"role\": \"assistant\"}, \"finish_reason\": null}]}"),
			toFormat: FormatGoogle,
			expected: []byte("data: {\"id\": \"chatcmpl-123\", \"object\": \"chat.completion.chunk\", \"choices\": [{\"index\": 0, \"delta\": {\"role\": \"assistant\"}, \"finish_reason\": null}]}"),
		},
		{
			name:     "Content delta",
			input:    []byte("data: {\"id\": \"chatcmpl-123\", \"object\": \"chat.completion.chunk\", \"choices\": [{\"index\": 0, \"delta\": {\"content\": \" world!\"}, \"finish_reason\": null}]}"),
			toFormat: FormatAWS,
			expected: []byte("data: {\"id\": \"chatcmpl-123\", \"object\": \"chat.completion.chunk\", \"choices\": [{\"index\": 0, \"delta\": {\"content\": \" world!\"}, \"finish_reason\": null}]}"),
		},
		{
			name:     "Finish reason",
			input:    []byte("data: {\"id\": \"chatcmpl-123\", \"object\": \"chat.completion.chunk\", \"choices\": [{\"index\": 0, \"delta\": {}, \"finish_reason\": \"stop\"}]}"),
			toFormat: FormatAnthropic,
			expected: []byte("data: {\"id\": \"chatcmpl-123\", \"object\": \"chat.completion.chunk\", \"choices\": [{\"index\": 0, \"delta\": {}, \"finish_reason\": \"stop\"}]}"),
		},
		{
			name:     "Done event",
			input:    []byte("data: [DONE]"),
			toFormat: FormatGoogle,
			expected: []byte("data: [DONE]"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := converter.ConvertStreamEvent(tc.input, tc.toFormat)
			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, string(tc.expected), string(result))
		})
	}
}

func TestGoogleConverter_StreamEvents(t *testing.T) {
	converter := NewGoogleConverter()

	testCases := []struct {
		name     string
		input    []byte
		toFormat MessageFormat
		expected []byte
	}{
		{
			name:     "Generation start",
			input:    []byte("data: {\"candidates\": [{\"content\": {\"role\": \"model\", \"parts\": []}, \"safetyRatings\": []}]}"),
			toFormat: FormatAnthropic,
			expected: []byte("data: {\"candidates\": [{\"content\": {\"role\": \"model\", \"parts\": []}, \"safetyRatings\": []}]}"),
		},
		{
			name:     "Content generation",
			input:    []byte("data: {\"candidates\": [{\"content\": {\"role\": \"model\", \"parts\": [{\"text\": \"Hello world\"}]}, \"finishReason\": \"STOP\"}]}"),
			toFormat: FormatOpenAI,
			expected: []byte("data: {\"candidates\": [{\"content\": {\"role\": \"model\", \"parts\": [{\"text\": \"Hello world\"}]}, \"finishReason\": \"STOP\"}]}"),
		},
		{
			name:     "Usage metadata",
			input:    []byte("data: {\"usageMetadata\": {\"promptTokenCount\": 10, \"candidatesTokenCount\": 5, \"totalTokenCount\": 15}}"),
			toFormat: FormatAWS,
			expected: []byte("data: {\"usageMetadata\": {\"promptTokenCount\": 10, \"candidatesTokenCount\": 5, \"totalTokenCount\": 15}}"),
		},
		{
			name:     "Safety ratings",
			input:    []byte("data: {\"candidates\": [{\"safetyRatings\": [{\"category\": \"HARM_CATEGORY_HARASSMENT\", \"probability\": \"NEGLIGIBLE\"}]}]}"),
			toFormat: FormatAnthropic,
			expected: []byte("data: {\"candidates\": [{\"safetyRatings\": [{\"category\": \"HARM_CATEGORY_HARASSMENT\", \"probability\": \"NEGLIGIBLE\"}]}]}"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := converter.ConvertStreamEvent(tc.input, tc.toFormat)
			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, string(tc.expected), string(result))
		})
	}
}

func TestAWSConverter_StreamEvents(t *testing.T) {
	converter := NewAWSConverter()

	testCases := []struct {
		name     string
		input    []byte
		toFormat MessageFormat
		expected []byte
	}{
		{
			name:     "Message start",
			input:    []byte("data: {\"type\": \"message_start\", \"message\": {\"id\": \"msg_bedrock_123\", \"type\": \"message\", \"role\": \"assistant\", \"model\": \"anthropic.claude-3-haiku-20240307-v1:0\"}}"),
			toFormat: FormatAnthropic,
			expected: []byte("data: {\"type\": \"message_start\", \"message\": {\"id\": \"msg_bedrock_123\", \"type\": \"message\", \"role\": \"assistant\", \"model\": \"anthropic.claude-3-haiku-20240307-v1:0\"}}"),
		},
		{
			name:     "Content block delta",
			input:    []byte("data: {\"type\": \"content_block_delta\", \"index\": 0, \"delta\": {\"type\": \"text_delta\", \"text\": \"Hello from Bedrock\"}}"),
			toFormat: FormatOpenAI,
			expected: []byte("data: {\"type\": \"content_block_delta\", \"index\": 0, \"delta\": {\"type\": \"text_delta\", \"text\": \"Hello from Bedrock\"}}"),
		},
		{
			name:     "Message delta with usage",
			input:    []byte("data: {\"type\": \"message_delta\", \"delta\": {\"stop_reason\": \"end_turn\"}, \"usage\": {\"input_tokens\": 20, \"output_tokens\": 10}}"),
			toFormat: FormatGoogle,
			expected: []byte("data: {\"type\": \"message_delta\", \"delta\": {\"stop_reason\": \"end_turn\"}, \"usage\": {\"input_tokens\": 20, \"output_tokens\": 10}}"),
		},
		{
			name:     "Message stop",
			input:    []byte("data: {\"type\": \"message_stop\"}"),
			toFormat: FormatAnthropic,
			expected: []byte("data: {\"type\": \"message_stop\"}"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := converter.ConvertStreamEvent(tc.input, tc.toFormat)
			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, string(tc.expected), string(result))
		})
	}
}

func TestStreamEvent_EdgeCases(t *testing.T) {
	converter := NewMessageConverter()

	t.Run("Empty stream events", func(t *testing.T) {
		formats := []MessageFormat{FormatAnthropic, FormatOpenAI, FormatGoogle, FormatAWS}
		
		for _, from := range formats {
			for _, to := range formats {
				t.Run(string(from)+"_to_"+string(to), func(t *testing.T) {
					result, err := converter.ConvertStreamEvent([]byte(""), from, to)
					testutil.AssertNoError(t, err)
					testutil.AssertEqual(t, "", string(result))
				})
			}
		}
	})

	t.Run("Whitespace only events", func(t *testing.T) {
		whitespaceEvents := [][]byte{
			[]byte("   "),
			[]byte("\n"),
			[]byte("\t"),
			[]byte("\r\n"),
			[]byte("  \t\n  "),
		}

		for _, event := range whitespaceEvents {
			result, err := converter.ConvertStreamEvent(event, FormatAnthropic, FormatOpenAI)
			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, string(event), string(result))
		}
	})

	t.Run("Binary data events", func(t *testing.T) {
		// Test with binary data that might contain null bytes
		binaryData := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}
		
		result, err := converter.ConvertStreamEvent(binaryData, FormatAnthropic, FormatOpenAI)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, string(binaryData), string(result))
	})

	t.Run("Very large stream events", func(t *testing.T) {
		// Test with large stream events
		largeContent := make([]byte, 10240) // 10KB
		for i := range largeContent {
			largeContent[i] = byte('A' + (i % 26))
		}

		result, err := converter.ConvertStreamEvent(largeContent, FormatGoogle, FormatAWS)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, string(largeContent), string(result))
	})

	t.Run("Stream events with special characters", func(t *testing.T) {
		specialChars := []byte("data: {\"content\": \"Hello \\n\\r\\t \\\"world\\\" 🌍\"}")
		
		result, err := converter.ConvertStreamEvent(specialChars, FormatOpenAI, FormatGoogle)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, string(specialChars), string(result))
	})
}

func TestStreamEvent_SSEFormat(t *testing.T) {
	converter := NewMessageConverter()

	t.Run("Standard SSE format events", func(t *testing.T) {
		sseEvents := [][]byte{
			[]byte("data: {\"test\": \"value\"}\n\n"),
			[]byte("event: message\ndata: {\"id\": \"123\"}\n\n"),
			[]byte("id: 456\nevent: update\ndata: {\"status\": \"complete\"}\n\n"),
			[]byte(": This is a comment\ndata: {\"comment\": \"ignored\"}\n\n"),
			[]byte("retry: 3000\ndata: {\"retry\": true}\n\n"),
		}

		for i, event := range sseEvents {
			t.Run("SSE_event_"+string(rune('A'+i)), func(t *testing.T) {
				result, err := converter.ConvertStreamEvent(event, FormatAnthropic, FormatOpenAI)
				testutil.AssertNoError(t, err)
				testutil.AssertEqual(t, string(event), string(result))
			})
		}
	})

	t.Run("Malformed SSE events", func(t *testing.T) {
		// These are malformed but should still pass through
		malformedEvents := [][]byte{
			[]byte("data: incomplete"),
			[]byte("event: no-data\n"),
			[]byte("data: {\"unclosed\": \"json"),
			[]byte("invalid: header\ndata: test"),
		}

		for i, event := range malformedEvents {
			t.Run("Malformed_SSE_"+string(rune('A'+i)), func(t *testing.T) {
				result, err := converter.ConvertStreamEvent(event, FormatOpenAI, FormatGoogle)
				testutil.AssertNoError(t, err)
				testutil.AssertEqual(t, string(event), string(result))
			})
		}
	})
}

func TestStreamEvent_ConcurrentAccess(t *testing.T) {
	// Test that stream event conversion is safe for concurrent use
	converter := NewMessageConverter()
	
	t.Run("Concurrent stream conversions", func(t *testing.T) {
		testEvent := []byte("data: {\"concurrent\": \"test\"}")
		
		// Run multiple conversions concurrently
		done := make(chan bool, 10)
		
		for i := 0; i < 10; i++ {
			go func(id int) {
				defer func() { done <- true }()
				
				for j := 0; j < 100; j++ {
					result, err := converter.ConvertStreamEvent(testEvent, FormatAnthropic, FormatOpenAI)
					testutil.AssertNoError(t, err)
					testutil.AssertEqual(t, string(testEvent), string(result))
				}
			}(i)
		}
		
		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}