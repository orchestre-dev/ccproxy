package server

import (
	"context"
	"runtime"
	"time"
)

// Safety limits to prevent system freeze
const (
	maxGoroutines = 1000
	maxMemoryMB   = 512
)

// checkResourceLimits monitors resource usage and returns error if limits exceeded
func checkResourceLimits() error {
	// Check goroutine count
	if runtime.NumGoroutine() > maxGoroutines {
		return fmt.Errorf("goroutine limit exceeded: %d > %d", runtime.NumGoroutine(), maxGoroutines)
	}
	
	// Check memory usage
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	allocMB := m.Alloc / 1024 / 1024
	if allocMB > maxMemoryMB {
		return fmt.Errorf("memory limit exceeded: %d MB > %d MB", allocMB, maxMemoryMB)
	}
	
	return nil
}

// startResourceMonitor starts a goroutine that monitors resource usage
func (s *Server) startResourceMonitor(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := checkResourceLimits(); err != nil {
					utils.GetLogger().Errorf("Resource limit exceeded: %v", err)
					// Trigger graceful shutdown
					s.Shutdown()
				}
			}
		}
	}()
}