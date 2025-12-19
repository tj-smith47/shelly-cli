// Package sync provides the sync command for synchronizing devices.
package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds the command options.
type Options struct {
	Push    bool
	Pull    bool
	DryRun  bool
	Devices []string
}

// pullResult holds the result of pulling a device config.
type pullResult struct {
	config map[string]any
	err    error
}

// fetchDeviceConfig fetches config from a device and returns it as a map.
func fetchDeviceConfig(ctx context.Context, svc *shelly.Service, device string) pullResult {
	conn, err := svc.Connect(ctx, device)
	if err != nil {
		return pullResult{err: err}
	}

	rawResult, err := conn.Call(ctx, "Shelly.GetConfig", nil)
	iostreams.CloseWithDebug("closing sync connection", conn)
	if err != nil {
		return pullResult{err: err}
	}

	jsonBytes, err := json.Marshal(rawResult)
	if err != nil {
		return pullResult{err: fmt.Errorf("marshal: %w", err)}
	}

	var deviceConfig map[string]any
	if err := json.Unmarshal(jsonBytes, &deviceConfig); err != nil {
		return pullResult{err: fmt.Errorf("unmarshal: %w", err)}
	}

	return pullResult{config: deviceConfig}
}

// saveDeviceConfig saves a device config to a file.
func saveDeviceConfig(syncDir, device string, cfg map[string]any) error {
	filename := filepath.Join(syncDir, fmt.Sprintf("%s.json", device))
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err := os.WriteFile(filename, data, 0o600); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}

// loadDeviceConfig loads a device config from a file.
func loadDeviceConfig(syncDir, filename string) (map[string]any, error) {
	fullPath := filepath.Join(syncDir, filename)
	data, err := os.ReadFile(fullPath) //nolint:gosec // User-managed sync directory
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	return cfg, nil
}

// pushDeviceConfig pushes config to a device.
func pushDeviceConfig(ctx context.Context, svc *shelly.Service, device string, cfg map[string]any) error {
	conn, err := svc.Connect(ctx, device)
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

// NewCommand creates the sync command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:     "sync",
		Aliases: []string{"synchronize"},
		Short:   "Synchronize device configurations",
		Long: `Synchronize device configurations between local storage and devices.

Operations:
  --pull  Download device configs to local storage
  --push  Upload local configs to devices

Configurations are stored in the CLI config directory.`,
		Example: `  # Pull all device configs to local storage
  shelly sync --pull

  # Push local config to specific device
  shelly sync --push --device kitchen-light

  # Preview sync without making changes
  shelly sync --pull --dry-run`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), f, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Push, "push", false, "Push local configs to devices")
	cmd.Flags().BoolVar(&opts.Pull, "pull", false, "Pull device configs to local storage")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Preview changes without applying")
	cmd.Flags().StringSliceVar(&opts.Devices, "device", nil, "Specific devices to sync (default: all)")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, opts *Options) error {
	if !opts.Push && !opts.Pull {
		return fmt.Errorf("specify --push or --pull")
	}

	if opts.Push && opts.Pull {
		return fmt.Errorf("cannot use --push and --pull together")
	}

	if opts.Pull {
		return runPull(ctx, f, opts)
	}

	return runPush(ctx, f, opts)
}

func runPull(ctx context.Context, f *cmdutil.Factory, opts *Options) error {
	ios := f.IOStreams()
	svc := f.ShellyService()
	cfg, err := f.Config()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get sync directory
	syncDir, err := getSyncDir()
	if err != nil {
		return err
	}

	// Determine devices to sync
	devices := opts.Devices
	if len(devices) == 0 {
		// Use configured devices
		for name := range cfg.Devices {
			devices = append(devices, name)
		}
	}

	if len(devices) == 0 {
		ios.Warning("No devices configured. Add devices with 'shelly config device add'")
		return nil
	}

	ios.Info("Pulling configurations from %d device(s)...", len(devices))
	if opts.DryRun {
		ios.Warning("[DRY RUN] No changes will be made")
	}
	ios.Println()

	var success, failed int

	for _, device := range devices {
		ios.Printf("  %s: ", device)

		if opts.DryRun {
			result := fetchDeviceConfig(ctx, svc, device)
			if result.err != nil {
				ios.Printf("failed (%v)\n", result.err)
				failed++
			} else {
				ios.Printf("would save config\n")
				success++
			}
			continue
		}

		result := fetchDeviceConfig(ctx, svc, device)
		if result.err != nil {
			ios.Printf("failed (%v)\n", result.err)
			failed++
			continue
		}

		if err := saveDeviceConfig(syncDir, device, result.config); err != nil {
			ios.Printf("failed (%v)\n", err)
			failed++
			continue
		}

		ios.Printf("saved\n")
		success++
	}

	ios.Println()
	if failed > 0 {
		ios.Warning("Completed: %d succeeded, %d failed", success, failed)
	} else {
		ios.Success("Completed: %d device(s) synced", success)
	}

	if !opts.DryRun {
		ios.Info("Configs saved to: %s", syncDir)
	}

	return nil
}

func runPush(ctx context.Context, f *cmdutil.Factory, opts *Options) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	// Get sync directory
	syncDir, err := getSyncDir()
	if err != nil {
		return err
	}

	// Find config files
	files, err := os.ReadDir(syncDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no sync directory found; run 'shelly sync --pull' first")
		}
		return fmt.Errorf("failed to read sync directory: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no config files found; run 'shelly sync --pull' first")
	}

	ios.Info("Pushing configurations to devices...")
	if opts.DryRun {
		ios.Warning("[DRY RUN] No changes will be made")
	}
	ios.Println()

	var success, failed, skipped int

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		deviceName := file.Name()[:len(file.Name())-5] // Remove .json

		// Filter by specified devices
		if len(opts.Devices) > 0 && !slices.Contains(opts.Devices, deviceName) {
			skipped++
			continue
		}

		ios.Printf("  %s: ", deviceName)

		if opts.DryRun {
			ios.Printf("would push config\n")
			success++
			continue
		}

		configData, err := loadDeviceConfig(syncDir, file.Name())
		if err != nil {
			ios.Printf("failed (%v)\n", err)
			failed++
			continue
		}

		if err := pushDeviceConfig(ctx, svc, deviceName, configData); err != nil {
			ios.Printf("failed (%v)\n", err)
			failed++
			continue
		}

		ios.Printf("pushed\n")
		success++
	}

	ios.Println()
	if failed > 0 {
		ios.Warning("Completed: %d succeeded, %d failed, %d skipped", success, failed, skipped)
	} else {
		ios.Success("Completed: %d device(s) updated", success)
	}

	return nil
}

func getSyncDir() (string, error) {
	configDir, err := config.Dir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}

	syncDir := filepath.Join(configDir, "sync")

	// Ensure directory exists
	if err := os.MkdirAll(syncDir, 0o700); err != nil {
		return "", fmt.Errorf("failed to create sync directory: %w", err)
	}

	return syncDir, nil
}
