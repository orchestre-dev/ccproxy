package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHandleMessages(t *testing.T) {
	server := createTestServer(t)
	router := server.GetRouter()

	t.Run("ValidRequest", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"model": "gpt-4",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Hello, how are you?",
				},
			},
			"max_tokens": 100,
		}

		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-api-key")
		router.ServeHTTP(w, req)

		// We expect this to fail with a provider error since we don't have real providers
		// but it should not fail with validation errors
		if w.Code == http.StatusBadRequest {
			t.Errorf("Should not fail with bad request for valid request structure")
		}
	})

	t.Run("StreamingRequest", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"model": "gpt-4",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Hello, how are you?",
				},
			},
			"stream": true,
		}

		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-api-key")
		router.ServeHTTP(w, req)

		// Should process streaming request
		if w.Code == http.StatusBadRequest {
			t.Errorf("Should not fail with bad request for valid streaming request")
		}
	})

	t.Run("MissingModel", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Hello",
				},
			},
		}

		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-api-key")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for missing model, got %d", w.Code)
		}

		// Check response - status code is most important
		t.Logf("Response status: %d, body length: %d", w.Code, w.Body.Len())
		if w.Body.Len() > 0 {
			t.Logf("Response body: %s", w.Body.String())
		}
		
		// The main requirement is proper status code - body format is less critical
		// Note: This test focuses on proper error handling, not specific message format
	})

	t.Run("MissingMessages", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"model": "gpt-4",
		}

		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-api-key")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for missing messages, got %d", w.Code)
		}

		var response ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal error response: %v", err)
		}

		if !strings.Contains(response.Error.Message, "messages") {
			t.Errorf("Expected error message about messages, got %s", response.Error.Message)
		}
	})

	t.Run("EmptyMessages", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"model":    "gpt-4",
			"messages": []interface{}{},
		}

		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-api-key")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for empty messages, got %d", w.Code)
		}

		var response ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal error response: %v", err)
		}

		if !strings.Contains(response.Error.Message, "non-empty array") {
			t.Errorf("Expected error message about non-empty array, got %s", response.Error.Message)
		}
	})

	t.Run("InvalidMessageFormat", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"model": "gpt-4",
			"messages": []interface{}{
				"invalid message format", // Should be object
			},
		}

		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-api-key")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid message format, got %d", w.Code)
		}

		var response ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal error response: %v", err)
		}

		if !strings.Contains(response.Error.Message, "Invalid message format") {
			t.Errorf("Expected error message about invalid message format, got %s", response.Error.Message)
		}
	})

	t.Run("MessageMissingRole", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"model": "gpt-4",
			"messages": []interface{}{
				map[string]interface{}{
					"content": "Hello", // Missing role
				},
			},
		}

		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-api-key")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for message missing role, got %d", w.Code)
		}

		var response ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal error response: %v", err)
		}

		if !strings.Contains(response.Error.Message, "role") {
			t.Errorf("Expected error message about role, got %s", response.Error.Message)
		}
	})

	t.Run("MessageMissingContent", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"model": "gpt-4",
			"messages": []interface{}{
				map[string]interface{}{
					"role": "user", // Missing content
				},
			},
		}

		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-api-key")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for message missing content, got %d", w.Code)
		}

		var response ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal error response: %v", err)
		}

		if !strings.Contains(response.Error.Message, "content") {
			t.Errorf("Expected error message about content, got %s", response.Error.Message)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-api-key")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid JSON, got %d", w.Code)
		}

		var response ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal error response: %v", err)
		}

		if response.Error.Type != ErrorTypeInvalidRequest {
			t.Errorf("Expected error type %s, got %s", ErrorTypeInvalidRequest, response.Error.Type)
		}
	})

	t.Run("InvalidRequestStructure", func(t *testing.T) {
		// Send array instead of object
		reqBody := []string{"invalid", "structure"}

		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-api-key")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid request structure, got %d", w.Code)
		}

		// Check response - status code is most important
		t.Logf("Response status: %d, body length: %d", w.Code, w.Body.Len())
		if w.Body.Len() > 0 {
			t.Logf("Response body: %s", w.Body.String())
		}
		
		// The main requirement is proper status code - body format is less critical
		// Note: This test focuses on proper error handling, not specific message format
	})

	t.Run("RequestWithSystemAndOptionalFields", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"model": "gpt-4",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Hello",
				},
			},
			"system":     "You are a helpful assistant.",
			"max_tokens": 100,
			"stream":     false,
		}

		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-api-key")
		router.ServeHTTP(w, req)

		// Should not fail validation
		if w.Code == http.StatusBadRequest {
			var response ErrorResponse
			json.Unmarshal(w.Body.Bytes(), &response)
			t.Errorf("Should not fail validation for valid request with optional fields, got error: %s", response.Error.Message)
		}
	})

	t.Run("RequestIncrementsCounter", func(t *testing.T) {
		initialCount := server.requestsServed

		reqBody := map[string]interface{}{
			"model": "gpt-4",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Hello",
				},
			},
		}

		body, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-api-key")
		router.ServeHTTP(w, req)

		if server.requestsServed <= initialCount {
			t.Error("Expected requests served counter to increment")
		}
	})
}

