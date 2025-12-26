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
// Uses DeviceInfoAuto to support both Gen1 and Gen2 devices.
func (s *Service) DevicePing(ctx context.Context, identifier string) (*DeviceInfo, error) {
	return s.DeviceInfoAuto(ctx, identifier)
}

// DeviceInfoAuto returns device info, auto-detecting generation (Gen1 vs Gen2).
// If generation is known from config, it tries that generation first for efficiency.
// Otherwise it tries Gen2 first (more common), then falls back to Gen1 if Gen2 fails.
// Use this for TUI/cache where we need to handle all device types.
func (s *Service) DeviceInfoAuto(ctx context.Context, identifier string) (*DeviceInfo, error) {
	// First resolve to check if we have a stored generation
	// Error is intentionally ignored - if resolution fails, we try Gen2 first
	device, err := s.ResolveWithGeneration(ctx, identifier)

	// If we know it's Gen1, try Gen1 first to avoid wasting time on Gen2
	if err == nil && device.Generation == 1 {
		gen1Result, gen1Err := s.DeviceInfoGen1(ctx, identifier)
		if gen1Err == nil {
			return gen1Result, nil
		}
		// Gen1 failed unexpectedly, try Gen2 as fallback
		result, err := s.DeviceInfo(ctx, identifier)
		if err == nil {
			return result, nil
		}
		// Both failed, return Gen1 error since we knew it was Gen1
		return nil, gen1Err
	}

	// Gen2+ or unknown: Try Gen2 first (more common)
	result, err := s.DeviceInfo(ctx, identifier)
	if err == nil {
		return result, nil
	}

	// Gen2 failed, try Gen1
	gen1Result, gen1Err := s.DeviceInfoGen1(ctx, identifier)
	if gen1Err == nil {
		return gen1Result, nil
	}

	// Both failed, return the original Gen2 error (more informative)
	return nil, err
}

// DeviceInfoGen1 returns information about a Gen1 device.
func (s *Service) DeviceInfoGen1(ctx context.Context, identifier string) (*DeviceInfo, error) {
	var result *DeviceInfo
	err := s.WithGen1Connection(ctx, identifier, func(conn *client.Gen1Client) error {
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
