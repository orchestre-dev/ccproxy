package security

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
)

func TestNewSecurityAuditor(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("with audit logging enabled", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = true
		config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

		auditor, err := NewSecurityAuditor(config)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, auditor)
		testutil.AssertNotEqual(t, nil, auditor.logFile)
		testutil.AssertNotEqual(t, nil, auditor.flushTicker)

		auditor.Close()
	})

	t.Run("with audit logging disabled", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = false

		auditor, err := NewSecurityAuditor(config)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, auditor)
		testutil.AssertEqual(t, (*os.File)(nil), auditor.logFile)
		testutil.AssertEqual(t, (*time.Ticker)(nil), auditor.flushTicker)

		auditor.Close()
	})

	t.Run("with nil config", func(t *testing.T) {
		auditor, err := NewSecurityAuditor(nil)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, auditor)

		auditor.Close()
	})

	t.Run("directory creation", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = true

		auditDir := filepath.Join(testConfig.TempDir, "nested", "audit")
		config.AuditLogPath = filepath.Join(auditDir, "audit.log")

		auditor, err := NewSecurityAuditor(config)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, auditor)

		// Check that directory was created
		_, err = os.Stat(auditDir)
		testutil.AssertNoError(t, err)

		auditor.Close()
	})

	t.Run("invalid directory path", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = true
		// Try to create a file under /root (should fail on most systems)
		config.AuditLogPath = "/root/nonexistent/audit.log"

		_, err := NewSecurityAuditor(config)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "failed to create audit log directory")
	})
}

func TestAuditorClose(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("close with logging enabled", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = true
		config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

		auditor, err := NewSecurityAuditor(config)
		testutil.AssertNoError(t, err)

		err = auditor.Close()
		testutil.AssertNoError(t, err)
	})

	t.Run("close with logging disabled", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = false

		auditor, err := NewSecurityAuditor(config)
		testutil.AssertNoError(t, err)

		err = auditor.Close()
		testutil.AssertNoError(t, err)
	})

	t.Run("multiple closes", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = true
		config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

		auditor, err := NewSecurityAuditor(config)
		testutil.AssertNoError(t, err)

		// Multiple closes should not panic
		err1 := auditor.Close()
		err2 := auditor.Close()

		testutil.AssertNoError(t, err1)
		testutil.AssertNoError(t, err2)
	})
}

func TestLogSecurityEvent(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("logs security event with audit enabled", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = true
		config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

		auditor, err := NewSecurityAuditor(config)
		testutil.AssertNoError(t, err)
		defer auditor.Close()

		event := SecurityEvent{
			ID:          "test-event-1",
			Type:        "test_event",
			Severity:    "high",
			Timestamp:   time.Now(),
			Source:      "test",
			Description: "Test security event",
			Data: map[string]interface{}{
				"key": "value",
			},
		}

		auditor.LogSecurityEvent(event)

		// Force flush to ensure write
		auditor.flush()

		// Read and verify log contents
		content, err := os.ReadFile(config.AuditLogPath)
		testutil.AssertNoError(t, err)
		testutil.AssertContains(t, string(content), "test-event-1")
		testutil.AssertContains(t, string(content), "test_event")
		testutil.AssertContains(t, string(content), "high")
	})

	t.Run("ignores event when audit disabled", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = false

		auditor, err := NewSecurityAuditor(config)
		testutil.AssertNoError(t, err)
		defer auditor.Close()

		event := SecurityEvent{
			ID:          "test-event-2",
			Type:        "test_event",
			Severity:    "medium",
			Timestamp:   time.Now(),
			Source:      "test",
			Description: "Test security event",
			Data:        map[string]interface{}{},
		}

		// Should not panic or error
		auditor.LogSecurityEvent(event)
	})

	t.Run("logs different severity levels", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = true
		config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

		auditor, err := NewSecurityAuditor(config)
		testutil.AssertNoError(t, err)
		defer auditor.Close()

		severities := []string{"critical", "high", "medium", "low", "info"}

		for i, severity := range severities {
			event := SecurityEvent{
				ID:          "test-event-" + severity,
				Type:        "test_event",
				Severity:    severity,
				Timestamp:   time.Now(),
				Source:      "test",
				Description: "Test " + severity + " security event",
				Data:        map[string]interface{}{"index": i},
			}

			auditor.LogSecurityEvent(event)
		}

		auditor.flush()

		// Verify all events logged
		content, err := os.ReadFile(config.AuditLogPath)
		testutil.AssertNoError(t, err)

		for _, severity := range severities {
			testutil.AssertContains(t, string(content), severity)
		}
	})
}

