package errors

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestErrorHandlerMiddleware(t *testing.T) {
	t.Run("NoErrors", func(t *testing.T) {
		w := httptest.NewRecorder()
		router := gin.New()
		router.Use(ErrorHandlerMiddleware())
		router.GET("/test", func(c *gin.Context) {
			// No errors
			c.Status(200)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		testutil.AssertEqual(t, 200, w.Code)
	})

	t.Run("WithGinError", func(t *testing.T) {
		w := httptest.NewRecorder()
		router := gin.New()
		router.Use(ErrorHandlerMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.Error(New(ErrorTypeBadRequest, "test error"))
		})

		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		testutil.AssertEqual(t, http.StatusBadRequest, w.Code)
		testutil.AssertEqual(t, "application/json", w.Header().Get("Content-Type"))

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		testutil.AssertNoError(t, err)

		errorObj := response["error"].(map[string]interface{})
		testutil.AssertEqual(t, "bad_request", errorObj["type"])
		testutil.AssertEqual(t, "test error", errorObj["message"])
	})

	t.Run("WithPanic", func(t *testing.T) {
		w := httptest.NewRecorder()
		router := gin.New()
		router.Use(ErrorHandlerMiddleware())
		router.GET("/test", func(c *gin.Context) {
			panic("test panic")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		testutil.AssertEqual(t, http.StatusInternalServerError, w.Code)
		testutil.AssertEqual(t, "application/json", w.Header().Get("Content-Type"))

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		testutil.AssertNoError(t, err)

		errorObj := response["error"].(map[string]interface{})
		testutil.AssertEqual(t, "internal_error", errorObj["type"])
		testutil.AssertEqual(t, "Internal server error", errorObj["message"])
	})

	t.Run("WithRequestID", func(t *testing.T) {
		w := httptest.NewRecorder()
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("request_id", "req-123")
			c.Next()
		})
		router.Use(ErrorHandlerMiddleware())
		router.GET("/test", func(c *gin.Context) {
			panic("test panic")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		testutil.AssertEqual(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		testutil.AssertNoError(t, err)

		errorObj := response["error"].(map[string]interface{})
		testutil.AssertEqual(t, "req-123", errorObj["request_id"])
	})
}

func TestHandleError(t *testing.T) {
	t.Run("NilError", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		HandleError(c, nil)

		testutil.AssertEqual(t, 200, w.Code) // Should not modify response
	})

	t.Run("CCProxyError", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		err := New(ErrorTypeBadRequest, "test error").
			WithCode("CODE123").
			WithProvider("test-provider").
			WithDetails(map[string]interface{}{"key": "value"})

		HandleError(c, err)

		testutil.AssertEqual(t, http.StatusBadRequest, w.Code)
		testutil.AssertEqual(t, "application/json", w.Header().Get("Content-Type"))

		var response map[string]interface{}
		unmarshalErr := json.Unmarshal(w.Body.Bytes(), &response)
		testutil.AssertNoError(t, unmarshalErr)

		errorObj := response["error"].(map[string]interface{})
		testutil.AssertEqual(t, "bad_request", errorObj["type"])
		testutil.AssertEqual(t, "test error", errorObj["message"])
		testutil.AssertEqual(t, "CODE123", errorObj["code"])
		testutil.AssertEqual(t, "test-provider", errorObj["provider"])

		details := errorObj["details"].(map[string]interface{})
		testutil.AssertEqual(t, "value", details["key"])
	})

	t.Run("RegularError", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		err := io.EOF

		HandleError(c, err)

		testutil.AssertEqual(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		unmarshalErr := json.Unmarshal(w.Body.Bytes(), &response)
		testutil.AssertNoError(t, unmarshalErr)

		errorObj := response["error"].(map[string]interface{})
		testutil.AssertEqual(t, "bad_request", errorObj["type"])
		testutil.AssertEqual(t, "Unexpected end of input", errorObj["message"])
	})

	t.Run("WithRequestID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Set("request_id", "req-456")

		err := New(ErrorTypeBadRequest, "test error")

		HandleError(c, err)

		var response map[string]interface{}
		unmarshalErr := json.Unmarshal(w.Body.Bytes(), &response)
		testutil.AssertNoError(t, unmarshalErr)

		errorObj := response["error"].(map[string]interface{})
		testutil.AssertEqual(t, "req-456", errorObj["request_id"])
	})
}

func TestHandleErrorWithStatus(t *testing.T) {
	t.Run("NilError", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		HandleErrorWithStatus(c, 500, nil)

		testutil.AssertEqual(t, 200, w.Code) // Should not modify response
	})

	t.Run("CCProxyError", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		err := New(ErrorTypeBadRequest, "test error")

		HandleErrorWithStatus(c, 418, err) // Teapot status

		testutil.AssertEqual(t, 418, w.Code)
	})

	t.Run("RegularError", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		HandleErrorWithStatus(c, 503, io.EOF)

		testutil.AssertEqual(t, 503, w.Code)

		var response map[string]interface{}
		unmarshalErr := json.Unmarshal(w.Body.Bytes(), &response)
		testutil.AssertNoError(t, unmarshalErr)

		errorObj := response["error"].(map[string]interface{})
		testutil.AssertEqual(t, "service_unavailable", errorObj["type"])
		testutil.AssertEqual(t, "EOF", errorObj["message"])
	})
}

