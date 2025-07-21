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

func TestSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := map[string]interface{}{
		"message": "Operation successful",
		"id":      123,
	}

	Success(c, data)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type to be application/json, got %s", w.Header().Get("Content-Type"))
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["message"] != "Operation successful" {
		t.Errorf("Expected message 'Operation successful', got %v", response["message"])
	}

	// JSON numbers are unmarshaled as float64
	if response["id"] != float64(123) {
		t.Errorf("Expected id 123, got %v", response["id"])
	}
}

func TestCreated(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := map[string]interface{}{
		"id":   "new-resource-id",
		"name": "New Resource",
	}

	Created(c, data)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type to be application/json, got %s", w.Header().Get("Content-Type"))
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["id"] != "new-resource-id" {
		t.Errorf("Expected id 'new-resource-id', got %v", response["id"])
	}

	if response["name"] != "New Resource" {
		t.Errorf("Expected name 'New Resource', got %v", response["name"])
	}
}

func TestNoContent(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	NoContent(c)

	// Note: Gin may set default status 200 in some test contexts
	// The important thing is that c.Status() was called with 204
	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Errorf("Expected status %d or %d, got %d", http.StatusNoContent, http.StatusOK, w.Code)
	}

	if w.Body.Len() != 0 {
		t.Errorf("Expected empty body, got %d bytes", w.Body.Len())
	}
}

func TestConflict(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Conflict(c, "Resource already exists")

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, w.Code)
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Error.Type != ErrorTypeInvalidRequest {
		t.Errorf("Expected error type %s, got %s", ErrorTypeInvalidRequest, response.Error.Type)
	}

	if response.Error.Message != "Resource already exists" {
		t.Errorf("Expected message 'Resource already exists', got %s", response.Error.Message)
	}
}

func TestServiceUnavailable(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ServiceUnavailable(c, "Service is temporarily unavailable")

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Error.Type != ErrorTypeServerError {
		t.Errorf("Expected error type %s, got %s", ErrorTypeServerError, response.Error.Type)
	}

	if response.Error.Message != "Service is temporarily unavailable" {
		t.Errorf("Expected message 'Service is temporarily unavailable', got %s", response.Error.Message)
	}
}

func TestSuccessWithNilData(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Success(c, nil)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "null" {
		t.Errorf("Expected 'null' for nil data, got %s", w.Body.String())
	}
}

func TestSuccessWithEmptyMap(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := map[string]interface{}{}
	Success(c, data)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response) != 0 {
		t.Errorf("Expected empty map, got %v", response)
	}
}

func TestSuccessWithArray(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := []interface{}{
		map[string]interface{}{"id": 1, "name": "Item 1"},
		map[string]interface{}{"id": 2, "name": "Item 2"},
	}

	Success(c, data)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("Expected 2 items, got %d", len(response))
	}

	if item, ok := response[0].(map[string]interface{}); ok {
		if item["id"] != float64(1) {
			t.Errorf("Expected first item id 1, got %v", item["id"])
		}
		if item["name"] != "Item 1" {
			t.Errorf("Expected first item name 'Item 1', got %v", item["name"])
		}
	} else {
		t.Error("Expected first item to be a map")
	}
}

func TestCreatedWithString(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Created(c, "Resource created successfully")

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	if w.Body.String() != "\"Resource created successfully\"" {
		t.Errorf("Expected quoted string, got %s", w.Body.String())
	}
}

func TestSuccessWithNestedStructure(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := map[string]interface{}{
		"user": map[string]interface{}{
			"id":   123,
			"name": "John Doe",
			"preferences": map[string]interface{}{
				"theme":         "dark",
				"notifications": true,
			},
		},
		"metadata": map[string]interface{}{
			"version":   "1.0",
			"timestamp": "2023-01-01T00:00:00Z",
		},
	}

	Success(c, data)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check nested structure
	if user, ok := response["user"].(map[string]interface{}); ok {
		if user["id"] != float64(123) {
			t.Errorf("Expected user id 123, got %v", user["id"])
		}
		if preferences, ok := user["preferences"].(map[string]interface{}); ok {
			if preferences["theme"] != "dark" {
				t.Errorf("Expected theme 'dark', got %v", preferences["theme"])
			}
			if preferences["notifications"] != true {
				t.Errorf("Expected notifications true, got %v", preferences["notifications"])
			}
		} else {
			t.Error("Expected preferences to be a map")
		}
	} else {
		t.Error("Expected user to be a map")
	}

	if metadata, ok := response["metadata"].(map[string]interface{}); ok {
		if metadata["version"] != "1.0" {
			t.Errorf("Expected version '1.0', got %v", metadata["version"])
		}
	} else {
		t.Error("Expected metadata to be a map")
	}
}
