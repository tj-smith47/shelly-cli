// Package set provides the light set subcommand.
package set

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// NewCommand creates the light set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		lightID    int
		brightness int
		on         bool
	)

	cmd := &cobra.Command{
		Use:     "set <device>",
		Aliases: []string{"brightness", "br"},
		Short:   "Set light parameters",
		Long: `Set parameters of a light component on the specified device.

You can set brightness and on/off state.
Values not specified will be left unchanged.`,
		Example: `  # Set brightness to 50%
  shelly light set kitchen --brightness 50

  # Turn on and set brightness
  shelly light br kitchen -b 75 --on`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], lightID, brightness, on)
		},
	}

	flags.AddComponentIDFlag(cmd, &lightID, "Light")
	cmd.Flags().IntVarP(&brightness, "brightness", "b", -1, "Brightness (0-100)")
	cmd.Flags().BoolVar(&on, "on", false, "Turn on")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, lightID, brightness int, on bool) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	svc := f.ShellyService()
	ios := f.IOStreams()

	var brightnessPtr *int
	if brightness >= 0 && brightness <= 100 {
		brightnessPtr = &brightness
	}

	var onPtr *bool
	if on {
		onPtr = &on
	}

	err := cmdutil.RunWithSpinner(ctx, ios, "Setting light parameters...", func(ctx context.Context) error {
		return svc.LightSet(ctx, device, lightID, brightnessPtr, onPtr)
	})
	if err != nil {
		return fmt.Errorf("failed to set light parameters: %w", err)
	}

	ios.Success("Light %d parameters set", lightID)
	return nil
}
