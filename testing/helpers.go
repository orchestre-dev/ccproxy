package testing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestHelpers provides additional test helper functions
type TestHelpers struct{}

// NewTestHelpers creates a new test helpers instance
func NewTestHelpers() *TestHelpers {
	return &TestHelpers{}
}

// GenerateTestData generates test data of specified size
func (th *TestHelpers) GenerateTestData(size int) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}
	return data
}

// CreateTestRequest creates a test HTTP request
func (th *TestHelpers) CreateTestRequest(method, path string, body interface{}) *http.Request {
	var bodyReader io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(data)
	}
	
	req, _ := http.NewRequest(method, path, bodyReader)
	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	return req
}

// GetFreePort returns a free port by binding to port 0 and letting the OS assign a port
func GetFreePort() (int, error) {
	// Listen on port 0 to get a free port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	
	// Get the port that was assigned
	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

// GetFreePorts returns multiple free ports
func GetFreePorts(count int) ([]int, error) {
	ports := make([]int, 0, count)
	listeners := make([]net.Listener, 0, count)
	
	// Get all ports first
	for i := 0; i < count; i++ {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			// Close any listeners we've already opened
			for _, l := range listeners {
				l.Close()
			}
			return nil, err
		}
		listeners = append(listeners, listener)
		
		addr := listener.Addr().(*net.TCPAddr)
		ports = append(ports, addr.Port)
	}
	
	// Close all listeners
	for _, l := range listeners {
		l.Close()
	}
	
	return ports, nil
}

// EnsurePortFree ensures a port is free before use
func EnsurePortFree(port int) error {
	// Try to listen on the port
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		// Port is in use
		return fmt.Errorf("port %d is already in use", port)
	}
	
	// Port is free, close the listener
	listener.Close()
	
	// Give OS time to fully release the port
	time.Sleep(100 * time.Millisecond)
	
	return nil
}

// ProgressReporter tracks and reports test progress
type ProgressReporter struct {
	t            *testing.T
	name         string
	totalSteps   int
	currentStep  int32
	startTime    time.Time
	stepTimes    []time.Duration
	mu           sync.Mutex
	ticker       *time.Ticker
	done         chan bool
	updateEvery  time.Duration
}

// NewProgressReporter creates a new progress reporter
func NewProgressReporter(t *testing.T, name string, totalSteps int) *ProgressReporter {
	pr := &ProgressReporter{
		t:           t,
		name:        name,
		totalSteps:  totalSteps,
		currentStep: 0,
		startTime:   time.Now(),
		stepTimes:   make([]time.Duration, 0, totalSteps),
		done:        make(chan bool),
		updateEvery: 5 * time.Second, // Default to 5 second updates
	}
	
	// Start periodic updates
	pr.startPeriodicUpdates()
	
	return pr
}

// SetUpdateInterval sets how often progress is reported
func (pr *ProgressReporter) SetUpdateInterval(interval time.Duration) {
	pr.updateEvery = interval
}

// Step marks the completion of a step
func (pr *ProgressReporter) Step(description string) {
	stepNum := atomic.AddInt32(&pr.currentStep, 1)
	
	pr.mu.Lock()
	if len(pr.stepTimes) > 0 {
		pr.stepTimes = append(pr.stepTimes, time.Since(pr.startTime)-sumDurations(pr.stepTimes))
	} else {
		pr.stepTimes = append(pr.stepTimes, time.Since(pr.startTime))
	}
	pr.mu.Unlock()
	
	pr.reportProgress(fmt.Sprintf("Step %d/%d: %s", stepNum, pr.totalSteps, description))
}

// Complete marks the test as complete
func (pr *ProgressReporter) Complete() {
	if pr.ticker != nil {
		pr.ticker.Stop()
	}
	close(pr.done)
	
	elapsed := time.Since(pr.startTime)
	pr.t.Logf("[%s] Completed in %s", pr.name, elapsed.Round(time.Millisecond))
}

// startPeriodicUpdates starts the periodic progress updates
func (pr *ProgressReporter) startPeriodicUpdates() {
	pr.ticker = time.NewTicker(pr.updateEvery)
	
	go func() {
		for {
			select {
			case <-pr.ticker.C:
				current := atomic.LoadInt32(&pr.currentStep)
				if current < int32(pr.totalSteps) {
					pr.reportProgress("")
				}
			case <-pr.done:
				return
			}
		}
	}()
}

