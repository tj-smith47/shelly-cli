// Package factories provides command factory functions for creating standard CLI commands.
package factories

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
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
func NewConfigDeleteCommand(f *cmdutil.Factory, opts ConfigDeleteOpts) *cobra.Command {
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

func runConfigDelete(f *cmdutil.Factory, opts ConfigDeleteOpts, name string, skipConfirm bool) error {
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

// ConfigListOpts configures a config-based list command.
type ConfigListOpts[T any] struct {
	// Resource name (e.g., "alias", "group", "scene", "template").
	Resource string

	// FetchFunc retrieves items from config.
	FetchFunc func() []T

	// DisplayFunc renders table output.
	DisplayFunc func(ios *iostreams.IOStreams, items []T)

	// EmptyMsg shown when no items exist.
	// If empty, defaults to "No <resource>s configured".
	EmptyMsg string

	// HintMsg shown after empty message to suggest next action.
	HintMsg string
}

// NewConfigListCommand creates a config-based list command.
func NewConfigListCommand[T any](f *cmdutil.Factory, opts ConfigListOpts[T]) *cobra.Command {
	short := fmt.Sprintf("List %ss", opts.Resource)
	long := fmt.Sprintf(`List all configured %ss.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.`, opts.Resource)

	examples := fmt.Sprintf(`  # List all %ss
  shelly %s list

  # Output as JSON
  shelly %s list -o json

  # Output as YAML
  shelly %s list -o yaml`, opts.Resource, opts.Resource, opts.Resource, opts.Resource)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   short,
		Long:    long,
		Example: examples,
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runConfigList(f, opts)
		},
	}

	return cmd
}

func runConfigList[T any](f *cmdutil.Factory, opts ConfigListOpts[T]) error {
	ios := f.IOStreams()
	items := opts.FetchFunc()

	if len(items) == 0 {
		msg := opts.EmptyMsg
		if msg == "" {
			msg = fmt.Sprintf("No %ss configured", opts.Resource)
		}
		ios.Info("%s", msg)
		if opts.HintMsg != "" {
			ios.Info("%s", opts.HintMsg)
		}
		return nil
	}

	// Handle structured output (JSON/YAML/template)
	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, items)
	}

	// Table output
	opts.DisplayFunc(ios, items)
	return nil
}

// capitalize returns the string with first letter capitalized.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
