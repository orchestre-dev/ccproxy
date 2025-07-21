package security

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/orchestre-dev/ccproxy/internal/errors"
	"github.com/orchestre-dev/ccproxy/internal/utils"
)

// RequestValidator validates incoming requests
type RequestValidator struct {
	config           *SecurityConfig
	compiledPatterns []*regexp.Regexp
	sensitiveRegexps []*regexp.Regexp
}

// NewRequestValidator creates a new request validator
func NewRequestValidator(config *SecurityConfig) (*RequestValidator, error) {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	validator := &RequestValidator{
		config:           config,
		compiledPatterns: make([]*regexp.Regexp, 0),
		sensitiveRegexps: make([]*regexp.Regexp, 0),
	}

	// Compile blocked patterns
	for _, pattern := range config.BlockedPatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid blocked pattern %s: %w", pattern, err)
		}
		validator.compiledPatterns = append(validator.compiledPatterns, re)
	}

	// Compile sensitive patterns
	for _, pattern := range config.SensitivePatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid sensitive pattern %s: %w", pattern, err)
		}
		validator.sensitiveRegexps = append(validator.sensitiveRegexps, re)
	}

	return validator, nil
}

// Validate validates generic data
func (v *RequestValidator) Validate(data interface{}) ValidationResult {
	result := ValidationResult{
		Valid:  true,
		Score:  1.0,
		Errors: []string{},
		Warnings: []string{},
	}

	// Convert to JSON for pattern matching
	jsonData, err := json.Marshal(data)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, "failed to marshal data for validation")
		result.Score = 0
		return result
	}

	dataStr := string(jsonData)

	// Check for blocked patterns
	if v.config.EnableContentFilter {
		for _, pattern := range v.compiledPatterns {
			if pattern.MatchString(dataStr) {
				result.Valid = false
				result.Errors = append(result.Errors, fmt.Sprintf("blocked pattern detected: %s", pattern.String()))
				result.Score *= 0.5
			}
		}
	}

	// Check for sensitive data
	for _, pattern := range v.sensitiveRegexps {
		if pattern.MatchString(dataStr) {
			result.Warnings = append(result.Warnings, "potential sensitive data detected")
			result.Score *= 0.8
		}
	}

	return result
}

// ValidateRequest validates HTTP requests
func (v *RequestValidator) ValidateRequest(req *http.Request) ValidationResult {
	result := ValidationResult{
		Valid:    true,
		Score:    1.0,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Check request size
	if req.ContentLength > v.config.MaxRequestSize {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("request size %d exceeds limit %d", req.ContentLength, v.config.MaxRequestSize))
		result.Score = 0
		return result
	}

	// Validate headers
	if err := v.validateHeaders(req); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
		result.Score *= 0.5
	}

	// Validate URL
	if err := v.validateURL(req.URL); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
		result.Score *= 0.5
	}

	// Check authentication if required
	if v.config.RequireAuth {
		if err := v.validateAuth(req); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, err.Error())
			result.Score = 0
		}
	}

	// Check for SQL injection patterns
	if v.detectSQLInjection(req) {
		result.Valid = false
		result.Errors = append(result.Errors, "potential SQL injection detected")
		result.Score = 0
	}

	// Check for XSS patterns
	if v.detectXSS(req) {
		result.Valid = false
		result.Errors = append(result.Errors, "potential XSS attack detected")
		result.Score = 0
	}

	// Check for path traversal
	if v.detectPathTraversal(req) {
		result.Valid = false
		result.Errors = append(result.Errors, "potential path traversal detected")
		result.Score = 0
	}

	return result
}

