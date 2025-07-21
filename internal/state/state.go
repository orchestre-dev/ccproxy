package state

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ServiceState represents the current state of the service
type ServiceState string

const (
	// StateStarting indicates the service is starting up
	StateStarting ServiceState = "starting"
	// StateReady indicates the service is ready to handle requests
	StateReady ServiceState = "ready"
	// StateDegraded indicates the service is partially functional
	StateDegraded ServiceState = "degraded"
	// StateStopping indicates the service is shutting down
	StateStopping ServiceState = "stopping"
	// StateStopped indicates the service has stopped
	StateStopped ServiceState = "stopped"
	// StateError indicates the service is in an error state
	StateError ServiceState = "error"
)

// Component represents a component that can have state
type Component struct {
	Name        string
	State       ServiceState
	Error       error
	LastChanged time.Time
}

// StateChangeHandler is called when state changes
type StateChangeHandler func(old, new ServiceState, component string)

// Manager manages service state and transitions
type Manager struct {
	mu         sync.RWMutex
	state      ServiceState
	components map[string]*Component
	handlers   []StateChangeHandler
	startTime  time.Time
	readyTime  time.Time
	errCount   int
}

// NewManager creates a new state manager
func NewManager() *Manager {
	return &Manager{
		state:      StateStarting,
		components: make(map[string]*Component),
		handlers:   make([]StateChangeHandler, 0),
		startTime:  time.Now(),
	}
}

// RegisterComponent registers a component for state tracking
func (m *Manager) RegisterComponent(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.components[name] = &Component{
		Name:        name,
		State:       StateStarting,
		LastChanged: time.Now(),
	}
}

// SetComponentState updates the state of a component
func (m *Manager) SetComponentState(name string, state ServiceState, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	component, exists := m.components[name]
	if !exists {
		m.components[name] = &Component{
			Name:        name,
			State:       state,
			Error:       err,
			LastChanged: time.Now(),
		}
		// Don't return early - we need to update service state
		oldState := StateStarting // Default old state for new components

		if err != nil {
			m.errCount++
		}

		// Update overall service state based on component states
		m.updateServiceState()

		// Notify handlers of component state change
		for _, handler := range m.handlers {
			handler(oldState, state, name)
		}
		return
	}

	oldState := component.State
	component.State = state
	component.Error = err
	component.LastChanged = time.Now()

	if err != nil {
		m.errCount++
	}

	// Update overall service state based on component states
	m.updateServiceState()

	// Notify handlers of component state change
	for _, handler := range m.handlers {
		handler(oldState, state, name)
	}
}

// updateServiceState calculates the overall service state based on components
func (m *Manager) updateServiceState() {
	var readyCount, errorCount, stoppingCount, startingCount int

	for _, component := range m.components {
		switch component.State {
		case StateReady:
			readyCount++
		case StateError:
			errorCount++
		case StateStopping, StateStopped:
			stoppingCount++
		case StateStarting:
			startingCount++
		}
	}

	oldState := m.state
	totalComponents := len(m.components)

	// Debug logging (commented out for production)
	// fmt.Printf("updateServiceState: total=%d, ready=%d, error=%d, stopping=%d, starting=%d\n",
	//     totalComponents, readyCount, errorCount, stoppingCount, startingCount)

	// Determine new state based on component states
	switch {
	case stoppingCount > 0:
		m.state = StateStopping
	case errorCount == totalComponents && totalComponents > 0:
		m.state = StateError
	case errorCount > 0 && readyCount > 0:
		m.state = StateDegraded
	case readyCount == totalComponents && totalComponents > 0:
		m.state = StateReady
		if m.readyTime.IsZero() {
			m.readyTime = time.Now()
		}
	case startingCount > 0:
		m.state = StateStarting
	case totalComponents == 0:
		m.state = StateStarting
	default:
		m.state = StateStarting
	}

	// Notify handlers of service state change
	if oldState != m.state {
		for _, handler := range m.handlers {
			handler(oldState, m.state, "service")
		}
	}
}

// GetState returns the current service state
func (m *Manager) GetState() ServiceState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

// GetComponentState returns the state of a specific component
func (m *Manager) GetComponentState(name string) (ServiceState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	component, exists := m.components[name]
	if !exists {
		return "", fmt.Errorf("component %s not found", name)
	}

	return component.State, component.Error
}

// GetComponents returns all registered components
func (m *Manager) GetComponents() map[string]Component {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]Component, len(m.components))
	for k, v := range m.components {
		result[k] = *v
	}
	return result
}

// OnStateChange registers a handler for state changes
func (m *Manager) OnStateChange(handler StateChangeHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers = append(m.handlers, handler)
}

// SetReady marks the service as ready
func (m *Manager) SetReady() {
	m.mu.Lock()
	defer m.mu.Unlock()

	oldState := m.state
	m.state = StateReady
	m.readyTime = time.Now()

	if oldState != m.state {
		for _, handler := range m.handlers {
			handler(oldState, m.state, "service")
		}
	}
}

// SetStopping marks the service as stopping
func (m *Manager) SetStopping() {
	m.mu.Lock()
	defer m.mu.Unlock()

	oldState := m.state
	m.state = StateStopping

	if oldState != m.state {
		for _, handler := range m.handlers {
			handler(oldState, m.state, "service")
		}
	}
}

// SetError marks the service as in error state
func (m *Manager) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	oldState := m.state
	m.state = StateError
	m.errCount++

	if oldState != m.state {
		for _, handler := range m.handlers {
			handler(oldState, m.state, "service")
		}
	}
}

// GetStatus returns the current status of the service
func (m *Manager) GetStatus() Status {
	m.mu.RLock()
	defer m.mu.RUnlock()

	components := make(map[string]ComponentStatus)
	for name, comp := range m.components {
		components[name] = ComponentStatus{
			Name:        comp.Name,
			State:       string(comp.State),
			Error:       "",
			LastChanged: comp.LastChanged,
		}
		if comp.Error != nil {
			components[name] = ComponentStatus{
				Name:        comp.Name,
				State:       string(comp.State),
				Error:       comp.Error.Error(),
				LastChanged: comp.LastChanged,
			}
		}
	}

	var uptime time.Duration
	if !m.readyTime.IsZero() {
		uptime = time.Since(m.readyTime)
	}

	return Status{
		State:      string(m.state),
		StartTime:  m.startTime,
		ReadyTime:  m.readyTime,
		Uptime:     uptime,
		Components: components,
		ErrorCount: m.errCount,
	}
}

// Status represents the service status
type Status struct {
	State      string                     `json:"state"`
	StartTime  time.Time                  `json:"start_time"`
	ReadyTime  time.Time                  `json:"ready_time,omitempty"`
	Uptime     time.Duration              `json:"uptime,omitempty"`
	Components map[string]ComponentStatus `json:"components"`
	ErrorCount int                        `json:"error_count"`
}

// ComponentStatus represents the status of a component
type ComponentStatus struct {
	Name        string    `json:"name"`
	State       string    `json:"state"`
	Error       string    `json:"error,omitempty"`
	LastChanged time.Time `json:"last_changed"`
}

// WaitForState waits for the service to reach a specific state
func (m *Manager) WaitForState(ctx context.Context, state ServiceState, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for state %s", state)
		case <-ticker.C:
			if m.GetState() == state {
				return nil
			}
		}
	}
}

// IsHealthy returns true if the service is in a healthy state
func (m *Manager) IsHealthy() bool {
	state := m.GetState()
	return state == StateReady || state == StateDegraded
}

// IsReady returns true if the service is ready
func (m *Manager) IsReady() bool {
	return m.GetState() == StateReady
}
