// Package disable provides the modbus disable command.
package disable

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// Options holds command options.
type Options struct {
	Device  string
	Factory *cmdutil.Factory
}

// NewCommand creates the modbus disable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "disable <device>",
		Aliases: []string{"off"},
		Short:   "Disable Modbus-TCP server",
		Long:    `Disable the Modbus-TCP server on a Shelly device.`,
		Example: `  # Disable Modbus
  shelly modbus disable kitchen`,
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
	svc := opts.Factory.ModbusService()

	err := cmdutil.RunWithSpinner(ctx, ios, "Disabling Modbus-TCP...", func(ctx context.Context) error {
		return svc.SetConfig(ctx, opts.Device, false)
	})
	if err != nil {
		return err
	}

	ios.Success("Modbus-TCP disabled")
	return nil
}
