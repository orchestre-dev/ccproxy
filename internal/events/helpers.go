package events

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Helper functions for creating common events

// NewRequestEvent creates a new request event
func NewRequestEvent(eventType EventType, requestID string, provider, model string) Event {
	return Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now(),
		Source:    "request-handler",
		Data: map[string]interface{}{
			"request_id": requestID,
			"provider":   provider,
			"model":      model,
		},
	}
}

// NewRequestReceivedEvent creates a request received event
func NewRequestReceivedEvent(requestID string, provider, model string, tokensIn int) Event {
	event := NewRequestEvent(EventRequestReceived, requestID, provider, model)
	event.Data["tokens_in"] = tokensIn
	return event
}

// NewRequestCompletedEvent creates a request completed event
func NewRequestCompletedEvent(requestID string, provider, model string, latency time.Duration, tokensOut int, statusCode int) Event {
	event := NewRequestEvent(EventRequestCompleted, requestID, provider, model)
	event.Data["latency"] = latency.Milliseconds()
	event.Data["tokens_out"] = tokensOut
	event.Data["status_code"] = statusCode
	return event
}

// NewRequestFailedEvent creates a request failed event
func NewRequestFailedEvent(requestID string, provider, model string, err error, statusCode int) Event {
	event := NewRequestEvent(EventRequestFailed, requestID, provider, model)
	event.Data["error"] = err.Error()
	event.Data["status_code"] = statusCode
	event.Error = err
	return event
}

// NewProviderEvent creates a new provider event
func NewProviderEvent(eventType EventType, providerName string) Event {
	return Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now(),
		Source:    "provider-service",
		Data: map[string]interface{}{
			"provider_name": providerName,
		},
	}
}

// NewProviderHealthyEvent creates a provider healthy event
func NewProviderHealthyEvent(providerName string, responseTime time.Duration) Event {
	event := NewProviderEvent(EventProviderHealthy, providerName)
	event.Data["healthy"] = true
	event.Data["response_time"] = responseTime.Milliseconds()
	return event
}

// NewProviderUnhealthyEvent creates a provider unhealthy event
func NewProviderUnhealthyEvent(providerName string, err error) Event {
	event := NewProviderEvent(EventProviderUnhealthy, providerName)
	event.Data["healthy"] = false
	event.Data["error"] = err.Error()
	event.Error = err
	return event
}

// NewSystemEvent creates a new system event
func NewSystemEvent(eventType EventType, component string) Event {
	return Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now(),
		Source:    "system",
		Data: map[string]interface{}{
			"component": component,
		},
	}
}

// NewSystemStateChangedEvent creates a system state changed event
func NewSystemStateChangedEvent(component, oldState, newState string) Event {
	event := NewSystemEvent(EventSystemStateChanged, component)
	event.Data["old_state"] = oldState
	event.Data["new_state"] = newState
	return event
}

// NewPerformanceEvent creates a new performance event
func NewPerformanceEvent(eventType EventType, metric string, value float64) Event {
	return Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now(),
		Source:    "performance-monitor",
		Data: map[string]interface{}{
			"metric": metric,
			"value":  value,
		},
	}
}

// NewRateLimitExceededEvent creates a rate limit exceeded event
func NewRateLimitExceededEvent(provider string, limit int, current int) Event {
	event := NewPerformanceEvent(EventRateLimitExceeded, "rate_limit", float64(current))
	event.Data["provider"] = provider
	event.Data["limit"] = limit
	event.Data["message"] = fmt.Sprintf("Rate limit exceeded for %s: %d/%d", provider, current, limit)
	return event
}

// NewCircuitBreakerEvent creates a circuit breaker event
func NewCircuitBreakerEvent(open bool, provider string, errorRate float64) Event {
	var eventType EventType
	if open {
		eventType = EventCircuitBreakerOpen
	} else {
		eventType = EventCircuitBreakerClose
	}
	
	event := NewPerformanceEvent(eventType, "circuit_breaker", errorRate)
	event.Data["provider"] = provider
	event.Data["open"] = open
	return event
}

// EventLogger provides structured logging for events
type EventLogger struct {
	eventBus *EventBus
}

// NewEventLogger creates a new event logger
func NewEventLogger(eventBus *EventBus) *EventLogger {
	return &EventLogger{
		eventBus: eventBus,
	}
}

// LogRequestReceived logs a request received event
func (el *EventLogger) LogRequestReceived(requestID string, provider, model string, tokensIn int) {
	event := NewRequestReceivedEvent(requestID, provider, model, tokensIn)
	el.eventBus.PublishEvent(event)
}

// LogRequestCompleted logs a request completed event
func (el *EventLogger) LogRequestCompleted(requestID string, provider, model string, latency time.Duration, tokensOut int, statusCode int) {
	event := NewRequestCompletedEvent(requestID, provider, model, latency, tokensOut, statusCode)
	el.eventBus.PublishEvent(event)
}

// LogRequestFailed logs a request failed event
func (el *EventLogger) LogRequestFailed(requestID string, provider, model string, err error, statusCode int) {
	event := NewRequestFailedEvent(requestID, provider, model, err, statusCode)
	el.eventBus.PublishEvent(event)
}

// LogProviderHealthy logs a provider healthy event
func (el *EventLogger) LogProviderHealthy(providerName string, responseTime time.Duration) {
	event := NewProviderHealthyEvent(providerName, responseTime)
	el.eventBus.PublishEvent(event)
}

// LogProviderUnhealthy logs a provider unhealthy event
func (el *EventLogger) LogProviderUnhealthy(providerName string, err error) {
	event := NewProviderUnhealthyEvent(providerName, err)
	el.eventBus.PublishEvent(event)
}

// LogSystemStateChanged logs a system state changed event
func (el *EventLogger) LogSystemStateChanged(component, oldState, newState string) {
	event := NewSystemStateChangedEvent(component, oldState, newState)
	el.eventBus.PublishEvent(event)
}

// LogRateLimitExceeded logs a rate limit exceeded event
func (el *EventLogger) LogRateLimitExceeded(provider string, limit int, current int) {
	event := NewRateLimitExceededEvent(provider, limit, current)
	el.eventBus.PublishEvent(event)
}

// LogCircuitBreakerStateChanged logs a circuit breaker state change
func (el *EventLogger) LogCircuitBreakerStateChanged(open bool, provider string, errorRate float64) {
	event := NewCircuitBreakerEvent(open, provider, errorRate)
	el.eventBus.PublishEvent(event)
}