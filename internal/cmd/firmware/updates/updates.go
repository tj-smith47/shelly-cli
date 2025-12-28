// Package updates provides the firmware updates subcommand.
// This command provides an interactive workflow for discovering and applying
// firmware updates across all devices, with support for both Shelly and plugin-managed platforms.
package updates

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

// NewCommand creates the firmware updates command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory:  f,
		Parallel: 3,
	}

	cmd := &cobra.Command{
		Use:     "updates",
		Aliases: []string{"upd"},
		Short:   "Interactive firmware update workflow",
		Long: `Check for and apply firmware updates with an interactive workflow.

By default, runs in interactive mode - displays devices with available updates
and prompts for selection. Use --all with --yes for non-interactive batch updates.

Supports both native Shelly devices and plugin-managed devices (Tasmota, etc.).`,
		Example: `  # Interactive mode - check and select devices
  shelly firmware updates

  # Update all devices to stable (non-interactive)
  shelly firmware updates --all --yes

  # Update all devices to beta (non-interactive)
  shelly firmware updates --all --beta --yes

  # Update specific devices to stable
  shelly firmware updates --devices=kitchen,bedroom --stable

  # Update only Tasmota devices
  shelly firmware updates --all --platform=tasmota --yes

  # Update specific devices interactively
  shelly firmware updates --devices=kitchen,bedroom`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.All, "all", false, "Update all devices with available updates")
	cmd.Flags().BoolVar(&opts.Stable, "stable", false, "Use stable release channel (default)")
	cmd.Flags().BoolVar(&opts.Beta, "beta", false, "Use beta/development release channel")
	cmd.Flags().StringVar(&opts.Devices, "devices", "", "Comma-separated list of specific devices to update")
	cmd.Flags().StringVar(&opts.Platform, "platform", "", "Only update devices of this platform (e.g., tasmota)")
	flags.AddYesOnlyFlag(cmd, &opts.ConfirmFlags)
	cmd.Flags().IntVar(&opts.Parallel, "parallel", 3, "Number of devices to update in parallel")

	return cmd
}

// Options holds command options.
type Options struct {
	flags.ConfirmFlags
	Factory  *cmdutil.Factory
	All      bool
	Stable   bool
	Beta     bool
	Devices  string
	Platform string
	Parallel int
}

func run(ctx context.Context, opts *Options) error {
	f := opts.Factory
	ios := f.IOStreams()
	svc := f.ShellyService()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Devices) == 0 {
		ios.Warning("No devices registered. Use 'shelly device add' to add devices.")
		return nil
	}

	// Filter devices and check for updates
	deviceConfigs := shelly.FilterDevicesByNameAndPlatform(cfg.Devices, opts.Devices, opts.Platform)
	if len(deviceConfigs) == 0 {
		ios.Warning("No devices matched the specified filters")
		return nil
	}

	entries := shelly.BuildFirmwareUpdateList(svc.CheckFirmwareAllPlatforms(ctx, ios, deviceConfigs), deviceConfigs)
	if len(entries) == 0 {
		ios.Info("All devices are up to date")
		return nil
	}

	// Display and select devices
	termEntries := term.ConvertToTermEntries(entries)
	term.DisplayFirmwareUpdatesTable(ios, termEntries)

	var selectedIndices []int
	var stage string
	if opts.All || opts.Devices != "" {
		selectedIndices, stage = shelly.SelectEntriesByStage(entries, opts.Beta)
	} else {
		selectedIndices, stage = term.InteractiveFirmwareSelect(ios, termEntries, shelly.AnyHasBeta(entries))
	}

	if len(selectedIndices) == 0 {
		ios.Info("No devices selected for update")
		return nil
	}

	toUpdate := shelly.GetEntriesByIndices(entries, selectedIndices)

	// Confirm if needed
	if !opts.Yes {
		msg := fmt.Sprintf("Update %d device(s) to %s?", len(toUpdate), stage)
		confirmed, confirmErr := ios.Confirm(msg, false)
		if confirmErr != nil {
			return confirmErr
		}
		if !confirmed {
			ios.Warning("Update cancelled")
			return nil
		}
	}

	// Execute updates
	updateResults := svc.UpdateDevices(ctx, ios, shelly.ToDeviceUpdateStatuses(toUpdate), shelly.UpdateOpts{
		Beta:        stage == "beta",
		Parallelism: opts.Parallel,
	})

	term.DisplayUpdateResults(ios, term.ConvertToTermResults(updateResults))

	// Check for failures
	for _, r := range updateResults {
		if !r.Success {
			return fmt.Errorf("some updates failed")
		}
	}
	return nil
}
