package transformer

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestNewMaxTokenTransformer(t *testing.T) {
	transformer := NewMaxTokenTransformer()

	if transformer == nil {
		t.Error("Expected non-nil MaxToken transformer")
	}

	if transformer.GetName() != "maxtoken" {
		t.Errorf("Expected name 'maxtoken', got %s", transformer.GetName())
	}

	if transformer.GetEndpoint() != "" {
		t.Errorf("Expected empty endpoint, got %s", transformer.GetEndpoint())
	}

	if transformer.defaultMaxTokens != 4096 {
		t.Errorf("Expected defaultMaxTokens 4096, got %d", transformer.defaultMaxTokens)
	}

	// Check provider limits
	expectedLimits := map[string]int{
		"anthropic":  200000,
		"openai":     128000,
		"groq":       32768,
		"gemini":     1048576,
		"deepseek":   32768,
		"openrouter": 200000,
		"mistral":    32768,
		"xai":        128000,
	}

	for provider, expectedLimit := range expectedLimits {
		actualLimit := transformer.GetProviderLimit(provider)
		if actualLimit != expectedLimit {
			t.Errorf("Provider %s: expected limit %d, got %d", provider, expectedLimit, actualLimit)
		}
	}
}

func TestMaxTokenTransformer_TransformRequestIn(t *testing.T) {
	transformer := NewMaxTokenTransformer()
	ctx := context.Background()

	t.Run("AddDefaultMaxTokens", func(t *testing.T) {
		request := map[string]interface{}{
			"model": "gpt-4",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Hello",
				},
			},
		}

		result, err := transformer.TransformRequestIn(ctx, request, "openai")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("Result is not a map")
		}

		if resultMap["max_tokens"] != 4096 {
			t.Errorf("Expected default max_tokens 4096, got %v", resultMap["max_tokens"])
		}
	})

	t.Run("ValidMaxTokens", func(t *testing.T) {
		request := map[string]interface{}{
			"model": "gpt-4",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Hello",
				},
			},
			"max_tokens": 1000,
		}

		result, err := transformer.TransformRequestIn(ctx, request, "openai")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		resultMap := result.(map[string]interface{})
		if resultMap["max_tokens"] != 1000 {
			t.Errorf("Expected max_tokens 1000, got %v", resultMap["max_tokens"])
		}
	})

	t.Run("ExcessiveMaxTokens_CappedToProviderLimit", func(t *testing.T) {
		request := map[string]interface{}{
			"model": "gpt-4",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Hello",
				},
			},
			"max_tokens": 500000, // Way over OpenAI's limit
		}

		result, err := transformer.TransformRequestIn(ctx, request, "openai")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		resultMap := result.(map[string]interface{})
		maxTokens := resultMap["max_tokens"]

		// Should be capped but allow for request token calculation
		if maxTokens.(int) > 128000 {
			t.Errorf("Expected max_tokens to be capped at or below provider limit, got %v", maxTokens)
		}
	})

	t.Run("ZeroMaxTokens_SetToDefault", func(t *testing.T) {
		request := map[string]interface{}{
			"model": "gpt-4",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Hello",
				},
			},
			"max_tokens": 0,
		}

		result, err := transformer.TransformRequestIn(ctx, request, "openai")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		resultMap := result.(map[string]interface{})
		if resultMap["max_tokens"] != 4096 {
			t.Errorf("Expected max_tokens to be set to default 4096, got %v", resultMap["max_tokens"])
		}
	})

	t.Run("NegativeMaxTokens_SetToDefault", func(t *testing.T) {
		request := map[string]interface{}{
			"model": "gpt-4",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Hello",
				},
			},
			"max_tokens": -100,
		}

		result, err := transformer.TransformRequestIn(ctx, request, "openai")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		resultMap := result.(map[string]interface{})
		if resultMap["max_tokens"] != 4096 {
			t.Errorf("Expected max_tokens to be set to default 4096, got %v", resultMap["max_tokens"])
		}
	})

	t.Run("FloatMaxTokens", func(t *testing.T) {
		request := map[string]interface{}{
			"model": "gpt-4",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Hello",
				},
			},
			"max_tokens": 1500.0,
		}

		result, err := transformer.TransformRequestIn(ctx, request, "openai")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		resultMap := result.(map[string]interface{})
		if resultMap["max_tokens"] != 1500 {
			t.Errorf("Expected max_tokens 1500, got %v", resultMap["max_tokens"])
		}
	})

	t.Run("InvalidMaxTokensType", func(t *testing.T) {
		request := map[string]interface{}{
			"model": "gpt-4",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Hello",
				},
			},
			"max_tokens": "invalid",
		}

		_, err := transformer.TransformRequestIn(ctx, request, "openai")
		if err == nil {
			t.Error("Expected error for invalid max_tokens type")
		}

		if !strings.Contains(err.Error(), "invalid max_tokens type") {
			t.Errorf("Expected 'invalid max_tokens type' error, got %v", err)
		}
	})

	t.Run("UnknownProvider_UsesDefault", func(t *testing.T) {
		request := map[string]interface{}{
			"model": "unknown-model",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Hello",
				},
			},
		}

		result, err := transformer.TransformRequestIn(ctx, request, "unknown")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		resultMap := result.(map[string]interface{})
		if resultMap["max_tokens"] != 4096 {
			t.Errorf("Expected default max_tokens 4096 for unknown provider, got %v", resultMap["max_tokens"])
		}
	})

	t.Run("RequestConfig_Handling", func(t *testing.T) {
		reqConfig := &RequestConfig{
			Body: map[string]interface{}{
				"model": "gpt-4",
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
				},
			},
			URL:     "https://api.openai.com/v1/chat/completions",
			Headers: map[string]string{"Authorization": "Bearer token"},
		}

		result, err := transformer.TransformRequestIn(ctx, reqConfig, "openai")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		resultConfig, ok := result.(*RequestConfig)
		if !ok {
			t.Fatal("Result is not a RequestConfig")
		}

		bodyMap := resultConfig.Body.(map[string]interface{})
		if bodyMap["max_tokens"] != 4096 {
			t.Errorf("Expected max_tokens 4096 in RequestConfig, got %v", bodyMap["max_tokens"])
		}

		// Other fields should be preserved
		if resultConfig.URL != reqConfig.URL {
			t.Error("URL should be preserved")
		}
		if resultConfig.Headers["Authorization"] != "Bearer token" {
			t.Error("Headers should be preserved")
		}
	})

	t.Run("NonMapRequest_PassThrough", func(t *testing.T) {
		request := "non-map request"

		result, err := transformer.TransformRequestIn(ctx, request, "openai")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result != request {
			t.Error("Non-map request should pass through unchanged")
		}
	})
}

