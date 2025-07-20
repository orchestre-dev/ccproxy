package testing

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// TestHelpers provides additional test helper functions
type TestHelpers struct{}

// NewTestHelpers creates a new test helpers instance
func NewTestHelpers() *TestHelpers {
	return &TestHelpers{}
}

// GenerateTestData generates test data of specified size
func (th *TestHelpers) GenerateTestData(size int) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}
	return data
}

// CreateTestRequest creates a test HTTP request
func (th *TestHelpers) CreateTestRequest(method, path string, body interface{}) *http.Request {
	var bodyReader io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(data)
	}
	
	req, _ := http.NewRequest(method, path, bodyReader)
	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	return req
}