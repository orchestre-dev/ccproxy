package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/musistudio/ccproxy/internal/config"
	"github.com/musistudio/ccproxy/internal/errors"
	"github.com/musistudio/ccproxy/internal/providers"
	"github.com/musistudio/ccproxy/internal/transformer"
	"github.com/musistudio/ccproxy/internal/utils"
)

// Mock provider service
type mockProviderService struct {
	providers map[string]*config.Provider
	health    map[string]*providers.HealthStatus
	stats     map[string]*providers.ProviderStats
}

func newMockProviderService() *mockProviderService {
	return &mockProviderService{
		providers: map[string]*config.Provider{
			"test-provider": {
				Name:       "test-provider",
				APIBaseURL: "https://api.test.com",
				APIKey:     "test-key",
				Models:     []string{"test-model", "streaming-model"},
				Enabled:    true,
			},
			"disabled-provider": {
				Name:       "disabled-provider",
				APIBaseURL: "https://api.disabled.com",
				APIKey:     "disabled-key",
				Models:     []string{"test-model"},
				Enabled:    false,
			},
		},
		health: map[string]*providers.HealthStatus{
			"test-provider": {
				Healthy:      true,
				LastCheck:    time.Now(),
				ResponseTime: 100 * time.Millisecond,
			},
			"disabled-provider": {
				Healthy:      false,
				LastCheck:    time.Now(),
				ErrorMessage: "Provider disabled",
			},
		},
		stats: map[string]*providers.ProviderStats{
			"test-provider": {
				TotalRequests:      100,
				SuccessfulRequests: 95,
				FailedRequests:     5,
				AverageLatency:     150 * time.Millisecond,
				LastUsed:           time.Now(),
			},
		},
	}
}

func (m *mockProviderService) GetProvider(name string) (*config.Provider, error) {
	if p, ok := m.providers[name]; ok {
		return p, nil
	}
	return nil, errors.ErrNotFound("provider")
}

func (m *mockProviderService) SelectProvider(criteria providers.SelectionCriteria) (*config.Provider, error) {
	for _, p := range m.providers {
		if !p.Enabled {
			continue
		}
		for _, model := range p.Models {
			if model == criteria.Model {
				return p, nil
			}
		}
	}
	return nil, errors.ErrNotFound("provider for model " + criteria.Model)
}

func (m *mockProviderService) RecordRequest(provider string, success bool, latency time.Duration) {
	if stats, ok := m.stats[provider]; ok {
		stats.TotalRequests++
		if success {
			stats.SuccessfulRequests++
		} else {
			stats.FailedRequests++
		}
		stats.LastUsed = time.Now()
	}
}

// Mock transformer service
type mockTransformerService struct {
	transformers map[string]transformer.Transformer
}

func newMockTransformerService() *mockTransformerService {
	return &mockTransformerService{
		transformers: map[string]transformer.Transformer{
			"test-provider": &mockTransformer{name: "test-provider"},
		},
	}
}

func (m *mockTransformerService) GetTransformer(provider string) (transformer.Transformer, error) {
	if t, ok := m.transformers[provider]; ok {
		return t, nil
	}
	return nil, errors.ErrNotFound("transformer for provider " + provider)
}

// Mock transformer
type mockTransformer struct {
	name                    string
	transformRequestError   error
	transformResponseError  error
	transformStreamingError error
	simulateStreamingData   bool
}

func (m *mockTransformer) TransformRequest(req *http.Request, provider *config.Provider) (*http.Request, error) {
	if m.transformRequestError != nil {
		return nil, m.transformRequestError
	}
	req.URL.Host = "transformed.api.com"
	req.Header.Set("Authorization", "Bearer "+provider.APIKey)
	return req, nil
}

func (m *mockTransformer) TransformResponse(resp *http.Response, provider *config.Provider) (*http.Response, error) {
	if m.transformResponseError != nil {
		return nil, m.transformResponseError
	}
	// Transform response body
	body := `{"transformed": true, "provider": "` + provider.Name + `"}`
	resp.Body = io.NopCloser(strings.NewReader(body))
	resp.ContentLength = int64(len(body))
	return resp, nil
}

func (m *mockTransformer) TransformStreamingResponse(ctx context.Context, reader io.Reader, writer io.Writer, provider *config.Provider) error {
	if m.transformStreamingError != nil {
		return m.transformStreamingError
	}
	
	if m.simulateStreamingData {
		// Simulate streaming data
		events := []string{
			"data: {\"chunk\": 1}\n\n",
			"data: {\"chunk\": 2}\n\n",
			"data: [DONE]\n\n",
		}
		
		for _, event := range events {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if _, err := writer.Write([]byte(event)); err != nil {
					return err
				}
				time.Sleep(10 * time.Millisecond) // Simulate delay
			}
		}
	}
	
	return nil
}

