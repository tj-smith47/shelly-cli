// Package factories provides command factory functions for creating standard CLI commands.
package factories

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// EnableDisableOpts configures an enable or disable command.
type EnableDisableOpts struct {
	// Feature is the display name of the feature being toggled (e.g., "Modbus-TCP", "cloud connection").
	Feature string

	// Enable is true for enable commands, false for disable commands.
	Enable bool

	// Aliases are command aliases. If empty, defaults to {"on"} for enable and {"off"} for disable.
	Aliases []string

	// Long is the long description. If empty, a default is generated.
	Long string

	// Example is the example usage text. If empty, a default is generated.
	Example string

	// ExampleParent is the parent command path for auto-generated examples (e.g., "modbus", "cloud").
	// Only used when Example is empty.
	ExampleParent string

	// ServiceFunc performs the enable/disable operation.
	ServiceFunc func(ctx context.Context, f *cmdutil.Factory, device string) error

	// PostSuccess is called after a successful operation to print additional info lines.
	// If nil, no additional output is printed.
	PostSuccess func(ios *cmdutil.Factory, device string)
}

// NewEnableDisableCommand creates an enable or disable command for a feature.
// This factory consolidates the common pattern across modbus, cloud, and similar enable/disable pairs.
func NewEnableDisableCommand(f *cmdutil.Factory, opts EnableDisableOpts) *cobra.Command {
	v := newToggleVerbs(opts.Enable)
	verb := v.Verb
	pastVerb := v.PastVerb
	gerund := v.Gerund

	aliases := opts.Aliases
	if len(aliases) == 0 {
		if opts.Enable {
			aliases = []string{"on"}
		} else {
			aliases = []string{"off"}
		}
	}

	long := opts.Long
	if long == "" {
		long = fmt.Sprintf("%s %s on a device.", capitalize(verb), opts.Feature)
	}

	example := opts.Example
	if example == "" && opts.ExampleParent != "" {
		example = fmt.Sprintf("  # %s %s\n  shelly %s %s <device>",
			capitalize(verb), opts.Feature, opts.ExampleParent, verb)
	}

	cmd := &cobra.Command{
		Use:               fmt.Sprintf("%s <device>", verb),
		Aliases:           aliases,
		Short:             fmt.Sprintf("%s %s", capitalize(verb), opts.Feature),
		Long:              long,
		Example:           example,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEnableDisable(cmd.Context(), f, opts, args[0], gerund, pastVerb)
		},
	}

	return cmd
}

func runEnableDisable(ctx context.Context, f *cmdutil.Factory, opts EnableDisableOpts, device, gerund, pastVerb string) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()

	err := cmdutil.RunWithSpinner(ctx, ios, fmt.Sprintf("%s %s...", gerund, opts.Feature), func(ctx context.Context) error {
		return opts.ServiceFunc(ctx, f, device)
	})
	if err != nil {
		return err
	}

	ios.Success("%s %s", opts.Feature, pastVerb)

	if opts.PostSuccess != nil {
		opts.PostSuccess(f, device)
	}

	return nil
}
