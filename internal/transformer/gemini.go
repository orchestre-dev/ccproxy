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

// GeminiTransformer handles Google Gemini-specific transformations
type GeminiTransformer struct {
	BaseTransformer
}

// NewGeminiTransformer creates a new Gemini transformer
func NewGeminiTransformer() *GeminiTransformer {
	return &GeminiTransformer{
		BaseTransformer: *NewBaseTransformer("gemini", "/v1beta/models/:modelAndAction"),
	}
}

// TransformRequestIn transforms OpenAI format to Gemini format
func (t *GeminiTransformer) TransformRequestIn(ctx context.Context, request interface{}, provider string) (interface{}, error) {
	// Parse the incoming request
	reqMap, ok := request.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid request format")
	}

	// Create transformed request
	transformed := make(map[string]interface{})

	// Transform messages
	messages, ok := reqMap["messages"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("missing or invalid messages field")
	}

	// Gemini expects "contents" instead of "messages"
	contents := []interface{}{}
	var systemInstruction string

	for _, msg := range messages {
		msgMap, ok := msg.(map[string]interface{})
		if !ok {
			continue
		}

		role, _ := msgMap["role"].(string)
		content := msgMap["content"]

		// Handle system messages
		if role == "system" {
			if str, ok := content.(string); ok {
				systemInstruction = str
			}
			continue
		}

		// Transform role: assistant -> model, others -> user
		geminiRole := "user"
		if role == "assistant" {
			geminiRole = "model"
		}

		// Create content parts
		parts := []interface{}{}

		// Handle text content
		if str, ok := content.(string); ok {
			parts = append(parts, map[string]interface{}{
				"text": str,
			})
		}

		// Handle tool calls from assistant
		if role == "assistant" && msgMap["tool_calls"] != nil {
			toolCalls, ok := msgMap["tool_calls"].([]interface{})
			if ok {
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
						// Safe to ignore error - args will remain nil on parse failure
						_ = json.Unmarshal([]byte(argsStr), &args)
					}

					// Create function call part
					parts = append(parts, map[string]interface{}{
						"functionCall": map[string]interface{}{
							"name": funcMap["name"],
							"args": args,
						},
					})
				}
			}
		}

		// Handle tool responses
		if role == "tool" {
			// In Gemini, tool responses are function response parts
			parts = append(parts, map[string]interface{}{
				"functionResponse": map[string]interface{}{
					"name": msgMap["name"], // Tool name should be provided
					"response": map[string]interface{}{
						"result": content,
					},
				},
			})
		}

		// Add to contents
		contents = append(contents, map[string]interface{}{
			"role":  geminiRole,
			"parts": parts,
		})
	}

	transformed["contents"] = contents

	// Add system instruction if present
	if systemInstruction != "" {
		transformed["systemInstruction"] = systemInstruction
	}

	// Transform generation config
	genConfig := make(map[string]interface{})

	if temperature, ok := reqMap["temperature"]; ok {
		genConfig["temperature"] = temperature
	}
	if maxTokens, ok := reqMap["max_tokens"]; ok {
		genConfig["maxOutputTokens"] = maxTokens
	}
	if topP, ok := reqMap["top_p"]; ok {
		genConfig["topP"] = topP
	}

	if len(genConfig) > 0 {
		transformed["generationConfig"] = genConfig
	}

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

			// Clean up parameters - remove unsupported fields
			if params, ok := funcMap["parameters"].(map[string]interface{}); ok {
				cleanParams := t.cleanJSONSchema(params)
				funcMap["parameters"] = cleanParams
			}

			// Gemini expects function declarations at the tool level
			transformedTools = append(transformedTools, map[string]interface{}{
				"function_declarations": []interface{}{funcMap},
			})
		}

		if len(transformedTools) > 0 {
			transformed["tools"] = transformedTools
		}
	}

	return transformed, nil
}

