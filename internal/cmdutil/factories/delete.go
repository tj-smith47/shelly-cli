// Package factories provides command factory functions for creating standard CLI commands.
package factories

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
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

	// ServiceFunc is the function that performs the deletion using shelly.Service.
	// Use this for webhook and other shelly.Service-based deletions.
	ServiceFunc func(ctx context.Context, svc *shelly.Service, device string, id int) error

	// AutomationServiceFunc is the function that performs deletion using automation.Service.
	// Use this for script and schedule deletions.
	AutomationServiceFunc func(ctx context.Context, svc *automation.Service, device string, id int) error

	// ShowWarning controls whether a warning message is shown before confirmation.
	// If true, displays "This will delete <resource> <id>." before asking for confirmation.
	ShowWarning bool
}

// NewDeviceDeleteCommand creates a device-based delete command.
// This factory consolidates the common pattern for schedule/del, script/del, webhook/del, etc.
func NewDeviceDeleteCommand(f *cmdutil.Factory, opts DeviceDeleteOpts) *cobra.Command {
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

	flags.AddYesFlag(cmd, &yesFlag)

	return cmd
}

func runDeviceDelete(ctx context.Context, f *cmdutil.Factory, opts DeviceDeleteOpts, device string, id int, skipConfirm bool) error {
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

	resourceCapitalized := capitalize(opts.Resource)
	err = cmdutil.RunWithSpinner(ctx, ios, fmt.Sprintf("Deleting %s...", opts.Resource), func(ctx context.Context) error {
		var deleteErr error
		if opts.AutomationServiceFunc != nil {
			deleteErr = opts.AutomationServiceFunc(ctx, f.AutomationService(), device, id)
		} else {
			deleteErr = opts.ServiceFunc(ctx, f.ShellyService(), device, id)
		}
		if deleteErr != nil {
			return fmt.Errorf("failed to delete %s: %w", opts.Resource, deleteErr)
		}
		ios.Success("%s %d deleted", resourceCapitalized, id)
		return nil
	})
	if err != nil {
		return err
	}

	// Invalidate cached data for this resource type
	cacheType := cacheTypeForResource(opts.Resource)
	if cacheType != "" {
		cmdutil.InvalidateCache(f, device, cacheType)
	}
	return nil
}

// cacheTypeForResource maps a resource name to its cache type constant.
func cacheTypeForResource(resource string) string {
	switch resource {
	case "schedule":
		return cache.TypeSchedules
	case "script":
		return cache.TypeScripts
	case "webhook":
		return cache.TypeWebhooks
	default:
		return ""
	}
}
