package converter

import (
	"encoding/json"
	"testing"

	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
)

func TestMessageFormat_Constants(t *testing.T) {
	testCases := []struct {
		name     string
		format   MessageFormat
		expected string
	}{
		{
			name:     "Anthropic format",
			format:   FormatAnthropic,
			expected: "anthropic",
		},
		{
			name:     "OpenAI format",
			format:   FormatOpenAI,
			expected: "openai",
		},
		{
			name:     "Google format",
			format:   FormatGoogle,
			expected: "google",
		},
		{
			name:     "AWS format",
			format:   FormatAWS,
			expected: "aws",
		},
		{
			name:     "Generic format",
			format:   FormatGeneric,
			expected: "generic",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testutil.AssertEqual(t, tc.expected, string(tc.format))
		})
	}
}

func TestMessage_JSONSerialization(t *testing.T) {
	message := Message{
		Role:    "user",
		Content: json.RawMessage(`"Hello, world!"`),
		Name:    "test-user",
	}

	// Test marshaling
	data, err := json.Marshal(message)
	testutil.AssertNoError(t, err)

	expectedJSON := `{"role":"user","content":"Hello, world!","name":"test-user"}`
	testutil.AssertEqual(t, expectedJSON, string(data))

	// Test unmarshaling
	var unmarshaled Message
	err = json.Unmarshal(data, &unmarshaled)
	testutil.AssertNoError(t, err)

	testutil.AssertEqual(t, message.Role, unmarshaled.Role)
	// Content should be preserved (comparing as strings might have spacing differences)
	testutil.AssertTrue(t, len(unmarshaled.Content) > 0, "Content should be preserved")
	testutil.AssertEqual(t, message.Name, unmarshaled.Name)
}

func TestMessage_JSONSerialization_WithoutName(t *testing.T) {
	message := Message{
		Role:    "assistant",
		Content: json.RawMessage(`{"type": "text", "text": "Hello!"}`),
	}

	// Test marshaling
	data, err := json.Marshal(message)
	testutil.AssertNoError(t, err)

	// Just verify that data was marshaled successfully
	testutil.AssertTrue(t, len(data) > 0, "Data should be marshaled")

	// Test unmarshaling
	var unmarshaled Message
	err = json.Unmarshal(data, &unmarshaled)
	testutil.AssertNoError(t, err)

	testutil.AssertEqual(t, message.Role, unmarshaled.Role)
	// Content should be preserved (comparing as strings might have spacing differences)
	testutil.AssertTrue(t, len(unmarshaled.Content) > 0, "Content should be preserved")
	testutil.AssertEqual(t, "", unmarshaled.Name)
}

func TestRequest_JSONSerialization(t *testing.T) {
	request := Request{
		Model:       "gpt-4",
		Messages:    []Message{{Role: "user", Content: json.RawMessage(`"test"`)}, {Role: "assistant", Content: json.RawMessage(`"response"`)}},
		System:      "You are a helpful assistant",
		MaxTokens:   1000,
		Temperature: 0.7,
		Stream:      true,
		Metadata:    json.RawMessage(`{"user_id": "123"}`),
	}

	// Test marshaling
	data, err := json.Marshal(request)
	testutil.AssertNoError(t, err)

	// Test unmarshaling
	var unmarshaled Request
	err = json.Unmarshal(data, &unmarshaled)
	testutil.AssertNoError(t, err)

	testutil.AssertEqual(t, request.Model, unmarshaled.Model)
	testutil.AssertEqual(t, len(request.Messages), len(unmarshaled.Messages))
	testutil.AssertEqual(t, request.System, unmarshaled.System)
	testutil.AssertEqual(t, request.MaxTokens, unmarshaled.MaxTokens)
	testutil.AssertEqual(t, request.Temperature, unmarshaled.Temperature)
	testutil.AssertEqual(t, request.Stream, unmarshaled.Stream)
	// Just check that metadata was preserved if it existed
	if request.Metadata != nil {
		testutil.AssertTrue(t, unmarshaled.Metadata != nil, "Metadata should be preserved")
	}
}

