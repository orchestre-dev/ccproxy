package config

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// Test default values
	if config.Host != "127.0.0.1" {
		t.Errorf("Expected host to be '127.0.0.1', got '%s'", config.Host)
	}
	if config.Port != 3456 {
		t.Errorf("Expected port to be 3456, got %d", config.Port)
	}
	if config.Log != false {
		t.Errorf("Expected log to be false, got %t", config.Log)
	}
	if config.LogFile != "" {
		t.Errorf("Expected log file to be empty, got '%s'", config.LogFile)
	}
	if config.APIKey != "" {
		t.Errorf("Expected API key to be empty, got '%s'", config.APIKey)
	}
	if config.ProxyURL != "" {
		t.Errorf("Expected proxy URL to be empty, got '%s'", config.ProxyURL)
	}

	// Test routes and providers are initialized as empty
	if config.Routes == nil {
		t.Error("Expected routes to be initialized, got nil")
	}
	if len(config.Routes) != 0 {
		t.Errorf("Expected routes to be empty by default, got %d routes", len(config.Routes))
	}
	if config.Providers == nil {
		t.Error("Expected providers to be initialized, got nil")
	}
	if len(config.Providers) != 0 {
		t.Errorf("Expected providers to be empty by default, got %d providers", len(config.Providers))
	}

	// Test performance defaults
	if !config.Performance.MetricsEnabled {
		t.Error("Expected metrics to be enabled by default")
	}
	if config.Performance.RateLimitEnabled {
		t.Error("Expected rate limit to be disabled by default")
	}
	if config.Performance.RateLimitRequestsPerMin != 1000 {
		t.Errorf("Expected default rate limit to be 1000, got %d", config.Performance.RateLimitRequestsPerMin)
	}
	if !config.Performance.CircuitBreakerEnabled {
		t.Error("Expected circuit breaker to be enabled by default")
	}
	if config.Performance.RequestTimeout != 30*time.Second {
		t.Errorf("Expected default request timeout to be 30s, got %v", config.Performance.RequestTimeout)
	}
	if config.Performance.MaxRequestBodySize != int64(10*1024*1024) {
		t.Errorf("Expected default max body size to be 10MB, got %d", config.Performance.MaxRequestBodySize)
	}

	// Test shutdown timeout
	if config.ShutdownTimeout != 30*time.Second {
		t.Errorf("Expected default shutdown timeout to be 30s, got %v", config.ShutdownTimeout)
	}
}

func TestConfig_StructFields(t *testing.T) {
	config := &Config{
		Providers: []Provider{
			{
				Name:          "test-provider",
				APIBaseURL:    "https://api.test.com",
				APIKey:        "test-key",
				Models:        []string{"model1", "model2"},
				Enabled:       true,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
				MessageFormat: "openai",
			},
		},
		Routes: map[string]Route{
			"default": {
				Provider: "test-provider",
				Model:    "model1",
				Conditions: []Condition{
					{
						Type:     "tokenCount",
						Operator: ">",
						Value:    1000,
					},
				},
			},
		},
		Log:             true,
		LogFile:         "/var/log/ccproxy.log",
		Host:            "0.0.0.0",
		Port:            8080,
		APIKey:          "global-api-key",
		ProxyURL:        "http://proxy.company.com:8080",
		Performance: PerformanceConfig{
			MetricsEnabled:          true,
			RateLimitEnabled:        true,
			RateLimitRequestsPerMin: 500,
			CircuitBreakerEnabled:   true,
			RequestTimeout:          60 * time.Second,
			MaxRequestBodySize:      5 * 1024 * 1024,
		},
		ShutdownTimeout: 45 * time.Second,
	}

	// Test that all fields are properly set
	if len(config.Providers) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(config.Providers))
	}
	if config.Providers[0].Name != "test-provider" {
		t.Errorf("Expected provider name 'test-provider', got '%s'", config.Providers[0].Name)
	}
	if config.Providers[0].APIBaseURL != "https://api.test.com" {
		t.Errorf("Expected API URL 'https://api.test.com', got '%s'", config.Providers[0].APIBaseURL)
	}
	if config.Providers[0].APIKey != "test-key" {
		t.Errorf("Expected API key 'test-key', got '%s'", config.Providers[0].APIKey)
	}
	if len(config.Providers[0].Models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(config.Providers[0].Models))
	}
	if !config.Providers[0].Enabled {
		t.Error("Expected provider to be enabled")
	}
	if config.Providers[0].MessageFormat != "openai" {
		t.Errorf("Expected message format 'openai', got '%s'", config.Providers[0].MessageFormat)
	}

	if len(config.Routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(config.Routes))
	}
	route, exists := config.Routes["default"]
	if !exists {
		t.Error("Expected 'default' route to exist")
	}
	if route.Provider != "test-provider" {
		t.Errorf("Expected route provider 'test-provider', got '%s'", route.Provider)
	}
	if route.Model != "model1" {
		t.Errorf("Expected route model 'model1', got '%s'", route.Model)
	}
	if len(route.Conditions) != 1 {
		t.Errorf("Expected 1 condition, got %d", len(route.Conditions))
	}

	if !config.Log {
		t.Error("Expected logging to be enabled")
	}
	if config.LogFile != "/var/log/ccproxy.log" {
		t.Errorf("Expected log file '/var/log/ccproxy.log', got '%s'", config.LogFile)
	}
	if config.Host != "0.0.0.0" {
		t.Errorf("Expected host '0.0.0.0', got '%s'", config.Host)
	}
	if config.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", config.Port)
	}
	if config.APIKey != "global-api-key" {
		t.Errorf("Expected API key 'global-api-key', got '%s'", config.APIKey)
	}
	if config.ProxyURL != "http://proxy.company.com:8080" {
		t.Errorf("Expected proxy URL 'http://proxy.company.com:8080', got '%s'", config.ProxyURL)
	}
	if config.ShutdownTimeout != 45*time.Second {
		t.Errorf("Expected shutdown timeout 45s, got %v", config.ShutdownTimeout)
	}
}

