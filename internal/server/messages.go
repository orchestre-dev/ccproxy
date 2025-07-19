package server

import (
	"github.com/gin-gonic/gin"
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
	var req MessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// TODO: Implement the full message processing pipeline:
	// 1. Extract provider from model (provider,model format)
	// 2. Count tokens for routing decision
	// 3. Select appropriate model based on rules
	// 4. Transform request to provider format
	// 5. Call provider API
	// 6. Transform response back to Anthropic format
	// 7. Handle streaming if requested

	// For now, return a not implemented response
	NotImplemented(c, "Message processing pipeline not yet implemented. This will be implemented in Phase 7.17")
}