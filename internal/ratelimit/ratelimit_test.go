package ratelimit

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestCircuitBreaker_InitialState(t *testing.T) {
	t.Parallel()
	cb := NewCircuitBreaker(3, 2, 60*time.Second)

	if cb.State() != StateClosed {
		t.Errorf("expected initial state Closed, got %v", cb.State())
	}

	if !cb.Allow() {
		t.Error("expected Allow() to return true in closed state")
	}
}

func TestCircuitBreaker_OpensAfterFailures(t *testing.T) {
	t.Parallel()
	cb := NewCircuitBreaker(3, 2, 60*time.Second)

	// First two failures should keep circuit closed
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.State() != StateClosed {
		t.Errorf("expected state Closed after 2 failures, got %v", cb.State())
	}

	// Third failure should open the circuit
	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Errorf("expected state Open after 3 failures, got %v", cb.State())
	}

	// Requests should be blocked
	if cb.Allow() {
		t.Error("expected Allow() to return false in open state")
	}
}

func TestCircuitBreaker_SuccessResetsFailCount(t *testing.T) {
	t.Parallel()
	cb := NewCircuitBreaker(3, 2, 60*time.Second)

	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordSuccess() // Should reset fail count

	// Now we need 3 more failures to open
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.State() != StateClosed {
		t.Errorf("expected state Closed, got %v", cb.State())
	}

	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Errorf("expected state Open, got %v", cb.State())
	}
}

