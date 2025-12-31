// Package set provides the config set subcommand.
package set

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds the command options.
type Options struct {
	Factory   *cmdutil.Factory
	Component string
	Device    string
	KeyValues []string
}

// NewCommand creates the config set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "set <device> <component> <key>=<value>...",
		Aliases: []string{"write", "update"},
		Short:   "Set device configuration",
		Long: `Set configuration values for a device component.

Specify key=value pairs to update. Only the specified keys will be modified.`,
		Example: `  # Set switch name
  shelly config set living-room switch:0 name="Main Light"

  # Set multiple values
  shelly config set living-room switch:0 name="Light" initial_state=on

  # Set light brightness default
  shelly config set living-room light:0 default.brightness=50`,
		Args: cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.Component = args[1]
			opts.KeyValues = args[2:]
			return run(cmd.Context(), opts)
		},
	}

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	svc := opts.Factory.ShellyService()
	ios := opts.Factory.IOStreams()

	// Parse key=value pairs
	cfg := make(map[string]any)
	for _, kv := range opts.KeyValues {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid key=value format: %s", kv)
		}
		key := parts[0]
		value := config.ParseValue(parts[1])
		cfg[key] = value
	}

	err := cmdutil.RunWithSpinner(ctx, ios, "Setting configuration...", func(ctx context.Context) error {
		return svc.SetComponentConfig(ctx, opts.Device, opts.Component, cfg)
	})
	if err != nil {
		return fmt.Errorf("failed to set configuration: %w", err)
	}

	ios.Success("Configuration updated for %s on %s", opts.Component, opts.Device)
	return nil
}
