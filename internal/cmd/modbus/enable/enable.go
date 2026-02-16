// Package enable provides the modbus enable command.
package enable

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
)

// NewCommand creates the modbus enable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewEnableDisableCommand(f, factories.EnableDisableOpts{
		Feature:       "Modbus-TCP",
		Enable:        true,
		ExampleParent: "modbus",
		Long: `Enable the Modbus-TCP server on a Shelly device.

When enabled, the device exposes Modbus registers on TCP port 502.`,
		Example: `  # Enable Modbus
  shelly modbus enable kitchen`,
		ServiceFunc: func(ctx context.Context, f *cmdutil.Factory, device string) error {
			return f.ModbusService().SetConfig(ctx, device, true)
		},
		PostSuccess: func(f *cmdutil.Factory, _ string) {
			f.IOStreams().Info("Listening on port 502")
		},
	})
}
