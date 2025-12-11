// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// RGBInfo holds RGB information for list operations.
type RGBInfo struct {
	ID         int
	Name       string
	Output     bool
	Brightness int
	Red        int
	Green      int
	Blue       int
	Power      float64
}

// RGBSetParams holds parameters for RGBSet operation.
type RGBSetParams struct {
	Red        *int
	Green      *int
	Blue       *int
	Brightness *int
	On         *bool
}

// RGBOn turns on an RGB component.
func (s *Service) RGBOn(ctx context.Context, identifier string, rgbID int) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.RGB(rgbID).On(ctx)
	})
}

// RGBOff turns off an RGB component.
func (s *Service) RGBOff(ctx context.Context, identifier string, rgbID int) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.RGB(rgbID).Off(ctx)
	})
}

// RGBToggle toggles an RGB component and returns the new status.
func (s *Service) RGBToggle(ctx context.Context, identifier string, rgbID int) (*model.RGBStatus, error) {
	var result *model.RGBStatus
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		_, err := conn.RGB(rgbID).Toggle(ctx)
		if err != nil {
			return err
		}
		// Get current status after toggle.
		result, err = conn.RGB(rgbID).GetStatus(ctx)
		return err
	})
	return result, err
}

// RGBBrightness sets the brightness of an RGB component (0-100).
func (s *Service) RGBBrightness(ctx context.Context, identifier string, rgbID, brightness int) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.RGB(rgbID).SetBrightness(ctx, brightness)
	})
}

// RGBColor sets the color of an RGB component.
func (s *Service) RGBColor(ctx context.Context, identifier string, rgbID, r, g, b int) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.RGB(rgbID).SetColor(ctx, r, g, b)
	})
}

// RGBColorAndBrightness sets both color and brightness of an RGB component.
func (s *Service) RGBColorAndBrightness(ctx context.Context, identifier string, rgbID, r, g, b, brightness int) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.RGB(rgbID).SetColorAndBrightness(ctx, r, g, b, brightness)
	})
}

// RGBStatus gets the status of an RGB component.
func (s *Service) RGBStatus(ctx context.Context, identifier string, rgbID int) (*model.RGBStatus, error) {
	var result *model.RGBStatus
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		status, err := conn.RGB(rgbID).GetStatus(ctx)
		if err != nil {
			return err
		}
		result = status
		return nil
	})
	return result, err
}

// RGBSet sets parameters of an RGB component.
func (s *Service) RGBSet(ctx context.Context, identifier string, rgbID int, params RGBSetParams) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.RGB(rgbID).Set(ctx, params.Red, params.Green, params.Blue, params.Brightness, params.On)
	})
}

// RGBList lists all RGB components on a device with their status.
func (s *Service) RGBList(ctx context.Context, identifier string) ([]RGBInfo, error) {
	var result []RGBInfo
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		components, err := conn.FilterComponents(ctx, model.ComponentRGB)
		if err != nil {
			return err
		}

		result = make([]RGBInfo, 0, len(components))
		for _, comp := range components {
			info := RGBInfo{ID: comp.ID, Brightness: -1}

			status, err := conn.RGB(comp.ID).GetStatus(ctx)
			if err != nil {
				continue
			}
			info.Output = status.Output
			if status.Brightness != nil {
				info.Brightness = *status.Brightness
			}
			if status.RGB != nil {
				info.Red = status.RGB.Red
				info.Green = status.RGB.Green
				info.Blue = status.RGB.Blue
			}
			if status.Power != nil {
				info.Power = *status.Power
			}

			config, err := conn.RGB(comp.ID).GetConfig(ctx)
			if err == nil && config.Name != nil {
				info.Name = *config.Name
			}

			result = append(result, info)
		}

		return nil
	})
	return result, err
}
