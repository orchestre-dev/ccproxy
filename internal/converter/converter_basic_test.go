package converter

import (
	"encoding/json"
	"testing"

	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
)

func TestNewMessageConverter(t *testing.T) {
	converter := NewMessageConverter()

	testutil.AssertNotEqual(t, nil, converter)
	testutil.AssertNotEqual(t, nil, converter.converters)
	testutil.AssertEqual(t, 4, len(converter.converters))

	// Check that all expected converters are present
	expectedFormats := []string{
		string(FormatAnthropic),
		string(FormatOpenAI),
		string(FormatGoogle),
		string(FormatAWS),
	}

	for _, format := range expectedFormats {
		testutil.AssertNotEqual(t, nil, converter.converters[format])
	}
}

func TestMessageConverter_SameFormat(t *testing.T) {
	converter := NewMessageConverter()

	// Test that same format conversion returns the same data
	requestData := json.RawMessage(`{"test": "data"}`)

	result, err := converter.ConvertRequest(requestData, FormatAnthropic, FormatAnthropic)
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, string(requestData), string(result))

	responseData := json.RawMessage(`{"test": "response"}`)
	result, err = converter.ConvertResponse(responseData, FormatOpenAI, FormatOpenAI)
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, string(responseData), string(result))

	eventData := []byte(`data: {"test": "event"}`)
	resultBytes, err := converter.ConvertStreamEvent(eventData, FormatGoogle, FormatGoogle)
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, string(eventData), string(resultBytes))
}

func TestMessageConverter_UnsupportedFormat(t *testing.T) {
	converter := NewMessageConverter()
	requestData := json.RawMessage(`{"test": "data"}`)

	// Test unsupported source format
	_, err := converter.ConvertRequest(requestData, MessageFormat("unsupported"), FormatOpenAI)
	testutil.AssertError(t, err)
	testutil.AssertContains(t, err.Error(), "unsupported source format")

	// Test unsupported response format
	_, err = converter.ConvertResponse(requestData, MessageFormat("invalid"), FormatAnthropic)
	testutil.AssertError(t, err)
	testutil.AssertContains(t, err.Error(), "unsupported source format")

	// Test unsupported stream event format
	eventData := []byte(`data: {"test": "event"}`)
	_, err = converter.ConvertStreamEvent(eventData, MessageFormat("bad"), FormatGoogle)
	testutil.AssertError(t, err)
	testutil.AssertContains(t, err.Error(), "unsupported source format")
}

func TestMessageFormatConstants(t *testing.T) {
	tests := []struct {
		name     string
		format   MessageFormat
		expected string
	}{
		{"Anthropic", FormatAnthropic, "anthropic"},
		{"OpenAI", FormatOpenAI, "openai"},
		{"Google", FormatGoogle, "google"},
		{"AWS", FormatAWS, "aws"},
		{"Generic", FormatGeneric, "generic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.AssertEqual(t, tt.expected, string(tt.format))
		})
	}
}

// Test basic struct initialization
func TestBasicStructs(t *testing.T) {
	t.Run("Message", func(t *testing.T) {
		msg := Message{
			Role:    "user",
			Content: json.RawMessage(`"Hello"`),
			Name:    "test",
		}
		testutil.AssertEqual(t, "user", msg.Role)
		testutil.AssertEqual(t, "test", msg.Name)
	})

	t.Run("Request", func(t *testing.T) {
		req := Request{
			Model:     "test-model",
			Messages:  []Message{},
			MaxTokens: 100,
		}
		testutil.AssertEqual(t, "test-model", req.Model)
		testutil.AssertEqual(t, 100, req.MaxTokens)
	})

	t.Run("Response", func(t *testing.T) {
		resp := Response{
			ID:   "test-id",
			Type: "message",
			Role: "assistant",
		}
		testutil.AssertEqual(t, "test-id", resp.ID)
		testutil.AssertEqual(t, "message", resp.Type)
		testutil.AssertEqual(t, "assistant", resp.Role)
	})

	t.Run("Usage", func(t *testing.T) {
		usage := Usage{
			InputTokens:  10,
			OutputTokens: 5,
			TotalTokens:  15,
		}
		testutil.AssertEqual(t, 10, usage.InputTokens)
		testutil.AssertEqual(t, 5, usage.OutputTokens)
		testutil.AssertEqual(t, 15, usage.TotalTokens)
	})

	t.Run("ContentPart", func(t *testing.T) {
		part := ContentPart{
			Type: "text",
			Text: "Hello world",
		}
		testutil.AssertEqual(t, "text", part.Type)
		testutil.AssertEqual(t, "Hello world", part.Text)
	})

	t.Run("StreamEvent", func(t *testing.T) {
		event := StreamEvent{
			Event: "message_start",
			Data:  json.RawMessage(`{"id": "msg_123"}`),
		}
		testutil.AssertEqual(t, "message_start", event.Event)
	})
}
