package converter

import (
	"encoding/json"
)

// MessageFormat represents different message formats used by LLM providers
type MessageFormat string

const (
	// FormatAnthropic represents Anthropic's message format
	FormatAnthropic MessageFormat = "anthropic"
	// FormatOpenAI represents OpenAI's message format
	FormatOpenAI MessageFormat = "openai"
	// FormatGoogle represents Google's message format
	FormatGoogle MessageFormat = "google"
	// FormatAWS represents AWS Bedrock's message format
	FormatAWS MessageFormat = "aws"
	// FormatGeneric represents a generic/standard format
	FormatGeneric MessageFormat = "generic"
)

// Message represents a generic message structure
type Message struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
	Name    string          `json:"name,omitempty"`
}

// Request represents a generic request structure
type Request struct {
	Model       string          `json:"model"`
	Messages    []Message       `json:"messages"`
	System      string          `json:"system,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
}

// Response represents a generic response structure
type Response struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
	Model   string          `json:"model"`
	Usage   *Usage          `json:"usage,omitempty"`
}

// Usage represents token usage information
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// ContentPart represents a part of message content
type ContentPart struct {
	Type string          `json:"type"`
	Text string          `json:"text,omitempty"`
	Data json.RawMessage `json:"data,omitempty"`
}

// StreamEvent represents a server-sent event
type StreamEvent struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}