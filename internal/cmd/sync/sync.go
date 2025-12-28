// Package sync provides the sync command for synchronizing devices.
package sync

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// Options holds the command options.
type Options struct {
	Push    bool
	Pull    bool
	DryRun  bool
	Devices []string
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
	flags.AddDryRunFlag(cmd, &opts.DryRun)
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

	syncDir, err := config.GetSyncDir()
	if err != nil {
		return err
	}

	devices := opts.Devices
	if len(devices) == 0 {
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

		result := svc.FetchDeviceConfig(ctx, device)
		if result.Err != nil {
			ios.Printf("failed (%v)\n", result.Err)
			failed++
			continue
		}

		if opts.DryRun {
			ios.Printf("would save config\n")
			success++
			continue
		}

		if err := config.SaveSyncConfig(syncDir, device, result.Config); err != nil {
			ios.Printf("failed (%v)\n", err)
			failed++
			continue
		}

		ios.Printf("saved\n")
		success++
	}

	printSyncSummary(ios, success, failed, opts.DryRun, syncDir)
	return nil
}

func runPush(ctx context.Context, f *cmdutil.Factory, opts *Options) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	syncDir, err := config.GetSyncDir()
	if err != nil {
		return err
	}

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

		configData, err := config.LoadSyncConfig(syncDir, file.Name())
		if err != nil {
			ios.Printf("failed (%v)\n", err)
			failed++
			continue
		}

		if err := svc.PushDeviceConfig(ctx, deviceName, configData); err != nil {
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

func printSyncSummary(ios *iostreams.IOStreams, success, failed int, dryRun bool, syncDir string) {
	ios.Println()
	if failed > 0 {
		ios.Warning("Completed: %d succeeded, %d failed", success, failed)
	} else {
		ios.Success("Completed: %d device(s) synced", success)
	}

	if !dryRun {
		ios.Info("Configs saved to: %s", syncDir)
	}
}
