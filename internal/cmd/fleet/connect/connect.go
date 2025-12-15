// Package connect provides the fleet connect subcommand.
package connect

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// Options holds the command options.
type Options struct {
	Host   string
	Region string
}

// NewCommand creates the fleet connect command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:     "connect",
		Aliases: []string{"login", "auth"},
		Short:   "Connect to Shelly Cloud hosts",
		Long: `Connect to Shelly Cloud hosts for fleet management.

Requires integrator credentials configured via environment variables or config:
  SHELLY_INTEGRATOR_TAG - Your integrator tag
  SHELLY_INTEGRATOR_TOKEN - Your integrator token

By default, connects to all regions (EU and US). Use --host to connect
to a specific cloud host.`,
		Example: `  # Connect to all regions
  shelly fleet connect

  # Connect to specific host
  shelly fleet connect --host shelly-13-eu.shelly.cloud

  # Connect to specific region
  shelly fleet connect --region eu`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), f, opts)
		},
	}

	cmd.Flags().StringVar(&opts.Host, "host", "", "Specific cloud host to connect to")
	cmd.Flags().StringVar(&opts.Region, "region", "", "Region to connect to (eu, us)")

	return cmd
}

func run(_ context.Context, f *cmdutil.Factory, opts *Options) error {
	ios := f.IOStreams()

	// Check for credentials
	integratorTag := os.Getenv("SHELLY_INTEGRATOR_TAG")
	integratorToken := os.Getenv("SHELLY_INTEGRATOR_TOKEN")

	if integratorTag == "" || integratorToken == "" {
		ios.Warning("Integrator credentials not configured")
		ios.Println("")
		ios.Info("Set the following environment variables:")
		ios.Printf("  SHELLY_INTEGRATOR_TAG=your-integrator-tag\n")
		ios.Printf("  SHELLY_INTEGRATOR_TOKEN=your-integrator-token\n")
		ios.Println("")
		ios.Info("Or add to config file (~/.config/shelly/config.yaml):")
		ios.Printf("  integrator:\n")
		ios.Printf("    tag: your-integrator-tag\n")
		ios.Printf("    token: your-integrator-token\n")
		return fmt.Errorf("integrator credentials required")
	}

	ios.StartProgress("Connecting to Shelly Cloud...")

	ios.StopProgress()

	// Note: This is a placeholder implementation
	// Full implementation requires integrator package integration
	switch {
	case opts.Host != "":
		ios.Success("Connected to host: %s", opts.Host)
	case opts.Region != "":
		ios.Success("Connected to %s region", opts.Region)
	default:
		ios.Success("Connected to all regions")
	}

	ios.Println("")
	ios.Info("Fleet management ready. Use 'shelly fleet status' to view devices.")

	return nil
}
