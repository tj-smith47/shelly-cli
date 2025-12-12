// Package status provides the energy status command.
package status

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

const (
	typeAuto = "auto"
	typeEM   = "em"
	typeEM1  = "em1"
)

// NewCmd creates the energy status command.
func NewCmd(f *cmdutil.Factory) *cobra.Command {
	var (
		componentID   int
		componentType string
	)

	cmd := &cobra.Command{
		Use:   "status <device> [id]",
		Short: "Show energy monitor status",
		Long: `Show current status of an energy monitoring component.

Displays real-time measurements including voltage, current, power,
power factor, and frequency. For 3-phase EM components, shows
per-phase data and totals.`,
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

	cmd.Flags().StringVar(&componentType, "type", typeAuto, "Component type (auto, em, em1)")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, id int, componentType string) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	// Auto-detect type if not specified
	if componentType == typeAuto {
		emIDs, err := svc.ListEMComponents(ctx, device)
		ios.DebugErr("list EM components", err)
		em1IDs, err := svc.ListEM1Components(ctx, device)
		ios.DebugErr("list EM1 components", err)

		for _, emID := range emIDs {
			if emID == id {
				componentType = typeEM
				break
			}
		}
		if componentType == typeAuto {
			for _, em1ID := range em1IDs {
				if em1ID == id {
					componentType = typeEM1
					break
				}
			}
		}
		if componentType == typeAuto && len(emIDs) > 0 {
			componentType = typeEM
		} else if componentType == typeAuto && len(em1IDs) > 0 {
			componentType = typeEM1
		}
	}

	switch componentType {
	case typeEM:
		return showEMStatus(ctx, ios, svc, device, id)
	case typeEM1:
		return showEM1Status(ctx, ios, svc, device, id)
	default:
		return fmt.Errorf("no energy monitoring components found")
	}
}

func showEMStatus(ctx context.Context, ios *iostreams.IOStreams, svc *shelly.Service, device string, id int) error {
	status, err := svc.GetEMStatus(ctx, device, id)
	if err != nil {
		return fmt.Errorf("failed to get EM status: %w", err)
	}

	if cmdutil.WantsStructured() {
		return cmdutil.FormatOutput(ios, status)
	}

	// Human-readable format
	ios.Printf("Energy Monitor (EM) #%d\n\n", status.ID)

	ios.Printf("Phase A:\n")
	ios.Printf("  Voltage:        %.2f V\n", status.AVoltage)
	ios.Printf("  Current:        %.2f A\n", status.ACurrent)
	ios.Printf("  Active Power:   %.2f W\n", status.AActivePower)
	ios.Printf("  Apparent Power: %.2f VA\n", status.AApparentPower)
	if status.APowerFactor != nil {
		ios.Printf("  Power Factor:   %.3f\n", *status.APowerFactor)
	}
	if status.AFreq != nil {
		ios.Printf("  Frequency:      %.2f Hz\n", *status.AFreq)
	}

	ios.Printf("\nPhase B:\n")
	ios.Printf("  Voltage:        %.2f V\n", status.BVoltage)
	ios.Printf("  Current:        %.2f A\n", status.BCurrent)
	ios.Printf("  Active Power:   %.2f W\n", status.BActivePower)
	ios.Printf("  Apparent Power: %.2f VA\n", status.BApparentPower)
	if status.BPowerFactor != nil {
		ios.Printf("  Power Factor:   %.3f\n", *status.BPowerFactor)
	}
	if status.BFreq != nil {
		ios.Printf("  Frequency:      %.2f Hz\n", *status.BFreq)
	}

	ios.Printf("\nPhase C:\n")
	ios.Printf("  Voltage:        %.2f V\n", status.CVoltage)
	ios.Printf("  Current:        %.2f A\n", status.CCurrent)
	ios.Printf("  Active Power:   %.2f W\n", status.CActivePower)
	ios.Printf("  Apparent Power: %.2f VA\n", status.CApparentPower)
	if status.CPowerFactor != nil {
		ios.Printf("  Power Factor:   %.3f\n", *status.CPowerFactor)
	}
	if status.CFreq != nil {
		ios.Printf("  Frequency:      %.2f Hz\n", *status.CFreq)
	}

	ios.Printf("\nTotals:\n")
	ios.Printf("  Current:        %.2f A\n", status.TotalCurrent)
	ios.Printf("  Active Power:   %.2f W\n", status.TotalActivePower)
	ios.Printf("  Apparent Power: %.2f VA\n", status.TotalAprtPower)

	if status.NCurrent != nil {
		ios.Printf("\nNeutral Current: %.2f A\n", *status.NCurrent)
	}

	if len(status.Errors) > 0 {
		ios.Printf("\nErrors: %v\n", status.Errors)
	}

	return nil
}

func showEM1Status(ctx context.Context, ios *iostreams.IOStreams, svc *shelly.Service, device string, id int) error {
	status, err := svc.GetEM1Status(ctx, device, id)
	if err != nil {
		return fmt.Errorf("failed to get EM1 status: %w", err)
	}

	if cmdutil.WantsStructured() {
		return cmdutil.FormatOutput(ios, status)
	}

	// Human-readable format
	ios.Printf("Energy Monitor (EM1) #%d\n\n", status.ID)
	ios.Printf("Voltage:        %.2f V\n", status.Voltage)
	ios.Printf("Current:        %.2f A\n", status.Current)
	ios.Printf("Active Power:   %.2f W\n", status.ActPower)
	ios.Printf("Apparent Power: %.2f VA\n", status.AprtPower)
	if status.PF != nil {
		ios.Printf("Power Factor:   %.3f\n", *status.PF)
	}
	if status.Freq != nil {
		ios.Printf("Frequency:      %.2f Hz\n", *status.Freq)
	}

	if len(status.Errors) > 0 {
		ios.Printf("\nErrors: %v\n", status.Errors)
	}

	return nil
}