// cleanJSONSchema removes unsupported fields from JSON schema
func (t *GeminiTransformer) cleanJSONSchema(schema map[string]interface{}) map[string]interface{} {
	cleaned := make(map[string]interface{})

	for k, v := range schema {
		// Skip unsupported fields
		if k == "$schema" || k == "additionalProperties" {
			continue
		}

		// Recursively clean nested objects
		if k == "properties" {
			if props, ok := v.(map[string]interface{}); ok {
				cleanedProps := make(map[string]interface{})
				for propName, propSchema := range props {
					if propMap, ok := propSchema.(map[string]interface{}); ok {
						cleanedProps[propName] = t.cleanJSONSchema(propMap)
					} else {
						cleanedProps[propName] = propSchema
					}
				}
				cleaned[k] = cleanedProps
			} else {
				cleaned[k] = v
			}
		} else {
			cleaned[k] = v
		}
	}

	return cleaned
}

// TransformRequestOut modifies the outgoing HTTP request
func (t *GeminiTransformer) TransformRequestOut(ctx context.Context, request interface{}) (interface{}, error) {
	// For Gemini, we need to modify the Authorization header
	// This is typically done at the HTTP client level, not in the transformer
	// The actual header transformation happens in the request processing pipeline
	return request, nil
}

// TransformResponseOut transforms Gemini response to OpenAI format
func (t *GeminiTransformer) TransformResponseOut(ctx context.Context, response *http.Response) (*http.Response, error) {
	// Check if it's a streaming response
	if strings.Contains(response.Header.Get("Content-Type"), "text/event-stream") {
		return t.transformStreamingResponse(ctx, response)
	}

	// Handle non-streaming response
	body, err := io.ReadAll(response.Body)
	_ = response.Body.Close() // Safe to ignore: already read all data
	if err != nil {
		return nil, err
	}

	// Parse Gemini response
	var geminiResp map[string]interface{}
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		// Return original response if we can't parse it
		response.Body = io.NopCloser(strings.NewReader(string(body)))
		return response, nil
	}

	// Transform to OpenAI format
	openaiResp := t.transformGeminiToOpenAI(geminiResp)

	// Marshal transformed response
	transformedBody, err := json.Marshal(openaiResp)
	if err != nil {
		return nil, err
	}

	// Create new response
	response.Body = io.NopCloser(strings.NewReader(string(transformedBody)))
	response.ContentLength = int64(len(transformedBody))
	return response, nil
}

