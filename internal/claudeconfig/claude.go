package claudeconfig

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/orchestre-dev/ccproxy/internal/utils"
)

// ClaudeConfig represents the ~/.claude.json configuration
type ClaudeConfig struct {
	NumStartups            int                    `json:"numStartups"`
	AutoUpdaterStatus      string                 `json:"autoUpdaterStatus"`
	UserID                 string                 `json:"userID"`
	HasCompletedOnboarding bool                   `json:"hasCompletedOnboarding"`
	LastOnboardingVersion  string                 `json:"lastOnboardingVersion"`
	Projects               map[string]interface{} `json:"projects"`
	LastActiveProject      string                 `json:"lastActiveProject,omitempty"`
	LastUpdateCheck        *time.Time             `json:"lastUpdateCheck,omitempty"`
	TelemetryEnabled       bool                   `json:"telemetryEnabled"`
	PreferredModel         string                 `json:"preferredModel,omitempty"`
	ExperimentalFeatures   []string               `json:"experimentalFeatures,omitempty"`
	CustomInstructions     string                 `json:"customInstructions,omitempty"`
	FirstLaunchDate        *time.Time             `json:"firstLaunchDate,omitempty"`
	LastLaunchDate         *time.Time             `json:"lastLaunchDate,omitempty"`
	TotalUsageMinutes      int                    `json:"totalUsageMinutes,omitempty"`
	FeedbackSubmitted      bool                   `json:"feedbackSubmitted,omitempty"`
	InstalledExtensions    []string               `json:"installedExtensions,omitempty"`
	KeyboardShortcuts      map[string]string      `json:"keyboardShortcuts,omitempty"`
	Theme                  string                 `json:"theme,omitempty"`
}

// DefaultClaudeConfig returns a new default Claude configuration
func DefaultClaudeConfig() *ClaudeConfig {
	now := time.Now()
	return &ClaudeConfig{
		NumStartups:            0,
		AutoUpdaterStatus:      "enabled",
		UserID:                 generateUserID(),
		HasCompletedOnboarding: false,
		LastOnboardingVersion:  "1.0.17",
		Projects:               make(map[string]interface{}),
		TelemetryEnabled:       true,
		FirstLaunchDate:        &now,
		Theme:                  "system",
	}
}

// Manager handles Claude configuration operations
type Manager struct {
	configPath string
	config     *ClaudeConfig
}

// NewManager creates a new Claude config manager
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".claude.json")
	return &Manager{
		configPath: configPath,
	}, nil
}

// Initialize ensures the Claude configuration exists and is valid
func (m *Manager) Initialize() error {
	// Check if config file exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		// Create default config
		m.config = DefaultClaudeConfig()
		if err := m.Save(); err != nil {
			return fmt.Errorf("failed to save default config: %w", err)
		}
		utils.GetLogger().Info("Created default ~/.claude.json configuration")
	} else {
		// Load existing config
		if err := m.Load(); err != nil {
			// If config is corrupted, create new one
			utils.GetLogger().Warnf("Failed to load existing config, creating new one: %v", err)
			m.config = DefaultClaudeConfig()
			if err := m.Save(); err != nil {
				return fmt.Errorf("failed to save new config: %w", err)
			}
		}
	}

	// Update launch info
	m.updateLaunchInfo()

	return m.Save()
}

// Load reads the Claude configuration from disk
func (m *Manager) Load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config ClaudeConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	m.config = &config
	return nil
}

// Save writes the Claude configuration to disk
func (m *Manager) Save() error {
	// Ensure directory exists
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config with pretty printing
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file with proper permissions
	if err := os.WriteFile(m.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Get returns the current Claude configuration
func (m *Manager) Get() *ClaudeConfig {
	if m.config == nil {
		m.config = DefaultClaudeConfig()
	}
	return m.config
}

// IncrementStartups increments the startup counter
func (m *Manager) IncrementStartups() error {
	if m.config == nil {
		if err := m.Load(); err != nil {
			return err
		}
	}

	m.config.NumStartups++
	return m.Save()
}

// SetUserID sets a new user ID
func (m *Manager) SetUserID(userID string) error {
	if m.config == nil {
		if err := m.Load(); err != nil {
			return err
		}
	}

	m.config.UserID = userID
	return m.Save()
}

// CompleteOnboarding marks onboarding as completed
func (m *Manager) CompleteOnboarding(version string) error {
	if m.config == nil {
		if err := m.Load(); err != nil {
			return err
		}
	}

	m.config.HasCompletedOnboarding = true
	m.config.LastOnboardingVersion = version
	return m.Save()
}

// AddProject adds a project to the configuration
func (m *Manager) AddProject(projectID string, projectData interface{}) error {
	if m.config == nil {
		if err := m.Load(); err != nil {
			return err
		}
	}

	m.config.Projects[projectID] = projectData
	m.config.LastActiveProject = projectID
	return m.Save()
}

// updateLaunchInfo updates launch-related information
func (m *Manager) updateLaunchInfo() {
	now := time.Now()
	m.config.LastLaunchDate = &now
	m.config.NumStartups++

	// Calculate usage time if last launch date exists
	if m.config.LastLaunchDate != nil {
		// This is a simplified calculation
		// In a real implementation, you'd track actual usage time
		m.config.TotalUsageMinutes += 5 // Assume 5 minutes per launch
	}
}

// generateUserID generates a random 64-character hex string
func generateUserID() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// GetConfigPath returns the path to the Claude configuration file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".claude.json"), nil
}

// Exists checks if the Claude configuration file exists
func Exists() bool {
	path, err := GetConfigPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}
