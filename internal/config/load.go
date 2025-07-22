package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadFromFile loads configuration from a specific file
func LoadFromFile(path string) (*Config, error) {
	// Read the file
	data, err := os.ReadFile(path) // #nosec G304 -- Path is provided by the user via CLI flag and is expected to be a trusted configuration file
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults for empty fields before validation
	if cfg.Host == "" {
		cfg.Host = "127.0.0.1"
	}
	if cfg.Port == 0 {
		cfg.Port = 3456
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}
