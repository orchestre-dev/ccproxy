package utils

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/pkoukk/tiktoken-go"
)

var (
	encoder     *tiktoken.Tiktoken
	encoderOnce sync.Once
	encoderErr  error
)

// InitTokenizer initializes the tiktoken encoder with cl100k_base encoding
func InitTokenizer() error {
	encoderOnce.Do(func() {
		encoder, encoderErr = tiktoken.GetEncoding("cl100k_base")
	})
	return encoderErr
}

// GetEncoder returns the initialized encoder
func GetEncoder() (*tiktoken.Tiktoken, error) {
	if err := InitTokenizer(); err != nil {
		return nil, err
	}
	return encoder, nil
}

// CountTokens counts the number of tokens in a string
func CountTokens(text string) (int, error) {
	enc, err := GetEncoder()
	if err != nil {
		return 0, fmt.Errorf("failed to get encoder: %w", err)
	}

	tokens := enc.Encode(text, nil, nil)
	return len(tokens), nil
}

// MessageContent represents content that can be either string or array
type MessageContent interface{}

// TextContent represents text content in a message
type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ToolUseContent represents tool use in a message
type ToolUseContent struct {
	Type  string      `json:"type"`
	ID    string      `json:"id"`
	Name  string      `json:"name"`
	Input interface{} `json:"input"`
}

// ToolResultContent represents tool result in a message
type ToolResultContent struct {
	Type    string      `json:"type"`
	ToolUse string      `json:"tool_use_id"`
	Content interface{} `json:"content"`
}

// SystemContent represents system message content
type SystemContent struct {
	Type string   `json:"type"`
	Text []string `json:"text"`
}

// Tool represents a tool with its schema
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	InputSchema interface{} `json:"input_schema,omitempty"`
}

// MessageCreateParams represents the parameters for creating a message
type MessageCreateParams struct {
	Model    string      `json:"model"`
	Messages []Message   `json:"messages"`
	System   interface{} `json:"system,omitempty"` // Can be string or []SystemContent
	Tools    []Tool      `json:"tools,omitempty"`
	Stream   bool        `json:"stream,omitempty"`
}

// Message represents a message in the conversation
type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // Can be string or array of content objects
}

// CountMessageTokens counts tokens in a MessageCreateParams structure
func CountMessageTokens(params *MessageCreateParams) (int, error) {
	enc, err := GetEncoder()
	if err != nil {
		return 0, fmt.Errorf("failed to get encoder: %w", err)
	}

	tokenCount := 0

	// Count tokens in messages
	for _, message := range params.Messages {
		count, err := countMessageContentTokens(enc, message.Content)
		if err != nil {
			return 0, fmt.Errorf("failed to count message tokens: %w", err)
		}
		tokenCount += count
	}

	// Count tokens in system prompt
	if params.System != nil {
		count := countSystemTokens(enc, params.System)
		tokenCount += count
	}

	// Count tokens in tools
	if params.Tools != nil {
		for _, tool := range params.Tools {
			// Count name + description
			if tool.Description != "" {
				tokens := enc.Encode(tool.Name+tool.Description, nil, nil)
				tokenCount += len(tokens)
			}

			// Count input schema as JSON
			if tool.InputSchema != nil {
				schemaJSON, err := json.Marshal(tool.InputSchema)
				if err != nil {
					return 0, fmt.Errorf("failed to marshal tool schema: %w", err)
				}
				tokens := enc.Encode(string(schemaJSON), nil, nil)
				tokenCount += len(tokens)
			}
		}
	}

	return tokenCount
}

// countMessageContentTokens counts tokens in message content
func countMessageContentTokens(enc *tiktoken.Tiktoken, content interface{}) (int, error) {
	tokenCount := 0

	switch c := content.(type) {
	case string:
		// Simple string content
		tokens := enc.Encode(c, nil, nil)
		tokenCount += len(tokens)

	case []interface{}:
		// Array of content objects
		for _, item := range c {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			contentType, ok := itemMap["type"].(string)
			if !ok {
				continue
			}

			switch contentType {
			case "text":
				if text, ok := itemMap["text"].(string); ok {
					tokens := enc.Encode(text, nil, nil)
					tokenCount += len(tokens)
				}

			case "tool_use":
				if input, ok := itemMap["input"]; ok {
					inputJSON, err := json.Marshal(input)
					if err != nil {
						return 0, fmt.Errorf("failed to marshal tool input: %w", err)
					}
					tokens := enc.Encode(string(inputJSON), nil, nil)
					tokenCount += len(tokens)
				}

			case "tool_result":
				if content, ok := itemMap["content"]; ok {
					var text string
					if str, ok := content.(string); ok {
						text = str
					} else {
						contentJSON, err := json.Marshal(content)
						if err != nil {
							return 0, fmt.Errorf("failed to marshal tool result: %w", err)
						}
						text = string(contentJSON)
					}
					tokens := enc.Encode(text, nil, nil)
					tokenCount += len(tokens)
				}
			}
		}
	}

	return tokenCount
}

// countSystemTokens counts tokens in system content
func countSystemTokens(enc *tiktoken.Tiktoken, system interface{}) int {
	tokenCount := 0

	switch s := system.(type) {
	case string:
		// Simple string system prompt
		tokens := enc.Encode(s, nil, nil)
		tokenCount += len(tokens)

	case []interface{}:
		// Array of system content objects
		for _, item := range s {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			if itemMap["type"] != "text" {
				continue
			}

			// Handle text which can be string or array
			if text, ok := itemMap["text"].(string); ok {
				tokens := enc.Encode(text, nil, nil)
				tokenCount += len(tokens)
			} else if textArray, ok := itemMap["text"].([]interface{}); ok {
				for _, textPart := range textArray {
					if str, ok := textPart.(string); ok {
						tokens := enc.Encode(str, nil, nil)
						tokenCount += len(tokens)
					}
				}
			}
		}
	}

	return tokenCount
}
