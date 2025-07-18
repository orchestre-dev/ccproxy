package converter

import (
	"context"
	"encoding/json"
	"testing"

	"ccproxy/internal/models"
)

// Test data for tool conversion compatibility tests
var toolConversionTests = []struct {
	name           string
	anthropicInput *models.MessagesRequest
	expectedError  bool
	description    string
}{
	{
		name: "Simple tool call conversion",
		anthropicInput: &models.MessagesRequest{
			Model: "claude-3-sonnet",
			Messages: []models.Message{
				{
					Role:    "user",
					Content: "What's the weather in New York?",
				},
			},
			Tools: []models.Tool{
				{
					Name:        "get_weather",
					Description: stringPtr("Get weather information for a location"),
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"location": map[string]interface{}{
								"type":        "string",
								"description": "The city and state, e.g. San Francisco, CA",
							},
							"unit": map[string]interface{}{
								"type": "string",
								"enum": []interface{}{"celsius", "fahrenheit"},
							},
						},
						"required": []interface{}{"location"},
					},
				},
			},
		},
		expectedError: false,
		description:   "Should convert simple tool definition correctly",
	},
	{
		name: "Complex nested tool schema",
		anthropicInput: &models.MessagesRequest{
			Model: "claude-3-sonnet",
			Messages: []models.Message{
				{
					Role:    "user",
					Content: "Create a user profile",
				},
			},
			Tools: []models.Tool{
				{
					Name:        "create_user_profile",
					Description: stringPtr("Create a user profile with nested data"),
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"user": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"name": map[string]interface{}{
										"type": "string",
									},
									"age": map[string]interface{}{
										"type": "integer",
										"minimum": 0,
										"maximum": 150,
									},
									"preferences": map[string]interface{}{
										"type": "array",
										"items": map[string]interface{}{
											"type": "string",
										},
									},
								},
								"required": []interface{}{"name"},
							},
							"metadata": map[string]interface{}{
								"type": "object",
								"additionalProperties": true,
							},
						},
						"required": []interface{}{"user"},
					},
				},
			},
		},
		expectedError: false,
		description:   "Should handle complex nested schemas correctly",
	},
	{
		name: "Invalid tool schema - missing type",
		anthropicInput: &models.MessagesRequest{
			Model: "claude-3-sonnet",
			Messages: []models.Message{
				{
					Role:    "user",
					Content: "Test invalid tool",
				},
			},
			Tools: []models.Tool{
				{
					Name: "invalid_tool",
					InputSchema: map[string]interface{}{
						"properties": map[string]interface{}{
							"param": map[string]interface{}{
								"type": "string",
							},
						},
					},
				},
			},
		},
		expectedError: true,
		description:   "Should reject tool schema without type field",
	},
	{
		name: "Empty tool array",
		anthropicInput: &models.MessagesRequest{
			Model: "claude-3-sonnet",
			Messages: []models.Message{
				{
					Role:    "user",
					Content: "No tools here",
				},
			},
			Tools: []models.Tool{},
		},
		expectedError: false,
		description:   "Should handle empty tool array correctly",
	},
}

