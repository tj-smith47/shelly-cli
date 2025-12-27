// Package updates provides the firmware updates subcommand.
// This command provides an interactive workflow for discovering and applying
// firmware updates across all devices, with support for both Shelly and plugin-managed platforms.
package updates

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// NewCommand creates the firmware updates command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		allFlag      bool
		stableFlag   bool
		betaFlag     bool
		devicesFlag  string
		platformFlag string
		yesFlag      bool
		parallelFlag int
	)

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
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := Options{
				All:      allFlag,
				Stable:   stableFlag,
				Beta:     betaFlag,
				Devices:  devicesFlag,
				Platform: platformFlag,
				Yes:      yesFlag,
				Parallel: parallelFlag,
			}
			return run(cmd.Context(), f, opts)
		},
	}

	cmd.Flags().BoolVar(&allFlag, "all", false, "Update all devices with available updates")
	cmd.Flags().BoolVar(&stableFlag, "stable", false, "Use stable release channel (default)")
	cmd.Flags().BoolVar(&betaFlag, "beta", false, "Use beta/development release channel")
	cmd.Flags().StringVar(&devicesFlag, "devices", "", "Comma-separated list of specific devices to update")
	cmd.Flags().StringVar(&platformFlag, "platform", "", "Only update devices of this platform (e.g., tasmota)")
	cmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompts")
	cmd.Flags().IntVar(&parallelFlag, "parallel", 3, "Number of devices to update in parallel")

	return cmd
}

// Options holds command options.
type Options struct {
	All      bool
	Stable   bool
	Beta     bool
	Devices  string
	Platform string
	Yes      bool
	Parallel int
}

func run(ctx context.Context, f *cmdutil.Factory, opts Options) error {
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
