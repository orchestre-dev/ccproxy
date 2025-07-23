package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Service handles configuration loading and management
type Service struct {
	config *Config
	viper  *viper.Viper
	mu     sync.RWMutex
}

// NewService creates a new configuration service
func NewService() *Service {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Set up configuration search paths
	v.SetConfigName("config")
	v.SetConfigType("json")

	// Add configuration paths in order of priority
	// 1. Current directory
	v.AddConfigPath(".")

	// 2. User's home directory under .ccproxy
	if home, err := os.UserHomeDir(); err == nil {
		v.AddConfigPath(filepath.Join(home, ".ccproxy"))
	}

	// 3. System configuration directory
	v.AddConfigPath("/etc/ccproxy")

	// Enable environment variable binding
	v.SetEnvPrefix("CCPROXY")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	return &Service{
		viper:  v,
		config: DefaultConfig(),
	}
}

// Load reads and parses the configuration from all sources
func (s *Service) Load() error {
	// Step 1: Load defaults (already set in NewService)

	// Step 2: Load from JSON config file if exists
	if err := s.viper.ReadInConfig(); err != nil {
		// It's okay if config file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Step 3: Load from .env file if exists
	if err := s.loadEnvFile(); err != nil {
		// Log but don't fail if .env file doesn't exist
		if !os.IsNotExist(err) {
			return fmt.Errorf("error loading .env file: %w", err)
		}
	}

	// Step 4: Environment variables are automatically loaded by Viper

	// Step 5: Unmarshal into config struct with custom decoder
	decoderConfig := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeHookFunc(time.RFC3339),
			mapstructure.StringToTimeDurationHookFunc(),
		),
		Result:           s.config,
		WeaklyTypedInput: true,
	}

	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return fmt.Errorf("error creating decoder: %w", err)
	}

	if err := decoder.Decode(s.viper.AllSettings()); err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Step 6: Validate configuration
	if err := s.config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Step 7: Apply special environment variable mappings
	s.applyEnvironmentMappings()

	return nil
}

// Get returns the current configuration (returns a copy to prevent race conditions)
func (s *Service) Get() *Config {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a deep copy of the config
	configCopy := *s.config

	// Deep copy the slices
	configCopy.Providers = make([]Provider, len(s.config.Providers))
	copy(configCopy.Providers, s.config.Providers)

	// Deep copy the routes map
	configCopy.Routes = make(map[string]Route)
	for k, v := range s.config.Routes {
		configCopy.Routes[k] = v
	}

	return &configCopy
}

// SetConfig sets the configuration (mainly for testing)
func (s *Service) SetConfig(cfg *Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = cfg
}

// Validate checks if the configuration is valid
func (s *Service) Validate() error {
	// Validate port
	if s.config.Port < 1 || s.config.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", s.config.Port)
	}

	// Validate providers
	providerNames := make(map[string]bool)
	for _, provider := range s.config.Providers {
		if provider.Name == "" {
			return fmt.Errorf("provider name cannot be empty")
		}
		if providerNames[provider.Name] {
			return fmt.Errorf("duplicate provider name: %s", provider.Name)
		}
		providerNames[provider.Name] = true

		if provider.APIBaseURL == "" {
			return fmt.Errorf("provider %s: api_base_url cannot be empty", provider.Name)
		}
	}

	// Validate routes
	for routeName, route := range s.config.Routes {
		if route.Provider == "" {
			return fmt.Errorf("route %s: provider cannot be empty", routeName)
		}
		if route.Model == "" {
			return fmt.Errorf("route %s: model cannot be empty", routeName)
		}
	}

	return nil
}

// Reload reloads the configuration
func (s *Service) Reload() error {
	// Create a new viper instance to avoid conflicts
	newService := NewService()
	if err := newService.Load(); err != nil {
		return err
	}

	// Update the configuration atomically
	s.config = newService.config
	s.viper = newService.viper

	return nil
}

// Save saves the current configuration to file
func (s *Service) Save() error {
	// Ensure config directory exists
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot get home directory: %w", err)
	}

	configDir := filepath.Join(home, ".ccproxy")
	if err := os.MkdirAll(configDir, 0750); err != nil {
		return fmt.Errorf("cannot create config directory: %w", err)
	}

	// Marshal configuration to JSON
	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal config: %w", err)
	}

	// Write to file
	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("cannot write config file: %w", err)
	}

	return nil
}

// GetProvider returns a provider by name
func (s *Service) GetProvider(name string) (*Provider, error) {
	for _, provider := range s.config.Providers {
		if provider.Name == name {
			return &provider, nil
		}
	}
	return nil, fmt.Errorf("provider not found: %s", name)
}

// UpdateProvider updates a provider by name
func (s *Service) UpdateProvider(name string, provider *Provider) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	provider.UpdatedAt = time.Now()

	// Find and update existing provider
	for i, p := range s.config.Providers {
		if p.Name == name {
			if provider.CreatedAt.IsZero() {
				provider.CreatedAt = p.CreatedAt
			}
			s.config.Providers[i] = *provider
			return s.Save()
		}
	}

	return fmt.Errorf("provider not found: %s", name)
}