func TestCircuitBreaker_TransitionsToHalfOpen(t *testing.T) {
	t.Parallel()
	// Use short duration for testing
	cb := NewCircuitBreaker(1, 2, 10*time.Millisecond)

	cb.RecordFailure() // Opens circuit
	if cb.State() != StateOpen {
		t.Errorf("expected state Open, got %v", cb.State())
	}

	// Wait for timeout
	time.Sleep(20 * time.Millisecond)

	// Next Allow() should transition to half-open
	if !cb.Allow() {
		t.Error("expected Allow() to return true after timeout")
	}
	if cb.State() != StateHalfOpen {
		t.Errorf("expected state HalfOpen, got %v", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenToClosedOnSuccess(t *testing.T) {
	t.Parallel()
	cb := NewCircuitBreaker(1, 2, 10*time.Millisecond)

	cb.RecordFailure() // Opens circuit
	time.Sleep(20 * time.Millisecond)
	cb.Allow() // Transitions to half-open

	// First success
	cb.RecordSuccess()
	if cb.State() != StateHalfOpen {
		t.Errorf("expected state HalfOpen after 1 success, got %v", cb.State())
	}

	// Second success should close
	cb.RecordSuccess()
	if cb.State() != StateClosed {
		t.Errorf("expected state Closed after 2 successes, got %v", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenToOpenOnFailure(t *testing.T) {
	t.Parallel()
	cb := NewCircuitBreaker(1, 2, 10*time.Millisecond)

	cb.RecordFailure() // Opens circuit
	time.Sleep(20 * time.Millisecond)
	cb.Allow() // Transitions to half-open

	// Any failure in half-open should reopen
	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Errorf("expected state Open after failure in half-open, got %v", cb.State())
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	t.Parallel()
	cb := NewCircuitBreaker(1, 2, 60*time.Second)

	cb.RecordFailure() // Opens circuit
	if cb.State() != StateOpen {
		t.Errorf("expected state Open, got %v", cb.State())
	}

	cb.Reset()
	if cb.State() != StateClosed {
		t.Errorf("expected state Closed after reset, got %v", cb.State())
	}
	if !cb.Allow() {
		t.Error("expected Allow() to return true after reset")
	}
}

func TestCircuitBreaker_Stats(t *testing.T) {
	t.Parallel()
	cb := NewCircuitBreaker(3, 2, 60*time.Second)

	cb.RecordFailure()
	cb.RecordFailure()

	stats := cb.Stats()
	if stats.State != StateClosed {
		t.Errorf("expected state Closed, got %v", stats.State)
	}
	if stats.FailCount != 2 {
		t.Errorf("expected FailCount 2, got %d", stats.FailCount)
	}
}

func TestDeviceRateLimiter_BasicAcquireRelease(t *testing.T) {
	t.Parallel()
	rl := New()
	ctx := context.Background()

	release, err := rl.Acquire(ctx, "192.168.1.100", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stats, ok := rl.Stats("192.168.1.100")
	if !ok {
		t.Fatal("expected device to be tracked")
	}
	if stats.InFlight != 1 {
		t.Errorf("expected InFlight 1, got %d", stats.InFlight)
	}

	release()

	stats, _ = rl.Stats("192.168.1.100")
	if stats.InFlight != 0 {
		t.Errorf("expected InFlight 0 after release, got %d", stats.InFlight)
	}
}

func TestDeviceRateLimiter_Gen1Limits(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()
	config.Gen1.MaxConcurrent = 1
	config.Gen1.MinInterval = 10 * time.Millisecond
	config.Global.MaxConcurrent = 10

	rl := NewWithConfig(config)
	ctx := context.Background()

	// First acquire should succeed
	release1, err := rl.Acquire(ctx, "192.168.1.100", 1)
	if err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}

	// Second acquire should block (Gen1 limit is 1)
	// Use TryAcquire to test without blocking
	_, ok := rl.TryAcquire("192.168.1.100", 1)
	if ok {
		t.Error("expected TryAcquire to fail when Gen1 device is at limit")
	}

	release1()

	// Now acquire should work (after interval)
	time.Sleep(15 * time.Millisecond)
	release2, ok := rl.TryAcquire("192.168.1.100", 1)
	if !ok {
		t.Error("expected TryAcquire to succeed after release and interval")
	}
	if release2 != nil {
		release2()
	}
}

func TestDeviceRateLimiter_Gen2Limits(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()
	config.Gen2.MaxConcurrent = 3
	config.Gen2.MinInterval = 1 * time.Millisecond
	config.Global.MaxConcurrent = 10

	rl := NewWithConfig(config)
	ctx := context.Background()

	// Acquire 3 slots (Gen2 limit)
	releases := make([]func(), 0, 3)
	for i := range 3 {
		release, err := rl.Acquire(ctx, "192.168.1.100", 2)
		if err != nil {
			t.Fatalf("acquire %d failed: %v", i, err)
		}
		releases = append(releases, release)
	}

	stats, _ := rl.Stats("192.168.1.100")
	if stats.InFlight != 3 {
		t.Errorf("expected InFlight 3, got %d", stats.InFlight)
	}

	// Fourth acquire should block
	_, ok := rl.TryAcquire("192.168.1.100", 2)
	if ok {
		t.Error("expected TryAcquire to fail when Gen2 device is at limit")
	}

	// Release all
	for _, release := range releases {
		release()
	}
}

func TestDeviceRateLimiter_GlobalLimit(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()
	config.Gen2.MaxConcurrent = 10
	config.Gen2.MinInterval = 1 * time.Millisecond
	config.Global.MaxConcurrent = 3

	rl := NewWithConfig(config)
	ctx := context.Background()

	// Acquire 3 slots across different devices (global limit)
	releases := make([]func(), 0, 3)
	for i := range 3 {
		addr := "192.168.1." + string(rune('1'+i))
		release, err := rl.Acquire(ctx, addr, 2)
		if err != nil {
			t.Fatalf("acquire %d failed: %v", i, err)
		}
		releases = append(releases, release)
	}

	// Fourth acquire should fail (global limit)
	_, ok := rl.TryAcquire("192.168.1.50", 2)
	if ok {
		t.Error("expected TryAcquire to fail when at global limit")
	}

	// Release one
	releases[0]()

	// Now should work
	time.Sleep(5 * time.Millisecond) // Wait for interval
	release, ok := rl.TryAcquire("192.168.1.50", 2)
	if !ok {
		t.Error("expected TryAcquire to succeed after releasing global slot")
	}
	if release != nil {
		release()
	}

	// Release remaining
	for _, r := range releases[1:] {
		r()
	}
}

func TestDeviceRateLimiter_CircuitBreaker(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()
	config.Gen2.CircuitThreshold = 2
	config.Global.CircuitOpenDuration = 10 * time.Millisecond

	rl := NewWithConfig(config)
	ctx := context.Background()

	// Acquire and record failures
	release, err := rl.Acquire(ctx, "192.168.1.100", 2)
	if err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}
	rl.RecordFailure("192.168.1.100")
	release()

	release, err = rl.Acquire(ctx, "192.168.1.100", 2)
	if err != nil {
		t.Fatalf("second acquire failed: %v", err)
	}
	rl.RecordFailure("192.168.1.100")
	release()

	// Circuit should now be open
	if !rl.IsCircuitOpen("192.168.1.100") {
		t.Error("expected circuit to be open after 2 failures")
	}

	// Acquire should fail with ErrCircuitOpen
	_, err = rl.Acquire(ctx, "192.168.1.100", 2)
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}

	// Wait for circuit to transition to half-open
	time.Sleep(15 * time.Millisecond)

	// Now acquire should work (half-open allows probe)
	release, err = rl.Acquire(ctx, "192.168.1.100", 2)
	if err != nil {
		t.Fatalf("expected acquire to succeed in half-open, got %v", err)
	}
	rl.RecordSuccess("192.168.1.100")
	release()

	// Need one more success to close circuit
	time.Sleep(5 * time.Millisecond)
	release, err = rl.Acquire(ctx, "192.168.1.100", 2)
	if err != nil {
		t.Fatalf("acquire after first success failed: %v", err)
	}
	rl.RecordSuccess("192.168.1.100")
	release()

	if rl.IsCircuitOpen("192.168.1.100") {
		t.Error("expected circuit to be closed after successes")
	}
}

func TestDeviceRateLimiter_MinInterval(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()
	config.Gen2.MaxConcurrent = 1
	config.Gen2.MinInterval = 50 * time.Millisecond
	config.Global.MaxConcurrent = 10

	rl := NewWithConfig(config)
	ctx := context.Background()

	// First acquire
	start := time.Now()
	release1, err := rl.Acquire(ctx, "192.168.1.100", 2)
	if err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}
	release1()

	// Second acquire should wait for interval
	release2, err := rl.Acquire(ctx, "192.168.1.100", 2)
	if err != nil {
		t.Fatalf("second acquire failed: %v", err)
	}
	elapsed := time.Since(start)
	release2()

	// Should have waited at least 50ms (some tolerance for timing)
	if elapsed < 45*time.Millisecond {
		t.Errorf("expected at least 45ms delay, got %v", elapsed)
	}
}

func TestDeviceRateLimiter_ContextCancellation(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()
	config.Gen2.MaxConcurrent = 1
	config.Global.MaxConcurrent = 10

	rl := NewWithConfig(config)

	// Acquire first slot
	ctx := context.Background()
	release, err := rl.Acquire(ctx, "192.168.1.100", 2)
	if err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}

	// Try to acquire with cancelled context
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = rl.Acquire(cancelCtx, "192.168.1.100", 2)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	release()
}

func TestDeviceRateLimiter_SetGeneration(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()
	config.Gen1.MaxConcurrent = 1
	config.Gen2.MaxConcurrent = 3

	rl := NewWithConfig(config)
	ctx := context.Background()

	// Initially treat as Gen1
	release, err := rl.Acquire(ctx, "192.168.1.100", 1)
	if err != nil {
		t.Fatalf("acquire failed: %v", err)
	}
	release()

	stats, _ := rl.Stats("192.168.1.100")
	if stats.Generation != 1 {
		t.Errorf("expected generation 1, got %d", stats.Generation)
	}

	// Update to Gen2
	rl.SetGeneration("192.168.1.100", 2)

	stats, _ = rl.Stats("192.168.1.100")
	if stats.Generation != 2 {
		t.Errorf("expected generation 2, got %d", stats.Generation)
	}
}

func TestDeviceRateLimiter_ResetAll(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()
	config.Gen2.CircuitThreshold = 1
	config.Global.CircuitOpenDuration = 60 * time.Second

	rl := NewWithConfig(config)
	ctx := context.Background()

	// Open circuits on two devices
	for _, addr := range []string{"192.168.1.100", "192.168.1.101"} {
		release, err := rl.Acquire(ctx, addr, 2)
		if err != nil {
			t.Fatalf("acquire for %s failed: %v", addr, err)
		}
		rl.RecordFailure(addr)
		release()
	}

	if !rl.IsCircuitOpen("192.168.1.100") || !rl.IsCircuitOpen("192.168.1.101") {
		t.Error("expected both circuits to be open")
	}

	rl.ResetAll()

	if rl.IsCircuitOpen("192.168.1.100") || rl.IsCircuitOpen("192.168.1.101") {
		t.Error("expected both circuits to be closed after ResetAll")
	}
}

func TestDeviceRateLimiter_AllStats(t *testing.T) {
	t.Parallel()
	rl := New()
	ctx := context.Background()

	// Create some devices
	for _, addr := range []string{"192.168.1.100", "192.168.1.101", "192.168.1.102"} {
		release, err := rl.Acquire(ctx, addr, 2)
		if err != nil {
			t.Fatalf("acquire for %s failed: %v", addr, err)
		}
		release()
	}

	stats := rl.AllStats()
	if len(stats) != 3 {
		t.Errorf("expected 3 devices, got %d", len(stats))
	}
}

func TestDeviceRateLimiter_ConcurrentAccess(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()
	config.Gen2.MaxConcurrent = 3
	config.Gen2.MinInterval = 1 * time.Millisecond
	config.Global.MaxConcurrent = 10

	rl := NewWithConfig(config)

	var wg sync.WaitGroup
	errChan := make(chan error, 100)

	// Launch 20 goroutines trying to acquire the same device
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()

			release, err := rl.Acquire(ctx, "192.168.1.100", 2)
			if err != nil {
				if !errors.Is(err, context.DeadlineExceeded) {
					errChan <- err
				}
				return
			}

			// Simulate some work
			time.Sleep(5 * time.Millisecond)
			release()
		}()
	}

	wg.Wait()
	close(errChan)

	// Check for unexpected errors (DeadlineExceeded is expected for some)
	for err := range errChan {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify no in-flight requests remain
	stats, _ := rl.Stats("192.168.1.100")
	if stats.InFlight != 0 {
		t.Errorf("expected InFlight 0 after all releases, got %d", stats.InFlight)
	}
}

func TestDeviceRateLimiter_UnknownGeneration(t *testing.T) {
	t.Parallel()
	rl := New()
	ctx := context.Background()

	// Generation 0 should be treated as Gen1 (conservative)
	release, err := rl.Acquire(ctx, "192.168.1.100", 0)
	if err != nil {
		t.Fatalf("acquire failed: %v", err)
	}
	release()

	stats, _ := rl.Stats("192.168.1.100")
	if stats.Generation != 1 {
		t.Errorf("expected generation 0 to be treated as 1, got %d", stats.Generation)
	}

	// Generation 99 should be treated as Gen2 (capped)
	release, err = rl.Acquire(ctx, "192.168.1.101", 99)
	if err != nil {
		t.Fatalf("acquire for gen99 failed: %v", err)
	}
	release()

	stats, _ = rl.Stats("192.168.1.101")
	if stats.Generation != 2 {
		t.Errorf("expected generation 99 to be treated as 2, got %d", stats.Generation)
	}
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()

	// Verify Gen1 defaults (conservative)
	if config.Gen1.MaxConcurrent != 1 {
		t.Errorf("Gen1.MaxConcurrent: expected 1, got %d", config.Gen1.MaxConcurrent)
	}
	if config.Gen1.MinInterval != 2*time.Second {
		t.Errorf("Gen1.MinInterval: expected 2s, got %v", config.Gen1.MinInterval)
	}

	// Verify Gen2 defaults (more permissive)
	if config.Gen2.MaxConcurrent != 3 {
		t.Errorf("Gen2.MaxConcurrent: expected 3, got %d", config.Gen2.MaxConcurrent)
	}
	if config.Gen2.MinInterval != 500*time.Millisecond {
		t.Errorf("Gen2.MinInterval: expected 500ms, got %v", config.Gen2.MinInterval)
	}

	// Verify global defaults
	if config.Global.MaxConcurrent != 20 {
		t.Errorf("Global.MaxConcurrent: expected 20, got %d", config.Global.MaxConcurrent)
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()
	rl := New(
		WithGen1MinInterval(5*time.Second),
		WithGen2MinInterval(1*time.Second),
		WithGen1MaxConcurrent(2),
		WithGen2MaxConcurrent(5),
		WithGlobalMaxConcurrent(10),
		WithCircuitOpenDuration(120*time.Second),
	)

	config := rl.Config()

	if config.Gen1.MinInterval != 5*time.Second {
		t.Errorf("Gen1.MinInterval: expected 5s, got %v", config.Gen1.MinInterval)
	}
	if config.Gen2.MinInterval != 1*time.Second {
		t.Errorf("Gen2.MinInterval: expected 1s, got %v", config.Gen2.MinInterval)
	}
	if config.Gen1.MaxConcurrent != 2 {
		t.Errorf("Gen1.MaxConcurrent: expected 2, got %d", config.Gen1.MaxConcurrent)
	}
	if config.Gen2.MaxConcurrent != 5 {
		t.Errorf("Gen2.MaxConcurrent: expected 5, got %d", config.Gen2.MaxConcurrent)
	}
	if config.Global.MaxConcurrent != 10 {
		t.Errorf("Global.MaxConcurrent: expected 10, got %d", config.Global.MaxConcurrent)
	}
	if config.Global.CircuitOpenDuration != 120*time.Second {
		t.Errorf("Global.CircuitOpenDuration: expected 120s, got %v", config.Global.CircuitOpenDuration)
	}
}

func TestState_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		state State
		want  string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
		{State(99), "unknown"},
	}

	for _, tt := range tests {
		got := tt.state.String()
		if got != tt.want {
			t.Errorf("State(%d).String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}

func TestDeviceRateLimiter_RecordOnUnknownDevice(t *testing.T) {
	t.Parallel()
	rl := New()

	// Should not panic on unknown device
	rl.RecordSuccess("unknown-device")
	rl.RecordFailure("unknown-device")

	// Should return false for unknown device
	if rl.IsCircuitOpen("unknown-device") {
		t.Error("expected IsCircuitOpen to return false for unknown device")
	}
}

func TestDeviceRateLimiter_ResetUnknownDevice(t *testing.T) {
	t.Parallel()
	rl := New()

	// Should not panic on unknown device
	rl.Reset("unknown-device")
}

func TestDeviceRateLimiter_SetGenerationUnknownDevice(t *testing.T) {
	t.Parallel()
	rl := New()

	// Should not panic on unknown device
	rl.SetGeneration("unknown-device", 2)
}

func TestDeviceRateLimiter_StatsUnknownDevice(t *testing.T) {
	t.Parallel()
	rl := New()

	_, ok := rl.Stats("unknown-device")
	if ok {
		t.Error("expected Stats to return false for unknown device")
	}
}

func TestCircuitBreaker_LastFailureTracking(t *testing.T) {
	t.Parallel()
	cb := NewCircuitBreaker(3, 2, 60*time.Second)

	// No failures yet
	stats := cb.Stats()
	if !stats.LastFailure.IsZero() {
		t.Error("expected LastFailure to be zero initially")
	}

	// Record a failure
	cb.RecordFailure()
	stats = cb.Stats()
	if stats.LastFailure.IsZero() {
		t.Error("expected LastFailure to be set after failure")
	}
}

func TestGenerationConfig(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()

	gen1 := config.generationConfig(1)
	if gen1.MaxConcurrent != config.Gen1.MaxConcurrent {
		t.Error("generationConfig(1) should return Gen1 config")
	}

	gen2 := config.generationConfig(2)
	if gen2.MaxConcurrent != config.Gen2.MaxConcurrent {
		t.Error("generationConfig(2) should return Gen2 config")
	}

	// Any other value should return Gen2 (default)
	gen3 := config.generationConfig(3)
	if gen3.MaxConcurrent != config.Gen2.MaxConcurrent {
		t.Error("generationConfig(3) should return Gen2 config as default")
	}
}

func TestCircuitBreaker_RecordSuccessInOpenState(t *testing.T) {
	t.Parallel()
	cb := NewCircuitBreaker(1, 2, 60*time.Second)

	// Open the circuit
	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Fatalf("expected state Open, got %v", cb.State())
	}

	// Recording success in open state (shouldn't normally happen, but handles edge case)
	cb.RecordSuccess()
	// Should transition to half-open
	if cb.State() != StateHalfOpen {
		t.Errorf("expected state HalfOpen after success in open, got %v", cb.State())
	}
}

func TestCircuitBreaker_RecordFailureInOpenState(t *testing.T) {
	t.Parallel()
	cb := NewCircuitBreaker(1, 2, 60*time.Second)

	// Open the circuit
	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Fatalf("expected state Open, got %v", cb.State())
	}

	initialChange := cb.Stats().LastStateChange

	// Wait a bit and record another failure while open
	time.Sleep(10 * time.Millisecond)
	cb.RecordFailure()

	// Should still be open but with updated timer
	if cb.State() != StateOpen {
		t.Errorf("expected state Open, got %v", cb.State())
	}

	// LastStateChange should be updated (timer refreshed)
	if !cb.Stats().LastStateChange.After(initialChange) {
		t.Error("expected LastStateChange to be updated after failure in open state")
	}
}

func TestDeviceRateLimiter_TryAcquire_CircuitOpen(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()
	config.Gen2.CircuitThreshold = 1
	config.Global.CircuitOpenDuration = 60 * time.Second

	rl := NewWithConfig(config)
	ctx := context.Background()

	// Acquire and fail to open circuit
	release, err := rl.Acquire(ctx, "192.168.1.100", 2)
	if err != nil {
		t.Fatalf("acquire failed: %v", err)
	}
	rl.RecordFailure("192.168.1.100")
	release()

	// TryAcquire should fail when circuit is open
	_, ok := rl.TryAcquire("192.168.1.100", 2)
	if ok {
		t.Error("expected TryAcquire to fail when circuit is open")
	}
}

func TestDeviceRateLimiter_TryAcquire_IntervalNotElapsed(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()
	config.Gen2.MaxConcurrent = 10
	config.Gen2.MinInterval = 100 * time.Millisecond
	config.Global.MaxConcurrent = 10

	rl := NewWithConfig(config)
	ctx := context.Background()

	// First acquire
	release, err := rl.Acquire(ctx, "192.168.1.100", 2)
	if err != nil {
		t.Fatalf("acquire failed: %v", err)
	}
	release()

	// Immediately try again - should fail due to interval not elapsed
	_, ok := rl.TryAcquire("192.168.1.100", 2)
	if ok {
		t.Error("expected TryAcquire to fail when interval has not elapsed")
	}
}

func TestDeviceRateLimiter_TryAcquire_GlobalSemaphoreFull(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()
	config.Gen2.MaxConcurrent = 10
	config.Gen2.MinInterval = 1 * time.Millisecond
	config.Global.MaxConcurrent = 1 // Only 1 global slot

	rl := NewWithConfig(config)
	ctx := context.Background()

	// First acquire takes the only global slot
	release, err := rl.Acquire(ctx, "192.168.1.100", 2)
	if err != nil {
		t.Fatalf("acquire failed: %v", err)
	}

	// TryAcquire on different device should fail (global full)
	_, ok := rl.TryAcquire("192.168.1.101", 2)
	if ok {
		t.Error("expected TryAcquire to fail when global semaphore is full")
	}

	release()
}

func TestDeviceRateLimiter_TryAcquire_DeviceSemaphoreFull(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()
	config.Gen2.MaxConcurrent = 1 // Only 1 per device
	config.Gen2.MinInterval = 1 * time.Millisecond
	config.Global.MaxConcurrent = 10

	rl := NewWithConfig(config)
	ctx := context.Background()

	// First acquire takes the device slot
	release, err := rl.Acquire(ctx, "192.168.1.100", 2)
	if err != nil {
		t.Fatalf("acquire failed: %v", err)
	}

	// TryAcquire on same device should fail (device full)
	_, ok := rl.TryAcquire("192.168.1.100", 2)
	if ok {
		t.Error("expected TryAcquire to fail when device semaphore is full")
	}

	release()
}

func TestDeviceRateLimiter_ConcurrentStateCreation(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()
	config.Gen2.MinInterval = 1 * time.Millisecond
	config.Global.MaxConcurrent = 100

	rl := NewWithConfig(config)

	// Launch many goroutines trying to create state for the same device simultaneously
	// This tests the double-check locking pattern in getOrCreateState
	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			release, err := rl.Acquire(ctx, "concurrent-device", 2)
			if err != nil {
				return // timeout is fine
			}
			time.Sleep(time.Millisecond)
			release()
		}()
	}

	wg.Wait()

	// Verify only one state was created
	stats := rl.AllStats()
	count := 0
	for _, s := range stats {
		if s.Address == "concurrent-device" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 device state, got %d", count)
	}
}

