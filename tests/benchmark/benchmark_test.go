package benchmark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/orchestre-dev/ccproxy/internal/router"
	testfw "github.com/orchestre-dev/ccproxy/internal/testing"
	"github.com/orchestre-dev/ccproxy/internal/token"
	"github.com/orchestre-dev/ccproxy/internal/transformer"
)

// BenchmarkTokenCounting benchmarks token counting performance
func BenchmarkTokenCounting(b *testing.B) {
	counter := token.NewCounter()
	fixtures := testfw.NewFixtures()
	
	testCases := []struct {
		name   string
		tokens int
	}{
		{"Small-100", 100},
		{"Medium-1000", 1000},
		{"Large-10000", 10000},
		{"XLarge-50000", 50000},
	}
	
	for _, tc := range testCases {
		message := fixtures.GenerateLargeMessage(tc.tokens)
		
		b.Run(tc.name, func(b *testing.B) {
			bf := testfw.NewBenchmarkFramework(b)
			
			bf.RunBenchmark(func(i int) {
				_, _ = counter.CountTokens(message, "claude-3-sonnet-20240229")
			})
			
			b.Log(bf.Report())
		})
	}
}

// BenchmarkRouting benchmarks routing decision performance
func BenchmarkRouting(b *testing.B) {
	framework := testfw.NewTestFramework(&testing.T{})
	fixtures := testfw.NewFixtures()
	
	// Create router with multiple providers
	cfg := framework.GetConfig()
	r, _ := router.New(cfg.Routing)
	
	// Test different message sizes
	testCases := []struct {
		name     string
		messages int
	}{
		{"SingleMessage", 1},
		{"ShortConversation", 5},
		{"LongConversation", 20},
		{"VeryLongConversation", 50},
	}
	
	for _, tc := range testCases {
		messages := fixtures.GenerateMessages(tc.messages)
		
		b.Run(tc.name, func(b *testing.B) {
			bf := testfw.NewBenchmarkFramework(b)
			
			bf.RunBenchmark(func(i int) {
				req := &router.Request{
					Model:    "claude-3-sonnet-20240229",
					Messages: messages,
				}
				_, _ = r.SelectProvider(req)
			})
			
			b.Log(bf.Report())
		})
	}
}

// BenchmarkTransformer benchmarks message transformation
func BenchmarkTransformer(b *testing.B) {
	transformerSvc := transformer.New()
	fixtures := testfw.NewFixtures()
	
	// Register transformers
	transformerSvc.Register("anthropic", transformer.NewAnthropicTransformer())
	transformerSvc.Register("openai", transformer.NewOpenAITransformer())
	
	testCases := []struct {
		name      string
		transform string
		fixture   string
	}{
		{"Anthropic-Small", "anthropic", "anthropic_messages"},
		{"OpenAI-Small", "openai", "openai_chat"},
		{"Anthropic-Large", "anthropic", "large_messages"},
		{"OpenAI-Large", "openai", "large_messages"},
	}
	
	// Add large message fixture
	largeMessages := fixtures.GenerateMessages(50)
	fixtures.AddRequest("large_messages", map[string]interface{}{
		"model":    "claude-3-sonnet-20240229",
		"messages": largeMessages,
	})
	
	for _, tc := range testCases {
		reqData, _ := fixtures.GetRequest(tc.fixture)
		
		b.Run(tc.name, func(b *testing.B) {
			bf := testfw.NewBenchmarkFramework(b)
			
			bf.RunBenchmark(func(i int) {
				_, _ = transformerSvc.TransformRequest(tc.transform, reqData)
			})
			
			b.Log(bf.Report())
		})
	}
}

// BenchmarkHTTPServer benchmarks HTTP server performance
func BenchmarkHTTPServer(b *testing.B) {
	framework := testfw.NewTestFramework(&testing.T{})
	fixtures := testfw.NewFixtures()
	
	// Start server
	server := framework.StartServer()
	serverURL := fmt.Sprintf("http://127.0.0.1:%d", server.GetPort())
	
	// Create mock provider
	mockProvider := testfw.NewMockProviderServer("anthropic")
	defer mockProvider.Close()
	
	// Update provider config with mock URL
	// This would require extending the framework to support updating provider configs
	
	testCases := []struct {
		name     string
		endpoint string
		fixture  string
	}{
		{"Messages-Small", "/v1/messages", "anthropic_messages"},
		{"Messages-Large", "/v1/messages", "large_request"},
		{"ChatCompletions-Small", "/v1/chat/completions", "openai_chat"},
		{"ChatCompletions-Large", "/v1/chat/completions", "large_chat"},
	}
	
	// Add large request fixtures
	fixtures.AddRequest("large_request", map[string]interface{}{
		"model":    "claude-3-sonnet-20240229",
		"messages": fixtures.GenerateMessages(50),
	})
	
	fixtures.AddRequest("large_chat", map[string]interface{}{
		"model":    "gpt-4",
		"messages": fixtures.GenerateMessages(50),
	})
	
	for _, tc := range testCases {
		reqBody, _ := fixtures.GetRequest(tc.fixture)
		reqData, _ := json.Marshal(reqBody)
		
		b.Run(tc.name, func(b *testing.B) {
			bf := testfw.NewBenchmarkFramework(b)
			
			client := &http.Client{}
			
			bf.RunBenchmark(func(i int) {
				req, _ := http.NewRequest("POST", serverURL+tc.endpoint, bytes.NewReader(reqData))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer test-key")
				
				resp, err := client.Do(req)
				if err == nil {
					resp.Body.Close()
				}
			})
			
			b.Log(bf.Report())
		})
	}
}

