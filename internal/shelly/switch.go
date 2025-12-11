// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// SwitchInfo holds switch information for list operations.
type SwitchInfo struct {
	ID     int
	Name   string
	Output bool
	Power  float64
}

// SwitchOn turns on a switch component.
func (s *Service) SwitchOn(ctx context.Context, identifier string, switchID int) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.Switch(switchID).On(ctx)
	})
}

// SwitchOff turns off a switch component.
func (s *Service) SwitchOff(ctx context.Context, identifier string, switchID int) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.Switch(switchID).Off(ctx)
	})
}

// SwitchToggle toggles a switch component and returns the new status.
func (s *Service) SwitchToggle(ctx context.Context, identifier string, switchID int) (*model.SwitchStatus, error) {
	var result *model.SwitchStatus
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		status, err := conn.Switch(switchID).Toggle(ctx)
		if err != nil {
			return err
		}
		// Get current status after toggle.
		result, err = conn.Switch(switchID).GetStatus(ctx)
		if err != nil {
			// Fall back to toggle result.
			result = status
		}
		return nil
	})
	return result, err
}

// SwitchStatus gets the status of a switch component.
func (s *Service) SwitchStatus(ctx context.Context, identifier string, switchID int) (*model.SwitchStatus, error) {
	var result *model.SwitchStatus
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		status, err := conn.Switch(switchID).GetStatus(ctx)
		if err != nil {
			return err
		}
		result = status
		return nil
	})
	return result, err
}

// SwitchList lists all switch components on a device with their status.
func (s *Service) SwitchList(ctx context.Context, identifier string) ([]SwitchInfo, error) {
	var result []SwitchInfo
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		components, err := conn.FilterComponents(ctx, model.ComponentSwitch)
		if err != nil {
			return err
		}

		result = make([]SwitchInfo, 0, len(components))
		for _, comp := range components {
			info := SwitchInfo{ID: comp.ID}

			status, err := conn.Switch(comp.ID).GetStatus(ctx)
			if err != nil {
				continue
			}
			info.Output = status.Output
			if status.Power != nil {
				info.Power = *status.Power
			}

			config, err := conn.Switch(comp.ID).GetConfig(ctx)
			if err == nil && config.Name != nil {
				info.Name = *config.Name
			}

			result = append(result, info)
		}

		return nil
	})
	return result, err
}
