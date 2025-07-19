package claudeconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDefaultClaudeConfig(t *testing.T) {
	config := DefaultClaudeConfig()
	
	// Verify default values
	if config.NumStartups != 0 {
		t.Errorf("Expected NumStartups to be 0, got %d", config.NumStartups)
	}
	if config.AutoUpdaterStatus != "enabled" {
		t.Errorf("Expected AutoUpdaterStatus to be 'enabled', got %s", config.AutoUpdaterStatus)
	}
	if config.UserID == "" {
		t.Error("Expected UserID to be generated")
	}
	if len(config.UserID) != 64 {
		t.Errorf("Expected UserID to be 64 characters, got %d", len(config.UserID))
	}
	if config.HasCompletedOnboarding != false {
		t.Error("Expected HasCompletedOnboarding to be false")
	}
	if config.LastOnboardingVersion != "1.0.17" {
		t.Errorf("Expected LastOnboardingVersion to be '1.0.17', got %s", config.LastOnboardingVersion)
	}
	if config.Projects == nil {
		t.Error("Expected Projects to be initialized")
	}
	if config.TelemetryEnabled != true {
		t.Error("Expected TelemetryEnabled to be true")
	}
	if config.FirstLaunchDate == nil {
		t.Error("Expected FirstLaunchDate to be set")
	}
	if config.Theme != "system" {
		t.Errorf("Expected Theme to be 'system', got %s", config.Theme)
	}
}

func TestGenerateUserID(t *testing.T) {
	// Test multiple generations to ensure uniqueness
	ids := make(map[string]bool)
	for i := 0; i < 10; i++ {
		id := generateUserID()
		if len(id) != 64 {
			t.Errorf("Expected UserID to be 64 characters, got %d", len(id))
		}
		if ids[id] {
			t.Error("Generated duplicate UserID")
		}
		ids[id] = true
		
		// Verify it's valid hex
		if !isValidHex(id) {
			t.Errorf("UserID is not valid hex: %s", id)
		}
	}
}

func TestManager_Initialize(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".claude.json")
	
	// Override home directory for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)
	
	// Create manager
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	// Initialize (should create new config)
	err = manager.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}
	
	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}
	
	// Verify content
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	
	var config ClaudeConfig
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}
	
	// Verify startup count was incremented
	if config.NumStartups != 1 {
		t.Errorf("Expected NumStartups to be 1, got %d", config.NumStartups)
	}
	
	// Initialize again (should load existing config)
	err = manager.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize with existing config: %v", err)
	}
	
	// Reload and verify startup count incremented again
	data, _ = os.ReadFile(configPath)
	json.Unmarshal(data, &config)
	if config.NumStartups != 2 {
		t.Errorf("Expected NumStartups to be 2, got %d", config.NumStartups)
	}
}

func TestManager_Operations(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", os.Getenv("HOME"))
	
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	// Initialize
	manager.Initialize()
	
	// Test IncrementStartups
	initialStartups := manager.Get().NumStartups
	err = manager.IncrementStartups()
	if err != nil {
		t.Errorf("Failed to increment startups: %v", err)
	}
	if manager.Get().NumStartups != initialStartups+1 {
		t.Error("Startups not incremented")
	}
	
	// Test SetUserID
	newUserID := "test-user-id-1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	err = manager.SetUserID(newUserID)
	if err != nil {
		t.Errorf("Failed to set user ID: %v", err)
	}
	if manager.Get().UserID != newUserID {
		t.Error("User ID not updated")
	}
	
	// Test CompleteOnboarding
	err = manager.CompleteOnboarding("1.0.18")
	if err != nil {
		t.Errorf("Failed to complete onboarding: %v", err)
	}
	if !manager.Get().HasCompletedOnboarding {
		t.Error("Onboarding not marked as completed")
	}
	if manager.Get().LastOnboardingVersion != "1.0.18" {
		t.Error("Onboarding version not updated")
	}
	
	// Test AddProject
	projectData := map[string]interface{}{
		"name": "test-project",
		"path": "/tmp/test-project",
	}
	err = manager.AddProject("project-123", projectData)
	if err != nil {
		t.Errorf("Failed to add project: %v", err)
	}
	if len(manager.Get().Projects) != 1 {
		t.Error("Project not added")
	}
	if manager.Get().LastActiveProject != "project-123" {
		t.Error("Last active project not updated")
	}
}

func TestManager_CorruptedConfig(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".claude.json")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", os.Getenv("HOME"))
	
	// Write corrupted config
	os.WriteFile(configPath, []byte("{ invalid json }"), 0600)
	
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	// Initialize should handle corrupted config
	err = manager.Initialize()
	if err != nil {
		t.Fatalf("Failed to handle corrupted config: %v", err)
	}
	
	// Verify new config was created
	if manager.Get().UserID == "" {
		t.Error("New config not created")
	}
	
	// Verify file is now valid JSON
	data, _ := os.ReadFile(configPath)
	var config ClaudeConfig
	if err := json.Unmarshal(data, &config); err != nil {
		t.Error("Config file is still invalid JSON")
	}
}

func TestGetConfigPath(t *testing.T) {
	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}
	
	if !strings.HasSuffix(path, ".claude.json") {
		t.Errorf("Config path doesn't end with .claude.json: %s", path)
	}
}

func TestExists(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".claude.json")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", os.Getenv("HOME"))
	
	// Should not exist initially
	if Exists() {
		t.Error("Config should not exist initially")
	}
	
	// Create file
	os.WriteFile(configPath, []byte("{}"), 0600)
	
	// Should exist now
	if !Exists() {
		t.Error("Config should exist after creation")
	}
}

func TestUpdateLaunchInfo(t *testing.T) {
	config := DefaultClaudeConfig()
	manager := &Manager{config: config}
	
	// Set initial values
	initialStartups := config.NumStartups
	initialMinutes := config.TotalUsageMinutes
	
	// Update launch info
	manager.updateLaunchInfo()
	
	// Verify updates
	if config.NumStartups != initialStartups+1 {
		t.Error("NumStartups not incremented")
	}
	if config.LastLaunchDate == nil {
		t.Error("LastLaunchDate not set")
	}
	if config.TotalUsageMinutes != initialMinutes+5 {
		t.Errorf("Expected TotalUsageMinutes to be %d, got %d", 
			initialMinutes+5, config.TotalUsageMinutes)
	}
	
	// Verify LastLaunchDate is recent
	timeDiff := time.Since(*config.LastLaunchDate)
	if timeDiff > time.Second {
		t.Error("LastLaunchDate is not recent")
	}
}

// Helper function to check if a string is valid hex
func isValidHex(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}