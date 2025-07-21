package testing

import (
	"testing"
)

func TestGetFreePort(t *testing.T) {
	// Test getting a single free port
	port, err := GetFreePort()
	if err != nil {
		t.Fatalf("GetFreePort failed: %v", err)
	}
	
	if port <= 0 || port > 65535 {
		t.Errorf("Invalid port number: %d", port)
	}
	
	t.Logf("Got free port: %d", port)
}

func TestGetFreePorts(t *testing.T) {
	// Test getting multiple free ports
	count := 5
	ports, err := GetFreePorts(count)
	if err != nil {
		t.Fatalf("GetFreePorts failed: %v", err)
	}
	
	if len(ports) != count {
		t.Errorf("Expected %d ports, got %d", count, len(ports))
	}
	
	// Check that all ports are valid and unique
	seen := make(map[int]bool)
	for i, port := range ports {
		if port <= 0 || port > 65535 {
			t.Errorf("Invalid port number at index %d: %d", i, port)
		}
		
		if seen[port] {
			t.Errorf("Duplicate port found: %d", port)
		}
		seen[port] = true
	}
	
	t.Logf("Got free ports: %v", ports)
}

func TestGetFreePortsParallel(t *testing.T) {
	// Test that multiple goroutines can get free ports without conflicts
	t.Parallel()
	
	done := make(chan int, 10)
	
	for i := 0; i < 10; i++ {
		go func() {
			port, err := GetFreePort()
			if err != nil {
				t.Errorf("GetFreePort failed: %v", err)
				done <- -1
				return
			}
			done <- port
		}()
	}
	
	// Collect all ports
	ports := make([]int, 0, 10)
	for i := 0; i < 10; i++ {
		port := <-done
		if port > 0 {
			ports = append(ports, port)
		}
	}
	
	// Check for duplicates
	seen := make(map[int]int)
	for _, port := range ports {
		seen[port]++
		if seen[port] > 1 {
			t.Errorf("Port %d was assigned to multiple goroutines", port)
		}
	}
	
	t.Logf("Parallel test got ports: %v", ports)
}