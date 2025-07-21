package converter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageConverter_ConvertRequest(t *testing.T) {
	converter := NewMessageConverter()

	t.Run("Anthropic to OpenAI", func(t *testing.T) {
		// Create Anthropic request
		anthropicReq := AnthropicRequest{
			Model: "claude-3-opus-20240229",
			Messages: []AnthropicMessage{
				{
					Role: "user",
					Content: []AnthropicContent{
						{Type: "text", Text: "Hello, how are you?"},
					},
				},
			},
			System:    "You are a helpful assistant",
			MaxTokens: 100,
		}

		data, err := json.Marshal(anthropicReq)
		require.NoError(t, err)

		// Convert to OpenAI format
		result, err := converter.ConvertRequest(data, FormatAnthropic, FormatOpenAI)
		require.NoError(t, err)

		// Verify OpenAI format
		var openAIReq OpenAIRequest
		err = json.Unmarshal(result, &openAIReq)
		require.NoError(t, err)

		assert.Equal(t, "claude-3-opus-20240229", openAIReq.Model)
		assert.Equal(t, 100, openAIReq.MaxTokens)
		assert.Len(t, openAIReq.Messages, 2) // System + user message
		assert.Equal(t, "system", openAIReq.Messages[0].Role)
		assert.Equal(t, "You are a helpful assistant", openAIReq.Messages[0].Content)
		assert.Equal(t, "user", openAIReq.Messages[1].Role)
		assert.Equal(t, "Hello, how are you?", openAIReq.Messages[1].Content)
	})

	t.Run("OpenAI to Anthropic", func(t *testing.T) {
		// Create OpenAI request
		openAIReq := OpenAIRequest{
			Model: "gpt-4",
			Messages: []OpenAIMessage{
				{Role: "system", Content: "You are a helpful assistant"},
				{Role: "user", Content: "What is 2+2?"},
				{Role: "assistant", Content: "2+2 equals 4"},
				{Role: "user", Content: "What about 3+3?"},
			},
			MaxTokens: 150,
		}

		data, err := json.Marshal(openAIReq)
		require.NoError(t, err)

		// Convert to Anthropic format
		result, err := converter.ConvertRequest(data, FormatOpenAI, FormatAnthropic)
		require.NoError(t, err)

		// Verify Anthropic format
		var anthropicReq AnthropicRequest
		err = json.Unmarshal(result, &anthropicReq)
		require.NoError(t, err)

		assert.Equal(t, "gpt-4", anthropicReq.Model)
		assert.Equal(t, "You are a helpful assistant", anthropicReq.System)
		assert.Equal(t, 150, anthropicReq.MaxTokens)
		assert.Len(t, anthropicReq.Messages, 3) // No system message in messages array
		assert.Equal(t, "user", anthropicReq.Messages[0].Role)
		assert.Equal(t, "assistant", anthropicReq.Messages[1].Role)
		assert.Equal(t, "user", anthropicReq.Messages[2].Role)
	})

	t.Run("Same format returns unchanged", func(t *testing.T) {
		original := json.RawMessage(`{"model":"test","messages":[]}`)
		result, err := converter.ConvertRequest(original, FormatAnthropic, FormatAnthropic)
		require.NoError(t, err)
		assert.Equal(t, original, result)
	})

	t.Run("Unsupported format error", func(t *testing.T) {
		data := json.RawMessage(`{}`)
		_, err := converter.ConvertRequest(data, MessageFormat("unknown"), FormatOpenAI)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported source format")
	})
}

func TestMessageConverter_ConvertResponse(t *testing.T) {
	converter := NewMessageConverter()

	t.Run("Anthropic to OpenAI Response", func(t *testing.T) {
		// Create Anthropic response
		anthropicResp := AnthropicResponse{
			ID:   "msg_123",
			Type: "message",
			Role: "assistant",
			Content: []AnthropicContent{
				{Type: "text", Text: "Hello! I'm doing well, thank you."},
			},
			Model: "claude-3-opus-20240229",
			Usage: &AnthropicUsage{
				InputTokens:  10,
				OutputTokens: 8,
			},
		}

		data, err := json.Marshal(anthropicResp)
		require.NoError(t, err)

		// Convert to OpenAI format
		result, err := converter.ConvertResponse(data, FormatAnthropic, FormatOpenAI)
		require.NoError(t, err)

		// Verify OpenAI format
		var openAIResp OpenAIResponse
		err = json.Unmarshal(result, &openAIResp)
		require.NoError(t, err)

		assert.Equal(t, "msg_123", openAIResp.ID)
		assert.Equal(t, "claude-3-opus-20240229", openAIResp.Model)
		assert.Len(t, openAIResp.Choices, 1)
		assert.Equal(t, "assistant", openAIResp.Choices[0].Message.Role)
		assert.Equal(t, "Hello! I'm doing well, thank you.", openAIResp.Choices[0].Message.Content)
		assert.Equal(t, 10, openAIResp.Usage.PromptTokens)
		assert.Equal(t, 8, openAIResp.Usage.CompletionTokens)
		assert.Equal(t, 18, openAIResp.Usage.TotalTokens)
	})

	t.Run("OpenAI to Anthropic Response", func(t *testing.T) {
		// Create OpenAI response
		openAIResp := OpenAIResponse{
			ID:     "chatcmpl-123",
			Object: "chat.completion",
			Model:  "gpt-4",
			Choices: []OpenAIChoice{
				{
					Index: 0,
					Message: OpenAIMessage{
						Role:    "assistant",
						Content: "The answer is 42",
					},
					FinishReason: "stop",
				},
			},
			Usage: &OpenAIUsage{
				PromptTokens:     20,
				CompletionTokens: 5,
				TotalTokens:      25,
			},
		}

		data, err := json.Marshal(openAIResp)
		require.NoError(t, err)

		// Convert to Anthropic format
		result, err := converter.ConvertResponse(data, FormatOpenAI, FormatAnthropic)
		require.NoError(t, err)

		// Verify Anthropic format
		var anthropicResp AnthropicResponse
		err = json.Unmarshal(result, &anthropicResp)
		require.NoError(t, err)

		assert.Equal(t, "chatcmpl-123", anthropicResp.ID)
		assert.Equal(t, "gpt-4", anthropicResp.Model)
		assert.Equal(t, "assistant", anthropicResp.Role)
		assert.Len(t, anthropicResp.Content, 1)
		assert.Equal(t, "text", anthropicResp.Content[0].Type)
		assert.Equal(t, "The answer is 42", anthropicResp.Content[0].Text)
		assert.Equal(t, 20, anthropicResp.Usage.InputTokens)
		assert.Equal(t, 5, anthropicResp.Usage.OutputTokens)
	})
}

