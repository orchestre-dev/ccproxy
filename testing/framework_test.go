package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFrameworkBasics tests the basic functionality of the test framework
func TestFrameworkBasics(t *testing.T) {
	// Test fixture loading
	fixtures := NewFixtures()
	
	// Test getting a request fixture
	req, err := fixtures.GetRequest("anthropic_messages")
	assert.NoError(t, err)
	assert.NotNil(t, req)
	assert.Equal(t, "claude-3-sonnet-20240229", req["model"])
	
	// Test getting a response fixture
	resp, err := fixtures.GetResponse("anthropic_messages")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	
	// Test getting a provider fixture
	prov, err := fixtures.GetProvider("anthropic")
	assert.NoError(t, err)
	assert.NotNil(t, prov)
	assert.Equal(t, "anthropic", prov["type"])
}

// TestAssertions tests the assertion helpers
func TestAssertions(t *testing.T) {
	a := NewAssertions(t)
	
	// Test JSON equality
	json1 := `{"name": "test", "value": 123}`
	json2 := `{"value": 123, "name": "test"}`
	a.AssertJSONEqual(json1, json2)
	
	// Test JSON path
	a.AssertJSONPath(json1, "name", "test")
	a.AssertJSONPath(json1, "value", float64(123))
}

// TestMockServer tests the mock server functionality
func TestMockServer(t *testing.T) {
	// Create mock server
	mock := NewMockServer()
	defer mock.Close()
	
	// Add a route
	mock.AddRoute("GET", "/test", map[string]string{"status": "ok"}, 200)
	
	// Test the route
	url := mock.GetURL()
	assert.NotEmpty(t, url)
	
	// Record requests
	requests := mock.GetRequests()
	assert.Len(t, requests, 0)
}

// TestTestHelpers tests the test helper functions
func TestTestHelpers(t *testing.T) {
	helpers := NewTestHelpers()
	
	// Test data generation
	data := helpers.GenerateTestData(100)
	assert.Len(t, data, 100)
	
	// Test request creation
	req := helpers.CreateTestRequest("POST", "/test", map[string]string{"key": "value"})
	assert.NotNil(t, req)
	assert.Equal(t, "POST", req.Method)
	assert.Equal(t, "/test", req.URL.Path)
}