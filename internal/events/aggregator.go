package events

import (
	"sync"
	"time"
)

// EventAggregator aggregates events for analytics and reporting
type EventAggregator struct {
	eventBus *EventBus
	mu       sync.RWMutex
	
	// Request metrics
	requestCounts      map[string]int64 // provider -> count
	requestLatencies   map[string][]time.Duration
	requestErrors      map[string]int64
	tokenCounts        map[string]int64 // provider -> total tokens
	
	// Provider metrics
	providerHealth     map[string]bool
	providerLastCheck  map[string]time.Time
	
	// System metrics
	stateChanges       []StateChange
	rateLimitHits      int64
	circuitBreakerTrips int64
	
	// Time window
	windowStart time.Time
	windowSize  time.Duration
}

// StateChange represents a state change event
type StateChange struct {
	Timestamp time.Time
	Component string
	OldState  string
	NewState  string
}

// NewEventAggregator creates a new event aggregator
func NewEventAggregator(eventBus *EventBus, windowSize time.Duration) *EventAggregator {
	if windowSize == 0 {
		windowSize = 5 * time.Minute
	}
	
	agg := &EventAggregator{
		eventBus:          eventBus,
		requestCounts:     make(map[string]int64),
		requestLatencies:  make(map[string][]time.Duration),
		requestErrors:     make(map[string]int64),
		tokenCounts:       make(map[string]int64),
		providerHealth:    make(map[string]bool),
		providerLastCheck: make(map[string]time.Time),
		stateChanges:      make([]StateChange, 0),
		windowStart:       time.Now(),
		windowSize:        windowSize,
	}
	
	return agg
}

// Start begins aggregating events
func (ea *EventAggregator) Start() {
	// Subscribe to all relevant event types
	ea.eventBus.Subscribe([]EventType{
		EventRequestReceived,
		EventRequestCompleted,
		EventRequestFailed,
	}, ea.handleRequestEvent, WithPriority(10))
	
	ea.eventBus.Subscribe([]EventType{
		EventProviderHealthy,
		EventProviderUnhealthy,
	}, ea.handleProviderEvent, WithPriority(10))
	
	ea.eventBus.Subscribe([]EventType{
		EventSystemStateChanged,
	}, ea.handleSystemEvent, WithPriority(10))
	
	ea.eventBus.Subscribe([]EventType{
		EventRateLimitExceeded,
		EventCircuitBreakerOpen,
	}, ea.handlePerformanceEvent, WithPriority(10))
	
	// Start window rotation
	go ea.rotateWindow()
}

// handleRequestEvent processes request events
func (ea *EventAggregator) handleRequestEvent(event Event) {
	ea.mu.Lock()
	defer ea.mu.Unlock()
	
	provider, _ := event.Data["provider"].(string)
	if provider == "" {
		return
	}
	
	switch event.Type {
	case EventRequestReceived:
		ea.requestCounts[provider]++
		
	case EventRequestCompleted:
		if latencyMs, ok := event.Data["latency"].(int64); ok {
			latency := time.Duration(latencyMs) * time.Millisecond
			ea.requestLatencies[provider] = append(ea.requestLatencies[provider], latency)
		}
		if tokensOut, ok := event.Data["tokens_out"].(int); ok {
			ea.tokenCounts[provider] += int64(tokensOut)
		}
		
	case EventRequestFailed:
		ea.requestErrors[provider]++
	}
}

// handleProviderEvent processes provider events
func (ea *EventAggregator) handleProviderEvent(event Event) {
	ea.mu.Lock()
	defer ea.mu.Unlock()
	
	providerName, _ := event.Data["provider_name"].(string)
	if providerName == "" {
		return
	}
	
	switch event.Type {
	case EventProviderHealthy:
		ea.providerHealth[providerName] = true
		ea.providerLastCheck[providerName] = event.Timestamp
		
	case EventProviderUnhealthy:
		ea.providerHealth[providerName] = false
		ea.providerLastCheck[providerName] = event.Timestamp
	}
}

