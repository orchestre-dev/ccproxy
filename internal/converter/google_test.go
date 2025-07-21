package converter

import (
	"encoding/json"
	"testing"

	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
)

func TestNewGoogleConverter(t *testing.T) {
	converter := NewGoogleConverter()
	testutil.AssertNotEqual(t, nil, converter)
}

func TestGoogleConverter_ToGeneric_Request(t *testing.T) {
	converter := NewGoogleConverter()

	tests := []struct {
		name             string
		input            string
		expectedMessages int
		expectError      bool
	}{
		{
			name: "basic request",
			input: `{
				"contents": [
					{
						"role": "user",
						"parts": [
							{"text": "Hello world"}
						]
					}
				],
				"generationConfig": {
					"temperature": 0.7,
					"maxOutputTokens": 100
				}
			}`,
			expectedMessages: 1,
			expectError:      false,
		},
		{
			name: "request with model role",
			input: `{
				"contents": [
					{
						"role": "user",
						"parts": [{"text": "Hello"}]
					},
					{
						"role": "model",
						"parts": [{"text": "Hi there!"}]
					},
					{
						"role": "user", 
						"parts": [{"text": "How are you?"}]
					}
				],
				"generationConfig": {
					"temperature": 0.5,
					"maxOutputTokens": 200,
					"topP": 0.95,
					"topK": 40
				}
			}`,
			expectedMessages: 3,
			expectError:      false,
		},
		{
			name: "request with multiple parts",
			input: `{
				"contents": [
					{
						"role": "user",
						"parts": [
							{"text": "First part. "},
							{"text": "Second part."},
							{"text": " Third part."}
						]
					}
				],
				"generationConfig": {
					"maxOutputTokens": 150,
					"stopSequences": ["END", "STOP"]
				}
			}`,
			expectedMessages: 1,
			expectError:      false,
		},
		{
			name: "request without generation config",
			input: `{
				"contents": [
					{
						"role": "user",
						"parts": [{"text": "Simple request"}]
					}
				]
			}`,
			expectedMessages: 1,
			expectError:      false,
		},
		{
			name: "request with empty parts",
			input: `{
				"contents": [
					{
						"role": "user",
						"parts": []
					}
				],
				"generationConfig": {
					"maxOutputTokens": 100
				}
			}`,
			expectedMessages: 1,
			expectError:      false,
		},
		{
			name:        "malformed JSON",
			input:       `{"contents": [`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.ToGeneric(json.RawMessage(tt.input), true)

			if tt.expectError {
				testutil.AssertError(t, err)
				return
			}

			testutil.AssertNoError(t, err)
			testutil.AssertNotEqual(t, nil, result)

			// Parse the result to verify structure
			var genericReq Request
			err = json.Unmarshal(result, &genericReq)
			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, tt.expectedMessages, len(genericReq.Messages))
		})
	}
}

func TestGoogleConverter_ToGeneric_Response(t *testing.T) {
	converter := NewGoogleConverter()

	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name: "basic response",
			input: `{
				"candidates": [
					{
						"content": {
							"role": "model",
							"parts": [
								{"text": "Hello there!"}
							]
						},
						"finishReason": "STOP",
						"safetyRatings": []
					}
				],
				"usageMetadata": {
					"promptTokenCount": 10,
					"candidatesTokenCount": 5,
					"totalTokenCount": 15
				}
			}`,
			expectError: false,
		},
		{
			name: "response without usage",
			input: `{
				"candidates": [
					{
						"content": {
							"role": "model",
							"parts": [
								{"text": "Response without usage"}
							]
						},
						"finishReason": "STOP"
					}
				]
			}`,
			expectError: false,
		},
		{
			name: "response with multiple parts",
			input: `{
				"candidates": [
					{
						"content": {
							"role": "model",
							"parts": [
								{"text": "First part. "},
								{"text": "Second part."}
							]
						},
						"finishReason": "MAX_TOKENS"
					}
				],
				"usageMetadata": {
					"promptTokenCount": 20,
					"candidatesTokenCount": 8,
					"totalTokenCount": 28
				}
			}`,
			expectError: false,
		},
		{
			name: "response with multiple candidates",
			input: `{
				"candidates": [
					{
						"content": {
							"role": "model",
							"parts": [{"text": "First candidate"}]
						},
						"finishReason": "STOP"
					},
					{
						"content": {
							"role": "model",
							"parts": [{"text": "Second candidate"}]
						},
						"finishReason": "STOP"
					}
				]
			}`,
			expectError: false, // Should use first candidate
		},
		{
			name: "response with no candidates",
			input: `{
				"candidates": []
			}`,
			expectError: true,
		},
		{
			name: "response with empty parts",
			input: `{
				"candidates": [
					{
						"content": {
							"role": "model",
							"parts": []
						},
						"finishReason": "STOP"
					}
				]
			}`,
			expectError: false,
		},
		{
			name:        "malformed JSON",
			input:       `{"candidates": [`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.ToGeneric(json.RawMessage(tt.input), false)

			if tt.expectError {
				testutil.AssertError(t, err)
				return
			}

			testutil.AssertNoError(t, err)
			testutil.AssertNotEqual(t, nil, result)

			// Parse the result to verify structure
			var genericResp Response
			err = json.Unmarshal(result, &genericResp)
			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, "message", genericResp.Type)
			testutil.AssertEqual(t, "assistant", genericResp.Role) // model -> assistant
		})
	}
}