func TestProvider_StructFields(t *testing.T) {
	now := time.Now()
	provider := Provider{
		Name:       "openai",
		APIBaseURL: "https://api.openai.com/v1",
		APIKey:     "sk-test123",
		Models:     []string{"gpt-4", "gpt-3.5-turbo"},
		Enabled:    true,
		Transformers: []TransformerConfig{
			{
				Name: "token-limiter",
				Config: map[string]interface{}{
					"max_tokens": 4000,
				},
			},
		},
		CreatedAt:     now,
		UpdatedAt:     now.Add(time.Hour),
		MessageFormat: "openai",
	}

	if provider.Name != "openai" {
		t.Errorf("Expected provider name 'openai', got '%s'", provider.Name)
	}
	if provider.APIBaseURL != "https://api.openai.com/v1" {
		t.Errorf("Expected API base URL 'https://api.openai.com/v1', got '%s'", provider.APIBaseURL)
	}
	if provider.APIKey != "sk-test123" {
		t.Errorf("Expected API key 'sk-test123', got '%s'", provider.APIKey)
	}
	if len(provider.Models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(provider.Models))
	}
	if provider.Models[0] != "gpt-4" {
		t.Errorf("Expected first model 'gpt-4', got '%s'", provider.Models[0])
	}
	if provider.Models[1] != "gpt-3.5-turbo" {
		t.Errorf("Expected second model 'gpt-3.5-turbo', got '%s'", provider.Models[1])
	}
	if !provider.Enabled {
		t.Error("Expected provider to be enabled")
	}
	if len(provider.Transformers) != 1 {
		t.Errorf("Expected 1 transformer, got %d", len(provider.Transformers))
	}
	if provider.Transformers[0].Name != "token-limiter" {
		t.Errorf("Expected transformer name 'token-limiter', got '%s'", provider.Transformers[0].Name)
	}
	if provider.Transformers[0].Config["max_tokens"] != 4000 {
		t.Errorf("Expected max_tokens 4000, got %v", provider.Transformers[0].Config["max_tokens"])
	}
	if provider.CreatedAt != now {
		t.Errorf("Expected CreatedAt %v, got %v", now, provider.CreatedAt)
	}
	if provider.UpdatedAt != now.Add(time.Hour) {
		t.Errorf("Expected UpdatedAt %v, got %v", now.Add(time.Hour), provider.UpdatedAt)
	}
	if provider.MessageFormat != "openai" {
		t.Errorf("Expected message format 'openai', got '%s'", provider.MessageFormat)
	}
}

