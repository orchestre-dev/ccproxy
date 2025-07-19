package transformer

import (
	"context"
	"testing"

	"github.com/musistudio/ccproxy/internal/config"
)

func TestToolTransformer_TransformRequest(t *testing.T) {
	transformer := NewToolTransformer()
	ctx := context.Background()

	tests := []struct {
		name     string
		provider *config.Provider
		request  interface{}
		validate func(t *testing.T, result interface{})
		wantErr  bool
	}{
		{
			name: "request without tools",
			provider: &config.Provider{
				Name: "test",
			},
			request: map[string]interface{}{
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
				},
			},
			validate: func(t *testing.T, result interface{}) {
				req := result.(map[string]interface{})
				if _, hasTools := req["tools"]; hasTools {
					t.Error("Expected no tools field")
				}
			},
		},
		{
			name: "request with valid tools",
			provider: &config.Provider{
				Name: "anthropic",
			},
			request: map[string]interface{}{
				"tools": []interface{}{
					map[string]interface{}{
						"name":        "calculator",
						"description": "Perform calculations",
						"input_schema": map[string]interface{}{
							"type": "object",
						},
					},
				},
			},
			validate: func(t *testing.T, result interface{}) {
				req := result.(map[string]interface{})
				tools := req["tools"].([]interface{})
				if len(tools) != 1 {
					t.Error("Expected tools to be preserved")
				}
			},
		},
		{
			name: "request with invalid tool",
			provider: &config.Provider{
				Name: "test",
			},
			request: map[string]interface{}{
				"tools": []interface{}{
					map[string]interface{}{
						// Missing required fields
					},
				},
			},
			wantErr: true,
		},
		{
			name: "transform tools for OpenAI",
			provider: &config.Provider{
				Name: "openai",
			},
			request: map[string]interface{}{
				"model": "gpt-4",
				"tools": []interface{}{
					map[string]interface{}{
						"name":        "search",
						"description": "Search the web",
						"input_schema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"query": map[string]interface{}{
									"type": "string",
								},
							},
						},
					},
				},
			},
			validate: func(t *testing.T, result interface{}) {
				req := result.(map[string]interface{})
				tools := req["tools"].([]interface{})
				tool := tools[0].(map[string]interface{})
				if tool["type"] != "function" {
					t.Error("Expected OpenAI tool format")
				}
				if _, hasFunction := tool["function"]; !hasFunction {
					t.Error("Expected function field")
				}
			},
		},
		{
			name: "transform tools for legacy OpenAI model",
			provider: &config.Provider{
				Name: "openai",
			},
			request: map[string]interface{}{
				"model": "gpt-3.5-turbo-0613",
				"tools": []interface{}{
					map[string]interface{}{
						"name":        "search",
						"description": "Search the web",
						"input_schema": map[string]interface{}{
							"type": "object",
						},
					},
				},
			},
			validate: func(t *testing.T, result interface{}) {
				req := result.(map[string]interface{})
				if _, hasTools := req["tools"]; hasTools {
					t.Error("Expected tools to be removed for legacy model")
				}
				functions, hasFunctions := req["functions"].([]interface{})
				if !hasFunctions {
					t.Error("Expected functions field for legacy model")
				}
				if len(functions) != 1 {
					t.Error("Expected function to be converted")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := transformer.TransformRequest(ctx, tt.provider, tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransformRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.validate != nil && !tt.wantErr {
				tt.validate(t, result)
			}
		})
	}
}

func TestToolTransformer_TransformResponse(t *testing.T) {
	transformer := NewToolTransformer()
	ctx := context.Background()

	tests := []struct {
		name     string
		provider *config.Provider
		response interface{}
		validate func(t *testing.T, result interface{})
	}{
		{
			name: "response with tool use",
			provider: &config.Provider{
				Name: "test",
			},
			response: map[string]interface{}{
				"content": []interface{}{
					map[string]interface{}{
						"type": "tool_use",
						"id":   "tool_123",
						"name": "calculator",
					},
				},
			},
			validate: func(t *testing.T, result interface{}) {
				resp := result.(map[string]interface{})
				// Check metadata was added
				metadata, ok := resp["metadata"].(map[string]interface{})
				if !ok || metadata["has_tool_use"] != true {
					t.Error("Expected tool use metadata")
				}
				// Check processing marker
				content := resp["content"].([]interface{})
				toolBlock := content[0].(map[string]interface{})
				if toolBlock["_processed"] != true {
					t.Error("Expected tool block to be marked as processed")
				}
			},
		},
		{
			name: "response without tool use",
			provider: &config.Provider{
				Name: "test",
			},
			response: map[string]interface{}{
				"content": "Just text",
			},
			validate: func(t *testing.T, result interface{}) {
				resp := result.(map[string]interface{})
				if resp["content"] != "Just text" {
					t.Error("Expected content to remain unchanged")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := transformer.TransformResponse(ctx, tt.provider, tt.response)
			if err != nil {
				t.Fatalf("TransformResponse() error = %v", err)
			}
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestIsLegacyOpenAIModel(t *testing.T) {
	tests := []struct {
		model  string
		legacy bool
	}{
		{"gpt-3.5-turbo-0613", true},
		{"gpt-3.5-turbo-16k-0613", true},
		{"gpt-4-0613", true},
		{"gpt-4-32k-0613", true},
		{"gpt-4", false},
		{"gpt-3.5-turbo", false},
		{"gpt-4o", false},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			result := isLegacyOpenAIModel(tt.model)
			if result != tt.legacy {
				t.Errorf("isLegacyOpenAIModel(%s) = %v, want %v", tt.model, result, tt.legacy)
			}
		})
	}
}