func TestGoogleConverter_FromGeneric_Request(t *testing.T) {
	converter := NewGoogleConverter()

	tests := []struct {
		name             string
		input            string
		expectedContents int
		expectSystem     bool
		expectError      bool
	}{
		{
			name: "basic request",
			input: `{
				"model": "gemini-pro",
				"messages": [
					{
						"role": "user",
						"content": "Hello world"
					}
				],
				"max_tokens": 100,
				"temperature": 0.7
			}`,
			expectedContents: 1,
			expectSystem:     false,
			expectError:      false,
		},
		{
			name: "request with system",
			input: `{
				"model": "gemini-pro",
				"messages": [
					{
						"role": "user",
						"content": "Hello"
					}
				],
				"system": "You are helpful",
				"max_tokens": 200,
				"temperature": 0.5
			}`,
			expectedContents: 2, // System + user message
			expectSystem:     true,
			expectError:      false,
		},
		{
			name: "request with assistant message",
			input: `{
				"model": "gemini-pro",
				"messages": [
					{
						"role": "user",
						"content": "Hello"
					},
					{
						"role": "assistant",
						"content": "Hi there!"
					}
				],
				"max_tokens": 150
			}`,
			expectedContents: 2,
			expectSystem:     false,
			expectError:      false,
		},
		{
			name: "request with object content",
			input: `{
				"model": "gemini-pro",
				"messages": [
					{
						"role": "user",
						"content": {"text": "Object content", "other": "data"}
					}
				],
				"max_tokens": 100
			}`,
			expectedContents: 1,
			expectSystem:     false,
			expectError:      false,
		},
		{
			name: "request with zero temperature and tokens",
			input: `{
				"model": "gemini-pro",
				"messages": [
					{
						"role": "user",
						"content": "Test"
					}
				],
				"max_tokens": 0,
				"temperature": 0.0
			}`,
			expectedContents: 1,
			expectSystem:     false,
			expectError:      false,
		},
		{
			name:        "malformed JSON",
			input:       `{"messages": [`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.FromGeneric(json.RawMessage(tt.input), true)

			if tt.expectError {
				testutil.AssertError(t, err)
				return
			}

			testutil.AssertNoError(t, err)
			testutil.AssertNotEqual(t, nil, result)

			// Parse the result to verify structure
			var googleReq GoogleRequest
			err = json.Unmarshal(result, &googleReq)
			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, tt.expectedContents, len(googleReq.Contents))

			// Check for system message
			if tt.expectSystem {
				testutil.AssertEqual(t, "user", googleReq.Contents[0].Role)
			}
		})
	}
}

