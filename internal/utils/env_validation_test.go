package utils

import (
	"os"
	"strings"
	"testing"
)

func TestValidateSpawnDepth(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
		errMsg  string
	}{
		{"empty value", "", false, ""},
		{"valid zero", "0", false, ""},
		{"valid positive", "5", false, ""},
		{"valid max", "10", false, ""},
		{"valid with spaces", " 5 ", false, ""}, // Should be trimmed and valid
		{"negative value", "-1", true, "must be non-negative"},
		{"exceeds max", "11", true, "exceeds maximum allowed depth"},
		{"exceeds max large", "100", true, "exceeds maximum allowed depth"},
		{"invalid integer", "abc", true, "invalid integer format"},
		{"invalid float", "3.14", true, "invalid integer format"},
		{"very large number", "2147483648", true, "value out of range"}, // Exceeds int32
		{"very negative number", "-2147483649", true, "value too long"}, // Below int32, caught by length check
		{"extremely long value", "12345678901234567890", true, "value too long"},
		{"invalid hex", "0x10", true, "invalid integer format"},
		{"invalid octal", "010", false, ""}, // Will be parsed as 10, not 8
		{"empty spaces only", "   ", true, "invalid integer format"},
		{"plus sign", "+5", false, ""}, // Should be valid
		{"double negative", "--5", true, "invalid integer format"},
		{"unicode digit", "５", true, "invalid integer format"}, // Full-width digit
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSpawnDepth(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSpawnDepth(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateSpawnDepth(%q) error = %v, want error containing %q", tt.value, err, tt.errMsg)
			}
		})
	}
}

func TestValidateBooleanFlag(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty value", "", false},
		{"valid 0", "0", false},
		{"valid 1", "1", false},
		{"invalid 2", "2", true},
		{"invalid true", "true", true},
		{"invalid false", "false", true},
		{"invalid text", "yes", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBooleanFlag(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBooleanFlag(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidateBoolean(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty value", "", false},
		{"valid true", "true", false},
		{"valid false", "false", false},
		{"valid TRUE", "TRUE", false},
		{"valid FALSE", "FALSE", false},
		{"valid 1", "1", false},
		{"valid 0", "0", false},
		{"invalid yes", "yes", true},
		{"invalid no", "no", true},
		{"invalid text", "maybe", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBoolean(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBoolean(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePort(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty value", "", false},
		{"valid min", "1", false},
		{"valid common", "8080", false},
		{"valid max", "65535", false},
		{"invalid zero", "0", true},
		{"invalid negative", "-1", true},
		{"invalid too high", "65536", true},
		{"invalid text", "http", true},
		{"invalid float", "80.5", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePort(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePort(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidateHostAddress(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty value", "", false},
		{"valid localhost", "localhost", false},
		{"valid IP", "127.0.0.1", false},
		{"valid all interfaces", "0.0.0.0", false},
		{"valid domain", "example.com", false},
		{"invalid empty after trim", "   ", true},
		{"invalid just spaces", "    ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHostAddress(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHostAddress(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidateFilePath(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty value", "", false},
		{"valid path", "/var/log/ccproxy.log", false},
		{"valid relative", "./logs/app.log", false},
		{"valid windows", "C:\\logs\\app.log", false},
		{"invalid null char", "/var/log\x00/app.log", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilePath(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilePath(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty value", "", false},
		{"valid http", "http://proxy.example.com", false},
		{"valid https", "https://proxy.example.com:8080", false},
		{"valid HTTP uppercase", "HTTP://proxy.example.com", false},
		{"invalid no scheme", "proxy.example.com", true},
		{"invalid ftp", "ftp://proxy.example.com", true},
		{"invalid socks", "socks5://proxy.example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidateDirectoryPath(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty value", "", false},
		{"valid path", "/var/log/ccproxy", false},
		{"valid relative", "./logs", false},
		{"valid windows", "C:\\logs", false},
		{"valid with spaces", "/path with spaces/logs", false},
		{"invalid null char", "/var/log\x00/ccproxy", true},
		{"invalid just spaces", "   ", true},
		{"invalid empty after trim", "\t\n\r ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDirectoryPath(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDirectoryPath(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidateEnvironmentVariables(t *testing.T) {
	// Save current environment
	oldEnv := make(map[string]string)
	for _, envVar := range EnvironmentVariables {
		oldEnv[envVar.Name] = os.Getenv(envVar.Name)
		os.Unsetenv(envVar.Name)
	}
	defer func() {
		// Restore environment
		for name, value := range oldEnv {
			if value != "" {
				os.Setenv(name, value)
			}
		}
	}()

	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name:    "all defaults valid",
			setup:   func() {},
			wantErr: false,
		},
		{
			name: "valid spawn depth",
			setup: func() {
				os.Setenv("CCPROXY_SPAWN_DEPTH", "5")
			},
			wantErr: false,
		},
		{
			name: "invalid spawn depth",
			setup: func() {
				os.Setenv("CCPROXY_SPAWN_DEPTH", "20")
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			setup: func() {
				os.Setenv("CCPROXY_PORT", "70000")
			},
			wantErr: true,
		},
		{
			name: "multiple invalid",
			setup: func() {
				os.Setenv("CCPROXY_SPAWN_DEPTH", "-1")
				os.Setenv("CCPROXY_PORT", "0")
				os.Setenv("CCPROXY_FOREGROUND", "yes")
			},
			wantErr: true,
		},
		{
			name: "all valid",
			setup: func() {
				os.Setenv("CCPROXY_SPAWN_DEPTH", "2")
				os.Setenv("CCPROXY_PORT", "8080")
				os.Setenv("CCPROXY_FOREGROUND", "1")
				os.Setenv("CCPROXY_LOG", "true")
				os.Setenv("CCPROXY_HOST", "0.0.0.0")
			},
			wantErr: false,
		},
		{
			name: "test directory paths",
			setup: func() {
				os.Setenv("CCPROXY_HOME", "/tmp/ccproxy")
				os.Setenv("CCPROXY_CONFIG_DIR", "/tmp/ccproxy/config")
				os.Setenv("CCPROXY_DATA_DIR", "/tmp/ccproxy/data")
				os.Setenv("CCPROXY_LOG_DIR", "/tmp/ccproxy/logs")
				os.Setenv("CCPROXY_PID_FILE", "/tmp/ccproxy/ccproxy.pid")
			},
			wantErr: false,
		},
		{
			name: "invalid directory path",
			setup: func() {
				os.Setenv("CCPROXY_HOME", "   ")
			},
			wantErr: true,
		},
		{
			name: "overflow spawn depth",
			setup: func() {
				os.Setenv("CCPROXY_SPAWN_DEPTH", "2147483648")
			},
			wantErr: true,
		},
		{
			name: "whitespace spawn depth",
			setup: func() {
				os.Setenv("CCPROXY_SPAWN_DEPTH", " 5 ")
			},
			wantErr: false, // Should be valid after trimming
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			for _, envVar := range EnvironmentVariables {
				os.Unsetenv(envVar.Name)
			}
			
			// Run setup
			tt.setup()
			
			// Test
			err := ValidateEnvironmentVariables()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEnvironmentVariables() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}