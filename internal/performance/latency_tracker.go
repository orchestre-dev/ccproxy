package performance

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// LatencyTracker tracks request latencies and calculates percentiles
type LatencyTracker struct {
	samples    []time.Duration
	maxSamples int
	mu         sync.RWMutex
}

// LatencyPercentiles represents latency percentiles
type LatencyPercentiles struct {
	Average time.Duration `json:"average"`
	P50     time.Duration `json:"p50"`
	P95     time.Duration `json:"p95"`
	P99     time.Duration `json:"p99"`
	Min     time.Duration `json:"min"`
	Max     time.Duration `json:"max"`
}

// NewLatencyTracker creates a new latency tracker
func NewLatencyTracker() *LatencyTracker {
	return &LatencyTracker{
		samples:    make([]time.Duration, 0, 10000),
		maxSamples: 10000, // Keep last 10k samples
	}
}

// Record records a latency sample
func (lt *LatencyTracker) Record(latency time.Duration) {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	lt.samples = append(lt.samples, latency)

	// Keep only the last maxSamples
	if len(lt.samples) > lt.maxSamples {
		// Remove oldest samples
		copy(lt.samples, lt.samples[len(lt.samples)-lt.maxSamples:])
		lt.samples = lt.samples[:lt.maxSamples]
	}
}

// GetPercentiles calculates and returns latency percentiles
func (lt *LatencyTracker) GetPercentiles() LatencyPercentiles {
	lt.mu.RLock()
	defer lt.mu.RUnlock()

	if len(lt.samples) == 0 {
		return LatencyPercentiles{}
	}

	// Copy samples for sorting
	samples := make([]time.Duration, len(lt.samples))
	copy(samples, lt.samples)

	// Sort samples
	sort.Slice(samples, func(i, j int) bool {
		return samples[i] < samples[j]
	})

	// Calculate percentiles
	percentiles := LatencyPercentiles{
		Min: samples[0],
		Max: samples[len(samples)-1],
	}

	// Calculate average
	var sum int64
	for _, s := range samples {
		sum += int64(s)
	}
	percentiles.Average = time.Duration(sum / int64(len(samples)))

	// Calculate specific percentiles
	percentiles.P50 = getPercentile(samples, 50)
	percentiles.P95 = getPercentile(samples, 95)
	percentiles.P99 = getPercentile(samples, 99)

	return percentiles
}

// Reset clears all samples
func (lt *LatencyTracker) Reset() {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	lt.samples = lt.samples[:0]
}

// getPercentile calculates a specific percentile from sorted samples
func getPercentile(sortedSamples []time.Duration, percentile float64) time.Duration {
	if len(sortedSamples) == 0 {
		return 0
	}

	if percentile <= 0 {
		return sortedSamples[0]
	}
	if percentile >= 100 {
		return sortedSamples[len(sortedSamples)-1]
	}

	// Calculate index
	index := (percentile / 100.0) * float64(len(sortedSamples)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sortedSamples[lower]
	}

	// Linear interpolation
	weight := index - float64(lower)
	return time.Duration(
		float64(sortedSamples[lower])*(1-weight) + float64(sortedSamples[upper])*weight,
	)
}

// GetSampleCount returns the current number of samples
func (lt *LatencyTracker) GetSampleCount() int {
	lt.mu.RLock()
	defer lt.mu.RUnlock()
	return len(lt.samples)
}

// GetHistogram returns a histogram of latencies
func (lt *LatencyTracker) GetHistogram(buckets []time.Duration) map[string]int {
	lt.mu.RLock()
	defer lt.mu.RUnlock()

	histogram := make(map[string]int)
	
	// Initialize all bucket keys
	for i := 0; i <= len(buckets); i++ {
		var key string
		if i == 0 {
			key = fmt.Sprintf("<%v", buckets[i])
		} else if i == len(buckets) {
			key = fmt.Sprintf(">=%v", buckets[i-1])
		} else {
			key = fmt.Sprintf("%v-%v", buckets[i-1], buckets[i])
		}
		histogram[key] = 0
	}

	// Count samples in each bucket
	for _, sample := range lt.samples {
		placed := false
		for i, bucket := range buckets {
			if sample < bucket {
				var key string
				if i == 0 {
					key = fmt.Sprintf("<%v", bucket)
				} else {
					key = fmt.Sprintf("%v-%v", buckets[i-1], bucket)
				}
				histogram[key]++
				placed = true
				break
			}
		}
		// If not placed in any bucket, it goes in the last bucket
		if !placed {
			key := fmt.Sprintf(">=%v", buckets[len(buckets)-1])
			histogram[key]++
		}
	}

	return histogram
}