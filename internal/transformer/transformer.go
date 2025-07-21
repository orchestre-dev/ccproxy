package transformer

import (
	"context"
	"net/http"
)

// Transformer defines the interface for request/response transformations
type Transformer interface {
	// GetName returns the transformer name
	GetName() string
	
	// GetEndpoint returns the endpoint this transformer handles (optional)
	GetEndpoint() string
	
	// TransformRequestIn transforms an incoming unified request to provider-specific format
	// Can return just the body, or RequestConfig with additional HTTP configuration
	TransformRequestIn(ctx context.Context, request interface{}, provider string) (interface{}, error)
	
	// TransformRequestOut transforms a provider-specific request to unified format
	TransformRequestOut(ctx context.Context, request interface{}) (interface{}, error)
	
	// TransformResponseIn transforms a provider response to internal format
	TransformResponseIn(ctx context.Context, response *http.Response) (*http.Response, error)
	
	// TransformResponseOut transforms an internal response to client format
	TransformResponseOut(ctx context.Context, response *http.Response) (*http.Response, error)
}

// StreamTransformer defines additional methods for streaming transformers
type StreamTransformer interface {
	Transformer
	
	// TransformStream handles streaming response transformation
	TransformStream(ctx context.Context, reader StreamReader, writer StreamWriter) error
}

// StreamReader provides methods to read from a stream
type StreamReader interface {
	// ReadEvent reads the next SSE event from the stream
	ReadEvent() (*SSEEvent, error)
	
	// Close closes the stream reader
	Close() error
}

// StreamWriter provides methods to write to a stream
type StreamWriter interface {
	// WriteEvent writes an SSE event to the stream
	WriteEvent(event *SSEEvent) error
	
	// Flush flushes any buffered data
	Flush() error
	
	// Close closes the stream writer
	Close() error
}

// SSEEvent represents a Server-Sent Event
type SSEEvent struct {
	Event string `json:"event,omitempty"`
	Data  string `json:"data"`
	ID    string `json:"id,omitempty"`
	Retry int    `json:"retry,omitempty"`
}

// RequestConfig allows transformers to modify HTTP request configuration
type RequestConfig struct {
	Body    interface{}            `json:"body"`
	URL     string                 `json:"url,omitempty"`
	Headers map[string]string      `json:"headers,omitempty"`
	Method  string                 `json:"method,omitempty"`
	Timeout int                    `json:"timeout,omitempty"`
}

// BaseTransformer provides default implementations
type BaseTransformer struct {
	name     string
	endpoint string
}

// NewBaseTransformer creates a new base transformer
func NewBaseTransformer(name string, endpoint string) *BaseTransformer {
	return &BaseTransformer{
		name:     name,
		endpoint: endpoint,
	}
}

func (t *BaseTransformer) GetName() string {
	return t.name
}

func (t *BaseTransformer) GetEndpoint() string {
	return t.endpoint
}

func (t *BaseTransformer) TransformRequestIn(ctx context.Context, request interface{}, provider string) (interface{}, error) {
	// Default: pass through unchanged
	return request, nil
}

func (t *BaseTransformer) TransformRequestOut(ctx context.Context, request interface{}) (interface{}, error) {
	// Default: pass through unchanged
	return request, nil
}

func (t *BaseTransformer) TransformResponseIn(ctx context.Context, response *http.Response) (*http.Response, error) {
	// Default: pass through unchanged
	return response, nil
}

func (t *BaseTransformer) TransformResponseOut(ctx context.Context, response *http.Response) (*http.Response, error) {
	// Default: pass through unchanged
	return response, nil
}

// TransformerChain represents a chain of transformers
type TransformerChain struct {
	transformers []Transformer
}

// NewTransformerChain creates a new transformer chain
func NewTransformerChain(transformers ...Transformer) *TransformerChain {
	return &TransformerChain{
		transformers: transformers,
	}
}

// Add adds a transformer to the chain
func (c *TransformerChain) Add(transformer Transformer) {
	c.transformers = append(c.transformers, transformer)
}

// TransformRequestIn applies all transformers' TransformRequestIn in order
func (c *TransformerChain) TransformRequestIn(ctx context.Context, request interface{}, provider string) (interface{}, error) {
	result := request
	for _, t := range c.transformers {
		var err error
		result, err = t.TransformRequestIn(ctx, result, provider)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// TransformResponseOut applies all transformers' TransformResponseOut in reverse order
func (c *TransformerChain) TransformResponseOut(ctx context.Context, response *http.Response) (*http.Response, error) {
	result := response
	// Apply in reverse order for responses
	for i := len(c.transformers) - 1; i >= 0; i-- {
		var err error
		result, err = c.transformers[i].TransformResponseOut(ctx, result)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// TransformSSEEvent transforms an SSE event through the chain
func (c *TransformerChain) TransformSSEEvent(ctx context.Context, event *SSEEvent, provider string) (*SSEEvent, error) {
	// For now, just return the event as-is
	// Individual transformers can override this if needed
	return event, nil
}