package transformer

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
)

func TestBaseTransformer_Complete(t *testing.T) {
	transformer := NewBaseTransformer("test-transformer", "/v1/chat")
	
	t.Run("BasicProperties", func(t *testing.T) {
		testutil.AssertEqual(t, "test-transformer", transformer.GetName())
		testutil.AssertEqual(t, "/v1/chat", transformer.GetEndpoint())
	})
	
	t.Run("TransformRequestIn_PassThrough", func(t *testing.T) {
		ctx := context.Background()
		request := map[string]interface{}{
			"model": "test-model",
			"messages": []interface{}{
				map[string]interface{}{"role": "user", "content": "Hello"},
			},
		}
		
		result, err := transformer.TransformRequestIn(ctx, request, "test-provider")
		testutil.AssertNoError(t, err)
		
		resultMap, ok := result.(map[string]interface{})
		testutil.AssertTrue(t, ok)
		testutil.AssertEqual(t, "test-model", resultMap["model"])
	})
	
	t.Run("TransformRequestOut_PassThrough", func(t *testing.T) {
		ctx := context.Background()
		request := map[string]interface{}{
			"provider_specific": "value",
		}
		
		result, err := transformer.TransformRequestOut(ctx, request)
		testutil.AssertNoError(t, err)
		
		resultMap, ok := result.(map[string]interface{})
		testutil.AssertTrue(t, ok)
		testutil.AssertEqual(t, "value", resultMap["provider_specific"])
	})
	
	t.Run("TransformResponseIn_PassThrough", func(t *testing.T) {
		ctx := context.Background()
		originalBody := `{"response": "test"}`
		
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(originalBody)),
		}
		
		result, err := transformer.TransformResponseIn(ctx, resp)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 200, result.StatusCode)
		testutil.AssertEqual(t, "application/json", result.Header.Get("Content-Type"))
		
		// Read body to verify it's preserved
		body, readErr := io.ReadAll(result.Body)
		testutil.AssertNoError(t, readErr)
		testutil.AssertEqual(t, originalBody, string(body))
	})
	
	t.Run("TransformResponseOut_PassThrough", func(t *testing.T) {
		ctx := context.Background()
		originalBody := `{"output": "transformed"}`
		
		resp := &http.Response{
			StatusCode: 201,
			Header:     http.Header{"X-Custom": []string{"header"}},
			Body:       io.NopCloser(strings.NewReader(originalBody)),
		}
		
		result, err := transformer.TransformResponseOut(ctx, resp)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 201, result.StatusCode)
		testutil.AssertEqual(t, "header", result.Header.Get("X-Custom"))
		
		// Read body to verify it's preserved  
		body, readErr := io.ReadAll(result.Body)
		testutil.AssertNoError(t, readErr)
		testutil.AssertEqual(t, originalBody, string(body))
	})
}

func TestSSEEvent_Complete(t *testing.T) {
	t.Run("FullEvent", func(t *testing.T) {
		event := &SSEEvent{
			Event: "message_start",
			Data:  `{"type": "message", "id": "msg_123"}`,
			ID:    "event_456",
			Retry: 3000,
		}
		
		testutil.AssertEqual(t, "message_start", event.Event)
		testutil.AssertEqual(t, `{"type": "message", "id": "msg_123"}`, event.Data)
		testutil.AssertEqual(t, "event_456", event.ID)
		testutil.AssertEqual(t, 3000, event.Retry)
	})
	
	t.Run("MinimalEvent", func(t *testing.T) {
		event := &SSEEvent{
			Data: "simple data",
		}
		
		testutil.AssertEqual(t, "", event.Event)
		testutil.AssertEqual(t, "simple data", event.Data)
		testutil.AssertEqual(t, "", event.ID)
		testutil.AssertEqual(t, 0, event.Retry)
	})
	
	t.Run("EventWithJSONData", func(t *testing.T) {
		jsonData := `{"content": "Hello", "role": "assistant"}`
		event := &SSEEvent{
			Event: "content_block_delta",
			Data:  jsonData,
		}
		
		testutil.AssertEqual(t, "content_block_delta", event.Event)
		testutil.AssertContains(t, event.Data, "Hello")
		testutil.AssertContains(t, event.Data, "assistant")
	})
}

