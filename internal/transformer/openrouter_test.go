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

func TestOpenRouterTransformer(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	t.Run("NewOpenRouterTransformer", func(t *testing.T) {
		transformer := NewOpenRouterTransformer()
		testutil.AssertEqual(t, "openrouter", transformer.GetName())
		testutil.AssertEqual(t, "/api/v1/chat/completions", transformer.GetEndpoint())
	})
}

func TestOpenRouterTransformResponseOut(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewOpenRouterTransformer()
	ctx := context.Background()

	t.Run("NonStreamingResponse", func(t *testing.T) {
		body := `{"choices": [{"message": {"content": "Hello"}}]}`
		resp := &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
		}

		result, err := transformer.TransformResponseOut(ctx, resp)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 200, result.StatusCode)

		// Should pass through unchanged for non-streaming
		resultBody, _ := io.ReadAll(result.Body)
		testutil.AssertEqual(t, body, string(resultBody))
	})

	t.Run("StreamingResponseTransformation", func(t *testing.T) {
		// Create streaming response with reasoning content
		streamingData := `data: {"choices": [{"delta": {"reasoning_content": "Let me analyze this..."}}]}

data: {"choices": [{"delta": {"content": "The answer is 42"}}]}

data: {"choices": [{"delta": {}, "finish_reason": "stop"}]}

data: [DONE]

`
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(streamingData)),
		}

		result, err := transformer.TransformResponseOut(ctx, resp)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 200, result.StatusCode)
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

		// Should have multiple events
		testutil.AssertEqual(t, true, len(events) >= 3)

		// Check that reasoning_content was transformed to thinking
		var firstEvent map[string]interface{}
		json.Unmarshal([]byte(events[0]), &firstEvent)

		choices, ok := firstEvent["choices"].([]interface{})
		testutil.AssertEqual(t, true, ok)
		testutil.AssertEqual(t, true, len(choices) > 0)

		choice, ok := choices[0].(map[string]interface{})
		testutil.AssertEqual(t, true, ok)

		delta, ok := choice["delta"].(map[string]interface{})
		testutil.AssertEqual(t, true, ok)

		// Should have thinking instead of reasoning_content
		thinking, hasThinking := delta["thinking"]
		testutil.AssertEqual(t, true, hasThinking)
		testutil.AssertNotEqual(t, nil, thinking)

		// reasoning_content should be removed
		_, hasReasoning := delta["reasoning_content"]
		testutil.AssertEqual(t, false, hasReasoning)
	})

	t.Run("StreamingWithToolCalls", func(t *testing.T) {
		streamingData := `data: {"choices": [{"delta": {"reasoning_content": "I need to use a tool..."}}]}

data: {"choices": [{"delta": {"tool_calls": [{"index": 0, "id": "call_123", "type": "function", "function": {"name": "get_weather", "arguments": "{\"location\": \"NYC\"}"}}]}}]}

data: {"choices": [{"finish_reason": "tool_calls"}]}

data: [DONE]

`
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(streamingData)),
		}

		result, err := transformer.TransformResponseOut(ctx, resp)
		testutil.AssertNoError(t, err)

		// Read and verify events
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

		// Should have reasoning → thinking block → tool call → finish
		testutil.AssertEqual(t, true, len(events) >= 3)

		// Verify tool call has correct index after reasoning transformation
		hasToolCall := false
		for _, eventData := range events {
			var event map[string]interface{}
			json.Unmarshal([]byte(eventData), &event)

			if choices, ok := event["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						if toolCalls, ok := delta["tool_calls"].([]interface{}); ok {
							hasToolCall = true
							toolCall := toolCalls[0].(map[string]interface{})
							// Index should be adjusted for reasoning content
							index := toolCall["index"].(float64)
							testutil.AssertEqual(t, true, index >= 0)
						}
					}
				}
			}
		}
		testutil.AssertEqual(t, true, hasToolCall)
	})
}

