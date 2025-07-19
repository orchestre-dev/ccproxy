package transformer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/musistudio/ccproxy/internal/utils"
)

// DeepSeekTransformer handles DeepSeek-specific transformations
type DeepSeekTransformer struct {
	BaseTransformer
}

// NewDeepSeekTransformer creates a new DeepSeek transformer
func NewDeepSeekTransformer() *DeepSeekTransformer {
	return &DeepSeekTransformer{
		BaseTransformer: *NewBaseTransformer("deepseek", "/v1/chat/completions"),
	}
}

// TransformRequestIn transforms the request for DeepSeek
func (t *DeepSeekTransformer) TransformRequestIn(ctx context.Context, request interface{}, provider string) (interface{}, error) {
	// Parse the incoming request
	reqMap, ok := request.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid request format")
	}

	// DeepSeek has a hard limit of 8192 max_tokens
	if maxTokens, ok := reqMap["max_tokens"].(float64); ok && maxTokens > 8192 {
		reqMap["max_tokens"] = 8192
		utils.GetLogger().Debugf("DeepSeek: Limited max_tokens from %.0f to 8192", maxTokens)
	}

	return reqMap, nil
}

// TransformResponseOut transforms DeepSeek streaming response
func (t *DeepSeekTransformer) TransformResponseOut(ctx context.Context, response *http.Response) (*http.Response, error) {
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
		state := &deepseekStreamState{
			reasoningContent:    "",
			isReasoningComplete: false,
			contentIndex:        0,
		}

		for {
			event, err := reader.ReadEvent()
			if err != nil {
				if err != io.EOF {
					utils.GetLogger().Errorf("DeepSeek: Error reading SSE event: %v", err)
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
				utils.GetLogger().Errorf("DeepSeek: Error transforming stream data: %v", err)
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

// deepseekStreamState tracks state during streaming
type deepseekStreamState struct {
	reasoningContent    string
	isReasoningComplete bool
	contentIndex        int
}

// transformStreamData transforms a single stream data chunk
func (t *DeepSeekTransformer) transformStreamData(data string, state *deepseekStreamState) ([]string, error) {
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

	// Get delta
	delta, ok := choice["delta"].(map[string]interface{})
	if !ok {
		return []string{data}, nil
	}

	// Check for reasoning content
	reasoningContent, hasReasoning := delta["reasoning_content"].(string)
	regularContent, hasContent := delta["content"].(string)

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

	// Handle transition from reasoning to regular content
	if !state.isReasoningComplete && hasContent && regularContent != "" {
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

	// Handle finish reason
	if finishReason := choice["finish_reason"]; finishReason != nil && finishReason != "" {
		// Make sure we send any remaining reasoning content
		if !state.isReasoningComplete && state.reasoningContent != "" {
			state.isReasoningComplete = true
			thinkingChunk := t.createThinkingBlockChunk(chunk, state.reasoningContent, state.contentIndex)
			thinkingData, _ := json.Marshal(thinkingChunk)
			results = append(results, string(thinkingData))
		}

		// Pass through the finish chunk
		modifiedData, _ := json.Marshal(chunk)
		results = append(results, string(modifiedData))
	}

	return results, nil
}

// createThinkingBlockChunk creates a chunk with a complete thinking block
func (t *DeepSeekTransformer) createThinkingBlockChunk(baseChunk map[string]interface{}, content string, index int) map[string]interface{} {
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