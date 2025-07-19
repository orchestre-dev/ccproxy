package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorHandling(t *testing.T) {
	// Setup
	configPath := createTestConfig(t)
	stopCCProxy := startCCProxy(t, configPath)
	defer stopCCProxy()
	
	mockProvider := startMockProvider(t)
	defer mockProvider.Stop()
	
	t.Run("Invalid Request Body", func(t *testing.T) {
		// Send invalid JSON
		headers := map[string]string{
			"Authorization": "Bearer " + testAPIKey,
			"Content-Type":  "application/json",
		}
		
		resp, body := makeRequest(t, "POST", "/v1/messages", "invalid json", headers)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		
		var errorResp map[string]interface{}
		err := json.Unmarshal(body, &errorResp)
		require.NoError(t, err)
		
		assert.NotNil(t, errorResp["error"])
	})
	
	t.Run("Missing Messages", func(t *testing.T) {
		// Request without messages field
		requestBody := map[string]interface{}{
			"model": "mock-model",
			// No messages field
		}
		
		resp, body := makeRequest(t, "POST", "/v1/messages", requestBody, nil)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		
		var errorResp map[string]interface{}
		err := json.Unmarshal(body, &errorResp)
		require.NoError(t, err)
		
		assert.NotNil(t, errorResp["error"])
		errorObj := errorResp["error"].(map[string]interface{})
		assert.Contains(t, errorObj["message"], "messages")
	})
	
	t.Run("Provider Error", func(t *testing.T) {
		mockProvider.ClearRequests()
		
		// Set up error response from provider
		mockProvider.SetResponse("/v1/messages", mockResponse{
			Status: 500,
			Body: map[string]interface{}{
				"error": map[string]interface{}{
					"type":    "internal_error",
					"message": "Provider internal error",
				},
			},
		})
		
		requestBody := map[string]interface{}{
			"model": "mock-model",
			"messages": []map[string]interface{}{
				{"role": "user", "content": "Hello"},
			},
		}
		
		resp, body := makeRequest(t, "POST", "/v1/messages", requestBody, nil)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		
		var errorResp map[string]interface{}
		err := json.Unmarshal(body, &errorResp)
		require.NoError(t, err)
		
		assert.NotNil(t, errorResp["error"])
		errorObj := errorResp["error"].(map[string]interface{})
		assert.Equal(t, "internal_error", errorObj["type"])
		assert.Equal(t, "Provider internal error", errorObj["message"])
	})
	
	t.Run("Rate Limit Error", func(t *testing.T) {
		mockProvider.ClearRequests()
		
		// Set up rate limit response from provider
		mockProvider.SetResponse("/v1/messages", mockResponse{
			Status: 429,
			Headers: map[string]string{
				"Retry-After": "60",
			},
			Body: map[string]interface{}{
				"error": map[string]interface{}{
					"type":    "rate_limit_error",
					"message": "Rate limit exceeded",
				},
			},
		})
		
		requestBody := map[string]interface{}{
			"model": "mock-model",
			"messages": []map[string]interface{}{
				{"role": "user", "content": "Hello"},
			},
		}
		
		resp, body := makeRequest(t, "POST", "/v1/messages", requestBody, nil)
		assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
		assert.Equal(t, "60", resp.Header.Get("Retry-After"))
		
		var errorResp map[string]interface{}
		err := json.Unmarshal(body, &errorResp)
		require.NoError(t, err)
		
		assert.NotNil(t, errorResp["error"])
		errorObj := errorResp["error"].(map[string]interface{})
		assert.Equal(t, "rate_limit_error", errorObj["type"])
	})
	
	t.Run("Invalid Endpoint", func(t *testing.T) {
		resp, body := makeRequest(t, "POST", "/v1/invalid-endpoint", map[string]interface{}{}, nil)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		assert.Contains(t, string(body), "not found")
	})
}