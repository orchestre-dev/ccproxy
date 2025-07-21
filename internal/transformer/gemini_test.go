package transformer

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
)

func TestGeminiTransformer(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	t.Run("NewGeminiTransformer", func(t *testing.T) {
		transformer := NewGeminiTransformer()
		testutil.AssertEqual(t, "gemini", transformer.GetName())
		testutil.AssertEqual(t, "/v1beta/models/:modelAndAction", transformer.GetEndpoint())
	})
}

func TestGeminiTransformRequestIn(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewGeminiTransformer()
	ctx := context.Background()

	t.Run("BasicMessageTransformation", func(t *testing.T) {
		request := map[string]interface{}{
			"model": "gemini-pro",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "system",
					"content": "You are a helpful assistant",
				},
				map[string]interface{}{
					"role":    "user",
					"content": "Hello",
				},
				map[string]interface{}{
					"role":    "assistant",
					"content": "Hi there!",
				},
			},
		}

		result, err := transformer.TransformRequestIn(ctx, request, "gemini")
		testutil.AssertNoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		testutil.AssertEqual(t, true, ok)

		// Should have systemInstruction
		testutil.AssertEqual(t, "You are a helpful assistant", resultMap["systemInstruction"])

		// Should have contents instead of messages
		contents, ok := resultMap["contents"].([]interface{})
		testutil.AssertEqual(t, true, ok)
		testutil.AssertEqual(t, 2, len(contents)) // system message excluded from contents

		// Check user message transformation
		userContent := contents[0].(map[string]interface{})
		testutil.AssertEqual(t, "user", userContent["role"])

		userParts := userContent["parts"].([]interface{})
		testutil.AssertEqual(t, 1, len(userParts))
		userPart := userParts[0].(map[string]interface{})
		testutil.AssertEqual(t, "Hello", userPart["text"])

		// Check assistant message transformation (role becomes "model")
		assistantContent := contents[1].(map[string]interface{})
		testutil.AssertEqual(t, "model", assistantContent["role"])

		assistantParts := assistantContent["parts"].([]interface{})
		testutil.AssertEqual(t, 1, len(assistantParts))
		assistantPart := assistantParts[0].(map[string]interface{})
		testutil.AssertEqual(t, "Hi there!", assistantPart["text"])
	})

	t.Run("ToolCallTransformation", func(t *testing.T) {
		request := map[string]interface{}{
			"model": "gemini-pro",
			"messages": []interface{}{
				map[string]interface{}{
					"role": "assistant",
					"tool_calls": []interface{}{
						map[string]interface{}{
							"id":   "call_123",
							"type": "function",
							"function": map[string]interface{}{
								"name":      "get_weather",
								"arguments": `{"location": "New York"}`,
							},
						},
					},
				},
			},
		}

		result, err := transformer.TransformRequestIn(ctx, request, "gemini")
		testutil.AssertNoError(t, err)

		resultMap := result.(map[string]interface{})
		contents := resultMap["contents"].([]interface{})
		testutil.AssertEqual(t, 1, len(contents))

		content := contents[0].(map[string]interface{})
		testutil.AssertEqual(t, "model", content["role"])

		parts := content["parts"].([]interface{})
		testutil.AssertEqual(t, 1, len(parts))

		part := parts[0].(map[string]interface{})
		functionCall := part["functionCall"].(map[string]interface{})
		testutil.AssertEqual(t, "get_weather", functionCall["name"])

		args := functionCall["args"].(map[string]interface{})
		testutil.AssertEqual(t, "New York", args["location"])
	})

	t.Run("ToolResponseTransformation", func(t *testing.T) {
		request := map[string]interface{}{
			"model": "gemini-pro",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "tool",
					"name":    "get_weather",
					"content": "Sunny, 72°F",
				},
			},
		}

		result, err := transformer.TransformRequestIn(ctx, request, "gemini")
		testutil.AssertNoError(t, err)

		resultMap := result.(map[string]interface{})
		contents := resultMap["contents"].([]interface{})

		// Find the tool response message
		var toolContent map[string]interface{}
		for _, cont := range contents {
			contentMap := cont.(map[string]interface{})
			if contentMap["role"] == "user" {
				parts := contentMap["parts"].([]interface{})
				for _, part := range parts {
					partMap := part.(map[string]interface{})
					if _, ok := partMap["functionResponse"]; ok {
						toolContent = contentMap
						break
					}
				}
			}
		}

		testutil.AssertNotEqual(t, nil, toolContent)
		testutil.AssertEqual(t, "user", toolContent["role"]) // tool becomes user

		parts := toolContent["parts"].([]interface{})
		testutil.AssertEqual(t, true, len(parts) >= 1)

		// Find the function response part
		var functionResponse map[string]interface{}
		for _, p := range parts {
			if partMap, ok := p.(map[string]interface{}); ok {
				if fr, exists := partMap["functionResponse"]; exists {
					functionResponse = fr.(map[string]interface{})
					break
				}
			}
		}

		testutil.AssertNotEqual(t, nil, functionResponse)
		testutil.AssertEqual(t, "get_weather", functionResponse["name"])

		response := functionResponse["response"].(map[string]interface{})
		testutil.AssertEqual(t, "Sunny, 72°F", response["result"])
	})

	t.Run("GenerationConfigTransformation", func(t *testing.T) {
		request := map[string]interface{}{
			"model":       "gemini-pro",
			"messages":    []interface{}{},
			"temperature": 0.7,
			"max_tokens":  1000,
			"top_p":       0.9,
		}

		result, err := transformer.TransformRequestIn(ctx, request, "gemini")
		testutil.AssertNoError(t, err)

		resultMap := result.(map[string]interface{})
		genConfig := resultMap["generationConfig"].(map[string]interface{})

		testutil.AssertEqual(t, 0.7, genConfig["temperature"])
		testutil.AssertEqual(t, 1000, genConfig["maxOutputTokens"])
		testutil.AssertEqual(t, 0.9, genConfig["topP"])
	})

	t.Run("ToolsTransformation", func(t *testing.T) {
		request := map[string]interface{}{
			"model":    "gemini-pro",
			"messages": []interface{}{},
			"tools": []interface{}{
				map[string]interface{}{
					"type": "function",
					"function": map[string]interface{}{
						"name":        "get_weather",
						"description": "Get weather information",
						"parameters": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"location": map[string]interface{}{
									"type":        "string",
									"description": "City name",
								},
							},
							"required":             []interface{}{"location"},
							"$schema":              "http://json-schema.org/draft-07/schema#",
							"additionalProperties": false,
						},
					},
				},
			},
		}

		result, err := transformer.TransformRequestIn(ctx, request, "gemini")
		testutil.AssertNoError(t, err)

		resultMap := result.(map[string]interface{})
		tools := resultMap["tools"].([]interface{})
		testutil.AssertEqual(t, 1, len(tools))

		tool := tools[0].(map[string]interface{})
		funcDeclarations := tool["function_declarations"].([]interface{})
		testutil.AssertEqual(t, 1, len(funcDeclarations))

		funcDecl := funcDeclarations[0].(map[string]interface{})
		testutil.AssertEqual(t, "get_weather", funcDecl["name"])
		testutil.AssertEqual(t, "Get weather information", funcDecl["description"])

		// Check that unsupported fields are removed
		params := funcDecl["parameters"].(map[string]interface{})
		_, hasSchema := params["$schema"]
		testutil.AssertEqual(t, false, hasSchema)
		_, hasAdditionalProps := params["additionalProperties"]
		testutil.AssertEqual(t, false, hasAdditionalProps)

		// Supported fields should remain
		testutil.AssertEqual(t, "object", params["type"])
		_, hasProperties := params["properties"]
		testutil.AssertEqual(t, true, hasProperties)
		_, hasRequired := params["required"]
		testutil.AssertEqual(t, true, hasRequired)
	})

	t.Run("InvalidRequestFormat", func(t *testing.T) {
		request := "invalid request"

		_, err := transformer.TransformRequestIn(ctx, request, "gemini")
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "invalid request format")
	})

	t.Run("MissingMessages", func(t *testing.T) {
		request := map[string]interface{}{
			"model": "gemini-pro",
		}

		_, err := transformer.TransformRequestIn(ctx, request, "gemini")
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "missing or invalid messages field")
	})
}