// reportProgress reports the current progress
func (pr *ProgressReporter) reportProgress(message string) {
	current := atomic.LoadInt32(&pr.currentStep)
	elapsed := time.Since(pr.startTime)
	
	// Calculate ETA
	var eta time.Duration
	if current > 0 {
		avgTimePerStep := elapsed / time.Duration(current)
		remainingSteps := pr.totalSteps - int(current)
		eta = avgTimePerStep * time.Duration(remainingSteps)
	}
	
	progressMsg := fmt.Sprintf("[%s] Progress: %d/%d (%.1f%%) - Elapsed: %s",
		pr.name,
		current,
		pr.totalSteps,
		float64(current)/float64(pr.totalSteps)*100,
		elapsed.Round(time.Second))
	
	if eta > 0 {
		progressMsg += fmt.Sprintf(" - ETA: %s", eta.Round(time.Second))
	}
	
	if message != "" {
		progressMsg += " - " + message
	}
	
	pr.t.Log(progressMsg)
}

// sumDurations calculates the sum of durations
func sumDurations(durations []time.Duration) time.Duration {
	var sum time.Duration
	for _, d := range durations {
		sum += d
	}
	return sum
}

// TestSuiteReporter tracks progress across multiple test suites
type TestSuiteReporter struct {
	t              *testing.T
	suiteName      string
	tests          []string
	currentTest    int32
	startTime      time.Time
	testStartTimes map[string]time.Time
	testDurations  map[string]time.Duration
	mu             sync.Mutex
}

// NewTestSuiteReporter creates a new test suite reporter
func NewTestSuiteReporter(t *testing.T, suiteName string, tests []string) *TestSuiteReporter {
	tsr := &TestSuiteReporter{
		t:              t,
		suiteName:      suiteName,
		tests:          tests,
		currentTest:    -1,
		startTime:      time.Now(),
		testStartTimes: make(map[string]time.Time),
		testDurations:  make(map[string]time.Duration),
	}
	
	t.Logf("[%s] Starting test suite with %d tests", suiteName, len(tests))
	return tsr
}

// StartTest marks the start of a test
func (tsr *TestSuiteReporter) StartTest(testName string) {
	testNum := atomic.AddInt32(&tsr.currentTest, 1)
	
	tsr.mu.Lock()
	tsr.testStartTimes[testName] = time.Now()
	tsr.mu.Unlock()
	
	elapsed := time.Since(tsr.startTime)
	tsr.t.Logf("[%s] Starting test %d/%d: %s (elapsed: %s)",
		tsr.suiteName,
		testNum+1,
		len(tsr.tests),
		testName,
		elapsed.Round(time.Second))
}

// EndTest marks the end of a test
func (tsr *TestSuiteReporter) EndTest(testName string, passed bool) {
	tsr.mu.Lock()
	if startTime, ok := tsr.testStartTimes[testName]; ok {
		duration := time.Since(startTime)
		tsr.testDurations[testName] = duration
	}
	tsr.mu.Unlock()
	
	status := "PASSED"
	if !passed {
		status = "FAILED"
	}
	
	duration := tsr.testDurations[testName]
	tsr.t.Logf("[%s] Test %s: %s (duration: %s)",
		tsr.suiteName,
		testName,
		status,
		duration.Round(time.Millisecond))
}

// Complete prints a summary of the test suite
func (tsr *TestSuiteReporter) Complete() {
	totalDuration := time.Since(tsr.startTime)
	
	tsr.t.Logf("[%s] Test suite completed in %s", tsr.suiteName, totalDuration.Round(time.Second))
	
	// Print slowest tests
	tsr.printSlowestTests(5)
}

