// Package status provides the power status command.
package status

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
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
		// Both PM and PM1 return PMStatus
		return showPMStatus(ctx, ios, svc, device, id, componentType)
	default:
		return fmt.Errorf("no power meter components found")
	}
}

func showPMStatus(ctx context.Context, ios *iostreams.IOStreams, svc *shelly.Service, device string, id int, componentType string) error {
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

	if cmdutil.WantsStructured() {
		return cmdutil.FormatOutput(ios, status)
	}

	// Human-readable format
	typeLabel := "Power Meter (PM)"
	if componentType == shelly.ComponentTypePM1 {
		typeLabel = "Power Meter (PM1)"
	}
	ios.Printf("%s #%d\n\n", typeLabel, status.ID)
	ios.Printf("Voltage: %.2f V\n", status.Voltage)
	ios.Printf("Current: %.2f A\n", status.Current)
	ios.Printf("Power:   %.2f W\n", status.APower)

	if status.Freq != nil {
		ios.Printf("Frequency: %.2f Hz\n", *status.Freq)
	}

	// Energy counters
	if status.AEnergy != nil {
		ios.Printf("\nAccumulated Energy:\n")
		ios.Printf("  Total: %.2f Wh\n", status.AEnergy.Total)
		if status.AEnergy.MinuteTs != nil && len(status.AEnergy.ByMinute) > 0 {
			ios.Printf("  Recent (by minute): %v\n", status.AEnergy.ByMinute[:min(5, len(status.AEnergy.ByMinute))])
		}
	}

	if status.RetAEnergy != nil {
		ios.Printf("\nReturn Energy:\n")
		ios.Printf("  Total: %.2f Wh\n", status.RetAEnergy.Total)
	}

	if len(status.Errors) > 0 {
		ios.Printf("\nErrors: %v\n", status.Errors)
	}

	return nil
}
