// Package disable provides the mqtt disable subcommand.
package disable

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the mqtt disable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "disable <device>",
		Aliases: []string{"off", "stop"},
		Short:   "Disable MQTT",
		Long: `Disable MQTT on a device.

This disconnects the device from the MQTT broker and disables MQTT functionality.`,
		Example: `  # Disable MQTT
  shelly mqtt disable living-room`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	enable := false
	return cmdutil.RunWithSpinner(ctx, ios, "Disabling MQTT...", func(ctx context.Context) error {
		if err := svc.SetMQTTConfig(ctx, device, &enable, "", "", "", ""); err != nil {
			return fmt.Errorf("failed to disable MQTT: %w", err)
		}
		ios.Success("MQTT disabled on %s", device)
		return nil
	})
}
