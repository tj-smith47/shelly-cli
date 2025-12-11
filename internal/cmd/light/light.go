// Package light provides the light command and its subcommands.
package light

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/light/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/light/off"
	"github.com/tj-smith47/shelly-cli/internal/cmd/light/on"
	lightset "github.com/tj-smith47/shelly-cli/internal/cmd/light/set"
	"github.com/tj-smith47/shelly-cli/internal/cmd/light/status"
	"github.com/tj-smith47/shelly-cli/internal/cmd/light/toggle"
)

// NewCommand creates the light command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "light",
		Aliases: []string{"lt"},
		Short:   "Control light components",
		Long:    `Control dimmable light components on Shelly devices.`,
	}

	cmd.AddCommand(on.NewCommand())
	cmd.AddCommand(off.NewCommand())
	cmd.AddCommand(toggle.NewCommand())
	cmd.AddCommand(status.NewCommand())
	cmd.AddCommand(lightset.NewCommand())
	cmd.AddCommand(list.NewCommand())

	return cmd
}
