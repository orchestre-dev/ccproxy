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
	if p.APIKey == "" && p.Enabled {
		// This is just a warning, not an error
		// The provider health check will handle this
	}

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
