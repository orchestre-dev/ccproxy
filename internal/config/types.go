package config

import (
	"time"
)

// Config represents the main configuration structure for CCProxy
type Config struct {
	Providers []Provider         `json:"providers" mapstructure:"providers"`
	Routes    map[string]Route   `json:"routes" mapstructure:"routes"`
	Log       bool               `json:"log" mapstructure:"log"`
	LogFile   string             `json:"log_file" mapstructure:"log_file"`
	Host      string             `json:"host" mapstructure:"host"`
	Port      int                `json:"port" mapstructure:"port"`
	APIKey    string             `json:"apikey" mapstructure:"apikey"`
	ProxyURL  string             `json:"proxy_url" mapstructure:"proxy_url"`
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

// Default configuration values
func DefaultConfig() *Config {
	return &Config{
		Host:      "127.0.0.1",
		Port:      3456,
		Log:       false,
		LogFile:   "",
		Routes:    map[string]Route{},
		Providers: []Provider{},
	}
}