package testing

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

// BenchmarkFramework provides utilities for running benchmarks
type BenchmarkFramework struct {
	b         *testing.B
	startTime time.Time
	metrics   BenchmarkMetrics
}

// BenchmarkMetrics contains benchmark performance metrics
type BenchmarkMetrics struct {
	Operations   int64
	Duration     time.Duration
	OpsPerSecond float64
	AllocBytes   uint64
	Allocs       uint64
	CustomMetrics map[string]float64
}

// NewBenchmarkFramework creates a new benchmark framework
func NewBenchmarkFramework(b *testing.B) *BenchmarkFramework {
	return &BenchmarkFramework{
		b: b,
		metrics: BenchmarkMetrics{
			CustomMetrics: make(map[string]float64),
		},
	}
}

// RunBenchmark runs a benchmark function and collects metrics
func (bf *BenchmarkFramework) RunBenchmark(fn func(int)) {
	// Reset timer and start measurement
	bf.b.ResetTimer()
	bf.startTime = time.Now()
	
	// Run the benchmark
	for i := 0; i < bf.b.N; i++ {
		fn(i)
	}
	
	// Stop timer and collect metrics
	bf.b.StopTimer()
	bf.collectMetrics()
}

// RunParallelBenchmark runs a parallel benchmark
func (bf *BenchmarkFramework) RunParallelBenchmark(fn func(*testing.PB)) {
	// Reset timer and start measurement
	bf.b.ResetTimer()
	bf.startTime = time.Now()
	
	// Run the parallel benchmark
	bf.b.RunParallel(fn)
	
	// Stop timer and collect metrics
	bf.b.StopTimer()
	bf.collectMetrics()
}

// collectMetrics collects benchmark metrics
func (bf *BenchmarkFramework) collectMetrics() {
	bf.metrics.Operations = int64(bf.b.N)
	bf.metrics.Duration = time.Since(bf.startTime)
	
	if bf.metrics.Duration > 0 {
		bf.metrics.OpsPerSecond = float64(bf.metrics.Operations) / bf.metrics.Duration.Seconds()
	}
	
	// Collect memory statistics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	bf.metrics.AllocBytes = m.Alloc
	bf.metrics.Allocs = m.Mallocs
}

// RecordCustomMetric records a custom metric
func (bf *BenchmarkFramework) RecordCustomMetric(name string, value float64) {
	bf.metrics.CustomMetrics[name] = value
}

// GetMetrics returns the collected metrics
func (bf *BenchmarkFramework) GetMetrics() BenchmarkMetrics {
	return bf.metrics
}

// Report generates a benchmark report
func (bf *BenchmarkFramework) Report() string {
	report := fmt.Sprintf("Benchmark Report:\n")
	report += fmt.Sprintf("  Operations: %d\n", bf.metrics.Operations)
	report += fmt.Sprintf("  Duration: %s\n", bf.metrics.Duration)
	report += fmt.Sprintf("  Ops/second: %.2f\n", bf.metrics.OpsPerSecond)
	report += fmt.Sprintf("  Alloc bytes: %d\n", bf.metrics.AllocBytes)
	report += fmt.Sprintf("  Allocations: %d\n", bf.metrics.Allocs)
	
	if len(bf.metrics.CustomMetrics) > 0 {
		report += "  Custom metrics:\n"
		for name, value := range bf.metrics.CustomMetrics {
			report += fmt.Sprintf("    %s: %.2f\n", name, value)
		}
	}
	
	return report
}

// SetParallelism sets the parallelism level for parallel benchmarks
func (bf *BenchmarkFramework) SetParallelism(p int) {
	bf.b.SetParallelism(p)
}