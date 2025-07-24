package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"
)

// mockSchedulingService is a mock implementation for testing
type mockSchedulingService struct {
	callCount int
	shouldErr bool
}

func (m *mockSchedulingService) CreateInstancesForActiveTemplates(ctx context.Context) error {
	m.callCount++
	if m.shouldErr {
		return context.DeadlineExceeded
	}
	return nil
}

func TestRunner_Start(t *testing.T) {
	mockService := &mockSchedulingService{}
	runner := NewRunner(mockService, 100*time.Millisecond) // Short interval for testing

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// Start the runner
	runner.Start(ctx, &wg)

	// Wait a bit to let it run a few times
	time.Sleep(250 * time.Millisecond)

	// Cancel and wait for completion
	cancel()
	wg.Wait()

	// Should have been called at least twice (initial + at least one scheduled)
	if mockService.callCount < 2 {
		t.Errorf("Expected at least 2 calls, got %d", mockService.callCount)
	}
}

func TestNewDailyRunner(t *testing.T) {
	mockService := &mockSchedulingService{}
	runner := NewDailyRunner(mockService)

	if runner.interval != 24*time.Hour {
		t.Errorf("Expected 24 hour interval, got %v", runner.interval)
	}

	if runner.schedulingService != mockService {
		t.Error("Expected scheduling service to be set correctly")
	}
}