package security

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/orchestre-dev/ccproxy/internal/errors"
	"github.com/orchestre-dev/ccproxy/internal/utils"
)

// Manager coordinates all security components
type Manager struct {
	config    *SecurityConfig
	validator *RequestValidator
	sanitizer *DataSanitizer
	auditor   *SecurityAuditor

	// IP management
	ipWhitelist map[string]bool
	ipBlacklist map[string]bool
	ipMu        sync.RWMutex

	// Rate limiting
	rateLimiter *IPRateLimiter

	// API key management
	apiKeys    map[string]APIKeyInfo
	keyMu      sync.RWMutex
	keyRotator *time.Ticker

	// Metrics
	requestCount   int64
	blockedCount   int64
	validationFail int64
	mu             sync.RWMutex
}

// APIKeyInfo stores API key information
type APIKeyInfo struct {
	Key         string
	Hash        string
	Created     time.Time
	LastUsed    time.Time
	Permissions []string
	RateLimit   int
	Active      bool
}

// NewManager creates a new security manager
func NewManager(config *SecurityConfig) (*Manager, error) {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	validator, err := NewRequestValidator(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create validator: %w", err)
	}

	sanitizer, err := NewDataSanitizer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create sanitizer: %w", err)
	}

	auditor, err := NewSecurityAuditor(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create auditor: %w", err)
	}

	manager := &Manager{
		config:      config,
		validator:   validator,
		sanitizer:   sanitizer,
		auditor:     auditor,
		ipWhitelist: make(map[string]bool),
		ipBlacklist: make(map[string]bool),
		apiKeys:     make(map[string]APIKeyInfo),
	}

	// Initialize IP lists
	for _, ip := range config.AllowedIPs {
		manager.ipWhitelist[ip] = true
	}
	for _, ip := range config.BlockedIPs {
		manager.ipBlacklist[ip] = true
	}

	// Initialize rate limiter
	if config.EnableRateLimiting {
		manager.rateLimiter = NewIPRateLimiter(100, time.Minute) // 100 requests per minute default
	}

	// Start API key rotation if enabled
	if config.EnableAPIKeyRotation {
		manager.keyRotator = time.NewTicker(24 * time.Hour) // Daily rotation check
		go manager.rotateAPIKeys()
	}

	utils.GetLogger().Info("Security manager initialized")
	return manager, nil
}

// Close closes the security manager
func (m *Manager) Close() error {
	if m.keyRotator != nil {
		m.keyRotator.Stop()
	}

	return m.auditor.Close()
}

// ValidateRequest validates an incoming HTTP request
func (m *Manager) ValidateRequest(req *http.Request) error {
	if req == nil {
		return errors.NewValidationError("request is nil", nil)
	}

	m.mu.Lock()
	m.requestCount++
	m.mu.Unlock()

	// Check IP restrictions
	clientIP := m.getClientIP(req)
	if err := m.checkIPRestrictions(clientIP); err != nil {
		m.recordBlockedRequest(req, "ip_restriction", err.Error())
		return err
	}

	// Check rate limiting
	if m.config.EnableRateLimiting {
		if !m.rateLimiter.Allow(clientIP) {
			m.recordBlockedRequest(req, "rate_limit", "rate limit exceeded")
			return errors.NewRateLimitError("rate limit exceeded", nil)
		}
	}

	// Validate request
	result := m.validator.ValidateRequest(req)
	if !result.Valid {
		m.mu.Lock()
		m.validationFail++
		m.mu.Unlock()

		// Log validation failure
		failure := ValidationFailure{
			ID:        uuid.New().String(),
			Timestamp: time.Now(),
			Type:      "request_validation",
			Errors:    result.Errors,
			Context: map[string]interface{}{
				"method": req.Method,
				"path":   req.URL.Path,
				"ip":     clientIP,
			},
		}
		m.auditor.LogValidationFailure(failure)

		return errors.NewValidationError(strings.Join(result.Errors, "; "), nil)
	}

	// Log successful access
	m.logAccessAttempt(req, true, "")

	return nil
}

// ValidateResponse validates an outgoing response
func (m *Manager) ValidateResponse(resp interface{}) error {
	result := m.validator.ValidateResponse(resp)
	if !result.Valid {
		return errors.NewValidationError(strings.Join(result.Errors, "; "), nil)
	}
	return nil
}

// SanitizeRequest sanitizes request data
func (m *Manager) SanitizeRequest(req interface{}) (interface{}, error) {
	return m.sanitizer.SanitizeRequest(req)
}

// SanitizeResponse sanitizes response data
func (m *Manager) SanitizeResponse(resp interface{}) (interface{}, error) {
	return m.sanitizer.SanitizeResponse(resp)
}

// ValidateAPIKey validates an API key
func (m *Manager) ValidateAPIKey(key string) error {
	m.keyMu.RLock()
	defer m.keyMu.RUnlock()

	// Hash the key for comparison
	hash := m.hashAPIKey(key)

	// Find key info
	var keyInfo *APIKeyInfo
	for _, info := range m.apiKeys {
		if info.Hash == hash {
			keyInfo = &info
			break
		}
	}

	if keyInfo == nil {
		return errors.NewAuthError("invalid API key", nil)
	}

	if !keyInfo.Active {
		return errors.NewAuthError("API key is inactive", nil)
	}

	// Update last used time
	keyInfo.LastUsed = time.Now()

	return nil
}

