// Package factories provides command factory functions for creating standard CLI commands.
package factories

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// ScheduleToggleOpts configures a schedule enable/disable command.
type ScheduleToggleOpts struct {
	// Enable is true for enable commands, false for disable commands.
	Enable bool

	// Aliases are command aliases.
	Aliases []string

	// Long is the long description.
	Long string

	// Example is the example usage text.
	Example string

	// ValidArgsFunc provides shell completion.
	ValidArgsFunc func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective)

	// ServiceFunc performs the enable/disable operation on the schedule.
	ServiceFunc func(ctx context.Context, f *cmdutil.Factory, device string, id int) error
}

// NewScheduleToggleCommand creates a schedule enable or disable command.
// This factory consolidates the common pattern for schedule enable/disable pairs
// that take <device> <id> arguments.
func NewScheduleToggleCommand(f *cmdutil.Factory, opts ScheduleToggleOpts) *cobra.Command {
	verb := "disable"
	if opts.Enable {
		verb = "enable"
	}
	pastVerb := "disabled"
	if opts.Enable {
		pastVerb = "enabled"
	}
	gerund := "Disabling"
	if opts.Enable {
		gerund = "Enabling"
	}

	cmd := &cobra.Command{
		Use:     fmt.Sprintf("%s <device> <id>", verb),
		Aliases: opts.Aliases,
		Short:   fmt.Sprintf("%s a schedule", capitalize(verb)),
		Long:    opts.Long,
		Example: opts.Example,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid schedule ID: %s", args[1])
			}
			return runScheduleToggle(cmd.Context(), f, opts, args[0], id, gerund, pastVerb)
		},
	}

	if opts.ValidArgsFunc != nil {
		cmd.ValidArgsFunction = opts.ValidArgsFunc
	}

	return cmd
}

func runScheduleToggle(ctx context.Context, f *cmdutil.Factory, opts ScheduleToggleOpts, device string, id int, gerund, pastVerb string) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()

	return cmdutil.RunWithSpinner(ctx, ios, fmt.Sprintf("%s schedule...", gerund), func(ctx context.Context) error {
		if err := opts.ServiceFunc(ctx, f, device, id); err != nil {
			return fmt.Errorf("failed to %s schedule: %w", pastVerb[:len(pastVerb)-1], err)
		}
		ios.Success("Schedule %d %s", id, pastVerb)
		return nil
	})
}
