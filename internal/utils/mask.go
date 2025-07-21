package utils

import "strings"

// MaskAPIKey masks an API key for safe display
func MaskAPIKey(key string) string {
	if key == "" {
		return "(not set)"
	}

	if len(key) <= 8 {
		return strings.Repeat("*", len(key))
	}

	// Show first 4 and last 4 characters
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}

