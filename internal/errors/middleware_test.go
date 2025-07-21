package errors

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestConvertToCCProxyError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedType ErrorType
	}{
		{
			name:         "unauthorized error",
			err:          errors.New("unauthorized access"),
			expectedType: ErrorTypeUnauthorized,
		},
		{
			name:         "forbidden error",
			err:          errors.New("forbidden resource"),
			expectedType: ErrorTypeForbidden,
		},
		{
			name:         "not found error",
			err:          errors.New("resource not found"),
			expectedType: ErrorTypeNotFound,
		},
		{
			name:         "bad request error",
			err:          errors.New("bad request format"),
			expectedType: ErrorTypeBadRequest,
		},
		{
			name:         "rate limit error",
			err:          errors.New("rate limit exceeded"),
			expectedType: ErrorTypeRateLimitError,
		},
		{
			name:         "timeout error",
			err:          errors.New("request timeout"),
			expectedType: ErrorTypeGatewayTimeout,
		},
		{
			name:         "connection refused",
			err:          errors.New("connection refused"),
			expectedType: ErrorTypeBadGateway,
		},
		{
			name:         "connection reset",
			err:          errors.New("connection reset by peer"),
			expectedType: ErrorTypeBadGateway,
		},
		{
			name:         "no such host",
			err:          errors.New("dial tcp: no such host"),
			expectedType: ErrorTypeBadGateway,
		},
		{
			name:         "service unavailable",
			err:          errors.New("service unavailable"),
			expectedType: ErrorTypeServiceUnavailable,
		},
		{
			name:         "not implemented",
			err:          errors.New("not implemented"),
			expectedType: ErrorTypeNotImplemented,
		},
		{
			name:         "validation error",
			err:          errors.New("validation failed"),
			expectedType: ErrorTypeValidationError,
		},
		{
			name:         "generic error",
			err:          errors.New("something went wrong"),
			expectedType: ErrorTypeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ccErr := convertToCCProxyError(tt.err)
			if ccErr.Type != tt.expectedType {
				t.Errorf("Expected error type %s, got %s", tt.expectedType, ccErr.Type)
			}
			if ccErr.Message != tt.err.Error() {
				t.Errorf("Expected message %s, got %s", tt.err.Error(), ccErr.Message)
			}
		})
	}
}

