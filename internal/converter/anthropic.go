package converter

import (
	"encoding/json"
	"fmt"
	"strings"
)

// AnthropicConverter handles Anthropic format conversions
type AnthropicConverter struct{}

// NewAnthropicConverter creates a new Anthropic converter
func NewAnthropicConverter() *AnthropicConverter {
	return &AnthropicConverter{}
}

// AnthropicRequest represents Anthropic's request format
type AnthropicRequest struct {
	Model       string                   `json:"model"`
	Messages    []AnthropicMessage       `json:"messages"`
	System      string                   `json:"system,omitempty"`
	MaxTokens   int                      `json:"max_tokens,omitempty"`
	Temperature float64                  `json:"temperature,omitempty"`
	Stream      bool                     `json:"stream,omitempty"`
	Metadata    map[string]interface{}   `json:"metadata,omitempty"`
}

// AnthropicMessage represents Anthropic's message format
type AnthropicMessage struct {
	Role    string                 `json:"role"`
	Content []AnthropicContent     `json:"content"`
}

// AnthropicContent represents Anthropic's content format
type AnthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// AnthropicResponse represents Anthropic's response format
type AnthropicResponse struct {
	ID      string             `json:"id"`
	Type    string             `json:"type"`
	Role    string             `json:"role"`
	Content []AnthropicContent `json:"content"`
	Model   string             `json:"model"`
	Usage   *AnthropicUsage    `json:"usage,omitempty"`
}

// AnthropicUsage represents Anthropic's usage format
type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ToGeneric converts Anthropic format to generic format
func (ac *AnthropicConverter) ToGeneric(data json.RawMessage, isRequest bool) (json.RawMessage, error) {
	if isRequest {
		var req AnthropicRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to unmarshal Anthropic request: %w", err)
		}

		// Convert messages
		messages := make([]Message, len(req.Messages))
		for i, msg := range req.Messages {
			// Convert content to generic format
			content, err := json.Marshal(msg.Content)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal content: %w", err)
			}
			
			messages[i] = Message{
				Role:    msg.Role,
				Content: content,
			}
		}

		// Create generic request
		genericReq := Request{
			Model:       req.Model,
			Messages:    messages,
			System:      req.System,
			MaxTokens:   req.MaxTokens,
			Temperature: req.Temperature,
			Stream:      req.Stream,
		}

		if req.Metadata != nil {
			metadata, err := json.Marshal(req.Metadata)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal metadata: %w", err)
			}
			genericReq.Metadata = metadata
		}

		return json.Marshal(genericReq)
	}

	// Handle response
	var resp AnthropicResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Anthropic response: %w", err)
	}

	// Convert content
	content, err := json.Marshal(resp.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal content: %w", err)
	}

	// Create generic response
	genericResp := Response{
		ID:      resp.ID,
		Type:    resp.Type,
		Role:    resp.Role,
		Content: content,
		Model:   resp.Model,
	}

	if resp.Usage != nil {
		genericResp.Usage = &Usage{
			InputTokens:  resp.Usage.InputTokens,
			OutputTokens: resp.Usage.OutputTokens,
			TotalTokens:  resp.Usage.InputTokens + resp.Usage.OutputTokens,
		}
	}

	return json.Marshal(genericResp)
}

// FromGeneric converts generic format to Anthropic format
func (ac *AnthropicConverter) FromGeneric(data json.RawMessage, isRequest bool) (json.RawMessage, error) {
	if isRequest {
		var req Request
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to unmarshal generic request: %w", err)
		}

		// Convert messages
		messages := make([]AnthropicMessage, len(req.Messages))
		for i, msg := range req.Messages {
			// Parse content
			var content []AnthropicContent
			if err := json.Unmarshal(msg.Content, &content); err != nil {
				// Try as string
				var textContent string
				if err := json.Unmarshal(msg.Content, &textContent); err == nil {
					content = []AnthropicContent{{Type: "text", Text: textContent}}
				} else {
					return nil, fmt.Errorf("failed to parse content: %w", err)
				}
			}
			
			messages[i] = AnthropicMessage{
				Role:    msg.Role,
				Content: content,
			}
		}

		// Create Anthropic request
		anthropicReq := AnthropicRequest{
			Model:       req.Model,
			Messages:    messages,
			System:      req.System,
			MaxTokens:   req.MaxTokens,
			Temperature: req.Temperature,
			Stream:      req.Stream,
		}

		if req.Metadata != nil {
			var metadata map[string]interface{}
			if err := json.Unmarshal(req.Metadata, &metadata); err == nil {
				anthropicReq.Metadata = metadata
			}
		}

		return json.Marshal(anthropicReq)
	}

	// Handle response
	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal generic response: %w", err)
	}

	// Parse content
	var content []AnthropicContent
	if err := json.Unmarshal(resp.Content, &content); err != nil {
		return nil, fmt.Errorf("failed to parse content: %w", err)
	}

	// Create Anthropic response
	anthropicResp := AnthropicResponse{
		ID:      resp.ID,
		Type:    resp.Type,
		Role:    resp.Role,
		Content: content,
		Model:   resp.Model,
	}

	if resp.Usage != nil {
		anthropicResp.Usage = &AnthropicUsage{
			InputTokens:  resp.Usage.InputTokens,
			OutputTokens: resp.Usage.OutputTokens,
		}
	}

	return json.Marshal(anthropicResp)
}

// ConvertStreamEvent converts stream events
func (ac *AnthropicConverter) ConvertStreamEvent(data []byte, toFormat MessageFormat) ([]byte, error) {
	// For now, just pass through stream events
	// In a full implementation, this would parse and convert SSE events
	return data, nil
}

// Helper function to convert string content to Anthropic format
func stringToAnthropicContent(text string) []AnthropicContent {
	return []AnthropicContent{{Type: "text", Text: text}}
}

// Helper function to convert Anthropic content to string
func anthropicContentToString(content []AnthropicContent) string {
	var texts []string
	for _, c := range content {
		if c.Type == "text" && c.Text != "" {
			texts = append(texts, c.Text)
		}
	}
	return strings.Join(texts, " ")
}