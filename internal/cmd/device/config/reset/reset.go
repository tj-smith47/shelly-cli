// Package reset provides the config reset subcommand.
package reset

import (
	"context"
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

var yesFlag bool

// NewCommand creates the config reset command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "reset <device> [component]",
		Aliases: []string{"factory", "clear"},
		Short:   "Reset configuration to defaults",
		Long: `Reset device or component configuration to factory defaults.

Without a component argument, shows available components that can be reset.
With a component argument, resets that component's configuration.

Note: This does not perform a full factory reset. For that, use:
  shelly device factory-reset <device>`,
		Example: `  # Reset switch:0 to defaults
  shelly config reset living-room switch:0

  # Reset with confirmation skipped
  shelly config reset living-room switch:0 --yes`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			device := args[0]
			component := ""
			if len(args) > 1 {
				component = args[1]
			}
			return run(cmd.Context(), f, device, component)
		},
	}

	cmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device, component string) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	svc := f.ShellyService()
	ios := f.IOStreams()

	if component == "" {
		// Show available components
		var config map[string]any
		err := cmdutil.RunWithSpinner(ctx, ios, "Getting device components...", func(ctx context.Context) error {
			var getErr error
			config, getErr = svc.GetConfig(ctx, device)
			return getErr
		})
		if err != nil {
			return fmt.Errorf("failed to get device configuration: %w", err)
		}

		keys := make([]string, 0, len(config))
		for key := range config {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		term.DisplayResetableComponents(ios, device, keys)
		return nil
	}

	// Confirm reset
	confirmed, err := f.ConfirmAction(
		fmt.Sprintf("Reset %s configuration on %s to defaults?", component, device),
		yesFlag,
	)
	if err != nil {
		return err
	}
	if !confirmed {
		ios.Warning("Reset cancelled")
		return nil
	}

	// Reset by setting config to empty/defaults
	err = cmdutil.RunWithSpinner(ctx, ios, "Resetting configuration...", func(ctx context.Context) error {
		return svc.SetComponentConfig(ctx, device, component, map[string]any{})
	})
	if err != nil {
		return fmt.Errorf("failed to reset configuration: %w", err)
	}

	ios.Success("Configuration reset for %s on %s", component, device)
	return nil
}
