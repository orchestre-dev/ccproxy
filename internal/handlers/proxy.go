// Package handlers provides HTTP request handlers for the CCProxy server
package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ccproxy/internal/converter"
	"ccproxy/internal/models"
	"ccproxy/internal/provider"
	"ccproxy/internal/provider/common"
	"ccproxy/pkg/logger"
)

// Handler holds dependencies for request handlers
type Handler struct {
	provider provider.Provider
	logger   *logger.Logger
}

// NewHandler creates a new handler instance
func NewHandler(provider provider.Provider, logger *logger.Logger) *Handler {
	return &Handler{
		provider: provider,
		logger:   logger,
	}
}

// RegisterRoutes registers all routes with the Gin router
func RegisterRoutes(router *gin.Engine, provider provider.Provider, logger *logger.Logger) {
	handler := NewHandler(provider, logger)

	// Health check endpoint
	router.GET("/", handler.HealthCheck)

	// Main proxy endpoint
	router.POST("/v1/messages", handler.ProxyMessages)

	// Additional health endpoint
	router.GET("/health", handler.DetailedHealthCheck)

	// Provider status endpoint
	router.GET("/status", handler.ProviderStatus)
}

// HealthCheck handles the root health check endpoint
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message":  "CCProxy Multi-Provider Anthropic API is alive 💡",
		"provider": h.provider.GetName(),
	})
}

// DetailedHealthCheck provides more detailed health information
func (h *Handler) DetailedHealthCheck(c *gin.Context) {
	// Check provider health
	err := h.provider.HealthCheck(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Provider health check failed")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"service": "ccproxy",
			"version": "1.0.0",
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "ccproxy",
		"version": "1.0.0",
	})
}

// ProviderStatus provides information about the current provider
func (h *Handler) ProviderStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"provider":   h.provider.GetName(),
		"model":      h.provider.GetModel(),
		"base_url":   h.provider.GetBaseURL(),
		"max_tokens": h.provider.GetMaxTokens(),
		"status":     "active",
		"service":    "ccproxy",
		"version":    "1.0.0",
	})
}

// ProxyMessages handles the main proxy endpoint for messages
func (h *Handler) ProxyMessages(c *gin.Context) {
	requestID := getRequestIDFromContext(c)

	// Parse request
	var req models.MessagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithRequestID(requestID).WithError(err).Error("Failed to parse request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid request format",
			"request_id": requestID,
		})
		return
	}

	// Basic validation
	if len(req.Messages) == 0 {
		h.logger.WithRequestID(requestID).Error("No messages provided in request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Messages array cannot be empty",
			"request_id": requestID,
		})
		return
	}

	// Log incoming request
	h.logger.APILog("anthropic_request", map[string]interface{}{
		"model":      req.Model,
		"messages":   len(req.Messages),
		"max_tokens": req.MaxTokens,
		"tools":      len(req.Tools),
		"provider":   h.provider.GetName(),
	}, requestID)

	// Convert Anthropic request to OpenAI format
	openaiReq, err := converter.ConvertAnthropicToOpenAI(&req)
	if err != nil {
		h.logger.WithRequestID(requestID).WithError(err).Error("Failed to convert request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Failed to convert request format",
			"request_id": requestID,
		})
		return
	}

	// Add request ID to context for provider
	type contextKey string
	const requestIDKey contextKey = "request_id"
	ctx := context.WithValue(c.Request.Context(), requestIDKey, requestID)

	// Send request to provider
	openaiResp, err := h.provider.CreateChatCompletion(ctx, openaiReq)
	if err != nil {
		h.logger.WithRequestID(requestID).WithError(err).Errorf("Failed to call %s API", h.provider.GetName())
		
		// Check if it's a ProviderError with a specific status code
		statusCode := http.StatusInternalServerError
		errorMessage := "Failed to process request"
		
		var providerErr *common.ProviderError
		if errors.As(err, &providerErr) {
			if providerErr.Code != 0 {
				statusCode = providerErr.Code
			}
			// For client errors, provide more specific messages
			if statusCode == http.StatusBadRequest || statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden {
				errorMessage = providerErr.Message
			}
		}
		
		c.JSON(statusCode, gin.H{
			"error":      errorMessage,
			"request_id": requestID,
		})
		return
	}

	// Convert OpenAI response back to Anthropic format
	anthropicResp, err := converter.ConvertOpenAIToAnthropic(openaiResp, generateMessageID(), h.provider.GetName())
	if err != nil {
		h.logger.WithRequestID(requestID).WithError(err).Error("Failed to convert response")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to convert response format",
			"request_id": requestID,
		})
		return
	}

	// Log response
	h.logger.APILog("anthropic_response", map[string]interface{}{
		"stop_reason":    anthropicResp.StopReason,
		"input_tokens":   anthropicResp.Usage.InputTokens,
		"output_tokens":  anthropicResp.Usage.OutputTokens,
		"content_blocks": len(anthropicResp.Content),
	}, requestID)

	// Return response
	c.JSON(http.StatusOK, anthropicResp)
}

// getRequestIDFromContext extracts request ID from Gin context
func getRequestIDFromContext(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return uuid.New().String()
}

// generateMessageID generates a new message ID
func generateMessageID() string {
	return "msg_" + uuid.New().String()[:12]
}
