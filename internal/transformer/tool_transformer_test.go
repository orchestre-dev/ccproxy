package transformer

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/orchestre-dev/ccproxy/internal/config"
	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
	"github.com/orchestre-dev/ccproxy/internal/tools"
)

func TestToolTransformer(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	t.Run("NewToolTransformer", func(t *testing.T) {
		transformer := NewToolTransformer()
		testutil.AssertEqual(t, "tool", transformer.GetName())
		testutil.AssertEqual(t, "", transformer.GetEndpoint())
		testutil.AssertNotEqual(t, nil, transformer.handler)
	})
}

func TestToolTransformRequest(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewToolTransformer()
	ctx := context.Background()

	t.Run("ValidToolDefinition", func(t *testing.T) {
		provider := &config.Provider{Name: "anthropic"}
		request := map[string]interface{}{
			"model":    "claude-3-sonnet",
			"messages": []interface{}{},
			"tools": []interface{}{
				map[string]interface{}{
					"name":        "get_weather",
					"description": "Get weather information for a location",
					"input_schema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"location": map[string]interface{}{
								"type":        "string",
								"description": "City name",
							},
						},
						"required": []interface{}{"location"},
					},
				},
			},
		}

		result, err := transformer.TransformRequest(ctx, provider, request)
		testutil.AssertNoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		testutil.AssertEqual(t, true, ok)

		// Should preserve the tools for Anthropic (no transformation needed)
		tools := resultMap["tools"].([]interface{})
		testutil.AssertEqual(t, 1, len(tools))

		tool := tools[0].(map[string]interface{})
		testutil.AssertEqual(t, "get_weather", tool["name"])
		testutil.AssertEqual(t, "Get weather information for a location", tool["description"])
	})

	t.Run("OpenAIModernModelTransformation", func(t *testing.T) {
		provider := &config.Provider{Name: "openai"}
		request := map[string]interface{}{
			"model": "gpt-4", // Modern model
			"tools": []interface{}{
				map[string]interface{}{
					"name":        "calculate",
					"description": "Perform a calculation",
					"input_schema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"expression": map[string]interface{}{
								"type": "string",
							},
						},
					},
				},
			},
		}

		result, err := transformer.TransformRequest(ctx, provider, request)
		testutil.AssertNoError(t, err)

		resultMap := result.(map[string]interface{})
		tools := resultMap["tools"].([]interface{})
		testutil.AssertEqual(t, 1, len(tools))

		// Should be transformed to OpenAI function format
		tool := tools[0].(map[string]interface{})
		testutil.AssertEqual(t, "function", tool["type"])

		function := tool["function"].(map[string]interface{})
		testutil.AssertEqual(t, "calculate", function["name"])
		testutil.AssertEqual(t, "Perform a calculation", function["description"])
		testutil.AssertNotEqual(t, nil, function["parameters"])
	})

	t.Run("OpenAILegacyModelTransformation", func(t *testing.T) {
		provider := &config.Provider{Name: "openai"}
		request := map[string]interface{}{
			"model": "gpt-3.5-turbo-0613", // Legacy model
			"tools": []interface{}{
				map[string]interface{}{
					"name":        "search",
					"description": "Search for information",
					"input_schema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"query": map[string]interface{}{
								"type": "string",
							},
						},
					},
				},
			},
		}

		result, err := transformer.TransformRequest(ctx, provider, request)
		testutil.AssertNoError(t, err)

		resultMap := result.(map[string]interface{})

		// Should have functions instead of tools for legacy models
		_, hasTools := resultMap["tools"]
		testutil.AssertEqual(t, false, hasTools)

		functions, hasFunctions := resultMap["functions"]
		testutil.AssertEqual(t, true, hasFunctions)

		functionList := functions.([]interface{})
		testutil.AssertEqual(t, 1, len(functionList))

		function := functionList[0].(map[string]interface{})
		testutil.AssertEqual(t, "search", function["name"])
		testutil.AssertEqual(t, "Search for information", function["description"])
	})

	t.Run("ModelWithProviderPrefix", func(t *testing.T) {
		provider := &config.Provider{Name: "openai"}
		request := map[string]interface{}{
			"model": "openai,gpt-3.5-turbo-0613", // With provider prefix
			"tools": []interface{}{
				map[string]interface{}{
					"name":        "test_tool",
					"description": "A test tool",
					"input_schema": map[string]interface{}{
						"type": "object",
					},
				},
			},
		}

		result, err := transformer.TransformRequest(ctx, provider, request)
		testutil.AssertNoError(t, err)

		resultMap := result.(map[string]interface{})

		// Should recognize legacy model after stripping prefix
		_, hasTools := resultMap["tools"]
		testutil.AssertEqual(t, false, hasTools)

		_, hasFunctions := resultMap["functions"]
		testutil.AssertEqual(t, true, hasFunctions)
	})

	t.Run("InvalidToolDefinition", func(t *testing.T) {
		provider := &config.Provider{Name: "anthropic"}
		request := map[string]interface{}{
			"tools": []interface{}{
				map[string]interface{}{
					// Missing name and description
					"input_schema": map[string]interface{}{},
				},
			},
		}

		_, err := transformer.TransformRequest(ctx, provider, request)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "invalid tool definition")
	})

	t.Run("NoTools", func(t *testing.T) {
		provider := &config.Provider{Name: "anthropic"}
		request := map[string]interface{}{
			"model":    "claude-3",
			"messages": []interface{}{},
		}

		result, err := transformer.TransformRequest(ctx, provider, request)
		testutil.AssertNoError(t, err)

		// Should pass through unchanged when no tools
		resultMap := result.(map[string]interface{})
		testutil.AssertEqual(t, "claude-3", resultMap["model"])
		testutil.AssertEqual(t, request["model"], resultMap["model"])
	})

	t.Run("InvalidRequestFormat", func(t *testing.T) {
		provider := &config.Provider{Name: "anthropic"}
		request := "invalid request"

		result, err := transformer.TransformRequest(ctx, provider, request)
		testutil.AssertNoError(t, err)

		// Should pass through unchanged when not a map
		testutil.AssertEqual(t, "invalid request", result.(string))
	})

	t.Run("InvalidToolsArray", func(t *testing.T) {
		provider := &config.Provider{Name: "anthropic"}
		request := map[string]interface{}{
			"tools": "invalid tools array",
		}

		result, err := transformer.TransformRequest(ctx, provider, request)
		testutil.AssertNoError(t, err)

		// Should pass through unchanged when tools is not an array
		resultMap := result.(map[string]interface{})
		testutil.AssertEqual(t, "invalid tools array", resultMap["tools"])
	})
}

