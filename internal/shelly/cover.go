// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"fmt"

	gen1comp "github.com/tj-smith47/shelly-go/gen1/components"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// Component list table column headers.
const (
	headerName       = "Name"
	headerState      = "State"
	headerPower      = "Power"
	headerBrightness = "Brightness"
)

// CoverInfo holds cover information for list operations.
type CoverInfo struct {
	ID       int     `json:"id" yaml:"id"`
	Name     string  `json:"name" yaml:"name"`
	State    string  `json:"state" yaml:"state"`
	Position int     `json:"current_pos" yaml:"current_pos"`
	Power    float64 `json:"power" yaml:"power"`
}

// ListHeaders returns the column headers for the table.
func (c CoverInfo) ListHeaders() []string {
	return []string{"ID", headerName, headerState, "Position", headerPower}
}

// ListRow returns the formatted row values for the table.
func (c CoverInfo) ListRow() []string {
	name := output.FormatComponentName(c.Name, "cover", c.ID)
	state := output.RenderCoverState(c.State)

	position := "-"
	if c.Position >= 0 {
		position = fmt.Sprintf("%d%%", c.Position)
	}

	power := output.FormatPowerTableValue(c.Power)
	return []string{fmt.Sprintf("%d", c.ID), name, state, position, power}
}

// CoverOpen opens a cover component with optional duration in seconds.
// For Gen1 devices, this controls the roller. Duration is only supported on Gen2+.
func (s *Service) CoverOpen(ctx context.Context, identifier string, coverID int, duration *int) error {
	return s.withComponentAction(ctx, identifier,
		func(conn *client.Gen1Client) error {
			roller, err := conn.Roller(coverID)
			if err != nil {
				return err
			}
			// Gen1 doesn't support duration in basic Open, use OpenForDuration if provided
			if duration != nil && *duration > 0 {
				return roller.OpenForDuration(ctx, float64(*duration))
			}
			return roller.Open(ctx)
		},
		func(conn *client.Client) error {
			return conn.Cover(coverID).Open(ctx, duration)
		},
	)
}

// CoverClose closes a cover component with optional duration in seconds.
// For Gen1 devices, this controls the roller.
func (s *Service) CoverClose(ctx context.Context, identifier string, coverID int, duration *int) error {
	return s.withComponentAction(ctx, identifier,
		func(conn *client.Gen1Client) error {
			roller, err := conn.Roller(coverID)
			if err != nil {
				return err
			}
			if duration != nil && *duration > 0 {
				return roller.CloseForDuration(ctx, float64(*duration))
			}
			return roller.Close(ctx)
		},
		func(conn *client.Client) error {
			return conn.Cover(coverID).Close(ctx, duration)
		},
	)
}

// CoverStop stops a cover component.
// For Gen1 devices, this controls the roller.
func (s *Service) CoverStop(ctx context.Context, identifier string, coverID int) error {
	return s.withComponentAction(ctx, identifier,
		func(conn *client.Gen1Client) error {
			roller, err := conn.Roller(coverID)
			if err != nil {
				return err
			}
			return roller.Stop(ctx)
		},
		func(conn *client.Client) error {
			return conn.Cover(coverID).Stop(ctx)
		},
	)
}

// CoverPosition moves a cover to a specific position (0-100).
// For Gen1 devices, this controls the roller position.
func (s *Service) CoverPosition(ctx context.Context, identifier string, coverID, position int) error {
	return s.withComponentAction(ctx, identifier,
		func(conn *client.Gen1Client) error {
			roller, err := conn.Roller(coverID)
			if err != nil {
				return err
			}
			return roller.GoToPosition(ctx, position)
		},
		func(conn *client.Client) error {
			return conn.Cover(coverID).GoToPosition(ctx, position)
		},
	)
}

// CoverStatus gets the status of a cover component.
// For Gen1 devices, this returns roller status.
func (s *Service) CoverStatus(ctx context.Context, identifier string, coverID int) (*model.CoverStatus, error) {
	var result *model.CoverStatus
	err := s.WithDevice(ctx, identifier, func(dev *DeviceClient) error {
		if dev.IsGen1() {
			roller, err := dev.Gen1().Roller(coverID)
			if err != nil {
				return err
			}
			status, err := roller.GetStatus(ctx)
			if err != nil {
				return err
			}
			result = gen1RollerStatusToCover(coverID, status)
			return nil
		}

		// Gen2+
		status, err := dev.Gen2().Cover(coverID).GetStatus(ctx)
		if err != nil {
			return err
		}
		result = status
		return nil
	})
	return result, err
}

// CoverCalibrate starts cover calibration.
// For Gen1 devices, this starts roller calibration.
func (s *Service) CoverCalibrate(ctx context.Context, identifier string, coverID int) error {
	return s.withComponentAction(ctx, identifier,
		func(conn *client.Gen1Client) error {
			roller, err := conn.Roller(coverID)
			if err != nil {
				return err
			}
			return roller.Calibrate(ctx)
		},
		func(conn *client.Client) error {
			return conn.Cover(coverID).Calibrate(ctx)
		},
	)
}

// CoverList lists all cover components on a device with their status.
// Gen2+ devices enumerate their components; a Gen1 device lists its rollers
// from /settings.
func (s *Service) CoverList(ctx context.Context, identifier string) ([]CoverInfo, error) {
	var result []CoverInfo
	err := s.withGenAwareAction(ctx, identifier, func(conn *client.Gen1Client) error {
		names, err := gen1ComponentNames(ctx, conn, model.ComponentCover)
		if err != nil {
			return err
		}
		result = make([]CoverInfo, 0, len(names))
		for id, name := range names {
			roller, err := conn.Roller(id)
			if err != nil {
				continue
			}
			status, err := roller.GetStatus(ctx)
			if err != nil {
				continue
			}
			result = append(result, coverInfo(id, name, gen1RollerStatusToCover(id, status)))
		}
		return nil
	}, func(conn *client.Client) error {
		components, err := conn.FilterComponents(ctx, model.ComponentCover)
		if err != nil {
			return err
		}

		result = make([]CoverInfo, 0, len(components))
		for _, comp := range components {
			status, err := conn.Cover(comp.ID).GetStatus(ctx)
			if err != nil {
				continue
			}
			name := ""
			config, err := conn.Cover(comp.ID).GetConfig(ctx)
			if err == nil && config.Name != nil {
				name = *config.Name
			}
			result = append(result, coverInfo(comp.ID, name, status))
		}

		return nil
	})
	return result, err
}

// coverInfo builds a list row from a cover status.
func coverInfo(id int, name string, status *model.CoverStatus) CoverInfo {
	info := CoverInfo{ID: id, Name: name, State: status.State}
	if status.CurrentPosition != nil {
		info.Position = *status.CurrentPosition
	}
	if status.Power != nil {
		info.Power = *status.Power
	}
	return info
}

// gen1RollerStatusToCover converts Gen1 roller status to model.CoverStatus.
func gen1RollerStatusToCover(id int, status *gen1comp.RollerStatus) *model.CoverStatus {
	result := &model.CoverStatus{
		ID:          id,
		State:       status.State,
		Calibrating: status.Calibrating,
	}
	if status.CurrentPos >= 0 && status.IsValid {
		result.CurrentPosition = &status.CurrentPos
	}
	// Always surface the reading: an idle roller legitimately reports 0 W, and the
	// prior `> 0` guard hid it so `cover status` printed "-" instead of "0.0 W".
	power := status.Power
	result.Power = &power
	return result
}
