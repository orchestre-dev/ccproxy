package router

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/musistudio/ccproxy/internal/config"
	"github.com/musistudio/ccproxy/internal/utils"
)

// RouterMiddleware creates a middleware that performs intelligent model routing
func RouterMiddleware(cfg *config.Config) gin.HandlerFunc {
	router := New(cfg)
	logger := utils.GetLogger()
	
	return func(c *gin.Context) {
		// Only process POST /v1/messages
		if c.Request.Method != http.MethodPost || c.Request.URL.Path != "/v1/messages" {
			c.Next()
			return
		}
		
		// Parse request body
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			// If parsing fails, let the handler deal with it
			c.Next()
			return
		}
		
		// Get model from request
		modelStr, ok := body["model"].(string)
		if !ok || modelStr == "" {
			c.Next()
			return
		}
		
		// Create request object
		req := Request{
			Model: modelStr,
		}
		
		// Check for thinking parameter
		if thinking, ok := body["thinking"].(bool); ok {
			req.Thinking = thinking
		}
		
		// Count tokens
		tokenCount := 0
		params := &utils.MessageCreateParams{
			Model: modelStr,
		}
		
		// Parse messages
		if messages, ok := body["messages"].([]interface{}); ok {
			params.Messages = make([]utils.Message, 0, len(messages))
			for _, m := range messages {
				if msgMap, ok := m.(map[string]interface{}); ok {
					msg := utils.Message{
						Role:    getStringValue(msgMap, "role"),
						Content: msgMap["content"],
					}
					params.Messages = append(params.Messages, msg)
				}
			}
		}
		
		// Parse system
		if system, ok := body["system"]; ok {
			params.System = system
		}
		
		// Parse tools
		if tools, ok := body["tools"].([]interface{}); ok {
			params.Tools = make([]utils.Tool, 0, len(tools))
			for _, t := range tools {
				if toolMap, ok := t.(map[string]interface{}); ok {
					tool := utils.Tool{
						Name:        getStringValue(toolMap, "name"),
						Description: getStringValue(toolMap, "description"),
						InputSchema: toolMap["input_schema"],
					}
					params.Tools = append(params.Tools, tool)
				}
			}
		}
		
		// Count tokens
		var err error
		tokenCount, err = utils.CountMessageTokens(params)
		if err != nil {
			logger.WithError(err).Error("Failed to count tokens in router middleware")
			// Continue with tokenCount = 0
		}
		
		// Perform routing
		decision := router.Route(req, tokenCount)
		
		// Update the model in the request
		newModel := FormatModelString(decision.Provider, decision.Model)
		body["model"] = newModel
		
		// Log routing decision
		logger.WithFields(map[string]interface{}{
			"original_model": modelStr,
			"routed_model":   newModel,
			"token_count":    tokenCount,
			"reason":         decision.Reason,
		}).Debug("Model routing decision")
		
		// Store routing metadata in context
		c.Set("routing_decision", decision)
		c.Set("token_count", tokenCount)
		
		// Re-bind the modified body
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			logger.WithError(err).Error("Failed to marshal modified request body")
			c.Next()
			return
		}
		
		// Replace request body
		c.Request.Body = &bodyReader{
			data: bodyBytes,
			pos:  0,
		}
		
		c.Next()
	}
}

// bodyReader implements io.ReadCloser for replacing request body
type bodyReader struct {
	data []byte
	pos  int
}

func (r *bodyReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func (r *bodyReader) Close() error {
	return nil
}

// getStringValue safely gets a string value from a map
func getStringValue(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}