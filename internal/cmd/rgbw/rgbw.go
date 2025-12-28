// Package rgbw provides RGBW LED control commands.
package rgbw

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/rgbw/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/rgbw/off"
	"github.com/tj-smith47/shelly-cli/internal/cmd/rgbw/on"
	"github.com/tj-smith47/shelly-cli/internal/cmd/rgbw/set"
	"github.com/tj-smith47/shelly-cli/internal/cmd/rgbw/status"
	"github.com/tj-smith47/shelly-cli/internal/cmd/rgbw/toggle"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the rgbw command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rgbw",
		Aliases: []string{"rgbwled"},
		Short:   "Control RGBW LED outputs",
		Long: `Control RGBW LED outputs on Shelly devices.

RGBW components support RGB color channels plus a separate white channel,
providing more control than RGB-only outputs.`,
		Example: `  # Turn on RGBW with warm white
  shelly rgbw on kitchen --rgb 255,200,150 --white 128

  # Set color and brightness
  shelly rgbw set kitchen --rgb 255,0,0 --brightness 75

  # Toggle RGBW state
  shelly rgbw toggle kitchen`,
	}

	cmd.AddCommand(
		list.NewCommand(f),
		status.NewCommand(f),
		on.NewCommand(f),
		off.NewCommand(f),
		toggle.NewCommand(f),
		set.NewCommand(f),
	)

	return cmd
}
