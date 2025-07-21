package converter

import (
	"encoding/json"
	"fmt"
	"strings"
)

// OpenAIConverter handles OpenAI format conversions
type OpenAIConverter struct{}

// NewOpenAIConverter creates a new OpenAI converter
func NewOpenAIConverter() *OpenAIConverter {
	return &OpenAIConverter{}
}

// OpenAIRequest represents OpenAI's request format
type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
	N           int             `json:"n,omitempty"`
	Stop        []string        `json:"stop,omitempty"`
}

// OpenAIMessage represents OpenAI's message format
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

// OpenAIResponse represents OpenAI's response format
type OpenAIResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   *OpenAIUsage   `json:"usage,omitempty"`
}

// OpenAIChoice represents a response choice
type OpenAIChoice struct {
	Index        int           `json:"index"`
	Message      OpenAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

// OpenAIUsage represents OpenAI's usage format
type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ToGeneric converts OpenAI format to generic format
func (oc *OpenAIConverter) ToGeneric(data json.RawMessage, isRequest bool) (json.RawMessage, error) {
	if isRequest {
		var req OpenAIRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to unmarshal OpenAI request: %w", err)
		}

		// Convert messages
		messages := make([]Message, 0, len(req.Messages))
		var system string

		for _, msg := range req.Messages {
			if msg.Role == "system" {
				// Extract system message
				system = msg.Content
			} else {
				// Convert content to JSON
				content, err := json.Marshal(msg.Content)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal content: %w", err)
				}

				messages = append(messages, Message{
					Role:    msg.Role,
					Content: content,
					Name:    msg.Name,
				})
			}
		}

		// Create generic request
		genericReq := Request{
			Model:       req.Model,
			Messages:    messages,
			System:      system,
			MaxTokens:   req.MaxTokens,
			Temperature: req.Temperature,
			Stream:      req.Stream,
		}

		return json.Marshal(genericReq)
	}

	// Handle response
	var resp OpenAIResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OpenAI response: %w", err)
	}

	// Get first choice (most common case)
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in OpenAI response")
	}

	choice := resp.Choices[0]

	// Convert content to Anthropic-style content array
	content, err := json.Marshal([]map[string]interface{}{
		{
			"type": "text",
			"text": choice.Message.Content,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal content: %w", err)
	}

	// Create generic response
	genericResp := Response{
		ID:      resp.ID,
		Type:    "message",
		Role:    choice.Message.Role,
		Content: content,
		Model:   resp.Model,
	}

	if resp.Usage != nil {
		genericResp.Usage = &Usage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
			TotalTokens:  resp.Usage.TotalTokens,
		}
	}

	return json.Marshal(genericResp)
}

// FromGeneric converts generic format to OpenAI format
func (oc *OpenAIConverter) FromGeneric(data json.RawMessage, isRequest bool) (json.RawMessage, error) {
	if isRequest {
		var req Request
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to unmarshal generic request: %w", err)
		}

		// Convert messages
		messages := make([]OpenAIMessage, 0, len(req.Messages)+1)

		// Add system message if present
		if req.System != "" {
			messages = append(messages, OpenAIMessage{
				Role:    "system",
				Content: req.System,
			})
		}

		// Add other messages
		for _, msg := range req.Messages {
			// Parse content - handle both string and Anthropic content array format
			var content string

			// First try as string
			if err := json.Unmarshal(msg.Content, &content); err == nil {
				// It's already a string
			} else {
				// Try as Anthropic content array
				var contentArray []map[string]interface{}
				if err := json.Unmarshal(msg.Content, &contentArray); err == nil {
					// Extract text from content array
					var texts []string
					for _, part := range contentArray {
						if part["type"] == "text" {
							if text, ok := part["text"].(string); ok {
								texts = append(texts, text)
							}
						}
					}
					content = strings.Join(texts, " ")
				} else {
					// Try as generic object
					var contentObj interface{}
					if err := json.Unmarshal(msg.Content, &contentObj); err == nil {
						content = fmt.Sprintf("%v", contentObj)
					} else {
						return nil, fmt.Errorf("failed to parse content: %w", err)
					}
				}
			}

			messages = append(messages, OpenAIMessage{
				Role:    msg.Role,
				Content: content,
				Name:    msg.Name,
			})
		}

		// Create OpenAI request
		openAIReq := OpenAIRequest{
			Model:       req.Model,
			Messages:    messages,
			MaxTokens:   req.MaxTokens,
			Temperature: req.Temperature,
			Stream:      req.Stream,
		}

		return json.Marshal(openAIReq)
	}

	// Handle response
	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal generic response: %w", err)
	}

	// Parse content - handle both string and Anthropic content array
	var content string

	// First try as string
	if err := json.Unmarshal(resp.Content, &content); err != nil {
		// Try as Anthropic content array
		var contentArray []map[string]interface{}
		if err := json.Unmarshal(resp.Content, &contentArray); err == nil {
			// Extract text from content array
			var texts []string
			for _, part := range contentArray {
				if part["type"] == "text" {
					if text, ok := part["text"].(string); ok {
						texts = append(texts, text)
					}
				}
			}
			content = strings.Join(texts, " ")
		} else {
			return nil, fmt.Errorf("failed to parse content: %w", err)
		}
	}

	// Create OpenAI response
	openAIResp := OpenAIResponse{
		ID:     resp.ID,
		Object: "chat.completion",
		Model:  resp.Model,
		Choices: []OpenAIChoice{
			{
				Index: 0,
				Message: OpenAIMessage{
					Role:    resp.Role,
					Content: content,
				},
				FinishReason: "stop",
			},
		},
	}

	if resp.Usage != nil {
		openAIResp.Usage = &OpenAIUsage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
	}

	return json.Marshal(openAIResp)
}

// ConvertStreamEvent converts stream events
func (oc *OpenAIConverter) ConvertStreamEvent(data []byte, toFormat MessageFormat) ([]byte, error) {
	// For now, just pass through stream events
	// In a full implementation, this would parse and convert SSE events
	return data, nil
}
