package transformer

import (
	"context"
	"encoding/json"
	"fmt"
)

// ParametersTransformer handles common parameters across different providers
type ParametersTransformer struct {
	*BaseTransformer
	parameterMappings map[string]map[string]string // provider -> parameter -> mapped name
	parameterLimits   map[string]map[string]Range  // provider -> parameter -> valid range
}

// Range defines min and max values for a parameter
type Range struct {
	Min float64
	Max float64
}

// NewParametersTransformer creates a new Parameters transformer
func NewParametersTransformer() *ParametersTransformer {
	return &ParametersTransformer{
		BaseTransformer: NewBaseTransformer("parameters", ""),
		parameterMappings: map[string]map[string]string{
			// Anthropic uses standard names
			"anthropic": {},
			// OpenAI uses standard names
			"openai": {},
			// Gemini has some different names
			"gemini": {
				"max_tokens": "maxOutputTokens",
				"top_p":      "topP",
				"top_k":      "topK",
			},
			// DeepSeek uses standard names
			"deepseek": {},
			// Groq uses standard names
			"groq": {},
		},
		parameterLimits: map[string]map[string]Range{
			"anthropic": {
				"temperature": {Min: 0, Max: 1},
				"top_p":       {Min: 0, Max: 1},
				"top_k":       {Min: 1, Max: 100000},
			},
			"openai": {
				"temperature":       {Min: 0, Max: 2},
				"top_p":             {Min: 0, Max: 1},
				"presence_penalty":  {Min: -2, Max: 2},
				"frequency_penalty": {Min: -2, Max: 2},
			},
			"gemini": {
				"temperature": {Min: 0, Max: 2},
				"topP":        {Min: 0, Max: 1},
				"topK":        {Min: 1, Max: 100},
			},
			"deepseek": {
				"temperature": {Min: 0, Max: 2},
				"top_p":       {Min: 0, Max: 1},
			},
			"groq": {
				"temperature": {Min: 0, Max: 2},
				"top_p":       {Min: 0, Max: 1},
			},
		},
	}
}

// TransformRequestIn validates and maps parameters for the provider
func (t *ParametersTransformer) TransformRequestIn(ctx context.Context, request interface{}, provider string) (interface{}, error) {
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

	// Process parameters
	if err := t.processParameters(bodyMap, provider); err != nil {
		return nil, err
	}

	return bodyMap, nil
}

// transformRequestConfig handles RequestConfig type
func (t *ParametersTransformer) transformRequestConfig(ctx context.Context, reqConfig *RequestConfig, provider string) (interface{}, error) {
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

	// Process parameters
	if err := t.processParameters(bodyMap, provider); err != nil {
		return nil, err
	}

	reqConfig.Body = bodyMap
	return reqConfig, nil
}

// processParameters validates and transforms parameters
func (t *ParametersTransformer) processParameters(bodyMap map[string]interface{}, provider string) error {
	// Get mappings and limits for provider
	mappings, hasMappings := t.parameterMappings[provider]
	limits, hasLimits := t.parameterLimits[provider]

	// Process common parameters
	commonParams := []string{"temperature", "top_p", "top_k", "presence_penalty", "frequency_penalty", "max_tokens"}

	for _, param := range commonParams {
		if value, exists := bodyMap[param]; exists {
			// Determine actual parameter name (might be mapped)
			actualParam := param
			if hasMappings {
				if mappedName, needsMapping := mappings[param]; needsMapping {
					actualParam = mappedName
				}
			}

			// Skip validation for non-numeric parameters like max_tokens mapping
			if param != "max_tokens" && hasLimits {
				// Check limit using the actual parameter name
				if limit, hasLimit := limits[actualParam]; hasLimit {
					if err := t.validateParameter(actualParam, value, limit); err != nil {
						return err
					}
				}
			}

			// Map parameter name if needed
			if actualParam != param {
				// Remove old parameter and add with new name
				delete(bodyMap, param)
				bodyMap[actualParam] = value
			}
		}
	}

	// Handle provider-specific validation
	switch provider {
	case "anthropic":
		// Anthropic doesn't support presence_penalty or frequency_penalty
		delete(bodyMap, "presence_penalty")
		delete(bodyMap, "frequency_penalty")

	case "gemini":
		// Gemini parameters need to be in generationConfig
		if err := t.wrapGeminiParameters(bodyMap); err != nil {
			return err
		}

	case "openai":
		// Ensure logprobs is boolean if present
		if logprobs, exists := bodyMap["logprobs"]; exists {
			bodyMap["logprobs"] = t.toBool(logprobs)
		}
	}

	return nil
}

// validateParameter checks if a parameter value is within valid range
func (t *ParametersTransformer) validateParameter(name string, value interface{}, limit Range) error {
	var floatVal float64

	switch v := value.(type) {
	case float64:
		floatVal = v
	case int:
		floatVal = float64(v)
	case int64:
		floatVal = float64(v)
	default:
		return fmt.Errorf("invalid %s type: %T", name, value)
	}

	if floatVal < limit.Min || floatVal > limit.Max {
		return fmt.Errorf("%s must be between %v and %v, got %v", name, limit.Min, limit.Max, floatVal)
	}

	return nil
}

// wrapGeminiParameters wraps parameters in generationConfig for Gemini
func (t *ParametersTransformer) wrapGeminiParameters(bodyMap map[string]interface{}) error {
	// Parameters that should be in generationConfig
	configParams := []string{"temperature", "topP", "topK", "maxOutputTokens"}

	// Check if generationConfig exists
	genConfig, exists := bodyMap["generationConfig"].(map[string]interface{})
	if !exists {
		genConfig = make(map[string]interface{})
	}

	// Move parameters to generationConfig
	for _, param := range configParams {
		if value, exists := bodyMap[param]; exists {
			genConfig[param] = value
			delete(bodyMap, param)
		}
	}

	// Only add generationConfig if it has parameters
	if len(genConfig) > 0 {
		bodyMap["generationConfig"] = genConfig
	}

	return nil
}

// toBool converts various types to boolean
func (t *ParametersTransformer) toBool(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return v == "true" || v == "1" || v == "yes"
	case int:
		return v != 0
	case float64:
		return v != 0
	default:
		return false
	}
}

// SetParameterMapping sets a custom parameter mapping for a provider
func (t *ParametersTransformer) SetParameterMapping(provider, original, mapped string) {
	if _, exists := t.parameterMappings[provider]; !exists {
		t.parameterMappings[provider] = make(map[string]string)
	}
	t.parameterMappings[provider][original] = mapped
}

// SetParameterLimit sets a parameter limit for a provider
func (t *ParametersTransformer) SetParameterLimit(provider, parameter string, min, max float64) {
	if _, exists := t.parameterLimits[provider]; !exists {
		t.parameterLimits[provider] = make(map[string]Range)
	}
	t.parameterLimits[provider][parameter] = Range{Min: min, Max: max}
}