func TestGoogleConverter_FromGeneric_Response(t *testing.T) {
	converter := NewGoogleConverter()

	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name: "basic response",
			input: `{
				"id": "resp_123",
				"type": "message",
				"role": "assistant",
				"content": "Hello there!",
				"model": "gemini-pro",
				"usage": {
					"input_tokens": 10,
					"output_tokens": 5,
					"total_tokens": 15
				}
			}`,
			expectError: false,
		},
		{
			name: "response without usage",
			input: `{
				"id": "resp_456",
				"type": "message",
				"role": "assistant",
				"content": "Response without usage",
				"model": "gemini-pro"
			}`,
			expectError: false,
		},
		{
			name: "response with model role",
			input: `{
				"id": "resp_789",
				"type": "message",
				"role": "model",
				"content": "Model response",
				"model": "gemini-pro"
			}`,
			expectError: false,
		},
		{
			name:        "malformed JSON",
			input:       `{"id": "resp_123"`,
			expectError: true,
		},
		{
			name:        "malformed content",
			input:       `{"id": "resp_123", "content": invalid_json}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.FromGeneric(json.RawMessage(tt.input), false)

			if tt.expectError {
				testutil.AssertError(t, err)
				return
			}

			testutil.AssertNoError(t, err)
			testutil.AssertNotEqual(t, nil, result)

			// Parse the result to verify structure
			var googleResp GoogleResponse
			err = json.Unmarshal(result, &googleResp)
			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, 1, len(googleResp.Candidates))
			testutil.AssertEqual(t, "STOP", googleResp.Candidates[0].FinishReason)
		})
	}
}

func TestGoogleConverter_ConvertStreamEvent(t *testing.T) {
	converter := NewGoogleConverter()

	tests := []struct {
		name     string
		input    []byte
		toFormat MessageFormat
	}{
		{
			name:     "basic stream event",
			input:    []byte("data: {\"candidates\": [{\"content\": {\"parts\": [{\"text\": \"Hello\"}]}}]}"),
			toFormat: FormatAnthropic,
		},
		{
			name:     "empty event",
			input:    []byte(""),
			toFormat: FormatOpenAI,
		},
		{
			name:     "complex event",
			input:    []byte("event: generation\ndata: {\"usageMetadata\": {\"totalTokenCount\": 25}}"),
			toFormat: FormatAWS,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.ConvertStreamEvent(tt.input, tt.toFormat)

			// Current implementation just passes through
			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, string(tt.input), string(result))
		})
	}
}

func TestGoogleConverter_RoleMapping(t *testing.T) {
	converter := NewGoogleConverter()

	t.Run("model to assistant mapping in ToGeneric", func(t *testing.T) {
		input := `{
			"candidates": [
				{
					"content": {
						"role": "model",
						"parts": [{"text": "Hello"}]
					},
					"finishReason": "STOP"
				}
			]
		}`

		result, err := converter.ToGeneric(json.RawMessage(input), false)
		testutil.AssertNoError(t, err)

		var genericResp Response
		err = json.Unmarshal(result, &genericResp)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, "assistant", genericResp.Role)
	})

	t.Run("assistant to model mapping in FromGeneric", func(t *testing.T) {
		input := `{
			"messages": [
				{
					"role": "assistant",
					"content": "Hello"
				}
			]
		}`

		result, err := converter.FromGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)

		var googleReq GoogleRequest
		err = json.Unmarshal(result, &googleReq)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, "model", googleReq.Contents[0].Role)
	})

	t.Run("user role preserved", func(t *testing.T) {
		input := `{
			"messages": [
				{
					"role": "user",
					"content": "Hello"
				}
			]
		}`

		result, err := converter.FromGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)

		var googleReq GoogleRequest
		err = json.Unmarshal(result, &googleReq)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, "user", googleReq.Contents[0].Role)
	})

	t.Run("response role mapping in FromGeneric", func(t *testing.T) {
		input := `{
			"id": "resp_123",
			"type": "message",
			"role": "assistant",
			"content": "Hello",
			"model": "gemini-pro"
		}`

		result, err := converter.FromGeneric(json.RawMessage(input), false)
		testutil.AssertNoError(t, err)

		var googleResp GoogleResponse
		err = json.Unmarshal(result, &googleResp)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, "model", googleResp.Candidates[0].Content.Role)
	})
}

func TestGoogleConverter_ContentHandling(t *testing.T) {
	converter := NewGoogleConverter()

	t.Run("multiple parts concatenation in ToGeneric", func(t *testing.T) {
		input := `{
			"contents": [
				{
					"role": "user",
					"parts": [
						{"text": "First part"},
						{"text": "Second part"},
						{"text": "Third part"}
					]
				}
			]
		}`

		result, err := converter.ToGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)

		var genericReq Request
		err = json.Unmarshal(result, &genericReq)
		testutil.AssertNoError(t, err)

		var content string
		err = json.Unmarshal(genericReq.Messages[0].Content, &content)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, "First partSecond partThird part", content)
	})

	t.Run("empty parts in ToGeneric", func(t *testing.T) {
		input := `{
			"contents": [
				{
					"role": "user",
					"parts": []
				}
			]
		}`

		result, err := converter.ToGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)

		var genericReq Request
		err = json.Unmarshal(result, &genericReq)
		testutil.AssertNoError(t, err)

		var content string
		err = json.Unmarshal(genericReq.Messages[0].Content, &content)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, "", content)
	})

	t.Run("content marshal error in ToGeneric", func(t *testing.T) {
		// This should succeed normally
		input := `{
			"contents": [
				{
					"role": "user",
					"parts": [{"text": "Hello"}]
				}
			]
		}`

		_, err := converter.ToGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)
	})

	t.Run("response parts concatenation in ToGeneric", func(t *testing.T) {
		input := `{
			"candidates": [
				{
					"content": {
						"role": "model",
						"parts": [
							{"text": "Part one "},
							{"text": "Part two"}
						]
					},
					"finishReason": "STOP"
				}
			]
		}`

		result, err := converter.ToGeneric(json.RawMessage(input), false)
		testutil.AssertNoError(t, err)

		var genericResp Response
		err = json.Unmarshal(result, &genericResp)
		testutil.AssertNoError(t, err)

		var content string
		err = json.Unmarshal(genericResp.Content, &content)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, "Part one Part two", content)
	})
}

func TestGoogleConverter_GenerationConfig(t *testing.T) {
	converter := NewGoogleConverter()

	t.Run("generation config in ToGeneric", func(t *testing.T) {
		input := `{
			"contents": [
				{
					"role": "user",
					"parts": [{"text": "Hello"}]
				}
			],
			"generationConfig": {
				"temperature": 0.8,
				"maxOutputTokens": 250,
				"topP": 0.95,
				"topK": 40,
				"stopSequences": ["END", "STOP"]
			}
		}`

		result, err := converter.ToGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)

		var genericReq Request
		err = json.Unmarshal(result, &genericReq)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 250, genericReq.MaxTokens)
		testutil.AssertEqual(t, 0.8, genericReq.Temperature)
	})

	t.Run("generation config creation in FromGeneric", func(t *testing.T) {
		input := `{
			"messages": [
				{
					"role": "user",
					"content": "Hello"
				}
			],
			"max_tokens": 300,
			"temperature": 0.6
		}`

		result, err := converter.FromGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)

		var googleReq GoogleRequest
		err = json.Unmarshal(result, &googleReq)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, googleReq.GenerationConfig)
		testutil.AssertEqual(t, 300, googleReq.GenerationConfig.MaxOutputTokens)
		testutil.AssertEqual(t, 0.6, googleReq.GenerationConfig.Temperature)
	})

	t.Run("no generation config when values are zero", func(t *testing.T) {
		input := `{
			"messages": [
				{
					"role": "user",
					"content": "Hello"
				}
			],
			"max_tokens": 0,
			"temperature": 0.0
		}`

		result, err := converter.FromGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)

		var googleReq GoogleRequest
		err = json.Unmarshal(result, &googleReq)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, (*GoogleGenerationConfig)(nil), googleReq.GenerationConfig)
	})

	t.Run("generation config with only max tokens", func(t *testing.T) {
		input := `{
			"messages": [
				{
					"role": "user",
					"content": "Hello"
				}
			],
			"max_tokens": 150,
			"temperature": 0.0
		}`

		result, err := converter.FromGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)

		var googleReq GoogleRequest
		err = json.Unmarshal(result, &googleReq)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, googleReq.GenerationConfig)
		testutil.AssertEqual(t, 150, googleReq.GenerationConfig.MaxOutputTokens)
		testutil.AssertEqual(t, 0.0, googleReq.GenerationConfig.Temperature)
	})
}

func TestGoogleConverter_ErrorHandling(t *testing.T) {
	converter := NewGoogleConverter()

	t.Run("invalid request JSON", func(t *testing.T) {
		input := `{"contents": invalid}`
		_, err := converter.ToGeneric(json.RawMessage(input), true)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "failed to unmarshal Google request")
	})

	t.Run("invalid response JSON", func(t *testing.T) {
		input := `{"candidates": invalid}`
		_, err := converter.ToGeneric(json.RawMessage(input), false)
		testutil.AssertError(t, err)
		testutil.AssertContains(t, err.Error(), "failed to unmarshal Google response")
	})

	t.Run("content marshal error in ToGeneric request", func(t *testing.T) {
		// This should succeed normally
		input := `{
			"contents": [
				{
					"role": "user",
					"parts": [{"text": "Hello"}]
				}
			]
		}`

		_, err := converter.ToGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)
	})

	t.Run("content marshal error in ToGeneric response", func(t *testing.T) {
		input := `{
			"candidates": [
				{
					"content": {
						"role": "model",
						"parts": [{"text": "Hello"}]
					},
					"finishReason": "STOP"
				}
			]
		}`

		_, err := converter.ToGeneric(json.RawMessage(input), false)
		testutil.AssertNoError(t, err)
	})
}

func TestGoogleConverter_EdgeCases(t *testing.T) {
	converter := NewGoogleConverter()

	t.Run("empty contents array", func(t *testing.T) {
		input := `{
			"contents": [],
			"generationConfig": {
				"maxOutputTokens": 100
			}
		}`

		result, err := converter.ToGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)

		var genericReq Request
		err = json.Unmarshal(result, &genericReq)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 0, len(genericReq.Messages))
	})

	t.Run("empty system message handling", func(t *testing.T) {
		input := `{
			"messages": [
				{
					"role": "user",
					"content": "Hello"
				}
			],
			"system": "",
			"max_tokens": 100
		}`

		result, err := converter.FromGeneric(json.RawMessage(input), true)
		testutil.AssertNoError(t, err)

		var googleReq GoogleRequest
		err = json.Unmarshal(result, &googleReq)
		testutil.AssertNoError(t, err)
		testutil.AssertEqual(t, 1, len(googleReq.Contents)) // Only user message
	})

	t.Run("object content parsing error in FromGeneric", func(t *testing.T) {
		input := `{
			"messages": [
				{
					"role": "user",
					"content": invalid_json
				}
			]
		}`

		_, err := converter.FromGeneric(json.RawMessage(input), true)
		testutil.AssertError(t, err)
	})

	t.Run("usage metadata mapping", func(t *testing.T) {
		input := `{
			"candidates": [
				{
					"content": {
						"role": "model",
						"parts": [{"text": "Hello"}]
					},
					"finishReason": "STOP"
				}
			],
			"usageMetadata": {
				"promptTokenCount": 25,
				"candidatesTokenCount": 15,
				"totalTokenCount": 40
			}
		}`

		result, err := converter.ToGeneric(json.RawMessage(input), false)
		testutil.AssertNoError(t, err)

		var genericResp Response
		err = json.Unmarshal(result, &genericResp)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, genericResp.Usage)
		testutil.AssertEqual(t, 25, genericResp.Usage.InputTokens)
		testutil.AssertEqual(t, 15, genericResp.Usage.OutputTokens)
		testutil.AssertEqual(t, 40, genericResp.Usage.TotalTokens)
	})

	t.Run("usage metadata creation in FromGeneric", func(t *testing.T) {
		input := `{
			"id": "resp_123",
			"type": "message",
			"role": "assistant",
			"content": "Hello",
			"model": "gemini-pro",
			"usage": {
				"input_tokens": 30,
				"output_tokens": 20,
				"total_tokens": 50
			}
		}`

		result, err := converter.FromGeneric(json.RawMessage(input), false)
		testutil.AssertNoError(t, err)

		var googleResp GoogleResponse
		err = json.Unmarshal(result, &googleResp)
		testutil.AssertNoError(t, err)
		testutil.AssertNotEqual(t, nil, googleResp.UsageMetadata)
		testutil.AssertEqual(t, 30, googleResp.UsageMetadata.PromptTokenCount)
		testutil.AssertEqual(t, 20, googleResp.UsageMetadata.CandidatesTokenCount)
		testutil.AssertEqual(t, 50, googleResp.UsageMetadata.TotalTokenCount)
	})
}