func TestAbortWithError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	err := New(ErrorTypeBadRequest, "test error")

	AbortWithError(c, err)

	testutil.AssertEqual(t, http.StatusBadRequest, w.Code)
	testutil.AssertTrue(t, c.IsAborted())
}

func TestAbortWithCCProxyError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	AbortWithCCProxyError(c, ErrorTypeNotFound, "resource not found")

	testutil.AssertEqual(t, http.StatusNotFound, w.Code)
	testutil.AssertTrue(t, c.IsAborted())

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	testutil.AssertNoError(t, err)

	errorObj := response["error"].(map[string]interface{})
	testutil.AssertEqual(t, "not_found", errorObj["type"])
	testutil.AssertEqual(t, "resource not found", errorObj["message"])
}

func TestHandleGinError(t *testing.T) {
	t.Run("AlreadyWritten", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		// Write something to mark as written using gin context
		c.Writer.WriteHeader(200)
		c.Writer.Write([]byte("already written"))

		ginErr := &gin.Error{Err: New(ErrorTypeBadRequest, "test error")}
		handleGinError(c, ginErr)

		// Should not override existing response
		testutil.AssertEqual(t, 200, w.Code)
		testutil.AssertEqual(t, "already written", w.Body.String())
	})

	t.Run("NotWritten", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		ginErr := &gin.Error{Err: New(ErrorTypeBadRequest, "test error")}
		handleGinError(c, ginErr)

		testutil.AssertEqual(t, http.StatusBadRequest, w.Code)
	})
}

func TestConvertToCCProxyError(t *testing.T) {
	tests := []struct {
		name         string
		inputError   error
		expectedType ErrorType
		expectedMsg  string
	}{
		{
			name:         "EOF",
			inputError:   io.EOF,
			expectedType: ErrorTypeBadRequest,
			expectedMsg:  "Unexpected end of input",
		},
		{
			name:         "Unauthorized",
			inputError:   &testError{"unauthorized access"},
			expectedType: ErrorTypeUnauthorized,
			expectedMsg:  "unauthorized access",
		},
		{
			name:         "Forbidden",
			inputError:   &testError{"forbidden resource"},
			expectedType: ErrorTypeForbidden,
			expectedMsg:  "forbidden resource",
		},
		{
			name:         "Not Found",
			inputError:   &testError{"not found"},
			expectedType: ErrorTypeNotFound,
			expectedMsg:  "not found",
		},
		{
			name:         "Bad Request",
			inputError:   &testError{"bad request"},
			expectedType: ErrorTypeBadRequest,
			expectedMsg:  "bad request",
		},
		{
			name:         "Rate Limit",
			inputError:   &testError{"rate limit exceeded"},
			expectedType: ErrorTypeRateLimitError,
			expectedMsg:  "rate limit exceeded",
		},
		{
			name:         "Timeout",
			inputError:   &testError{"request timeout"},
			expectedType: ErrorTypeGatewayTimeout,
			expectedMsg:  "request timeout",
		},
		{
			name:         "Connection Refused",
			inputError:   &testError{"connection refused"},
			expectedType: ErrorTypeBadGateway,
			expectedMsg:  "connection refused",
		},
		{
			name:         "Connection Reset",
			inputError:   &testError{"connection reset by peer"},
			expectedType: ErrorTypeBadGateway,
			expectedMsg:  "connection reset by peer",
		},
		{
			name:         "No Such Host",
			inputError:   &testError{"no such host"},
			expectedType: ErrorTypeBadGateway,
			expectedMsg:  "no such host",
		},
		{
			name:         "Service Unavailable",
			inputError:   &testError{"service unavailable"},
			expectedType: ErrorTypeServiceUnavailable,
			expectedMsg:  "service unavailable",
		},
		{
			name:         "Not Implemented",
			inputError:   &testError{"not implemented"},
			expectedType: ErrorTypeNotImplemented,
			expectedMsg:  "not implemented",
		},
		{
			name:         "Validation",
			inputError:   &testError{"validation failed"},
			expectedType: ErrorTypeValidationError,
			expectedMsg:  "validation failed",
		},
		{
			name:         "Default",
			inputError:   &testError{"unknown error"},
			expectedType: ErrorTypeInternal,
			expectedMsg:  "unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToCCProxyError(tt.inputError)
			testutil.AssertEqual(t, tt.expectedType, result.Type)
			testutil.AssertEqual(t, tt.expectedMsg, result.Message)
		})
	}
}

