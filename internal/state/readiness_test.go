package state

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReadinessProbe(t *testing.T) {
	manager := NewManager()
	probe := NewReadinessProbe(manager, 100*time.Millisecond, 50*time.Millisecond)

	assert.NotNil(t, probe)
	assert.Equal(t, manager, probe.manager)
	assert.Equal(t, 100*time.Millisecond, probe.interval)
	assert.Equal(t, 50*time.Millisecond, probe.timeout)
}

func TestRegisterCheck(t *testing.T) {
	manager := NewManager()
	probe := NewReadinessProbe(manager, 100*time.Millisecond, 50*time.Millisecond)

	checkCalled := false
	check := func(ctx context.Context) error {
		checkCalled = true
		return nil
	}

	probe.RegisterCheck("test", check)

	// Component should be registered with manager
	components := manager.GetComponents()
	assert.Contains(t, components, "test")

	// Run check once
	probe.RunOnce(context.Background())
	assert.True(t, checkCalled)
}

func TestRunCheck(t *testing.T) {
	manager := NewManager()
	probe := NewReadinessProbe(manager, 100*time.Millisecond, 50*time.Millisecond)

	// Successful check
	probe.RegisterCheck("success", func(ctx context.Context) error {
		return nil
	})

	// Failing check
	probe.RegisterCheck("failure", func(ctx context.Context) error {
		return errors.New("check failed")
	})

	// Run checks
	probe.RunOnce(context.Background())

	// Verify results
	successResult, ok := probe.GetResult("success")
	require.True(t, ok)
	assert.True(t, successResult.Success)
	assert.Nil(t, successResult.Error)
	assert.NotZero(t, successResult.LastChecked)
	assert.NotZero(t, successResult.Duration)

	failureResult, ok := probe.GetResult("failure")
	require.True(t, ok)
	assert.False(t, failureResult.Success)
	assert.Error(t, failureResult.Error)

	// Verify component states
	state, _ := manager.GetComponentState("success")
	assert.Equal(t, StateReady, state)

	state, err := manager.GetComponentState("failure")
	assert.Equal(t, StateError, state)
	assert.Error(t, err)
}

func TestIsReady(t *testing.T) {
	probe := NewReadinessProbe(nil, 100*time.Millisecond, 50*time.Millisecond)

	// No checks registered
	assert.False(t, probe.IsReady())

	// All checks passing
	probe.RegisterCheck("check1", func(ctx context.Context) error { return nil })
	probe.RegisterCheck("check2", func(ctx context.Context) error { return nil })
	probe.RunOnce(context.Background())
	assert.True(t, probe.IsReady())

	// One check failing
	probe.RegisterCheck("check3", func(ctx context.Context) error { return errors.New("fail") })
	probe.RunOnce(context.Background())
	assert.False(t, probe.IsReady())
}

func TestWaitForReady(t *testing.T) {
	probe := NewReadinessProbe(nil, 100*time.Millisecond, 50*time.Millisecond)

	var ready atomic.Bool
	probe.RegisterCheck("async", func(ctx context.Context) error {
		if ready.Load() {
			return nil
		}
		return errors.New("not ready")
	})

	// Start probe
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	probe.Start(ctx)
	defer probe.Stop()

	// Set ready after a short delay
	go func() {
		time.Sleep(150 * time.Millisecond)
		ready.Store(true)
	}()

	// Wait for ready
	err := probe.WaitForReady(context.Background(), 500*time.Millisecond)
	assert.NoError(t, err)
	assert.True(t, probe.IsReady())
}

func TestWaitForReadyTimeout(t *testing.T) {
	probe := NewReadinessProbe(nil, 100*time.Millisecond, 50*time.Millisecond)

	// Always failing check
	probe.RegisterCheck("failing", func(ctx context.Context) error {
		return errors.New("always fails")
	})

	// Start probe
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	probe.Start(ctx)
	defer probe.Stop()

	// Wait should timeout
	err := probe.WaitForReady(context.Background(), 100*time.Millisecond)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func TestCheckTimeout(t *testing.T) {
	probe := NewReadinessProbe(nil, 100*time.Millisecond, 20*time.Millisecond)

	// Check that exceeds timeout
	probe.RegisterCheck("slow", func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return nil
		}
	})

	start := time.Now()
	probe.RunOnce(context.Background())
	duration := time.Since(start)

	// Check should have timed out
	result, ok := probe.GetResult("slow")
	require.True(t, ok)
	assert.False(t, result.Success)
	assert.Error(t, result.Error)
	assert.True(t, duration < 50*time.Millisecond, "check should timeout quickly")
}