func TestGeminiCleanJSONSchema(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewGeminiTransformer()

	t.Run("RemoveUnsupportedFields", func(t *testing.T) {
		schema := map[string]interface{}{
			"type":                 "object",
			"$schema":              "http://json-schema.org/draft-07/schema#",
			"additionalProperties": false,
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":                 "string",
					"$schema":              "should be removed",
					"additionalProperties": true,
				},
			},
			"required": []interface{}{"name"},
		}

		cleaned := transformer.cleanJSONSchema(schema)

		// Should remove unsupported top-level fields
		_, hasSchema := cleaned["$schema"]
		testutil.AssertEqual(t, false, hasSchema)
		_, hasAdditionalProps := cleaned["additionalProperties"]
		testutil.AssertEqual(t, false, hasAdditionalProps)

		// Should keep supported fields
		testutil.AssertEqual(t, "object", cleaned["type"])
		_, hasRequired := cleaned["required"]
		testutil.AssertEqual(t, true, hasRequired)

		// Should recursively clean properties
		properties := cleaned["properties"].(map[string]interface{})
		nameProperty := properties["name"].(map[string]interface{})
		testutil.AssertEqual(t, "string", nameProperty["type"])
		_, hasNestedSchema := nameProperty["$schema"]
		testutil.AssertEqual(t, false, hasNestedSchema)
		_, hasNestedAdditionalProps := nameProperty["additionalProperties"]
		testutil.AssertEqual(t, false, hasNestedAdditionalProps)
	})

	t.Run("PreserveValidFields", func(t *testing.T) {
		schema := map[string]interface{}{
			"type":        "object",
			"description": "Test schema",
			"properties": map[string]interface{}{
				"field1": map[string]interface{}{
					"type":        "string",
					"description": "First field",
				},
			},
		}

		cleaned := transformer.cleanJSONSchema(schema)

		testutil.AssertEqual(t, schema["type"], cleaned["type"])
		testutil.AssertEqual(t, schema["description"], cleaned["description"])

		// Check properties are preserved (compare field by field)
		originalProps := schema["properties"].(map[string]interface{})
		cleanedProps := cleaned["properties"].(map[string]interface{})
		testutil.AssertEqual(t, len(originalProps), len(cleanedProps))

		for key := range originalProps {
			_, exists := cleanedProps[key]
			testutil.AssertEqual(t, true, exists)
		}
	})
}

