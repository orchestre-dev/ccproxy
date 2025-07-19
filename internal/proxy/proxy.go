package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/musistudio/ccproxy/internal/utils"
)

// Config represents proxy configuration
type Config struct {
	URL      string   // Proxy URL (http://proxy.example.com:8080)
	NoProxy  []string // List of hosts to bypass proxy
	Username string   // Proxy username (optional)
	Password string   // Proxy password (optional)
}

// NewConfig creates a new proxy configuration
func NewConfig(proxyURL string) (*Config, error) {
	if proxyURL == "" {
		return nil, nil
	}

	// Parse proxy URL
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}

	config := &Config{
		URL: proxyURL,
	}

	// Extract username and password if present
	if parsedURL.User != nil {
		config.Username = parsedURL.User.Username()
		config.Password, _ = parsedURL.User.Password()
	}

	// Get no-proxy list from environment
	noProxy := os.Getenv("NO_PROXY")
	if noProxy == "" {
		noProxy = os.Getenv("no_proxy")
	}
	
	if noProxy != "" {
		config.NoProxy = parseNoProxy(noProxy)
	}

	return config, nil
}

// CreateHTTPClient creates an HTTP client with proxy configuration
func CreateHTTPClient(proxyConfig *Config, timeout time.Duration) (*http.Client, error) {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment, // Default to environment
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	// Configure proxy if provided
	if proxyConfig != nil && proxyConfig.URL != "" {
		proxyURL, err := url.Parse(proxyConfig.URL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}

		// Create proxy function that respects no-proxy list
		transport.Proxy = createProxyFunc(proxyURL, proxyConfig.NoProxy)
		
		utils.GetLogger().Infof("Using corporate proxy: %s", sanitizeProxyURL(proxyConfig.URL))
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	return client, nil
}

// createProxyFunc creates a proxy function that respects the no-proxy list
func createProxyFunc(proxyURL *url.URL, noProxy []string) func(*http.Request) (*url.URL, error) {
	return func(req *http.Request) (*url.URL, error) {
		// Check if host should bypass proxy
		if shouldBypassProxy(req.URL.Host, noProxy) {
			return nil, nil
		}
		return proxyURL, nil
	}
}

// shouldBypassProxy checks if a host should bypass the proxy
func shouldBypassProxy(host string, noProxy []string) bool {
	if len(noProxy) == 0 {
		return false
	}

	// Remove port from host for comparison
	hostname := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		hostname = h
	}

	for _, pattern := range noProxy {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}

		// Handle wildcard patterns
		if strings.HasPrefix(pattern, "*") {
			suffix := strings.TrimPrefix(pattern, "*")
			if strings.HasSuffix(hostname, suffix) {
				return true
			}
		} else if strings.HasPrefix(pattern, ".") {
			// .example.com matches sub.example.com but not example.com
			if strings.HasSuffix(hostname, pattern) {
				return true
			}
		} else {
			// Exact match or suffix match
			if hostname == pattern || strings.HasSuffix(hostname, "."+pattern) {
				return true
			}
		}

		// Check for IP addresses
		if net.ParseIP(hostname) != nil && hostname == pattern {
			return true
		}

		// Check for CIDR ranges
		if strings.Contains(pattern, "/") {
			_, cidr, err := net.ParseCIDR(pattern)
			if err == nil {
				if ip := net.ParseIP(hostname); ip != nil && cidr.Contains(ip) {
					return true
				}
			}
		}
	}

	return false
}

// parseNoProxy parses the NO_PROXY environment variable
func parseNoProxy(noProxy string) []string {
	var result []string
	for _, host := range strings.Split(noProxy, ",") {
		host = strings.TrimSpace(host)
		if host != "" {
			result = append(result, host)
		}
	}
	return result
}

// sanitizeProxyURL removes credentials from proxy URL for logging
func sanitizeProxyURL(proxyURL string) string {
	u, err := url.Parse(proxyURL)
	if err != nil {
		// Return original string if parsing fails
		return proxyURL
	}
	
	// Remove user info
	u.User = nil
	return u.String()
}

// GetProxyFromEnvironment gets proxy configuration from environment variables
func GetProxyFromEnvironment() *Config {
	// Check various proxy environment variables
	proxyVars := []string{
		"HTTPS_PROXY",
		"https_proxy",
		"HTTP_PROXY",
		"http_proxy",
		"ALL_PROXY",
		"all_proxy",
	}

	var proxyURL string
	for _, v := range proxyVars {
		if proxy := os.Getenv(v); proxy != "" {
			proxyURL = proxy
			break
		}
	}

	if proxyURL == "" {
		return nil
	}

	config, err := NewConfig(proxyURL)
	if err != nil {
		utils.GetLogger().Warnf("Invalid proxy configuration: %v", err)
		return nil
	}

	return config
}

// ValidateProxy tests if the proxy is working
func ValidateProxy(client *http.Client, testURL string) error {
	if testURL == "" {
		testURL = "https://api.anthropic.com/health"
	}

	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}

	// Set a shorter timeout for validation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("proxy validation failed: %w", err)
	}
	defer resp.Body.Close()

	// We don't care about the status code, just that we can connect
	utils.GetLogger().Infof("Proxy validation successful, connected to %s", testURL)
	return nil
}