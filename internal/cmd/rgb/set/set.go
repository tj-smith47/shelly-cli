// Package set provides the rgb set subcommand.
package set

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the rgb set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewColorSetCommand(f, factories.ColorSetOpts{
		Component: "RGB",
		HasWhite:  false,
		SetFunc: func(ctx context.Context, f *cmdutil.Factory, device string, id, red, green, blue, _, brightness int, on bool) error {
			params := shelly.BuildRGBSetParams(red, green, blue, brightness, on)
			return f.ShellyService().RGBSet(ctx, device, id, params)
		},
	})
}
