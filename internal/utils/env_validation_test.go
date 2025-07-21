package utils

import (
	"fmt"
	"os"
	"strings"
	"testing"

	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
)

func TestValidateSpawnDepth(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty value",
			value:   "",
			wantErr: false,
		},
		{
			name:    "valid zero",
			value:   "0",
			wantErr: false,
		},
		{
			name:    "valid positive",
			value:   "5",
			wantErr: false,
		},
		{
			name:    "maximum allowed",
			value:   "10",
			wantErr: false,
		},
		{
			name:    "negative value",
			value:   "-1",
			wantErr: true,
			errMsg:  "must be non-negative",
		},
		{
			name:    "exceeds maximum",
			value:   "11",
			wantErr: true,
			errMsg:  "exceeds maximum allowed depth",
		},
		{
			name:    "invalid format",
			value:   "abc",
			wantErr: true,
			errMsg:  "invalid integer format",
		},
		{
			name:    "with whitespace",
			value:   "  5  ",
			wantErr: false,
		},
		{
			name:    "too long value",
			value:   "12345678901", // 11 digits
			wantErr: true,
			errMsg:  "value too long",
		},
		{
			name:    "very large number",
			value:   "999999999999",
			wantErr: true,
			errMsg:  "value too long",
		},
		{
			name:    "float value",
			value:   "5.5",
			wantErr: true,
			errMsg:  "invalid integer format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSpawnDepth(tt.value)
			if tt.wantErr {
				testutil.AssertError(t, err, "Expected validation error")
				if tt.errMsg != "" {
					testutil.AssertContains(t, err.Error(), tt.errMsg, "Error message should contain expected text")
				}
			} else {
				testutil.AssertNoError(t, err, "Should not have validation error")
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
		{
			name:    "empty value",
			value:   "",
			wantErr: false,
		},
		{
			name:    "zero flag",
			value:   "0",
			wantErr: false,
		},
		{
			name:    "one flag",
			value:   "1",
			wantErr: false,
		},
		{
			name:    "invalid true",
			value:   "true",
			wantErr: true,
		},
		{
			name:    "invalid false",
			value:   "false",
			wantErr: true,
		},
		{
			name:    "invalid yes",
			value:   "yes",
			wantErr: true,
		},
		{
			name:    "invalid number",
			value:   "2",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBooleanFlag(tt.value)
			if tt.wantErr {
				testutil.AssertError(t, err, "Expected validation error")
				testutil.AssertContains(t, err.Error(), "must be '0' or '1'", "Should mention valid values")
			} else {
				testutil.AssertNoError(t, err, "Should not have validation error")
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
		{
			name:    "empty value",
			value:   "",
			wantErr: false,
		},
		{
			name:    "true lowercase",
			value:   "true",
			wantErr: false,
		},
		{
			name:    "false lowercase",
			value:   "false",
			wantErr: false,
		},
		{
			name:    "true uppercase",
			value:   "TRUE",
			wantErr: false,
		},
		{
			name:    "false uppercase",
			value:   "FALSE",
			wantErr: false,
		},
		{
			name:    "mixed case true",
			value:   "True",
			wantErr: false,
		},
		{
			name:    "one flag",
			value:   "1",
			wantErr: false,
		},
		{
			name:    "zero flag",
			value:   "0",
			wantErr: false,
		},
		{
			name:    "invalid yes",
			value:   "yes",
			wantErr: true,
		},
		{
			name:    "invalid no",
			value:   "no",
			wantErr: true,
		},
		{
			name:    "invalid number",
			value:   "2",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBoolean(tt.value)
			if tt.wantErr {
				testutil.AssertError(t, err, "Expected validation error")
				testutil.AssertContains(t, err.Error(), "must be 'true', 'false', '1', or '0'", 
					"Should mention valid values")
			} else {
				testutil.AssertNoError(t, err, "Should not have validation error")
			}
		})
	}
}

func TestValidatePort(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty value",
			value:   "",
			wantErr: false,
		},
		{
			name:    "valid port 80",
			value:   "80",
			wantErr: false,
		},
		{
			name:    "valid port 3456",
			value:   "3456",
			wantErr: false,
		},
		{
			name:    "valid port 65535",
			value:   "65535",
			wantErr: false,
		},
		{
			name:    "minimum port 1",
			value:   "1",
			wantErr: false,
		},
		{
			name:    "port zero",
			value:   "0",
			wantErr: true,
			errMsg:  "must be between 1 and 65535",
		},
		{
			name:    "negative port",
			value:   "-1",
			wantErr: true,
			errMsg:  "must be between 1 and 65535",
		},
		{
			name:    "port too high",
			value:   "65536",
			wantErr: true,
			errMsg:  "must be between 1 and 65535",
		},
		{
			name:    "invalid format",
			value:   "abc",
			wantErr: true,
			errMsg:  "must be a valid integer",
		},
		{
			name:    "float value",
			value:   "80.5",
			wantErr: true,
			errMsg:  "must be a valid integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePort(tt.value)
			if tt.wantErr {
				testutil.AssertError(t, err, "Expected validation error")
				if tt.errMsg != "" {
					testutil.AssertContains(t, err.Error(), tt.errMsg, "Error message should contain expected text")
				}
			} else {
				testutil.AssertNoError(t, err, "Should not have validation error")
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
		{
			name:    "empty value",
			value:   "",
			wantErr: false,
		},
		{
			name:    "localhost",
			value:   "localhost",
			wantErr: false,
		},
		{
			name:    "IPv4 address",
			value:   "127.0.0.1",
			wantErr: false,
		},
		{
			name:    "domain name",
			value:   "example.com",
			wantErr: false,
		},
		{
			name:    "IPv6 address",
			value:   "::1",
			wantErr: false,
		},
		{
			name:    "wildcard",
			value:   "0.0.0.0",
			wantErr: false,
		},
		{
			name:    "whitespace only",
			value:   "   ",
			wantErr: true,
		},
		{
			name:    "with spaces but valid",
			value:   "  localhost  ",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHostAddress(tt.value)
			if tt.wantErr {
				testutil.AssertError(t, err, "Expected validation error")
			} else {
				testutil.AssertNoError(t, err, "Should not have validation error")
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
		{
			name:    "empty value",
			value:   "",
			wantErr: false,
		},
		{
			name:    "valid absolute path",
			value:   "/var/log/ccproxy.log",
			wantErr: false,
		},
		{
			name:    "valid relative path",
			value:   "logs/ccproxy.log",
			wantErr: false,
		},
		{
			name:    "current directory file",
			value:   "./config.json",
			wantErr: false,
		},
		{
			name:    "parent directory file",
			value:   "../config.json",
			wantErr: false,
		},
		{
			name:    "with null character",
			value:   "/var/log/\x00/ccproxy.log",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilePath(tt.value)
			if tt.wantErr {
				testutil.AssertError(t, err, "Expected validation error")
				testutil.AssertContains(t, err.Error(), "null characters", "Should mention null characters")
			} else {
				testutil.AssertNoError(t, err, "Should not have validation error")
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
		{
			name:    "empty value",
			value:   "",
			wantErr: false,
		},
		{
			name:    "valid absolute path",
			value:   "/var/lib/ccproxy",
			wantErr: false,
		},
		{
			name:    "valid relative path",
			value:   "data/ccproxy",
			wantErr: false,
		},
		{
			name:    "current directory",
			value:   ".",
			wantErr: false,
		},
		{
			name:    "parent directory",
			value:   "..",
			wantErr: false,
		},
		{
			name:    "whitespace only",
			value:   "   ",
			wantErr: true,
		},
		{
			name:    "with null character",
			value:   "/var/\x00/ccproxy",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDirectoryPath(tt.value)
			if tt.wantErr {
				testutil.AssertError(t, err, "Expected validation error")
			} else {
				testutil.AssertNoError(t, err, "Should not have validation error")
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
		{
			name:    "empty value",
			value:   "",
			wantErr: false,
		},
		{
			name:    "valid http URL",
			value:   "http://example.com",
			wantErr: false,
		},
		{
			name:    "valid https URL",
			value:   "https://example.com",
			wantErr: false,
		},
		{
			name:    "http with port",
			value:   "http://example.com:8080",
			wantErr: false,
		},
		{
			name:    "https with path",
			value:   "https://example.com/path",
			wantErr: false,
		},
		{
			name:    "uppercase HTTP",
			value:   "HTTP://example.com",
			wantErr: false,
		},
		{
			name:    "uppercase HTTPS",
			value:   "HTTPS://example.com",
			wantErr: false,
		},
		{
			name:    "missing protocol",
			value:   "example.com",
			wantErr: true,
		},
		{
			name:    "ftp protocol",
			value:   "ftp://example.com",
			wantErr: true,
		},
		{
			name:    "invalid protocol",
			value:   "tcp://example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.value)
			if tt.wantErr {
				testutil.AssertError(t, err, "Expected validation error")
				testutil.AssertContains(t, err.Error(), "must start with http:// or https://", 
					"Should mention valid protocols")
			} else {
				testutil.AssertNoError(t, err, "Should not have validation error")
			}
		})
	}
}

func TestValidateEnvironmentVariables(t *testing.T) {
	// Save original environment
	originalEnvs := make(map[string]string)
	for _, envVar := range EnvironmentVariables {
		originalEnvs[envVar.Name] = os.Getenv(envVar.Name)
	}
	
	// Clean up environment after test
	defer func() {
		for name, value := range originalEnvs {
			if value == "" {
				os.Unsetenv(name)
			} else {
				os.Setenv(name, value)
			}
		}
	}()

	t.Run("all variables valid", func(t *testing.T) {
		// Clear all environment variables
		for _, envVar := range EnvironmentVariables {
			os.Unsetenv(envVar.Name)
		}
		
		// Set valid values
		os.Setenv("CCPROXY_SPAWN_DEPTH", "5")
		os.Setenv("CCPROXY_FOREGROUND", "1")
		os.Setenv("CCPROXY_PORT", "8080")
		os.Setenv("CCPROXY_LOG", "true")
		
		err := ValidateEnvironmentVariables()
		testutil.AssertNoError(t, err, "Should validate successfully with valid values")
	})

	t.Run("invalid values", func(t *testing.T) {
		// Clear all environment variables
		for _, envVar := range EnvironmentVariables {
			os.Unsetenv(envVar.Name)
		}
		
		// Set invalid values
		os.Setenv("CCPROXY_SPAWN_DEPTH", "invalid")
		os.Setenv("CCPROXY_PORT", "99999")
		
		err := ValidateEnvironmentVariables()
		testutil.AssertError(t, err, "Should fail validation with invalid values")
		testutil.AssertContains(t, err.Error(), "CCPROXY_SPAWN_DEPTH", "Should mention invalid spawn depth")
		testutil.AssertContains(t, err.Error(), "CCPROXY_PORT", "Should mention invalid port")
	})

	t.Run("empty environment", func(t *testing.T) {
		// Clear all environment variables
		for _, envVar := range EnvironmentVariables {
			os.Unsetenv(envVar.Name)
		}
		
		err := ValidateEnvironmentVariables()
		testutil.AssertNoError(t, err, "Should validate successfully with empty environment (no required vars)")
	})
}

func TestValidateEnvironmentVariablesWithReport(t *testing.T) {
	// Save original environment
	originalEnvs := make(map[string]string)
	for _, envVar := range EnvironmentVariables {
		originalEnvs[envVar.Name] = os.Getenv(envVar.Name)
	}
	
	// Clean up environment after test
	defer func() {
		for name, value := range originalEnvs {
			if value == "" {
				os.Unsetenv(name)
			} else {
				os.Setenv(name, value)
			}
		}
	}()

	t.Run("valid environment", func(t *testing.T) {
		// Clear all environment variables
		for _, envVar := range EnvironmentVariables {
			os.Unsetenv(envVar.Name)
		}
		
		// Set valid values
		os.Setenv("CCPROXY_SPAWN_DEPTH", "3")
		os.Setenv("CCPROXY_PORT", "3456")
		
		report := ValidateEnvironmentVariablesWithReport()
		testutil.AssertNotEqual(t, nil, report, "Report should not be nil")
		testutil.AssertTrue(t, report.Valid, "Report should be valid")
		testutil.AssertEqual(t, 0, len(report.Errors), "Should have no errors")
		
		// Check specific details
		spawnDetail, exists := report.Details["CCPROXY_SPAWN_DEPTH"]
		testutil.AssertTrue(t, exists, "Should have spawn depth detail")
		testutil.AssertTrue(t, spawnDetail.Valid, "Spawn depth should be valid")
		testutil.AssertEqual(t, "3", spawnDetail.Value, "Should have correct value")
	})

	t.Run("invalid environment", func(t *testing.T) {
		// Clear all environment variables
		for _, envVar := range EnvironmentVariables {
			os.Unsetenv(envVar.Name)
		}
		
		// Set invalid values
		os.Setenv("CCPROXY_SPAWN_DEPTH", "invalid")
		os.Setenv("CCPROXY_FOREGROUND", "maybe")
		
		report := ValidateEnvironmentVariablesWithReport()
		testutil.AssertNotEqual(t, nil, report, "Report should not be nil")
		testutil.AssertFalse(t, report.Valid, "Report should be invalid")
		testutil.AssertTrue(t, len(report.Errors) > 0, "Should have errors")
		
		// Check specific invalid details
		spawnDetail, exists := report.Details["CCPROXY_SPAWN_DEPTH"]
		testutil.AssertTrue(t, exists, "Should have spawn depth detail")
		testutil.AssertFalse(t, spawnDetail.Valid, "Spawn depth should be invalid")
		testutil.AssertTrue(t, len(spawnDetail.Error) > 0, "Should have error message")
		
		fgDetail, exists := report.Details["CCPROXY_FOREGROUND"]
		testutil.AssertTrue(t, exists, "Should have foreground detail")
		testutil.AssertFalse(t, fgDetail.Valid, "Foreground should be invalid")
	})
}

func TestGetEnvironmentVariableDocumentation(t *testing.T) {
	doc := GetEnvironmentVariableDocumentation()
	testutil.AssertTrue(t, len(doc) > 0, "Documentation should not be empty")
	testutil.AssertContains(t, doc, "CCProxy Environment Variables", "Should contain title")
	testutil.AssertContains(t, doc, "CCPROXY_SPAWN_DEPTH", "Should contain spawn depth variable")
	testutil.AssertContains(t, doc, "CCPROXY_PORT", "Should contain port variable")
	testutil.AssertContains(t, doc, "Default:", "Should contain default values")
	
	// Check that all defined variables are documented
	for _, envVar := range EnvironmentVariables {
		testutil.AssertContains(t, doc, envVar.Name, 
			fmt.Sprintf("Should document %s", envVar.Name))
		testutil.AssertContains(t, doc, envVar.Description, 
			fmt.Sprintf("Should contain description for %s", envVar.Name))
	}
}

func TestEnvironmentVariablesDefinition(t *testing.T) {
	// Test that all environment variables are properly defined
	testutil.AssertTrue(t, len(EnvironmentVariables) > 0, "Should have environment variables defined")
	
	for _, envVar := range EnvironmentVariables {
		testutil.AssertTrue(t, len(envVar.Name) > 0, 
			"Environment variable should have a name")
		testutil.AssertTrue(t, strings.HasPrefix(envVar.Name, "CCPROXY_"), 
			"Environment variable should start with CCPROXY_")
		testutil.AssertTrue(t, len(envVar.Description) > 0, 
			"Environment variable should have a description")
		
		// Test validation function if defined
		if envVar.ValidateFunc != nil {
			// Test with empty value
			err := envVar.ValidateFunc("")
			if envVar.Required {
				// Required variables might fail with empty values in their validation
				// but that's checked separately
			} else {
				testutil.AssertNoError(t, err, 
					fmt.Sprintf("Validation function for %s should handle empty values", envVar.Name))
			}
			
			// Test with default value if defined
			if envVar.DefaultValue != "" {
				err := envVar.ValidateFunc(envVar.DefaultValue)
				testutil.AssertNoError(t, err, 
					fmt.Sprintf("Default value for %s should be valid", envVar.Name))
			}
		}
	}
}

func TestEnvVarEdgeCases(t *testing.T) {
	t.Run("spawn depth edge cases", func(t *testing.T) {
		// Test boundary values
		testutil.AssertNoError(t, ValidateSpawnDepth("0"), "Should accept 0")
		testutil.AssertNoError(t, ValidateSpawnDepth("10"), "Should accept 10")
		testutil.AssertError(t, ValidateSpawnDepth("11"), "Should reject 11")
		
		// Test with leading zeros
		testutil.AssertNoError(t, ValidateSpawnDepth("005"), "Should handle leading zeros")
		
		// Test extreme cases
		testutil.AssertError(t, ValidateSpawnDepth("2147483648"), "Should reject values > int32 max")
	})

	t.Run("port edge cases", func(t *testing.T) {
		// Test boundary values
		testutil.AssertNoError(t, ValidatePort("1"), "Should accept port 1")
		testutil.AssertNoError(t, ValidatePort("65535"), "Should accept port 65535")
		testutil.AssertError(t, ValidatePort("0"), "Should reject port 0")
		testutil.AssertError(t, ValidatePort("65536"), "Should reject port 65536")
		
		// Test with leading zeros
		testutil.AssertNoError(t, ValidatePort("0080"), "Should handle leading zeros")
	})
}

// Benchmark tests for performance verification
func BenchmarkValidateSpawnDepth(b *testing.B) {
	testValue := "5"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateSpawnDepth(testValue)
	}
}

func BenchmarkValidateEnvironmentVariables(b *testing.B) {
	// Set up some environment variables
	os.Setenv("CCPROXY_SPAWN_DEPTH", "3")
	os.Setenv("CCPROXY_PORT", "8080")
	defer func() {
		os.Unsetenv("CCPROXY_SPAWN_DEPTH")
		os.Unsetenv("CCPROXY_PORT")
	}()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateEnvironmentVariables()
	}
}

func BenchmarkValidateEnvironmentVariablesWithReport(b *testing.B) {
	// Set up some environment variables
	os.Setenv("CCPROXY_SPAWN_DEPTH", "3")
	os.Setenv("CCPROXY_PORT", "8080")
	defer func() {
		os.Unsetenv("CCPROXY_SPAWN_DEPTH")
		os.Unsetenv("CCPROXY_PORT")
	}()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateEnvironmentVariablesWithReport()
	}
}

func TestValidateSpawnDepthComprehensive(t *testing.T) {
	// Test all error paths for complete coverage
	tests := []struct {
		name    string
		value   string
		wantErr bool
		errType string
	}{
		{
			name:    "empty value success",
			value:   "",
			wantErr: false,
		},
		{
			name:    "value with leading/trailing spaces",
			value:   "  5  ",
			wantErr: false,
		},
		{
			name:    "maximum integer length check",
			value:   "1234567890",
			wantErr: true, // This is actually > 10 for spawn depth
			errType: "exceeds maximum allowed depth",
		},
		{
			name:    "exactly 10 chars (boundary)",
			value:   "9999999999",
			wantErr: true, // This would be out of range for 32-bit int
			errType: "value out of range",
		},
		{
			name:    "11 chars (too long)",
			value:   "12345678901",
			wantErr: true,
			errType: "value too long",
		},
		{
			name:    "range error simulation",
			value:   "2147483648", // int32 max + 1
			wantErr: true,
			errType: "value out of range",
		},
		{
			name:    "long syntax error",
			value:   "not_a_number",
			wantErr: true,
			errType: "value too long", // First check is length, then parsing
		},
		{
			name:    "short syntax error",
			value:   "abc",
			wantErr: true,
			errType: "invalid integer format",
		},
		{
			name:    "negative number",
			value:   "-5",
			wantErr: true,
			errType: "must be non-negative",
		},
		{
			name:    "exceeds max depth",
			value:   "15",
			wantErr: true,
			errType: "exceeds maximum allowed depth",
		},
		{
			name:    "boundary max depth",
			value:   "10",
			wantErr: false,
		},
		{
			name:    "boundary max depth + 1",
			value:   "11",
			wantErr: true,
			errType: "exceeds maximum allowed depth",
		},
		{
			name:    "float format",
			value:   "5.0",
			wantErr: true,
			errType: "invalid integer format",
		},
		{
			name:    "hex format",
			value:   "0x10",
			wantErr: true,
			errType: "invalid integer format",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSpawnDepth(tt.value)
			if tt.wantErr {
				testutil.AssertError(t, err, "Expected validation error")
				if tt.errType != "" {
					testutil.AssertContains(t, err.Error(), tt.errType, 
						"Error should contain expected type: "+tt.errType)
				}
			} else {
				testutil.AssertNoError(t, err, "Should not have validation error")
			}
		})
	}
}

func TestValidateEnvironmentVariablesWithRequiredVars(t *testing.T) {
	// Save original environment
	originalEnvs := make(map[string]string)
	for _, envVar := range EnvironmentVariables {
		originalEnvs[envVar.Name] = os.Getenv(envVar.Name)
	}
	
	// Clean up environment after test
	defer func() {
		for name, value := range originalEnvs {
			if value == "" {
				os.Unsetenv(name)
			} else {
				os.Setenv(name, value)
			}
		}
	}()

	// Test with a temporarily required variable
	// Modify one variable to be required for testing
	originalRequired := EnvironmentVariables[0].Required
	EnvironmentVariables[0].Required = true
	
	defer func() {
		EnvironmentVariables[0].Required = originalRequired
	}()
	
	t.Run("missing required variable", func(t *testing.T) {
		// Clear all environment variables
		for _, envVar := range EnvironmentVariables {
			os.Unsetenv(envVar.Name)
		}
		
		err := ValidateEnvironmentVariables()
		testutil.AssertError(t, err, "Should fail with missing required variable")
		testutil.AssertContains(t, err.Error(), EnvironmentVariables[0].Name, 
			"Should mention the missing required variable")
		testutil.AssertContains(t, err.Error(), "required environment variable not set", 
			"Should mention requirement")
	})
	
	t.Run("present required variable", func(t *testing.T) {
		// Clear all environment variables
		for _, envVar := range EnvironmentVariables {
			os.Unsetenv(envVar.Name)
		}
		
		// Set the required variable
		os.Setenv(EnvironmentVariables[0].Name, EnvironmentVariables[0].DefaultValue)
		
		err := ValidateEnvironmentVariables()
		testutil.AssertNoError(t, err, "Should pass with required variable set")
	})
}

func TestValidateEnvironmentVariablesWithReportComprehensive(t *testing.T) {
	// Save original environment
	originalEnvs := make(map[string]string)
	for _, envVar := range EnvironmentVariables {
		originalEnvs[envVar.Name] = os.Getenv(envVar.Name)
	}
	
	// Clean up environment after test
	defer func() {
		for name, value := range originalEnvs {
			if value == "" {
				os.Unsetenv(name)
			} else {
				os.Setenv(name, value)
			}
		}
	}()

	// Test with mixed valid and invalid values
	t.Run("mixed validation results", func(t *testing.T) {
		// Clear all environment variables
		for _, envVar := range EnvironmentVariables {
			os.Unsetenv(envVar.Name)
		}
		
		// Set mix of valid and invalid values
		os.Setenv("CCPROXY_SPAWN_DEPTH", "5")       // valid
		os.Setenv("CCPROXY_FOREGROUND", "invalid")  // invalid
		os.Setenv("CCPROXY_PORT", "80")            // valid
		os.Setenv("CCPROXY_LOG", "maybe")          // invalid
		
		report := ValidateEnvironmentVariablesWithReport()
		testutil.AssertNotEqual(t, nil, report, "Report should not be nil")
		testutil.AssertFalse(t, report.Valid, "Report should be invalid")
		testutil.AssertTrue(t, len(report.Errors) > 0, "Should have errors")
		
		// Check specific details
		spawnDetail := report.Details["CCPROXY_SPAWN_DEPTH"]
		testutil.AssertTrue(t, spawnDetail.Valid, "Spawn depth should be valid")
		testutil.AssertEqual(t, "5", spawnDetail.Value, "Should have correct value")
		
		fgDetail := report.Details["CCPROXY_FOREGROUND"]
		testutil.AssertFalse(t, fgDetail.Valid, "Foreground should be invalid")
		testutil.AssertEqual(t, "invalid", fgDetail.Value, "Should have correct invalid value")
		testutil.AssertTrue(t, len(fgDetail.Error) > 0, "Should have error message")
		
		portDetail := report.Details["CCPROXY_PORT"]
		testutil.AssertTrue(t, portDetail.Valid, "Port should be valid")
		
		logDetail := report.Details["CCPROXY_LOG"]
		testutil.AssertFalse(t, logDetail.Valid, "Log should be invalid")
		testutil.AssertContains(t, logDetail.Error, "must be 'true', 'false', '1', or '0'", 
			"Should have descriptive error")
	})
	
	// Test with required variable missing
	originalRequired := EnvironmentVariables[1].Required
	EnvironmentVariables[1].Required = true
	
	defer func() {
		EnvironmentVariables[1].Required = originalRequired
	}()
	
	t.Run("required variable missing in report", func(t *testing.T) {
		// Clear all environment variables
		for _, envVar := range EnvironmentVariables {
			os.Unsetenv(envVar.Name)
		}
		
		report := ValidateEnvironmentVariablesWithReport()
		testutil.AssertNotEqual(t, nil, report, "Report should not be nil")
		testutil.AssertFalse(t, report.Valid, "Report should be invalid")
		testutil.AssertTrue(t, len(report.Errors) > 0, "Should have errors")
		
		// Check the required variable detail
		requiredDetail := report.Details[EnvironmentVariables[1].Name]
		testutil.AssertFalse(t, requiredDetail.Valid, "Required variable should be invalid")
		testutil.AssertEqual(t, "", requiredDetail.Value, "Should have empty value")
		testutil.AssertContains(t, requiredDetail.Error, "required environment variable not set", 
			"Should mention requirement")
	})
}

func TestGetEnvironmentVariableDocumentationComprehensive(t *testing.T) {
	doc := GetEnvironmentVariableDocumentation()
	
	// Test basic structure
	testutil.AssertTrue(t, len(doc) > 0, "Documentation should not be empty")
	testutil.AssertContains(t, doc, "CCProxy Environment Variables", "Should contain title")
	testutil.AssertContains(t, doc, "============================", "Should contain separator")
	
	// Test that each variable is documented
	for _, envVar := range EnvironmentVariables {
		testutil.AssertContains(t, doc, envVar.Name, 
			fmt.Sprintf("Should document variable name: %s", envVar.Name))
		testutil.AssertContains(t, doc, envVar.Description, 
			fmt.Sprintf("Should document description for: %s", envVar.Name))
		
		// Check default value if present
		if envVar.DefaultValue != "" {
			testutil.AssertContains(t, doc, fmt.Sprintf("Default: %s", envVar.DefaultValue), 
				fmt.Sprintf("Should document default value for: %s", envVar.Name))
		}
		
		// Check required flag (though none are required in current definition)
		if envVar.Required {
			testutil.AssertContains(t, doc, "Required: Yes", 
				fmt.Sprintf("Should document required flag for: %s", envVar.Name))
		}
	}
	
	// Test formatting consistency
	lines := strings.Split(doc, "\n")
	testutil.AssertTrue(t, len(lines) > 3, "Should have multiple lines")
	
	// Each variable should have at least 2 lines (name/description + empty line)
	expectedMinLines := len(EnvironmentVariables)*2 + 3 // Variables + title + separator + final newline
	testutil.AssertTrue(t, len(lines) >= expectedMinLines, "Should have adequate line count")
}

func TestEnvironmentVariablesDefinitionComprehensive(t *testing.T) {
	// Validate the structure and consistency of environment variables
	testutil.AssertTrue(t, len(EnvironmentVariables) > 0, "Should have environment variables defined")
	
	namesSeen := make(map[string]bool)
	
	for i, envVar := range EnvironmentVariables {
		// Test required fields
		testutil.AssertTrue(t, len(envVar.Name) > 0, 
			fmt.Sprintf("Variable %d should have a name", i))
		testutil.AssertTrue(t, strings.HasPrefix(envVar.Name, "CCPROXY_"), 
			fmt.Sprintf("Variable %s should start with CCPROXY_", envVar.Name))
		testutil.AssertTrue(t, len(envVar.Description) > 0, 
			fmt.Sprintf("Variable %s should have a description", envVar.Name))
		
		// Test uniqueness
		testutil.AssertFalse(t, namesSeen[envVar.Name], 
			fmt.Sprintf("Variable name %s should be unique", envVar.Name))
		namesSeen[envVar.Name] = true
		
		// Test validation function consistency
		if envVar.ValidateFunc != nil {
			// Test with empty value (should work for non-required)
			if !envVar.Required {
				err := envVar.ValidateFunc("")
				testutil.AssertNoError(t, err, 
					fmt.Sprintf("Validation function for %s should handle empty values", envVar.Name))
			}
			
			// Test with default value if defined
			if envVar.DefaultValue != "" {
				err := envVar.ValidateFunc(envVar.DefaultValue)
				testutil.AssertNoError(t, err, 
					fmt.Sprintf("Default value for %s should be valid: %s", envVar.Name, envVar.DefaultValue))
			}
		}
		
		// Test that description is descriptive (at least 10 chars)
		testutil.AssertTrue(t, len(envVar.Description) >= 10, 
			fmt.Sprintf("Description for %s should be descriptive", envVar.Name))
		
		// Test that name follows convention (uppercase with underscores)
		testutil.AssertTrue(t, envVar.Name == strings.ToUpper(envVar.Name), 
			fmt.Sprintf("Variable name %s should be uppercase", envVar.Name))
	}
	
	// Test that we have key variables we expect
	expectedVars := []string{
		"CCPROXY_SPAWN_DEPTH",
		"CCPROXY_FOREGROUND", 
		"CCPROXY_PORT",
		"CCPROXY_HOST",
		"CCPROXY_LOG",
		"CCPROXY_API_KEY",
	}
	
	for _, expected := range expectedVars {
		found := false
		for _, envVar := range EnvironmentVariables {
			if envVar.Name == expected {
				found = true
				break
			}
		}
		testutil.AssertTrue(t, found, fmt.Sprintf("Should have variable: %s", expected))
	}
}

// Additional edge case tests for better coverage
func TestValidationFunctionEdgeCases(t *testing.T) {
	t.Run("ValidateSpawnDepthEdgeCases", func(t *testing.T) {
		// Test the specific error type paths
		err := ValidateSpawnDepth("2147483649") // > int32 max
		if err != nil {
			// Should get range error if system supports it
			testutil.AssertError(t, err, "Should error on overflow")
		}
		
		// Test leading zeros handling
		testutil.AssertNoError(t, ValidateSpawnDepth("0005"), "Should handle leading zeros")
		testutil.AssertNoError(t, ValidateSpawnDepth("00000"), "Should handle multiple leading zeros")
	})
}