// BenchmarkConcurrentRequests benchmarks concurrent request handling
func BenchmarkConcurrentRequests(b *testing.B) {
	framework := testfw.NewTestFramework(&testing.T{})
	fixtures := testfw.NewFixtures()
	
	// Start server
	server := framework.StartServer()
	serverURL := fmt.Sprintf("http://127.0.0.1:%d", server.GetPort())
	
	// Create mock provider
	mockProvider := testfw.NewMockProviderServer("anthropic")
	defer mockProvider.Close()
	
	reqBody, _ := fixtures.GetRequest("anthropic_messages")
	reqData, _ := json.Marshal(reqBody)
	
	concurrencyLevels := []int{1, 10, 50, 100}
	
	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrent-%d", concurrency), func(b *testing.B) {
			bf := testfw.NewBenchmarkFramework(b)
			
			client := &http.Client{}
			
			bf.RunParallelBenchmark(func(pb *testing.PB) {
				req, _ := http.NewRequest("POST", serverURL+"/v1/messages", bytes.NewReader(reqData))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer test-key")
				
				resp, err := client.Do(req)
				if err == nil {
					resp.Body.Close()
				}
			})
			
			b.Log(bf.Report())
			bf.RecordCustomMetric("concurrency", concurrency)
		})
	}
}

// BenchmarkMemoryUsage benchmarks memory usage patterns
func BenchmarkMemoryUsage(b *testing.B) {
	fixtures := testfw.NewFixtures()
	
	testCases := []struct {
		name  string
		size  int
		count int
	}{
		{"SmallMessages", 100, 10},
		{"MediumMessages", 1000, 10},
		{"LargeMessages", 10000, 10},
		{"ManySmallMessages", 100, 100},
		{"FewLargeMessages", 50000, 5},
	}
	
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			bf := testfw.NewBenchmarkFramework(b)
			
			// Generate test data
			messages := make([]string, tc.count)
			for i := 0; i < tc.count; i++ {
				messages[i] = fixtures.GenerateLargeMessage(tc.size)
			}
			
			bf.RunBenchmark(func(i int) {
				// Simulate processing messages
				var result []map[string]interface{}
				for _, msg := range messages {
					result = append(result, map[string]interface{}{
						"role":    "user",
						"content": msg,
					})
				}
				
				// Force some allocations
				data, _ := json.Marshal(result)
				_ = data
			})
			
			metrics := bf.GetMetrics()
			b.Logf("Memory per operation: %s", formatBytes(metrics.AllocBytes/uint64(b.N)))
			b.Log(bf.Report())
		})
	}
}

// BenchmarkStreamingPerformance benchmarks streaming response handling
func BenchmarkStreamingPerformance(b *testing.B) {
	framework := testfw.NewTestFramework(&testing.T{})
	
	// Create mock streaming server
	mockServer := testfw.NewMockServer()
	defer mockServer.Close()
	
	// Setup streaming response with various chunk sizes
	testCases := []struct {
		name       string
		chunkCount int
		chunkSize  int
	}{
		{"SmallChunks", 100, 100},
		{"MediumChunks", 50, 1000},
		{"LargeChunks", 10, 10000},
		{"ManyTinyChunks", 1000, 10},
	}
	
	for _, tc := range testCases {
		// Generate chunks
		chunks := make([]string, tc.chunkCount)
		for i := 0; i < tc.chunkCount; i++ {
			content := make([]byte, tc.chunkSize)
			for j := range content {
				content[j] = byte('a' + (j % 26))
			}
			chunks[i] = fmt.Sprintf("data: {\"delta\":{\"text\":\"%s\"}}\n", string(content))
		}
		
		mockServer.AddStreamingRoute("POST", "/stream", chunks)
		
		b.Run(tc.name, func(b *testing.B) {
			bf := testfw.NewBenchmarkFramework(b)
			
			client := &http.Client{}
			
			bf.RunBenchmark(func(i int) {
				req, _ := http.NewRequest("POST", mockServer.GetURL()+"/stream", nil)
				resp, err := client.Do(req)
				if err == nil {
					// Read all chunks
					buf := make([]byte, 8192)
					for {
						_, err := resp.Body.Read(buf)
						if err != nil {
							break
						}
					}
					resp.Body.Close()
				}
			})
			
			bf.RecordCustomMetric("total_bytes", tc.chunkCount*tc.chunkSize)
			b.Log(bf.Report())
		})
	}
}

// Helper function
func formatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}