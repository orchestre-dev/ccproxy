package config

import (
	"time"
)

// Config represents the main configuration structure for CCProxy
type Config struct {
	Providers        []Provider           `json:"providers" mapstructure:"providers"`
	Routes           map[string]Route     `json:"routes" mapstructure:"routes"`
	Log              bool                 `json:"log" mapstructure:"log"`
	LogFile          string               `json:"log_file" mapstructure:"log_file"`
	Host             string               `json:"host" mapstructure:"host"`
	Port             int                  `json:"port" mapstructure:"port"`
	APIKey           string               `json:"apikey" mapstructure:"apikey"`
	ProxyURL         string               `json:"proxy_url" mapstructure:"proxy_url"`
	Performance      PerformanceConfig    `json:"performance" mapstructure:"performance"`
	ShutdownTimeout  time.Duration        `json:"shutdown_timeout" mapstructure:"shutdown_timeout"`
}

// Provider represents a LLM provider configuration
type Provider struct {
	Name         string               `json:"name" mapstructure:"name"`
	APIBaseURL   string               `json:"api_base_url" mapstructure:"api_base_url"`
	APIKey       string               `json:"api_key" mapstructure:"api_key"`
	Models       []string             `json:"models" mapstructure:"models"`
	Enabled      bool                 `json:"enabled" mapstructure:"enabled"`
	Transformers []TransformerConfig  `json:"transformers" mapstructure:"transformers"`
	CreatedAt    time.Time            `json:"created_at" mapstructure:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at" mapstructure:"updated_at"`
	MessageFormat string              `json:"message_format,omitempty" mapstructure:"message_format"` // Message format used by provider
}

// Route represents a routing configuration
type Route struct {
	Provider   string      `json:"provider" mapstructure:"provider"`
	Model      string      `json:"model" mapstructure:"model"`
	Conditions []Condition `json:"conditions" mapstructure:"conditions"`
}

// Condition represents a routing condition
type Condition struct {
	Type     string      `json:"type" mapstructure:"type"`         // "tokenCount", "parameter", "model"
	Operator string      `json:"operator" mapstructure:"operator"` // ">", "<", "==", "contains"
	Value    interface{} `json:"value" mapstructure:"value"`
}

// TransformerConfig represents transformer configuration
type TransformerConfig struct {
	Name   string                 `json:"name" mapstructure:"name"`
	Config map[string]interface{} `json:"config,omitempty" mapstructure:"config"`
}

// PerformanceConfig represents performance monitoring configuration
type PerformanceConfig struct {
	MetricsEnabled          bool          `json:"metrics_enabled" mapstructure:"metrics_enabled"`
	RateLimitEnabled        bool          `json:"rate_limit_enabled" mapstructure:"rate_limit_enabled"`
	RateLimitRequestsPerMin int           `json:"rate_limit_requests_per_min" mapstructure:"rate_limit_requests_per_min"`
	CircuitBreakerEnabled   bool          `json:"circuit_breaker_enabled" mapstructure:"circuit_breaker_enabled"`
	RequestTimeout          time.Duration `json:"request_timeout" mapstructure:"request_timeout"`
	MaxRequestBodySize      int64         `json:"max_request_body_size" mapstructure:"max_request_body_size"`
}

// Default configuration values
func DefaultConfig() *Config {
	return &Config{
		Host:      "127.0.0.1",
		Port:      3456,
		Log:       false,
		LogFile:   "",
		Routes:    map[string]Route{},
		Providers: []Provider{},
		Performance: PerformanceConfig{
			MetricsEnabled:          true,
			RateLimitEnabled:        false,
			RateLimitRequestsPerMin: 1000,
			CircuitBreakerEnabled:   true,
			RequestTimeout:          30 * time.Second, // Much more reasonable default
			MaxRequestBodySize:      10 * 1024 * 1024, // 10MB limit
		},
		ShutdownTimeout: 30 * time.Second,
	}
}