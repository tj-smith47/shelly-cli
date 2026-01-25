// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly/connection"
	"github.com/tj-smith47/shelly-cli/internal/tui/debug"
)

// Action constants for device control.
const (
	ActionOn     = "on"
	ActionOff    = "off"
	ActionToggle = "toggle"
)

// QuickResult holds the result of a quick operation.
type QuickResult struct {
	// Count is the number of components affected.
	Count int
	// PluginResult holds the result from a plugin dispatch (nil for Shelly devices).
	PluginResult *PluginQuickResult
}

// QuickOn turns on controllable components on a device.
// If componentID is nil, turns on all controllable components.
// If componentID is set, turns on only that specific component.
// For plugin-managed devices, dispatches to the plugin's control hook.
// Supports both Gen1 and Gen2+ devices.
func (s *Service) QuickOn(ctx context.Context, identifier string, componentID *int) (*QuickResult, error) {
	// Resolve the device to check if it's plugin-managed
	device, err := s.resolver.Resolve(identifier)
	if err != nil {
		return nil, err
	}

	// Dispatch to plugin for non-Shelly devices
	if device.IsPluginManaged() {
		pluginResult, err := s.dispatchToPlugin(ctx, device, ActionOn, "switch", componentID)
		if err != nil {
			return nil, err
		}
		return &QuickResult{Count: 1, PluginResult: pluginResult}, nil
	}

	// Shelly device - use native control with Gen1/Gen2 support
	result := &QuickResult{}

	err = s.WithDevice(ctx, identifier, func(dev *connection.DeviceClient) error {
		if dev.IsGen1() {
			return quickOnGen1(ctx, dev.Gen1(), componentID, result)
		}
		return quickOnGen2(ctx, dev.Gen2(), componentID, result)
	})

	return result, err
}

// QuickOff turns off controllable components on a device.
// If componentID is nil, turns off all controllable components.
// If componentID is set, turns off only that specific component.
// For plugin-managed devices, dispatches to the plugin's control hook.
// Supports both Gen1 and Gen2+ devices.
func (s *Service) QuickOff(ctx context.Context, identifier string, componentID *int) (*QuickResult, error) {
	// Resolve the device to check if it's plugin-managed
	device, err := s.resolver.Resolve(identifier)
	if err != nil {
		return nil, err
	}

	// Dispatch to plugin for non-Shelly devices
	if device.IsPluginManaged() {
		pluginResult, err := s.dispatchToPlugin(ctx, device, ActionOff, "switch", componentID)
		if err != nil {
			return nil, err
		}
		return &QuickResult{Count: 1, PluginResult: pluginResult}, nil
	}

	// Shelly device - use native control with Gen1/Gen2 support
	result := &QuickResult{}

	err = s.WithDevice(ctx, identifier, func(dev *connection.DeviceClient) error {
		if dev.IsGen1() {
			return quickOffGen1(ctx, dev.Gen1(), componentID, result)
		}
		return quickOffGen2(ctx, dev.Gen2(), componentID, result)
	})

	return result, err
}

// QuickToggle toggles controllable components on a device.
// If componentID is nil, toggles all controllable components.
// If componentID is set, toggles only that specific component.
// For plugin-managed devices, dispatches to the plugin's control hook.
// Supports both Gen1 and Gen2+ devices.
func (s *Service) QuickToggle(ctx context.Context, identifier string, componentID *int) (*QuickResult, error) {
	// Resolve the device to check if it's plugin-managed
	device, err := s.resolver.Resolve(identifier)
	if err != nil {
		return nil, err
	}

	// Dispatch to plugin for non-Shelly devices
	if device.IsPluginManaged() {
		pluginResult, err := s.dispatchToPlugin(ctx, device, ActionToggle, "switch", componentID)
		if err != nil {
			return nil, err
		}
		return &QuickResult{Count: 1, PluginResult: pluginResult}, nil
	}

	// Shelly device - use native control with Gen1/Gen2 support
	result := &QuickResult{}

	err = s.WithDevice(ctx, identifier, func(dev *connection.DeviceClient) error {
		if dev.IsGen1() {
			return quickToggleGen1(ctx, dev.Gen1(), componentID, result)
		}
		return quickToggleGen2(ctx, dev.Gen2(), componentID, result)
	})

	return result, err
}

