package converter

import (
	"encoding/json"
	"fmt"
)

// AWSConverter handles AWS Bedrock format conversions
type AWSConverter struct{}

// NewAWSConverter creates a new AWS converter
func NewAWSConverter() *AWSConverter {
	return &AWSConverter{}
}

// AWSRequest represents AWS Bedrock's request format
type AWSRequest struct {
	AnthropicVersion string       `json:"anthropic_version"`
	Messages         []AWSMessage `json:"messages"`
	System           string       `json:"system,omitempty"`
	MaxTokens        int          `json:"max_tokens"`
	Temperature      float64      `json:"temperature,omitempty"`
	TopP             float64      `json:"top_p,omitempty"`
	TopK             int          `json:"top_k,omitempty"`
	StopSequences    []string     `json:"stop_sequences,omitempty"`
}

// AWSMessage represents AWS Bedrock's message format
type AWSMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

// AWSResponse represents AWS Bedrock's response format
type AWSResponse struct {
	ID           string          `json:"id"`
	Model        string          `json:"model"`
	Type         string          `json:"type"`
	Role         string          `json:"role"`
	Content      json.RawMessage `json:"content"`
	StopReason   string          `json:"stop_reason,omitempty"`
	StopSequence string          `json:"stop_sequence,omitempty"`
	Usage        *AWSUsage       `json:"usage,omitempty"`
}

// AWSUsage represents AWS Bedrock's usage format
type AWSUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ToGeneric converts AWS format to generic format
func (ac *AWSConverter) ToGeneric(data json.RawMessage, isRequest bool) (json.RawMessage, error) {
	if isRequest {
		var req AWSRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to unmarshal AWS request: %w", err)
		}

		// Convert messages
		messages := make([]Message, len(req.Messages))
		for i, msg := range req.Messages {
			messages[i] = Message{
				Role:    msg.Role,
				Content: msg.Content,
			}
		}

		// Create generic request
		genericReq := Request{
			Messages:    messages,
			System:      req.System,
			MaxTokens:   req.MaxTokens,
			Temperature: req.Temperature,
		}

		return json.Marshal(genericReq)
	}

	// Handle response
	var resp AWSResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal AWS response: %w", err)
	}

	// Create generic response
	genericResp := Response{
		ID:      resp.ID,
		Type:    resp.Type,
		Role:    resp.Role,
		Content: resp.Content,
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

// FromGeneric converts generic format to AWS format
func (ac *AWSConverter) FromGeneric(data json.RawMessage, isRequest bool) (json.RawMessage, error) {
	if isRequest {
		var req Request
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to unmarshal generic request: %w", err)
		}

		// Convert messages
		messages := make([]AWSMessage, len(req.Messages))
		for i, msg := range req.Messages {
			messages[i] = AWSMessage{
				Role:    msg.Role,
				Content: msg.Content,
			}
		}

		// Create AWS request
		awsReq := AWSRequest{
			AnthropicVersion: "bedrock-2023-05-31",
			Messages:         messages,
			System:           req.System,
			MaxTokens:        req.MaxTokens,
			Temperature:      req.Temperature,
		}

		// Default max tokens if not specified
		if awsReq.MaxTokens == 0 {
			awsReq.MaxTokens = 4096
		}

		return json.Marshal(awsReq)
	}

	// Handle response
	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal generic response: %w", err)
	}

	// Create AWS response
	awsResp := AWSResponse{
		ID:         resp.ID,
		Model:      resp.Model,
		Type:       resp.Type,
		Role:       resp.Role,
		Content:    resp.Content,
		StopReason: "stop_sequence",
	}

	if resp.Usage != nil {
		awsResp.Usage = &AWSUsage{
			InputTokens:  resp.Usage.InputTokens,
			OutputTokens: resp.Usage.OutputTokens,
		}
	}

	return json.Marshal(awsResp)
}

// ConvertStreamEvent converts stream events
func (ac *AWSConverter) ConvertStreamEvent(data []byte, toFormat MessageFormat) ([]byte, error) {
	// For now, just pass through stream events
	// In a full implementation, this would parse and convert SSE events
	return data, nil
}
