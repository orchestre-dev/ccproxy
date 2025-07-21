package testing

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestGoroutineLeakDetection verifies goroutine leak detection works
func TestGoroutineLeakDetection(t *testing.T) {
	t.Run("NoLeak", func(t *testing.T) {
		WithLeakDetection(t, func() {
			// Start goroutine that properly exits
			done := make(chan struct{})
			go func() {
				<-done
			}()
			close(done)
			time.Sleep(50 * time.Millisecond)
		})
	})
}

// TestResourceMonitoring verifies memory monitoring works
func TestResourceMonitoring(t *testing.T) {
	t.Run("WithinLimits", func(t *testing.T) {
		WithResourceMonitoring(t, 50, func() {
			// Allocate some memory
			data := make([]byte, 10*1024*1024) // 10MB
			_ = data
			time.Sleep(100 * time.Millisecond)
		})
	})
}

// TestConnectionTracking verifies connection leak detection
func TestConnectionTracking(t *testing.T) {
	t.Run("HTTPConnectionLeak", func(t *testing.T) {
		// Create a test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test"))
		}))
		defer server.Close()

		tracker := NewConnectionTracker(t)
		client := TrackedHTTPClient(tracker)

		// Make a request
		resp, err := client.Get(server.URL)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		
		// Properly close the response body
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		// Check for leaks (should find none)
		tracker.CheckLeaks()
	})

	t.Run("TCPConnectionTracking", func(t *testing.T) {
		// Start a test TCP server
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("Failed to start listener: %v", err)
		}
		defer listener.Close()

		go func() {
			for {
				conn, err := listener.Accept()
				if err != nil {
					return
				}
				conn.Close()
			}
		}()

		tracker := NewConnectionTracker(t)
		
		// Create a tracked connection
		conn, err := net.Dial("tcp", listener.Addr().String())
		if err != nil {
			t.Fatalf("Failed to dial: %v", err)
		}
		
		trackedConn := tracker.TrackConnection(conn, "tcp")
		
		// Properly close the connection
		trackedConn.Close()
		
		// Check for leaks (should find none)
		tracker.CheckLeaks()
	})
}

// TestComprehensiveLeakDetection demonstrates the full leak detector
func TestComprehensiveLeakDetection(t *testing.T) {
	WithComprehensiveLeakDetection(t, 100, func(detector *LeakDetector) {
		// Create a test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test response"))
		}))
		defer server.Close()

		// Use the tracked HTTP client
		client := detector.HTTPClient()
		
		// Make multiple requests
		for i := 0; i < 5; i++ {
			resp, err := client.Get(server.URL)
			if err != nil {
				t.Errorf("Request %d failed: %v", i, err)
				continue
			}
			
			// Properly consume and close response
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}

		// Start some goroutines that exit properly
		done := make(chan struct{})
		for i := 0; i < 3; i++ {
			go func(id int) {
				select {
				case <-done:
					return
				case <-time.After(50 * time.Millisecond):
					return
				}
			}(i)
		}
		close(done)
		
		// Give everything time to clean up
		time.Sleep(100 * time.Millisecond)
	})
}

// Example_leakDetectorUsage shows how to use the leak detector in tests
func Example_leakDetectorUsage() {
	// In your test:
	t := &testing.T{} // This would be your actual test
	
	// Option 1: Just goroutine detection
	WithLeakDetection(t, func() {
		// Your test code here
	})
	
	// Option 2: Just resource monitoring
	WithResourceMonitoring(t, 100, func() {
		// Your test code here
	})
	
	// Option 3: Comprehensive detection
	WithComprehensiveLeakDetection(t, 100, func(detector *LeakDetector) {
		// Use detector.HTTPClient() for HTTP requests
		// Use detector.Dialer() for TCP connections
		// All leaks will be detected automatically
	})
}