// findControllable lists and filters components to controllable ones.
func findControllable(ctx context.Context, conn *client.Client) ([]model.Component, error) {
	components, err := conn.ListComponents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list components: %w", err)
	}

	var controllable []model.Component
	for _, comp := range components {
		if isControllable(comp.Type) {
			controllable = append(controllable, comp)
		}
	}

	if len(controllable) == 0 {
		return nil, fmt.Errorf("no controllable components found on device (total components: %d)", len(components))
	}

	return controllable, nil
}

// isControllable returns true if the component type can be controlled.
func isControllable(t model.ComponentType) bool {
	switch t {
	case model.ComponentSwitch, model.ComponentLight, model.ComponentRGB, model.ComponentCover:
		return true
	default:
		return false
	}
}

// selectComponents selects which components to control.
// If componentID is nil, returns all controllable components.
// If componentID is set, returns only the component with that ID.
func selectComponents(controllable []model.Component, componentID *int) []model.Component {
	if componentID == nil {
		return controllable
	}
	for _, comp := range controllable {
		if comp.ID == *componentID {
			return []model.Component{comp}
		}
	}
	return nil
}

// ComponentControlResult holds the result of controlling a single component.
type ComponentControlResult struct {
	Type    model.ComponentType
	ID      int
	Success bool
	Err     error
}

// ControlAllComponents performs an action on all controllable components.
// Returns detailed results for each component.
func (s *Service) ControlAllComponents(ctx context.Context, device, action string) ([]ComponentControlResult, error) {
	var results []ComponentControlResult

	err := s.WithConnection(ctx, device, func(conn *client.Client) error {
		comps, err := conn.ListComponents(ctx)
		if err != nil {
			return err
		}

		for _, comp := range comps {
			if !isControllable(comp.Type) {
				continue
			}

			var opErr error
			switch comp.Type {
			case model.ComponentSwitch:
				opErr = controlSwitchAction(ctx, conn, comp.ID, action)
			case model.ComponentLight:
				opErr = controlLightAction(ctx, conn, comp.ID, action)
			case model.ComponentRGB:
				opErr = controlRGBAction(ctx, conn, comp.ID, action)
			case model.ComponentCover:
				opErr = controlCoverAction(ctx, conn, comp.ID, action)
			default:
				continue // Skip non-controllable components
			}

			results = append(results, ComponentControlResult{
				Type:    comp.Type,
				ID:      comp.ID,
				Success: opErr == nil,
				Err:     opErr,
			})
		}

		return nil
	})

	return results, err
}

func controlSwitchAction(ctx context.Context, c *client.Client, id int, action string) error {
	switch action {
	case ActionOn:
		return c.Switch(id).On(ctx)
	case ActionOff:
		return c.Switch(id).Off(ctx)
	case ActionToggle:
		_, err := c.Switch(id).Toggle(ctx)
		return err
	default:
		return nil
	}
}

func controlLightAction(ctx context.Context, c *client.Client, id int, action string) error {
	switch action {
	case ActionOn:
		return c.Light(id).On(ctx)
	case ActionOff:
		return c.Light(id).Off(ctx)
	case ActionToggle:
		_, err := c.Light(id).Toggle(ctx)
		return err
	default:
		return nil
	}
}

func controlRGBAction(ctx context.Context, c *client.Client, id int, action string) error {
	switch action {
	case ActionOn:
		return c.RGB(id).On(ctx)
	case ActionOff:
		return c.RGB(id).Off(ctx)
	case ActionToggle:
		_, err := c.RGB(id).Toggle(ctx)
		return err
	default:
		return nil
	}
}

func controlCoverAction(ctx context.Context, c *client.Client, id int, action string) error {
	switch action {
	case ActionOn:
		return c.Cover(id).Open(ctx, nil)
	case ActionOff:
		return c.Cover(id).Close(ctx, nil)
	case ActionToggle:
		return c.Cover(id).Stop(ctx)
	default:
		return nil
	}
}