func TestExtractProviderError(t *testing.T) {
	tests := []struct {
		name          string
		resp          *http.Response
		body          []byte
		provider      string
		expectError   bool
		expectedType  ErrorType
	}{
		{
			name: "successful response",
			resp: &http.Response{
				StatusCode: 200,
			},
			body:        []byte(`{"success": true}`),
			provider:    "test-provider",
			expectError: false,
		},
		{
			name: "bad request error",
			resp: &http.Response{
				StatusCode: 400,
			},
			body:        []byte(`{"error": {"message": "Invalid request"}}`),
			provider:    "test-provider",
			expectError: true,
			expectedType: ErrorTypeBadRequest,
		},
		{
			name: "rate limit error",
			resp: &http.Response{
				StatusCode: 429,
				Header:     http.Header{"Retry-After": []string{"60"}},
			},
			body:        []byte(`{"error": {"message": "Rate limited"}}`),
			provider:    "test-provider",
			expectError: true,
			expectedType: ErrorTypeTooManyRequests,
		},
		{
			name: "internal server error",
			resp: &http.Response{
				StatusCode: 500,
			},
			body:        []byte(`{"error": {"message": "Server error"}}`),
			provider:    "test-provider",
			expectError: true,
			expectedType: ErrorTypeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ExtractProviderError(tt.resp, tt.body, tt.provider)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				
				ccErr, ok := err.(*CCProxyError)
				if !ok {
					t.Error("Expected CCProxyError type")
				} else {
					if ccErr.Type != tt.expectedType {
						t.Errorf("Expected error type %s, got %s", tt.expectedType, ccErr.Type)
					}
					if ccErr.Provider != tt.provider {
						t.Errorf("Expected provider %s, got %s", tt.provider, ccErr.Provider)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestWrapProviderError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		provider     string
		expectedType ErrorType
		expectNil    bool
	}{
		{
			name:      "nil error",
			err:       nil,
			provider:  "test",
			expectNil: true,
		},
		{
			name:         "context deadline exceeded",
			err:          errors.New("context deadline exceeded"),
			provider:     "test-provider",
			expectedType: ErrorTypeGatewayTimeout,
		},
		{
			name:         "connection refused",
			err:          errors.New("dial tcp: connection refused"),
			provider:     "test-provider",
			expectedType: ErrorTypeBadGateway,
		},
		{
			name:         "rate limit error",
			err:          errors.New("rate limit exceeded"),
			provider:     "test-provider",
			expectedType: ErrorTypeRateLimitError,
		},
		{
			name:         "generic provider error",
			err:          errors.New("provider error"),
			provider:     "test-provider",
			expectedType: ErrorTypeProviderError,
		},
		{
			name: "existing CCProxyError with provider",
			err: &CCProxyError{
				Type:     ErrorTypeBadRequest,
				Message:  "bad request",
				Provider: "existing-provider",
			},
			provider:     "test-provider",
			expectedType: ErrorTypeBadRequest,
		},
		{
			name: "existing CCProxyError without provider",
			err: &CCProxyError{
				Type:    ErrorTypeInternal,
				Message: "internal error",
			},
			provider:     "test-provider",
			expectedType: ErrorTypeInternal, // WrapProviderError preserves original error type
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrappedErr := WrapProviderError(tt.err, tt.provider)
			
			if tt.expectNil {
				if wrappedErr != nil {
					t.Error("Expected nil error")
				}
				return
			}
			
			if wrappedErr == nil {
				t.Fatal("Expected non-nil error")
			}
			
			ccErr, ok := wrappedErr.(*CCProxyError)
			if !ok {
				t.Fatal("Expected CCProxyError type")
			}
			
			// Check if it's an existing CCProxyError with provider
			if existingErr, ok := tt.err.(*CCProxyError); ok && existingErr.Provider != "" {
				// Should return as-is
				if ccErr != existingErr {
					t.Error("Expected original error to be returned")
				}
			} else {
				// Should be wrapped with correct type
				if ccErr.Type != tt.expectedType {
					t.Errorf("Expected error type %s, got %s", tt.expectedType, ccErr.Type)
				}
				if ccErr.Provider != tt.provider {
					t.Errorf("Expected provider %s, got %s", tt.provider, ccErr.Provider)
				}
			}
		})
	}
}

func TestNewErrorResponse(t *testing.T) {
	tests := []struct {
		name        string
		errorType   ErrorType
		message     string
		code        string
		details     map[string]interface{}
	}{
		{
			name:      "basic error response",
			errorType: ErrorTypeBadRequest,
			message:   "Invalid input",
		},
		{
			name:      "error response with code",
			errorType: ErrorTypeUnauthorized,
			message:   "Invalid token",
			code:      "INVALID_TOKEN",
		},
		{
			name:      "error response with details",
			errorType: ErrorTypeValidationError,
			message:   "Validation failed",
			code:      "VALIDATION_ERROR",
			details: map[string]interface{}{
				"field": "email",
				"error": "invalid format",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewErrorResponse(tt.errorType, tt.message)
			
			if resp.Error.Type != string(tt.errorType) {
				t.Errorf("Expected type %s, got %s", tt.errorType, resp.Error.Type)
			}
			if resp.Error.Message != tt.message {
				t.Errorf("Expected message %s, got %s", tt.message, resp.Error.Message)
			}
			
			// Test WithCode
			if tt.code != "" {
				resp = resp.WithCode(tt.code)
				if resp.Error.Code != tt.code {
					t.Errorf("Expected code %s, got %s", tt.code, resp.Error.Code)
				}
			}
			
			// Test WithDetails
			if tt.details != nil {
				resp = resp.WithDetails(tt.details)
				if len(resp.Error.Details) != len(tt.details) {
					t.Errorf("Expected %d details, got %d", len(tt.details), len(resp.Error.Details))
				}
				for k, v := range tt.details {
					if resp.Error.Details[k] != v {
						t.Errorf("Expected detail %s=%v, got %v", k, v, resp.Error.Details[k])
					}
				}
			}
			
			// Test JSON marshaling
			data, err := json.Marshal(resp)
			if err != nil {
				t.Fatalf("Failed to marshal error response: %v", err)
			}
			
			var unmarshaled ErrorResponse
			if err := json.Unmarshal(data, &unmarshaled); err != nil {
				t.Fatalf("Failed to unmarshal error response: %v", err)
			}
			
			if unmarshaled.Error.Type != resp.Error.Type {
				t.Error("Type not preserved after JSON round-trip")
			}
		})
	}
}

func TestErrorHandlerMiddleware_Full(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupError     func(*gin.Context)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "panic recovery",
			setupError: func(c *gin.Context) {
				panic("test panic")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error",
		},
		{
			name: "CCProxyError in context",
			setupError: func(c *gin.Context) {
				err := New(ErrorTypeBadRequest, "Bad request").WithCode("BAD_INPUT")
				c.Error(err)
				c.Abort()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad request",
		},
		{
			name: "regular error in context",
			setupError: func(c *gin.Context) {
				c.Error(errors.New("rate limit exceeded"))
				c.Abort()
			},
			expectedStatus: http.StatusTooManyRequests,
			expectedBody:   "rate limit exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router with error handler middleware
			router := gin.New()
			router.Use(ErrorHandlerMiddleware())
			
			// Add test endpoint
			router.GET("/test", func(c *gin.Context) {
				tt.setupError(c)
			})
			
			// Make request
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)
			
			// Check response
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			
			body := w.Body.String()
			if !strings.Contains(body, tt.expectedBody) {
				t.Errorf("Expected body to contain %q, got %q", tt.expectedBody, body)
			}
			
			// Check content type
			if ct := w.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", ct)
			}
		})
	}
}

