package process

import (
	"os"
	"sync"
	"testing"
)

func TestNewReferenceCounter(t *testing.T) {
	rc, err := NewReferenceCounter()
	if err != nil {
		t.Fatalf("NewReferenceCounter failed: %v", err)
	}
	
	if rc == nil {
		t.Fatal("NewReferenceCounter returned nil")
	}
	
	if rc.path == "" {
		t.Error("Reference counter path is empty")
	}
	
	// Cleanup
	rc.Cleanup()
}

func TestIncrementDecrement(t *testing.T) {
	rc, err := NewReferenceCounter()
	if err != nil {
		t.Fatalf("NewReferenceCounter failed: %v", err)
	}
	defer rc.Cleanup()
	
	// Initial count should be 0
	if count := rc.GetCount(); count != 0 {
		t.Errorf("Expected initial count 0, got %d", count)
	}
	
	// Increment
	if err := rc.Increment(); err != nil {
		t.Fatalf("Increment failed: %v", err)
	}
	
	if count := rc.GetCount(); count != 1 {
		t.Errorf("Expected count 1 after increment, got %d", count)
	}
	
	// Increment again
	if err := rc.Increment(); err != nil {
		t.Fatalf("Second increment failed: %v", err)
	}
	
	if count := rc.GetCount(); count != 2 {
		t.Errorf("Expected count 2 after second increment, got %d", count)
	}
	
	// Decrement
	shouldStop, err := rc.Decrement()
	if err != nil {
		t.Fatalf("Decrement failed: %v", err)
	}
	
	if shouldStop {
		t.Error("Should not stop when count is still positive")
	}
	
	if count := rc.GetCount(); count != 1 {
		t.Errorf("Expected count 1 after decrement, got %d", count)
	}
	
	// Decrement to zero
	shouldStop, err = rc.Decrement()
	if err != nil {
		t.Fatalf("Second decrement failed: %v", err)
	}
	
	if !shouldStop {
		t.Error("Should stop when count reaches zero")
	}
	
	if count := rc.GetCount(); count != 0 {
		t.Errorf("Expected count 0 after decrement to zero, got %d", count)
	}
}

func TestDecrementBelowZero(t *testing.T) {
	rc, err := NewReferenceCounter()
	if err != nil {
		t.Fatalf("NewReferenceCounter failed: %v", err)
	}
	defer rc.Cleanup()
	
	// Decrement when already at 0
	shouldStop, err := rc.Decrement()
	if err != nil {
		t.Fatalf("Decrement failed: %v", err)
	}
	
	if !shouldStop {
		t.Error("Should stop when already at zero")
	}
	
	// Count should still be 0 (not negative)
	if count := rc.GetCount(); count != 0 {
		t.Errorf("Expected count to remain 0, got %d", count)
	}
}

func TestReset(t *testing.T) {
	rc, err := NewReferenceCounter()
	if err != nil {
		t.Fatalf("NewReferenceCounter failed: %v", err)
	}
	defer rc.Cleanup()
	
	// Set count to 5
	for i := 0; i < 5; i++ {
		if err := rc.Increment(); err != nil {
			t.Fatalf("Increment failed: %v", err)
		}
	}
	
	if count := rc.GetCount(); count != 5 {
		t.Errorf("Expected count 5, got %d", count)
	}
	
	// Reset
	if err := rc.Reset(); err != nil {
		t.Fatalf("Reset failed: %v", err)
	}
	
	if count := rc.GetCount(); count != 0 {
		t.Errorf("Expected count 0 after reset, got %d", count)
	}
}

func TestCorruptedFile(t *testing.T) {
	rc, err := NewReferenceCounter()
	if err != nil {
		t.Fatalf("NewReferenceCounter failed: %v", err)
	}
	defer rc.Cleanup()
	
	// Write corrupted data
	if err := os.WriteFile(rc.path, []byte("invalid"), 0644); err != nil {
		t.Fatalf("Failed to write corrupted data: %v", err)
	}
	
	// Should handle gracefully and return 0
	if count := rc.GetCount(); count != 0 {
		t.Errorf("Expected count 0 for corrupted file, got %d", count)
	}
	
	// Should be able to increment from corrupted state
	if err := rc.Increment(); err != nil {
		t.Fatalf("Increment after corruption failed: %v", err)
	}
	
	if count := rc.GetCount(); count != 1 {
		t.Errorf("Expected count 1 after increment from corrupted state, got %d", count)
	}
}

func TestConcurrentAccess(t *testing.T) {
	rc, err := NewReferenceCounter()
	if err != nil {
		t.Fatalf("NewReferenceCounter failed: %v", err)
	}
	defer rc.Cleanup()
	
	// Run 100 increments concurrently
	var wg sync.WaitGroup
	numGoroutines := 100
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := rc.Increment(); err != nil {
				t.Errorf("Concurrent increment failed: %v", err)
			}
		}()
	}
	
	wg.Wait()
	
	// Should have exactly 100
	if count := rc.GetCount(); count != numGoroutines {
		t.Errorf("Expected count %d after concurrent increments, got %d", numGoroutines, count)
	}
	
	// Run 50 decrements concurrently
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := rc.Decrement(); err != nil {
				t.Errorf("Concurrent decrement failed: %v", err)
			}
		}()
	}
	
	wg.Wait()
	
	// Should have exactly 50
	if count := rc.GetCount(); count != 50 {
		t.Errorf("Expected count 50 after concurrent decrements, got %d", count)
	}
}

func TestIncrementAndCheck(t *testing.T) {
	rc, err := NewReferenceCounter()
	if err != nil {
		t.Fatalf("NewReferenceCounter failed: %v", err)
	}
	defer rc.Cleanup()
	
	// Test IncrementAndCheck
	newCount, err := rc.IncrementAndCheck()
	if err != nil {
		t.Fatalf("IncrementAndCheck failed: %v", err)
	}
	
	if newCount != 1 {
		t.Errorf("Expected new count 1, got %d", newCount)
	}
	
	newCount, err = rc.IncrementAndCheck()
	if err != nil {
		t.Fatalf("Second IncrementAndCheck failed: %v", err)
	}
	
	if newCount != 2 {
		t.Errorf("Expected new count 2, got %d", newCount)
	}
}

func TestDecrementAndCheck(t *testing.T) {
	rc, err := NewReferenceCounter()
	if err != nil {
		t.Fatalf("NewReferenceCounter failed: %v", err)
	}
	defer rc.Cleanup()
	
	// Set count to 2
	rc.Increment()
	rc.Increment()
	
	// Test DecrementAndCheck
	shouldStop, newCount, err := rc.DecrementAndCheck()
	if err != nil {
		t.Fatalf("DecrementAndCheck failed: %v", err)
	}
	
	if shouldStop {
		t.Error("Should not stop when count is still positive")
	}
	
	if newCount != 1 {
		t.Errorf("Expected new count 1, got %d", newCount)
	}
	
	// Decrement to zero
	shouldStop, newCount, err = rc.DecrementAndCheck()
	if err != nil {
		t.Fatalf("Second DecrementAndCheck failed: %v", err)
	}
	
	if !shouldStop {
		t.Error("Should stop when count reaches zero")
	}
	
	if newCount != 0 {
		t.Errorf("Expected new count 0, got %d", newCount)
	}
}