// Package rgb provides the rgb command and its subcommands.
package rgb

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/rgb/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/rgb/off"
	"github.com/tj-smith47/shelly-cli/internal/cmd/rgb/on"
	rgbset "github.com/tj-smith47/shelly-cli/internal/cmd/rgb/set"
	"github.com/tj-smith47/shelly-cli/internal/cmd/rgb/status"
	"github.com/tj-smith47/shelly-cli/internal/cmd/rgb/toggle"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the rgb command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rgb",
		Aliases: []string{"color"},
		Short:   "Control RGB light components",
		Long:    `Control RGB light components on Shelly devices.`,
		Example: `  # Turn on RGB light
  shelly rgb on living-room

  # Set RGB color to red
  shelly rgb set living-room --red 255 --green 0 --blue 0

  # Check RGB status
  shelly rgb status living-room`,
	}

	cmd.AddCommand(on.NewCommand(f))
	cmd.AddCommand(off.NewCommand(f))
	cmd.AddCommand(toggle.NewCommand(f))
	cmd.AddCommand(status.NewCommand(f))
	cmd.AddCommand(rgbset.NewCommand(f))
	cmd.AddCommand(list.NewCommand(f))

	return cmd
}
