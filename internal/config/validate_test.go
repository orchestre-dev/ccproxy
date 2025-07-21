package config

import (
	"strings"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	t.Run("Valid config", func(t *testing.T) {
		config := &Config{
			Host: "127.0.0.1",
			Port: 3456,
			Providers: []Provider{
				{
					Name:       "openai",
					APIBaseURL: "https://api.openai.com/v1",
					APIKey:     "sk-test123",
					Models:     []string{"gpt-4"},
					Enabled:    true,
				},
			},
			Routes: map[string]Route{
				"default": {
					Provider: "openai",
					Model:    "gpt-4",
					Conditions: []Condition{
						{
							Type:     "tokenCount",
							Operator: ">",
							Value:    1000,
						},
					},
				},
			},
			Log:     true,
			LogFile: "/var/log/ccproxy.log",
		}

		err := config.Validate()
		if err != nil {
			t.Errorf("Expected no error for valid config, got: %v", err)
		}
	})

	t.Run("Empty host gets default", func(t *testing.T) {
		config := &Config{
			Host: "",
			Port: 3456,
		}

		err := config.Validate()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if config.Host != "127.0.0.1" {
			t.Errorf("Expected host to be set to default '127.0.0.1', got '%s'", config.Host)
		}
	})

	t.Run("Invalid port - zero", func(t *testing.T) {
		config := &Config{
			Host: "127.0.0.1",
			Port: 0,
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected error for port 0")
		}
		if !strings.Contains(err.Error(), "invalid port number: 0") {
			t.Errorf("Expected port error message, got: %v", err)
		}
	})

	t.Run("Invalid port - negative", func(t *testing.T) {
		config := &Config{
			Host: "127.0.0.1",
			Port: -1,
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected error for negative port")
		}
		if !strings.Contains(err.Error(), "invalid port number: -1") {
			t.Errorf("Expected port error message, got: %v", err)
		}
	})

	t.Run("Invalid port - too high", func(t *testing.T) {
		config := &Config{
			Host: "127.0.0.1",
			Port: 65536,
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected error for port 65536")
		}
		if !strings.Contains(err.Error(), "invalid port number: 65536") {
			t.Errorf("Expected port error message, got: %v", err)
		}
	})

	t.Run("Duplicate provider names", func(t *testing.T) {
		config := &Config{
			Host: "127.0.0.1",
			Port: 3456,
			Providers: []Provider{
				{
					Name:       "openai",
					APIBaseURL: "https://api.openai.com/v1",
					APIKey:     "sk-test123",
					Models:     []string{"gpt-4"},
					Enabled:    true,
				},
				{
					Name:       "openai", // Duplicate name
					APIBaseURL: "https://api.openai.com/v2",
					APIKey:     "sk-test456",
					Models:     []string{"gpt-3.5-turbo"},
					Enabled:    true,
				},
			},
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected error for duplicate provider names")
		}
		if !strings.Contains(err.Error(), "duplicate provider name: openai") {
			t.Errorf("Expected duplicate provider error message, got: %v", err)
		}
	})

	t.Run("Route references unknown provider", func(t *testing.T) {
		config := &Config{
			Host: "127.0.0.1",
			Port: 3456,
			Providers: []Provider{
				{
					Name:       "openai",
					APIBaseURL: "https://api.openai.com/v1",
					APIKey:     "sk-test123",
					Models:     []string{"gpt-4"},
					Enabled:    true,
				},
			},
			Routes: map[string]Route{
				"default": {
					Provider: "unknown-provider", // Provider doesn't exist
					Model:    "gpt-4",
				},
			},
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected error for unknown provider in route")
		}
		if !strings.Contains(err.Error(), "route default references unknown provider: unknown-provider") {
			t.Errorf("Expected unknown provider error message, got: %v", err)
		}
	})

	t.Run("Route with empty provider is allowed", func(t *testing.T) {
		config := &Config{
			Host: "127.0.0.1",
			Port: 3456,
			Routes: map[string]Route{
				"default": {
					Provider: "", // Empty provider should be allowed
					Model:    "gpt-4",
				},
			},
		}

		err := config.Validate()
		if err != nil {
			t.Errorf("Expected no error for empty provider in route, got: %v", err)
		}
	})

	t.Run("Invalid condition in route", func(t *testing.T) {
		config := &Config{
			Host: "127.0.0.1",
			Port: 3456,
			Providers: []Provider{
				{
					Name:       "openai",
					APIBaseURL: "https://api.openai.com/v1",
					APIKey:     "sk-test123",
					Models:     []string{"gpt-4"},
					Enabled:    true,
				},
			},
			Routes: map[string]Route{
				"default": {
					Provider: "openai",
					Model:    "gpt-4",
					Conditions: []Condition{
						{
							Type:     "invalid-type", // Invalid condition type
							Operator: ">",
							Value:    1000,
						},
					},
				},
			},
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected error for invalid condition")
		}
		if !strings.Contains(err.Error(), "invalid condition in route default") {
			t.Errorf("Expected invalid condition error message, got: %v", err)
		}
	})

	t.Run("Log file with null character", func(t *testing.T) {
		config := &Config{
			Host:    "127.0.0.1",
			Port:    3456,
			Log:     true,
			LogFile: "/var/log/test\x00file.log", // Null character in path
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected error for log file with null character")
		}
		if !strings.Contains(err.Error(), "invalid log file path") {
			t.Errorf("Expected invalid log file path error message, got: %v", err)
		}
	})

	t.Run("Valid log file path", func(t *testing.T) {
		config := &Config{
			Host:    "127.0.0.1",
			Port:    3456,
			Log:     true,
			LogFile: "/var/log/ccproxy.log",
		}

		err := config.Validate()
		if err != nil {
			t.Errorf("Expected no error for valid log file path, got: %v", err)
		}
	})

	t.Run("Logging disabled with log file set", func(t *testing.T) {
		config := &Config{
			Host:    "127.0.0.1",
			Port:    3456,
			Log:     false,
			LogFile: "/var/log/ccproxy.log", // Log file set but logging disabled
		}

		err := config.Validate()
		if err != nil {
			t.Errorf("Expected no error when logging disabled with log file set, got: %v", err)
		}
	})

	t.Run("Invalid provider error", func(t *testing.T) {
		config := &Config{
			Host: "127.0.0.1",
			Port: 3456,
			Providers: []Provider{
				{
					Name:       "", // Empty name should cause error
					APIBaseURL: "https://api.openai.com/v1",
					APIKey:     "sk-test123",
					Models:     []string{"gpt-4"},
					Enabled:    true,
				},
			},
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected error for invalid provider")
		}
		if !strings.Contains(err.Error(), "invalid provider") {
			t.Errorf("Expected invalid provider error message, got: %v", err)
		}
	})
}

