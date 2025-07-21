package security

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
)

func TestNewRequestValidator(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("with nil config", func(t *testing.T) {
		validator, err := NewRequestValidator(nil)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, validator)
		testutil.AssertEqual(t, SecurityLevelBasic, validator.config.Level)
	})

	t.Run("with custom patterns", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.BlockedPatterns = []string{"\\btest\\b", "\\bdebug\\b"}
		config.SensitivePatterns = []string{"\\bsecret\\b", "\\bpassword\\b"}

		validator, err := NewRequestValidator(config)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, validator)
		testutil.AssertEqual(t, 2, len(validator.compiledPatterns))
		testutil.AssertEqual(t, 2, len(validator.sensitiveRegexps))
	})

	t.Run("with invalid blocked pattern", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.BlockedPatterns = []string{"[invalid regex"}

		_, err := NewRequestValidator(config)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "invalid blocked pattern")
	})

	t.Run("with invalid sensitive pattern", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.SensitivePatterns = []string{"[invalid regex"}

		_, err := NewRequestValidator(config)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "invalid sensitive pattern")
	})
}

func TestValidateGenericData(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("valid data", func(t *testing.T) {
		config := DefaultSecurityConfig()
		validator, err := NewRequestValidator(config)
		testutil.AssertNoError(t, err)

		data := map[string]interface{}{
			"message": "Hello, World!",
			"count":   42,
		}

		result := validator.Validate(data)
		testutil.AssertTrue(t, result.Valid)
		testutil.AssertEqual(t, 1.0, result.Score)
		testutil.AssertEqual(t, 0, len(result.Errors))
	})

	t.Run("data with blocked patterns", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableContentFilter = true
		config.BlockedPatterns = []string{"\\bsecret\\b", "\\bdebug\\b"}

		validator, err := NewRequestValidator(config)
		testutil.AssertNoError(t, err)

		data := map[string]interface{}{
			"message": "This contains secret information",
			"mode":    "debug",
		}

		result := validator.Validate(data)
		testutil.AssertFalse(t, result.Valid)
		testutil.AssertTrue(t, result.Score < 1.0)
		testutil.AssertTrue(t, len(result.Errors) > 0)
		testutil.AssertContains(t, result.Errors[0], "blocked pattern")
	})

	t.Run("data with sensitive patterns", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.SensitivePatterns = []string{"\\bapi_key\\b", "\\bpassword\\b"}

		validator, err := NewRequestValidator(config)
		testutil.AssertNoError(t, err)

		data := map[string]interface{}{
			"auth": "my api_key is here",
			"user": "test_password_field",
		}

		result := validator.Validate(data)
		testutil.AssertTrue(t, result.Valid) // Warnings don't make it invalid
		testutil.AssertTrue(t, result.Score < 1.0)
		testutil.AssertTrue(t, len(result.Warnings) > 0)
		testutil.AssertContains(t, result.Warnings[0], "sensitive data")
	})

	t.Run("content filter disabled", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableContentFilter = false
		config.BlockedPatterns = []string{"\\bsecret\\b"}

		validator, err := NewRequestValidator(config)
		testutil.AssertNoError(t, err)

		data := map[string]interface{}{
			"message": "This contains secret information",
		}

		result := validator.Validate(data)
		testutil.AssertTrue(t, result.Valid)
		testutil.AssertEqual(t, 1.0, result.Score)
	})

	t.Run("unmarshallable data", func(t *testing.T) {
		config := DefaultSecurityConfig()
		validator, err := NewRequestValidator(config)
		testutil.AssertNoError(t, err)

		// Create data that can't be marshaled to JSON
		data := make(chan int)

		result := validator.Validate(data)
		testutil.AssertFalse(t, result.Valid)
		testutil.AssertEqual(t, 0.0, result.Score)
		testutil.AssertContains(t, result.Errors[0], "failed to marshal")
	})
}

