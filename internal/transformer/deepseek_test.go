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

func TestDeepSeekTransformer(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	t.Run("NewDeepSeekTransformer", func(t *testing.T) {
		transformer := NewDeepSeekTransformer()
		testutil.AssertEqual(t, "deepseek", transformer.GetName())
		testutil.AssertEqual(t, "/v1/chat/completions", transformer.GetEndpoint())
	})
}

func TestDeepSeekTransformRequestIn(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewDeepSeekTransformer()
	ctx := context.Background()

	t.Run("ValidRequest", func(t *testing.T) {
		request := map[string]interface{}{
			"model":      "deepseek-chat",
			"messages":   []interface{}{},
			"max_tokens": float64(4096),
		}

		result, err := transformer.TransformRequestIn(ctx, request, "deepseek")
		testutil.AssertNoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		testutil.AssertEqual(t, true, ok)
		testutil.AssertEqual(t, "deepseek-chat", resultMap["model"])
		testutil.AssertEqual(t, float64(4096), resultMap["max_tokens"])
	})

	t.Run("LimitMaxTokens", func(t *testing.T) {
		request := map[string]interface{}{
			"model":      "deepseek-chat",
			"messages":   []interface{}{},
			"max_tokens": float64(16384), // Exceeds DeepSeek limit
		}

		result, err := transformer.TransformRequestIn(ctx, request, "deepseek")
		testutil.AssertNoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		testutil.AssertEqual(t, true, ok)
		testutil.AssertEqual(t, 8192, resultMap["max_tokens"])
	})

	t.Run("MaxTokensWithinLimit", func(t *testing.T) {
		request := map[string]interface{}{
			"model":      "deepseek-chat",
			"messages":   []interface{}{},
			"max_tokens": float64(4096), // Within limit
		}

		result, err := transformer.TransformRequestIn(ctx, request, "deepseek")
		testutil.AssertNoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		testutil.AssertEqual(t, true, ok)
		testutil.AssertEqual(t, float64(4096), resultMap["max_tokens"])
	})

	t.Run("NoMaxTokens", func(t *testing.T) {
		request := map[string]interface{}{
			"model":    "deepseek-chat",
			"messages": []interface{}{},
		}

		result, err := transformer.TransformRequestIn(ctx, request, "deepseek")
		testutil.AssertNoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		testutil.AssertEqual(t, true, ok)
		testutil.AssertEqual(t, "deepseek-chat", resultMap["model"])
		// Should not add max_tokens if not present
		_, hasMaxTokens := resultMap["max_tokens"]
		testutil.AssertEqual(t, false, hasMaxTokens)
	})

	t.Run("InvalidRequestFormat", func(t *testing.T) {
		request := "invalid request"

		_, err := transformer.TransformRequestIn(ctx, request, "deepseek")
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "invalid request format")
	})
}