func TestValidateProvider(t *testing.T) {
	t.Run("Valid provider", func(t *testing.T) {
		provider := &Provider{
			Name:       "openai",
			APIBaseURL: "https://api.openai.com/v1",
			APIKey:     "sk-test123",
			Models:     []string{"gpt-4"},
			Enabled:    true,
		}

		err := validateProvider(provider)
		if err != nil {
			t.Errorf("Expected no error for valid provider, got: %v", err)
		}
	})

	t.Run("Empty name", func(t *testing.T) {
		provider := &Provider{
			Name:       "",
			APIBaseURL: "https://api.openai.com/v1",
			APIKey:     "sk-test123",
			Models:     []string{"gpt-4"},
			Enabled:    true,
		}

		err := validateProvider(provider)
		if err == nil {
			t.Error("Expected error for empty provider name")
		}
		if !strings.Contains(err.Error(), "provider name is required") {
			t.Errorf("Expected name required error message, got: %v", err)
		}
	})

	t.Run("Empty API base URL", func(t *testing.T) {
		provider := &Provider{
			Name:       "openai",
			APIBaseURL: "",
			APIKey:     "sk-test123",
			Models:     []string{"gpt-4"},
			Enabled:    true,
		}

		err := validateProvider(provider)
		if err == nil {
			t.Error("Expected error for empty API base URL")
		}
		if !strings.Contains(err.Error(), "API base URL is required") {
			t.Errorf("Expected API base URL required error message, got: %v", err)
		}
	})

	t.Run("Invalid API base URL", func(t *testing.T) {
		provider := &Provider{
			Name:       "openai",
			APIBaseURL: "not-a-valid-url",
			APIKey:     "sk-test123",
			Models:     []string{"gpt-4"},
			Enabled:    true,
		}

		err := validateProvider(provider)
		if err == nil {
			t.Error("Expected error for invalid API base URL")
		}
		if !strings.Contains(err.Error(), "API base URL") {
			t.Errorf("Expected API base URL error message, got: %v", err)
		}
	})

	t.Run("API base URL with non-HTTP scheme", func(t *testing.T) {
		provider := &Provider{
			Name:       "openai",
			APIBaseURL: "ftp://api.openai.com/v1",
			APIKey:     "sk-test123",
			Models:     []string{"gpt-4"},
			Enabled:    true,
		}

		err := validateProvider(provider)
		if err == nil {
			t.Error("Expected error for non-HTTP scheme")
		}
		if !strings.Contains(err.Error(), "API base URL must use http or https scheme") {
			t.Errorf("Expected scheme error message, got: %v", err)
		}
	})

	t.Run("HTTP scheme is allowed", func(t *testing.T) {
		provider := &Provider{
			Name:       "local",
			APIBaseURL: "http://localhost:8080/v1",
			APIKey:     "test-key",
			Models:     []string{"local-model"},
			Enabled:    true,
		}

		err := validateProvider(provider)
		if err != nil {
			t.Errorf("Expected no error for HTTP scheme, got: %v", err)
		}
	})

	t.Run("Empty API key for enabled provider (warning only)", func(t *testing.T) {
		provider := &Provider{
			Name:       "openai",
			APIBaseURL: "https://api.openai.com/v1",
			APIKey:     "", // Empty API key
			Models:     []string{"gpt-4"},
			Enabled:    true,
		}

		err := validateProvider(provider)
		// This should not return an error (just a warning)
		if err != nil {
			t.Errorf("Expected no error for empty API key (should be warning only), got: %v", err)
		}
	})

	t.Run("Empty API key for disabled provider", func(t *testing.T) {
		provider := &Provider{
			Name:       "openai",
			APIBaseURL: "https://api.openai.com/v1",
			APIKey:     "", // Empty API key
			Models:     []string{"gpt-4"},
			Enabled:    false, // Disabled
		}

		err := validateProvider(provider)
		if err != nil {
			t.Errorf("Expected no error for empty API key on disabled provider, got: %v", err)
		}
	})

	t.Run("No models for enabled provider", func(t *testing.T) {
		provider := &Provider{
			Name:       "openai",
			APIBaseURL: "https://api.openai.com/v1",
			APIKey:     "sk-test123",
			Models:     []string{}, // Empty models
			Enabled:    true,
		}

		err := validateProvider(provider)
		if err == nil {
			t.Error("Expected error for no models on enabled provider")
		}
		if !strings.Contains(err.Error(), "at least one model must be specified for enabled provider") {
			t.Errorf("Expected models required error message, got: %v", err)
		}
	})

	t.Run("No models for disabled provider", func(t *testing.T) {
		provider := &Provider{
			Name:       "openai",
			APIBaseURL: "https://api.openai.com/v1",
			APIKey:     "sk-test123",
			Models:     []string{}, // Empty models
			Enabled:    false,      // Disabled
		}

		err := validateProvider(provider)
		if err != nil {
			t.Errorf("Expected no error for no models on disabled provider, got: %v", err)
		}
	})

	t.Run("Transformer with empty name", func(t *testing.T) {
		provider := &Provider{
			Name:       "openai",
			APIBaseURL: "https://api.openai.com/v1",
			APIKey:     "sk-test123",
			Models:     []string{"gpt-4"},
			Enabled:    true,
			Transformers: []TransformerConfig{
				{
					Name:   "", // Empty transformer name
					Config: map[string]interface{}{"key": "value"},
				},
			},
		}

		err := validateProvider(provider)
		if err == nil {
			t.Error("Expected error for transformer with empty name")
		}
		if !strings.Contains(err.Error(), "transformer name is required") {
			t.Errorf("Expected transformer name required error message, got: %v", err)
		}
	})

	t.Run("Valid transformer", func(t *testing.T) {
		provider := &Provider{
			Name:       "openai",
			APIBaseURL: "https://api.openai.com/v1",
			APIKey:     "sk-test123",
			Models:     []string{"gpt-4"},
			Enabled:    true,
			Transformers: []TransformerConfig{
				{
					Name:   "rate-limiter",
					Config: map[string]interface{}{"max_requests": 100},
				},
			},
		}

		err := validateProvider(provider)
		if err != nil {
			t.Errorf("Expected no error for valid transformer, got: %v", err)
		}
	})

	t.Run("Multiple valid transformers", func(t *testing.T) {
		provider := &Provider{
			Name:       "openai",
			APIBaseURL: "https://api.openai.com/v1",
			APIKey:     "sk-test123",
			Models:     []string{"gpt-4"},
			Enabled:    true,
			Transformers: []TransformerConfig{
				{
					Name:   "rate-limiter",
					Config: map[string]interface{}{"max_requests": 100},
				},
				{
					Name:   "token-counter",
					Config: map[string]interface{}{"max_tokens": 4000},
				},
			},
		}

		err := validateProvider(provider)
		if err != nil {
			t.Errorf("Expected no error for multiple valid transformers, got: %v", err)
		}
	})
}

