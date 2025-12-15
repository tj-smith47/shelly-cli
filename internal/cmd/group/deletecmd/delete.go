// Package deletecmd provides the group delete subcommand.
package deletecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

var yesFlag bool

// NewCommand creates the group delete command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <name>",
		Aliases: []string{"rm", "del"},
		Short:   "Delete a device group",
		Long: `Delete a device group.

This only removes the group definition, not the devices themselves.`,
		Example: `  # Delete a group (with confirmation)
  shelly group delete living-room

  # Delete without confirmation
  shelly group delete living-room --yes

  # Using alias
  shelly group rm bedroom

  # Short form
  shelly grp del office`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.GroupNames(),
		RunE: func(_ *cobra.Command, args []string) error {
			return run(f, args[0])
		},
	}

	cmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func run(f *cmdutil.Factory, name string) error {
	ios := f.IOStreams()

	// Check if group exists
	group := f.GetGroup(name)
	if group == nil {
		return fmt.Errorf("group %q not found", name)
	}

	// Confirm unless --yes is specified
	msg := fmt.Sprintf("Delete group %q?", name)
	if len(group.Devices) > 0 {
		msg = fmt.Sprintf("Delete group %q with %d device(s)?", name, len(group.Devices))
	}
	confirmed, err := f.ConfirmAction(msg, yesFlag)
	if err != nil {
		return fmt.Errorf("confirmation failed: %w", err)
	}
	if !confirmed {
		ios.Info("Deletion cancelled")
		return nil
	}

	if err := config.DeleteGroup(name); err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	ios.Success("Group %q deleted", name)

	return nil
}
