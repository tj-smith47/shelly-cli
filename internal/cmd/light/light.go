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
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the light command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "light",
		Aliases: []string{"lt"},
		Short:   "Control light components",
		Long:    `Control dimmable light components on Shelly devices.`,
	}

	cmd.AddCommand(on.NewCommand(f))
	cmd.AddCommand(off.NewCommand(f))
	cmd.AddCommand(toggle.NewCommand(f))
	cmd.AddCommand(status.NewCommand(f))
	cmd.AddCommand(lightset.NewCommand(f))
	cmd.AddCommand(list.NewCommand(f))

	return cmd
}
