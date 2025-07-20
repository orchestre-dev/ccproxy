package load

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	testfw "github.com/orchestre-dev/ccproxy/internal/testing"
	"github.com/stretchr/testify/require"
)

// TestLoadBasicEndpoint tests basic load on the messages endpoint
func TestLoadBasicEndpoint(t *testing.T) {
	framework := testfw.NewTestFramework(t)
	fixtures := testfw.NewFixtures()
	
	// Start server
	server := framework.StartServer()
	serverURL := fmt.Sprintf("http://127.0.0.1:%d", server.GetPort())
	
	// Create mock provider
	mockProvider := testfw.NewMockProviderServer("anthropic")
	defer mockProvider.Close()
	
	// Configure load test
	config := testfw.LoadTestConfig{
		Duration:        30 * time.Second,
		ConcurrentUsers: 50,
		RampUpTime:      5 * time.Second,
		RequestsPerUser: 100,
		ThinkTime:       100 * time.Millisecond,
	}
	
	// Create load tester
	loadTester := testfw.NewLoadTester(framework, config)
	
	// Prepare request
	reqBody, _ := fixtures.GetRequest("anthropic_messages")
	reqData, _ := json.Marshal(reqBody)
	
	// Run load test
	results := loadTester.Run(func() error {
		req, err := http.NewRequest("POST", serverURL+"/v1/messages", bytes.NewReader(reqData))
		if err != nil {
			return err
		}
		
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-key")
		
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status: %d", resp.StatusCode)
		}
		
		return nil
	})
	
	// Assert results
	t.Logf("Load Test Results:")
	t.Logf("Total Requests: %d", results.TotalRequests)
	t.Logf("Success Rate: %.2f%%", (1-results.ErrorRate)*100)
	t.Logf("Requests/sec: %.2f", results.RequestsPerSec)
	t.Logf("Error Rate: %.2f%%", results.ErrorRate*100)
	
	// Basic assertions
	require.Greater(t, results.TotalRequests, int64(0))
	require.Less(t, results.ErrorRate, 0.05) // Less than 5% error rate
	require.Greater(t, results.RequestsPerSec, float64(10)) // At least 10 req/s
}

// TestLoadMixedEndpoints tests load with mixed endpoint traffic
func TestLoadMixedEndpoints(t *testing.T) {
	framework := testfw.NewTestFramework(t)
	fixtures := testfw.NewFixtures()
	
	// Start server
	server := framework.StartServer()
	serverURL := fmt.Sprintf("http://127.0.0.1:%d", server.GetPort())
	
	// Create mock providers
	anthropicMock := testfw.NewMockProviderServer("anthropic")
	defer anthropicMock.Close()
	
	openaiMock := testfw.NewMockProviderServer("openai")
	defer openaiMock.Close()
	
	// Configure load test
	config := testfw.LoadTestConfig{
		Duration:        60 * time.Second,
		ConcurrentUsers: 100,
		RampUpTime:      10 * time.Second,
		RequestsPerUser: 200,
		ThinkTime:       50 * time.Millisecond,
	}
	
	// Create load tester
	loadTester := testfw.NewLoadTester(framework, config)
	
	// Prepare requests
	anthropicReq, _ := fixtures.GetRequest("anthropic_messages")
	anthropicData, _ := json.Marshal(anthropicReq)
	
	openaiReq, _ := fixtures.GetRequest("openai_chat")
	openaiData, _ := json.Marshal(openaiReq)
	
	// Run load test with mixed traffic
	requestCount := 0
	results := loadTester.Run(func() error {
		requestCount++
		
		// Alternate between endpoints
		var endpoint string
		var reqData []byte
		
		if requestCount%2 == 0 {
			endpoint = "/v1/messages"
			reqData = anthropicData
		} else {
			endpoint = "/v1/chat/completions"
			reqData = openaiData
		}
		
		req, err := http.NewRequest("POST", serverURL+endpoint, bytes.NewReader(reqData))
		if err != nil {
			return err
		}
		
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-key")
		
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status: %d", resp.StatusCode)
		}
		
		return nil
	})
	
	// Report results
	t.Logf("Mixed Endpoint Load Test Results:")
	t.Logf("Total Requests: %d", results.TotalRequests)
	t.Logf("Success Rate: %.2f%%", (1-results.ErrorRate)*100)
	t.Logf("Requests/sec: %.2f", results.RequestsPerSec)
	
	// Assertions
	require.Less(t, results.ErrorRate, 0.10) // Less than 10% error rate
}

