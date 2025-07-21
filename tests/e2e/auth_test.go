package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthentication(t *testing.T) {
	// Setup
	configPath := createTestConfig(t)
	stopCCProxy := startCCProxy(t, configPath)
	defer stopCCProxy()
	
	mockProvider := startMockProvider(t)
	defer mockProvider.Stop()
	
	t.Run("No API Key", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "", // Empty authorization
		}
		
		t.Logf("Making request with headers: %v", headers)
		resp, body := makeRequest(t, "POST", "/v1/messages", map[string]interface{}{
			"model":    "mock-model",
			"messages": []map[string]interface{}{{"role": "user", "content": "test"}},
		}, headers)
		t.Logf("Got response body: %s", string(body))
		
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Contains(t, string(body), "Invalid API key")
	})
	
	t.Run("Invalid API Key", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Bearer invalid-key",
		}
		
		resp, body := makeRequest(t, "POST", "/v1/messages", map[string]interface{}{
			"model":    "mock-model",
			"messages": []map[string]interface{}{{"role": "user", "content": "test"}},
		}, headers)
		
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Contains(t, string(body), "Invalid API key")
	})
	
	t.Run("Valid API Key", func(t *testing.T) {
		// Clear previous requests
		mockProvider.ClearRequests()
		
		headers := map[string]string{
			"Authorization": "Bearer " + testAPIKey,
		}
		
		resp, _ := makeRequest(t, "POST", "/v1/messages", map[string]interface{}{
			"model":    "mock-model",
			"messages": []map[string]interface{}{{"role": "user", "content": "test"}},
		}, headers)
		
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		
		// Verify request was forwarded to provider
		requests := mockProvider.GetRequests()
		assert.Len(t, requests, 1)
	})
	
	t.Run("Health Endpoint No Auth", func(t *testing.T) {
		// Health endpoint should not require authentication
		headers := map[string]string{
			"Authorization": "", // No auth
		}
		
		resp, _ := makeRequest(t, "GET", "/health", nil, headers)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
	
	t.Run("Status Endpoint No Auth", func(t *testing.T) {
		// Status endpoint should not require authentication
		headers := map[string]string{
			"Authorization": "", // No auth
		}
		
		resp, _ := makeRequest(t, "GET", "/status", nil, headers)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}