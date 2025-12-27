// Package automation provides script, schedule, and event automation for Shelly devices.
package automation

import (
	"context"
	"encoding/json"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// ConnectionProvider provides device connection capabilities.
// This interface is implemented by shelly.Service.
type ConnectionProvider interface {
	// WithConnection executes a function with a device connection.
	WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error
	// ResolveWithGeneration resolves a device identifier with generation auto-detection.
	ResolveWithGeneration(ctx context.Context, identifier string) (model.Device, error)
}

// EventStreamProvider extends ConnectionProvider with Gen1 monitoring support for EventStream.
// This interface is implemented by shelly.Service.
type EventStreamProvider interface {
	ConnectionProvider
	// GetGen1StatusJSON returns the Gen1 device status as JSON for event streaming.
	// This is used by EventStream to poll Gen1 devices.
	GetGen1StatusJSON(ctx context.Context, identifier string) (json.RawMessage, error)
}

// Service provides automation operations for Shelly devices.
// It wraps a ConnectionProvider (typically shelly.Service) to provide
// script, schedule, and event streaming functionality.
type Service struct {
	parent ConnectionProvider
}

// New creates a new automation service.
func New(parent ConnectionProvider) *Service {
	return &Service{parent: parent}
}
