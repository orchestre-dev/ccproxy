package transformer

import (
	"context"
	"testing"

	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
)

func TestParametersTransformer(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	t.Run("NewParametersTransformer", func(t *testing.T) {
		transformer := NewParametersTransformer()
		testutil.AssertEqual(t, "parameters", transformer.GetName())
		testutil.AssertEqual(t, "", transformer.GetEndpoint())
		testutil.AssertNotEqual(t, nil, transformer.parameterMappings)
		testutil.AssertNotEqual(t, nil, transformer.parameterLimits)
	})
}

func TestParametersTransformRequestIn(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewParametersTransformer()
	ctx := context.Background()

	t.Run("AnthropicParameterValidation", func(t *testing.T) {
		request := map[string]interface{}{
			"model":       "claude-3-sonnet",
			"temperature": 0.8,
			"top_p":       0.9,
			"top_k":       50.0,
		}

		result, err := transformer.TransformRequestIn(ctx, request, "anthropic")
		testutil.AssertNoError(t, err)

		resultMap := result.(map[string]interface{})
		testutil.AssertEqual(t, 0.8, resultMap["temperature"])
		testutil.AssertEqual(t, 0.9, resultMap["top_p"])
		testutil.AssertEqual(t, 50.0, resultMap["top_k"])

		// Anthropic shouldn't have these parameters
		_, hasPresencePenalty := resultMap["presence_penalty"]
		testutil.AssertEqual(t, false, hasPresencePenalty)
		_, hasFrequencyPenalty := resultMap["frequency_penalty"]
		testutil.AssertEqual(t, false, hasFrequencyPenalty)
	})

	t.Run("OpenAIParameterValidation", func(t *testing.T) {
		request := map[string]interface{}{
			"model":             "gpt-4",
			"temperature":       1.5,
			"top_p":             0.8,
			"presence_penalty":  1.0,
			"frequency_penalty": 0.5,
			"logprobs":          "true",
		}

		result, err := transformer.TransformRequestIn(ctx, request, "openai")
		testutil.AssertNoError(t, err)

		resultMap := result.(map[string]interface{})
		testutil.AssertEqual(t, 1.5, resultMap["temperature"])
		testutil.AssertEqual(t, 0.8, resultMap["top_p"])
		testutil.AssertEqual(t, 1.0, resultMap["presence_penalty"])
		testutil.AssertEqual(t, 0.5, resultMap["frequency_penalty"])
		testutil.AssertEqual(t, true, resultMap["logprobs"]) // Should be converted to boolean
	})

	t.Run("GeminiParameterMapping", func(t *testing.T) {
		request := map[string]interface{}{
			"model":       "gemini-pro",
			"temperature": 0.7,
			"max_tokens":  1000,
			"top_p":       0.95,
			"top_k":       40.0,
		}

		result, err := transformer.TransformRequestIn(ctx, request, "gemini")
		testutil.AssertNoError(t, err)

		resultMap := result.(map[string]interface{})

		// Should have generationConfig
		genConfig := resultMap["generationConfig"].(map[string]interface{})
		testutil.AssertEqual(t, 0.7, genConfig["temperature"])
		testutil.AssertEqual(t, 1000, genConfig["maxOutputTokens"])
		testutil.AssertEqual(t, 0.95, genConfig["topP"])
		testutil.AssertEqual(t, 40.0, genConfig["topK"])

		// Parameters should be removed from top level
		_, hasTemperature := resultMap["temperature"]
		testutil.AssertEqual(t, false, hasTemperature)
		_, hasMaxTokens := resultMap["max_tokens"]
		testutil.AssertEqual(t, false, hasMaxTokens)
	})

	t.Run("ParameterValidationErrors", func(t *testing.T) {
		testCases := []struct {
			name      string
			provider  string
			request   map[string]interface{}
			expectErr bool
		}{
			{
				name:     "AnthropicTemperatureOutOfRange",
				provider: "anthropic",
				request: map[string]interface{}{
					"temperature": 2.0, // Max is 1.0 for Anthropic
				},
				expectErr: true,
			},
			{
				name:     "OpenAITemperatureValid",
				provider: "openai",
				request: map[string]interface{}{
					"temperature": 2.0, // Max is 2.0 for OpenAI
				},
				expectErr: false,
			},
			{
				name:     "AnthropicTopKOutOfRange",
				provider: "anthropic",
				request: map[string]interface{}{
					"top_k": 200000.0, // Max is 100000 for Anthropic
				},
				expectErr: true,
			},
			{
				name:     "GeminiTopKValid",
				provider: "gemini",
				request: map[string]interface{}{
					"top_k": 50.0, // Within Gemini limits
				},
				expectErr: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := transformer.TransformRequestIn(ctx, tc.request, tc.provider)
				if tc.expectErr {
					testutil.AssertError(t, err)
				} else {
					testutil.AssertNoError(t, err)
				}
			})
		}
	})

	t.Run("RequestConfigHandling", func(t *testing.T) {
		reqConfig := &RequestConfig{
			Body: map[string]interface{}{
				"model":       "claude-3-sonnet",
				"temperature": 0.5,
			},
			URL: "https://api.anthropic.com/v1/messages",
		}

		result, err := transformer.TransformRequestIn(ctx, reqConfig, "anthropic")
		testutil.AssertNoError(t, err)

		resultConfig := result.(*RequestConfig)
		testutil.AssertEqual(t, "https://api.anthropic.com/v1/messages", resultConfig.URL)

		bodyMap := resultConfig.Body.(map[string]interface{})
		testutil.AssertEqual(t, "claude-3-sonnet", bodyMap["model"])
		testutil.AssertEqual(t, 0.5, bodyMap["temperature"])
	})

	t.Run("InvalidParameterType", func(t *testing.T) {
		request := map[string]interface{}{
			"temperature": "invalid_type", // Should be numeric
		}

		_, err := transformer.TransformRequestIn(ctx, request, "anthropic")
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "invalid temperature type")
	})

	t.Run("NonMapRequestHandling", func(t *testing.T) {
		request := "invalid request"

		result, err := transformer.TransformRequestIn(ctx, request, "anthropic")
		testutil.AssertNoError(t, err)

		// Should pass through unchanged for non-map requests
		testutil.AssertEqual(t, request, result)
	})

	t.Run("JSONUnmarshalableRequest", func(t *testing.T) {
		// Create a request that can't be marshaled to JSON
		request := make(chan int)

		result, err := transformer.TransformRequestIn(ctx, request, "anthropic")
		testutil.AssertNoError(t, err)

		// Should pass through unchanged when JSON operations fail
		testutil.AssertEqual(t, request, result)
	})
}