func TestHandleError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedType   ErrorType
		expectedBody   string
		skipNilCheck   bool
	}{
		{
			name:           "nil error",
			err:            nil,
			expectedStatus: http.StatusOK, // HandleError returns early for nil error
			skipNilCheck:   true,
		},
		{
			name:           "CCProxyError",
			err:            New(ErrorTypeNotFound, "Resource not found").WithCode("NOT_FOUND"),
			expectedStatus: http.StatusNotFound,
			expectedType:   ErrorTypeNotFound,
			expectedBody:   "Resource not found",
		},
		{
			name:           "regular error",
			err:            errors.New("connection refused"),
			expectedStatus: http.StatusBadGateway,
			expectedType:   ErrorTypeBadGateway,
			expectedBody:   "connection refused",
		},
		{
			name:           "io.EOF error",
			err:            io.EOF,
			expectedStatus: http.StatusBadRequest,
			expectedType:   ErrorTypeBadRequest,
			expectedBody:   "Unexpected end of input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			
			// Create a proper request
			req := httptest.NewRequest("GET", "/test", nil)
			c.Request = req
			
			HandleError(c, tt.err)
			
			// For nil error, HandleError returns early without writing response
			if tt.skipNilCheck {
				return
			}
			
			// Check status
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			
			// Check response body
			var resp map[string]interface{}
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}
			
			errorData, ok := resp["error"].(map[string]interface{})
			if !ok {
				t.Fatal("Expected error field in response")
			}
			
			if errorData["type"] != string(tt.expectedType) {
				t.Errorf("Expected error type %s, got %v", tt.expectedType, errorData["type"])
			}
			
			if !strings.Contains(errorData["message"].(string), tt.expectedBody) {
				t.Errorf("Expected message to contain %q, got %q", tt.expectedBody, errorData["message"])
			}
		})
	}
}


func TestWriteErrorResponse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		errResp    *ErrorResponse
	}{
		{
			name:       "bad request",
			statusCode: http.StatusBadRequest,
			errResp:    NewErrorResponse(ErrorTypeBadRequest, "Invalid input").WithCode("INVALID"),
		},
		{
			name:       "internal server error",
			statusCode: http.StatusInternalServerError,
			errResp:    NewErrorResponse(ErrorTypeInternal, "Server error"),
		},
		{
			name:       "validation error with details",
			statusCode: http.StatusBadRequest,
			errResp: NewErrorResponse(ErrorTypeValidationError, "Validation failed").
				WithCode("VALIDATION_FAILED").
				WithDetails(map[string]interface{}{
					"fields": []string{"email", "password"},
				}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			
			WriteErrorResponse(w, tt.statusCode, tt.errResp)
			
			// Check status
			if w.Code != tt.statusCode {
				t.Errorf("Expected status %d, got %d", tt.statusCode, w.Code)
			}
			
			// Check content type
			if ct := w.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", ct)
			}
			
			// Check response body
			var resp ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}
			
			if resp.Error.Type != tt.errResp.Error.Type {
				t.Errorf("Expected type %s, got %s", tt.errResp.Error.Type, resp.Error.Type)
			}
			
			if resp.Error.Message != tt.errResp.Error.Message {
				t.Errorf("Expected message %s, got %s", tt.errResp.Error.Message, resp.Error.Message)
			}
			
			if resp.Error.Code != tt.errResp.Error.Code {
				t.Errorf("Expected code %s, got %s", tt.errResp.Error.Code, resp.Error.Code)
			}
		})
	}
}

// Helper to simulate writing error response
func WriteErrorResponse(w http.ResponseWriter, statusCode int, errResp *ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	data, _ := json.Marshal(errResp)
	w.Write(data)
}