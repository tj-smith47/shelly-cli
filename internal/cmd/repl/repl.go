// Package repl provides the interactive REPL command.
package repl

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

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
without prefixing them with 'shelly'. It supports command history within
the session and provides a more interactive experience.

Available commands in REPL:
  help             Show available commands
  devices          List registered devices
  connect <device> Set active device for subsequent commands
  disconnect       Clear active device
  status           Show status of active device
  on               Turn on active device
  off              Turn off active device
  toggle           Toggle active device
  rpc <method>     Execute raw RPC call on active device
  exit, quit, q    Exit the REPL

You can also run any shelly subcommand by typing it directly.`,
		Example: `  # Start REPL mode
  shelly repl

  # Start with a default device
  shelly repl --device living-room

  # Example session:
  > devices
  > connect living-room
  > status
  > on
  > exit`,
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

	scanner := bufio.NewScanner(os.Stdin)

	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			ios.Println("\nSession terminated")
			return nil
		default:
		}

		// Print prompt
		prompt := term.FormatREPLPrompt(session.ActiveDevice)
		if _, err := fmt.Fprint(ios.Out, prompt); err != nil {
			ios.DebugErr("failed to print prompt", err)
		}

		// Read input
		if !scanner.Scan() {
			// EOF or error
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

		// Parse and execute command
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		cmd := strings.ToLower(parts[0])
		args := parts[1:]

		shouldExit := session.ExecuteCommand(ctx, cmd, args)
		if shouldExit {
			ios.Println("Goodbye!")
			return nil
		}
	}
}
