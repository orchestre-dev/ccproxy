package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// EnvVar represents an environment variable with validation rules
type EnvVar struct {
	Name         string
	Description  string
	Required     bool
	DefaultValue string
	ValidateFunc func(value string) error
}

// EnvironmentVariables defines all CCProxy environment variables
var EnvironmentVariables = []EnvVar{
	// Core environment variables
	{
		Name:         "CCPROXY_SPAWN_DEPTH",
		Description:  "Tracks the depth of process spawning to prevent infinite loops",
		Required:     false,
		DefaultValue: "0",
		ValidateFunc: ValidateSpawnDepth,
	},
	{
		Name:         "CCPROXY_FOREGROUND",
		Description:  "Set to '1' to indicate the process is running in foreground mode",
		Required:     false,
		DefaultValue: "0",
		ValidateFunc: ValidateBooleanFlag,
	},
	{
		Name:         "CCPROXY_TEST_MODE",
		Description:  "Set to '1' to enable test mode which disables background spawning",
		Required:     false,
		DefaultValue: "0",
		ValidateFunc: ValidateBooleanFlag,
	},
	{
		Name:         "CCPROXY_VERSION",
		Description:  "Override the version string reported by the server",
		Required:     false,
		DefaultValue: "",
		ValidateFunc: nil, // No validation needed
	},
	{
		Name:         "CCPROXY_HOST",
		Description:  "Host address to bind the server to",
		Required:     false,
		DefaultValue: "127.0.0.1",
		ValidateFunc: ValidateHostAddress,
	},
	{
		Name:         "CCPROXY_PORT",
		Description:  "Port number to bind the server to",
		Required:     false,
		DefaultValue: "3456",
		ValidateFunc: ValidatePort,
	},
	{
		Name:         "CCPROXY_LOG",
		Description:  "Enable logging (true/false)",
		Required:     false,
		DefaultValue: "false",
		ValidateFunc: ValidateBoolean,
	},
	{
		Name:         "CCPROXY_LOG_FILE",
		Description:  "Path to the log file",
		Required:     false,
		DefaultValue: "",
		ValidateFunc: ValidateFilePath,
	},
	{
		Name:         "CCPROXY_API_KEY",
		Description:  "API key for authentication",
		Required:     false,
		DefaultValue: "",
		ValidateFunc: nil, // No validation needed
	},
	{
		Name:         "CCPROXY_PROXY_URL",
		Description:  "HTTP/HTTPS proxy URL for outbound connections",
		Required:     false,
		DefaultValue: "",
		ValidateFunc: ValidateURL,
	},
	// Test-specific environment variables
	{
		Name:         "CCPROXY_HOME",
		Description:  "Home directory for CCProxy (used in tests)",
		Required:     false,
		DefaultValue: "",
		ValidateFunc: ValidateDirectoryPath,
	},
	{
		Name:         "CCPROXY_CONFIG_DIR",
		Description:  "Configuration directory path (used in tests)",
		Required:     false,
		DefaultValue: "",
		ValidateFunc: ValidateDirectoryPath,
	},
	{
		Name:         "CCPROXY_DATA_DIR",
		Description:  "Data directory path (used in tests)",
		Required:     false,
		DefaultValue: "",
		ValidateFunc: ValidateDirectoryPath,
	},
	{
		Name:         "CCPROXY_LOG_DIR",
		Description:  "Log directory path (used in tests)",
		Required:     false,
		DefaultValue: "",
		ValidateFunc: ValidateDirectoryPath,
	},
	{
		Name:         "CCPROXY_PID_FILE",
		Description:  "PID file path (used in tests)",
		Required:     false,
		DefaultValue: "",
		ValidateFunc: ValidateFilePath,
	},
}

// ValidateSpawnDepth validates the CCPROXY_SPAWN_DEPTH environment variable
func ValidateSpawnDepth(value string) error {
	if value == "" {
		return nil
	}

	// Trim whitespace to prevent issues
	value = strings.TrimSpace(value)
	
	// Check for extremely long values that might cause overflow
	if len(value) > 10 {
		return fmt.Errorf("value too long to be a valid spawn depth")
	}

	// Use ParseInt with base 10 and 32-bit size to prevent overflow
	depth64, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		// Provide more specific error messages
		if numErr, ok := err.(*strconv.NumError); ok {
			switch numErr.Err {
			case strconv.ErrRange:
				return fmt.Errorf("value out of range for 32-bit integer")
			case strconv.ErrSyntax:
				return fmt.Errorf("invalid integer format: %q", value)
			}
		}
		return fmt.Errorf("must be a valid integer: %v", err)
	}

	depth := int(depth64)

	if depth < 0 {
		return fmt.Errorf("must be non-negative, got %d", depth)
	}

	if depth > 10 {
		return fmt.Errorf("exceeds maximum allowed depth of 10, got %d", depth)
	}

	return nil
}

// ValidateBooleanFlag validates environment variables that should be "0" or "1"
func ValidateBooleanFlag(value string) error {
	if value == "" || value == "0" || value == "1" {
		return nil
	}
	return fmt.Errorf("must be '0' or '1', got '%s'", value)
}