func TestParametersProcessParameters(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewParametersTransformer()

	t.Run("ValidParameterProcessing", func(t *testing.T) {
		bodyMap := map[string]interface{}{
			"temperature":       0.7,
			"top_p":             0.9,
			"presence_penalty":  0.5,
			"frequency_penalty": -0.5,
			"max_tokens":        1000,
		}

		err := transformer.processParameters(bodyMap, "openai")
		testutil.AssertNoError(t, err)

		// All parameters should be present for OpenAI
		testutil.AssertEqual(t, 0.7, bodyMap["temperature"])
		testutil.AssertEqual(t, 0.9, bodyMap["top_p"])
		testutil.AssertEqual(t, 0.5, bodyMap["presence_penalty"])
		testutil.AssertEqual(t, -0.5, bodyMap["frequency_penalty"])
		testutil.AssertEqual(t, 1000, bodyMap["max_tokens"])
	})

	t.Run("AnthropicParameterFiltering", func(t *testing.T) {
		bodyMap := map[string]interface{}{
			"temperature":       0.8,
			"top_p":             0.95,
			"presence_penalty":  0.5, // Should be removed
			"frequency_penalty": 0.3, // Should be removed
		}

		err := transformer.processParameters(bodyMap, "anthropic")
		testutil.AssertNoError(t, err)

		testutil.AssertEqual(t, 0.8, bodyMap["temperature"])
		testutil.AssertEqual(t, 0.95, bodyMap["top_p"])

		// These should be removed for Anthropic
		_, hasPresence := bodyMap["presence_penalty"]
		testutil.AssertEqual(t, false, hasPresence)
		_, hasFrequency := bodyMap["frequency_penalty"]
		testutil.AssertEqual(t, false, hasFrequency)
	})

	t.Run("GeminiParameterWrapping", func(t *testing.T) {
		bodyMap := map[string]interface{}{
			"model":       "gemini-pro",
			"temperature": 0.6,
			"max_tokens":  800,
			"top_p":       0.8,
			"top_k":       30.0,
		}

		err := transformer.processParameters(bodyMap, "gemini")
		testutil.AssertNoError(t, err)

		// Should have generationConfig
		genConfig := bodyMap["generationConfig"].(map[string]interface{})
		testutil.AssertEqual(t, 0.6, genConfig["temperature"])
		testutil.AssertEqual(t, 800, genConfig["maxOutputTokens"])
		testutil.AssertEqual(t, 0.8, genConfig["topP"])
		testutil.AssertEqual(t, 30.0, genConfig["topK"])

		// Top-level parameters should be removed
		_, hasTemp := bodyMap["temperature"]
		testutil.AssertEqual(t, false, hasTemp)
	})

	t.Run("ExistingGenerationConfigMerge", func(t *testing.T) {
		bodyMap := map[string]interface{}{
			"generationConfig": map[string]interface{}{
				"candidateCount": 1,
			},
			"temperature": 0.7,
			"max_tokens":  1200,
		}

		err := transformer.processParameters(bodyMap, "gemini")
		testutil.AssertNoError(t, err)

		genConfig := bodyMap["generationConfig"].(map[string]interface{})
		testutil.AssertEqual(t, 1, genConfig["candidateCount"])     // Preserved
		testutil.AssertEqual(t, 0.7, genConfig["temperature"])      // Added
		testutil.AssertEqual(t, 1200, genConfig["maxOutputTokens"]) // Added
	})
}

