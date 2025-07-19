package security

import (
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strings"

	"github.com/orchestre-dev/ccproxy/internal/utils"
)

// DataSanitizer sanitizes data to prevent security issues
type DataSanitizer struct {
	config           *SecurityConfig
	sensitiveRegexps []*regexp.Regexp
	redactPatterns   map[string]*regexp.Regexp
}

// NewDataSanitizer creates a new data sanitizer
func NewDataSanitizer(config *SecurityConfig) (*DataSanitizer, error) {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	sanitizer := &DataSanitizer{
		config:           config,
		sensitiveRegexps: make([]*regexp.Regexp, 0),
		redactPatterns:   make(map[string]*regexp.Regexp),
	}

	// Compile sensitive patterns
	for _, pattern := range config.SensitivePatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid sensitive pattern %s: %w", pattern, err)
		}
		sanitizer.sensitiveRegexps = append(sanitizer.sensitiveRegexps, re)
	}

	// Default redaction patterns
	sanitizer.redactPatterns["api_key"] = regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[:=]\s*["']?([^"'\s]+)["']?`)
	sanitizer.redactPatterns["token"] = regexp.MustCompile(`(?i)(token|bearer)\s*[:=]\s*["']?([^"'\s]+)["']?`)
	sanitizer.redactPatterns["password"] = regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[:=]\s*["']?([^"'\s]+)["']?`)
	sanitizer.redactPatterns["secret"] = regexp.MustCompile(`(?i)(secret)\s*[:=]\s*["']?([^"'\s]+)["']?`)
	sanitizer.redactPatterns["credit_card"] = regexp.MustCompile(`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`)
	sanitizer.redactPatterns["ssn"] = regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
	sanitizer.redactPatterns["email"] = regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)

	return sanitizer, nil
}

// SanitizeRequest sanitizes incoming requests
func (s *DataSanitizer) SanitizeRequest(req interface{}) (interface{}, error) {
	// Convert to JSON for processing
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Parse as generic map
	var genericData map[string]interface{}
	if err := json.Unmarshal(data, &genericData); err != nil {
		// If not a map, try array
		var arrayData []interface{}
		if err := json.Unmarshal(data, &arrayData); err != nil {
			// Return as-is if can't parse
			return req, nil
		}
		return s.sanitizeArray(arrayData), nil
	}

	// Sanitize the data
	sanitized := s.sanitizeMap(genericData)

	// Convert back to original type
	sanitizedJSON, err := json.Marshal(sanitized)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sanitized data: %w", err)
	}

	// Create new instance of same type
	result := req
	if err := json.Unmarshal(sanitizedJSON, &result); err != nil {
		// Return sanitized map if can't convert back
		return sanitized, nil
	}

	return result, nil
}

// SanitizeResponse sanitizes outgoing responses
func (s *DataSanitizer) SanitizeResponse(resp interface{}) (interface{}, error) {
	// For responses, we're less strict but still remove sensitive data
	return s.SanitizeRequest(resp)
}

// SanitizeString sanitizes a string value
func (s *DataSanitizer) SanitizeString(str string) string {
	// HTML escape
	sanitized := html.EscapeString(str)

	// URL decode to check for encoded attacks
	decoded, _ := url.QueryUnescape(sanitized)
	
	// Remove null bytes
	sanitized = strings.ReplaceAll(decoded, "\x00", "")
	
	// Remove control characters
	sanitized = regexp.MustCompile(`[\x00-\x1F\x7F]`).ReplaceAllString(sanitized, "")

	// Redact sensitive information if logging is disabled
	if !s.config.LogSensitiveData {
		for name, pattern := range s.redactPatterns {
			if pattern.MatchString(sanitized) {
				replacement := fmt.Sprintf("[REDACTED_%s]", strings.ToUpper(name))
				sanitized = pattern.ReplaceAllString(sanitized, replacement)
			}
		}
	}

	return sanitized
}

