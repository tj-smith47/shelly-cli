// Package factories provides command factory functions for creating standard CLI commands.
package factories

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// QuickAction represents a quick device action.
type QuickAction string

// Quick action constants.
const (
	QuickOn     QuickAction = "on"
	QuickOff    QuickAction = "off"
	QuickToggle QuickAction = "toggle"
)

// QuickOpts configures a quick command (on/off/toggle for auto-detected devices).
type QuickOpts struct {
	// Action is the quick action type.
	Action QuickAction

	// Aliases are alternate command names.
	Aliases []string

	// Short is the short description.
	Short string

	// Long is the detailed description.
	Long string

	// Example shows usage examples.
	Example string

	// SpinnerText is the text shown during operation.
	SpinnerText string

	// SuccessSingular is the success message for single component (uses %q for device).
	SuccessSingular string

	// SuccessPlural is the success message for multiple components (uses %d for count, %q for device).
	SuccessPlural string
}

// quickOptions holds the runtime options for a quick command.
type quickOptions struct {
	flags.QuickComponentFlags
	Device  string
	Factory *cmdutil.Factory
	Config  QuickOpts
}

// NewQuickCommand creates a quick on/off/toggle command with auto-detection.
func NewQuickCommand(f *cmdutil.Factory, config QuickOpts) *cobra.Command {
	opts := &quickOptions{
		Factory: f,
		Config:  config,
	}

	cmd := &cobra.Command{
		Use:               string(config.Action) + " <device>",
		Aliases:           config.Aliases,
		Short:             config.Short,
		Long:              config.Long,
		Example:           config.Example,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runQuick(cmd.Context(), opts)
		},
	}

	flags.AddQuickComponentFlags(cmd, &opts.QuickComponentFlags)

	return cmd
}

func runQuick(ctx context.Context, opts *quickOptions) error {
	f := opts.Factory
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	var result *shelly.QuickResult
	err := cmdutil.RunWithSpinner(ctx, ios, opts.Config.SpinnerText, func(ctx context.Context) error {
		var opErr error
		componentID := opts.ComponentIDPointer()
		switch opts.Config.Action {
		case QuickOn:
			result, opErr = svc.QuickOn(ctx, opts.Device, componentID)
		case QuickOff:
			result, opErr = svc.QuickOff(ctx, opts.Device, componentID)
		case QuickToggle:
			result, opErr = svc.QuickToggle(ctx, opts.Device, componentID)
		}
		return opErr
	})
	if err != nil {
		return err
	}

	if result.Count == 1 {
		ios.Success(opts.Config.SuccessSingular, opts.Device)
	} else {
		ios.Success(opts.Config.SuccessPlural, result.Count, opts.Device)
	}
	return nil
}
