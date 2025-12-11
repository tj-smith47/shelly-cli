// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// LightInfo holds light information for list operations.
type LightInfo struct {
	ID         int
	Name       string
	Output     bool
	Brightness int
	Power      float64
}

// LightOn turns on a light component.
func (s *Service) LightOn(ctx context.Context, identifier string, lightID int) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.Light(lightID).On(ctx)
	})
}

// LightOff turns off a light component.
func (s *Service) LightOff(ctx context.Context, identifier string, lightID int) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.Light(lightID).Off(ctx)
	})
}

// LightToggle toggles a light component and returns the new status.
func (s *Service) LightToggle(ctx context.Context, identifier string, lightID int) (*model.LightStatus, error) {
	var result *model.LightStatus
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		_, err := conn.Light(lightID).Toggle(ctx)
		if err != nil {
			return err
		}
		// Get current status after toggle.
		result, err = conn.Light(lightID).GetStatus(ctx)
		return err
	})
	return result, err
}

// LightBrightness sets the brightness of a light component (0-100).
func (s *Service) LightBrightness(ctx context.Context, identifier string, lightID, brightness int) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.Light(lightID).SetBrightness(ctx, brightness)
	})
}

// LightStatus gets the status of a light component.
func (s *Service) LightStatus(ctx context.Context, identifier string, lightID int) (*model.LightStatus, error) {
	var result *model.LightStatus
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		status, err := conn.Light(lightID).GetStatus(ctx)
		if err != nil {
			return err
		}
		result = status
		return nil
	})
	return result, err
}

// LightSet sets parameters of a light component.
func (s *Service) LightSet(ctx context.Context, identifier string, lightID int, brightness *int, on *bool) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.Light(lightID).Set(ctx, brightness, on)
	})
}

// LightList lists all light components on a device with their status.
func (s *Service) LightList(ctx context.Context, identifier string) ([]LightInfo, error) {
	var result []LightInfo
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		components, err := conn.FilterComponents(ctx, model.ComponentLight)
		if err != nil {
			return err
		}

		result = make([]LightInfo, 0, len(components))
		for _, comp := range components {
			info := LightInfo{ID: comp.ID, Brightness: -1}

			status, err := conn.Light(comp.ID).GetStatus(ctx)
			if err != nil {
				continue
			}
			info.Output = status.Output
			if status.Brightness != nil {
				info.Brightness = *status.Brightness
			}
			if status.Power != nil {
				info.Power = *status.Power
			}

			config, err := conn.Light(comp.ID).GetConfig(ctx)
			if err == nil && config.Name != nil {
				info.Name = *config.Name
			}

			result = append(result, info)
		}

		return nil
	})
	return result, err
}