func TestLogAccessAttempt(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("logs successful access attempt", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = true
		config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

		auditor, err := NewSecurityAuditor(config)
		testutil.AssertNoError(t, err)
		defer auditor.Close()

		attempt := AccessAttempt{
			ID:         "access-1",
			Timestamp:  time.Now(),
			IP:         "192.168.1.1",
			UserAgent:  "test-agent",
			Method:     "GET",
			Path:       "/api/test",
			Success:    true,
			Reason:     "",
			APIKeyHash: "hash123",
		}

		auditor.LogAccessAttempt(attempt)
		auditor.flush()

		content, err := os.ReadFile(config.AuditLogPath)
		testutil.AssertNoError(t, err)
		testutil.AssertContains(t, string(content), "access-1")
		testutil.AssertContains(t, string(content), "192.168.1.1")
		testutil.AssertContains(t, string(content), "GET")
		testutil.AssertContains(t, string(content), "success=true")
	})

	t.Run("logs failed access attempt", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = true
		config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

		auditor, err := NewSecurityAuditor(config)
		testutil.AssertNoError(t, err)
		defer auditor.Close()

		attempt := AccessAttempt{
			ID:        "access-2",
			Timestamp: time.Now(),
			IP:        "1.2.3.4",
			UserAgent: "bad-agent",
			Method:    "POST",
			Path:      "/api/admin",
			Success:   false,
			Reason:    "invalid credentials",
		}

		auditor.LogAccessAttempt(attempt)
		auditor.flush()

		content, err := os.ReadFile(config.AuditLogPath)
		testutil.AssertNoError(t, err)
		testutil.AssertContains(t, string(content), "access-2")
		testutil.AssertContains(t, string(content), "1.2.3.4")
		testutil.AssertContains(t, string(content), "success=false")
		testutil.AssertContains(t, string(content), "invalid credentials")
	})

	t.Run("ignores attempt when audit disabled", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = false

		auditor, err := NewSecurityAuditor(config)
		testutil.AssertNoError(t, err)
		defer auditor.Close()

		attempt := AccessAttempt{
			ID:        "access-3",
			Timestamp: time.Now(),
			IP:        "192.168.1.1",
			Success:   true,
		}

		// Should not panic
		auditor.LogAccessAttempt(attempt)
	})
}

func TestLogValidationFailure(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("logs validation failure", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = true
		config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

		auditor, err := NewSecurityAuditor(config)
		testutil.AssertNoError(t, err)
		defer auditor.Close()

		failure := ValidationFailure{
			ID:        "validation-1",
			Timestamp: time.Now(),
			Type:      "request_validation",
			Field:     "email",
			Errors:    []string{"invalid format", "too long"},
			Context: map[string]interface{}{
				"request_id": "req-123",
				"ip":         "192.168.1.1",
			},
		}

		auditor.LogValidationFailure(failure)
		auditor.flush()

		content, err := os.ReadFile(config.AuditLogPath)
		testutil.AssertNoError(t, err)
		testutil.AssertContains(t, string(content), "validation-1")
		testutil.AssertContains(t, string(content), "request_validation")
		testutil.AssertContains(t, string(content), "invalid format")
		testutil.AssertContains(t, string(content), "req-123")
	})

	t.Run("logs validation failure without context", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = true
		config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

		auditor, err := NewSecurityAuditor(config)
		testutil.AssertNoError(t, err)
		defer auditor.Close()

		failure := ValidationFailure{
			ID:        "validation-2",
			Timestamp: time.Now(),
			Type:      "input_validation",
			Errors:    []string{"required field missing"},
			Context:   nil,
		}

		auditor.LogValidationFailure(failure)
		auditor.flush()

		content, err := os.ReadFile(config.AuditLogPath)
		testutil.AssertNoError(t, err)
		testutil.AssertContains(t, string(content), "validation-2")
		testutil.AssertContains(t, string(content), "required field missing")
	})

	t.Run("ignores failure when audit disabled", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = false

		auditor, err := NewSecurityAuditor(config)
		testutil.AssertNoError(t, err)
		defer auditor.Close()

		failure := ValidationFailure{
			ID:        "validation-3",
			Timestamp: time.Now(),
			Type:      "test",
			Errors:    []string{"test error"},
		}

		// Should not panic
		auditor.LogValidationFailure(failure)
	})
}

