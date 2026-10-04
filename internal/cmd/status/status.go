// Package status provides the quick status command.
package status

import (
	"context"
	"sort"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// commandUse is the cobra Use string for the quick status command.
const commandUse = "status [device]"

// Options holds command options.
type Options struct {
	Device  string
	Factory *cmdutil.Factory
}

// NewCommand creates the status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     commandUse,
		Aliases: []string{"st", "state"},
		Short:   "Show device status (quick overview)",
		Long: `Show a quick status overview for a device or all registered devices.

If no device is specified, shows a summary of all registered devices
with their online/offline status and primary component state.

Use -o json or -o yaml for structured output. The all-devices list has the
fields name, model, online and link_state (set when an offline device is
linked to a parent switch).`,
		Example: `  # Show status for a specific device
  shelly status living-room

  # Show status for all devices
  shelly status

  # Online state of every device as JSON
  shelly status -o json

  # Names of the devices that are offline
  shelly status -o json | jq -r '.[] | select(.online == false) | .name'`,
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Device = args[0]
			}
			return run(cmd.Context(), opts)
		},
	}

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Single device status
	if opts.Device != "" {
		ctx, cancel := context.WithTimeout(ctx, 2*shelly.DefaultTimeout)
		defer cancel()

		var componentStates []term.ComponentState

		err := cmdutil.RunWithSpinner(ctx, ios, "Getting status...", func(ctx context.Context) error {
			return svc.WithDevice(ctx, opts.Device, func(dev *shelly.DeviceClient) error {
				var fetchErr error
				componentStates, fetchErr = term.GetSingleDeviceStatus(ctx, ios, dev)
				return fetchErr
			})
		})
		if err != nil {
			return err
		}

		if cmdutil.StructuredOutput() {
			componentStates = term.PlainComponentStates(componentStates)
		}
		return cmdutil.PrintListResult(ios, componentStates, term.DisplayQuickDeviceStatus)
	}

	// All devices status
	devices := config.ListDevices()
	if len(devices) == 0 {
		return cmdutil.PrintList(ios, []term.QuickDeviceStatus(nil), term.DisplayAllDevicesQuickStatus, func() {
			ios.Warning("No devices registered. Use 'shelly device add' to add devices.")
		})
	}

	names := make([]string, 0, len(devices))
	for name := range devices {
		names = append(names, name)
	}
	sort.Strings(names)

	statuses := make([]term.QuickDeviceStatus, len(names))

	err := cmdutil.RunWithSpinner(ctx, ios, "Checking devices...", func(ctx context.Context) error {
		shelly.ForEachDevice(ctx, names, shelly.DefaultTimeout, func(devCtx context.Context, idx int, deviceName string) {
			ds := term.QuickDeviceStatus{Name: deviceName}
			connErr := svc.WithDevice(devCtx, deviceName, func(dev *shelly.DeviceClient) error {
				ds.Model = dev.Info().Model
				ds.Online = true
				return nil
			})
			if connErr != nil {
				// The device's own timeout may be used up by the failed
				// connection, so the link lookup gets a fresh one.
				linkCtx, linkCancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
				defer linkCancel()
				if ls, linkErr := svc.ResolveLinkStatus(linkCtx, deviceName); linkErr == nil && ls != nil {
					ds.LinkState = ls.State
				}
			}
			statuses[idx] = ds
		})
		return nil
	})
	if err != nil {
		return err
	}

	return cmdutil.PrintListResult(ios, statuses, term.DisplayAllDevicesQuickStatus)
}