func TestAnthropicConverter(t *testing.T) {
	conv := NewAnthropicConverter()

	t.Run("ToGeneric Request", func(t *testing.T) {
		req := AnthropicRequest{
			Model: "claude-3",
			Messages: []AnthropicMessage{
				{
					Role: "user",
					Content: []AnthropicContent{
						{Type: "text", Text: "Hello"},
					},
				},
			},
			System: "Be helpful",
		}

		data, _ := json.Marshal(req)
		result, err := conv.ToGeneric(data, true)
		require.NoError(t, err)

		var genericReq Request
		err = json.Unmarshal(result, &genericReq)
		require.NoError(t, err)

		assert.Equal(t, "claude-3", genericReq.Model)
		assert.Equal(t, "Be helpful", genericReq.System)
		assert.Len(t, genericReq.Messages, 1)
	})

	t.Run("FromGeneric Request", func(t *testing.T) {
		genericReq := Request{
			Model: "claude-3",
			Messages: []Message{
				{
					Role:    "user",
					Content: json.RawMessage(`"Hello world"`),
				},
			},
			System: "Be helpful",
		}

		data, _ := json.Marshal(genericReq)
		result, err := conv.FromGeneric(data, true)
		require.NoError(t, err)

		var anthropicReq AnthropicRequest
		err = json.Unmarshal(result, &anthropicReq)
		require.NoError(t, err)

		assert.Equal(t, "claude-3", anthropicReq.Model)
		assert.Equal(t, "Be helpful", anthropicReq.System)
		assert.Len(t, anthropicReq.Messages, 1)
		assert.Len(t, anthropicReq.Messages[0].Content, 1)
		assert.Equal(t, "Hello world", anthropicReq.Messages[0].Content[0].Text)
	})
}

func TestOpenAIConverter(t *testing.T) {
	conv := NewOpenAIConverter()

	t.Run("ToGeneric with System Message", func(t *testing.T) {
		req := OpenAIRequest{
			Model: "gpt-4",
			Messages: []OpenAIMessage{
				{Role: "system", Content: "You are helpful"},
				{Role: "user", Content: "Hi"},
			},
		}

		data, _ := json.Marshal(req)
		result, err := conv.ToGeneric(data, true)
		require.NoError(t, err)

		var genericReq Request
		err = json.Unmarshal(result, &genericReq)
		require.NoError(t, err)

		assert.Equal(t, "You are helpful", genericReq.System)
		assert.Len(t, genericReq.Messages, 1) // Only user message
		assert.Equal(t, "user", genericReq.Messages[0].Role)
	})
}

func TestGoogleConverter(t *testing.T) {
	conv := NewGoogleConverter()

	t.Run("Role Mapping", func(t *testing.T) {
		req := GoogleRequest{
			Contents: []GoogleContent{
				{
					Role:  "user",
					Parts: []GooglePart{{Text: "Hello"}},
				},
				{
					Role:  "model",
					Parts: []GooglePart{{Text: "Hi there"}},
				},
			},
		}

		data, _ := json.Marshal(req)
		result, err := conv.ToGeneric(data, true)
		require.NoError(t, err)

		var genericReq Request
		err = json.Unmarshal(result, &genericReq)
		require.NoError(t, err)

		assert.Len(t, genericReq.Messages, 2)
		assert.Equal(t, "user", genericReq.Messages[0].Role)
		assert.Equal(t, "assistant", genericReq.Messages[1].Role) // model -> assistant
	})
}

func TestAWSConverter(t *testing.T) {
	conv := NewAWSConverter()

	t.Run("Default Max Tokens", func(t *testing.T) {
		genericReq := Request{
			Model: "anthropic.claude-v2",
			Messages: []Message{
				{
					Role:    "user",
					Content: json.RawMessage(`"Hello"`),
				},
			},
		}

		data, _ := json.Marshal(genericReq)
		result, err := conv.FromGeneric(data, true)
		require.NoError(t, err)

		var awsReq AWSRequest
		err = json.Unmarshal(result, &awsReq)
		require.NoError(t, err)

		assert.Equal(t, 4096, awsReq.MaxTokens) // Default value
		assert.Equal(t, "bedrock-2023-05-31", awsReq.AnthropicVersion)
	})
}