// Package set provides the light set subcommand.
package set

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the light set command.
func NewCommand() *cobra.Command {
	var (
		lightID    int
		brightness int
		on         bool
	)

	cmd := &cobra.Command{
		Use:   "set <device>",
		Short: "Set light parameters",
		Long: `Set parameters of a light component on the specified device.

You can set brightness and on/off state.
Values not specified will be left unchanged.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), args[0], lightID, brightness, on)
		},
	}

	cmd.Flags().IntVarP(&lightID, "id", "i", 0, "Light ID (default 0)")
	cmd.Flags().IntVarP(&brightness, "brightness", "b", -1, "Brightness (0-100)")
	cmd.Flags().BoolVar(&on, "on", false, "Turn on")

	return cmd
}

func run(ctx context.Context, device string, lightID, brightness int, on bool) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	svc := shelly.NewService()

	spin := iostreams.NewSpinner("Setting light parameters...")
	spin.Start()

	var brightnessPtr *int
	if brightness >= 0 && brightness <= 100 {
		brightnessPtr = &brightness
	}

	var onPtr *bool
	if on {
		onPtr = &on
	}

	err := svc.LightSet(ctx, device, lightID, brightnessPtr, onPtr)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to set light parameters: %w", err)
	}

	iostreams.Success("Light %d parameters set", lightID)
	return nil
}
