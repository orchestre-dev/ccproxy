package tools

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

func TestProcessToolResponse(t *testing.T) {
	h := NewHandler()
	ctx := context.Background()

	tests := []struct {
		name     string
		response interface{}
		wantErr  bool
		validate func(t *testing.T, result interface{})
	}{
		{
			name: "simple text response",
			response: map[string]interface{}{
				"content": "Hello, world!",
			},
			validate: func(t *testing.T, result interface{}) {
				resp := result.(map[string]interface{})
				if resp["content"] != "Hello, world!" {
					t.Errorf("Expected content to remain unchanged")
				}
			},
		},
		{
			name: "response with tool use",
			response: map[string]interface{}{
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "Let me help you with that.",
					},
					map[string]interface{}{
						"type": "tool_use",
						"id":   "tool_123",
						"name": "calculator",
						"input": map[string]interface{}{
							"operation": "add",
							"a":         5,
							"b":         3,
						},
					},
				},
			},
			validate: func(t *testing.T, result interface{}) {
				resp := result.(map[string]interface{})
				
				// Check metadata was added
				metadata, ok := resp["metadata"].(map[string]interface{})
				if !ok {
					t.Fatal("Expected metadata to be added")
				}
				if metadata["has_tool_use"] != true {
					t.Error("Expected has_tool_use to be true")
				}
				
				// Check tool use block was processed
				content := resp["content"].([]interface{})
				toolBlock := content[1].(map[string]interface{})
				if toolBlock["_processed"] != true {
					t.Error("Expected tool block to be marked as processed")
				}
			},
		},
		{
			name: "non-map response",
			response: "plain string",
			validate: func(t *testing.T, result interface{}) {
				if result != "plain string" {
					t.Error("Expected non-map response to remain unchanged")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := h.ProcessToolResponse(ctx, tt.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessToolResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestExtractToolUses(t *testing.T) {
	h := NewHandler()

	tests := []struct {
		name    string
		message interface{}
		want    []ToolUse
		wantErr bool
	}{
		{
			name: "message with tool uses",
			message: map[string]interface{}{
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "I'll calculate that for you.",
					},
					map[string]interface{}{
						"type": "tool_use",
						"id":   "calc_001",
						"name": "calculator",
						"input": map[string]interface{}{
							"expression": "2 + 2",
						},
					},
					map[string]interface{}{
						"type": "tool_use",
						"id":   "search_001",
						"name": "web_search",
						"input": map[string]interface{}{
							"query": "weather today",
						},
					},
				},
			},
			want: []ToolUse{
				{
					Type: "tool_use",
					ID:   "calc_001",
					Name: "calculator",
					Input: json.RawMessage(`{"expression":"2 + 2"}`),
				},
				{
					Type: "tool_use",
					ID:   "search_001",
					Name: "web_search",
					Input: json.RawMessage(`{"query":"weather today"}`),
				},
			},
		},
		{
			name: "message without tool uses",
			message: map[string]interface{}{
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "Here's the answer.",
					},
				},
			},
			want: []ToolUse{},
		},
		{
			name:    "non-map message",
			message: "plain string",
			want:    nil,
		},
		{
			name: "string content",
			message: map[string]interface{}{
				"content": "Just a text response",
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.ExtractToolUses(tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractToolUses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if len(got) != len(tt.want) {
				t.Errorf("ExtractToolUses() got %d tools, want %d", len(got), len(tt.want))
				return
			}
			
			for i, tool := range got {
				if tool.Type != tt.want[i].Type || tool.ID != tt.want[i].ID || tool.Name != tt.want[i].Name {
					t.Errorf("Tool %d mismatch: got %+v, want %+v", i, tool, tt.want[i])
				}
				
				// Compare Input JSON
				if tt.want[i].Input != nil {
					var gotInput, wantInput interface{}
					json.Unmarshal(tool.Input, &gotInput)
					json.Unmarshal(tt.want[i].Input, &wantInput)
					if !reflect.DeepEqual(gotInput, wantInput) {
						t.Errorf("Tool %d input mismatch: got %v, want %v", i, gotInput, wantInput)
					}
				}
			}
		})
	}
}

func TestCreateToolResultMessage(t *testing.T) {
	h := NewHandler()

	errorMsg := "Tool execution failed"
	results := []ToolResult{
		{
			Type:      "tool_result",
			ToolUseID: "calc_001",
			Content:   json.RawMessage(`{"result": 4}`),
		},
		{
			Type:      "tool_result",
			ToolUseID: "search_001",
			Error:     &errorMsg,
		},
	}

	msg := h.CreateToolResultMessage(results)

	if msg.Role != "user" {
		t.Errorf("Expected role to be 'user', got %s", msg.Role)
	}

	if len(msg.Content) != 2 {
		t.Fatalf("Expected 2 content blocks, got %d", len(msg.Content))
	}

	// Check first result
	if msg.Content[0].Type != "tool_result" {
		t.Errorf("Expected type 'tool_result', got %s", msg.Content[0].Type)
	}
	if msg.Content[0].ToolUseID != "calc_001" {
		t.Errorf("Expected tool_use_id 'calc_001', got %s", msg.Content[0].ToolUseID)
	}

	// Check error result
	if msg.Content[1].ToolUseID != "search_001" {
		t.Errorf("Expected tool_use_id 'search_001', got %s", msg.Content[1].ToolUseID)
	}
	
	var errorContent map[string]string
	if err := json.Unmarshal(msg.Content[1].Content, &errorContent); err != nil {
		t.Fatalf("Failed to unmarshal error content: %v", err)
	}
	if errorContent["error"] != errorMsg {
		t.Errorf("Expected error message %q, got %q", errorMsg, errorContent["error"])
	}
}

func TestValidateToolDefinition(t *testing.T) {
	h := NewHandler()

	tests := []struct {
		name    string
		tool    interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid tool",
			tool: map[string]interface{}{
				"name":        "calculator",
				"description": "Perform mathematical calculations",
				"input_schema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"expression": map[string]interface{}{
							"type": "string",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			tool: map[string]interface{}{
				"description": "A tool without a name",
			},
			wantErr: true,
			errMsg:  "must have a non-empty name",
		},
		{
			name: "missing description",
			tool: map[string]interface{}{
				"name": "tool",
			},
			wantErr: true,
			errMsg:  "must have a non-empty description",
		},
		{
			name:    "non-map tool",
			tool:    "not a tool",
			wantErr: true,
			errMsg:  "must be an object",
		},
		{
			name: "invalid input_schema",
			tool: map[string]interface{}{
				"name":         "tool",
				"description":  "A tool",
				"input_schema": "not an object",
			},
			wantErr: true,
			errMsg:  "input_schema must be an object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := h.ValidateToolDefinition(tt.tool)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToolDefinition() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
				t.Errorf("Expected error containing %q, got %q", tt.errMsg, err.Error())
			}
		})
	}
}