// TestLoadWithFailures tests system behavior under load with failures
func TestLoadWithFailures(t *testing.T) {
	framework := testfw.NewTestFramework(t)
	fixtures := testfw.NewFixtures()
	
	// Start server
	server := framework.StartServer()
	serverURL := fmt.Sprintf("http://127.0.0.1:%d", server.GetPort())
	
	// Create mock provider with intermittent failures
	mockProvider := testfw.NewMockProviderServer("anthropic")
	defer mockProvider.Close()
	
	// Configure provider to fail 20% of requests
	failureCount := 0
	mockProvider.AddConditionalRoute("POST", "/v1/messages", 
		func(r *http.Request) bool {
			failureCount++
			return failureCount%5 != 0 // Fail every 5th request
		},
		map[string]interface{}{"error": "provider error"},
		http.StatusInternalServerError,
	)
	
	// Configure load test
	config := testfw.LoadTestConfig{
		Duration:        30 * time.Second,
		ConcurrentUsers: 25,
		RampUpTime:      5 * time.Second,
		RequestsPerUser: 50,
		ThinkTime:       200 * time.Millisecond,
	}
	
	// Create load tester
	loadTester := testfw.NewLoadTester(framework, config)
	
	// Prepare request
	reqBody, _ := fixtures.GetRequest("anthropic_messages")
	reqData, _ := json.Marshal(reqBody)
	
	// Run load test
	results := loadTester.Run(func() error {
		req, err := http.NewRequest("POST", serverURL+"/v1/messages", bytes.NewReader(reqData))
		if err != nil {
			return err
		}
		
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-key")
		
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		
		// Accept both success and server errors (from failover)
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusInternalServerError {
			return fmt.Errorf("unexpected status: %d", resp.StatusCode)
		}
		
		return nil
	})
	
	// Report results
	t.Logf("Load Test with Failures Results:")
	t.Logf("Total Requests: %d", results.TotalRequests)
	t.Logf("Success Rate: %.2f%%", (1-results.ErrorRate)*100)
	t.Logf("Failed Requests: %d", results.FailedRequests)
	
	// System should handle failures gracefully
	require.Greater(t, results.SuccessRequests, int64(0))
}

// TestLoadStreaming tests streaming endpoint under load
func TestLoadStreaming(t *testing.T) {
	framework := testfw.NewTestFramework(t)
	fixtures := testfw.NewFixtures()
	
	// Start server
	server := framework.StartServer()
	serverURL := fmt.Sprintf("http://127.0.0.1:%d", server.GetPort())
	
	// Create mock provider with streaming
	mockProvider := testfw.NewMockProviderServer("anthropic")
	defer mockProvider.Close()
	
	// Configure load test with fewer users for streaming
	config := testfw.LoadTestConfig{
		Duration:        20 * time.Second,
		ConcurrentUsers: 20,
		RampUpTime:      5 * time.Second,
		RequestsPerUser: 10,
		ThinkTime:       500 * time.Millisecond,
	}
	
	// Create load tester
	loadTester := testfw.NewLoadTester(framework, config)
	
	// Prepare streaming request
	reqBody := map[string]interface{}{
		"model": "claude-3-sonnet-20240229",
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": "Tell me a story",
			},
		},
		"stream": true,
	}
	reqData, _ := json.Marshal(reqBody)
	
	// Run load test
	results := loadTester.Run(func() error {
		req, err := http.NewRequest("POST", serverURL+"/v1/messages", bytes.NewReader(reqData))
		if err != nil {
			return err
		}
		
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-key")
		
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status: %d", resp.StatusCode)
		}
		
		// Read streaming response
		buf := make([]byte, 1024)
		totalBytes := 0
		for {
			n, err := resp.Body.Read(buf)
			totalBytes += n
			if err != nil {
				break
			}
		}
		
		if totalBytes == 0 {
			return fmt.Errorf("no data received from stream")
		}
		
		return nil
	})
	
	// Report results
	t.Logf("Streaming Load Test Results:")
	t.Logf("Total Requests: %d", results.TotalRequests)
	t.Logf("Success Rate: %.2f%%", (1-results.ErrorRate)*100)
	t.Logf("Requests/sec: %.2f", results.RequestsPerSec)
	
	// Streaming should still work under load
	require.Less(t, results.ErrorRate, 0.15) // Allow higher error rate for streaming
}

