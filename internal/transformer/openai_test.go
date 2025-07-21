package transformer

import (
	"testing"
)

func TestNewOpenAITransformer(t *testing.T) {
	trans := NewOpenAITransformer()
	
	if trans == nil {
		t.Fatal("Expected non-nil transformer")
	}
	
	// Check name
	if trans.GetName() != "openai" {
		t.Errorf("Expected name 'openai', got %s", trans.GetName())
	}
	
	// Check endpoint
	if trans.GetEndpoint() != "/v1/chat/completions" {
		t.Errorf("Expected endpoint '/v1/chat/completions', got %s", trans.GetEndpoint())
	}
	
	// Verify it's using BaseTransformer behavior
	// The OpenAITransformer should pass through requests/responses unchanged
	testData := map[string]interface{}{
		"model": "gpt-4",
		"messages": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": "Hello",
			},
		},
	}
	
	// Test TransformRequestIn - should return input unchanged
	result, err := trans.TransformRequestIn(nil, testData, "openai")
	if err != nil {
		t.Errorf("TransformRequestIn failed: %v", err)
	}
	
	// Should return the same data
	if resultMap, ok := result.(map[string]interface{}); !ok {
		t.Error("Expected map result")
	} else if resultMap["model"] != "gpt-4" {
		t.Error("Expected request to be unchanged")
	}
}