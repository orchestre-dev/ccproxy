package common

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// BenchmarkHTTPClient benchmarks the configured HTTP client performance
func BenchmarkHTTPClient(b *testing.B) {
	// Create a test server that responds quickly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	b.Run("Default_Timeout", func(b *testing.B) {
		client := NewConfiguredHTTPClient(30 * time.Second)
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
			resp, err := client.Do(req)
			if err != nil {
				b.Fatal(err)
			}
			resp.Body.Close()
		}
	})

	b.Run("Short_Timeout", func(b *testing.B) {
		client := NewConfiguredHTTPClient(5 * time.Second)
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
			resp, err := client.Do(req)
			if err != nil {
				b.Fatal(err)
			}
			resp.Body.Close()
		}
	})

	b.Run("With_Keep_Alive", func(b *testing.B) {
		client := NewConfiguredHTTPClient(30 * time.Second)
		ctx := context.Background()

		// Warm up connection
		req, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
		resp, _ := client.Do(req)
		resp.Body.Close()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
			resp, err := client.Do(req)
			if err != nil {
				b.Fatal(err)
			}
			resp.Body.Close()
		}
	})
}

// BenchmarkErrorCreation benchmarks error creation performance
func BenchmarkErrorCreation(b *testing.B) {
	b.Run("ProviderError", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewProviderError("test-provider", "test error message", nil)
		}
	})

	b.Run("ConfigError", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewConfigError("test-provider", "test-field", "validation failed")
		}
	})

	b.Run("HTTPError", func(b *testing.B) {
		resp := &http.Response{
			StatusCode: 500,
			Body:       http.NoBody,
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewHTTPError("test-provider", resp, nil)
		}
	})

	b.Run("RateLimitError", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewRateLimitError("test-provider", "60s")
		}
	})
}

// BenchmarkLargePayloadHandling benchmarks handling of large payloads
func BenchmarkLargePayloadHandling(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"1KB", 1024},
		{"10KB", 10 * 1024},
		{"100KB", 100 * 1024},
		{"1MB", 1024 * 1024},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			// Create a server that echoes back the request body
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				// Create a response of similar size
				w.Write([]byte(strings.Repeat("a", size.size)))
			}))
			defer server.Close()

			client := NewConfiguredHTTPClient(30 * time.Second)
			ctx := context.Background()
			payload := strings.NewReader(strings.Repeat("b", size.size))

			b.ResetTimer()
			b.SetBytes(int64(size.size))
			for i := 0; i < b.N; i++ {
				req, _ := http.NewRequestWithContext(ctx, "POST", server.URL, payload)
				req.Header.Set("Content-Type", "application/json")
				resp, err := client.Do(req)
				if err != nil {
					b.Fatal(err)
				}
				// Read and discard the response
				buf := make([]byte, 4096)
				for {
					_, err := resp.Body.Read(buf)
					if err != nil {
						break
					}
				}
				resp.Body.Close()
				payload.Seek(0, 0) // Reset reader for next iteration
			}
		})
	}
}

// BenchmarkConcurrentRequests benchmarks concurrent request handling
func BenchmarkConcurrentRequests(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate some processing time
		time.Sleep(10 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewConfiguredHTTPClient(30 * time.Second)

	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			req, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
			resp, err := client.Do(req)
			if err != nil {
				b.Fatal(err)
			}
			resp.Body.Close()
		}
	})
}

// BenchmarkHealthCheck benchmarks health check implementations
func BenchmarkHealthCheck(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewConfiguredHTTPClient(5 * time.Second)

	b.Run("BasicHealthCheck", func(b *testing.B) {
		ctx := context.Background()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := BasicHealthCheck(ctx, client, server.URL, "test-provider")
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("WithRetry", func(b *testing.B) {
		// Mock health checker that always succeeds
		checker := &mockHealthChecker{healthy: true}
		config := HealthCheckConfig{
			Timeout:    1 * time.Second,
			Retries:    3,
			RetryDelay: 10 * time.Millisecond,
		}

		ctx := context.Background()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := PerformHealthCheckWithRetry(ctx, checker, config)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// mockHealthChecker for benchmarking
type mockHealthChecker struct {
	healthy bool
}

func (m *mockHealthChecker) HealthCheck(ctx context.Context) error {
	if m.healthy {
		return nil
	}
	return NewProviderError("mock", "unhealthy", nil)
}