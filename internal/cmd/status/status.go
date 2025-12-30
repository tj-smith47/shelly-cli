// Package statuscmd provides the quick status command.
package status

import (
	"context"
	"sort"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	Device  string
	Factory *cmdutil.Factory
}

// NewCommand creates the status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "status [device]",
		Aliases: []string{"st", "state"},
		Short:   "Show device status (quick overview)",
		Long: `Show a quick status overview for a device or all registered devices.

If no device is specified, shows a summary of all registered devices
with their online/offline status and primary component state.`,
		Example: `  # Show status for a specific device
  shelly status living-room

  # Show status for all devices
  shelly status`,
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
	ctx, cancel := context.WithTimeout(ctx, 2*shelly.DefaultTimeout)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	if opts.Device != "" {
		return runSingleDevice(ctx, ios, svc, opts.Device)
	}
	return runAllDevices(ctx, ios, svc)
}

func runSingleDevice(ctx context.Context, ios *iostreams.IOStreams, svc *shelly.Service, device string) error {
	var info *term.QuickDeviceInfo
	var componentStates []term.ComponentState

	err := cmdutil.RunWithSpinner(ctx, ios, "Getting status...", func(ctx context.Context) error {
		return svc.WithDevice(ctx, device, func(dev *shelly.DeviceClient) error {
			if dev.IsGen1() {
				// Gen1 devices have limited status info
				devInfo := dev.Gen1().Info()
				info = &term.QuickDeviceInfo{
					Model:      devInfo.Model,
					Generation: devInfo.Generation,
					Firmware:   devInfo.Firmware,
				}
				return nil
			}

			conn := dev.Gen2()
			devInfo := conn.Info()
			info = &term.QuickDeviceInfo{
				Model:      devInfo.Model,
				Generation: devInfo.Generation,
				Firmware:   devInfo.Firmware,
			}

			components, err := conn.ListComponents(ctx)
			if err != nil {
				return err
			}

			for _, comp := range components {
				state := term.GetComponentState(ctx, ios, conn, comp)
				if state != nil {
					componentStates = append(componentStates, *state)
				}
			}

			return nil
		})
	})
	if err != nil {
		return err
	}

	term.DisplayQuickDeviceStatus(ios, device, info, componentStates)
	return nil
}

func runAllDevices(ctx context.Context, ios *iostreams.IOStreams, svc *shelly.Service) error {
	devices := config.ListDevices()
	if len(devices) == 0 {
		ios.Warning("No devices registered. Use 'shelly device add' to add devices.")
		return nil
	}

	names := make([]string, 0, len(devices))
	for name := range devices {
		names = append(names, name)
	}
	sort.Strings(names)

	var statuses []term.QuickDeviceStatus

	err := cmdutil.RunWithSpinner(ctx, ios, "Checking devices...", func(ctx context.Context) error {
		for _, name := range names {
			ds := term.QuickDeviceStatus{Name: name}

			connErr := svc.WithDevice(ctx, name, func(dev *shelly.DeviceClient) error {
				devInfo := dev.Info()
				ds.Model = devInfo.Model
				ds.Online = true
				return nil
			})
			if connErr != nil {
				ds.Online = false
			}

			statuses = append(statuses, ds)
		}
		return nil
	})
	if err != nil {
		return err
	}

	term.DisplayAllDevicesQuickStatus(ios, statuses)
	return nil
}
