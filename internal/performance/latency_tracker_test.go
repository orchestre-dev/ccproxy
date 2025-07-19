package performance

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLatencyTracker(t *testing.T) {
	t.Run("NewLatencyTracker", func(t *testing.T) {
		tracker := NewLatencyTracker()
		require.NotNil(t, tracker)
		assert.Equal(t, 0, tracker.GetSampleCount())
	})

	t.Run("Record and GetPercentiles", func(t *testing.T) {
		tracker := NewLatencyTracker()

		// Add samples
		samples := []time.Duration{
			10 * time.Millisecond,
			20 * time.Millisecond,
			30 * time.Millisecond,
			40 * time.Millisecond,
			50 * time.Millisecond,
			60 * time.Millisecond,
			70 * time.Millisecond,
			80 * time.Millisecond,
			90 * time.Millisecond,
			100 * time.Millisecond,
		}

		for _, sample := range samples {
			tracker.Record(sample)
		}

		assert.Equal(t, len(samples), tracker.GetSampleCount())

		percentiles := tracker.GetPercentiles()
		assert.Equal(t, 10*time.Millisecond, percentiles.Min)
		assert.Equal(t, 100*time.Millisecond, percentiles.Max)
		assert.Equal(t, 55*time.Millisecond, percentiles.Average)
		assert.InDelta(t, 50*time.Millisecond, percentiles.P50, float64(5*time.Millisecond))
		assert.InDelta(t, 95*time.Millisecond, percentiles.P95, float64(10*time.Millisecond))
		assert.InDelta(t, 99*time.Millisecond, percentiles.P99, float64(10*time.Millisecond))
	})

	t.Run("Empty Percentiles", func(t *testing.T) {
		tracker := NewLatencyTracker()
		percentiles := tracker.GetPercentiles()
		assert.Equal(t, time.Duration(0), percentiles.Average)
		assert.Equal(t, time.Duration(0), percentiles.P50)
		assert.Equal(t, time.Duration(0), percentiles.P95)
		assert.Equal(t, time.Duration(0), percentiles.P99)
	})

	t.Run("Max Samples", func(t *testing.T) {
		tracker := NewLatencyTracker()

		// Add more than max samples
		for i := 0; i < 15000; i++ {
			tracker.Record(time.Duration(i) * time.Millisecond)
		}

		// Should only keep last 10000
		assert.Equal(t, 10000, tracker.GetSampleCount())

		// Verify oldest samples were removed
		percentiles := tracker.GetPercentiles()
		assert.Greater(t, percentiles.Min, time.Duration(0))
	})

	t.Run("Reset", func(t *testing.T) {
		tracker := NewLatencyTracker()

		// Add samples
		for i := 0; i < 100; i++ {
			tracker.Record(time.Duration(i) * time.Millisecond)
		}

		assert.Equal(t, 100, tracker.GetSampleCount())

		// Reset
		tracker.Reset()
		assert.Equal(t, 0, tracker.GetSampleCount())
	})

	t.Run("GetHistogram", func(t *testing.T) {
		tracker := NewLatencyTracker()

		// Add samples
		for i := 0; i < 100; i++ {
			tracker.Record(time.Duration(i) * time.Millisecond)
		}

		buckets := []time.Duration{
			25 * time.Millisecond,
			50 * time.Millisecond,
			75 * time.Millisecond,
			100 * time.Millisecond,
		}

		histogram := tracker.GetHistogram(buckets)
		// Check that the bucket labels exist
		assert.Contains(t, histogram, "<25ms")
		assert.Contains(t, histogram, "25ms-50ms")
		assert.Contains(t, histogram, "50ms-75ms")
		assert.Contains(t, histogram, ">=100ms")
		
		// Check distribution
		total := 0
		for _, count := range histogram {
			total += count
		}
		assert.Equal(t, 100, total) // All samples should be counted
	})

	t.Run("Concurrent Access", func(t *testing.T) {
		tracker := NewLatencyTracker()

		// Concurrent writes
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func(idx int) {
				for j := 0; j < 100; j++ {
					tracker.Record(time.Duration(idx*100+j) * time.Millisecond)
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		assert.Equal(t, 1000, tracker.GetSampleCount())

		// Concurrent reads
		for i := 0; i < 10; i++ {
			go func() {
				percentiles := tracker.GetPercentiles()
				assert.NotNil(t, percentiles)
				done <- true
			}()
		}

		// Wait for reads
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

func TestGetPercentile(t *testing.T) {
	samples := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
		50 * time.Millisecond,
	}

	t.Run("Edge cases", func(t *testing.T) {
		assert.Equal(t, 10*time.Millisecond, getPercentile(samples, 0))
		assert.Equal(t, 50*time.Millisecond, getPercentile(samples, 100))
		assert.Equal(t, 30*time.Millisecond, getPercentile(samples, 50))
	})

	t.Run("Empty samples", func(t *testing.T) {
		assert.Equal(t, time.Duration(0), getPercentile([]time.Duration{}, 50))
	})

	t.Run("Interpolation", func(t *testing.T) {
		// Test interpolation between values
		p25 := getPercentile(samples, 25)
		assert.Greater(t, p25, 10*time.Millisecond)
		assert.Less(t, p25, 30*time.Millisecond)
	})
}