func TestMaxTokenTransformer_TransformResponseOut(t *testing.T) {
	transformer := NewMaxTokenTransformer()
	ctx := context.Background()

	t.Run("AddMissingUsageFields", func(t *testing.T) {
		responseData := map[string]interface{}{
			"id":     "chatcmpl-123",
			"object": "chat.completion",
			"usage": map[string]interface{}{
				"prompt_tokens": float64(10),
				// completion_tokens missing
			},
		}

		body, _ := json.Marshal(responseData)
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(bytes.NewReader(body)),
		}

		result, err := transformer.TransformResponseOut(ctx, resp)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		resultBody, _ := io.ReadAll(result.Body)
		var resultData map[string]interface{}
		json.Unmarshal(resultBody, &resultData)

		usage := resultData["usage"].(map[string]interface{})
		if usage["completion_tokens"] != float64(0) {
			t.Errorf("Expected completion_tokens 0, got %v", usage["completion_tokens"])
		}

		if usage["total_tokens"] != float64(10) {
			t.Errorf("Expected total_tokens 10, got %v", usage["total_tokens"])
		}
	})

	t.Run("CalculateTotalTokens", func(t *testing.T) {
		responseData := map[string]interface{}{
			"id":     "chatcmpl-123",
			"object": "chat.completion",
			"usage": map[string]interface{}{
				"prompt_tokens":     float64(15),
				"completion_tokens": float64(25),
				// total_tokens missing
			},
		}

		body, _ := json.Marshal(responseData)
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(bytes.NewReader(body)),
		}

		result, err := transformer.TransformResponseOut(ctx, resp)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		resultBody, _ := io.ReadAll(result.Body)
		var resultData map[string]interface{}
		json.Unmarshal(resultBody, &resultData)

		usage := resultData["usage"].(map[string]interface{})
		totalTokensValue := usage["total_tokens"]
		if totalTokensValue != float64(40) {
			t.Errorf("Expected total_tokens 40, got %v (type: %T)", totalTokensValue, totalTokensValue)
		}
	})

	t.Run("NoUsageField", func(t *testing.T) {
		responseData := map[string]interface{}{
			"id":     "chatcmpl-123",
			"object": "chat.completion",
			"choices": []interface{}{
				map[string]interface{}{
					"message": map[string]interface{}{
						"content": "Hello!",
					},
				},
			},
		}

		body, _ := json.Marshal(responseData)
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(bytes.NewReader(body)),
		}

		result, err := transformer.TransformResponseOut(ctx, resp)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		resultBody, _ := io.ReadAll(result.Body)
		var resultData map[string]interface{}
		json.Unmarshal(resultBody, &resultData)

		// Should pass through unchanged when no usage field
		if resultData["id"] != "chatcmpl-123" {
			t.Error("Response should pass through unchanged when no usage field")
		}
	})

	t.Run("NonJSONResponse", func(t *testing.T) {
		body := "not json content"
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/plain"}},
			Body:       io.NopCloser(strings.NewReader(body)),
		}

		result, err := transformer.TransformResponseOut(ctx, resp)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		resultBody, _ := io.ReadAll(result.Body)
		if string(resultBody) != body {
			t.Error("Non-JSON response should pass through unchanged")
		}
	})
}

