package shelly

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// SyncResult holds the result of syncing a device config.
type SyncResult struct {
	Config map[string]any
	Err    error
}

// FetchDeviceConfig fetches config from a device and returns it as a map.
func (s *Service) FetchDeviceConfig(ctx context.Context, device string) SyncResult {
	conn, err := s.Connect(ctx, device)
	if err != nil {
		return SyncResult{Err: err}
	}

	rawResult, err := conn.Call(ctx, "Shelly.GetConfig", nil)
	iostreams.CloseWithDebug("closing sync connection", conn)
	if err != nil {
		return SyncResult{Err: err}
	}

	jsonBytes, err := json.Marshal(rawResult)
	if err != nil {
		return SyncResult{Err: fmt.Errorf("marshal: %w", err)}
	}

	var deviceConfig map[string]any
	if err := json.Unmarshal(jsonBytes, &deviceConfig); err != nil {
		return SyncResult{Err: fmt.Errorf("unmarshal: %w", err)}
	}

	return SyncResult{Config: deviceConfig}
}

// PushDeviceConfig pushes config to a device.
func (s *Service) PushDeviceConfig(ctx context.Context, device string, cfg map[string]any) error {
	conn, err := s.Connect(ctx, device)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	_, err = conn.Call(ctx, "Shelly.SetConfig", map[string]any{"config": cfg})
	iostreams.CloseWithDebug("closing sync push connection", conn)
	if err != nil {
		return err
	}
	return nil
}

// SyncDeviceResult holds the result of a device sync operation.
type SyncDeviceResult struct {
	Device string
	Status string
	Err    error
}

// SyncProgressCallback is called for each device during batch sync operations.
type SyncProgressCallback func(result SyncDeviceResult)

// PullDeviceConfigs pulls configurations from multiple devices.
// It calls the progress callback for each device as it completes.
func (s *Service) PullDeviceConfigs(ctx context.Context, devices []string, syncDir string, dryRun bool, progress SyncProgressCallback) (success, failed int) {
	for _, device := range devices {
		result := s.FetchDeviceConfig(ctx, device)
		if result.Err != nil {
			progress(SyncDeviceResult{Device: device, Status: fmt.Sprintf("failed (%v)", result.Err), Err: result.Err})
			failed++
			continue
		}

		if dryRun {
			progress(SyncDeviceResult{Device: device, Status: "would save config"})
			success++
			continue
		}

		if err := config.SaveSyncConfig(syncDir, device, result.Config); err != nil {
			progress(SyncDeviceResult{Device: device, Status: fmt.Sprintf("failed (%v)", err), Err: err})
			failed++
			continue
		}

		progress(SyncDeviceResult{Device: device, Status: "saved"})
		success++
	}
	return success, failed
}

// PushDeviceConfigs pushes configurations to multiple devices from local files.
// It calls the progress callback for each device as it completes.
func (s *Service) PushDeviceConfigs(ctx context.Context, syncDir string, deviceFilter []string, dryRun bool, progress SyncProgressCallback) (success, failed, skipped int, err error) {
	fs := config.Fs()
	files, err := afero.ReadDir(fs, syncDir)
	if err != nil {
		exists, existsErr := afero.Exists(fs, syncDir)
		if existsErr == nil && !exists {
			return 0, 0, 0, fmt.Errorf("no sync directory found; run 'shelly sync --pull' first")
		}
		return 0, 0, 0, fmt.Errorf("failed to read sync directory: %w", err)
	}

	if len(files) == 0 {
		return 0, 0, 0, fmt.Errorf("no config files found; run 'shelly sync --pull' first")
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		deviceName := file.Name()[:len(file.Name())-5] // Remove .json

		if len(deviceFilter) > 0 && !slices.Contains(deviceFilter, deviceName) {
			skipped++
			continue
		}

		if dryRun {
			progress(SyncDeviceResult{Device: deviceName, Status: "would push config"})
			success++
			continue
		}

		configData, loadErr := config.LoadSyncConfig(syncDir, file.Name())
		if loadErr != nil {
			progress(SyncDeviceResult{Device: deviceName, Status: fmt.Sprintf("failed (%v)", loadErr), Err: loadErr})
			failed++
			continue
		}

		if pushErr := s.PushDeviceConfig(ctx, deviceName, configData); pushErr != nil {
			progress(SyncDeviceResult{Device: deviceName, Status: fmt.Sprintf("failed (%v)", pushErr), Err: pushErr})
			failed++
			continue
		}

		progress(SyncDeviceResult{Device: deviceName, Status: "pushed"})
		success++
	}
	return success, failed, skipped, nil
}
