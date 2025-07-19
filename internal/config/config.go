package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Service handles configuration loading and management
type Service struct {
	config *Config
	viper  *viper.Viper
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
	
	// Step 5: Unmarshal into config struct
	if err := s.viper.Unmarshal(s.config); err != nil {
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

// Get returns the current configuration
func (s *Service) Get() *Config {
	return s.config
}

// SetConfig sets the configuration (mainly for testing)
func (s *Service) SetConfig(cfg *Config) {
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
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("cannot create config directory: %w", err)
	}
	
	// Marshal configuration to JSON
	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal config: %w", err)
	}
	
	// Write to file
	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
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
		data, err := os.ReadFile(loc)
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
					os.Setenv(key, value)
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
	
	if port := os.Getenv("PORT"); port != "" {
		// PORT env var is already handled by viper's AutomaticEnv
		// This is just for documentation purposes
	}
	
	// Check for corporate proxy settings
	proxyVars := []string{"HTTPS_PROXY", "https_proxy", "httpsProxy", "PROXY_URL"}
	for _, v := range proxyVars {
		if proxy := os.Getenv(v); proxy != "" {
			s.config.ProxyURL = proxy
			break
		}
	}
}