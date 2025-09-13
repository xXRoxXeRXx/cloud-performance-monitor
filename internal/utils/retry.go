package utils

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"
)

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxRetries      int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	RetryableErrors []string
}

// DefaultRetryConfig returns a sensible default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:    3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		RetryableErrors: []string{
			"connection refused",
			"timeout",
			"temporary failure",
			"network is unreachable",
			"no such host",
		},
	}
}

// RetryableFunc is a function that can be retried
type RetryableFunc func(ctx context.Context) error

// IsRetryableError checks if an error is retryable
func (rc *RetryConfig) IsRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	for _, retryable := range rc.RetryableErrors {
		if contains(errStr, retryable) {
			return true
		}
	}
	return false
}

// WithRetry executes a function with retry logic
func (rc *RetryConfig) WithRetry(ctx context.Context, operation string, fn RetryableFunc) error {
	var lastErr error
	
	for attempt := 0; attempt <= rc.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := rc.calculateDelay(attempt)
			log.Printf("Retrying %s (attempt %d/%d) after %v delay", 
				operation, attempt, rc.MaxRetries, delay)
			
			select {
			case <-ctx.Done():
				return fmt.Errorf("operation cancelled during retry: %w", ctx.Err())
			case <-time.After(delay):
			}
		}
		
		lastErr = fn(ctx)
		if lastErr == nil {
			if attempt > 0 {
				log.Printf("Operation %s succeeded after %d retries", operation, attempt)
			}
			return nil
		}
		
		// Check if error is retryable
		if !rc.IsRetryableError(lastErr) {
			log.Printf("Non-retryable error in %s: %v", operation, lastErr)
			return lastErr
		}
		
		log.Printf("Retryable error in %s (attempt %d): %v", operation, attempt+1, lastErr)
	}
	
	return fmt.Errorf("operation %s failed after %d retries: %w", operation, rc.MaxRetries, lastErr)
}

// calculateDelay calculates the delay for exponential backoff
func (rc *RetryConfig) calculateDelay(attempt int) time.Duration {
	delay := float64(rc.InitialDelay) * math.Pow(rc.BackoffFactor, float64(attempt-1))
	if delay > float64(rc.MaxDelay) {
		delay = float64(rc.MaxDelay)
	}
	return time.Duration(delay)
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    (len(s) > len(substr) && 
		     (s[:len(substr)] == substr || 
		      s[len(s)-len(substr):] == substr ||
		      hasSubstring(s, substr))))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
