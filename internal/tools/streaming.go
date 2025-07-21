package tools

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/orchestre-dev/ccproxy/internal/utils"
	"github.com/sirupsen/logrus"
)

// StreamingHandler processes tool-related SSE streams
type StreamingHandler struct {
	handler *Handler
	logger  *logrus.Logger
}

// NewStreamingHandler creates a new streaming handler
func NewStreamingHandler() *StreamingHandler {
	return &StreamingHandler{
		handler: NewHandler(),
		logger:  utils.GetLogger(),
	}
}

// SSEEvent represents a Server-Sent Event
type SSEEvent struct {
	Event string
	Data  string
	ID    string
	Retry string
}

// ProcessToolStream processes a stream that may contain tool use events
func (sh *StreamingHandler) ProcessToolStream(ctx context.Context, reader io.Reader, writer io.Writer) error {
	scanner := bufio.NewScanner(reader)
	var currentEvent SSEEvent
	var buffer bytes.Buffer

	for scanner.Scan() {
		line := scanner.Text()

		// Empty line indicates end of event
		if line == "" {
			if currentEvent.Data != "" {
				// Process the event
				processedEvent, err := sh.processSSEEvent(ctx, currentEvent)
				if err != nil {
					sh.logger.Warnf("Failed to process SSE event: %v", err)
					// Write original event on error
					sh.writeSSEEvent(writer, currentEvent)
				} else {
					sh.writeSSEEvent(writer, processedEvent)
				}
			}
			// Reset for next event
			currentEvent = SSEEvent{}
			buffer.Reset()
			continue
		}

		// Parse SSE fields
		if strings.HasPrefix(line, "event:") {
			currentEvent.Event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if currentEvent.Data != "" {
				currentEvent.Data += "\n" + data
			} else {
				currentEvent.Data = data
			}
		} else if strings.HasPrefix(line, "id:") {
			currentEvent.ID = strings.TrimSpace(strings.TrimPrefix(line, "id:"))
		} else if strings.HasPrefix(line, "retry:") {
			currentEvent.Retry = strings.TrimSpace(strings.TrimPrefix(line, "retry:"))
		}
	}

	// Handle last event if exists
	if currentEvent.Data != "" {
		processedEvent, err := sh.processSSEEvent(ctx, currentEvent)
		if err != nil {
			sh.writeSSEEvent(writer, currentEvent)
		} else {
			sh.writeSSEEvent(writer, processedEvent)
		}
	}

	return scanner.Err()
}

// processSSEEvent processes a single SSE event
func (sh *StreamingHandler) processSSEEvent(ctx context.Context, event SSEEvent) (SSEEvent, error) {
	// Only process data events
	if event.Data == "" || event.Data == "[DONE]" {
		return event, nil
	}

	// Try to parse as JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(event.Data), &data); err != nil {
		// Not JSON, return as-is
		return event, nil
	}

	// Check for content delta with tool use
	if delta, ok := data["delta"].(map[string]interface{}); ok {
		if processedDelta, modified := sh.processContentDelta(ctx, delta); modified {
			data["delta"] = processedDelta
			
			// Add metadata
			if metadata, ok := data["metadata"].(map[string]interface{}); ok {
				metadata["has_tool_use"] = true
			} else {
				data["metadata"] = map[string]interface{}{
					"has_tool_use": true,
				}
			}
			
			// Marshal back to JSON
			newData, err := json.Marshal(data)
			if err != nil {
				return event, err
			}
			event.Data = string(newData)
		}
	}

	// Check for message with content blocks
	if content, ok := data["content"].([]interface{}); ok {
		if processedContent, modified := sh.processStreamingContent(ctx, content); modified {
			data["content"] = processedContent
			
			// Marshal back to JSON
			newData, err := json.Marshal(data)
			if err != nil {
				return event, err
			}
			event.Data = string(newData)
		}
	}

	return event, nil
}

