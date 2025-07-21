package utils

import (
	"testing"
	"time"
)

func TestToJSONString(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: "{}",
		},
		{
			name:     "empty map",
			input:    map[string]interface{}{},
			expected: "{}",
		},
		{
			name:     "simple map",
			input:    map[string]interface{}{"key": "value"},
			expected: `{"key":"value"}`,
		},
		{
			name: "complex object",
			input: map[string]interface{}{
				"name": "test",
				"age":  30,
				"nested": map[string]interface{}{
					"field": "value",
				},
			},
			expected: `{"age":30,"name":"test","nested":{"field":"value"}}`,
		},
		{
			name:     "array",
			input:    []int{1, 2, 3},
			expected: "[1,2,3]",
		},
		{
			name:     "string",
			input:    "test string",
			expected: `"test string"`,
		},
		{
			name:     "number",
			input:    42,
			expected: "42",
		},
		{
			name:     "boolean",
			input:    true,
			expected: "true",
		},
		{
			name:     "invalid input (channel)",
			input:    make(chan int),
			expected: "{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToJSONString(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetTimestamp(t *testing.T) {
	// Test that GetTimestamp returns a value close to current time
	before := time.Now().Unix()
	timestamp := GetTimestamp()
	after := time.Now().Unix()
	
	if timestamp < before || timestamp > after {
		t.Errorf("timestamp %d is not between %d and %d", timestamp, before, after)
	}
	
	// Test that consecutive calls return incrementing or equal values
	ts1 := GetTimestamp()
	time.Sleep(10 * time.Millisecond)
	ts2 := GetTimestamp()
	
	if ts2 < ts1 {
		t.Error("second timestamp should be >= first timestamp")
	}
}