func TestOpenRouterTransformStreamData(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewOpenRouterTransformer()

	t.Run("ReasoningContentTransformation", func(t *testing.T) {
		state := &openrouterStreamState{
			reasoningContent:    "",
			isReasoningComplete: false,
			contentIndex:        0,
			hasToolCalls:        false,
		}

		data := `{"choices": [{"delta": {"reasoning_content": "I'm thinking about this problem..."}}]}`

		results, err := transformer.transformStreamData(data, state)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 1, len(results))

		// Should have accumulated reasoning content
		testutil.AssertEqual(t, "I'm thinking about this problem...", state.reasoningContent)
		testutil.AssertEqual(t, false, state.isReasoningComplete)

		// Check transformed result
		var result map[string]interface{}
		json.Unmarshal([]byte(results[0]), &result)

		choices := result["choices"].([]interface{})
		choice := choices[0].(map[string]interface{})
		delta := choice["delta"].(map[string]interface{})

		thinking := delta["thinking"].(map[string]interface{})
		testutil.AssertEqual(t, "I'm thinking about this problem...", thinking["content"])

		// reasoning_content should be removed
		_, hasReasoning := delta["reasoning_content"]
		testutil.AssertEqual(t, false, hasReasoning)
	})

	t.Run("TransitionToRegularContent", func(t *testing.T) {
		state := &openrouterStreamState{
			reasoningContent:    "Previous reasoning...",
			isReasoningComplete: false,
			contentIndex:        0,
			hasToolCalls:        false,
		}

		data := `{"choices": [{"delta": {"content": "Here's my answer"}}]}`

		results, err := transformer.transformStreamData(data, state)
		testutil.AssertNoError(t, err)

		// Should have 2 results: thinking block + regular content
		testutil.AssertEqual(t, 2, len(results))
		testutil.AssertEqual(t, true, state.isReasoningComplete)
		testutil.AssertEqual(t, 1, state.contentIndex)

		// First result should be thinking block
		var thinkingResult map[string]interface{}
		json.Unmarshal([]byte(results[0]), &thinkingResult)

		choices := thinkingResult["choices"].([]interface{})
		choice := choices[0].(map[string]interface{})
		testutil.AssertEqual(t, 0, int(choice["index"].(float64)))

		delta := choice["delta"].(map[string]interface{})
		content := delta["content"].(map[string]interface{})
		testutil.AssertEqual(t, "Previous reasoning...", content["content"])

		// Second result should be regular content with updated index
		var contentResult map[string]interface{}
		json.Unmarshal([]byte(results[1]), &contentResult)

		choices2 := contentResult["choices"].([]interface{})
		choice2 := choices2[0].(map[string]interface{})
		testutil.AssertEqual(t, 1, int(choice2["index"].(float64)))

		delta2 := choice2["delta"].(map[string]interface{})
		testutil.AssertEqual(t, "Here's my answer", delta2["content"])
	})

	t.Run("ToolCallsHandling", func(t *testing.T) {
		state := &openrouterStreamState{
			reasoningContent:    "Need to call a tool...",
			isReasoningComplete: false,
			contentIndex:        0,
			hasToolCalls:        false,
		}

		data := `{"choices": [{"delta": {"tool_calls": [{"index": 0, "id": "call_123", "type": "function", "function": {"name": "test_tool"}}]}}]}`

		results, err := transformer.transformStreamData(data, state)
		testutil.AssertNoError(t, err)

		// Should have 2 results: thinking block + tool calls
		testutil.AssertEqual(t, 2, len(results))
		testutil.AssertEqual(t, true, state.isReasoningComplete)
		testutil.AssertEqual(t, true, state.hasToolCalls)

		// Check tool call index adjustment
		var toolResult map[string]interface{}
		json.Unmarshal([]byte(results[1]), &toolResult)

		choices := toolResult["choices"].([]interface{})
		choice := choices[0].(map[string]interface{})
		delta := choice["delta"].(map[string]interface{})
		toolCalls := delta["tool_calls"].([]interface{})
		toolCall := toolCalls[0].(map[string]interface{})

		// Index should be adjusted by contentIndex
		testutil.AssertEqual(t, float64(1), toolCall["index"])
	})

	t.Run("FinishReasonWithPendingReasoning", func(t *testing.T) {
		state := &openrouterStreamState{
			reasoningContent:    "Final thoughts...",
			isReasoningComplete: false,
			contentIndex:        0,
			hasToolCalls:        false,
		}

		data := `{"choices": [{"delta": {}, "finish_reason": "stop"}]}`

		results, err := transformer.transformStreamData(data, state)
		testutil.AssertNoError(t, err)

		// Should have 2 results: thinking block + finish chunk
		testutil.AssertEqual(t, 2, len(results))
		testutil.AssertEqual(t, true, state.isReasoningComplete)

		// First result should be thinking block
		var thinkingResult map[string]interface{}
		json.Unmarshal([]byte(results[0]), &thinkingResult)

		choices := thinkingResult["choices"].([]interface{})
		choice := choices[0].(map[string]interface{})
		delta := choice["delta"].(map[string]interface{})
		content := delta["content"].(map[string]interface{})
		testutil.AssertEqual(t, "Final thoughts...", content["content"])

		// Second result should be finish chunk with adjusted index
		var finishResult map[string]interface{}
		json.Unmarshal([]byte(results[1]), &finishResult)

		choices2 := finishResult["choices"].([]interface{})
		choice2 := choices2[0].(map[string]interface{})
		testutil.AssertEqual(t, 1, int(choice2["index"].(float64)))
		testutil.AssertEqual(t, "stop", choice2["finish_reason"])
	})

	t.Run("EmptyDelta", func(t *testing.T) {
		state := &openrouterStreamState{}
		data := `{"choices": [{"finish_reason": "stop"}]}`

		results, err := transformer.transformStreamData(data, state)
		testutil.AssertNoError(t, err)

		// Should handle chunks without delta
		testutil.AssertEqual(t, 1, len(results))

		var result map[string]interface{}
		json.Unmarshal([]byte(results[0]), &result)

		choices := result["choices"].([]interface{})
		choice := choices[0].(map[string]interface{})
		testutil.AssertEqual(t, "stop", choice["finish_reason"])
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		state := &openrouterStreamState{}
		data := `invalid json`

		results, err := transformer.transformStreamData(data, state)
		testutil.AssertNoError(t, err)

		// Should pass through unchanged on parse error
		testutil.AssertEqual(t, 1, len(results))
		testutil.AssertEqual(t, data, results[0])
	})

	t.Run("NoChoices", func(t *testing.T) {
		state := &openrouterStreamState{}
		data := `{"model": "some-model"}`

		results, err := transformer.transformStreamData(data, state)
		testutil.AssertNoError(t, err)

		// Should pass through unchanged
		testutil.AssertEqual(t, 1, len(results))
		testutil.AssertEqual(t, data, results[0])
	})
}

