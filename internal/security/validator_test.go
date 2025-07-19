package security

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRequestValidator(t *testing.T) {
	t.Run("with default config", func(t *testing.T) {
		validator, err := NewRequestValidator(nil)
		assert.NoError(t, err)
		assert.NotNil(t, validator)
		assert.NotNil(t, validator.config)
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &SecurityConfig{
			BlockedPatterns:   []string{`test\d+`},
			SensitivePatterns: []string{`password:\s*\w+`},
		}
		
		validator, err := NewRequestValidator(config)
		assert.NoError(t, err)
		assert.NotNil(t, validator)
		assert.Len(t, validator.compiledPatterns, 1)
		assert.Len(t, validator.sensitiveRegexps, 1)
	})

	t.Run("with invalid pattern", func(t *testing.T) {
		config := &SecurityConfig{
			BlockedPatterns: []string{`[`}, // Invalid regex
		}
		
		validator, err := NewRequestValidator(config)
		assert.Error(t, err)
		assert.Nil(t, validator)
	})
}

func TestValidate(t *testing.T) {
	config := &SecurityConfig{
		EnableContentFilter: true,
		BlockedPatterns:    []string{`malicious`},
		SensitivePatterns:  []string{`password:\s*\w+`},
	}
	
	validator, err := NewRequestValidator(config)
	require.NoError(t, err)

	t.Run("clean data", func(t *testing.T) {
		data := map[string]string{
			"message": "This is clean data",
		}
		
		result := validator.Validate(data)
		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
		assert.Equal(t, 1.0, result.Score)
	})

	t.Run("blocked pattern", func(t *testing.T) {
		data := map[string]string{
			"message": "This contains malicious content",
		}
		
		result := validator.Validate(data)
		assert.False(t, result.Valid)
		assert.NotEmpty(t, result.Errors)
		assert.Less(t, result.Score, 1.0)
	})

	t.Run("sensitive data", func(t *testing.T) {
		data := map[string]string{
			"config": "password: secret123",
		}
		
		result := validator.Validate(data)
		assert.True(t, result.Valid) // Sensitive data produces warning, not error
		assert.NotEmpty(t, result.Warnings)
		assert.Less(t, result.Score, 1.0)
	})
}

func TestValidatorValidateRequest(t *testing.T) {
	config := &SecurityConfig{
		MaxRequestSize: 1024,
		RequireAuth:    true,
		APIKeyHeader:   "X-API-Key",
	}
	
	validator, err := NewRequestValidator(config)
	require.NoError(t, err)

	t.Run("valid request", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "https://example.com/api", nil)
		req.Header.Set("X-API-Key", "test-key-12345")
		req.ContentLength = 100
		
		result := validator.ValidateRequest(req)
		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
	})

	t.Run("oversized request", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "https://example.com/api", nil)
		req.ContentLength = 2048 // Exceeds limit
		
		result := validator.ValidateRequest(req)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0], "exceeds limit")
	})

	t.Run("missing auth", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "https://example.com/api", nil)
		
		result := validator.ValidateRequest(req)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0], "missing authentication")
	})

	t.Run("SQL injection attempt", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "https://example.com/api?id=1' OR '1'='1", nil)
		req.Header.Set("X-API-Key", "test-key")
		
		result := validator.ValidateRequest(req)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0], "SQL injection")
	})

	t.Run("XSS attempt", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "https://example.com/api?name=<script>alert('xss')</script>", nil)
		req.Header.Set("X-API-Key", "test-key")
		
		result := validator.ValidateRequest(req)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0], "XSS attack")
	})

	t.Run("path traversal attempt", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "https://example.com/../../etc/passwd", nil)
		req.Header.Set("X-API-Key", "test-key")
		
		result := validator.ValidateRequest(req)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0], "path traversal")
	})
}

func TestValidateHeaders(t *testing.T) {
	config := &SecurityConfig{
		RequireAuth:  true,
		APIKeyHeader: "X-API-Key",
	}
	
	validator, err := NewRequestValidator(config)
	require.NoError(t, err)

	t.Run("with API key", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "https://example.com", nil)
		req.Header.Set("X-API-Key", "test-key")
		
		err := validator.validateHeaders(req)
		assert.NoError(t, err)
	})

	t.Run("with Authorization header", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "https://example.com", nil)
		req.Header.Set("Authorization", "Bearer token")
		
		err := validator.validateHeaders(req)
		assert.NoError(t, err)
	})

	t.Run("missing auth headers", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "https://example.com", nil)
		
		err := validator.validateHeaders(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing authentication")
	})
}

func TestValidateURL(t *testing.T) {
	validator, err := NewRequestValidator(nil)
	require.NoError(t, err)

	t.Run("valid URL", func(t *testing.T) {
		u, _ := url.Parse("https://example.com/api/v1/users")
		err := validator.validateURL(u)
		assert.NoError(t, err)
	})

	t.Run("path traversal", func(t *testing.T) {
		u, _ := url.Parse("https://example.com/../etc/passwd")
		err := validator.validateURL(u)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid path segment")
	})

	t.Run("long query parameter", func(t *testing.T) {
		longValue := strings.Repeat("a", 1001)
		u, _ := url.Parse("https://example.com/api?param=" + longValue)
		err := validator.validateURL(u)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "too long")
	})

	t.Run("XSS in query", func(t *testing.T) {
		u, _ := url.Parse("https://example.com/api?name=%3Cscript%3Ealert%28%27xss%27%29%3C%2Fscript%3E")
		err := validator.validateURL(u)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "XSS")
	})
}

func TestValidateAuth(t *testing.T) {
	config := &SecurityConfig{
		AllowedAuthMethods: []string{"bearer", "api_key"},
	}
	
	validator, err := NewRequestValidator(config)
	require.NoError(t, err)

	t.Run("valid bearer token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "https://example.com", nil)
		req.Header.Set("Authorization", "Bearer valid-token-12345")
		
		err := validator.validateAuth(req)
		assert.NoError(t, err)
	})

	t.Run("invalid auth format", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "https://example.com", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		
		err := validator.validateAuth(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid authorization header format")
	})

	t.Run("disallowed auth method", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "https://example.com", nil)
		req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
		
		err := validator.validateAuth(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not allowed")
	})

	t.Run("short token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "https://example.com", nil)
		req.Header.Set("Authorization", "Bearer short")
		
		err := validator.validateAuth(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token format")
	})
}