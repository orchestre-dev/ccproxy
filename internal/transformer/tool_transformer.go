package transformer

import (
	"context"
	"fmt"
	"strings"

	"github.com/orchestre-dev/ccproxy/internal/config"
	"github.com/orchestre-dev/ccproxy/internal/tools"
)

// ToolTransformer handles tool-related transformations
type ToolTransformer struct {
	BaseTransformer
	handler *tools.Handler
}

// NewToolTransformer creates a new tool transformer
func NewToolTransformer() *ToolTransformer {
	return &ToolTransformer{
		BaseTransformer: BaseTransformer{
			name: "tool",
		},
		handler: tools.NewHandler(),
	}
}

// TransformRequest transforms tool definitions in the request
func (t *ToolTransformer) TransformRequest(ctx context.Context, provider *config.Provider, request interface{}) (interface{}, error) {
	reqMap, ok := request.(map[string]interface{})
	if !ok {
		return request, nil
	}

	// Check for tools in the request
	toolsRaw, hasTools := reqMap["tools"]
	if !hasTools {
		return request, nil
	}

	tools, ok := toolsRaw.([]interface{})
	if !ok {
		return request, nil
	}

	// Validate tools
	for _, tool := range tools {
		if err := t.handler.ValidateToolDefinition(tool); err != nil {
			return nil, fmt.Errorf("invalid tool definition: %w", err)
		}
	}

	// Transform tools for the specific provider
	transformedTools, err := t.handler.TransformToolsForProvider(tools, provider.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to transform tools: %w", err)
	}

	// Update request with transformed tools
	reqMap["tools"] = transformedTools

	// For OpenAI, tools are passed differently
	if provider.Name == "openai" || provider.Name == "gpt" {
		// OpenAI uses "functions" instead of "tools" for older models
		// Modern models use tools with function type
		if model, ok := reqMap["model"].(string); ok {
			// Strip provider prefix if present
			if strings.Contains(model, ",") {
				parts := strings.SplitN(model, ",", 2)
				if len(parts) == 2 {
					model = parts[1]
				}
			}
			if isLegacyOpenAIModel(model) {
				// Convert to legacy format
				var functions []interface{}
				for _, tool := range transformedTools {
					if toolMap, ok := tool.(map[string]interface{}); ok {
						if function, ok := toolMap["function"]; ok {
							functions = append(functions, function)
						}
					}
				}
				delete(reqMap, "tools")
				reqMap["functions"] = functions
			}
		}
	}

	return reqMap, nil
}

// TransformResponse processes tool use in responses
func (t *ToolTransformer) TransformResponse(ctx context.Context, provider *config.Provider, response interface{}) (interface{}, error) {
	// Use the tool handler to process the response
	return t.handler.ProcessToolResponse(ctx, response)
}

// TransformSSEEvent transforms tool use in SSE events
func (t *ToolTransformer) TransformSSEEvent(ctx context.Context, provider *config.Provider, event string) (string, error) {
	// For now, pass through - streaming tool handling is done in the streaming processor
	return event, nil
}

// isLegacyOpenAIModel checks if the model uses legacy function calling
func isLegacyOpenAIModel(model string) bool {
	legacyModels := []string{
		"gpt-3.5-turbo-0613",
		"gpt-3.5-turbo-16k-0613",
		"gpt-4-0613",
		"gpt-4-32k-0613",
	}

	for _, legacy := range legacyModels {
		if model == legacy {
			return true
		}
	}

	return false
}

// ExtractToolCalls extracts tool calls from a response message
func (t *ToolTransformer) ExtractToolCalls(message interface{}) ([]tools.ToolUse, error) {
	return t.handler.ExtractToolUses(message)
}

// CreateToolResultMessage creates a tool result message
func (t *ToolTransformer) CreateToolResultMessage(results []tools.ToolResult) interface{} {
	msg := t.handler.CreateToolResultMessage(results)

	// Convert to generic interface for flexibility
	return map[string]interface{}{
		"role":    msg.Role,
		"content": msg.Content,
	}
}
