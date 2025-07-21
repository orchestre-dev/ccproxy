package transformer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/orchestre-dev/ccproxy/internal/utils"
)

// ToolUseTransformer handles tool mode transformations
type ToolUseTransformer struct {
	BaseTransformer
}

// NewToolUseTransformer creates a new ToolUse transformer
func NewToolUseTransformer() *ToolUseTransformer {
	return &ToolUseTransformer{
		BaseTransformer: *NewBaseTransformer("tooluse", ""),
	}
}

// TransformRequestIn adds ExitTool and enforces tool usage
func (t *ToolUseTransformer) TransformRequestIn(ctx context.Context, request interface{}, provider string) (interface{}, error) {
	// Parse the incoming request
	reqMap, ok := request.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid request format")
	}

	// Add system reminder about tool mode
	messages, ok := reqMap["messages"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("missing or invalid messages field")
	}

	// Create system reminder
	systemReminder := map[string]interface{}{
		"role": "system",
		"content": "You are in tool mode. When you have completed the user's request, " +
			"you MUST call the ExitTool function to exit tool mode and return to normal conversation.",
	}

	// Add system reminder at the beginning
	newMessages := []interface{}{systemReminder}
	newMessages = append(newMessages, messages...)
	reqMap["messages"] = newMessages

	// Add ExitTool to tools
	tools, _ := reqMap["tools"].([]interface{})
	if tools == nil {
		tools = []interface{}{}
	}

	// Create ExitTool
	exitTool := map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "ExitTool",
			"description": "Exit tool mode and return to normal conversation. Call this when you have completed the user's request.",
			"parameters": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []interface{}{},
			},
		},
	}

	// Add ExitTool to tools
	tools = append(tools, exitTool)
	reqMap["tools"] = tools

	// Set tool_choice to required
	reqMap["tool_choice"] = "required"

	return reqMap, nil
}

// TransformResponseOut intercepts ExitTool calls and converts to content
func (t *ToolUseTransformer) TransformResponseOut(ctx context.Context, response *http.Response) (*http.Response, error) {
	// Check if it's a streaming response
	if strings.Contains(response.Header.Get("Content-Type"), "text/event-stream") {
		return t.transformStreamingResponse(ctx, response)
	}

	// Handle non-streaming response
	return t.transformNonStreamingResponse(ctx, response)
}

// transformNonStreamingResponse handles non-streaming responses
func (t *ToolUseTransformer) transformNonStreamingResponse(ctx context.Context, response *http.Response) (*http.Response, error) {
	// Read the response body
	body, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return nil, err
	}

	// Parse the response
	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		// Return original response if we can't parse it
		response.Body = io.NopCloser(strings.NewReader(string(body)))
		return response, nil
	}

	// Check for ExitTool in choices
	if choices, ok := resp["choices"].([]interface{}); ok {
		for _, choice := range choices {
			choiceMap, ok := choice.(map[string]interface{})
			if !ok {
				continue
			}

			message, ok := choiceMap["message"].(map[string]interface{})
			if !ok {
				continue
			}

			// Check for tool calls
			if toolCalls, ok := message["tool_calls"].([]interface{}); ok {
				hasExitTool := false
				nonExitTools := []interface{}{}

				for _, tc := range toolCalls {
					tcMap, ok := tc.(map[string]interface{})
					if !ok {
						nonExitTools = append(nonExitTools, tc)
						continue
					}

					function, ok := tcMap["function"].(map[string]interface{})
					if !ok {
						nonExitTools = append(nonExitTools, tc)
						continue
					}

					// Check if it's ExitTool
					if function["name"] == "ExitTool" {
						hasExitTool = true
					} else {
						nonExitTools = append(nonExitTools, tc)
					}
				}

				// If ExitTool was called, convert to content
				if hasExitTool {
					// Remove tool_calls if no other tools
					if len(nonExitTools) == 0 {
						delete(message, "tool_calls")
					} else {
						message["tool_calls"] = nonExitTools
					}

					// Add default content if none exists
					if message["content"] == nil || message["content"] == "" {
						message["content"] = "I have completed the requested task."
					}
				}
			}
		}
	}

	// Marshal the transformed response
	transformedBody, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}

	// Create new response
	response.Body = io.NopCloser(strings.NewReader(string(transformedBody)))
	response.ContentLength = int64(len(transformedBody))
	return response, nil
}