// toggleCover toggles a cover based on its current state.
func toggleCover(ctx context.Context, cover *client.CoverComponent) error {
	status, err := cover.GetStatus(ctx)
	if err != nil {
		return err
	}

	switch status.State {
	case "open", "opening":
		return cover.Close(ctx, nil)
	case "closed", "closing":
		return cover.Open(ctx, nil)
	default:
		// If stopped mid-way or unknown, open
		return cover.Open(ctx, nil)
	}
}

// quickOnGen2 turns on controllable components on a Gen2+ device.
func quickOnGen2(ctx context.Context, conn *client.Client, componentID *int, result *QuickResult) error {
	controllable, err := findControllable(ctx, conn)
	if err != nil {
		return err
	}

	toControl := selectComponents(controllable, componentID)
	if len(toControl) == 0 && componentID != nil {
		return fmt.Errorf("component ID %d not found on device", *componentID)
	}

	for _, comp := range toControl {
		var opErr error
		switch comp.Type {
		case model.ComponentSwitch:
			opErr = conn.Switch(comp.ID).On(ctx)
		case model.ComponentLight:
			opErr = conn.Light(comp.ID).On(ctx)
		case model.ComponentRGB:
			opErr = conn.RGB(comp.ID).On(ctx)
		case model.ComponentCover:
			opErr = conn.Cover(comp.ID).Open(ctx, nil)
		default:
			continue
		}
		if opErr != nil {
			return fmt.Errorf("failed to turn on %s:%d: %w", comp.Type, comp.ID, opErr)
		}
		result.Count++
	}

	return nil
}

// quickOffGen2 turns off controllable components on a Gen2+ device.
func quickOffGen2(ctx context.Context, conn *client.Client, componentID *int, result *QuickResult) error {
	controllable, err := findControllable(ctx, conn)
	if err != nil {
		return err
	}

	toControl := selectComponents(controllable, componentID)
	if len(toControl) == 0 && componentID != nil {
		return fmt.Errorf("component ID %d not found on device", *componentID)
	}

	for _, comp := range toControl {
		var opErr error
		switch comp.Type {
		case model.ComponentSwitch:
			opErr = conn.Switch(comp.ID).Off(ctx)
		case model.ComponentLight:
			opErr = conn.Light(comp.ID).Off(ctx)
		case model.ComponentRGB:
			opErr = conn.RGB(comp.ID).Off(ctx)
		case model.ComponentCover:
			opErr = conn.Cover(comp.ID).Close(ctx, nil)
		default:
			continue
		}
		if opErr != nil {
			return fmt.Errorf("failed to turn off %s:%d: %w", comp.Type, comp.ID, opErr)
		}
		result.Count++
	}

	return nil
}

// quickToggleGen2 toggles controllable components on a Gen2+ device.
func quickToggleGen2(ctx context.Context, conn *client.Client, componentID *int, result *QuickResult) error {
	controllable, err := findControllable(ctx, conn)
	if err != nil {
		return err
	}

	toControl := selectComponents(controllable, componentID)
	if len(toControl) == 0 && componentID != nil {
		return fmt.Errorf("component ID %d not found on device", *componentID)
	}

	debug.TraceEvent("QuickToggle: found %d controllable, toggling %d components", len(controllable), len(toControl))

	for _, comp := range toControl {
		debug.TraceEvent("QuickToggle: toggling %s:%d", comp.Type, comp.ID)
		var opErr error
		switch comp.Type {
		case model.ComponentSwitch:
			_, opErr = conn.Switch(comp.ID).Toggle(ctx)
		case model.ComponentLight:
			_, opErr = conn.Light(comp.ID).Toggle(ctx)
		case model.ComponentRGB:
			_, opErr = conn.RGB(comp.ID).Toggle(ctx)
		case model.ComponentCover:
			opErr = toggleCover(ctx, conn.Cover(comp.ID))
		default:
			continue
		}
		if opErr != nil {
			debug.TraceEvent("QuickToggle: error toggling %s:%d: %v", comp.Type, comp.ID, opErr)
			return fmt.Errorf("failed to toggle %s:%d: %w", comp.Type, comp.ID, opErr)
		}
		debug.TraceEvent("QuickToggle: toggled %s:%d successfully", comp.Type, comp.ID)
		result.Count++
	}

	return nil
}

