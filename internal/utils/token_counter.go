package utils

import (
	"strings"
)

// CountRequestTokens estimates token count for a request
// This is a simplified implementation - in production you would use tiktoken-go
func CountRequestTokens(bodyMap map[string]interface{}) int {
	// Simple estimation based on message content
	tokenCount := 0

	// Count tokens in messages
	if messages, ok := bodyMap["messages"].([]interface{}); ok {
		for _, msg := range messages {
			if msgMap, ok := msg.(map[string]interface{}); ok {
				if content, ok := msgMap["content"].(string); ok {
					// Rough estimation: 1 token per 4 characters
					tokenCount += len(content) / 4
				}
			}
		}
	}

	// Count tokens in system message
	if system, ok := bodyMap["system"].(string); ok {
		tokenCount += len(system) / 4
	}

	// Add base tokens for request structure
	tokenCount += 50

	return tokenCount
}

// CountResponseTokens estimates token count for a response
func CountResponseTokens(content string) int {
	// Remove common formatting
	content = strings.TrimSpace(content)

	// Rough estimation: 1 token per 4 characters
	return len(content) / 4
}