func TestGetAllResults(t *testing.T) {
	probe := NewReadinessProbe(nil, 100*time.Millisecond, 50*time.Millisecond)

	probe.RegisterCheck("check1", func(ctx context.Context) error { return nil })
	probe.RegisterCheck("check2", func(ctx context.Context) error { return errors.New("fail") })
	probe.RegisterCheck("check3", func(ctx context.Context) error { return nil })

	probe.RunOnce(context.Background())

	results := probe.GetAllResults()
	assert.Len(t, results, 3)
	assert.True(t, results["check1"].Success)
	assert.False(t, results["check2"].Success)
	assert.True(t, results["check3"].Success)
}

func TestStartStop(t *testing.T) {
	probe := NewReadinessProbe(nil, 50*time.Millisecond, 20*time.Millisecond)

	var callCount atomic.Int32
	probe.RegisterCheck("counter", func(ctx context.Context) error {
		callCount.Add(1)
		return nil
	})

	// Start probe
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	probe.Start(ctx)

	// Initial check happens immediately, then wait for at least 2 more intervals
	time.Sleep(150 * time.Millisecond)

	// Stop probe
	probe.Stop()

	finalCount := callCount.Load()
	assert.GreaterOrEqual(t, finalCount, int32(2), "check should run at least twice")

	// Count should not increase after stop
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, finalCount, callCount.Load())
}

func TestDoubleStart(t *testing.T) {
	probe := NewReadinessProbe(nil, 100*time.Millisecond, 50*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// First start
	probe.Start(ctx)
	defer probe.Stop()

	// Second start should be ignored
	probe.Start(ctx) // Should not panic or cause issues
}

func TestGetResultNotFound(t *testing.T) {
	probe := NewReadinessProbe(nil, 100*time.Millisecond, 50*time.Millisecond)

	result, ok := probe.GetResult("nonexistent")
	assert.Nil(t, result)
	assert.False(t, ok)
}

func TestConcurrentChecks(t *testing.T) {
	probe := NewReadinessProbe(nil, 100*time.Millisecond, 50*time.Millisecond)

	// Register multiple checks
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("check%d", i)
		idx := i
		probe.RegisterCheck(name, func(ctx context.Context) error {
			time.Sleep(time.Duration(idx) * time.Millisecond)
			if idx%2 == 0 {
				return nil
			}
			return errors.New("odd check fails")
		})
	}

	// Run checks concurrently
	probe.RunOnce(context.Background())

	// Verify all checks completed
	results := probe.GetAllResults()
	assert.Len(t, results, 10)

	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("check%d", i)
		if i%2 == 0 {
			assert.True(t, results[name].Success)
		} else {
			assert.False(t, results[name].Success)
		}
	}
}

func TestProbeWithStateManager(t *testing.T) {
	manager := NewManager()
	probe := NewReadinessProbe(manager, 50*time.Millisecond, 20*time.Millisecond)

	// Track state changes
	var stateChanges []string
	manager.OnStateChange(func(old, new ServiceState, component string) {
		stateChanges = append(stateChanges, fmt.Sprintf("%s:%s->%s", component, old, new))
	})

	// Register checks
	probe.RegisterCheck("database", func(ctx context.Context) error { return nil })
	probe.RegisterCheck("cache", func(ctx context.Context) error { return nil })

	// Start probe
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	probe.Start(ctx)
	defer probe.Stop()

	// Wait for checks to complete
	err := probe.WaitForReady(context.Background(), 200*time.Millisecond)
	require.NoError(t, err)

	// Verify state manager shows ready
	assert.Equal(t, StateReady, manager.GetState())
	assert.True(t, manager.IsReady())

	// Verify state changes were tracked
	assert.NotEmpty(t, stateChanges)
}