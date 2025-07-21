package transformer

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewParametersTransformer(t *testing.T) {
	transformer := NewParametersTransformer()
	
	assert.NotNil(t, transformer)
	assert.Equal(t, "parameters", transformer.GetName())
	assert.Equal(t, "", transformer.GetEndpoint())
	assert.NotEmpty(t, transformer.parameterMappings)
	assert.NotEmpty(t, transformer.parameterLimits)
}

func TestParametersTransformer_TransformRequestIn(t *testing.T) {
	tests := []struct {
		name        string
		provider    string
		request     interface{}
		expected    map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:     "anthropic standard parameters",
			provider: "anthropic",
			request: map[string]interface{}{
				"messages":          []interface{}{},
				"temperature":       0.7,
				"top_p":             0.9,
				"top_k":             50,
				"presence_penalty":  0.5, // Should be removed
				"frequency_penalty": 0.5, // Should be removed
			},
			expected: map[string]interface{}{
				"messages":    []interface{}{},
				"temperature": 0.7,
				"top_p":       0.9,
				"top_k":       50,
			},
		},
		{
			name:     "openai with valid parameters",
			provider: "openai",
			request: map[string]interface{}{
				"messages":          []interface{}{},
				"temperature":       1.5,
				"top_p":             0.95,
				"presence_penalty":  1.0,
				"frequency_penalty": -0.5,
				"logprobs":          "true",
			},
			expected: map[string]interface{}{
				"messages":          []interface{}{},
				"temperature":       1.5,
				"top_p":             0.95,
				"presence_penalty":  1.0,
				"frequency_penalty": -0.5,
				"logprobs":          true,
			},
		},
		{
			name:     "gemini parameter mapping",
			provider: "gemini",
			request: map[string]interface{}{
				"messages":    []interface{}{},
				"temperature": 0.8,
				"top_p":       0.9,
				"top_k":       40,
				"max_tokens":  1024,
			},
			expected: map[string]interface{}{
				"messages": []interface{}{},
				"generationConfig": map[string]interface{}{
					"temperature":      0.8,
					"topP":             0.9,
					"topK":             40,
					"maxOutputTokens": 1024,
				},
			},
		},
		{
			name:     "temperature out of range",
			provider: "anthropic",
			request: map[string]interface{}{
				"messages":    []interface{}{},
				"temperature": 2.5, // Max is 1 for anthropic
			},
			expectError: true,
			errorMsg:    "temperature must be between 0 and 1",
		},
		{
			name:     "invalid parameter type",
			provider: "openai",
			request: map[string]interface{}{
				"messages":    []interface{}{},
				"temperature": "not a number",
			},
			expectError: true,
			errorMsg:    "invalid temperature type",
		},
		{
			name:     "request config handling",
			provider: "openai",
			request: &RequestConfig{
				Body: map[string]interface{}{
					"messages":    []interface{}{},
					"temperature": 0.5,
				},
			},
			expected: map[string]interface{}{
				"messages":    []interface{}{},
				"temperature": 0.5,
			},
		},
		{
			name:     "groq parameters",
			provider: "groq",
			request: map[string]interface{}{
				"messages":    []interface{}{},
				"temperature": 1.2,
				"top_p":       0.8,
			},
			expected: map[string]interface{}{
				"messages":    []interface{}{},
				"temperature": 1.2,
				"top_p":       0.8,
			},
		},
		{
			name:     "deepseek parameters",
			provider: "deepseek",
			request: map[string]interface{}{
				"messages":    []interface{}{},
				"temperature": 1.8,
				"top_p":       0.95,
			},
			expected: map[string]interface{}{
				"messages":    []interface{}{},
				"temperature": 1.8,
				"top_p":       0.95,
			},
		},
		{
			name:     "gemini with existing generationConfig",
			provider: "gemini",
			request: map[string]interface{}{
				"messages": []interface{}{},
				"generationConfig": map[string]interface{}{
					"candidateCount": 1,
				},
				"temperature": 0.7,
				"top_k":       30,
			},
			expected: map[string]interface{}{
				"messages": []interface{}{},
				"generationConfig": map[string]interface{}{
					"candidateCount": 1,
					"temperature":    0.7,
					"topK":           30,
				},
			},
		},
	}

	transformer := NewParametersTransformer()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := transformer.TransformRequestIn(ctx, tt.request, tt.provider)
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
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
			
			// Check expected values
			assert.Equal(t, tt.expected, bodyMap)
		})
	}
}

func TestParametersTransformer_ToBool(t *testing.T) {
	transformer := NewParametersTransformer()
	
	tests := []struct {
		input    interface{}
		expected bool
	}{
		{true, true},
		{false, false},
		{"true", true},
		{"false", false},
		{"1", true},
		{"0", false},
		{"yes", true},
		{"no", false},
		{1, true},
		{0, false},
		{1.0, true},
		{0.0, false},
		{struct{}{}, false}, // Unknown type
	}
	
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.input), func(t *testing.T) {
			result := transformer.toBool(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParametersTransformer_CustomSettings(t *testing.T) {
	transformer := NewParametersTransformer()
	
	// Test custom parameter mapping
	transformer.SetParameterMapping("custom", "temperature", "temp")
	
	request := map[string]interface{}{
		"messages":    []interface{}{},
		"temperature": 0.8,
	}
	
	result, err := transformer.TransformRequestIn(context.Background(), request, "custom")
	require.NoError(t, err)
	
	bodyMap := result.(map[string]interface{})
	assert.NotContains(t, bodyMap, "temperature")
	assert.Equal(t, 0.8, bodyMap["temp"])
	
	// Test custom parameter limit
	transformer.SetParameterLimit("custom", "temp", 0, 0.9)
	
	request2 := map[string]interface{}{
		"messages":    []interface{}{},
		"temperature": 1.5, // Will be mapped to "temp" and validated
	}
	
	_, err = transformer.TransformRequestIn(context.Background(), request2, "custom")
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "temp must be between 0 and 0.9")
	}
}

func TestParametersTransformer_PassThrough(t *testing.T) {
	transformer := NewParametersTransformer()
	ctx := context.Background()
	
	// Test non-map request passes through
	nonMapRequest := "not a map"
	result, err := transformer.TransformRequestIn(ctx, nonMapRequest, "openai")
	assert.NoError(t, err)
	assert.Equal(t, nonMapRequest, result)
	
	// Test request with no parameters
	emptyRequest := map[string]interface{}{
		"messages": []interface{}{},
	}
	result2, err := transformer.TransformRequestIn(ctx, emptyRequest, "openai")
	assert.NoError(t, err)
	assert.Equal(t, emptyRequest, result2.(map[string]interface{}))
}