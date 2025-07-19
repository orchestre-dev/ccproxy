package events

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/musistudio/ccproxy/internal/utils"
)

// EventBus manages event publishing and subscription
type EventBus struct {
	config          *EventBusConfig
	subscriptions   map[string]*Subscription
	subsByType      map[EventType][]*Subscription
	eventChan       chan Event
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	mu              sync.RWMutex
	
	// Metrics
	eventsPublished int64
	eventsProcessed int64
	eventsDropped   int64
	handlerErrors   int64
	
	// Event history for debugging
	eventHistory    []Event
	historyMu       sync.RWMutex
	maxHistorySize  int
}

// NewEventBus creates a new event bus
func NewEventBus(config *EventBusConfig) *EventBus {
	if config == nil {
		config = DefaultEventBusConfig()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	eb := &EventBus{
		config:         config,
		subscriptions:  make(map[string]*Subscription),
		subsByType:     make(map[EventType][]*Subscription),
		eventChan:      make(chan Event, config.BufferSize),
		ctx:            ctx,
		cancel:         cancel,
		eventHistory:   make([]Event, 0, 1000),
		maxHistorySize: 1000,
	}
	
	return eb
}

// Start begins processing events
func (eb *EventBus) Start() {
	// Start worker pool
	for i := 0; i < eb.config.Workers; i++ {
		eb.wg.Add(1)
		go eb.worker(i)
	}
	
	// Start metrics reporter if enabled
	if eb.config.EnableMetrics {
		eb.wg.Add(1)
		go eb.metricsReporter()
	}
	
	utils.GetLogger().Info("Event bus started")
}

// Stop stops the event bus
func (eb *EventBus) Stop() {
	utils.GetLogger().Info("Stopping event bus...")
	
	// Cancel context to stop workers
	eb.cancel()
	
	// Wait for workers to finish processing
	eb.wg.Wait()
	
	// Now safe to close event channel
	close(eb.eventChan)
	
	utils.GetLogger().Info("Event bus stopped")
}

// Subscribe creates a new subscription
func (eb *EventBus) Subscribe(types []EventType, handler EventHandler, options ...SubscriptionOption) string {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	
	// Create subscription
	sub := &Subscription{
		ID:       uuid.New().String(),
		Types:    types,
		Handler:  handler,
		Priority: 0,
		Async:    true,
	}
	
	// Apply options
	for _, opt := range options {
		opt(sub)
	}
	
	// Add to subscriptions map
	eb.subscriptions[sub.ID] = sub
	
	// Add to type-based index
	for _, eventType := range types {
		eb.subsByType[eventType] = append(eb.subsByType[eventType], sub)
		// Sort by priority
		eb.sortSubscriptionsByPriority(eventType)
	}
	
	utils.GetLogger().Debugf("Added subscription %s for event types: %v", sub.ID, types)
	
	return sub.ID
}

// Unsubscribe removes a subscription
func (eb *EventBus) Unsubscribe(subscriptionID string) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	
	sub, exists := eb.subscriptions[subscriptionID]
	if !exists {
		return fmt.Errorf("subscription %s not found", subscriptionID)
	}
	
	// Remove from subscriptions map
	delete(eb.subscriptions, subscriptionID)
	
	// Remove from type-based index
	for _, eventType := range sub.Types {
		subs := eb.subsByType[eventType]
		for i, s := range subs {
			if s.ID == subscriptionID {
				eb.subsByType[eventType] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
	}
	
	utils.GetLogger().Debugf("Removed subscription %s", subscriptionID)
	
	return nil
}

// Publish publishes an event
func (eb *EventBus) Publish(eventType EventType, source string, data map[string]interface{}) {
	event := Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now(),
		Source:    source,
		Data:      data,
	}
	
	eb.PublishEvent(event)
}

// PublishEvent publishes a pre-constructed event
func (eb *EventBus) PublishEvent(event Event) {
	atomic.AddInt64(&eb.eventsPublished, 1)
	
	// Add to history
	eb.addToHistory(event)
	
	// Check if channel is closed
	select {
	case <-eb.ctx.Done():
		// Event bus is stopped, don't publish
		return
	default:
	}
	
	// Try to send to channel
	select {
	case eb.eventChan <- event:
		// Event sent successfully
	default:
		// Channel is full, drop event
		atomic.AddInt64(&eb.eventsDropped, 1)
		utils.GetLogger().Warnf("Event dropped: channel full (type: %s)", event.Type)
	}
}

// worker processes events from the channel
func (eb *EventBus) worker(id int) {
	defer eb.wg.Done()
	
	for {
		select {
		case <-eb.ctx.Done():
			return
		case event, ok := <-eb.eventChan:
			if !ok {
				return
			}
			
			eb.processEvent(event)
		}
	}
}

// processEvent processes a single event
func (eb *EventBus) processEvent(event Event) {
	atomic.AddInt64(&eb.eventsProcessed, 1)
	
	eb.mu.RLock()
	subs := eb.subsByType[event.Type]
	eb.mu.RUnlock()
	
	for _, sub := range subs {
		// Apply filter if present
		if sub.Filter != nil && !sub.Filter(event) {
			continue
		}
		
		if sub.Async {
			// Handle asynchronously
			go eb.handleEvent(sub, event)
		} else {
			// Handle synchronously
			eb.handleEvent(sub, event)
		}
	}
}

// handleEvent handles an event with a subscription
func (eb *EventBus) handleEvent(sub *Subscription, event Event) {
	defer func() {
		if r := recover(); r != nil {
			atomic.AddInt64(&eb.handlerErrors, 1)
			utils.GetLogger().Errorf("Event handler panic: %v (subscription: %s, event: %s)", r, sub.ID, event.ID)
		}
	}()
	
	// Retry logic
	var lastErr error
	for attempt := 0; attempt <= eb.config.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(eb.config.RetryDelay * time.Duration(attempt))
		}
		
		// Create a timeout context for the handler
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		// Run handler in goroutine to respect timeout
		done := make(chan error, 1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					done <- fmt.Errorf("handler panic: %v", r)
				}
			}()
			sub.Handler(event)
			done <- nil
		}()
		
		select {
		case err := <-done:
			if err == nil {
				return // Success
			}
			lastErr = err
		case <-ctx.Done():
			lastErr = fmt.Errorf("handler timeout")
		}
	}
	
	if lastErr != nil {
		atomic.AddInt64(&eb.handlerErrors, 1)
		utils.GetLogger().Errorf("Event handler failed after %d retries: %v (subscription: %s, event: %s)", 
			eb.config.MaxRetries, lastErr, sub.ID, event.ID)
	}
}

