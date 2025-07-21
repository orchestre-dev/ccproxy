package security

import (
	"testing"
	"time"

	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
)

func TestSecurityLevel(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	_ = testConfig

	// Test security level constants
	testutil.AssertEqual(t, SecurityLevelNone, SecurityLevel("none"))
	testutil.AssertEqual(t, SecurityLevelBasic, SecurityLevel("basic"))
	testutil.AssertEqual(t, SecurityLevelStrict, SecurityLevel("strict"))
	testutil.AssertEqual(t, SecurityLevelParanoid, SecurityLevel("paranoid"))
}

func TestValidationResult(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	_ = testConfig

	t.Run("valid result", func(t *testing.T) {
		result := ValidationResult{
			Valid:    true,
			Errors:   []string{},
			Warnings: []string{},
			Score:    1.0,
		}

		testutil.AssertTrue(t, result.Valid)
		testutil.AssertEqual(t, 0, len(result.Errors))
		testutil.AssertEqual(t, 0, len(result.Warnings))
		testutil.AssertEqual(t, 1.0, result.Score)
	})

	t.Run("invalid result with errors", func(t *testing.T) {
		result := ValidationResult{
			Valid:    false,
			Errors:   []string{"error1", "error2"},
			Warnings: []string{"warning1"},
			Score:    0.5,
		}

		testutil.AssertFalse(t, result.Valid)
		testutil.AssertEqual(t, 2, len(result.Errors))
		testutil.AssertEqual(t, 1, len(result.Warnings))
		testutil.AssertEqual(t, 0.5, result.Score)
		testutil.AssertEqual(t, "error1", result.Errors[0])
		testutil.AssertEqual(t, "error2", result.Errors[1])
		testutil.AssertEqual(t, "warning1", result.Warnings[0])
	})
}

func TestDefaultSecurityConfig(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	_ = testConfig

	config := DefaultSecurityConfig()

	// Test default values
	testutil.AssertEqual(t, SecurityLevelBasic, config.Level)
	testutil.AssertFalse(t, config.EnableRequestSigning)
	testutil.AssertTrue(t, config.EnableTLS)
	testutil.AssertEqual(t, "1.2", config.TLSMinVersion)
	testutil.AssertTrue(t, config.EnableRateLimiting)
	testutil.AssertFalse(t, config.EnableIPWhitelist)
	testutil.AssertFalse(t, config.EnableAPIKeyRotation)

	// Test request validation settings
	testutil.AssertEqual(t, int64(10*1024*1024), config.MaxRequestSize) // 10MB
	testutil.AssertEqual(t, 100000, config.MaxTokenLength)
	testutil.AssertEqual(t, 50000, config.MaxPromptLength)
	testutil.AssertEqual(t, 5*time.Minute, config.RequestTimeout)

	// Test content filtering
	testutil.AssertTrue(t, config.EnableContentFilter)
	testutil.AssertEqual(t, 0, len(config.BlockedPatterns))
	testutil.AssertEqual(t, 0, len(config.SensitivePatterns))

	// Test authentication
	testutil.AssertTrue(t, config.RequireAuth)
	testutil.AssertEqual(t, 2, len(config.AllowedAuthMethods))
	testutil.AssertEqual(t, "api_key", config.AllowedAuthMethods[0])
	testutil.AssertEqual(t, "bearer", config.AllowedAuthMethods[1])
	testutil.AssertEqual(t, "X-API-Key", config.APIKeyHeader)

	// Test IP restrictions
	testutil.AssertEqual(t, 0, len(config.AllowedIPs))
	testutil.AssertEqual(t, 0, len(config.BlockedIPs))
	testutil.AssertEqual(t, 0, len(config.TrustedProxies))

	// Test audit and logging
	testutil.AssertTrue(t, config.EnableAuditLog)
	testutil.AssertFalse(t, config.LogSensitiveData)
	testutil.AssertEqual(t, "./audit.log", config.AuditLogPath)
	testutil.AssertEqual(t, 30, config.RetentionDays)
}

func TestStrictSecurityConfig(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	_ = testConfig

	config := StrictSecurityConfig()

	// Test strict security overrides
	testutil.AssertEqual(t, SecurityLevelStrict, config.Level)
	testutil.AssertTrue(t, config.EnableRequestSigning)
	testutil.AssertTrue(t, config.EnableAPIKeyRotation)
	testutil.AssertEqual(t, int64(5*1024*1024), config.MaxRequestSize) // 5MB
	testutil.AssertEqual(t, 50000, config.MaxTokenLength)
	testutil.AssertEqual(t, 25000, config.MaxPromptLength)
	testutil.AssertEqual(t, 2*time.Minute, config.RequestTimeout)
	testutil.AssertTrue(t, config.EnableIPWhitelist)

	// Test that other defaults are preserved
	testutil.AssertTrue(t, config.EnableTLS)
	testutil.AssertTrue(t, config.RequireAuth)
	testutil.AssertTrue(t, config.EnableAuditLog)
}

