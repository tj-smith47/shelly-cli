// Package cmdutil provides command utilities and shared infrastructure for CLI commands.
package cmdutil

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ConfigDeleteOpts configures a config-based delete command.
type ConfigDeleteOpts struct {
	// Resource is the name (e.g., "scene", "group", "alias", "template").
	Resource string

	// ValidArgsFunc provides shell completion.
	ValidArgsFunc func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective)

	// ExistsFunc checks if resource exists and returns it.
	// Returns (resource, exists). Resource can be any type.
	ExistsFunc func(name string) (any, bool)

	// DeleteFunc performs the deletion.
	DeleteFunc func(name string) error

	// InfoFunc returns extra info for confirmation message.
	// If nil, uses simple "Delete <resource> <name>?" message.
	// Example: "Delete scene \"movie-night\" with 5 action(s)?"
	InfoFunc func(resource any, name string) string

	// SkipConfirmation if true, deletes without asking (e.g., alias delete).
	SkipConfirmation bool
}

// NewConfigDeleteCommand creates a config-based delete command.
func NewConfigDeleteCommand(f *Factory, opts ConfigDeleteOpts) *cobra.Command {
	var yesFlag bool

	use := fmt.Sprintf("delete <%s>", opts.Resource)
	short := fmt.Sprintf("Delete a %s", opts.Resource)
	long := fmt.Sprintf("Delete a saved %s permanently.", opts.Resource)

	examples := fmt.Sprintf(`  # Delete a %s (with confirmation)
  shelly %s delete my-%s

  # Delete without confirmation
  shelly %s delete my-%s --yes

  # Using alias
  shelly %s rm my-%s`, opts.Resource, opts.Resource, opts.Resource,
		opts.Resource, opts.Resource, opts.Resource, opts.Resource)

	cmd := &cobra.Command{
		Use:     use,
		Aliases: []string{"rm", "del", "remove"},
		Short:   short,
		Long:    long,
		Example: examples,
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runConfigDelete(f, opts, args[0], yesFlag)
		},
	}

	if opts.ValidArgsFunc != nil {
		cmd.ValidArgsFunction = opts.ValidArgsFunc
	}

	if !opts.SkipConfirmation {
		cmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompt")
	}

	return cmd
}

func runConfigDelete(f *Factory, opts ConfigDeleteOpts, name string, skipConfirm bool) error {
	ios := f.IOStreams()

	// Check if resource exists
	resource, exists := opts.ExistsFunc(name)
	if !exists {
		return fmt.Errorf("%s %q not found", opts.Resource, name)
	}

	// Confirm unless skipped
	if !opts.SkipConfirmation {
		msg := fmt.Sprintf("Delete %s %q?", opts.Resource, name)
		if opts.InfoFunc != nil {
			msg = opts.InfoFunc(resource, name)
		}

		confirmed, err := f.ConfirmAction(msg, skipConfirm)
		if err != nil {
			return fmt.Errorf("confirmation failed: %w", err)
		}
		if !confirmed {
			ios.Info("Deletion cancelled")
			return nil
		}
	}

	if err := opts.DeleteFunc(name); err != nil {
		return fmt.Errorf("failed to delete %s: %w", opts.Resource, err)
	}

	ios.Success("%s %q deleted", capitalize(opts.Resource), name)
	return nil
}
