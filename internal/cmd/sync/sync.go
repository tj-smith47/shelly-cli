// Package sync provides the sync command for synchronizing devices.
package sync

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Devices []string
	DryRun  bool
	Pull    bool
	Push    bool
}

// NewCommand creates the sync command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Push, "push", false, "Push local configs to devices")
	cmd.Flags().BoolVar(&opts.Pull, "pull", false, "Pull device configs to local storage")
	flags.AddDryRunFlag(cmd, &opts.DryRun)
	cmd.Flags().StringSliceVar(&opts.Devices, "device", nil, "Specific devices to sync (default: all)")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	if !opts.Push && !opts.Pull {
		return fmt.Errorf("specify --push or --pull")
	}

	if opts.Push && opts.Pull {
		return fmt.Errorf("cannot use --push and --pull together")
	}

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	syncDir, err := config.GetSyncDir()
	if err != nil {
		return err
	}

	// Progress callback displays each device result
	progress := func(result shelly.SyncDeviceResult) {
		term.DisplaySyncProgress(ios, result.Device, result.Status)
	}

	if opts.Pull {
		cfg, err := opts.Factory.Config()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
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

		term.DisplaySyncHeader(ios, "Pulling", len(devices), opts.DryRun)
		success, failed := svc.PullDeviceConfigs(ctx, devices, syncDir, opts.DryRun, progress)
		term.DisplaySyncSummary(ios, success, failed, opts.DryRun, syncDir)
		return nil
	}

	// Push operation
	term.DisplaySyncHeader(ios, "Pushing", 0, opts.DryRun)
	success, failed, skipped, err := svc.PushDeviceConfigs(ctx, syncDir, opts.Devices, opts.DryRun, progress)
	if err != nil {
		return err
	}
	term.DisplayPushSummary(ios, success, failed, skipped)
	return nil
}
