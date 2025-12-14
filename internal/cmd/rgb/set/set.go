// Package set provides the rgb set subcommand.
package set

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the rgb set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		rgbID      int
		red        int
		green      int
		blue       int
		brightness int
		on         bool
	)

	cmd := &cobra.Command{
		Use:   "set <device>",
		Short: "Set RGB parameters",
		Long: `Set parameters of an RGB light component on the specified device.

You can set color values (red, green, blue), brightness, and on/off state.
Values not specified will be left unchanged.`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], rgbID, red, green, blue, brightness, on)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &rgbID, "RGB")
	cmd.Flags().IntVarP(&red, "red", "r", -1, "Red value (0-255)")
	cmd.Flags().IntVarP(&green, "green", "g", -1, "Green value (0-255)")
	cmd.Flags().IntVarP(&blue, "blue", "b", -1, "Blue value (0-255)")
	cmd.Flags().IntVar(&brightness, "brightness", -1, "Brightness (0-100)")
	cmd.Flags().BoolVar(&on, "on", false, "Turn on")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, rgbID, red, green, blue, brightness int, on bool) error {
	params := buildParams(red, green, blue, brightness, on)

	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	svc := f.ShellyService()
	ios := f.IOStreams()

	ios.StartProgress("Setting RGB parameters...")

	err := svc.RGBSet(ctx, device, rgbID, params)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("failed to set RGB parameters: %w", err)
	}

	ios.Success("RGB %d parameters set", rgbID)
	return nil
}

func buildParams(red, green, blue, brightness int, on bool) shelly.RGBSetParams {
	params := shelly.RGBSetParams{}

	// Color values are valid from 0-255, -1 means not set
	if red >= 0 && red <= 255 {
		params.Red = &red
	}
	if green >= 0 && green <= 255 {
		params.Green = &green
	}
	if blue >= 0 && blue <= 255 {
		params.Blue = &blue
	}

	// Brightness is valid from 0-100
	if brightness >= 0 && brightness <= 100 {
		params.Brightness = &brightness
	}

	// Only set on if explicitly requested
	if on {
		params.On = &on
	}

	return params
}
