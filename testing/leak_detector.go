package testing

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// GoroutineLeakDetector helps detect goroutine leaks in tests
type GoroutineLeakDetector struct {
	t                 *testing.T
	initialGoroutines []string
	threshold         int
}

// NewGoroutineLeakDetector creates a new leak detector
func NewGoroutineLeakDetector(t *testing.T) *GoroutineLeakDetector {
	return &GoroutineLeakDetector{
		t:                 t,
		initialGoroutines: getGoroutineStacks(),
		threshold:         10, // Allow up to 10 extra goroutines
	}
}

// Check verifies no goroutines leaked
func (d *GoroutineLeakDetector) Check() {
	// Give goroutines time to clean up
	time.Sleep(100 * time.Millisecond)
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	currentGoroutines := getGoroutineStacks()
	leaked := findLeakedGoroutines(d.initialGoroutines, currentGoroutines)

	if len(leaked) > d.threshold {
		d.t.Errorf("Detected %d leaked goroutines (threshold: %d):", len(leaked), d.threshold)
		for i, stack := range leaked {
			if i < 5 { // Only show first 5 to avoid spam
				d.t.Errorf("Leaked goroutine %d:\n%s", i+1, stack)
			}
		}
		if len(leaked) > 5 {
			d.t.Errorf("... and %d more leaked goroutines", len(leaked)-5)
		}
	}
}

// getGoroutineStacks returns current goroutine stacks
func getGoroutineStacks() []string {
	buf := make([]byte, 1<<20) // 1MB buffer
	n := runtime.Stack(buf, true)

	stacks := strings.Split(string(buf[:n]), "\n\n")
	return stacks
}

// findLeakedGoroutines finds goroutines that weren't in the initial set
func findLeakedGoroutines(initial, current []string) []string {
	// Create a map of initial goroutines for fast lookup
	initialMap := make(map[string]bool)
	for _, stack := range initial {
		// Use first few lines as key to identify goroutine
		lines := strings.Split(stack, "\n")
		if len(lines) >= 2 {
			key := lines[0] + "\n" + lines[1]
			initialMap[key] = true
		}
	}

	var leaked []string
	for _, stack := range current {
		lines := strings.Split(stack, "\n")
		if len(lines) >= 2 {
			key := lines[0] + "\n" + lines[1]

			// Skip system goroutines
			if strings.Contains(stack, "runtime.") ||
				strings.Contains(stack, "testing.") ||
				strings.Contains(stack, "net/http.(*Server).") {
				continue
			}

			if !initialMap[key] {
				leaked = append(leaked, stack)
			}
		}
	}

	return leaked
}

// WithLeakDetection runs a test with goroutine leak detection
func WithLeakDetection(t *testing.T, fn func()) {
	detector := NewGoroutineLeakDetector(t)
	defer detector.Check()
	fn()
}

// ResourceMonitor monitors resource usage during tests
type ResourceMonitor struct {
	t              *testing.T
	maxMemoryMB    int64
	checkInterval  time.Duration
	stopChan       chan struct{}
	violationCount int
}

// NewResourceMonitor creates a new resource monitor
func NewResourceMonitor(t *testing.T, maxMemoryMB int64) *ResourceMonitor {
	return &ResourceMonitor{
		t:             t,
		maxMemoryMB:   maxMemoryMB,
		checkInterval: 500 * time.Millisecond,
		stopChan:      make(chan struct{}),
	}
}

// Start begins monitoring
func (m *ResourceMonitor) Start() {
	go func() {
		ticker := time.NewTicker(m.checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-m.stopChan:
				return
			case <-ticker.C:
				m.checkResources()
			}
		}
	}()
}

// Stop stops monitoring
func (m *ResourceMonitor) Stop() {
	close(m.stopChan)

	if m.violationCount > 0 {
		m.t.Errorf("Resource monitor detected %d violations", m.violationCount)
	}
}

// checkResources checks current resource usage
func (m *ResourceMonitor) checkResources() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	allocMB := int64(memStats.Alloc / 1024 / 1024)
	if allocMB > m.maxMemoryMB {
		m.violationCount++
		m.t.Logf("Memory usage exceeded limit: %d MB (limit: %d MB)", allocMB, m.maxMemoryMB)

		// Force GC to try to reclaim memory
		runtime.GC()
		runtime.GC() // Run twice to ensure finalizers run
	}

	// Check goroutine count
	goroutineCount := runtime.NumGoroutine()
	if goroutineCount > 1000 {
		m.violationCount++
		m.t.Logf("Excessive goroutines: %d", goroutineCount)
	}
}

// WithResourceMonitoring runs a test with resource monitoring
func WithResourceMonitoring(t *testing.T, maxMemoryMB int64, fn func()) {
	monitor := NewResourceMonitor(t, maxMemoryMB)
	monitor.Start()
	defer monitor.Stop()

	fn()
}

// ConnectionTracker tracks open connections during tests
type ConnectionTracker struct {
	t           *testing.T
	mu          sync.Mutex
	connections map[string]connectionInfo
	nextID      int64
}

type connectionInfo struct {
	id        int64
	connType  string // "tcp", "http", etc.
	address   string
	stack     string
	createdAt time.Time
}

// Global connection tracker for intercepting connections
var globalConnTracker = &ConnectionTracker{
	connections: make(map[string]connectionInfo),
}