func TestGeminiTransformResponseOut(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewGeminiTransformer()
	ctx := context.Background()

	t.Run("NonStreamingResponse", func(t *testing.T) {
		geminiResponse := map[string]interface{}{
			"candidates": []interface{}{
				map[string]interface{}{
					"content": map[string]interface{}{
						"parts": []interface{}{
							map[string]interface{}{
								"text": "Hello, how can I help you?",
							},
						},
					},
					"finishReason": "STOP",
				},
			},
			"usageMetadata": map[string]interface{}{
				"promptTokenCount":     float64(10),
				"candidatesTokenCount": float64(15),
				"totalTokenCount":      float64(25),
			},
		}

		responseBody, _ := json.Marshal(geminiResponse)
		resp := &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(string(responseBody))),
		}

		result, err := transformer.TransformResponseOut(ctx, resp)
		testutil.AssertNoError(t, err)

		body, _ := io.ReadAll(result.Body)
		var transformedResp map[string]interface{}
		json.Unmarshal(body, &transformedResp)

		// Check OpenAI format
		testutil.AssertEqual(t, "chat.completion", transformedResp["object"])
		testutil.AssertEqual(t, "gemini-pro", transformedResp["model"])

		// Check choices
		choices := transformedResp["choices"].([]interface{})
		testutil.AssertEqual(t, 1, len(choices))

		choice := choices[0].(map[string]interface{})
		testutil.AssertEqual(t, 0, int(choice["index"].(float64)))
		testutil.AssertEqual(t, "stop", choice["finish_reason"])

		message := choice["message"].(map[string]interface{})
		testutil.AssertEqual(t, "assistant", message["role"])
		testutil.AssertEqual(t, "Hello, how can I help you?", message["content"])

		// Check usage
		usage := transformedResp["usage"].(map[string]interface{})
		testutil.AssertEqual(t, 10, int(usage["prompt_tokens"].(float64)))
		testutil.AssertEqual(t, 15, int(usage["completion_tokens"].(float64)))
		testutil.AssertEqual(t, 25, int(usage["total_tokens"].(float64)))
	})

	t.Run("ResponseWithFunctionCall", func(t *testing.T) {
		geminiResponse := map[string]interface{}{
			"candidates": []interface{}{
				map[string]interface{}{
					"content": map[string]interface{}{
						"parts": []interface{}{
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
		}

		responseBody, _ := json.Marshal(geminiResponse)
		resp := &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(string(responseBody))),
		}

		result, err := transformer.TransformResponseOut(ctx, resp)
		testutil.AssertNoError(t, err)

		body, _ := io.ReadAll(result.Body)
		var transformedResp map[string]interface{}
		json.Unmarshal(body, &transformedResp)

		choices := transformedResp["choices"].([]interface{})
		choice := choices[0].(map[string]interface{})
		message := choice["message"].(map[string]interface{})

		toolCalls := message["tool_calls"].([]interface{})
		testutil.AssertEqual(t, 1, len(toolCalls))

		toolCall := toolCalls[0].(map[string]interface{})
		testutil.AssertEqual(t, "function", toolCall["type"])

		function := toolCall["function"].(map[string]interface{})
		testutil.AssertEqual(t, "get_weather", function["name"])

		// Arguments should be JSON string
		args := function["arguments"].(string)
		var parsedArgs map[string]interface{}
		json.Unmarshal([]byte(args), &parsedArgs)
		testutil.AssertEqual(t, "San Francisco", parsedArgs["location"])
	})

	t.Run("StreamingResponse", func(t *testing.T) {
		streamData := `data: {"candidates": [{"content": {"parts": [{"text": "Hello"}]}}]}

data: [DONE]

`
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(streamData)),
		}

		result, err := transformer.TransformResponseOut(ctx, resp)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, "text/event-stream", result.Header.Get("Content-Type"))

		// Read the transformed stream
		reader := NewSSEReader(result.Body)
		events := []string{}

		for {
			event, err := reader.ReadEvent()
			if err == io.EOF {
				break
			}
			testutil.AssertNoError(t, err)

			if event.Data != "" && event.Data != "[DONE]" {
				events = append(events, event.Data)
			}
		}

		// Should have transformed to OpenAI format
		testutil.AssertEqual(t, true, len(events) > 0)

		var firstEvent map[string]interface{}
		json.Unmarshal([]byte(events[0]), &firstEvent)

		testutil.AssertEqual(t, "chat.completion.chunk", firstEvent["object"])
		testutil.AssertEqual(t, "gemini-pro", firstEvent["model"])

		choices := firstEvent["choices"].([]interface{})
		choice := choices[0].(map[string]interface{})
		delta := choice["delta"].(map[string]interface{})
		testutil.AssertEqual(t, "Hello", delta["content"])
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("invalid json")),
		}

		result, err := transformer.TransformResponseOut(ctx, resp)
		testutil.AssertNoError(t, err)

		// Should return original response
		body, _ := io.ReadAll(result.Body)
		testutil.AssertEqual(t, "invalid json", string(body))
	})
}

