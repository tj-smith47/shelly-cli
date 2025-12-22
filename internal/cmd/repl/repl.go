// Package repl provides the interactive REPL command.
package repl

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds the command options.
type Options struct {
	Factory  *cmdutil.Factory
	Device   string
	NoPrompt bool
}

// NewCommand creates the interactive command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:     "repl",
		Aliases: []string{"interactive", "i"},
		Short:   "Launch interactive REPL",
		Long: `Launch an interactive REPL (Read-Eval-Print Loop) for Shelly CLI.

This provides a command-line shell where you can enter Shelly commands
without prefixing them with 'shelly'. It supports command history and
readline-style line editing (arrow keys, Ctrl+A/E, etc.).

Available commands in REPL:
  help             Show available commands
  devices          List registered devices
  connect <device> Set active device for subsequent commands
  disconnect       Clear active device
  status           Show status of active device
  on               Turn on all device components
  off              Turn off all device components
  toggle           Toggle all device components
  rpc <method>     Execute raw RPC call on active device
  exit, quit, q    Exit the REPL

To control a specific component (e.g., switch ID 1 on a multi-switch device),
use the rpc command with the appropriate method and parameters.`,
		Example: `  # Start REPL mode
  shelly repl

  # Start with a default device
  shelly repl --device living-room

  # Example session:
  > devices
  > connect living-room
  > status
  > on
  > exit

  # Control specific switch on multi-switch device (after connecting):
  > connect dual-switch
  > rpc Switch.Set {"id":1,"on":true}
  > rpc Switch.Toggle {"id":0}

  # Or without connecting first (device as first arg):
  > rpc dual-switch Switch.Set {"id":1,"on":true}`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Device, "device", "d", "", "Default device to connect to")
	cmd.Flags().BoolVar(&opts.NoPrompt, "no-prompt", false, "Disable interactive prompt (for scripting)")

	return cmd
}

// run executes the interactive REPL.
func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	if !ios.IsStdinTTY() && !opts.NoPrompt {
		ios.Info("Running in non-interactive mode (stdin is not a TTY)")
	}

	ios.Println(theme.Bold().Render("Shelly Interactive REPL"))
	ios.Info("Type 'help' for available commands, 'exit' to quit")
	ios.Println()

	svc := opts.Factory.ShellyService()
	session := term.NewREPLSession(ios, svc, opts.Device)

	return session.RunREPLLoop(ctx)
}