// GetMetrics returns current event bus metrics
func (eb *EventBus) GetMetrics() map[string]int64 {
	return map[string]int64{
		"events_published": atomic.LoadInt64(&eb.eventsPublished),
		"events_processed": atomic.LoadInt64(&eb.eventsProcessed),
		"events_dropped":   atomic.LoadInt64(&eb.eventsDropped),
		"handler_errors":   atomic.LoadInt64(&eb.handlerErrors),
		"subscriptions":    int64(len(eb.subscriptions)),
		"pending_events":   int64(len(eb.eventChan)),
	}
}

// GetEventHistory returns recent event history
func (eb *EventBus) GetEventHistory(limit int) []Event {
	eb.historyMu.RLock()
	defer eb.historyMu.RUnlock()
	
	if limit <= 0 || limit > len(eb.eventHistory) {
		limit = len(eb.eventHistory)
	}
	
	// Return most recent events
	start := len(eb.eventHistory) - limit
	if start < 0 {
		start = 0
	}
	
	result := make([]Event, limit)
	copy(result, eb.eventHistory[start:])
	return result
}

// addToHistory adds an event to the history buffer
func (eb *EventBus) addToHistory(event Event) {
	eb.historyMu.Lock()
	defer eb.historyMu.Unlock()
	
	eb.eventHistory = append(eb.eventHistory, event)
	
	// Trim history if needed
	if len(eb.eventHistory) > eb.maxHistorySize {
		// Keep last N events
		eb.eventHistory = eb.eventHistory[len(eb.eventHistory)-eb.maxHistorySize:]
	}
}

// sortSubscriptionsByPriority sorts subscriptions by priority (higher first)
func (eb *EventBus) sortSubscriptionsByPriority(eventType EventType) {
	subs := eb.subsByType[eventType]
	for i := 0; i < len(subs)-1; i++ {
		for j := i + 1; j < len(subs); j++ {
			if subs[i].Priority < subs[j].Priority {
				subs[i], subs[j] = subs[j], subs[i]
			}
		}
	}
}

// metricsReporter periodically logs metrics
func (eb *EventBus) metricsReporter() {
	defer eb.wg.Done()
	
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-eb.ctx.Done():
			return
		case <-ticker.C:
			metrics := eb.GetMetrics()
			utils.GetLogger().Infof("EventBus metrics: published=%d, processed=%d, dropped=%d, errors=%d, pending=%d",
				metrics["events_published"],
				metrics["events_processed"],
				metrics["events_dropped"],
				metrics["handler_errors"],
				metrics["pending_events"])
		}
	}
}

// SubscriptionOption is a function that configures a subscription
type SubscriptionOption func(*Subscription)

// WithPriority sets the subscription priority
func WithPriority(priority int) SubscriptionOption {
	return func(s *Subscription) {
		s.Priority = priority
	}
}

// WithFilter sets the subscription filter
func WithFilter(filter EventFilter) SubscriptionOption {
	return func(s *Subscription) {
		s.Filter = filter
	}
}

// WithSync makes the subscription synchronous
func WithSync() SubscriptionOption {
	return func(s *Subscription) {
		s.Async = false
	}
}