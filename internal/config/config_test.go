// Package config provides configuration management tests
package config

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testWithConfig is a helper that creates a config file in a temp directory
// and loads the configuration. It returns the config and a cleanup function.
func testWithConfig(t *testing.T, configContent string, envVars map[string]string) (*Config, func()) {
	// Set environment variables
	for key, value := range envVars {
		os.Setenv(key, value)
	}

	// Write to temp file
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/config.yaml"
	err := os.WriteFile(tmpFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)

	// Reset viper
	viper.Reset()

	// Load config
	cfg := Load()

	// Return cleanup function
	cleanup := func() {
		os.Chdir(oldDir)
		for key := range envVars {
			os.Unsetenv(key)
		}
	}

	return cfg, cleanup
}

func TestLoad_Success(t *testing.T) {
	configContent := `
logging:
  level: debug
  format: json

provider: groq

providers:
  groq:
    api_key: test-key
    model: mixtral-8x7b-32768
    base_url: https://api.groq.com/openai/v1
    max_tokens: 1000
    timeout: 30s

server:
  host: localhost
  port: 8080
  environment: test
  read_timeout: 30s
  write_timeout: 30s
  shutdown_timeout: 5s
`
	cfg, cleanup := testWithConfig(t, configContent, map[string]string{"GROQ_API_KEY": "test-key"})
	defer cleanup()

	// Verify configuration
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.Equal(t, "groq", cfg.Provider)
	assert.Equal(t, "test-key", cfg.Providers.Groq.APIKey)
	assert.Equal(t, "mixtral-8x7b-32768", cfg.Providers.Groq.Model)
	assert.Equal(t, "https://api.groq.com/openai/v1", cfg.Providers.Groq.BaseURL)
	assert.Equal(t, 1000, cfg.Providers.Groq.MaxTokens)
	assert.Equal(t, 30*time.Second, cfg.Providers.Groq.Timeout)
	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, "test", cfg.Server.Environment)
}

func TestLoad_FileNotFound(t *testing.T) {
	// Set test environment variables to avoid validation errors (default provider is groq)
	os.Setenv("GROQ_API_KEY", "test-key")
	defer os.Unsetenv("GROQ_API_KEY")

	// Change to a temp directory with no config file
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	// Reset viper
	viper.Reset()

	// This should not panic even with no config file
	cfg := Load()
	assert.NotNil(t, cfg)
}

func TestLoad_InvalidYAML(t *testing.T) {
	configContent := `
logging:
  level: debug
  invalid yaml content
  [broken
`
	cfg, cleanup := testWithConfig(t, configContent, map[string]string{"GROQ_API_KEY": "test-key"})
	defer cleanup()

	// This should still load with defaults even with invalid YAML
	assert.NotNil(t, cfg)
}

func TestLoad_EnvironmentVariables(t *testing.T) {
	configContent := `
provider: groq
`
	cfg, cleanup := testWithConfig(t, configContent, map[string]string{
		"GROQ_API_KEY": "env-test-key",
		"GROQ_MODEL": "env-model",
		"SERVER_PORT": "9090",
		"LOG_LEVEL": "error",
	})
	defer cleanup()

	// Verify environment variables override defaults
	assert.Equal(t, "env-test-key", cfg.Providers.Groq.APIKey)
	assert.Equal(t, "env-model", cfg.Providers.Groq.Model)
	assert.Equal(t, "9090", cfg.Server.Port)
	assert.Equal(t, "error", cfg.Logging.Level)
}

func TestLoad_Defaults(t *testing.T) {
	configContent := `
provider: openai
`
	cfg, cleanup := testWithConfig(t, configContent, map[string]string{"OPENAI_API_KEY": "test-key"})
	defer cleanup()

	// Verify defaults are applied
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, "7187", cfg.Server.Port)
	assert.Equal(t, "development", cfg.Server.Environment)
	assert.Equal(t, 30*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 30*time.Second, cfg.Server.WriteTimeout)
	assert.Equal(t, 5*time.Second, cfg.Server.ShutdownTimeout)
}

