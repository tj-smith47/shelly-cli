// Package get provides the config get subcommand.
package get

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// NewCommand creates the config get command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get <device> [component]",
		Aliases: []string{"show", "read"},
		Short:   "Get device configuration",
		Long: `Get configuration for a device or specific component.

Without a component argument, returns the full device configuration.
With a component argument (e.g., "switch:0", "sys", "wifi"), returns
only that component's configuration.`,
		Example: `  # Get full device configuration
  shelly config get living-room

  # Get switch:0 configuration
  shelly config get living-room switch:0

  # Get system configuration
  shelly config get living-room sys

  # Get WiFi configuration
  shelly config get living-room wifi

  # Output as JSON
  shelly config get living-room -o json

  # Output as YAML
  shelly config get living-room -o yaml`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			device := args[0]
			component := ""
			if len(args) > 1 {
				component = args[1]
			}
			return run(cmd.Context(), f, device, component)
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device, component string) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	svc := f.ShellyService()
	ios := f.IOStreams()

	ios.StartProgress("Getting configuration...")

	config, err := svc.GetConfig(ctx, device)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("failed to get configuration: %w", err)
	}

	// If a specific component was requested, extract just that
	var result any = config
	if component != "" {
		compConfig, ok := config[component]
		if !ok {
			return fmt.Errorf("component %q not found in configuration", component)
		}
		result = map[string]any{component: compConfig}
	}

	// Output based on format
	if output.WantsJSON() {
		return output.PrintJSON(result)
	}
	if output.WantsYAML() {
		return output.PrintYAML(result)
	}

	return cmdutil.DisplayConfigTable(ios, result)
}