func TestCircuitBreaker_Allow_DefaultCase(t *testing.T) {
	t.Parallel()
	// Test the default case in Allow() - when state is invalid
	// This requires accessing internal state, so we use reflection or
	// just verify normal operation covers the expected paths
	cb := NewCircuitBreaker(1, 2, 10*time.Millisecond)

	// Test all valid states through Allow
	// Closed state
	if !cb.Allow() {
		t.Error("expected Allow() to return true in closed state")
	}

	// Open state
	cb.RecordFailure()
	if cb.Allow() {
		t.Error("expected Allow() to return false in open state")
	}

	// Wait for timeout then half-open
	time.Sleep(15 * time.Millisecond)
	if !cb.Allow() {
		t.Error("expected Allow() to return true in half-open state")
	}
}

func TestDeviceRateLimiter_Reset_NonExistingDevice(t *testing.T) {
	t.Parallel()
	rl := New()

	// Reset on non-existing device should not create state
	rl.Reset("non-existing-device")

	// Verify no state was created
	_, exists := rl.Stats("non-existing-device")
	if exists {
		t.Error("Reset should not create device state")
	}
}

func TestDeviceRateLimiter_EnforceInterval_ContextCancelled(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()
	config.Gen2.MaxConcurrent = 1
	config.Gen2.MinInterval = 200 * time.Millisecond // Interval to wait
	config.Global.MaxConcurrent = 10

	rl := NewWithConfig(config)
	ctx := context.Background()

	// First acquire
	release1, err := rl.Acquire(ctx, "192.168.1.100", 2)
	if err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}
	release1()

	// Second acquire with short timeout - context cancelled during interval wait
	// The enforceInterval will return early when context is cancelled
	shortCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	release2, err := rl.Acquire(shortCtx, "192.168.1.100", 2)
	elapsed := time.Since(start)

	// Acquire still succeeds (semaphores were obtained), but interval wait was cut short
	// The context cancellation in enforceInterval just returns early, it doesn't fail the acquire
	if err != nil {
		// If it failed with context error, that's also valid (semaphore acquisition might fail)
		if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
			t.Errorf("unexpected error: %v", err)
		}
		return
	}

	// If successful, release
	if release2 != nil {
		release2()
	}

	// Should have returned quickly (context timeout ~50ms), not waited full 200ms
	if elapsed > 100*time.Millisecond {
		t.Errorf("expected quick return on context cancel, but waited %v", elapsed)
	}
}

