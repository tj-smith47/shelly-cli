// Package connect provides the fleet connect subcommand.
package connect

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/integrator"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// Options holds the command options.
type Options struct {
	Host   string
	Region string
}

// Cloud hosts by region.
var cloudHosts = map[string][]string{
	"eu": {"shelly-13-eu.shelly.cloud", "shelly-14-eu.shelly.cloud"},
	"us": {"shelly-15-us.shelly.cloud", "shelly-16-us.shelly.cloud"},
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

func run(ctx context.Context, f *cmdutil.Factory, opts *Options) error { //nolint:gocyclo // will fix soon
	ios := f.IOStreams()

	// Get credentials from environment or config
	integratorTag := os.Getenv("SHELLY_INTEGRATOR_TAG")
	integratorToken := os.Getenv("SHELLY_INTEGRATOR_TOKEN")

	// Try config if not in environment
	if integratorTag == "" || integratorToken == "" { //nolint:nestif // will fix soon
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
	}

	if integratorTag == "" || integratorToken == "" {
		ios.Warning("Integrator credentials not configured")
		ios.Println()
		ios.Info("Set the following environment variables:")
		ios.Printf("  export SHELLY_INTEGRATOR_TAG=your-integrator-tag\n")
		ios.Printf("  export SHELLY_INTEGRATOR_TOKEN=your-integrator-token\n")
		ios.Println()
		ios.Info("Or add to config file (~/.config/shelly/config.yaml):")
		ios.Printf("  integrator:\n")
		ios.Printf("    tag: your-integrator-tag\n")
		ios.Printf("    token: your-integrator-token\n")
		return fmt.Errorf("integrator credentials required")
	}

	// Create integrator client
	client := integrator.New(integratorTag, integratorToken)

	// Authenticate first
	ios.Info("Authenticating with Shelly Cloud...")
	if err := client.Authenticate(ctx); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Create fleet manager
	fm := integrator.NewFleetManager(client)

	// Determine hosts to connect to
	var hosts []string
	switch {
	case opts.Host != "":
		hosts = []string{opts.Host}
	case opts.Region != "":
		regionHosts, ok := cloudHosts[opts.Region]
		if !ok {
			return fmt.Errorf("unknown region: %s (valid: eu, us)", opts.Region)
		}
		hosts = regionHosts
	default:
		// Connect to all regions
		for _, h := range cloudHosts {
			hosts = append(hosts, h...)
		}
	}

	// Connect to hosts
	connectOpts := &integrator.ConnectOptions{}
	var successCount, failCount int

	for _, host := range hosts {
		ios.Info("Connecting to %s...", host)
		if _, err := fm.Connect(ctx, host, connectOpts); err != nil {
			ios.Warning("  Failed: %v", err)
			failCount++
		} else {
			ios.Success("  Connected")
			successCount++
		}
	}

	ios.Println()
	if successCount > 0 {
		ios.Success("Connected to %d host(s)", successCount)
		if failCount > 0 {
			ios.Warning("%d host(s) failed to connect", failCount)
		}
		ios.Println()
		ios.Info("Fleet management ready. Use 'shelly fleet status' to view devices.")
	} else {
		return fmt.Errorf("failed to connect to any hosts")
	}

	return nil
}
