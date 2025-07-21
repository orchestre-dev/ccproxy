package transformer

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/orchestre-dev/ccproxy/internal/utils"
)

// OpenRouterTransformer handles OpenRouter-specific transformations
type OpenRouterTransformer struct {
	BaseTransformer
}

// NewOpenRouterTransformer creates a new OpenRouter transformer
func NewOpenRouterTransformer() *OpenRouterTransformer {
	return &OpenRouterTransformer{
		BaseTransformer: *NewBaseTransformer("openrouter", "/api/v1/chat/completions"),
	}
}

// TransformResponseOut transforms OpenRouter streaming response
func (t *OpenRouterTransformer) TransformResponseOut(ctx context.Context, response *http.Response) (*http.Response, error) {
	// Only process streaming responses
	if !strings.Contains(response.Header.Get("Content-Type"), "text/event-stream") {
		return response, nil
	}

	// Create SSE reader and writer
	reader := NewSSEReader(response.Body)
	pr, pw := io.Pipe()

	// Create new response with transformed body
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

		// Track state for reasoning transformation
		state := &openrouterStreamState{
			reasoningContent:    "",
			isReasoningComplete: false,
			contentIndex:        0,
		}

		for {
			event, err := reader.ReadEvent()
			if err != nil {
				if err != io.EOF {
					utils.GetLogger().Errorf("OpenRouter: Error reading SSE event: %v", err)
				}
				break
			}

			// Pass through non-data events
			if event.Data == "" || event.Data == "[DONE]" {
				writer.WriteEvent(event)
				continue
			}

			// Transform the data event
			transformedData, err := t.transformStreamData(event.Data, state)
			if err != nil {
				utils.GetLogger().Errorf("OpenRouter: Error transforming stream data: %v", err)
				// Pass through original on error
				writer.WriteEvent(event)
				continue
			}

			// Write transformed events
			for _, data := range transformedData {
				if data != "" {
					writer.WriteEvent(&SSEEvent{Data: data})
				}
			}
		}
	}()

	return newResp, nil
}

// openrouterStreamState tracks state during streaming
type openrouterStreamState struct {
	reasoningContent    string
	isReasoningComplete bool
	contentIndex        int
	hasToolCalls        bool
}

// transformStreamData transforms a single stream data chunk
func (t *OpenRouterTransformer) transformStreamData(data string, state *openrouterStreamState) ([]string, error) {
	// Parse the JSON data
	var chunk map[string]interface{}
	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		return []string{data}, nil // Pass through on parse error
	}

	// Get choices array
	choices, ok := chunk["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return []string{data}, nil
	}

	// Process first choice
	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return []string{data}, nil
	}

	// Get delta (might not exist for finish_reason only chunks)
	delta, hasDelta := choice["delta"].(map[string]interface{})
	if !hasDelta {
		// Create empty delta for finish_reason only chunks
		delta = make(map[string]interface{})
	}

	// Check for reasoning content
	reasoningContent, hasReasoning := delta["reasoning_content"].(string)
	regularContent, hasContent := delta["content"].(string)
	toolCalls, hasToolCalls := delta["tool_calls"]

	results := []string{}

	// Handle reasoning content
	if hasReasoning && reasoningContent != "" {
		state.reasoningContent += reasoningContent

		// Transform to thinking delta
		delta["thinking"] = map[string]interface{}{
			"content": reasoningContent,
		}
		delete(delta, "reasoning_content")

		// Serialize and add to results
		modifiedData, _ := json.Marshal(chunk)
		results = append(results, string(modifiedData))
	}

	// Handle transition from reasoning to regular content or tool calls
	if !state.isReasoningComplete && (hasContent || hasToolCalls) {
		state.isReasoningComplete = true

		// First, send the complete thinking block
		if state.reasoningContent != "" {
			thinkingChunk := t.createThinkingBlockChunk(chunk, state.reasoningContent, state.contentIndex)
			thinkingData, _ := json.Marshal(thinkingChunk)
			results = append(results, string(thinkingData))

			// Increment index for subsequent content
			state.contentIndex++
		}
	}

	// Handle regular content (after reasoning or without reasoning)
	if hasContent && regularContent != "" {
		// Update index if we had reasoning content
		if state.isReasoningComplete {
			choice["index"] = state.contentIndex
		}

		// Remove reasoning_content if it exists
		delete(delta, "reasoning_content")

		// Serialize and add to results
		modifiedData, _ := json.Marshal(chunk)
		results = append(results, string(modifiedData))
	}

	// Handle tool calls
	if hasToolCalls {
		state.hasToolCalls = true

		// Update index if we had reasoning content
		if state.isReasoningComplete {
			// For tool calls, we need to update the index in each tool call
			if toolCallsArray, ok := toolCalls.([]interface{}); ok {
				for _, tc := range toolCallsArray {
					if tcMap, ok := tc.(map[string]interface{}); ok {
						if idx, hasIdx := tcMap["index"].(float64); hasIdx {
							tcMap["index"] = idx + float64(state.contentIndex)
						}
					}
				}
			}
		}

		// Remove reasoning_content if it exists
		delete(delta, "reasoning_content")

		// Serialize and add to results
		modifiedData, _ := json.Marshal(chunk)
		results = append(results, string(modifiedData))
	}

	// Handle finish reason - need to check if results is empty
	if finishReason := choice["finish_reason"]; finishReason != nil && finishReason != "" {
		// Debug logging
		// fmt.Printf("DEBUG: finish_reason found, isReasoningComplete=%v, reasoningContent='%s'\n", state.isReasoningComplete, state.reasoningContent)

		// Make sure we send any remaining reasoning content first
		if !state.isReasoningComplete && state.reasoningContent != "" {
			state.isReasoningComplete = true
			thinkingChunk := t.createThinkingBlockChunk(chunk, state.reasoningContent, state.contentIndex)
			thinkingData, _ := json.Marshal(thinkingChunk)
			results = append(results, string(thinkingData))

			// Increment index for the finish reason chunk
			state.contentIndex++
		}

		// Update the index for finish reason if we had reasoning
		if state.isReasoningComplete {
			choice["index"] = state.contentIndex
		}

		// Pass through the finish chunk
		modifiedData, _ := json.Marshal(chunk)
		results = append(results, string(modifiedData))

		return results, nil
	}

	return results, nil
}

// createThinkingBlockChunk creates a chunk with a complete thinking block
func (t *OpenRouterTransformer) createThinkingBlockChunk(baseChunk map[string]interface{}, content string, index int) map[string]interface{} {
	// Deep copy the chunk structure
	chunk := make(map[string]interface{})
	for k, v := range baseChunk {
		chunk[k] = v
	}

	// Create new choices array with thinking block
	choices := []interface{}{
		map[string]interface{}{
			"index": index,
			"delta": map[string]interface{}{
				"content": map[string]interface{}{
					"content":   content,
					"signature": strconv.FormatInt(time.Now().Unix(), 10),
				},
			},
		},
	}
	chunk["choices"] = choices

	return chunk
}
