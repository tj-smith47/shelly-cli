// Package wireless provides wireless protocol operations for Shelly devices.
// This includes Zigbee, BTHome, LoRa, Matter, and BLE functionality.
package wireless

import (
	"context"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Parent defines the interface to the parent shelly.Service that wireless operations need.
// This avoids import cycles while allowing wireless to use shared service functionality.
type Parent interface {
	// WithConnection executes a function with a device connection, handling cleanup.
	WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error

	// RawRPC sends a raw RPC command to a device and returns the response.
	RawRPC(ctx context.Context, identifier, method string, params map[string]any) (any, error)
}

// Service provides wireless protocol operations for Shelly devices.
type Service struct {
	parent Parent
}

// New creates a new wireless Service with the given parent service.
func New(parent Parent) *Service {
	return &Service{parent: parent}
}

// config returns the global configuration, or nil if not configured.
func getConfig() *config.Config {
	return config.Get()
}