func TestDeviceRateLimiter_TryAcquire_GenerationNormalization(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()
	config.Gen2.MaxConcurrent = 10
	config.Gen2.MinInterval = 1 * time.Millisecond
	config.Global.MaxConcurrent = 10

	rl := NewWithConfig(config)

	// Test generation 0 normalization in TryAcquire (treated as Gen1)
	release, ok := rl.TryAcquire("192.168.1.100", 0)
	if !ok {
		t.Error("expected TryAcquire to succeed with generation 0")
	}
	if release != nil {
		release()
	}

	stats, _ := rl.Stats("192.168.1.100")
	if stats.Generation != 1 {
		t.Errorf("expected generation 0 to be normalized to 1, got %d", stats.Generation)
	}

	// Wait for interval
	time.Sleep(config.Gen1.MinInterval + 10*time.Millisecond)

	// Test generation > 2 normalization in TryAcquire (treated as Gen2)
	release, ok = rl.TryAcquire("192.168.1.101", 99)
	if !ok {
		t.Error("expected TryAcquire to succeed with generation 99")
	}
	if release != nil {
		release()
	}

	stats, _ = rl.Stats("192.168.1.101")
	if stats.Generation != 2 {
		t.Errorf("expected generation 99 to be normalized to 2, got %d", stats.Generation)
	}
}

