// Package set provides the rgbw set subcommand.
package set

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the rgbw set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewColorSetCommand(f, factories.ColorSetOpts{
		Component: "RGBW",
		HasWhite:  true,
		SetFunc: func(ctx context.Context, f *cmdutil.Factory, device string, id, red, green, blue, white, brightness int, on bool) error {
			params := shelly.BuildRGBWSetParams(red, green, blue, white, brightness, on)
			return f.ShellyService().RGBWSet(ctx, device, id, params)
		},
	})
}
