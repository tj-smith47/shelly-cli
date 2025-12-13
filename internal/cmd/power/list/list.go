// Package list provides the power list command.
package list

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// NewCommand creates the power list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <device>",
		Short: "List power meter components",
		Long: `List all power meter components (PM/PM1) on a device.

Shows component IDs and types for all power meters found on the device.`,
		Aliases: []string{"ls"},
		Args:    cobra.ExactArgs(1),
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

	// List PM components
	pmIDs, err := svc.ListPMComponents(ctx, device)
	if err != nil {
		return fmt.Errorf("failed to list PM components: %w", err)
	}

	// List PM1 components
	pm1IDs, err := svc.ListPM1Components(ctx, device)
	if err != nil {
		return fmt.Errorf("failed to list PM1 components: %w", err)
	}

	// Combine results
	components := make([]componentInfo, 0, len(pmIDs)+len(pm1IDs))
	for _, id := range pmIDs {
		components = append(components, componentInfo{
			ID:   id,
			Type: "PM",
		})
	}
	for _, id := range pm1IDs {
		components = append(components, componentInfo{
			ID:   id,
			Type: "PM1",
		})
	}

	if len(components) == 0 {
		_, err := fmt.Fprintf(ios.Out, "No power meter components found\n")
		ios.DebugErr("writing output", err)
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