func TestDeviceRateLimiter_Acquire_DeviceSemaphoreCancelledAfterGlobal(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()
	config.Gen2.MaxConcurrent = 1 // Only 1 per device
	config.Gen2.MinInterval = 1 * time.Millisecond
	config.Global.MaxConcurrent = 10

	rl := NewWithConfig(config)
	ctx := context.Background()

	// First acquire takes the device slot
	release1, err := rl.Acquire(ctx, "192.168.1.100", 2)
	if err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}

	// Second acquire with very short timeout
	// Should acquire global semaphore but fail on device semaphore
	shortCtx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err = rl.Acquire(shortCtx, "192.168.1.100", 2)
	if err == nil {
		t.Error("expected error when device semaphore times out")
	}
	// Should be context deadline or canceled
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Errorf("expected context error, got: %v", err)
	}

	release1()
}

func TestDeviceRateLimiter_SetGeneration_SameGeneration(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()
	rl := NewWithConfig(config)
	ctx := context.Background()

	// Create device with Gen2
	release, err := rl.Acquire(ctx, "192.168.1.100", 2)
	if err != nil {
		t.Fatalf("acquire failed: %v", err)
	}
	release()

	// Set to same generation - should not recreate semaphore
	rl.SetGeneration("192.168.1.100", 2)

	stats, _ := rl.Stats("192.168.1.100")
	if stats.Generation != 2 {
		t.Errorf("expected generation 2, got %d", stats.Generation)
	}
}

