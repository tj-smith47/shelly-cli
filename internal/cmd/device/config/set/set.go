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

// NewCommand creates the config set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
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
			device := args[0]
			component := args[1]
			keyValues := args[2:]
			return run(cmd.Context(), f, device, component, keyValues)
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device, component string, keyValues []string) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	svc := f.ShellyService()
	ios := f.IOStreams()

	// Parse key=value pairs
	cfg := make(map[string]any)
	for _, kv := range keyValues {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid key=value format: %s", kv)
		}
		key := parts[0]
		value := config.ParseValue(parts[1])
		cfg[key] = value
	}

	ios.StartProgress("Setting configuration...")

	err := svc.SetComponentConfig(ctx, device, component, cfg)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("failed to set configuration: %w", err)
	}

	ios.Success("Configuration updated for %s on %s", component, device)
	return nil
}