// transformStreamingResponse handles streaming responses
func (t *ToolUseTransformer) transformStreamingResponse(ctx context.Context, response *http.Response) (*http.Response, error) {
	reader := NewSSEReader(response.Body)
	pr, pw := io.Pipe()

	// Create new response
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

	// Start transformation in goroutine
	go func() {
		defer pw.Close()
		writer := NewSSEWriter(pw)
		
		// Track state
		state := &toolUseStreamState{
			hasExitTool:       false,
			exitToolID:        "",
			suppressToolCalls: false,
		}

		for {
			event, err := reader.ReadEvent()
			if err != nil {
				if err != io.EOF {
					utils.GetLogger().Errorf("ToolUse: Error reading SSE event: %v", err)
				}
				break
			}

			// Pass through non-data events
			if event.Data == "" || event.Data == "[DONE]" {
				writer.WriteEvent(event)
				continue
			}

			// Transform the data event
			transformedEvent := t.transformStreamEvent(event, state)
			if transformedEvent != nil {
				writer.WriteEvent(transformedEvent)
			}
		}
	}()

	return newResp, nil
}

// toolUseStreamState tracks state during streaming
type toolUseStreamState struct {
	hasExitTool       bool
	exitToolID        string
	suppressToolCalls bool
}

// transformStreamEvent transforms a single stream event
func (t *ToolUseTransformer) transformStreamEvent(event *SSEEvent, state *toolUseStreamState) *SSEEvent {
	// Parse the JSON data
	var chunk map[string]interface{}
	if err := json.Unmarshal([]byte(event.Data), &chunk); err != nil {
		return event // Pass through on parse error
	}

	// Get choices array
	choices, ok := chunk["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return event
	}

	// Process first choice
	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return event
	}

	// Get delta
	delta, ok := choice["delta"].(map[string]interface{})
	if !ok {
		return event
	}

	// Check for tool calls
	if toolCalls, ok := delta["tool_calls"].([]interface{}); ok {
		filteredToolCalls := []interface{}{}
		
		for _, tc := range toolCalls {
			tcMap, ok := tc.(map[string]interface{})
			if !ok {
				filteredToolCalls = append(filteredToolCalls, tc)
				continue
			}

			// Check for function
			if function, ok := tcMap["function"].(map[string]interface{}); ok {
				// Check if it's ExitTool
				if function["name"] == "ExitTool" {
					state.hasExitTool = true
					if id, ok := tcMap["id"].(string); ok {
						state.exitToolID = id
					}
					// Don't include ExitTool in output
					continue
				}
			}

			// Include non-ExitTool calls
			filteredToolCalls = append(filteredToolCalls, tc)
		}

		// Update or remove tool_calls
		if len(filteredToolCalls) > 0 {
			delta["tool_calls"] = filteredToolCalls
		} else {
			delete(delta, "tool_calls")
			
			// If we filtered out ExitTool and there's no other content, add default content
			if state.hasExitTool && delta["content"] == nil {
				delta["content"] = "I have completed the requested task."
			}
		}
	}

	// If this is a finish event and we had ExitTool, ensure we have content
	if finishReason := choice["finish_reason"]; finishReason != nil && state.hasExitTool {
		if delta["content"] == nil && len(delta) == 0 {
			delta["content"] = "I have completed the requested task."
		}
	}

	// Serialize the modified chunk
	modifiedData, _ := json.Marshal(chunk)
	return &SSEEvent{Data: string(modifiedData)}
}