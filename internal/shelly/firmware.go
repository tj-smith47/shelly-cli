// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"sync"

	"github.com/tj-smith47/shelly-go/firmware"
	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// FirmwareInfo contains firmware update information.
type FirmwareInfo struct {
	Current     string
	Available   string
	Beta        string
	HasUpdate   bool
	DeviceModel string
	DeviceID    string
	Generation  int
}

// FirmwareStatus contains the current firmware status.
type FirmwareStatus struct {
	Status      string
	HasUpdate   bool
	NewVersion  string
	Progress    int
	CanRollback bool
}

// CheckFirmware checks for firmware updates on a device.
func (s *Service) CheckFirmware(ctx context.Context, identifier string) (*FirmwareInfo, error) {
	var result *FirmwareInfo
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		mgr := firmware.New(conn.RPCClient())
		info, err := mgr.CheckForUpdate(ctx)
		if err != nil {
			return err
		}

		deviceInfo := conn.Info()
		result = &FirmwareInfo{
			Current:     info.Current,
			Available:   info.Available,
			Beta:        info.Beta,
			HasUpdate:   info.HasUpdate(),
			DeviceModel: deviceInfo.Model,
			DeviceID:    deviceInfo.ID,
			Generation:  deviceInfo.Generation,
		}
		return nil
	})
	return result, err
}

// GetFirmwareStatus gets the current firmware status.
func (s *Service) GetFirmwareStatus(ctx context.Context, identifier string) (*FirmwareStatus, error) {
	var result *FirmwareStatus
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		mgr := firmware.New(conn.RPCClient())

		status, err := mgr.GetStatus(ctx)
		if err != nil {
			return err
		}

		rollbackStatus, rollbackErr := mgr.GetRollbackStatus(ctx)
		canRollback := false
		if rollbackErr == nil && rollbackStatus != nil {
			canRollback = rollbackStatus.CanRollback
		}

		result = &FirmwareStatus{
			Status:      status.Status,
			HasUpdate:   status.HasUpdate,
			NewVersion:  status.NewVersion,
			Progress:    status.Progress,
			CanRollback: canRollback,
		}
		return nil
	})
	return result, err
}

// UpdateFirmware starts a firmware update on a device.
func (s *Service) UpdateFirmware(ctx context.Context, identifier string, opts *firmware.UpdateOptions) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		mgr := firmware.New(conn.RPCClient())
		return mgr.Update(ctx, opts)
	})
}

// UpdateFirmwareStable updates to the latest stable firmware.
func (s *Service) UpdateFirmwareStable(ctx context.Context, identifier string) error {
	return s.UpdateFirmware(ctx, identifier, &firmware.UpdateOptions{Stage: "stable"})
}

// UpdateFirmwareBeta updates to the latest beta firmware.
func (s *Service) UpdateFirmwareBeta(ctx context.Context, identifier string) error {
	return s.UpdateFirmware(ctx, identifier, &firmware.UpdateOptions{Stage: "beta"})
}

// UpdateFirmwareFromURL updates from a custom firmware URL.
func (s *Service) UpdateFirmwareFromURL(ctx context.Context, identifier, url string) error {
	return s.UpdateFirmware(ctx, identifier, &firmware.UpdateOptions{URL: url})
}

// RollbackFirmware rolls back to the previous firmware version.
func (s *Service) RollbackFirmware(ctx context.Context, identifier string) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		mgr := firmware.New(conn.RPCClient())
		return mgr.Rollback(ctx)
	})
}

// GetFirmwareURL gets the firmware download URL for a device.
func (s *Service) GetFirmwareURL(ctx context.Context, identifier, stage string) (string, error) {
	var result string
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		mgr := firmware.New(conn.RPCClient())
		url, err := mgr.GetFirmwareURL(ctx, stage)
		if err != nil {
			return err
		}
		result = url
		return nil
	})
	return result, err
}

// FirmwareCheckResult holds the result of a firmware check for a single device.
type FirmwareCheckResult struct {
	Name string
	Info *FirmwareInfo
	Err  error
}

// CheckFirmwareAll checks firmware on multiple devices concurrently.
func (s *Service) CheckFirmwareAll(ctx context.Context, ios *iostreams.IOStreams, devices []string) []FirmwareCheckResult {
	var (
		results []FirmwareCheckResult
		mu      sync.Mutex
	)

	ios.StartProgress("Checking firmware on all devices...")

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(5) // Limit concurrent checks

	for _, name := range devices {
		deviceName := name
		g.Go(func() error {
			info, checkErr := s.CheckFirmware(gctx, deviceName)
			mu.Lock()
			results = append(results, FirmwareCheckResult{Name: deviceName, Info: info, Err: checkErr})
			mu.Unlock()
			return nil // Don't fail the whole group on individual errors
		})
	}

	if err := g.Wait(); err != nil {
		ios.DebugErr("errgroup wait", err)
	}
	ios.StopProgress()

	return results
}
