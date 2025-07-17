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
	"ccproxy/pkg/logger"
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
		return nil, fmt.Errorf("gemini config cannot be nil")
	}
	
	return &Provider{
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		config: cfg,
		logger: logger,
	}, nil
}

// CreateChatCompletion sends a chat completion request to Google Gemini API
func (p *Provider) CreateChatCompletion(
	ctx context.Context,
	req *models.ChatCompletionRequest,
) (*models.ChatCompletionResponse, error) {
	// Convert OpenAI format to Gemini format
	geminiReq := p.convertToGeminiRequest(req)

	// Marshal request
	reqBody, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request URL with API key
	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", p.config.BaseURL, p.config.Model, p.config.APIKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
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
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := httpResp.Body.Close(); err != nil {
			p.logger.WithError(err).Warn("Failed to close response body")
		}
	}()

	duration := time.Since(start)

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if httpResp.StatusCode != http.StatusOK {
		p.logger.WithField("status_code", httpResp.StatusCode).
			WithField("response_body", string(respBody)).
			WithField("request_id", getRequestID(ctx)).
			Error("Gemini API returned error")

		return nil, fmt.Errorf("gemini API error: %d %s", httpResp.StatusCode, string(respBody))
	}

	// Parse Gemini response and convert to OpenAI format
	var geminiResp GeminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Convert to OpenAI format
	resp := p.convertFromGeminiResponse(&geminiResp)

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
	return "gemini"
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
		return fmt.Errorf("GEMINI_API_KEY is required")
	}
	if p.config.BaseURL == "" {
		return fmt.Errorf("GEMINI_BASE_URL is required")
	}
	if p.config.Model == "" {
		return fmt.Errorf("GEMINI_MODEL is required")
	}
	return nil
}

// GetBaseURL returns the base URL
func (p *Provider) GetBaseURL() string {
	return p.config.BaseURL
}

// GeminiRequest represents the request format for Gemini API
type GeminiRequest struct {
	GenerationConfig *GeminiGenerationConfig `json:"generationConfig,omitempty"`
	Contents         []GeminiContent         `json:"contents"`
	Tools            []GeminiTool            `json:"tools,omitempty"`
}

// GeminiContent represents content in Gemini format
type GeminiContent struct {
	Role  string       `json:"role"`
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart represents a part of content
type GeminiPart struct {
	FunctionCall     *GeminiFunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *GeminiFunctionResponse `json:"functionResponse,omitempty"`
	Text             string                  `json:"text,omitempty"`
}

// GeminiFunctionCall represents a function call in Gemini format
type GeminiFunctionCall struct {
	Args map[string]interface{} `json:"args"`
	Name string                 `json:"name"`
}

// GeminiFunctionResponse represents a function response in Gemini format
type GeminiFunctionResponse struct {
	Response map[string]interface{} `json:"response"`
	Name     string                 `json:"name"`
}

// GeminiTool represents a tool in Gemini format
type GeminiTool struct {
	FunctionDeclarations []GeminiFunctionDeclaration `json:"functionDeclarations"`
}

// GeminiFunctionDeclaration represents a function declaration
type GeminiFunctionDeclaration struct {
	Parameters  map[string]interface{} `json:"parameters"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
}

// GeminiGenerationConfig represents generation configuration
type GeminiGenerationConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`
	TopP            *float64 `json:"topP,omitempty"`
	TopK            *int     `json:"topK,omitempty"`
	MaxOutputTokens *int     `json:"maxOutputTokens,omitempty"`
}

// GeminiResponse represents the response format from Gemini API
type GeminiResponse struct {
	Candidates    []GeminiCandidate   `json:"candidates"`
	UsageMetadata GeminiUsageMetadata `json:"usageMetadata"`
}

// GeminiCandidate represents a candidate response
type GeminiCandidate struct {
	FinishReason string        `json:"finishReason"`
	Content      GeminiContent `json:"content"`
	Index        int           `json:"index"`
}

// GeminiUsageMetadata represents usage metadata
type GeminiUsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// convertToGeminiRequest converts OpenAI format to Gemini format
func (p *Provider) convertToGeminiRequest(req *models.ChatCompletionRequest) *GeminiRequest {
	geminiReq := &GeminiRequest{
		Contents: make([]GeminiContent, 0, len(req.Messages)),
	}

	// Convert messages
	for _, msg := range req.Messages {
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}

		content := GeminiContent{
			Role:  role,
			Parts: []GeminiPart{{Text: msg.Content}},
		}

		// Handle tool calls
		if len(msg.ToolCalls) > 0 {
			for _, toolCall := range msg.ToolCalls {
				if toolCall.Function.Name != "" {
					var args map[string]interface{}
					if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err == nil {
						content.Parts = append(content.Parts, GeminiPart{
							FunctionCall: &GeminiFunctionCall{
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
		geminiTool := GeminiTool{
			FunctionDeclarations: make([]GeminiFunctionDeclaration, 0, len(req.Tools)),
		}

		for _, tool := range req.Tools {
			if tool.Function.Name != "" {
				decl := GeminiFunctionDeclaration{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  tool.Function.Parameters,
				}
				geminiTool.FunctionDeclarations = append(geminiTool.FunctionDeclarations, decl)
			}
		}

		geminiReq.Tools = []GeminiTool{geminiTool}
	}

	// Convert generation config
	config := &GeminiGenerationConfig{}
	if req.Temperature != nil {
		config.Temperature = req.Temperature
	}
	// Note: TopP is not available in current ChatCompletionRequest model
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

// convertFromGeminiResponse converts Gemini response to OpenAI format
func (p *Provider) convertFromGeminiResponse(geminiResp *GeminiResponse) *models.ChatCompletionResponse {
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
				argsBytes, _ := json.Marshal(part.FunctionCall.Args)
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
		return fmt.Errorf("gemini health check failed: %w", err)
	}
	
	return nil
}
