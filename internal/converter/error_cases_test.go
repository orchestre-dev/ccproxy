package converter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverterErrorHandling(t *testing.T) {
	t.Run("AnthropicConverter errors", func(t *testing.T) {
		conv := NewAnthropicConverter()
		
		// Invalid request JSON
		_, err := conv.ToGeneric(json.RawMessage(`{"invalid`), true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal Anthropic request")
		
		// Invalid response JSON
		_, err = conv.ToGeneric(json.RawMessage(`{"invalid`), false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal Anthropic response")
		
		// Invalid generic request
		_, err = conv.FromGeneric(json.RawMessage(`{"invalid`), true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal generic request")
		
		// Invalid generic response
		_, err = conv.FromGeneric(json.RawMessage(`{"invalid`), false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal generic response")
		
		// Invalid content format in FromGeneric
		invalidReq := Request{
			Messages: []Message{
				{
					Role:    "user",
					Content: json.RawMessage(`{"not": "valid", "content": true}`),
				},
			},
		}
		data, _ := json.Marshal(invalidReq)
		_, err = conv.FromGeneric(data, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse content")
	})

	t.Run("OpenAIConverter errors", func(t *testing.T) {
		conv := NewOpenAIConverter()
		
		// Invalid request JSON
		_, err := conv.ToGeneric(json.RawMessage(`{"invalid`), true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal OpenAI request")
		
		// Invalid response JSON
		_, err = conv.ToGeneric(json.RawMessage(`{"invalid`), false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal OpenAI response")
		
		// Invalid generic request
		_, err = conv.FromGeneric(json.RawMessage(`{"invalid`), true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal generic request")
		
		// Invalid generic response
		_, err = conv.FromGeneric(json.RawMessage(`{"invalid`), false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal generic response")
		
		// Invalid content in response
		invalidResp := Response{
			Content: json.RawMessage(`{"invalid": "content"}`),
		}
		data, _ := json.Marshal(invalidResp)
		_, err = conv.FromGeneric(data, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse content")
	})

	t.Run("GoogleConverter errors", func(t *testing.T) {
		conv := NewGoogleConverter()
		
		// Invalid request JSON
		_, err := conv.ToGeneric(json.RawMessage(`{"invalid`), true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal Google request")
		
		// Invalid response JSON
		_, err = conv.ToGeneric(json.RawMessage(`{"invalid`), false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal Google response")
		
		// Invalid generic request
		_, err = conv.FromGeneric(json.RawMessage(`{"invalid`), true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal generic request")
		
		// Invalid generic response
		_, err = conv.FromGeneric(json.RawMessage(`{"invalid`), false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal generic response")
		
		// Invalid content in response
		invalidResp := Response{
			Content: json.RawMessage(`{"invalid": "content"}`),
		}
		data, _ := json.Marshal(invalidResp)
		_, err = conv.FromGeneric(data, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse content")
	})

	t.Run("AWSConverter errors", func(t *testing.T) {
		conv := NewAWSConverter()
		
		// Invalid request JSON
		_, err := conv.ToGeneric(json.RawMessage(`{"invalid`), true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal AWS request")
		
		// Invalid response JSON
		_, err = conv.ToGeneric(json.RawMessage(`{"invalid`), false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal AWS response")
		
		// Invalid generic request
		_, err = conv.FromGeneric(json.RawMessage(`{"invalid`), true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal generic request")
		
		// Invalid generic response
		_, err = conv.FromGeneric(json.RawMessage(`{"invalid`), false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal generic response")
	})
}

func TestComplexContentConversion(t *testing.T) {
	conv := NewOpenAIConverter()

	t.Run("FromGeneric with invalid content object", func(t *testing.T) {
		genericReq := Request{
			Model: "gpt-4",
			Messages: []Message{
				{
					Role:    "user",
					Content: json.RawMessage(`{"complex": {"nested": ["invalid"]}, "object": true}`),
				},
			},
		}

		data, _ := json.Marshal(genericReq)
		result, err := conv.FromGeneric(data, true)
		require.NoError(t, err)

		var openAIReq OpenAIRequest
		err = json.Unmarshal(result, &openAIReq)
		require.NoError(t, err)

		// Should convert complex object to string representation
		assert.Len(t, openAIReq.Messages, 1)
		assert.Contains(t, openAIReq.Messages[0].Content, "complex")
	})
}

func TestGoogleConverter_MoreCases(t *testing.T) {
	conv := NewGoogleConverter()

	t.Run("FromGeneric with object content", func(t *testing.T) {
		genericReq := Request{
			Messages: []Message{
				{
					Role:    "user",
					Content: json.RawMessage(`{"type": "custom", "data": "test"}`),
				},
			},
		}

		data, _ := json.Marshal(genericReq)
		result, err := conv.FromGeneric(data, true)
		require.NoError(t, err)

		var googleReq GoogleRequest
		err = json.Unmarshal(result, &googleReq)
		require.NoError(t, err)

		// Should convert object to string
		assert.Len(t, googleReq.Contents, 1)
		assert.Contains(t, googleReq.Contents[0].Parts[0].Text, "custom")
	})

	t.Run("ToGeneric with empty config", func(t *testing.T) {
		googleReq := GoogleRequest{
			Contents: []GoogleContent{
				{
					Role:  "user",
					Parts: []GooglePart{{Text: "Hello"}},
				},
			},
			// No GenerationConfig
		}

		data, _ := json.Marshal(googleReq)
		result, err := conv.ToGeneric(data, true)
		require.NoError(t, err)

		var genericReq Request
		err = json.Unmarshal(result, &genericReq)
		require.NoError(t, err)

		assert.Equal(t, 0, genericReq.MaxTokens)
		assert.Equal(t, float64(0), genericReq.Temperature)
	})
}

func TestStreamEventConversionMore(t *testing.T) {
	t.Run("All converters stream event support", func(t *testing.T) {
		testData := []byte("event: test\ndata: {}\n\n")
		
		converters := []FormatConverter{
			NewAnthropicConverter(),
			NewOpenAIConverter(),
			NewGoogleConverter(),
			NewAWSConverter(),
		}
		
		for _, conv := range converters {
			result, err := conv.ConvertStreamEvent(testData, FormatGeneric)
			assert.NoError(t, err)
			assert.Equal(t, testData, result) // Currently pass-through
		}
	})
}

func TestMessageConverter_GenericFormat(t *testing.T) {
	converter := NewMessageConverter()

	t.Run("Convert to generic format", func(t *testing.T) {
		anthropicReq := AnthropicRequest{
			Model: "claude-3",
			Messages: []AnthropicMessage{
				{
					Role: "user",
					Content: []AnthropicContent{
						{Type: "text", Text: "Test"},
					},
				},
			},
		}

		data, _ := json.Marshal(anthropicReq)
		result, err := converter.ConvertRequest(data, FormatAnthropic, FormatGeneric)
		require.NoError(t, err)

		// Should return generic format
		var genericReq Request
		err = json.Unmarshal(result, &genericReq)
		require.NoError(t, err)
		assert.Equal(t, "claude-3", genericReq.Model)
	})

	t.Run("Convert response to generic", func(t *testing.T) {
		openAIResp := OpenAIResponse{
			ID:     "test",
			Model:  "gpt-4",
			Choices: []OpenAIChoice{
				{
					Message: OpenAIMessage{
						Role:    "assistant",
						Content: "Test response",
					},
				},
			},
		}

		data, _ := json.Marshal(openAIResp)
		result, err := converter.ConvertResponse(data, FormatOpenAI, FormatGeneric)
		require.NoError(t, err)

		// Should return generic format
		var genericResp Response
		err = json.Unmarshal(result, &genericResp)
		require.NoError(t, err)
		assert.Equal(t, "test", genericResp.ID)
	})
}

func TestAnthropicConverter_ComplexContent(t *testing.T) {
	conv := NewAnthropicConverter()

	t.Run("FromGeneric with invalid content array", func(t *testing.T) {
		genericResp := Response{
			ID:      "test",
			Type:    "message",
			Role:    "assistant",
			Content: json.RawMessage(`"not an array"`),
		}

		data, _ := json.Marshal(genericResp)
		_, err := conv.FromGeneric(data, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse content")
	})
}