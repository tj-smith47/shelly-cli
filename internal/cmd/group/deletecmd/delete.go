// Package deletecmd provides the group delete subcommand.
package deletecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// NewCommand creates the group delete command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var yes bool

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
		ValidArgsFunction: cmdutil.CompleteGroupNames(),
		RunE: func(_ *cobra.Command, args []string) error {
			return run(args[0], yes)
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func run(name string, skipConfirm bool) error {
	// Check if group exists
	group, exists := config.GetGroup(name)
	if !exists {
		return fmt.Errorf("group %q not found", name)
	}

	// Confirm unless --yes is specified
	if !skipConfirm {
		msg := fmt.Sprintf("Delete group %q?", name)
		if len(group.Devices) > 0 {
			msg = fmt.Sprintf("Delete group %q with %d device(s)?", name, len(group.Devices))
		}
		confirmed, err := iostreams.Confirm(msg, false)
		if err != nil {
			return fmt.Errorf("confirmation failed: %w", err)
		}
		if !confirmed {
			iostreams.Info("Deletion cancelled")
			return nil
		}
	}

	err := config.DeleteGroup(name)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	iostreams.Success("Group %q deleted", name)

	return nil
}
