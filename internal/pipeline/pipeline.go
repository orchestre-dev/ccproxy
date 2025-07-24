package pipeline

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/orchestre-dev/ccproxy/internal/config"
	"github.com/orchestre-dev/ccproxy/internal/converter"
	"github.com/orchestre-dev/ccproxy/internal/performance"
	"github.com/orchestre-dev/ccproxy/internal/providers"
	"github.com/orchestre-dev/ccproxy/internal/proxy"
	"github.com/orchestre-dev/ccproxy/internal/router"
	"github.com/orchestre-dev/ccproxy/internal/transformer"
	"github.com/orchestre-dev/ccproxy/internal/utils"
)

// Pipeline handles the complete request processing flow
type Pipeline struct {
	config             *config.Config
	providerService    *providers.Service
	transformerService *transformer.Service
	router             *router.Router
	httpClient         *http.Client
	streamingProcessor *StreamingProcessor
	performanceMonitor *performance.Monitor
	requestCounter     int64
	messageConverter   *converter.MessageConverter
}

// NewPipeline creates a new request processing pipeline
func NewPipeline(
	cfg *config.Config,
	providerService *providers.Service,
	transformerService *transformer.Service,
	router *router.Router,
) *Pipeline {
	// Create HTTP client with configurable timeout
	timeout := cfg.Performance.RequestTimeout
	if timeout == 0 {
		// Fallback to reasonable default if not configured
		timeout = 30 * time.Second
	}

	// Create proxy configuration
	var proxyConfig *proxy.Config
	if cfg.ProxyURL != "" {
		var err error
		proxyConfig, err = proxy.NewConfig(cfg.ProxyURL)
		if err != nil {
			utils.GetLogger().Warnf("Invalid proxy configuration: %v", err)
		}
	} else {
		// Try to get proxy from environment
		proxyConfig = proxy.GetProxyFromEnvironment()
	}

	// Create HTTP client with proxy support
	httpClient, err := proxy.CreateHTTPClient(proxyConfig, timeout)
	if err != nil {
		utils.GetLogger().Errorf("Failed to create HTTP client: %v", err)
		// Fallback to simple client
		httpClient = &http.Client{Timeout: timeout}
	}

	// Validate proxy if configured
	if proxyConfig != nil {
		if err := proxy.ValidateProxy(httpClient, ""); err != nil {
			utils.GetLogger().Warnf("Proxy validation failed: %v", err)
		}
	}

	return &Pipeline{
		config:             cfg,
		providerService:    providerService,
		transformerService: transformerService,
		router:             router,
		httpClient:         httpClient,
		streamingProcessor: NewStreamingProcessor(transformerService),
		messageConverter:   converter.NewMessageConverter(),
		performanceMonitor: performance.NewMonitor(&performance.PerformanceConfig{
			MetricsEnabled:  true,
			MetricsInterval: 30 * time.Second,
			ResourceLimits: performance.ResourceLimits{
				RequestTimeout:   timeout,
				MaxRequestBodyMB: int(cfg.Performance.MaxRequestBodySize / 1024 / 1024),
			},
		}),
	}
}

