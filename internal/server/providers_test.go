package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orchestre-dev/ccproxy/internal/config"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func createTestServerWithProviders(t *testing.T) *Server {
	cfg := &config.Config{
		Host:   "127.0.0.1",
		Port:   3456,
		APIKey: "test-api-key",
		Performance: config.PerformanceConfig{
			RequestTimeout:     30 * time.Second,
			MaxRequestBodySize: 10 * 1024 * 1024,
		},
		Routes: map[string]config.Route{
			"default": {
				Provider: "openai",
				Model:    "gpt-4",
			},
		},
		Providers: []config.Provider{
			{
				Name:       "openai",
				APIBaseURL: "https://api.openai.com",
				APIKey:     "test-openai-key",
				Models:     []string{"gpt-4", "gpt-3.5-turbo"},
				Enabled:    true,
			},
			{
				Name:       "anthropic",
				APIBaseURL: "https://api.anthropic.com",
				APIKey:     "test-anthropic-key",
				Models:     []string{"claude-3-opus", "claude-3-sonnet"},
				Enabled:    false,
			},
		},
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}

	return server
}

func TestHandleListProviders(t *testing.T) {
	server := createTestServerWithProviders(t)
	router := server.GetRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/providers", nil)
	req.Header.Set("Authorization", "Bearer test-api-key")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response []config.Provider
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(response))
	}

	if response[0].Name != "openai" {
		t.Errorf("Expected first provider to be 'openai', got %s", response[0].Name)
	}

	if response[1].Name != "anthropic" {
		t.Errorf("Expected second provider to be 'anthropic', got %s", response[1].Name)
	}
}

func TestHandleGetProvider(t *testing.T) {
	server := createTestServerWithProviders(t)
	router := server.GetRouter()

	t.Run("ExistingProvider", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/providers/openai", nil)
		req.Header.Set("Authorization", "Bearer test-api-key")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response config.Provider
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Name != "openai" {
			t.Errorf("Expected provider name 'openai', got %s", response.Name)
		}

		if response.APIBaseURL != "https://api.openai.com" {
			t.Errorf("Expected API base URL 'https://api.openai.com', got %s", response.APIBaseURL)
		}

		if !response.Enabled {
			t.Error("Expected provider to be enabled")
		}
	})

	t.Run("NonExistentProvider", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/providers/nonexistent", nil)
		req.Header.Set("Authorization", "Bearer test-api-key")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}

		var response ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal error response: %v", err)
		}

		if response.Error.Type != ErrorTypeNotFound {
			t.Errorf("Expected error type %s, got %s", ErrorTypeNotFound, response.Error.Type)
		}
	})
}

func TestHandleCreateProvider(t *testing.T) {
	server := createTestServerWithProviders(t)
	router := server.GetRouter()

	t.Run("ValidProvider", func(t *testing.T) {
		reqBody := CreateProviderRequest{
			Name:       "groq",
			APIBaseURL: "https://api.groq.com",
			APIKey:     "test-groq-key",
			Models:     []string{"llama-3-8b", "mixtral-8x7b"},
			Enabled:    true,
		}

		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/providers", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-api-key")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", w.Code)
		}
	})

	t.Run("DuplicateProvider", func(t *testing.T) {
		reqBody := CreateProviderRequest{
			Name:       "openai", // Already exists
			APIBaseURL: "https://api.openai.com",
			APIKey:     "test-key",
			Models:     []string{"gpt-4"},
			Enabled:    true,
		}

		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/providers", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-api-key")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("Expected status 409, got %d", w.Code)
		}

		var response ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal error response: %v", err)
		}

		if response.Error.Type != ErrorTypeInvalidRequest {
			t.Errorf("Expected error type %s, got %s", ErrorTypeInvalidRequest, response.Error.Type)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/providers", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-api-key")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

func TestHandleUpdateProvider(t *testing.T) {
	server := createTestServerWithProviders(t)
	router := server.GetRouter()

	t.Run("UpdateExistingProvider", func(t *testing.T) {
		reqBody := UpdateProviderRequest{
			APIBaseURL: "https://new.api.openai.com",
			Models:     []string{"gpt-4", "gpt-3.5-turbo", "gpt-4-turbo"},
		}

		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/providers/openai", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-api-key")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("UpdateNonExistentProvider", func(t *testing.T) {
		reqBody := UpdateProviderRequest{
			APIBaseURL: "https://api.nonexistent.com",
		}

		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/providers/nonexistent", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-api-key")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("UpdateWithDuplicateName", func(t *testing.T) {
		reqBody := UpdateProviderRequest{
			Name: "anthropic", // Change openai to anthropic (conflict)
		}

		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/providers/openai", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-api-key")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("Expected status 409, got %d", w.Code)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/providers/openai", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-api-key")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

func TestHandleDeleteProvider(t *testing.T) {
	server := createTestServerWithProviders(t)
	router := server.GetRouter()

	t.Run("DeleteExistingProvider", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/providers/anthropic", nil)
		req.Header.Set("Authorization", "Bearer test-api-key")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("DeleteNonExistentProvider", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/providers/nonexistent", nil)
		req.Header.Set("Authorization", "Bearer test-api-key")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})
}

func TestHandleToggleProvider(t *testing.T) {
	server := createTestServerWithProviders(t)
	router := server.GetRouter()

	t.Run("ToggleExistingProvider", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/providers/anthropic/toggle", nil)
		req.Header.Set("Authorization", "Bearer test-api-key")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("ToggleNonExistentProvider", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/providers/nonexistent/toggle", nil)
		req.Header.Set("Authorization", "Bearer test-api-key")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})
}

func TestProviderRequestStructures(t *testing.T) {
	t.Run("CreateProviderRequest", func(t *testing.T) {
		req := CreateProviderRequest{
			Name:       "test-provider",
			APIBaseURL: "https://api.test.com",
			APIKey:     "test-key",
			Models:     []string{"model1", "model2"},
			Enabled:    true,
		}

		if req.Name != "test-provider" {
			t.Errorf("Expected name 'test-provider', got %s", req.Name)
		}

		if req.APIBaseURL != "https://api.test.com" {
			t.Errorf("Expected API base URL 'https://api.test.com', got %s", req.APIBaseURL)
		}

		if req.APIKey != "test-key" {
			t.Errorf("Expected API key 'test-key', got %s", req.APIKey)
		}

		if len(req.Models) != 2 {
			t.Errorf("Expected 2 models, got %d", len(req.Models))
		}

		if !req.Enabled {
			t.Error("Expected enabled to be true")
		}
	})

	t.Run("UpdateProviderRequest", func(t *testing.T) {
		enabled := true
		req := UpdateProviderRequest{
			Name:       "updated-provider",
			APIBaseURL: "https://api.updated.com",
			APIKey:     "updated-key",
			Models:     []string{"model3", "model4"},
			Enabled:    &enabled,
		}

		if req.Name != "updated-provider" {
			t.Errorf("Expected name 'updated-provider', got %s", req.Name)
		}

		if req.APIBaseURL != "https://api.updated.com" {
			t.Errorf("Expected API base URL 'https://api.updated.com', got %s", req.APIBaseURL)
		}

		if req.APIKey != "updated-key" {
			t.Errorf("Expected API key 'updated-key', got %s", req.APIKey)
		}

		if len(req.Models) != 2 {
			t.Errorf("Expected 2 models, got %d", len(req.Models))
		}

		if req.Enabled == nil || !*req.Enabled {
			t.Error("Expected enabled to be true")
		}
	})
}