// handleSystemEvent processes system events
func (ea *EventAggregator) handleSystemEvent(event Event) {
	ea.mu.Lock()
	defer ea.mu.Unlock()
	
	if event.Type == EventSystemStateChanged {
		component, _ := event.Data["component"].(string)
		oldState, _ := event.Data["old_state"].(string)
		newState, _ := event.Data["new_state"].(string)
		
		ea.stateChanges = append(ea.stateChanges, StateChange{
			Timestamp: event.Timestamp,
			Component: component,
			OldState:  oldState,
			NewState:  newState,
		})
	}
}

// handlePerformanceEvent processes performance events
func (ea *EventAggregator) handlePerformanceEvent(event Event) {
	ea.mu.Lock()
	defer ea.mu.Unlock()
	
	switch event.Type {
	case EventRateLimitExceeded:
		ea.rateLimitHits++
		
	case EventCircuitBreakerOpen:
		ea.circuitBreakerTrips++
	}
}

// GetMetrics returns aggregated metrics
func (ea *EventAggregator) GetMetrics() AggregatedMetrics {
	ea.mu.RLock()
	defer ea.mu.RUnlock()
	
	// Calculate provider metrics
	providerMetrics := make(map[string]ProviderAggregatedMetrics)
	for provider, count := range ea.requestCounts {
		var avgLatency time.Duration
		if latencies := ea.requestLatencies[provider]; len(latencies) > 0 {
			var total time.Duration
			for _, l := range latencies {
				total += l
			}
			avgLatency = total / time.Duration(len(latencies))
		}
		
		errorCount := ea.requestErrors[provider]
		errorRate := float64(0)
		if count > 0 {
			errorRate = float64(errorCount) / float64(count)
		}
		
		providerMetrics[provider] = ProviderAggregatedMetrics{
			RequestCount:    count,
			ErrorCount:      errorCount,
			ErrorRate:       errorRate,
			AverageLatency:  avgLatency,
			TokensProcessed: ea.tokenCounts[provider],
			Healthy:         ea.providerHealth[provider],
			LastHealthCheck: ea.providerLastCheck[provider],
		}
	}
	
	return AggregatedMetrics{
		WindowStart:         ea.windowStart,
		WindowEnd:           time.Now(),
		ProviderMetrics:     providerMetrics,
		StateChanges:        ea.stateChanges,
		RateLimitHits:       ea.rateLimitHits,
		CircuitBreakerTrips: ea.circuitBreakerTrips,
	}
}

// rotateWindow periodically resets metrics
func (ea *EventAggregator) rotateWindow() {
	ticker := time.NewTicker(ea.windowSize)
	defer ticker.Stop()
	
	for range ticker.C {
		ea.mu.Lock()
		
		// Reset counters
		ea.requestCounts = make(map[string]int64)
		ea.requestLatencies = make(map[string][]time.Duration)
		ea.requestErrors = make(map[string]int64)
		ea.tokenCounts = make(map[string]int64)
		ea.stateChanges = make([]StateChange, 0)
		ea.rateLimitHits = 0
		ea.circuitBreakerTrips = 0
		ea.windowStart = time.Now()
		
		ea.mu.Unlock()
	}
}

// AggregatedMetrics represents aggregated event metrics
type AggregatedMetrics struct {
	WindowStart         time.Time                             `json:"window_start"`
	WindowEnd           time.Time                             `json:"window_end"`
	ProviderMetrics     map[string]ProviderAggregatedMetrics `json:"provider_metrics"`
	StateChanges        []StateChange                         `json:"state_changes"`
	RateLimitHits       int64                                 `json:"rate_limit_hits"`
	CircuitBreakerTrips int64                                 `json:"circuit_breaker_trips"`
}

// ProviderAggregatedMetrics represents aggregated metrics for a provider
type ProviderAggregatedMetrics struct {
	RequestCount    int64         `json:"request_count"`
	ErrorCount      int64         `json:"error_count"`
	ErrorRate       float64       `json:"error_rate"`
	AverageLatency  time.Duration `json:"average_latency"`
	TokensProcessed int64         `json:"tokens_processed"`
	Healthy         bool          `json:"healthy"`
	LastHealthCheck time.Time     `json:"last_health_check"`
}