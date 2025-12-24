// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/ratelimit"
)

// DefaultTimeout is the default timeout for device operations.
const DefaultTimeout = 10 * time.Second

// Service provides high-level operations on Shelly devices.
type Service struct {
	resolver    DeviceResolver
	rateLimiter *ratelimit.DeviceRateLimiter
}

// DeviceResolver resolves device identifiers to device configurations.
type DeviceResolver interface {
	Resolve(identifier string) (model.Device, error)
}

// GenerationAwareResolver extends DeviceResolver with generation detection.
type GenerationAwareResolver interface {
	DeviceResolver
	ResolveWithGeneration(ctx context.Context, identifier string) (model.Device, error)
}

// ServiceOption configures a Service.
type ServiceOption func(*Service)

// WithRateLimiter configures the service to use rate limiting.
// If not provided, no rate limiting is applied (backward compatible).
func WithRateLimiter(rl *ratelimit.DeviceRateLimiter) ServiceOption {
	return func(s *Service) {
		s.rateLimiter = rl
	}
}

// WithDefaultRateLimiter configures the service with default rate limiting.
// This is recommended for TUI usage to prevent overloading Shelly devices.
func WithDefaultRateLimiter() ServiceOption {
	return func(s *Service) {
		s.rateLimiter = ratelimit.New()
	}
}

// New creates a new Shelly service with optional configuration.
func New(resolver DeviceResolver, opts ...ServiceOption) *Service {
	svc := &Service{resolver: resolver}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

// Connect establishes a connection to a device by identifier (name or address).
func (s *Service) Connect(ctx context.Context, identifier string) (*client.Client, error) {
	device, err := s.resolver.Resolve(identifier)
	if err != nil {
		return nil, err
	}

	return client.Connect(ctx, device)
}

// WithConnection executes a function with a device connection, handling cleanup.
func (s *Service) WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error {
	conn, err := s.Connect(ctx, identifier)
	if err != nil {
		return err
	}
	defer iostreams.CloseWithDebug("closing device connection", conn)

	return fn(conn)
}

// RawRPC sends a raw RPC command to a device and returns the response.
func (s *Service) RawRPC(ctx context.Context, identifier, method string, params map[string]any) (any, error) {
	var result any
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		res, err := conn.Call(ctx, method, params)
		if err != nil {
			return err
		}
		result = res
		return nil
	})
	return result, err
}

// RawGen1Call sends a raw REST API call to a Gen1 device and returns the response as bytes.
func (s *Service) RawGen1Call(ctx context.Context, identifier, path string) ([]byte, error) {
	var result []byte
	err := s.WithGen1Connection(ctx, identifier, func(conn *client.Gen1Client) error {
		res, err := conn.Call(ctx, path)
		if err != nil {
			return err
		}
		result = res
		return nil
	})
	return result, err
}

// ResolveWithGeneration resolves a device identifier with generation auto-detection.
// If the resolver implements GenerationAwareResolver, it uses that; otherwise falls back to basic resolution.
func (s *Service) ResolveWithGeneration(ctx context.Context, identifier string) (model.Device, error) {
	if gar, ok := s.resolver.(GenerationAwareResolver); ok {
		return gar.ResolveWithGeneration(ctx, identifier)
	}
	return s.resolver.Resolve(identifier)
}

// ConnectGen1 establishes a connection to a Gen1 device by identifier.
func (s *Service) ConnectGen1(ctx context.Context, identifier string) (*client.Gen1Client, error) {
	device, err := s.ResolveWithGeneration(ctx, identifier)
	if err != nil {
		return nil, err
	}

	return client.ConnectGen1(ctx, device)
}

// WithGen1Connection executes a function with a Gen1 device connection, handling cleanup.
func (s *Service) WithGen1Connection(ctx context.Context, identifier string, fn func(*client.Gen1Client) error) error {
	conn, err := s.ConnectGen1(ctx, identifier)
	if err != nil {
		return err
	}
	defer iostreams.CloseWithDebug("closing gen1 device connection", conn)

	return fn(conn)
}

// IsGen1Device checks if a device is Gen1.
func (s *Service) IsGen1Device(ctx context.Context, identifier string) (bool, model.Device, error) {
	device, err := s.ResolveWithGeneration(ctx, identifier)
	if err != nil {
		return false, model.Device{}, err
	}
	return device.Generation == 1, device, nil
}

// withGenAwareAction executes gen1Fn for Gen1 devices, gen2Fn for Gen2+ devices.
// This centralizes the generation detection and routing logic.
func (s *Service) withGenAwareAction(
	ctx context.Context,
	identifier string,
	gen1Fn func(*client.Gen1Client) error,
	gen2Fn func(*client.Client) error,
) error {
	isGen1, _, err := s.IsGen1Device(ctx, identifier)
	if err != nil {
		return err
	}

	if isGen1 {
		return s.WithGen1Connection(ctx, identifier, gen1Fn)
	}
	return s.WithConnection(ctx, identifier, gen2Fn)
}

// WithRateLimitedCall wraps a device operation with rate limiting.
// If no rate limiter is configured, the operation executes directly.
//
// The generation parameter controls rate limiting behavior:
//   - 1: Gen1 limits (1 concurrent, 2s interval - ESP8266 constraints)
//   - 2: Gen2 limits (3 concurrent, 500ms interval - ESP32 constraints)
//   - 0: Unknown, treated as Gen1 for safety
//
// Returns ErrCircuitOpen if the device's circuit breaker is open.
func (s *Service) WithRateLimitedCall(ctx context.Context, address string, generation int, fn func() error) error {
	// No rate limiter configured - execute directly (backward compatible)
	if s.rateLimiter == nil {
		return fn()
	}

	// Acquire rate limiter slot
	release, err := s.rateLimiter.Acquire(ctx, address, generation)
	if err != nil {
		return err
	}
	defer release()

	// Execute the operation
	err = fn()

	// Record success/failure for circuit breaker
	if err != nil {
		s.rateLimiter.RecordFailure(address)
	} else {
		s.rateLimiter.RecordSuccess(address)
	}

	return err
}

// RateLimiter returns the service's rate limiter, if configured.
// Returns nil if no rate limiting is enabled.
func (s *Service) RateLimiter() *ratelimit.DeviceRateLimiter {
	return s.rateLimiter
}

// SetDeviceGeneration updates the rate limiter's generation info for a device.
// This is useful after auto-detection to optimize rate limiting.
// No-op if rate limiting is not enabled.
func (s *Service) SetDeviceGeneration(address string, generation int) {
	if s.rateLimiter != nil {
		s.rateLimiter.SetGeneration(address, generation)
	}
}