// transformGeminiToOpenAI transforms Gemini response format to OpenAI format
func (t *GeminiTransformer) transformGeminiToOpenAI(geminiResp map[string]interface{}) map[string]interface{} {
	openaiResp := map[string]interface{}{
		"id":      fmt.Sprintf("chatcmpl-%d", utils.GetTimestamp()),
		"object":  "chat.completion",
		"created": utils.GetTimestamp(),
		"model":   "gemini-pro", // Default model name
	}

	// Transform candidates to choices
	choices := []interface{}{}

	if candidates, ok := geminiResp["candidates"].([]interface{}); ok {
		for i, candidate := range candidates {
			candMap, ok := candidate.(map[string]interface{})
			if !ok {
				continue
			}

			// Extract content
			message := map[string]interface{}{
				"role": "assistant",
			}

			var textContent strings.Builder
			var toolCalls []interface{}

			if content, ok := candMap["content"].(map[string]interface{}); ok {
				if parts, ok := content["parts"].([]interface{}); ok {
					for _, part := range parts {
						partMap, ok := part.(map[string]interface{})
						if !ok {
							continue
						}

						// Handle text parts
						if text, ok := partMap["text"].(string); ok {
							textContent.WriteString(text)
						}

						// Handle function calls
						if funcCall, ok := partMap["functionCall"].(map[string]interface{}); ok {
							toolCall := map[string]interface{}{
								"id":   fmt.Sprintf("call_%d", utils.GetTimestamp()),
								"type": "function",
								"function": map[string]interface{}{
									"name":      funcCall["name"],
									"arguments": utils.ToJSONString(funcCall["args"]),
								},
							}
							toolCalls = append(toolCalls, toolCall)
						}
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

			// Get finish reason
			finishReason := "stop"
			if reason, ok := candMap["finishReason"].(string); ok {
				finishReason = t.convertFinishReason(reason)
			}

			choice := map[string]interface{}{
				"index":         i,
				"message":       message,
				"finish_reason": finishReason,
			}
			choices = append(choices, choice)
		}
	}

	openaiResp["choices"] = choices

	// Transform usage metadata
	if metadata, ok := geminiResp["usageMetadata"].(map[string]interface{}); ok {
		usage := map[string]interface{}{}

		if promptTokens, ok := metadata["promptTokenCount"].(float64); ok {
			usage["prompt_tokens"] = int(promptTokens)
		}
		if candidateTokens, ok := metadata["candidatesTokenCount"].(float64); ok {
			usage["completion_tokens"] = int(candidateTokens)
		}
		if totalTokens, ok := metadata["totalTokenCount"].(float64); ok {
			usage["total_tokens"] = int(totalTokens)
		}

		openaiResp["usage"] = usage
	}

	return openaiResp
}

// convertFinishReason converts Gemini finish reason to OpenAI format
func (t *GeminiTransformer) convertFinishReason(reason string) string {
	switch reason {
	case "STOP":
		return "stop"
	case "MAX_TOKENS":
		return "length"
	case "SAFETY":
		return "content_filter"
	case "RECITATION":
		return "content_filter"
	default:
		return "stop"
	}
}

// transformStreamingResponse transforms Gemini streaming response
func (t *GeminiTransformer) transformStreamingResponse(ctx context.Context, response *http.Response) (*http.Response, error) {
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

		for {
			event, err := reader.ReadEvent()
			if err != nil {
				if err != io.EOF {
					utils.GetLogger().Errorf("Gemini: Error reading SSE event: %v", err)
				}
				break
			}

			// Transform the event
			transformed := t.transformStreamEvent(event)
			if transformed != nil {
				// Safe to ignore error for streaming output
				_ = writer.WriteEvent(transformed)
			}
		}

		// Send [DONE] event
		// Safe to ignore error for final streaming event
		_ = writer.WriteEvent(&SSEEvent{Data: "[DONE]"})
	}()

	return newResp, nil
}

// transformStreamEvent transforms a single Gemini SSE event
func (t *GeminiTransformer) transformStreamEvent(event *SSEEvent) *SSEEvent {
	// Parse the event data
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(event.Data), &data); err != nil {
		return nil
	}

	// Transform to OpenAI format
	chunk := map[string]interface{}{
		"id":      fmt.Sprintf("chatcmpl-%d", utils.GetTimestamp()),
		"object":  "chat.completion.chunk",
		"created": utils.GetTimestamp(),
		"model":   "gemini-pro",
		"choices": []interface{}{},
	}

	// Process candidates
	if candidates, ok := data["candidates"].([]interface{}); ok && len(candidates) > 0 {
		if candidate, ok := candidates[0].(map[string]interface{}); ok {
			delta := map[string]interface{}{}

			// Extract content
			if content, ok := candidate["content"].(map[string]interface{}); ok {
				if parts, ok := content["parts"].([]interface{}); ok && len(parts) > 0 {
					if part, ok := parts[0].(map[string]interface{}); ok {
						// Handle text
						if text, ok := part["text"].(string); ok {
							delta["content"] = text
						}

						// Handle function calls
						if funcCall, ok := part["functionCall"].(map[string]interface{}); ok {
							delta["tool_calls"] = []interface{}{
								map[string]interface{}{
									"index": 0,
									"id":    fmt.Sprintf("call_%d", utils.GetTimestamp()),
									"type":  "function",
									"function": map[string]interface{}{
										"name":      funcCall["name"],
										"arguments": utils.ToJSONString(funcCall["args"]),
									},
								},
							}
						}
					}
				}
			}

			// Handle finish reason
			finishReason := ""
			if reason, ok := candidate["finishReason"].(string); ok {
				finishReason = t.convertFinishReason(reason)
			}

			choice := map[string]interface{}{
				"index":         0,
				"delta":         delta,
				"finish_reason": nil,
			}

			if finishReason != "" {
				choice["finish_reason"] = finishReason
			}

			chunk["choices"] = []interface{}{choice}
		}
	}

	// Create transformed event
	transformedData, _ := json.Marshal(chunk)
	return &SSEEvent{
		Data: string(transformedData),
	}
}
