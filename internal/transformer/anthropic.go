package transformer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/orchestre-dev/ccproxy/internal/utils"
)

// AnthropicTransformer converts between OpenAI and Anthropic message formats
type AnthropicTransformer struct {
	BaseTransformer
}

// NewAnthropicTransformer creates a new Anthropic transformer
func NewAnthropicTransformer() *AnthropicTransformer {
	return &AnthropicTransformer{
		BaseTransformer: *NewBaseTransformer("anthropic", "/v1/messages"),
	}
}

// TransformRequestIn transforms OpenAI format to Anthropic format
func (t *AnthropicTransformer) TransformRequestIn(ctx context.Context, request interface{}, provider string) (interface{}, error) {
	// Parse the incoming request
	reqMap, ok := request.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid request format")
	}

	// Create transformed request
	transformed := make(map[string]interface{})

	// Copy basic fields
	if model, ok := reqMap["model"].(string); ok {
		// Strip provider prefix if present (e.g., "anthropic,claude-3" -> "claude-3")
		if strings.Contains(model, ",") {
			parts := strings.SplitN(model, ",", 2)
			if len(parts) == 2 {
				transformed["model"] = parts[1]
			} else {
				transformed["model"] = model
			}
		} else {
			transformed["model"] = model
		}
	}
	if maxTokens, ok := reqMap["max_tokens"]; ok {
		transformed["max_tokens"] = maxTokens
	}
	if temperature, ok := reqMap["temperature"]; ok {
		transformed["temperature"] = temperature
	}
	if stream, ok := reqMap["stream"]; ok {
		transformed["stream"] = stream
	}

	// Transform messages
	messages, ok := reqMap["messages"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("missing or invalid messages field")
	}

	transformedMessages := []interface{}{}
	var systemMessage string

	for _, msg := range messages {
		msgMap, ok := msg.(map[string]interface{})
		if !ok {
			continue
		}

		role, _ := msgMap["role"].(string)
		content := msgMap["content"]

		// Extract system message
		if role == "system" {
			if str, ok := content.(string); ok {
				systemMessage = str
			}
			continue
		}

		// Transform tool results
		if role == "tool" {
			// Convert OpenAI tool message to Anthropic tool_result
			transformedMsg := map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{
						"type":        "tool_result",
						"tool_use_id": msgMap["tool_call_id"],
						"content":     content,
					},
				},
			}
			transformedMessages = append(transformedMessages, transformedMsg)
			continue
		}

		// Handle assistant messages with tool calls
		if role == "assistant" && msgMap["tool_calls"] != nil {
			toolCalls, ok := msgMap["tool_calls"].([]interface{})
			if ok {
				contentBlocks := []interface{}{}
				
				// Add text content if present
				if content != nil && content != "" {
					contentBlocks = append(contentBlocks, map[string]interface{}{
						"type": "text",
						"text": content,
					})
				}

				// Convert tool calls
				for _, tc := range toolCalls {
					tcMap, ok := tc.(map[string]interface{})
					if !ok {
						continue
					}

					funcMap, ok := tcMap["function"].(map[string]interface{})
					if !ok {
						continue
					}

					// Parse arguments
					argsStr, _ := funcMap["arguments"].(string)
					var args interface{}
					if argsStr != "" {
						json.Unmarshal([]byte(argsStr), &args)
					}

					toolUse := map[string]interface{}{
						"type":  "tool_use",
						"id":    tcMap["id"],
						"name":  funcMap["name"],
						"input": args,
					}
					contentBlocks = append(contentBlocks, toolUse)
				}

				transformedMsg := map[string]interface{}{
					"role":    "assistant",
					"content": contentBlocks,
				}
				transformedMessages = append(transformedMessages, transformedMsg)
				continue
			}
		}

		// Regular message
		transformedMessages = append(transformedMessages, msgMap)
	}

	// Set system message if found
	if systemMessage != "" {
		transformed["system"] = systemMessage
	}

	transformed["messages"] = transformedMessages

	// Transform tools
	if tools, ok := reqMap["tools"].([]interface{}); ok {
		transformedTools := []interface{}{}
		for _, tool := range tools {
			toolMap, ok := tool.(map[string]interface{})
			if !ok {
				continue
			}

			funcMap, ok := toolMap["function"].(map[string]interface{})
			if !ok {
				continue
			}

			anthropicTool := map[string]interface{}{
				"name":         funcMap["name"],
				"description":  funcMap["description"],
				"input_schema": funcMap["parameters"],
			}
			transformedTools = append(transformedTools, anthropicTool)
		}
		transformed["tools"] = transformedTools

		// Transform tool_choice
		if toolChoice := reqMap["tool_choice"]; toolChoice != nil {
			switch tc := toolChoice.(type) {
			case string:
				if tc == "auto" {
					transformed["tool_choice"] = map[string]interface{}{"type": "any"}
				} else if tc == "required" {
					transformed["tool_choice"] = map[string]interface{}{"type": "any"}
				} else if tc == "none" {
					transformed["tool_choice"] = map[string]interface{}{"type": "none"}
				}
			case map[string]interface{}:
				if funcName, ok := tc["function"].(map[string]interface{})["name"].(string); ok {
					transformed["tool_choice"] = map[string]interface{}{
						"type": "tool",
						"name": funcName,
					}
				}
			}
		}
	}

	// Handle thinking parameter
	if thinking, ok := reqMap["thinking"]; ok && thinking == true {
		transformed["thinking"] = map[string]interface{}{
			"type":         "enabled",
			"budget_tokens": 16000,
		}
	}

	return transformed, nil
}