func TestGetAuditTrail(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("returns filtered audit entries", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = true
		config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

		auditor, err := NewSecurityAuditor(config)
		testutil.AssertNoError(t, err)
		defer auditor.Close()

		// Add multiple entries
		now := time.Now()
		entries := []AuditEntry{
			{
				ID:        "audit-1",
				Timestamp: now,
				Type:      "security_event",
				Actor:     "user1",
				Action:    "login",
				Result:    "success",
			},
			{
				ID:        "audit-2",
				Timestamp: now.Add(time.Minute),
				Type:      "access_attempt",
				Actor:     "user2",
				Action:    "access",
				Result:    "failed",
			},
			{
				ID:        "audit-3",
				Timestamp: now.Add(2 * time.Minute),
				Type:      "security_event",
				Actor:     "user1",
				Action:    "logout",
				Result:    "success",
			},
		}

		for _, entry := range entries {
			auditor.addEntry(entry)
		}

		// Test filtering by type
		filter := AuditFilter{
			Type: "security_event",
		}
		results := auditor.GetAuditTrail(filter)
		testutil.AssertEqual(t, 2, len(results))
		testutil.AssertEqual(t, "security_event", results[0].Type)
		testutil.AssertEqual(t, "security_event", results[1].Type)

		// Test filtering by actor
		filter = AuditFilter{
			Actor: "user1",
		}
		results = auditor.GetAuditTrail(filter)
		testutil.AssertEqual(t, 2, len(results))
		testutil.AssertEqual(t, "user1", results[0].Actor)
		testutil.AssertEqual(t, "user1", results[1].Actor)

		// Test limit and offset
		filter = AuditFilter{
			Limit:  1,
			Offset: 1,
		}
		results = auditor.GetAuditTrail(filter)
		testutil.AssertEqual(t, 1, len(results))
		testutil.AssertEqual(t, "audit-2", results[0].ID)
	})

	t.Run("handles empty results", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = false

		auditor, err := NewSecurityAuditor(config)
		testutil.AssertNoError(t, err)
		defer auditor.Close()

		filter := AuditFilter{
			Type: "nonexistent",
		}
		results := auditor.GetAuditTrail(filter)
		testutil.AssertEqual(t, 0, len(results))
	})

	t.Run("handles offset beyond results", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = true
		config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

		auditor, err := NewSecurityAuditor(config)
		testutil.AssertNoError(t, err)
		defer auditor.Close()

		auditor.addEntry(AuditEntry{
			ID:        "audit-1",
			Timestamp: time.Now(),
			Type:      "test",
			Actor:     "test",
		})

		filter := AuditFilter{
			Offset: 10,
		}
		results := auditor.GetAuditTrail(filter)
		testutil.AssertEqual(t, 0, len(results))
	})
}

func TestAuditLogFlushing(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("auto flush on buffer full", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = true
		config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

		auditor, err := NewSecurityAuditor(config)
		testutil.AssertNoError(t, err)
		defer auditor.Close()

		// Add entries to fill buffer (default size is 100)
		for i := 0; i < 101; i++ {
			entry := AuditEntry{
				ID:        "audit-" + string(rune('0'+i%10)),
				Timestamp: time.Now(),
				Type:      "test",
				Actor:     "test",
			}
			auditor.addEntry(entry)
		}

		// Buffer should have been flushed
		content, err := os.ReadFile(config.AuditLogPath)
		testutil.AssertNoError(t, err)
		testutil.AssertTrue(t, len(content) > 0)
	})

	t.Run("manual flush", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = true
		config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

		auditor, err := NewSecurityAuditor(config)
		testutil.AssertNoError(t, err)
		defer auditor.Close()

		entry := AuditEntry{
			ID:        "audit-manual",
			Timestamp: time.Now(),
			Type:      "test",
			Actor:     "test",
		}
		auditor.addEntry(entry)

		// Manually flush
		auditor.flush()

		content, err := os.ReadFile(config.AuditLogPath)
		testutil.AssertNoError(t, err)
		testutil.AssertContains(t, string(content), "audit-manual")
	})

	t.Run("flush with no entries", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = true
		config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

		auditor, err := NewSecurityAuditor(config)
		testutil.AssertNoError(t, err)
		defer auditor.Close()

		// Should not error
		auditor.flush()

		content, err := os.ReadFile(config.AuditLogPath)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, "", string(content))
	})

	t.Run("flush with nil log file", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = false

		auditor, err := NewSecurityAuditor(config)
		testutil.AssertNoError(t, err)
		defer auditor.Close()

		// Should not panic
		auditor.flush()
	})
}

