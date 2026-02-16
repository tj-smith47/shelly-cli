// Package disable provides the modbus disable command.
package disable

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
)

// NewCommand creates the modbus disable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewEnableDisableCommand(f, factories.EnableDisableOpts{
		Feature:       "Modbus-TCP",
		Enable:        false,
		ExampleParent: "modbus",
		Long:          `Disable the Modbus-TCP server on a Shelly device.`,
		Example: `  # Disable Modbus
  shelly modbus disable kitchen`,
		ServiceFunc: func(ctx context.Context, f *cmdutil.Factory, device string) error {
			return f.ModbusService().SetConfig(ctx, device, false)
		},
	})
}
