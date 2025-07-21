package events

import (
	"time"
)

// EventType represents the type of event
type EventType string

const (
	// Request events
	EventRequestReceived  EventType = "request.received"
	EventRequestRouted    EventType = "request.routed"
	EventRequestCompleted EventType = "request.completed"
	EventRequestFailed    EventType = "request.failed"
	
	// Provider events
	EventProviderHealthy   EventType = "provider.healthy"
	EventProviderUnhealthy EventType = "provider.unhealthy"
	EventProviderAdded     EventType = "provider.added"
	EventProviderRemoved   EventType = "provider.removed"
	EventProviderUpdated   EventType = "provider.updated"
	
	// System events
	EventSystemStarted      EventType = "system.started"
	EventSystemStopping     EventType = "system.stopping"
	EventSystemStopped      EventType = "system.stopped"
	EventSystemError        EventType = "system.error"
	EventSystemStateChanged EventType = "system.state_changed"
	
	// Performance events
	EventRateLimitExceeded   EventType = "performance.rate_limit_exceeded"
	EventCircuitBreakerOpen  EventType = "performance.circuit_breaker_open"
	EventCircuitBreakerClose EventType = "performance.circuit_breaker_close"
	EventResourceLimitHit    EventType = "performance.resource_limit_hit"
	
	// Configuration events
	EventConfigReloaded EventType = "config.reloaded"
	EventConfigError    EventType = "config.error"
)

// Event represents a system event
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
	Error     error                  `json:"error,omitempty"`
}

// EventHandler is a function that handles events
type EventHandler func(event Event)

// EventFilter is a function that filters events
type EventFilter func(event Event) bool

// Subscription represents an event subscription
type Subscription struct {
	ID       string
	Types    []EventType
	Handler  EventHandler
	Filter   EventFilter
	Priority int
	Async    bool
}

// EventBusConfig represents configuration for the event bus
type EventBusConfig struct {
	BufferSize       int           `json:"buffer_size"`
	Workers          int           `json:"workers"`
	MaxRetries       int           `json:"max_retries"`
	RetryDelay       time.Duration `json:"retry_delay"`
	PersistEvents    bool          `json:"persist_events"`
	EventTTL         time.Duration `json:"event_ttl"`
	EnableMetrics    bool          `json:"enable_metrics"`
	EnableAuditLog   bool          `json:"enable_audit_log"`
}

// DefaultEventBusConfig returns default event bus configuration
func DefaultEventBusConfig() *EventBusConfig {
	return &EventBusConfig{
		BufferSize:    10000,
		Workers:       10,
		MaxRetries:    3,
		RetryDelay:    100 * time.Millisecond,
		PersistEvents: false,
		EventTTL:      24 * time.Hour,
		EnableMetrics: true,
		EnableAuditLog: false,
	}
}

// RequestEvent represents a request-related event
type RequestEvent struct {
	RequestID    string        `json:"request_id"`
	Provider     string        `json:"provider"`
	Model        string        `json:"model"`
	Latency      time.Duration `json:"latency,omitempty"`
	TokensIn     int           `json:"tokens_in,omitempty"`
	TokensOut    int           `json:"tokens_out,omitempty"`
	StatusCode   int           `json:"status_code,omitempty"`
	Error        string        `json:"error,omitempty"`
}

// ProviderEvent represents a provider-related event
type ProviderEvent struct {
	ProviderName string        `json:"provider_name"`
	Healthy      bool          `json:"healthy"`
	ResponseTime time.Duration `json:"response_time,omitempty"`
	ErrorMessage string        `json:"error_message,omitempty"`
}

// SystemEvent represents a system-related event
type SystemEvent struct {
	Component string `json:"component"`
	OldState  string `json:"old_state,omitempty"`
	NewState  string `json:"new_state,omitempty"`
	Message   string `json:"message,omitempty"`
}

// PerformanceEvent represents a performance-related event
type PerformanceEvent struct {
	Metric    string  `json:"metric"`
	Value     float64 `json:"value"`
	Threshold float64 `json:"threshold,omitempty"`
	Provider  string  `json:"provider,omitempty"`
	Message   string  `json:"message,omitempty"`
}