package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Provider  string         `mapstructure:"provider"`
	Server    ServerConfig   `mapstructure:"server"`
	Providers ProvidersConfig `mapstructure:"providers"`
	Logging   LoggingConfig  `mapstructure:"logging"`
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
	MaxTokens int           `mapstructure:"max_tokens"`
	Timeout   time.Duration `mapstructure:"timeout"`
	SiteURL   string        `mapstructure:"site_url"`
	SiteName  string        `mapstructure:"site_name"`
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
	viper.SetDefault("providers.groq.max_tokens", 16384)
	viper.SetDefault("providers.groq.timeout", "60s")

	// OpenRouter defaults
	viper.SetDefault("providers.openrouter.base_url", "https://openrouter.ai/api/v1")
	viper.SetDefault("providers.openrouter.model", "openai/gpt-4o")
	viper.SetDefault("providers.openrouter.max_tokens", 4096)
	viper.SetDefault("providers.openrouter.timeout", "60s")

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

	// Logging environment variables
	_ = viper.BindEnv("logging.level", "LOG_LEVEL")
	_ = viper.BindEnv("logging.format", "LOG_FORMAT")
}

// validate validates required configuration
func validate(config *Config) {
	// Validate provider selection
	validProviders := []string{"groq", "openrouter"}
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

	if config.MaxTokens > 16384 {
		log.Printf("Warning: GROQ_MAX_TOKENS (%d) exceeds recommended limit (16384)", config.MaxTokens)
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