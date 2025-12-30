package connection

import (
	"context"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// WithDevice executes a function with a unified device connection.
// The DeviceClient auto-detects the device generation and provides
// access to both Gen1 and Gen2 APIs through a single interface.
//
// Example usage:
//
//	err := mgr.WithDevice(ctx, "living-room", func(dev *DeviceClient) error {
//	    if dev.IsGen1() {
//	        relay, _ := dev.Gen1().Relay(0)
//	        return relay.TurnOn(ctx)
//	    }
//	    return dev.Gen2().Switch(0).On(ctx)
//	})
func (m *Manager) WithDevice(ctx context.Context, identifier string, fn func(*DeviceClient) error) error {
	isGen1, _, err := m.IsGen1Device(ctx, identifier)
	if err != nil {
		return err
	}

	if isGen1 {
		return m.WithGen1Connection(ctx, identifier, func(conn *client.Gen1Client) error {
			return fn(NewGen1Client(conn))
		})
	}

	return m.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return fn(NewGen2Client(conn))
	})
}

// WithDevices executes a function for multiple devices concurrently.
// Each device gets its own DeviceClient with auto-detected generation.
// Errors are collected and returned as a combined error.
func (m *Manager) WithDevices(ctx context.Context, devices []string, concurrency int, fn func(device string, dev *DeviceClient) error) error {
	if len(devices) == 0 {
		return nil
	}

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
			err := m.WithDevice(ctx, device, func(dev *DeviceClient) error {
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