func TestParametersValidateParameter(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewParametersTransformer()

	t.Run("ValidNumericValues", func(t *testing.T) {
		testCases := []struct {
			name  string
			value interface{}
			limit Range
		}{
			{"ValidFloat", 0.5, Range{Min: 0, Max: 1}},
			{"ValidInt", 5, Range{Min: 1, Max: 10}},
			{"ValidInt64", int64(8), Range{Min: 0, Max: 100}},
			{"EdgeCaseMin", 0.0, Range{Min: 0, Max: 1}},
			{"EdgeCaseMax", 1.0, Range{Min: 0, Max: 1}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := transformer.validateParameter("test_param", tc.value, tc.limit)
				testutil.AssertNoError(t, err)
			})
		}
	})

	t.Run("InvalidValues", func(t *testing.T) {
		testCases := []struct {
			name  string
			value interface{}
			limit Range
			error string
		}{
			{"BelowMin", -0.5, Range{Min: 0, Max: 1}, "must be between"},
			{"AboveMax", 1.5, Range{Min: 0, Max: 1}, "must be between"},
			{"InvalidType", "invalid", Range{Min: 0, Max: 1}, "invalid test_param type"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := transformer.validateParameter("test_param", tc.value, tc.limit)
				testutil.AssertError(t, err)
				testutil.AssertContains(t, err.Error(), tc.error)
			})
		}
	})
}

func TestParametersWrapGeminiParameters(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewParametersTransformer()

	t.Run("WrapNewGenerationConfig", func(t *testing.T) {
		bodyMap := map[string]interface{}{
			"model":           "gemini-pro",
			"temperature":     0.8,
			"topP":            0.95,
			"topK":            25.0,
			"maxOutputTokens": 2000,
		}

		err := transformer.wrapGeminiParameters(bodyMap)
		testutil.AssertNoError(t, err)

		// Should create generationConfig
		genConfig := bodyMap["generationConfig"].(map[string]interface{})
		testutil.AssertEqual(t, 0.8, genConfig["temperature"])
		testutil.AssertEqual(t, 0.95, genConfig["topP"])
		testutil.AssertEqual(t, 25.0, genConfig["topK"])
		testutil.AssertEqual(t, 2000, genConfig["maxOutputTokens"])

		// Parameters should be removed from top level
		_, hasTemp := bodyMap["temperature"]
		testutil.AssertEqual(t, false, hasTemp)
		_, hasTopP := bodyMap["topP"]
		testutil.AssertEqual(t, false, hasTopP)
	})

	t.Run("MergeExistingGenerationConfig", func(t *testing.T) {
		bodyMap := map[string]interface{}{
			"generationConfig": map[string]interface{}{
				"stopSequences": []string{"END"},
			},
			"temperature": 0.9,
			"topK":        50.0,
		}

		err := transformer.wrapGeminiParameters(bodyMap)
		testutil.AssertNoError(t, err)

		genConfig := bodyMap["generationConfig"].(map[string]interface{})

		// Should preserve existing fields
		stopSeqs := genConfig["stopSequences"].([]string)
		testutil.AssertEqual(t, "END", stopSeqs[0])

		// Should add new fields
		testutil.AssertEqual(t, 0.9, genConfig["temperature"])
		testutil.AssertEqual(t, 50.0, genConfig["topK"])
	})

	t.Run("NoParametersToWrap", func(t *testing.T) {
		bodyMap := map[string]interface{}{
			"model":    "gemini-pro",
			"messages": []interface{}{},
		}

		err := transformer.wrapGeminiParameters(bodyMap)
		testutil.AssertNoError(t, err)

		// Should not create generationConfig if no parameters
		_, hasGenConfig := bodyMap["generationConfig"]
		testutil.AssertEqual(t, false, hasGenConfig)
	})
}