func TestParanoidSecurityConfig(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	_ = testConfig

	config := ParanoidSecurityConfig()

	// Test paranoid security settings
	testutil.AssertEqual(t, SecurityLevelParanoid, config.Level)
	testutil.AssertEqual(t, int64(1*1024*1024), config.MaxRequestSize) // 1MB
	testutil.AssertEqual(t, 10000, config.MaxTokenLength)
	testutil.AssertEqual(t, 5000, config.MaxPromptLength)
	testutil.AssertEqual(t, 1*time.Minute, config.RequestTimeout)
	testutil.AssertFalse(t, config.LogSensitiveData)
	testutil.AssertEqual(t, 90, config.RetentionDays)

	// Test that strict settings are inherited
	testutil.AssertTrue(t, config.EnableRequestSigning)
	testutil.AssertTrue(t, config.EnableAPIKeyRotation)
	testutil.AssertTrue(t, config.EnableIPWhitelist)
}

func TestSecurityEvent(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	_ = testConfig

	now := time.Now()
	event := SecurityEvent{
		ID:          "test-event-1",
		Type:        "test_event",
		Severity:    "high",
		Timestamp:   now,
		Source:      "test_source",
		Description: "Test security event",
		Data: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
	}

	testutil.AssertEqual(t, "test-event-1", event.ID)
	testutil.AssertEqual(t, "test_event", event.Type)
	testutil.AssertEqual(t, "high", event.Severity)
	testutil.AssertEqual(t, now, event.Timestamp)
	testutil.AssertEqual(t, "test_source", event.Source)
	testutil.AssertEqual(t, "Test security event", event.Description)
	testutil.AssertEqual(t, "value1", event.Data["key1"])
	testutil.AssertEqual(t, 42, event.Data["key2"])
}

func TestAccessAttempt(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	_ = testConfig

	now := time.Now()
	attempt := AccessAttempt{
		ID:         "attempt-1",
		Timestamp:  now,
		IP:         "192.168.1.1",
		UserAgent:  "test-agent",
		Method:     "POST",
		Path:       "/api/test",
		Success:    true,
		Reason:     "",
		APIKey:     "secret-key",
		APIKeyHash: "hash123",
	}

	testutil.AssertEqual(t, "attempt-1", attempt.ID)
	testutil.AssertEqual(t, now, attempt.Timestamp)
	testutil.AssertEqual(t, "192.168.1.1", attempt.IP)
	testutil.AssertEqual(t, "test-agent", attempt.UserAgent)
	testutil.AssertEqual(t, "POST", attempt.Method)
	testutil.AssertEqual(t, "/api/test", attempt.Path)
	testutil.AssertTrue(t, attempt.Success)
	testutil.AssertEqual(t, "", attempt.Reason)
	testutil.AssertEqual(t, "secret-key", attempt.APIKey)
	testutil.AssertEqual(t, "hash123", attempt.APIKeyHash)
}

func TestValidationFailure(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	_ = testConfig

	now := time.Now()
	failure := ValidationFailure{
		ID:        "failure-1",
		Timestamp: now,
		Type:      "request_validation",
		Field:     "email",
		Value:     "invalid-email",
		Errors:    []string{"invalid format", "too short"},
		Context: map[string]interface{}{
			"request_id": "req-123",
		},
	}

	testutil.AssertEqual(t, "failure-1", failure.ID)
	testutil.AssertEqual(t, now, failure.Timestamp)
	testutil.AssertEqual(t, "request_validation", failure.Type)
	testutil.AssertEqual(t, "email", failure.Field)
	testutil.AssertEqual(t, "invalid-email", failure.Value)
	testutil.AssertEqual(t, 2, len(failure.Errors))
	testutil.AssertEqual(t, "invalid format", failure.Errors[0])
	testutil.AssertEqual(t, "too short", failure.Errors[1])
	testutil.AssertEqual(t, "req-123", failure.Context["request_id"])
}

func TestAuditEntry(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	_ = testConfig

	now := time.Now()
	entry := AuditEntry{
		ID:        "audit-1",
		Timestamp: now,
		Type:      "security_event",
		Actor:     "user123",
		Action:    "login",
		Resource:  "authentication",
		Result:    "success",
		Details: map[string]interface{}{
			"ip": "192.168.1.1",
		},
	}

	testutil.AssertEqual(t, "audit-1", entry.ID)
	testutil.AssertEqual(t, now, entry.Timestamp)
	testutil.AssertEqual(t, "security_event", entry.Type)
	testutil.AssertEqual(t, "user123", entry.Actor)
	testutil.AssertEqual(t, "login", entry.Action)
	testutil.AssertEqual(t, "authentication", entry.Resource)
	testutil.AssertEqual(t, "success", entry.Result)
	testutil.AssertEqual(t, "192.168.1.1", entry.Details["ip"])
}