func TestMaxTokenTransformer_ProviderLimits(t *testing.T) {
	transformer := NewMaxTokenTransformer()

	t.Run("GetProviderLimit", func(t *testing.T) {
		tests := []struct {
			provider      string
			expectedLimit int
		}{
			{"anthropic", 200000},
			{"openai", 128000},
			{"gemini", 1048576},
			{"unknown", 4096}, // Should use default
		}

		for _, test := range tests {
			actual := transformer.GetProviderLimit(test.provider)
			if actual != test.expectedLimit {
				t.Errorf("Provider %s: expected limit %d, got %d", test.provider, test.expectedLimit, actual)
			}
		}
	})

	t.Run("SetProviderLimit", func(t *testing.T) {
		// Set a custom limit
		transformer.SetProviderLimit("custom", 50000)

		actual := transformer.GetProviderLimit("custom")
		if actual != 50000 {
			t.Errorf("Expected custom limit 50000, got %d", actual)
		}

		// Override existing limit
		transformer.SetProviderLimit("openai", 256000)
		actual = transformer.GetProviderLimit("openai")
		if actual != 256000 {
			t.Errorf("Expected updated OpenAI limit 256000, got %d", actual)
		}
	})
}

func TestMaxTokenTransformer_Integration(t *testing.T) {
	t.Run("WithDifferentProviders", func(t *testing.T) {
		transformer := NewMaxTokenTransformer()
		ctx := context.Background()

		request := map[string]interface{}{
			"model": "test-model",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Test",
				},
			},
			"max_tokens": 1000000, // Excessive for all providers
		}

		providers := []string{"anthropic", "openai", "gemini", "unknown"}

		for _, provider := range providers {
			result, err := transformer.TransformRequestIn(ctx, request, provider)
			if err != nil {
				t.Fatalf("Provider %s: unexpected error: %v", provider, err)
			}

			resultMap := result.(map[string]interface{})
			maxTokens := resultMap["max_tokens"].(int)

			expectedLimit := transformer.GetProviderLimit(provider)
			if maxTokens > expectedLimit {
				t.Errorf("Provider %s: max_tokens %d exceeds provider limit %d", provider, maxTokens, expectedLimit)
			}

			// Should be reasonable (accounting for request tokens and buffer)
			if maxTokens < 10 {
				t.Errorf("Provider %s: max_tokens %d too low", provider, maxTokens)
			}
		}
	})
}