func TestOpenRouterCreateThinkingBlockChunk(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewOpenRouterTransformer()

	t.Run("CreateValidThinkingBlock", func(t *testing.T) {
		baseChunk := map[string]interface{}{
			"id":      "chatcmpl-123",
			"object":  "chat.completion.chunk",
			"created": 1234567890,
			"model":   "openrouter-model",
		}

		content := "Here's my reasoning process..."
		index := 2

		result := transformer.createThinkingBlockChunk(baseChunk, content, index)

		// Should preserve base fields
		testutil.AssertEqual(t, "chatcmpl-123", result["id"])
		testutil.AssertEqual(t, "chat.completion.chunk", result["object"])
		testutil.AssertEqual(t, 1234567890, result["created"])
		testutil.AssertEqual(t, "openrouter-model", result["model"])

		// Should have correct choices structure
		choices := result["choices"].([]interface{})
		testutil.AssertEqual(t, 1, len(choices))

		choice := choices[0].(map[string]interface{})
		testutil.AssertEqual(t, index, choice["index"])

		delta := choice["delta"].(map[string]interface{})
		contentMap := delta["content"].(map[string]interface{})
		testutil.AssertEqual(t, content, contentMap["content"])

		// Should have signature
		signature := contentMap["signature"].(string)
		testutil.AssertEqual(t, true, len(signature) > 0)
	})
}

