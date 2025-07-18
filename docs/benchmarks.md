# Performance Benchmarks

CCProxy includes comprehensive performance benchmarks to ensure optimal performance for local AI proxy operations.

## Running Benchmarks

### Quick Start

Run all benchmarks:
```bash
./scripts/run-benchmarks.sh
```

### Specific Package Benchmarks

Run benchmarks for specific packages:

```bash
# HTTP handler benchmarks
go test -bench=. -benchmem ccproxy/internal/handlers

# Converter benchmarks
go test -bench=. -benchmem ccproxy/internal/converter

# Provider common benchmarks
go test -bench=. -benchmem ccproxy/internal/provider/common
```

### Specific Benchmark Tests

Run individual benchmarks:

```bash
# Test simple request handling
go test -bench=BenchmarkProxyMessages_SimpleRequest -benchmem ccproxy/internal/handlers

# Test concurrent request handling
go test -bench=BenchmarkProxyMessages_Concurrent -benchmem ccproxy/internal/handlers

# Test conversion performance
go test -bench=BenchmarkConvertAnthropicToOpenAI -benchmem ccproxy/internal/converter
```

## Benchmark Coverage

### HTTP Handler Benchmarks (`internal/handlers`)

- **BenchmarkProxyMessages_SimpleRequest**: Basic single message request
- **BenchmarkProxyMessages_LargePayload**: Large message handling (~24KB)
- **BenchmarkProxyMessages_MultiTurn**: Multi-turn conversation (10 messages)
- **BenchmarkProxyMessages_Concurrent**: Concurrent request handling
- **BenchmarkHealthCheck**: Health check endpoint performance
- **BenchmarkJSONMarshaling**: JSON encoding/decoding performance
- **BenchmarkErrorHandling**: Error response generation
- **BenchmarkMemoryAllocation**: Memory allocation tracking
- **BenchmarkHighConcurrency**: Performance under 100 concurrent requests

### Converter Benchmarks (`internal/converter`)

- **BenchmarkConvertAnthropicToOpenAI**: Anthropic to OpenAI format conversion
  - Simple messages
  - Multi-turn conversations
  - System prompts
  - Large messages (~10KB)
  - Tool/function calls
- **BenchmarkConvertOpenAIToAnthropic**: OpenAI to Anthropic format conversion
  - Simple responses
  - Tool call responses
  - Large responses (~10KB)
- **BenchmarkComplexToolConversion**: Complex tool call scenarios
- **BenchmarkMemoryEfficiency**: Memory allocation during conversion

### Provider Common Benchmarks (`internal/provider/common`)

- **BenchmarkHTTPClient**: HTTP client performance with different timeouts
- **BenchmarkErrorCreation**: Error object creation overhead
- **BenchmarkLargePayloadHandling**: Large payload handling (1KB to 1MB)
- **BenchmarkConcurrentRequests**: Concurrent HTTP request performance
- **BenchmarkHealthCheck**: Provider health check implementations

## Benchmark Options

### CPU Profiling

Generate CPU profiles for detailed analysis:

```bash
go test -bench=. -cpuprofile=cpu.prof ccproxy/internal/handlers
go tool pprof cpu.prof
```

### Memory Profiling

Generate memory profiles:

```bash
go test -bench=. -memprofile=mem.prof ccproxy/internal/handlers
go tool pprof mem.prof
```

### Benchmark Duration

Control benchmark duration:

```bash
# Run each benchmark for 30 seconds
go test -bench=. -benchtime=30s ccproxy/internal/handlers

# Run each benchmark 1000 times
go test -bench=. -benchtime=1000x ccproxy/internal/handlers
```

### CPU Variations

Test with different CPU counts:

```bash
go test -bench=. -cpu=1,2,4,8 ccproxy/internal/handlers
```

## Performance Goals

CCProxy is designed for local use with minimal overhead:

- **Latency**: < 1ms overhead for request proxying
- **Throughput**: > 10,000 requests/second on modern hardware
- **Memory**: < 10MB baseline memory usage
- **Concurrency**: Handle 1000+ concurrent connections

## Benchmark Results

Results are stored in the `benchmark-results/` directory with timestamps. Each run generates a detailed report including:

- Execution time (ns/op)
- Memory allocation (B/op)
- Number of allocations (allocs/op)
- Results for different CPU counts

## Continuous Performance Monitoring

To track performance over time:

1. Run benchmarks regularly:
   ```bash
   ./scripts/run-benchmarks.sh
   ```

2. Compare results using benchstat:
   ```bash
   go install golang.org/x/perf/cmd/benchstat@latest
   benchstat old.txt new.txt
   ```

3. Monitor for performance regressions in CI/CD pipelines

## Optimization Tips

Based on benchmark results, here are key optimization areas:

1. **JSON Processing**: Use streaming JSON for large payloads
2. **Memory Allocation**: Reuse buffers where possible
3. **Concurrent Handling**: Leverage Go's goroutines effectively
4. **HTTP Client**: Use connection pooling and keep-alive

## Example Benchmark Output

```
BenchmarkProxyMessages_SimpleRequest-8    50000    23456 ns/op    4096 B/op    42 allocs/op
BenchmarkProxyMessages_Concurrent-8      100000    15678 ns/op    3072 B/op    35 allocs/op
BenchmarkConvertAnthropicToOpenAI-8      200000     8901 ns/op    2048 B/op    18 allocs/op
```

This shows:
- Simple requests: ~23μs per request
- Concurrent handling: ~15μs per request (better due to parallelism)
- Format conversion: ~9μs per conversion