func TestValidateCondition(t *testing.T) {
	t.Run("Valid tokenCount condition", func(t *testing.T) {
		condition := &Condition{
			Type:     "tokenCount",
			Operator: ">",
			Value:    1000,
		}

		err := validateCondition(condition)
		if err != nil {
			t.Errorf("Expected no error for valid tokenCount condition, got: %v", err)
		}
	})

	t.Run("Valid parameter condition", func(t *testing.T) {
		condition := &Condition{
			Type:     "parameter",
			Operator: "==",
			Value:    "test-value",
		}

		err := validateCondition(condition)
		if err != nil {
			t.Errorf("Expected no error for valid parameter condition, got: %v", err)
		}
	})

	t.Run("Valid model condition", func(t *testing.T) {
		condition := &Condition{
			Type:     "model",
			Operator: "contains",
			Value:    "gpt",
		}

		err := validateCondition(condition)
		if err != nil {
			t.Errorf("Expected no error for valid model condition, got: %v", err)
		}
	})

	t.Run("Invalid condition type", func(t *testing.T) {
		condition := &Condition{
			Type:     "invalidType",
			Operator: ">",
			Value:    1000,
		}

		err := validateCondition(condition)
		if err == nil {
			t.Error("Expected error for invalid condition type")
		}
		if !strings.Contains(err.Error(), "invalid condition type: invalidType") {
			t.Errorf("Expected invalid condition type error message, got: %v", err)
		}
	})

	t.Run("All valid operators", func(t *testing.T) {
		validOperators := []string{">", "<", "==", "!=", ">=", "<=", "contains"}

		for _, operator := range validOperators {
			t.Run("operator_"+operator, func(t *testing.T) {
				condition := &Condition{
					Type:     "tokenCount",
					Operator: operator,
					Value:    1000,
				}

				err := validateCondition(condition)
				if err != nil {
					t.Errorf("Expected no error for valid operator '%s', got: %v", operator, err)
				}
			})
		}
	})

	t.Run("Invalid operator", func(t *testing.T) {
		condition := &Condition{
			Type:     "tokenCount",
			Operator: "invalidOperator",
			Value:    1000,
		}

		err := validateCondition(condition)
		if err == nil {
			t.Error("Expected error for invalid operator")
		}
		if !strings.Contains(err.Error(), "invalid operator: invalidOperator") {
			t.Errorf("Expected invalid operator error message, got: %v", err)
		}
	})

	t.Run("Nil value", func(t *testing.T) {
		condition := &Condition{
			Type:     "tokenCount",
			Operator: ">",
			Value:    nil,
		}

		err := validateCondition(condition)
		if err == nil {
			t.Error("Expected error for nil condition value")
		}
		if !strings.Contains(err.Error(), "condition value is required") {
			t.Errorf("Expected condition value required error message, got: %v", err)
		}
	})

	t.Run("Different value types", func(t *testing.T) {
		testCases := []struct {
			name  string
			value interface{}
		}{
			{"string value", "test"},
			{"int value", 42},
			{"float value", 3.14},
			{"bool value", true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				condition := &Condition{
					Type:     "parameter",
					Operator: "==",
					Value:    tc.value,
				}

				err := validateCondition(condition)
				if err != nil {
					t.Errorf("Expected no error for %s, got: %v", tc.name, err)
				}
			})
		}
	})
}