// printSlowestTests prints the N slowest tests
func (tsr *TestSuiteReporter) printSlowestTests(n int) {
	if len(tsr.testDurations) == 0 {
		return
	}
	
	// Convert map to slice for sorting
	type testDuration struct {
		name     string
		duration time.Duration
	}
	
	tests := make([]testDuration, 0, len(tsr.testDurations))
	for name, duration := range tsr.testDurations {
		tests = append(tests, testDuration{name, duration})
	}
	
	// Sort by duration (descending)
	for i := 0; i < len(tests)-1; i++ {
		for j := i + 1; j < len(tests); j++ {
			if tests[i].duration < tests[j].duration {
				tests[i], tests[j] = tests[j], tests[i]
			}
		}
	}
	
	// Print top N
	limit := n
	if limit > len(tests) {
		limit = len(tests)
	}
	
	tsr.t.Logf("[%s] Slowest %d tests:", tsr.suiteName, limit)
	for i := 0; i < limit; i++ {
		tsr.t.Logf("  %d. %s: %s", i+1, tests[i].name, tests[i].duration.Round(time.Millisecond))
	}
}

// SpinnerReporter provides a visual spinner for long operations
type SpinnerReporter struct {
	t          *testing.T
	message    string
	done       chan bool
	ticker     *time.Ticker
	spinChars  []string
	spinIndex  int32
}

// NewSpinnerReporter creates a new spinner reporter
func NewSpinnerReporter(t *testing.T, message string) *SpinnerReporter {
	sr := &SpinnerReporter{
		t:         t,
		message:   message,
		done:      make(chan bool),
		spinChars: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		spinIndex: 0,
	}
	
	sr.start()
	return sr
}

// start begins the spinner animation
func (sr *SpinnerReporter) start() {
	sr.ticker = time.NewTicker(100 * time.Millisecond)
	
	go func() {
		for {
			select {
			case <-sr.ticker.C:
				index := atomic.AddInt32(&sr.spinIndex, 1) % int32(len(sr.spinChars))
				sr.t.Logf("%s %s", sr.spinChars[index], sr.message)
			case <-sr.done:
				return
			}
		}
	}()
}

// Stop stops the spinner
func (sr *SpinnerReporter) Stop() {
	if sr.ticker != nil {
		sr.ticker.Stop()
	}
	close(sr.done)
}

// LongTestReporter provides comprehensive reporting for long-running tests
type LongTestReporter struct {
	t            *testing.T
	name         string
	phases       []string
	currentPhase int32
	phaseStart   time.Time
	startTime    time.Time
	mu           sync.Mutex
}

// NewLongTestReporter creates a reporter for long-running tests
func NewLongTestReporter(t *testing.T, name string, phases []string) *LongTestReporter {
	ltr := &LongTestReporter{
		t:            t,
		name:         name,
		phases:       phases,
		currentPhase: -1,
		startTime:    time.Now(),
	}
	
	t.Logf("[%s] Starting long-running test with %d phases", name, len(phases))
	for i, phase := range phases {
		t.Logf("  Phase %d: %s", i+1, phase)
	}
	
	return ltr
}

// StartPhase marks the beginning of a test phase
func (ltr *LongTestReporter) StartPhase(phaseName string) {
	phaseNum := atomic.AddInt32(&ltr.currentPhase, 1)
	
	ltr.mu.Lock()
	ltr.phaseStart = time.Now()
	ltr.mu.Unlock()
	
	elapsed := time.Since(ltr.startTime)
	ltr.t.Logf("[%s] Phase %d/%d: Starting '%s' (total elapsed: %s)",
		ltr.name,
		phaseNum+1,
		len(ltr.phases),
		phaseName,
		elapsed.Round(time.Second))
}

// EndPhase marks the completion of a test phase
func (ltr *LongTestReporter) EndPhase(phaseName string) {
	ltr.mu.Lock()
	phaseDuration := time.Since(ltr.phaseStart)
	ltr.mu.Unlock()
	
	current := atomic.LoadInt32(&ltr.currentPhase)
	ltr.t.Logf("[%s] Phase %d/%d: Completed '%s' in %s",
		ltr.name,
		current+1,
		len(ltr.phases),
		phaseName,
		phaseDuration.Round(time.Millisecond))
}

// Complete marks the test as complete
func (ltr *LongTestReporter) Complete() {
	totalDuration := time.Since(ltr.startTime)
	ltr.t.Logf("[%s] Test completed in %s", ltr.name, totalDuration.Round(time.Second))
}