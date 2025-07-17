// Package common provides shared utilities for providers
package common

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"ccproxy/internal/constants"
	"ccproxy/internal/models"
)

// MarshalJSONRequest marshals a request to JSON and handles errors
func MarshalJSONRequest(req interface{}, provider string) ([]byte, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, NewProviderError(provider, "failed to marshal request", err)
	}
	return reqBody, nil
}

// CreateHTTPRequest creates an HTTP request with JSON body
func CreateHTTPRequest(ctx context.Context, method, url string, body []byte, provider string) (*http.Request, error) {
	httpReq, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, NewProviderError(provider, "failed to create HTTP request", err)
	}
	return httpReq, nil
}

// SetStandardHeaders sets standard headers for API requests
func SetStandardHeaders(req *http.Request, apiKey string) {
	req.Header.Set(constants.HeaderContentType, constants.ContentTypeJSON)
	if apiKey != "" {
		req.Header.Set(constants.HeaderAuthorization, "Bearer "+apiKey)
	}
}

// UnmarshalJSONResponse unmarshals a JSON response and handles errors
func UnmarshalJSONResponse(body []byte, resp interface{}, provider string) error {
	if err := json.Unmarshal(body, resp); err != nil {
		return NewProviderError(provider, "failed to unmarshal response", err)
	}
	return nil
}

// GetFinishReasonFromResponse extracts finish reason from ChatCompletionResponse
func GetFinishReasonFromResponse(resp models.ChatCompletionResponse) string {
	if len(resp.Choices) > 0 {
		return resp.Choices[0].FinishReason
	}
	return constants.DefaultFinishReason
}

// CreateRequestMetrics creates standard request metrics for logging
func CreateRequestMetrics(req *models.ChatCompletionRequest) map[string]interface{} {
	metrics := map[string]interface{}{
		"model":    req.Model,
		"messages": len(req.Messages),
	}
	
	if req.MaxTokens != nil {
		metrics["max_tokens"] = *req.MaxTokens
	}
	
	if req.Tools != nil {
		metrics["tools"] = len(req.Tools)
	}
	
	return metrics
}

// CreateResponseMetrics creates standard response metrics for logging
func CreateResponseMetrics(resp models.ChatCompletionResponse, durationMs int64) map[string]interface{} {
	metrics := map[string]interface{}{
		"duration_ms":  durationMs,
		"finish_reason": GetFinishReasonFromResponse(resp),
	}
	
	// Add usage metrics if available
	metrics["prompt_tokens"] = resp.Usage.PromptTokens
	metrics["completion_tokens"] = resp.Usage.CompletionTokens
	metrics["total_tokens"] = resp.Usage.TotalTokens
	
	return metrics
}