// quickOnGen1 turns on controllable components on a Gen1 device.
func quickOnGen1(ctx context.Context, conn *client.Gen1Client, componentID *int, result *QuickResult) error {
	controllable, err := findControllableGen1(ctx, conn)
	if err != nil {
		return err
	}

	toControl := selectComponents(controllable, componentID)
	if len(toControl) == 0 && componentID != nil {
		return fmt.Errorf("component ID %d not found on device", *componentID)
	}

	for _, comp := range toControl {
		var opErr error
		switch comp.Type {
		case model.ComponentSwitch:
			relay, relayErr := conn.Relay(comp.ID)
			if relayErr != nil {
				opErr = relayErr
			} else {
				opErr = relay.TurnOn(ctx)
			}
		case model.ComponentLight:
			light, lightErr := conn.Light(comp.ID)
			if lightErr != nil {
				opErr = lightErr
			} else {
				opErr = light.TurnOn(ctx)
			}
		case model.ComponentRGB:
			color, colorErr := conn.Color(comp.ID)
			if colorErr != nil {
				opErr = colorErr
			} else {
				opErr = color.TurnOn(ctx)
			}
		case model.ComponentCover:
			roller, rollerErr := conn.Roller(comp.ID)
			if rollerErr != nil {
				opErr = rollerErr
			} else {
				opErr = roller.Open(ctx)
			}
		default:
			continue
		}
		if opErr != nil {
			return fmt.Errorf("failed to turn on %s:%d: %w", comp.Type, comp.ID, opErr)
		}
		result.Count++
	}

	return nil
}

// quickOffGen1 turns off controllable components on a Gen1 device.
func quickOffGen1(ctx context.Context, conn *client.Gen1Client, componentID *int, result *QuickResult) error {
	controllable, err := findControllableGen1(ctx, conn)
	if err != nil {
		return err
	}

	toControl := selectComponents(controllable, componentID)
	if len(toControl) == 0 && componentID != nil {
		return fmt.Errorf("component ID %d not found on device", *componentID)
	}

	for _, comp := range toControl {
		var opErr error
		switch comp.Type {
		case model.ComponentSwitch:
			relay, relayErr := conn.Relay(comp.ID)
			if relayErr != nil {
				opErr = relayErr
			} else {
				opErr = relay.TurnOff(ctx)
			}
		case model.ComponentLight:
			light, lightErr := conn.Light(comp.ID)
			if lightErr != nil {
				opErr = lightErr
			} else {
				opErr = light.TurnOff(ctx)
			}
		case model.ComponentRGB:
			color, colorErr := conn.Color(comp.ID)
			if colorErr != nil {
				opErr = colorErr
			} else {
				opErr = color.TurnOff(ctx)
			}
		case model.ComponentCover:
			roller, rollerErr := conn.Roller(comp.ID)
			if rollerErr != nil {
				opErr = rollerErr
			} else {
				opErr = roller.Close(ctx)
			}
		default:
			continue
		}
		if opErr != nil {
			return fmt.Errorf("failed to turn off %s:%d: %w", comp.Type, comp.ID, opErr)
		}
		result.Count++
	}

	return nil
}

// quickToggleGen1 toggles controllable components on a Gen1 device.
func quickToggleGen1(ctx context.Context, conn *client.Gen1Client, componentID *int, result *QuickResult) error {
	controllable, err := findControllableGen1(ctx, conn)
	if err != nil {
		return err
	}

	toControl := selectComponents(controllable, componentID)
	if len(toControl) == 0 && componentID != nil {
		return fmt.Errorf("component ID %d not found on device", *componentID)
	}

	for _, comp := range toControl {
		var opErr error
		switch comp.Type {
		case model.ComponentSwitch:
			relay, relayErr := conn.Relay(comp.ID)
			if relayErr != nil {
				opErr = relayErr
			} else {
				opErr = relay.Toggle(ctx)
			}
		case model.ComponentLight:
			light, lightErr := conn.Light(comp.ID)
			if lightErr != nil {
				opErr = lightErr
			} else {
				opErr = light.Toggle(ctx)
			}
		case model.ComponentRGB:
			color, colorErr := conn.Color(comp.ID)
			if colorErr != nil {
				opErr = colorErr
			} else {
				opErr = color.Toggle(ctx)
			}
		case model.ComponentCover:
			roller, rollerErr := conn.Roller(comp.ID)
			if rollerErr != nil {
				opErr = rollerErr
			} else {
				opErr = roller.Stop(ctx)
			}
		default:
			continue
		}
		if opErr != nil {
			return fmt.Errorf("failed to toggle %s:%d: %w", comp.Type, comp.ID, opErr)
		}
		result.Count++
	}

	return nil
}

