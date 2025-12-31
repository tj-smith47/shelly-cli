// Package create provides the group create subcommand.
package create

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds the options for the create command.
type Options struct {
	Factory *cmdutil.Factory
	Name    string
}

// NewCommand creates the group create command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
			opts.Name = args[0]
			return run(opts)
		},
	}

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	if err := config.CreateGroup(opts.Name); err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}

	ios.Success("Group %q created", opts.Name)
	ios.Info("Add devices with: shelly group add %s <device>...", opts.Name)

	return nil
}