func TestExtractProviderError(t *testing.T) {
	t.Run("SuccessResponse", func(t *testing.T) {
		resp := &http.Response{StatusCode: 200}
		err := ExtractProviderError(resp, []byte(""), "test-provider")
		testutil.AssertEqual(t, true, err == nil)
	})

	t.Run("ErrorResponse", func(t *testing.T) {
		resp := &http.Response{StatusCode: 400}
		body := []byte(`{"error":{"message":"Bad request"}}`)

		err := ExtractProviderError(resp, body, "test-provider")
		testutil.AssertEqual(t, false, err == nil)

		ccErr, ok := err.(*CCProxyError)
		testutil.AssertTrue(t, ok)
		testutil.AssertEqual(t, "test-provider", ccErr.Provider)
		testutil.AssertEqual(t, 400, ccErr.StatusCode)
	})
}

func TestWrapProviderError(t *testing.T) {
	t.Run("NilError", func(t *testing.T) {
		result := WrapProviderError(nil, "test-provider")
		testutil.AssertEqual(t, true, result == nil)
	})

	t.Run("CCProxyErrorWithProvider", func(t *testing.T) {
		originalErr := New(ErrorTypeBadRequest, "test error").WithProvider("existing-provider")
		result := WrapProviderError(originalErr, "new-provider")

		// Should return as-is since it already has provider info
		testutil.AssertEqual(t, originalErr, result)

		ccErr := result.(*CCProxyError)
		testutil.AssertEqual(t, "existing-provider", ccErr.Provider)
	})

	t.Run("CCProxyErrorWithoutProvider", func(t *testing.T) {
		originalErr := New(ErrorTypeBadRequest, "test error")
		result := WrapProviderError(originalErr, "new-provider")

		// Should add provider info
		testutil.AssertEqual(t, originalErr, result)

		ccErr := result.(*CCProxyError)
		testutil.AssertEqual(t, "new-provider", ccErr.Provider)
	})

	t.Run("RegularErrorWithTimeout", func(t *testing.T) {
		originalErr := &testError{"context deadline exceeded"}
		result := WrapProviderError(originalErr, "test-provider")

		ccErr, ok := result.(*CCProxyError)
		testutil.AssertTrue(t, ok)
		testutil.AssertEqual(t, ErrorTypeGatewayTimeout, ccErr.Type)
		testutil.AssertEqual(t, "test-provider", ccErr.Provider)
		testutil.AssertEqual(t, originalErr, ccErr.wrapped)
	})

	t.Run("RegularErrorWithConnectionRefused", func(t *testing.T) {
		originalErr := &testError{"connection refused"}
		result := WrapProviderError(originalErr, "test-provider")

		ccErr, ok := result.(*CCProxyError)
		testutil.AssertTrue(t, ok)
		testutil.AssertEqual(t, ErrorTypeBadGateway, ccErr.Type)
		testutil.AssertEqual(t, "test-provider", ccErr.Provider)
	})

	t.Run("RegularErrorWithRateLimit", func(t *testing.T) {
		originalErr := &testError{"rate limit exceeded"}
		result := WrapProviderError(originalErr, "test-provider")

		ccErr, ok := result.(*CCProxyError)
		testutil.AssertTrue(t, ok)
		testutil.AssertEqual(t, ErrorTypeRateLimitError, ccErr.Type)
		testutil.AssertEqual(t, "test-provider", ccErr.Provider)
	})

	t.Run("RegularErrorDefault", func(t *testing.T) {
		originalErr := &testError{"generic error"}
		result := WrapProviderError(originalErr, "test-provider")

		ccErr, ok := result.(*CCProxyError)
		testutil.AssertTrue(t, ok)
		testutil.AssertEqual(t, ErrorTypeProviderError, ccErr.Type)
		testutil.AssertEqual(t, "Provider test-provider error", ccErr.Message)
		testutil.AssertEqual(t, "test-provider", ccErr.Provider)
	})
}

func TestNewErrorResponse(t *testing.T) {
	resp := NewErrorResponse(ErrorTypeBadRequest, "test message")

	testutil.AssertEqual(t, "bad_request", resp.Error.Type)
	testutil.AssertEqual(t, "test message", resp.Error.Message)
	testutil.AssertEqual(t, "", resp.Error.Code)
	testutil.AssertEqual(t, true, resp.Error.Details == nil)
}

func TestErrorResponse_WithCode(t *testing.T) {
	resp := NewErrorResponse(ErrorTypeBadRequest, "test message").
		WithCode("CODE123")

	testutil.AssertEqual(t, "CODE123", resp.Error.Code)
}

func TestErrorResponse_WithDetails(t *testing.T) {
	details := map[string]interface{}{"key": "value"}
	resp := NewErrorResponse(ErrorTypeBadRequest, "test message").
		WithDetails(details)

	testutil.AssertEqual(t, "value", resp.Error.Details["key"])
}

func TestErrorResponse_Chaining(t *testing.T) {
	resp := NewErrorResponse(ErrorTypeBadRequest, "test message").
		WithCode("CODE123").
		WithDetails(map[string]interface{}{"key": "value"})

	testutil.AssertEqual(t, "bad_request", resp.Error.Type)
	testutil.AssertEqual(t, "test message", resp.Error.Message)
	testutil.AssertEqual(t, "CODE123", resp.Error.Code)
	testutil.AssertEqual(t, "value", resp.Error.Details["key"])
}

// testError is a helper type for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}