func TestParametersToBool(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewParametersTransformer()

	testCases := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"TrueBoolean", true, true},
		{"FalseBoolean", false, false},
		{"TrueString", "true", true},
		{"FalseString", "false", false},
		{"OneString", "1", true},
		{"YesString", "yes", true},
		{"OtherString", "no", false},
		{"NonZeroInt", 5, true},
		{"ZeroInt", 0, false},
		{"NonZeroFloat", 1.5, true},
		{"ZeroFloat", 0.0, false},
		{"InvalidType", []string{"invalid"}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := transformer.toBool(tc.input)
			testutil.AssertEqual(t, tc.expected, result)
		})
	}
}

func TestParametersSetParameterMapping(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewParametersTransformer()

	t.Run("SetNewMapping", func(t *testing.T) {
		transformer.SetParameterMapping("custom_provider", "temperature", "temp")

		// Test the mapping
		bodyMap := map[string]interface{}{
			"temperature": 0.7,
		}

		err := transformer.processParameters(bodyMap, "custom_provider")
		testutil.AssertNoError(t, err)

		// Should have mapped parameter name
		testutil.AssertEqual(t, 0.7, bodyMap["temp"])
		_, hasOriginal := bodyMap["temperature"]
		testutil.AssertEqual(t, false, hasOriginal)
	})

	t.Run("UpdateExistingMapping", func(t *testing.T) {
		transformer.SetParameterMapping("gemini", "temperature", "customTemp")

		bodyMap := map[string]interface{}{
			"temperature": 0.8,
		}

		err := transformer.processParameters(bodyMap, "gemini")
		testutil.AssertNoError(t, err)

		// Should use new mapping instead of default
		if genConfig, ok := bodyMap["generationConfig"]; ok && genConfig != nil {
			genConfigMap := genConfig.(map[string]interface{})
			testutil.AssertEqual(t, 0.8, genConfigMap["customTemp"])
		} else {
			// Temperature might not be moved to generationConfig with custom mapping
			testutil.AssertEqual(t, 0.8, bodyMap["customTemp"])
		}
	})
}

func TestParametersSetParameterLimit(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewParametersTransformer()

	t.Run("SetNewLimit", func(t *testing.T) {
		transformer.SetParameterLimit("custom_provider", "temperature", 0.1, 0.9)

		// Test valid value
		bodyMap := map[string]interface{}{
			"temperature": 0.5,
		}
		err := transformer.processParameters(bodyMap, "custom_provider")
		testutil.AssertNoError(t, err)

		// Test invalid value
		bodyMap2 := map[string]interface{}{
			"temperature": 1.0, // Above limit
		}
		err = transformer.processParameters(bodyMap2, "custom_provider")
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "must be between 0.1 and 0.9")
	})

	t.Run("UpdateExistingLimit", func(t *testing.T) {
		transformer.SetParameterLimit("anthropic", "temperature", 0.2, 0.8)

		bodyMap := map[string]interface{}{
			"temperature": 0.9, // Would be valid with original limit (0-1) but not with new limit (0.2-0.8)
		}
		err := transformer.processParameters(bodyMap, "anthropic")
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "must be between 0.2 and 0.8")
	})
}

