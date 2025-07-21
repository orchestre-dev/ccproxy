package proxy

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name        string
		proxyURL    string
		wantErr     bool
		wantNil     bool
		wantUser    string
		wantPass    string
	}{
		{
			name:     "empty URL returns nil",
			proxyURL: "",
			wantNil:  true,
		},
		{
			name:     "valid URL without auth",
			proxyURL: "http://proxy.example.com:8080",
		},
		{
			name:     "valid URL with auth",
			proxyURL: "http://user:pass@proxy.example.com:8080",
			wantUser: "user",
			wantPass: "pass",
		},
		{
			name:     "invalid URL",
			proxyURL: "://invalid",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewConfig(tt.proxyURL)
			
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if tt.wantNil {
				if config != nil {
					t.Error("Expected nil config")
				}
				return
			}
			
			if config == nil {
				t.Fatal("Expected non-nil config")
			}
			
			if config.URL != tt.proxyURL {
				t.Errorf("Expected URL %s, got %s", tt.proxyURL, config.URL)
			}
			
			if config.Username != tt.wantUser {
				t.Errorf("Expected username %s, got %s", tt.wantUser, config.Username)
			}
			
			if config.Password != tt.wantPass {
				t.Errorf("Expected password %s, got %s", tt.wantPass, config.Password)
			}
		})
	}
}

func TestParseNoProxy(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single host",
			input:    "localhost",
			expected: []string{"localhost"},
		},
		{
			name:     "multiple hosts",
			input:    "localhost,example.com,192.168.1.1",
			expected: []string{"localhost", "example.com", "192.168.1.1"},
		},
		{
			name:     "with spaces",
			input:    " localhost , example.com , 192.168.1.1 ",
			expected: []string{"localhost", "example.com", "192.168.1.1"},
		},
		{
			name:     "with wildcards",
			input:    "*.local,.example.com,10.0.0.0/8",
			expected: []string{"*.local", ".example.com", "10.0.0.0/8"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "only commas",
			input:    ",,,",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseNoProxy(tt.input)
			
			if len(result) != len(tt.expected) {
				t.Fatalf("Expected %d hosts, got %d", len(tt.expected), len(result))
			}
			
			for i, host := range result {
				if host != tt.expected[i] {
					t.Errorf("Expected host[%d] to be %s, got %s", i, tt.expected[i], host)
				}
			}
		})
	}
}

