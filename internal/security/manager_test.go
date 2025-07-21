package security

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
)

func TestNewManager(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("with default config", func(t *testing.T) {
		manager, err := NewManager(nil)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, manager)
		testutil.AssertEqual(t, SecurityLevelBasic, manager.config.Level)
		testutil.AssertNotEqual(t, nil, manager.validator)
		testutil.AssertNotEqual(t, nil, manager.sanitizer)
		testutil.AssertNotEqual(t, nil, manager.auditor)

		// Cleanup
		manager.Close()
	})

	t.Run("with custom config", func(t *testing.T) {
		config := StrictSecurityConfig()
		config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

		manager, err := NewManager(config)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, manager)
		testutil.AssertEqual(t, SecurityLevelStrict, manager.config.Level)

		// Cleanup
		manager.Close()
	})

	t.Run("with IP restrictions", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.AllowedIPs = []string{"192.168.1.1", "10.0.0.1"}
		config.BlockedIPs = []string{"1.2.3.4", "5.6.7.8"}
		config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

		manager, err := NewManager(config)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, manager)

		// Check IP lists are initialized
		testutil.AssertTrue(t, manager.ipWhitelist["192.168.1.1"])
		testutil.AssertTrue(t, manager.ipWhitelist["10.0.0.1"])
		testutil.AssertTrue(t, manager.ipBlacklist["1.2.3.4"])
		testutil.AssertTrue(t, manager.ipBlacklist["5.6.7.8"])

		// Cleanup
		manager.Close()
	})
}

func TestManagerValidateRequest(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	config := DefaultSecurityConfig()
	config.RequireAuth = false        // Disable auth for these tests
	config.EnableRateLimiting = false // Disable rate limiting for these tests
	config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

	manager, err := NewManager(config)
	testutil.AssertNoError(t, err)
	defer manager.Close()

	t.Run("valid request", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://example.com/api/test", nil)
		testutil.AssertNoError(t, err)
		req.RemoteAddr = "192.168.1.1:12345"

		err = manager.ValidateRequest(req)
		testutil.AssertNoError(t, err)

		// Check metrics
		metrics := manager.GetMetrics()
		testutil.AssertTrue(t, metrics["total_requests"].(int64) > 0)
	})

	t.Run("blocked IP", func(t *testing.T) {
		manager.AddIPToBlacklist("1.2.3.4")

		req, err := http.NewRequest("GET", "http://example.com/api/test", nil)
		testutil.AssertNoError(t, err)
		req.RemoteAddr = "1.2.3.4:12345"

		err = manager.ValidateRequest(req)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "blocked", "Error should mention blocked IP")

		// Check metrics
		metrics := manager.GetMetrics()
		testutil.AssertTrue(t, metrics["blocked_requests"].(int64) > 0)
	})

	t.Run("IP whitelist enabled", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableIPWhitelist = true
		config.AllowedIPs = []string{"192.168.1.1"}
		config.RequireAuth = false
		config.EnableRateLimiting = false
		config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit2.log", "")

		manager2, err := NewManager(config)
		testutil.AssertNoError(t, err)
		defer manager2.Close()

		// Request from allowed IP
		req, err := http.NewRequest("GET", "http://example.com/api/test", nil)
		testutil.AssertNoError(t, err)
		req.RemoteAddr = "192.168.1.1:12345"

		err = manager2.ValidateRequest(req)
		testutil.AssertNoError(t, err)

		// Request from non-allowed IP
		req.RemoteAddr = "10.0.0.1:12345"
		err = manager2.ValidateRequest(req)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "whitelist", "Error should mention whitelist")
	})

	t.Run("request too large", func(t *testing.T) {
		req, err := http.NewRequest("POST", "http://example.com/api/test", strings.NewReader("test body"))
		testutil.AssertNoError(t, err)
		req.RemoteAddr = "192.168.1.1:12345"
		req.ContentLength = manager.config.MaxRequestSize + 1

		err = manager.ValidateRequest(req)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "exceeds limit", "Error should mention size limit")
	})
}

