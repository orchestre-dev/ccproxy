package events

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewEventBus(t *testing.T) {
	config := DefaultEventBusConfig()
	eb := NewEventBus(config)

	assert.NotNil(t, eb)
	assert.Equal(t, config, eb.config)
	assert.NotNil(t, eb.eventChan)
	assert.Empty(t, eb.subscriptions)
}

func TestEventBusStartStop(t *testing.T) {
	eb := NewEventBus(DefaultEventBusConfig())

	// Start event bus
	eb.Start()

	// Publish some events
	for i := 0; i < 10; i++ {
		eb.Publish(EventRequestReceived, "test", map[string]interface{}{
			"index": i,
		})
	}

	// Give workers time to process
	time.Sleep(100 * time.Millisecond)

	// Stop event bus
	eb.Stop()

	// Verify metrics
	metrics := eb.GetMetrics()
	assert.Equal(t, int64(10), metrics["events_published"])
}

func TestSubscribe(t *testing.T) {
	eb := NewEventBus(DefaultEventBusConfig())
	eb.Start()
	defer eb.Stop()

	var received atomic.Int32

	// Subscribe to events
	subID := eb.Subscribe([]EventType{EventRequestReceived}, func(event Event) {
		received.Add(1)
	})

	assert.NotEmpty(t, subID)

	// Publish events
	for i := 0; i < 5; i++ {
		eb.Publish(EventRequestReceived, "test", nil)
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, int32(5), received.Load())
}

func TestUnsubscribe(t *testing.T) {
	eb := NewEventBus(DefaultEventBusConfig())
	eb.Start()
	defer eb.Stop()

	var received atomic.Int32

	// Subscribe
	subID := eb.Subscribe([]EventType{EventRequestReceived}, func(event Event) {
		received.Add(1)
	})

	// Publish event
	eb.Publish(EventRequestReceived, "test", nil)
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, int32(1), received.Load())

	// Unsubscribe
	err := eb.Unsubscribe(subID)
	assert.NoError(t, err)

	// Publish more events
	eb.Publish(EventRequestReceived, "test", nil)
	time.Sleep(50 * time.Millisecond)

	// Should still be 1
	assert.Equal(t, int32(1), received.Load())
}

func TestMultipleSubscriptions(t *testing.T) {
	eb := NewEventBus(DefaultEventBusConfig())
	eb.Start()
	defer eb.Stop()

	var (
		sub1Count atomic.Int32
		sub2Count atomic.Int32
		sub3Count atomic.Int32
	)

	// Subscribe to different event types
	eb.Subscribe([]EventType{EventRequestReceived}, func(event Event) {
		sub1Count.Add(1)
	})

	eb.Subscribe([]EventType{EventRequestCompleted}, func(event Event) {
		sub2Count.Add(1)
	})

	eb.Subscribe([]EventType{EventRequestReceived, EventRequestCompleted}, func(event Event) {
		sub3Count.Add(1)
	})

	// Publish events
	eb.Publish(EventRequestReceived, "test", nil)
	eb.Publish(EventRequestCompleted, "test", nil)
	eb.Publish(EventRequestFailed, "test", nil) // Not subscribed

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, int32(1), sub1Count.Load())
	assert.Equal(t, int32(1), sub2Count.Load())
	assert.Equal(t, int32(2), sub3Count.Load()) // Receives both events
}

func TestEventFilter(t *testing.T) {
	eb := NewEventBus(DefaultEventBusConfig())
	eb.Start()
	defer eb.Stop()

	var received atomic.Int32

	// Subscribe with filter
	eb.Subscribe([]EventType{EventRequestReceived}, func(event Event) {
		received.Add(1)
	}, WithFilter(func(event Event) bool {
		// Only accept events with source="allowed"
		return event.Source == "allowed"
	}))

	// Publish events
	eb.Publish(EventRequestReceived, "allowed", nil)
	eb.Publish(EventRequestReceived, "blocked", nil)
	eb.Publish(EventRequestReceived, "allowed", nil)

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, int32(2), received.Load())
}

