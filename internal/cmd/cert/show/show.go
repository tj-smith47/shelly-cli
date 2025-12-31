// Package show provides the cert show subcommand.
package show

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
}

// NewCommand creates the cert show command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "show <device>",
		Aliases: []string{"info", "view", "get"},
		Short:   "Show device TLS configuration",
		Long:    `Display TLS certificate configuration for a Gen2+ device.`,
		Example: `  # Show TLS config
  shelly cert show kitchen`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	var config map[string]any
	err := cmdutil.RunWithSpinner(ctx, ios, "Fetching TLS configuration...", func(ctx context.Context) error {
		return svc.WithDevice(ctx, opts.Device, func(dev *shelly.DeviceClient) error {
			if dev.IsGen1() {
				return fmt.Errorf("TLS configuration is only supported on Gen2+ devices")
			}

			conn := dev.Gen2()

			result, callErr := conn.Call(ctx, "Shelly.GetConfig", nil)
			if callErr != nil {
				return fmt.Errorf("get config: %w", callErr)
			}

			// Handle different result types from RPC client
			switch v := result.(type) {
			case map[string]any:
				config = v
			case json.RawMessage:
				if err := json.Unmarshal(v, &config); err != nil {
					return fmt.Errorf("unmarshal config: %w", err)
				}
			default:
				return fmt.Errorf("unexpected response type: %T", result)
			}

			return nil
		})
	})
	if err != nil {
		return err
	}

	ios.Success("TLS Configuration for %s", opts.Device)
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
