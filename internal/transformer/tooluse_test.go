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

func TestToolUseTransformer(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	t.Run("NewToolUseTransformer", func(t *testing.T) {
		transformer := NewToolUseTransformer()
		testutil.AssertEqual(t, "tooluse", transformer.GetName())
		testutil.AssertEqual(t, "", transformer.GetEndpoint())
	})
}

func TestToolUseTransformRequestIn(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewToolUseTransformer()
	ctx := context.Background()

	t.Run("ValidRequestTransformation", func(t *testing.T) {
		request := map[string]interface{}{
			"model": "claude-3-sonnet",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Please help me with this task",
				},
			},
			"tools": []interface{}{
				map[string]interface{}{
					"name":        "existing_tool",
					"description": "An existing tool",
				},
			},
		}

		result, err := transformer.TransformRequestIn(ctx, request, "anthropic")
		testutil.AssertNoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		testutil.AssertEqual(t, true, ok)

		// Should add system reminder message
		messages := resultMap["messages"].([]interface{})
		testutil.AssertEqual(t, 2, len(messages)) // original + system reminder

		// Check system reminder
		systemMsg := messages[0].(map[string]interface{})
		testutil.AssertEqual(t, "system", systemMsg["role"])
		content := systemMsg["content"].(string)
		testutil.AssertContains(t, content, "tool mode")
		testutil.AssertContains(t, content, "ExitTool")

		// Check original user message is preserved
		userMsg := messages[1].(map[string]interface{})
		testutil.AssertEqual(t, "user", userMsg["role"])
		testutil.AssertEqual(t, "Please help me with this task", userMsg["content"])

		// Should add ExitTool to tools
		tools := resultMap["tools"].([]interface{})
		testutil.AssertEqual(t, 2, len(tools)) // existing + ExitTool

		// Check ExitTool structure
		var exitTool map[string]interface{}
		for _, tool := range tools {
			toolMap := tool.(map[string]interface{})
			if function, ok := toolMap["function"].(map[string]interface{}); ok {
				if function["name"] == "ExitTool" {
					exitTool = toolMap
					break
				}
			}
		}

		testutil.AssertNotEqual(t, nil, exitTool)
		testutil.AssertEqual(t, "function", exitTool["type"])

		function := exitTool["function"].(map[string]interface{})
		testutil.AssertEqual(t, "ExitTool", function["name"])
		testutil.AssertContains(t, function["description"].(string), "Exit tool mode")

		// Check parameters structure
		params := function["parameters"].(map[string]interface{})
		testutil.AssertEqual(t, "object", params["type"])
		testutil.AssertNotEqual(t, nil, params["properties"])
		testutil.AssertNotEqual(t, nil, params["required"])

		// Should set tool_choice to required
		testutil.AssertEqual(t, "required", resultMap["tool_choice"])
	})

	t.Run("RequestWithNoExistingTools", func(t *testing.T) {
		request := map[string]interface{}{
			"model": "claude-3-sonnet",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Help me",
				},
			},
		}

		result, err := transformer.TransformRequestIn(ctx, request, "anthropic")
		testutil.AssertNoError(t, err)

		resultMap := result.(map[string]interface{})

		// Should still add ExitTool
		tools := resultMap["tools"].([]interface{})
		testutil.AssertEqual(t, 1, len(tools))

		tool := tools[0].(map[string]interface{})
		function := tool["function"].(map[string]interface{})
		testutil.AssertEqual(t, "ExitTool", function["name"])
	})

	t.Run("InvalidRequestFormat", func(t *testing.T) {
		request := "invalid request"

		_, err := transformer.TransformRequestIn(ctx, request, "anthropic")
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "invalid request format")
	})

	t.Run("MissingMessages", func(t *testing.T) {
		request := map[string]interface{}{
			"model": "claude-3-sonnet",
		}

		_, err := transformer.TransformRequestIn(ctx, request, "anthropic")
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "missing or invalid messages field")
	})
}