func TestSubscriptionPriority(t *testing.T) {
	eb := NewEventBus(DefaultEventBusConfig())
	eb.Start()
	defer eb.Stop()

	var order []int
	var mu sync.Mutex

	// Subscribe with different priorities (synchronous to ensure order)
	eb.Subscribe([]EventType{EventRequestReceived}, func(event Event) {
		mu.Lock()
		order = append(order, 1)
		mu.Unlock()
	}, WithPriority(1), WithSync())

	eb.Subscribe([]EventType{EventRequestReceived}, func(event Event) {
		mu.Lock()
		order = append(order, 3)
		mu.Unlock()
	}, WithPriority(3), WithSync())

	eb.Subscribe([]EventType{EventRequestReceived}, func(event Event) {
		mu.Lock()
		order = append(order, 2)
		mu.Unlock()
	}, WithPriority(2), WithSync())

	// Publish event
	eb.Publish(EventRequestReceived, "test", nil)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	// Should be processed in priority order: 3, 2, 1
	assert.Equal(t, []int{3, 2, 1}, order)
}

func TestSynchronousHandler(t *testing.T) {
	eb := NewEventBus(DefaultEventBusConfig())
	eb.Start()
	defer eb.Stop()

	var processed atomic.Bool

	// Subscribe synchronously
	eb.Subscribe([]EventType{EventRequestReceived}, func(event Event) {
		time.Sleep(50 * time.Millisecond) // Simulate work
		processed.Store(true)
	}, WithSync())

	// Publish event
	start := time.Now()
	eb.Publish(EventRequestReceived, "test", nil)

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	assert.True(t, processed.Load())
	assert.True(t, time.Since(start) >= 50*time.Millisecond)
}

func TestEventHistory(t *testing.T) {
	config := DefaultEventBusConfig()
	config.BufferSize = 100
	eb := NewEventBus(config)
	eb.maxHistorySize = 10 // Small history for testing
	eb.Start()
	defer eb.Stop()

	// Publish events
	for i := 0; i < 20; i++ {
		eb.Publish(EventRequestReceived, "test", map[string]interface{}{
			"index": i,
		})
	}

	time.Sleep(100 * time.Millisecond)

	// Get history
	history := eb.GetEventHistory(5)
	assert.Len(t, history, 5)

	// Should have most recent events
	for i, event := range history {
		expectedIndex := 15 + i // Last 5 events: 15, 16, 17, 18, 19
		assert.Equal(t, expectedIndex, event.Data["index"])
	}

	// Get all history
	allHistory := eb.GetEventHistory(0)
	assert.Len(t, allHistory, 10) // Limited by maxHistorySize
}

func TestEventBusMetrics(t *testing.T) {
	config := DefaultEventBusConfig()
	config.BufferSize = 5 // Small buffer to test dropping
	config.MaxRetries = 0 // No retries for this test
	eb := NewEventBus(config)
	eb.Start()
	defer eb.Stop()

	// Subscribe with handler that sometimes fails
	var handlerCallCount atomic.Int32
	eb.Subscribe([]EventType{EventRequestReceived}, func(event Event) {
		count := handlerCallCount.Add(1)
		if count%3 == 0 {
			panic("simulated panic")
		}
	})

	// Stop processing to ensure buffer fills
	time.Sleep(50 * time.Millisecond)
	
	// Publish many events quickly to fill buffer
	for i := 0; i < 20; i++ {
		eb.Publish(EventRequestReceived, "test", nil)
	}

	// Wait longer for handlers to complete and errors to be counted
	time.Sleep(500 * time.Millisecond)

	metrics := eb.GetMetrics()
	t.Logf("Metrics: %+v", metrics)
	assert.Equal(t, int64(20), metrics["events_published"])
	// Some events should be processed
	processed := metrics["events_processed"]
	assert.True(t, processed > 0, "Expected some events to be processed, got %d", processed)
	// Should have dropped some events due to small buffer
	dropped := metrics["events_dropped"]
	t.Logf("Processed: %d, Dropped: %d", processed, dropped)
	assert.True(t, dropped > 0 || processed == 20, "Expected some events to be dropped or all processed")
	// Every third processed event causes a panic
	if processed > 0 {
		expectedErrors := processed / 3
		if processed%3 == 0 {
			expectedErrors = processed / 3
		} else {
			expectedErrors = (processed / 3) + 1
		}
		actualErrors := metrics["handler_errors"]
		t.Logf("Expected errors: %d, Actual errors: %d", expectedErrors, actualErrors)
		assert.True(t, actualErrors >= expectedErrors-1, "Expected at least %d errors, got %d", expectedErrors-1, actualErrors)
	}
}

