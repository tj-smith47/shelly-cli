// Package ratelimit provides per-device request throttling and circuit breaker
// functionality for Shelly IoT devices.
//
// Shelly devices have hardware limitations:
//   - Gen1 (ESP8266): MAX 2 concurrent HTTP connections
//   - Gen2 (ESP32): MAX 5 concurrent HTTP transactions
//
// This package prevents overloading devices by:
//   - Limiting concurrent requests per device
//   - Enforcing minimum intervals between requests
//   - Circuit breaking unresponsive devices
package ratelimit

import "time"

// Config holds rate limiting configuration.
type Config struct {
	Gen1   GenerationConfig
	Gen2   GenerationConfig
	Global GlobalConfig
}

// GenerationConfig holds settings for a specific device generation.
type GenerationConfig struct {
	// MinInterval is the minimum time between requests to the same device.
	// Gen1 devices need more breathing room than Gen2.
	MinInterval time.Duration

	// MaxConcurrent is the maximum in-flight requests per device.
	// Gen1 can only handle 2 connections total, so we use 1 to leave headroom.
	// Gen2 can handle 5, so we use 3 for safety.
	MaxConcurrent int

	// CircuitThreshold is the number of consecutive failures before
	// the circuit breaker opens for a device.
	CircuitThreshold int
}

// GlobalConfig holds global rate limiting settings.
type GlobalConfig struct {
	// MaxConcurrent is the total concurrent requests across all devices.
	// Prevents network saturation when polling many devices.
	MaxConcurrent int

	// CircuitOpenDuration is how long a circuit stays open before
	// transitioning to half-open to test if the device recovered.
	CircuitOpenDuration time.Duration

	// CircuitSuccessThreshold is the number of consecutive successes
	// needed in half-open state to close the circuit.
	CircuitSuccessThreshold int
}

// DefaultConfig returns sensible defaults based on Shelly hardware constraints.
//
// Gen1 (ESP8266) limits:
//   - 2 concurrent HTTP connections max
//   - 4 total sockets
//   - Easily overwhelmed, needs conservative limits
//
// Gen2 (ESP32) limits:
//   - 5 concurrent HTTP transactions max
//   - 10-second timeout per transaction
//   - More resilient but still resource-constrained
func DefaultConfig() Config {
	return Config{
		Gen1: GenerationConfig{
			MinInterval:      2 * time.Second, // Gen1 needs breathing room
			MaxConcurrent:    1,               // Leave 1 connection for safety
			CircuitThreshold: 3,               // Open circuit after 3 failures
		},
		Gen2: GenerationConfig{
			MinInterval:      500 * time.Millisecond, // Gen2 handles faster polling
			MaxConcurrent:    3,                      // Leave 2 connections for safety
			CircuitThreshold: 5,                      // Gen2 is more resilient
		},
		Global: GlobalConfig{
			MaxConcurrent:           5,                // Total across all devices
			CircuitOpenDuration:     60 * time.Second, // Standard backoff
			CircuitSuccessThreshold: 2,                // Successes to close circuit
		},
	}
}

// Option configures the rate limiter.
type Option func(*Config)

// WithGen1MinInterval sets the minimum interval between requests for Gen1 devices.
func WithGen1MinInterval(d time.Duration) Option {
	return func(c *Config) {
		c.Gen1.MinInterval = d
	}
}

// WithGen2MinInterval sets the minimum interval between requests for Gen2 devices.
func WithGen2MinInterval(d time.Duration) Option {
	return func(c *Config) {
		c.Gen2.MinInterval = d
	}
}

// WithGen1MaxConcurrent sets the max concurrent requests per Gen1 device.
func WithGen1MaxConcurrent(n int) Option {
	return func(c *Config) {
		c.Gen1.MaxConcurrent = n
	}
}

// WithGen2MaxConcurrent sets the max concurrent requests per Gen2 device.
func WithGen2MaxConcurrent(n int) Option {
	return func(c *Config) {
		c.Gen2.MaxConcurrent = n
	}
}

// WithGlobalMaxConcurrent sets the total concurrent requests across all devices.
func WithGlobalMaxConcurrent(n int) Option {
	return func(c *Config) {
		c.Global.MaxConcurrent = n
	}
}

// WithCircuitOpenDuration sets how long circuits stay open.
func WithCircuitOpenDuration(d time.Duration) Option {
	return func(c *Config) {
		c.Global.CircuitOpenDuration = d
	}
}

// generationConfig returns the config for a specific generation.
func (c *Config) generationConfig(gen int) GenerationConfig {
	if gen == 1 {
		return c.Gen1
	}
	return c.Gen2
}