func TestGeminiConvertFinishReason(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewGeminiTransformer()

	testCases := []struct {
		geminiReason string
		openaiReason string
	}{
		{"STOP", "stop"},
		{"MAX_TOKENS", "length"},
		{"SAFETY", "content_filter"},
		{"RECITATION", "content_filter"},
		{"UNKNOWN_REASON", "stop"},
	}

	for _, tc := range testCases {
		t.Run(tc.geminiReason, func(t *testing.T) {
			result := transformer.convertFinishReason(tc.geminiReason)
			testutil.AssertEqual(t, tc.openaiReason, result)
		})
	}
}

func TestGeminiTransformStreamEvent(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewGeminiTransformer()

	t.Run("ValidStreamEvent", func(t *testing.T) {
		eventData := map[string]interface{}{
			"candidates": []interface{}{
				map[string]interface{}{
					"content": map[string]interface{}{
						"parts": []interface{}{
							map[string]interface{}{
								"text": "Hello world",
							},
						},
					},
				},
			},
		}

		jsonData, _ := json.Marshal(eventData)
		event := &SSEEvent{Data: string(jsonData)}

		result := transformer.transformStreamEvent(event)
		testutil.AssertNotEqual(t, nil, result)

		var transformedData map[string]interface{}
		json.Unmarshal([]byte(result.Data), &transformedData)

		testutil.AssertEqual(t, "chat.completion.chunk", transformedData["object"])
		testutil.AssertEqual(t, "gemini-pro", transformedData["model"])

		choices := transformedData["choices"].([]interface{})
		choice := choices[0].(map[string]interface{})
		delta := choice["delta"].(map[string]interface{})
		testutil.AssertEqual(t, "Hello world", delta["content"])
	})

	t.Run("EventWithFunctionCall", func(t *testing.T) {
		eventData := map[string]interface{}{
			"candidates": []interface{}{
				map[string]interface{}{
					"content": map[string]interface{}{
						"parts": []interface{}{
							map[string]interface{}{
								"functionCall": map[string]interface{}{
									"name": "test_function",
									"args": map[string]interface{}{
										"param": "value",
									},
								},
							},
						},
					},
				},
			},
		}

		jsonData, _ := json.Marshal(eventData)
		event := &SSEEvent{Data: string(jsonData)}

		result := transformer.transformStreamEvent(event)
		testutil.AssertNotEqual(t, nil, result)

		var transformedData map[string]interface{}
		json.Unmarshal([]byte(result.Data), &transformedData)

		choices := transformedData["choices"].([]interface{})
		choice := choices[0].(map[string]interface{})
		delta := choice["delta"].(map[string]interface{})

		toolCalls := delta["tool_calls"].([]interface{})
		testutil.AssertEqual(t, 1, len(toolCalls))

		toolCall := toolCalls[0].(map[string]interface{})
		testutil.AssertEqual(t, 0, int(toolCall["index"].(float64)))
		testutil.AssertEqual(t, "function", toolCall["type"])

		function := toolCall["function"].(map[string]interface{})
		testutil.AssertEqual(t, "test_function", function["name"])
	})

	t.Run("InvalidEventData", func(t *testing.T) {
		event := &SSEEvent{Data: "invalid json"}

		result := transformer.transformStreamEvent(event)
		testutil.AssertEqual(t, (*SSEEvent)(nil), result)
	})
}