func TestRoute_StructFields(t *testing.T) {
	route := Route{
		Provider: "openai",
		Model:    "gpt-4",
		Conditions: []Condition{
			{
				Type:     "tokenCount",
				Operator: ">",
				Value:    1000,
			},
			{
				Type:     "parameter",
				Operator: "==",
				Value:    "high-quality",
			},
		},
	}

	if route.Provider != "openai" {
		t.Errorf("Expected route provider 'openai', got '%s'", route.Provider)
	}
	if route.Model != "gpt-4" {
		t.Errorf("Expected route model 'gpt-4', got '%s'", route.Model)
	}
	if len(route.Conditions) != 2 {
		t.Errorf("Expected 2 conditions, got %d", len(route.Conditions))
	}

	// Test first condition
	condition1 := route.Conditions[0]
	if condition1.Type != "tokenCount" {
		t.Errorf("Expected first condition type 'tokenCount', got '%s'", condition1.Type)
	}
	if condition1.Operator != ">" {
		t.Errorf("Expected first condition operator '>', got '%s'", condition1.Operator)
	}
	if condition1.Value != 1000 {
		t.Errorf("Expected first condition value 1000, got %v", condition1.Value)
	}

	// Test second condition
	condition2 := route.Conditions[1]
	if condition2.Type != "parameter" {
		t.Errorf("Expected second condition type 'parameter', got '%s'", condition2.Type)
	}
	if condition2.Operator != "==" {
		t.Errorf("Expected second condition operator '==', got '%s'", condition2.Operator)
	}
	if condition2.Value != "high-quality" {
		t.Errorf("Expected second condition value 'high-quality', got %v", condition2.Value)
	}
}

func TestCondition_StructFields(t *testing.T) {
	tests := []struct {
		name      string
		condition Condition
	}{
		{
			name: "TokenCount condition",
			condition: Condition{
				Type:     "tokenCount",
				Operator: ">",
				Value:    1000,
			},
		},
		{
			name: "Parameter condition",
			condition: Condition{
				Type:     "parameter",
				Operator: "==",
				Value:    "test-value",
			},
		},
		{
			name: "Model condition",
			condition: Condition{
				Type:     "model",
				Operator: "contains",
				Value:    "gpt",
			},
		},
		{
			name: "Numeric condition with float",
			condition: Condition{
				Type:     "tokenCount",
				Operator: "<=",
				Value:    1500.5,
			},
		},
		{
			name: "Boolean condition",
			condition: Condition{
				Type:     "parameter",
				Operator: "==",
				Value:    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition := tt.condition
			if condition.Type != tt.condition.Type {
				t.Errorf("Expected condition type '%s', got '%s'", tt.condition.Type, condition.Type)
			}
			if condition.Operator != tt.condition.Operator {
				t.Errorf("Expected condition operator '%s', got '%s'", tt.condition.Operator, condition.Operator)
			}
			if condition.Value != tt.condition.Value {
				t.Errorf("Expected condition value %v, got %v", tt.condition.Value, condition.Value)
			}
		})
	}
}

func TestTransformerConfig_StructFields(t *testing.T) {
	transformer := TransformerConfig{
		Name: "rate-limiter",
		Config: map[string]interface{}{
			"requests_per_minute": 100,
			"burst_size":          10,
			"enabled":             true,
			"timeout":             "30s",
		},
	}

	if transformer.Name != "rate-limiter" {
		t.Errorf("Expected transformer name 'rate-limiter', got '%s'", transformer.Name)
	}
	if transformer.Config == nil {
		t.Error("Expected transformer config to not be nil")
	}
	if transformer.Config["requests_per_minute"] != 100 {
		t.Errorf("Expected requests_per_minute 100, got %v", transformer.Config["requests_per_minute"])
	}
	if transformer.Config["burst_size"] != 10 {
		t.Errorf("Expected burst_size 10, got %v", transformer.Config["burst_size"])
	}
	if transformer.Config["enabled"] != true {
		t.Errorf("Expected enabled true, got %v", transformer.Config["enabled"])
	}
	if transformer.Config["timeout"] != "30s" {
		t.Errorf("Expected timeout '30s', got %v", transformer.Config["timeout"])
	}
}