func TestDeepSeekTransformResponseOut(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewDeepSeekTransformer()
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
		streamingData := `data: {"choices": [{"delta": {"reasoning_content": "Let me think..."}}]}

data: {"choices": [{"delta": {"content": "Hello world"}}]}

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

		// Should have transformed reasoning_content to thinking
		testutil.AssertEqual(t, true, len(events) > 0)
		
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
	})
}

func TestDeepSeekTransformStreamData(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewDeepSeekTransformer()

	t.Run("ReasoningContentTransformation", func(t *testing.T) {
		state := &deepseekStreamState{
			reasoningContent:    "",
			isReasoningComplete: false,
			contentIndex:        0,
		}

		data := `{"choices": [{"delta": {"reasoning_content": "Let me think about this..."}}]}`
		
		results, err := transformer.transformStreamData(data, state)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 1, len(results))

		// Should have accumulated reasoning content
		testutil.AssertEqual(t, "Let me think about this...", state.reasoningContent)
		testutil.AssertEqual(t, false, state.isReasoningComplete)

		// Check transformed result
		var result map[string]interface{}
		json.Unmarshal([]byte(results[0]), &result)
		
		choices := result["choices"].([]interface{})
		choice := choices[0].(map[string]interface{})
		delta := choice["delta"].(map[string]interface{})
		
		thinking := delta["thinking"].(map[string]interface{})
		testutil.AssertEqual(t, "Let me think about this...", thinking["content"])
		
		// reasoning_content should be removed
		_, hasReasoning := delta["reasoning_content"]
		testutil.AssertEqual(t, false, hasReasoning)
	})

	t.Run("TransitionToRegularContent", func(t *testing.T) {
		state := &deepseekStreamState{
			reasoningContent:    "Previous thinking...",
			isReasoningComplete: false,
			contentIndex:        0,
		}

		data := `{"choices": [{"delta": {"content": "Hello world"}}]}`
		
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
		testutil.AssertEqual(t, "Previous thinking...", content["content"])

		// Second result should be regular content with updated index
		var contentResult map[string]interface{}
		json.Unmarshal([]byte(results[1]), &contentResult)
		
		choices2 := contentResult["choices"].([]interface{})
		choice2 := choices2[0].(map[string]interface{})
		testutil.AssertEqual(t, 1, int(choice2["index"].(float64)))
		
		delta2 := choice2["delta"].(map[string]interface{})
		testutil.AssertEqual(t, "Hello world", delta2["content"])
	})

	t.Run("FinishReasonWithRemainingReasoning", func(t *testing.T) {
		state := &deepseekStreamState{
			reasoningContent:    "Final thoughts...",
			isReasoningComplete: false,
			contentIndex:        0,
		}

		data := `{"choices": [{"delta": {}, "finish_reason": "stop"}]}`
		
		results, err := transformer.transformStreamData(data, state)
		testutil.AssertNoError(t, err)
		
		// Check we got some results
		testutil.AssertEqual(t, true, len(results) > 0) // At least some result
		
		// Only check state if we have results
		if len(results) > 0 {
			testutil.AssertEqual(t, true, state.isReasoningComplete)
		}

		// First result should be thinking block
		var thinkingResult map[string]interface{}
		json.Unmarshal([]byte(results[0]), &thinkingResult)
		
		choices := thinkingResult["choices"].([]interface{})
		choice := choices[0].(map[string]interface{})
		delta := choice["delta"].(map[string]interface{})
		content := delta["content"].(map[string]interface{})
		testutil.AssertEqual(t, "Final thoughts...", content["content"])

		// Check for finish chunk if there are multiple results
		if len(results) > 1 {
			var finishResult map[string]interface{}
			json.Unmarshal([]byte(results[1]), &finishResult)
			
			choices2 := finishResult["choices"].([]interface{})
			choice2 := choices2[0].(map[string]interface{})
			testutil.AssertEqual(t, "stop", choice2["finish_reason"])
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		state := &deepseekStreamState{}
		data := `invalid json`
		
		results, err := transformer.transformStreamData(data, state)
		testutil.AssertNoError(t, err)
		
		// Should pass through unchanged on parse error
		testutil.AssertEqual(t, 1, len(results))
		testutil.AssertEqual(t, data, results[0])
	})

	t.Run("NoChoices", func(t *testing.T) {
		state := &deepseekStreamState{}
		data := `{"model": "deepseek-chat"}`
		
		results, err := transformer.transformStreamData(data, state)
		testutil.AssertNoError(t, err)
		
		// Should pass through unchanged
		testutil.AssertEqual(t, 1, len(results))
		testutil.AssertEqual(t, data, results[0])
	})
}

func TestDeepSeekCreateThinkingBlockChunk(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewDeepSeekTransformer()

	t.Run("CreateValidThinkingBlock", func(t *testing.T) {
		baseChunk := map[string]interface{}{
			"id":      "chat-123",
			"object":  "chat.completion.chunk",
			"created": 1234567890,
			"model":   "deepseek-chat",
		}
		
		content := "This is my thinking process..."
		index := 1

		result := transformer.createThinkingBlockChunk(baseChunk, content, index)

		// Should preserve base fields
		testutil.AssertEqual(t, "chat-123", result["id"])
		testutil.AssertEqual(t, "chat.completion.chunk", result["object"])
		testutil.AssertEqual(t, 1234567890, result["created"])
		testutil.AssertEqual(t, "deepseek-chat", result["model"])

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

func TestDeepSeekStreamState(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	t.Run("InitialState", func(t *testing.T) {
		state := &deepseekStreamState{
			reasoningContent:    "",
			isReasoningComplete: false,
			contentIndex:        0,
		}

		testutil.AssertEqual(t, "", state.reasoningContent)
		testutil.AssertEqual(t, false, state.isReasoningComplete)
		testutil.AssertEqual(t, 0, state.contentIndex)
	})

	t.Run("StateTransitions", func(t *testing.T) {
		state := &deepseekStreamState{}

		// Add reasoning content
		state.reasoningContent = "Some thinking"
		testutil.AssertEqual(t, "Some thinking", state.reasoningContent)
		testutil.AssertEqual(t, false, state.isReasoningComplete)

		// Mark reasoning complete
		state.isReasoningComplete = true
		state.contentIndex = 1
		testutil.AssertEqual(t, true, state.isReasoningComplete)
		testutil.AssertEqual(t, 1, state.contentIndex)
	})
}

func TestDeepSeekIntegration(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewDeepSeekTransformer()
	ctx := context.Background()

	t.Run("CompleteWorkflow", func(t *testing.T) {
		// Test request transformation
		request := map[string]interface{}{
			"model":      "deepseek-chat",
			"messages":   []interface{}{},
			"max_tokens": float64(16384), // Will be limited
		}

		transformedReq, err := transformer.TransformRequestIn(ctx, request, "deepseek")
		testutil.AssertNoError(t, err)

		reqMap := transformedReq.(map[string]interface{})
		testutil.AssertEqual(t, 8192, reqMap["max_tokens"])

		// Test streaming response transformation
		streamData := `data: {"choices": [{"delta": {"reasoning_content": "Thinking..."}}]}

data: {"choices": [{"delta": {"content": "Response"}}]}

data: [DONE]

`
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(streamData)),
		}

		transformedResp, err := transformer.TransformResponseOut(ctx, resp)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 200, transformedResp.StatusCode)

		// Verify transformation worked
		reader := NewSSEReader(transformedResp.Body)
		hasThinkingEvent := false
		hasContentEvent := false

		for {
			event, err := reader.ReadEvent()
			if err == io.EOF {
				break
			}
			testutil.AssertNoError(t, err)

			if event.Data != "" && event.Data != "[DONE]" {
				var eventData map[string]interface{}
				json.Unmarshal([]byte(event.Data), &eventData)

				if choices, ok := eventData["choices"].([]interface{}); ok && len(choices) > 0 {
					if choice, ok := choices[0].(map[string]interface{}); ok {
						if delta, ok := choice["delta"].(map[string]interface{}); ok {
							if thinking, ok := delta["thinking"]; ok {
								hasThinkingEvent = true
								testutil.AssertNotEqual(t, nil, thinking)
							}
							if content, ok := delta["content"].(string); ok && content == "Response" {
								hasContentEvent = true
							}
						}
					}
				}
			}
		}

		testutil.AssertEqual(t, true, hasThinkingEvent)
		testutil.AssertEqual(t, true, hasContentEvent)
	})
}