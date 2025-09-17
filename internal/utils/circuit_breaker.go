package utils

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	// CircuitClosed - normal operation
	CircuitClosed CircuitState = iota
	// CircuitOpen - circuit is open, requests fail fast
	CircuitOpen
	// CircuitHalfOpen - testing if service recovered
	CircuitHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "CLOSED"
	case CircuitOpen:
		return "OPEN"
	case CircuitHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreakerConfig configures circuit breaker behavior
type CircuitBreakerConfig struct {
	MaxFailures     int           // Number of failures before opening
	ResetTimeout    time.Duration // Time to wait before trying again
	SuccessThreshold int          // Successes needed to close from half-open
}

// DefaultCircuitBreakerConfig returns sensible defaults
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		MaxFailures:     5,
		ResetTimeout:    60 * time.Second,
		SuccessThreshold: 2,
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	config          *CircuitBreakerConfig
	state           CircuitState
	failures        int
	successes       int
	lastFailureTime time.Time
	mutex           sync.RWMutex
	name            string
	logger          ClientLogger
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}
	
	return &CircuitBreaker{
		config: config,
		state:  CircuitClosed,
		name:   name,
		logger: &DefaultClientLogger{},
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	if !cb.canExecute() {
		return fmt.Errorf("circuit breaker '%s' is OPEN", cb.name)
	}
	
	err := fn(ctx)
	cb.recordResult(err)
	return err
}

// canExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) canExecute() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	
	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		return time.Since(cb.lastFailureTime) >= cb.config.ResetTimeout
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

// recordResult records the result of an execution
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	if err != nil {
		cb.recordFailure()
	} else {
		cb.recordSuccess()
	}
}

// recordFailure records a failure
func (cb *CircuitBreaker) recordFailure() {
	cb.failures++
	cb.lastFailureTime = time.Now()
	
	switch cb.state {
	case CircuitClosed:
		if cb.failures >= cb.config.MaxFailures {
			cb.setState(CircuitOpen)
		}
	case CircuitHalfOpen:
		cb.setState(CircuitOpen)
	}
}

// recordSuccess records a success
func (cb *CircuitBreaker) recordSuccess() {
	switch cb.state {
	case CircuitClosed:
		cb.failures = 0
	case CircuitHalfOpen:
		cb.successes++
		if cb.successes >= cb.config.SuccessThreshold {
			cb.setState(CircuitClosed)
		}
	case CircuitOpen:
		// Transition to half-open on first success after timeout
		cb.setState(CircuitHalfOpen)
		cb.successes = 1
	}
}

// setState changes the circuit breaker state
func (cb *CircuitBreaker) setState(newState CircuitState) {
	if cb.state != newState {
		oldState := cb.state
		cb.state = newState
		
		// Reset counters on state change
		switch newState {
		case CircuitClosed:
			cb.failures = 0
			cb.successes = 0
		case CircuitHalfOpen:
			cb.successes = 0
		}
		
		cb.logger.LogOperation(INFO, "utils", "circuit_breaker", "state", "change", 
			fmt.Sprintf("Circuit breaker '%s' state changed: %s -> %s", cb.name, oldState, newState), 
			map[string]interface{}{"circuit_breaker": cb.name, "old_state": oldState.String(), "new_state": newState.String()})
	}
}

// GetState returns the current state (thread-safe)
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// GetStats returns current statistics
func (cb *CircuitBreaker) GetStats() (CircuitState, int, int) {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state, cb.failures, cb.successes
}
