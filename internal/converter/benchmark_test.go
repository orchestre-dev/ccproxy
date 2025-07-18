package converter

import (
	"strings"
	"testing"

	"ccproxy/internal/models"
)

// BenchmarkConvertAnthropicToOpenAI benchmarks conversion from Anthropic to OpenAI format
func BenchmarkConvertAnthropicToOpenAI(b *testing.B) {

	b.Run("Simple_Message", func(b *testing.B) {
		maxTokens := 100
		req := &models.MessagesRequest{
			Model: "claude-3-opus",
			Messages: []models.Message{
				{Role: "user", Content: "Hello, how are you?"},
			},
			MaxTokens: &maxTokens,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := ConvertAnthropicToOpenAI(req)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Multi_Turn_Conversation", func(b *testing.B) {
		maxTokens := 200
		req := &models.MessagesRequest{
			Model: "claude-3-opus",
			Messages: []models.Message{
				{Role: "user", Content: "What's the weather like?"},
				{Role: "assistant", Content: "I don't have access to real-time weather data."},
				{Role: "user", Content: "Can you tell me about climate patterns?"},
				{Role: "assistant", Content: "Climate patterns are long-term weather trends..."},
				{Role: "user", Content: "Thanks for the explanation!"},
			},
			MaxTokens: &maxTokens,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := ConvertAnthropicToOpenAI(req)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("With_System_Prompt", func(b *testing.B) {
		maxTokens := 500
		req := &models.MessagesRequest{
			Model: "claude-3-opus",
			// System prompts are typically added as first message
			Messages: []models.Message{
				{Role: "system", Content: "You are a helpful AI assistant focused on technical topics."},
				{Role: "user", Content: "Explain Docker containers"},
			},
			MaxTokens: &maxTokens,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := ConvertAnthropicToOpenAI(req)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Large_Message", func(b *testing.B) {
		// Create a large message (~10KB)
		largeContent := strings.Repeat("This is a test message with some content. ", 250)
		maxTokens := 1000
		req := &models.MessagesRequest{
			Model: "claude-3-opus",
			Messages: []models.Message{
				{Role: "user", Content: largeContent},
			},
			MaxTokens: &maxTokens,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := ConvertAnthropicToOpenAI(req)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("With_Tools", func(b *testing.B) {
		maxTokens := 200
		req := &models.MessagesRequest{
			Model: "claude-3-opus",
			Messages: []models.Message{
				{Role: "user", Content: "What's the weather in San Francisco?"},
			},
			Tools: []models.Tool{
				{
					Name:        "get_weather",
					Description: stringPtr("Get the current weather for a location"),
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"location": map[string]interface{}{
								"type":        "string",
								"description": "The city and state",
							},
						},
						"required": []string{"location"},
					},
				},
			},
			MaxTokens: &maxTokens,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := ConvertAnthropicToOpenAI(req)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkConvertOpenAIToAnthropic benchmarks conversion from OpenAI to Anthropic format
func BenchmarkConvertOpenAIToAnthropic(b *testing.B) {

	b.Run("Simple_Response", func(b *testing.B) {
		resp := &models.ChatCompletionResponse{
			ID:      "chatcmpl-123",
			Object:  "chat.completion",
			Created: 1677652288,
			Model:   "gpt-4",
			Choices: []models.ChatCompletionChoice{
				{
					Index: 0,
					Message: models.ChatMessage{
						Role:    "assistant",
						Content: "Hello! I'm doing well, thank you for asking.",
					},
					FinishReason: "stop",
				},
			},
			Usage: models.ChatCompletionUsage{
				PromptTokens:     10,
				CompletionTokens: 12,
				TotalTokens:      22,
			},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = ConvertOpenAIToAnthropic(resp, "bench-123", "benchmark")
		}
	})

	b.Run("With_Tool_Calls", func(b *testing.B) {
		resp := &models.ChatCompletionResponse{
			ID:      "chatcmpl-456",
			Object:  "chat.completion",
			Created: 1677652288,
			Model:   "gpt-4",
			Choices: []models.ChatCompletionChoice{
				{
					Index: 0,
					Message: models.ChatMessage{
						Role:    "assistant",
						Content: "",
						ToolCalls: []models.ToolCall{
							{
								ID:   "call_abc123",
								Type: "function",
								Function: models.FunctionCall{
									Name:      "get_weather",
									Arguments: `{"location": "San Francisco, CA"}`,
								},
							},
						},
					},
					FinishReason: "tool_calls",
				},
			},
			Usage: models.ChatCompletionUsage{
				PromptTokens:     25,
				CompletionTokens: 15,
				TotalTokens:      40,
			},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = ConvertOpenAIToAnthropic(resp, "bench-123", "benchmark")
		}
	})

	b.Run("Large_Response", func(b *testing.B) {
		// Create a large response (~10KB)
		largeContent := strings.Repeat("This is a test response with detailed content. ", 250)
		resp := &models.ChatCompletionResponse{
			ID:      "chatcmpl-789",
			Object:  "chat.completion",
			Created: 1677652288,
			Model:   "gpt-4",
			Choices: []models.ChatCompletionChoice{
				{
					Index: 0,
					Message: models.ChatMessage{
						Role:    "assistant",
						Content: largeContent,
					},
					FinishReason: "stop",
				},
			},
			Usage: models.ChatCompletionUsage{
				PromptTokens:     50,
				CompletionTokens: 2000,
				TotalTokens:      2050,
			},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = ConvertOpenAIToAnthropic(resp, "bench-123", "benchmark")
		}
	})
}

// BenchmarkComplexToolConversion benchmarks complex tool call conversions
func BenchmarkComplexToolConversion(b *testing.B) {

	// Create a complex request with multiple tools
	maxTokens := 500
	req := &models.MessagesRequest{
		Model: "claude-3-opus",
		Messages: []models.Message{
			{
				Role: "user",
				Content: []models.Content{
					{Type: "text", Text: "Analyze this data and create a visualization"},
				},
			},
			{
				Role: "assistant",
				Content: []models.Content{
					{Type: "text", Text: "I'll analyze the data and create a visualization for you."},
					{
						Type: "tool_use",
						ID:   "tool1",
						Name: "analyze_data",
						Input: map[string]interface{}{
							"data": []int{1, 2, 3, 4, 5},
							"type": "time_series",
						},
					},
				},
			},
			{
				Role: "user",
				Content: []models.Content{
					{
						Type:      "tool_result",
						ToolUseID: "tool1",
						Content:   "Analysis complete: trend is increasing",
					},
				},
			},
		},
		Tools: []models.Tool{
			{
				Name:        "analyze_data",
				Description: stringPtr("Analyze numerical data"),
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"data": map[string]interface{}{"type": "array"},
						"type": map[string]interface{}{"type": "string"},
					},
				},
			},
			{
				Name:        "create_chart",
				Description: stringPtr("Create a chart visualization"),
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"data":       map[string]interface{}{"type": "array"},
						"chart_type": map[string]interface{}{"type": "string"},
					},
				},
			},
		},
		MaxTokens: &maxTokens,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ConvertAnthropicToOpenAI(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMemoryEfficiency measures memory allocations during conversion
func BenchmarkMemoryEfficiency(b *testing.B) {

	maxTokens := 100
	req := &models.MessagesRequest{
		Model: "claude-3-opus",
		Messages: []models.Message{
			{Role: "user", Content: "Test message"},
			{Role: "assistant", Content: "Test response"},
		},
		MaxTokens: &maxTokens,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ConvertAnthropicToOpenAI(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}