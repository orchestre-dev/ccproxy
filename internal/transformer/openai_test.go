package transformer

import (
	"context"
	"testing"
)

func TestNewOpenAITransformer(t *testing.T) {
	transformer := NewOpenAITransformer()

	if transformer == nil {
		t.Error("Expected non-nil OpenAI transformer")
	}

	if transformer.GetName() != "openai" {
		t.Errorf("Expected name 'openai', got %s", transformer.GetName())
	}

	if transformer.GetEndpoint() != "/v1/chat/completions" {
		t.Errorf("Expected endpoint '/v1/chat/completions', got %s", transformer.GetEndpoint())
	}
}

func TestOpenAITransformer_DefaultBehavior(t *testing.T) {
	transformer := NewOpenAITransformer()
	ctx := context.Background()

	t.Run("TransformRequestIn_PassThrough", func(t *testing.T) {
		request := map[string]interface{}{
			"model": "gpt-4",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Hello, world!",
				},
			},
			"max_tokens":  1000,
			"temperature": 0.7,
		}

		result, err := transformer.TransformRequestIn(ctx, request, "openai")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("Result is not a map")
		}

		// Should pass through unchanged
		if resultMap["model"] != "gpt-4" {
			t.Errorf("Expected model 'gpt-4', got %v", resultMap["model"])
		}

		if resultMap["max_tokens"] != 1000 {
			t.Errorf("Expected max_tokens 1000, got %v", resultMap["max_tokens"])
		}

		if resultMap["temperature"] != 0.7 {
			t.Errorf("Expected temperature 0.7, got %v", resultMap["temperature"])
		}
	})

	t.Run("TransformRequestOut_PassThrough", func(t *testing.T) {
		request := map[string]interface{}{
			"model":  "gpt-4",
			"stream": true,
		}

		result, err := transformer.TransformRequestOut(ctx, request)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("Result is not a map")
		}

		// Should pass through unchanged
		if resultMap["model"] != "gpt-4" {
			t.Errorf("Expected model 'gpt-4', got %v", resultMap["model"])
		}

		if resultMap["stream"] != true {
			t.Errorf("Expected stream true, got %v", resultMap["stream"])
		}
	})

	t.Run("NilInput", func(t *testing.T) {
		result, err := transformer.TransformRequestIn(ctx, nil, "openai")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result != nil {
			t.Error("Expected nil result for nil input")
		}
	})

	t.Run("StringInput", func(t *testing.T) {
		input := "test string"
		result, err := transformer.TransformRequestIn(ctx, input, "openai")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result != input {
			t.Error("Expected string input to pass through unchanged")
		}
	})
}

func TestOpenAITransformer_Integration(t *testing.T) {
	t.Run("ChainWithOtherTransformers", func(t *testing.T) {
		// Test that OpenAI transformer works well in chains
		openaiTransformer := NewOpenAITransformer()
		maxTokenTransformer := NewMaxTokenTransformer()

		chain := NewTransformerChain(openaiTransformer, maxTokenTransformer)

		request := map[string]interface{}{
			"model": "gpt-4",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Test message",
				},
			},
			// No max_tokens specified - should be added by MaxToken transformer
		}

		result, err := chain.TransformRequestIn(context.Background(), request, "openai")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("Result is not a map")
		}

		// MaxToken transformer should have added max_tokens
		if _, exists := resultMap["max_tokens"]; !exists {
			t.Error("Expected max_tokens to be added by MaxToken transformer")
		}

		// Original fields should be preserved
		if resultMap["model"] != "gpt-4" {
			t.Errorf("Expected model 'gpt-4', got %v", resultMap["model"])
		}
	})
}
