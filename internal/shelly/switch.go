// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"fmt"

	gen1comp "github.com/tj-smith47/shelly-go/gen1/components"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// SwitchInfo holds switch information for list operations.
type SwitchInfo struct {
	ID     int
	Name   string
	Output bool
	Power  float64
}

// ListHeaders returns the column headers for the table.
func (s SwitchInfo) ListHeaders() []string {
	return []string{"ID", "Name", "State", "Power"}
}

// ListRow returns the formatted row values for the table.
func (s SwitchInfo) ListRow() []string {
	name := output.FormatComponentName(s.Name, "switch", s.ID)
	state := output.RenderOnOff(s.Output, output.CaseUpper, theme.FalseError)
	power := output.FormatPowerTableValue(s.Power)
	return []string{fmt.Sprintf("%d", s.ID), name, state, power}
}

// SwitchOn turns on a switch component.
// For Gen1 devices, this controls the relay.
func (s *Service) SwitchOn(ctx context.Context, identifier string, switchID int) error {
	return s.withGenAwareAction(ctx, identifier,
		func(conn *client.Gen1Client) error {
			relay, err := conn.Relay(switchID)
			if err != nil {
				return err
			}
			return relay.TurnOn(ctx)
		},
		func(conn *client.Client) error {
			return conn.Switch(switchID).On(ctx)
		},
	)
}

// SwitchOff turns off a switch component.
// For Gen1 devices, this controls the relay.
func (s *Service) SwitchOff(ctx context.Context, identifier string, switchID int) error {
	return s.withGenAwareAction(ctx, identifier,
		func(conn *client.Gen1Client) error {
			relay, err := conn.Relay(switchID)
			if err != nil {
				return err
			}
			return relay.TurnOff(ctx)
		},
		func(conn *client.Client) error {
			return conn.Switch(switchID).Off(ctx)
		},
	)
}

// SwitchToggle toggles a switch component and returns the new status.
// For Gen1 devices, this controls the relay.
func (s *Service) SwitchToggle(ctx context.Context, identifier string, switchID int) (*model.SwitchStatus, error) {
	var result *model.SwitchStatus
	err := s.WithDevice(ctx, identifier, func(dev *DeviceClient) error {
		if dev.IsGen1() {
			relay, err := dev.Gen1().Relay(switchID)
			if err != nil {
				return err
			}
			if err := relay.Toggle(ctx); err != nil {
				return err
			}
			// Get status after toggle
			status, err := relay.GetStatus(ctx)
			if err != nil {
				return err
			}
			result = gen1RelayStatusToSwitch(switchID, status)
			return nil
		}

		// Gen2+
		status, err := dev.Gen2().Switch(switchID).Toggle(ctx)
		if err != nil {
			return err
		}
		// Get current status after toggle.
		result, err = dev.Gen2().Switch(switchID).GetStatus(ctx)
		if err != nil {
			// Fall back to toggle result.
			result = status
		}
		return nil
	})
	return result, err
}

// SwitchStatus gets the status of a switch component.
// For Gen1 devices, this returns relay status.
func (s *Service) SwitchStatus(ctx context.Context, identifier string, switchID int) (*model.SwitchStatus, error) {
	var result *model.SwitchStatus
	err := s.WithDevice(ctx, identifier, func(dev *DeviceClient) error {
		if dev.IsGen1() {
			relay, err := dev.Gen1().Relay(switchID)
			if err != nil {
				return err
			}
			status, err := relay.GetStatus(ctx)
			if err != nil {
				return err
			}
			result = gen1RelayStatusToSwitch(switchID, status)
			return nil
		}

		// Gen2+
		status, err := dev.Gen2().Switch(switchID).GetStatus(ctx)
		if err != nil {
			return err
		}
		result = status
		return nil
	})
	return result, err
}

// SwitchList lists all switch components on a device with their status.
// Note: Gen1 devices don't have a component enumeration API, so this only works for Gen2+.
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

// gen1RelayStatusToSwitch converts Gen1 relay status to model.SwitchStatus.
func gen1RelayStatusToSwitch(id int, status *gen1comp.RelayStatus) *model.SwitchStatus {
	return &model.SwitchStatus{
		ID:        id,
		Output:    status.IsOn,
		Source:    status.Source,
		Overpower: status.Overpower,
	}
}
