package proxy_test

import (
	"fmt"
	"log"
	"time"

	"github.com/orchestre-dev/ccproxy/internal/proxy"
)

func ExampleCreateHTTPClient() {
	// Example 1: Create client with explicit proxy configuration
	proxyConfig := &proxy.Config{
		URL:      "http://proxy.example.com:8080",
		Username: "user",
		Password: "pass",
		NoProxy:  []string{"localhost", "127.0.0.1", ".internal.com"},
	}

	client, err := proxy.CreateHTTPClient(proxyConfig, 30*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	// Use the client for HTTP requests
	_ = client

	// Example 2: Create client that uses environment proxy settings
	envProxyConfig := proxy.GetProxyFromEnvironment()
	client2, err := proxy.CreateHTTPClient(envProxyConfig, 30*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	// Use the client for HTTP requests
	_ = client2

	fmt.Println("HTTP clients created successfully")
	// Output: HTTP clients created successfully
}

func ExampleValidateProxy() {
	// Create a client with proxy configuration
	proxyConfig := &proxy.Config{
		URL: "http://proxy.example.com:8080",
	}

	client, err := proxy.CreateHTTPClient(proxyConfig, 30*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	// Validate proxy connectivity
	err = proxy.ValidateProxy(client, "https://api.anthropic.com/health")
	if err != nil {
		fmt.Printf("Proxy validation failed: %v\n", err)
	} else {
		fmt.Println("Proxy validation successful")
	}
}

func ExampleNewConfig() {
	// Create proxy configuration from URL
	config, err := proxy.NewConfig("http://user:pass@proxy.example.com:8080")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Proxy URL: %s\n", config.URL)
	fmt.Printf("Username: %s\n", config.Username)
	// Note: Password is not printed for security reasons

	// Create configuration from environment
	// This would read NO_PROXY environment variable automatically
	config2, _ := proxy.NewConfig("http://proxy.example.com:8080")
	if config2 != nil && len(config2.NoProxy) > 0 {
		fmt.Printf("No-proxy list: %v\n", config2.NoProxy)
	}
}