// GenerateAPIKey generates a new API key
func (m *Manager) GenerateAPIKey(permissions []string, rateLimit int) (string, error) {
	// Generate random key
	key := uuid.New().String()
	hash := m.hashAPIKey(key)

	info := APIKeyInfo{
		Key:         key,
		Hash:        hash,
		Created:     time.Now(),
		LastUsed:    time.Now(),
		Permissions: permissions,
		RateLimit:   rateLimit,
		Active:      true,
	}

	m.keyMu.Lock()
	m.apiKeys[key] = info
	m.keyMu.Unlock()

	// Log key generation
	m.auditor.LogSecurityEvent(SecurityEvent{
		ID:          uuid.New().String(),
		Type:        "api_key_generated",
		Severity:    "info",
		Timestamp:   time.Now(),
		Source:      "security_manager",
		Description: "New API key generated",
		Data: map[string]interface{}{
			"key_hash":    hash,
			"permissions": permissions,
			"rate_limit":  rateLimit,
		},
	})

	return key, nil
}

// RevokeAPIKey revokes an API key
func (m *Manager) RevokeAPIKey(key string) error {
	m.keyMu.Lock()
	defer m.keyMu.Unlock()

	info, exists := m.apiKeys[key]
	if !exists {
		return errors.NewNotFoundError("API key not found", nil)
	}

	info.Active = false
	m.apiKeys[key] = info

	// Log revocation
	m.auditor.LogSecurityEvent(SecurityEvent{
		ID:          uuid.New().String(),
		Type:        "api_key_revoked",
		Severity:    "warning",
		Timestamp:   time.Now(),
		Source:      "security_manager",
		Description: "API key revoked",
		Data: map[string]interface{}{
			"key_hash": info.Hash,
		},
	})

	return nil
}

// GetMetrics returns security metrics
func (m *Manager) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"total_requests":      m.requestCount,
		"blocked_requests":    m.blockedCount,
		"validation_failures": m.validationFail,
		"active_api_keys":     m.getActiveKeyCount(),
		"security_level":      m.config.Level,
	}
}

// Helper methods

func (m *Manager) getClientIP(req *http.Request) string {
	// Check trusted proxy headers
	if len(m.config.TrustedProxies) > 0 {
		if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
			ips := strings.Split(xff, ",")
			return strings.TrimSpace(ips[0])
		}
		if xri := req.Header.Get("X-Real-IP"); xri != "" {
			return xri
		}
	}

	// Fall back to remote address
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}
	return ip
}

func (m *Manager) checkIPRestrictions(ip string) error {
	m.ipMu.RLock()
	defer m.ipMu.RUnlock()

	// Check blacklist first
	if m.ipBlacklist[ip] {
		return errors.NewForbiddenError("IP address is blocked", nil)
	}

	// Check whitelist if enabled
	if m.config.EnableIPWhitelist && len(m.ipWhitelist) > 0 {
		if !m.ipWhitelist[ip] {
			return errors.NewForbiddenError("IP address not in whitelist", nil)
		}
	}

	return nil
}

func (m *Manager) recordBlockedRequest(req *http.Request, reason, details string) {
	m.mu.Lock()
	m.blockedCount++
	m.mu.Unlock()

	m.logAccessAttempt(req, false, fmt.Sprintf("%s: %s", reason, details))
}

func (m *Manager) logAccessAttempt(req *http.Request, success bool, reason string) {
	attempt := AccessAttempt{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		IP:        m.getClientIP(req),
		UserAgent: req.UserAgent(),
		Method:    req.Method,
		Path:      req.URL.Path,
		Success:   success,
		Reason:    reason,
	}

	// Add API key hash if present
	if apiKey := req.Header.Get(m.config.APIKeyHeader); apiKey != "" {
		attempt.APIKeyHash = m.hashAPIKey(apiKey)
	}

	m.auditor.LogAccessAttempt(attempt)
}

func (m *Manager) hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

func (m *Manager) getActiveKeyCount() int {
	m.keyMu.RLock()
	defer m.keyMu.RUnlock()

	count := 0
	for _, info := range m.apiKeys {
		if info.Active {
			count++
		}
	}
	return count
}

func (m *Manager) rotateAPIKeys() {
	for range m.keyRotator.C {
		m.keyMu.Lock()

		// Check for keys that haven't been used in 30 days
		cutoff := time.Now().AddDate(0, 0, -30)
		for key, info := range m.apiKeys {
			if info.Active && info.LastUsed.Before(cutoff) {
				info.Active = false
				m.apiKeys[key] = info

				utils.GetLogger().Warnf("Auto-rotated unused API key: %s", info.Hash)
			}
		}

		m.keyMu.Unlock()
	}
}

// AddIPToWhitelist adds an IP to the whitelist
func (m *Manager) AddIPToWhitelist(ip string) {
	m.ipMu.Lock()
	defer m.ipMu.Unlock()
	m.ipWhitelist[ip] = true
}

// RemoveIPFromWhitelist removes an IP from the whitelist
func (m *Manager) RemoveIPFromWhitelist(ip string) {
	m.ipMu.Lock()
	defer m.ipMu.Unlock()
	delete(m.ipWhitelist, ip)
}

// AddIPToBlacklist adds an IP to the blacklist
func (m *Manager) AddIPToBlacklist(ip string) {
	m.ipMu.Lock()
	defer m.ipMu.Unlock()
	m.ipBlacklist[ip] = true
}

// RemoveIPFromBlacklist removes an IP from the blacklist
func (m *Manager) RemoveIPFromBlacklist(ip string) {
	m.ipMu.Lock()
	defer m.ipMu.Unlock()
	delete(m.ipBlacklist, ip)
}
