package security

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/orchestre-dev/ccproxy/internal/utils"
)

// SecurityAuditor handles security audit logging
type SecurityAuditor struct {
	config      *SecurityConfig
	logFile     *os.File
	mu          sync.Mutex
	buffer      []AuditEntry
	bufferSize  int
	flushTicker *time.Ticker
	done        chan struct{}
	closeOnce   sync.Once
}

// NewSecurityAuditor creates a new security auditor
func NewSecurityAuditor(config *SecurityConfig) (*SecurityAuditor, error) {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	auditor := &SecurityAuditor{
		config:     config,
		buffer:     make([]AuditEntry, 0, 100),
		bufferSize: 100,
		done:       make(chan struct{}),
	}

	if config.EnableAuditLog {
		// Create audit log directory if needed
		dir := filepath.Dir(config.AuditLogPath)
		if err := os.MkdirAll(dir, 0750); err != nil {
			return nil, fmt.Errorf("failed to create audit log directory: %w", err)
		}

		// Open audit log file
		file, err := os.OpenFile(config.AuditLogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640)
		if err != nil {
			return nil, fmt.Errorf("failed to open audit log: %w", err)
		}
		auditor.logFile = file

		// Start flush ticker
		auditor.flushTicker = time.NewTicker(5 * time.Second)
		go auditor.periodicFlush()
	}

	return auditor, nil
}

// Close closes the auditor
func (a *SecurityAuditor) Close() error {
	var err error
	a.closeOnce.Do(func() {
		close(a.done)

		if a.flushTicker != nil {
			a.flushTicker.Stop()
		}

		// Flush remaining entries
		a.flush()

		if a.logFile != nil {
			err = a.logFile.Close()
		}
	})
	return err
}

// LogSecurityEvent logs a security event
func (a *SecurityAuditor) LogSecurityEvent(event SecurityEvent) {
	if !a.config.EnableAuditLog {
		return
	}

	entry := AuditEntry{
		ID:        event.ID,
		Timestamp: event.Timestamp,
		Type:      "security_event",
		Actor:     event.Source,
		Action:    event.Type,
		Resource:  event.Description,
		Result:    event.Severity,
		Details:   event.Data,
	}

	a.addEntry(entry)

	// Also log to standard logger based on severity
	switch event.Severity {
	case "critical", "high":
		utils.GetLogger().Errorf("Security Event [%s]: %s - %s", event.Severity, event.Type, event.Description)
	case "medium":
		utils.GetLogger().Warnf("Security Event [%s]: %s - %s", event.Severity, event.Type, event.Description)
	default:
		utils.GetLogger().Infof("Security Event [%s]: %s - %s", event.Severity, event.Type, event.Description)
	}
}

// LogAccessAttempt logs an access attempt
func (a *SecurityAuditor) LogAccessAttempt(attempt AccessAttempt) {
	if !a.config.EnableAuditLog {
		return
	}

	details := map[string]interface{}{
		"ip":         attempt.IP,
		"user_agent": attempt.UserAgent,
		"method":     attempt.Method,
		"path":       attempt.Path,
		"success":    attempt.Success,
	}

	if attempt.Reason != "" {
		details["reason"] = attempt.Reason
	}

	if attempt.APIKeyHash != "" {
		details["api_key_hash"] = attempt.APIKeyHash
	}

	entry := AuditEntry{
		ID:        attempt.ID,
		Timestamp: attempt.Timestamp,
		Type:      "access_attempt",
		Actor:     attempt.IP,
		Action:    fmt.Sprintf("%s %s", attempt.Method, attempt.Path),
		Resource:  "api_endpoint",
		Result:    fmt.Sprintf("success=%v", attempt.Success),
		Details:   details,
	}

	a.addEntry(entry)

	// Log failed attempts
	if !attempt.Success {
		utils.GetLogger().Warnf("Failed access attempt from %s: %s", attempt.IP, attempt.Reason)
	}
}

// LogValidationFailure logs a validation failure
func (a *SecurityAuditor) LogValidationFailure(failure ValidationFailure) {
	if !a.config.EnableAuditLog {
		return
	}

	details := map[string]interface{}{
		"errors": failure.Errors,
	}

	if failure.Field != "" {
		details["field"] = failure.Field
	}

	if failure.Context != nil {
		details["context"] = failure.Context
	}

	entry := AuditEntry{
		ID:        failure.ID,
		Timestamp: failure.Timestamp,
		Type:      "validation_failure",
		Actor:     "validator",
		Action:    failure.Type,
		Resource:  "request_validation",
		Result:    "failed",
		Details:   details,
	}

	a.addEntry(entry)

	utils.GetLogger().Warnf("Validation failure [%s]: %v", failure.Type, failure.Errors)
}

