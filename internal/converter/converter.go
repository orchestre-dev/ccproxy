// Package converter provides conversion utilities between different API formats
package converter

import (
	"encoding/json"
	"fmt"
	"strings"

	"ccproxy/internal/models"
)

// ConvertAnthropicToOpenAI converts Anthropic messages request to OpenAI format
func ConvertAnthropicToOpenAI(req *models.MessagesRequest) (*models.ChatCompletionRequest, error) {
	chatReq := &models.ChatCompletionRequest{
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}

	// Convert messages
	messages, err := convertMessages(req.Messages)
	if err != nil {
		return nil, fmt.Errorf("failed to convert messages: %w", err)
	}
	chatReq.Messages = messages

	// Convert tools
	if req.Tools != nil {
		tools := convertTools(req.Tools)
		chatReq.Tools = tools
		chatReq.ToolChoice = req.ToolChoice
	}

	return chatReq, nil
}

// ConvertOpenAIToAnthropic converts OpenAI response to Anthropic format
func ConvertOpenAIToAnthropic(
	resp *models.ChatCompletionResponse,
	requestID, providerName string,
) (*models.MessagesResponse, error) {
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := resp.Choices[0]
	message := choice.Message

	var content []models.Content
	var stopReason string

	// Handle tool calls
	if len(message.ToolCalls) > 0 {
		for _, toolCall := range message.ToolCalls {
			var input map[string]interface{}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &input); err != nil {
				return nil, fmt.Errorf("failed to parse tool arguments: %w", err)
			}

			content = append(content, models.Content{
				Type:  "tool_use",
				ID:    toolCall.ID,
				Name:  toolCall.Function.Name,
				Input: input,
			})
		}
		stopReason = "tool_use"
	} else {
		// Regular text response
		content = append(content, models.Content{
			Type: "text",
			Text: message.Content,
		})
		stopReason = "end_turn"
	}

	return &models.MessagesResponse{
		ID:         requestID,
		Type:       "message",
		Role:       "assistant",
		Model:      fmt.Sprintf("%s/%s", providerName, resp.Model),
		Content:    content,
		StopReason: stopReason,
		Usage: models.Usage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
		},
	}, nil
}

// convertMessages converts Anthropic messages to OpenAI format
func convertMessages(messages []models.Message) ([]models.ChatMessage, error) {
	var chatMessages []models.ChatMessage

	for _, msg := range messages {
		chatMsg := models.ChatMessage{
			Role: msg.Role,
		}

		switch content := msg.Content.(type) {
		case string:
			chatMsg.Content = content
		case []models.Content:
			var parts []string
			for _, block := range content {
				switch block.Type {
				case "text":
					parts = append(parts, block.Text)
				case "tool_use":
					toolInfo := fmt.Sprintf("[Tool Use: %s] %s", block.Name, mustMarshalJSON(block.Input))
					parts = append(parts, toolInfo)
				case "tool_result":
					result := fmt.Sprintf("<tool_result>%s</tool_result>", mustMarshalJSON(block.Content))
					parts = append(parts, result)
				}
			}
			chatMsg.Content = strings.Join(parts, "\n")
		default:
			return nil, fmt.Errorf("unsupported content type: %T", content)
		}

		chatMessages = append(chatMessages, chatMsg)
	}

	return chatMessages, nil
}

// convertTools converts Anthropic tools to OpenAI format
func convertTools(tools []models.Tool) []models.ChatCompletionTool {
	var chatTools []models.ChatCompletionTool

	for _, tool := range tools {
		description := ""
		if tool.Description != nil {
			description = *tool.Description
		}

		chatTool := models.ChatCompletionTool{
			Type: "function",
			Function: models.ChatCompletionToolFunction{
				Name:        tool.Name,
				Description: description,
				Parameters:  tool.InputSchema,
			},
		}

		chatTools = append(chatTools, chatTool)
	}

	return chatTools
}

// mustMarshalJSON marshals to JSON or returns empty string on error
func mustMarshalJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(data)
}
