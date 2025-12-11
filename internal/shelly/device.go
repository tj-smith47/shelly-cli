// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

// DeviceInfo holds extended device information.
type DeviceInfo struct {
	ID         string
	MAC        string
	Model      string
	Generation int
	Firmware   string
	App        string
	AuthEn     bool
}

// DeviceStatus holds device status information.
type DeviceStatus struct {
	Info   *DeviceInfo
	Status map[string]any
}

// DeviceReboot reboots the device.
func (s *Service) DeviceReboot(ctx context.Context, identifier string, delayMS int) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.Reboot(ctx, delayMS)
	})
}

// DeviceFactoryReset performs a factory reset on the device.
func (s *Service) DeviceFactoryReset(ctx context.Context, identifier string) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.FactoryReset(ctx)
	})
}

// DeviceInfo returns information about the device.
func (s *Service) DeviceInfo(ctx context.Context, identifier string) (*DeviceInfo, error) {
	var result *DeviceInfo
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		info := conn.Info()
		result = &DeviceInfo{
			ID:         info.ID,
			MAC:        info.MAC,
			Model:      info.Model,
			Generation: info.Generation,
			Firmware:   info.Firmware,
			App:        info.App,
			AuthEn:     info.AuthEn,
		}
		return nil
	})
	return result, err
}

// DeviceStatus returns the full status of the device.
func (s *Service) DeviceStatus(ctx context.Context, identifier string) (*DeviceStatus, error) {
	var result *DeviceStatus
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		info := conn.Info()
		status, err := conn.GetStatus(ctx)
		if err != nil {
			return err
		}

		result = &DeviceStatus{
			Info: &DeviceInfo{
				ID:         info.ID,
				MAC:        info.MAC,
				Model:      info.Model,
				Generation: info.Generation,
				Firmware:   info.Firmware,
				App:        info.App,
				AuthEn:     info.AuthEn,
			},
			Status: status,
		}
		return nil
	})
	return result, err
}

// DevicePing checks if the device is reachable by attempting to connect.
func (s *Service) DevicePing(ctx context.Context, identifier string) (*DeviceInfo, error) {
	return s.DeviceInfo(ctx, identifier)
}
