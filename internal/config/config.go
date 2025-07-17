// Package config provides configuration management for CCProxy
package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Logging   LoggingConfig   `mapstructure:"logging"`
	Provider  string          `mapstructure:"provider"`
	Providers ProvidersConfig `mapstructure:"providers"`
	Server    ServerConfig    `mapstructure:"server"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	Environment     string        `mapstructure:"environment"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// ProvidersConfig holds configuration for all providers
type ProvidersConfig struct {
	Groq       GroqConfig       `mapstructure:"groq"`
	OpenRouter OpenRouterConfig `mapstructure:"openrouter"`
	OpenAI     OpenAIConfig     `mapstructure:"openai"`
	XAI        XAIConfig        `mapstructure:"xai"`
	Gemini     GeminiConfig     `mapstructure:"gemini"`
	Mistral    MistralConfig    `mapstructure:"mistral"`
	Ollama     OllamaConfig     `mapstructure:"ollama"`
}

// GroqConfig holds Groq API configuration
type GroqConfig struct {
	APIKey    string        `mapstructure:"api_key"`
	BaseURL   string        `mapstructure:"base_url"`
	Model     string        `mapstructure:"model"`
	MaxTokens int           `mapstructure:"max_tokens"`
	Timeout   time.Duration `mapstructure:"timeout"`
}

// OpenRouterConfig holds OpenRouter API configuration
type OpenRouterConfig struct {
	APIKey    string        `mapstructure:"api_key"`
	BaseURL   string        `mapstructure:"base_url"`
	Model     string        `mapstructure:"model"`
	SiteURL   string        `mapstructure:"site_url"`
	SiteName  string        `mapstructure:"site_name"`
	MaxTokens int           `mapstructure:"max_tokens"`
	Timeout   time.Duration `mapstructure:"timeout"`
}

// OpenAIConfig holds OpenAI API configuration
type OpenAIConfig struct {
	APIKey    string        `mapstructure:"api_key"`
	BaseURL   string        `mapstructure:"base_url"`
	Model     string        `mapstructure:"model"`
	MaxTokens int           `mapstructure:"max_tokens"`
	Timeout   time.Duration `mapstructure:"timeout"`
}

// XAIConfig holds XAI (Grok) API configuration
type XAIConfig struct {
	APIKey    string        `mapstructure:"api_key"`
	BaseURL   string        `mapstructure:"base_url"`
	Model     string        `mapstructure:"model"`
	MaxTokens int           `mapstructure:"max_tokens"`
	Timeout   time.Duration `mapstructure:"timeout"`
}

// GeminiConfig holds Google Gemini API configuration
type GeminiConfig struct {
	APIKey    string        `mapstructure:"api_key"`
	BaseURL   string        `mapstructure:"base_url"`
	Model     string        `mapstructure:"model"`
	MaxTokens int           `mapstructure:"max_tokens"`
	Timeout   time.Duration `mapstructure:"timeout"`
}

// MistralConfig holds Mistral AI API configuration
type MistralConfig struct {
	APIKey    string        `mapstructure:"api_key"`
	BaseURL   string        `mapstructure:"base_url"`
	Model     string        `mapstructure:"model"`
	MaxTokens int           `mapstructure:"max_tokens"`
	Timeout   time.Duration `mapstructure:"timeout"`
}

