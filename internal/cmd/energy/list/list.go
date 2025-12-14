// Package list provides the energy list command.
package list

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// NewCommand creates the energy list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <device>",
		Short: "List energy monitoring components",
		Long: `List all energy monitoring components (EM/EM1) on a device.

Shows component IDs and types for all energy monitors found on the device.
EM components are 3-phase monitors, EM1 components are single-phase.`,
		Aliases:           []string{"ls"},
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	return cmd
}

type componentInfo struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	// List EM components
	emIDs, err := svc.ListEMComponents(ctx, device)
	if err != nil {
		return fmt.Errorf("failed to list EM components: %w", err)
	}

	// List EM1 components
	em1IDs, err := svc.ListEM1Components(ctx, device)
	if err != nil {
		return fmt.Errorf("failed to list EM1 components: %w", err)
	}

	// Combine results
	components := make([]componentInfo, 0, len(emIDs)+len(em1IDs))
	for _, id := range emIDs {
		components = append(components, componentInfo{
			ID:   id,
			Type: "EM (3-phase)",
		})
	}
	for _, id := range em1IDs {
		components = append(components, componentInfo{
			ID:   id,
			Type: "EM1 (single-phase)",
		})
	}

	if len(components) == 0 {
		_, err := fmt.Fprintf(ios.Out, "No energy monitoring components found\n")
		iostreams.DebugErr("writing output", err)
		return nil
	}

	// Output results
	return cmdutil.PrintListResult(ios, components, func(ios *iostreams.IOStreams, items []componentInfo) {
		table := output.NewTable("ID", "Type")
		for _, comp := range items {
			table.AddRow(fmt.Sprintf("%d", comp.ID), comp.Type)
		}
		if err := table.PrintTo(ios.Out); err != nil {
			ios.DebugErr("print table", err)
		}
	})
}
