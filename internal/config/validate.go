package config

import (
	"fmt"
	"net/url"
	"strings"
)

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate host
	if c.Host == "" {
		c.Host = "127.0.0.1"
	}

	// Validate port
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", c.Port)
	}

	// Validate providers
	providerNames := make(map[string]bool)
	for i, provider := range c.Providers {
		// Check for duplicate names
		if providerNames[provider.Name] {
			return fmt.Errorf("duplicate provider name: %s", provider.Name)
		}
		providerNames[provider.Name] = true

		// Validate provider
		if err := validateProvider(&c.Providers[i]); err != nil {
			return fmt.Errorf("invalid provider %s: %w", provider.Name, err)
		}
	}

	// Validate routes
	for routeName, route := range c.Routes {
		// Check if provider exists
		if route.Provider != "" && !providerNames[route.Provider] {
			return fmt.Errorf("route %s references unknown provider: %s", routeName, route.Provider)
		}

		// Validate conditions
		for _, condition := range route.Conditions {
			if err := validateCondition(&condition); err != nil {
				return fmt.Errorf("invalid condition in route %s: %w", routeName, err)
			}
		}

		// Validate parameters
		if err := validateRouteParameters(route.Parameters); err != nil {
			return fmt.Errorf("invalid parameters in route %s: %w", routeName, err)
		}
	}

	// Validate log file path if logging is enabled
	if c.Log && c.LogFile != "" {
		// Just check if it's a valid path format
		if strings.ContainsAny(c.LogFile, "\x00") {
			return fmt.Errorf("invalid log file path")
		}
	}

	return nil
}

// validateProvider validates a provider configuration
func validateProvider(p *Provider) error {
	// Name is required
	if p.Name == "" {
		return fmt.Errorf("provider name is required")
	}

	// API base URL is required and must be valid
	if p.APIBaseURL == "" {
		return fmt.Errorf("API base URL is required")
	}

	// Parse and validate URL
	u, err := url.Parse(p.APIBaseURL)
	if err != nil {
		return fmt.Errorf("invalid API base URL: %w", err)
	}

	// Ensure it's HTTP or HTTPS
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("API base URL must use http or https scheme")
	}

	// API key validation - warn if empty but don't fail
	// This is just a warning, not an error
	// The provider health check will handle this

	// Validate models list
	if len(p.Models) == 0 && p.Enabled {
		return fmt.Errorf("at least one model must be specified for enabled provider")
	}

	// Validate transformer configs
	for _, transformer := range p.Transformers {
		if transformer.Name == "" {
			return fmt.Errorf("transformer name is required")
		}
	}

	return nil
}

// validateCondition validates a routing condition
func validateCondition(c *Condition) error {
	// Validate condition type
	switch c.Type {
	case "tokenCount", "parameter", "model":
		// Valid types
	default:
		return fmt.Errorf("invalid condition type: %s", c.Type)
	}

	// Validate operator
	switch c.Operator {
	case ">", "<", "==", "!=", ">=", "<=", "contains":
		// Valid operators
	default:
		return fmt.Errorf("invalid operator: %s", c.Operator)
	}

	// Value is required
	if c.Value == nil {
		return fmt.Errorf("condition value is required")
	}

	return nil
}

// validateRouteParameters validates parameters configured for a route
func validateRouteParameters(params map[string]interface{}) error {
	if params == nil {
		return nil // Parameters are optional
	}

	// Validate temperature if present
	if temp, exists := params["temperature"]; exists {
		var tempValue float64
		switch v := temp.(type) {
		case float64:
			tempValue = v
		case int:
			tempValue = float64(v)
		case float32:
			tempValue = float64(v)
		default:
			return fmt.Errorf("temperature must be a number, got %T", temp)
		}

		// Temperature range validation (most providers support 0-2)
		if tempValue < 0 || tempValue > 2 {
			return fmt.Errorf("temperature must be between 0 and 2, got %v", tempValue)
		}
	}

	// Validate other common parameters if present
	if topP, exists := params["top_p"]; exists {
		var topPValue float64
		switch v := topP.(type) {
		case float64:
			topPValue = v
		case int:
			topPValue = float64(v)
		case float32:
			topPValue = float64(v)
		default:
			return fmt.Errorf("top_p must be a number, got %T", topP)
		}

		if topPValue < 0 || topPValue > 1 {
			return fmt.Errorf("top_p must be between 0 and 1, got %v", topPValue)
		}
	}

	if maxTokens, exists := params["max_tokens"]; exists {
		var maxTokensValue int
		switch v := maxTokens.(type) {
		case int:
			maxTokensValue = v
		case float64:
			maxTokensValue = int(v)
		case int64:
			maxTokensValue = int(v)
		default:
			return fmt.Errorf("max_tokens must be an integer, got %T", maxTokens)
		}

		if maxTokensValue <= 0 {
			return fmt.Errorf("max_tokens must be positive, got %v", maxTokensValue)
		}
	}

	return nil
}