func TestRequestConfig_Complete(t *testing.T) {
	t.Run("FullRequestConfig", func(t *testing.T) {
		body := map[string]interface{}{
			"model":    "claude-3-haiku",
			"messages": []interface{}{map[string]interface{}{"role": "user", "content": "Hi"}},
		}
		
		config := &RequestConfig{
			Body: body,
			URL:  "https://api.anthropic.com/v1/messages",
			Headers: map[string]string{
				"Authorization": "Bearer sk-test",
				"Content-Type":  "application/json",
			},
			Method:  "POST",
			Timeout: 30,
		}
		
		testutil.AssertEqual(t, "https://api.anthropic.com/v1/messages", config.URL)
		testutil.AssertEqual(t, "POST", config.Method)
		testutil.AssertEqual(t, 30, config.Timeout)
		testutil.AssertEqual(t, "Bearer sk-test", config.Headers["Authorization"])
		testutil.AssertEqual(t, "application/json", config.Headers["Content-Type"])
		
		bodyMap, ok := config.Body.(map[string]interface{})
		testutil.AssertTrue(t, ok)
		testutil.AssertEqual(t, "claude-3-haiku", bodyMap["model"])
	})
	
	t.Run("MinimalRequestConfig", func(t *testing.T) {
		config := &RequestConfig{
			Body: "simple body",
		}
		
		testutil.AssertEqual(t, "simple body", config.Body)
		testutil.AssertEqual(t, "", config.URL)
		testutil.AssertEqual(t, "", config.Method)
		testutil.AssertEqual(t, 0, config.Timeout)
		testutil.AssertEqual(t, true, config.Headers == nil)
	})
	
	t.Run("RequestConfigWithCustomHeaders", func(t *testing.T) {
		headers := map[string]string{
			"X-Custom-Header": "custom-value",
			"User-Agent":      "ccproxy/1.0",
		}
		
		config := &RequestConfig{
			Body:    map[string]string{"key": "value"},
			Headers: headers,
		}
		
		testutil.AssertEqual(t, "custom-value", config.Headers["X-Custom-Header"])
		testutil.AssertEqual(t, "ccproxy/1.0", config.Headers["User-Agent"])
		testutil.AssertEqual(t, 2, len(config.Headers))
	})
}

func TestTransformerInterface_Implementation(t *testing.T) {
	// Test that BaseTransformer implements the Transformer interface
	var transformer Transformer = NewBaseTransformer("test", "/endpoint")
	
	testutil.AssertEqual(t, "test", transformer.GetName())
	testutil.AssertEqual(t, "/endpoint", transformer.GetEndpoint())
	
	ctx := context.Background()
	
	// Test that all interface methods can be called
	_, err := transformer.TransformRequestIn(ctx, map[string]string{"test": "value"}, "provider")
	testutil.AssertNoError(t, err)
	
	_, err = transformer.TransformRequestOut(ctx, map[string]string{"test": "value"})
	testutil.AssertNoError(t, err)
	
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("{}")),
	}
	_, err = transformer.TransformResponseIn(ctx, resp)
	testutil.AssertNoError(t, err)
	
	resp.Body = io.NopCloser(strings.NewReader("{}"))
	_, err = transformer.TransformResponseOut(ctx, resp)
	testutil.AssertNoError(t, err)
}

func TestTransformerError_Handling(t *testing.T) {
	transformer := NewBaseTransformer("error-test", "/test")
	ctx := context.Background()
	
	t.Run("NilRequest", func(t *testing.T) {
		result, err := transformer.TransformRequestIn(ctx, nil, "provider")
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, true, result == nil)
	})
	
	t.Run("NilResponse", func(t *testing.T) {
		result, err := transformer.TransformResponseIn(ctx, nil)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, true, result == nil)
	})
}

func TestSSEEvent_EdgeCases(t *testing.T) {
	t.Run("EmptyData", func(t *testing.T) {
		event := &SSEEvent{
			Event: "empty",
			Data:  "",
		}
		
		testutil.AssertEqual(t, "empty", event.Event)
		testutil.AssertEqual(t, "", event.Data)
	})
	
	t.Run("LargeRetryValue", func(t *testing.T) {
		event := &SSEEvent{
			Data:  "test",
			Retry: 999999,
		}
		
		testutil.AssertEqual(t, 999999, event.Retry)
	})
	
	t.Run("SpecialCharactersInData", func(t *testing.T) {
		specialData := "data with\nnewlines\tand\ttabs"
		event := &SSEEvent{
			Data: specialData,
		}
		
		testutil.AssertEqual(t, specialData, event.Data)
	})
}

func TestRequestConfig_EdgeCases(t *testing.T) {
	t.Run("EmptyHeaders", func(t *testing.T) {
		config := &RequestConfig{
			Body:    "test",
			Headers: map[string]string{},
		}
		
		testutil.AssertEqual(t, 0, len(config.Headers))
	})
	
	t.Run("NilBody", func(t *testing.T) {
		config := &RequestConfig{
			Body: nil,
			URL:  "https://api.example.com",
		}
		
		testutil.AssertEqual(t, true, config.Body == nil)
		testutil.AssertEqual(t, "https://api.example.com", config.URL)
	})
	
	t.Run("ComplexBodyStructure", func(t *testing.T) {
		complexBody := map[string]interface{}{
			"nested": map[string]interface{}{
				"array": []interface{}{1, 2, 3},
				"bool":  true,
			},
			"number": 42.5,
		}
		
		config := &RequestConfig{Body: complexBody}
		
		bodyMap := config.Body.(map[string]interface{})
		nested := bodyMap["nested"].(map[string]interface{})
		testutil.AssertEqual(t, true, nested["bool"])
		testutil.AssertEqual(t, 42.5, bodyMap["number"])
	})
}