func TestEnhancedConverter_ConvertRequest(t *testing.T) {
	converter := NewEnhancedConverter(nil, nil)
	
	for _, tt := range toolConversionTests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.ConvertRequest(context.Background(), tt.anthropicInput)
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error for test '%s', but got none", tt.name)
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error for test '%s': %v", tt.name, err)
				return
			}
			
			if result == nil {
				t.Errorf("Expected result for test '%s', but got nil", tt.name)
				return
			}
			
			// Verify basic structure
			if result.Model != tt.anthropicInput.Model {
				t.Errorf("Model mismatch for test '%s': expected %s, got %s", tt.name, tt.anthropicInput.Model, result.Model)
			}
			
			if len(result.Messages) != len(tt.anthropicInput.Messages) {
				t.Errorf("Message count mismatch for test '%s': expected %d, got %d", tt.name, len(tt.anthropicInput.Messages), len(result.Messages))
			}
			
			// Verify tool conversion
			if len(tt.anthropicInput.Tools) > 0 {
				if len(result.Tools) != len(tt.anthropicInput.Tools) {
					t.Errorf("Tool count mismatch for test '%s': expected %d, got %d", tt.name, len(tt.anthropicInput.Tools), len(result.Tools))
				}
				
				// Verify tool structure preservation
				for i, originalTool := range tt.anthropicInput.Tools {
					if i >= len(result.Tools) {
						break
					}
					
					convertedTool := result.Tools[i]
					if convertedTool.Function.Name != originalTool.Name {
						t.Errorf("Tool name mismatch for test '%s', tool %d: expected %s, got %s", tt.name, i, originalTool.Name, convertedTool.Function.Name)
					}
					
					// Verify schema preservation
					if !deepEqual(convertedTool.Function.Parameters, originalTool.InputSchema) {
						t.Errorf("Tool schema not preserved for test '%s', tool %d", tt.name, i)
					}
				}
			}
		})
	}
}

func TestConvertAnthropicToolsToOpenAI(t *testing.T) {
	tests := []struct {
		name          string
		tools         []models.Tool
		expectedError bool
		description   string
	}{
		{
			name: "Valid tool conversion",
			tools: []models.Tool{
				{
					Name:        "test_tool",
					Description: stringPtr("A test tool"),
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"param1": map[string]interface{}{
								"type": "string",
							},
						},
					},
				},
			},
			expectedError: false,
			description:   "Should convert valid tool correctly",
		},
		{
			name: "Tool with nil description",
			tools: []models.Tool{
				{
					Name:        "test_tool",
					Description: nil,
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"param1": map[string]interface{}{
								"type": "string",
							},
						},
					},
				},
			},
			expectedError: false,
			description:   "Should handle nil description correctly",
		},
		{
			name: "Invalid tool - empty name",
			tools: []models.Tool{
				{
					Name:        "",
					Description: stringPtr("Invalid tool"),
					InputSchema: map[string]interface{}{
						"type": "object",
					},
				},
			},
			expectedError: true,
			description:   "Should reject tool with empty name",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertAnthropicToolsToOpenAI(tt.tools)
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error for test '%s', but got none", tt.name)
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error for test '%s': %v", tt.name, err)
				return
			}
			
			if len(result) != len(tt.tools) {
				t.Errorf("Result length mismatch for test '%s': expected %d, got %d", tt.name, len(tt.tools), len(result))
			}
		})
	}
}

func TestConvertOpenAIToolCallsToAnthropic(t *testing.T) {
	tests := []struct {
		name          string
		toolCalls     []models.ToolCall
		expectedError bool
		description   string
	}{
		{
			name: "Valid tool call conversion",
			toolCalls: []models.ToolCall{
				{
					ID:   "call_123",
					Type: "function",
					Function: models.FunctionCall{
						Name:      "get_weather",
						Arguments: `{"location": "New York", "unit": "celsius"}`,
					},
				},
			},
			expectedError: false,
			description:   "Should convert valid tool call correctly",
		},
		{
			name: "Tool call with empty arguments",
			toolCalls: []models.ToolCall{
				{
					ID:   "call_456",
					Type: "function",
					Function: models.FunctionCall{
						Name:      "simple_tool",
						Arguments: "",
					},
				},
			},
			expectedError: false,
			description:   "Should handle empty arguments correctly",
		},
		{
			name: "Invalid tool call - malformed JSON",
			toolCalls: []models.ToolCall{
				{
					ID:   "call_789",
					Type: "function",
					Function: models.FunctionCall{
						Name:      "broken_tool",
						Arguments: `{"invalid": json}`,
					},
				},
			},
			expectedError: true,
			description:   "Should reject tool call with malformed JSON",
		},
		{
			name: "Invalid tool call - empty ID",
			toolCalls: []models.ToolCall{
				{
					ID:   "",
					Type: "function",
					Function: models.FunctionCall{
						Name:      "test_tool",
						Arguments: `{}`,
					},
				},
			},
			expectedError: true,
			description:   "Should reject tool call with empty ID",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertOpenAIToolCallsToAnthropic(tt.toolCalls)
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error for test '%s', but got none", tt.name)
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error for test '%s': %v", tt.name, err)
				return
			}
			
			if len(result) != len(tt.toolCalls) {
				t.Errorf("Result length mismatch for test '%s': expected %d, got %d", tt.name, len(tt.toolCalls), len(result))
			}
			
			// Verify correlation IDs are preserved
			for i, originalCall := range tt.toolCalls {
				if i >= len(result) {
					break
				}
				
				if result[i].ID != originalCall.ID {
					t.Errorf("Correlation ID not preserved for test '%s', call %d: expected %s, got %s", tt.name, i, originalCall.ID, result[i].ID)
				}
				
				if result[i].Name != originalCall.Function.Name {
					t.Errorf("Function name not preserved for test '%s', call %d: expected %s, got %s", tt.name, i, originalCall.Function.Name, result[i].Name)
				}
			}
		})
	}
}