func TestValidation_EdgeCases(t *testing.T) {
	t.Run("Boundary port values", func(t *testing.T) {
		testCases := []struct {
			port        int
			shouldError bool
			description string
		}{
			{1, false, "minimum valid port"},
			{65535, false, "maximum valid port"},
			{80, false, "common HTTP port"},
			{443, false, "common HTTPS port"},
			{3456, false, "default port"},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				config := &Config{
					Host: "127.0.0.1",
					Port: tc.port,
				}

				err := config.Validate()
				if tc.shouldError && err == nil {
					t.Errorf("Expected error for port %d", tc.port)
				}
				if !tc.shouldError && err != nil {
					t.Errorf("Expected no error for port %d, got: %v", tc.port, err)
				}
			})
		}
	})

	t.Run("Empty config validation", func(t *testing.T) {
		config := &Config{}

		err := config.Validate()
		if err == nil {
			t.Error("Expected error for completely empty config")
		}
	})

	t.Run("URL edge cases", func(t *testing.T) {
		testCases := []struct {
			url         string
			shouldError bool
			description string
		}{
			{"https://api.openai.com", false, "URL without path"},
			{"https://api.openai.com/", false, "URL with trailing slash"},
			{"https://api.openai.com/v1/", false, "URL with path and trailing slash"},
			{"https://localhost:8080", false, "localhost with port"},
			{"http://192.168.1.1:3000/api", false, "IP address with port and path"},
			{"https://", false, "incomplete HTTPS URL - Go parses this as valid"},
			{"://api.openai.com", true, "missing scheme"},
			{"", true, "empty URL handled by required check"},
		}

		for _, tc := range testCases {
			if tc.url == "" {
				continue // Skip empty URL as it's tested by required check
			}

			t.Run(tc.description, func(t *testing.T) {
				provider := &Provider{
					Name:       "test",
					APIBaseURL: tc.url,
					APIKey:     "test-key",
					Models:     []string{"test-model"},
					Enabled:    true,
				}

				err := validateProvider(provider)
				if tc.shouldError && err == nil {
					t.Errorf("Expected error for URL '%s'", tc.url)
				}
				if !tc.shouldError && err != nil {
					t.Errorf("Expected no error for URL '%s', got: %v", tc.url, err)
				}
			})
		}
	})

	t.Run("Complex config validation", func(t *testing.T) {
		config := &Config{
			Host: "0.0.0.0",
			Port: 8080,
			Providers: []Provider{
				{
					Name:       "provider1",
					APIBaseURL: "https://api1.com/v1",
					APIKey:     "key1",
					Models:     []string{"model1", "model2"},
					Enabled:    true,
					Transformers: []TransformerConfig{
						{Name: "transformer1", Config: map[string]interface{}{"setting": "value"}},
					},
				},
				{
					Name:       "provider2",
					APIBaseURL: "http://localhost:9000",
					APIKey:     "key2",
					Models:     []string{"local-model"},
					Enabled:    false,
				},
			},
			Routes: map[string]Route{
				"route1": {
					Provider: "provider1",
					Model:    "model1",
					Conditions: []Condition{
						{Type: "tokenCount", Operator: ">", Value: 500},
						{Type: "parameter", Operator: "==", Value: "premium"},
					},
				},
				"route2": {
					Provider: "provider2",
					Model:    "local-model",
					Conditions: []Condition{
						{Type: "model", Operator: "contains", Value: "local"},
					},
				},
			},
			Log:     true,
			LogFile: "/var/log/app.log",
		}

		err := config.Validate()
		if err != nil {
			t.Errorf("Expected no error for complex valid config, got: %v", err)
		}
	})
}
