package state

import (
	"fmt"
	"sync"
)

// Transition represents a state transition
type Transition struct {
	From ServiceState
	To   ServiceState
	Name string
}

// TransitionFunc is called during a state transition
type TransitionFunc func(from, to ServiceState) error

// StateMachine manages state transitions with validation
type StateMachine struct {
	mu          sync.RWMutex
	current     ServiceState
	transitions map[string][]ServiceState // from -> allowed to states
	handlers    map[string]TransitionFunc // "from->to" -> handler
	history     []Transition
	maxHistory  int
}

// NewStateMachine creates a new state machine
func NewStateMachine(initial ServiceState) *StateMachine {
	sm := &StateMachine{
		current:     initial,
		transitions: make(map[string][]ServiceState),
		handlers:    make(map[string]TransitionFunc),
		history:     make([]Transition, 0),
		maxHistory:  100,
	}

	// Define allowed transitions
	sm.defineTransitions()

	return sm
}

// defineTransitions sets up the allowed state transitions
func (sm *StateMachine) defineTransitions() {
	// From Starting
	sm.transitions[string(StateStarting)] = []ServiceState{
		StateReady,
		StateError,
		StateStopping,
	}

	// From Ready
	sm.transitions[string(StateReady)] = []ServiceState{
		StateDegraded,
		StateError,
		StateStopping,
	}

	// From Degraded
	sm.transitions[string(StateDegraded)] = []ServiceState{
		StateReady,
		StateError,
		StateStopping,
	}

	// From Error
	sm.transitions[string(StateError)] = []ServiceState{
		StateStarting, // Allow restart
		StateStopping,
	}

	// From Stopping
	sm.transitions[string(StateStopping)] = []ServiceState{
		StateStopped,
		StateError, // Error during shutdown
	}

	// From Stopped - terminal state, but allow restart
	sm.transitions[string(StateStopped)] = []ServiceState{
		StateStarting,
	}
}

// SetTransitionHandler registers a handler for a specific transition
func (sm *StateMachine) SetTransitionHandler(from, to ServiceState, handler TransitionFunc) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	key := fmt.Sprintf("%s->%s", from, to)
	sm.handlers[key] = handler
}

// Transition attempts to transition to a new state
func (sm *StateMachine) Transition(to ServiceState, name string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	from := sm.current

	// Check if transition is allowed
	if !sm.isTransitionAllowed(from, to) {
		return fmt.Errorf("transition from %s to %s not allowed", from, to)
	}

	// Execute transition handler if registered
	key := fmt.Sprintf("%s->%s", from, to)
	if handler, exists := sm.handlers[key]; exists {
		if err := handler(from, to); err != nil {
			return fmt.Errorf("transition handler failed: %w", err)
		}
	}

	// Update state
	sm.current = to

	// Record transition
	sm.recordTransition(Transition{
		From: from,
		To:   to,
		Name: name,
	})

	return nil
}

// isTransitionAllowed checks if a transition is valid
func (sm *StateMachine) isTransitionAllowed(from, to ServiceState) bool {
	// Same state transition is always allowed
	if from == to {
		return true
	}

	allowed, exists := sm.transitions[string(from)]
	if !exists {
		return false
	}

	for _, state := range allowed {
		if state == to {
			return true
		}
	}

	return false
}

// recordTransition adds a transition to history
func (sm *StateMachine) recordTransition(t Transition) {
	sm.history = append(sm.history, t)

	// Trim history if needed
	if len(sm.history) > sm.maxHistory {
		sm.history = sm.history[len(sm.history)-sm.maxHistory:]
	}
}

// GetCurrent returns the current state
func (sm *StateMachine) GetCurrent() ServiceState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.current
}

// GetHistory returns the transition history
func (sm *StateMachine) GetHistory() []Transition {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make([]Transition, len(sm.history))
	copy(result, sm.history)
	return result
}

// CanTransitionTo checks if a transition to the given state is allowed
func (sm *StateMachine) CanTransitionTo(to ServiceState) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.isTransitionAllowed(sm.current, to)
}

// GetAllowedTransitions returns all allowed transitions from current state
func (sm *StateMachine) GetAllowedTransitions() []ServiceState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	allowed, exists := sm.transitions[string(sm.current)]
	if !exists {
		return []ServiceState{}
	}

	result := make([]ServiceState, len(allowed))
	copy(result, allowed)
	return result
}

// Reset resets the state machine to initial state
func (sm *StateMachine) Reset(initial ServiceState) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.current = initial
	sm.history = make([]Transition, 0)
}