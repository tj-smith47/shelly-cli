package term

import (
	"context"
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// QuickDeviceInfo holds device info for quick status display.
type QuickDeviceInfo struct {
	Model      string
	Generation int
	Firmware   string
}

// ComponentState holds a component's display name and state string.
type ComponentState struct {
	Name  string
	State string
}

// QuickDeviceStatus holds status for the all-devices view.
type QuickDeviceStatus struct {
	Name   string
	Model  string
	Online bool
}

// DisplayQuickDeviceStatus displays quick status for a single device.
func DisplayQuickDeviceStatus(ios *iostreams.IOStreams, device string, info *QuickDeviceInfo, states []ComponentState) {
	ios.Info("Device: %s", theme.Bold().Render(device))
	ios.Info("Model: %s (Gen%d)", info.Model, info.Generation)
	ios.Info("Firmware: %s", info.Firmware)
	ios.Println()

	if len(states) == 0 {
		ios.Printf("No controllable components found\n")
		return
	}

	table := output.NewTable("Component", "State")
	for _, cs := range states {
		table.AddRow(cs.Name, cs.State)
	}
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print component status table", err)
	}
}

// DisplayAllDevicesQuickStatus displays quick status for all registered devices.
func DisplayAllDevicesQuickStatus(ios *iostreams.IOStreams, statuses []QuickDeviceStatus) {
	if len(statuses) == 0 {
		ios.Warning("No devices registered. Use 'shelly device add' to add devices.")
		return
	}

	table := output.NewTable("Device", "Model", "Status")
	for _, ds := range statuses {
		status := output.RenderOnline(ds.Online, output.CaseLower)
		deviceModel := ds.Model
		if !ds.Online {
			deviceModel = "-"
		}
		table.AddRow(ds.Name, deviceModel, status)
	}
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print devices status table", err)
	}
}

// RenderSwitchState returns the state string for a switch component.
func RenderSwitchState(status *model.SwitchStatus) string {
	return output.RenderOnOff(status.Output, output.CaseUpper, theme.FalseDim)
}

// RenderLightState returns the state string for a light component.
func RenderLightState(status *model.LightStatus) string {
	return output.RenderOnOffStateWithBrightness(status.Output, status.Brightness)
}

// RenderRGBState returns the state string for an RGB component.
func RenderRGBState(status *model.RGBStatus) string {
	return output.RenderOnOffStateWithBrightness(status.Output, status.Brightness)
}

// RenderCoverState returns the state string for a cover component.
func RenderCoverState(status *model.CoverStatus) string {
	state := status.State
	if status.CurrentPosition != nil && *status.CurrentPosition >= 0 {
		state = fmt.Sprintf("%s (%d%%)", status.State, *status.CurrentPosition)
	}
	return state
}

// RenderInputState returns the state string for an input component.
func RenderInputState(status *model.InputStatus) string {
	return output.RenderInputTriggeredState(status.State)
}

// GetComponentState returns the state for a controllable component.
func GetComponentState(ctx context.Context, ios *iostreams.IOStreams, conn *client.Client, comp model.Component) *ComponentState {
	name := fmt.Sprintf("%s:%d", comp.Type, comp.ID)

	switch comp.Type {
	case model.ComponentSwitch:
		status, err := conn.Switch(comp.ID).GetStatus(ctx)
		if err != nil {
			ios.DebugErr("get switch status", err)
			return &ComponentState{Name: name, State: output.RenderErrorState()}
		}
		return &ComponentState{Name: name, State: RenderSwitchState(status)}

	case model.ComponentLight:
		status, err := conn.Light(comp.ID).GetStatus(ctx)
		if err != nil {
			ios.DebugErr("get light status", err)
			return &ComponentState{Name: name, State: output.RenderErrorState()}
		}
		return &ComponentState{Name: name, State: RenderLightState(status)}

	case model.ComponentRGB:
		status, err := conn.RGB(comp.ID).GetStatus(ctx)
		if err != nil {
			ios.DebugErr("get RGB status", err)
			return &ComponentState{Name: name, State: output.RenderErrorState()}
		}
		return &ComponentState{Name: name, State: RenderRGBState(status)}

	case model.ComponentCover:
		status, err := conn.Cover(comp.ID).GetStatus(ctx)
		if err != nil {
			ios.DebugErr("get cover status", err)
			return &ComponentState{Name: name, State: output.RenderErrorState()}
		}
		return &ComponentState{Name: name, State: RenderCoverState(status)}

	case model.ComponentInput:
		status, err := conn.Input(comp.ID).GetStatus(ctx)
		if err != nil {
			ios.DebugErr("get input status", err)
			return &ComponentState{Name: name, State: output.RenderErrorState()}
		}
		return &ComponentState{Name: name, State: RenderInputState(status)}

	default:
		return nil
	}
}
