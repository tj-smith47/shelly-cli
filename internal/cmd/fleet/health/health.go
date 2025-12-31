// Package health provides the fleet health subcommand.
package health

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/integrator"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Factory   *cmdutil.Factory
	Threshold time.Duration
}

// NewCommand creates the fleet health command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory:   f,
		Threshold: 10 * time.Minute,
	}

	cmd := &cobra.Command{
		Use:     "health",
		Aliases: []string{"check", "diagnose"},
		Short:   "Check device health",
		Long: `Check the health status of devices in your fleet.

Reports devices that haven't been seen recently or have frequent
online/offline transitions indicating connectivity issues.

Requires an active fleet connection. Run 'shelly fleet connect' first.`,
		Example: `  # Check device health
  shelly fleet health

  # Custom threshold for "unhealthy"
  shelly fleet health --threshold 30m

  # JSON output
  shelly fleet health -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().DurationVar(&opts.Threshold, "threshold", 10*time.Minute, "Time threshold for unhealthy status")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Get credentials
	integratorTag := os.Getenv("SHELLY_INTEGRATOR_TAG")
	integratorToken := os.Getenv("SHELLY_INTEGRATOR_TOKEN")

	if cfg := cmdutil.SafeConfig(opts.Factory); cfg != nil {
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

	// Get health data from health monitor
	healthMonitor := fm.HealthMonitor()
	healthData := healthMonitor.ListDeviceHealth()

	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, healthData)
	}

	if len(healthData) == 0 {
		ios.Warning("No health data available")
		ios.Info("Device health is tracked over time. Try again after devices have been connected.")
		return nil
	}

	ios.Success("Fleet Health Report")
	ios.Println()

	term.DisplayFleetHealth(ios, healthData, opts.Threshold)

	return nil
}
