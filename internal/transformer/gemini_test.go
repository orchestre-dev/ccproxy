package transformer

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestGeminiTransformer_TransformRequestIn(t *testing.T) {
	transformer := NewGeminiTransformer()
	ctx := context.Background()

	tests := []struct {
		name     string
		input    map[string]interface{}
		validate func(t *testing.T, result interface{})
		wantErr  bool
	}{
		{
			name: "basic message transformation",
			input: map[string]interface{}{
				"model":      "gemini-pro",
				"max_tokens": 1000,
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
					map[string]interface{}{
						"role":    "assistant",
						"content": "Hi there!",
					},
				},
			},
			validate: func(t *testing.T, result interface{}) {
				res := result.(map[string]interface{})
				
				// Check contents
				contents, ok := res["contents"].([]interface{})
				if !ok || len(contents) != 2 {
					t.Errorf("Expected 2 contents, got %v", contents)
				}
				
				// Check role transformation
				msg1 := contents[0].(map[string]interface{})
				if msg1["role"] != "user" {
					t.Errorf("Expected user role, got %v", msg1["role"])
				}
				
				msg2 := contents[1].(map[string]interface{})
				if msg2["role"] != "model" {
					t.Errorf("Expected model role for assistant, got %v", msg2["role"])
				}
				
				// Check generation config
				genConfig, ok := res["generationConfig"].(map[string]interface{})
				if !ok {
					t.Error("Expected generationConfig")
				}
				if genConfig["maxOutputTokens"] != 1000 {
					t.Errorf("Expected maxOutputTokens 1000, got %v", genConfig["maxOutputTokens"])
				}
			},
		},
		{
			name: "system message extraction",
			input: map[string]interface{}{
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "system",
						"content": "You are a helpful assistant",
					},
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
				},
			},
			validate: func(t *testing.T, result interface{}) {
				res := result.(map[string]interface{})
				
				// Check system instruction
				if res["systemInstruction"] != "You are a helpful assistant" {
					t.Errorf("Expected system instruction, got %v", res["systemInstruction"])
				}
				
				// Check contents don't include system message
				contents := res["contents"].([]interface{})
				if len(contents) != 1 {
					t.Errorf("Expected 1 content (system excluded), got %d", len(contents))
				}
			},
		},
		{
			name: "tool transformation",
			input: map[string]interface{}{
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "What's the weather?",
					},
				},
				"tools": []interface{}{
					map[string]interface{}{
						"type": "function",
						"function": map[string]interface{}{
							"name":        "get_weather",
							"description": "Get weather info",
							"parameters": map[string]interface{}{
								"type":                 "object",
								"$schema":              "http://json-schema.org/draft-07/schema#",
								"additionalProperties": false,
								"properties": map[string]interface{}{
									"location": map[string]interface{}{
										"type": "string",
									},
								},
								"required": []interface{}{"location"},
							},
						},
					},
				},
			},
			validate: func(t *testing.T, result interface{}) {
				res := result.(map[string]interface{})
				
				// Check tools transformation
				tools, ok := res["tools"].([]interface{})
				if !ok || len(tools) != 1 {
					t.Errorf("Expected 1 tool, got %v", tools)
				}
				
				tool := tools[0].(map[string]interface{})
				funcDecls, ok := tool["function_declarations"].([]interface{})
				if !ok || len(funcDecls) != 1 {
					t.Error("Expected function_declarations")
				}
				
				// Check schema cleaning
				funcDecl := funcDecls[0].(map[string]interface{})
				params := funcDecl["parameters"].(map[string]interface{})
				
				// Should not have $schema or additionalProperties
				if _, hasSchema := params["$schema"]; hasSchema {
					t.Error("$schema should be removed")
				}
				if _, hasAdditional := params["additionalProperties"]; hasAdditional {
					t.Error("additionalProperties should be removed")
				}
				
				// Should still have other fields
				if params["type"] != "object" {
					t.Error("type should be preserved")
				}
			},
		},
		{
			name: "tool call transformation",
			input: map[string]interface{}{
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "assistant",
						"content": "I'll check the weather",
						"tool_calls": []interface{}{
							map[string]interface{}{
								"id":   "call_123",
								"type": "function",
								"function": map[string]interface{}{
									"name":      "get_weather",
									"arguments": `{"location": "San Francisco"}`,
								},
							},
						},
					},
				},
			},
			validate: func(t *testing.T, result interface{}) {
				res := result.(map[string]interface{})
				contents := res["contents"].([]interface{})
				content := contents[0].(map[string]interface{})
				parts := content["parts"].([]interface{})
				
				// Should have text and function call parts
				if len(parts) != 2 {
					t.Errorf("Expected 2 parts, got %d", len(parts))
				}
				
				// Check function call part
				var foundFuncCall bool
				for _, part := range parts {
					partMap := part.(map[string]interface{})
					if funcCall, ok := partMap["functionCall"].(map[string]interface{}); ok {
						foundFuncCall = true
						if funcCall["name"] != "get_weather" {
							t.Errorf("Expected get_weather, got %v", funcCall["name"])
						}
						args := funcCall["args"].(map[string]interface{})
						if args["location"] != "San Francisco" {
							t.Errorf("Expected San Francisco, got %v", args["location"])
						}
					}
				}
				
				if !foundFuncCall {
					t.Error("Function call not found in parts")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := transformer.TransformRequestIn(ctx, tt.input, "gemini")
			if (err != nil) != tt.wantErr {
				t.Errorf("TransformRequestIn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestGeminiTransformer_TransformResponseOut(t *testing.T) {
	transformer := NewGeminiTransformer()
	ctx := context.Background()

	tests := []struct {
		name     string
		input    map[string]interface{}
		validate func(t *testing.T, result map[string]interface{})
	}{
		{
			name: "basic text response",
			input: map[string]interface{}{
				"candidates": []interface{}{
					map[string]interface{}{
						"content": map[string]interface{}{
							"parts": []interface{}{
								map[string]interface{}{
									"text": "Hello from Gemini!",
								},
							},
							"role": "model",
						},
						"finishReason": "STOP",
					},
				},
				"usageMetadata": map[string]interface{}{
					"promptTokenCount":     10.0,
					"candidatesTokenCount": 15.0,
					"totalTokenCount":      25.0,
				},
			},
			validate: func(t *testing.T, result map[string]interface{}) {
				// Check basic structure
				if result["object"] != "chat.completion" {
					t.Errorf("Expected chat.completion, got %v", result["object"])
				}
				
				// Check choices
				choices := result["choices"].([]interface{})
				if len(choices) != 1 {
					t.Fatalf("Expected 1 choice, got %d", len(choices))
				}
				
				choice := choices[0].(map[string]interface{})
				message := choice["message"].(map[string]interface{})
				
				if message["role"] != "assistant" {
					t.Errorf("Expected assistant role, got %v", message["role"])
				}
				if message["content"] != "Hello from Gemini!" {
					t.Errorf("Expected content, got %v", message["content"])
				}
				if choice["finish_reason"] != "stop" {
					t.Errorf("Expected stop, got %v", choice["finish_reason"])
				}
				
				// Check usage
				usage := result["usage"].(map[string]interface{})
				if usage["prompt_tokens"].(float64) != 10 {
					t.Errorf("Expected 10 prompt tokens, got %v", usage["prompt_tokens"])
				}
				if usage["completion_tokens"].(float64) != 15 {
					t.Errorf("Expected 15 completion tokens, got %v", usage["completion_tokens"])
				}
			},
		},
		{
			name: "function call response",
			input: map[string]interface{}{
				"candidates": []interface{}{
					map[string]interface{}{
						"content": map[string]interface{}{
							"parts": []interface{}{
								map[string]interface{}{
									"text": "I'll check the weather for you.",
								},
								map[string]interface{}{
									"functionCall": map[string]interface{}{
										"name": "get_weather",
										"args": map[string]interface{}{
											"location": "San Francisco",
										},
									},
								},
							},
						},
						"finishReason": "STOP",
					},
				},
			},
			validate: func(t *testing.T, result map[string]interface{}) {
				choices := result["choices"].([]interface{})
				choice := choices[0].(map[string]interface{})
				message := choice["message"].(map[string]interface{})
				
				// Check content
				if message["content"] != "I'll check the weather for you." {
					t.Errorf("Expected content, got %v", message["content"])
				}
				
				// Check tool calls
				toolCalls, ok := message["tool_calls"].([]interface{})
				if !ok || len(toolCalls) != 1 {
					t.Fatalf("Expected 1 tool call, got %v", toolCalls)
				}
				
				toolCall := toolCalls[0].(map[string]interface{})
				if toolCall["type"] != "function" {
					t.Errorf("Expected function type, got %v", toolCall["type"])
				}
				
				function := toolCall["function"].(map[string]interface{})
				if function["name"] != "get_weather" {
					t.Errorf("Expected get_weather, got %v", function["name"])
				}
				
				// Check arguments are stringified
				args, ok := function["arguments"].(string)
				if !ok {
					t.Errorf("Expected string arguments, got %T", function["arguments"])
				}
				if !strings.Contains(args, "San Francisco") {
					t.Errorf("Expected San Francisco in args, got %v", args)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create response body
			body, _ := json.Marshal(tt.input)
			resp := &http.Response{
				StatusCode: 200,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(string(body))),
			}
			resp.Header.Set("Content-Type", "application/json")

			// Transform response
			transformed, err := transformer.TransformResponseOut(ctx, resp)
			if err != nil {
				t.Fatalf("TransformResponseOut() error = %v", err)
			}

			// Read transformed body
			transformedBody, err := io.ReadAll(transformed.Body)
			if err != nil {
				t.Fatalf("Failed to read transformed body: %v", err)
			}

			// Parse result
			var result map[string]interface{}
			if err := json.Unmarshal(transformedBody, &result); err != nil {
				t.Fatalf("Failed to parse transformed response: %v", err)
			}

			tt.validate(t, result)
		})
	}
}

func TestGeminiTransformer_StreamingResponse(t *testing.T) {
	transformer := NewGeminiTransformer()
	ctx := context.Background()

	// Create streaming response
	events := []string{
		`data: {"candidates":[{"content":{"parts":[{"text":"Hello"}],"role":"model"}}]}`,
		`data: {"candidates":[{"content":{"parts":[{"text":" from"}],"role":"model"}}]}`,
		`data: {"candidates":[{"content":{"parts":[{"text":" Gemini!"}],"role":"model"}}]}`,
		`data: {"candidates":[{"content":{"parts":[]},"role":"model"},"finishReason":"STOP"}]}`,
	}

	sseData := strings.Join(events, "\n\n") + "\n\n"

	resp := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(sseData)),
	}
	resp.Header.Set("Content-Type", "text/event-stream")

	// Transform response
	transformed, err := transformer.TransformResponseOut(ctx, resp)
	if err != nil {
		t.Fatalf("TransformResponseOut() error = %v", err)
	}

	// Read transformed events
	reader := NewSSEReader(transformed.Body)
	var contentParts []string
	var foundDone bool

	for {
		event, err := reader.ReadEvent()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Failed to read transformed event: %v", err)
		}

		if event.Data == "[DONE]" {
			foundDone = true
			continue
		}

		var chunk map[string]interface{}
		if err := json.Unmarshal([]byte(event.Data), &chunk); err == nil {
			if chunk["object"] != "chat.completion.chunk" {
				t.Errorf("Expected chat.completion.chunk, got %v", chunk["object"])
			}

			choices, _ := chunk["choices"].([]interface{})
			if len(choices) > 0 {
				choice := choices[0].(map[string]interface{})
				delta := choice["delta"].(map[string]interface{})
				
				if content, ok := delta["content"].(string); ok {
					contentParts = append(contentParts, content)
				}
			}
		}
	}

	// Verify content
	fullContent := strings.Join(contentParts, "")
	if fullContent != "Hello from Gemini!" {
		t.Errorf("Expected 'Hello from Gemini!', got '%s'", fullContent)
	}

	if !foundDone {
		t.Error("Expected [DONE] event")
	}
}