// findControllableGen1 discovers controllable components on a Gen1 device.
// Gen1 devices don't have a ListComponents API, so we query settings to
// discover which components are available.
func findControllableGen1(ctx context.Context, conn *client.Gen1Client) ([]model.Component, error) {
	settings, err := conn.GetSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get device settings: %w", err)
	}

	var controllable []model.Component

	// Check relays (switches)
	if settings.Relays != nil {
		for i := range settings.Relays {
			controllable = append(controllable, model.Component{
				Type: model.ComponentSwitch,
				ID:   i,
			})
		}
	}

	// Check lights
	if settings.Lights != nil {
		for i := range settings.Lights {
			controllable = append(controllable, model.Component{
				Type: model.ComponentLight,
				ID:   i,
			})
		}
	}

	// Check rollers (covers)
	if settings.Rollers != nil {
		for i := range settings.Rollers {
			controllable = append(controllable, model.Component{
				Type: model.ComponentCover,
				ID:   i,
			})
		}
	}

	if len(controllable) == 0 {
		return nil, fmt.Errorf("no controllable components found on Gen1 device")
	}

	return controllable, nil
}

// PartyColors defines RGB colors for party mode.
var PartyColors = []struct{ R, G, B int }{
	{255, 0, 0},     // Red
	{0, 255, 0},     // Green
	{0, 0, 255},     // Blue
	{255, 255, 0},   // Yellow
	{255, 0, 255},   // Magenta
	{0, 255, 255},   // Cyan
	{255, 128, 0},   // Orange
	{128, 0, 255},   // Purple
	{255, 255, 255}, // White
}

// PartyToggleDevice handles toggling a single device on or off with fallback to switch.
func (s *Service) PartyToggleDevice(ctx context.Context, ios IOStreamsDebugger, dev string, on bool) {
	if on {
		s.PartyToggleOn(ctx, ios, dev)
	} else {
		s.PartyToggleOff(ctx, ios, dev)
	}
}

// PartyToggleOn turns a device on with light/switch fallback and sets random color.
func (s *Service) PartyToggleOn(ctx context.Context, ios IOStreamsDebugger, dev string) {
	if err := s.LightOn(ctx, dev, 0); err != nil {
		// Try as switch (expected to fail for non-switch devices)
		if switchErr := s.SwitchOn(ctx, dev, 0); switchErr != nil {
			ios.DebugErr("party toggle on "+dev, switchErr)
		}
	}

	// Try to set random color for RGB lights (expected to fail for non-RGB)
	color := PartyColors[rand.Intn(len(PartyColors))] //nolint:gosec // Not crypto, just random colors
	if rgbErr := s.RGBColor(ctx, dev, 0, color.R, color.G, color.B); rgbErr != nil {
		ios.DebugErr("party RGB "+dev, rgbErr)
	}
}

// PartyToggleOff turns a device off with light/switch fallback.
func (s *Service) PartyToggleOff(ctx context.Context, ios IOStreamsDebugger, dev string) {
	if err := s.LightOff(ctx, dev, 0); err != nil {
		// Try as switch (expected to fail for non-switch devices)
		if switchErr := s.SwitchOff(ctx, dev, 0); switchErr != nil {
			ios.DebugErr("party toggle off "+dev, switchErr)
		}
	}
}

// IOStreamsDebugger is an interface for debug error logging.
type IOStreamsDebugger interface {
	DebugErr(context string, err error)
}