func TestToolTransformResponse(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewToolTransformer()
	ctx := context.Background()
	provider := &config.Provider{Name: "anthropic"}

	t.Run("ResponseWithToolUse", func(t *testing.T) {
		response := map[string]interface{}{
			"role": "assistant",
			"content": []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "I'll help you get the weather information.",
				},
				map[string]interface{}{
					"type": "tool_use",
					"id":   "tool_123",
					"name": "get_weather",
					"input": map[string]interface{}{
						"location": "New York",
					},
				},
			},
		}

		result, err := transformer.TransformResponse(ctx, provider, response)
		testutil.AssertNoError(t, err)

		// Should process the tool use blocks
		resultMap := result.(map[string]interface{})
		content := resultMap["content"].([]interface{})

		// Find the tool use block
		var toolUseBlock map[string]interface{}
		for _, block := range content {
			if blockMap, ok := block.(map[string]interface{}); ok {
				if blockMap["type"] == "tool_use" {
					toolUseBlock = blockMap
					break
				}
			}
		}

		testutil.AssertNotEqual(t, nil, toolUseBlock)
		testutil.AssertEqual(t, true, toolUseBlock["_processed"])
		testutil.AssertEqual(t, "ccproxy", toolUseBlock["_processor"])
	})

	t.Run("ResponseWithoutToolUse", func(t *testing.T) {
		response := map[string]interface{}{
			"role":    "assistant",
			"content": "This is a simple text response.",
		}

		result, err := transformer.TransformResponse(ctx, provider, response)
		testutil.AssertNoError(t, err)

		// Should pass through unchanged
		resultMap := result.(map[string]interface{})
		testutil.AssertEqual(t, "assistant", resultMap["role"])
		testutil.AssertEqual(t, "This is a simple text response.", resultMap["content"])
	})

	t.Run("InvalidResponseFormat", func(t *testing.T) {
		response := "invalid response"

		result, err := transformer.TransformResponse(ctx, provider, response)
		testutil.AssertNoError(t, err)

		// Should pass through unchanged
		testutil.AssertEqual(t, "invalid response", result.(string))
	})
}

