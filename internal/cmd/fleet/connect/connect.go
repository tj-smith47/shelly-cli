// Package connect provides the fleet connect subcommand.
package connect

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/integrator"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
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

func run(ctx context.Context, f *cmdutil.Factory, opts *Options) error {
	ios := f.IOStreams()

	// Get credentials
	cfg, cfgErr := f.Config()
	if cfgErr != nil {
		ios.DebugErr("load config", cfgErr)
	}
	tag, token, err := loadIntegratorCredentials(ios, cfg)
	if err != nil {
		return err
	}

	// Create integrator client and authenticate
	client := integrator.New(tag, token)
	ios.Info("Authenticating with Shelly Cloud...")
	if err := client.Authenticate(ctx); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Determine hosts to connect to
	hosts, err := determineHosts(opts)
	if err != nil {
		return err
	}

	// Connect to hosts
	fm := integrator.NewFleetManager(client)
	successCount, failCount := connectToHosts(ctx, ios, fm, hosts)

	// Report results
	ios.Println()
	if successCount > 0 {
		ios.Success("Connected to %d host(s)", successCount)
		if failCount > 0 {
			ios.Warning("%d host(s) failed to connect", failCount)
		}
		ios.Println()
		ios.Info("Fleet management ready. Use 'shelly fleet status' to view devices.")
		return nil
	}
	return fmt.Errorf("failed to connect to any hosts")
}

// loadIntegratorCredentials loads integrator credentials from environment or config.
func loadIntegratorCredentials(ios *iostreams.IOStreams, cfg *config.Config) (tag, token string, err error) {
	tag = os.Getenv("SHELLY_INTEGRATOR_TAG")
	token = os.Getenv("SHELLY_INTEGRATOR_TOKEN")

	// Try config if not in environment
	if cfg != nil {
		if tag == "" {
			tag = cfg.Integrator.Tag
		}
		if token == "" {
			token = cfg.Integrator.Token
		}
	}

	if tag == "" || token == "" {
		printCredentialHelp(ios)
		return "", "", fmt.Errorf("integrator credentials required")
	}
	return tag, token, nil
}

// printCredentialHelp prints help for configuring integrator credentials.
func printCredentialHelp(ios *iostreams.IOStreams) {
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
}

// determineHosts returns the list of hosts to connect to based on options.
func determineHosts(opts *Options) ([]string, error) {
	switch {
	case opts.Host != "":
		return []string{opts.Host}, nil
	case opts.Region != "":
		regionHosts, ok := cloudHosts[opts.Region]
		if !ok {
			return nil, fmt.Errorf("unknown region: %s (valid: eu, us)", opts.Region)
		}
		return regionHosts, nil
	default:
		var hosts []string
		for _, h := range cloudHosts {
			hosts = append(hosts, h...)
		}
		return hosts, nil
	}
}

// connectToHosts connects to the specified hosts and returns success/fail counts.
func connectToHosts(ctx context.Context, ios *iostreams.IOStreams, fm *integrator.FleetManager, hosts []string) (success, fail int) {
	connectOpts := &integrator.ConnectOptions{}
	for _, host := range hosts {
		ios.Info("Connecting to %s...", host)
		if _, err := fm.Connect(ctx, host, connectOpts); err != nil {
			ios.Warning("  Failed: %v", err)
			fail++
		} else {
			ios.Success("  Connected")
			success++
		}
	}
	return success, fail
}
