// Package converter provides conversion utilities between different API formats
package converter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ccproxy/internal/models"
	"ccproxy/pkg/logger"
)

// Converter handles bidirectional API format conversion with enhanced error handling and validation
type Converter interface {
	// ConvertRequest converts Anthropic request to OpenAI format
	ConvertRequest(ctx context.Context, req *models.MessagesRequest) (*models.ChatCompletionRequest, error)
	
	// ConvertResponse converts OpenAI response to Anthropic format
	ConvertResponse(ctx context.Context, resp *models.ChatCompletionResponse, requestID string, providerName string) (*models.MessagesResponse, error)
	
	// ValidateRequest validates Anthropic request before conversion
	ValidateRequest(req *models.MessagesRequest) error
	
	// SupportedFeatures returns what features this converter supports
	SupportedFeatures() ConverterFeatures
}

// ConverterFeatures tracks supported capabilities per provider
type ConverterFeatures struct {
	SupportsTools     bool `json:"supports_tools"`
	SupportsStreaming bool `json:"supports_streaming"`
	MaxTokens         int  `json:"max_tokens"`
	MaxMessageSize    int  `json:"max_message_size"`
	MaxMessages       int  `json:"max_messages"`
	MaxToolCalls      int  `json:"max_tool_calls"`
}

// ConverterConfig provides configurable validation and processing options
type ConverterConfig struct {
	MaxRequestSize   int64 `json:"max_request_size"`
	MaxMessages      int   `json:"max_messages"`
	MaxToolCalls     int   `json:"max_tool_calls"`
	ValidateSchemas  bool  `json:"validate_schemas"`
	StrictMode       bool  `json:"strict_mode"`
	LogSanitization  bool  `json:"log_sanitization"`
}