func TestAuditFilter(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	_ = testConfig

	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now()

	filter := AuditFilter{
		StartTime: startTime,
		EndTime:   endTime,
		Type:      "access_attempt",
		Actor:     "user123",
		Action:    "login",
		Result:    "success",
		Limit:     100,
		Offset:    0,
	}

	testutil.AssertEqual(t, startTime, filter.StartTime)
	testutil.AssertEqual(t, endTime, filter.EndTime)
	testutil.AssertEqual(t, "access_attempt", filter.Type)
	testutil.AssertEqual(t, "user123", filter.Actor)
	testutil.AssertEqual(t, "login", filter.Action)
	testutil.AssertEqual(t, "success", filter.Result)
	testutil.AssertEqual(t, 100, filter.Limit)
	testutil.AssertEqual(t, 0, filter.Offset)
}

func TestRateLimitInfo(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	_ = testConfig

	resetTime := time.Now().Add(1 * time.Hour)
	info := RateLimitInfo{
		Key:       "192.168.1.1",
		Limit:     100,
		Window:    time.Hour,
		Used:      25,
		Reset:     resetTime,
		Remaining: 75,
	}

	testutil.AssertEqual(t, "192.168.1.1", info.Key)
	testutil.AssertEqual(t, 100, info.Limit)
	testutil.AssertEqual(t, time.Hour, info.Window)
	testutil.AssertEqual(t, 25, info.Used)
	testutil.AssertEqual(t, resetTime, info.Reset)
	testutil.AssertEqual(t, 75, info.Remaining)
}

func TestIPInfo(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	_ = testConfig

	info := IPInfo{
		IP:          "192.168.1.1",
		Country:     "US",
		Region:      "CA",
		City:        "San Francisco",
		ISP:         "Example ISP",
		ThreatLevel: 0,
		IsProxy:     false,
		IsVPN:       false,
		IsTor:       false,
	}

	testutil.AssertEqual(t, "192.168.1.1", info.IP)
	testutil.AssertEqual(t, "US", info.Country)
	testutil.AssertEqual(t, "CA", info.Region)
	testutil.AssertEqual(t, "San Francisco", info.City)
	testutil.AssertEqual(t, "Example ISP", info.ISP)
	testutil.AssertEqual(t, 0, info.ThreatLevel)
	testutil.AssertFalse(t, info.IsProxy)
	testutil.AssertFalse(t, info.IsVPN)
	testutil.AssertFalse(t, info.IsTor)
}

func TestSecurityConfigCustomization(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	_ = testConfig

	config := DefaultSecurityConfig()
	
	// Test customizing config
	config.Level = SecurityLevelStrict
	config.MaxRequestSize = 1024 * 1024 // 1MB
	config.BlockedPatterns = []string{"\\btest\\b", "\\bdebug\\b"}
	config.AllowedIPs = []string{"192.168.1.1", "10.0.0.1"}

	testutil.AssertEqual(t, SecurityLevelStrict, config.Level)
	testutil.AssertEqual(t, int64(1024*1024), config.MaxRequestSize)
	testutil.AssertEqual(t, 2, len(config.BlockedPatterns))
	testutil.AssertEqual(t, "\\btest\\b", config.BlockedPatterns[0])
	testutil.AssertEqual(t, "\\bdebug\\b", config.BlockedPatterns[1])
	testutil.AssertEqual(t, 2, len(config.AllowedIPs))
	testutil.AssertEqual(t, "192.168.1.1", config.AllowedIPs[0])
	testutil.AssertEqual(t, "10.0.0.1", config.AllowedIPs[1])
}

func TestSecurityConfigEdgeCases(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	_ = testConfig

	t.Run("zero values", func(t *testing.T) {
		config := &SecurityConfig{}
		
		testutil.AssertEqual(t, SecurityLevel(""), config.Level)
		testutil.AssertFalse(t, config.EnableTLS)
		testutil.AssertEqual(t, int64(0), config.MaxRequestSize)
		testutil.AssertEqual(t, 0, config.MaxTokenLength)
		testutil.AssertEqual(t, time.Duration(0), config.RequestTimeout)
	})

	t.Run("negative values", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.MaxRequestSize = -1
		config.MaxTokenLength = -1
		config.RetentionDays = -1

		testutil.AssertEqual(t, int64(-1), config.MaxRequestSize)
		testutil.AssertEqual(t, -1, config.MaxTokenLength)
		testutil.AssertEqual(t, -1, config.RetentionDays)
	})

	t.Run("empty arrays", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.BlockedPatterns = []string{}
		config.AllowedIPs = []string{}
		config.AllowedAuthMethods = []string{}

		testutil.AssertEqual(t, 0, len(config.BlockedPatterns))
		testutil.AssertEqual(t, 0, len(config.AllowedIPs))
		testutil.AssertEqual(t, 0, len(config.AllowedAuthMethods))
	})
}