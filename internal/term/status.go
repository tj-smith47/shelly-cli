package term

import (
	"context"
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/output/table"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
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

	builder := table.NewBuilder("Component", "State")
	for _, cs := range states {
		builder.AddRow(cs.Name, cs.State)
	}
	tbl := builder.WithModeStyle(ios).Build()
	if err := tbl.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print component status table", err)
	}
}

// DisplayAllDevicesQuickStatus displays quick status for all registered devices.
func DisplayAllDevicesQuickStatus(ios *iostreams.IOStreams, statuses []QuickDeviceStatus) {
	if len(statuses) == 0 {
		ios.Warning("No devices registered. Use 'shelly device add' to add devices.")
		return
	}

	builder := table.NewBuilder("Device", "Model", "Status")
	for _, ds := range statuses {
		status := output.RenderOnline(ds.Online, output.CaseLower)
		deviceModel := ds.Model
		if !ds.Online {
			deviceModel = "-"
		}
		builder.AddRow(ds.Name, deviceModel, status)
	}
	tbl := builder.WithModeStyle(ios).Build()
	if err := tbl.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print devices status table", err)
	}
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
		return &ComponentState{Name: name, State: output.RenderSwitchState(status)}

	case model.ComponentLight:
		status, err := conn.Light(comp.ID).GetStatus(ctx)
		if err != nil {
			ios.DebugErr("get light status", err)
			return &ComponentState{Name: name, State: output.RenderErrorState()}
		}
		return &ComponentState{Name: name, State: output.RenderLightState(status)}

	case model.ComponentRGB:
		status, err := conn.RGB(comp.ID).GetStatus(ctx)
		if err != nil {
			ios.DebugErr("get RGB status", err)
			return &ComponentState{Name: name, State: output.RenderErrorState()}
		}
		return &ComponentState{Name: name, State: output.RenderRGBState(status)}

	case model.ComponentCover:
		status, err := conn.Cover(comp.ID).GetStatus(ctx)
		if err != nil {
			ios.DebugErr("get cover status", err)
			return &ComponentState{Name: name, State: output.RenderErrorState()}
		}
		return &ComponentState{Name: name, State: output.RenderCoverStatusState(status)}

	case model.ComponentInput:
		status, err := conn.Input(comp.ID).GetStatus(ctx)
		if err != nil {
			ios.DebugErr("get input status", err)
			return &ComponentState{Name: name, State: output.RenderErrorState()}
		}
		return &ComponentState{Name: name, State: output.RenderInputState(status)}

	default:
		return nil
	}
}

// GetSingleDeviceStatus fetches status for a single device.
func GetSingleDeviceStatus(ctx context.Context, ios *iostreams.IOStreams, dev *shelly.DeviceClient) (*QuickDeviceInfo, []ComponentState, error) {
	if dev.IsGen1() {
		devInfo := dev.Gen1().Info()
		return &QuickDeviceInfo{
			Model:      devInfo.Model,
			Generation: devInfo.Generation,
			Firmware:   devInfo.Firmware,
		}, nil, nil
	}

	conn := dev.Gen2()
	devInfo := conn.Info()
	info := &QuickDeviceInfo{
		Model:      devInfo.Model,
		Generation: devInfo.Generation,
		Firmware:   devInfo.Firmware,
	}

	components, err := conn.ListComponents(ctx)
	if err != nil {
		return nil, nil, err
	}

	var componentStates []ComponentState
	for _, comp := range components {
		state := GetComponentState(ctx, ios, conn, comp)
		if state != nil {
			componentStates = append(componentStates, *state)
		}
	}

	return info, componentStates, nil
}