// OllamaConfig holds Ollama API configuration
type OllamaConfig struct {
	APIKey    string        `mapstructure:"api_key"`
	BaseURL   string        `mapstructure:"base_url"`
	Model     string        `mapstructure:"model"`
	MaxTokens int           `mapstructure:"max_tokens"`
	Timeout   time.Duration `mapstructure:"timeout"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load loads configuration from environment variables and config files
func Load() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set default values
	setDefaults()

	// Bind environment variables
	bindEnvVars()

	// Read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Config file not found or error reading: %v", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to decode config: %v", err)
	}

	// Validate required configuration
	validate(&config)

	return &config
}

// Default token limits for various providers
const (
	groqMaxTokens       = 16384
	openRouterMaxTokens = 4096
	openAIMaxTokens     = 4096
	xaiMaxTokens        = 128000
	geminiMaxTokens     = 32768
	mistralMaxTokens    = 32768
	ollamaMaxTokens     = 4096
)

// setDefaults sets default configuration values
func setDefaults() {
	// Provider defaults
	viper.SetDefault("provider", "groq")

	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", "7187")
	viper.SetDefault("server.environment", "development")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.shutdown_timeout", "5s")

	// Groq defaults
	viper.SetDefault("providers.groq.base_url", "https://api.groq.com/openai/v1")
	viper.SetDefault("providers.groq.model", "moonshotai/kimi-k2-instruct")
	viper.SetDefault("providers.groq.max_tokens", groqMaxTokens)
	viper.SetDefault("providers.groq.timeout", "60s")

	// OpenRouter defaults
	viper.SetDefault("providers.openrouter.base_url", "https://openrouter.ai/api/v1")
	viper.SetDefault("providers.openrouter.model", "openai/gpt-4o")
	viper.SetDefault("providers.openrouter.max_tokens", openRouterMaxTokens)
	viper.SetDefault("providers.openrouter.timeout", "60s")

	// OpenAI defaults
	viper.SetDefault("providers.openai.base_url", "https://api.openai.com/v1")
	viper.SetDefault("providers.openai.model", "gpt-4o")
	viper.SetDefault("providers.openai.max_tokens", openAIMaxTokens)
	viper.SetDefault("providers.openai.timeout", "60s")

	// XAI defaults
	viper.SetDefault("providers.xai.base_url", "https://api.x.ai/v1")
	viper.SetDefault("providers.xai.model", "grok-beta")
	viper.SetDefault("providers.xai.max_tokens", xaiMaxTokens)
	viper.SetDefault("providers.xai.timeout", "60s")

	// Gemini defaults
	viper.SetDefault("providers.gemini.base_url", "https://generativelanguage.googleapis.com")
	viper.SetDefault("providers.gemini.model", "gemini-2.0-flash")
	viper.SetDefault("providers.gemini.max_tokens", geminiMaxTokens)
	viper.SetDefault("providers.gemini.timeout", "60s")

	// Mistral defaults
	viper.SetDefault("providers.mistral.base_url", "https://api.mistral.ai/v1")
	viper.SetDefault("providers.mistral.model", "mistral-large-latest")
	viper.SetDefault("providers.mistral.max_tokens", mistralMaxTokens)
	viper.SetDefault("providers.mistral.timeout", "60s")

	// Ollama defaults
	viper.SetDefault("providers.ollama.base_url", "http://localhost:11434")
	viper.SetDefault("providers.ollama.model", "llama3.2")
	viper.SetDefault("providers.ollama.max_tokens", ollamaMaxTokens)
	viper.SetDefault("providers.ollama.timeout", "120s")
	viper.SetDefault("providers.ollama.api_key", "ollama")

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
}

// bindEnvVars binds environment variables to config keys
func bindEnvVars() {
	viper.AutomaticEnv()

	// Provider selection
	_ = viper.BindEnv("provider", "PROVIDER")

	// Server environment variables
	_ = viper.BindEnv("server.host", "SERVER_HOST")
	_ = viper.BindEnv("server.port", "SERVER_PORT", "PORT")
	_ = viper.BindEnv("server.environment", "SERVER_ENVIRONMENT", "ENV", "ENVIRONMENT")

	// Groq environment variables
	_ = viper.BindEnv("providers.groq.api_key", "GROQ_API_KEY")
	_ = viper.BindEnv("providers.groq.base_url", "GROQ_BASE_URL")
	_ = viper.BindEnv("providers.groq.model", "GROQ_MODEL")
	_ = viper.BindEnv("providers.groq.max_tokens", "GROQ_MAX_TOKENS")

	// OpenRouter environment variables
	_ = viper.BindEnv("providers.openrouter.api_key", "OPENROUTER_API_KEY")
	_ = viper.BindEnv("providers.openrouter.base_url", "OPENROUTER_BASE_URL")
	_ = viper.BindEnv("providers.openrouter.model", "OPENROUTER_MODEL")
	_ = viper.BindEnv("providers.openrouter.max_tokens", "OPENROUTER_MAX_TOKENS")
	_ = viper.BindEnv("providers.openrouter.site_url", "OPENROUTER_SITE_URL")
	_ = viper.BindEnv("providers.openrouter.site_name", "OPENROUTER_SITE_NAME")

	// OpenAI environment variables
	_ = viper.BindEnv("providers.openai.api_key", "OPENAI_API_KEY")
	_ = viper.BindEnv("providers.openai.base_url", "OPENAI_BASE_URL")
	_ = viper.BindEnv("providers.openai.model", "OPENAI_MODEL")
	_ = viper.BindEnv("providers.openai.max_tokens", "OPENAI_MAX_TOKENS")

	// XAI environment variables
	_ = viper.BindEnv("providers.xai.api_key", "XAI_API_KEY")
	_ = viper.BindEnv("providers.xai.base_url", "XAI_BASE_URL")
	_ = viper.BindEnv("providers.xai.model", "XAI_MODEL")
	_ = viper.BindEnv("providers.xai.max_tokens", "XAI_MAX_TOKENS")

	// Gemini environment variables
	_ = viper.BindEnv("providers.gemini.api_key", "GEMINI_API_KEY", "GOOGLE_API_KEY")
	_ = viper.BindEnv("providers.gemini.base_url", "GEMINI_BASE_URL")
	_ = viper.BindEnv("providers.gemini.model", "GEMINI_MODEL")
	_ = viper.BindEnv("providers.gemini.max_tokens", "GEMINI_MAX_TOKENS")

	// Mistral environment variables
	_ = viper.BindEnv("providers.mistral.api_key", "MISTRAL_API_KEY")
	_ = viper.BindEnv("providers.mistral.base_url", "MISTRAL_BASE_URL")
	_ = viper.BindEnv("providers.mistral.model", "MISTRAL_MODEL")
	_ = viper.BindEnv("providers.mistral.max_tokens", "MISTRAL_MAX_TOKENS")

	// Ollama environment variables
	_ = viper.BindEnv("providers.ollama.api_key", "OLLAMA_API_KEY")
	_ = viper.BindEnv("providers.ollama.base_url", "OLLAMA_BASE_URL")
	_ = viper.BindEnv("providers.ollama.model", "OLLAMA_MODEL")
	_ = viper.BindEnv("providers.ollama.max_tokens", "OLLAMA_MAX_TOKENS")

	// Logging environment variables
	_ = viper.BindEnv("logging.level", "LOG_LEVEL")
	_ = viper.BindEnv("logging.format", "LOG_FORMAT")
}

// validate validates required configuration
func validate(config *Config) {
	// Validate provider selection
	validProviders := []string{"groq", "openrouter", "openai", "xai", "gemini", "mistral", "ollama"}
	isValid := false
	for _, provider := range validProviders {
		if config.Provider == provider {
			isValid = true
			break
		}
	}
	if !isValid {
		log.Fatalf("Invalid provider '%s'. Valid providers are: %v", config.Provider, validProviders)
	}

	// Validate provider-specific configuration
	switch config.Provider {
	case "groq":
		validateGroqConfig(&config.Providers.Groq)
	case "openrouter":
		validateOpenRouterConfig(&config.Providers.OpenRouter)
	case "openai":
		validateOpenAIConfig(&config.Providers.OpenAI)
	case "xai":
		validateXAIConfig(&config.Providers.XAI)
	case "gemini":
		validateGeminiConfig(&config.Providers.Gemini)
	case "mistral":
		validateMistralConfig(&config.Providers.Mistral)
	case "ollama":
		validateOllamaConfig(&config.Providers.Ollama)
	}
}

// validateGroqConfig validates Groq-specific configuration
func validateGroqConfig(config *GroqConfig) {
	if config.APIKey == "" {
		log.Fatal("GROQ_API_KEY environment variable is required when using Groq provider")
	}

	if config.MaxTokens <= 0 {
		log.Fatal("GROQ_MAX_TOKENS must be greater than 0")
	}

	if config.MaxTokens > groqMaxTokens {
		log.Printf("Warning: GROQ_MAX_TOKENS (%d) exceeds recommended limit (%d)", config.MaxTokens, groqMaxTokens)
	}
}

// validateOpenRouterConfig validates OpenRouter-specific configuration
func validateOpenRouterConfig(config *OpenRouterConfig) {
	if config.APIKey == "" {
		log.Fatal("OPENROUTER_API_KEY environment variable is required when using OpenRouter provider")
	}

	if config.MaxTokens < 0 {
		log.Fatal("OPENROUTER_MAX_TOKENS cannot be negative")
	}

	// Note: OpenRouter max tokens can be 0 (unlimited), so we don't enforce a minimum
}

// validateOpenAIConfig validates OpenAI-specific configuration
func validateOpenAIConfig(config *OpenAIConfig) {
	if config.APIKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required when using OpenAI provider")
	}

	if config.MaxTokens <= 0 {
		log.Fatal("OPENAI_MAX_TOKENS must be greater than 0")
	}

	if config.MaxTokens > xaiMaxTokens {
		log.Printf("Warning: OPENAI_MAX_TOKENS (%d) exceeds GPT-4o context limit (%d)", config.MaxTokens, xaiMaxTokens)
	}
}

// validateXAIConfig validates XAI-specific configuration
func validateXAIConfig(config *XAIConfig) {
	if config.APIKey == "" {
		log.Fatal("XAI_API_KEY environment variable is required when using XAI provider")
	}

	if config.MaxTokens <= 0 {
		log.Fatal("XAI_MAX_TOKENS must be greater than 0")
	}

	if config.MaxTokens > xaiMaxTokens {
		log.Printf("Warning: XAI_MAX_TOKENS (%d) exceeds Grok context limit (%d)", config.MaxTokens, xaiMaxTokens)
	}
}

// validateGeminiConfig validates Gemini-specific configuration
func validateGeminiConfig(config *GeminiConfig) {
	if config.APIKey == "" {
		log.Fatal("GEMINI_API_KEY or GOOGLE_API_KEY environment variable is required when using Gemini provider")
	}

	if config.MaxTokens <= 0 {
		log.Fatal("GEMINI_MAX_TOKENS must be greater than 0")
	}

	if config.MaxTokens > geminiMaxTokens {
		log.Printf("Warning: GEMINI_MAX_TOKENS (%d) exceeds Gemini context limit (%d)", config.MaxTokens, geminiMaxTokens)
	}
}

// validateMistralConfig validates Mistral-specific configuration
func validateMistralConfig(config *MistralConfig) {
	if config.APIKey == "" {
		log.Fatal("MISTRAL_API_KEY environment variable is required when using Mistral provider")
	}

	if config.MaxTokens <= 0 {
		log.Fatal("MISTRAL_MAX_TOKENS must be greater than 0")
	}

	if config.MaxTokens > mistralMaxTokens {
		log.Printf("Warning: MISTRAL_MAX_TOKENS (%d) exceeds Mistral context limit (%d)", config.MaxTokens, mistralMaxTokens)
	}
}

// validateOllamaConfig validates Ollama-specific configuration
func validateOllamaConfig(config *OllamaConfig) {
	if config.BaseURL == "" {
		log.Fatal("OLLAMA_BASE_URL environment variable is required when using Ollama provider")
	}

	if config.Model == "" {
		log.Fatal("OLLAMA_MODEL environment variable is required when using Ollama provider")
	}

	if config.MaxTokens < 0 {
		log.Fatal("OLLAMA_MAX_TOKENS cannot be negative")
	}

	// Note: Ollama models can have varying context limits, so we don't enforce a strict upper limit
	if config.MaxTokens > xaiMaxTokens {
		log.Printf("Warning: OLLAMA_MAX_TOKENS (%d) is very large and may not be supported by all models", config.MaxTokens)
	}
}
