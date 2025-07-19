package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDataSanitizer(t *testing.T) {
	t.Run("with default config", func(t *testing.T) {
		sanitizer, err := NewDataSanitizer(nil)
		assert.NoError(t, err)
		assert.NotNil(t, sanitizer)
		assert.NotNil(t, sanitizer.config)
		assert.NotEmpty(t, sanitizer.redactPatterns)
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &SecurityConfig{
			SensitivePatterns: []string{`custom_secret:\s*\w+`},
		}
		
		sanitizer, err := NewDataSanitizer(config)
		assert.NoError(t, err)
		assert.NotNil(t, sanitizer)
		assert.Len(t, sanitizer.sensitiveRegexps, 1)
	})

	t.Run("with invalid pattern", func(t *testing.T) {
		config := &SecurityConfig{
			SensitivePatterns: []string{`[`}, // Invalid regex
		}
		
		sanitizer, err := NewDataSanitizer(config)
		assert.Error(t, err)
		assert.Nil(t, sanitizer)
	})
}

func TestSanitizeString(t *testing.T) {
	config := &SecurityConfig{
		LogSensitiveData: false,
	}
	
	sanitizer, err := NewDataSanitizer(config)
	require.NoError(t, err)

	t.Run("HTML escape", func(t *testing.T) {
		input := `<script>alert("xss")</script>`
		expected := `&lt;script&gt;alert(&#34;xss&#34;)&lt;/script&gt;`
		
		result := sanitizer.SanitizeString(input)
		assert.Equal(t, expected, result)
	})

	t.Run("null bytes removal", func(t *testing.T) {
		input := "test\x00data"
		expected := "testdata"
		
		result := sanitizer.SanitizeString(input)
		assert.Equal(t, expected, result)
	})

	t.Run("control characters removal", func(t *testing.T) {
		input := "test\x01\x02\x03data"
		expected := "testdata"
		
		result := sanitizer.SanitizeString(input)
		assert.Equal(t, expected, result)
	})

	t.Run("API key redaction", func(t *testing.T) {
		input := `api_key: "sk-1234567890abcdef"`
		
		result := sanitizer.SanitizeString(input)
		assert.Contains(t, result, "[REDACTED_API_KEY]")
		assert.NotContains(t, result, "sk-1234567890abcdef")
	})

	t.Run("password redaction", func(t *testing.T) {
		input := `password: mysecretpassword123`
		
		result := sanitizer.SanitizeString(input)
		assert.Contains(t, result, "[REDACTED_PASSWORD]")
		assert.NotContains(t, result, "mysecretpassword123")
	})

	t.Run("credit card redaction", func(t *testing.T) {
		input := `card number: 4111 1111 1111 1111`
		
		result := sanitizer.SanitizeString(input)
		assert.Contains(t, result, "[REDACTED_CREDIT_CARD]")
		assert.NotContains(t, result, "4111 1111 1111 1111")
	})

	t.Run("email redaction", func(t *testing.T) {
		input := `email: user@example.com`
		
		result := sanitizer.SanitizeString(input)
		assert.Contains(t, result, "[REDACTED_EMAIL]")
		assert.NotContains(t, result, "user@example.com")
	})

	t.Run("with logging enabled", func(t *testing.T) {
		config := &SecurityConfig{
			LogSensitiveData: true,
		}
		
		sanitizer, err := NewDataSanitizer(config)
		require.NoError(t, err)
		
		input := `password: mysecret`
		result := sanitizer.SanitizeString(input)
		assert.Contains(t, result, "mysecret") // Not redacted when logging is enabled
	})
}

func TestSanitizeRequest(t *testing.T) {
	sanitizer, err := NewDataSanitizer(nil)
	require.NoError(t, err)

	t.Run("map request", func(t *testing.T) {
		request := map[string]interface{}{
			"username": "john",
			"password": "secret123",
			"email":    "john@example.com",
		}
		
		result, err := sanitizer.SanitizeRequest(request)
		assert.NoError(t, err)
		
		sanitized, ok := result.(map[string]interface{})
		require.True(t, ok)
		
		assert.Equal(t, "john", sanitized["username"])
		assert.Equal(t, "[REDACTED]", sanitized["password"])
		assert.Contains(t, sanitized["email"].(string), "[REDACTED_EMAIL]")
	})

	t.Run("nested map request", func(t *testing.T) {
		request := map[string]interface{}{
			"user": map[string]interface{}{
				"name":     "john",
				"api_key":  "sk-12345",
			},
		}
		
		result, err := sanitizer.SanitizeRequest(request)
		assert.NoError(t, err)
		
		sanitized, ok := result.(map[string]interface{})
		require.True(t, ok)
		
		userMap := sanitized["user"].(map[string]interface{})
		assert.Equal(t, "john", userMap["name"])
		assert.Equal(t, "[REDACTED]", userMap["api_key"])
	})

	t.Run("array request", func(t *testing.T) {
		request := []interface{}{
			map[string]interface{}{
				"token": "secret-token",
			},
			"normal string",
		}
		
		result, err := sanitizer.SanitizeRequest(request)
		assert.NoError(t, err)
		
		sanitized, ok := result.([]interface{})
		require.True(t, ok)
		require.Len(t, sanitized, 2)
		
		firstItem := sanitized[0].(map[string]interface{})
		assert.Equal(t, "[REDACTED]", firstItem["token"])
		assert.Equal(t, "normal string", sanitized[1])
	})
}

func TestRemoveSensitiveData(t *testing.T) {
	sanitizer, err := NewDataSanitizer(nil)
	require.NoError(t, err)

	data := map[string]interface{}{
		"username":     "john",
		"password":     "secret",
		"access_token": "token123",
		"public_data":  "visible",
	}
	
	result := sanitizer.RemoveSensitiveData(data)
	
	assert.Equal(t, "john", result["username"])
	assert.Equal(t, "[REDACTED]", result["password"])
	assert.Equal(t, "[REDACTED]", result["access_token"])
	assert.Equal(t, "visible", result["public_data"])
}

func TestIsSensitiveKey(t *testing.T) {
	sanitizer, err := NewDataSanitizer(nil)
	require.NoError(t, err)

	tests := []struct {
		key       string
		sensitive bool
	}{
		{"password", true},
		{"user_password", true},
		{"api_key", true},
		{"apikey", true},
		{"api-key", true},
		{"token", true},
		{"auth_token", true},
		{"secret", true},
		{"private_key", true},
		{"username", false},
		{"email", false},
		{"name", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := sanitizer.isSensitiveKey(tt.key)
			assert.Equal(t, tt.sensitive, result)
		})
	}
}

func TestMaskValue(t *testing.T) {
	tests := []struct {
		value     string
		showChars int
		expected  string
	}{
		{"secret123", 3, "sec******"},
		{"short", 10, "*****"},
		{"verylongsecret", 4, "very**********"},
		{"test", 0, "****"},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			result := MaskValue(tt.value, tt.showChars)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		email    string
		expected string
	}{
		{"user@example.com", "us***@example.com"},
		{"a@example.com", "*@example.com"},
		{"longusername@example.com", "lo***********@example.com"},
		{"notanemail", "not*******"},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := MaskEmail(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		apiKey   string
		expected string
	}{
		{"sk-1234567890abcdef", "sk-1********cdef"},
		{"short", "*****"},
		{"12345678", "********"},
		{"verylongapikey123456", "very************3456"},
	}

	for _, tt := range tests {
		t.Run(tt.apiKey, func(t *testing.T) {
			result := MaskAPIKey(tt.apiKey)
			assert.Equal(t, tt.expected, result)
		})
	}
}