// TransformResponseOut transforms Anthropic streaming response to OpenAI format
func (t *AnthropicTransformer) TransformResponseOut(ctx context.Context, response *http.Response) (*http.Response, error) {
	// Check if it's a streaming response
	if !strings.Contains(response.Header.Get("Content-Type"), "text/event-stream") {
		// For non-streaming, parse and transform the JSON response
		return t.transformNonStreamingResponse(ctx, response)
	}

	// For streaming responses, we need to transform the SSE events
	reader := NewSSEReader(response.Body)
	pr, pw := io.Pipe()

	// Create a new response with the pipe reader
	newResp := &http.Response{
		Status:        response.Status,
		StatusCode:    response.StatusCode,
		Proto:         response.Proto,
		ProtoMajor:    response.ProtoMajor,
		ProtoMinor:    response.ProtoMinor,
		Header:        response.Header.Clone(),
		Body:          pr,
		ContentLength: -1,
		Request:       response.Request,
	}

	// Start the transformation in a goroutine
	go func() {
		defer pw.Close()
		writer := NewSSEWriter(pw)
		
		// Track state for streaming transformation
		state := &anthropicStreamState{
			contentBlocks: make(map[int]interface{}),
			messageID:     "",
			model:         "",
		}

		for {
			event, err := reader.ReadEvent()
			if err != nil {
				if err != io.EOF {
					utils.GetLogger().Errorf("Error reading SSE event: %v", err)
				}
				break
			}

			// Transform the event
			transformed, err := t.transformStreamEvent(event, state)
			if err != nil {
				utils.GetLogger().Errorf("Error transforming event: %v", err)
				continue
			}

			// Write transformed events
			for _, evt := range transformed {
				if err := writer.WriteEvent(evt); err != nil {
					utils.GetLogger().Errorf("Error writing transformed event: %v", err)
					break
				}
			}
		}
	}()

	return newResp, nil
}

