package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/orchestre-dev/ccproxy/internal/utils"
	"github.com/sirupsen/logrus"
)

// Handler processes tool-related responses and requests
type Handler struct {
	logger *logrus.Logger
}

// NewHandler creates a new tool handler
func NewHandler() *Handler {
	return &Handler{
		logger: utils.GetLogger(),
	}
}

// ToolUse represents a tool use request from Claude
type ToolUse struct {
	Type  string          `json:"type"`
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// ToolResult represents a tool execution result
type ToolResult struct {
	Type      string          `json:"type"`
	ToolUseID string          `json:"tool_use_id"`
	Content   json.RawMessage `json:"content,omitempty"`
	Error     *string         `json:"error,omitempty"`
}

// ContentBlock represents a content block in Claude's response
type ContentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
}

// Message represents a message with content blocks
type Message struct {
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

// ProcessToolResponse processes a response that may contain tool use blocks
func (h *Handler) ProcessToolResponse(ctx context.Context, response interface{}) (interface{}, error) {
	// Check if response contains tool use blocks
	respMap, ok := response.(map[string]interface{})
	if !ok {
		return response, nil
	}

	content, exists := respMap["content"]
	if !exists {
		return response, nil
	}

	// Handle both string content and array content
	switch v := content.(type) {
	case string:
		// Simple text response, no tool use
		return response, nil
	case []interface{}:
		// Array of content blocks, check for tool use
		return h.processContentBlocks(ctx, respMap, v)
	default:
		return response, nil
	}
}

// processContentBlocks processes an array of content blocks
func (h *Handler) processContentBlocks(ctx context.Context, response map[string]interface{}, blocks []interface{}) (interface{}, error) {
	var hasToolUse bool
	var processedBlocks []interface{}

	for _, block := range blocks {
		blockMap, ok := block.(map[string]interface{})
		if !ok {
			processedBlocks = append(processedBlocks, block)
			continue
		}

		blockType, _ := blockMap["type"].(string)
		if blockType == "tool_use" {
			hasToolUse = true
			// Process tool use block
			processedBlock, err := h.processToolUseBlock(ctx, blockMap)
			if err != nil {
				h.logger.Warnf("Failed to process tool use block: %v", err)
				processedBlocks = append(processedBlocks, block)
			} else {
				processedBlocks = append(processedBlocks, processedBlock)
			}
		} else {
			processedBlocks = append(processedBlocks, block)
		}
	}

	if hasToolUse {
		// Update response metadata to indicate tool use
		if metadata, ok := response["metadata"].(map[string]interface{}); ok {
			metadata["has_tool_use"] = true
		} else {
			response["metadata"] = map[string]interface{}{
				"has_tool_use": true,
			}
		}
	}

	response["content"] = processedBlocks
	return response, nil
}

// processToolUseBlock processes a single tool use block
func (h *Handler) processToolUseBlock(ctx context.Context, block map[string]interface{}) (interface{}, error) {
	toolName, _ := block["name"].(string)
	toolID, _ := block["id"].(string)

	h.logger.Debugf("Processing tool use: %s (ID: %s)", toolName, toolID)

	// Add metadata to the block
	block["_processed"] = true
	block["_processor"] = "ccproxy"

	return block, nil
}

// ExtractToolUses extracts tool use blocks from a message
func (h *Handler) ExtractToolUses(message interface{}) ([]ToolUse, error) {
	msgMap, ok := message.(map[string]interface{})
	if !ok {
		return nil, nil
	}

	content, exists := msgMap["content"]
	if !exists {
		return nil, nil
	}

	blocks, ok := content.([]interface{})
	if !ok {
		return nil, nil
	}

	var toolUses []ToolUse
	for _, block := range blocks {
		blockMap, ok := block.(map[string]interface{})
		if !ok {
			continue
		}

		blockType, _ := blockMap["type"].(string)
		if blockType != "tool_use" {
			continue
		}

		toolUse := ToolUse{
			Type: blockType,
		}

		if id, ok := blockMap["id"].(string); ok {
			toolUse.ID = id
		}
		if name, ok := blockMap["name"].(string); ok {
			toolUse.Name = name
		}
		if input, ok := blockMap["input"]; ok {
			inputBytes, err := json.Marshal(input)
			if err == nil {
				toolUse.Input = inputBytes
			}
		}

		toolUses = append(toolUses, toolUse)
	}

	return toolUses, nil
}

// CreateToolResultMessage creates a message containing tool results
func (h *Handler) CreateToolResultMessage(results []ToolResult) Message {
	var content []ContentBlock

	for _, result := range results {
		block := ContentBlock{
			Type:      "tool_result",
			ToolUseID: result.ToolUseID,
		}

		if result.Error != nil {
			block.Content = json.RawMessage(fmt.Sprintf(`{"error": %q}`, *result.Error))
		} else {
			block.Content = result.Content
		}

		content = append(content, block)
	}

	return Message{
		Role:    "user",
		Content: content,
	}
}

// ValidateToolDefinition validates a tool definition
func (h *Handler) ValidateToolDefinition(tool interface{}) error {
	toolMap, ok := tool.(map[string]interface{})
	if !ok {
		return fmt.Errorf("tool must be an object")
	}

	// Check required fields
	name, hasName := toolMap["name"].(string)
	if !hasName || name == "" {
		return fmt.Errorf("tool must have a non-empty name")
	}

	description, hasDesc := toolMap["description"].(string)
	if !hasDesc || description == "" {
		return fmt.Errorf("tool must have a non-empty description")
	}

	// Validate input schema if present
	if inputSchema, hasSchema := toolMap["input_schema"]; hasSchema {
		if _, ok := inputSchema.(map[string]interface{}); !ok {
			return fmt.Errorf("input_schema must be an object")
		}
	}

	return nil
}

// TransformToolsForProvider transforms tool definitions for a specific provider
func (h *Handler) TransformToolsForProvider(tools []interface{}, provider string) ([]interface{}, error) {
	if tools == nil || len(tools) == 0 {
		return tools, nil
	}

	provider = strings.ToLower(provider)

	switch provider {
	case "anthropic":
		// Anthropic format is already the standard
		return tools, nil
	case "openai", "gpt":
		return h.transformToolsForOpenAI(tools)
	case "google", "gemini":
		return h.transformToolsForGoogle(tools)
	default:
		// Return tools as-is for unknown providers
		h.logger.Warnf("Unknown provider for tool transformation: %s", provider)
		return tools, nil
	}
}

// transformToolsForOpenAI transforms tools to OpenAI function calling format
func (h *Handler) transformToolsForOpenAI(tools []interface{}) ([]interface{}, error) {
	var transformed []interface{}

	for _, tool := range tools {
		toolMap, ok := tool.(map[string]interface{})
		if !ok {
			continue
		}

		// OpenAI uses "function" type
		function := map[string]interface{}{
			"name":        toolMap["name"],
			"description": toolMap["description"],
		}

		if inputSchema, ok := toolMap["input_schema"]; ok {
			function["parameters"] = inputSchema
		}

		transformed = append(transformed, map[string]interface{}{
			"type":     "function",
			"function": function,
		})
	}

	return transformed, nil
}

// transformToolsForGoogle transforms tools to Google/Gemini format
func (h *Handler) transformToolsForGoogle(tools []interface{}) ([]interface{}, error) {
	var transformed []interface{}

	for _, tool := range tools {
		toolMap, ok := tool.(map[string]interface{})
		if !ok {
			continue
		}

		// Google uses a flatter structure
		googleTool := map[string]interface{}{
			"name":        toolMap["name"],
			"description": toolMap["description"],
		}

		if inputSchema, ok := toolMap["input_schema"].(map[string]interface{}); ok {
			// Transform schema properties
			if properties, ok := inputSchema["properties"]; ok {
				googleTool["parameters"] = properties
			}
		}

		transformed = append(transformed, googleTool)
	}

	return transformed, nil
}
