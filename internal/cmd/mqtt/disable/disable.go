// Package disable provides the mqtt disable subcommand.
package disable

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
}

// NewCommand creates the mqtt disable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "disable <device>",
		Aliases: []string{"off", "stop"},
		Short:   "Disable MQTT",
		Long: `Disable MQTT on a device.

This disconnects the device from the MQTT broker and disables MQTT functionality.`,
		Example: `  # Disable MQTT
  shelly mqtt disable living-room`,
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
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	enable := false
	return cmdutil.RunWithSpinner(ctx, ios, "Disabling MQTT...", func(ctx context.Context) error {
		if err := svc.SetMQTTConfig(ctx, opts.Device, &enable, "", "", "", ""); err != nil {
			return fmt.Errorf("failed to disable MQTT: %w", err)
		}
		ios.Success("MQTT disabled on %s", opts.Device)
		return nil
	})
}
