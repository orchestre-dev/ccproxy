package utils

import (
	"testing"
	"time"

	testutil "github.com/orchestre-dev/ccproxy/internal/testing"
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
			name:     "empty string",
			input:    "",
			expected: `""`,
		},
		{
			name:     "simple string",
			input:    "hello",
			expected: `"hello"`,
		},
		{
			name:     "integer",
			input:    42,
			expected: "42",
		},
		{
			name:     "boolean true",
			input:    true,
			expected: "true",
		},
		{
			name:     "boolean false",
			input:    false,
			expected: "false",
		},
		{
			name:     "simple map",
			input:    map[string]string{"key": "value"},
			expected: `{"key":"value"}`,
		},
		{
			name:     "simple slice",
			input:    []int{1, 2, 3},
			expected: `[1,2,3]`,
		},
		{
			name:     "empty map",
			input:    map[string]string{},
			expected: `{}`,
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: `[]`,
		},
		{
			name: "nested structure",
			input: map[string]interface{}{
				"name": "test",
				"data": map[string]interface{}{
					"count": 5,
					"items": []string{"a", "b"},
				},
			},
			expected: `{"data":{"count":5,"items":["a","b"]},"name":"test"}`,
		},
		{
			name: "struct with json tags",
			input: struct {
				Name  string `json:"name"`
				Value int    `json:"value"`
			}{
				Name:  "test",
				Value: 123,
			},
			expected: `{"name":"test","value":123}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToJSONString(tt.input)
			testutil.AssertEqual(t, tt.expected, result, "ToJSONString result mismatch")
		})
	}
}

func TestToJSONStringWithUnmarshallableType(t *testing.T) {
	// Function type cannot be marshalled to JSON
	input := func() {}
	result := ToJSONString(input)
	testutil.AssertEqual(t, "{}", result, "Should return {} for unmarshallable types")
}

func TestToJSONStringWithComplexTypes(t *testing.T) {
	// Test with channels, which cannot be marshalled
	input := make(chan int)
	defer close(input)

	result := ToJSONString(input)
	testutil.AssertEqual(t, "{}", result, "Should return {} for channels")

	// Test with complex numbers
	complex := complex(1, 2)
	result = ToJSONString(complex)
	testutil.AssertEqual(t, "{}", result, "Should return {} for complex numbers")
}

func TestGetTimestamp(t *testing.T) {
	// Get current time before calling function
	before := time.Now().Unix()

	// Call the function
	timestamp := GetTimestamp()

	// Get current time after calling function
	after := time.Now().Unix()

	// Timestamp should be between before and after
	testutil.AssertTrue(t, timestamp >= before, "Timestamp should be >= time before call")
	testutil.AssertTrue(t, timestamp <= after, "Timestamp should be <= time after call")

	// Timestamp should be a valid Unix timestamp (positive number)
	testutil.AssertTrue(t, timestamp > 0, "Timestamp should be positive")
}

func TestGetTimestampConsistency(t *testing.T) {
	// Call function twice in quick succession
	timestamp1 := GetTimestamp()
	timestamp2 := GetTimestamp()

	// Second timestamp should be >= first (time moves forward)
	testutil.AssertTrue(t, timestamp2 >= timestamp1, "Second timestamp should be >= first")

	// Difference should be small (less than 1 second for this test)
	diff := timestamp2 - timestamp1
	testutil.AssertTrue(t, diff < 2, "Timestamps should be within 2 seconds of each other")
}

func TestGetTimestampPrecision(t *testing.T) {
	// Test that we get Unix timestamp (seconds since epoch)
	timestamp := GetTimestamp()

	// Unix timestamp should be in reasonable range
	// After 2020-01-01 (1577836800) and before 2040-01-01 (2208988800)
	testutil.AssertTrue(t, timestamp > 1577836800, "Timestamp should be after 2020")
	testutil.AssertTrue(t, timestamp < 2208988800, "Timestamp should be before 2040")
}

// Benchmark tests for performance verification
func BenchmarkToJSONString(b *testing.B) {
	testData := map[string]interface{}{
		"name":   "benchmark test",
		"count":  100,
		"active": true,
		"items":  []string{"a", "b", "c"},
		"nested": map[string]int{"x": 1, "y": 2},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ToJSONString(testData)
	}
}

func BenchmarkToJSONStringSimple(b *testing.B) {
	testData := "simple string"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ToJSONString(testData)
	}
}

func BenchmarkGetTimestamp(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetTimestamp()
	}
}
