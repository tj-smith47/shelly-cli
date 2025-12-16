// Package cmdutil provides command utilities and shared infrastructure for CLI commands.
package cmdutil

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Action represents the type of component action.
type Action string

// Component action constants.
const (
	ActionOn     Action = "on"
	ActionOff    Action = "off"
	ActionToggle Action = "toggle"
)

// ComponentToggleFunc is a function that toggles a component and returns the new output state.
type ComponentToggleFunc func(ctx context.Context, svc *shelly.Service, device string, id int) (outputOn bool, err error)

// ComponentOpts configures a component command (on/off/toggle for Light, RGB, Switch, etc.).
type ComponentOpts struct {
	// Component is the display name (e.g., "Light", "RGB", "Switch").
	Component string

	// Action is the operation type: ActionOn, ActionOff, or ActionToggle.
	Action Action

	// SimpleFunc is used for on/off actions (returns error only).
	// Ignored for toggle actions.
	SimpleFunc ComponentAction

	// ToggleFunc is used for toggle actions (returns output state and error).
	// Ignored for on/off actions.
	ToggleFunc ComponentToggleFunc
}

// NewComponentCommand creates a component on/off/toggle command.
// This factory consolidates the common pattern across Light, RGB, Switch, etc.
func NewComponentCommand(f *Factory, opts ComponentOpts) *cobra.Command {
	var componentID int

	componentLower := strings.ToLower(opts.Component)
	actionStr := string(opts.Action)

	// Build command metadata based on action type
	var (
		use      string
		aliases  []string
		short    string
		long     string
		examples string
	)

	switch opts.Action {
	case ActionOn:
		use = "on <device>"
		aliases = []string{"enable", "1"}
		short = fmt.Sprintf("Turn %s on", componentLower)
		long = fmt.Sprintf("Turn on a %s component on the specified device.", componentLower)
		examples = fmt.Sprintf(`  # Turn on %s
  shelly %s on <device>

  # Turn on specific %s ID
  shelly %s on <device> --id 1`, componentLower, componentLower, componentLower, componentLower)

	case ActionOff:
		use = "off <device>"
		aliases = []string{"disable", "0"}
		short = fmt.Sprintf("Turn %s off", componentLower)
		long = fmt.Sprintf("Turn off a %s component on the specified device.", componentLower)
		examples = fmt.Sprintf(`  # Turn off %s
  shelly %s off <device>

  # Turn off specific %s ID
  shelly %s off <device> --id 1`, componentLower, componentLower, componentLower, componentLower)

	case ActionToggle:
		use = "toggle <device>"
		aliases = []string{"flip", "t"}
		short = fmt.Sprintf("Toggle %s on/off", componentLower)
		long = fmt.Sprintf("Toggle a %s component on or off on the specified device.", componentLower)
		examples = fmt.Sprintf(`  # Toggle %s
  shelly %s toggle <device>

  # Toggle specific %s ID
  shelly %s flip <device> --id 1`, componentLower, componentLower, componentLower, componentLower)
	}

	cmd := &cobra.Command{
		Use:               use,
		Aliases:           aliases,
		Short:             short,
		Long:              long,
		Example:           examples,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runComponent(cmd.Context(), f, opts, args[0], componentID, actionStr)
		},
	}

	AddComponentIDFlag(cmd, &componentID, opts.Component)

	return cmd
}

func runComponent(ctx context.Context, f *Factory, opts ComponentOpts, device string, componentID int, actionStr string) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()
	componentLower := strings.ToLower(opts.Component)

	switch opts.Action {
	case ActionOn, ActionOff:
		spinnerMsg := fmt.Sprintf("Turning %s %s...", componentLower, actionStr)
		successMsg := fmt.Sprintf("%s %d turned %s", opts.Component, componentID, actionStr)

		return RunSimple(ctx, ios, svc, device, componentID, spinnerMsg, successMsg, opts.SimpleFunc)

	case ActionToggle:
		spinnerMsg := fmt.Sprintf("Toggling %s...", componentLower)

		return RunWithSpinner(ctx, ios, spinnerMsg, func(ctx context.Context) error {
			outputOn, err := opts.ToggleFunc(ctx, svc, device, componentID)
			if err != nil {
				return err
			}

			state := "off"
			if outputOn {
				state = "on"
			}
			ios.Success("%s %d toggled %s", opts.Component, componentID, state)
			return nil
		})

	default:
		return fmt.Errorf("unknown action: %s", opts.Action)
	}
}
