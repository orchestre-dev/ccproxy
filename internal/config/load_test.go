package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadFromFile(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("Valid config file", func(t *testing.T) {
		// Create a valid config file
		config := Config{
			Host: "0.0.0.0",
			Port: 8080,
			Log:  true,
			LogFile: "/var/log/ccproxy.log",
			Providers: []Provider{
				{
					Name:       "openai",
					APIBaseURL: "https://api.openai.com/v1",
					APIKey:     "sk-test123",
					Models:     []string{"gpt-4", "gpt-3.5-turbo"},
					Enabled:    true,
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
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
		}

		configPath := filepath.Join(tempDir, "config.json")
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			t.Fatalf("Should marshal config: %v", err)
		}

		err = os.WriteFile(configPath, data, 0644)
		if err != nil {
			t.Fatalf("Should write config file: %v", err)
		}

		// Test loading
		loadedConfig, err := LoadFromFile(configPath)
		if err != nil {
			t.Errorf("Should load config without error: %v", err)
		}
		if loadedConfig == nil {
			t.Fatal("Should return config")
		}

		// Verify loaded values
		if loadedConfig.Host != "0.0.0.0" {
			t.Errorf("Expected host '0.0.0.0', got '%s'", loadedConfig.Host)
		}
		if loadedConfig.Port != 8080 {
			t.Errorf("Expected port 8080, got %d", loadedConfig.Port)
		}
		if !loadedConfig.Log {
			t.Error("Expected log to be true")
		}
		if len(loadedConfig.Providers) != 1 {
			t.Errorf("Expected 1 provider, got %d", len(loadedConfig.Providers))
		}
		if len(loadedConfig.Routes) != 1 {
			t.Errorf("Expected 1 route, got %d", len(loadedConfig.Routes))
		}
	})

	t.Run("File not found", func(t *testing.T) {
		nonExistentPath := filepath.Join(tempDir, "nonexistent.json")
		
		config, err := LoadFromFile(nonExistentPath)
		if err == nil {
			t.Error("Should error for non-existent file")
		}
		if config != nil {
			t.Error("Should return nil config")
		}
		if !strings.Contains(err.Error(), "failed to read config file") {
			t.Errorf("Expected read failure error, got: %v", err)
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		invalidJSONPath := filepath.Join(tempDir, "invalid.json")
		invalidJSON := `{
			"host": "127.0.0.1",
			"port": 3456,
			"invalid": json syntax
		}`

		err := os.WriteFile(invalidJSONPath, []byte(invalidJSON), 0644)
		if err != nil {
			t.Fatalf("Should write invalid JSON file: %v", err)
		}

		config, err := LoadFromFile(invalidJSONPath)
		if err == nil {
			t.Error("Should error for invalid JSON")
		}
		if config != nil {
			t.Error("Should return nil config")
		}
		if !strings.Contains(err.Error(), "failed to parse config file") {
			t.Errorf("Expected parse failure error, got: %v", err)
		}
	})

	t.Run("Invalid config", func(t *testing.T) {
		invalidConfigPath := filepath.Join(tempDir, "invalid_config.json")
		invalidConfig := `{
			"host": "127.0.0.1",
			"port": 0,
			"providers": [
				{
					"name": "",
					"api_base_url": "https://api.openai.com/v1"
				}
			]
		}`

		err := os.WriteFile(invalidConfigPath, []byte(invalidConfig), 0644)
		if err != nil {
			t.Fatalf("Should write invalid config file: %v", err)
		}

		config, err := LoadFromFile(invalidConfigPath)
		if err == nil {
			t.Error("Should error for invalid config")
		}
		if config != nil {
			t.Error("Should return nil config")
		}
		if !strings.Contains(err.Error(), "invalid configuration") {
			t.Errorf("Expected config validation error, got: %v", err)
		}
	})

	t.Run("Config with defaults applied", func(t *testing.T) {
		// Create config with missing host/port
		partialConfig := `{
			"log": true,
			"providers": []
		}`

		partialPath := filepath.Join(tempDir, "partial.json")
		err := os.WriteFile(partialPath, []byte(partialConfig), 0644)
		if err != nil {
			t.Fatalf("Should write partial config: %v", err)
		}

		config, err := LoadFromFile(partialPath)
		if err != nil {
			t.Errorf("Should load partial config: %v", err)
		}

		// Check defaults were applied
		if config.Host != "127.0.0.1" {
			t.Errorf("Expected default host '127.0.0.1', got '%s'", config.Host)
		}
		if config.Port != 3456 {
			t.Errorf("Expected default port 3456, got %d", config.Port)
		}
		if !config.Log {
			t.Error("Expected log to be true")
		}
	})

	t.Run("Empty file", func(t *testing.T) {
		emptyPath := filepath.Join(tempDir, "empty.json")
		
		err := os.WriteFile(emptyPath, []byte(""), 0644)
		if err != nil {
			t.Fatalf("Should write empty file: %v", err)
		}

		config, err := LoadFromFile(emptyPath)
		if err == nil {
			t.Error("Should error for empty file")
		}
		if config != nil {
			t.Error("Should return nil config")
		}
		if !strings.Contains(err.Error(), "failed to parse config file") {
			t.Errorf("Expected parse failure error, got: %v", err)
		}
	})
}