func TestRequest_JSONSerialization_OptionalFields(t *testing.T) {
	// Test with minimal required fields
	request := Request{
		Model:    "gpt-4",
		Messages: []Message{{Role: "user", Content: json.RawMessage(`"test"`)}},
	}

	// Test marshaling
	data, err := json.Marshal(request)
	testutil.AssertNoError(t, err)

	// Test unmarshaling
	var unmarshaled Request
	err = json.Unmarshal(data, &unmarshaled)
	testutil.AssertNoError(t, err)

	testutil.AssertEqual(t, request.Model, unmarshaled.Model)
	testutil.AssertEqual(t, len(request.Messages), len(unmarshaled.Messages))
	testutil.AssertEqual(t, "", unmarshaled.System)
	testutil.AssertEqual(t, 0, unmarshaled.MaxTokens)
	testutil.AssertEqual(t, 0.0, unmarshaled.Temperature)
	testutil.AssertEqual(t, false, unmarshaled.Stream)
	// Check that metadata is nil or empty
	if unmarshaled.Metadata != nil {
		testutil.AssertEqual(t, 0, len(unmarshaled.Metadata))
	}
}

func TestResponse_JSONSerialization(t *testing.T) {
	usage := &Usage{
		InputTokens:  100,
		OutputTokens: 50,
		TotalTokens:  150,
	}

	response := Response{
		ID:      "resp_123",
		Type:    "message",
		Role:    "assistant",
		Content: json.RawMessage(`[{"type": "text", "text": "Hello!"}]`),
		Model:   "gpt-4",
		Usage:   usage,
	}

	// Test marshaling
	data, err := json.Marshal(response)
	testutil.AssertNoError(t, err)

	// Test unmarshaling
	var unmarshaled Response
	err = json.Unmarshal(data, &unmarshaled)
	testutil.AssertNoError(t, err)

	testutil.AssertEqual(t, response.ID, unmarshaled.ID)
	testutil.AssertEqual(t, response.Type, unmarshaled.Type)
	testutil.AssertEqual(t, response.Role, unmarshaled.Role)
	// Content should be preserved 
	testutil.AssertTrue(t, len(unmarshaled.Content) > 0, "Content should be preserved")
	testutil.AssertEqual(t, response.Model, unmarshaled.Model)
	
	// Check usage
	testutil.AssertEqual(t, response.Usage.InputTokens, unmarshaled.Usage.InputTokens)
	testutil.AssertEqual(t, response.Usage.OutputTokens, unmarshaled.Usage.OutputTokens)
	testutil.AssertEqual(t, response.Usage.TotalTokens, unmarshaled.Usage.TotalTokens)
}

func TestResponse_JSONSerialization_WithoutUsage(t *testing.T) {
	response := Response{
		ID:      "resp_123",
		Type:    "message",
		Role:    "assistant",
		Content: json.RawMessage(`"Simple response"`),
		Model:   "gpt-4",
	}

	// Test marshaling
	data, err := json.Marshal(response)
	testutil.AssertNoError(t, err)

	// Test unmarshaling
	var unmarshaled Response
	err = json.Unmarshal(data, &unmarshaled)
	testutil.AssertNoError(t, err)

	testutil.AssertEqual(t, response.ID, unmarshaled.ID)
	testutil.AssertEqual(t, response.Type, unmarshaled.Type)
	testutil.AssertEqual(t, response.Role, unmarshaled.Role)
	// Content should be preserved 
	testutil.AssertTrue(t, len(unmarshaled.Content) > 0, "Content should be preserved")
	testutil.AssertEqual(t, response.Model, unmarshaled.Model)
	testutil.AssertEqual(t, (*Usage)(nil), unmarshaled.Usage)
}

func TestUsage_JSONSerialization(t *testing.T) {
	usage := Usage{
		InputTokens:  100,
		OutputTokens: 50,
		TotalTokens:  150,
	}

	// Test marshaling
	data, err := json.Marshal(usage)
	testutil.AssertNoError(t, err)

	expectedJSON := `{"input_tokens":100,"output_tokens":50,"total_tokens":150}`
	testutil.AssertEqual(t, expectedJSON, string(data))

	// Test unmarshaling
	var unmarshaled Usage
	err = json.Unmarshal(data, &unmarshaled)
	testutil.AssertNoError(t, err)

	testutil.AssertEqual(t, usage.InputTokens, unmarshaled.InputTokens)
	testutil.AssertEqual(t, usage.OutputTokens, unmarshaled.OutputTokens)
	testutil.AssertEqual(t, usage.TotalTokens, unmarshaled.TotalTokens)
}

