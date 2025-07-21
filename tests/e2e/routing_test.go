package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouting(t *testing.T) {
	// Setup
	configPath := createTestConfig(t)
	stopCCProxy := startCCProxy(t, configPath)
	defer stopCCProxy()
	
	mockProvider := startMockProvider(t)
	defer mockProvider.Stop()
	
	t.Run("Default Routing", func(t *testing.T) {
		mockProvider.ClearRequests()
		
		// Request with just model name
		requestBody := map[string]interface{}{
			"model": "mock-model",
			"messages": []map[string]interface{}{
				{"role": "user", "content": "Hello"},
			},
		}
		
		resp, _ := makeRequest(t, "POST", "/v1/messages", requestBody, nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		
		// Verify request was routed to mock provider
		requests := mockProvider.GetRequests()
		require.Len(t, requests, 1)
		
		var capturedRequest map[string]interface{}
		err := json.Unmarshal(requests[0].Body, &capturedRequest)
		require.NoError(t, err)
		
		t.Logf("Captured request: %+v", capturedRequest)
		
		// Model should have provider prefix added by router
		assert.Equal(t, "mock,mock-model", capturedRequest["model"])
	})
	
	t.Run("Explicit Provider Routing", func(t *testing.T) {
		mockProvider.ClearRequests()
		
		// Request with provider,model format
		requestBody := map[string]interface{}{
			"model": "mock,mock-model-2",
			"messages": []map[string]interface{}{
				{"role": "user", "content": "Hello"},
			},
		}
		
		resp, _ := makeRequest(t, "POST", "/v1/messages", requestBody, nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		
		// Verify request was routed to mock provider
		requests := mockProvider.GetRequests()
		require.Len(t, requests, 1)
		
		var capturedRequest map[string]interface{}
		err := json.Unmarshal(requests[0].Body, &capturedRequest)
		require.NoError(t, err)
		
		// Model should keep the provider prefix (no transformer configured)
		assert.Equal(t, "mock,mock-model-2", capturedRequest["model"])
	})
	
	t.Run("Model Not Found", func(t *testing.T) {
		// Request with unknown model
		requestBody := map[string]interface{}{
			"model": "unknown-model",
			"messages": []map[string]interface{}{
				{"role": "user", "content": "Hello"},
			},
		}
		
		resp, _ := makeRequest(t, "POST", "/v1/messages", requestBody, nil)
		// Should still route to default provider
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
	
	t.Run("Invalid Provider", func(t *testing.T) {
		// Request with invalid provider
		requestBody := map[string]interface{}{
			"model": "invalid-provider,some-model",
			"messages": []map[string]interface{}{
				{"role": "user", "content": "Hello"},
			},
		}
		
		resp, body := makeRequest(t, "POST", "/v1/messages", requestBody, nil)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.Contains(t, string(body), "provider not found")
	})
}