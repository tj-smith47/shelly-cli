// Package reset provides the config reset subcommand.
package reset

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
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
	ios := f.IOStreams()

	if component == "" {
		// Show available components
		return showComponents(f, ctx, device)
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

	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	svc := f.ShellyService()

	ios.StartProgress("Resetting configuration...")

	// Reset by setting config to empty/defaults
	// Note: The actual reset behavior depends on the component type
	// For now, we use a raw RPC call if available, or set to empty config
	err = resetComponent(ctx, svc, device, component)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("failed to reset configuration: %w", err)
	}

	ios.Success("Configuration reset for %s on %s", component, device)
	return nil
}

// showComponents lists available components that can be reset.
func showComponents(f *cmdutil.Factory, ctx context.Context, device string) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	svc := f.ShellyService()
	ios := f.IOStreams()

	ios.StartProgress("Getting device components...")

	config, err := svc.GetConfig(ctx, device)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("failed to get device configuration: %w", err)
	}

	ios.Title("Available components")
	ios.Printf("Specify a component to reset its configuration:\n")
	ios.Printf("\n")

	for key := range config {
		ios.Printf("  shelly config reset %s %s\n", device, key)
	}

	return nil
}

// resetComponent resets a specific component's configuration.
func resetComponent(ctx context.Context, svc *shelly.Service, device, component string) error {
	// For most components, setting an empty config or specific defaults works
	// This is a simplified implementation - some components may need special handling
	return svc.SetComponentConfig(ctx, device, component, map[string]any{})
}
