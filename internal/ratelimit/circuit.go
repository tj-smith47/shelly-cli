package ratelimit

import (
	"sync"
	"time"
)

// State represents the circuit breaker state.
type State int

const (
	// StateClosed is normal operation - requests are allowed.
	StateClosed State = iota
	// StateOpen blocks requests - device is unresponsive.
	StateOpen
	// StateHalfOpen allows probe requests to test if device recovered.
	StateHalfOpen
)

// String returns a human-readable state name.
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker prevents hammering unresponsive devices.
//
// State machine:
//   - Closed: Normal operation, requests allowed
//   - Open: Blocking requests, device unresponsive
//   - HalfOpen: Testing if device recovered (allows limited probe requests)
//
// Transitions:
//   - Closed → Open: After N consecutive failures
//   - Open → HalfOpen: After timeout duration
//   - HalfOpen → Closed: After M consecutive successes
//   - HalfOpen → Open: On any failure
type CircuitBreaker struct {
	mu              sync.Mutex
	state           State
	failCount       int       // Consecutive failures
	successCount    int       // Consecutive successes (in half-open)
	lastStateChange time.Time // When state last changed
	lastFailure     time.Time // Last failure time

	// Configuration
	failThreshold    int           // Failures to open (default: 3)
	successThreshold int           // Successes to close (default: 2)
	openDuration     time.Duration // How long to stay open (default: 60s)
}

// NewCircuitBreaker creates a circuit breaker with the given thresholds.
func NewCircuitBreaker(failThreshold, successThreshold int, openDuration time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            StateClosed,
		failThreshold:    failThreshold,
		successThreshold: successThreshold,
		openDuration:     openDuration,
		lastStateChange:  time.Now(),
	}
}

// Allow checks if a request should be allowed.
// Returns true if the request can proceed, false if blocked.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true

	case StateOpen:
		// Check if we should transition to half-open
		if time.Since(cb.lastStateChange) >= cb.openDuration {
			cb.transitionTo(StateHalfOpen)
			return true // Allow probe request
		}
		return false // Still blocking

	case StateHalfOpen:
		// Allow probe requests in half-open state
		return true

	default:
		return false
	}
}

// RecordSuccess records a successful request.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failCount = 0 // Reset consecutive failures

	switch cb.state {
	case StateClosed:
		// Nothing to do, already healthy

	case StateHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.transitionTo(StateClosed)
		}

	case StateOpen:
		// Shouldn't happen (requests blocked), but reset anyway
		cb.transitionTo(StateHalfOpen)
	}
}

// RecordFailure records a failed request.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.lastFailure = time.Now()
	cb.failCount++
	cb.successCount = 0 // Reset consecutive successes

	switch cb.state {
	case StateClosed:
		if cb.failCount >= cb.failThreshold {
			cb.transitionTo(StateOpen)
		}

	case StateHalfOpen:
		// Any failure in half-open goes back to open
		cb.transitionTo(StateOpen)

	case StateOpen:
		// Already open, just refresh the timer
		cb.lastStateChange = time.Now()
	}
}

// State returns the current circuit state.
func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Stats returns circuit breaker statistics.
func (cb *CircuitBreaker) Stats() CircuitStats {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	return CircuitStats{
		State:           cb.state,
		FailCount:       cb.failCount,
		SuccessCount:    cb.successCount,
		LastStateChange: cb.lastStateChange,
		LastFailure:     cb.lastFailure,
	}
}

// CircuitStats holds circuit breaker statistics.
type CircuitStats struct {
	State           State
	FailCount       int
	SuccessCount    int
	LastStateChange time.Time
	LastFailure     time.Time
}

// Reset resets the circuit breaker to closed state.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failCount = 0
	cb.successCount = 0
	cb.lastStateChange = time.Now()
}

// transitionTo changes state (must be called with lock held).
func (cb *CircuitBreaker) transitionTo(newState State) {
	cb.state = newState
	cb.lastStateChange = time.Now()
	cb.successCount = 0
	if newState == StateClosed {
		cb.failCount = 0
	}
}