func TestValidateHeaders(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("valid headers with auth required", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.RequireAuth = true
		validator, err := NewRequestValidator(config)
		testutil.AssertNoError(t, err)

		req, _ := http.NewRequest("GET", "http://example.com/test", nil)
		req.Header.Set("Authorization", "Bearer token123")

		err = validator.validateHeaders(req)
		testutil.AssertNoError(t, err)
	})

	t.Run("valid headers with API key", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.RequireAuth = true
		validator, err := NewRequestValidator(config)
		testutil.AssertNoError(t, err)

		req, _ := http.NewRequest("GET", "http://example.com/test", nil)
		req.Header.Set(config.APIKeyHeader, "apikey123")

		err = validator.validateHeaders(req)
		testutil.AssertNoError(t, err)
	})

	t.Run("missing auth headers when required", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.RequireAuth = true
		validator, err := NewRequestValidator(config)
		testutil.AssertNoError(t, err)

		req, _ := http.NewRequest("GET", "http://example.com/test", nil)

		err = validator.validateHeaders(req)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "missing authentication")
	})

	t.Run("suspicious headers", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.RequireAuth = false
		validator, err := NewRequestValidator(config)
		testutil.AssertNoError(t, err)

		req, _ := http.NewRequest("GET", "http://example.com/test", nil)
		req.Header.Set("X-Forwarded-Host", "malicious.com")
		req.Header.Set("X-Original-URL", "http://evil.com/admin")

		// Should not error but will log warnings
		err = validator.validateHeaders(req)
		testutil.AssertNoError(t, err)
	})

	t.Run("auth disabled", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.RequireAuth = false
		validator, err := NewRequestValidator(config)
		testutil.AssertNoError(t, err)

		req, _ := http.NewRequest("GET", "http://example.com/test", nil)

		err = validator.validateHeaders(req)
		testutil.AssertNoError(t, err)
	})
}

func TestValidateURL(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	config := DefaultSecurityConfig()
	validator, err := NewRequestValidator(config)
	testutil.AssertNoError(t, err)

	t.Run("valid URL", func(t *testing.T) {
		u, _ := url.Parse("http://example.com/api/users")
		err := validator.validateURL(u)
		testutil.AssertNoError(t, err)
	})

	t.Run("path traversal with dots", func(t *testing.T) {
		u, _ := url.Parse("http://example.com/../../../etc/passwd")
		err := validator.validateURL(u)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "invalid path segment")
	})

	t.Run("path traversal with single dot", func(t *testing.T) {
		u, _ := url.Parse("http://example.com/./secret")
		err := validator.validateURL(u)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "invalid path segment")
	})

	t.Run("long query parameter", func(t *testing.T) {
		longValue := strings.Repeat("a", 1001)
		u, _ := url.Parse("http://example.com/test?param=" + longValue)
		err := validator.validateURL(u)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "too long")
	})

	t.Run("invalid query encoding", func(t *testing.T) {
		// Create URL with invalid percent encoding that will actually fail parsing
		u := &url.URL{
			Scheme:   "http",
			Host:     "example.com",
			Path:     "/test",
			RawQuery: "param=%zz", // Invalid hex characters after %
		}
		err := validator.validateURL(u)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "invalid query parameter encoding")
	})

	t.Run("XSS in query parameters", func(t *testing.T) {
		u, _ := url.Parse("http://example.com/test?q=%3Cscript%3Ealert('xss')%3C/script%3E")
		err := validator.validateURL(u)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "potential XSS")
	})

	t.Run("JavaScript in query parameters", func(t *testing.T) {
		u, _ := url.Parse("http://example.com/test?redirect=javascript:alert('xss')")
		err := validator.validateURL(u)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "potential XSS")
	})

	t.Run("valid long parameter under limit", func(t *testing.T) {
		validValue := strings.Repeat("a", 500)
		u, _ := url.Parse("http://example.com/test?param=" + validValue)
		err := validator.validateURL(u)
		testutil.AssertNoError(t, err)
	})

	t.Run("multiple query parameters", func(t *testing.T) {
		u, _ := url.Parse("http://example.com/test?param1=value1&param2=value2")
		err := validator.validateURL(u)
		testutil.AssertNoError(t, err)
	})
}

