// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// DeviceClient provides a unified interface for both Gen1 and Gen2 device connections.
// Use IsGen1() to determine which generation, then Gen1() or Gen2() to access
// generation-specific APIs.
type DeviceClient struct {
	gen1 *client.Gen1Client
	gen2 *client.Client
}

// IsGen1 returns true if this is a Gen1 device connection.
func (c *DeviceClient) IsGen1() bool {
	return c.gen1 != nil
}

// IsGen2 returns true if this is a Gen2+ device connection.
func (c *DeviceClient) IsGen2() bool {
	return c.gen2 != nil
}

// Generation returns the device generation (1 for Gen1, 2+ for Gen2).
func (c *DeviceClient) Generation() int {
	if c.gen1 != nil {
		return c.gen1.Info().Generation
	}
	if c.gen2 != nil {
		return c.gen2.Info().Generation
	}
	return 0
}

// Gen1 returns the Gen1 client. Panics if this is not a Gen1 connection.
// Check IsGen1() first.
func (c *DeviceClient) Gen1() *client.Gen1Client {
	if c.gen1 == nil {
		panic("DeviceClient: not a Gen1 connection, check IsGen1() first")
	}
	return c.gen1
}

// Gen2 returns the Gen2+ client. Panics if this is not a Gen2 connection.
// Check IsGen2() first.
func (c *DeviceClient) Gen2() *client.Client {
	if c.gen2 == nil {
		panic("DeviceClient: not a Gen2 connection, check IsGen2() first")
	}
	return c.gen2
}

// Info returns the device information. Works for both generations.
func (c *DeviceClient) Info() *client.DeviceInfo {
	if c.gen1 != nil {
		return c.gen1.Info()
	}
	if c.gen2 != nil {
		return c.gen2.Info()
	}
	return nil
}

// Close closes the device connection. Works for both generations.
func (c *DeviceClient) Close() error {
	if c.gen1 != nil {
		return c.gen1.Close()
	}
	if c.gen2 != nil {
		return c.gen2.Close()
	}
	return nil
}

// WithDevice executes a function with a unified device connection.
// The DeviceClient auto-detects the device generation and provides
// access to both Gen1 and Gen2 APIs through a single interface.
//
// Example usage:
//
//	err := svc.WithDevice(ctx, "living-room", func(dev *shelly.DeviceClient) error {
//	    if dev.IsGen1() {
//	        relay, _ := dev.Gen1().Relay(0)
//	        return relay.TurnOn(ctx)
//	    }
//	    return dev.Gen2().Switch(0).On(ctx)
//	})
func (s *Service) WithDevice(ctx context.Context, identifier string, fn func(*DeviceClient) error) error {
	isGen1, _, err := s.IsGen1Device(ctx, identifier)
	if err != nil {
		return err
	}

	if isGen1 {
		return s.WithGen1Connection(ctx, identifier, func(conn *client.Gen1Client) error {
			return fn(&DeviceClient{gen1: conn})
		})
	}

	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return fn(&DeviceClient{gen2: conn})
	})
}

// WithDevices executes a function for multiple devices concurrently.
// Each device gets its own DeviceClient with auto-detected generation.
// Errors are collected and returned as a combined error.
func (s *Service) WithDevices(ctx context.Context, devices []string, concurrency int, fn func(device string, dev *DeviceClient) error) error {
	if len(devices) == 0 {
		return nil
	}

	// Use the existing batch infrastructure
	type result struct {
		device string
		err    error
	}

	results := make(chan result, len(devices))
	sem := make(chan struct{}, concurrency)

	for _, device := range devices {
		sem <- struct{}{}
		go func() {
			defer func() { <-sem }()
			err := s.WithDevice(ctx, device, func(dev *DeviceClient) error {
				return fn(device, dev)
			})
			results <- result{device: device, err: err}
		}()
	}

	// Collect results
	var errs []error
	for range devices {
		r := <-results
		if r.err != nil {
			errs = append(errs, r.err)
			iostreams.DebugErr("WithDevices: "+r.device, r.err)
		}
	}

	if len(errs) > 0 {
		return errs[0] // Return first error; could be improved with multi-error
	}
	return nil
}