// GetAuditTrail retrieves audit entries based on filter
func (a *SecurityAuditor) GetAuditTrail(filter AuditFilter) []AuditEntry {
	a.mu.Lock()
	defer a.mu.Unlock()

	// For now, return entries from buffer
	// In production, this would query from persistent storage
	var filtered []AuditEntry

	for _, entry := range a.buffer {
		if a.matchesFilter(entry, filter) {
			filtered = append(filtered, entry)
		}
	}

	// Apply limit and offset
	start := filter.Offset
	if start > len(filtered) {
		return []AuditEntry{}
	}

	end := start + filter.Limit
	if end > len(filtered) || filter.Limit == 0 {
		end = len(filtered)
	}

	return filtered[start:end]
}

// addEntry adds an entry to the buffer
func (a *SecurityAuditor) addEntry(entry AuditEntry) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.buffer = append(a.buffer, entry)

	// Flush if buffer is full
	if len(a.buffer) >= a.bufferSize {
		a.flushLocked()
	}
}

// flush writes buffered entries to disk
func (a *SecurityAuditor) flush() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.flushLocked()
}

// flushLocked writes buffered entries to disk (must be called with lock held)
func (a *SecurityAuditor) flushLocked() {
	if len(a.buffer) == 0 || a.logFile == nil {
		return
	}

	// Write each entry as JSON line
	for _, entry := range a.buffer {
		data, err := json.Marshal(entry)
		if err != nil {
			utils.GetLogger().Errorf("Failed to marshal audit entry: %v", err)
			continue
		}

		if _, err := a.logFile.Write(append(data, '\n')); err != nil {
			utils.GetLogger().Errorf("Failed to write audit entry: %v", err)
		}
	}

	// Sync to disk
	if err := a.logFile.Sync(); err != nil {
		utils.GetLogger().Errorf("Failed to sync audit log: %v", err)
	}

	// Clear buffer
	a.buffer = a.buffer[:0]
}

// periodicFlush periodically flushes the buffer
func (a *SecurityAuditor) periodicFlush() {
	for {
		select {
		case <-a.flushTicker.C:
			a.flush()
		case <-a.done:
			return
		}
	}
}

// matchesFilter checks if an entry matches the filter
func (a *SecurityAuditor) matchesFilter(entry AuditEntry, filter AuditFilter) bool {
	// Check time range
	if !filter.StartTime.IsZero() && entry.Timestamp.Before(filter.StartTime) {
		return false
	}
	if !filter.EndTime.IsZero() && entry.Timestamp.After(filter.EndTime) {
		return false
	}

	// Check type
	if filter.Type != "" && entry.Type != filter.Type {
		return false
	}

	// Check actor
	if filter.Actor != "" && entry.Actor != filter.Actor {
		return false
	}

	// Check action
	if filter.Action != "" && entry.Action != filter.Action {
		return false
	}

	// Check result
	if filter.Result != "" && entry.Result != filter.Result {
		return false
	}

	return true
}

// LogSecurityAlert logs a security alert
func (a *SecurityAuditor) LogSecurityAlert(alertType, description string, data map[string]interface{}) {
	event := SecurityEvent{
		ID:          uuid.New().String(),
		Type:        alertType,
		Severity:    "high",
		Timestamp:   time.Now(),
		Source:      "security_monitor",
		Description: description,
		Data:        data,
	}

	a.LogSecurityEvent(event)
}

// LogSuspiciousActivity logs suspicious activity
func (a *SecurityAuditor) LogSuspiciousActivity(activityType, source, description string) {
	event := SecurityEvent{
		ID:          uuid.New().String(),
		Type:        "suspicious_activity",
		Severity:    "medium",
		Timestamp:   time.Now(),
		Source:      source,
		Description: description,
		Data: map[string]interface{}{
			"activity_type": activityType,
		},
	}

	a.LogSecurityEvent(event)
}

// RotateLogs rotates old audit logs based on retention policy
func (a *SecurityAuditor) RotateLogs() error {
	if !a.config.EnableAuditLog || a.config.RetentionDays <= 0 {
		return nil
	}

	// Calculate cutoff date
	cutoff := time.Now().AddDate(0, 0, -a.config.RetentionDays)

	// Get audit log directory
	dir := filepath.Dir(a.config.AuditLogPath)

	// Find old log files
	files, err := filepath.Glob(filepath.Join(dir, "audit*.log"))
	if err != nil {
		return fmt.Errorf("failed to list audit logs: %w", err)
	}

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		// Delete if older than retention period
		if info.ModTime().Before(cutoff) {
			if err := os.Remove(file); err != nil {
				utils.GetLogger().Errorf("Failed to remove old audit log %s: %v", file, err)
			} else {
				utils.GetLogger().Infof("Removed old audit log: %s", file)
			}
		}
	}

	return nil
}
