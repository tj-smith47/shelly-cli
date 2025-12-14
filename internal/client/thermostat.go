// Package client provides device communication abstraction over shelly-go SDK.
package client

import (
	"context"

	"github.com/tj-smith47/shelly-go/gen2/components"
	"github.com/tj-smith47/shelly-go/rpc"
)

// ThermostatComponent provides access to thermostat component operations.
type ThermostatComponent struct {
	th  *components.Thermostat
	rpc *rpc.Client
	id  int
}

// GetStatus returns the current status of the thermostat.
func (t *ThermostatComponent) GetStatus(ctx context.Context) (*components.ThermostatStatus, error) {
	return t.th.GetStatus(ctx)
}

// GetConfig returns the configuration of the thermostat.
func (t *ThermostatComponent) GetConfig(ctx context.Context) (*components.ThermostatConfig, error) {
	return t.th.GetConfig(ctx)
}

// SetConfig updates the thermostat configuration.
func (t *ThermostatComponent) SetConfig(ctx context.Context, config *components.ThermostatConfig) error {
	return t.th.SetConfig(ctx, config)
}

// SetTarget sets the target temperature.
func (t *ThermostatComponent) SetTarget(ctx context.Context, targetC float64) error {
	return t.th.SetTarget(ctx, targetC)
}

// Enable enables or disables the thermostat.
func (t *ThermostatComponent) Enable(ctx context.Context, enable bool) error {
	return t.th.Enable(ctx, enable)
}

// SetMode sets the thermostat operating mode.
func (t *ThermostatComponent) SetMode(ctx context.Context, mode string) error {
	return t.th.SetMode(ctx, mode)
}

// Boost activates boost mode with optional duration.
func (t *ThermostatComponent) Boost(ctx context.Context, durationSec int) error {
	return t.th.Boost(ctx, durationSec)
}

// CancelBoost cancels an active boost mode.
func (t *ThermostatComponent) CancelBoost(ctx context.Context) error {
	return t.th.CancelBoost(ctx)
}

// Override activates temperature override mode.
func (t *ThermostatComponent) Override(ctx context.Context, targetC float64, durationSec int) error {
	return t.th.Override(ctx, targetC, durationSec)
}

// CancelOverride cancels an active override mode.
func (t *ThermostatComponent) CancelOverride(ctx context.Context) error {
	return t.th.CancelOverride(ctx)
}

// Calibrate initiates valve calibration.
func (t *ThermostatComponent) Calibrate(ctx context.Context) error {
	return t.th.Calibrate(ctx)
}

// ID returns the thermostat component ID.
func (t *ThermostatComponent) ID() int {
	return t.id
}
