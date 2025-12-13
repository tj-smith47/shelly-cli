// Package cover provides the cover command and its subcommands.
package cover

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/cover/calibrate"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cover/closecmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cover/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cover/open"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cover/position"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cover/status"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cover/stop"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the cover command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cover",
		Aliases: []string{"cv"},
		Short:   "Control cover/roller components",
		Long:    `Control cover and roller shutter components on Shelly devices.`,
	}

	cmd.AddCommand(open.NewCommand(f))
	cmd.AddCommand(closecmd.NewCommand(f))
	cmd.AddCommand(stop.NewCommand(f))
	cmd.AddCommand(status.NewCommand(f))
	cmd.AddCommand(position.NewCommand(f))
	cmd.AddCommand(calibrate.NewCommand(f))
	cmd.AddCommand(list.NewCommand(f))

	return cmd
}
