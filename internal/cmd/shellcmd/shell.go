// Package shellcmd provides the device shell command.
package shellcmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
}

// NewCommand creates the shell command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:     "shell <device>",
		Aliases: []string{"sh", "console"},
		Short:   "Interactive shell for a specific device",
		Long: `Open an interactive shell for a specific Shelly device.

This provides direct access to execute RPC commands on the device.
It maintains a persistent connection and allows you to explore the
device's capabilities interactively. Supports readline-style line
editing (arrow keys, Ctrl+A/E, etc.) with command history.

Available commands:
  help           Show available commands
  info           Show device information
  status         Show device status
  config         Show device configuration
  methods        List available RPC methods
  components     List device components
  <method>       Execute RPC method (e.g., Switch.GetStatus, Shelly.GetConfig)
  exit           Close shell

RPC methods can be called directly by typing the method name.
For methods requiring parameters, provide JSON after the method name.

For multi-switch devices, use the RPC methods with component ID to
control specific switches.`,
		Example: `  # Open shell for a device
  shelly shell living-room

  # Example session:
  shell> info
  shell> methods
  shell> Switch.GetStatus {"id":0}
  shell> Shelly.GetConfig
  shell> exit

  # Control specific switch on multi-switch device:
  shell> Switch.Set {"id":1,"on":true}
  shell> Switch.Toggle {"id":0}
  shell> Switch.GetStatus {"id":1}`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	return cmd
}

// run executes the device shell.
func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Connect to device
	ios.Info("Connecting to %s...", opts.Device)

	return svc.WithDevice(ctx, opts.Device, func(dev *shelly.DeviceClient) error {
		if dev.IsGen1() {
			return fmt.Errorf("interactive shell is only supported on Gen2+ devices")
		}

		conn := dev.Gen2()
		info := conn.Info()
		ios.Println()
		ios.Println(theme.Bold().Render(fmt.Sprintf("Connected to %s", info.Model)))
		ios.Printf("  ID: %s | MAC: %s | FW: %s\n", info.ID, info.MAC, info.Firmware)
		ios.Info("Type 'help' for commands, 'methods' to list RPC methods, 'exit' to quit")
		ios.Println()

		session := term.NewShellSession(ios, conn, opts.Device)

		return session.RunShellLoop(ctx)
	})
}