// transformNonStreamingResponse transforms non-streaming Anthropic response
func (t *AnthropicTransformer) transformNonStreamingResponse(ctx context.Context, response *http.Response) (*http.Response, error) {
	// Read the response body
	body, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return nil, err
	}

	// Parse the response
	var anthropicResp map[string]interface{}
	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		// Return original response if we can't parse it
		response.Body = io.NopCloser(bytes.NewReader(body))
		return response, nil
	}

	// Transform to OpenAI format
	openaiResp := t.transformAnthropicToOpenAI(anthropicResp)

	// Marshal the transformed response
	transformedBody, err := json.Marshal(openaiResp)
	if err != nil {
		return nil, err
	}

	// Create new response with transformed body
	response.Body = io.NopCloser(bytes.NewReader(transformedBody))
	response.ContentLength = int64(len(transformedBody))
	return response, nil
}

// transformAnthropicToOpenAI transforms Anthropic response format to OpenAI format
func (t *AnthropicTransformer) transformAnthropicToOpenAI(anthropicResp map[string]interface{}) map[string]interface{} {
	openaiResp := map[string]interface{}{
		"id":      anthropicResp["id"],
		"object":  "chat.completion",
		"created": utils.GetTimestamp(),
		"model":   anthropicResp["model"],
	}

	// Transform content blocks to choices
	choices := []interface{}{}
	if content, ok := anthropicResp["content"].([]interface{}); ok {
		message := map[string]interface{}{
			"role": "assistant",
		}

		var textContent strings.Builder
		var toolCalls []interface{}

		for _, block := range content {
			blockMap, ok := block.(map[string]interface{})
			if !ok {
				continue
			}

			blockType, _ := blockMap["type"].(string)
			switch blockType {
			case "text":
				if text, ok := blockMap["text"].(string); ok {
					textContent.WriteString(text)
				}
			case "tool_use":
				toolCall := map[string]interface{}{
					"id":   blockMap["id"],
					"type": "function",
					"function": map[string]interface{}{
						"name":      blockMap["name"],
						"arguments": utils.ToJSONString(blockMap["input"]),
					},
				}
				toolCalls = append(toolCalls, toolCall)
			case "thinking":
				// Handle thinking blocks if needed
				if thinking, ok := blockMap["thinking"].(string); ok {
					// Add thinking as a special field or ignore based on requirements
					_ = thinking
				}
			}
		}

		// Set message content
		if textContent.Len() > 0 {
			message["content"] = textContent.String()
		}
		if len(toolCalls) > 0 {
			message["tool_calls"] = toolCalls
		}

		choice := map[string]interface{}{
			"index":   0,
			"message": message,
			"finish_reason": t.convertStopReason(anthropicResp["stop_reason"]),
		}
		choices = append(choices, choice)
	}

	openaiResp["choices"] = choices

	// Transform usage
	if usage, ok := anthropicResp["usage"].(map[string]interface{}); ok {
		openaiResp["usage"] = map[string]interface{}{
			"prompt_tokens":     usage["input_tokens"],
			"completion_tokens": usage["output_tokens"],
			"total_tokens":      t.sumTokens(usage["input_tokens"], usage["output_tokens"]),
		}
	}

	return openaiResp
}

// convertStopReason converts Anthropic stop reason to OpenAI finish reason
func (t *AnthropicTransformer) convertStopReason(stopReason interface{}) string {
	reason, _ := stopReason.(string)
	switch reason {
	case "end_turn":
		return "stop"
	case "max_tokens":
		return "length"
	case "tool_use":
		return "tool_calls"
	default:
		return "stop"
	}
}

// sumTokens safely adds token counts
func (t *AnthropicTransformer) sumTokens(input, output interface{}) int {
	inputTokens, _ := input.(float64)
	outputTokens, _ := output.(float64)
	return int(inputTokens + outputTokens)
}

// anthropicStreamState tracks state during streaming transformation
type anthropicStreamState struct {
	contentBlocks map[int]interface{}
	messageID     string
	model         string
	currentIndex  int
	usage         map[string]interface{}
}

