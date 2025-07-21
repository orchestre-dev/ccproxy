package transformer

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
)

// MockTransformer for testing
type MockTransformer struct {
	BaseTransformer
	RequestInCalled  bool
	RequestOutCalled bool
	ResponseInCalled bool
	ResponseOutCalled bool
	ReturnError      error
}

func NewMockTransformer(name string) *MockTransformer {
	return &MockTransformer{
		BaseTransformer: *NewBaseTransformer(name, ""),
	}
}

func (m *MockTransformer) TransformRequestIn(ctx context.Context, request interface{}, provider string) (interface{}, error) {
	m.RequestInCalled = true
	if m.ReturnError != nil {
		return nil, m.ReturnError
	}
	// Add suffix to track transformation
	if str, ok := request.(string); ok {
		return str + "-in", nil
	}
	return request, nil
}

func (m *MockTransformer) TransformRequestOut(ctx context.Context, request interface{}) (interface{}, error) {
	m.RequestOutCalled = true
	if m.ReturnError != nil {
		return nil, m.ReturnError
	}
	// Add suffix to track transformation
	if str, ok := request.(string); ok {
		return str + "-out", nil
	}
	return request, nil
}

func (m *MockTransformer) TransformResponseIn(ctx context.Context, response *http.Response) (*http.Response, error) {
	m.ResponseInCalled = true
	if m.ReturnError != nil {
		return nil, m.ReturnError
	}
	return response, nil
}

func (m *MockTransformer) TransformResponseOut(ctx context.Context, response *http.Response) (*http.Response, error) {
	m.ResponseOutCalled = true
	if m.ReturnError != nil {
		return nil, m.ReturnError
	}
	return response, nil
}

func TestBaseTransformer(t *testing.T) {
	transformer := NewBaseTransformer("test-transformer", "/test/endpoint")
	
	// Test GetName
	if transformer.GetName() != "test-transformer" {
		t.Errorf("Expected name 'test-transformer', got '%s'", transformer.GetName())
	}
	
	// Test GetEndpoint
	if transformer.GetEndpoint() != "/test/endpoint" {
		t.Errorf("Expected endpoint '/test/endpoint', got '%s'", transformer.GetEndpoint())
	}
	
	// Test pass-through behavior
	ctx := context.Background()
	
	// TransformRequestIn should pass through
	result, err := transformer.TransformRequestIn(ctx, "test-request", "provider")
	if err != nil {
		t.Errorf("TransformRequestIn failed: %v", err)
	}
	if result != "test-request" {
		t.Errorf("Expected 'test-request', got '%v'", result)
	}
	
	// TransformRequestOut should pass through
	result, err = transformer.TransformRequestOut(ctx, "test-request")
	if err != nil {
		t.Errorf("TransformRequestOut failed: %v", err)
	}
	if result != "test-request" {
		t.Errorf("Expected 'test-request', got '%v'", result)
	}
}

func TestTransformerChain(t *testing.T) {
	ctx := context.Background()
	
	// Create mock transformers
	t1 := NewMockTransformer("transformer1")
	t2 := NewMockTransformer("transformer2")
	t3 := NewMockTransformer("transformer3")
	
	// Create chain
	chain := NewTransformerChain(t1, t2, t3)
	
	// Test TransformRequestIn - should apply in order
	result, err := chain.TransformRequestIn(ctx, "test", "provider")
	if err != nil {
		t.Fatalf("TransformRequestIn failed: %v", err)
	}
	
	// Each transformer adds "-in", so we should get "test-in-in-in"
	expected := "test-in-in-in"
	if result != expected {
		t.Errorf("Expected '%s', got '%v'", expected, result)
	}
	
	// Verify all transformers were called
	if !t1.RequestInCalled || !t2.RequestInCalled || !t3.RequestInCalled {
		t.Error("Not all transformers were called")
	}
	
	// Test error propagation
	t2.ReturnError = fmt.Errorf("transform error")
	_, err = chain.TransformRequestIn(ctx, "test", "provider")
	if err == nil {
		t.Error("Expected error from chain")
	}
}