func TestToolUseTransformResponseOut(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewToolUseTransformer()
	ctx := context.Background()

	t.Run("NonStreamingResponseWithExitTool", func(t *testing.T) {
		response := map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{
					"message": map[string]interface{}{
						"role": "assistant",
						"tool_calls": []interface{}{
							map[string]interface{}{
								"id":   "call_regular",
								"type": "function",
								"function": map[string]interface{}{
									"name": "regular_tool",
								},
							},
							map[string]interface{}{
								"id":   "call_exit",
								"type": "function",
								"function": map[string]interface{}{
									"name": "ExitTool",
								},
							},
						},
					},
				},
			},
		}

		responseBody, _ := json.Marshal(response)
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

		// Should have only one tool call (ExitTool removed)
		toolCalls := message["tool_calls"].([]interface{})
		testutil.AssertEqual(t, 1, len(toolCalls))

		remainingCall := toolCalls[0].(map[string]interface{})
		function := remainingCall["function"].(map[string]interface{})
		testutil.AssertEqual(t, "regular_tool", function["name"])

		// Should have default content added
		testutil.AssertEqual(t, "I have completed the requested task.", message["content"])
	})

	t.Run("NonStreamingResponseOnlyExitTool", func(t *testing.T) {
		response := map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{
					"message": map[string]interface{}{
						"role": "assistant",
						"tool_calls": []interface{}{
							map[string]interface{}{
								"id":   "call_exit",
								"type": "function",
								"function": map[string]interface{}{
									"name": "ExitTool",
								},
							},
						},
					},
				},
			},
		}

		responseBody, _ := json.Marshal(response)
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

		// Should have no tool calls
		_, hasToolCalls := message["tool_calls"]
		testutil.AssertEqual(t, false, hasToolCalls)

		// Should have default content
		testutil.AssertEqual(t, "I have completed the requested task.", message["content"])
	})

	t.Run("NonStreamingResponseWithExistingContent", func(t *testing.T) {
		response := map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "Task completed successfully!",
						"tool_calls": []interface{}{
							map[string]interface{}{
								"id":   "call_exit",
								"type": "function",
								"function": map[string]interface{}{
									"name": "ExitTool",
								},
							},
						},
					},
				},
			},
		}

		responseBody, _ := json.Marshal(response)
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

		// Should preserve existing content
		testutil.AssertEqual(t, "Task completed successfully!", message["content"])
	})

	t.Run("NonStreamingResponseNoExitTool", func(t *testing.T) {
		response := map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "Here's my response",
					},
				},
			},
		}

		responseBody, _ := json.Marshal(response)
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

		// Should pass through unchanged
		testutil.AssertEqual(t, "Here's my response", transformedResp["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"])
	})

	t.Run("StreamingResponse", func(t *testing.T) {
		streamData := `data: {"choices": [{"delta": {"tool_calls": [{"id": "call_exit", "type": "function", "function": {"name": "ExitTool"}}]}}]}

data: {"choices": [{"finish_reason": "tool_calls"}]}

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

		// Should have events but ExitTool should be filtered out
		testutil.AssertEqual(t, true, len(events) > 0)

		// Check that ExitTool was filtered and content added
		hasDefaultContent := false
		for _, eventData := range events {
			var event map[string]interface{}
			json.Unmarshal([]byte(eventData), &event)

			if choices, ok := event["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						if content, ok := delta["content"].(string); ok && content == "I have completed the requested task." {
							hasDefaultContent = true
						}
					}
				}
			}
		}

		testutil.AssertEqual(t, true, hasDefaultContent)
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

func TestToolUseTransformStreamEvent(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewToolUseTransformer()

	t.Run("FilterExitToolCall", func(t *testing.T) {
		state := &toolUseStreamState{}
		
		eventData := map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{
					"delta": map[string]interface{}{
						"tool_calls": []interface{}{
							map[string]interface{}{
								"id":   "call_regular",
								"type": "function",
								"function": map[string]interface{}{
									"name": "regular_tool",
								},
							},
							map[string]interface{}{
								"id":   "call_exit",
								"type": "function",
								"function": map[string]interface{}{
									"name": "ExitTool",
								},
							},
						},
					},
				},
			},
		}

		jsonData, _ := json.Marshal(eventData)
		event := &SSEEvent{Data: string(jsonData)}

		result := transformer.transformStreamEvent(event, state)
		testutil.AssertNotEqual(t, nil, result)

		var transformedData map[string]interface{}
		json.Unmarshal([]byte(result.Data), &transformedData)

		choices := transformedData["choices"].([]interface{})
		choice := choices[0].(map[string]interface{})
		delta := choice["delta"].(map[string]interface{})
		toolCalls := delta["tool_calls"].([]interface{})

		// Should have only one tool call (ExitTool filtered out)
		testutil.AssertEqual(t, 1, len(toolCalls))
		remainingCall := toolCalls[0].(map[string]interface{})
		function := remainingCall["function"].(map[string]interface{})
		testutil.AssertEqual(t, "regular_tool", function["name"])

		// State should track that ExitTool was found
		testutil.AssertEqual(t, true, state.hasExitTool)
	})

	t.Run("FilterOnlyExitTool", func(t *testing.T) {
		state := &toolUseStreamState{}
		
		eventData := map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{
					"delta": map[string]interface{}{
						"tool_calls": []interface{}{
							map[string]interface{}{
								"id":   "call_exit",
								"type": "function",
								"function": map[string]interface{}{
									"name": "ExitTool",
								},
							},
						},
					},
				},
			},
		}

		jsonData, _ := json.Marshal(eventData)
		event := &SSEEvent{Data: string(jsonData)}

		result := transformer.transformStreamEvent(event, state)
		testutil.AssertNotEqual(t, nil, result)

		var transformedData map[string]interface{}
		json.Unmarshal([]byte(result.Data), &transformedData)

		choices := transformedData["choices"].([]interface{})
		choice := choices[0].(map[string]interface{})
		delta := choice["delta"].(map[string]interface{})

		// Should have no tool calls
		_, hasToolCalls := delta["tool_calls"]
		testutil.AssertEqual(t, false, hasToolCalls)

		// Should have default content added
		testutil.AssertEqual(t, "I have completed the requested task.", delta["content"])
		testutil.AssertEqual(t, true, state.hasExitTool)
	})

	t.Run("FinishEventWithExitTool", func(t *testing.T) {
		state := &toolUseStreamState{
			hasExitTool: true,
		}
		
		eventData := map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{
					"finish_reason": "tool_calls",
					"delta":         map[string]interface{}{},
				},
			},
		}

		jsonData, _ := json.Marshal(eventData)
		event := &SSEEvent{Data: string(jsonData)}

		result := transformer.transformStreamEvent(event, state)
		testutil.AssertNotEqual(t, nil, result)

		var transformedData map[string]interface{}
		json.Unmarshal([]byte(result.Data), &transformedData)

		choices := transformedData["choices"].([]interface{})
		choice := choices[0].(map[string]interface{})
		delta := choice["delta"].(map[string]interface{})

		// Should add default content for finish event with ExitTool
		testutil.AssertEqual(t, "I have completed the requested task.", delta["content"])
	})

	t.Run("RegularEventPassThrough", func(t *testing.T) {
		state := &toolUseStreamState{}
		
		eventData := map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{
					"delta": map[string]interface{}{
						"content": "Regular response content",
					},
				},
			},
		}

		jsonData, _ := json.Marshal(eventData)
		event := &SSEEvent{Data: string(jsonData)}

		result := transformer.transformStreamEvent(event, state)
		testutil.AssertNotEqual(t, nil, result)

		var transformedData map[string]interface{}
		json.Unmarshal([]byte(result.Data), &transformedData)

		choices := transformedData["choices"].([]interface{})
		choice := choices[0].(map[string]interface{})
		delta := choice["delta"].(map[string]interface{})

		// Should pass through unchanged
		testutil.AssertEqual(t, "Regular response content", delta["content"])
		testutil.AssertEqual(t, false, state.hasExitTool)
	})

	t.Run("InvalidEventData", func(t *testing.T) {
		state := &toolUseStreamState{}
		event := &SSEEvent{Data: "invalid json"}

		result := transformer.transformStreamEvent(event, state)
		
		// Should return original event on parse error
		testutil.AssertEqual(t, event, result)
	})
}

func TestToolUseStreamState(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	t.Run("InitialState", func(t *testing.T) {
		state := &toolUseStreamState{
			hasExitTool:       false,
			exitToolID:        "",
			suppressToolCalls: false,
		}

		testutil.AssertEqual(t, false, state.hasExitTool)
		testutil.AssertEqual(t, "", state.exitToolID)
		testutil.AssertEqual(t, false, state.suppressToolCalls)
	})

	t.Run("StateTransitions", func(t *testing.T) {
		state := &toolUseStreamState{}

		// Mark ExitTool found
		state.hasExitTool = true
		state.exitToolID = "call_123"
		testutil.AssertEqual(t, true, state.hasExitTool)
		testutil.AssertEqual(t, "call_123", state.exitToolID)

		// Set suppression
		state.suppressToolCalls = true
		testutil.AssertEqual(t, true, state.suppressToolCalls)
	})
}

func TestToolUseIntegration(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewToolUseTransformer()
	ctx := context.Background()

	t.Run("CompleteToolModeWorkflow", func(t *testing.T) {
		// Test request transformation
		request := map[string]interface{}{
			"model": "claude-3-sonnet",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Help me complete a task",
				},
			},
			"tools": []interface{}{
				map[string]interface{}{
					"name":        "helper_tool",
					"description": "A helpful tool",
				},
			},
		}

		transformedReq, err := transformer.TransformRequestIn(ctx, request, "anthropic")
		testutil.AssertNoError(t, err)

		reqMap := transformedReq.(map[string]interface{})
		
		// Should have system reminder and ExitTool
		messages := reqMap["messages"].([]interface{})
		testutil.AssertEqual(t, 2, len(messages))
		
		tools := reqMap["tools"].([]interface{})
		testutil.AssertEqual(t, 2, len(tools)) // original + ExitTool
		
		testutil.AssertEqual(t, "required", reqMap["tool_choice"])

		// Test response transformation with ExitTool
		response := map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "Task completed!",
						"tool_calls": []interface{}{
							map[string]interface{}{
								"id":   "call_helper",
								"type": "function",
								"function": map[string]interface{}{
									"name": "helper_tool",
								},
							},
							map[string]interface{}{
								"id":   "call_exit",
								"type": "function",
								"function": map[string]interface{}{
									"name": "ExitTool",
								},
							},
						},
					},
				},
			},
		}

		responseBody, _ := json.Marshal(response)
		resp := &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(string(responseBody))),
		}

		transformedResp, err := transformer.TransformResponseOut(ctx, resp)
		testutil.AssertNoError(t, err)

		body, _ := io.ReadAll(transformedResp.Body)
		var finalResp map[string]interface{}
		json.Unmarshal(body, &finalResp)

		choices := finalResp["choices"].([]interface{})
		choice := choices[0].(map[string]interface{})
		message := choice["message"].(map[string]interface{})

		// Should preserve content and filter ExitTool
		testutil.AssertEqual(t, "Task completed!", message["content"])
		
		toolCalls := message["tool_calls"].([]interface{})
		testutil.AssertEqual(t, 1, len(toolCalls)) // Only helper_tool, ExitTool filtered

		remainingCall := toolCalls[0].(map[string]interface{})
		function := remainingCall["function"].(map[string]interface{})
		testutil.AssertEqual(t, "helper_tool", function["name"])
	})

	t.Run("StreamingToolModeWorkflow", func(t *testing.T) {
		// Test streaming response with mixed tool calls
		streamData := `data: {"choices": [{"delta": {"tool_calls": [{"id": "call_helper", "function": {"name": "helper_tool"}}, {"id": "call_exit", "function": {"name": "ExitTool"}}]}}]}

data: {"choices": [{"finish_reason": "tool_calls"}]}

data: [DONE]

`
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(streamData)),
		}

		result, err := transformer.TransformResponseOut(ctx, resp)
		testutil.AssertNoError(t, err)

		// Read and analyze events
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

		// Should have filtered ExitTool and potentially added content
		testutil.AssertEqual(t, true, len(events) > 0)

		// Verify ExitTool was filtered
		hasExitTool := false
		hasHelperTool := false
		
		for _, eventData := range events {
			var event map[string]interface{}
			json.Unmarshal([]byte(eventData), &event)

			if choices, ok := event["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						if toolCalls, ok := delta["tool_calls"].([]interface{}); ok {
							for _, tc := range toolCalls {
								if tcMap, ok := tc.(map[string]interface{}); ok {
									if function, ok := tcMap["function"].(map[string]interface{}); ok {
										name := function["name"].(string)
										if name == "ExitTool" {
											hasExitTool = true
										} else if name == "helper_tool" {
											hasHelperTool = true
										}
									}
								}
							}
						}
					}
				}
			}
		}

		testutil.AssertEqual(t, false, hasExitTool)   // Should be filtered out
		testutil.AssertEqual(t, true, hasHelperTool) // Should be preserved
	})
}