// ToolConversionContext maintains request context during conversion
type ToolConversionContext struct {
	RequestID    string                `json:"request_id"`
	ProviderName string                `json:"provider_name"`
	Capabilities ConverterFeatures     `json:"capabilities"`
	Logger       *logger.Logger        `json:"-"`
	StartTime    time.Time             `json:"start_time"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// EnhancedConverter implements the Converter interface with proper error handling and validation
type EnhancedConverter struct {
	logger *logger.Logger
	config *ConverterConfig
}

// NewEnhancedConverter creates a new enhanced converter with the given configuration
func NewEnhancedConverter(config *ConverterConfig, log *logger.Logger) *EnhancedConverter {
	if config == nil {
		config = &ConverterConfig{
			MaxRequestSize:  10 * 1024 * 1024, // 10MB
			MaxMessages:     100,
			MaxToolCalls:    50,
			ValidateSchemas: true,
			StrictMode:      false,
			LogSanitization: true,
		}
	}
	
	if log == nil {
		log = logger.New()
	}
	
	return &EnhancedConverter{
		logger: log,
		config: config,
	}
}

// SupportedFeatures returns the features supported by this converter
func (c *EnhancedConverter) SupportedFeatures() ConverterFeatures {
	return ConverterFeatures{
		SupportsTools:     true,
		SupportsStreaming: true,
		MaxTokens:         200000,
		MaxMessageSize:    1024 * 1024, // 1MB per message
		MaxMessages:       c.config.MaxMessages,
		MaxToolCalls:      c.config.MaxToolCalls,
	}
}

// ValidateRequest validates Anthropic request before conversion
func (c *EnhancedConverter) ValidateRequest(req *models.MessagesRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}
	
	if req.Model == "" {
		return fmt.Errorf("model cannot be empty")
	}
	
	if len(req.Messages) == 0 {
		return fmt.Errorf("messages array cannot be empty")
	}
	
	if len(req.Messages) > c.config.MaxMessages {
		return fmt.Errorf("too many messages: %d (max %d)", len(req.Messages), c.config.MaxMessages)
	}
	
	// Validate individual messages
	totalSize := 0
	for i, msg := range req.Messages {
		if err := c.validateMessage(msg); err != nil {
			return fmt.Errorf("message %d: %w", i, err)
		}
		totalSize += c.calculateMessageSize(msg)
	}
	
	if int64(totalSize) > c.config.MaxRequestSize {
		return fmt.Errorf("request too large: %d bytes (max %d)", totalSize, c.config.MaxRequestSize)
	}
	
	// Validate tools if present
	if req.Tools != nil {
		if len(req.Tools) > c.config.MaxToolCalls {
			return fmt.Errorf("too many tools: %d (max %d)", len(req.Tools), c.config.MaxToolCalls)
		}
		
		for i, tool := range req.Tools {
			if err := c.validateTool(tool); err != nil {
				return fmt.Errorf("tool %d (%s): %w", i, tool.Name, err)
			}
		}
	}
	
	return nil
}

// validateMessage validates an individual message
func (c *EnhancedConverter) validateMessage(msg models.Message) error {
	if msg.Role == "" {
		return fmt.Errorf("role cannot be empty")
	}
	
	if msg.Role != "user" && msg.Role != "assistant" && msg.Role != "system" {
		return fmt.Errorf("invalid role: %s", msg.Role)
	}
	
	if msg.Content == nil {
		return fmt.Errorf("content cannot be nil")
	}
	
	// Validate content based on type
	switch content := msg.Content.(type) {
	case string:
		if len(content) > c.SupportedFeatures().MaxMessageSize {
			return fmt.Errorf("message content too large: %d bytes (max %d)", len(content), c.SupportedFeatures().MaxMessageSize)
		}
	case []models.Content:
		for i, block := range content {
			if err := c.validateContentBlock(block); err != nil {
				return fmt.Errorf("content block %d: %w", i, err)
			}
		}
	default:
		return fmt.Errorf("invalid content type: %T", content)
	}
	
	return nil
}

// validateContentBlock validates a content block
func (c *EnhancedConverter) validateContentBlock(block models.Content) error {
	if block.Type == "" {
		return fmt.Errorf("content block type cannot be empty")
	}
	
	switch block.Type {
	case "text":
		if block.Text == "" {
			return fmt.Errorf("text content cannot be empty")
		}
	case "tool_use":
		if block.Name == "" {
			return fmt.Errorf("tool_use name cannot be empty")
		}
		if block.ID == "" {
			return fmt.Errorf("tool_use ID cannot be empty")
		}
	case "tool_result":
		if block.ToolUseID == "" {
			return fmt.Errorf("tool_result must have tool_use_id")
		}
	default:
		return fmt.Errorf("unsupported content block type: %s", block.Type)
	}
	
	return nil
}

// validateTool validates a tool definition
func (c *EnhancedConverter) validateTool(tool models.Tool) error {
	if tool.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	
	if tool.InputSchema == nil {
		return fmt.Errorf("tool input_schema cannot be nil")
	}
	
	// Basic JSON schema validation
	if c.config.ValidateSchemas {
		if schemaType, ok := tool.InputSchema["type"]; ok {
			if schemaType != "object" {
				return fmt.Errorf("tool input_schema type must be 'object', got: %v", schemaType)
			}
		} else {
			return fmt.Errorf("tool input_schema must have 'type' field")
		}
	}
	
	return nil
}

// calculateMessageSize calculates the approximate size of a message in bytes
func (c *EnhancedConverter) calculateMessageSize(msg models.Message) int {
	data, err := json.Marshal(msg)
	if err != nil {
		return 0
	}
	return len(data)
}

// ConvertRequest converts Anthropic request to OpenAI format with enhanced validation and error handling
func (c *EnhancedConverter) ConvertRequest(ctx context.Context, req *models.MessagesRequest) (*models.ChatCompletionRequest, error) {
	// Validate request first
	if err := c.ValidateRequest(req); err != nil {
		c.logger.WithError(err).Error("Request validation failed")
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	
	// Create conversion context
	convCtx := &ToolConversionContext{
		RequestID:    fmt.Sprintf("req_%d", time.Now().UnixNano()),
		ProviderName: "unknown", // Will be set by caller
		Capabilities: c.SupportedFeatures(),
		Logger:       c.logger,
		StartTime:    time.Now(),
		Metadata:     make(map[string]interface{}),
	}
	
	c.logger.WithField("request_id", convCtx.RequestID).
		WithField("model", req.Model).
		WithField("message_count", len(req.Messages)).
		WithField("tool_count", len(req.Tools)).
		Info("Starting request conversion")
	
	chatReq := &models.ChatCompletionRequest{
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}

	// Convert messages with enhanced logic
	messages, err := c.convertMessagesEnhanced(req.Messages, convCtx)
	if err != nil {
		c.logger.WithError(err).WithField("request_id", convCtx.RequestID).Error("Failed to convert messages")
		return nil, fmt.Errorf("failed to convert messages: %w", err)
	}
	chatReq.Messages = messages

	// Convert tools with enhanced logic
	if req.Tools != nil {
		tools, err := c.convertToolsEnhanced(req.Tools, convCtx)
		if err != nil {
			c.logger.WithError(err).WithField("request_id", convCtx.RequestID).Error("Failed to convert tools")
			return nil, fmt.Errorf("failed to convert tools: %w", err)
		}
		chatReq.Tools = tools
		chatReq.ToolChoice = req.ToolChoice
	}

	c.logger.WithField("request_id", convCtx.RequestID).
		WithField("conversion_time_ms", time.Since(convCtx.StartTime).Milliseconds()).
		Info("Request conversion completed successfully")

	return chatReq, nil
}

// ConvertResponse converts OpenAI response to Anthropic format with enhanced validation and error handling
func (c *EnhancedConverter) ConvertResponse(ctx context.Context, resp *models.ChatCompletionResponse, requestID string, providerName string) (*models.MessagesResponse, error) {
	if resp == nil {
		return nil, fmt.Errorf("response cannot be nil")
	}
	
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	// Create conversion context
	convCtx := &ToolConversionContext{
		RequestID:    requestID,
		ProviderName: providerName,
		Capabilities: c.SupportedFeatures(),
		Logger:       c.logger,
		StartTime:    time.Now(),
		Metadata:     make(map[string]interface{}),
	}
	
	c.logger.WithField("request_id", requestID).
		WithField("provider", providerName).
		WithField("choice_count", len(resp.Choices)).
		Info("Starting response conversion")

	choice := resp.Choices[0]
	message := choice.Message

	var content []models.Content
	var stopReason string

	// Handle tool calls with enhanced logic
	if len(message.ToolCalls) > 0 {
		toolContent, err := c.convertOpenAIToolCallsToAnthropicEnhanced(message.ToolCalls, convCtx)
		if err != nil {
			c.logger.WithError(err).WithField("request_id", requestID).Error("Failed to convert tool calls")
			return nil, fmt.Errorf("failed to convert tool calls: %w", err)
		}
		content = toolContent
		stopReason = "tool_use"
	} else {
		// Regular text response
		if message.Content != "" {
			content = append(content, models.Content{
				Type: "text",
				Text: message.Content,
			})
		}
		
		// Map OpenAI finish reasons to Anthropic stop reasons
		switch choice.FinishReason {
		case "stop":
			stopReason = "end_turn"
		case "length":
			stopReason = "max_tokens"
		case "tool_calls":
			stopReason = "tool_use"
		case "content_filter":
			stopReason = "stop_sequence"
		default:
			stopReason = "end_turn"
		}
	}

	// Generate message ID in the same format as Python proxy
	messageID := requestID
	if messageID == "" {
		messageID = fmt.Sprintf("msg_%x", time.Now().UnixNano()&0xFFFFFFFFFFFF)
	}

	response := &models.MessagesResponse{
		ID:         messageID,
		Type:       "message",
		Role:       "assistant",
		Model:      fmt.Sprintf("%s/%s", providerName, resp.Model),
		Content:    content,
		StopReason: stopReason,
		Usage: models.Usage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
		},
	}

	c.logger.WithField("request_id", requestID).
		WithField("conversion_time_ms", time.Since(convCtx.StartTime).Milliseconds()).
		WithField("stop_reason", stopReason).
		WithField("content_blocks", len(content)).
		Info("Response conversion completed successfully")

	return response, nil
}

// convertMessagesEnhanced converts Anthropic messages to OpenAI format with semantic preservation
func (c *EnhancedConverter) convertMessagesEnhanced(messages []models.Message, ctx *ToolConversionContext) ([]models.ChatMessage, error) {
	var chatMessages []models.ChatMessage

	for i, msg := range messages {
		chatMsg := models.ChatMessage{
			Role: msg.Role,
		}

		switch content := msg.Content.(type) {
		case string:
			chatMsg.Content = content
		case []models.Content:
			// Enhanced conversion that preserves semantic structure
			convertedContent, err := c.convertContentBlocks(content, ctx)
			if err != nil {
				return nil, fmt.Errorf("message %d: failed to convert content blocks: %w", i, err)
			}
			chatMsg.Content = convertedContent
		default:
			return nil, fmt.Errorf("message %d: unsupported content type: %T", i, content)
		}

		chatMessages = append(chatMessages, chatMsg)
	}

	return chatMessages, nil
}

// convertContentBlocks converts Anthropic content blocks to OpenAI format while preserving semantics
func (c *EnhancedConverter) convertContentBlocks(blocks []models.Content, ctx *ToolConversionContext) (string, error) {
	var parts []string
	
	for i, block := range blocks {
		switch block.Type {
		case "text":
			if block.Text == "" {
				ctx.Logger.WithField("request_id", ctx.RequestID).
					WithField("block_index", i).
					Warn("Empty text block encountered")
			}
			parts = append(parts, block.Text)
			
		case "tool_use":
			// Convert tool_use to a structured format that can be parsed back
			toolUse := map[string]interface{}{
				"type": "tool_use",
				"id":   block.ID,
				"name": block.Name,
				"input": block.Input,
			}
			
			toolJSON, err := json.Marshal(toolUse)
			if err != nil {
				return "", fmt.Errorf("block %d: failed to marshal tool_use: %w", i, err)
			}
			
			// Use a special marker that can be detected during response conversion
			parts = append(parts, fmt.Sprintf("__TOOL_USE_START__%s__TOOL_USE_END__", string(toolJSON)))
			
			ctx.Logger.WithField("request_id", ctx.RequestID).
				WithField("tool_name", block.Name).
				WithField("tool_id", block.ID).
				Debug("Converted tool_use block")
				
		case "tool_result":
			// Convert tool_result to a structured format
			toolResult := map[string]interface{}{
				"type": "tool_result",
				"tool_use_id": block.ToolUseID,
				"content": block.Content,
			}
			
			resultJSON, err := json.Marshal(toolResult)
			if err != nil {
				return "", fmt.Errorf("block %d: failed to marshal tool_result: %w", i, err)
			}
			
			parts = append(parts, fmt.Sprintf("__TOOL_RESULT_START__%s__TOOL_RESULT_END__", string(resultJSON)))
			
			ctx.Logger.WithField("request_id", ctx.RequestID).
				WithField("tool_use_id", block.ToolUseID).
				Debug("Converted tool_result block")
				
		default:
			return "", fmt.Errorf("block %d: unsupported content block type: %s", i, block.Type)
		}
	}
	
	return strings.Join(parts, "\n"), nil
}

// convertToolsEnhanced converts Anthropic tools to OpenAI format with enhanced validation
func (c *EnhancedConverter) convertToolsEnhanced(tools []models.Tool, ctx *ToolConversionContext) ([]models.ChatCompletionTool, error) {
	var chatTools []models.ChatCompletionTool

	for i, tool := range tools {
		// Additional validation beyond basic validation
		if err := c.validateToolSchema(tool, ctx); err != nil {
			return nil, fmt.Errorf("tool %d (%s): schema validation failed: %w", i, tool.Name, err)
		}
		
		description := ""
		if tool.Description != nil {
			description = *tool.Description
		}

		chatTool := models.ChatCompletionTool{
			Type: "function",
			Function: models.ChatCompletionToolFunction{
				Name:        tool.Name,
				Description: description,
				Parameters:  tool.InputSchema,
			},
		}

		chatTools = append(chatTools, chatTool)
		
		ctx.Logger.WithField("request_id", ctx.RequestID).
			WithField("tool_name", tool.Name).
			WithField("has_description", tool.Description != nil).
			Debug("Converted tool definition")
	}

	return chatTools, nil
}

// validateToolSchema performs enhanced validation of tool schemas including complex nested objects and arrays
func (c *EnhancedConverter) validateToolSchema(tool models.Tool, ctx *ToolConversionContext) error {
	if !c.config.ValidateSchemas {
		return nil
	}
	
	schema := tool.InputSchema
	
	// Validate required fields in schema
	if schemaType, ok := schema["type"]; !ok || schemaType != "object" {
		return fmt.Errorf("schema must have type 'object'")
	}
	
	// Validate properties if present with enhanced nested validation
	if properties, ok := schema["properties"]; ok {
		propsMap, ok := properties.(map[string]interface{})
		if !ok {
			return fmt.Errorf("properties must be an object")
		}
		
		// Enhanced validation for complex nested objects and arrays
		for propName, propDef := range propsMap {
			if err := c.validateSchemaProperty(propName, propDef, 0, ctx); err != nil {
				return fmt.Errorf("property '%s': %w", propName, err)
			}
		}
	}
	
	// Validate required array if present
	if required, ok := schema["required"]; ok {
		requiredArray, ok := required.([]interface{})
		if !ok {
			return fmt.Errorf("required must be an array")
		}
		
		for i, req := range requiredArray {
			if _, ok := req.(string); !ok {
				return fmt.Errorf("required[%d] must be a string", i)
			}
		}
	}
	
	return nil
}

// validateSchemaProperty validates a single schema property with support for complex nested structures
func (c *EnhancedConverter) validateSchemaProperty(propName string, propDef interface{}, depth int, ctx *ToolConversionContext) error {
	// Prevent infinite recursion in deeply nested schemas
	const maxDepth = 10
	if depth > maxDepth {
		return fmt.Errorf("schema nesting too deep (max %d levels)", maxDepth)
	}
	
	propDefMap, ok := propDef.(map[string]interface{})
	if !ok {
		return fmt.Errorf("property definition must be an object")
	}
	
	// Validate property type
	propType, hasType := propDefMap["type"]
	if !hasType {
		// Type is optional in some cases, but we'll warn about it
		ctx.Logger.WithField("request_id", ctx.RequestID).
			WithField("property", propName).
			WithField("depth", depth).
			Warn("Property missing type definition")
		return nil
	}
	
	typeStr, ok := propType.(string)
	if !ok {
		return fmt.Errorf("type must be a string")
	}
	
	// Validate against allowed types
	validTypes := []string{"string", "number", "integer", "boolean", "array", "object", "null"}
	isValidType := false
	for _, validType := range validTypes {
		if typeStr == validType {
			isValidType = true
			break
		}
	}
	
	if !isValidType {
		return fmt.Errorf("invalid type: %s (allowed: %v)", typeStr, validTypes)
	}
	
	// Enhanced validation for complex types
	switch typeStr {
	case "array":
		if err := c.validateArraySchema(propName, propDefMap, depth, ctx); err != nil {
			return fmt.Errorf("array validation failed: %w", err)
		}
		
	case "object":
		if err := c.validateObjectSchema(propName, propDefMap, depth, ctx); err != nil {
			return fmt.Errorf("object validation failed: %w", err)
		}
		
	case "string":
		if err := c.validateStringSchema(propName, propDefMap, ctx); err != nil {
			return fmt.Errorf("string validation failed: %w", err)
		}
		
	case "number", "integer":
		if err := c.validateNumberSchema(propName, propDefMap, ctx); err != nil {
			return fmt.Errorf("number validation failed: %w", err)
		}
	}
	
	return nil
}

// validateArraySchema validates array type schemas with support for nested items
func (c *EnhancedConverter) validateArraySchema(propName string, schema map[string]interface{}, depth int, ctx *ToolConversionContext) error {
	// Validate items schema if present
	if items, ok := schema["items"]; ok {
		if err := c.validateSchemaProperty(propName+"[items]", items, depth+1, ctx); err != nil {
			return fmt.Errorf("items schema invalid: %w", err)
		}
	}
	
	// Validate array constraints
	if minItems, ok := schema["minItems"]; ok {
		if minItemsNum, ok := minItems.(float64); ok {
			if minItemsNum < 0 {
				return fmt.Errorf("minItems cannot be negative")
			}
		} else {
			return fmt.Errorf("minItems must be a number")
		}
	}
	
	if maxItems, ok := schema["maxItems"]; ok {
		if maxItemsNum, ok := maxItems.(float64); ok {
			if maxItemsNum < 0 {
				return fmt.Errorf("maxItems cannot be negative")
			}
		} else {
			return fmt.Errorf("maxItems must be a number")
		}
	}
	
	return nil
}

// validateObjectSchema validates object type schemas with support for nested properties
func (c *EnhancedConverter) validateObjectSchema(propName string, schema map[string]interface{}, depth int, ctx *ToolConversionContext) error {
	// Validate nested properties if present
	if properties, ok := schema["properties"]; ok {
		propsMap, ok := properties.(map[string]interface{})
		if !ok {
			return fmt.Errorf("properties must be an object")
		}
		
		// Recursively validate nested properties
		for nestedPropName, nestedPropDef := range propsMap {
			fullPropName := fmt.Sprintf("%s.%s", propName, nestedPropName)
			if err := c.validateSchemaProperty(fullPropName, nestedPropDef, depth+1, ctx); err != nil {
				return fmt.Errorf("nested property '%s': %w", nestedPropName, err)
			}
		}
	}
	
	// Validate required array for nested object
	if required, ok := schema["required"]; ok {
		requiredArray, ok := required.([]interface{})
		if !ok {
			return fmt.Errorf("required must be an array")
		}
		
		for i, req := range requiredArray {
			if _, ok := req.(string); !ok {
				return fmt.Errorf("required[%d] must be a string", i)
			}
		}
	}
	
	// Validate additionalProperties if present
	if additionalProps, ok := schema["additionalProperties"]; ok {
		switch additionalProps := additionalProps.(type) {
		case bool:
			// Boolean value is valid
		case map[string]interface{}:
			// Schema object for additional properties
			if err := c.validateSchemaProperty(propName+"[additionalProperties]", additionalProps, depth+1, ctx); err != nil {
				return fmt.Errorf("additionalProperties schema invalid: %w", err)
			}
		default:
			return fmt.Errorf("additionalProperties must be boolean or schema object")
		}
	}
	
	return nil
}

// validateStringSchema validates string type schemas with format and constraint validation
func (c *EnhancedConverter) validateStringSchema(propName string, schema map[string]interface{}, ctx *ToolConversionContext) error {
	// Validate string constraints
	if minLength, ok := schema["minLength"]; ok {
		if minLengthNum, ok := minLength.(float64); ok {
			if minLengthNum < 0 {
				return fmt.Errorf("minLength cannot be negative")
			}
		} else {
			return fmt.Errorf("minLength must be a number")
		}
	}
	
	if maxLength, ok := schema["maxLength"]; ok {
		if maxLengthNum, ok := maxLength.(float64); ok {
			if maxLengthNum < 0 {
				return fmt.Errorf("maxLength cannot be negative")
			}
		} else {
			return fmt.Errorf("maxLength must be a number")
		}
	}
	
	// Validate string format if present
	if format, ok := schema["format"]; ok {
		formatStr, ok := format.(string)
		if !ok {
			return fmt.Errorf("format must be a string")
		}
		
		// Common string formats
		validFormats := []string{
			"date-time", "date", "time", "email", "hostname", "ipv4", "ipv6", "uri", "uuid",
		}
		
		isValidFormat := false
		for _, validFormat := range validFormats {
			if formatStr == validFormat {
				isValidFormat = true
				break
			}
		}
		
		if !isValidFormat {
			ctx.Logger.WithField("request_id", ctx.RequestID).
				WithField("property", propName).
				WithField("format", formatStr).
				Warn("Unknown string format specified")
		}
	}
	
	// Validate enum values if present
	if enum, ok := schema["enum"]; ok {
		enumArray, ok := enum.([]interface{})
		if !ok {
			return fmt.Errorf("enum must be an array")
		}
		
		if len(enumArray) == 0 {
			return fmt.Errorf("enum array cannot be empty")
		}
		
		// All enum values should be strings for string type
		for i, enumValue := range enumArray {
			if _, ok := enumValue.(string); !ok {
				return fmt.Errorf("enum[%d] must be a string for string type", i)
			}
		}
	}
	
	return nil
}

// validateNumberSchema validates number/integer type schemas with range validation
func (c *EnhancedConverter) validateNumberSchema(propName string, schema map[string]interface{}, ctx *ToolConversionContext) error {
	// Validate number constraints - handle both int and float64 types
	if minimum, ok := schema["minimum"]; ok {
		if !isNumber(minimum) {
			return fmt.Errorf("minimum must be a number")
		}
	}
	
	if maximum, ok := schema["maximum"]; ok {
		if !isNumber(maximum) {
			return fmt.Errorf("maximum must be a number")
		}
	}
	
	if exclusiveMinimum, ok := schema["exclusiveMinimum"]; ok {
		if !isNumber(exclusiveMinimum) {
			return fmt.Errorf("exclusiveMinimum must be a number")
		}
	}
	
	if exclusiveMaximum, ok := schema["exclusiveMaximum"]; ok {
		if !isNumber(exclusiveMaximum) {
			return fmt.Errorf("exclusiveMaximum must be a number")
		}
	}
	
	if multipleOf, ok := schema["multipleOf"]; ok {
		if !isNumber(multipleOf) {
			return fmt.Errorf("multipleOf must be a number")
		}
		
		// Convert to float64 for comparison
		var multipleOfNum float64
		switch v := multipleOf.(type) {
		case float64:
			multipleOfNum = v
		case int:
			multipleOfNum = float64(v)
		case int64:
			multipleOfNum = float64(v)
		}
		
		if multipleOfNum <= 0 {
			return fmt.Errorf("multipleOf must be greater than 0")
		}
	}
	
	return nil
}

// isNumber checks if a value is a numeric type (int, int64, float64)
func isNumber(v interface{}) bool {
	switch v.(type) {
	case int, int64, float64:
		return true
	default:
		return false
	}
}

// convertOpenAIToolCallsToAnthropicEnhanced converts OpenAI tool calls to Anthropic format with enhanced error handling
func (c *EnhancedConverter) convertOpenAIToolCallsToAnthropicEnhanced(toolCalls []models.ToolCall, ctx *ToolConversionContext) ([]models.Content, error) {
	if len(toolCalls) == 0 {
		return nil, nil
	}
	
	content := make([]models.Content, len(toolCalls))
	
	for i, call := range toolCalls {
		// Enhanced JSON parsing with better error handling
		var input map[string]interface{}
		if call.Function.Arguments == "" {
			// Handle empty arguments
			input = make(map[string]interface{})
		} else {
			if err := json.Unmarshal([]byte(call.Function.Arguments), &input); err != nil {
				ctx.Logger.WithField("request_id", ctx.RequestID).
					WithField("tool_call_id", call.ID).
					WithField("function_name", call.Function.Name).
					WithField("arguments", call.Function.Arguments).
					WithError(err).
					Error("Failed to parse tool call arguments")
				return nil, fmt.Errorf("tool call %d (%s): invalid arguments JSON: %w", i, call.Function.Name, err)
			}
		}
		
		// Validate tool call structure
		if call.ID == "" {
			return nil, fmt.Errorf("tool call %d: missing ID", i)
		}
		
		if call.Function.Name == "" {
			return nil, fmt.Errorf("tool call %d: missing function name", i)
		}
		
		content[i] = models.Content{
			Type:  "tool_use",
			ID:    call.ID,
			Name:  call.Function.Name,
			Input: input,
		}
		
		// Log successful conversion with sanitized data
		ctx.Logger.WithField("request_id", ctx.RequestID).
			WithField("tool_name", call.Function.Name).
			WithField("tool_id", call.ID).
			WithField("input_keys", c.getMapKeys(input)).
			Info("Successfully converted tool call to Anthropic format")
	}
	
	return content, nil
}

// getMapKeys returns the keys of a map for logging (helps with debugging without exposing sensitive data)
func (c *EnhancedConverter) getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// ConvertAnthropicToolsToOpenAI converts Anthropic tools to OpenAI format while maintaining exact tool schema structure
func ConvertAnthropicToolsToOpenAI(tools []models.Tool) ([]models.ChatCompletionTool, error) {
	if len(tools) == 0 {
		return nil, nil
	}
	
	converted := make([]models.ChatCompletionTool, len(tools))
	for i, tool := range tools {
		// Validate tool before conversion
		if err := validateToolDefinition(tool); err != nil {
			return nil, fmt.Errorf("tool %d (%s): %w", i, tool.Name, err)
		}
		
		// Get description, handling nil pointer
		description := ""
		if tool.Description != nil {
			description = *tool.Description
		}
		
		// Convert while preserving exact schema structure
		converted[i] = models.ChatCompletionTool{
			Type: "function",
			Function: models.ChatCompletionToolFunction{
				Name:        tool.Name,
				Description: description,
				Parameters:  tool.InputSchema, // Preserve exact schema structure
			},
		}
	}
	
	return converted, nil
}

// validateToolDefinition validates a tool definition structure
func validateToolDefinition(tool models.Tool) error {
	if tool.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	
	if tool.InputSchema == nil {
		return fmt.Errorf("tool input_schema cannot be nil")
	}
	
	// Validate that input_schema has required structure
	if schemaType, ok := tool.InputSchema["type"]; ok {
		if schemaType != "object" {
			return fmt.Errorf("tool input_schema type must be 'object', got: %v", schemaType)
		}
	} else {
		return fmt.Errorf("tool input_schema must have 'type' field")
	}
	
	// Validate properties if present
	if properties, ok := tool.InputSchema["properties"]; ok {
		if _, ok := properties.(map[string]interface{}); !ok {
			return fmt.Errorf("tool input_schema properties must be an object")
		}
	}
	
	// Validate required array if present
	if required, ok := tool.InputSchema["required"]; ok {
		if requiredArray, ok := required.([]interface{}); ok {
			for i, req := range requiredArray {
				if _, ok := req.(string); !ok {
					return fmt.Errorf("tool input_schema required[%d] must be a string", i)
				}
			}
		} else {
			return fmt.Errorf("tool input_schema required must be an array")
		}
	}
	
	return nil
}

// ToolConversionError represents a structured error for tool conversion failures
type ToolConversionError struct {
	Type        string                 `json:"type"`
	Message     string                 `json:"message"`
	ToolName    string                 `json:"tool_name,omitempty"`
	ToolID      string                 `json:"tool_id,omitempty"`
	Field       string                 `json:"field,omitempty"`
	Value       interface{}            `json:"value,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
}

