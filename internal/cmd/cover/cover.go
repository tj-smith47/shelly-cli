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
)

// NewCommand creates the cover command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cover",
		Aliases: []string{"cv"},
		Short:   "Control cover/roller components",
		Long:    `Control cover and roller shutter components on Shelly devices.`,
	}

	cmd.AddCommand(open.NewCommand())
	cmd.AddCommand(closecmd.NewCommand())
	cmd.AddCommand(stop.NewCommand())
	cmd.AddCommand(status.NewCommand())
	cmd.AddCommand(position.NewCommand())
	cmd.AddCommand(calibrate.NewCommand())
	cmd.AddCommand(list.NewCommand())

	return cmd
}
