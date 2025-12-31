// Package status provides the energy status command.
package status

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Factory       *cmdutil.Factory
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
		Use:   "status <device> [id]",
		Short: "Show energy monitor status",
		Long: `Show current status of an energy monitoring component.

Displays real-time measurements including voltage, current, power,
power factor, and frequency. For 3-phase EM components, shows
per-phase data and totals.`,
		Example: `  # Show energy monitor status
  shelly energy status shelly-3em-pro

  # Show specific component by ID
  shelly energy status shelly-em 0

  # Specify component type explicitly
  shelly energy status shelly-em --type em1

  # Output as JSON for scripting
  shelly energy status shelly-3em-pro -o json`,
		Aliases: []string{"st"},
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
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

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

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