func TestGeminiIntegration(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewGeminiTransformer()
	ctx := context.Background()

	t.Run("CompleteWorkflow", func(t *testing.T) {
		// Test request transformation
		request := map[string]interface{}{
			"model": "gemini-pro",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "system",
					"content": "You are helpful",
				},
				map[string]interface{}{
					"role":    "user",
					"content": "Hello",
				},
			},
			"temperature": 0.7,
			"max_tokens":  1000,
			"tools": []interface{}{
				map[string]interface{}{
					"type": "function",
					"function": map[string]interface{}{
						"name":        "test_tool",
						"description": "A test tool",
						"parameters": map[string]interface{}{
							"type":    "object",
							"$schema": "should be removed",
						},
					},
				},
			},
		}

		transformedReq, err := transformer.TransformRequestIn(ctx, request, "gemini")
		testutil.AssertNoError(t, err)

		reqMap := transformedReq.(map[string]interface{})

		// Check system instruction extraction
		testutil.AssertEqual(t, "You are helpful", reqMap["systemInstruction"])

		// Check contents structure
		contents := reqMap["contents"].([]interface{})
		testutil.AssertEqual(t, 1, len(contents)) // Only user message in contents

		// Check generation config
		genConfig := reqMap["generationConfig"].(map[string]interface{})
		testutil.AssertEqual(t, 0.7, genConfig["temperature"])
		testutil.AssertEqual(t, 1000, genConfig["maxOutputTokens"])

		// Check tools transformation and schema cleaning
		tools := reqMap["tools"].([]interface{})
		tool := tools[0].(map[string]interface{})
		funcDecls := tool["function_declarations"].([]interface{})
		funcDecl := funcDecls[0].(map[string]interface{})
		params := funcDecl["parameters"].(map[string]interface{})
		_, hasSchema := params["$schema"]
		testutil.AssertEqual(t, false, hasSchema) // Should be cleaned

		// Test response transformation
		geminiResponse := map[string]interface{}{
			"candidates": []interface{}{
				map[string]interface{}{
					"content": map[string]interface{}{
						"parts": []interface{}{
							map[string]interface{}{
								"text": "Hello! How can I help you?",
							},
						},
					},
					"finishReason": "STOP",
				},
			},
		}

		responseBody, _ := json.Marshal(geminiResponse)
		resp := &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(string(responseBody))),
		}

		transformedResp, err := transformer.TransformResponseOut(ctx, resp)
		testutil.AssertNoError(t, err)

		body, _ := io.ReadAll(transformedResp.Body)
		var openaiResp map[string]interface{}
		json.Unmarshal(body, &openaiResp)

		testutil.AssertEqual(t, "chat.completion", openaiResp["object"])
		choices := openaiResp["choices"].([]interface{})
		choice := choices[0].(map[string]interface{})
		message := choice["message"].(map[string]interface{})
		testutil.AssertEqual(t, "Hello! How can I help you?", message["content"])
	})
}