func TestValidateAuth(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	config := DefaultSecurityConfig()
	validator, err := NewRequestValidator(config)
	testutil.AssertNoError(t, err)

	t.Run("valid Bearer token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test", nil)
		req.Header.Set("Authorization", "Bearer validtoken123456")

		err := validator.validateAuth(req)
		testutil.AssertNoError(t, err)
	})

	t.Run("valid API key", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test", nil)
		req.Header.Set(config.APIKeyHeader, "validapikey123")

		err := validator.validateAuth(req)
		testutil.AssertNoError(t, err)
	})

	t.Run("invalid auth header format", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test", nil)
		req.Header.Set("Authorization", "InvalidFormat")

		err := validator.validateAuth(req)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "invalid authorization header format")
	})

	t.Run("disallowed auth method", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.AllowedAuthMethods = []string{"bearer"} // Only allow bearer
		validator, err := NewRequestValidator(config)
		testutil.AssertNoError(t, err)

		req, _ := http.NewRequest("GET", "http://example.com/test", nil)
		req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")

		err = validator.validateAuth(req)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "authentication method basic not allowed")
	})

	t.Run("token too short", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test", nil)
		req.Header.Set("Authorization", "Bearer short")

		err := validator.validateAuth(req)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "invalid token format")
	})

	t.Run("API key too short", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test", nil)
		req.Header.Set(config.APIKeyHeader, "short")

		err := validator.validateAuth(req)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "invalid API key format")
	})

	t.Run("case insensitive auth type", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test", nil)
		req.Header.Set("Authorization", "BEARER validtoken123456")

		err := validator.validateAuth(req)
		testutil.AssertNoError(t, err)
	})

	t.Run("no auth headers", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test", nil)

		err := validator.validateAuth(req)
		testutil.AssertNoError(t, err) // No error if no headers, validation happens elsewhere
	})
}

func TestDetectSQLInjection(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	config := DefaultSecurityConfig()
	validator, err := NewRequestValidator(config)
	testutil.AssertNoError(t, err)

	t.Run("union select in URL", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test?id=1 UNION SELECT password FROM users", nil)
		detected := validator.detectSQLInjection(req)
		testutil.AssertTrue(t, detected)
	})

	t.Run("select from in URL", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test?query=SELECT * FROM sensitive_table", nil)
		detected := validator.detectSQLInjection(req)
		testutil.AssertTrue(t, detected)
	})

	t.Run("insert into in URL", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test?cmd=INSERT INTO users VALUES('hacker')", nil)
		detected := validator.detectSQLInjection(req)
		testutil.AssertTrue(t, detected)
	})

	t.Run("delete from in URL", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test?cmd=DELETE FROM users WHERE 1=1", nil)
		detected := validator.detectSQLInjection(req)
		testutil.AssertTrue(t, detected)
	})

	t.Run("drop table in URL", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test?cmd=DROP TABLE users", nil)
		detected := validator.detectSQLInjection(req)
		testutil.AssertTrue(t, detected)
	})

	t.Run("or 1=1 injection", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test?id=1 OR 1=1", nil)
		detected := validator.detectSQLInjection(req)
		testutil.AssertTrue(t, detected)
	})

	t.Run("comment injection", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test?id=1; --", nil)
		detected := validator.detectSQLInjection(req)
		testutil.AssertTrue(t, detected)
	})

	t.Run("quote injection", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test?name=' OR '1'='1", nil)
		detected := validator.detectSQLInjection(req)
		testutil.AssertTrue(t, detected)
	})

	t.Run("SQL injection in headers", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test", nil)
		req.Header.Set("User-Agent", "Mozilla' UNION SELECT password FROM users --")
		detected := validator.detectSQLInjection(req)
		testutil.AssertTrue(t, detected)
	})

	t.Run("case insensitive detection", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test?query=union select password from users", nil)
		detected := validator.detectSQLInjection(req)
		testutil.AssertTrue(t, detected)
	})

	t.Run("safe URL", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test?id=123&name=john", nil)
		detected := validator.detectSQLInjection(req)
		testutil.AssertFalse(t, detected)
	})

	t.Run("false positive words", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test?search=selection criteria", nil)
		detected := validator.detectSQLInjection(req)
		testutil.AssertFalse(t, detected)
	})
}