func TestToolTransformSSEEvent(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewToolTransformer()
	ctx := context.Background()
	provider := &config.Provider{Name: "anthropic"}

	t.Run("PassThroughEvent", func(t *testing.T) {
		event := "test event data"

		result, err := transformer.TransformSSEEvent(ctx, provider, event)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, event, result)
	})
}

func TestToolExtractToolCalls(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewToolTransformer()

	t.Run("ExtractValidToolCalls", func(t *testing.T) {
		message := map[string]interface{}{
			"role": "assistant",
			"content": []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "Let me search for that information.",
				},
				map[string]interface{}{
					"type": "tool_use",
					"id":   "tool_456",
					"name": "search",
					"input": map[string]interface{}{
						"query": "weather forecast",
					},
				},
			},
		}

		toolUses, err := transformer.ExtractToolCalls(message)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 1, len(toolUses))

		toolUse := toolUses[0]
		testutil.AssertEqual(t, "tool_use", toolUse.Type)
		testutil.AssertEqual(t, "tool_456", toolUse.ID)
		testutil.AssertEqual(t, "search", toolUse.Name)

		// Check input parsing
		var input map[string]interface{}
		json.Unmarshal(toolUse.Input, &input)
		testutil.AssertEqual(t, "weather forecast", input["query"])
	})

	t.Run("NoToolCalls", func(t *testing.T) {
		message := map[string]interface{}{
			"role":    "assistant",
			"content": "Just a text response",
		}

		toolUses, err := transformer.ExtractToolCalls(message)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 0, len(toolUses))
	})

	t.Run("InvalidMessage", func(t *testing.T) {
		message := "not a valid message"

		toolUses, err := transformer.ExtractToolCalls(message)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 0, len(toolUses))
	})
}

func TestToolCreateToolResultMessage(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewToolTransformer()

	t.Run("CreateSuccessResult", func(t *testing.T) {
		results := []tools.ToolResult{
			{
				Type:      "tool_result",
				ToolUseID: "tool_123",
				Content:   json.RawMessage(`{"weather": "sunny", "temperature": "75F"}`),
			},
		}

		message := transformer.CreateToolResultMessage(results)
		messageMap := message.(map[string]interface{})
		testutil.AssertEqual(t, "user", messageMap["role"])

		content := messageMap["content"].([]tools.ContentBlock)
		testutil.AssertEqual(t, 1, len(content))

		block := content[0]
		testutil.AssertEqual(t, "tool_result", block.Type)
		testutil.AssertEqual(t, "tool_123", block.ToolUseID)
		testutil.AssertEqual(t, `{"weather": "sunny", "temperature": "75F"}`, string(block.Content))
	})

	t.Run("CreateErrorResult", func(t *testing.T) {
		errorMsg := "API key not found"
		results := []tools.ToolResult{
			{
				Type:      "tool_result",
				ToolUseID: "tool_456",
				Error:     &errorMsg,
			},
		}

		message := transformer.CreateToolResultMessage(results)
		messageMap := message.(map[string]interface{})
		content := messageMap["content"].([]tools.ContentBlock)
		block := content[0]

		testutil.AssertEqual(t, "tool_result", block.Type)
		testutil.AssertEqual(t, "tool_456", block.ToolUseID)

		// Check error content format
		var errorContent map[string]interface{}
		json.Unmarshal(block.Content, &errorContent)
		testutil.AssertEqual(t, errorMsg, errorContent["error"])
	})

	t.Run("MultipleResults", func(t *testing.T) {
		errorMsg := "timeout"
		results := []tools.ToolResult{
			{
				Type:      "tool_result",
				ToolUseID: "tool_1",
				Content:   json.RawMessage(`{"result": "success"}`),
			},
			{
				Type:      "tool_result",
				ToolUseID: "tool_2",
				Error:     &errorMsg,
			},
		}

		message := transformer.CreateToolResultMessage(results)
		messageMap := message.(map[string]interface{})
		content := messageMap["content"].([]tools.ContentBlock)
		testutil.AssertEqual(t, 2, len(content))

		// Check first result (success)
		testutil.AssertEqual(t, "tool_1", content[0].ToolUseID)
		testutil.AssertEqual(t, `{"result": "success"}`, string(content[0].Content))

		// Check second result (error)
		testutil.AssertEqual(t, "tool_2", content[1].ToolUseID)
		var errorContent map[string]interface{}
		json.Unmarshal(content[1].Content, &errorContent)
		testutil.AssertEqual(t, errorMsg, errorContent["error"])
	})
}

