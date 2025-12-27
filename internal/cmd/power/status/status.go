// Package status provides the power status command.
package status

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// NewCommand creates the power status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		componentID   int
		componentType string
	)

	cmd := &cobra.Command{
		Use:   "status <device> [id]",
		Short: "Show power meter status",
		Long: `Show current status of a power meter component.

Displays real-time measurements including voltage, current, power,
frequency, and accumulated energy.`,
		Example: `  # Show power meter status
  shelly power status living-room

  # Show specific component by ID
  shelly power status living-room 0

  # Specify component type explicitly
  shelly power status living-room --type pm1

  # Output as JSON for scripting
  shelly power status living-room -o json`,
		Aliases: []string{"st"},
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			device := args[0]
			if len(args) == 2 {
				_, err := fmt.Sscanf(args[1], "%d", &componentID)
				if err != nil {
					return fmt.Errorf("invalid component ID: %w", err)
				}
			}
			return run(cmd.Context(), f, device, componentID, componentType)
		},
	}

	cmd.Flags().StringVar(&componentType, "type", shelly.ComponentTypeAuto, "Component type (auto, pm, pm1)")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, id int, componentType string) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	// Auto-detect type if not specified
	if componentType == shelly.ComponentTypeAuto {
		componentType = svc.DetectPowerComponentType(ctx, ios, device, id)
	}

	switch componentType {
	case shelly.ComponentTypePM, shelly.ComponentTypePM1:
		var status *shelly.PMStatus
		var err error

		if componentType == shelly.ComponentTypePM {
			status, err = svc.GetPMStatus(ctx, device, id)
		} else {
			status, err = svc.GetPM1Status(ctx, device, id)
		}

		if err != nil {
			return fmt.Errorf("failed to get %s status: %w", componentType, err)
		}

		if output.WantsStructured() {
			return output.FormatOutput(ios.Out, status)
		}

		term.DisplayPMStatusDetails(ios, status, componentType)
		return nil
	default:
		return fmt.Errorf("no power meter components found")
	}
}