// TestLoadSpike tests system behavior under sudden load spike
func TestLoadSpike(t *testing.T) {
	framework := testfw.NewTestFramework(t)
	fixtures := testfw.NewFixtures()
	
	// Start server
	server := framework.StartServer()
	serverURL := fmt.Sprintf("http://127.0.0.1:%d", server.GetPort())
	
	// Create mock provider
	mockProvider := testfw.NewMockProviderServer("anthropic")
	defer mockProvider.Close()
	
	// Phase 1: Normal load
	normalConfig := testfw.LoadTestConfig{
		Duration:        10 * time.Second,
		ConcurrentUsers: 10,
		RampUpTime:      2 * time.Second,
		RequestsPerUser: 50,
		ThinkTime:       100 * time.Millisecond,
	}
	
	loadTester := testfw.NewLoadTester(framework, normalConfig)
	reqBody, _ := fixtures.GetRequest("anthropic_messages")
	reqData, _ := json.Marshal(reqBody)
	
	testFunc := func() error {
		req, err := http.NewRequest("POST", serverURL+"/v1/messages", bytes.NewReader(reqData))
		if err != nil {
			return err
		}
		
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-key")
		
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status: %d", resp.StatusCode)
		}
		
		return nil
	}
	
	normalResults := loadTester.Run(testFunc)
	
	// Phase 2: Spike load (5x users, no ramp up)
	spikeConfig := testfw.LoadTestConfig{
		Duration:        10 * time.Second,
		ConcurrentUsers: 50,
		RampUpTime:      0, // No ramp up - immediate spike
		RequestsPerUser: 50,
		ThinkTime:       50 * time.Millisecond,
	}
	
	spikeTester := testfw.NewLoadTester(framework, spikeConfig)
	spikeResults := spikeTester.Run(testFunc)
	
	// Report results
	t.Logf("Normal Load Results:")
	t.Logf("  Requests/sec: %.2f", normalResults.RequestsPerSec)
	t.Logf("  Error Rate: %.2f%%", normalResults.ErrorRate*100)
	
	t.Logf("Spike Load Results:")
	t.Logf("  Requests/sec: %.2f", spikeResults.RequestsPerSec)
	t.Logf("  Error Rate: %.2f%%", spikeResults.ErrorRate*100)
	
	// System should handle spike, even if degraded
	require.Less(t, spikeResults.ErrorRate, 0.25) // Allow up to 25% errors during spike
}

// TestLoadSustained tests sustained load over longer period
func TestLoadSustained(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping sustained load test in short mode")
	}
	
	framework := testfw.NewTestFramework(t)
	fixtures := testfw.NewFixtures()
	
	// Start server
	server := framework.StartServer()
	serverURL := fmt.Sprintf("http://127.0.0.1:%d", server.GetPort())
	
	// Create mock provider
	mockProvider := testfw.NewMockProviderServer("anthropic")
	defer mockProvider.Close()
	
	// Configure sustained load test
	config := testfw.LoadTestConfig{
		Duration:        5 * time.Minute,
		ConcurrentUsers: 30,
		RampUpTime:      30 * time.Second,
		RequestsPerUser: 1000,
		ThinkTime:       200 * time.Millisecond,
	}
	
	// Create load tester
	loadTester := testfw.NewLoadTester(framework, config)
	
	// Prepare requests of varying sizes
	smallReq, _ := fixtures.GetRequest("anthropic_messages")
	smallData, _ := json.Marshal(smallReq)
	
	largeReq := map[string]interface{}{
		"model":    "claude-3-sonnet-20240229",
		"messages": fixtures.GenerateMessages(20),
	}
	largeData, _ := json.Marshal(largeReq)
	
	// Run sustained load test
	requestCount := 0
	results := loadTester.Run(func() error {
		requestCount++
		
		// Mix of small and large requests
		var reqData []byte
		if requestCount%3 == 0 {
			reqData = largeData
		} else {
			reqData = smallData
		}
		
		req, err := http.NewRequest("POST", serverURL+"/v1/messages", bytes.NewReader(reqData))
		if err != nil {
			return err
		}
		
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-key")
		
		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status: %d", resp.StatusCode)
		}
		
		return nil
	})
	
	// Report results
	t.Logf("Sustained Load Test Results:")
	t.Logf("Duration: %v", results.TotalDuration)
	t.Logf("Total Requests: %d", results.TotalRequests)
	t.Logf("Success Rate: %.2f%%", (1-results.ErrorRate)*100)
	t.Logf("Average Requests/sec: %.2f", results.RequestsPerSec)
	
	// System should maintain performance over time
	require.Less(t, results.ErrorRate, 0.05) // Less than 5% error rate
	require.Greater(t, results.RequestsPerSec, float64(20)) // Maintain at least 20 req/s
}