func TestCircuitBreaker_HalfOpen_AllowsRequests(t *testing.T) {
	t.Parallel()
	cb := NewCircuitBreaker(1, 2, 10*time.Millisecond)

	// Open the circuit
	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Fatalf("expected state Open, got %v", cb.State())
	}

	// Wait for transition
	time.Sleep(15 * time.Millisecond)

	// First Allow() transitions to half-open and returns true
	if !cb.Allow() {
		t.Error("expected Allow() to return true transitioning to half-open")
	}
	if cb.State() != StateHalfOpen {
		t.Errorf("expected state HalfOpen, got %v", cb.State())
	}

	// Subsequent Allow() calls in half-open should still return true
	if !cb.Allow() {
		t.Error("expected Allow() to return true in half-open state")
	}
}

func TestDeviceRateLimiter_Reset_ExistingDevice(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()
	config.Gen2.CircuitThreshold = 1
	config.Global.CircuitOpenDuration = 60 * time.Second

	rl := NewWithConfig(config)
	ctx := context.Background()

	// Create device and open its circuit
	release, err := rl.Acquire(ctx, "192.168.1.100", 2)
	if err != nil {
		t.Fatalf("acquire failed: %v", err)
	}
	rl.RecordFailure("192.168.1.100")
	release()

	// Verify circuit is open
	if !rl.IsCircuitOpen("192.168.1.100") {
		t.Fatal("expected circuit to be open")
	}

	// Reset the existing device - this covers the exists=true branch
	rl.Reset("192.168.1.100")

	// Circuit should now be closed
	if rl.IsCircuitOpen("192.168.1.100") {
		t.Error("expected circuit to be closed after Reset")
	}
}

