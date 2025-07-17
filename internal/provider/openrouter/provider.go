// Package openrouter implements the OpenRouter provider for CCProxy.
package openrouter

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

// Provider represents an OpenRouter provider implementation
type Provider struct {
	httpClient *http.Client
	config     *config.OpenRouterConfig
	logger     *logger.Logger
}

// NewProvider creates a new OpenRouter provider
func NewProvider(cfg *config.OpenRouterConfig, logger *logger.Logger) (*Provider, error) {
	if cfg == nil {
		return nil, common.NewConfigError("openrouter", "config", "config cannot be nil")
	}

	return &Provider{
		httpClient: common.NewConfiguredHTTPClient(cfg.Timeout),
		config: cfg,
		logger: logger,
	}, nil
}

// CreateChatCompletion sends a chat completion request to OpenRouter API
func (p *Provider) CreateChatCompletion(
	ctx context.Context,
	req *models.ChatCompletionRequest,
) (*models.ChatCompletionResponse, error) {
	// Apply max tokens limit if configured
	if p.config.MaxTokens > 0 {
		if req.MaxTokens == nil || *req.MaxTokens > p.config.MaxTokens {
			if req.MaxTokens != nil && *req.MaxTokens > p.config.MaxTokens {
				p.logger.Warnf("Capping max_tokens from %d to %d", *req.MaxTokens, p.config.MaxTokens)
			}
			req.MaxTokens = &p.config.MaxTokens
		}
	}

	// Use configured model
	req.Model = p.config.Model

	// Marshal request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, common.NewProviderError("openrouter", "failed to marshal request", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, common.NewProviderError("openrouter", "failed to create HTTP request", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	// Add optional OpenRouter-specific headers
	if p.config.SiteURL != "" {
		httpReq.Header.Set("HTTP-Referer", p.config.SiteURL)
	}
	if p.config.SiteName != "" {
		httpReq.Header.Set("X-Title", p.config.SiteName)
	}

	// Log request
	p.logger.APILog("openrouter_request", map[string]interface{}{
		"model":      req.Model,
		"messages":   len(req.Messages),
		"max_tokens": req.MaxTokens,
		"tools":      len(req.Tools),
		"site_url":   p.config.SiteURL,
		"site_name":  p.config.SiteName,
	}, getRequestID(ctx))

	// Send request
	start := time.Now()
	httpResp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, common.NewProviderError("openrouter", "failed to send HTTP request", err)
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
		return nil, common.NewProviderError("openrouter", "failed to read response body", err)
	}

	// Check for HTTP errors
	if httpResp.StatusCode != http.StatusOK {
		p.logger.WithField("status_code", httpResp.StatusCode).
			WithField("response_body", string(respBody)).
			WithField("request_id", getRequestID(ctx)).
			Error("OpenRouter API returned error")

		return nil, common.NewHTTPError("openrouter", httpResp, nil)
	}

	// Parse response
	var resp models.ChatCompletionResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, common.NewProviderError("openrouter", "failed to unmarshal response", err)
	}

	// Log response
	p.logger.APILog("openrouter_response", map[string]interface{}{
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
	return "openrouter"
}

// GetModel returns the configured model
func (p *Provider) GetModel() string {
	return p.config.Model
}

// GetMaxTokens returns the maximum tokens allowed
func (p *Provider) GetMaxTokens() int {
	return p.config.MaxTokens
}

// GetBaseURL returns the API base URL
func (p *Provider) GetBaseURL() string {
	return p.config.BaseURL
}

// ValidateConfig validates the provider configuration
func (p *Provider) ValidateConfig() error {
	if p.config.APIKey == "" {
		return common.NewConfigError("openrouter", "APIKey", "API key is required")
	}
	if p.config.BaseURL == "" {
		return common.NewConfigError("openrouter", "BaseURL", "base URL is required")
	}
	if p.config.Model == "" {
		return common.NewConfigError("openrouter", "Model", "model is required")
	}
	// MaxTokens is optional for OpenRouter, so we don't validate it
	return nil
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

// HealthCheck performs a health check on the openrouter provider
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
		return common.NewProviderError("openrouter", "health check failed", err)
	}

	return nil
}