func TestToolIsLegacyOpenAIModel(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	testCases := []struct {
		model    string
		expected bool
	}{
		{"gpt-3.5-turbo-0613", true},
		{"gpt-3.5-turbo-16k-0613", true},
		{"gpt-4-0613", true},
		{"gpt-4-32k-0613", true},
		{"gpt-4", false},
		{"gpt-3.5-turbo", false},
		{"gpt-4-turbo", false},
		{"claude-3", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.model, func(t *testing.T) {
			result := isLegacyOpenAIModel(tc.model)
			testutil.AssertEqual(t, tc.expected, result)
		})
	}
}

func TestToolTransformerIntegration(t *testing.T) {
	cfg := testutil.SetupTest(t)
	_ = cfg

	transformer := NewToolTransformer()
	ctx := context.Background()

	t.Run("CompleteWorkflow", func(t *testing.T) {
		// Test request transformation
		provider := &config.Provider{Name: "openai"}
		request := map[string]interface{}{
			"model": "gpt-4",
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "What's the weather like?",
				},
			},
			"tools": []interface{}{
				map[string]interface{}{
					"name":        "get_weather",
					"description": "Get current weather information",
					"input_schema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"location": map[string]interface{}{
								"type":        "string",
								"description": "City name",
							},
							"units": map[string]interface{}{
								"type":    "string",
								"enum":    []interface{}{"celsius", "fahrenheit"},
								"default": "celsius",
							},
						},
						"required": []interface{}{"location"},
					},
				},
			},
		}

		// Transform request
		transformedReq, err := transformer.TransformRequest(ctx, provider, request)
		testutil.AssertNoError(t, err)

		reqMap := transformedReq.(map[string]interface{})
		tools := reqMap["tools"].([]interface{})
		tool := tools[0].(map[string]interface{})

		// Should be in OpenAI function format
		testutil.AssertEqual(t, "function", tool["type"])
		function := tool["function"].(map[string]interface{})
		testutil.AssertEqual(t, "get_weather", function["name"])
		testutil.AssertEqual(t, "Get current weather information", function["description"])

		// Test response processing
		response := map[string]interface{}{
			"role": "assistant",
			"content": []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "I'll check the weather for you.",
				},
				map[string]interface{}{
					"type": "tool_use",
					"id":   "weather_call_123",
					"name": "get_weather",
					"input": map[string]interface{}{
						"location": "San Francisco",
						"units":    "fahrenheit",
					},
				},
			},
		}

		transformedResp, err := transformer.TransformResponse(ctx, provider, response)
		testutil.AssertNoError(t, err)

		respMap := transformedResp.(map[string]interface{})

		// Should have tool use metadata
		metadata := respMap["metadata"].(map[string]interface{})
		testutil.AssertEqual(t, true, metadata["has_tool_use"])

		// Tool use block should be processed
		content := respMap["content"].([]interface{})
		var toolBlock map[string]interface{}
		for _, block := range content {
			if blockMap, ok := block.(map[string]interface{}); ok {
				if blockMap["type"] == "tool_use" {
					toolBlock = blockMap
					break
				}
			}
		}

		testutil.AssertEqual(t, true, toolBlock["_processed"])
		testutil.AssertEqual(t, "ccproxy", toolBlock["_processor"])

		// Test tool extraction
		toolUses, err := transformer.ExtractToolCalls(transformedResp)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 1, len(toolUses))

		toolUse := toolUses[0]
		testutil.AssertEqual(t, "weather_call_123", toolUse.ID)
		testutil.AssertEqual(t, "get_weather", toolUse.Name)
	})
}