func TestManagerAPIKeyManagement(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	config := DefaultSecurityConfig()
	config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

	manager, err := NewManager(config)
	testutil.AssertNoError(t, err)
	defer manager.Close()

	t.Run("generate API key", func(t *testing.T) {
		permissions := []string{"read", "write"}
		rateLimit := 100

		key, err := manager.GenerateAPIKey(permissions, rateLimit)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, "", key)
		testutil.AssertTrue(t, len(key) > 10)

		// Validate the generated key
		err = manager.ValidateAPIKey(key)
		testutil.AssertNoError(t, err)
	})

	t.Run("validate invalid API key", func(t *testing.T) {
		err := manager.ValidateAPIKey("invalid-key")
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "invalid", "Error should mention invalid key")
	})

	t.Run("revoke API key", func(t *testing.T) {
		key, err := manager.GenerateAPIKey([]string{"read"}, 50)
		testutil.AssertNoError(t, err)

		// Key should be valid initially
		err = manager.ValidateAPIKey(key)
		testutil.AssertNoError(t, err)

		// Revoke the key
		err = manager.RevokeAPIKey(key)
		testutil.AssertNoError(t, err)

		// Key should now be invalid
		err = manager.ValidateAPIKey(key)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "inactive", "Error should mention inactive key")
	})

	t.Run("revoke non-existent key", func(t *testing.T) {
		err := manager.RevokeAPIKey("non-existent-key")
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "not found", "Error should mention key not found")
	})
}

func TestManagerIPManagement(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	config := DefaultSecurityConfig()
	config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

	manager, err := NewManager(config)
	testutil.AssertNoError(t, err)
	defer manager.Close()

	t.Run("manage whitelist", func(t *testing.T) {
		ip := "192.168.1.100"

		// Add to whitelist
		manager.AddIPToWhitelist(ip)
		testutil.AssertTrue(t, manager.ipWhitelist[ip])

		// Remove from whitelist
		manager.RemoveIPFromWhitelist(ip)
		testutil.AssertFalse(t, manager.ipWhitelist[ip])
	})

	t.Run("manage blacklist", func(t *testing.T) {
		ip := "1.2.3.5"

		// Add to blacklist
		manager.AddIPToBlacklist(ip)
		testutil.AssertTrue(t, manager.ipBlacklist[ip])

		// Remove from blacklist
		manager.RemoveIPFromBlacklist(ip)
		testutil.AssertFalse(t, manager.ipBlacklist[ip])
	})
}

func TestManagerGetClientIP(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	config := DefaultSecurityConfig()
	config.TrustedProxies = []string{"proxy1", "proxy2"}
	config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

	manager, err := NewManager(config)
	testutil.AssertNoError(t, err)
	defer manager.Close()

	t.Run("X-Forwarded-For header", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://example.com/test", nil)
		testutil.AssertNoError(t, err)
		req.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.1")
		req.RemoteAddr = "proxy:12345"

		ip := manager.getClientIP(req)
		testutil.AssertEqual(t, "192.168.1.1", ip)
	})

	t.Run("X-Real-IP header", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://example.com/test", nil)
		testutil.AssertNoError(t, err)
		req.Header.Set("X-Real-IP", "192.168.1.2")
		req.RemoteAddr = "proxy:12345"

		ip := manager.getClientIP(req)
		testutil.AssertEqual(t, "192.168.1.2", ip)
	})

	t.Run("RemoteAddr fallback", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://example.com/test", nil)
		testutil.AssertNoError(t, err)
		req.RemoteAddr = "192.168.1.3:12345"

		ip := manager.getClientIP(req)
		testutil.AssertEqual(t, "192.168.1.3", ip)
	})

	t.Run("invalid RemoteAddr", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://example.com/test", nil)
		testutil.AssertNoError(t, err)
		req.RemoteAddr = "invalid-addr"

		ip := manager.getClientIP(req)
		testutil.AssertEqual(t, "invalid-addr", ip)
	})
}