// transformStreamEvent transforms a single Anthropic SSE event to OpenAI format
func (t *AnthropicTransformer) transformStreamEvent(event *SSEEvent, state *anthropicStreamState) ([]*SSEEvent, error) {
	// Parse the event data
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(event.Data), &data); err != nil {
		return nil, err
	}

	eventType, _ := data["type"].(string)
	transformed := []*SSEEvent{}

	switch eventType {
	case "message_start":
		// Extract message info
		if msg, ok := data["message"].(map[string]interface{}); ok {
			state.messageID, _ = msg["id"].(string)
			state.model, _ = msg["model"].(string)
		}

	case "content_block_start":
		index := 0
		if idx, ok := data["index"].(float64); ok {
			index = int(idx)
		}
		state.currentIndex = index
		
		if block, ok := data["content_block"].(map[string]interface{}); ok {
			blockType, _ := block["type"].(string)
			
			if blockType == "tool_use" {
				// Start a tool call
				chunk := t.createStreamChunk(state, map[string]interface{}{
					"tool_calls": []interface{}{
						map[string]interface{}{
							"index": index,
							"id":    block["id"],
							"type":  "function",
							"function": map[string]interface{}{
								"name":      block["name"],
								"arguments": "",
							},
						},
					},
				})
				transformed = append(transformed, t.createSSEEvent(chunk))
			}
		}

	case "content_block_delta":
		index := 0
		if idx, ok := data["index"].(float64); ok {
			index = int(idx)
		}
		
		if delta, ok := data["delta"].(map[string]interface{}); ok {
			deltaType, _ := delta["type"].(string)
			
			switch deltaType {
			case "text_delta":
				// Stream text content
				chunk := t.createStreamChunk(state, map[string]interface{}{
					"content": delta["text"],
				})
				transformed = append(transformed, t.createSSEEvent(chunk))
				
			case "input_json_delta":
				// Stream tool arguments
				chunk := t.createStreamChunk(state, map[string]interface{}{
					"tool_calls": []interface{}{
						map[string]interface{}{
							"index": index,
							"function": map[string]interface{}{
								"arguments": delta["partial_json"],
							},
						},
					},
				})
				transformed = append(transformed, t.createSSEEvent(chunk))
			}
		}

	case "message_delta":
		// Handle stop reason and usage
		if delta, ok := data["delta"].(map[string]interface{}); ok {
			if stopReason := delta["stop_reason"]; stopReason != nil {
				chunk := t.createStreamChunk(state, nil)
				chunk["choices"].([]interface{})[0].(map[string]interface{})["finish_reason"] = t.convertStopReason(stopReason)
				transformed = append(transformed, t.createSSEEvent(chunk))
			}
		}
		
		if usage, ok := data["usage"].(map[string]interface{}); ok {
			state.usage = usage
		}

	case "message_stop":
		// Send final chunk with usage
		if state.usage != nil {
			chunk := t.createStreamChunk(state, nil)
			chunk["usage"] = map[string]interface{}{
				"prompt_tokens":     state.usage["input_tokens"],
				"completion_tokens": state.usage["output_tokens"],
				"total_tokens":      t.sumTokens(state.usage["input_tokens"], state.usage["output_tokens"]),
			}
			transformed = append(transformed, t.createSSEEvent(chunk))
		}
		
		// Send [DONE] event
		transformed = append(transformed, &SSEEvent{Data: "[DONE]"})
	}

	return transformed, nil
}

// createStreamChunk creates an OpenAI stream chunk
func (t *AnthropicTransformer) createStreamChunk(state *anthropicStreamState, delta interface{}) map[string]interface{} {
	chunk := map[string]interface{}{
		"id":      state.messageID,
		"object":  "chat.completion.chunk",
		"created": utils.GetTimestamp(),
		"model":   state.model,
		"choices": []interface{}{
			map[string]interface{}{
				"index": 0,
				"delta": delta,
			},
		},
	}
	return chunk
}

// createSSEEvent creates an SSE event with JSON data
func (t *AnthropicTransformer) createSSEEvent(data interface{}) *SSEEvent {
	jsonData, _ := json.Marshal(data)
	return &SSEEvent{
		Data: string(jsonData),
	}
}