// ProcessRequest handles the complete request processing pipeline
func (p *Pipeline) ProcessRequest(ctx context.Context, req *RequestContext) (*ResponseContext, error) {
	// Extract model and count tokens from request
	var routeReq router.Request
	var tokenCount int

	if bodyMap, ok := req.Body.(map[string]interface{}); ok {
		if model, ok := bodyMap["model"].(string); ok {
			routeReq.Model = model
		}
		if thinking, ok := bodyMap["thinking"].(bool); ok {
			routeReq.Thinking = thinking
		}

		// Count tokens
		tokenCount = utils.CountRequestTokens(bodyMap)
	}

	// 1. Route to appropriate model/provider
	routingDecision := p.router.Route(routeReq, tokenCount)

	// 2. Get provider configuration
	selectedProvider, err := p.providerService.GetProvider(routingDecision.Provider)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %s", routingDecision.Provider)
	}

	// 3. Apply route parameters to request body
	requestBody := req.Body
	if len(routingDecision.Parameters) > 0 {
		// Apply parameters to request body
		if bodyMap, ok := requestBody.(map[string]interface{}); ok {
			for key, value := range routingDecision.Parameters {
				// Only set if not already present in the request
				if _, exists := bodyMap[key]; !exists {
					bodyMap[key] = value
				}
			}
			requestBody = bodyMap
		}
	}

	// 4. Get transformer chain for provider
	chain := p.transformerService.GetChainForProvider(routingDecision.Provider)

	// 5. Apply request transformations
	transformedRequest, err := chain.TransformRequestIn(ctx, requestBody, routingDecision.Provider)
	if err != nil {
		return nil, fmt.Errorf("request transformation failed: %w", err)
	}

	// 6. Build HTTP request with transformed data
	httpReq, err := p.buildHTTPRequest(ctx, selectedProvider, transformedRequest, req.IsStreaming, routingDecision.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to build HTTP request: %w", err)
	}

	// 7. Send request to provider
	startTime := time.Now()
	httpResp, err := p.httpClient.Do(httpReq)
	duration := time.Since(startTime)

	// Track provider metrics atomically
	atomic.AddInt64(&p.requestCounter, 1)

	if err != nil {
		// Track provider failure
		if p.performanceMonitor != nil {
			p.performanceMonitor.RecordRequest(performance.RequestMetrics{
				Provider:  selectedProvider.Name,
				StartTime: startTime,
				EndTime:   time.Now(),
				Latency:   duration,
				Success:   false,
				Error:     err,
			})
		}
		return nil, fmt.Errorf("provider request failed: %w", err)
	}

	// Track provider success
	if p.performanceMonitor != nil {
		p.performanceMonitor.RecordRequest(performance.RequestMetrics{
			Provider:  selectedProvider.Name,
			StartTime: startTime,
			EndTime:   time.Now(),
			Latency:   duration,
			Success:   true,
		})
	}

	// 8. Transform response through chain
	transformedResp, err := chain.TransformResponseOut(ctx, httpResp)
	if err != nil {
		// Close response body to prevent leak
		if httpResp.Body != nil {
			_ = httpResp.Body.Close() // Safe to ignore: closing on error path
		}
		return nil, fmt.Errorf("response transformation failed: %w", err)
	}

	// 9. Build response context
	respCtx := &ResponseContext{
		Response:        transformedResp,
		Provider:        routingDecision.Provider,
		Model:           routingDecision.Model,
		TokenCount:      tokenCount,
		RoutingStrategy: routingDecision.Reason,
	}

	return respCtx, nil
}

