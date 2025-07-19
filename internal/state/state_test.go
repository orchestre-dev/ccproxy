package state

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	m := NewManager()

	assert.NotNil(t, m)
	assert.Equal(t, StateStarting, m.GetState())
	assert.Empty(t, m.GetComponents())
	assert.NotZero(t, m.startTime)
}

func TestRegisterComponent(t *testing.T) {
	m := NewManager()

	m.RegisterComponent("server")
	m.RegisterComponent("database")

	components := m.GetComponents()
	assert.Len(t, components, 2)
	assert.Equal(t, StateStarting, components["server"].State)
	assert.Equal(t, StateStarting, components["database"].State)
}

func TestSetComponentState(t *testing.T) {
	m := NewManager()

	// Set state for unregistered component
	m.SetComponentState("server", StateReady, nil)

	state, err := m.GetComponentState("server")
	assert.Equal(t, StateReady, state)
	assert.Nil(t, err)

	// Set error state
	testErr := assert.AnError
	m.SetComponentState("database", StateError, testErr)

	state, err = m.GetComponentState("database")
	assert.Equal(t, StateError, state)
	assert.Equal(t, testErr, err)
}

func TestServiceStateTransitions(t *testing.T) {
	tests := []struct {
		name           string
		componentStates map[string]ServiceState
		expectedState   ServiceState
	}{
		{
			name:           "all ready",
			componentStates: map[string]ServiceState{
				"server":   StateReady,
				"database": StateReady,
			},
			expectedState: StateReady,
		},
		{
			name:           "all error",
			componentStates: map[string]ServiceState{
				"server":   StateError,
				"database": StateError,
			},
			expectedState: StateError,
		},
		{
			name:           "mixed ready and error",
			componentStates: map[string]ServiceState{
				"server":   StateReady,
				"database": StateError,
			},
			expectedState: StateDegraded,
		},
		{
			name:           "one stopping",
			componentStates: map[string]ServiceState{
				"server":   StateReady,
				"database": StateStopping,
			},
			expectedState: StateStopping,
		},
		{
			name:           "still starting",
			componentStates: map[string]ServiceState{
				"server":   StateReady,
				"database": StateStarting,
			},
			expectedState: StateStarting,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager()

			for name, state := range tt.componentStates {
				m.SetComponentState(name, state, nil)
			}

			// Debug: Print actual state vs expected
			actualState := m.GetState()
			components := m.GetComponents()
			t.Logf("Test: %s, Components: %+v, Actual state: %s, Expected: %s", 
				tt.name, components, actualState, tt.expectedState)

			assert.Equal(t, tt.expectedState, actualState)
		})
	}
}

func TestStateChangeHandlers(t *testing.T) {
	m := NewManager()

	var (
		mu          sync.Mutex
		changes     []string
		handlerFunc = func(old, new ServiceState, component string) {
			mu.Lock()
			defer mu.Unlock()
			changes = append(changes, fmt.Sprintf("%s: %s -> %s", component, old, new))
		}
	)

	m.OnStateChange(handlerFunc)

	// Register and update components
	m.RegisterComponent("server")
	m.SetComponentState("server", StateReady, nil)

	// Wait for handlers to execute
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	// Should have component change and service change
	assert.Contains(t, changes, "server: starting -> ready")
	assert.Contains(t, changes, "service: starting -> ready")
}

func TestSetReady(t *testing.T) {
	m := NewManager()

	assert.Equal(t, StateStarting, m.GetState())
	assert.Zero(t, m.readyTime)

	m.SetReady()

	assert.Equal(t, StateReady, m.GetState())
	assert.NotZero(t, m.readyTime)
}

func TestSetStopping(t *testing.T) {
	m := NewManager()
	m.SetReady()

	assert.Equal(t, StateReady, m.GetState())

	m.SetStopping()

	assert.Equal(t, StateStopping, m.GetState())
}

func TestSetError(t *testing.T) {
	m := NewManager()
	m.SetReady()

	assert.Equal(t, StateReady, m.GetState())
	assert.Equal(t, 0, m.errCount)

	m.SetError(assert.AnError)

	assert.Equal(t, StateError, m.GetState())
	assert.Equal(t, 1, m.errCount)
}

func TestGetStatus(t *testing.T) {
	m := NewManager()

	// Register components
	m.RegisterComponent("server")
	m.RegisterComponent("database")

	// Set states
	m.SetComponentState("server", StateReady, nil)
	m.SetComponentState("database", StateError, assert.AnError)

	// Get status
	status := m.GetStatus()

	assert.Equal(t, string(StateDegraded), status.State)
	assert.NotZero(t, status.StartTime)
	assert.Len(t, status.Components, 2)
	assert.Equal(t, "ready", status.Components["server"].State)
	assert.Equal(t, "error", status.Components["database"].State)
	assert.NotEmpty(t, status.Components["database"].Error)
	assert.Equal(t, 1, status.ErrorCount)
}

