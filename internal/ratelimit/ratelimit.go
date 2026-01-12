package ratelimit

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tj-smith47/shelly-go/types"
	"golang.org/x/sync/semaphore"

	"github.com/tj-smith47/shelly-cli/internal/tui/debug"
)

// ErrCircuitOpen is returned when a device's circuit breaker is open.
var ErrCircuitOpen = errors.New("circuit breaker is open: device unresponsive")

// DeviceRateLimiter provides per-device request throttling with circuit breakers.
//
// It enforces:
//   - Global concurrent request limit across all devices
//   - Per-device concurrent request limit (lower for Gen1)
//   - Minimum interval between requests to the same device
//   - Circuit breaking for unresponsive devices
type DeviceRateLimiter struct {
	mu        sync.RWMutex
	devices   map[string]*deviceState
	config    Config
	globalSem *semaphore.Weighted
}

// deviceState tracks rate limiting state for a single device.
type deviceState struct {
	address     string
	generation  int
	lastRequest time.Time
	deviceSem   *semaphore.Weighted
	circuit     *CircuitBreaker
	inFlight    atomic.Int32
}

// New creates a new DeviceRateLimiter with the given options.
func New(opts ...Option) *DeviceRateLimiter {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(&config)
	}

	return &DeviceRateLimiter{
		devices:   make(map[string]*deviceState),
		config:    config,
		globalSem: semaphore.NewWeighted(int64(config.Global.MaxConcurrent)),
	}
}

// NewWithConfig creates a DeviceRateLimiter with explicit config.
func NewWithConfig(config Config) *DeviceRateLimiter {
	return &DeviceRateLimiter{
		devices:   make(map[string]*deviceState),
		config:    config,
		globalSem: semaphore.NewWeighted(int64(config.Global.MaxConcurrent)),
	}
}

// Acquire obtains permission to make a request to the specified device.
// It blocks until a slot is available or the context is cancelled.
//
// Parameters:
//   - ctx: Context for cancellation
//   - address: Device address (IP or hostname)
//   - generation: Device generation (1 or 2), use 0 if unknown (treated as Gen1 for safety)
//
// Returns a release function that MUST be called when the request completes,
// or an error if the circuit is open or context cancelled.
func (rl *DeviceRateLimiter) Acquire(ctx context.Context, address string, generation int) (release func(), err error) {
	// Treat unknown generation as Gen1 for safety (more conservative limits)
	if generation < 1 {
		generation = 1
	}
	if generation > 2 {
		generation = 2
	}

	state := rl.getOrCreateState(address, generation)

	// Check circuit breaker first (fast path)
	if !state.circuit.Allow() {
		return nil, ErrCircuitOpen
	}

	// Acquire global semaphore
	if err := rl.globalSem.Acquire(ctx, 1); err != nil {
		return nil, err
	}

	// Acquire per-device semaphore
	if err := state.deviceSem.Acquire(ctx, 1); err != nil {
		rl.globalSem.Release(1)
		return nil, err
	}

	// Enforce minimum interval between requests
	genConfig := rl.config.generationConfig(generation)
	rl.enforceInterval(ctx, state, genConfig.MinInterval)

	// Update state
	rl.mu.Lock()
	state.lastRequest = time.Now()
	rl.mu.Unlock()
	state.inFlight.Add(1)

	// Return release function
	return func() {
		state.inFlight.Add(-1)
		state.deviceSem.Release(1)
		rl.globalSem.Release(1)
	}, nil
}

// TryAcquire attempts to acquire without blocking.
// Returns immediately if a slot is not available.
func (rl *DeviceRateLimiter) TryAcquire(address string, generation int) (release func(), ok bool) {
	if generation < 1 {
		generation = 1
	}
	if generation > 2 {
		generation = 2
	}

	state := rl.getOrCreateState(address, generation)

	// Check circuit breaker
	if !state.circuit.Allow() {
		return nil, false
	}

	// Try global semaphore
	if !rl.globalSem.TryAcquire(1) {
		return nil, false
	}

	// Try per-device semaphore
	if !state.deviceSem.TryAcquire(1) {
		rl.globalSem.Release(1)
		return nil, false
	}

	// Check if we need to wait for interval (non-blocking check)
	rl.mu.RLock()
	genConfig := rl.config.generationConfig(generation)
	elapsed := time.Since(state.lastRequest)
	rl.mu.RUnlock()

	if elapsed < genConfig.MinInterval {
		// Would need to wait, release and return false
		state.deviceSem.Release(1)
		rl.globalSem.Release(1)
		return nil, false
	}

	// Update state
	rl.mu.Lock()
	state.lastRequest = time.Now()
	rl.mu.Unlock()
	state.inFlight.Add(1)

	return func() {
		state.inFlight.Add(-1)
		state.deviceSem.Release(1)
		rl.globalSem.Release(1)
	}, true
}

// RecordSuccess records a successful request for the device.
// This should be called after a request completes successfully.
func (rl *DeviceRateLimiter) RecordSuccess(address string) {
	rl.mu.RLock()
	state, exists := rl.devices[address]
	rl.mu.RUnlock()

	if exists {
		state.circuit.RecordSuccess()
	}
}

// RecordFailure records a failed request for the device.
// This should be called when a request fails (timeout, connection error, etc).
func (rl *DeviceRateLimiter) RecordFailure(address string) {
	rl.mu.RLock()
	state, exists := rl.devices[address]
	rl.mu.RUnlock()

	if exists {
		debug.TraceEvent("circuit: RecordFailure for %s", address)
		state.circuit.RecordFailure()
	}
}