// SaveProvider saves a new provider
func (s *Service) SaveProvider(provider *Provider) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if provider already exists
	for _, p := range s.config.Providers {
		if p.Name == provider.Name {
			return fmt.Errorf("provider already exists: %s", provider.Name)
		}
	}

	// Add new provider
	if provider.CreatedAt.IsZero() {
		provider.CreatedAt = time.Now()
	}
	provider.UpdatedAt = time.Now()
	s.config.Providers = append(s.config.Providers, *provider)
	return s.Save()
}

// DeleteProvider removes a provider by name
func (s *Service) DeleteProvider(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	providers := make([]Provider, 0, len(s.config.Providers))
	found := false

	for _, p := range s.config.Providers {
		if p.Name != name {
			providers = append(providers, p)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("provider not found: %s", name)
	}

	s.config.Providers = providers
	return s.Save()
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	v.SetDefault("host", "127.0.0.1")
	v.SetDefault("port", 3456)
	v.SetDefault("log", false)
	v.SetDefault("log_file", "")
	// Don't set default routes - let user configure them
}

// loadEnvFile loads environment variables from .env file
func (s *Service) loadEnvFile() error {
	// Check common locations for .env file
	locations := []string{
		".env",
		filepath.Join(".ccproxy", ".env"),
	}

	if home, err := os.UserHomeDir(); err == nil {
		locations = append(locations, filepath.Join(home, ".ccproxy", ".env"))
	}

	for _, loc := range locations {
		data, err := os.ReadFile(loc) // #nosec G304 -- Reading from known safe .env file locations (current dir, .ccproxy dir, and user home)
		if err == nil {
			// Parse .env file
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}

				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					// Remove quotes if present
					value = strings.Trim(value, `"'`)
					_ = os.Setenv(key, value) // Best effort env var setting
				}
			}
			return nil
		}
	}

	return os.ErrNotExist
}

// applyEnvironmentMappings applies special environment variable mappings
func (s *Service) applyEnvironmentMappings() {
	// Map common environment variables to config
	if apiKey := os.Getenv("APIKEY"); apiKey != "" {
		s.config.APIKey = apiKey
	}

	// Map CCPROXY_API_KEY (common alternate name)
	if apiKey := os.Getenv("CCPROXY_API_KEY"); apiKey != "" {
		s.config.APIKey = apiKey
	}

	// PORT env var is already handled by viper's AutomaticEnv

	// Check for corporate proxy settings
	proxyVars := []string{"HTTPS_PROXY", "https_proxy", "httpsProxy", "PROXY_URL"}
	for _, v := range proxyVars {
		if proxy := os.Getenv(v); proxy != "" {
			s.config.ProxyURL = proxy
			break
		}
	}

	// Map provider-specific API keys
	providerEnvMap := map[string]string{
		"anthropic":   "ANTHROPIC_API_KEY",
		"openai":      "OPENAI_API_KEY",
		"gemini":      "GEMINI_API_KEY",
		"google":      "GOOGLE_API_KEY", // Alternate for Gemini
		"deepseek":    "DEEPSEEK_API_KEY",
		"openrouter":  "OPENROUTER_API_KEY",
		"groq":        "GROQ_API_KEY",
		"mistral":     "MISTRAL_API_KEY",
		"xai":         "XAI_API_KEY",
		"grok":        "GROK_API_KEY", // Alternate for XAI
		"ollama":      "OLLAMA_API_KEY",
		"bedrock":     "AWS_ACCESS_KEY_ID", // AWS Bedrock uses AWS credentials
	}

	// Apply provider-specific environment variables
	for i, provider := range s.config.Providers {
		// Check if there's a provider-specific environment variable
		if envVar, exists := providerEnvMap[strings.ToLower(provider.Name)]; exists {
			if apiKey := os.Getenv(envVar); apiKey != "" {
				s.config.Providers[i].APIKey = apiKey
			}
		}

		// Special handling for AWS Bedrock which needs two credentials
		if strings.ToLower(provider.Name) == "bedrock" {
			if accessKey := os.Getenv("AWS_ACCESS_KEY_ID"); accessKey != "" {
				s.config.Providers[i].APIKey = accessKey
				// Also check for secret key and region
				if secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY"); secretKey != "" {
					// Store secret key in a custom field or append to API key
					// This depends on how the Bedrock provider is implemented
					s.config.Providers[i].APIKey = accessKey + ":" + secretKey
				}
			}
		}

		// Also check for indexed environment variables for backward compatibility
		// This allows both ANTHROPIC_API_KEY and CCPROXY_PROVIDERS_0_API_KEY to work
		indexedKey := fmt.Sprintf("CCPROXY_PROVIDERS_%d_API_KEY", i)
		if apiKey := os.Getenv(indexedKey); apiKey != "" {
			s.config.Providers[i].APIKey = apiKey
		}
	}
}