func TestDeviceRateLimiter_GetOrCreateState_RaceCondition(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()
	config.Gen2.MinInterval = 1 * time.Millisecond
	config.Global.MaxConcurrent = 200

	rl := NewWithConfig(config)

	// Use a barrier to maximize chance of hitting the race condition
	// where another goroutine creates the state between RUnlock and Lock
	var ready sync.WaitGroup
	ready.Add(100)
	start := make(chan struct{})

	var wg sync.WaitGroup
	for range 100 {
		wg.Go(func() {
			ready.Done()
			<-start // All goroutines start at the same time

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			release, err := rl.Acquire(ctx, "race-device", 2)
			if err != nil {
				return // timeout is fine
			}
			release()
		})
	}

	ready.Wait() // Wait for all goroutines to be ready
	close(start) // Release them all at once
	wg.Wait()

	// Verify only one state exists
	stats := rl.AllStats()
	count := 0
	for _, s := range stats {
		if s.Address == "race-device" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 device state, got %d", count)
	}
}

func TestCircuitBreaker_Allow_InvalidState(t *testing.T) {
	t.Parallel()
	// Test the default case in Allow() by setting an invalid state
	// This is defensive code that handles corrupted state
	cb := NewCircuitBreaker(3, 2, 60*time.Second)

	// Directly set an invalid state value (same package has access)
	cb.mu.Lock()
	cb.state = State(99) // Invalid state
	cb.mu.Unlock()

	// Allow should return false for invalid state (default case)
	if cb.Allow() {
		t.Error("expected Allow() to return false for invalid state")
	}
}

