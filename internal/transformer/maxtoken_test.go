package transformer

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMaxTokenTransformer(t *testing.T) {
	transformer := NewMaxTokenTransformer()
	
	assert.NotNil(t, transformer)
	assert.Equal(t, "maxtoken", transformer.GetName())
	assert.Equal(t, "", transformer.GetEndpoint())
	assert.Equal(t, 4096, transformer.defaultMaxTokens)
	assert.NotEmpty(t, transformer.providerLimits)
}

func TestMaxTokenTransformer_TransformRequestIn(t *testing.T) {
	tests := []struct {
		name           string
		provider       string
		request        interface{}
		expectedTokens int
		expectError    bool
	}{
		{
			name:     "no max_tokens specified",
			provider: "openai",
			request: map[string]interface{}{
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
				},
			},
			expectedTokens: 4096, // default
		},
		{
			name:     "max_tokens within limit",
			provider: "openai",
			request: map[string]interface{}{
				"max_tokens": 1000,
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
				},
			},
			expectedTokens: 1000,
		},
		{
			name:     "max_tokens exceeds provider limit",
			provider: "groq",
			request: map[string]interface{}{
				"max_tokens": 50000,
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
				},
			},
			expectedTokens: 32617, // groq limit minus request tokens and buffer
		},
		{
			name:     "negative max_tokens",
			provider: "anthropic",
			request: map[string]interface{}{
				"max_tokens": -100,
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
				},
			},
			expectedTokens: 4096, // default
		},
		{
			name:     "request config with max_tokens",
			provider: "anthropic",
			request: &RequestConfig{
				Body: map[string]interface{}{
					"max_tokens": 2048,
					"messages": []interface{}{
						map[string]interface{}{
							"role":    "user",
							"content": "Hello",
						},
					},
				},
			},
			expectedTokens: 2048,
		},
		{
			name:     "very long input adjusts max_tokens",
			provider: "openai",
			request: map[string]interface{}{
				"max_tokens": 100000,
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": string(bytes.Repeat([]byte("Hello world! "), 10000)), // ~30k tokens
					},
				},
			},
			expectedTokens: 95350, // Adjusted for input tokens (32k) + buffer (100) = 128000 - 32650 = 95350
		},
	}

	transformer := NewMaxTokenTransformer()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := transformer.TransformRequestIn(ctx, tt.request, tt.provider)
			
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			
			// Extract the body map
			var bodyMap map[string]interface{}
			switch v := result.(type) {
			case map[string]interface{}:
				bodyMap = v
			case *RequestConfig:
				bodyMap = v.Body.(map[string]interface{})
			default:
				t.Fatalf("unexpected result type: %T", result)
			}
			
			// Check max_tokens
			maxTokens, ok := bodyMap["max_tokens"].(int)
			assert.True(t, ok, "max_tokens should be an int")
			assert.Equal(t, tt.expectedTokens, maxTokens)
		})
	}
}

func TestMaxTokenTransformer_TransformResponseOut(t *testing.T) {
	tests := []struct {
		name             string
		responseBody     interface{}
		expectedUsage    map[string]interface{}
		expectUsageAdded bool
	}{
		{
			name: "response with complete usage",
			responseBody: map[string]interface{}{
				"id": "test-123",
				"usage": map[string]interface{}{
					"prompt_tokens":     100,
					"completion_tokens": 50,
					"total_tokens":      150,
				},
			},
			expectedUsage: map[string]interface{}{
				"prompt_tokens":     100,
				"completion_tokens": 50,
				"total_tokens":      150,
			},
			expectUsageAdded: true,
		},
		{
			name: "response with partial usage",
			responseBody: map[string]interface{}{
				"id": "test-123",
				"usage": map[string]interface{}{
					"prompt_tokens":     100,
					"completion_tokens": 50,
				},
			},
			expectedUsage: map[string]interface{}{
				"prompt_tokens":     100,
				"completion_tokens": 50,
				"total_tokens":      150, // calculated
			},
			expectUsageAdded: true,
		},
		{
			name: "response without usage",
			responseBody: map[string]interface{}{
				"id":      "test-123",
				"content": "Hello world",
			},
			expectUsageAdded: false,
		},
		{
			name:             "non-JSON response",
			responseBody:     "plain text response",
			expectUsageAdded: false,
		},
	}

	transformer := NewMaxTokenTransformer()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create response body
			var body []byte
			if str, ok := tt.responseBody.(string); ok {
				body = []byte(str)
			} else {
				var err error
				body, err = json.Marshal(tt.responseBody)
				require.NoError(t, err)
			}

			// Create HTTP response
			response := &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(body)),
				Header:     http.Header{},
			}

			// Transform
			result, err := transformer.TransformResponseOut(ctx, response)
			require.NoError(t, err)
			assert.NotNil(t, result)

			// Read transformed body
			transformedBody, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			// Check if it's JSON
			var responseData map[string]interface{}
			err = json.Unmarshal(transformedBody, &responseData)
			
			if tt.expectUsageAdded && err == nil {
				usage, ok := responseData["usage"].(map[string]interface{})
				assert.True(t, ok, "usage should exist")
				
				if tt.expectedUsage != nil {
					// Compare as float64 since JSON unmarshaling returns numbers as float64
					assert.Equal(t, float64(tt.expectedUsage["prompt_tokens"].(int)), usage["prompt_tokens"])
					assert.Equal(t, float64(tt.expectedUsage["completion_tokens"].(int)), usage["completion_tokens"])
					assert.Equal(t, float64(tt.expectedUsage["total_tokens"].(int)), usage["total_tokens"])
				}
			}
		})
	}
}

func TestMaxTokenTransformer_ProviderLimits(t *testing.T) {
	transformer := NewMaxTokenTransformer()

	// Test getting existing provider limit
	assert.Equal(t, 200000, transformer.GetProviderLimit("anthropic"))
	assert.Equal(t, 128000, transformer.GetProviderLimit("openai"))
	
	// Test getting unknown provider limit (should return default)
	assert.Equal(t, 4096, transformer.GetProviderLimit("unknown"))
	
	// Test setting custom provider limit
	transformer.SetProviderLimit("custom", 65536)
	assert.Equal(t, 65536, transformer.GetProviderLimit("custom"))
}

func TestMaxTokenTransformer_EdgeCases(t *testing.T) {
	transformer := NewMaxTokenTransformer()
	ctx := context.Background()

	t.Run("invalid max_tokens type", func(t *testing.T) {
		request := map[string]interface{}{
			"max_tokens": "not a number",
			"messages":   []interface{}{},
		}
		
		_, err := transformer.TransformRequestIn(ctx, request, "openai")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid max_tokens type")
	})

	t.Run("pass through non-map request", func(t *testing.T) {
		request := "not a map"
		result, err := transformer.TransformRequestIn(ctx, request, "openai")
		assert.NoError(t, err)
		assert.Equal(t, request, result)
	})

	t.Run("minimum token enforcement", func(t *testing.T) {
		// Create a request with very long input that would leave no room for output
		request := map[string]interface{}{
			"max_tokens": 10000,
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": string(bytes.Repeat([]byte("x"), 500000)), // ~125k tokens
				},
			},
		}
		
		result, err := transformer.TransformRequestIn(ctx, request, "openai")
		require.NoError(t, err)
		
		bodyMap := result.(map[string]interface{})
		maxTokens := bodyMap["max_tokens"].(int)
		assert.Equal(t, 2850, maxTokens) // Should be adjusted based on available tokens
	})
}