func TestPerformanceConfig_StructFields(t *testing.T) {
	perfConfig := PerformanceConfig{
		MetricsEnabled:          true,
		RateLimitEnabled:        true,
		RateLimitRequestsPerMin: 2000,
		CircuitBreakerEnabled:   false,
		RequestTimeout:          120 * time.Second,
		MaxRequestBodySize:      50 * 1024 * 1024, // 50MB
	}

	if !perfConfig.MetricsEnabled {
		t.Error("Expected metrics to be enabled")
	}
	if !perfConfig.RateLimitEnabled {
		t.Error("Expected rate limit to be enabled")
	}
	if perfConfig.RateLimitRequestsPerMin != 2000 {
		t.Errorf("Expected rate limit 2000, got %d", perfConfig.RateLimitRequestsPerMin)
	}
	if perfConfig.CircuitBreakerEnabled {
		t.Error("Expected circuit breaker to be disabled")
	}
	if perfConfig.RequestTimeout != 120*time.Second {
		t.Errorf("Expected request timeout 120s, got %v", perfConfig.RequestTimeout)
	}
	if perfConfig.MaxRequestBodySize != int64(50*1024*1024) {
		t.Errorf("Expected max body size 50MB, got %d", perfConfig.MaxRequestBodySize)
	}
}

func TestEmptyStructs(t *testing.T) {
	// Test that empty structs can be created without panicking
	emptyConfig := &Config{}
	if emptyConfig.Host != "" {
		t.Errorf("Expected empty config host to be empty, got '%s'", emptyConfig.Host)
	}
	if emptyConfig.Port != 0 {
		t.Errorf("Expected empty config port to be 0, got %d", emptyConfig.Port)
	}
	if emptyConfig.Routes != nil {
		t.Error("Expected empty config routes to be nil")
	}
	if emptyConfig.Providers != nil {
		t.Error("Expected empty config providers to be nil")
	}

	emptyProvider := &Provider{}
	if emptyProvider.Name != "" {
		t.Errorf("Expected empty provider name to be empty, got '%s'", emptyProvider.Name)
	}
	if emptyProvider.APIBaseURL != "" {
		t.Errorf("Expected empty provider URL to be empty, got '%s'", emptyProvider.APIBaseURL)
	}
	if emptyProvider.Enabled {
		t.Error("Expected empty provider to be disabled")
	}

	emptyRoute := &Route{}
	if emptyRoute.Provider != "" {
		t.Errorf("Expected empty route provider to be empty, got '%s'", emptyRoute.Provider)
	}
	if emptyRoute.Model != "" {
		t.Errorf("Expected empty route model to be empty, got '%s'", emptyRoute.Model)
	}
	if emptyRoute.Conditions != nil {
		t.Error("Expected empty route conditions to be nil")
	}

	emptyCondition := &Condition{}
	if emptyCondition.Type != "" {
		t.Errorf("Expected empty condition type to be empty, got '%s'", emptyCondition.Type)
	}
	if emptyCondition.Operator != "" {
		t.Errorf("Expected empty condition operator to be empty, got '%s'", emptyCondition.Operator)
	}
	if emptyCondition.Value != nil {
		t.Errorf("Expected empty condition value to be nil, got %v", emptyCondition.Value)
	}
}

func TestTimeFields(t *testing.T) {
	now := time.Now()
	provider := Provider{
		Name:      "test",
		CreatedAt: now,
		UpdatedAt: now.Add(time.Hour),
	}

	if provider.CreatedAt != now {
		t.Errorf("Expected CreatedAt %v, got %v", now, provider.CreatedAt)
	}
	if provider.UpdatedAt != now.Add(time.Hour) {
		t.Errorf("Expected UpdatedAt %v, got %v", now.Add(time.Hour), provider.UpdatedAt)
	}
	if !provider.UpdatedAt.After(provider.CreatedAt) {
		t.Error("Expected UpdatedAt to be after CreatedAt")
	}
}

func TestConfigNilHandling(t *testing.T) {
	// Test that nil maps and slices are handled gracefully
	config := &Config{
		Routes:    nil,
		Providers: nil,
	}

	// These should not panic
	if config.Routes != nil {
		t.Error("Expected nil routes to remain nil")
	}
	if config.Providers != nil {
		t.Error("Expected nil providers to remain nil")
	}

	// Test accessing length of nil slices (should return 0)
	if len(config.Providers) != 0 {
		t.Errorf("Expected length of nil providers slice to be 0, got %d", len(config.Providers))
	}
}