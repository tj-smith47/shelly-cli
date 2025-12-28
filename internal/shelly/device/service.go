// Package device provides device-level operations for Shelly devices.
package device

import (
	"context"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly/firmware"
)

// ConnectionProvider provides device connection capabilities.
// This interface is implemented by shelly.Service.
type ConnectionProvider interface {
	// WithConnection executes a function with a Gen2+ device connection.
	WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error
	// WithGen1Connection executes a function with a Gen1 device connection.
	WithGen1Connection(ctx context.Context, identifier string, fn func(*client.Gen1Client) error) error
	// ResolveWithGeneration resolves a device identifier with generation auto-detection.
	ResolveWithGeneration(ctx context.Context, identifier string) (model.Device, error)
	// GetCachedFirmware returns cached firmware info for a device.
	GetCachedFirmware(ctx context.Context, deviceName string, maxAge time.Duration) *firmware.CacheEntry
}

// Service provides device-level operations for Shelly devices.
type Service struct {
	parent ConnectionProvider
}

// New creates a new device service.
func New(parent ConnectionProvider) *Service {
	return &Service{parent: parent}
}
