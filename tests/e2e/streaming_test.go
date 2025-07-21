package e2e

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStreaming(t *testing.T) {
	// Setup
	configPath := createTestConfig(t)
	stopCCProxy := startCCProxy(t, configPath)
	defer stopCCProxy()
	
	mockProvider := startMockProvider(t)
	defer mockProvider.Stop()
	
	t.Run("Simple Streaming", func(t *testing.T) {
		mockProvider.ClearRequests()
		
		// Set up streaming response
		streamChunks := createMockStreamChunks("Hello", " from", " streaming!")
		mockProvider.SetStreamData("/v1/messages", streamChunks)
		mockProvider.SetResponse("/v1/messages", mockResponse{
			Status: 200,
			Stream: true,
		})
		
		// Make streaming request
		requestBody := map[string]interface{}{
			"model": "mock-model",
			"messages": []map[string]interface{}{
				{"role": "user", "content": "Hello"},
			},
			"stream": true,
		}
		
		client := &http.Client{}
		bodyData, _ := json.Marshal(requestBody)
		req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/v1/messages", ccproxyPort), strings.NewReader(string(bodyData)))
		require.NoError(t, err)
		
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testAPIKey))
		req.Header.Set("Accept", "text/event-stream")
		
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))
		
		// Read streaming response
		scanner := bufio.NewScanner(resp.Body)
		var events []string
		
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data: ") {
				events = append(events, strings.TrimPrefix(line, "data: "))
			}
		}
		
		// Verify we got all events
		assert.GreaterOrEqual(t, len(events), 5) // At least start, content blocks, and stop events
		
		// Verify request was forwarded with stream=true
		requests := mockProvider.GetRequests()
		require.Len(t, requests, 1)
		
		var capturedRequest map[string]interface{}
		err = json.Unmarshal(requests[0].Body, &capturedRequest)
		require.NoError(t, err)
		assert.Equal(t, true, capturedRequest["stream"])
	})
	
	t.Run("Non-Streaming Request", func(t *testing.T) {
		mockProvider.ClearRequests()
		
		// Reset to non-streaming response
		mockProvider.SetResponse("/v1/messages", mockResponse{
			Status: 200,
			Body: map[string]interface{}{
				"id":   "msg_test",
				"type": "message",
				"role": "assistant",
				"content": []map[string]interface{}{
					{"type": "text", "text": "Non-streaming response"},
				},
			},
			Stream: false,
		})
		
		// Make non-streaming request
		requestBody := map[string]interface{}{
			"model": "mock-model",
			"messages": []map[string]interface{}{
				{"role": "user", "content": "Hello"},
			},
			"stream": false,
		}
		
		resp, body := makeRequest(t, "POST", "/v1/messages", requestBody, nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotEqual(t, "text/event-stream", resp.Header.Get("Content-Type"))
		
		var response map[string]interface{}
		err := json.Unmarshal(body, &response)
		require.NoError(t, err)
		
		assert.Equal(t, "msg_test", response["id"])
		assert.Equal(t, "message", response["type"])
	})
}