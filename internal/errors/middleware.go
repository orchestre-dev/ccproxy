package errors

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/musistudio/ccproxy/internal/utils"
	"github.com/sirupsen/logrus"
)

// ErrorHandlerMiddleware returns a Gin middleware for handling errors
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic
				logger := utils.GetLogger()
				logger.WithFields(logrus.Fields{
					"panic":  r,
					"stack":  string(debug.Stack()),
					"method": c.Request.Method,
					"path":   c.Request.URL.Path,
				}).Error("Panic recovered")
				
				// Create internal error
				err := New(ErrorTypeInternal, "Internal server error")
				if reqID := c.GetString("request_id"); reqID != "" {
					err.WithRequestID(reqID)
				}
				
				// Write error response
				err.WriteHTTPResponse(c.Writer)
				c.Abort()
			}
		}()
		
		// Process request
		c.Next()
		
		// Check for errors set in context
		if len(c.Errors) > 0 {
			// Get the last error
			err := c.Errors.Last()
			handleGinError(c, err)
		}
	}
}

// HandleError handles errors in Gin context
func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	
	logger := utils.GetLogger()
	
	// Check if it's already a CCProxyError
	var ccErr *CCProxyError
	if e, ok := err.(*CCProxyError); ok {
		ccErr = e
	} else {
		// Convert to CCProxyError based on error type
		ccErr = convertToCCProxyError(err)
	}
	
	// Add request ID if available
	if reqID := c.GetString("request_id"); reqID != "" && ccErr.RequestID == "" {
		ccErr.WithRequestID(reqID)
	}
	
	// Log the error
	fields := logrus.Fields{
		"error_type": ccErr.Type,
		"status":     ccErr.StatusCode,
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"retryable":  ccErr.Retryable,
	}
	
	if ccErr.Provider != "" {
		fields["provider"] = ccErr.Provider
	}
	if ccErr.RequestID != "" {
		fields["request_id"] = ccErr.RequestID
	}
	if ccErr.Details != nil {
		fields["details"] = ccErr.Details
	}
	
	// Log at appropriate level
	if ccErr.StatusCode >= 500 {
		logger.WithFields(fields).Error(ccErr.Message)
	} else if ccErr.StatusCode >= 400 {
		logger.WithFields(fields).Warn(ccErr.Message)
	} else {
		logger.WithFields(fields).Info(ccErr.Message)
	}
	
	// Write response
	ccErr.WriteHTTPResponse(c.Writer)
	c.Abort()
}

// HandleErrorWithStatus handles errors with a specific status code
func HandleErrorWithStatus(c *gin.Context, statusCode int, err error) {
	if err == nil {
		return
	}
	
	var ccErr *CCProxyError
	if e, ok := err.(*CCProxyError); ok {
		ccErr = e
		ccErr.StatusCode = statusCode
	} else {
		ccErr = &CCProxyError{
			Type:       getErrorTypeFromStatusCode(statusCode),
			Message:    err.Error(),
			StatusCode: statusCode,
			Retryable:  isRetryable(getErrorTypeFromStatusCode(statusCode)),
		}
	}
	
	HandleError(c, ccErr)
}

// AbortWithError aborts the request with an error
func AbortWithError(c *gin.Context, err error) {
	HandleError(c, err)
}

// AbortWithCCProxyError aborts with a CCProxyError
func AbortWithCCProxyError(c *gin.Context, errorType ErrorType, message string) {
	err := New(errorType, message)
	HandleError(c, err)
}

// handleGinError handles errors from Gin's error list
func handleGinError(c *gin.Context, ginErr *gin.Error) {
	err := ginErr.Err
	
	// Check if it's already handled
	if c.Writer.Written() {
		return
	}
	
	// Handle the error
	HandleError(c, err)
}

// convertToCCProxyError converts various error types to CCProxyError
func convertToCCProxyError(err error) *CCProxyError {
	errStr := err.Error()
	errLower := strings.ToLower(errStr)
	
	// Check for specific error patterns
	switch {
	case strings.Contains(errLower, "unauthorized"):
		return New(ErrorTypeUnauthorized, err.Error())
		
	case strings.Contains(errLower, "forbidden"):
		return New(ErrorTypeForbidden, err.Error())
		
	case strings.Contains(errLower, "not found"):
		return New(ErrorTypeNotFound, err.Error())
		
	case strings.Contains(errLower, "bad request"):
		return New(ErrorTypeBadRequest, err.Error())
		
	case strings.Contains(errLower, "rate limit"):
		return New(ErrorTypeRateLimitError, err.Error())
		
	case strings.Contains(errLower, "timeout"):
		return New(ErrorTypeGatewayTimeout, err.Error())
		
	case strings.Contains(errLower, "connection refused"),
		strings.Contains(errLower, "connection reset"),
		strings.Contains(errLower, "no such host"):
		return New(ErrorTypeBadGateway, err.Error())
		
	case strings.Contains(errLower, "service unavailable"):
		return New(ErrorTypeServiceUnavailable, err.Error())
		
	case strings.Contains(errLower, "not implemented"):
		return New(ErrorTypeNotImplemented, err.Error())
		
	case strings.Contains(errLower, "validation"):
		return New(ErrorTypeValidationError, err.Error())
		
	default:
		return New(ErrorTypeInternal, err.Error())
	}
}

// ExtractProviderError extracts error information from provider response
func ExtractProviderError(resp *http.Response, body []byte, provider string) error {
	if resp.StatusCode < 400 {
		return nil
	}
	
	return FromProviderResponse(resp.StatusCode, body, provider)
}

// WrapProviderError wraps a provider error with additional context
func WrapProviderError(err error, provider string) error {
	if err == nil {
		return nil
	}
	
	// If it's already a CCProxyError with provider info, return as-is
	if ccErr, ok := err.(*CCProxyError); ok && ccErr.Provider != "" {
		return err
	}
	
	// Check for specific provider error patterns
	errStr := err.Error()
	var errorType ErrorType
	
	switch {
	case strings.Contains(errStr, "context deadline exceeded"):
		errorType = ErrorTypeGatewayTimeout
	case strings.Contains(errStr, "connection refused"):
		errorType = ErrorTypeBadGateway
	case strings.Contains(errStr, "rate limit"):
		errorType = ErrorTypeRateLimitError
	default:
		errorType = ErrorTypeProviderError
	}
	
	return Wrap(err, errorType, fmt.Sprintf("Provider %s error", provider)).
		WithProvider(provider)
}

// ErrorResponse represents the API error response format
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error details
type ErrorDetail struct {
	Type    string                 `json:"type"`
	Message string                 `json:"message"`
	Code    string                 `json:"code,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(errorType ErrorType, message string) *ErrorResponse {
	return &ErrorResponse{
		Error: ErrorDetail{
			Type:    string(errorType),
			Message: message,
		},
	}
}

// WithCode adds a code to the error response
func (e *ErrorResponse) WithCode(code string) *ErrorResponse {
	e.Error.Code = code
	return e
}

// WithDetails adds details to the error response
func (e *ErrorResponse) WithDetails(details map[string]interface{}) *ErrorResponse {
	e.Error.Details = details
	return e
}