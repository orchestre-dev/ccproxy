// Package gemini implements the Gemini provider for CCProxy.
package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"ccproxy/internal/config"
	"ccproxy/internal/models"
	"ccproxy/internal/provider/common"
	"ccproxy/pkg/logger"
)

const (
	providerName = "gemini"
)

// Provider implements the provider interface for Google Gemini API
type Provider struct {
	httpClient *http.Client
	config     *config.GeminiConfig
	logger     *logger.Logger
}

// NewProvider creates a new Gemini provider instance
func NewProvider(cfg *config.GeminiConfig, logger *logger.Logger) (*Provider, error) {
	if cfg == nil {
		return nil, common.NewConfigError("gemini", "config", "config cannot be nil")
	}

	return &Provider{
		httpClient: common.NewConfiguredHTTPClient(cfg.Timeout),
		config:     cfg,
		logger:     logger,
	}, nil
}

// CreateChatCompletion sends a chat completion request to Google Gemini API
func (p *Provider) CreateChatCompletion(
	ctx context.Context,
	req *models.ChatCompletionRequest,
) (*models.ChatCompletionResponse, error) {
	// Convert OpenAI format to Gemini format
	geminiReq := p.convertToRequest(req)

	// Marshal request
	reqBody, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, common.NewProviderError("gemini", "failed to marshal request", err)
	}

	// Create HTTP request URL with API key
	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", p.config.BaseURL, p.config.Model, p.config.APIKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, common.NewProviderError("gemini", "failed to create HTTP request", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")

	// Log request
	p.logger.APILog("gemini_request", map[string]interface{}{
		"model":    p.config.Model,
		"messages": len(req.Messages),
		"tools":    len(req.Tools),
	}, getRequestID(ctx))

	// Send request
	start := time.Now()
	httpResp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, common.NewProviderError("gemini", "failed to send HTTP request", err)
	}
	defer func() {
		if closeErr := httpResp.Body.Close(); closeErr != nil {
			p.logger.WithError(closeErr).Warn("Failed to close response body")
		}
	}()

	duration := time.Since(start)

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, common.NewProviderError("gemini", "failed to read response body", err)
	}

	// Check for HTTP errors
	if httpResp.StatusCode != http.StatusOK {
		p.logger.WithField("status_code", httpResp.StatusCode).
			WithField("response_body", string(respBody)).
			WithField("request_id", getRequestID(ctx)).
			Error("Gemini API returned error")

		return nil, common.NewHTTPError("gemini", httpResp, nil)
	}

	// Parse Gemini response and convert to OpenAI format
	var geminiResp Response
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return nil, common.NewProviderError("gemini", "failed to unmarshal response", err)
	}

	// Convert to OpenAI format
	resp := p.convertFromResponse(&geminiResp)

	// Log response
	p.logger.APILog("gemini_response", map[string]interface{}{
		"duration_ms":   duration.Milliseconds(),
		"finish_reason": getFinishReason(*resp),
		"candidates":    len(geminiResp.Candidates),
	}, getRequestID(ctx))

	return resp, nil
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return providerName
}

// GetModel returns the configured model
func (p *Provider) GetModel() string {
	return p.config.Model
}

// GetMaxTokens returns the maximum tokens limit
func (p *Provider) GetMaxTokens() int {
	return p.config.MaxTokens
}

// ValidateConfig validates the provider configuration
func (p *Provider) ValidateConfig() error {
	if p.config.APIKey == "" {
		return common.NewConfigError("gemini", "APIKey", "API key is required")
	}
	if p.config.BaseURL == "" {
		return common.NewConfigError("gemini", "BaseURL", "base URL is required")
	}
	if p.config.Model == "" {
		return common.NewConfigError("gemini", "Model", "model is required")
	}
	return nil
}

// GetBaseURL returns the base URL
func (p *Provider) GetBaseURL() string {
	return p.config.BaseURL
}

// Request represents the request format for Gemini API
type Request struct {
	GenerationConfig *GenerationConfig `json:"generationConfig,omitempty"`
	Contents         []Content         `json:"contents"`
	Tools            []Tool            `json:"tools,omitempty"`
}

// Content represents content in Gemini format
type Content struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}

// Part represents a part of content
type Part struct {
	FunctionCall     *FunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *FunctionResponse `json:"functionResponse,omitempty"`
	Text             string            `json:"text,omitempty"`
}

// FunctionCall represents a function call in Gemini format
type FunctionCall struct {
	Args map[string]interface{} `json:"args"`
	Name string                 `json:"name"`
}

// FunctionResponse represents a function response in Gemini format
type FunctionResponse struct {
	Response map[string]interface{} `json:"response"`
	Name     string                 `json:"name"`
}

// Tool represents a tool in Gemini format
type Tool struct {
	FunctionDeclarations []FunctionDeclaration `json:"functionDeclarations"`
}

// FunctionDeclaration represents a function declaration
type FunctionDeclaration struct {
	Parameters  map[string]interface{} `json:"parameters"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
}

// GenerationConfig represents generation configuration
type GenerationConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`
	TopP            *float64 `json:"topP,omitempty"`
	TopK            *int     `json:"topK,omitempty"`
	MaxOutputTokens *int     `json:"maxOutputTokens,omitempty"`
}

