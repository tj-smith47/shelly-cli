// Package device provides device-level operations for Shelly devices.
package device

import (
	"context"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

// Info holds extended device information.
type Info struct {
	ID         string
	MAC        string
	Model      string
	Generation int
	Firmware   string
	App        string
	AuthEn     bool
}

// Status holds device status information.
type Status struct {
	Info   *Info
	Status map[string]any
}

// GetInfo returns information about the device.
func (s *Service) GetInfo(ctx context.Context, identifier string) (*Info, error) {
	var result *Info
	err := s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		info := conn.Info()
		result = &Info{
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

// GetStatus returns the full status of the device.
func (s *Service) GetStatus(ctx context.Context, identifier string) (*Status, error) {
	var result *Status
	err := s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		info := conn.Info()
		status, err := conn.GetStatus(ctx)
		if err != nil {
			return err
		}

		result = &Status{
			Info: &Info{
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

// Ping checks if the device is reachable by attempting to connect.
// Uses GetInfoAuto to support both Gen1 and Gen2 devices.
func (s *Service) Ping(ctx context.Context, identifier string) (*Info, error) {
	return s.GetInfoAuto(ctx, identifier)
}

// GetInfoAuto returns device info, auto-detecting generation (Gen1 vs Gen2).
// If generation is known from config, it tries that generation first for efficiency.
// Otherwise it tries Gen2 first (more common), then falls back to Gen1 if Gen2 fails.
// Use this for TUI/cache where we need to handle all device types.
func (s *Service) GetInfoAuto(ctx context.Context, identifier string) (*Info, error) {
	// First resolve to check if we have a stored generation
	// Error is intentionally ignored - if resolution fails, we try Gen2 first
	device, err := s.parent.ResolveWithGeneration(ctx, identifier)

	// If we know it's Gen1, try Gen1 first to avoid wasting time on Gen2
	if err == nil && device.Generation == 1 {
		gen1Result, gen1Err := s.GetInfoGen1(ctx, identifier)
		if gen1Err == nil {
			return gen1Result, nil
		}
		// Gen1 failed unexpectedly, try Gen2 as fallback
		result, err := s.GetInfo(ctx, identifier)
		if err == nil {
			return result, nil
		}
		// Both failed, return Gen1 error since we knew it was Gen1
		return nil, gen1Err
	}

	// Gen2+ or unknown: Try Gen2 first (more common)
	result, err := s.GetInfo(ctx, identifier)
	if err == nil {
		return result, nil
	}

	// Gen2 failed, try Gen1
	gen1Result, gen1Err := s.GetInfoGen1(ctx, identifier)
	if gen1Err == nil {
		return gen1Result, nil
	}

	// Both failed, return the original Gen2 error (more informative)
	return nil, err
}

// GetInfoGen1 returns information about a Gen1 device.
func (s *Service) GetInfoGen1(ctx context.Context, identifier string) (*Info, error) {
	var result *Info
	err := s.parent.WithGen1Connection(ctx, identifier, func(conn *client.Gen1Client) error {
		info := conn.Info()
		result = &Info{
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
