package pipeline

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/orchestre-dev/ccproxy/internal/transformer"
	"github.com/orchestre-dev/ccproxy/internal/utils"
)

// StreamingProcessor handles streaming response processing
type StreamingProcessor struct {
	transformerService *transformer.Service
}

// NewStreamingProcessor creates a new streaming processor
func NewStreamingProcessor(transformerService *transformer.Service) *StreamingProcessor {
	return &StreamingProcessor{
		transformerService: transformerService,
	}
}

// ProcessStreamingResponse handles the complete streaming response flow
func (p *StreamingProcessor) ProcessStreamingResponse(
	ctx context.Context,
	w http.ResponseWriter,
	resp *http.Response,
	provider string,
) error {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable Nginx buffering
	
	// Ensure we can flush
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("response writer does not support flushing")
	}
	
	// Create SSE reader and writer
	reader := transformer.NewSSEReader(resp.Body)
	writer := transformer.NewSSEWriter(w)
	
	// Handle context cancellation
	done := make(chan struct{})
	defer close(done)
	
	go func() {
		select {
		case <-ctx.Done():
			reader.Close()
			writer.Close()
		case <-done:
			// Normal completion
		}
	}()
	
	// Get transformer chain for the provider
	chain := p.transformerService.GetChainForProvider(provider)
	if chain == nil {
		// If no chain, just pass through
		return p.passThrough(reader, writer, flusher)
	}
	
	// Process events through transformer chain
	eventCount := 0
	errorCount := 0
	
	for {
		// Read event
		event, err := reader.ReadEvent()
		if err != nil {
			if err == io.EOF {
				// Normal end of stream
				break
			}
			// Log error but try to continue
			utils.GetLogger().Warnf("Error reading SSE event: %v", err)
			errorCount++
			if errorCount > 10 {
				return fmt.Errorf("too many errors reading SSE stream")
			}
			continue
		}
		
		// Skip empty events
		if event.Data == "" && event.Event == "" {
			continue
		}
		
		// Apply transformations if this is a data event
		if event.Data != "" && !strings.HasPrefix(event.Data, "[DONE]") {
			transformedEvent, err := chain.TransformSSEEvent(ctx, event, provider)
			if err != nil {
				utils.GetLogger().Warnf("Error transforming SSE event: %v", err)
				// Continue with original event
				transformedEvent = event
			}
			event = transformedEvent
		}
		
		// Write event
		if err := writer.WriteEvent(event); err != nil {
			// Client disconnected or context cancelled
			if strings.Contains(err.Error(), "broken pipe") || 
			   strings.Contains(err.Error(), "connection reset") ||
			   strings.Contains(err.Error(), "writer is closed") {
				utils.GetLogger().Info("Client disconnected or context cancelled during streaming")
				return nil
			}
			return fmt.Errorf("error writing SSE event: %w", err)
		}
		
		// Flush after each event
		flusher.Flush()
		eventCount++
		
		// Check if this is the end marker
		if event.Data == "[DONE]" {
			break
		}
	}
	
	utils.GetLogger().Infof("Streamed %d events to client", eventCount)
	return nil
}

// passThrough handles streaming without transformation
func (p *StreamingProcessor) passThrough(
	reader *transformer.SSEReader,
	writer *transformer.SSEWriter,
	flusher http.Flusher,
) error {
	defer reader.Close()
	
	for {
		event, err := reader.ReadEvent()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		
		if err := writer.WriteEvent(event); err != nil {
			// Check for expected errors during cancellation
			if strings.Contains(err.Error(), "writer is closed") {
				return nil
			}
			return err
		}
		
		flusher.Flush()
		
		if event.Data == "[DONE]" {
			break
		}
	}
	
	return nil
}