// RemoveSensitiveData removes sensitive data from a map
func (s *DataSanitizer) RemoveSensitiveData(data map[string]interface{}) map[string]interface{} {
	return s.sanitizeMap(data)
}

// sanitizeMap recursively sanitizes a map
func (s *DataSanitizer) sanitizeMap(data map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})

	for key, value := range data {
		// Check if key indicates sensitive data
		lowerKey := strings.ToLower(key)
		if s.isSensitiveKey(lowerKey) && !s.config.LogSensitiveData {
			sanitized[key] = "[REDACTED]"
			continue
		}

		// Recursively sanitize value
		switch v := value.(type) {
		case string:
			sanitized[key] = s.SanitizeString(v)
		case map[string]interface{}:
			sanitized[key] = s.sanitizeMap(v)
		case []interface{}:
			sanitized[key] = s.sanitizeArray(v)
		default:
			sanitized[key] = value
		}
	}

	return sanitized
}

// sanitizeArray recursively sanitizes an array
func (s *DataSanitizer) sanitizeArray(data []interface{}) []interface{} {
	sanitized := make([]interface{}, len(data))

	for i, value := range data {
		switch v := value.(type) {
		case string:
			sanitized[i] = s.SanitizeString(v)
		case map[string]interface{}:
			sanitized[i] = s.sanitizeMap(v)
		case []interface{}:
			sanitized[i] = s.sanitizeArray(v)
		default:
			sanitized[i] = value
		}
	}

	return sanitized
}

// isSensitiveKey checks if a key name indicates sensitive data
func (s *DataSanitizer) isSensitiveKey(key string) bool {
	sensitiveKeys := []string{
		"password", "passwd", "pwd",
		"secret", "api_key", "apikey", "api-key",
		"token", "auth", "authorization",
		"credit_card", "creditcard", "cc_number",
		"ssn", "social_security", "social-security",
		"private_key", "private-key", "privatekey",
		"access_key", "access-key", "accesskey",
		"refresh_token", "refresh-token",
		"session", "session_id", "session-id",
		"cookie", "csrf", "xsrf",
	}

	for _, sensitive := range sensitiveKeys {
		if strings.Contains(key, sensitive) {
			return true
		}
	}

	return false
}

// RedactSecrets redacts secrets from log messages
func (s *DataSanitizer) RedactSecrets(message string) string {
	redacted := message

	// Apply all redaction patterns
	for name, pattern := range s.redactPatterns {
		if pattern.MatchString(redacted) {
			replacement := fmt.Sprintf("[REDACTED_%s]", strings.ToUpper(name))
			redacted = pattern.ReplaceAllString(redacted, replacement)
			utils.GetLogger().Debugf("Redacted %s from log message", name)
		}
	}

	return redacted
}

// MaskValue masks a sensitive value for display
func MaskValue(value string, showChars int) string {
	if len(value) <= showChars {
		return strings.Repeat("*", len(value))
	}

	if showChars <= 0 {
		return strings.Repeat("*", len(value))
	}

	// Show first few characters and mask the rest
	visible := value[:showChars]
	masked := strings.Repeat("*", len(value)-showChars)
	return visible + masked
}

// MaskEmail masks an email address
func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return MaskValue(email, 3)
	}

	// Mask username part
	username := parts[0]
	domain := parts[1]

	if len(username) <= 3 {
		username = strings.Repeat("*", len(username))
	} else {
		username = username[:2] + strings.Repeat("*", len(username)-2)
	}

	return username + "@" + domain
}

// MaskAPIKey masks an API key
func MaskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return strings.Repeat("*", len(apiKey))
	}

	// Show first 4 and last 4 characters
	prefix := apiKey[:4]
	suffix := apiKey[len(apiKey)-4:]
	masked := strings.Repeat("*", len(apiKey)-8)
	
	return prefix + masked + suffix
}