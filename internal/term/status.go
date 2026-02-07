package term

import (
	"context"
	"fmt"
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/output/table"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// ComponentState holds a component's display info and state.
type ComponentState struct {
	Type  string // Formatted type like "Switch 0", "Input 1"
	Name  string // User-assigned name from config (empty if none)
	State string // State string like "ON (45W)", "idle"
}

// QuickDeviceStatus holds status for the all-devices view.
type QuickDeviceStatus struct {
	Name      string
	Model     string
	Online    bool
	LinkState string // Derived state from parent link (empty if not linked or online)
}

// DisplayQuickDeviceStatus displays quick status for a single device.
func DisplayQuickDeviceStatus(ios *iostreams.IOStreams, states []ComponentState) {
	if len(states) == 0 {
		ios.Printf("No controllable components found\n")
		return
	}

	builder := table.NewBuilder("Component", "Name", "State")
	for _, cs := range states {
		name := cs.Name
		if name == "" {
			name = "-"
		}
		builder.AddRow(cs.Type, name, cs.State)
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
		var status string
		deviceModel := ds.Model
		switch {
		case ds.Online:
			status = output.RenderOnline(true, output.CaseLower)
		case ds.LinkState != "":
			status = ds.LinkState
			deviceModel = "-"
		default:
			status = output.RenderOnline(false, output.CaseLower)
			deviceModel = "-"
		}
		builder.AddRow(ds.Name, deviceModel, status)
	}
	tbl := builder.WithModeStyle(ios).Build()
	if err := tbl.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print devices status table", err)
	}
}

// formatComponentType formats a component type and ID into "Switch 0" style.
func formatComponentType(compType model.ComponentType, id int) string {
	typeName := string(compType)
	if typeName != "" {
		typeName = strings.ToUpper(typeName[:1]) + typeName[1:]
	}
	return fmt.Sprintf("%s %d", typeName, id)
}

// componentGetter abstracts getting status and config for any component type.
type componentGetter[S, C any] interface {
	GetStatus(ctx context.Context) (S, error)
	GetConfig(ctx context.Context) (C, error)
}

// fetchComponentState is a generic helper to fetch status and config name.
func fetchComponentState[S, C any](
	ctx context.Context,
	comp componentGetter[S, C],
	renderState func(S) string,
	getName func(C) *string,
) (state, name string, err error) {
	status, err := comp.GetStatus(ctx)
	if err != nil {
		return "", "", err
	}
	state = renderState(status)
	if cfg, cfgErr := comp.GetConfig(ctx); cfgErr == nil {
		if n := getName(cfg); n != nil {
			name = *n
		}
	}
	return state, name, nil
}

// GetComponentState returns the state for a controllable component.
func GetComponentState(ctx context.Context, ios *iostreams.IOStreams, conn *client.Client, comp model.Component) *ComponentState {
	typeStr := formatComponentType(comp.Type, comp.ID)

	var state, name string
	var err error

	switch comp.Type {
	case model.ComponentSwitch:
		state, name, err = fetchComponentState(ctx, conn.Switch(comp.ID),
			func(s *model.SwitchStatus) string {
				st := output.RenderSwitchState(s)
				if s.Power != nil && *s.Power > 0 {
					return fmt.Sprintf("%s (%.0fW)", st, *s.Power)
				}
				return st
			},
			func(c *model.SwitchConfig) *string { return c.Name })

	case model.ComponentInput:
		state, name, err = fetchInputState(ctx, conn.Input(comp.ID))

	case model.ComponentLight:
		state, name, err = fetchComponentState(ctx, conn.Light(comp.ID),
			output.RenderLightState,
			func(c *model.LightConfig) *string { return c.Name })

	case model.ComponentRGB:
		state, name, err = fetchComponentState(ctx, conn.RGB(comp.ID),
			output.RenderRGBState,
			func(c *model.RGBConfig) *string { return c.Name })

	case model.ComponentRGBW:
		state, name, err = fetchComponentState(ctx, conn.RGBW(comp.ID),
			output.RenderRGBWState,
			func(c *model.RGBWConfig) *string { return c.Name })

	case model.ComponentCover:
		state, name, err = fetchComponentState(ctx, conn.Cover(comp.ID),
			output.RenderCoverStatusState,
			func(c *model.CoverConfig) *string { return c.Name })

	default:
		return nil
	}

	if err != nil {
		ios.DebugErr(fmt.Sprintf("get %s status", comp.Type), err)
		return &ComponentState{Type: typeStr, State: output.RenderErrorState()}
	}

	return &ComponentState{Type: typeStr, Name: name, State: state}
}

// fetchInputState handles input's special Enable field.
func fetchInputState(ctx context.Context, input *client.InputComponent) (state, name string, err error) {
	status, err := input.GetStatus(ctx)
	if err != nil {
		return "", "", err
	}
	state = output.RenderInputState(status)
	if cfg, cfgErr := input.GetConfig(ctx); cfgErr == nil {
		if cfg.Name != nil {
			name = *cfg.Name
		}
		if !cfg.Enable {
			state = "disabled"
		}
	}
	return state, name, nil
}

// GetSingleDeviceStatus fetches component states for a single device.
func GetSingleDeviceStatus(ctx context.Context, ios *iostreams.IOStreams, dev *shelly.DeviceClient) ([]ComponentState, error) {
	if dev.IsGen1() {
		return nil, nil
	}

	conn := dev.Gen2()
	components, err := conn.ListComponents(ctx)
	if err != nil {
		return nil, err
	}

	var states []ComponentState
	for _, comp := range components {
		if state := GetComponentState(ctx, ios, conn, comp); state != nil {
			states = append(states, *state)
		}
	}

	return states, nil
}
