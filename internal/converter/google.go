package converter

import (
	"encoding/json"
	"fmt"
)

// GoogleConverter handles Google format conversions
type GoogleConverter struct{}

// NewGoogleConverter creates a new Google converter
func NewGoogleConverter() *GoogleConverter {
	return &GoogleConverter{}
}

// GoogleRequest represents Google's request format
type GoogleRequest struct {
	Contents         []GoogleContent         `json:"contents"`
	GenerationConfig *GoogleGenerationConfig `json:"generationConfig,omitempty"`
}

// GoogleContent represents Google's content format
type GoogleContent struct {
	Role  string       `json:"role"`
	Parts []GooglePart `json:"parts"`
}

// GooglePart represents a content part
type GooglePart struct {
	Text string `json:"text"`
}

// GoogleGenerationConfig represents generation configuration
type GoogleGenerationConfig struct {
	Temperature     float64  `json:"temperature,omitempty"`
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
	TopP            float64  `json:"topP,omitempty"`
	TopK            int      `json:"topK,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}

// GoogleResponse represents Google's response format
type GoogleResponse struct {
	Candidates    []GoogleCandidate `json:"candidates"`
	UsageMetadata *GoogleUsage      `json:"usageMetadata,omitempty"`
}

// GoogleCandidate represents a response candidate
type GoogleCandidate struct {
	Content       GoogleContent `json:"content"`
	FinishReason  string        `json:"finishReason"`
	SafetyRatings []interface{} `json:"safetyRatings,omitempty"`
}

// GoogleUsage represents Google's usage format
type GoogleUsage struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// ToGeneric converts Google format to generic format
func (gc *GoogleConverter) ToGeneric(data json.RawMessage, isRequest bool) (json.RawMessage, error) {
	if isRequest {
		var req GoogleRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to unmarshal Google request: %w", err)
		}

		// Convert contents to messages
		messages := make([]Message, len(req.Contents))
		for i, content := range req.Contents {
			// Combine parts into single text
			var text string
			for _, part := range content.Parts {
				text += part.Text
			}

			// Convert content to JSON
			contentJSON, err := json.Marshal(text)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal content: %w", err)
			}

			// Map Google roles to generic roles
			role := content.Role
			if role == "model" {
				role = "assistant"
			}

			messages[i] = Message{
				Role:    role,
				Content: contentJSON,
			}
		}

		// Create generic request
		genericReq := Request{
			Messages: messages,
		}

		if req.GenerationConfig != nil {
			genericReq.MaxTokens = req.GenerationConfig.MaxOutputTokens
			genericReq.Temperature = req.GenerationConfig.Temperature
		}

		return json.Marshal(genericReq)
	}

	// Handle response
	var resp GoogleResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Google response: %w", err)
	}

	// Get first candidate
	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in Google response")
	}

	candidate := resp.Candidates[0]

	// Extract text from parts
	var text string
	for _, part := range candidate.Content.Parts {
		text += part.Text
	}

	// Convert content
	content, err := json.Marshal(text)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal content: %w", err)
	}

	// Map role
	role := candidate.Content.Role
	if role == "model" {
		role = "assistant"
	}

	// Create generic response
	genericResp := Response{
		Type:    "message",
		Role:    role,
		Content: content,
	}

	if resp.UsageMetadata != nil {
		genericResp.Usage = &Usage{
			InputTokens:  resp.UsageMetadata.PromptTokenCount,
			OutputTokens: resp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:  resp.UsageMetadata.TotalTokenCount,
		}
	}

	return json.Marshal(genericResp)
}

// FromGeneric converts generic format to Google format
func (gc *GoogleConverter) FromGeneric(data json.RawMessage, isRequest bool) (json.RawMessage, error) {
	if isRequest {
		var req Request
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to unmarshal generic request: %w", err)
		}

		// Convert messages to contents
		contents := make([]GoogleContent, 0, len(req.Messages))

		// Add system message as first user message if present
		if req.System != "" {
			contents = append(contents, GoogleContent{
				Role:  "user",
				Parts: []GooglePart{{Text: req.System}},
			})
		}

		for _, msg := range req.Messages {
			// Parse content
			var text string
			if err := json.Unmarshal(msg.Content, &text); err != nil {
				// Try as object
				var contentObj interface{}
				if err := json.Unmarshal(msg.Content, &contentObj); err == nil {
					text = fmt.Sprintf("%v", contentObj)
				} else {
					return nil, fmt.Errorf("failed to parse content: %w", err)
				}
			}

			// Map role
			role := msg.Role
			if role == "assistant" {
				role = "model"
			}

			contents = append(contents, GoogleContent{
				Role:  role,
				Parts: []GooglePart{{Text: text}},
			})
		}

		// Create Google request
		googleReq := GoogleRequest{
			Contents: contents,
		}

		if req.MaxTokens > 0 || req.Temperature > 0 {
			googleReq.GenerationConfig = &GoogleGenerationConfig{
				MaxOutputTokens: req.MaxTokens,
				Temperature:     req.Temperature,
			}
		}

		return json.Marshal(googleReq)
	}

	// Handle response
	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal generic response: %w", err)
	}

	// Parse content
	var text string
	if err := json.Unmarshal(resp.Content, &text); err != nil {
		return nil, fmt.Errorf("failed to parse content: %w", err)
	}

	// Map role
	role := resp.Role
	if role == "assistant" {
		role = "model"
	}

	// Create Google response
	googleResp := GoogleResponse{
		Candidates: []GoogleCandidate{
			{
				Content: GoogleContent{
					Role:  role,
					Parts: []GooglePart{{Text: text}},
				},
				FinishReason: "STOP",
			},
		},
	}

	if resp.Usage != nil {
		googleResp.UsageMetadata = &GoogleUsage{
			PromptTokenCount:     resp.Usage.InputTokens,
			CandidatesTokenCount: resp.Usage.OutputTokens,
			TotalTokenCount:      resp.Usage.TotalTokens,
		}
	}

	return json.Marshal(googleResp)
}

// ConvertStreamEvent converts stream events
func (gc *GoogleConverter) ConvertStreamEvent(data []byte, toFormat MessageFormat) ([]byte, error) {
	// For now, just pass through stream events
	// In a full implementation, this would parse and convert SSE events
	return data, nil
}
