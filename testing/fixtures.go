package testing

import (
	"encoding/json"
	"fmt"
	"time"
)

// Fixtures provides test data fixtures
type Fixtures struct {
	providers map[string]interface{}
	requests  map[string]interface{}
	responses map[string]interface{}
}

// NewFixtures creates a new fixtures instance
func NewFixtures() *Fixtures {
	f := &Fixtures{
		providers: make(map[string]interface{}),
		requests:  make(map[string]interface{}),
		responses: make(map[string]interface{}),
	}
	
	f.loadDefaults()
	return f
}

// loadDefaults loads default fixtures
func (f *Fixtures) loadDefaults() {
	// Anthropic request fixture
	f.requests["anthropic_messages"] = map[string]interface{}{
		"model": "claude-3-sonnet-20240229",
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": "Hello, Claude!",
			},
		},
		"max_tokens": 100,
		"temperature": 0.7,
	}
	
	// Anthropic response fixture
	f.responses["anthropic_messages"] = map[string]interface{}{
		"id":   "msg_test_123",
		"type": "message",
		"role": "assistant",
		"model": "claude-3-sonnet-20240229",
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": "Hello! How can I assist you today?",
			},
		},
		"usage": map[string]interface{}{
			"input_tokens":  10,
			"output_tokens": 8,
		},
	}
	
	// OpenAI request fixture
	f.requests["openai_chat"] = map[string]interface{}{
		"model": "gpt-4",
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": "Hello, GPT!",
			},
		},
		"max_tokens": 100,
		"temperature": 0.7,
	}
	
	// OpenAI response fixture
	f.responses["openai_chat"] = map[string]interface{}{
		"id":      "chatcmpl-test123",
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   "gpt-4",
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"message": map[string]interface{}{
					"role": "assistant",
					"content": "Hello! How can I help you today?",
				},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]interface{}{
			"prompt_tokens": 10,
			"completion_tokens": 8,
			"total_tokens": 18,
		},
	}
	
	// Provider configurations
	f.providers["anthropic"] = map[string]interface{}{
		"name":     "anthropic-test",
		"type":     "anthropic",
		"api_key":  "test-anthropic-key",
		"base_url": "https://api.anthropic.com",
		"enabled":  true,
	}
	
	f.providers["openai"] = map[string]interface{}{
		"name":     "openai-test",
		"type":     "openai",
		"api_key":  "test-openai-key",
		"base_url": "https://api.openai.com/v1",
		"enabled":  true,
	}
}

// GetRequest returns a request fixture
func (f *Fixtures) GetRequest(name string) (map[string]interface{}, error) {
	if req, ok := f.requests[name]; ok {
		// Deep copy to avoid mutations
		data, _ := json.Marshal(req)
		var result map[string]interface{}
		json.Unmarshal(data, &result)
		return result, nil
	}
	return nil, fmt.Errorf("request fixture %s not found", name)
}

// GetResponse returns a response fixture
func (f *Fixtures) GetResponse(name string) (interface{}, error) {
	if resp, ok := f.responses[name]; ok {
		// Deep copy to avoid mutations
		data, _ := json.Marshal(resp)
		var result interface{}
		json.Unmarshal(data, &result)
		return result, nil
	}
	return nil, fmt.Errorf("response fixture %s not found", name)
}

// GetProvider returns a provider fixture
func (f *Fixtures) GetProvider(name string) (map[string]interface{}, error) {
	if prov, ok := f.providers[name]; ok {
		// Deep copy to avoid mutations
		data, _ := json.Marshal(prov)
		var result map[string]interface{}
		json.Unmarshal(data, &result)
		return result, nil
	}
	return nil, fmt.Errorf("provider fixture %s not found", name)
}

// GenerateLargeMessage generates a message with approximately the specified number of tokens
func (f *Fixtures) GenerateLargeMessage(tokens int) string {
	// Rough approximation: 1 token ≈ 4 characters
	chars := tokens * 4
	result := make([]byte, chars)
	
	// Fill with repeated pattern
	pattern := "The quick brown fox jumps over the lazy dog. "
	patternLen := len(pattern)
	
	for i := 0; i < chars; i++ {
		result[i] = pattern[i%patternLen]
	}
	
	return string(result)
}

// GenerateMessages generates multiple test messages
func (f *Fixtures) GenerateMessages(count int) []map[string]interface{} {
	messages := make([]map[string]interface{}, count)
	
	for i := 0; i < count; i++ {
		messages[i] = map[string]interface{}{
			"role": "user",
			"content": fmt.Sprintf("Test message %d: %s", i+1, f.GenerateLargeMessage(10)),
		}
	}
	
	return messages
}

// AddRequest adds a request fixture
func (f *Fixtures) AddRequest(name string, request map[string]interface{}) {
	f.requests[name] = request
}

// AddResponse adds a response fixture
func (f *Fixtures) AddResponse(name string, response interface{}) {
	f.responses[name] = response
}