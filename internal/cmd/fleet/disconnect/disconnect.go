// Package disconnect provides the fleet disconnect subcommand.
package disconnect

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
}

// NewCommand creates the fleet disconnect command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "disconnect",
		Aliases: []string{"logout", "close"},
		Short:   "Disconnect from Shelly Cloud hosts",
		Long: `Disconnect from all connected Shelly Cloud hosts.

This command closes any active WebSocket connections to Shelly Cloud hosts.
Note: In CLI mode, connections are typically ephemeral per command. This
command is useful for explicitly verifying connectivity and cleanup.

Requires integrator credentials configured via environment variables or config:
  SHELLY_INTEGRATOR_TAG - Your integrator tag
  SHELLY_INTEGRATOR_TOKEN - Your integrator token`,
		Example: `  # Disconnect from all hosts
  shelly fleet disconnect`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Get credentials
	cfg, cfgErr := opts.Factory.Config()
	if cfgErr != nil {
		ios.DebugErr("load config", cfgErr)
	}

	creds, err := shelly.GetIntegratorCredentials(ios, cfg)
	if err != nil {
		return err
	}

	// Connect so we can disconnect
	conn, err := shelly.ConnectFleet(ctx, ios, creds)
	if err != nil {
		return err
	}

	// Get stats before disconnect
	stats := conn.Manager.GetStats()

	// Disconnect from all hosts
	ios.Info("Disconnecting from Shelly Cloud...")
	conn.Close()

	if stats.TotalConnections > 0 {
		ios.Success("Disconnected from %d host(s)", stats.TotalConnections)
	} else {
		ios.Info("No active connections to disconnect")
	}

	return nil
}
