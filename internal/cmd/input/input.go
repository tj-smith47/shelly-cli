// Package input provides the input command and its subcommands.
package input

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/input/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/input/status"
	"github.com/tj-smith47/shelly-cli/internal/cmd/input/trigger"
)

// NewCommand creates the input command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "input",
		Aliases: []string{"in"},
		Short:   "Manage input components",
		Long:    `Manage input components on Shelly devices.`,
	}

	cmd.AddCommand(list.NewCommand())
	cmd.AddCommand(status.NewCommand())
	cmd.AddCommand(trigger.NewCommand())

	return cmd
}
