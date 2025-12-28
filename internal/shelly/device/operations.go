// Package device provides device-level operations for Shelly devices.
package device

import (
	"context"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

// Reboot reboots the device.
func (s *Service) Reboot(ctx context.Context, identifier string, delayMS int) error {
	return s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.Reboot(ctx, delayMS)
	})
}

// FactoryReset performs a factory reset on the device.
func (s *Service) FactoryReset(ctx context.Context, identifier string) error {
	return s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.FactoryReset(ctx)
	})
}
