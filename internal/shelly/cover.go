// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// CoverInfo holds cover information for list operations.
type CoverInfo struct {
	ID       int
	Name     string
	State    string
	Position int
	Power    float64
}

// CoverOpen opens a cover component with optional duration in seconds.
func (s *Service) CoverOpen(ctx context.Context, identifier string, coverID int, duration *int) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.Cover(coverID).Open(ctx, duration)
	})
}

// CoverClose closes a cover component with optional duration in seconds.
func (s *Service) CoverClose(ctx context.Context, identifier string, coverID int, duration *int) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.Cover(coverID).Close(ctx, duration)
	})
}

// CoverStop stops a cover component.
func (s *Service) CoverStop(ctx context.Context, identifier string, coverID int) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.Cover(coverID).Stop(ctx)
	})
}

// CoverPosition moves a cover to a specific position (0-100).
func (s *Service) CoverPosition(ctx context.Context, identifier string, coverID, position int) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.Cover(coverID).GoToPosition(ctx, position)
	})
}

// CoverStatus gets the status of a cover component.
func (s *Service) CoverStatus(ctx context.Context, identifier string, coverID int) (*model.CoverStatus, error) {
	var result *model.CoverStatus
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		status, err := conn.Cover(coverID).GetStatus(ctx)
		if err != nil {
			return err
		}
		result = status
		return nil
	})
	return result, err
}

// CoverCalibrate starts cover calibration.
func (s *Service) CoverCalibrate(ctx context.Context, identifier string, coverID int) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.Cover(coverID).Calibrate(ctx)
	})
}

// CoverList lists all cover components on a device with their status.
func (s *Service) CoverList(ctx context.Context, identifier string) ([]CoverInfo, error) {
	var result []CoverInfo
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		components, err := conn.FilterComponents(ctx, model.ComponentCover)
		if err != nil {
			return err
		}

		result = make([]CoverInfo, 0, len(components))
		for _, comp := range components {
			info := CoverInfo{ID: comp.ID}

			status, err := conn.Cover(comp.ID).GetStatus(ctx)
			if err != nil {
				continue
			}
			info.State = status.State
			if status.CurrentPosition != nil {
				info.Position = *status.CurrentPosition
			}
			if status.Power != nil {
				info.Power = *status.Power
			}

			config, err := conn.Cover(comp.ID).GetConfig(ctx)
			if err == nil && config.Name != nil {
				info.Name = *config.Name
			}

			result = append(result, info)
		}

		return nil
	})
	return result, err
}