// Response represents the response format from Gemini API
type Response struct {
	Candidates    []Candidate   `json:"candidates"`
	UsageMetadata UsageMetadata `json:"usageMetadata"`
}

// Candidate represents a candidate response
type Candidate struct {
	FinishReason string  `json:"finishReason"`
	Content      Content `json:"content"`
	Index        int     `json:"index"`
}

// UsageMetadata represents usage metadata
type UsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// convertToRequest converts OpenAI format to Gemini format
func (p *Provider) convertToRequest(req *models.ChatCompletionRequest) *Request {
	geminiReq := &Request{
		Contents: make([]Content, 0, len(req.Messages)),
	}

	// Convert messages
	for _, msg := range req.Messages {
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}

		content := Content{
			Role:  role,
			Parts: []Part{{Text: msg.Content}},
		}

		// Handle tool calls
		if len(msg.ToolCalls) > 0 {
			for _, toolCall := range msg.ToolCalls {
				if toolCall.Function.Name != "" {
					var args map[string]interface{}
					if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err == nil {
						content.Parts = append(content.Parts, Part{
							FunctionCall: &FunctionCall{
								Name: toolCall.Function.Name,
								Args: args,
							},
						})
					}
				}
			}
		}

		geminiReq.Contents = append(geminiReq.Contents, content)
	}

	// Convert tools
	if len(req.Tools) > 0 {
		geminiTool := Tool{
			FunctionDeclarations: make([]FunctionDeclaration, 0, len(req.Tools)),
		}

		for _, tool := range req.Tools {
			if tool.Function.Name != "" {
				decl := FunctionDeclaration{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  tool.Function.Parameters,
				}
				geminiTool.FunctionDeclarations = append(geminiTool.FunctionDeclarations, decl)
			}
		}

		geminiReq.Tools = []Tool{geminiTool}
	}

	// Convert generation config
	config := &GenerationConfig{}
	if req.Temperature != nil {
		config.Temperature = req.Temperature
	}
	// TopP is not available in current ChatCompletionRequest model
	// This would need to be added if required
	if req.MaxTokens != nil {
		maxTokens := *req.MaxTokens
		if maxTokens > p.config.MaxTokens {
			maxTokens = p.config.MaxTokens
		}
		config.MaxOutputTokens = &maxTokens
	}

	geminiReq.GenerationConfig = config

	return geminiReq
}

// convertFromResponse converts Gemini response to OpenAI format
func (p *Provider) convertFromResponse(geminiResp *Response) *models.ChatCompletionResponse {
	resp := &models.ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%d", time.Now().Unix()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   p.config.Model,
		Choices: make([]models.ChatCompletionChoice, 0, len(geminiResp.Candidates)),
		Usage: models.ChatCompletionUsage{
			PromptTokens:     geminiResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: geminiResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      geminiResp.UsageMetadata.TotalTokenCount,
		},
	}

	// Convert candidates to choices
	for _, candidate := range geminiResp.Candidates {
		choice := models.ChatCompletionChoice{
			Index:        candidate.Index,
			FinishReason: convertFinishReason(candidate.FinishReason),
			Message: models.ChatMessage{
				Role:    "assistant",
				Content: "",
			},
		}

		// Extract text content
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				choice.Message.Content += part.Text
			}

			// Handle function calls
			if part.FunctionCall != nil {
				argsBytes, err := json.Marshal(part.FunctionCall.Args)
				if err != nil {
					argsBytes = []byte("{}")
				}
				toolCall := models.ToolCall{
					ID:   fmt.Sprintf("call_%d", time.Now().UnixNano()),
					Type: "function",
					Function: models.FunctionCall{
						Name:      part.FunctionCall.Name,
						Arguments: string(argsBytes),
					},
				}
				choice.Message.ToolCalls = append(choice.Message.ToolCalls, toolCall)
			}
		}

		resp.Choices = append(resp.Choices, choice)
	}

	return resp
}

// convertFinishReason converts Gemini finish reason to OpenAI format
func convertFinishReason(geminiReason string) string {
	switch geminiReason {
	case "STOP":
		return "stop"
	case "MAX_TOKENS":
		return "length"
	case "SAFETY":
		return "content_filter"
	case "RECITATION":
		return "content_filter"
	default:
		return "stop"
	}
}

// getRequestID extracts request ID from context
func getRequestID(ctx context.Context) string {
	if requestID := ctx.Value("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return "unknown"
}

// getFinishReason extracts finish reason from response
func getFinishReason(resp models.ChatCompletionResponse) string {
	if len(resp.Choices) > 0 {
		return resp.Choices[0].FinishReason
	}
	return "unknown"
}

// HealthCheck performs a health check on the gemini provider
func (p *Provider) HealthCheck(ctx context.Context) error {
	// Simple health check by making a minimal request
	req := &models.ChatCompletionRequest{
		Model: p.config.Model,
		Messages: []models.ChatMessage{
			{
				Role:    "user",
				Content: "health check",
			},
		},
		MaxTokens: &[]int{1}[0],
	}

	_, err := p.CreateChatCompletion(ctx, req)
	if err != nil {
		return common.NewProviderError("gemini", "health check failed", err)
	}

	return nil
}
