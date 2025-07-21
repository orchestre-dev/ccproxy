package security

import (
	"net/http"
	"time"
)

// SecurityLevel represents the level of security enforcement
type SecurityLevel string

const (
	SecurityLevelNone     SecurityLevel = "none"
	SecurityLevelBasic    SecurityLevel = "basic"
	SecurityLevelStrict   SecurityLevel = "strict"
	SecurityLevelParanoid SecurityLevel = "paranoid"
)

// ValidationResult represents the result of a security validation
type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
	Score    float64  `json:"score"` // 0.0 to 1.0
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	Level                SecurityLevel `json:"level"`
	EnableRequestSigning bool         `json:"enable_request_signing"`
	EnableTLS           bool         `json:"enable_tls"`
	TLSMinVersion       string       `json:"tls_min_version"`
	EnableRateLimiting  bool         `json:"enable_rate_limiting"`
	EnableIPWhitelist   bool         `json:"enable_ip_whitelist"`
	EnableAPIKeyRotation bool        `json:"enable_api_key_rotation"`
	
	// Request validation
	MaxRequestSize      int64         `json:"max_request_size"`
	MaxTokenLength      int           `json:"max_token_length"`
	MaxPromptLength     int           `json:"max_prompt_length"`
	RequestTimeout      time.Duration `json:"request_timeout"`
	
	// Content filtering
	EnableContentFilter bool     `json:"enable_content_filter"`
	BlockedPatterns    []string `json:"blocked_patterns"`
	SensitivePatterns  []string `json:"sensitive_patterns"`
	
	// Authentication
	RequireAuth        bool     `json:"require_auth"`
	AllowedAuthMethods []string `json:"allowed_auth_methods"`
	APIKeyHeader       string   `json:"api_key_header"`
	
	// IP restrictions
	AllowedIPs     []string `json:"allowed_ips"`
	BlockedIPs     []string `json:"blocked_ips"`
	TrustedProxies []string `json:"trusted_proxies"`
	
	// Audit and logging
	EnableAuditLog      bool   `json:"enable_audit_log"`
	LogSensitiveData    bool   `json:"log_sensitive_data"`
	AuditLogPath        string `json:"audit_log_path"`
	RetentionDays       int    `json:"retention_days"`
}

// Validator interface for security validation
type Validator interface {
	Validate(data interface{}) ValidationResult
	ValidateRequest(req *http.Request) ValidationResult
	ValidateResponse(resp interface{}) ValidationResult
}

// Sanitizer interface for data sanitization
type Sanitizer interface {
	SanitizeRequest(req interface{}) (interface{}, error)
	SanitizeResponse(resp interface{}) (interface{}, error)
	SanitizeString(s string) string
	RemoveSensitiveData(data map[string]interface{}) map[string]interface{}
}

// Auditor interface for security auditing
type Auditor interface {
	LogSecurityEvent(event SecurityEvent)
	LogAccessAttempt(attempt AccessAttempt)
	LogValidationFailure(failure ValidationFailure)
	GetAuditTrail(filter AuditFilter) []AuditEntry
}

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Timestamp   time.Time              `json:"timestamp"`
	Source      string                 `json:"source"`
	Description string                 `json:"description"`
	Data        map[string]interface{} `json:"data"`
}

// AccessAttempt represents an access attempt
type AccessAttempt struct {
	ID         string    `json:"id"`
	Timestamp  time.Time `json:"timestamp"`
	IP         string    `json:"ip"`
	UserAgent  string    `json:"user_agent"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	Success    bool      `json:"success"`
	Reason     string    `json:"reason,omitempty"`
	APIKey     string    `json:"-"` // Don't log the actual key
	APIKeyHash string    `json:"api_key_hash,omitempty"`
}

// ValidationFailure represents a validation failure
type ValidationFailure struct {
	ID        string           `json:"id"`
	Timestamp time.Time        `json:"timestamp"`
	Type      string           `json:"type"`
	Field     string           `json:"field,omitempty"`
	Value     interface{}      `json:"-"` // Don't log potentially sensitive values
	Errors    []string         `json:"errors"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// AuditEntry represents an audit log entry
type AuditEntry struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Type      string                 `json:"type"`
	Actor     string                 `json:"actor"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Result    string                 `json:"result"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// AuditFilter for querying audit logs
type AuditFilter struct {
	StartTime time.Time
	EndTime   time.Time
	Type      string
	Actor     string
	Action    string
	Result    string
	Limit     int
	Offset    int
}

// RateLimitInfo represents rate limit information
type RateLimitInfo struct {
	Key       string
	Limit     int
	Window    time.Duration
	Used      int
	Reset     time.Time
	Remaining int
}

// IPInfo represents IP address information
type IPInfo struct {
	IP          string
	Country     string
	Region      string
	City        string
	ISP         string
	ThreatLevel int
	IsProxy     bool
	IsVPN       bool
	IsTor       bool
}

// DefaultSecurityConfig returns default security configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		Level:                SecurityLevelBasic,
		EnableRequestSigning: false,
		EnableTLS:           true,
		TLSMinVersion:       "1.2",
		EnableRateLimiting:  true,
		EnableIPWhitelist:   false,
		EnableAPIKeyRotation: false,
		
		MaxRequestSize:  10 * 1024 * 1024, // 10MB
		MaxTokenLength:  100000,
		MaxPromptLength: 50000,
		RequestTimeout:  5 * time.Minute,
		
		EnableContentFilter: true,
		BlockedPatterns:    []string{},
		SensitivePatterns:  []string{},
		
		RequireAuth:        true,
		AllowedAuthMethods: []string{"api_key", "bearer"},
		APIKeyHeader:       "X-API-Key",
		
		AllowedIPs:     []string{},
		BlockedIPs:     []string{},
		TrustedProxies: []string{},
		
		EnableAuditLog:   true,
		LogSensitiveData: false,
		AuditLogPath:     "./audit.log",
		RetentionDays:    30,
	}
}

// StrictSecurityConfig returns strict security configuration
func StrictSecurityConfig() *SecurityConfig {
	config := DefaultSecurityConfig()
	config.Level = SecurityLevelStrict
	config.EnableRequestSigning = true
	config.EnableAPIKeyRotation = true
	config.MaxRequestSize = 5 * 1024 * 1024 // 5MB
	config.MaxTokenLength = 50000
	config.MaxPromptLength = 25000
	config.RequestTimeout = 2 * time.Minute
	config.EnableIPWhitelist = true
	return config
}

// ParanoidSecurityConfig returns paranoid security configuration
func ParanoidSecurityConfig() *SecurityConfig {
	config := StrictSecurityConfig()
	config.Level = SecurityLevelParanoid
	config.MaxRequestSize = 1 * 1024 * 1024 // 1MB
	config.MaxTokenLength = 10000
	config.MaxPromptLength = 5000
	config.RequestTimeout = 1 * time.Minute
	config.LogSensitiveData = false
	config.RetentionDays = 90
	return config
}