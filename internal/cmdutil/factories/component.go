// Package factories provides command factory functions for creating standard CLI commands.
package factories

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/component"
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
	SimpleFunc cmdutil.ComponentAction

	// ToggleFunc is used for toggle actions (returns output state and error).
	// Ignored for on/off actions.
	ToggleFunc ComponentToggleFunc
}

// NewComponentCommand creates a component on/off/toggle command.
// This factory consolidates the common pattern across Light, RGB, Switch, etc.
func NewComponentCommand(f *cmdutil.Factory, opts ComponentOpts) *cobra.Command {
	var componentFlags flags.ComponentNameFlags

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

  # Turn on specific %s by ID
  shelly %s on <device> --id 1

  # Turn on %s by name
  shelly %s on <device> --name "Kitchen Light"`, componentLower, componentLower, componentLower, componentLower, componentLower, componentLower)

	case ActionOff:
		use = "off <device>"
		aliases = []string{"disable", "0"}
		short = fmt.Sprintf("Turn %s off", componentLower)
		long = fmt.Sprintf("Turn off a %s component on the specified device.", componentLower)
		examples = fmt.Sprintf(`  # Turn off %s
  shelly %s off <device>

  # Turn off specific %s by ID
  shelly %s off <device> --id 1

  # Turn off %s by name
  shelly %s off <device> --name "Kitchen Light"`, componentLower, componentLower, componentLower, componentLower, componentLower, componentLower)

	case ActionToggle:
		use = "toggle <device>"
		aliases = []string{"flip", "t"}
		short = fmt.Sprintf("Toggle %s on/off", componentLower)
		long = fmt.Sprintf("Toggle a %s component on or off on the specified device.", componentLower)
		examples = fmt.Sprintf(`  # Toggle %s
  shelly %s toggle <device>

  # Toggle specific %s by ID
  shelly %s flip <device> --id 1

  # Toggle %s by name
  shelly %s toggle <device> --name "Kitchen Light"`, componentLower, componentLower, componentLower, componentLower, componentLower, componentLower)
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
			return runComponent(cmd.Context(), f, opts, args[0], &componentFlags, actionStr)
		},
	}

	flags.AddComponentNameFlags(cmd, &componentFlags, opts.Component)

	return cmd
}

func runComponent(ctx context.Context, f *cmdutil.Factory, opts ComponentOpts, device string, compFlags *flags.ComponentNameFlags, actionStr string) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()
	componentLower := strings.ToLower(opts.Component)

	// Resolve component name to ID if provided
	componentID := compFlags.ID
	if compFlags.HasName() {
		compType, ok := component.GetType(componentLower)
		if !ok {
			return fmt.Errorf("unknown component type: %s", componentLower)
		}

		configFetcher := makeConfigFetcher(svc)
		resolvedID, err := component.ResolveIDWithGen(ctx, configFetcher, device, compType, compFlags.ID, compFlags.Name)
		if err != nil {
			return fmt.Errorf("failed to resolve %s name %q: %w", componentLower, compFlags.Name, err)
		}
		componentID = resolvedID
	}

	switch opts.Action {
	case ActionOn, ActionOff:
		spinnerMsg := fmt.Sprintf("Turning %s %s...", componentLower, actionStr)
		successMsg := formatComponentSuccessMsg(opts.Component, componentID, compFlags.Name, actionStr)

		return cmdutil.RunSimple(ctx, ios, svc, device, componentID, spinnerMsg, successMsg, opts.SimpleFunc)

	case ActionToggle:
		spinnerMsg := fmt.Sprintf("Toggling %s...", componentLower)

		return cmdutil.RunWithSpinner(ctx, ios, spinnerMsg, func(ctx context.Context) error {
			outputOn, err := opts.ToggleFunc(ctx, svc, device, componentID)
			if err != nil {
				return err
			}

			state := "off"
			if outputOn {
				state = "on"
			}
			ios.Success("%s toggled %s", formatComponentDisplayName(opts.Component, componentID, compFlags.Name), state)
			return nil
		})

	default:
		return fmt.Errorf("unknown action: %s", opts.Action)
	}
}

// makeConfigFetcher creates a config fetcher function for component name resolution.
func makeConfigFetcher(svc *shelly.Service) func(ctx context.Context, device string) (map[string]json.RawMessage, error) {
	return func(ctx context.Context, device string) (map[string]json.RawMessage, error) {
		cfg, err := svc.GetConfig(ctx, device)
		if err != nil {
			return nil, err
		}

		// Convert map[string]any to map[string]json.RawMessage
		result := make(map[string]json.RawMessage)
		for k, v := range cfg {
			raw, err := json.Marshal(v)
			if err != nil {
				continue
			}
			result[k] = raw
		}
		return result, nil
	}
}

// formatComponentSuccessMsg formats the success message with name if provided.
func formatComponentSuccessMsg(componentType string, id int, name, actionStr string) string {
	if name != "" {
		return fmt.Sprintf("%s %q turned %s", componentType, name, actionStr)
	}
	return fmt.Sprintf("%s %d turned %s", componentType, id, actionStr)
}

// formatComponentDisplayName formats a component display name with name or ID.
func formatComponentDisplayName(componentType string, id int, name string) string {
	if name != "" {
		return fmt.Sprintf("%s %q", componentType, name)
	}
	return fmt.Sprintf("%s %d", componentType, id)
}

// ListOpts configures a component list command.
type ListOpts[T any] struct {
	// Component is the display name (e.g., "Light", "RGB", "Switch", "Cover").
	Component string

	// Long is the long description for the command.
	Long string

	// Example is the example usage text.
	Example string

	// Fetcher retrieves the list of components from the device.
	Fetcher cmdutil.ListFetcher[T]

	// Display renders the list in human-readable format.
	Display cmdutil.ListDisplay[T]
}

// NewListCommand creates a generic component list command.
// This factory consolidates the common pattern across Light, RGB, Switch, Cover list commands.
func NewListCommand[T any](f *cmdutil.Factory, opts ListOpts[T]) *cobra.Command {
	componentLower := strings.ToLower(opts.Component)

	cmd := &cobra.Command{
		Use:               "list <device>",
		Aliases:           []string{"ls", "l"},
		Short:             fmt.Sprintf("List %s components", componentLower),
		Long:              opts.Long,
		Example:           opts.Example,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd.Context(), f, opts, args[0])
		},
	}

	return cmd
}

func runList[T any](ctx context.Context, f *cmdutil.Factory, opts ListOpts[T], device string) error {
	ctx, cancel := f.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	componentLower := strings.ToLower(opts.Component)
	spinnerMsg := fmt.Sprintf("Fetching %s components...", componentLower)
	emptyMsg := fmt.Sprintf("%s components", componentLower)

	return cmdutil.RunList(ctx, ios, svc, device, spinnerMsg, emptyMsg, opts.Fetcher, opts.Display)
}
