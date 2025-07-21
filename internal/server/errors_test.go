package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		errorType ErrorType
		expected  string
	}{
		{ErrorTypeInvalidRequest, "invalid_request"},
		{ErrorTypeNotFound, "not_found"},
		{ErrorTypeAuthentication, "authentication_error"},
		{ErrorTypePermission, "permission_error"},
		{ErrorTypeRateLimit, "rate_limit_error"},
		{ErrorTypeProviderError, "provider_error"},
		{ErrorTypeServerError, "server_error"},
		{ErrorTypeNotImplemented, "not_implemented"},
	}

	for _, test := range tests {
		if string(test.errorType) != test.expected {
			t.Errorf("Expected error type %s, got %s", test.expected, string(test.errorType))
		}
	}
}

func TestRespondWithError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RespondWithError(c, http.StatusBadRequest, ErrorTypeInvalidRequest, "Test error message")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type to be application/json, got %s", w.Header().Get("Content-Type"))
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Error.Type != ErrorTypeInvalidRequest {
		t.Errorf("Expected error type %s, got %s", ErrorTypeInvalidRequest, response.Error.Type)
	}

	if response.Error.Message != "Test error message" {
		t.Errorf("Expected message 'Test error message', got %s", response.Error.Message)
	}

	if response.Error.Code != "" {
		t.Errorf("Expected empty code, got %s", response.Error.Code)
	}
}

func TestRespondWithErrorCode(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RespondWithErrorCode(c, http.StatusInternalServerError, ErrorTypeServerError, "Server error", "ERR_500")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Error.Type != ErrorTypeServerError {
		t.Errorf("Expected error type %s, got %s", ErrorTypeServerError, response.Error.Type)
	}

	if response.Error.Message != "Server error" {
		t.Errorf("Expected message 'Server error', got %s", response.Error.Message)
	}

	if response.Error.Code != "ERR_500" {
		t.Errorf("Expected code 'ERR_500', got %s", response.Error.Code)
	}
}

func TestBadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	BadRequest(c, "Invalid input")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Error.Type != ErrorTypeInvalidRequest {
		t.Errorf("Expected error type %s, got %s", ErrorTypeInvalidRequest, response.Error.Type)
	}

	if response.Error.Message != "Invalid input" {
		t.Errorf("Expected message 'Invalid input', got %s", response.Error.Message)
	}
}

func TestUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Unauthorized(c, "Invalid credentials")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Error.Type != ErrorTypeAuthentication {
		t.Errorf("Expected error type %s, got %s", ErrorTypeAuthentication, response.Error.Type)
	}

	if response.Error.Message != "Invalid credentials" {
		t.Errorf("Expected message 'Invalid credentials', got %s", response.Error.Message)
	}
}

func TestForbidden(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Forbidden(c, "Access denied")

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Error.Type != ErrorTypePermission {
		t.Errorf("Expected error type %s, got %s", ErrorTypePermission, response.Error.Type)
	}

	if response.Error.Message != "Access denied" {
		t.Errorf("Expected message 'Access denied', got %s", response.Error.Message)
	}
}

func TestNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	NotFound(c, "Resource not found")

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Error.Type != ErrorTypeNotFound {
		t.Errorf("Expected error type %s, got %s", ErrorTypeNotFound, response.Error.Type)
	}

	if response.Error.Message != "Resource not found" {
		t.Errorf("Expected message 'Resource not found', got %s", response.Error.Message)
	}
}

func TestInternalServerError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	InternalServerError(c, "Something went wrong")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Error.Type != ErrorTypeServerError {
		t.Errorf("Expected error type %s, got %s", ErrorTypeServerError, response.Error.Type)
	}

	if response.Error.Message != "Something went wrong" {
		t.Errorf("Expected message 'Something went wrong', got %s", response.Error.Message)
	}
}

func TestNotImplemented(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	NotImplemented(c, "Feature not implemented")

	if w.Code != http.StatusNotImplemented {
		t.Errorf("Expected status %d, got %d", http.StatusNotImplemented, w.Code)
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Error.Type != ErrorTypeNotImplemented {
		t.Errorf("Expected error type %s, got %s", ErrorTypeNotImplemented, response.Error.Type)
	}

	if response.Error.Message != "Feature not implemented" {
		t.Errorf("Expected message 'Feature not implemented', got %s", response.Error.Message)
	}
}

func TestProviderError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ProviderError(c, "Provider is down")

	if w.Code != http.StatusBadGateway {
		t.Errorf("Expected status %d, got %d", http.StatusBadGateway, w.Code)
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Error.Type != ErrorTypeProviderError {
		t.Errorf("Expected error type %s, got %s", ErrorTypeProviderError, response.Error.Type)
	}

	if response.Error.Message != "Provider is down" {
		t.Errorf("Expected message 'Provider is down', got %s", response.Error.Message)
	}
}

func TestRateLimitError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RateLimitError(c, "Too many requests")

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status %d, got %d", http.StatusTooManyRequests, w.Code)
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Error.Type != ErrorTypeRateLimit {
		t.Errorf("Expected error type %s, got %s", ErrorTypeRateLimit, response.Error.Type)
	}

	if response.Error.Message != "Too many requests" {
		t.Errorf("Expected message 'Too many requests', got %s", response.Error.Message)
	}
}

func TestErrorResponse(t *testing.T) {
	errorResp := ErrorResponse{
		Error: ErrorDetail{
			Message: "Test error",
			Type:    ErrorTypeInvalidRequest,
			Code:    "TEST_001",
		},
	}

	if errorResp.Error.Message != "Test error" {
		t.Errorf("Expected message 'Test error', got %s", errorResp.Error.Message)
	}

	if errorResp.Error.Type != ErrorTypeInvalidRequest {
		t.Errorf("Expected type %s, got %s", ErrorTypeInvalidRequest, errorResp.Error.Type)
	}

	if errorResp.Error.Code != "TEST_001" {
		t.Errorf("Expected code 'TEST_001', got %s", errorResp.Error.Code)
	}
}

func TestErrorDetail(t *testing.T) {
	detail := ErrorDetail{
		Message: "Detailed error message",
		Type:    ErrorTypeServerError,
		Code:    "ERR_500",
	}

	if detail.Message != "Detailed error message" {
		t.Errorf("Expected message 'Detailed error message', got %s", detail.Message)
	}

	if detail.Type != ErrorTypeServerError {
		t.Errorf("Expected type %s, got %s", ErrorTypeServerError, detail.Type)
	}

	if detail.Code != "ERR_500" {
		t.Errorf("Expected code 'ERR_500', got %s", detail.Code)
	}
}

func TestErrorResponseSerialization(t *testing.T) {
	errorResp := ErrorResponse{
		Error: ErrorDetail{
			Message: "Serialization test",
			Type:    ErrorTypeNotFound,
			Code:    "NOT_FOUND",
		},
	}

	data, err := json.Marshal(errorResp)
	if err != nil {
		t.Fatalf("Failed to marshal error response: %v", err)
	}

	var unmarshaled ErrorResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal error response: %v", err)
	}

	if unmarshaled.Error.Message != errorResp.Error.Message {
		t.Errorf("Message changed during serialization")
	}

	if unmarshaled.Error.Type != errorResp.Error.Type {
		t.Errorf("Type changed during serialization")
	}

	if unmarshaled.Error.Code != errorResp.Error.Code {
		t.Errorf("Code changed during serialization")
	}
}