func TestTransformToolsForProvider(t *testing.T) {
	h := NewHandler()

	anthropicTools := []interface{}{
		map[string]interface{}{
			"name":        "calculator",
			"description": "Perform calculations",
			"input_schema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"expression": map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
	}

	tests := []struct {
		name     string
		provider string
		validate func(t *testing.T, result []interface{})
	}{
		{
			name:     "anthropic",
			provider: "anthropic",
			validate: func(t *testing.T, result []interface{}) {
				if len(result) != 1 {
					t.Fatal("Expected 1 tool")
				}
				// Should remain unchanged
				tool := result[0].(map[string]interface{})
				if tool["name"] != "calculator" {
					t.Error("Tool should remain unchanged for Anthropic")
				}
			},
		},
		{
			name:     "openai",
			provider: "openai",
			validate: func(t *testing.T, result []interface{}) {
				if len(result) != 1 {
					t.Fatal("Expected 1 tool")
				}
				tool := result[0].(map[string]interface{})
				if tool["type"] != "function" {
					t.Error("Expected type to be 'function' for OpenAI")
				}
				function := tool["function"].(map[string]interface{})
				if function["name"] != "calculator" {
					t.Error("Expected function name to be preserved")
				}
				if function["parameters"] == nil {
					t.Error("Expected parameters to be set from input_schema")
				}
			},
		},
		{
			name:     "google",
			provider: "google",
			validate: func(t *testing.T, result []interface{}) {
				if len(result) != 1 {
					t.Fatal("Expected 1 tool")
				}
				tool := result[0].(map[string]interface{})
				if tool["name"] != "calculator" {
					t.Error("Expected name to be preserved")
				}
				if tool["parameters"] == nil {
					t.Error("Expected parameters to be set")
				}
			},
		},
		{
			name:     "unknown",
			provider: "unknown-provider",
			validate: func(t *testing.T, result []interface{}) {
				// Should return unchanged
				if len(result) != 1 {
					t.Fatal("Expected 1 tool")
				}
				tool := result[0].(map[string]interface{})
				if tool["name"] != "calculator" {
					t.Error("Tool should remain unchanged for unknown provider")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := h.TransformToolsForProvider(anthropicTools, tt.provider)
			if err != nil {
				t.Fatalf("TransformToolsForProvider() error = %v", err)
			}
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestTransformToolsForProviderEmpty(t *testing.T) {
	h := NewHandler()

	// Test with nil tools
	result, err := h.TransformToolsForProvider(nil, "openai")
	if err != nil {
		t.Errorf("Expected no error for nil tools, got %v", err)
	}
	if result != nil {
		t.Error("Expected nil result for nil input")
	}

	// Test with empty slice
	result, err = h.TransformToolsForProvider([]interface{}{}, "openai")
	if err != nil {
		t.Errorf("Expected no error for empty tools, got %v", err)
	}
	if len(result) != 0 {
		t.Error("Expected empty result for empty input")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && contains(s[1:], substr)
}