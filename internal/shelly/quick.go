// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
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
}

// QuickOn turns on controllable components on a device.
// If all is true, turns on all controllable components.
// If all is false, turns on only the first controllable component.
func (s *Service) QuickOn(ctx context.Context, device string, all bool) (*QuickResult, error) {
	result := &QuickResult{}

	err := s.WithConnection(ctx, device, func(conn *client.Client) error {
		controllable, err := findControllable(ctx, conn)
		if err != nil {
			return err
		}

		toControl := selectComponents(controllable, all)

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
	})

	return result, err
}

// QuickOff turns off controllable components on a device.
// If all is true, turns off all controllable components.
// If all is false, turns off only the first controllable component.
func (s *Service) QuickOff(ctx context.Context, device string, all bool) (*QuickResult, error) {
	result := &QuickResult{}

	err := s.WithConnection(ctx, device, func(conn *client.Client) error {
		controllable, err := findControllable(ctx, conn)
		if err != nil {
			return err
		}

		toControl := selectComponents(controllable, all)

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
	})

	return result, err
}

// QuickToggle toggles controllable components on a device.
// If all is true, toggles all controllable components.
// If all is false, toggles only the first controllable component.
func (s *Service) QuickToggle(ctx context.Context, device string, all bool) (*QuickResult, error) {
	result := &QuickResult{}

	err := s.WithConnection(ctx, device, func(conn *client.Client) error {
		controllable, err := findControllable(ctx, conn)
		if err != nil {
			return err
		}

		toControl := selectComponents(controllable, all)

		for _, comp := range toControl {
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
				return fmt.Errorf("failed to toggle %s:%d: %w", comp.Type, comp.ID, opErr)
			}
			result.Count++
		}

		return nil
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
		return nil, fmt.Errorf("no controllable components found on device")
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

// selectComponents selects which components to control based on the all flag.
func selectComponents(controllable []model.Component, all bool) []model.Component {
	if !all && len(controllable) > 1 {
		return controllable[:1]
	}
	return controllable
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