// SetGeneration updates the generation for a device.
// This is useful when generation is detected after initial contact.
func (rl *DeviceRateLimiter) SetGeneration(address string, generation int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if state, exists := rl.devices[address]; exists {
		if state.generation != generation {
			// Update generation and recreate semaphore with new limit
			state.generation = generation
			genConfig := rl.config.generationConfig(generation)
			state.deviceSem = semaphore.NewWeighted(int64(genConfig.MaxConcurrent))
		}
	}
}

// Reset resets the circuit breaker for a device.
func (rl *DeviceRateLimiter) Reset(address string) {
	rl.mu.RLock()
	state, exists := rl.devices[address]
	rl.mu.RUnlock()

	if exists {
		state.circuit.Reset()
	}
}

// ResetAll resets all circuit breakers.
func (rl *DeviceRateLimiter) ResetAll() {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	for _, state := range rl.devices {
		state.circuit.Reset()
	}
}

// Stats returns statistics for a device.
func (rl *DeviceRateLimiter) Stats(address string) (DeviceStats, bool) {
	rl.mu.RLock()
	state, exists := rl.devices[address]
	rl.mu.RUnlock()

	if !exists {
		return DeviceStats{}, false
	}

	return DeviceStats{
		Address:     address,
		Generation:  state.generation,
		InFlight:    int(state.inFlight.Load()),
		LastRequest: state.lastRequest,
		Circuit:     state.circuit.Stats(),
	}, true
}

// AllStats returns statistics for all tracked devices.
func (rl *DeviceRateLimiter) AllStats() []DeviceStats {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	stats := make([]DeviceStats, 0, len(rl.devices))
	for _, state := range rl.devices {
		stats = append(stats, DeviceStats{
			Address:     state.address,
			Generation:  state.generation,
			InFlight:    int(state.inFlight.Load()),
			LastRequest: state.lastRequest,
			Circuit:     state.circuit.Stats(),
		})
	}
	return stats
}

// DeviceStats holds statistics for a single device.
type DeviceStats struct {
	Address     string
	Generation  int
	InFlight    int
	LastRequest time.Time
	Circuit     CircuitStats
}

// IsCircuitOpen returns true if the device's circuit is open.
func (rl *DeviceRateLimiter) IsCircuitOpen(address string) bool {
	rl.mu.RLock()
	state, exists := rl.devices[address]
	rl.mu.RUnlock()

	if !exists {
		return false
	}
	return state.circuit.State() == StateOpen
}

// Config returns the current configuration (read-only).
func (rl *DeviceRateLimiter) Config() Config {
	return rl.config
}

// getOrCreateState returns the state for a device, creating it if needed.
func (rl *DeviceRateLimiter) getOrCreateState(address string, generation int) *deviceState {
	// Fast path: check if already exists
	rl.mu.RLock()
	state, exists := rl.devices[address]
	rl.mu.RUnlock()

	if exists {
		return state
	}

	// Slow path: create new state
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check after acquiring write lock
	if state, exists = rl.devices[address]; exists {
		return state
	}

	genConfig := rl.config.generationConfig(generation)

	state = &deviceState{
		address:    address,
		generation: generation,
		deviceSem:  semaphore.NewWeighted(int64(genConfig.MaxConcurrent)),
		circuit: NewCircuitBreaker(
			genConfig.CircuitThreshold,
			rl.config.Global.CircuitSuccessThreshold,
			rl.config.Global.CircuitOpenDuration,
		),
	}

	rl.devices[address] = state
	return state
}

// enforceInterval waits if needed to respect the minimum interval.
func (rl *DeviceRateLimiter) enforceInterval(ctx context.Context, state *deviceState, minInterval time.Duration) {
	rl.mu.RLock()
	elapsed := time.Since(state.lastRequest)
	rl.mu.RUnlock()

	if elapsed < minInterval {
		waitTime := minInterval - elapsed
		select {
		case <-ctx.Done():
			return
		case <-time.After(waitTime):
			return
		}
	}
}

// IsConnectivityFailure determines if an error represents an actual connectivity
// failure vs an expected API response (like "component not found").
//
// Returns true for errors that indicate the device is unreachable or unresponsive:
//   - Timeouts (context deadline exceeded)
//   - Device offline
//   - Connection refused/reset
//
// Returns false for errors where the device responded correctly:
//   - Component/resource not found (404)
//   - Method not supported
//   - Authentication errors
//   - Invalid parameters
func IsConnectivityFailure(err error) bool {
	if err == nil {
		return false
	}

	// These errors mean the device IS responding, just not with what we wanted.
	// They should NOT count as circuit breaker failures.
	if errors.Is(err, types.ErrNotFound) ||
		errors.Is(err, types.ErrNotSupported) ||
		errors.Is(err, types.ErrAuth) ||
		errors.Is(err, types.ErrInvalidParam) ||
		errors.Is(err, types.ErrInvalidResponse) {
		return false
	}

	// These errors indicate actual connectivity problems
	if errors.Is(err, types.ErrDeviceOffline) ||
		errors.Is(err, types.ErrTimeout) ||
		errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, context.Canceled) {
		return true
	}

	// For other errors, check if they contain connectivity-related keywords
	errStr := strings.ToLower(err.Error())
	connectivityKeywords := []string{
		"connection refused",
		"connection reset",
		"no route to host",
		"network is unreachable",
		"i/o timeout",
		"deadline exceeded",
	}
	for _, kw := range connectivityKeywords {
		if strings.Contains(errStr, kw) {
			return true
		}
	}

	// Default: treat unknown errors as NOT connectivity failures
	// This prevents false circuit breaker trips from unexpected error types
	return false
}
