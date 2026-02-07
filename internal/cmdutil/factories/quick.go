// Package factories provides command factory functions for creating standard CLI commands.
package factories

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/ratelimit"
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
func NewQuickCommand(f *cmdutil.Factory, cfg QuickOpts) *cobra.Command {
	opts := &quickOptions{
		Factory: f,
		Config:  cfg,
	}

	cmd := &cobra.Command{
		Use:               string(cfg.Action) + " <device>",
		Aliases:           cfg.Aliases,
		Short:             cfg.Short,
		Long:              cfg.Long,
		Example:           cfg.Example,
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
		// Try link fallback on connectivity errors
		if proxyResult, proxyErr := tryLinkProxy(ctx, svc, opts.Device, opts.Config.Action, err); proxyErr == nil {
			ios.Success("%s", proxyResult)
			return nil
		}
		return err
	}

	if result.Count == 1 {
		ios.Success(opts.Config.SuccessSingular, opts.Device)
	} else {
		ios.Success(opts.Config.SuccessPlural, result.Count, opts.Device)
	}
	return nil
}

// tryLinkProxy attempts to control the parent switch when a linked child device is unreachable.
// Returns the success message and nil error on success, or empty string and error if not applicable.
func tryLinkProxy(ctx context.Context, svc *shelly.Service, device string, action QuickAction, originalErr error) (string, error) {
	// Only proxy on connectivity failures, not on auth/param errors
	if !ratelimit.IsConnectivityFailure(originalErr) {
		return "", originalErr
	}

	link, ok := config.GetLink(device)
	if !ok {
		return "", originalErr
	}

	var err error
	switch action {
	case QuickOn:
		err = svc.SwitchOn(ctx, link.ParentDevice, link.SwitchID)
	case QuickOff:
		err = svc.SwitchOff(ctx, link.ParentDevice, link.SwitchID)
	case QuickToggle:
		_, err = svc.SwitchToggle(ctx, link.ParentDevice, link.SwitchID)
	}
	if err != nil {
		return "", fmt.Errorf("link proxy failed: %w", err)
	}

	return fmt.Sprintf("Turned %s %q switch:%d (linked from %q)", action, link.ParentDevice, link.SwitchID, device), nil
}