func TestShouldBypassProxy(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		noProxy  []string
		expected bool
	}{
		{
			name:     "empty no-proxy list",
			host:     "example.com",
			noProxy:  []string{},
			expected: false,
		},
		{
			name:     "exact match",
			host:     "localhost",
			noProxy:  []string{"localhost"},
			expected: true,
		},
		{
			name:     "exact match with port",
			host:     "localhost:8080",
			noProxy:  []string{"localhost"},
			expected: true,
		},
		{
			name:     "subdomain match with dot prefix",
			host:     "api.example.com",
			noProxy:  []string{".example.com"},
			expected: true,
		},
		{
			name:     "no match for domain with dot prefix",
			host:     "example.com",
			noProxy:  []string{".example.com"},
			expected: false,
		},
		{
			name:     "wildcard match",
			host:     "api.local",
			noProxy:  []string{"*.local"},
			expected: true,
		},
		{
			name:     "IP address match",
			host:     "192.168.1.1",
			noProxy:  []string{"192.168.1.1"},
			expected: true,
		},
		{
			name:     "CIDR range match",
			host:     "10.0.0.5",
			noProxy:  []string{"10.0.0.0/24"},
			expected: true,
		},
		{
			name:     "CIDR range no match",
			host:     "10.0.1.5",
			noProxy:  []string{"10.0.0.0/24"},
			expected: false,
		},
		{
			name:     "suffix match",
			host:     "api.example.com",
			noProxy:  []string{"example.com"},
			expected: true,
		},
		{
			name:     "no match",
			host:     "other.com",
			noProxy:  []string{"example.com"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldBypassProxy(tt.host, tt.noProxy)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestSanitizeProxyURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "URL without auth",
			input:    "http://proxy.example.com:8080",
			expected: "http://proxy.example.com:8080",
		},
		{
			name:     "URL with auth",
			input:    "http://user:pass@proxy.example.com:8080",
			expected: "http://proxy.example.com:8080",
		},
		{
			name:     "HTTPS URL with auth",
			input:    "https://user:pass@proxy.example.com:8080",
			expected: "https://proxy.example.com:8080",
		},
		{
			name:     "invalid URL",
			input:    "://invalid",
			expected: "://invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeProxyURL(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestCreateHTTPClient(t *testing.T) {
	tests := []struct {
		name        string
		proxyConfig *Config
		timeout     time.Duration
		wantErr     bool
	}{
		{
			name:        "no proxy config",
			proxyConfig: nil,
			timeout:     30 * time.Second,
		},
		{
			name: "with valid proxy config",
			proxyConfig: &Config{
				URL: "http://proxy.example.com:8080",
			},
			timeout: 30 * time.Second,
		},
		{
			name: "with invalid proxy URL",
			proxyConfig: &Config{
				URL: "://invalid",
			},
			timeout: 30 * time.Second,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := CreateHTTPClient(tt.proxyConfig, tt.timeout)
			
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if client == nil {
				t.Fatal("Expected non-nil client")
			}
			
			if client.Timeout != tt.timeout {
				t.Errorf("Expected timeout %v, got %v", tt.timeout, client.Timeout)
			}
		})
	}
}

func TestGetProxyFromEnvironment(t *testing.T) {
	// Save original env vars
	oldHTTPSProxy := os.Getenv("HTTPS_PROXY")
	oldHTTPProxy := os.Getenv("HTTP_PROXY")
	oldNoProxy := os.Getenv("NO_PROXY")
	
	// Restore env vars after test
	defer func() {
		os.Setenv("HTTPS_PROXY", oldHTTPSProxy)
		os.Setenv("HTTP_PROXY", oldHTTPProxy)
		os.Setenv("NO_PROXY", oldNoProxy)
	}()
	
	tests := []struct {
		name      string
		setupEnv  func()
		wantNil   bool
		wantURL   string
		wantNoProxy []string
	}{
		{
			name: "HTTPS_PROXY set",
			setupEnv: func() {
				os.Setenv("HTTPS_PROXY", "https://proxy.example.com:8080")
				os.Setenv("NO_PROXY", "localhost,127.0.0.1")
			},
			wantURL: "https://proxy.example.com:8080",
			wantNoProxy: []string{"localhost", "127.0.0.1"},
		},
		{
			name: "HTTP_PROXY set",
			setupEnv: func() {
				os.Unsetenv("HTTPS_PROXY")
				os.Setenv("HTTP_PROXY", "http://proxy.example.com:3128")
			},
			wantURL: "http://proxy.example.com:3128",
		},
		{
			name: "no proxy set",
			setupEnv: func() {
				os.Unsetenv("HTTPS_PROXY")
				os.Unsetenv("HTTP_PROXY")
				os.Unsetenv("https_proxy")
				os.Unsetenv("http_proxy")
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all proxy env vars
			os.Unsetenv("HTTPS_PROXY")
			os.Unsetenv("https_proxy")
			os.Unsetenv("HTTP_PROXY")
			os.Unsetenv("http_proxy")
			os.Unsetenv("NO_PROXY")
			os.Unsetenv("no_proxy")
			
			// Setup test environment
			tt.setupEnv()
			
			config := GetProxyFromEnvironment()
			
			if tt.wantNil {
				if config != nil {
					t.Error("Expected nil config")
				}
				return
			}
			
			if config == nil {
				t.Fatal("Expected non-nil config")
			}
			
			if config.URL != tt.wantURL {
				t.Errorf("Expected URL %s, got %s", tt.wantURL, config.URL)
			}
			
			if len(tt.wantNoProxy) > 0 {
				if len(config.NoProxy) != len(tt.wantNoProxy) {
					t.Errorf("Expected %d no-proxy entries, got %d", len(tt.wantNoProxy), len(config.NoProxy))
				}
			}
		})
	}
}

func TestValidateProxy(t *testing.T) {
	// Create a test server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer testServer.Close()

	// Create a client without proxy
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Test validation with test server
	err := ValidateProxy(client, testServer.URL)
	if err != nil {
		t.Errorf("Validation failed: %v", err)
	}

	// Test validation with unreachable URL
	err = ValidateProxy(client, "http://localhost:1")
	if err == nil {
		t.Error("Expected error for unreachable URL")
	}
}

func TestCreateProxyFunc(t *testing.T) {
	proxyURL, _ := url.Parse("http://proxy.example.com:8080")
	noProxy := []string{"localhost", ".internal.com"}
	
	proxyFunc := createProxyFunc(proxyURL, noProxy)
	
	tests := []struct {
		name        string
		requestURL  string
		expectProxy bool
	}{
		{
			name:        "should use proxy for external host",
			requestURL:  "https://api.anthropic.com/v1/messages",
			expectProxy: true,
		},
		{
			name:        "should bypass proxy for localhost",
			requestURL:  "http://localhost:8080/test",
			expectProxy: false,
		},
		{
			name:        "should bypass proxy for internal domain",
			requestURL:  "https://api.internal.com/test",
			expectProxy: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.requestURL, nil)
			resultURL, err := proxyFunc(req)
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if tt.expectProxy {
				if resultURL == nil {
					t.Error("Expected proxy URL but got nil")
				} else if resultURL.String() != proxyURL.String() {
					t.Errorf("Expected proxy URL %s, got %s", proxyURL.String(), resultURL.String())
				}
			} else {
				if resultURL != nil {
					t.Errorf("Expected nil proxy URL but got %s", resultURL.String())
				}
			}
		})
	}
}

func TestProxyWithAuth(t *testing.T) {
	// Create a test proxy server that checks authentication
	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Proxy-Authorization")
		if auth == "" {
			w.WriteHeader(http.StatusProxyAuthRequired)
			return
		}
		
		// Simple check for basic auth
		if !strings.HasPrefix(auth, "Basic ") {
			w.WriteHeader(http.StatusProxyAuthRequired)
			return
		}
		
		w.WriteHeader(http.StatusOK)
	}))
	defer proxyServer.Close()

	// Create proxy config with auth
	proxyConfig := &Config{
		URL:      fmt.Sprintf("http://testuser:testpass@%s", proxyServer.URL[7:]), // Remove http://
		Username: "testuser",
		Password: "testpass",
	}

	client, err := CreateHTTPClient(proxyConfig, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// The actual auth test would require a proper proxy setup
	// This test mainly verifies that the client is created correctly
	if client == nil {
		t.Error("Expected non-nil client")
	}
}