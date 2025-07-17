package openai

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

// Provider implements the provider interface for OpenAI direct API
type Provider struct {
	httpClient *http.Client
	config     *config.OpenAIConfig
	logger     *logger.Logger
}

// NewProvider creates a new OpenAI provider instance
func NewProvider(cfg *config.OpenAIConfig, logger *logger.Logger) *Provider {
	return &Provider{
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		config: cfg,
		logger: logger,
	}
}

// CreateChatCompletion sends a chat completion request to OpenAI API
func (p *Provider) CreateChatCompletion(ctx context.Context, req *models.ChatCompletionRequest) (*models.ChatCompletionResponse, error) {
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
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	// Log request
	p.logger.APILog("openai_request", map[string]interface{}{
		"model":      req.Model,
		"messages":   len(req.Messages),
		"max_tokens": req.MaxTokens,
		"tools":      len(req.Tools),
	}, getRequestID(ctx))

	// Send request
	start := time.Now()
	httpResp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

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
			Error("OpenAI API returned error")

		return nil, fmt.Errorf("openai API error: %d %s", httpResp.StatusCode, string(respBody))
	}

	// Parse response
	var resp models.ChatCompletionResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Log response
	p.logger.APILog("openai_response", map[string]interface{}{
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
		return fmt.Errorf("OPENAI_API_KEY is required")
	}
	if p.config.BaseURL == "" {
		return fmt.Errorf("OPENAI_BASE_URL is required")
	}
	if p.config.Model == "" {
		return fmt.Errorf("OPENAI_MODEL is required")
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
