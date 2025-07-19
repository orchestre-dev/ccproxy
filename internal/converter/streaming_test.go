package converter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStreamEventConversion(t *testing.T) {
	converter := NewMessageConverter()

	t.Run("ConvertStreamEvent pass-through", func(t *testing.T) {
		// For now, stream events are passed through unchanged
		testData := []byte("data: {\"test\": \"event\"}\n\n")
		
		result, err := converter.ConvertStreamEvent(testData, FormatAnthropic, FormatOpenAI)
		require.NoError(t, err)
		assert.Equal(t, testData, result)
	})

	t.Run("Same format stream event", func(t *testing.T) {
		testData := []byte("event: message_start\ndata: {\"id\": \"123\"}\n\n")
		
		result, err := converter.ConvertStreamEvent(testData, FormatOpenAI, FormatOpenAI)
		require.NoError(t, err)
		assert.Equal(t, testData, result)
	})
}

func TestGoogleConverter_EdgeCases(t *testing.T) {
	conv := NewGoogleConverter()

	t.Run("FromGeneric with system message", func(t *testing.T) {
		genericReq := Request{
			Model: "gemini-pro",
			System: "You are a helpful assistant",
			Messages: []Message{
				{
					Role:    "user",
					Content: json.RawMessage(`"Hello"`),
				},
			},
			MaxTokens:   1000,
			Temperature: 0.7,
		}

		data, _ := json.Marshal(genericReq)
		result, err := conv.FromGeneric(data, true)
		require.NoError(t, err)

		var googleReq GoogleRequest
		err = json.Unmarshal(result, &googleReq)
		require.NoError(t, err)

		// System message should be added as first user message
		assert.Len(t, googleReq.Contents, 2)
		assert.Equal(t, "user", googleReq.Contents[0].Role)
		assert.Equal(t, "You are a helpful assistant", googleReq.Contents[0].Parts[0].Text)
		
		// Check generation config
		assert.NotNil(t, googleReq.GenerationConfig)
		assert.Equal(t, 1000, googleReq.GenerationConfig.MaxOutputTokens)
		assert.Equal(t, 0.7, googleReq.GenerationConfig.Temperature)
	})

	t.Run("ToGeneric response without candidates", func(t *testing.T) {
		googleResp := GoogleResponse{
			Candidates: []GoogleCandidate{},
		}

		data, _ := json.Marshal(googleResp)
		_, err := conv.ToGeneric(data, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no candidates")
	})
}

func TestAnthropicConverter_EdgeCases(t *testing.T) {
	conv := NewAnthropicConverter()

	t.Run("FromGeneric with text string content", func(t *testing.T) {
		genericReq := Request{
			Model: "claude-3",
			Messages: []Message{
				{
					Role:    "user",
					Content: json.RawMessage(`"Plain text message"`),
				},
			},
		}

		data, _ := json.Marshal(genericReq)
		result, err := conv.FromGeneric(data, true)
		require.NoError(t, err)

		var anthropicReq AnthropicRequest
		err = json.Unmarshal(result, &anthropicReq)
		require.NoError(t, err)

		assert.Len(t, anthropicReq.Messages, 1)
		assert.Len(t, anthropicReq.Messages[0].Content, 1)
		assert.Equal(t, "text", anthropicReq.Messages[0].Content[0].Type)
		assert.Equal(t, "Plain text message", anthropicReq.Messages[0].Content[0].Text)
	})

	t.Run("ToGeneric with metadata", func(t *testing.T) {
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
			Metadata: map[string]interface{}{
				"user_id": "123",
				"session": "abc",
			},
		}

		data, _ := json.Marshal(anthropicReq)
		result, err := conv.ToGeneric(data, true)
		require.NoError(t, err)

		var genericReq Request
		err = json.Unmarshal(result, &genericReq)
		require.NoError(t, err)

		assert.NotNil(t, genericReq.Metadata)
		
		var metadata map[string]interface{}
		err = json.Unmarshal(genericReq.Metadata, &metadata)
		require.NoError(t, err)
		assert.Equal(t, "123", metadata["user_id"])
		assert.Equal(t, "abc", metadata["session"])
	})
}

