package transformer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/orchestre-dev/ccproxy/internal/utils"
)

// MaxTokenTransformer handles max_tokens parameter across different providers
type MaxTokenTransformer struct {
	*BaseTransformer
	defaultMaxTokens int
	providerLimits   map[string]int
}

// NewMaxTokenTransformer creates a new MaxToken transformer
func NewMaxTokenTransformer() *MaxTokenTransformer {
	return &MaxTokenTransformer{
		BaseTransformer:  NewBaseTransformer("maxtoken", ""),
		defaultMaxTokens: 4096,
		providerLimits: map[string]int{
			"anthropic":  200000,  // Claude 3 models
			"openai":     128000,  // GPT-4 Turbo
			"groq":       32768,   // Typical Groq limit
			"gemini":     1048576, // Gemini 1.5 Pro
			"deepseek":   32768,   // DeepSeek default
			"openrouter": 200000,  // Varies by model
			"mistral":    32768,   // Mistral default
			"xai":        128000,  // Grok models
		},
	}
}

// TransformRequestIn ensures max_tokens is within provider limits
func (t *MaxTokenTransformer) TransformRequestIn(ctx context.Context, request interface{}, provider string) (interface{}, error) {
	// Handle RequestConfig
	if reqConfig, ok := request.(*RequestConfig); ok {
		return t.transformRequestConfig(ctx, reqConfig, provider)
	}

	// Handle direct body
	bodyMap, ok := request.(map[string]interface{})
	if !ok {
		// Try to convert from other types
		data, err := json.Marshal(request)
		if err != nil {
			return request, nil // Pass through on error
		}
		if err := json.Unmarshal(data, &bodyMap); err != nil {
			return request, nil // Pass through on error
		}
	}

	// Process max_tokens
	if err := t.processMaxTokens(bodyMap, provider); err != nil {
		return nil, err
	}

	return bodyMap, nil
}

// transformRequestConfig handles RequestConfig type
func (t *MaxTokenTransformer) transformRequestConfig(ctx context.Context, reqConfig *RequestConfig, provider string) (interface{}, error) {
	bodyMap, ok := reqConfig.Body.(map[string]interface{})
	if !ok {
		// Try to convert
		data, err := json.Marshal(reqConfig.Body)
		if err != nil {
			return reqConfig, nil
		}
		if err := json.Unmarshal(data, &bodyMap); err != nil {
			return reqConfig, nil
		}
	}

	// Process max_tokens
	if err := t.processMaxTokens(bodyMap, provider); err != nil {
		return nil, err
	}

	reqConfig.Body = bodyMap
	return reqConfig, nil
}

// processMaxTokens handles the max_tokens parameter
func (t *MaxTokenTransformer) processMaxTokens(bodyMap map[string]interface{}, provider string) error {
	// Get provider limit
	providerLimit, exists := t.providerLimits[provider]
	if !exists {
		providerLimit = t.defaultMaxTokens
	}

	// Check if max_tokens is specified
	maxTokensInterface, exists := bodyMap["max_tokens"]
	if !exists {
		// Set default if not specified
		bodyMap["max_tokens"] = t.defaultMaxTokens
		return nil
	}

	// Convert to int
	var maxTokens int
	switch v := maxTokensInterface.(type) {
	case float64:
		maxTokens = int(v)
	case int:
		maxTokens = v
	case int64:
		maxTokens = int(v)
	default:
		return fmt.Errorf("invalid max_tokens type: %T", v)
	}

	// Validate and adjust
	if maxTokens <= 0 {
		maxTokens = t.defaultMaxTokens
	} else if maxTokens > providerLimit {
		// Log warning but cap to provider limit
		maxTokens = providerLimit
	}

	// Calculate tokens already used in the request
	requestTokens := utils.CountRequestTokens(bodyMap)
	
	// Ensure we don't exceed context window
	// Most models have input + output <= context_window
	// Leave some buffer for safety
	maxAvailable := providerLimit - requestTokens - 100 // 100 token safety buffer
	if maxAvailable < maxTokens {
		maxTokens = maxAvailable
	}
	
	// Ensure minimum viable response
	if maxTokens < 10 {
		maxTokens = 10
	}

	bodyMap["max_tokens"] = maxTokens
	return nil
}

// TransformResponseOut adds token usage information if available
func (t *MaxTokenTransformer) TransformResponseOut(ctx context.Context, response *http.Response) (*http.Response, error) {
	// Read response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return response, nil // Pass through on error
	}
	response.Body.Close()

	// Try to parse as JSON
	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		// Not JSON, restore body and pass through
		response.Body = io.NopCloser(bytes.NewReader(body))
		return response, nil
	}

	// Check if usage information exists
	if usage, ok := responseData["usage"].(map[string]interface{}); ok {
		// Ensure all token counts are present
		if _, hasPrompt := usage["prompt_tokens"]; !hasPrompt {
			usage["prompt_tokens"] = 0
		}
		if _, hasCompletion := usage["completion_tokens"]; !hasCompletion {
			usage["completion_tokens"] = 0
		}
		if _, hasTotal := usage["total_tokens"]; !hasTotal {
			// Calculate total if not present
			promptTokens := 0
			completionTokens := 0
			
			if pt, ok := usage["prompt_tokens"].(float64); ok {
				promptTokens = int(pt)
			}
			if ct, ok := usage["completion_tokens"].(float64); ok {
				completionTokens = int(ct)
			}
			
			usage["total_tokens"] = promptTokens + completionTokens
		}
		
		// Add to response
		responseData["usage"] = usage
	}

	// Re-encode response
	newBody, err := json.Marshal(responseData)
	if err != nil {
		// Restore original body on error
		response.Body = io.NopCloser(bytes.NewReader(body))
		return response, nil
	}

	// Update response
	response.Body = io.NopCloser(bytes.NewReader(newBody))
	response.ContentLength = int64(len(newBody))
	response.Header.Set("Content-Length", fmt.Sprintf("%d", len(newBody)))

	return response, nil
}

// GetProviderLimit returns the token limit for a provider
func (t *MaxTokenTransformer) GetProviderLimit(provider string) int {
	if limit, exists := t.providerLimits[provider]; exists {
		return limit
	}
	return t.defaultMaxTokens
}

// SetProviderLimit sets the token limit for a provider
func (t *MaxTokenTransformer) SetProviderLimit(provider string, limit int) {
	t.providerLimits[provider] = limit
}