func TestTransformerChain_Add(t *testing.T) {
	chain := NewTransformerChain()
	
	// Add transformers
	t1 := NewMockTransformer("transformer1")
	t2 := NewMockTransformer("transformer2")
	
	chain.Add(t1)
	chain.Add(t2)
	
	// Test that both transformers are in the chain
	ctx := context.Background()
	result, err := chain.TransformRequestIn(ctx, "test", "provider")
	if err != nil {
		t.Fatalf("TransformRequestIn failed: %v", err)
	}
	
	if result != "test-in-in" {
		t.Errorf("Expected 'test-in-in', got '%v'", result)
	}
}

// OrderTrackingTransformer embeds MockTransformer and tracks call order
type OrderTrackingTransformer struct {
	MockTransformer
	Name      string
	CallOrder *[]string
	Mutex     *sync.Mutex
}

func (t *OrderTrackingTransformer) TransformResponseOut(ctx context.Context, response *http.Response) (*http.Response, error) {
	t.Mutex.Lock()
	*t.CallOrder = append(*t.CallOrder, t.Name)
	t.Mutex.Unlock()
	t.ResponseOutCalled = true
	return response, nil
}

func TestTransformerChain_ResponseOrder(t *testing.T) {
	ctx := context.Background()
	
	// Create mock transformers that track call order
	var callOrder []string
	var mu sync.Mutex
	
	t1 := &OrderTrackingTransformer{
		MockTransformer: MockTransformer{
			BaseTransformer: *NewBaseTransformer("t1", ""),
		},
		Name:      "t1",
		CallOrder: &callOrder,
		Mutex:     &mu,
	}
	
	t2 := &OrderTrackingTransformer{
		MockTransformer: MockTransformer{
			BaseTransformer: *NewBaseTransformer("t2", ""),
		},
		Name:      "t2",
		CallOrder: &callOrder,
		Mutex:     &mu,
	}
	
	t3 := &OrderTrackingTransformer{
		MockTransformer: MockTransformer{
			BaseTransformer: *NewBaseTransformer("t3", ""),
		},
		Name:      "t3",
		CallOrder: &callOrder,
		Mutex:     &mu,
	}
	
	// Create chain
	chain := NewTransformerChain(t1, t2, t3)
	
	// Test TransformResponseOut - should apply in reverse order
	resp := &http.Response{}
	_, err := chain.TransformResponseOut(ctx, resp)
	if err != nil {
		t.Fatalf("TransformResponseOut failed: %v", err)
	}
	
	// Check order is reversed: t3, t2, t1
	expectedOrder := []string{"t3", "t2", "t1"}
	if len(callOrder) != len(expectedOrder) {
		t.Errorf("Expected %d calls, got %d", len(expectedOrder), len(callOrder))
	}
	
	for i, name := range expectedOrder {
		if i >= len(callOrder) || callOrder[i] != name {
			t.Errorf("Expected call %d to be '%s', got '%v'", i, name, callOrder[i])
		}
	}
}

func TestRequestConfig(t *testing.T) {
	config := RequestConfig{
		Body: map[string]string{"key": "value"},
		URL:  "https://api.example.com/v1/endpoint",
		Headers: map[string]string{
			"Authorization": "Bearer token",
			"Content-Type":  "application/json",
		},
		Method:  "POST",
		Timeout: 30,
	}
	
	// Verify fields are set correctly
	if config.URL != "https://api.example.com/v1/endpoint" {
		t.Errorf("Expected URL to be set correctly")
	}
	
	if config.Headers["Authorization"] != "Bearer token" {
		t.Errorf("Expected Authorization header to be set")
	}
	
	if config.Method != "POST" {
		t.Errorf("Expected method to be POST")
	}
	
	if config.Timeout != 30 {
		t.Errorf("Expected timeout to be 30")
	}
}