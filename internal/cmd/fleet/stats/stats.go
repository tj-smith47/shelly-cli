// Package stats provides the fleet stats subcommand.
package stats

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/integrator"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// NewCommand creates the fleet stats command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stats",
		Aliases: []string{"statistics", "summary"},
		Short:   "View fleet statistics",
		Long: `View aggregate statistics for your device fleet.

Requires an active fleet connection. Run 'shelly fleet connect' first.`,
		Example: `  # View fleet statistics
  shelly fleet stats

  # JSON output
  shelly fleet stats -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), f)
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory) error {
	ios := f.IOStreams()

	// Get credentials
	integratorTag := os.Getenv("SHELLY_INTEGRATOR_TAG")
	integratorToken := os.Getenv("SHELLY_INTEGRATOR_TOKEN")

	cfg, cfgErr := f.Config()
	if cfgErr != nil {
		ios.DebugErr("load config", cfgErr)
	}
	if cfg != nil {
		if integratorTag == "" {
			integratorTag = cfg.Integrator.Tag
		}
		if integratorToken == "" {
			integratorToken = cfg.Integrator.Token
		}
	}

	if integratorTag == "" || integratorToken == "" {
		return fmt.Errorf("integrator credentials required. Run 'shelly fleet connect' first")
	}

	// Create client and authenticate
	client := integrator.New(integratorTag, integratorToken)
	if err := client.Authenticate(ctx); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Create fleet manager and connect
	fm := integrator.NewFleetManager(client)

	ios.Info("Connecting to fleet...")
	errors := fm.ConnectAll(ctx, nil)
	if len(errors) > 0 {
		for host, err := range errors {
			ios.Warning("Failed to connect to %s: %v", host, err)
		}
	}

	// Get fleet statistics
	stats := fm.GetStats()

	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, stats)
	}

	ios.Success("Fleet Statistics")
	ios.Println()

	term.DisplayFleetStats(ios, stats)

	return nil
}
