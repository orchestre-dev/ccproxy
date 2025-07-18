package converter

import (
	"testing"

	"ccproxy/internal/models"
)

//nolint:gocognit // Test function needs comprehensive test cases
func TestConvertAnthropicToOpenAI(t *testing.T) {
	tests := []struct {
		input    *models.MessagesRequest
		expected *models.ChatCompletionRequest
		name     string
	}{
		{
			name: "simple text message",
			input: &models.MessagesRequest{
				Model: "claude-3-sonnet",
				Messages: []models.Message{
					{
						Role:    "user",
						Content: "Hello, how are you?",
					},
				},
				MaxTokens:   intPtr(100),
				Temperature: floatPtr(0.7),
			},
			expected: &models.ChatCompletionRequest{
				Model: "claude-3-sonnet",
				Messages: []models.ChatMessage{
					{
						Role:    "user",
						Content: "Hello, how are you?",
					},
				},
				MaxTokens:   intPtr(100),
				Temperature: floatPtr(0.7),
			},
		},
		{
			name: "complex content with tool use",
			input: &models.MessagesRequest{
				Model: "claude-3-sonnet",
				Messages: []models.Message{
					{
						Role: "user",
						Content: []models.Content{
							{
								Type: "text",
								Text: "What's the weather like?",
							},
						},
					},
					{
						Role: "assistant",
						Content: []models.Content{
							{
								Type: "tool_use",
								ID:   "call_123",
								Name: "get_weather",
								Input: map[string]interface{}{
									"location": "New York",
								},
							},
						},
					},
					{
						Role: "user",
						Content: []models.Content{
							{
								Type:      "tool_result",
								ToolUseID: "call_123",
								Content:   map[string]interface{}{"temperature": 72, "conditions": "sunny"},
							},
						},
					},
				},
			},
			expected: &models.ChatCompletionRequest{
				Model: "claude-3-sonnet",
				Messages: []models.ChatMessage{
					{
						Role:    "user",
						Content: "What's the weather like?",
					},
					{
						Role:    "assistant",
						Content: "[Tool Use: get_weather] {\"location\":\"New York\"}",
					},
					{
						Role:    "user",
						Content: "<tool_result>{\"conditions\":\"sunny\",\"temperature\":72}</tool_result>",
					},
				},
			},
		},
		{
			name: "with tools",
			input: &models.MessagesRequest{
				Model: "claude-3-sonnet",
				Messages: []models.Message{
					{
						Role:    "user",
						Content: "Get the weather",
					},
				},
				Tools: []models.Tool{
					{
						Name:        "get_weather",
						Description: stringPtr("Get current weather"),
						InputSchema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"location": map[string]interface{}{
									"type": "string",
								},
							},
						},
					},
				},
				ToolChoice: "auto",
			},
			expected: &models.ChatCompletionRequest{
				Model: "claude-3-sonnet",
				Messages: []models.ChatMessage{
					{
						Role:    "user",
						Content: "Get the weather",
					},
				},
				Tools: []models.ChatCompletionTool{
					{
						Type: "function",
						Function: models.ChatCompletionToolFunction{
							Name:        "get_weather",
							Description: "Get current weather",
							Parameters: map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"location": map[string]interface{}{
										"type": "string",
									},
								},
							},
						},
					},
				},
				ToolChoice: "auto",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertAnthropicToOpenAI(tt.input)
			if err != nil {
				t.Fatalf("ConvertAnthropicToOpenAI() error = %v", err)
			}

			// Compare basic fields
			if result.Model != tt.expected.Model {
				t.Errorf("Model = %v, expected %v", result.Model, tt.expected.Model)
			}

			if len(result.Messages) != len(tt.expected.Messages) {
				t.Errorf("Messages length = %v, expected %v", len(result.Messages), len(tt.expected.Messages))
			}

			// Compare messages
			for i, msg := range result.Messages {
				if i >= len(tt.expected.Messages) {
					break
				}
				expectedMsg := tt.expected.Messages[i]
				if msg.Role != expectedMsg.Role {
					t.Errorf("Message[%d].Role = %v, expected %v", i, msg.Role, expectedMsg.Role)
				}
				if msg.Content != expectedMsg.Content {
					t.Errorf("Message[%d].Content = %v, expected %v", i, msg.Content, expectedMsg.Content)
				}
			}

			// Compare tools if present
			if len(tt.expected.Tools) > 0 {
				if len(result.Tools) != len(tt.expected.Tools) {
					t.Errorf("Tools length = %v, expected %v", len(result.Tools), len(tt.expected.Tools))
				}
				for i, tool := range result.Tools {
					if i >= len(tt.expected.Tools) {
						break
					}
					expectedTool := tt.expected.Tools[i]
					if tool.Function.Name != expectedTool.Function.Name {
						t.Errorf("Tool[%d].Function.Name = %v, expected %v", i, tool.Function.Name, expectedTool.Function.Name)
					}
				}
			}
		})
	}
}