func TestWaitForState(t *testing.T) {
	m := NewManager()

	// Test successful wait
	go func() {
		time.Sleep(50 * time.Millisecond)
		m.SetReady()
	}()

	ctx := context.Background()
	err := m.WaitForState(ctx, StateReady, 200*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, StateReady, m.GetState())

	// Test timeout
	err = m.WaitForState(ctx, StateStopped, 50*time.Millisecond)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func TestIsHealthy(t *testing.T) {
	m := NewManager()

	tests := []struct {
		state    ServiceState
		healthy  bool
	}{
		{StateStarting, false},
		{StateReady, true},
		{StateDegraded, true},
		{StateStopping, false},
		{StateStopped, false},
		{StateError, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			m.mu.Lock()
			m.state = tt.state
			m.mu.Unlock()

			assert.Equal(t, tt.healthy, m.IsHealthy())
		})
	}
}

func TestManagerIsReady(t *testing.T) {
	m := NewManager()

	assert.False(t, m.IsReady())

	m.SetReady()
	assert.True(t, m.IsReady())

	m.SetStopping()
	assert.False(t, m.IsReady())
}

func TestGetComponentStateNotFound(t *testing.T) {
	m := NewManager()

	state, err := m.GetComponentState("nonexistent")
	assert.Equal(t, ServiceState(""), state)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestConcurrentOperations(t *testing.T) {
	m := NewManager()

	// Register components
	for i := 0; i < 10; i++ {
		m.RegisterComponent(fmt.Sprintf("component-%d", i))
	}

	// Concurrent updates
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				state := StateReady
				if j%2 == 0 {
					state = StateError
				}
				m.SetComponentState(fmt.Sprintf("component-%d", idx), state, nil)
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = m.GetState()
				_ = m.GetComponents()
				_ = m.GetStatus()
			}
		}()
	}

	wg.Wait()

	// Verify final state
	components := m.GetComponents()
	assert.Len(t, components, 10)
}

func TestUptimeCalculation(t *testing.T) {
	m := NewManager()

	// Initially no uptime
	status := m.GetStatus()
	assert.Zero(t, status.Uptime)

	// Set ready and check uptime
	m.SetReady()
	time.Sleep(100 * time.Millisecond)

	status = m.GetStatus()
	assert.True(t, status.Uptime >= 100*time.Millisecond)
	assert.True(t, status.Uptime < 200*time.Millisecond)
}

func TestMultipleStateChangeHandlers(t *testing.T) {
	m := NewManager()

	var (
		mu       sync.Mutex
		counter1 int
		counter2 int
	)

	m.OnStateChange(func(old, new ServiceState, component string) {
		mu.Lock()
		defer mu.Unlock()
		counter1++
	})

	m.OnStateChange(func(old, new ServiceState, component string) {
		mu.Lock()
		defer mu.Unlock()
		counter2++
	})

	m.SetReady()
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	assert.Equal(t, 1, counter1)
	assert.Equal(t, 1, counter2)
}

func TestStateManager_FullLifecycle(t *testing.T) {
	m := NewManager()

	// Track state changes
	var stateChanges []string
	m.OnStateChange(func(old, new ServiceState, component string) {
		stateChanges = append(stateChanges, fmt.Sprintf("%s: %s->%s", component, old, new))
	})

	// Starting state
	assert.Equal(t, StateStarting, m.GetState())

	// Register components
	m.RegisterComponent("api")
	m.RegisterComponent("database")
	m.RegisterComponent("cache")

	// Components come online
	m.SetComponentState("database", StateReady, nil)
	m.SetComponentState("cache", StateReady, nil)
	m.SetComponentState("api", StateReady, nil)

	// Service should be ready
	assert.Equal(t, StateReady, m.GetState())
	assert.True(t, m.IsReady())
	assert.True(t, m.IsHealthy())

	// Database has issues
	m.SetComponentState("database", StateError, assert.AnError)
	assert.Equal(t, StateDegraded, m.GetState())
	assert.False(t, m.IsReady())
	assert.True(t, m.IsHealthy()) // Still healthy in degraded state

	// Database recovers
	m.SetComponentState("database", StateReady, nil)
	assert.Equal(t, StateReady, m.GetState())

	// Shutdown sequence
	m.SetStopping()
	assert.Equal(t, StateStopping, m.GetState())
	assert.False(t, m.IsHealthy())

	// Verify state changes were tracked
	assert.NotEmpty(t, stateChanges)
}