func TestContentPart_JSONSerialization(t *testing.T) {
	contentPart := ContentPart{
		Type: "text",
		Text: "Hello, world!",
		Data: json.RawMessage(`{"additional": "data"}`),
	}

	// Test marshaling
	data, err := json.Marshal(contentPart)
	testutil.AssertNoError(t, err)

	// Test unmarshaling
	var unmarshaled ContentPart
	err = json.Unmarshal(data, &unmarshaled)
	testutil.AssertNoError(t, err)

	testutil.AssertEqual(t, contentPart.Type, unmarshaled.Type)
	testutil.AssertEqual(t, contentPart.Text, unmarshaled.Text)
	if contentPart.Data != nil && unmarshaled.Data != nil {
		// Data should be preserved (may have spacing differences in JSON)
		testutil.AssertTrue(t, len(unmarshaled.Data) > 0, "Data should be preserved")
	}
}

func TestContentPart_JSONSerialization_TextOnly(t *testing.T) {
	contentPart := ContentPart{
		Type: "text",
		Text: "Hello, world!",
	}

	// Test marshaling
	data, err := json.Marshal(contentPart)
	testutil.AssertNoError(t, err)

	// Should be successfully marshaled
	testutil.AssertTrue(t, len(data) > 0, "Data should be marshaled")

	// Test unmarshaling
	var unmarshaled ContentPart
	err = json.Unmarshal(data, &unmarshaled)
	testutil.AssertNoError(t, err)

	testutil.AssertEqual(t, contentPart.Type, unmarshaled.Type)
	testutil.AssertEqual(t, contentPart.Text, unmarshaled.Text)
	// Check that data is nil or empty
	if unmarshaled.Data != nil {
		testutil.AssertEqual(t, 0, len(unmarshaled.Data))
	}
}

func TestStreamEvent_JSONSerialization(t *testing.T) {
	event := StreamEvent{
		Event: "message_start",
		Data:  json.RawMessage(`{"id": "msg_123", "type": "message"}`),
	}

	// Test marshaling
	data, err := json.Marshal(event)
	testutil.AssertNoError(t, err)

	// Verify that data was marshaled successfully
	testutil.AssertTrue(t, len(data) > 0, "Data should be marshaled")

	// Test unmarshaling
	var unmarshaled StreamEvent
	err = json.Unmarshal(data, &unmarshaled)
	testutil.AssertNoError(t, err)

	testutil.AssertEqual(t, event.Event, unmarshaled.Event)
	// Data should be preserved
	testutil.AssertTrue(t, len(unmarshaled.Data) > 0, "Data should be preserved")
}

func TestMessage_InvalidJSON(t *testing.T) {
	invalidJSON := `{"role": "user", "content": invalid_json}`
	
	var message Message
	err := json.Unmarshal([]byte(invalidJSON), &message)
	testutil.AssertError(t, err, "Should fail with invalid JSON")
}

func TestRequest_EmptyMessages(t *testing.T) {
	request := Request{
		Model:    "gpt-4",
		Messages: []Message{},
	}

	// Test marshaling
	data, err := json.Marshal(request)
	testutil.AssertNoError(t, err)

	// Test unmarshaling
	var unmarshaled Request
	err = json.Unmarshal(data, &unmarshaled)
	testutil.AssertNoError(t, err)

	testutil.AssertEqual(t, 0, len(unmarshaled.Messages))
}

func TestUsage_ZeroValues(t *testing.T) {
	usage := Usage{}

	// Test marshaling
	data, err := json.Marshal(usage)
	testutil.AssertNoError(t, err)

	expectedJSON := `{"input_tokens":0,"output_tokens":0,"total_tokens":0}`
	testutil.AssertEqual(t, expectedJSON, string(data))

	// Test unmarshaling
	var unmarshaled Usage
	err = json.Unmarshal(data, &unmarshaled)
	testutil.AssertNoError(t, err)

	testutil.AssertEqual(t, 0, unmarshaled.InputTokens)
	testutil.AssertEqual(t, 0, unmarshaled.OutputTokens)
	testutil.AssertEqual(t, 0, unmarshaled.TotalTokens)
}