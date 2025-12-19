// Package shellcmd provides the device shell command.
package shellcmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
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
device's capabilities interactively.

Available commands:
  help           Show available commands
  info           Show device information
  status         Show device status
  config         Show device configuration
  methods        List available RPC methods
  <method>       Execute RPC method (e.g., Switch.GetStatus, Shelly.GetConfig)
  exit           Close shell

RPC methods can be called directly by typing the method name.
For methods requiring parameters, provide JSON after the method name.`,
		Example: `  # Open shell for a device
  shelly shell living-room

  # Example session:
  shell> info
  shell> methods
  shell> Switch.GetStatus {"id":0}
  shell> Shelly.GetConfig
  shell> exit`,
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
	conn, err := svc.Connect(ctx, opts.Device)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", opts.Device, err)
	}
	defer iostreams.CloseWithDebug("closing device shell connection", conn)

	info := conn.Info()
	ios.Println()
	ios.Println(theme.Bold().Render(fmt.Sprintf("Connected to %s", info.Model)))
	ios.Printf("  ID: %s | MAC: %s | FW: %s\n", info.ID, info.MAC, info.Firmware)
	ios.Info("Type 'help' for commands, 'methods' to list RPC methods, 'exit' to quit")
	ios.Println()

	session := term.NewShellSession(ios, conn, opts.Device)
	scanner := bufio.NewScanner(os.Stdin)
	prompt := term.FormatShellPrompt(opts.Device)

	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			ios.Println("\nSession terminated")
			return nil
		default:
		}

		// Print prompt
		if _, err := fmt.Fprint(ios.Out, prompt); err != nil {
			ios.DebugErr("failed to print prompt", err)
		}

		// Read input
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("error reading input: %w", err)
			}
			ios.Println("\nGoodbye!")
			return nil
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		shouldExit := session.ExecuteCommand(ctx, line)
		if shouldExit {
			ios.Println("Goodbye!")
			return nil
		}
	}
}
