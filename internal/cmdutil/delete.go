// Package cmdutil provides command utilities and shared infrastructure for CLI commands.
package cmdutil

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// DeviceDeleteOpts configures a device-based delete command.
type DeviceDeleteOpts struct {
	// Resource is the name of the resource being deleted (e.g., "schedule", "script", "webhook").
	Resource string

	// Use is the command use string (e.g., "delete <device> <id>").
	// If empty, defaults to "delete <device> <id>".
	Use string

	// Aliases are command aliases.
	// If empty, defaults to {"del", "rm"}.
	Aliases []string

	// Long is the command's long description.
	// If empty, defaults to a generic description.
	Long string

	// ValidArgsFunc provides completion for the command arguments.
	// If nil, no completion is provided.
	ValidArgsFunc func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective)

	// ServiceFunc is the function that performs the deletion.
	ServiceFunc func(ctx context.Context, svc *shelly.Service, device string, id int) error

	// ShowWarning controls whether a warning message is shown before confirmation.
	// If true, displays "This will delete <resource> <id>." before asking for confirmation.
	ShowWarning bool
}

// NewDeviceDeleteCommand creates a device-based delete command.
// This factory consolidates the common pattern for schedule/del, script/del, webhook/del, etc.
func NewDeviceDeleteCommand(f *Factory, opts DeviceDeleteOpts) *cobra.Command {
	var yesFlag bool

	// Apply defaults
	use := opts.Use
	if use == "" {
		use = fmt.Sprintf("delete <device> <%s-id>", opts.Resource)
	}

	aliases := opts.Aliases
	if len(aliases) == 0 {
		aliases = []string{"del", "rm"}
	}

	long := opts.Long
	if long == "" {
		long = fmt.Sprintf("Delete a %s from a device.", opts.Resource)
	}

	short := fmt.Sprintf("Delete a %s", opts.Resource)

	examples := fmt.Sprintf(`  # Delete a %s
  shelly %s delete <device> 1

  # Delete without confirmation
  shelly %s delete <device> 1 --yes`, opts.Resource, opts.Resource, opts.Resource)

	cmd := &cobra.Command{
		Use:     use,
		Aliases: aliases,
		Short:   short,
		Long:    long,
		Example: examples,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid %s ID: %s", opts.Resource, args[1])
			}
			return runDeviceDelete(cmd.Context(), f, opts, args[0], id, yesFlag)
		},
	}

	if opts.ValidArgsFunc != nil {
		cmd.ValidArgsFunction = opts.ValidArgsFunc
	}

	AddYesFlag(cmd, &yesFlag)

	return cmd
}

func runDeviceDelete(ctx context.Context, f *Factory, opts DeviceDeleteOpts, device string, id int, skipConfirm bool) error {
	ios := f.IOStreams()

	// Show warning if configured
	if opts.ShowWarning {
		ios.Warning("This will delete %s %d.", opts.Resource, id)
	}

	// Confirm deletion
	confirmMsg := fmt.Sprintf("Delete %s %d?", opts.Resource, id)
	confirmed, err := f.ConfirmAction(confirmMsg, skipConfirm)
	if err != nil {
		return err
	}
	if !confirmed {
		ios.Warning("Delete cancelled")
		return nil
	}

	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	svc := f.ShellyService()

	return RunWithSpinner(ctx, ios, fmt.Sprintf("Deleting %s...", opts.Resource), func(ctx context.Context) error {
		if err := opts.ServiceFunc(ctx, svc, device, id); err != nil {
			return fmt.Errorf("failed to delete %s: %w", opts.Resource, err)
		}
		ios.Success("%s %d deleted", capitalize(opts.Resource), id)
		return nil
	})
}

// capitalize returns the string with first letter capitalized.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
