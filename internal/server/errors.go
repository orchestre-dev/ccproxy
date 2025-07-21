package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorType represents the type of error
type ErrorType string

const (
	ErrorTypeInvalidRequest ErrorType = "invalid_request"
	ErrorTypeNotFound       ErrorType = "not_found"
	ErrorTypeAuthentication ErrorType = "authentication_error"
	ErrorTypePermission     ErrorType = "permission_error"
	ErrorTypeRateLimit      ErrorType = "rate_limit_error"
	ErrorTypeProviderError  ErrorType = "provider_error"
	ErrorTypeServerError    ErrorType = "server_error"
	ErrorTypeNotImplemented ErrorType = "not_implemented"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error details
type ErrorDetail struct {
	Message string    `json:"message"`
	Type    ErrorType `json:"type"`
	Code    string    `json:"code,omitempty"`
}

// RespondWithError sends a standardized error response
func RespondWithError(c *gin.Context, statusCode int, errorType ErrorType, message string) {
	c.JSON(statusCode, ErrorResponse{
		Error: ErrorDetail{
			Type:    errorType,
			Message: message,
		},
	})
}

// RespondWithErrorCode sends a standardized error response with error code
func RespondWithErrorCode(c *gin.Context, statusCode int, errorType ErrorType, message string, code string) {
	c.JSON(statusCode, ErrorResponse{
		Error: ErrorDetail{
			Type:    errorType,
			Message: message,
			Code:    code,
		},
	})
}

// Common error responses

// BadRequest sends a 400 Bad Request error
func BadRequest(c *gin.Context, message string) {
	RespondWithError(c, http.StatusBadRequest, ErrorTypeInvalidRequest, message)
}

// Unauthorized sends a 401 Unauthorized error
func Unauthorized(c *gin.Context, message string) {
	RespondWithError(c, http.StatusUnauthorized, ErrorTypeAuthentication, message)
}

// Forbidden sends a 403 Forbidden error
func Forbidden(c *gin.Context, message string) {
	RespondWithError(c, http.StatusForbidden, ErrorTypePermission, message)
}

// NotFound sends a 404 Not Found error
func NotFound(c *gin.Context, message string) {
	RespondWithError(c, http.StatusNotFound, ErrorTypeNotFound, message)
}

// InternalServerError sends a 500 Internal Server Error
func InternalServerError(c *gin.Context, message string) {
	RespondWithError(c, http.StatusInternalServerError, ErrorTypeServerError, message)
}

// NotImplemented sends a 501 Not Implemented error
func NotImplemented(c *gin.Context, message string) {
	RespondWithError(c, http.StatusNotImplemented, ErrorTypeNotImplemented, message)
}

// ProviderError sends a 502 Bad Gateway error for provider issues
func ProviderError(c *gin.Context, message string) {
	RespondWithError(c, http.StatusBadGateway, ErrorTypeProviderError, message)
}

// RateLimitError sends a 429 Too Many Requests error
func RateLimitError(c *gin.Context, message string) {
	RespondWithError(c, http.StatusTooManyRequests, ErrorTypeRateLimit, message)
}
