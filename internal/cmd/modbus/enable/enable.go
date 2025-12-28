// Package enable provides the modbus enable command.
package enable

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

// NewCommand creates the modbus enable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "enable <device>",
		Aliases: []string{"on"},
		Short:   "Enable Modbus-TCP server",
		Long: `Enable the Modbus-TCP server on a Shelly device.

When enabled, the device exposes Modbus registers on TCP port 502.`,
		Example: `  # Enable Modbus
  shelly modbus enable kitchen`,
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

	err := cmdutil.RunWithSpinner(ctx, ios, "Enabling Modbus-TCP...", func(ctx context.Context) error {
		return svc.SetConfig(ctx, opts.Device, true)
	})
	if err != nil {
		return err
	}

	ios.Success("Modbus-TCP enabled on port 502")
	return nil
}