func TestDetectXSS(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	config := DefaultSecurityConfig()
	validator, err := NewRequestValidator(config)
	testutil.AssertNoError(t, err)

	t.Run("script tag", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test?q=<script>alert('xss')</script>", nil)
		detected := validator.detectXSS(req)
		testutil.AssertTrue(t, detected)
	})

	t.Run("javascript protocol", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test?redirect=javascript:alert('xss')", nil)
		detected := validator.detectXSS(req)
		testutil.AssertTrue(t, detected)
	})

	t.Run("event handler", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test?html=<img onclick='alert(1)'/>", nil)
		detected := validator.detectXSS(req)
		testutil.AssertTrue(t, detected)
	})

	t.Run("iframe tag", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test?content=<iframe src='javascript:alert(1)'></iframe>", nil)
		detected := validator.detectXSS(req)
		testutil.AssertTrue(t, detected)
	})

	t.Run("object tag", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test?content=<object data='javascript:alert(1)'></object>", nil)
		detected := validator.detectXSS(req)
		testutil.AssertTrue(t, detected)
	})

	t.Run("embed tag", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test?content=<embed src='javascript:alert(1)'/>", nil)
		detected := validator.detectXSS(req)
		testutil.AssertTrue(t, detected)
	})

	t.Run("link tag", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test?content=<link rel='stylesheet' href='javascript:alert(1)'/>", nil)
		detected := validator.detectXSS(req)
		testutil.AssertTrue(t, detected)
	})

	t.Run("vbscript protocol", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test?redirect=vbscript:msgbox('xss')", nil)
		detected := validator.detectXSS(req)
		testutil.AssertTrue(t, detected)
	})

	t.Run("safe URL", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test?q=hello world", nil)
		detected := validator.detectXSS(req)
		testutil.AssertFalse(t, detected)
	})

	t.Run("safe script reference", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test?file=script.js", nil)
		detected := validator.detectXSS(req)
		testutil.AssertFalse(t, detected)
	})
}

func TestDetectPathTraversal(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	config := DefaultSecurityConfig()
	validator, err := NewRequestValidator(config)
	testutil.AssertNoError(t, err)

	t.Run("basic path traversal", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/../../../etc/passwd", nil)
		detected := validator.detectPathTraversal(req)
		testutil.AssertTrue(t, detected)
	})

	t.Run("windows path traversal", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/..\\..\\..\\windows\\system32\\config\\sam", nil)
		detected := validator.detectPathTraversal(req)
		testutil.AssertTrue(t, detected)
	})

	t.Run("URL encoded path traversal", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/%2e%2e%2f%2e%2e%2f%2e%2e%2fetc/passwd", nil)
		detected := validator.detectPathTraversal(req)
		testutil.AssertTrue(t, detected)
	})

	t.Run("URL encoded windows path traversal", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/%2e%2e%5c%2e%2e%5c%2e%2e%5cwindows", nil)
		detected := validator.detectPathTraversal(req)
		testutil.AssertTrue(t, detected)
	})

	t.Run("double encoded path traversal", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test", nil)
		// Set path directly to test the pattern
		req.URL.Path = "/%%32%65%%32%65%%32%66etc/passwd"
		detected := validator.detectPathTraversal(req)
		testutil.AssertTrue(t, detected)
	})

	t.Run("unicode path traversal", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test", nil)
		req.URL.Path = "/..%c0%af/etc/passwd"
		detected := validator.detectPathTraversal(req)
		testutil.AssertTrue(t, detected)
	})

	t.Run("alternative unicode encoding", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/test", nil)
		req.URL.Path = "/..%c1%9c/etc/passwd"
		detected := validator.detectPathTraversal(req)
		testutil.AssertTrue(t, detected)
	})

	t.Run("safe path", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/api/users/123", nil)
		detected := validator.detectPathTraversal(req)
		testutil.AssertFalse(t, detected)
	})

	t.Run("safe relative path", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/images/profile.jpg", nil)
		detected := validator.detectPathTraversal(req)
		testutil.AssertFalse(t, detected)
	})
}

