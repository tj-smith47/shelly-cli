// Package reset provides the config reset subcommand.
package reset

import (
	"context"
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	flags.ConfirmFlags
	Factory   *cmdutil.Factory
	Device    string
	Component string
}

// NewCommand creates the config reset command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "reset <device> [component]",
		Aliases: []string{"factory", "clear"},
		Short:   "Reset configuration to defaults",
		Long: `Reset device or component configuration to factory defaults.

Without a component argument, shows available components that can be reset.
With a component argument, resets that component's configuration.

Note: This does not perform a full factory reset. For that, use:
  shelly device factory-reset <device>`,
		Example: `  # Reset switch:0 to defaults
  shelly config reset living-room switch:0

  # Reset with confirmation skipped
  shelly config reset living-room switch:0 --yes`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			if len(args) > 1 {
				opts.Component = args[1]
			}
			return run(cmd.Context(), opts)
		},
	}

	flags.AddYesOnlyFlag(cmd, &opts.ConfirmFlags)

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	svc := opts.Factory.ShellyService()
	ios := opts.Factory.IOStreams()

	if opts.Component == "" {
		// Show available components
		var config map[string]any
		err := cmdutil.RunWithSpinner(ctx, ios, "Getting device components...", func(ctx context.Context) error {
			var getErr error
			config, getErr = svc.GetConfig(ctx, opts.Device)
			return getErr
		})
		if err != nil {
			return fmt.Errorf("failed to get device configuration: %w", err)
		}

		keys := make([]string, 0, len(config))
		for key := range config {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		term.DisplayResetableComponents(ios, opts.Device, keys)
		return nil
	}

	// Confirm reset
	confirmed, err := opts.Factory.ConfirmAction(
		fmt.Sprintf("Reset %s configuration on %s to defaults?", opts.Component, opts.Device),
		opts.Yes,
	)
	if err != nil {
		return err
	}
	if !confirmed {
		ios.Warning("Reset cancelled")
		return nil
	}

	// Reset by setting config to empty/defaults
	err = cmdutil.RunWithSpinner(ctx, ios, "Resetting configuration...", func(ctx context.Context) error {
		return svc.SetComponentConfig(ctx, opts.Device, opts.Component, map[string]any{})
	})
	if err != nil {
		return fmt.Errorf("failed to reset configuration: %w", err)
	}

	ios.Success("Configuration reset for %s on %s", opts.Component, opts.Device)
	return nil
}
