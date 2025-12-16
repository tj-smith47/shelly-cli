// Package statuscmd provides the quick status command.
package statuscmd

import (
	"context"
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
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

	// Single device status
	if opts.Device != "" {
		return showDeviceStatus(ctx, ios, svc, opts.Device)
	}

	// All devices status
	return showAllDevicesStatus(ctx, ios, svc)
}

func showDeviceStatus(ctx context.Context, ios *iostreams.IOStreams, svc *shelly.Service, device string) error {
	var info *client.DeviceInfo
	var components []model.Component
	var componentStates []componentState

	err := cmdutil.RunWithSpinner(ctx, ios, "Getting status...", func(ctx context.Context) error {
		return svc.WithConnection(ctx, device, func(conn *client.Client) error {
			info = conn.Info()

			var err error
			components, err = conn.ListComponents(ctx)
			if err != nil {
				return err
			}

			// Get state of controllable components
			for _, comp := range components {
				state := getComponentState(ctx, ios, conn, comp)
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

	// Display device info
	ios.Info("Device: %s", theme.Bold().Render(device))
	ios.Info("Model: %s (Gen%d)", info.Model, info.Generation)
	ios.Info("Firmware: %s", info.Firmware)
	ios.Println()

	if len(componentStates) == 0 {
		ios.Printf("No controllable components found\n")
		return nil
	}

	// Display component states
	table := output.NewTable("Component", "State")
	for _, cs := range componentStates {
		table.AddRow(cs.Name, cs.State)
	}
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print component status table", err)
	}

	return nil
}

type componentState struct {
	Name  string
	State string
}

func getComponentState(ctx context.Context, ios *iostreams.IOStreams, conn *client.Client, comp model.Component) *componentState {
	name := fmt.Sprintf("%s:%d", comp.Type, comp.ID)

	switch comp.Type {
	case model.ComponentSwitch:
		return getSwitchState(ctx, ios, conn, comp.ID, name)
	case model.ComponentLight:
		return getLightState(ctx, ios, conn, comp.ID, name)
	case model.ComponentRGB:
		return getRGBState(ctx, ios, conn, comp.ID, name)
	case model.ComponentCover:
		return getCoverState(ctx, ios, conn, comp.ID, name)
	case model.ComponentInput:
		return getInputState(ctx, ios, conn, comp.ID, name)
	default:
		return nil
	}
}

func getSwitchState(ctx context.Context, ios *iostreams.IOStreams, conn *client.Client, id int, name string) *componentState {
	status, err := conn.Switch(id).GetStatus(ctx)
	if err != nil {
		ios.DebugErr("failed to get switch status", err)
		return &componentState{Name: name, State: theme.StatusError().Render("error")}
	}
	if status.Output {
		return &componentState{Name: name, State: theme.StatusOK().Render("ON")}
	}
	return &componentState{Name: name, State: theme.Dim().Render("off")}
}

func getLightState(ctx context.Context, ios *iostreams.IOStreams, conn *client.Client, id int, name string) *componentState {
	status, err := conn.Light(id).GetStatus(ctx)
	if err != nil {
		ios.DebugErr("failed to get light status", err)
		return &componentState{Name: name, State: theme.StatusError().Render("error")}
	}
	if status.Output {
		if status.Brightness != nil && *status.Brightness > 0 {
			return &componentState{Name: name, State: theme.StatusOK().Render(fmt.Sprintf("ON (%d%%)", *status.Brightness))}
		}
		return &componentState{Name: name, State: theme.StatusOK().Render("ON")}
	}
	return &componentState{Name: name, State: theme.Dim().Render("off")}
}

func getRGBState(ctx context.Context, ios *iostreams.IOStreams, conn *client.Client, id int, name string) *componentState {
	status, err := conn.RGB(id).GetStatus(ctx)
	if err != nil {
		ios.DebugErr("failed to get RGB status", err)
		return &componentState{Name: name, State: theme.StatusError().Render("error")}
	}
	if status.Output {
		if status.Brightness != nil {
			return &componentState{Name: name, State: theme.StatusOK().Render(fmt.Sprintf("ON (%d%%)", *status.Brightness))}
		}
		return &componentState{Name: name, State: theme.StatusOK().Render("ON")}
	}
	return &componentState{Name: name, State: theme.Dim().Render("off")}
}

func getCoverState(ctx context.Context, ios *iostreams.IOStreams, conn *client.Client, id int, name string) *componentState {
	status, err := conn.Cover(id).GetStatus(ctx)
	if err != nil {
		ios.DebugErr("failed to get cover status", err)
		return &componentState{Name: name, State: theme.StatusError().Render("error")}
	}
	state := status.State
	if status.CurrentPosition != nil && *status.CurrentPosition >= 0 {
		state = fmt.Sprintf("%s (%d%%)", status.State, *status.CurrentPosition)
	}
	return &componentState{Name: name, State: state}
}

func getInputState(ctx context.Context, ios *iostreams.IOStreams, conn *client.Client, id int, name string) *componentState {
	status, err := conn.Input(id).GetStatus(ctx)
	if err != nil {
		ios.DebugErr("failed to get input status", err)
		return &componentState{Name: name, State: theme.StatusError().Render("error")}
	}
	if status.State {
		return &componentState{Name: name, State: theme.StatusWarn().Render("triggered")}
	}
	return &componentState{Name: name, State: theme.Dim().Render("idle")}
}

func showAllDevicesStatus(ctx context.Context, ios *iostreams.IOStreams, svc *shelly.Service) error {
	devices := config.ListDevices()
	if len(devices) == 0 {
		ios.Warning("No devices registered. Use 'shelly device add' to add devices.")
		return nil
	}

	type deviceStatus struct {
		Name   string
		Model  string
		Status string
	}

	var statuses []deviceStatus

	// Sort device names for consistent ordering
	names := make([]string, 0, len(devices))
	for name := range devices {
		names = append(names, name)
	}
	sort.Strings(names)

	err := cmdutil.RunWithSpinner(ctx, ios, "Checking devices...", func(ctx context.Context) error {
		for _, name := range names {
			ds := deviceStatus{Name: name}

			err := svc.WithConnection(ctx, name, func(conn *client.Client) error {
				info := conn.Info()
				ds.Model = info.Model
				ds.Status = theme.StatusOK().Render("online")
				return nil
			})
			if err != nil {
				ds.Model = "-"
				ds.Status = theme.StatusError().Render("offline")
			}

			statuses = append(statuses, ds)
		}
		return nil
	})
	if err != nil {
		return err
	}

	table := output.NewTable("Device", "Model", "Status")
	for _, ds := range statuses {
		table.AddRow(ds.Name, ds.Model, ds.Status)
	}
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print devices status table", err)
	}

	return nil
}
