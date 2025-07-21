package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/orchestre-dev/ccproxy/internal/config"
	"github.com/orchestre-dev/ccproxy/internal/server"
	testfw "github.com/orchestre-dev/ccproxy/testing"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite provides a comprehensive integration test suite
type IntegrationTestSuite struct {
	suite.Suite
	framework       *testfw.TestFramework
	mockProviders   map[string]*testfw.MockProviderServer
	server          *server.Server
	client          *http.Client
	fixtures        *testfw.Fixtures
	serverURL       string
	suiteReporter   *testfw.TestSuiteReporter
}

// SetupSuite runs once before all tests
func (s *IntegrationTestSuite) SetupSuite() {
	// Create test framework
	s.framework = testfw.NewTestFramework(s.T())
	s.fixtures = testfw.NewFixtures()
	s.mockProviders = make(map[string]*testfw.MockProviderServer)
	
	// Create mock providers
	s.mockProviders["anthropic"] = testfw.NewMockProviderServer()
	s.mockProviders["openai"] = testfw.NewMockProviderServer()
	s.mockProviders["google"] = testfw.NewMockProviderServer()
	s.mockProviders["aws"] = testfw.NewMockProviderServer()
	
	// Update config with mock provider URLs
	cfg := s.framework.GetConfig()
	cfg.Providers = []config.Provider{
		{
			Name:       "anthropic-test",
			APIBaseURL: s.mockProviders["anthropic"].URL(),
			APIKey:     "test-key",
			Enabled:    true,
			Models:     []string{"claude-3-sonnet"},
		},
		{
			Name:       "openai-test",
			APIBaseURL: s.mockProviders["openai"].URL(),
			APIKey:     "test-key",
			Enabled:    true,
			Models:     []string{"gpt-4"},
		},
		{
			Name:       "google-test",
			APIBaseURL: s.mockProviders["google"].URL(),
			APIKey:     "test-key",
			Enabled:    true,
			Models:     []string{"gemini-pro"},
		},
	}
	
	// Start server
	s.server = s.framework.StartServerWithConfig(cfg)
	s.serverURL = fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port)
	
	// Create HTTP client with connection pooling limits
	s.client = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     90 * time.Second,
		},
	}
}

// TearDownSuite runs once after all tests
func (s *IntegrationTestSuite) TearDownSuite() {
	// Shutdown server first
	if s.server != nil {
		if err := s.server.Shutdown(); err != nil {
			s.T().Logf("Error shutting down server: %v", err)
		}
	}
	
	// Close mock providers
	for name, mock := range s.mockProviders {
		s.T().Logf("Closing mock provider: %s", name)
		mock.Close()
	}
	
	// Clear references
	s.mockProviders = nil
	s.server = nil
}

// SetupTest runs before each test
func (s *IntegrationTestSuite) SetupTest() {
	// Clear mock provider requests
	for _, mock := range s.mockProviders {
		mock.ClearRequests()
	}
}

// TestAnthropicMessages tests Anthropic messages endpoint
func (s *IntegrationTestSuite) TestAnthropicMessages() {
	if s.suiteReporter != nil {
		s.suiteReporter.StartTest("TestAnthropicMessages")
		defer func() {
			s.suiteReporter.EndTest("TestAnthropicMessages", !s.T().Failed())
		}()
	}
	
	// Get request fixture
	reqBody, err := s.fixtures.GetRequest("anthropic_messages")
	s.Require().NoError(err)
	
	// Make request
	resp, err := s.makeRequest("POST", "/v1/messages", reqBody)
	s.Require().NoError(err)
	defer resp.Body.Close()
	
	// Assert response
	s.Assert().Equal(http.StatusOK, resp.StatusCode)
	
	// Parse response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	s.Require().NoError(err)
	
	// Verify response structure
	s.Assert().Equal("message", result["type"])
	s.Assert().Equal("assistant", result["role"])
	s.Assert().NotEmpty(result["content"])
	
	// Verify mock was called
	requests := s.mockProviders["anthropic"].GetRequestsForPath("/v1/messages")
	s.Assert().Len(requests, 1)
}

