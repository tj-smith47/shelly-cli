// Package cmdutil provides command utilities and shared infrastructure for CLI commands.
package cmdutil

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/helpers"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// pluralize returns the plural form of a word.
// Handles common cases for component names (switch->switches, light->lights).
func pluralize(word string) string {
	word = strings.ToLower(word)
	switch {
	case strings.HasSuffix(word, "ch") || strings.HasSuffix(word, "sh") ||
		strings.HasSuffix(word, "x") || strings.HasSuffix(word, "s"):
		return word + "es"
	default:
		return word + "s"
	}
}

// BatchComponentOpts configures a batch component command.
type BatchComponentOpts struct {
	// Component type: "Switch", "Light", "RGB", "Cover"
	// If empty, defaults to "Switch" for backward compatibility.
	Component string

	// Action is the operation type: ActionOn, ActionOff, or ActionToggle.
	Action Action

	// ServiceFunc is the function that performs the action on a device.
	// It receives (ctx, svc, device, componentID) and returns error.
	ServiceFunc func(ctx context.Context, svc *shelly.Service, device string, componentID int) error
}

// NewBatchComponentCommand creates a batch on/off/toggle command for any component type.
// This factory consolidates the common pattern across batch commands and supports
// multiple component types (Switch, Light, RGB, etc.).
func NewBatchComponentCommand(f *Factory, opts BatchComponentOpts) *cobra.Command {
	var (
		groupName   string
		all         bool
		timeout     time.Duration
		componentID int
		concurrent  int
	)

	// Default to "Switch" for backward compatibility
	component := opts.Component
	if component == "" {
		component = "Switch"
	}
	componentLower := strings.ToLower(component)
	actionStr := string(opts.Action)

	// Build command metadata based on action type
	var (
		use     string
		aliases []string
		short   string
	)

	componentPlural := pluralize(component)

	switch opts.Action {
	case ActionOn:
		use = "on [device...]"
		aliases = []string{"enable"}
		short = fmt.Sprintf("Turn on %s", componentPlural)
	case ActionOff:
		use = "off [device...]"
		aliases = []string{"disable"}
		short = fmt.Sprintf("Turn off %s", componentPlural)
	case ActionToggle:
		use = "toggle [device...]"
		aliases = []string{"flip"}
		short = fmt.Sprintf("Toggle %s", componentPlural)
	}

	// Build the Long description with action-specific text
	longDesc := fmt.Sprintf(`%s multiple %s components simultaneously.

By default, targets %s component 0 on each device.
Use --%s to specify a different component ID.

Target devices can be specified multiple ways:
  - As arguments: device names or addresses
  - Via stdin: pipe device names (one per line or space-separated)
  - Via group: --group flag targets all devices in a group
  - Via all: --all flag targets all registered devices

Priority: explicit args > stdin > group > all

Stdin input supports comments (lines starting with #) and
blank lines are ignored, making it easy to use device lists
from files or other commands.`, short, componentLower, componentLower, componentLower)

	// Build examples with action-specific text
	examples := fmt.Sprintf(`  # %s specific devices
  shelly batch %s light-1 light-2

  # %s all devices in a group
  shelly batch %s --group living-room

  # %s all registered devices
  shelly batch %s --all

  # Control %s 1 on all devices in group
  shelly batch %s --group bedroom --%s 1

  # Control concurrency and timeout
  shelly batch %s --all --concurrent 10 --timeout 30s

  # Pipe device names from a file
  cat devices.txt | shelly batch %s

  # Pipe from device list command
  shelly device list -o json | jq -r '.[].name' | shelly batch %s`,
		short, actionStr,
		short, actionStr,
		short, actionStr,
		componentLower, actionStr, componentLower,
		actionStr,
		actionStr,
		actionStr)

	cmd := &cobra.Command{
		Use:     use,
		Aliases: aliases,
		Short:   short,
		Long:    longDesc,
		Example: examples,
		RunE: func(cmd *cobra.Command, args []string) error {
			targets, err := helpers.ResolveBatchTargets(groupName, all, args)
			if err != nil {
				return err
			}
			return runBatchComponent(cmd.Context(), f, opts, targets, componentID, timeout, concurrent)
		},
	}

	cmd.Flags().StringVarP(&groupName, "group", "g", "", "Target device group")
	cmd.Flags().BoolVarP(&all, "all", "a", false, "Target all registered devices")
	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 10*time.Second, "Timeout per device")
	// Use short flag -s for switch component for backward compatibility
	if componentLower == "switch" {
		cmd.Flags().IntVarP(&componentID, componentLower, "s", 0, fmt.Sprintf("%s component ID", component))
	} else {
		cmd.Flags().IntVar(&componentID, componentLower, 0, fmt.Sprintf("%s component ID", component))
	}
	cmd.Flags().IntVarP(&concurrent, "concurrent", "c", 5, "Max concurrent operations")

	return cmd
}

func runBatchComponent(ctx context.Context, f *Factory, opts BatchComponentOpts, targets []string, componentID int, timeout time.Duration, concurrent int) error {
	if len(targets) == 0 {
		return fmt.Errorf("no target devices specified")
	}

	ios := f.IOStreams()
	svc := f.ShellyService()

	ctx, cancel := context.WithTimeout(ctx, timeout*time.Duration(len(targets)))
	defer cancel()

	return RunBatch(ctx, ios, svc, targets, concurrent, func(ctx context.Context, svc *shelly.Service, device string) error {
		return opts.ServiceFunc(ctx, svc, device, componentID)
	})
}