func TestOpenAIConverter_EdgeCases(t *testing.T) {
	conv := NewOpenAIConverter()

	t.Run("ToGeneric response no choices", func(t *testing.T) {
		openAIResp := OpenAIResponse{
			ID:      "test",
			Model:   "gpt-4",
			Choices: []OpenAIChoice{},
		}

		data, _ := json.Marshal(openAIResp)
		_, err := conv.ToGeneric(data, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no choices")
	})

	t.Run("FromGeneric with name field", func(t *testing.T) {
		genericReq := Request{
			Model: "gpt-4",
			Messages: []Message{
				{
					Role:    "user",
					Content: json.RawMessage(`"Hello"`),
					Name:    "test_user",
				},
			},
		}

		data, _ := json.Marshal(genericReq)
		result, err := conv.FromGeneric(data, true)
		require.NoError(t, err)

		var openAIReq OpenAIRequest
		err = json.Unmarshal(result, &openAIReq)
		require.NoError(t, err)

		assert.Len(t, openAIReq.Messages, 1)
		assert.Equal(t, "test_user", openAIReq.Messages[0].Name)
	})
}

func TestAWSConverter_EdgeCases(t *testing.T) {
	conv := NewAWSConverter()

	t.Run("ToGeneric with stop reason", func(t *testing.T) {
		awsResp := AWSResponse{
			ID:         "msg_123",
			Type:       "message",
			Role:       "assistant",
			Content:    json.RawMessage(`[{"type": "text", "text": "Hello"}]`),
			Model:      "claude-v2",
			StopReason: "stop_sequence",
			Usage: &AWSUsage{
				InputTokens:  10,
				OutputTokens: 5,
			},
		}

		data, _ := json.Marshal(awsResp)
		result, err := conv.ToGeneric(data, false)
		require.NoError(t, err)

		var genericResp Response
		err = json.Unmarshal(result, &genericResp)
		require.NoError(t, err)

		assert.Equal(t, "msg_123", genericResp.ID)
		assert.Equal(t, 15, genericResp.Usage.TotalTokens)
	})
}

func TestMessageConverter_Errors(t *testing.T) {
	converter := NewMessageConverter()

	t.Run("Invalid JSON request", func(t *testing.T) {
		invalidJSON := json.RawMessage(`{invalid json`)
		_, err := converter.ConvertRequest(invalidJSON, FormatAnthropic, FormatOpenAI)
		assert.Error(t, err)
	})

	t.Run("Invalid JSON response", func(t *testing.T) {
		invalidJSON := json.RawMessage(`{invalid json`)
		_, err := converter.ConvertResponse(invalidJSON, FormatOpenAI, FormatAnthropic)
		assert.Error(t, err)
	})

	t.Run("Unsupported target format", func(t *testing.T) {
		data := json.RawMessage(`{}`)
		_, err := converter.ConvertRequest(data, FormatAnthropic, MessageFormat("unknown"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported target format")
	})
}

func TestConverterHelperFunctions(t *testing.T) {
	t.Run("stringToAnthropicContent", func(t *testing.T) {
		content := stringToAnthropicContent("Hello world")
		assert.Len(t, content, 1)
		assert.Equal(t, "text", content[0].Type)
		assert.Equal(t, "Hello world", content[0].Text)
	})

	t.Run("anthropicContentToString", func(t *testing.T) {
		content := []AnthropicContent{
			{Type: "text", Text: "Hello"},
			{Type: "text", Text: "world"},
			{Type: "image", Text: ""}, // Should be ignored
		}
		result := anthropicContentToString(content)
		assert.Equal(t, "Hello world", result)
	})
}