// TestOpenAIChatCompletions tests OpenAI chat completions endpoint
func (s *IntegrationTestSuite) TestOpenAIChatCompletions() {
	// Get request fixture
	reqBody, err := s.fixtures.GetRequest("openai_chat")
	s.Require().NoError(err)
	
	// Make request
	resp, err := s.makeRequest("POST", "/v1/chat/completions", reqBody)
	s.Require().NoError(err)
	defer resp.Body.Close()
	
	// Assert response
	s.Assert().Equal(http.StatusOK, resp.StatusCode)
	
	// Parse response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	s.Require().NoError(err)
	
	// Verify response structure
	s.Assert().Equal("chat.completion", result["object"])
	s.Assert().NotEmpty(result["choices"])
	
	// Verify routing - should use OpenAI provider
	anthropicReqs := s.mockProviders["anthropic"].GetRequestsForPath("/v1/messages")
	openaiReqs := s.mockProviders["openai"].GetRequestsForPath("/v1/chat/completions")
	s.Assert().Len(anthropicReqs, 0, "Should not call Anthropic for OpenAI endpoint")
	s.Assert().Len(openaiReqs, 1, "Should call OpenAI provider")
}

// TestStreamingResponse tests streaming responses
func (s *IntegrationTestSuite) TestStreamingResponse() {
	// Prepare streaming request
	reqBody := map[string]interface{}{
		"model": "claude-3-sonnet-20240229",
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": "Hello!",
			},
		},
		"stream": true,
	}
	
	// Make request
	resp, err := s.makeRequest("POST", "/v1/messages", reqBody)
	s.Require().NoError(err)
	defer resp.Body.Close()
	
	// Assert streaming response
	s.Assert().Equal(http.StatusOK, resp.StatusCode)
	s.Assert().Contains(resp.Header.Get("Content-Type"), "text/event-stream")
	
	// Read streaming chunks
	chunks := make([]string, 0)
	reader := io.Reader(resp.Body)
	buf := make([]byte, 1024)
	
	for {
		n, err := reader.Read(buf)
		if err == io.EOF {
			break
		}
		s.Require().NoError(err)
		chunks = append(chunks, string(buf[:n]))
	}
	
	// Verify we received chunks
	s.Assert().NotEmpty(chunks)
}

// TestProviderFailover tests failover between providers
func (s *IntegrationTestSuite) TestProviderFailover() {
	// Set primary provider to return error
	s.mockProviders["anthropic"].SetErrorRate(1.0)
	
	// Make request
	reqBody, _ := s.fixtures.GetRequest("anthropic_messages")
	resp, err := s.makeRequest("POST", "/v1/messages", reqBody)
	s.Require().NoError(err)
	defer resp.Body.Close()
	
	// Should still succeed with fallback provider
	s.Assert().Equal(http.StatusOK, resp.StatusCode)
	
	// Verify both providers were called
	anthropicReqs := s.mockProviders["anthropic"].GetRequestsForPath("/v1/messages")
	openaiReqs := s.mockProviders["openai"].GetRequestsForPath("/v1/chat/completions")
	
	s.Assert().Len(anthropicReqs, 1, "Should try primary provider first")
	s.Assert().Len(openaiReqs, 1, "Should fallback to secondary provider")
}

// TestRateLimiting tests rate limiting functionality
func (s *IntegrationTestSuite) TestRateLimiting() {
	// Make multiple rapid requests
	reqBody, _ := s.fixtures.GetRequest("anthropic_messages")
	
	successCount := 0
	rateLimitCount := 0
	
	for i := 0; i < 20; i++ {
		resp, err := s.makeRequest("POST", "/v1/messages", reqBody)
		s.Require().NoError(err)
		
		if resp.StatusCode == http.StatusOK {
			successCount++
		} else if resp.StatusCode == http.StatusTooManyRequests {
			rateLimitCount++
		}
		
		resp.Body.Close()
	}
	
	// Should have some rate limited requests
	s.Assert().Greater(rateLimitCount, 0, "Should have rate limited some requests")
	s.Assert().Greater(successCount, 0, "Should have allowed some requests")
}

// TestAuthentication tests authentication middleware
func (s *IntegrationTestSuite) TestAuthentication() {
	// Test without auth header
	resp, err := s.makeRequestWithAuth("POST", "/v1/messages", nil, "")
	s.Require().NoError(err)
	defer resp.Body.Close()
	
	s.Assert().Equal(http.StatusUnauthorized, resp.StatusCode)
	
	// Test with invalid auth
	resp2, err := s.makeRequestWithAuth("POST", "/v1/messages", nil, "Bearer invalid-token")
	s.Require().NoError(err)
	defer resp2.Body.Close()
	
	s.Assert().Equal(http.StatusUnauthorized, resp2.StatusCode)
	
	// Test with valid auth
	reqBody, _ := s.fixtures.GetRequest("anthropic_messages")
	resp3, err := s.makeRequestWithAuth("POST", "/v1/messages", reqBody, "Bearer test-api-key")
	s.Require().NoError(err)
	defer resp3.Body.Close()
	
	s.Assert().Equal(http.StatusOK, resp3.StatusCode)
}

