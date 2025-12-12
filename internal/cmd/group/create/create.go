// Package create provides the group create subcommand.
package create

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

// NewCommand creates the group create command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create <name>",
		Aliases: []string{"new", "add-group"},
		Short:   "Create a new device group",
		Long: `Create a new device group.

Group names must be unique and cannot contain spaces or special characters.`,
		Example: `  # Create a new group
  shelly group create living-room

  # Create using alias
  shelly group new bedroom

  # Short form
  shelly grp create office`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return run(args[0])
		},
	}

	return cmd
}

func run(name string) error {
	err := config.CreateGroup(name)
	if err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}

	iostreams.Success("Group %q created", name)
	iostreams.Info("Add devices with: shelly group add %s <device>...", name)

	return nil
}