func TestManagerRateLimit(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	config := DefaultSecurityConfig()
	config.EnableRateLimiting = true
	config.RequireAuth = false
	config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

	manager, err := NewManager(config)
	testutil.AssertNoError(t, err)
	defer manager.Close()

	t.Run("rate limit enforcement", func(t *testing.T) {
		// Create requests from same IP
		ip := "192.168.1.10"

		// First request should pass
		req, err := http.NewRequest("GET", "http://example.com/test", nil)
		testutil.AssertNoError(t, err)
		req.RemoteAddr = ip + ":12345"

		err = manager.ValidateRequest(req)
		testutil.AssertNoError(t, err)

		// Exhaust rate limit
		for i := 0; i < 100; i++ {
			req, err := http.NewRequest("GET", "http://example.com/test", nil)
			testutil.AssertNoError(t, err)
			req.RemoteAddr = ip + ":12345"
			manager.ValidateRequest(req) // Ignore errors in loop
		}

		// Next request should be rate limited
		req, err = http.NewRequest("GET", "http://example.com/test", nil)
		testutil.AssertNoError(t, err)
		req.RemoteAddr = ip + ":12345"

		err = manager.ValidateRequest(req)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "rate limit", "Error should mention rate limit")
	})
}

func TestManagerSanitization(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	config := DefaultSecurityConfig()
	config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

	manager, err := NewManager(config)
	testutil.AssertNoError(t, err)
	defer manager.Close()

	t.Run("sanitize request", func(t *testing.T) {
		data := map[string]interface{}{
			"message": "Hello <script>alert('xss')</script>",
			"api_key": "secret-key-123",
		}

		sanitized, err := manager.SanitizeRequest(data)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, sanitized)
	})

	t.Run("sanitize response", func(t *testing.T) {
		data := map[string]interface{}{
			"result": "success",
			"token":  "sensitive-token",
		}

		sanitized, err := manager.SanitizeResponse(data)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, sanitized)
	})
}

func TestManagerMetrics(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	config := DefaultSecurityConfig()
	config.RequireAuth = false
	config.EnableRateLimiting = false
	config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

	manager, err := NewManager(config)
	testutil.AssertNoError(t, err)
	defer manager.Close()

	t.Run("initial metrics", func(t *testing.T) {
		metrics := manager.GetMetrics()
		testutil.AssertEqual(t, int64(0), metrics["total_requests"])
		testutil.AssertEqual(t, int64(0), metrics["blocked_requests"])
		testutil.AssertEqual(t, int64(0), metrics["validation_failures"])
		testutil.AssertEqual(t, SecurityLevelBasic, metrics["security_level"])
	})

	t.Run("metrics after requests", func(t *testing.T) {
		// Valid request
		req, err := http.NewRequest("GET", "http://example.com/test", nil)
		testutil.AssertNoError(t, err)
		req.RemoteAddr = "192.168.1.1:12345"

		err = manager.ValidateRequest(req)
		testutil.AssertNoError(t, err)

		// Blocked request
		manager.AddIPToBlacklist("1.2.3.4")
		req.RemoteAddr = "1.2.3.4:12345"
		err = manager.ValidateRequest(req)
		testutil.AssertError(t, err)

		metrics := manager.GetMetrics()
		testutil.AssertTrue(t, metrics["total_requests"].(int64) > 0)
		testutil.AssertTrue(t, metrics["blocked_requests"].(int64) > 0)
	})
}

func TestManagerWithAuth(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	config := DefaultSecurityConfig()
	config.RequireAuth = true
	config.EnableRateLimiting = false
	config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

	manager, err := NewManager(config)
	testutil.AssertNoError(t, err)
	defer manager.Close()

	// Generate a valid API key
	apiKey, err := manager.GenerateAPIKey([]string{"read"}, 100)
	testutil.AssertNoError(t, err)

	t.Run("request with valid API key", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://example.com/test", nil)
		testutil.AssertNoError(t, err)
		req.RemoteAddr = "192.168.1.1:12345"
		req.Header.Set(config.APIKeyHeader, apiKey)

		err = manager.ValidateRequest(req)
		testutil.AssertNoError(t, err)
	})

	t.Run("request without API key", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://example.com/test", nil)
		testutil.AssertNoError(t, err)
		req.RemoteAddr = "192.168.1.1:12345"

		err = manager.ValidateRequest(req)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "authentication", "Error should mention missing authentication")
	})
}

