package safety

import (
	"testing"
	"time"
	
	testfw "github.com/orchestre-dev/ccproxy/testing"
)

// TestWithLeakDetection demonstrates goroutine leak detection
func TestWithLeakDetection(t *testing.T) {
	testfw.WithLeakDetection(t, func() {
		// This test should pass - no goroutines leaked
		ch := make(chan struct{})
		go func() {
			<-ch
		}()
		close(ch)
		time.Sleep(100 * time.Millisecond)
	})
}

// TestWithResourceMonitoring demonstrates resource monitoring
func TestWithResourceMonitoring(t *testing.T) {
	testfw.WithResourceMonitoring(t, 100, func() {
		// This test allocates memory within limits
		data := make([]byte, 10*1024*1024) // 10MB
		_ = data
		
		// Let GC run
		time.Sleep(100 * time.Millisecond)
	})
}

// TestSafeGoroutineUsage shows proper goroutine management
func TestSafeGoroutineUsage(t *testing.T) {
	detector := testfw.NewGoroutineLeakDetector(t)
	defer detector.Check()
	
	// Use a worker pool instead of unbounded goroutines
	workerCount := 10
	jobs := make(chan int, 100)
	done := make(chan struct{})
	
	// Start workers
	for i := 0; i < workerCount; i++ {
		go func() {
			for job := range jobs {
				// Process job
				_ = job * 2
			}
			done <- struct{}{}
		}()
	}
	
	// Send jobs
	for i := 0; i < 100; i++ {
		jobs <- i
	}
	close(jobs)
	
	// Wait for workers to finish
	for i := 0; i < workerCount; i++ {
		<-done
	}
}