// processContentDelta processes a content delta that may contain tool use
func (sh *StreamingHandler) processContentDelta(ctx context.Context, delta map[string]interface{}) (map[string]interface{}, bool) {
	deltaType, _ := delta["type"].(string)
	if deltaType != "tool_use" {
		return delta, false
	}

	// Mark as processed
	delta["_processed"] = true
	delta["_processor"] = "ccproxy-streaming"

	// Log tool use detection
	if name, ok := delta["name"].(string); ok {
		sh.logger.Debugf("Detected streaming tool use: %s", name)
	}

	return delta, true
}

// processStreamingContent processes content blocks in streaming
func (sh *StreamingHandler) processStreamingContent(ctx context.Context, content []interface{}) ([]interface{}, bool) {
	var modified bool
	var processed []interface{}

	for _, block := range content {
		blockMap, ok := block.(map[string]interface{})
		if !ok {
			processed = append(processed, block)
			continue
		}

		blockType, _ := blockMap["type"].(string)
		if blockType == "tool_use" {
			// Process tool use block
			blockMap["_processed"] = true
			blockMap["_processor"] = "ccproxy-streaming"
			modified = true
			
			if name, ok := blockMap["name"].(string); ok {
				sh.logger.Debugf("Processing streaming tool use block: %s", name)
			}
		}
		
		processed = append(processed, blockMap)
	}

	return processed, modified
}

// writeSSEEvent writes an SSE event to the writer
func (sh *StreamingHandler) writeSSEEvent(w io.Writer, event SSEEvent) error {
	var err error
	
	if event.ID != "" {
		_, err = fmt.Fprintf(w, "id: %s\n", event.ID)
		if err != nil {
			return err
		}
	}
	
	if event.Event != "" {
		_, err = fmt.Fprintf(w, "event: %s\n", event.Event)
		if err != nil {
			return err
		}
	}
	
	if event.Retry != "" {
		_, err = fmt.Fprintf(w, "retry: %s\n", event.Retry)
		if err != nil {
			return err
		}
	}
	
	if event.Data != "" {
		// Handle multi-line data
		lines := strings.Split(event.Data, "\n")
		for _, line := range lines {
			_, err = fmt.Fprintf(w, "data: %s\n", line)
			if err != nil {
				return err
			}
		}
	}
	
	// Empty line to end event
	_, err = fmt.Fprintln(w)
	return err
}

// ExtractToolUseFromStream extracts tool use information from a streaming event
func (sh *StreamingHandler) ExtractToolUseFromStream(event SSEEvent) (*ToolUse, error) {
	if event.Data == "" || event.Data == "[DONE]" {
		return nil, nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(event.Data), &data); err != nil {
		return nil, nil
	}

	// Check in delta
	if delta, ok := data["delta"].(map[string]interface{}); ok {
		if deltaType, _ := delta["type"].(string); deltaType == "tool_use" {
			return sh.extractToolUseFromMap(delta)
		}
	}

	// Check in content blocks
	if content, ok := data["content"].([]interface{}); ok {
		for _, block := range content {
			if blockMap, ok := block.(map[string]interface{}); ok {
				if blockType, _ := blockMap["type"].(string); blockType == "tool_use" {
					return sh.extractToolUseFromMap(blockMap)
				}
			}
		}
	}

	return nil, nil
}

// extractToolUseFromMap extracts tool use from a map
func (sh *StreamingHandler) extractToolUseFromMap(m map[string]interface{}) (*ToolUse, error) {
	toolUse := &ToolUse{
		Type: "tool_use",
	}

	if id, ok := m["id"].(string); ok {
		toolUse.ID = id
	}
	if name, ok := m["name"].(string); ok {
		toolUse.Name = name
	}
	if input, ok := m["input"]; ok {
		inputBytes, err := json.Marshal(input)
		if err != nil {
			return nil, err
		}
		toolUse.Input = inputBytes
	}

	return toolUse, nil
}