func TestManagerXSSDetection(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	config := DefaultSecurityConfig()
	config.RequireAuth = false
	config.EnableRateLimiting = false
	config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

	manager, err := NewManager(config)
	testutil.AssertNoError(t, err)
	defer manager.Close()

	t.Run("XSS in URL", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://example.com/test?q=<script>alert('xss')</script>", nil)
		testutil.AssertNoError(t, err)
		req.RemoteAddr = "192.168.1.1:12345"

		err = manager.ValidateRequest(req)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "XSS", "Error should mention XSS")
	})

	t.Run("JavaScript in URL", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://example.com/test?redirect=javascript:alert('xss')", nil)
		testutil.AssertNoError(t, err)
		req.RemoteAddr = "192.168.1.1:12345"

		err = manager.ValidateRequest(req)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "XSS", "Error should mention XSS")
	})
}

func TestManagerPathTraversalDetection(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	config := DefaultSecurityConfig()
	config.RequireAuth = false
	config.EnableRateLimiting = false
	config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

	manager, err := NewManager(config)
	testutil.AssertNoError(t, err)
	defer manager.Close()

	t.Run("path traversal with dots", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://example.com/../../../etc/passwd", nil)
		testutil.AssertNoError(t, err)
		req.RemoteAddr = "192.168.1.1:12345"

		err = manager.ValidateRequest(req)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "path traversal", "Error should mention path traversal")
	})

	t.Run("encoded path traversal", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://example.com/%2e%2e%2f%2e%2e%2f%2e%2e%2fetc/passwd", nil)
		testutil.AssertNoError(t, err)
		req.RemoteAddr = "192.168.1.1:12345"

		err = manager.ValidateRequest(req)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "path traversal", "Error should mention path traversal")
	})
}

func TestManagerValidateResponse(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	config := DefaultSecurityConfig()
	config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

	manager, err := NewManager(config)
	testutil.AssertNoError(t, err)
	defer manager.Close()

	t.Run("valid response", func(t *testing.T) {
		response := map[string]interface{}{
			"status": "success",
			"data":   "Hello, World!",
		}

		err := manager.ValidateResponse(response)
		testutil.AssertNoError(t, err)
	})

	t.Run("response with blocked content", func(t *testing.T) {
		// Update config to block certain patterns
		config.BlockedPatterns = []string{"secret"}
		config.EnableContentFilter = true

		manager2, err := NewManager(config)
		testutil.AssertNoError(t, err)
		defer manager2.Close()

		response := map[string]interface{}{
			"status": "success",
			"data":   "This contains secret information",
		}

		err = manager2.ValidateResponse(response)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "blocked pattern", "Error should mention blocked pattern")
	})
}

func TestManagerEdgeCases(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("nil request", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

		manager, err := NewManager(config)
		testutil.AssertNoError(t, err)
		defer manager.Close()

		// This should not panic
		err = manager.ValidateRequest(nil)
		testutil.AssertError(t, err)
	})

	t.Run("empty URL", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.RequireAuth = false
		config.EnableRateLimiting = false
		config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

		manager, err := NewManager(config)
		testutil.AssertNoError(t, err)
		defer manager.Close()

		req := &http.Request{
			Method:     "GET",
			URL:        &url.URL{},
			RemoteAddr: "192.168.1.1:12345",
			Header:     make(http.Header),
		}

		err = manager.ValidateRequest(req)
		testutil.AssertNoError(t, err)
	})

	t.Run("close multiple times", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

		manager, err := NewManager(config)
		testutil.AssertNoError(t, err)

		// Should not panic on multiple closes
		manager.Close()
		manager.Close()
	})
}
