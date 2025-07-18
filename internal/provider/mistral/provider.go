// Package mistral implements the Mistral provider for CCProxy.
package mistral

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"ccproxy/internal/config"
	"ccproxy/internal/models"
	"ccproxy/internal/provider/common"
	"ccproxy/pkg/logger"
)

// Provider implements the provider interface for Mistral AI API
type Provider struct {
	httpClient *http.Client
	config     *config.MistralConfig
	logger     *logger.Logger
}

// NewProvider creates a new Mistral provider instance
func NewProvider(cfg *config.MistralConfig, logger *logger.Logger) (*Provider, error) {
	if cfg == nil {
		return nil, common.NewConfigError("mistral", "config", "config cannot be nil")
	}

	return &Provider{
		httpClient: common.NewConfiguredHTTPClient(cfg.Timeout),
		config:     cfg,
		logger:     logger,
	}, nil
}

// CreateChatCompletion sends a chat completion request to Mistral AI API
func (p *Provider) CreateChatCompletion(
	ctx context.Context,
	req *models.ChatCompletionRequest,
) (*models.ChatCompletionResponse, error) {
	// Apply max tokens limit
	if req.MaxTokens == nil || *req.MaxTokens > p.config.MaxTokens {
		if req.MaxTokens != nil && *req.MaxTokens > p.config.MaxTokens {
			p.logger.Warnf("Capping max_tokens from %d to %d", *req.MaxTokens, p.config.MaxTokens)
		}
		req.MaxTokens = &p.config.MaxTokens
	}

	// Use configured model
	req.Model = p.config.Model

	// Marshal request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, common.NewProviderError("mistral", "failed to marshal request", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, common.NewProviderError("mistral", "failed to create HTTP request", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	// Log request
	p.logger.APILog("mistral_request", map[string]interface{}{
		"model":      req.Model,
		"messages":   len(req.Messages),
		"max_tokens": req.MaxTokens,
		"tools":      len(req.Tools),
	}, getRequestID(ctx))

	// Send request
	start := time.Now()
	httpResp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, common.NewProviderError("mistral", "failed to send HTTP request", err)
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
		return nil, common.NewProviderError("mistral", "failed to read response body", err)
	}

	// Check for HTTP errors
	if httpResp.StatusCode != http.StatusOK {
		p.logger.WithField("status_code", httpResp.StatusCode).
			WithField("response_body", string(respBody)).
			WithField("request_id", getRequestID(ctx)).
			Error("Mistral API returned error")

		return nil, common.NewHTTPError("mistral", httpResp, nil)
	}

	// Parse response
	var resp models.ChatCompletionResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, common.NewProviderError("mistral", "failed to unmarshal response", err)
	}

	// Log response
	p.logger.APILog("mistral_response", map[string]interface{}{
		"duration_ms":       duration.Milliseconds(),
		"prompt_tokens":     resp.Usage.PromptTokens,
		"completion_tokens": resp.Usage.CompletionTokens,
		"total_tokens":      resp.Usage.TotalTokens,
		"finish_reason":     getFinishReason(resp),
	}, getRequestID(ctx))

	return &resp, nil
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return "mistral"
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
		return common.NewConfigError("mistral", "APIKey", "API key is required")
	}
	if p.config.BaseURL == "" {
		return common.NewConfigError("mistral", "BaseURL", "base URL is required")
	}
	if p.config.Model == "" {
		return common.NewConfigError("mistral", "Model", "model is required")
	}
	return nil
}

// GetBaseURL returns the base URL
func (p *Provider) GetBaseURL() string {
	return p.config.BaseURL
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

// HealthCheck performs a health check on the mistral provider
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
		return common.NewProviderError("mistral", "health check failed", err)
	}

	return nil
}
