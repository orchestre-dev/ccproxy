package server

import (
	"context"
	"net/http"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/musistudio/ccproxy/internal/pipeline"
	"github.com/musistudio/ccproxy/internal/utils"
)

// Message request structure (Anthropic format)
type MessageRequest struct {
	Model     string    `json:"model" binding:"required"`
	Messages  []Message `json:"messages" binding:"required,min=1"`
	MaxTokens int       `json:"max_tokens,omitempty"`
	Stream    bool      `json:"stream,omitempty"`
	System    string    `json:"system,omitempty"`
}

// Message structure
type Message struct {
	Role    string `json:"role" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// Message response structure
type MessageResponse struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Role     string   `json:"role"`
	Content  []Content `json:"content"`
	Model    string   `json:"model"`
	Usage    Usage    `json:"usage,omitempty"`
}

// Content structure
type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Usage structure
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// handleMessages processes the main Claude API endpoint
func (s *Server) handleMessages(c *gin.Context) {
	// Increment request counter
	atomic.AddInt64(&s.requestsServed, 1)
	
	// Parse raw body for pipeline processing
	var rawBody interface{}
	if err := c.ShouldBindJSON(&rawBody); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// Check if streaming is requested
	isStreaming := false
	if bodyMap, ok := rawBody.(map[string]interface{}); ok {
		if stream, ok := bodyMap["stream"].(bool); ok {
			isStreaming = stream
		}
	}

	// Create request context
	reqCtx := &pipeline.RequestContext{
		Body:        rawBody,
		Headers:     extractHeaders(c),
		IsStreaming: isStreaming,
		Metadata:    make(map[string]interface{}),
	}

	// Process through pipeline
	ctx := context.Background()
	respCtx, err := s.pipeline.ProcessRequest(ctx, reqCtx)
	if err != nil {
		utils.GetLogger().Errorf("Pipeline processing failed: %v", err)
		
		// Return appropriate error response
		errResp := pipeline.NewErrorResponse(
			err.Error(),
			"api_error",
			"pipeline_error",
		)
		pipeline.WriteErrorResponse(c.Writer, http.StatusInternalServerError, errResp)
		return
	}

	// Log routing decision
	utils.GetLogger().Infof("Routed to provider=%s, model=%s, tokens=%d, strategy=%s",
		respCtx.Provider, respCtx.Model, respCtx.TokenCount, respCtx.RoutingStrategy)

	// Handle response based on streaming
	if isStreaming {
		// Stream the response with transformation support
		if err := s.pipeline.StreamResponse(ctx, c.Writer, respCtx); err != nil {
			utils.GetLogger().Errorf("Streaming failed: %v", err)
			// Try to send error event if possible
			pipeline.HandleStreamingError(c.Writer, err)
		}
	} else {
		// Copy non-streaming response
		if err := pipeline.CopyResponse(c.Writer, respCtx.Response); err != nil {
			utils.GetLogger().Errorf("Response copy failed: %v", err)
		}
	}
}

// extractHeaders extracts relevant headers from the request
func extractHeaders(c *gin.Context) map[string]string {
	headers := make(map[string]string)
	
	// Extract relevant headers
	relevantHeaders := []string{
		"Authorization",
		"X-Api-Key",
		"Content-Type",
		"Accept",
		"User-Agent",
	}
	
	for _, header := range relevantHeaders {
		if value := c.GetHeader(header); value != "" {
			headers[header] = value
		}
	}
	
	return headers
}