// Mock HTTP client transport
type mockTransport struct {
	response *http.Response
	err      error
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.response != nil {
		return m.response, nil
	}
	// Default response
	return &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(`{"message": "success"}`)),
		Request:    req,
	}, nil
}

func TestNewPipeline(t *testing.T) {
	cfg := &config.Config{
		Routes: map[string]config.Route{
			"default": {Provider: "test-provider", Model: "test-model"},
		},
	}
	
	p := NewPipeline(cfg, nil, nil)
	if p == nil {
		t.Fatal("Expected pipeline to be created")
	}
	
	if p.config != cfg {
		t.Error("Expected config to be set")
	}
}

func TestProcessRequest(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		headers        map[string]string
		mockResponse   *http.Response
		mockError      error
		transformError error
		expectedStatus int
		expectedBody   string
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:   "successful request",
			method: "POST",
			path:   "/v1/messages",
			body: map[string]interface{}{
				"model":    "test-model",
				"messages": []map[string]string{{"role": "user", "content": "Hello"}},
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if resp["transformed"] != true {
					t.Error("Expected transformed response")
				}
			},
		},
		{
			name:   "missing model",
			method: "POST",
			path:   "/v1/messages",
			body: map[string]interface{}{
				"messages": []map[string]string{{"role": "user", "content": "Hello"}},
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "model is required",
		},
		{
			name:   "invalid model",
			method: "POST",
			path:   "/v1/messages",
			body: map[string]interface{}{
				"model":    "invalid-model",
				"messages": []map[string]string{{"role": "user", "content": "Hello"}},
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "provider for model invalid-model not found",
		},
		{
			name:   "transform request error",
			method: "POST",
			path:   "/v1/messages",
			body: map[string]interface{}{
				"model":    "test-model",
				"messages": []map[string]string{{"role": "user", "content": "Hello"}},
			},
			transformError: errors.New(errors.ErrorTypeInternal, "Transform failed"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Transform failed",
		},
		{
			name:   "provider error",
			method: "POST",
			path:   "/v1/messages",
			body: map[string]interface{}{
				"model":    "test-model",
				"messages": []map[string]string{{"role": "user", "content": "Hello"}},
			},
			mockError:      errors.New(errors.ErrorTypeServiceUnavailable, "Provider unavailable"),
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   "Provider unavailable",
		},
		{
			name:   "rate limit response",
			method: "POST",
			path:   "/v1/messages",
			body: map[string]interface{}{
				"model":    "test-model",
				"messages": []map[string]string{{"role": "user", "content": "Hello"}},
			},
			mockResponse: &http.Response{
				StatusCode: 429,
				Header:     http.Header{"Retry-After": []string{"60"}},
				Body:       io.NopCloser(strings.NewReader(`{"error": {"message": "Rate limited"}}`)),
			},
			expectedStatus: http.StatusTooManyRequests,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				if w.Header().Get("Retry-After") != "60" {
					t.Error("Expected Retry-After header to be preserved")
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test environment
			gin.SetMode(gin.TestMode)
			
			cfg := &config.Config{
				Routes: map[string]config.Route{
					"default": {Provider: "test-provider", Model: "test-model"},
				},
			}
			
			providerService := newMockProviderService()
			transformerService := newMockTransformerService()
			
			// Configure mock transformer
			if tt.transformError != nil {
				mockTrans := transformerService.transformers["test-provider"].(*mockTransformer)
				mockTrans.transformRequestError = tt.transformError
			}
			
			p := NewPipeline(cfg, providerService, transformerService)
			
			// Configure mock HTTP client
			transport := &mockTransport{
				response: tt.mockResponse,
				err:      tt.mockError,
			}
			p.httpClient.Transport = transport
			
			// Create request
			var body io.Reader
			if tt.body != nil {
				bodyBytes, _ := json.Marshal(tt.body)
				body = bytes.NewReader(bodyBytes)
			}
			
			req := httptest.NewRequest(tt.method, tt.path, body)
			req.Header.Set("Content-Type", "application/json")
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			
			// Create response recorder
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			
			// Process request
			p.ProcessRequest(c)
			
			// Check status
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			
			// Check body
			if tt.expectedBody != "" && !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("Expected body to contain %q, got %q", tt.expectedBody, w.Body.String())
			}
			
			// Custom checks
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestProcessStreamingRequest(t *testing.T) {
	tests := []struct {
		name              string
		body              interface{}
		simulateStreaming bool
		streamingError    error
		contextCancel     bool
		expectedStatus    int
		expectedEvents    []string
	}{
		{
			name: "successful streaming",
			body: map[string]interface{}{
				"model":    "streaming-model",
				"messages": []map[string]string{{"role": "user", "content": "Hello"}},
				"stream":   true,
			},
			simulateStreaming: true,
			expectedStatus:    http.StatusOK,
			expectedEvents:    []string{"chunk\": 1", "chunk\": 2", "[DONE]"},
		},
		{
			name: "streaming error",
			body: map[string]interface{}{
				"model":    "streaming-model",
				"messages": []map[string]string{{"role": "user", "content": "Hello"}},
				"stream":   true,
			},
			streamingError: errors.New(errors.ErrorTypeInternal, "Streaming failed"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "context cancellation",
			body: map[string]interface{}{
				"model":    "streaming-model",
				"messages": []map[string]string{{"role": "user", "content": "Hello"}},
				"stream":   true,
			},
			simulateStreaming: true,
			contextCancel:     true,
			expectedStatus:    http.StatusOK,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			
			cfg := &config.Config{
				Routes: map[string]config.Route{
					"default": {Provider: "test-provider", Model: "streaming-model"},
				},
			}
			
			providerService := newMockProviderService()
			transformerService := newMockTransformerService()
			
			// Configure mock transformer
			mockTrans := transformerService.transformers["test-provider"].(*mockTransformer)
			mockTrans.simulateStreamingData = tt.simulateStreaming
			mockTrans.transformStreamingError = tt.streamingError
			
			p := NewPipeline(cfg, providerService, transformerService)
			
			// Configure mock HTTP client for streaming response
			streamingBody := `data: {"test": "stream"}\n\ndata: [DONE]\n\n`
			transport := &mockTransport{
				response: &http.Response{
					StatusCode: 200,
					Header: http.Header{
						"Content-Type": []string{"text/event-stream"},
					},
					Body: io.NopCloser(strings.NewReader(streamingBody)),
				},
			}
			p.httpClient.Transport = transport
			
			// Create request
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			
			// Create response recorder that supports flushing
			w := &flushableResponseRecorder{
				ResponseRecorder: httptest.NewRecorder(),
				flushed:          false,
			}
			
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			
			// Handle context cancellation
			if tt.contextCancel {
				ctx, cancel := context.WithCancel(context.Background())
				c.Request = c.Request.WithContext(ctx)
				go func() {
					time.Sleep(50 * time.Millisecond)
					cancel()
				}()
			}
			
			// Process request
			p.ProcessRequest(c)
			
			// Check status
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			
			// Check streaming events
			if tt.expectedEvents != nil && tt.streamingError == nil {
				body := w.Body.String()
				for _, event := range tt.expectedEvents {
					if !strings.Contains(body, event) {
						t.Errorf("Expected body to contain %q, got %q", event, body)
					}
				}
			}
		})
	}
}

// flushableResponseRecorder implements http.Flusher
type flushableResponseRecorder struct {
	*httptest.ResponseRecorder
	flushed bool
}

func (f *flushableResponseRecorder) Flush() {
	f.flushed = true
}

func TestExtractModel(t *testing.T) {
	tests := []struct {
		name          string
		body          interface{}
		expectedModel string
		expectedError bool
	}{
		{
			name:          "model in body",
			body:          map[string]interface{}{"model": "test-model"},
			expectedModel: "test-model",
		},
		{
			name:          "missing model",
			body:          map[string]interface{}{"messages": []string{"hello"}},
			expectedError: true,
		},
		{
			name:          "empty model",
			body:          map[string]interface{}{"model": ""},
			expectedError: true,
		},
		{
			name:          "non-string model",
			body:          map[string]interface{}{"model": 123},
			expectedError: true,
		},
		{
			name:          "nil body",
			body:          nil,
			expectedError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := extractModel(tt.body)
			
			if tt.expectedError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if model != tt.expectedModel {
					t.Errorf("Expected model %q, got %q", tt.expectedModel, model)
				}
			}
		})
	}
}

func TestIsStreamingRequest(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		expectedStream bool
	}{
		{
			name:           "streaming enabled",
			body:           map[string]interface{}{"stream": true},
			expectedStream: true,
		},
		{
			name:           "streaming disabled",
			body:           map[string]interface{}{"stream": false},
			expectedStream: false,
		},
		{
			name:           "no stream field",
			body:           map[string]interface{}{"model": "test"},
			expectedStream: false,
		},
		{
			name:           "non-bool stream",
			body:           map[string]interface{}{"stream": "true"},
			expectedStream: false,
		},
		{
			name:           "nil body",
			body:           nil,
			expectedStream: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isStreamingRequest(tt.body)
			if result != tt.expectedStream {
				t.Errorf("Expected %v, got %v", tt.expectedStream, result)
			}
		})
	}
}

