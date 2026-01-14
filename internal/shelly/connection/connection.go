// Package connection provides device connection management for Shelly devices.
// It handles Gen1 and Gen2 device connections with rate limiting and automatic
// IP remapping when devices change addresses.
package connection

import (
	"context"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/ratelimit"
	"github.com/tj-smith47/shelly-cli/internal/tui/debug"
)

// Provider abstracts connection execution for services.
// This is the minimal interface that most services need.
type Provider interface {
	// WithConnection executes a function with a Gen2+ device connection.
	WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error
	// WithGen1Connection executes a function with a Gen1 device connection.
	WithGen1Connection(ctx context.Context, identifier string, fn func(*client.Gen1Client) error) error
}

// Discoverer provides MAC-based IP discovery for connection recovery.
// When a device's IP changes (e.g., DHCP reassignment), this interface
// enables automatic discovery via mDNS.
type Discoverer interface {
	DiscoverByMAC(ctx context.Context, mac string) (string, error)
}

// Resolver resolves device identifiers to model.Device with generation info.
type Resolver interface {
	ResolveWithGeneration(ctx context.Context, identifier string) (model.Device, error)
}

// Manager handles device connections with rate limiting and IP remapping.
// It provides connection management for both Gen1 and Gen2 Shelly devices.
type Manager struct {
	resolver    Resolver
	discoverer  Discoverer
	rateLimiter *ratelimit.DeviceRateLimiter
}

// Option configures a Manager.
type Option func(*Manager)

// WithRateLimiter configures the manager to use rate limiting.
// If not provided, no rate limiting is applied.
func WithRateLimiter(rl *ratelimit.DeviceRateLimiter) Option {
	return func(m *Manager) {
		m.rateLimiter = rl
	}
}

// NewManager creates a new connection Manager.
// The resolver is required for device resolution.
// The discoverer is optional; if nil, IP remapping is disabled.
func NewManager(resolver Resolver, discoverer Discoverer, opts ...Option) *Manager {
	m := &Manager{
		resolver:   resolver,
		discoverer: discoverer,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// Resolver returns the manager's resolver.
func (m *Manager) Resolver() Resolver {
	return m.resolver
}

// RateLimiter returns the manager's rate limiter, if configured.
func (m *Manager) RateLimiter() *ratelimit.DeviceRateLimiter {
	return m.rateLimiter
}

// WithConnection executes a function with a Gen2+ device connection.
// Rate limiting is automatically applied if configured.
func (m *Manager) WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error {
	// Resolve device to get address and generation for rate limiting
	dev, err := m.resolver.ResolveWithGeneration(ctx, identifier)
	if err != nil {
		return err
	}

	// No rate limiter configured - execute directly
	if m.rateLimiter == nil {
		return m.ExecuteGen2(ctx, dev, fn)
	}

	// Acquire rate limiter slot
	release, err := m.rateLimiter.Acquire(ctx, dev.Address, dev.Generation)
	if err != nil {
		return err
	}
	defer release()

	// Execute the operation
	err = m.ExecuteGen2(ctx, dev, fn)

	// Record success/failure for circuit breaker
	m.recordCircuitResult(ctx, dev, err)

	return err
}

// WithGen1Connection executes a function with a Gen1 device connection.
// Rate limiting is automatically applied if configured.
func (m *Manager) WithGen1Connection(ctx context.Context, identifier string, fn func(*client.Gen1Client) error) error {
	// Resolve device to get address for rate limiting
	dev, err := m.resolver.ResolveWithGeneration(ctx, identifier)
	if err != nil {
		return err
	}

	// No rate limiter configured - execute directly
	if m.rateLimiter == nil {
		return m.ExecuteGen1(ctx, dev, fn)
	}

	// Gen1 devices always use generation=1 for rate limiting
	release, err := m.rateLimiter.Acquire(ctx, dev.Address, 1)
	if err != nil {
		return err
	}
	defer release()

	// Execute the operation
	err = m.ExecuteGen1(ctx, dev, fn)

	// Record success/failure for circuit breaker
	m.recordCircuitResult(ctx, dev, err)

	return err
}

// IsGen1Device checks if a device is Gen1.
func (m *Manager) IsGen1Device(ctx context.Context, identifier string) (bool, model.Device, error) {
	dev, err := m.resolver.ResolveWithGeneration(ctx, identifier)
	if err != nil {
		return false, model.Device{}, err
	}
	return dev.Generation == 1, dev, nil
}

// recordCircuitResult records success/failure for the circuit breaker.
// Only counts actual connectivity failures, not expected API responses.
// Skips failure recording for polling requests (BUG-015) - polling failures
// shouldn't block user-initiated actions like toggles.
func (m *Manager) recordCircuitResult(ctx context.Context, dev model.Device, err error) {
	if err == nil {
		m.rateLimiter.RecordSuccess(dev.Address)
		return
	}

	if !ratelimit.IsConnectivityFailure(err) {
		debug.TraceEvent("circuit: Ignoring non-connectivity error for %s (%s): %v", dev.Name, dev.Address, err)
		m.rateLimiter.RecordSuccess(dev.Address)
		return
	}

	// Connectivity failure - skip for polling, record for user actions
	if ratelimit.IsPolling(ctx) {
		debug.TraceEvent("circuit: Skipping RecordFailure for polling %s (%s): %v", dev.Name, dev.Address, err)
		return
	}

	debug.TraceEvent("circuit: RecordFailure for %s (%s): %v", dev.Name, dev.Address, err)
	m.rateLimiter.RecordFailure(dev.Address)
}