func TestDeviceRateLimiter_GetOrCreateState_DoubleCheck(t *testing.T) {
	t.Parallel()
	// Test the double-check locking pattern in getOrCreateState
	// by calling it concurrently on a fresh address

	config := DefaultConfig()
	config.Gen2.MinInterval = 0 // No interval delay
	config.Global.MaxConcurrent = 1000

	// Run multiple iterations to increase chance of hitting the race
	for iteration := range 10 {
		rl := NewWithConfig(config)
		address := "double-check-device-" + string(rune('0'+iteration))

		// Launch many goroutines that all call getOrCreateState simultaneously
		const numGoroutines = 50
		var start sync.WaitGroup
		start.Add(numGoroutines)
		var barrier sync.WaitGroup
		barrier.Add(1)
		var done sync.WaitGroup
		done.Add(numGoroutines)

		for range numGoroutines {
			go func() {
				start.Done()
				barrier.Wait()
				// Directly call getOrCreateState (same package)
				_ = rl.getOrCreateState(address, 2)
				done.Done()
			}()
		}

		start.Wait()   // Wait for all goroutines to be ready
		barrier.Done() // Release all goroutines at once
		done.Wait()    // Wait for all to complete

		// Verify exactly one state was created
		count := 0
		for _, s := range rl.AllStats() {
			if s.Address == address {
				count++
			}
		}
		if count != 1 {
			t.Errorf("iteration %d: expected 1 device state for %s, got %d", iteration, address, count)
		}
	}
}

func TestMarkAsPolling(t *testing.T) {
	t.Parallel()

	// Background context should not be polling
	ctx := context.Background()
	if IsPolling(ctx) {
		t.Error("expected background context to not be marked as polling")
	}

	// Marked context should be polling
	pollingCtx := MarkAsPolling(ctx)
	if !IsPolling(pollingCtx) {
		t.Error("expected marked context to be polling")
	}

	// Original context should still not be polling
	if IsPolling(ctx) {
		t.Error("expected original context to remain unmarked")
	}
}

func TestIsPolling_WithTimeout(t *testing.T) {
	t.Parallel()

	// Mark as polling, then add timeout
	ctx := MarkAsPolling(context.Background())
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	// Should still be polling after adding timeout
	if !IsPolling(ctx) {
		t.Error("expected polling flag to be preserved through WithTimeout")
	}
}

func TestIsPolling_WithCancel(t *testing.T) {
	t.Parallel()

	// Mark as polling, then add cancel
	ctx := MarkAsPolling(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Should still be polling after adding cancel
	if !IsPolling(ctx) {
		t.Error("expected polling flag to be preserved through WithCancel")
	}
}

func TestIsPolling_NestedValues(t *testing.T) {
	t.Parallel()

	// Add another value, then mark as polling
	type otherKey struct{}
	ctx := context.WithValue(context.Background(), otherKey{}, "test")
	ctx = MarkAsPolling(ctx)

	// Both values should be accessible
	if !IsPolling(ctx) {
		t.Error("expected context to be polling")
	}
	if ctx.Value(otherKey{}) != "test" {
		t.Error("expected other value to be preserved")
	}
}

func TestIsPolling_WrongType(t *testing.T) {
	t.Parallel()

	// Simulate a context where pollingContextKey exists but with wrong type
	type pollingContextKey struct{}
	ctx := context.WithValue(context.Background(), pollingContextKey{}, "not-a-bool")

	// IsPolling should return false because the type assertion fails
	if IsPolling(ctx) {
		t.Error("expected context with wrong value type to not be polling")
	}
}
