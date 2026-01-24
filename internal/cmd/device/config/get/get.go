// Package get provides the config get subcommand.
package get

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Factory   *cmdutil.Factory
	Component string
	Device    string
}

// NewCommand creates the config get command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
			opts.Device = args[0]
			if len(args) > 1 {
				opts.Component = args[1]
			}
			return run(cmd.Context(), opts)
		},
	}

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	svc := opts.Factory.ShellyService()
	ios := opts.Factory.IOStreams()

	var config map[string]any
	err := cmdutil.RunWithSpinner(ctx, ios, "Getting configuration...", func(ctx context.Context) error {
		var err error
		config, err = svc.GetConfigAuto(ctx, opts.Device)
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to get configuration: %w", err)
	}

	// If a specific component was requested, extract just that
	var result any = config
	if opts.Component != "" {
		compConfig, ok := config[opts.Component]
		if !ok {
			return fmt.Errorf("component %q not found in configuration", opts.Component)
		}
		result = map[string]any{opts.Component: compConfig}
	}

	// Output based on format
	if output.WantsJSON() {
		return output.PrintJSON(result)
	}
	if output.WantsYAML() {
		return output.PrintYAML(result)
	}

	return term.DisplayConfigTable(ios, result)
}
