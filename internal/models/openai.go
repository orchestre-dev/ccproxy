package models

// ChatCompletionRequest represents an OpenAI chat completion request
type ChatCompletionRequest struct {
	ToolChoice  any                  `json:"tool_choice,omitempty"`
	MaxTokens   *int                 `json:"max_tokens,omitempty"`
	Temperature *float64             `json:"temperature,omitempty"`
	Model       string               `json:"model"`
	Messages    []ChatMessage        `json:"messages"`
	Tools       []ChatCompletionTool `json:"tools,omitempty"`
}

// ChatCompletionResponse represents an OpenAI chat completion response
type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Model   string                 `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
	Usage   ChatCompletionUsage    `json:"usage"`
	Created int64                  `json:"created"`
}

// ChatMessage represents a message in the chat completion
type ChatMessage struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ChatCompletionChoice represents a choice in the response
type ChatCompletionChoice struct {
	FinishReason string      `json:"finish_reason"`
	Message      ChatMessage `json:"message"`
	Index        int         `json:"index"`
}

// ChatCompletionTool represents a tool definition in OpenAI format
type ChatCompletionTool struct {
	Function ChatCompletionToolFunction `json:"function"`
	Type     string                     `json:"type"`
}

// ChatCompletionToolFunction represents a function tool
type ChatCompletionToolFunction struct {
	Parameters  map[string]interface{} `json:"parameters"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
}

// ToolCall represents a tool call in a message
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall represents a function call
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ChatCompletionUsage represents usage statistics
type ChatCompletionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
