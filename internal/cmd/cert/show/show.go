// Package show provides the cert show subcommand.
package show

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// NewCommand creates the cert show command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show <device>",
		Aliases: []string{"info", "view", "get"},
		Short:   "Show device TLS configuration",
		Long:    `Display TLS certificate configuration for a Gen2+ device.`,
		Example: `  # Show TLS config
  shelly cert show kitchen`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	ios.StartProgress("Fetching TLS configuration...")

	conn, err := svc.Connect(ctx, device)
	if err != nil {
		ios.StopProgress()
		return fmt.Errorf("connect: %w", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			ios.DebugErr("close connection", err)
		}
	}()

	// Get full config to check TLS settings
	result, err := conn.Call(ctx, "Shelly.GetConfig", nil)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	config, ok := result.(map[string]any)
	if !ok {
		return fmt.Errorf("unexpected response type")
	}

	ios.Success("TLS Configuration for %s", device)
	ios.Println("")

	hasCustomCA := term.DisplayTLSConfig(ios, config)

	// Show guidance if no custom CA is configured
	if !hasCustomCA {
		ios.Println("")
		ios.Info("Use 'shelly cert install' to add a custom CA certificate")
	}

	// Show raw TLS-related config for debugging
	if viper.GetBool("verbose") {
		ios.Println("")
		ios.Info("Raw configuration:")
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			ios.DebugErr("marshal config", err)
		} else {
			ios.Println(string(data))
		}
	}

	return nil
}
