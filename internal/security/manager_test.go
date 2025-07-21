package security

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	t.Run("with default config", func(t *testing.T) {
		manager, err := NewManager(nil)
		assert.NoError(t, err)
		assert.NotNil(t, manager)
		assert.NotNil(t, manager.config)
		assert.NotNil(t, manager.validator)
		assert.NotNil(t, manager.sanitizer)
		assert.NotNil(t, manager.auditor)
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &SecurityConfig{
			EnableRateLimiting:   true,
			EnableAPIKeyRotation: false,
			AllowedIPs:          []string{"192.168.1.1"},
			BlockedIPs:          []string{"10.0.0.1"},
		}
		
		manager, err := NewManager(config)
		assert.NoError(t, err)
		assert.NotNil(t, manager)
		assert.NotNil(t, manager.rateLimiter)
		assert.Nil(t, manager.keyRotator)
		assert.True(t, manager.ipWhitelist["192.168.1.1"])
		assert.True(t, manager.ipBlacklist["10.0.0.1"])
	})
}

func TestValidateRequest(t *testing.T) {
	config := &SecurityConfig{
		RequireAuth:        false,
		EnableRateLimiting: true,
		BlockedIPs:         []string{"10.0.0.1"},
	}
	
	manager, err := NewManager(config)
	require.NoError(t, err)

	t.Run("valid request", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "https://example.com/api", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		
		err := manager.ValidateRequest(req)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), manager.requestCount)
	})

	t.Run("blocked IP", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "https://example.com/api", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		
		err := manager.ValidateRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "blocked")
		assert.Equal(t, int64(1), manager.blockedCount)
	})

	t.Run("rate limit exceeded", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "https://example.com/api", nil)
		req.RemoteAddr = "192.168.1.2:1234"
		
		// Make many requests to exceed rate limit
		for i := 0; i < 101; i++ {
			manager.ValidateRequest(req)
		}
		
		err := manager.ValidateRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate limit")
	})
}

func TestAPIKeyManagement(t *testing.T) {
	manager, err := NewManager(nil)
	require.NoError(t, err)

	t.Run("generate API key", func(t *testing.T) {
		permissions := []string{"read", "write"}
		key, err := manager.GenerateAPIKey(permissions, 100)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, key)
		assert.Len(t, manager.apiKeys, 1)
		
		// Validate the generated key
		err = manager.ValidateAPIKey(key)
		assert.NoError(t, err)
	})

	t.Run("validate invalid key", func(t *testing.T) {
		err := manager.ValidateAPIKey("invalid-key")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid API key")
	})

	t.Run("revoke API key", func(t *testing.T) {
		// Generate a key first
		key, err := manager.GenerateAPIKey([]string{"read"}, 50)
		require.NoError(t, err)
		
		// Revoke it
		err = manager.RevokeAPIKey(key)
		assert.NoError(t, err)
		
		// Try to use revoked key
		err = manager.ValidateAPIKey(key)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "inactive")
	})

	t.Run("revoke non-existent key", func(t *testing.T) {
		err := manager.RevokeAPIKey("non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name           string
		trustedProxies []string
		headers        map[string]string
		remoteAddr     string
		expected       string
	}{
		{
			name:       "direct connection",
			remoteAddr: "192.168.1.1:1234",
			expected:   "192.168.1.1",
		},
		{
			name:           "X-Forwarded-For with trusted proxy",
			trustedProxies: []string{"10.0.0.1"},
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.1, 10.0.0.1",
			},
			remoteAddr: "10.0.0.1:1234",
			expected:   "203.0.113.1",
		},
		{
			name:           "X-Real-IP with trusted proxy",
			trustedProxies: []string{"10.0.0.1"},
			headers: map[string]string{
				"X-Real-IP": "203.0.113.2",
			},
			remoteAddr: "10.0.0.1:1234",
			expected:   "203.0.113.2",
		},
		{
			name: "untrusted proxy headers ignored",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.1",
			},
			remoteAddr: "192.168.1.1:1234",
			expected:   "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &SecurityConfig{
				TrustedProxies: tt.trustedProxies,
			}
			
			manager, err := NewManager(config)
			require.NoError(t, err)
			
			req, _ := http.NewRequest("GET", "https://example.com", nil)
			req.RemoteAddr = tt.remoteAddr
			
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			
			ip := manager.getClientIP(req)
			assert.Equal(t, tt.expected, ip)
		})
	}
}

func TestIPRestrictions(t *testing.T) {
	config := &SecurityConfig{
		EnableIPWhitelist: true,
		AllowedIPs:       []string{"192.168.1.1", "192.168.1.2"},
		BlockedIPs:       []string{"10.0.0.1"},
	}
	
	manager, err := NewManager(config)
	require.NoError(t, err)

	t.Run("whitelisted IP", func(t *testing.T) {
		err := manager.checkIPRestrictions("192.168.1.1")
		assert.NoError(t, err)
	})

	t.Run("blacklisted IP", func(t *testing.T) {
		err := manager.checkIPRestrictions("10.0.0.1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "blocked")
	})

	t.Run("non-whitelisted IP", func(t *testing.T) {
		err := manager.checkIPRestrictions("192.168.1.3")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not in whitelist")
	})

	t.Run("add/remove IP whitelist", func(t *testing.T) {
		// Add IP
		manager.AddIPToWhitelist("192.168.1.3")
		err := manager.checkIPRestrictions("192.168.1.3")
		assert.NoError(t, err)
		
		// Remove IP
		manager.RemoveIPFromWhitelist("192.168.1.3")
		err = manager.checkIPRestrictions("192.168.1.3")
		assert.Error(t, err)
	})

	t.Run("add/remove IP blacklist", func(t *testing.T) {
		// Add IP
		manager.AddIPToBlacklist("192.168.1.4")
		err := manager.checkIPRestrictions("192.168.1.4")
		assert.Error(t, err)
		
		// Remove IP
		manager.RemoveIPFromBlacklist("192.168.1.4")
		err = manager.checkIPRestrictions("192.168.1.4")
		assert.Error(t, err) // Still fails due to whitelist
	})
}

func TestGetMetrics(t *testing.T) {
	// Create manager with auth disabled for this test
	config := &SecurityConfig{
		RequireAuth: false,
		Level:       SecurityLevelBasic,
	}
	manager, err := NewManager(config)
	require.NoError(t, err)

	// Generate some activity
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	
	err = manager.ValidateRequest(req)
	assert.NoError(t, err)
	
	manager.GenerateAPIKey([]string{"read"}, 100)
	
	metrics := manager.GetMetrics()
	
	assert.Equal(t, int64(1), metrics["total_requests"])
	assert.Equal(t, int64(0), metrics["blocked_requests"])
	assert.Equal(t, int64(0), metrics["validation_failures"])
	assert.Equal(t, 1, metrics["active_api_keys"])
	assert.Equal(t, SecurityLevelBasic, metrics["security_level"])
}

func TestHashAPIKey(t *testing.T) {
	manager, err := NewManager(nil)
	require.NoError(t, err)

	key := "test-api-key-12345"
	hash1 := manager.hashAPIKey(key)
	hash2 := manager.hashAPIKey(key)
	
	// Same key should produce same hash
	assert.Equal(t, hash1, hash2)
	
	// Different key should produce different hash
	differentHash := manager.hashAPIKey("different-key")
	assert.NotEqual(t, hash1, differentHash)
	
	// Hash should be hex string
	assert.Regexp(t, "^[a-f0-9]{64}$", hash1)
}