func TestValidateRequestIntegration(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("comprehensive attack detection", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.RequireAuth = false // Disable for testing attack detection
		validator, err := NewRequestValidator(config)
		testutil.AssertNoError(t, err)

		// Request with multiple attack vectors
		req, _ := http.NewRequest("POST", "http://example.com/../admin?cmd=DROP TABLE users;--&redirect=javascript:alert('xss')", nil)
		req.ContentLength = 500 // Within limits

		result := validator.ValidateRequest(req)
		testutil.AssertFalse(t, result.Valid)
		testutil.AssertTrue(t, len(result.Errors) > 0)

		// Should detect multiple attack types
		errorStr := strings.Join(result.Errors, " ")
		testutil.AssertTrue(t,
			strings.Contains(errorStr, "XSS") ||
				strings.Contains(errorStr, "SQL") ||
				strings.Contains(errorStr, "path traversal"))
	})

	t.Run("valid request passes all checks", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.RequireAuth = false
		validator, err := NewRequestValidator(config)
		testutil.AssertNoError(t, err)

		req, _ := http.NewRequest("GET", "http://example.com/api/users?page=1&limit=10", nil)
		req.ContentLength = 0
		req.Header.Set("User-Agent", "MyApp/1.0")

		result := validator.ValidateRequest(req)
		testutil.AssertTrue(t, result.Valid)
		testutil.AssertEqual(t, 1.0, result.Score)
		testutil.AssertEqual(t, 0, len(result.Errors))
	})
}

func TestValidateResponse(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("valid response", func(t *testing.T) {
		config := DefaultSecurityConfig()
		validator, err := NewRequestValidator(config)
		testutil.AssertNoError(t, err)

		response := map[string]interface{}{
			"status": "success",
			"data":   []string{"item1", "item2"},
		}

		result := validator.ValidateResponse(response)
		testutil.AssertTrue(t, result.Valid)
		testutil.AssertEqual(t, 1.0, result.Score)
	})

	t.Run("response too large", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.MaxRequestSize = 100 // Very small limit for testing
		validator, err := NewRequestValidator(config)
		testutil.AssertNoError(t, err)

		largeData := strings.Repeat("a", 1000)
		response := map[string]interface{}{
			"data": largeData,
		}

		result := validator.ValidateResponse(response)
		testutil.AssertTrue(t, result.Valid) // Still valid, just a warning
		testutil.AssertTrue(t, len(result.Warnings) > 0)
		testutil.AssertTrue(t, result.Score < 1.0)
		testutil.AssertContains(t, result.Warnings[0], "exceeds recommended limit")
	})

	t.Run("response with blocked content", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableContentFilter = true
		config.BlockedPatterns = []string{"\\bsecret\\b"}
		validator, err := NewRequestValidator(config)
		testutil.AssertNoError(t, err)

		response := map[string]interface{}{
			"message": "This contains secret information",
		}

		result := validator.ValidateResponse(response)
		testutil.AssertFalse(t, result.Valid)
		testutil.AssertTrue(t, len(result.Errors) > 0)
	})

	t.Run("response validation error in content", func(t *testing.T) {
		config := DefaultSecurityConfig()
		validator, err := NewRequestValidator(config)
		testutil.AssertNoError(t, err)

		// Create response that will fail JSON marshaling during content validation
		response := make(chan int)

		result := validator.ValidateResponse(response)
		testutil.AssertFalse(t, result.Valid)
		testutil.AssertTrue(t, len(result.Errors) > 0)
	})
}