func TestToolConversionError(t *testing.T) {
	err := NewToolConversionError("validation_error", "test error").
		WithTool("test_tool", "call_123").
		WithField("param", "invalid_value").
		WithSuggestions("Fix the parameter", "Check the documentation")
	
	if err.Type != "validation_error" {
		t.Errorf("Expected error type 'validation_error', got '%s'", err.Type)
	}
	
	if err.ToolName != "test_tool" {
		t.Errorf("Expected tool name 'test_tool', got '%s'", err.ToolName)
	}
	
	if len(err.Suggestions) != 2 {
		t.Errorf("Expected 2 suggestions, got %d", len(err.Suggestions))
	}
	
	errorMsg := err.Error()
	if errorMsg == "" {
		t.Error("Error message should not be empty")
	}
}

func TestValidateToolCallArguments(t *testing.T) {
	toolSchema := models.Tool{
		Name: "test_tool",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"required_param": map[string]interface{}{
					"type": "string",
				},
				"optional_param": map[string]interface{}{
					"type": "integer",
				},
			},
			"required": []interface{}{"required_param"},
		},
	}
	
	tests := []struct {
		name          string
		toolCall      models.ToolCall
		expectedError bool
		description   string
	}{
		{
			name: "Valid tool call",
			toolCall: models.ToolCall{
				ID:   "call_123",
				Type: "function",
				Function: models.FunctionCall{
					Name:      "test_tool",
					Arguments: `{"required_param": "value", "optional_param": 42}`,
				},
			},
			expectedError: false,
			description:   "Should validate correct tool call",
		},
		{
			name: "Missing required parameter",
			toolCall: models.ToolCall{
				ID:   "call_456",
				Type: "function",
				Function: models.FunctionCall{
					Name:      "test_tool",
					Arguments: `{"optional_param": 42}`,
				},
			},
			expectedError: true,
			description:   "Should reject tool call missing required parameter",
		},
		{
			name: "Invalid JSON arguments",
			toolCall: models.ToolCall{
				ID:   "call_789",
				Type: "function",
				Function: models.FunctionCall{
					Name:      "test_tool",
					Arguments: `{"invalid": json}`,
				},
			},
			expectedError: true,
			description:   "Should reject tool call with invalid JSON",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateToolCallArguments(tt.toolCall, toolSchema)
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error for test '%s', but got none", tt.name)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for test '%s': %v", tt.name, err)
				}
			}
		})
	}
}

// Helper functions

func stringPtr(s string) *string {
	return &s
}

func deepEqual(a, b interface{}) bool {
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	return string(aJSON) == string(bJSON)
}