func TestValidateRequest(t *testing.T) {
	tests := []struct {
		name          string
		body          []byte
		expectedError bool
		errorContains string
	}{
		{
			name: "valid request",
			body: []byte(`{"model": "test", "messages": [{"role": "user", "content": "Hello"}]}`),
		},
		{
			name:          "empty body",
			body:          []byte{},
			expectedError: true,
			errorContains: "Request body is required",
		},
		{
			name:          "invalid JSON",
			body:          []byte(`{"model": "test", invalid json`),
			expectedError: true,
			errorContains: "Invalid JSON",
		},
		{
			name:          "missing model",
			body:          []byte(`{"messages": [{"role": "user", "content": "Hello"}]}`),
			expectedError: true,
			errorContains: "model is required",
		},
		{
			name:          "empty model",
			body:          []byte(`{"model": "", "messages": []}`),
			expectedError: true,
			errorContains: "model is required",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pipeline{}
			err := p.validateRequest(tt.body)
			
			if tt.expectedError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestCopyHeaders(t *testing.T) {
	tests := []struct {
		name           string
		sourceHeaders  map[string][]string
		expectedCopied []string
		notCopied      []string
	}{
		{
			name: "copy allowed headers",
			sourceHeaders: map[string][]string{
				"Content-Type":          {"application/json"},
				"X-Request-Id":          {"123"},
				"Accept":                {"*/*"},
				"Accept-Encoding":       {"gzip"},
				"User-Agent":            {"test-agent"},
				"Authorization":         {"Bearer token"},
				"Content-Length":        {"100"},
				"X-Custom-Header":       {"custom"},
				"Anthropic-Beta":        {"beta-feature"},
				"Anthropic-Version":     {"2023-01-01"},
				"X-Stainless-Os":        {"linux"},
				"X-Stainless-Retry-Count": {"2"},
			},
			expectedCopied: []string{
				"Content-Type",
				"X-Request-Id",
				"Accept",
				"Accept-Encoding",
				"User-Agent",
				"X-Custom-Header",
				"Anthropic-Beta",
				"Anthropic-Version",
				"X-Stainless-Os",
				"X-Stainless-Retry-Count",
			},
			notCopied: []string{
				"Authorization",
				"Content-Length",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := make(http.Header)
			for k, v := range tt.sourceHeaders {
				source[k] = v
			}
			
			dest := make(http.Header)
			copyHeaders(source, dest)
			
			// Check expected headers were copied
			for _, h := range tt.expectedCopied {
				if dest.Get(h) == "" {
					t.Errorf("Expected header %q to be copied", h)
				}
			}
			
			// Check headers that shouldn't be copied
			for _, h := range tt.notCopied {
				if dest.Get(h) != "" {
					t.Errorf("Header %q should not be copied", h)
				}
			}
		})
	}
}

func TestRecordMetrics(t *testing.T) {
	providerService := newMockProviderService()
	logger := utils.GetLogger()
	
	tests := []struct {
		name     string
		provider string
		success  bool
		latency  time.Duration
	}{
		{
			name:     "successful request",
			provider: "test-provider",
			success:  true,
			latency:  100 * time.Millisecond,
		},
		{
			name:     "failed request",
			provider: "test-provider",
			success:  false,
			latency:  50 * time.Millisecond,
		},
		{
			name:     "unknown provider",
			provider: "unknown",
			success:  true,
			latency:  200 * time.Millisecond,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Record metrics
			recordMetrics(providerService, tt.provider, logger, tt.success, time.Now().Add(-tt.latency))
			
			// For known providers, verify stats were updated
			if tt.provider == "test-provider" {
				stats := providerService.stats["test-provider"]
				if stats.TotalRequests == 0 {
					t.Error("Expected total requests to be updated")
				}
			}
		})
	}
}