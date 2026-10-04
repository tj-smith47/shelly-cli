// Package status provides the energy status command.
package status

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Factory       *cmdutil.Factory
	All           bool
	ComponentID   int
	ComponentType string
	Device        string
}

// NewCommand creates the energy status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory:       f,
		ComponentType: shelly.ComponentTypeAuto,
	}

	cmd := &cobra.Command{
		Use:   "status [device] [id]",
		Short: "Show energy monitor status",
		Long: `Show current status of an energy monitoring component.

Displays real-time measurements including voltage, current, power,
power factor, and frequency. For 3-phase EM components, shows
per-phase data and totals.

With --all, shows every EM and EM1 component on every registered
device as one list. Devices that are offline or have no EM or EM1
component are skipped with a note on stderr. With -o json or -o yaml
the list is printed as one array whose items carry name, type, id,
power (watts) and the full em or em1 reading.`,
		Example: `  # Show energy monitor status
  shelly energy status shelly-3em-pro

  # Show specific component by ID
  shelly energy status shelly-em 0

  # Specify component type explicitly
  shelly energy status shelly-em --type em1

  # Output as JSON for scripting
  shelly energy status shelly-3em-pro -o json

  # Show every energy monitor on every registered device
  shelly energy status --all

  # Every energy monitor as one JSON list
  shelly energy status --all -o json`,
		Aliases: []string{"st"},
		Args:    cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch {
			case opts.All && len(args) > 0:
				return fmt.Errorf("--all reports every registered device; do not pass a device or component ID")
			case !opts.All && len(args) == 0:
				return fmt.Errorf("requires a device argument, or --all")
			}
			if len(args) > 0 {
				opts.Device = args[0]
			}
			if len(args) == 2 {
				_, err := fmt.Sscanf(args[1], "%d", &opts.ComponentID)
				if err != nil {
					return fmt.Errorf("invalid component ID: %w", err)
				}
			}
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.ComponentType, "type", shelly.ComponentTypeAuto, "Component type (auto, em, em1)")
	flags.AddAllFlag(cmd, &opts.All)
	cmd.MarkFlagsMutuallyExclusive("all", "type")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	if opts.All {
		cfg, err := opts.Factory.Config()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		devices := slices.Sorted(maps.Keys(cfg.Devices))
		if len(devices) == 0 {
			ios.NoResults("registered devices", "Use 'shelly device add' to register devices first")
			return nil
		}

		var entries []model.EnergyStatusEntry
		var skipped []model.EnergyStatusSkip
		err = cmdutil.RunWithSpinner(ctx, ios, fmt.Sprintf("Reading energy monitors on %d devices...", len(devices)), func(ctx context.Context) error {
			entries, skipped = svc.CollectEnergyStatuses(ctx, devices)
			return nil
		})
		if err != nil {
			return err
		}
		for _, sk := range skipped {
			ios.Warning("Skipped %s: %s", sk.Device, sk.Reason)
		}
		return cmdutil.PrintDiscovered(ios, entries, term.DisplayEnergyStatusList, "energy monitor components")
	}

	// Auto-detect type if not specified
	componentType := opts.ComponentType
	if componentType == shelly.ComponentTypeAuto {
		componentType = svc.DetectEnergyComponentByID(ctx, ios, opts.Device, opts.ComponentID)
	}

	switch componentType {
	case shelly.ComponentTypeEM:
		status, err := svc.GetEMStatus(ctx, opts.Device, opts.ComponentID)
		if err != nil {
			return fmt.Errorf("failed to get EM status: %w", err)
		}
		term.DisplayEMStatus(ios, status)
		return nil
	case shelly.ComponentTypeEM1:
		status, err := svc.GetEM1Status(ctx, opts.Device, opts.ComponentID)
		if err != nil {
			return fmt.Errorf("failed to get EM1 status: %w", err)
		}
		term.DisplayEM1Status(ios, status)
		return nil
	default:
		return fmt.Errorf("no energy monitoring components found")
	}
}
