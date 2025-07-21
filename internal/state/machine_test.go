package state

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStateMachine(t *testing.T) {
	sm := NewStateMachine(StateStarting)

	assert.NotNil(t, sm)
	assert.Equal(t, StateStarting, sm.GetCurrent())
	assert.Empty(t, sm.GetHistory())
}

func TestStateMachine_AllowedTransitions(t *testing.T) {
	tests := []struct {
		name     string
		from     ServiceState
		allowed  []ServiceState
	}{
		{
			name: "from starting",
			from: StateStarting,
			allowed: []ServiceState{StateReady, StateError, StateStopping},
		},
		{
			name: "from ready",
			from: StateReady,
			allowed: []ServiceState{StateDegraded, StateError, StateStopping},
		},
		{
			name: "from degraded",
			from: StateDegraded,
			allowed: []ServiceState{StateReady, StateError, StateStopping},
		},
		{
			name: "from error",
			from: StateError,
			allowed: []ServiceState{StateStarting, StateStopping},
		},
		{
			name: "from stopping",
			from: StateStopping,
			allowed: []ServiceState{StateStopped, StateError},
		},
		{
			name: "from stopped",
			from: StateStopped,
			allowed: []ServiceState{StateStarting},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewStateMachine(tt.from)
			allowed := sm.GetAllowedTransitions()

			assert.ElementsMatch(t, tt.allowed, allowed)

			// Test CanTransitionTo
			for _, state := range tt.allowed {
				assert.True(t, sm.CanTransitionTo(state), 
					"should be able to transition from %s to %s", tt.from, state)
			}
		})
	}
}

func TestStateMachine_Transition(t *testing.T) {
	sm := NewStateMachine(StateStarting)

	// Valid transition
	err := sm.Transition(StateReady, "initialization complete")
	assert.NoError(t, err)
	assert.Equal(t, StateReady, sm.GetCurrent())

	// Check history
	history := sm.GetHistory()
	require.Len(t, history, 1)
	assert.Equal(t, StateStarting, history[0].From)
	assert.Equal(t, StateReady, history[0].To)
	assert.Equal(t, "initialization complete", history[0].Name)

	// Invalid transition
	err = sm.Transition(StateStarting, "invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not allowed")
	assert.Equal(t, StateReady, sm.GetCurrent()) // State unchanged

	// Same state transition (always allowed)
	err = sm.Transition(StateReady, "refresh")
	assert.NoError(t, err)
	assert.Equal(t, StateReady, sm.GetCurrent())
}

func TestStateMachine_TransitionHandler(t *testing.T) {
	sm := NewStateMachine(StateStarting)

	handlerCalled := false
	var handlerFrom, handlerTo ServiceState

	// Set transition handler
	sm.SetTransitionHandler(StateStarting, StateReady, func(from, to ServiceState) error {
		handlerCalled = true
		handlerFrom = from
		handlerTo = to
		return nil
	})

	// Perform transition
	err := sm.Transition(StateReady, "test")
	assert.NoError(t, err)
	assert.True(t, handlerCalled)
	assert.Equal(t, StateStarting, handlerFrom)
	assert.Equal(t, StateReady, handlerTo)
}

func TestStateMachine_TransitionHandlerError(t *testing.T) {
	sm := NewStateMachine(StateStarting)

	// Set failing handler
	sm.SetTransitionHandler(StateStarting, StateReady, func(from, to ServiceState) error {
		return errors.New("handler error")
	})

	// Attempt transition
	err := sm.Transition(StateReady, "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handler failed")
	assert.Equal(t, StateStarting, sm.GetCurrent()) // State unchanged
}

func TestStateMachine_HistoryLimit(t *testing.T) {
	sm := NewStateMachine(StateStarting)
	sm.maxHistory = 5

	// Create more transitions than history limit
	transitions := []struct {
		to   ServiceState
		name string
	}{
		{StateReady, "1"},
		{StateDegraded, "2"},
		{StateReady, "3"},
		{StateDegraded, "4"},
		{StateReady, "5"},
		{StateDegraded, "6"},
		{StateReady, "7"},
	}

	for _, tr := range transitions {
		err := sm.Transition(tr.to, tr.name)
		require.NoError(t, err)
	}

	// Check history is limited
	history := sm.GetHistory()
	assert.Len(t, history, 5)

	// Verify we have the most recent transitions
	assert.Equal(t, "3", history[0].Name)
	assert.Equal(t, "7", history[4].Name)
}

func TestStateMachine_Reset(t *testing.T) {
	sm := NewStateMachine(StateStarting)

	// Make some transitions
	sm.Transition(StateReady, "init")
	sm.Transition(StateDegraded, "degraded")

	assert.Equal(t, StateDegraded, sm.GetCurrent())
	assert.Len(t, sm.GetHistory(), 2)

	// Reset
	sm.Reset(StateStarting)

	assert.Equal(t, StateStarting, sm.GetCurrent())
	assert.Empty(t, sm.GetHistory())
}

func TestStateMachine_ConcurrentOperations(t *testing.T) {
	sm := NewStateMachine(StateStarting)

	// Set up valid transition path
	sm.Transition(StateReady, "init")

	var wg sync.WaitGroup

	// Concurrent transitions
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				if idx%2 == 0 {
					sm.Transition(StateDegraded, "concurrent")
				} else {
					sm.Transition(StateReady, "concurrent")
				}
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				_ = sm.GetCurrent()
				_ = sm.GetHistory()
				_ = sm.GetAllowedTransitions()
			}
		}()
	}

	wg.Wait()

	// Verify final state is valid
	current := sm.GetCurrent()
	assert.True(t, current == StateReady || current == StateDegraded)
}

func TestStateMachine_CompleteLifecycle(t *testing.T) {
	sm := NewStateMachine(StateStarting)

	// Define lifecycle handlers
	var events []string
	var mu sync.Mutex

	recordEvent := func(event string) TransitionFunc {
		return func(from, to ServiceState) error {
			mu.Lock()
			defer mu.Unlock()
			events = append(events, event)
			return nil
		}
	}

	// Set up handlers for key transitions
	sm.SetTransitionHandler(StateStarting, StateReady, recordEvent("service_started"))
	sm.SetTransitionHandler(StateReady, StateDegraded, recordEvent("service_degraded"))
	sm.SetTransitionHandler(StateDegraded, StateReady, recordEvent("service_recovered"))
	sm.SetTransitionHandler(StateReady, StateStopping, recordEvent("shutdown_initiated"))
	sm.SetTransitionHandler(StateStopping, StateStopped, recordEvent("shutdown_complete"))

	// Run through lifecycle
	assert.NoError(t, sm.Transition(StateReady, "startup complete"))
	assert.NoError(t, sm.Transition(StateDegraded, "component failure"))
	assert.NoError(t, sm.Transition(StateReady, "component recovered"))
	assert.NoError(t, sm.Transition(StateStopping, "shutdown requested"))
	assert.NoError(t, sm.Transition(StateStopped, "shutdown finished"))

	// Verify events
	mu.Lock()
	defer mu.Unlock()

	expectedEvents := []string{
		"service_started",
		"service_degraded", 
		"service_recovered",
		"shutdown_initiated",
		"shutdown_complete",
	}
	assert.Equal(t, expectedEvents, events)

	// Verify history
	history := sm.GetHistory()
	assert.Len(t, history, 5)
	assert.Equal(t, StateStopped, sm.GetCurrent())

	// Can restart from stopped
	assert.True(t, sm.CanTransitionTo(StateStarting))
}