// buildHTTPRequest builds the HTTP request for the provider
func (p *Pipeline) buildHTTPRequest(ctx context.Context, provider *config.Provider, body interface{}, isStreaming bool, providerName string) (*http.Request, error) {
	// Check if body is a RequestConfig with custom URL/headers
	var reqConfig *transformer.RequestConfig
	var actualBody interface{}

	if rc, ok := body.(*transformer.RequestConfig); ok {
		reqConfig = rc
		actualBody = rc.Body
	} else {
		actualBody = body
	}

	// Marshal request body
	bodyData, err := json.Marshal(actualBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Build URL
	url := provider.APIBaseURL

	// Use custom URL if provided by transformer
	if reqConfig != nil && reqConfig.URL != "" {
		url = reqConfig.URL
	} else {
		// Determine endpoint based on provider type
		endpoint := p.getProviderEndpoint(providerName)

		// Add endpoint to URL
		if !strings.HasSuffix(url, "/") {
			url += "/"
		}
		url += strings.TrimPrefix(endpoint, "/")
	}

	// Create request
	method := "POST"
	if reqConfig != nil && reqConfig.Method != "" {
		method = reqConfig.Method
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(bodyData))
	if err != nil {
		return nil, err
	}

	// Set default headers
	req.Header.Set("Content-Type", "application/json")

	// Set authentication header based on provider
	p.setAuthenticationHeader(req, provider, providerName)

	// Set streaming header if needed
	if isStreaming {
		req.Header.Set("Accept", "text/event-stream")
	}

	// Add user agent
	req.Header.Set("User-Agent", "ccproxy/1.0")

	// Apply custom headers from transformer
	if reqConfig != nil && reqConfig.Headers != nil {
		for key, value := range reqConfig.Headers {
			req.Header.Set(key, value)
		}
	}

	// Set timeout if specified
	if reqConfig != nil && reqConfig.Timeout > 0 {
		ctx, cancel := context.WithTimeout(ctx, time.Duration(reqConfig.Timeout)*time.Millisecond)
		defer cancel()
		req = req.WithContext(ctx)
	}

	return req, nil
}

// getProviderEndpoint returns the appropriate endpoint for a provider
func (p *Pipeline) getProviderEndpoint(providerName string) string {
	// Map provider names to their endpoints
	endpoints := map[string]string{
		"anthropic":  "/v1/messages",
		"openai":     "/v1/chat/completions",
		"groq":       "/openai/v1/chat/completions",
		"deepseek":   "/v1/chat/completions",
		"gemini":     "/v1beta/models/generateContent",
		"openrouter": "/api/v1/chat/completions",
		"mistral":    "/v1/chat/completions",
		"xai":        "/v1/chat/completions",
		"ollama":     "/api/chat",
	}

	if endpoint, exists := endpoints[providerName]; exists {
		return endpoint
	}

	// Default to OpenAI-compatible endpoint
	return "/v1/chat/completions"
}

// setAuthenticationHeader sets the appropriate authentication header for a provider
func (p *Pipeline) setAuthenticationHeader(req *http.Request, provider *config.Provider, providerName string) {
	if provider.APIKey == "" {
		return
	}

	// Provider-specific authentication headers
	switch providerName {
	case "anthropic":
		req.Header.Set("X-API-Key", provider.APIKey)
		req.Header.Set("anthropic-version", "2023-06-01")

	case "gemini":
		// Gemini uses API key as query parameter, handled by transformer
		// But also accepts Authorization header
		req.Header.Set("Authorization", "Bearer "+provider.APIKey)

	case "openrouter":
		req.Header.Set("Authorization", "Bearer "+provider.APIKey)
		// OpenRouter also supports custom headers for app identification
		// TODO: Add support for custom headers in provider configuration

	case "groq":
		req.Header.Set("Authorization", "Bearer "+provider.APIKey)

	case "ollama":
		// Ollama typically doesn't require authentication
		// but support it if configured
		if provider.APIKey != "" {
			req.Header.Set("Authorization", "Bearer "+provider.APIKey)
		}

	default:
		// Default to Bearer token for OpenAI-compatible providers
		req.Header.Set("Authorization", "Bearer "+provider.APIKey)
	}
}

// RequestContext contains the incoming request information
type RequestContext struct {
	Body        interface{}            // Parsed request body
	Headers     map[string]string      // Request headers
	IsStreaming bool                   // Whether this is a streaming request
	Metadata    map[string]interface{} // Additional metadata
}

// ResponseContext contains the response information
type ResponseContext struct {
	Response        *http.Response // HTTP response
	Provider        string         // Selected provider
	Model           string         // Selected model
	TokenCount      int            // Token count
	RoutingStrategy string         // Routing strategy used
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error details
type ErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// NewErrorResponse creates a standardized error response
func NewErrorResponse(message, errorType, code string) *ErrorResponse {
	return &ErrorResponse{
		Error: ErrorDetail{
			Message: message,
			Type:    errorType,
			Code:    code,
		},
	}
}

// WriteErrorResponse writes an error response to the response writer
func WriteErrorResponse(w http.ResponseWriter, statusCode int, err *ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	// Safe to ignore encoding error for error response
	_ = json.NewEncoder(w).Encode(err)
}

// StreamResponse handles streaming responses with transformation support
func (p *Pipeline) StreamResponse(ctx context.Context, w http.ResponseWriter, respCtx *ResponseContext) error {
	// Use the streaming processor for enhanced streaming support
	return p.streamingProcessor.ProcessStreamingResponse(ctx, w, respCtx.Response, respCtx.Provider)
}

// StreamResponse is a compatibility function for simple streaming
func StreamResponse(w http.ResponseWriter, resp *http.Response) error {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	// Get flusher for real-time streaming
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("response writer does not support flushing")
	}

	// Copy response body with flushing
	reader := bufio.NewReader(resp.Body)
	defer resp.Body.Close()

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// Write line
		if _, err := w.Write(line); err != nil {
			return err
		}

		// Flush to client
		flusher.Flush()
	}

	return nil
}

// CopyResponse copies a non-streaming response
func CopyResponse(w http.ResponseWriter, resp *http.Response) error {
	// Copy headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set status code
	w.WriteHeader(resp.StatusCode)

	// Copy body
	defer resp.Body.Close()
	_, err := io.Copy(w, resp.Body)
	return err
}

// HandleStreamingError attempts to send an error event in SSE format
func HandleStreamingError(w http.ResponseWriter, err error) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Try to write error as SSE event
	errorEvent := fmt.Sprintf("event: error\ndata: %s\n\n", err.Error())
	// Safe to ignore write errors for SSE cleanup
	_, _ = w.Write([]byte(errorEvent))

	// Send [DONE] marker
	_, _ = w.Write([]byte("data: [DONE]\n\n"))

	// Flush if possible
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}
