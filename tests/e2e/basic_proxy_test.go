package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicProxy(t *testing.T) {
	// Use isolated test environment
	env := newTestEnv(t)
	
	t.Run("Health Check", func(t *testing.T) {
		resp, body := env.request("GET", "/health", nil, nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		
		var health map[string]interface{}
		err := json.Unmarshal(body, &health)
		require.NoError(t, err)
		
		assert.Equal(t, "ok", health["status"])
		assert.NotEmpty(t, health["timestamp"])
		
		// Check providers
		providers, ok := health["providers"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, float64(1), providers["total"])
	})
	
	t.Run("Status Check", func(t *testing.T) {
		resp, body := env.request("GET", "/status", nil, nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		
		var status map[string]interface{}
		err := json.Unmarshal(body, &status)
		require.NoError(t, err)
		
		assert.NotEmpty(t, status["status"])
		assert.NotEmpty(t, status["timestamp"])
		assert.NotEmpty(t, status["proxy"])
		assert.NotEmpty(t, status["provider"])
	})
	
	t.Run("Simple Message Request", func(t *testing.T) {
		// Clear any previous requests (like health checks)
		env.mock.ClearRequests()
		
		// Set mock response
		env.mock.SetResponse("/v1/messages", mockResponse{
			Status: 200,
			Body: map[string]interface{}{
				"id":   "msg_test_123",
				"type": "message",
				"role": "assistant",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": "Hello from mock provider!",
					},
				},
				"model": "mock-model",
			},
		})
		
		// Make request
		requestBody := map[string]interface{}{
			"model": "mock-model",
			"messages": []map[string]interface{}{
				{
					"role":    "user",
					"content": "Hello",
				},
			},
		}
		
		resp, body := env.request("POST", "/v1/messages", requestBody, nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		
		var response map[string]interface{}
		err := json.Unmarshal(body, &response)
		require.NoError(t, err)
		
		assert.Equal(t, "msg_test_123", response["id"])
		assert.Equal(t, "message", response["type"])
		assert.Equal(t, "assistant", response["role"])
		
		// Verify mock provider received the request
		requests := env.mock.GetRequests()
		assert.Len(t, requests, 1)
		assert.Equal(t, "/v1/messages", requests[0].Path)
		assert.Equal(t, "POST", requests[0].Method)
		assert.Contains(t, requests[0].Headers.Get("Authorization"), "mock-key")
	})
	
	t.Run("Request with System Message", func(t *testing.T) {
		env.mock.ClearRequests()
		
		requestBody := map[string]interface{}{
			"model": "mock-model",
			"system": "You are a helpful assistant.",
			"messages": []map[string]interface{}{
				{
					"role":    "user",
					"content": "What is 2+2?",
				},
			},
		}
		
		resp, _ := env.request("POST", "/v1/messages", requestBody, nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		
		// Verify request was forwarded
		requests := env.mock.GetRequests()
		require.Len(t, requests, 1)
		
		var capturedRequest map[string]interface{}
		err := json.Unmarshal(requests[0].Body, &capturedRequest)
		require.NoError(t, err)
		
		// System message should be passed through as-is
		assert.Equal(t, "You are a helpful assistant.", capturedRequest["system"])
	})
	
	t.Run("Request with Max Tokens", func(t *testing.T) {
		env.mock.ClearRequests()
		
		requestBody := map[string]interface{}{
			"model": "mock-model",
			"messages": []map[string]interface{}{
				{
					"role":    "user",
					"content": "Tell me a story",
				},
			},
			"max_tokens": 100,
		}
		
		resp, _ := env.request("POST", "/v1/messages", requestBody, nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		
		// Verify max_tokens was passed through
		requests := env.mock.GetRequests()
		require.Len(t, requests, 1)
		
		var capturedRequest map[string]interface{}
		err := json.Unmarshal(requests[0].Body, &capturedRequest)
		require.NoError(t, err)
		
		assert.Equal(t, float64(100), capturedRequest["max_tokens"])
	})
}