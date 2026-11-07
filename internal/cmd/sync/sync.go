// Package sync provides the sync command for synchronizing devices.
package sync

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// commandUse is the command's Use string and primary name.
const commandUse = "sync"

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Devices []string
	DryRun  bool
	Pull    bool
	Push    bool
	Yes     bool
}

// NewCommand creates the sync command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     commandUse,
		Aliases: []string{"synchronize"},
		Short:   "Synchronize device configurations",
		Long: `Synchronize device configurations between local storage and devices.

Operations:
  --pull  Download device configs to local storage
  --push  Upload local configs to devices

Configurations are stored in the CLI config directory.

WiFi stations on --push: each saved config goes back to the device it was
pulled from, so a station keeps its saved address. A station on the network
the device is already on is written without a password, and the device keeps
its own. A station on another network is written with the password stored on
this host for it; with no stored password it is left out with a warning, and
the network can be set with 'shelly wifi set'.

--push overwrites device configuration, so it asks for confirmation first;
pass --yes to skip the prompt (required when there is no terminal, such as in
a script). --dry-run never prompts.`,
		Example: `  # Pull all device configs to local storage
  shelly sync --pull

  # Push local config to specific device
  shelly sync --push --device kitchen-light

  # Preview sync without making changes
  shelly sync --pull --dry-run

  # Push from a script (no confirmation prompt)
  shelly sync --push --yes`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Push, "push", false, "Push local configs to devices")
	cmd.Flags().BoolVar(&opts.Pull, "pull", false, "Pull device configs to local storage")
	flags.AddDryRunFlag(cmd, &opts.DryRun)
	flags.AddYesFlag(cmd, &opts.Yes)
	cmd.Flags().StringSliceVar(&opts.Devices, "device", nil, "Specific devices to sync (default: all)")

	return cmd
}

// pull downloads the selected (default: every configured) device's config
// into syncDir.
func (o *Options) pull(ctx context.Context, svc *shelly.Service, syncDir string, progress func(shelly.SyncDeviceResult)) error {
	ios := o.Factory.IOStreams()
	cfg, err := o.Factory.Config()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	devices := o.Devices
	if len(devices) == 0 {
		for name := range cfg.Devices {
			devices = append(devices, name)
		}
	}

	if len(devices) == 0 {
		ios.Warning("No devices configured. Add devices with 'shelly config device add'")
		return nil
	}

	term.DisplaySyncHeader(ios, "Pulling", len(devices), o.DryRun)
	success, failed := svc.PullDeviceConfigs(ctx, devices, syncDir, o.DryRun, progress)
	term.DisplaySyncSummary(ios, success, failed, o.DryRun, syncDir)
	return nil
}

// pushConfirmMessage names the sync directory and the devices --push would
// overwrite.
func (o *Options) pushConfirmMessage(syncDir string) string {
	scope := "every device with a saved config"
	if len(o.Devices) > 0 {
		scope = strings.Join(o.Devices, ", ")
	}
	return fmt.Sprintf("Push saved configs from %s to %s? This overwrites device configuration", syncDir, scope)
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
		return opts.pull(ctx, svc, syncDir, progress)
	}

	if !opts.DryRun {
		confirmed, err := opts.Factory.ConfirmAction(opts.pushConfirmMessage(syncDir), opts.Yes)
		if err != nil {
			return err
		}
		if !confirmed {
			ios.Info("Cancelled")
			return nil
		}
	}

	term.DisplaySyncHeader(ios, "Pushing", 0, opts.DryRun)
	success, failed, skipped, err := svc.PushDeviceConfigs(ctx, syncDir, opts.Devices, opts.DryRun, progress)
	if err != nil {
		return err
	}
	term.DisplayPushSummary(ios, success, failed, skipped)
	return nil
}