//nolint:gocognit,gocyclo // Test function needs comprehensive test cases
func TestConvertOpenAIToAnthropic(t *testing.T) {
	tests := []struct {
		input        *models.ChatCompletionResponse
		expected     *models.MessagesResponse
		name         string
		requestID    string
		providerName string
	}{
		{
			name: "simple text response",
			input: &models.ChatCompletionResponse{
				ID:      "chatcmpl-123",
				Model:   "gpt-4",
				Created: 1234567890,
				Choices: []models.ChatCompletionChoice{
					{
						Index: 0,
						Message: models.ChatMessage{
							Role:    "assistant",
							Content: "Hello! I'm doing well, thank you for asking.",
						},
						FinishReason: "stop",
					},
				},
				Usage: models.ChatCompletionUsage{
					PromptTokens:     10,
					CompletionTokens: 15,
					TotalTokens:      25,
				},
			},
			requestID:    "msg_123",
			providerName: "openai",
			expected: &models.MessagesResponse{
				ID:    "msg_123",
				Type:  "message",
				Role:  "assistant",
				Model: "openai/gpt-4",
				Content: []models.Content{
					{
						Type: "text",
						Text: "Hello! I'm doing well, thank you for asking.",
					},
				},
				StopReason: "end_turn",
				Usage: models.Usage{
					InputTokens:  10,
					OutputTokens: 15,
				},
			},
		},
		{
			name: "tool call response",
			input: &models.ChatCompletionResponse{
				ID:      "chatcmpl-456",
				Model:   "gpt-4",
				Created: 1234567890,
				Choices: []models.ChatCompletionChoice{
					{
						Index: 0,
						Message: models.ChatMessage{
							Role: "assistant",
							ToolCalls: []models.ToolCall{
								{
									ID:   "call_abc123",
									Type: "function",
									Function: models.FunctionCall{
										Name:      "get_weather",
										Arguments: "{\"location\":\"San Francisco\"}",
									},
								},
							},
						},
						FinishReason: "tool_calls",
					},
				},
				Usage: models.ChatCompletionUsage{
					PromptTokens:     20,
					CompletionTokens: 5,
					TotalTokens:      25,
				},
			},
			requestID:    "msg_456",
			providerName: "groq",
			expected: &models.MessagesResponse{
				ID:    "msg_456",
				Type:  "message",
				Role:  "assistant",
				Model: "groq/gpt-4",
				Content: []models.Content{
					{
						Type: "tool_use",
						ID:   "call_abc123",
						Name: "get_weather",
						Input: map[string]interface{}{
							"location": "San Francisco",
						},
					},
				},
				StopReason: "tool_use",
				Usage: models.Usage{
					InputTokens:  20,
					OutputTokens: 5,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertOpenAIToAnthropic(tt.input, tt.requestID, tt.providerName)
			if err != nil {
				t.Fatalf("ConvertOpenAIToAnthropic() error = %v", err)
			}

			// Compare basic fields
			if result.ID != tt.expected.ID {
				t.Errorf("ID = %v, expected %v", result.ID, tt.expected.ID)
			}
			if result.Type != tt.expected.Type {
				t.Errorf("Type = %v, expected %v", result.Type, tt.expected.Type)
			}
			if result.Role != tt.expected.Role {
				t.Errorf("Role = %v, expected %v", result.Role, tt.expected.Role)
			}
			if result.Model != tt.expected.Model {
				t.Errorf("Model = %v, expected %v", result.Model, tt.expected.Model)
			}
			if result.StopReason != tt.expected.StopReason {
				t.Errorf("StopReason = %v, expected %v", result.StopReason, tt.expected.StopReason)
			}

			// Compare usage
			if result.Usage.InputTokens != tt.expected.Usage.InputTokens {
				t.Errorf("Usage.InputTokens = %v, expected %v", result.Usage.InputTokens, tt.expected.Usage.InputTokens)
			}
			if result.Usage.OutputTokens != tt.expected.Usage.OutputTokens {
				t.Errorf("Usage.OutputTokens = %v, expected %v", result.Usage.OutputTokens, tt.expected.Usage.OutputTokens)
			}

			// Compare content
			if len(result.Content) != len(tt.expected.Content) {
				t.Errorf("Content length = %v, expected %v", len(result.Content), len(tt.expected.Content))
			}

			for i, content := range result.Content {
				if i >= len(tt.expected.Content) {
					break
				}
				expectedContent := tt.expected.Content[i]
				if content.Type != expectedContent.Type {
					t.Errorf("Content[%d].Type = %v, expected %v", i, content.Type, expectedContent.Type)
				}
				if content.Text != expectedContent.Text {
					t.Errorf("Content[%d].Text = %v, expected %v", i, content.Text, expectedContent.Text)
				}
				if content.Name != expectedContent.Name {
					t.Errorf("Content[%d].Name = %v, expected %v", i, content.Name, expectedContent.Name)
				}
			}
		})
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}

func stringPtr(s string) *string {
	return &s
}
