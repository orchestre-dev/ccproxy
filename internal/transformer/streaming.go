package transformer

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

// SSEReader implements StreamReader for Server-Sent Events
type SSEReader struct {
	reader *bufio.Reader
	closer io.Closer
	mu     sync.Mutex
	closed bool
}

// NewSSEReader creates a new SSE reader
func NewSSEReader(r io.ReadCloser) *SSEReader {
	return &SSEReader{
		reader: bufio.NewReader(r),
		closer: r,
	}
}

// ReadEvent reads the next SSE event
func (r *SSEReader) ReadEvent() (*SSEEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil, io.EOF
	}

	event := &SSEEvent{}
	var dataLines []string

	for {
		line, err := r.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF && len(dataLines) > 0 {
				// Process any remaining data
				event.Data = strings.Join(dataLines, "\n")
				return event, nil
			}
			return nil, err
		}

		line = strings.TrimRight(line, "\r\n")

		// Empty line signals end of event
		if line == "" && (event.Event != "" || len(dataLines) > 0) {
			event.Data = strings.Join(dataLines, "\n")
			return event, nil
		}

		// Parse field
		if strings.HasPrefix(line, "event: ") {
			event.Event = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "data: ") {
			dataLines = append(dataLines, strings.TrimPrefix(line, "data: "))
		} else if strings.HasPrefix(line, "id: ") {
			event.ID = strings.TrimPrefix(line, "id: ")
		} else if strings.HasPrefix(line, "retry: ") {
			// Parse retry value (currently ignored)
			_ = strings.TrimPrefix(line, "retry: ")
		}
	}
}

// Close closes the reader
func (r *SSEReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	r.closed = true
	return r.closer.Close()
}

// SSEWriter implements StreamWriter for Server-Sent Events
type SSEWriter struct {
	writer  io.Writer
	flusher http.Flusher
	mu      sync.Mutex
	closed  bool
}

// NewSSEWriter creates a new SSE writer
func NewSSEWriter(w io.Writer) *SSEWriter {
	writer := &SSEWriter{
		writer: w,
	}

	// Check if writer supports flushing
	if f, ok := w.(http.Flusher); ok {
		writer.flusher = f
	}

	return writer
}

// WriteEvent writes an SSE event
func (w *SSEWriter) WriteEvent(event *SSEEvent) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return fmt.Errorf("writer is closed")
	}

	var buf bytes.Buffer

	// Write event type
	if event.Event != "" {
		fmt.Fprintf(&buf, "event: %s\n", event.Event)
	}

	// Write ID
	if event.ID != "" {
		fmt.Fprintf(&buf, "id: %s\n", event.ID)
	}

	// Write retry
	if event.Retry > 0 {
		fmt.Fprintf(&buf, "retry: %d\n", event.Retry)
	}

	// Write data (can be multiline)
	if event.Data != "" {
		lines := strings.Split(event.Data, "\n")
		for _, line := range lines {
			fmt.Fprintf(&buf, "data: %s\n", line)
		}
	}

	// End of event
	buf.WriteString("\n")

	// Write to underlying writer
	_, err := w.writer.Write(buf.Bytes())
	if err != nil {
		return err
	}

	// Flush if possible
	if w.flusher != nil {
		w.flusher.Flush()
	}

	return nil
}

// Flush flushes any buffered data
func (w *SSEWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.flusher != nil {
		w.flusher.Flush()
	}

	return nil
}

// Close closes the writer
func (w *SSEWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil
	}

	w.closed = true

	// Flush any remaining data
	if w.flusher != nil {
		w.flusher.Flush()
	}

	// If the writer is also a closer, close it
	if closer, ok := w.writer.(io.Closer); ok {
		return closer.Close()
	}

	return nil
}

// StreamPipe pipes data from reader to writer with optional transformation
func StreamPipe(reader StreamReader, writer StreamWriter, transform func(*SSEEvent) (*SSEEvent, error)) error {
	defer reader.Close()
	defer writer.Close()

	for {
		event, err := reader.ReadEvent()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		// Apply transformation if provided
		if transform != nil {
			event, err = transform(event)
			if err != nil {
				return err
			}

			// Skip nil events (filtered out)
			if event == nil {
				continue
			}
		}

		// Write transformed event
		if err := writer.WriteEvent(event); err != nil {
			return err
		}
	}
}