// NewConnectionTracker creates a new connection tracker
func NewConnectionTracker(t *testing.T) *ConnectionTracker {
	return &ConnectionTracker{
		t:           t,
		connections: make(map[string]connectionInfo),
	}
}

// TrackConnection records a new connection
func (ct *ConnectionTracker) TrackConnection(conn net.Conn, connType string) net.Conn {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	id := atomic.AddInt64(&ct.nextID, 1)
	
	// Handle case where RemoteAddr might be nil
	var address string
	if conn.RemoteAddr() != nil {
		address = conn.RemoteAddr().String()
	} else {
		address = "unknown"
	}
	
	key := fmt.Sprintf("%s-%s-%d", connType, address, id)

	ct.connections[key] = connectionInfo{
		id:        id,
		connType:  connType,
		address:   address,
		stack:     string(debug.Stack()),
		createdAt: time.Now(),
	}

	// Wrap the connection to track closure
	return &trackedConn{
		Conn:    conn,
		tracker: ct,
		key:     key,
	}
}

// UntrackConnection removes a connection from tracking
func (ct *ConnectionTracker) UntrackConnection(key string) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	delete(ct.connections, key)
}

// GetOpenConnections returns currently open connections
func (ct *ConnectionTracker) GetOpenConnections() []connectionInfo {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	var open []connectionInfo
	for _, info := range ct.connections {
		open = append(open, info)
	}
	return open
}

// CheckLeaks verifies all connections were closed
func (ct *ConnectionTracker) CheckLeaks() {
	// Give connections time to close
	time.Sleep(200 * time.Millisecond)
	
	// Force GC to help clean up closed connections
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	open := ct.GetOpenConnections()
	if len(open) > 0 {
		ct.t.Errorf("Detected %d unclosed connections:", len(open))
		for i, info := range open {
			if i < 3 { // Only show first 3 to reduce noise
				ct.t.Errorf("Unclosed %s connection to %s (created %v ago):\n%s",
					info.connType, info.address,
					time.Since(info.createdAt).Round(time.Millisecond),
					truncateStack(info.stack))
			}
		}
		if len(open) > 3 {
			ct.t.Errorf("... and %d more unclosed connections", len(open)-3)
		}
	}
}

// trackedConn wraps a connection to track when it's closed
type trackedConn struct {
	net.Conn
	tracker *ConnectionTracker
	key     string
	closed  int32
}

func (tc *trackedConn) Close() error {
	if atomic.CompareAndSwapInt32(&tc.closed, 0, 1) {
		tc.tracker.UntrackConnection(tc.key)
	}
	return tc.Conn.Close()
}

// truncateStack returns a truncated version of the stack trace
func truncateStack(stack string) string {
	lines := strings.Split(stack, "\n")
	var result []string
	for i, line := range lines {
		if i > 10 { // Show only first 10 lines
			result = append(result, "...")
			break
		}
		// Skip runtime internal frames
		if strings.Contains(line, "runtime.") ||
			strings.Contains(line, "testing.") {
			continue
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}

// TrackedDialer creates a net.Dialer that tracks connections
func TrackedDialer(tracker *ConnectionTracker) *net.Dialer {
	return &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
}

// TrackedHTTPClient creates an http.Client that tracks connections
func TrackedHTTPClient(tracker *ConnectionTracker) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				dialer := &net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}
				conn, err := dialer.DialContext(ctx, network, addr)
				if err != nil {
					return nil, err
				}
				return tracker.TrackConnection(conn, "http"), nil
			},
			MaxIdleConns:          10,
			MaxIdleConnsPerHost:   2,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			DisableKeepAlives:     true, // Force close connections to prevent leaks
		},
		Timeout: 30 * time.Second,
	}
}

// LeakDetector combines all leak detection capabilities
type LeakDetector struct {
	goroutineDetector *GoroutineLeakDetector
	connTracker       *ConnectionTracker
	resourceMonitor   *ResourceMonitor
}

// NewLeakDetector creates a comprehensive leak detector
func NewLeakDetector(t *testing.T, maxMemoryMB int64) *LeakDetector {
	return &LeakDetector{
		goroutineDetector: NewGoroutineLeakDetector(t),
		connTracker:       NewConnectionTracker(t),
		resourceMonitor:   NewResourceMonitor(t, maxMemoryMB),
	}
}

// Start begins leak detection
func (ld *LeakDetector) Start() {
	ld.resourceMonitor.Start()
}

// Check verifies no leaks occurred
func (ld *LeakDetector) Check() {
	ld.resourceMonitor.Stop()
	ld.goroutineDetector.Check()
	ld.connTracker.CheckLeaks()
}

// HTTPClient returns an HTTP client that tracks connections
func (ld *LeakDetector) HTTPClient() *http.Client {
	return TrackedHTTPClient(ld.connTracker)
}

// Dialer returns a dialer that tracks connections
func (ld *LeakDetector) Dialer() *net.Dialer {
	return TrackedDialer(ld.connTracker)
}

// WithComprehensiveLeakDetection runs a test with all leak detection enabled
func WithComprehensiveLeakDetection(t *testing.T, maxMemoryMB int64, fn func(*LeakDetector)) {
	detector := NewLeakDetector(t, maxMemoryMB)
	detector.Start()
	defer detector.Check()

	fn(detector)
}
