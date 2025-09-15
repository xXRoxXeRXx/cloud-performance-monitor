package utils

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryWithSuccess(t *testing.T) {
	config := DefaultRetryConfig()
	attempts := 0
	
	err := config.WithRetry(context.Background(), "test_operation", func(ctx context.Context) error {
		attempts++
		if attempts < 2 {
			return errors.New("temporary failure")
		}
		return nil
	})
	
	if err != nil {
		t.Errorf("Expected success after retry, got error: %v", err)
	}
	
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestRetryWithNonRetryableError(t *testing.T) {
	config := DefaultRetryConfig()
	attempts := 0
	
	err := config.WithRetry(context.Background(), "test_operation", func(ctx context.Context) error {
		attempts++
		return errors.New("non retryable error")
	})
	
	if err == nil {
		t.Error("Expected error for non-retryable error")
	}
	
	if attempts != 1 {
		t.Errorf("Expected 1 attempt for non-retryable error, got %d", attempts)
	}
}

func TestRetryWithMaxRetriesExceeded(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:      2,
		InitialDelay:    1 * time.Millisecond,
		MaxDelay:        10 * time.Millisecond,
		BackoffFactor:   2.0,
		RetryableErrors: []string{"timeout"},
	}
	
	attempts := 0
	err := config.WithRetry(context.Background(), "test_operation", func(ctx context.Context) error {
		attempts++
		return errors.New("timeout error")
	})
	
	if err == nil {
		t.Error("Expected error after max retries exceeded")
	}
	
	if attempts != 3 { // initial + 2 retries
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestCircuitBreakerClosed(t *testing.T) {
	cb := NewCircuitBreaker("test", DefaultCircuitBreakerConfig())
	
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})
	
	if err != nil {
		t.Errorf("Expected success in closed state, got error: %v", err)
	}
	
	if cb.GetState() != CircuitClosed {
		t.Errorf("Expected circuit to remain closed, got state: %v", cb.GetState())
	}
}

func TestCircuitBreakerOpening(t *testing.T) {
	config := &CircuitBreakerConfig{
		MaxFailures:     2,
		ResetTimeout:    100 * time.Millisecond,
		SuccessThreshold: 1,
	}
	
	cb := NewCircuitBreaker("test", config)
	
	// First failure
	_ = cb.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("failure 1")
	})
	
	if cb.GetState() != CircuitClosed {
		t.Errorf("Expected circuit to remain closed after 1 failure, got: %v", cb.GetState())
	}
	
	// Second failure - should open circuit
	_ = cb.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("failure 2")
	})
	
	if cb.GetState() != CircuitOpen {
		t.Errorf("Expected circuit to open after 2 failures, got: %v", cb.GetState())
	}
	
	// Next execution should fail fast
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})
	
	if err == nil {
		t.Error("Expected error when circuit is open")
	}
}
