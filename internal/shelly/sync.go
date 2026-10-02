package shelly

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

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

// PushDeviceConfig pushes a config captured from the same device back to it
// and returns a warning for each WiFi station left out. A station on the
// network the device is already on is written without a key so the device
// keeps its own; a changed network takes this host's stored password, or the
// station is left out.
func (s *Service) PushDeviceConfig(ctx context.Context, device string, cfg map[string]any) ([]string, error) {
	conn, err := s.Connect(ctx, device)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	defer iostreams.CloseWithDebug("closing sync push connection", conn)

	var current map[string]any
	if configHasStation(cfg) {
		if current, err = conn.GetConfig(ctx); err != nil {
			return nil, fmt.Errorf("read current config: %w", err)
		}
	}
	planned, warnings := s.planConfigStations(ctx, cfg, current, stationsSameDevice)
	if _, err := conn.Call(ctx, "Shelly.SetConfig", map[string]any{fieldConfig: planned}); err != nil {
		return warnings, err
	}
	return warnings, nil
}

// SyncDeviceResult holds the result of a device sync operation.
type SyncDeviceResult struct {
	Device string
	Status string
	Err    error
}

// SyncProgressCallback is called for each device during batch sync operations.
type SyncProgressCallback func(result SyncDeviceResult)

// PullDeviceConfigs pulls configurations from multiple devices concurrently.
// It calls the progress callback for each device as it completes.
func (s *Service) PullDeviceConfigs(ctx context.Context, devices []string, syncDir string, dryRun bool, progress SyncProgressCallback) (success, failed int) {
	var successCount, failedCount atomic.Int64
	var wg sync.WaitGroup

	for _, device := range devices {
		d := device
		wg.Go(func() {
			devCtx, cancel := context.WithTimeout(ctx, DefaultTimeout)
			defer cancel()

			result := s.FetchDeviceConfig(devCtx, d)
			if result.Err != nil {
				progress(SyncDeviceResult{Device: d, Status: fmt.Sprintf("failed (%v)", result.Err), Err: result.Err})
				failedCount.Add(1)
				return
			}

			if dryRun {
				progress(SyncDeviceResult{Device: d, Status: "would save config"})
				successCount.Add(1)
				return
			}

			if err := config.SaveSyncConfig(syncDir, d, result.Config); err != nil {
				progress(SyncDeviceResult{Device: d, Status: fmt.Sprintf("failed (%v)", err), Err: err})
				failedCount.Add(1)
				return
			}

			progress(SyncDeviceResult{Device: d, Status: "saved"})
			successCount.Add(1)
		})
	}

	wg.Wait()
	return int(successCount.Load()), int(failedCount.Load())
}

// PushDeviceConfigs pushes configurations to multiple devices concurrently from local files.
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

	// Collect push work items (cheap local I/O), then execute network calls concurrently.
	type pushItem struct {
		deviceName string
		fileName   string
	}
	var items []pushItem
	var skippedCount int

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		deviceName := file.Name()[:len(file.Name())-5] // Remove .json

		if len(deviceFilter) > 0 && !slices.Contains(deviceFilter, deviceName) {
			skippedCount++
			continue
		}

		if dryRun {
			progress(SyncDeviceResult{Device: deviceName, Status: "would push config"})
			items = append(items, pushItem{deviceName: deviceName}) // count as success
			continue
		}

		items = append(items, pushItem{deviceName: deviceName, fileName: file.Name()})
	}

	if dryRun {
		return len(items), 0, skippedCount, nil
	}

	var successCount, failedCount atomic.Int64
	var wg sync.WaitGroup

	for _, item := range items {
		it := item
		wg.Go(func() {
			if s.pushSyncFile(ctx, syncDir, it.deviceName, it.fileName, progress) {
				successCount.Add(1)
			} else {
				failedCount.Add(1)
			}
		})
	}

	wg.Wait()
	return int(successCount.Load()), int(failedCount.Load()), skippedCount, nil
}

// pushSyncFile pushes one device's sync file and reports the outcome to
// progress; it returns false when the push failed.
func (s *Service) pushSyncFile(ctx context.Context, syncDir, deviceName, fileName string, progress SyncProgressCallback) bool {
	configData, err := config.LoadSyncConfig(syncDir, fileName)
	if err != nil {
		progress(SyncDeviceResult{Device: deviceName, Status: fmt.Sprintf("failed (%v)", err), Err: err})
		return false
	}

	devCtx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	warnings, err := s.PushDeviceConfig(devCtx, deviceName, configData)
	if err != nil {
		progress(SyncDeviceResult{Device: deviceName, Status: fmt.Sprintf("failed (%v)", err), Err: err})
		return false
	}

	status := "pushed"
	if len(warnings) > 0 {
		status += " (" + strings.Join(warnings, "; ") + ")"
	}
	progress(SyncDeviceResult{Device: deviceName, Status: status})
	return true
}
