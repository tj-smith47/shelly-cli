// Package component provides component-level operations for Shelly devices.
package component

import (
	"context"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// DeviceClient provides a unified interface for both Gen1 and Gen2 device connections.
type DeviceClient interface {
	IsGen1() bool
	Gen1() *client.Gen1Client
	Gen2() *client.Client
}

// ConnectionProvider provides device connection capabilities.
// This interface is implemented by shelly.Service.
type ConnectionProvider interface {
	// WithConnection executes a function with a Gen2+ device connection.
	WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error
	// WithGen1Connection executes a function with a Gen1 device connection.
	WithGen1Connection(ctx context.Context, identifier string, fn func(*client.Gen1Client) error) error
	// WithDevice executes a function with a unified device client.
	WithDevice(ctx context.Context, identifier string, fn func(DeviceClient) error) error
	// IsGen1Device checks if a device is Gen1.
	IsGen1Device(ctx context.Context, identifier string) (bool, model.Device, error)
}

// Service provides component-level operations for Shelly devices.
type Service struct {
	parent ConnectionProvider
}

// New creates a new component service.
func New(parent ConnectionProvider) *Service {
	return &Service{parent: parent}
}

// withGenAwareAction executes gen1Fn for Gen1 devices, gen2Fn for Gen2+ devices.
// This centralizes the generation detection and routing logic.
func (s *Service) withGenAwareAction(
	ctx context.Context,
	identifier string,
	gen1Fn func(*client.Gen1Client) error,
	gen2Fn func(*client.Client) error,
) error {
	isGen1, _, err := s.parent.IsGen1Device(ctx, identifier)
	if err != nil {
		return err
	}

	if isGen1 {
		return s.parent.WithGen1Connection(ctx, identifier, gen1Fn)
	}
	return s.parent.WithConnection(ctx, identifier, gen2Fn)
}