func (e *ToolConversionError) Error() string {
	if e.ToolName != "" {
		return fmt.Sprintf("tool conversion error for '%s': %s", e.ToolName, e.Message)
	}
	return fmt.Sprintf("tool conversion error: %s", e.Message)
}

// NewToolConversionError creates a new structured tool conversion error
func NewToolConversionError(errorType, message string) *ToolConversionError {
	return &ToolConversionError{
		Type:    errorType,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// WithTool adds tool information to the error
func (e *ToolConversionError) WithTool(name, id string) *ToolConversionError {
	e.ToolName = name
	e.ToolID = id
	return e
}

// WithField adds field information to the error
func (e *ToolConversionError) WithField(field string, value interface{}) *ToolConversionError {
	e.Field = field
	e.Value = value
	return e
}

// WithContext adds context information to the error
func (e *ToolConversionError) WithContext(key string, value interface{}) *ToolConversionError {
	e.Context[key] = value
	return e
}

// WithSuggestions adds suggestions to fix the error
func (e *ToolConversionError) WithSuggestions(suggestions ...string) *ToolConversionError {
	e.Suggestions = suggestions
	return e
}

// ConvertWithFallback converts a request with fallback behavior for providers that don't support tools
func (c *EnhancedConverter) ConvertWithFallback(ctx context.Context, req *models.MessagesRequest, providerSupportsTools bool) (*models.ChatCompletionRequest, error) {
	if !providerSupportsTools && len(req.Tools) > 0 {
		// Create a copy of the request without tools and convert tools to text descriptions
		fallbackReq := *req
		fallbackReq.Tools = nil
		fallbackReq.ToolChoice = nil
		
		// Add tool descriptions to the system message or create one
		toolDescriptions := c.generateToolDescriptions(req.Tools)
		if toolDescriptions != "" {
			systemMessage := models.Message{
				Role:    "system",
				Content: fmt.Sprintf("Available tools (for reference only, cannot be called directly):\n%s", toolDescriptions),
			}
			
			// Prepend system message or merge with existing system message
			if len(fallbackReq.Messages) > 0 && fallbackReq.Messages[0].Role == "system" {
				// Merge with existing system message
				existingContent := ""
				if content, ok := fallbackReq.Messages[0].Content.(string); ok {
					existingContent = content
				}
				fallbackReq.Messages[0].Content = fmt.Sprintf("%s\n\n%s", existingContent, systemMessage.Content)
			} else {
				// Prepend new system message
				fallbackReq.Messages = append([]models.Message{systemMessage}, fallbackReq.Messages...)
			}
		}
		
		c.logger.WithField("tool_count", len(req.Tools)).
			Warn("Provider doesn't support tools, converting to text descriptions")
		
		return c.ConvertRequest(ctx, &fallbackReq)
	}
	
	return c.ConvertRequest(ctx, req)
}

// generateToolDescriptions generates text descriptions of tools for fallback behavior
func (c *EnhancedConverter) generateToolDescriptions(tools []models.Tool) string {
	if len(tools) == 0 {
		return ""
	}
	
	var descriptions []string
	for _, tool := range tools {
		desc := fmt.Sprintf("- %s", tool.Name)
		if tool.Description != nil && *tool.Description != "" {
			desc += fmt.Sprintf(": %s", *tool.Description)
		}
		
		// Add parameter information
		if tool.InputSchema != nil {
			if properties, ok := tool.InputSchema["properties"]; ok {
				if propsMap, ok := properties.(map[string]interface{}); ok {
					var params []string
					for paramName := range propsMap {
						params = append(params, paramName)
					}
					if len(params) > 0 {
						desc += fmt.Sprintf(" (parameters: %s)", strings.Join(params, ", "))
					}
				}
			}
		}
		
		descriptions = append(descriptions, desc)
	}
	
	return strings.Join(descriptions, "\n")
}

// ValidateToolCallArguments validates tool call arguments against the tool schema
func ValidateToolCallArguments(toolCall models.ToolCall, toolSchema models.Tool) error {
	if toolCall.Function.Arguments == "" {
		// Check if the tool requires any parameters
		if required, ok := toolSchema.InputSchema["required"]; ok {
			if requiredArray, ok := required.([]interface{}); ok && len(requiredArray) > 0 {
				return NewToolConversionError("validation_error", "tool call has empty arguments but tool requires parameters").
					WithTool(toolCall.Function.Name, toolCall.ID).
					WithField("required_parameters", requiredArray).
					WithSuggestions("Provide the required parameters in the function arguments")
			}
		}
		return nil
	}
	
	// Parse arguments
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return NewToolConversionError("json_error", "invalid JSON in tool call arguments").
			WithTool(toolCall.Function.Name, toolCall.ID).
			WithField("arguments", toolCall.Function.Arguments).
			WithContext("parse_error", err.Error()).
			WithSuggestions("Ensure the arguments are valid JSON", "Check for missing quotes or commas")
	}
	
	// Validate against schema
	if err := validateArgumentsAgainstSchema(args, toolSchema.InputSchema, toolCall.Function.Name); err != nil {
		return err
	}
	
	return nil
}

// validateArgumentsAgainstSchema validates arguments against a JSON schema
func validateArgumentsAgainstSchema(args map[string]interface{}, schema map[string]interface{}, toolName string) error {
	// Check required fields
	if required, ok := schema["required"]; ok {
		if requiredArray, ok := required.([]interface{}); ok {
			for _, reqField := range requiredArray {
				if reqFieldStr, ok := reqField.(string); ok {
					if _, exists := args[reqFieldStr]; !exists {
						return NewToolConversionError("validation_error", fmt.Sprintf("missing required parameter '%s'", reqFieldStr)).
							WithTool(toolName, "").
							WithField("missing_parameter", reqFieldStr).
							WithSuggestions(fmt.Sprintf("Add the '%s' parameter to the tool call", reqFieldStr))
					}
				}
			}
		}
	}
	
	// Validate properties
	if properties, ok := schema["properties"]; ok {
		if propsMap, ok := properties.(map[string]interface{}); ok {
			for argName, argValue := range args {
				if propSchema, exists := propsMap[argName]; exists {
					if err := validateArgumentValue(argName, argValue, propSchema, toolName); err != nil {
						return err
					}
				} else {
					// Check if additional properties are allowed
					if additionalProps, ok := schema["additionalProperties"]; ok {
						if allowed, ok := additionalProps.(bool); ok && !allowed {
							return NewToolConversionError("validation_error", fmt.Sprintf("unexpected parameter '%s'", argName)).
								WithTool(toolName, "").
								WithField("unexpected_parameter", argName).
								WithSuggestions("Remove the unexpected parameter", "Check the tool schema for allowed parameters")
						}
					}
				}
			}
		}
	}
	
	return nil
}

// validateArgumentValue validates a single argument value against its schema
func validateArgumentValue(argName string, argValue interface{}, propSchema interface{}, toolName string) error {
	propSchemaMap, ok := propSchema.(map[string]interface{})
	if !ok {
		return nil // Skip validation if schema is not a map
	}
	
	expectedType, hasType := propSchemaMap["type"]
	if !hasType {
		return nil // Skip validation if no type specified
	}
	
	expectedTypeStr, ok := expectedType.(string)
	if !ok {
		return nil
	}
	
	// Validate type
	switch expectedTypeStr {
	case "string":
		if _, ok := argValue.(string); !ok {
			return NewToolConversionError("type_error", fmt.Sprintf("parameter '%s' must be a string", argName)).
				WithTool(toolName, "").
				WithField("parameter", argName).
				WithField("expected_type", "string").
				WithField("actual_value", argValue).
				WithSuggestions("Ensure the parameter value is a string")
		}
		
	case "number":
		switch argValue.(type) {
		case float64, int, int64:
			// Valid number types
		default:
			return NewToolConversionError("type_error", fmt.Sprintf("parameter '%s' must be a number", argName)).
				WithTool(toolName, "").
				WithField("parameter", argName).
				WithField("expected_type", "number").
				WithField("actual_value", argValue).
				WithSuggestions("Ensure the parameter value is a number")
		}
		
	case "integer":
		switch v := argValue.(type) {
		case float64:
			if v != float64(int64(v)) {
				return NewToolConversionError("type_error", fmt.Sprintf("parameter '%s' must be an integer", argName)).
					WithTool(toolName, "").
					WithField("parameter", argName).
					WithField("expected_type", "integer").
					WithField("actual_value", argValue).
					WithSuggestions("Ensure the parameter value is a whole number")
			}
		case int, int64:
			// Valid integer types
		default:
			return NewToolConversionError("type_error", fmt.Sprintf("parameter '%s' must be an integer", argName)).
				WithTool(toolName, "").
				WithField("parameter", argName).
				WithField("expected_type", "integer").
				WithField("actual_value", argValue).
				WithSuggestions("Ensure the parameter value is a whole number")
		}
		
	case "boolean":
		if _, ok := argValue.(bool); !ok {
			return NewToolConversionError("type_error", fmt.Sprintf("parameter '%s' must be a boolean", argName)).
				WithTool(toolName, "").
				WithField("parameter", argName).
				WithField("expected_type", "boolean").
				WithField("actual_value", argValue).
				WithSuggestions("Ensure the parameter value is true or false")
		}
		
	case "array":
		if _, ok := argValue.([]interface{}); !ok {
			return NewToolConversionError("type_error", fmt.Sprintf("parameter '%s' must be an array", argName)).
				WithTool(toolName, "").
				WithField("parameter", argName).
				WithField("expected_type", "array").
				WithField("actual_value", argValue).
				WithSuggestions("Ensure the parameter value is an array")
		}
		
	case "object":
		if _, ok := argValue.(map[string]interface{}); !ok {
			return NewToolConversionError("type_error", fmt.Sprintf("parameter '%s' must be an object", argName)).
				WithTool(toolName, "").
				WithField("parameter", argName).
				WithField("expected_type", "object").
				WithField("actual_value", argValue).
				WithSuggestions("Ensure the parameter value is an object")
		}
	}
	
	return nil
}

// ConvertOpenAIToolCallsToAnthropic converts OpenAI tool calls to Anthropic format while preserving tool call semantics
func ConvertOpenAIToolCallsToAnthropic(toolCalls []models.ToolCall) ([]models.Content, error) {
	if len(toolCalls) == 0 {
		return nil, nil
	}
	
	content := make([]models.Content, len(toolCalls))
	
	for i, call := range toolCalls {
		// Validate tool call structure first
		if err := validateToolCall(call); err != nil {
			return nil, fmt.Errorf("tool call %d: %w", i, err)
		}
		
		// Enhanced JSON parsing with proper error handling
		var input map[string]interface{}
		if call.Function.Arguments == "" {
			// Handle empty arguments gracefully
			input = make(map[string]interface{})
		} else {
			if err := json.Unmarshal([]byte(call.Function.Arguments), &input); err != nil {
				return nil, fmt.Errorf("tool call %d (%s): invalid arguments JSON: %w", i, call.Function.Name, err)
			}
		}
		
		// Preserve tool call order and correlation IDs correctly
		content[i] = models.Content{
			Type:  "tool_use",
			ID:    call.ID, // Maintain correlation ID
			Name:  call.Function.Name,
			Input: input,
		}
	}
	
	return content, nil
}

// validateToolCall validates the structure of an OpenAI tool call
func validateToolCall(call models.ToolCall) error {
	if call.ID == "" {
		return fmt.Errorf("tool call ID cannot be empty")
	}
	
	if call.Type != "function" {
		return fmt.Errorf("unsupported tool call type: %s (expected 'function')", call.Type)
	}
	
	if call.Function.Name == "" {
		return fmt.Errorf("function name cannot be empty")
	}
	
	// Validate that arguments is valid JSON if not empty
	if call.Function.Arguments != "" {
		var temp interface{}
		if err := json.Unmarshal([]byte(call.Function.Arguments), &temp); err != nil {
			return fmt.Errorf("function arguments must be valid JSON: %w", err)
		}
	}
	
	return nil
}

// ConvertAnthropicToOpenAI converts Anthropic messages request to OpenAI format
// Deprecated: Use EnhancedConverter.ConvertRequest instead for better error handling and validation
func ConvertAnthropicToOpenAI(req *models.MessagesRequest) (*models.ChatCompletionRequest, error) {
	// Use the enhanced converter for better semantic preservation
	converter := NewEnhancedConverter(nil, nil)
	return converter.ConvertRequest(context.Background(), req)
}

// ConvertOpenAIToAnthropic converts OpenAI response to Anthropic format
// Deprecated: Use EnhancedConverter.ConvertResponse instead for better error handling and validation
func ConvertOpenAIToAnthropic(
	resp *models.ChatCompletionResponse,
	requestID, providerName string,
) (*models.MessagesResponse, error) {
	// Use the enhanced converter for better semantic preservation and error handling
	converter := NewEnhancedConverter(nil, nil)
	return converter.ConvertResponse(context.Background(), resp, requestID, providerName)
}

// Legacy functions - kept for backward compatibility but deprecated

// convertMessages converts Anthropic messages to OpenAI format
// Deprecated: This function converts tool calls to plain text, losing semantic information.
// Use EnhancedConverter.convertMessagesEnhanced instead.
func convertMessages(messages []models.Message) ([]models.ChatMessage, error) {
	// Use enhanced converter for better semantic preservation
	converter := NewEnhancedConverter(nil, nil)
	ctx := &ToolConversionContext{
		RequestID:    fmt.Sprintf("legacy_%d", time.Now().UnixNano()),
		ProviderName: "unknown",
		Capabilities: converter.SupportedFeatures(),
		Logger:       converter.logger,
		StartTime:    time.Now(),
		Metadata:     make(map[string]interface{}),
	}
	
	return converter.convertMessagesEnhanced(messages, ctx)
}

// convertTools converts Anthropic tools to OpenAI format
// Deprecated: Use ConvertAnthropicToolsToOpenAI for better validation and error handling
func convertTools(tools []models.Tool) []models.ChatCompletionTool {
	converted, err := ConvertAnthropicToolsToOpenAI(tools)
	if err != nil {
		// For backward compatibility, return empty slice on error
		return []models.ChatCompletionTool{}
	}
	return converted
}

// mustMarshalJSON marshals to JSON or returns empty string on error
// Deprecated: This function silently ignores errors. Use proper error handling instead.
func mustMarshalJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(data)
}
