package fcm

import (
	"context"
	"sync"
	"time"
)

// CircuitState represents the state of the circuit breaker.
type CircuitState int

const (
	// CircuitStateClosed means the circuit is closed and requests pass through.
	CircuitStateClosed CircuitState = iota
	// CircuitStateOpen means the circuit is open and requests fail fast.
	CircuitStateOpen
	// CircuitStateHalfOpen means the circuit is testing if it should close.
	CircuitStateHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case CircuitStateClosed:
		return "closed"
	case CircuitStateOpen:
		return "open"
	case CircuitStateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig holds configuration for the circuit breaker.
type CircuitBreakerConfig struct {
	// FailureThreshold is the number of consecutive failures before opening the circuit.
	FailureThreshold int
	// SuccessThreshold is the number of consecutive successes in half-open state before closing.
	SuccessThreshold int
	// OpenDuration is how long the circuit stays open before transitioning to half-open.
	OpenDuration time.Duration
	// HalfOpenMaxCalls is the maximum number of calls allowed in half-open state.
	HalfOpenMaxCalls int
}

// DefaultCircuitBreakerConfig returns default configuration for FCM circuit breaker.
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold: 5,
		SuccessThreshold: 3,
		OpenDuration:     30 * time.Second,
		HalfOpenMaxCalls: 3,
	}
}

// CircuitBreaker implements the circuit breaker pattern for FCM calls.
// Bug 50 fix: Prevents cascading failures when FCM is unavailable.
type CircuitBreaker struct {
	config  CircuitBreakerConfig
	state   CircuitState
	mu      sync.RWMutex

	failures       int
	successes     int
	halfOpenCalls int

	lastFailureTime time.Time
}

// NewCircuitBreaker creates a new CircuitBreaker with the given config.
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  CircuitStateClosed,
	}
}

// Allow checks if a request should be allowed through the circuit breaker.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitStateClosed:
		return true
	case CircuitStateOpen:
		// Check if it's time to transition to half-open
		if time.Since(cb.lastFailureTime) >= cb.config.OpenDuration {
			cb.transitionTo(CircuitStateHalfOpen)
			return true
		}
		return false
	case CircuitStateHalfOpen:
		// Allow limited calls in half-open state
		if cb.halfOpenCalls < cb.config.HalfOpenMaxCalls {
			cb.halfOpenCalls++
			return true
		}
		return false
	default:
		return false
	}
}

// RecordSuccess records a successful call.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitStateClosed:
		cb.failures = 0
	case CircuitStateHalfOpen:
		cb.successes++
		if cb.successes >= cb.config.SuccessThreshold {
			cb.transitionTo(CircuitStateClosed)
		}
	}
}

// RecordFailure records a failed call.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.lastFailureTime = time.Now()

	switch cb.state {
	case CircuitStateClosed:
		cb.failures++
		if cb.failures >= cb.config.FailureThreshold {
			cb.transitionTo(CircuitStateOpen)
		}
	case CircuitStateHalfOpen:
		// Any failure in half-open state opens the circuit
		cb.transitionTo(CircuitStateOpen)
	}
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// transitionTo changes the circuit state (must be called with lock held).
func (cb *CircuitBreaker) transitionTo(state CircuitState) {
	cb.state = state

	switch state {
	case CircuitStateClosed:
		cb.failures = 0
		cb.successes = 0
		cb.halfOpenCalls = 0
	case CircuitStateOpen:
		// Already set lastFailureTime
	case CircuitStateHalfOpen:
		cb.successes = 0
		cb.halfOpenCalls = 0
	}
}

// CircuitBreakerClient wraps an FCMClient with circuit breaker protection.
type CircuitBreakerClient struct {
	client        *Client
	circuitBreaker *CircuitBreaker
}

// NewCircuitBreakerClient creates a new CircuitBreakerClient.
func NewCircuitBreakerClient(client *Client) *CircuitBreakerClient {
	return &CircuitBreakerClient{
		client:        client,
		circuitBreaker: NewCircuitBreaker(DefaultCircuitBreakerConfig()),
	}
}

// SendSilentWake sends a silent wake notification with circuit breaker protection.
// Returns ErrFCMCircuitOpen if the circuit is open.
func (c *CircuitBreakerClient) SendSilentWake(ctx context.Context, wake SilentWake) error {
	if !c.circuitBreaker.Allow() {
		return ErrFCMCircuitOpen
	}

	err := c.client.SendSilentWake(ctx, wake)
	if err != nil {
		c.circuitBreaker.RecordFailure()
		return err
	}

	c.circuitBreaker.RecordSuccess()
	return nil
}

// State returns the current circuit breaker state.
func (c *CircuitBreakerClient) State() CircuitState {
	return c.circuitBreaker.State()
}