// TestConcurrentRequests tests concurrent request handling
func (s *IntegrationTestSuite) TestConcurrentRequests() {
	// Prepare request
	reqBody, _ := s.fixtures.GetRequest("anthropic_messages")
	
	// Make concurrent requests
	concurrency := 10
	results := make(chan int, concurrency)
	
	for i := 0; i < concurrency; i++ {
		go func() {
			resp, err := s.makeRequest("POST", "/v1/messages", reqBody)
			if err != nil {
				results <- 0
				return
			}
			defer resp.Body.Close()
			results <- resp.StatusCode
		}()
	}
	
	// Collect results
	successCount := 0
	for i := 0; i < concurrency; i++ {
		status := <-results
		if status == http.StatusOK {
			successCount++
		}
	}
	
	// All requests should succeed
	s.Assert().Equal(concurrency, successCount)
}

// TestLargePayload tests handling of large payloads
func (s *IntegrationTestSuite) TestLargePayload() {
	// Create large message
	largeContent := s.fixtures.GenerateLargeMessage(10000) // ~10k tokens
	
	reqBody := map[string]interface{}{
		"model": "claude-3-sonnet-20240229",
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": largeContent,
			},
		},
		"max_tokens": 100,
	}
	
	// Make request
	resp, err := s.makeRequest("POST", "/v1/messages", reqBody)
	s.Require().NoError(err)
	defer resp.Body.Close()
	
	// Should handle large payload
	s.Assert().Equal(http.StatusOK, resp.StatusCode)
}

// TestHealthCheck tests health check endpoint
func (s *IntegrationTestSuite) TestHealthCheck() {
	resp, err := s.makeRequest("GET", "/health", nil)
	s.Require().NoError(err)
	defer resp.Body.Close()
	
	s.Assert().Equal(http.StatusOK, resp.StatusCode)
	
	// Parse response
	var health map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&health)
	s.Require().NoError(err)
	
	s.Assert().Equal("healthy", health["status"])
	s.Assert().NotEmpty(health["version"])
}

// TestMetrics tests metrics endpoint
func (s *IntegrationTestSuite) TestMetrics() {
	// Make some requests first
	reqBody, _ := s.fixtures.GetRequest("anthropic_messages")
	for i := 0; i < 5; i++ {
		resp, _ := s.makeRequest("POST", "/v1/messages", reqBody)
		if resp != nil {
			resp.Body.Close()
		}
	}
	
	// Get metrics
	resp, err := s.makeRequest("GET", "/metrics", nil)
	s.Require().NoError(err)
	defer resp.Body.Close()
	
	s.Assert().Equal(http.StatusOK, resp.StatusCode)
	
	// Parse metrics
	var metrics map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&metrics)
	s.Require().NoError(err)
	
	// Verify metrics exist
	s.Assert().NotEmpty(metrics["requests"])
	s.Assert().NotEmpty(metrics["providers"])
}

// Helper methods

func (s *IntegrationTestSuite) makeRequest(method, path string, body interface{}) (*http.Response, error) {
	return s.makeRequestWithAuth(method, path, body, "Bearer test-api-key")
}

func (s *IntegrationTestSuite) makeRequestWithAuth(method, path string, body interface{}, auth string) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}
	
	req, err := http.NewRequest(method, s.serverURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	
	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	
	return s.client.Do(req)
}

// TestSuite runs the integration test suite
func TestIntegrationSuite(t *testing.T) {
	// Create suite reporter to track all tests
	testNames := []string{
		"TestAnthropicMessages",
		"TestOpenAIChatCompletions", 
		"TestStreamingResponse",
		"TestProviderFailover",
		"TestRateLimiting",
		"TestAuthentication",
		"TestConcurrentRequests",
		"TestLargePayload",
		"TestHealthCheck",
		"TestMetrics",
	}
	
	reporter := testfw.NewTestSuiteReporter(t, "IntegrationSuite", testNames)
	
	// Hook into suite to report progress
	s := new(IntegrationTestSuite)
	s.suiteReporter = reporter
	
	suite.Run(t, s)
	
	reporter.Complete()
}