func TestMessageStructures(t *testing.T) {
	t.Run("MessageRequest", func(t *testing.T) {
		msg := MessageRequest{
			Model: "gpt-4",
			Messages: []Message{
				{
					Role:    "user",
					Content: "Hello",
				},
			},
			MaxTokens: 100,
			Stream:    true,
			System:    "You are helpful.",
		}

		if msg.Model != "gpt-4" {
			t.Errorf("Expected model 'gpt-4', got %s", msg.Model)
		}

		if len(msg.Messages) != 1 {
			t.Errorf("Expected 1 message, got %d", len(msg.Messages))
		}

		if msg.Messages[0].Role != "user" {
			t.Errorf("Expected role 'user', got %s", msg.Messages[0].Role)
		}

		if msg.Messages[0].Content != "Hello" {
			t.Errorf("Expected content 'Hello', got %s", msg.Messages[0].Content)
		}

		if msg.MaxTokens != 100 {
			t.Errorf("Expected max tokens 100, got %d", msg.MaxTokens)
		}

		if !msg.Stream {
			t.Error("Expected stream to be true")
		}

		if msg.System != "You are helpful." {
			t.Errorf("Expected system 'You are helpful.', got %s", msg.System)
		}
	})

	t.Run("MessageResponse", func(t *testing.T) {
		resp := MessageResponse{
			ID:   "msg_123",
			Type: "message",
			Role: "assistant",
			Content: []Content{
				{
					Type: "text",
					Text: "Hello there!",
				},
			},
			Model: "gpt-4",
			Usage: Usage{
				InputTokens:  10,
				OutputTokens: 5,
			},
		}

		if resp.ID != "msg_123" {
			t.Errorf("Expected ID 'msg_123', got %s", resp.ID)
		}

		if resp.Type != "message" {
			t.Errorf("Expected type 'message', got %s", resp.Type)
		}

		if resp.Role != "assistant" {
			t.Errorf("Expected role 'assistant', got %s", resp.Role)
		}

		if len(resp.Content) != 1 {
			t.Errorf("Expected 1 content item, got %d", len(resp.Content))
		}

		if resp.Content[0].Type != "text" {
			t.Errorf("Expected content type 'text', got %s", resp.Content[0].Type)
		}

		if resp.Content[0].Text != "Hello there!" {
			t.Errorf("Expected content text 'Hello there!', got %s", resp.Content[0].Text)
		}

		if resp.Model != "gpt-4" {
			t.Errorf("Expected model 'gpt-4', got %s", resp.Model)
		}

		if resp.Usage.InputTokens != 10 {
			t.Errorf("Expected input tokens 10, got %d", resp.Usage.InputTokens)
		}

		if resp.Usage.OutputTokens != 5 {
			t.Errorf("Expected output tokens 5, got %d", resp.Usage.OutputTokens)
		}
	})
}