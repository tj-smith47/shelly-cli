// Package enable provides the input enable command.
package enable

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// Options holds command options.
type Options struct {
	flags.ComponentFlags
	Factory *cmdutil.Factory
	Device  string
}

// NewCommand creates the input enable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "enable <device>",
		Aliases: []string{"on"},
		Short:   "Enable input component",
		Long: `Enable an input component on a Shelly device.

When enabled, the input will respond to physical button presses or switch
state changes and trigger associated actions.`,
		Example: `  # Enable input on a device
  shelly input enable kitchen

  # Enable specific input by ID
  shelly input enable living-room --id 1

  # Using alias
  shelly input on bedroom`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	flags.AddComponentFlags(cmd, &opts.ComponentFlags, "Input")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	f := opts.Factory
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	// Get current config to preserve other settings
	cfg, err := svc.InputGetConfig(ctx, opts.Device, opts.ID)
	if err != nil {
		return fmt.Errorf("failed to get input config: %w", err)
	}

	// Check if already enabled
	if cfg.Enable {
		ios.Info("Input %d is already enabled", opts.ID)
		return nil
	}

	// Set enable to true while preserving other settings
	cfg.Enable = true

	err = cmdutil.RunWithSpinner(ctx, ios, "Enabling input...", func(ctx context.Context) error {
		return svc.InputSetConfig(ctx, opts.Device, opts.ID, cfg)
	})
	if err != nil {
		return fmt.Errorf("failed to enable input: %w", err)
	}

	ios.Success("Input %d enabled", opts.ID)
	return nil
}
