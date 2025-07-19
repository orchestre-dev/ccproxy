package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/musistudio/ccproxy/internal/config"
)

func TestProviderEndpoints(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		apiKey         string
		expectedStatus int
		expectedBody   string
		checkResponse  func(t *testing.T, body []byte)
	}{
		// List providers
		{
			name:           "list providers with auth",
			method:         "GET",
			path:           "/providers",
			apiKey:         "test-key",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var providers []config.Provider
				if err := json.Unmarshal(body, &providers); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
			},
		},
		{
			name:           "list providers without auth",
			method:         "GET",
			path:           "/providers",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid API key",
		},

		// Create provider
		{
			name:   "create valid provider",
			method: "POST",
			path:   "/providers",
			body: CreateProviderRequest{
				Name:       "test-provider",
				APIBaseURL: "https://api.test.com",
				APIKey:     "test-provider-key",
				Models:     []string{"model1", "model2"},
				Enabled:    true,
			},
			apiKey:         "test-key",
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var provider config.Provider
				if err := json.Unmarshal(body, &provider); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if provider.Name != "test-provider" {
					t.Errorf("Expected provider name 'test-provider', got '%s'", provider.Name)
				}
			},
		},
		{
			name:   "create provider missing name",
			method: "POST",
			path:   "/providers",
			body: CreateProviderRequest{
				APIBaseURL: "https://api.test.com",
				APIKey:     "test-provider-key",
			},
			apiKey:         "test-key",
			expectedStatus: http.StatusBadRequest,
		},

		// Get provider
		{
			name:           "get existing provider",
			method:         "GET",
			path:           "/providers/test-provider",
			apiKey:         "test-key",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get non-existent provider",
			method:         "GET",
			path:           "/providers/non-existent",
			apiKey:         "test-key",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Provider 'non-existent' not found",
		},

		// Toggle provider
		{
			name:           "toggle provider enabled",
			method:         "PATCH",
			path:           "/providers/test-provider/toggle",
			apiKey:         "test-key",
			expectedStatus: http.StatusOK,
			expectedBody:   "successfully",
		},

		// Delete provider
		{
			name:           "delete existing provider",
			method:         "DELETE",
			path:           "/providers/test-provider",
			apiKey:         "test-key",
			expectedStatus: http.StatusOK,
			expectedBody:   "Provider deleted successfully",
		},
		{
			name:           "delete non-existent provider",
			method:         "DELETE",
			path:           "/providers/non-existent",
			apiKey:         "test-key",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Provider 'non-existent' not found",
		},
	}

	// Create test server with in-memory config
	cfg := &config.Config{
		APIKey: "test-key",
		Port:   3456,
		Host:   "127.0.0.1",
		Providers: []config.Provider{
			{
				Name:       "test-provider",
				APIBaseURL: "https://api.test.com",
				APIKey:     "provider-key",
				Models:     []string{"model1"},
				Enabled:    true,
			},
		},
	}

	srv, err := NewWithPath(cfg, "")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request body
			var reqBody []byte
			if tt.body != nil {
				reqBody, _ = json.Marshal(tt.body)
			}

			// Create request
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewReader(reqBody))
			if tt.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			if tt.apiKey != "" {
				req.Header.Set("x-api-key", tt.apiKey)
			}

			// Perform request
			w := httptest.NewRecorder()
			srv.router.ServeHTTP(w, req)

			// Check status
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			// Check body
			if tt.expectedBody != "" && !contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("Expected body to contain '%s', got '%s'", tt.expectedBody, w.Body.String())
			}

			// Custom response check
			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

func TestUpdateProvider(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test server
	cfg := &config.Config{
		APIKey: "test-key",
		Port:   3456,
		Host:   "127.0.0.1",
		Providers: []config.Provider{
			{
				Name:       "provider1",
				APIBaseURL: "https://api1.test.com",
				APIKey:     "key1",
				Models:     []string{"model1"},
				Enabled:    true,
			},
			{
				Name:       "provider2",
				APIBaseURL: "https://api2.test.com",
				APIKey:     "key2",
				Models:     []string{"model2"},
				Enabled:    false,
			},
		},
	}

	srv, err := NewWithPath(cfg, "")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tests := []struct {
		name           string
		providerName   string
		update         UpdateProviderRequest
		expectedStatus int
		checkResponse  func(t *testing.T, provider *config.Provider)
	}{
		{
			name:         "update API URL",
			providerName: "provider1",
			update: UpdateProviderRequest{
				APIBaseURL: "https://new-api.test.com",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, provider *config.Provider) {
				if provider.APIBaseURL != "https://new-api.test.com" {
					t.Errorf("Expected API URL to be updated")
				}
			},
		},
		{
			name:         "update enabled state",
			providerName: "provider2",
			update: UpdateProviderRequest{
				Enabled: ptrBool(true),
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, provider *config.Provider) {
				if !provider.Enabled {
					t.Errorf("Expected provider to be enabled")
				}
			},
		},
		{
			name:         "rename provider",
			providerName: "provider1",
			update: UpdateProviderRequest{
				Name: "provider1-renamed",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, provider *config.Provider) {
				if provider.Name != "provider1-renamed" {
					t.Errorf("Expected provider name to be updated")
				}
			},
		},
		{
			name:         "rename to existing name",
			providerName: "provider1",
			update: UpdateProviderRequest{
				Name: "provider2",
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:         "update non-existent provider",
			providerName: "non-existent",
			update: UpdateProviderRequest{
				APIKey: "new-key",
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request
			body, _ := json.Marshal(tt.update)
			req := httptest.NewRequest("PUT", "/providers/"+tt.providerName, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("x-api-key", "test-key")

			// Perform request
			w := httptest.NewRecorder()
			srv.router.ServeHTTP(w, req)

			// Check status
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			// Check response
			if tt.checkResponse != nil && w.Code == http.StatusOK {
				var provider config.Provider
				if err := json.Unmarshal(w.Body.Bytes(), &provider); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				tt.checkResponse(t, &provider)
			}
		})
	}
}

// Helper function to create bool pointer
func ptrBool(b bool) *bool {
	return &b
}