func TestAuditLogRotation(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("rotates old logs", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = true
		config.RetentionDays = 1
		config.AuditLogPath = filepath.Join(testConfig.TempDir, "audit.log")

		auditor, err := NewSecurityAuditor(config)
		testutil.AssertNoError(t, err)
		defer auditor.Close()

		// Create old log files
		oldLogPath := filepath.Join(testConfig.TempDir, "audit-old.log")
		oldFile, err := os.Create(oldLogPath)
		testutil.AssertNoError(t, err)
		oldFile.WriteString("old log content")
		oldFile.Close()

		// Set modification time to 2 days ago
		oldTime := time.Now().AddDate(0, 0, -2)
		err = os.Chtimes(oldLogPath, oldTime, oldTime)
		testutil.AssertNoError(t, err)

		// Create recent log file
		recentLogPath := filepath.Join(testConfig.TempDir, "audit-recent.log")
		recentFile, err := os.Create(recentLogPath)
		testutil.AssertNoError(t, err)
		recentFile.WriteString("recent log content")
		recentFile.Close()

		// Rotate logs
		err = auditor.RotateLogs()
		testutil.AssertNoError(t, err)

		// Old file should be removed
		_, err = os.Stat(oldLogPath)
		testutil.AssertTrue(t, os.IsNotExist(err))

		// Recent file should still exist
		_, err = os.Stat(recentLogPath)
		testutil.AssertNoError(t, err)
	})

	t.Run("rotation disabled", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = false
		config.RetentionDays = 0

		auditor, err := NewSecurityAuditor(config)
		testutil.AssertNoError(t, err)
		defer auditor.Close()

		// Should not error
		err = auditor.RotateLogs()
		testutil.AssertNoError(t, err)
	})

	t.Run("invalid directory for rotation", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = true
		config.RetentionDays = 1
		config.AuditLogPath = "/nonexistent/path/audit.log"

		auditor, err := NewSecurityAuditor(config)
		if err != nil {
			// If we can't create the auditor due to path issues, that's expected
			return
		}
		defer auditor.Close()

		err = auditor.RotateLogs()
		testutil.AssertError(t, err)
	})
}

func TestAuditLogSecurityAlert(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("logs security alert", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = true
		config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

		auditor, err := NewSecurityAuditor(config)
		testutil.AssertNoError(t, err)
		defer auditor.Close()

		alertData := map[string]interface{}{
			"ip":       "1.2.3.4",
			"attempts": 5,
		}

		auditor.LogSecurityAlert("brute_force_detected", "Multiple failed login attempts", alertData)
		auditor.flush()

		content, err := os.ReadFile(config.AuditLogPath)
		testutil.AssertNoError(t, err)
		testutil.AssertContains(t, string(content), "brute_force_detected")
		testutil.AssertContains(t, string(content), "Multiple failed login attempts")
		testutil.AssertContains(t, string(content), "high")
		testutil.AssertContains(t, string(content), "1.2.3.4")
	})
}

func TestLogSuspiciousActivity(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	t.Run("logs suspicious activity", func(t *testing.T) {
		config := DefaultSecurityConfig()
		config.EnableAuditLog = true
		config.AuditLogPath = testutil.CreateTempFile(t, testConfig.TempDir, "audit.log", "")

		auditor, err := NewSecurityAuditor(config)
		testutil.AssertNoError(t, err)
		defer auditor.Close()

		auditor.LogSuspiciousActivity("unusual_access_pattern", "scanner", "Rapid sequential requests")
		auditor.flush()

		content, err := os.ReadFile(config.AuditLogPath)
		testutil.AssertNoError(t, err)
		testutil.AssertContains(t, string(content), "suspicious_activity")
		testutil.AssertContains(t, string(content), "unusual_access_pattern")
		testutil.AssertContains(t, string(content), "Rapid sequential requests")
		testutil.AssertContains(t, string(content), "medium")
	})
}

func TestMatchesFilter(t *testing.T) {
	testConfig := testutil.SetupTest(t)
	defer func() {
		if testConfig.CleanupFunc != nil {
			testConfig.CleanupFunc()
		}
	}()

	config := DefaultSecurityConfig()
	config.EnableAuditLog = false

	auditor, err := NewSecurityAuditor(config)
	testutil.AssertNoError(t, err)
	defer auditor.Close()

	now := time.Now()
	entry := AuditEntry{
		ID:        "test-1",
		Timestamp: now,
		Type:      "security_event",
		Actor:     "user1",
		Action:    "login",
		Result:    "success",
	}

	t.Run("matches all fields", func(t *testing.T) {
		filter := AuditFilter{
			StartTime: now.Add(-time.Minute),
			EndTime:   now.Add(time.Minute),
			Type:      "security_event",
			Actor:     "user1",
			Action:    "login",
			Result:    "success",
		}

		matches := auditor.matchesFilter(entry, filter)
		testutil.AssertTrue(t, matches)
	})

	t.Run("fails time range filter", func(t *testing.T) {
		filter := AuditFilter{
			StartTime: now.Add(time.Minute),
			EndTime:   now.Add(2 * time.Minute),
		}

		matches := auditor.matchesFilter(entry, filter)
		testutil.AssertFalse(t, matches)
	})

	t.Run("fails type filter", func(t *testing.T) {
		filter := AuditFilter{
			Type: "different_type",
		}

		matches := auditor.matchesFilter(entry, filter)
		testutil.AssertFalse(t, matches)
	})

	t.Run("empty filter matches all", func(t *testing.T) {
		filter := AuditFilter{}

		matches := auditor.matchesFilter(entry, filter)
		testutil.AssertTrue(t, matches)
	})
}
