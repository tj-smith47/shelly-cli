// Package monitoring provides monitoring operations for Shelly devices.
// This includes energy metering, snapshots, metrics collection, and event subscriptions.
package monitoring

import (
	"context"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// ShellyConnector provides connectivity to Shelly devices.
// This interface is implemented by *shelly.Service.
type ShellyConnector interface {
	// WithConnection executes a function with a Gen2+ device connection.
	WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error

	// WithGen1Connection executes a function with a Gen1 device connection.
	WithGen1Connection(ctx context.Context, identifier string, fn func(*client.Gen1Client) error) error

	// Resolve resolves a device identifier to its full device info.
	Resolve(identifier string) (model.Device, error)

	// ResolveWithGeneration resolves a device and detects its generation.
	ResolveWithGeneration(ctx context.Context, identifier string) (*ResolvedDevice, error)

	// DeviceInfo returns device information.
	DeviceInfo(ctx context.Context, identifier string) (*DeviceInfo, error)

	// DeviceStatus returns device status.
	DeviceStatus(ctx context.Context, identifier string) (*DeviceStatusResult, error)
}

// ResolvedDevice holds resolved device info with generation.
type ResolvedDevice struct {
	Device     model.Device
	Generation int
}

// DeviceInfo holds device information returned by the connector.
type DeviceInfo struct {
	ID         string
	MAC        string
	Model      string
	Generation int
	Firmware   string
	App        string
	AuthEn     bool
}

// DeviceStatusResult holds device status returned by the connector.
type DeviceStatusResult struct {
	Status map[string]any
}

// Options configures real-time monitoring behavior.
type Options struct {
	Interval      time.Duration // Refresh interval for polling
	Count         int           // Number of updates (0 = unlimited)
	IncludePower  bool          // Include power meter data
	IncludeEnergy bool          // Include energy meter data
}

// Callback is called with each status update during monitoring.
type Callback func(model.MonitoringSnapshot) error

// EventHandler handles device events received via WebSocket.
type EventHandler func(model.DeviceEvent) error

// DeviceSnapshot holds the latest status for a device in multi-device monitoring.
type DeviceSnapshot struct {
	Device   string
	Address  string
	Info     *DeviceInfo
	Snapshot *model.MonitoringSnapshot
	Error    error
}

// Service provides monitoring operations for Shelly devices.
type Service struct {
	connector ShellyConnector
}

// NewService creates a new monitoring service.
func NewService(connector ShellyConnector) *Service {
	return &Service{connector: connector}
}
