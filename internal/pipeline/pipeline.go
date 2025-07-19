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
	"time"

	"github.com/musistudio/ccproxy/internal/config"
	"github.com/musistudio/ccproxy/internal/converter"
	"github.com/musistudio/ccproxy/internal/providers"
	"github.com/musistudio/ccproxy/internal/proxy"
	"github.com/musistudio/ccproxy/internal/router"
	"github.com/musistudio/ccproxy/internal/transformer"
	"github.com/musistudio/ccproxy/internal/utils"
)

// Pipeline handles the complete request processing flow
type Pipeline struct {
	config             *config.Config
	providerService    *providers.Service
	transformerService *transformer.Service
	router             *router.Router
	httpClient         *http.Client
	streamingProcessor *StreamingProcessor
	messageConverter   *converter.MessageConverter
}

// NewPipeline creates a new request processing pipeline
func NewPipeline(
	cfg *config.Config,
	providerService *providers.Service,
	transformerService *transformer.Service,
	router *router.Router,
) *Pipeline {
	// Create HTTP client with timeout
	// Default 60 minutes timeout
	timeout := 60 * time.Minute

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
	provider, err := p.providerService.GetProvider(routingDecision.Provider)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %s", routingDecision.Provider)
	}

	// 3. Apply request transformations
	transformedRequest, err := p.transformerService.ApplyRequestTransformation(ctx, provider, req.Body)
	if err != nil {
		return nil, fmt.Errorf("request transformation failed: %w", err)
	}

	// 4. Build HTTP request
	httpReq, err := p.buildHTTPRequest(ctx, provider, transformedRequest, req.IsStreaming)
	if err != nil {
		return nil, fmt.Errorf("failed to build HTTP request: %w", err)
	}

	// 5. Send request to provider
	httpResp, err := p.httpClient.Do(httpReq)
	if err != nil {
		// Track provider failure
		// TODO: Add provider request tracking when implementing metrics
		return nil, fmt.Errorf("provider request failed: %w", err)
	}

	// Track provider success (initially)
	// TODO: Add provider request tracking when implementing metrics

	// 6. Transform response
	transformedResp, err := p.transformerService.ApplyResponseTransformation(ctx, provider, &transformer.Response{Response: httpResp})
	if err != nil {
		return nil, fmt.Errorf("response transformation failed: %w", err)
	}

	// 7. Build response context
	respCtx := &ResponseContext{
		Response:        transformedResp.(*transformer.Response).Response,
		Provider:        routingDecision.Provider,
		Model:           routingDecision.Model,
		TokenCount:      tokenCount,
		RoutingStrategy: routingDecision.Reason,
	}

	return respCtx, nil
}

// buildHTTPRequest builds the HTTP request for the provider
func (p *Pipeline) buildHTTPRequest(ctx context.Context, provider *config.Provider, body interface{}, isStreaming bool) (*http.Request, error) {
	// Marshal request body
	bodyData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Build URL
	url := provider.APIBaseURL
	
	// Determine endpoint based on provider type
	// TODO: This should be determined by the transformer
	endpoint := "/v1/messages" // Default Anthropic endpoint

	// Add endpoint to URL
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	url += strings.TrimPrefix(endpoint, "/")

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyData))
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	
	// Set authentication header
	if provider.APIKey != "" {
		// Default to Authorization Bearer
		req.Header.Set("Authorization", "Bearer "+provider.APIKey)
	}

	// Set streaming header if needed
	if isStreaming {
		req.Header.Set("Accept", "text/event-stream")
	}

	// Add user agent
	req.Header.Set("User-Agent", "ccproxy/1.0")

	return req, nil
}

// RequestContext contains the incoming request information
type RequestContext struct {
	Body        interface{}          // Parsed request body
	Headers     map[string]string    // Request headers
	IsStreaming bool                 // Whether this is a streaming request
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
	json.NewEncoder(w).Encode(err)
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