func TestParametersIntegration(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewParametersTransformer()
	ctx := context.Background()

	t.Run("CompleteWorkflow", func(t *testing.T) {
		// Test with various provider scenarios
		testCases := []struct {
			name     string
			provider string
			request  map[string]interface{}
			validate func(t *testing.T, result interface{})
		}{
			{
				name:     "AnthropicWorkflow",
				provider: "anthropic",
				request: map[string]interface{}{
					"model":             "claude-3-sonnet",
					"temperature":       0.7,
					"top_p":             0.9,
					"top_k":             100.0,
					"presence_penalty":  0.5, // Should be removed
					"frequency_penalty": 0.3, // Should be removed
					"max_tokens":        2000,
				},
				validate: func(t *testing.T, result interface{}) {
					resultMap := result.(map[string]interface{})
					testutil.AssertEqual(t, 0.7, resultMap["temperature"])
					testutil.AssertEqual(t, 0.9, resultMap["top_p"])
					testutil.AssertEqual(t, 100.0, resultMap["top_k"])
					testutil.AssertEqual(t, 2000, resultMap["max_tokens"])

					_, hasPresence := resultMap["presence_penalty"]
					testutil.AssertEqual(t, false, hasPresence)
					_, hasFrequency := resultMap["frequency_penalty"]
					testutil.AssertEqual(t, false, hasFrequency)
				},
			},
			{
				name:     "GeminiWorkflow",
				provider: "gemini",
				request: map[string]interface{}{
					"model":       "gemini-pro",
					"temperature": 0.8,
					"max_tokens":  1500,
					"top_p":       0.95,
					"top_k":       40.0,
				},
				validate: func(t *testing.T, result interface{}) {
					resultMap := result.(map[string]interface{})
					testutil.AssertEqual(t, "gemini-pro", resultMap["model"])

					genConfig := resultMap["generationConfig"].(map[string]interface{})
					testutil.AssertEqual(t, 0.8, genConfig["temperature"])
					testutil.AssertEqual(t, 1500, genConfig["maxOutputTokens"])
					testutil.AssertEqual(t, 0.95, genConfig["topP"])
					testutil.AssertEqual(t, 40.0, genConfig["topK"])

					// Top-level parameters should be moved
					_, hasTemp := resultMap["temperature"]
					testutil.AssertEqual(t, false, hasTemp)
				},
			},
			{
				name:     "OpenAIWorkflow",
				provider: "openai",
				request: map[string]interface{}{
					"model":             "gpt-4",
					"temperature":       1.2,
					"top_p":             0.85,
					"presence_penalty":  0.6,
					"frequency_penalty": -0.3,
					"logprobs":          "false",
				},
				validate: func(t *testing.T, result interface{}) {
					resultMap := result.(map[string]interface{})
					testutil.AssertEqual(t, 1.2, resultMap["temperature"])
					testutil.AssertEqual(t, 0.85, resultMap["top_p"])
					testutil.AssertEqual(t, 0.6, resultMap["presence_penalty"])
					testutil.AssertEqual(t, -0.3, resultMap["frequency_penalty"])
					testutil.AssertEqual(t, false, resultMap["logprobs"]) // Converted to boolean
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := transformer.TransformRequestIn(ctx, tc.request, tc.provider)
				testutil.AssertNoError(t, err)
				tc.validate(t, result)
			})
		}
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		// Test validation errors across different providers
		errorCases := []struct {
			name      string
			provider  string
			request   map[string]interface{}
			errorText string
		}{
			{
				name:      "AnthropicTemperatureHigh",
				provider:  "anthropic",
				request:   map[string]interface{}{"temperature": 1.5},
				errorText: "temperature must be between 0 and 1",
			},
			{
				name:      "GeminiTopKHigh",
				provider:  "gemini",
				request:   map[string]interface{}{"top_k": 150.0},
				errorText: "topK must be between 1 and 100",
			},
			{
				name:      "OpenAIPresencePenaltyLow",
				provider:  "openai",
				request:   map[string]interface{}{"presence_penalty": -3.0},
				errorText: "presence_penalty must be between -2 and 2",
			},
		}

		for _, tc := range errorCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := transformer.TransformRequestIn(ctx, tc.request, tc.provider)
				testutil.AssertError(t, err)
				testutil.AssertContains(t, err.Error(), tc.errorText)
			})
		}
	})
}