func TestOpenRouterStreamState(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	t.Run("InitialState", func(t *testing.T) {
		state := &openrouterStreamState{
			reasoningContent:    "",
			isReasoningComplete: false,
			contentIndex:        0,
			hasToolCalls:        false,
		}

		testutil.AssertEqual(t, "", state.reasoningContent)
		testutil.AssertEqual(t, false, state.isReasoningComplete)
		testutil.AssertEqual(t, 0, state.contentIndex)
		testutil.AssertEqual(t, false, state.hasToolCalls)
	})

	t.Run("StateTransitions", func(t *testing.T) {
		state := &openrouterStreamState{}

		// Add reasoning content
		state.reasoningContent = "Analyzing the problem"
		testutil.AssertEqual(t, "Analyzing the problem", state.reasoningContent)
		testutil.AssertEqual(t, false, state.isReasoningComplete)

		// Mark reasoning complete and update index
		state.isReasoningComplete = true
		state.contentIndex = 1
		testutil.AssertEqual(t, true, state.isReasoningComplete)
		testutil.AssertEqual(t, 1, state.contentIndex)

		// Mark as having tool calls
		state.hasToolCalls = true
		testutil.AssertEqual(t, true, state.hasToolCalls)
	})
}

func TestOpenRouterIntegration(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewOpenRouterTransformer()
	ctx := context.Background()

	t.Run("ComplexStreamingScenario", func(t *testing.T) {
		// Complex streaming scenario with reasoning → tool call → content
		streamData := `data: {"choices": [{"delta": {"reasoning_content": "I need to think about this step by step..."}}]}

data: {"choices": [{"delta": {"reasoning_content": " Let me analyze the request."}}]}

data: {"choices": [{"delta": {"tool_calls": [{"index": 0, "id": "call_123", "type": "function", "function": {"name": "analyze", "arguments": "{\"query\": \"test\"}"}}]}}]}

data: {"choices": [{"delta": {"content": "Based on the analysis, here's my response."}}]}

data: {"choices": [{"delta": {}, "finish_reason": "stop"}]}

data: [DONE]

`
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(streamData)),
		}

		result, err := transformer.TransformResponseOut(ctx, resp)
		testutil.AssertNoError(t, err)

		// Read and analyze all events
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

		// Should have multiple events including thinking block, tool calls, content, and finish
		testutil.AssertEqual(t, true, len(events) >= 4)

		// Analyze the sequence of events
		hasThinkingBlock := false
		hasToolCall := false
		hasContent := false
		hasFinish := false
		thinkingBlockIndex := -1
		toolCallIndex := -1

		for i, eventData := range events {
			var event map[string]interface{}
			json.Unmarshal([]byte(eventData), &event)

			if choices, ok := event["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					// Check for thinking in delta
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						if thinking, ok := delta["thinking"]; ok && thinking != nil {
							hasThinkingBlock = true
							if thinkingBlockIndex == -1 {
								thinkingBlockIndex = i
							}
						}
						if toolCalls, ok := delta["tool_calls"]; ok && toolCalls != nil {
							hasToolCall = true
							toolCallIndex = i
						}
						if content, ok := delta["content"].(string); ok && content != "" && !strings.Contains(content, "content") {
							hasContent = true
						}
					}

					// Check for finish reason
					if finishReason, ok := choice["finish_reason"]; ok && finishReason != nil {
						hasFinish = true
					}
				}
			}
		}

		testutil.AssertEqual(t, true, hasThinkingBlock)
		testutil.AssertEqual(t, true, hasToolCall)
		testutil.AssertEqual(t, true, hasContent)
		testutil.AssertEqual(t, true, hasFinish)

		// Thinking block should appear before tool call
		if thinkingBlockIndex != -1 && toolCallIndex != -1 {
			testutil.AssertEqual(t, true, thinkingBlockIndex < toolCallIndex)
		}
	})

	t.Run("OnlyReasoningWithFinish", func(t *testing.T) {
		streamData := `data: {"choices": [{"delta": {"reasoning_content": "Just reasoning, no other content."}}]}

data: {"choices": [{"delta": {}, "finish_reason": "stop"}]}

data: [DONE]

`
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(streamData)),
		}

		result, err := transformer.TransformResponseOut(ctx, resp)
		testutil.AssertNoError(t, err)

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

		// Should have reasoning transformation + thinking block + finish
		testutil.AssertEqual(t, true, len(events) >= 2)

		// Last event should be the finish with proper index
		var lastEvent map[string]interface{}
		json.Unmarshal([]byte(events[len(events)-1]), &lastEvent)

		choices := lastEvent["choices"].([]interface{})
		choice := choices[0].(map[string]interface{})
		testutil.AssertEqual(t, "stop", choice["finish_reason"])
		// Index should be incremented due to thinking block
		testutil.AssertEqual(t, 1, int(choice["index"].(float64)))
	})
}
