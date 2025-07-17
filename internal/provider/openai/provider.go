// Package openai implements the OpenAI provider for CCProxy.
package openai

import (
	"context"
	"io"
	"net/http"
	"time"

	"ccproxy/internal/config"
	"ccproxy/internal/constants"
	"ccproxy/internal/models"
	"ccproxy/internal/provider/common"
	"ccproxy/pkg/logger"
)

// Provider implements the provider interface for OpenAI direct API
type Provider struct {
	httpClient *http.Client
	config     *config.OpenAIConfig
	logger     *logger.Logger
}

// NewProvider creates a new OpenAI provider instance
func NewProvider(cfg *config.OpenAIConfig, logger *logger.Logger) (*Provider, error) {
	if cfg == nil {
		return nil, common.NewConfigError("openai", "config", "config cannot be nil")
	}

	return &Provider{
		httpClient: common.NewConfiguredHTTPClient(cfg.Timeout),
		config:     cfg,
		logger:     logger,
	}, nil
}

// CreateChatCompletion sends a chat completion request to OpenAI API
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
	reqBody, err := common.MarshalJSONRequest(req, "openai")
	if err != nil {
		return nil, err
	}

	// Create HTTP request
	httpReq, err := common.CreateHTTPRequest(ctx, "POST", p.config.BaseURL+"/chat/completions", reqBody, "openai")
	if err != nil {
		return nil, err
	}

	// Set headers
	common.SetStandardHeaders(httpReq, p.config.APIKey)

	// Log request
	requestMetrics := common.CreateRequestMetrics(req)
	p.logger.APILog("openai_request", requestMetrics, common.GetRequestID(ctx))

	// Send request
	start := time.Now()
	httpResp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, common.NewProviderError("openai", "failed to send HTTP request", err)
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
		return nil, common.NewProviderError("openai", "failed to read response body", err)
	}

	// Check for HTTP errors
	if httpResp.StatusCode != http.StatusOK {
		p.logger.WithField("status_code", httpResp.StatusCode).
			WithField("response_body", string(respBody)).
			WithField("request_id", common.GetRequestID(ctx)).
			Error("OpenAI API returned error")

		return nil, common.NewHTTPError("openai", httpResp, nil)
	}

	// Parse response
	var resp models.ChatCompletionResponse
	if err := common.UnmarshalJSONResponse(respBody, &resp, "openai"); err != nil {
		return nil, err
	}

	// Log response
	responseMetrics := common.CreateResponseMetrics(resp, duration.Milliseconds())
	p.logger.APILog("openai_response", responseMetrics, common.GetRequestID(ctx))

	return &resp, nil
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return "openai"
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
		return common.NewConfigError("openai", "OPENAI_API_KEY", "API key is required")
	}
	if p.config.BaseURL == "" {
		return common.NewConfigError("openai", "OPENAI_BASE_URL", "base URL is required")
	}
	if p.config.Model == "" {
		return common.NewConfigError("openai", "OPENAI_MODEL", "model is required")
	}
	return nil
}

// GetBaseURL returns the base URL
func (p *Provider) GetBaseURL() string {
	return p.config.BaseURL
}


// HealthCheck performs a health check on the openai provider
func (p *Provider) HealthCheck(ctx context.Context) error {
	// Simple health check by making a minimal request
	req := &models.ChatCompletionRequest{
		Model: p.config.Model,
		Messages: []models.ChatMessage{
			{
				Role:    "user",
				Content: constants.HealthCheckMessage,
			},
		},
		MaxTokens: &[]int{constants.HealthCheckTokens}[0],
	}

	_, err := p.CreateChatCompletion(ctx, req)
	if err != nil {
		return common.NewProviderError("openai", "health check failed", err)
	}

	return nil
}
