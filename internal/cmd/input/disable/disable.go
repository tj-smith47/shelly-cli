// Package disable provides the input disable command.
package disable

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

// NewCommand creates the input disable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "disable <device>",
		Aliases: []string{"off"},
		Short:   "Disable input component",
		Long: `Disable an input component on a Shelly device.

When disabled, the input will not respond to physical button presses or
switch state changes. This can be useful for maintenance or to prevent
accidental triggers.`,
		Example: `  # Disable input on a device
  shelly input disable kitchen

  # Disable specific input by ID
  shelly input disable living-room --id 1

  # Using alias
  shelly input off bedroom`,
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

	// Check if already disabled
	if !cfg.Enable {
		ios.Info("Input %d is already disabled", opts.ID)
		return nil
	}

	// Set enable to false while preserving other settings
	cfg.Enable = false

	err = cmdutil.RunWithSpinner(ctx, ios, "Disabling input...", func(ctx context.Context) error {
		return svc.InputSetConfig(ctx, opts.Device, opts.ID, cfg)
	})
	if err != nil {
		return fmt.Errorf("failed to disable input: %w", err)
	}

	ios.Success("Input %d disabled", opts.ID)
	return nil
}