// ValidateResponse validates responses
func (v *RequestValidator) ValidateResponse(resp interface{}) ValidationResult {
	result := ValidationResult{
		Valid:    true,
		Score:    1.0,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Check response size
	if respData, err := json.Marshal(resp); err == nil {
		if int64(len(respData)) > v.config.MaxRequestSize {
			result.Warnings = append(result.Warnings, "response size exceeds recommended limit")
			result.Score *= 0.9
		}
	}

	// Validate content
	contentResult := v.Validate(resp)
	if !contentResult.Valid {
		result.Valid = false
		result.Errors = append(result.Errors, contentResult.Errors...)
		result.Score = contentResult.Score
	}
	result.Warnings = append(result.Warnings, contentResult.Warnings...)

	return result
}

// validateHeaders validates HTTP headers
func (v *RequestValidator) validateHeaders(req *http.Request) error {
	// Check for required headers
	if v.config.RequireAuth {
		authHeader := req.Header.Get("Authorization")
		apiKeyHeader := req.Header.Get(v.config.APIKeyHeader)
		
		if authHeader == "" && apiKeyHeader == "" {
			return errors.NewAuthError("missing authentication headers", nil)
		}
	}

	// Check for suspicious headers
	suspiciousHeaders := []string{
		"X-Forwarded-Host",
		"X-Original-URL",
		"X-Rewrite-URL",
	}

	for _, header := range suspiciousHeaders {
		if val := req.Header.Get(header); val != "" {
			// Log suspicious header but don't block
			utils.GetLogger().Warnf("Suspicious header detected: %s=%s", header, val)
		}
	}

	return nil
}

// validateURL validates the request URL
func (v *RequestValidator) validateURL(u *url.URL) error {
	// Check for suspicious path segments
	pathSegments := strings.Split(u.Path, "/")
	for _, segment := range pathSegments {
		if segment == ".." || segment == "." {
			return fmt.Errorf("invalid path segment: %s", segment)
		}
	}

	// Check query parameters
	for key, values := range u.Query() {
		for _, value := range values {
			// Check length
			if len(value) > 1000 {
				return fmt.Errorf("query parameter %s too long", key)
			}
			
			// Check for encoded characters that might be malicious
			decoded, err := url.QueryUnescape(value)
			if err != nil {
				return fmt.Errorf("invalid query parameter encoding: %s", key)
			}
			
			if strings.Contains(decoded, "<script") || strings.Contains(decoded, "javascript:") {
				return fmt.Errorf("potential XSS in query parameter: %s", key)
			}
		}
	}

	return nil
}

// validateAuth validates authentication
func (v *RequestValidator) validateAuth(req *http.Request) error {
	authHeader := req.Header.Get("Authorization")
	apiKeyHeader := req.Header.Get(v.config.APIKeyHeader)

	// Check Authorization header
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 {
			return errors.NewAuthError("invalid authorization header format", nil)
		}

		authType := strings.ToLower(parts[0])
		allowed := false
		for _, method := range v.config.AllowedAuthMethods {
			if authType == method {
				allowed = true
				break
			}
		}

		if !allowed {
			return errors.NewAuthError(fmt.Sprintf("authentication method %s not allowed", authType), nil)
		}

		// Validate token format
		token := parts[1]
		if len(token) < 10 {
			return errors.NewAuthError("invalid token format", nil)
		}
	}

	// Check API key header
	if apiKeyHeader != "" {
		if len(apiKeyHeader) < 10 {
			return errors.NewAuthError("invalid API key format", nil)
		}
	}

	return nil
}

// detectSQLInjection detects potential SQL injection
func (v *RequestValidator) detectSQLInjection(req *http.Request) bool {
	patterns := []string{
		`(?i)(union.*select)`,
		`(?i)(select.*from)`,
		`(?i)(insert.*into)`,
		`(?i)(delete.*from)`,
		`(?i)(drop.*table)`,
		`(?i)(or\s+1\s*=\s*1)`,
		`(?i)(;\s*--)`,
		`(?i)(\'\s*or\s*\')`,
	}

	// Check URL
	urlStr := req.URL.String()
	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, urlStr); matched {
			return true
		}
	}

	// Check headers
	for _, values := range req.Header {
		for _, value := range values {
			for _, pattern := range patterns {
				if matched, _ := regexp.MatchString(pattern, value); matched {
					return true
				}
			}
		}
	}

	return false
}

// detectXSS detects potential XSS attacks
func (v *RequestValidator) detectXSS(req *http.Request) bool {
	patterns := []string{
		`<script[^>]*>`,
		`javascript:`,
		`on\w+\s*=`,
		`<iframe[^>]*>`,
		`<object[^>]*>`,
		`<embed[^>]*>`,
		`<link[^>]*>`,
		`vbscript:`,
	}

	// Check URL
	urlStr := req.URL.String()
	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, urlStr); matched {
			return true
		}
	}

	return false
}

// detectPathTraversal detects path traversal attempts
func (v *RequestValidator) detectPathTraversal(req *http.Request) bool {
	patterns := []string{
		`\.\.\/`,
		`\.\.\\`,
		`%2e%2e%2f`,
		`%2e%2e%5c`,
		`%%32%65%%32%65%%32%66`,
		`\.\.%c0%af`,
		`\.\.%c1%9c`,
	}

	path := req.URL.Path
	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, path); matched {
			return true
		}
	}

	return false
}