func TestGeminiTransformer_FinishReasonConversion(t *testing.T) {
	transformer := NewGeminiTransformer()

	tests := []struct {
		input    string
		expected string
	}{
		{"STOP", "stop"},
		{"MAX_TOKENS", "length"},
		{"SAFETY", "content_filter"},
		{"RECITATION", "content_filter"},
		{"OTHER", "stop"},
		{"", "stop"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := transformer.convertFinishReason(tt.input)
			if result != tt.expected {
				t.Errorf("convertFinishReason(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGeminiTransformer_SchemaClean(t *testing.T) {
	transformer := NewGeminiTransformer()

	input := map[string]interface{}{
		"type":                 "object",
		"$schema":              "http://json-schema.org/draft-07/schema#",
		"additionalProperties": false,
		"properties": map[string]interface{}{
			"location": map[string]interface{}{
				"type":                 "string",
				"$schema":              "should also be removed",
				"additionalProperties": true,
			},
			"temperature": map[string]interface{}{
				"type": "number",
			},
		},
		"required": []interface{}{"location"},
	}

	cleaned := transformer.cleanJSONSchema(input)

	// Check top-level cleaning
	if _, hasSchema := cleaned["$schema"]; hasSchema {
		t.Error("$schema should be removed")
	}
	if _, hasAdditional := cleaned["additionalProperties"]; hasAdditional {
		t.Error("additionalProperties should be removed")
	}
	
	// Check nested cleaning
	props := cleaned["properties"].(map[string]interface{})
	locationProp := props["location"].(map[string]interface{})
	
	if _, hasSchema := locationProp["$schema"]; hasSchema {
		t.Error("Nested $schema should be removed")
	}
	if _, hasAdditional := locationProp["additionalProperties"]; hasAdditional {
		t.Error("Nested additionalProperties should be removed")
	}

	// Check preserved fields
	if cleaned["type"] != "object" {
		t.Error("type should be preserved")
	}
	if locationProp["type"] != "string" {
		t.Error("Nested type should be preserved")
	}
}