// Package list provides the energy list command.
package list

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// Options holds command options.
type Options struct {
	Device  string
	Factory *cmdutil.Factory
}

// NewCommand creates the energy list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:   "list <device>",
		Short: "List energy monitoring components",
		Long: `List all energy monitoring components (EM/EM1) on a device.

Shows component IDs and types for all energy monitors found on the device.
EM components are 3-phase monitors (Shelly Pro 3EM, etc.), EM1 components
are single-phase monitors (Shelly EM, Shelly Plus 1PM, etc.).

Use 'shelly energy status' with a component ID to get real-time readings.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

Columns: ID, Type`,
		Example: `  # List energy monitoring components on a device
  shelly energy list shelly-3em-pro

  # Output as JSON for scripting
  shelly energy list shelly-3em-pro -o json

  # Get IDs of 3-phase monitors
  shelly energy list shelly-3em-pro -o json | jq -r '.[] | select(.type | contains("3-phase")) | .id'

  # Count total energy components
  shelly energy list shelly-3em-pro -o json | jq length

  # Short form
  shelly energy ls shelly-3em-pro`,
		Aliases:           []string{"ls"},
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// List EM components
	emIDs, err := svc.ListEMComponents(ctx, opts.Device)
	if err != nil {
		return fmt.Errorf("failed to list EM components: %w", err)
	}

	// List EM1 components
	em1IDs, err := svc.ListEM1Components(ctx, opts.Device)
	if err != nil {
		return fmt.Errorf("failed to list EM1 components: %w", err)
	}

	// Combine results
	components := make([]model.ComponentListItem, 0, len(emIDs)+len(em1IDs))
	for _, id := range emIDs {
		components = append(components, model.ComponentListItem{
			ID:   id,
			Type: "EM (3-phase)",
		})
	}
	for _, id := range em1IDs {
		components = append(components, model.ComponentListItem{
			ID:   id,
			Type: "EM1 (single-phase)",
		})
	}

	if len(components) == 0 {
		ios.NoResults("energy monitoring components")
		return nil
	}

	// Output results
	return cmdutil.PrintListResult(ios, components, func(ios *iostreams.IOStreams, items []model.ComponentListItem) {
		table := output.NewTable("ID", "Type")
		for _, comp := range items {
			table.AddRow(fmt.Sprintf("%d", comp.ID), comp.Type)
		}
		if err := table.PrintTo(ios.Out); err != nil {
			ios.DebugErr("print table", err)
		}
	})
}
