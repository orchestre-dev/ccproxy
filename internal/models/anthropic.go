// Package models defines data structures for API requests and responses
package models

import "encoding/json"

// MessagesRequest represents an Anthropic API messages request
type MessagesRequest struct {
	ToolChoice  any       `json:"tool_choice,omitempty"`
	MaxTokens   *int      `json:"max_tokens,omitempty"`
	Temperature *float64  `json:"temperature,omitempty"`
	Stream      *bool     `json:"stream,omitempty"`
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Tools       []Tool    `json:"tools,omitempty"`
}

// MessagesResponse represents an Anthropic API messages response
type MessagesResponse struct {
	StopSequence *string   `json:"stop_sequence"`
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Role         string    `json:"role"`
	Model        string    `json:"model"`
	StopReason   string    `json:"stop_reason"`
	Content      []Content `json:"content"`
	Usage        Usage     `json:"usage"`
}

// Message represents a conversation message
type Message struct {
	Content any    `json:"content"`
	Role    string `json:"role"`
}

// Content represents a content block in a message
type Content struct {
	Type      string                 `json:"type"`
	Text      string                 `json:"text,omitempty"`
	ID        string                 `json:"id,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Input     map[string]interface{} `json:"input,omitempty"`
	Content   interface{}            `json:"content,omitempty"` // For tool_result blocks
	ToolUseID string                 `json:"tool_use_id,omitempty"`
}

// Tool represents a tool definition
type Tool struct {
	Description *string                `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"input_schema"`
	Name        string                 `json:"name"`
}

// Usage represents token usage information
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// UnmarshalJSON custom unmarshaling for Message.Content
func (m *Message) UnmarshalJSON(data []byte) error {
	type Alias Message
	aux := &struct {
		*Alias
		Content json.RawMessage `json:"content"`
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Try to unmarshal as string first
	var contentStr string
	if err := json.Unmarshal(aux.Content, &contentStr); err == nil {
		m.Content = contentStr
		return nil
	}

	// Try to unmarshal as array of Content
	var contentArray []Content
	if err := json.Unmarshal(aux.Content, &contentArray); err == nil {
		m.Content = contentArray
		return nil
	}

	return nil
}
