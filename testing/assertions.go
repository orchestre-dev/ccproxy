package testing

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Assertions provides enhanced assertion methods
type Assertions struct {
	t *testing.T
}

// NewAssertions creates a new assertions instance
func NewAssertions(t *testing.T) *Assertions {
	return &Assertions{t: t}
}

// AssertJSONEqual asserts that two JSON strings are equivalent
func (a *Assertions) AssertJSONEqual(expected, actual string, msgAndArgs ...interface{}) bool {
	a.t.Helper()
	
	var expectedObj, actualObj interface{}
	
	err := json.Unmarshal([]byte(expected), &expectedObj)
	require.NoError(a.t, err, "Failed to unmarshal expected JSON")
	
	err = json.Unmarshal([]byte(actual), &actualObj)
	require.NoError(a.t, err, "Failed to unmarshal actual JSON")
	
	return assert.Equal(a.t, expectedObj, actualObj, msgAndArgs...)
}

// AssertJSONPath asserts a value at a specific JSON path
func (a *Assertions) AssertJSONPath(jsonStr string, path string, expected interface{}, msgAndArgs ...interface{}) bool {
	a.t.Helper()
	
	var data interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	require.NoError(a.t, err, "Failed to unmarshal JSON")
	
	value := getJSONPath(data, path)
	return assert.Equal(a.t, expected, value, msgAndArgs...)
}

// getJSONPath extracts a value from a simple JSON path (e.g., "field.subfield")
func getJSONPath(data interface{}, path string) interface{} {
	if path == "" {
		return data
	}
	
	// Simple implementation - just handles dot notation
	if m, ok := data.(map[string]interface{}); ok {
		return m[path]
	}
	
	return nil
}