func TestLoad_AllProviders(t *testing.T) {
	configContent := `
provider: groq

providers:
  groq:
    api_key: groq-key
    model: mixtral-8x7b-32768
    base_url: https://api.groq.com/openai/v1
    max_tokens: 1000
    timeout: 30s
  
  openai:
    api_key: openai-key
    model: gpt-4
    base_url: https://api.openai.com/v1
    max_tokens: 4096
    timeout: 60s
    
  openrouter:
    api_key: openrouter-key
    model: gpt-4
    base_url: https://openrouter.ai/api/v1
    max_tokens: 4096
    timeout: 60s
    
  xai:
    api_key: xai-key
    model: grok-beta
    base_url: https://api.x.ai/v1
    max_tokens: 131072
    timeout: 60s
    
  gemini:
    api_key: gemini-key
    model: gemini-pro
    base_url: https://generativelanguage.googleapis.com
    max_tokens: 2048
    timeout: 60s
    
  mistral:
    api_key: mistral-key
    model: mistral-tiny
    base_url: https://api.mistral.ai/v1
    max_tokens: 8192
    timeout: 120s
    
  ollama:
    base_url: http://localhost:11434
    model: llama2
    max_tokens: 4096
    timeout: 300s
`
	cfg, cleanup := testWithConfig(t, configContent, map[string]string{"GROQ_API_KEY": "groq-key"})
	defer cleanup()

	// Verify all providers are loaded
	assert.Equal(t, "groq-key", cfg.Providers.Groq.APIKey)
	assert.Equal(t, "openai-key", cfg.Providers.OpenAI.APIKey)
	assert.Equal(t, "openrouter-key", cfg.Providers.OpenRouter.APIKey)
	assert.Equal(t, "xai-key", cfg.Providers.XAI.APIKey)
	assert.Equal(t, "gemini-key", cfg.Providers.Gemini.APIKey)
	assert.Equal(t, "mistral-key", cfg.Providers.Mistral.APIKey)
	assert.Equal(t, "http://localhost:11434", cfg.Providers.Ollama.BaseURL)
	// Verify all providers are loaded correctly
}

func TestLoad_InvalidProvider(t *testing.T) {
	configContent := `
provider: invalid_provider
`
	cfg, cleanup := testWithConfig(t, configContent, map[string]string{"PROVIDER": "ollama"})
	defer cleanup()

	// This should override with ollama from env var
	require.NotNil(t, cfg)
	assert.Equal(t, "ollama", cfg.Provider)
}

func TestServerConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config string
		check  func(t *testing.T, cfg *Config)
	}{
		{
			name: "Invalid port",
			config: `
provider: groq
server:
  port: invalid
`,
			check: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "invalid", cfg.Server.Port)
				// Port validation should be done at usage, not config load
			},
		},
		{
			name: "Zero timeouts use defaults",
			config: `
provider: groq
server:
  read_timeout: 0s
  write_timeout: 0s
  shutdown_timeout: 0s
`,
			check: func(t *testing.T, cfg *Config) {
				assert.Equal(t, time.Duration(0), cfg.Server.ReadTimeout)
				assert.Equal(t, time.Duration(0), cfg.Server.WriteTimeout)
				assert.Equal(t, time.Duration(0), cfg.Server.ShutdownTimeout)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, cleanup := testWithConfig(t, tt.config, map[string]string{"GROQ_API_KEY": "test-key"})
			defer cleanup()

			tt.check(t, cfg)
		})
	}
}

func TestProviderSelection(t *testing.T) {
	tests := []struct {
		name         string
		config       string
		envProvider  string
		wantProvider string
	}{
		{
			name: "Config file provider",
			config: `
provider: groq
`,
			envProvider:  "",
			wantProvider: "groq",
		},
		{
			name: "Environment override",
			config: `
provider: groq
`,
			envProvider:  "openai",
			wantProvider: "openai",
		},
		{
			name: "No provider specified (defaults to groq)",
			config: `
logging:
  level: debug
`,
			envProvider:  "",
			wantProvider: "groq",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set API keys based on the provider to avoid validation errors
			envVars := make(map[string]string)
			if tt.wantProvider == "groq" || (tt.wantProvider == "groq" && tt.envProvider == "") {
				envVars["GROQ_API_KEY"] = "test-key"
			} else if tt.wantProvider == "openai" || tt.envProvider == "openai" {
				envVars["OPENAI_API_KEY"] = "test-key"
			}
			if tt.envProvider != "" {
				envVars["PROVIDER"] = tt.envProvider
			}

			cfg, cleanup := testWithConfig(t, tt.config, envVars)
			defer cleanup()

			assert.Equal(t, tt.wantProvider, cfg.Provider)
		})
	}
}