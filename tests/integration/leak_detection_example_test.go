package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/orchestre-dev/ccproxy/internal/config"
	"github.com/orchestre-dev/ccproxy/internal/server"
	testfw "github.com/orchestre-dev/ccproxy/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServerWithGoroutineLeakDetection demonstrates goroutine leak detection
func TestServerWithGoroutineLeakDetection(t *testing.T) {
	testfw.WithLeakDetection(t, func() {
		// Create test framework
		tf := testfw.NewTestFramework(t)
		
		// Start server
		srv, err := tf.StartServerWithError(tf.GetConfig())
		require.NoError(t, err)
		defer srv.Shutdown()
		
		// Make some requests
		client := &http.Client{
			Timeout: 5 * time.Second,
		}
		
		for i := 0; i < 5; i++ {
			resp, err := client.Get(fmt.Sprintf("http://localhost:%d/health", tf.GetConfig().Port))
			require.NoError(t, err)
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
		
		// Give time for cleanup
		time.Sleep(100 * time.Millisecond)
	})
}

// TestServerWithResourceMonitoring demonstrates resource monitoring
func TestServerWithResourceMonitoring(t *testing.T) {
	testfw.WithResourceMonitoring(t, 100, func() {
		// Create test framework
		tf := testfw.NewTestFramework(t)
		
		// Start server
		srv, err := tf.StartServerWithError(tf.GetConfig())
		require.NoError(t, err)
		defer srv.Shutdown()
		
		// Generate some load
		client := &http.Client{
			Timeout: 5 * time.Second,
		}
		
		// Send large requests
		for i := 0; i < 10; i++ {
			largePayload := make([]byte, 1024*1024) // 1MB
			for j := range largePayload {
				largePayload[j] = byte(j % 256)
			}
			
			resp, err := client.Post(
				fmt.Sprintf("http://localhost:%d/v1/messages", tf.GetConfig().Port),
				"application/json",
				bytes.NewReader(largePayload),
			)
			if err == nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		}
		
		// Let GC run
		time.Sleep(200 * time.Millisecond)
	})
}

// TestServerWithComprehensiveLeakDetection demonstrates full leak detection
func TestServerWithComprehensiveLeakDetection(t *testing.T) {
	testfw.WithComprehensiveLeakDetection(t, 200, func(detector *testfw.LeakDetector) {
		// Create configuration
		cfg := &config.Config{
			Host:      "127.0.0.1",
			Port:      0, // Get free port
			Log:       false,
			APIKey:    "test-key",
			Providers: []config.Provider{
				{
					Name:       "openai",
					APIBaseURL: "https://api.openai.com/v1",
					APIKey:     "test-key",
					Enabled:    true,
				},
			},
		}
		
		// Get free port
		port, err := testfw.GetFreePort()
		require.NoError(t, err)
		cfg.Port = port
		
		// Create and start server
		srv, err := server.New(cfg)
		require.NoError(t, err)
		
		// Start server
		serverErr := make(chan error, 1)
		go func() {
			if err := srv.Run(); err != nil && err != http.ErrServerClosed {
				serverErr <- err
			}
		}()
		
		// Wait for server to start
		time.Sleep(200 * time.Millisecond)
		
		// Use tracked HTTP client for leak detection
		client := detector.HTTPClient()
		
		// Test various endpoints
		t.Run("HealthCheck", func(t *testing.T) {
			resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/health", cfg.Port))
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		})
		
		t.Run("MessagesEndpoint", func(t *testing.T) {
			payload := map[string]interface{}{
				"model": "gpt-3.5-turbo",
				"messages": []map[string]string{
					{"role": "user", "content": "Hello"},
				},
			}
			
			data, _ := json.Marshal(payload)
			req, _ := http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:%d/v1/messages", cfg.Port), bytes.NewReader(data))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test-key")
			
			resp, err := client.Do(req)
			if err == nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		})
		
		// Test concurrent requests
		t.Run("ConcurrentRequests", func(t *testing.T) {
			done := make(chan struct{})
			for i := 0; i < 5; i++ {
				go func(id int) {
					defer func() { done <- struct{}{} }()
					
					resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/health", cfg.Port))
					if err == nil {
						io.Copy(io.Discard, resp.Body)
						resp.Body.Close()
					}
				}(i)
			}
			
			// Wait for all goroutines
			for i := 0; i < 5; i++ {
				<-done
			}
		})
		
		// Shutdown server properly
		err = srv.Shutdown()
		assert.NoError(t, err)
		
		// Give time for final cleanup
		time.Sleep(200 * time.Millisecond)
	})
}

// TestLeakDetectionWithSimpleServer shows leak detection with the simple test server
func TestLeakDetectionWithSimpleServer(t *testing.T) {
	testfw.WithComprehensiveLeakDetection(t, 150, func(detector *testfw.LeakDetector) {
		// Create simple test server
		testServer := testfw.NewSimpleTestServer()
		defer testServer.Close()
		
		// Setup handlers
		testServer.HandleJSON("/api/test", map[string]string{"status": "ok"}, http.StatusOK)
		testServer.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			w.Write(body)
		})
		
		// Use tracked client
		client := detector.HTTPClient()
		
		// Make requests
		for i := 0; i < 3; i++ {
			// Test JSON endpoint
			resp, err := client.Get(testServer.URL() + "/api/test")
			require.NoError(t, err)
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			
			// Test echo endpoint
			payload := []byte("test data")
			resp, err = client.Post(testServer.URL()+"/echo", "text/plain", bytes.NewReader(payload))
			require.NoError(t, err)
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
		
		// Give time for cleanup
		time.Sleep(100 * time.Millisecond)
	})
}

// BenchmarkWithLeakDetection shows how to use leak detection in benchmarks
func BenchmarkWithLeakDetection(b *testing.B) {
	// Convert testing.B to testing.T for leak detection
	t := &testing.T{}
	
	detector := testfw.NewLeakDetector(t, 500)
	detector.Start()
	defer detector.Check()
	
	// Setup
	tf := testfw.NewTestFramework(t)
	srv, err := tf.StartServerWithError(tf.GetConfig())
	if err != nil {
		b.Fatal(err)
	}
	defer srv.Shutdown()
	
	client := detector.HTTPClient()
	
	// Reset timer after setup
	b.ResetTimer()
	
	// Run benchmark
	for i := 0; i < b.N; i++ {
		resp, err := client.Get(fmt.Sprintf("http://localhost:%d/health", tf.GetConfig().Port))
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}
	
	// Stop timer before cleanup
	b.StopTimer()
}