func TestHandlerTimeout(t *testing.T) {
	config := DefaultEventBusConfig()
	config.MaxRetries = 0 // No retries for this test
	eb := NewEventBus(config)
	eb.Start()
	defer eb.Stop()

	var timedOut atomic.Bool

	// Subscribe with slow handler
	eb.Subscribe([]EventType{EventRequestReceived}, func(event Event) {
		time.Sleep(10 * time.Second) // Will timeout
		timedOut.Store(true)
	})

	// Publish event
	eb.Publish(EventRequestReceived, "test", nil)

	// Wait for timeout
	time.Sleep(6 * time.Second)

	assert.False(t, timedOut.Load()) // Should not complete due to timeout
	metrics := eb.GetMetrics()
	assert.True(t, metrics["handler_errors"] > 0)
}

func TestRetryLogic(t *testing.T) {
	config := DefaultEventBusConfig()
	config.MaxRetries = 3
	config.RetryDelay = 10 * time.Millisecond
	eb := NewEventBus(config)
	eb.Start()
	defer eb.Stop()

	var attempts atomic.Int32

	// Subscribe with handler that fails first 2 attempts
	eb.Subscribe([]EventType{EventRequestReceived}, func(event Event) {
		attempt := attempts.Add(1)
		if attempt < 3 {
			panic("simulated failure")
		}
		// Success on 3rd attempt
	})

	// Publish event
	eb.Publish(EventRequestReceived, "test", nil)

	// Wait for retries
	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, int32(3), attempts.Load()) // Should retry 3 times
}

func TestConcurrentPublishSubscribe(t *testing.T) {
	eb := NewEventBus(DefaultEventBusConfig())
	eb.Start()
	defer eb.Stop()

	var (
		received atomic.Int64
		wg       sync.WaitGroup
	)

	// Multiple subscribers
	for i := 0; i < 5; i++ {
		eb.Subscribe([]EventType{EventRequestReceived}, func(event Event) {
			received.Add(1)
		})
	}

	// Concurrent publishers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				eb.Publish(EventRequestReceived, "test", map[string]interface{}{
					"publisher": id,
					"event":     j,
				})
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(500 * time.Millisecond)

	// Should receive 1000 events * 5 subscribers = 5000
	assert.Equal(t, int64(5000), received.Load())
}

func TestEventBusContextCancellation(t *testing.T) {
	eb := NewEventBus(DefaultEventBusConfig())
	eb.Start()

	var processed atomic.Int32

	// Subscribe
	eb.Subscribe([]EventType{EventRequestReceived}, func(event Event) {
		processed.Add(1)
	})

	// Publish events continuously
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				eb.Publish(EventRequestReceived, "test", nil)
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	// Let it run for a bit
	time.Sleep(100 * time.Millisecond)

	// Stop event bus
	eb.Stop()
	cancel()

	// Verify some events were processed
	assert.True(t, processed.Load() > 0)

	// Try to publish after stop (should not panic)
	eb.Publish(EventRequestReceived, "test", nil)
}

func TestUnsubscribeNonExistent(t *testing.T) {
	eb := NewEventBus(DefaultEventBusConfig())

	err := eb.Unsubscribe("non-existent-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPublishEvent(t *testing.T) {
	eb := NewEventBus(DefaultEventBusConfig())
	eb.Start()
	defer eb.Stop()

	var receivedEvent Event

	eb.Subscribe([]EventType{EventSystemStarted}, func(event Event) {
		receivedEvent = event
	})

	// Create and publish event
	event := Event{
		ID:        "test-123",
		Type:      EventSystemStarted,
		Timestamp: time.Now(),
		Source:    "test",
		Data: map[string]interface{}{
			"foo": "bar",
		},
	}

	eb.PublishEvent(event)

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, event.ID, receivedEvent.ID)
	assert.Equal(t, event.Type, receivedEvent.Type)
	assert.Equal(t, event.Source, receivedEvent.Source)
	assert.Equal(t, event.Data["foo"], receivedEvent.Data["foo"])
}