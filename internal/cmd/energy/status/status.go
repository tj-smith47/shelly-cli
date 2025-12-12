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

// NewCmd creates the energy status command.
func NewCmd(f *cmdutil.Factory) *cobra.Command {
	var (
		componentID int
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

	cmd.Flags().StringVar(&componentType, "type", "auto", "Component type (auto, em, em1)")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, id int, componentType string) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	// Auto-detect type if not specified
	if componentType == "auto" {
		emIDs, _ := svc.ListEMComponents(ctx, device)
		em1IDs, _ := svc.ListEM1Components(ctx, device)

		for _, emID := range emIDs {
			if emID == id {
				componentType = "em"
				break
			}
		}
		if componentType == "auto" {
			for _, em1ID := range em1IDs {
				if em1ID == id {
					componentType = "em1"
					break
				}
			}
		}
		if componentType == "auto" && len(emIDs) > 0 {
			componentType = "em"
		} else if componentType == "auto" && len(em1IDs) > 0 {
			componentType = "em1"
		}
	}

	switch componentType {
	case "em":
		return showEMStatus(ctx, ios, svc, device, id)
	case "em1":
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
	fmt.Fprintf(ios.Out, "Energy Monitor (EM) #%d\n\n", status.ID)

	fmt.Fprintf(ios.Out, "Phase A:\n")
	fmt.Fprintf(ios.Out, "  Voltage:        %.2f V\n", status.AVoltage)
	fmt.Fprintf(ios.Out, "  Current:        %.2f A\n", status.ACurrent)
	fmt.Fprintf(ios.Out, "  Active Power:   %.2f W\n", status.AActivePower)
	fmt.Fprintf(ios.Out, "  Apparent Power: %.2f VA\n", status.AApparentPower)
	if status.APowerFactor != nil {
		fmt.Fprintf(ios.Out, "  Power Factor:   %.3f\n", *status.APowerFactor)
	}
	if status.AFreq != nil {
		fmt.Fprintf(ios.Out, "  Frequency:      %.2f Hz\n", *status.AFreq)
	}

	fmt.Fprintf(ios.Out, "\nPhase B:\n")
	fmt.Fprintf(ios.Out, "  Voltage:        %.2f V\n", status.BVoltage)
	fmt.Fprintf(ios.Out, "  Current:        %.2f A\n", status.BCurrent)
	fmt.Fprintf(ios.Out, "  Active Power:   %.2f W\n", status.BActivePower)
	fmt.Fprintf(ios.Out, "  Apparent Power: %.2f VA\n", status.BApparentPower)
	if status.BPowerFactor != nil {
		fmt.Fprintf(ios.Out, "  Power Factor:   %.3f\n", *status.BPowerFactor)
	}
	if status.BFreq != nil {
		fmt.Fprintf(ios.Out, "  Frequency:      %.2f Hz\n", *status.BFreq)
	}

	fmt.Fprintf(ios.Out, "\nPhase C:\n")
	fmt.Fprintf(ios.Out, "  Voltage:        %.2f V\n", status.CVoltage)
	fmt.Fprintf(ios.Out, "  Current:        %.2f A\n", status.CCurrent)
	fmt.Fprintf(ios.Out, "  Active Power:   %.2f W\n", status.CActivePower)
	fmt.Fprintf(ios.Out, "  Apparent Power: %.2f VA\n", status.CApparentPower)
	if status.CPowerFactor != nil {
		fmt.Fprintf(ios.Out, "  Power Factor:   %.3f\n", *status.CPowerFactor)
	}
	if status.CFreq != nil {
		fmt.Fprintf(ios.Out, "  Frequency:      %.2f Hz\n", *status.CFreq)
	}

	fmt.Fprintf(ios.Out, "\nTotals:\n")
	fmt.Fprintf(ios.Out, "  Current:        %.2f A\n", status.TotalCurrent)
	fmt.Fprintf(ios.Out, "  Active Power:   %.2f W\n", status.TotalActivePower)
	fmt.Fprintf(ios.Out, "  Apparent Power: %.2f VA\n", status.TotalAprtPower)

	if status.NCurrent != nil {
		fmt.Fprintf(ios.Out, "\nNeutral Current: %.2f A\n", *status.NCurrent)
	}

	if len(status.Errors) > 0 {
		fmt.Fprintf(ios.Out, "\nErrors: %v\n", status.Errors)
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
	fmt.Fprintf(ios.Out, "Energy Monitor (EM1) #%d\n\n", status.ID)
	fmt.Fprintf(ios.Out, "Voltage:        %.2f V\n", status.Voltage)
	fmt.Fprintf(ios.Out, "Current:        %.2f A\n", status.Current)
	fmt.Fprintf(ios.Out, "Active Power:   %.2f W\n", status.ActPower)
	fmt.Fprintf(ios.Out, "Apparent Power: %.2f VA\n", status.AprtPower)
	if status.PF != nil {
		fmt.Fprintf(ios.Out, "Power Factor:   %.3f\n", *status.PF)
	}
	if status.Freq != nil {
		fmt.Fprintf(ios.Out, "Frequency:      %.2f Hz\n", *status.Freq)
	}

	if len(status.Errors) > 0 {
		fmt.Fprintf(ios.Out, "\nErrors: %v\n", status.Errors)
	}

	return nil
}