// ValidateBoolean validates boolean environment variables
func ValidateBoolean(value string) error {
	if value == "" {
		return nil
	}
	
	lower := strings.ToLower(value)
	if lower == "true" || lower == "false" || lower == "1" || lower == "0" {
		return nil
	}
	return fmt.Errorf("must be 'true', 'false', '1', or '0', got '%s'", value)
}

// ValidatePort validates port number
func ValidatePort(value string) error {
	if value == "" {
		return nil
	}

	port, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("must be a valid integer: %v", err)
	}

	if port < 1 || port > 65535 {
		return fmt.Errorf("must be between 1 and 65535, got %d", port)
	}

	return nil
}

// ValidateHostAddress validates host address (basic validation)
func ValidateHostAddress(value string) error {
	if value == "" {
		return nil
	}

	// Basic validation - just check it's not empty after trimming
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fmt.Errorf("host address cannot be empty or whitespace")
	}

	return nil
}

// ValidateFilePath validates file path (basic validation)
func ValidateFilePath(value string) error {
	if value == "" {
		return nil
	}

	// Check if path contains invalid characters
	if strings.ContainsAny(value, "\x00") {
		return fmt.Errorf("file path contains null characters")
	}

	return nil
}

// ValidateDirectoryPath validates directory path (basic validation)
func ValidateDirectoryPath(value string) error {
	if value == "" {
		return nil
	}

	// Check if path contains invalid characters
	if strings.ContainsAny(value, "\x00") {
		return fmt.Errorf("directory path contains null characters")
	}

	// Basic validation - just check it's not empty after trimming
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fmt.Errorf("directory path cannot be empty or whitespace")
	}

	return nil
}

// ValidateURL validates URL format (basic validation)
func ValidateURL(value string) error {
	if value == "" {
		return nil
	}

	lower := strings.ToLower(value)
	if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
		return fmt.Errorf("must start with http:// or https://")
	}

	return nil
}

// ValidateEnvironmentVariables validates all defined environment variables
func ValidateEnvironmentVariables() error {
	var errors []string

	for _, envVar := range EnvironmentVariables {
		value := os.Getenv(envVar.Name)
		
		// Check if required variable is missing
		if envVar.Required && value == "" {
			errors = append(errors, fmt.Sprintf("%s: required environment variable not set", envVar.Name))
			continue
		}

		// Skip validation if no value and not required
		if value == "" {
			continue
		}

		// Run validation function if defined
		if envVar.ValidateFunc != nil {
			if err := envVar.ValidateFunc(value); err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", envVar.Name, err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("environment variable validation failed:\n  %s", strings.Join(errors, "\n  "))
	}

	return nil
}

// GetEnvironmentVariableDocumentation returns a formatted string documenting all environment variables
func GetEnvironmentVariableDocumentation() string {
	var sb strings.Builder
	
	sb.WriteString("CCProxy Environment Variables\n")
	sb.WriteString("============================\n\n")
	
	for _, envVar := range EnvironmentVariables {
		sb.WriteString(fmt.Sprintf("%-25s %s\n", envVar.Name, envVar.Description))
		if envVar.Required {
			sb.WriteString(fmt.Sprintf("%-25s Required: Yes\n", ""))
		}
		if envVar.DefaultValue != "" {
			sb.WriteString(fmt.Sprintf("%-25s Default: %s\n", "", envVar.DefaultValue))
		}
		sb.WriteString("\n")
	}
	
	return sb.String()
}

// ValidationReport represents the result of environment variable validation
type ValidationReport struct {
	Valid   bool
	Errors  []string
	Details map[string]ValidationDetail
}

// ValidationDetail contains details about a specific environment variable validation
type ValidationDetail struct {
	Name    string
	Value   string
	Valid   bool
	Error   string
	Default string
}

// ValidateEnvironmentVariablesWithReport validates all environment variables and returns a detailed report
func ValidateEnvironmentVariablesWithReport() *ValidationReport {
	report := &ValidationReport{
		Valid:   true,
		Errors:  []string{},
		Details: make(map[string]ValidationDetail),
	}

	for _, envVar := range EnvironmentVariables {
		value := os.Getenv(envVar.Name)
		detail := ValidationDetail{
			Name:    envVar.Name,
			Value:   value,
			Valid:   true,
			Default: envVar.DefaultValue,
		}
		
		// Check if required variable is missing
		if envVar.Required && value == "" {
			detail.Valid = false
			detail.Error = "required environment variable not set"
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %s", envVar.Name, detail.Error))
			report.Valid = false
		} else if value != "" && envVar.ValidateFunc != nil {
			// Run validation function if defined
			if err := envVar.ValidateFunc(value); err != nil {
				detail.Valid = false
				detail.Error = err.Error()
				report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", envVar.Name, err))
				report.Valid = false
			}
		}
		
		report.Details[envVar.Name] = detail
	}

	return report
}