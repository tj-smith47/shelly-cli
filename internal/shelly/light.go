// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"

	gen1comp "github.com/tj-smith47/shelly-go/gen1/components"

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
// For Gen1 devices, this controls the light/dimmer.
func (s *Service) LightOn(ctx context.Context, identifier string, lightID int) error {
	return s.withGenAwareAction(ctx, identifier,
		func(conn *client.Gen1Client) error {
			light, err := conn.Light(lightID)
			if err != nil {
				return err
			}
			return light.TurnOn(ctx)
		},
		func(conn *client.Client) error {
			return conn.Light(lightID).On(ctx)
		},
	)
}

// LightOff turns off a light component.
// For Gen1 devices, this controls the light/dimmer.
func (s *Service) LightOff(ctx context.Context, identifier string, lightID int) error {
	return s.withGenAwareAction(ctx, identifier,
		func(conn *client.Gen1Client) error {
			light, err := conn.Light(lightID)
			if err != nil {
				return err
			}
			return light.TurnOff(ctx)
		},
		func(conn *client.Client) error {
			return conn.Light(lightID).Off(ctx)
		},
	)
}

// LightToggle toggles a light component and returns the new status.
// For Gen1 devices, this controls the light/dimmer.
func (s *Service) LightToggle(ctx context.Context, identifier string, lightID int) (*model.LightStatus, error) {
	isGen1, _, err := s.IsGen1Device(ctx, identifier)
	if err != nil {
		return nil, err
	}

	if isGen1 {
		var result *model.LightStatus
		err := s.WithGen1Connection(ctx, identifier, func(conn *client.Gen1Client) error {
			light, err := conn.Light(lightID)
			if err != nil {
				return err
			}
			if err := light.Toggle(ctx); err != nil {
				return err
			}
			status, err := light.GetStatus(ctx)
			if err != nil {
				return err
			}
			result = gen1LightStatusToLight(lightID, status)
			return nil
		})
		return result, err
	}

	var result *model.LightStatus
	err = s.WithConnection(ctx, identifier, func(conn *client.Client) error {
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
// For Gen1 devices, this controls the light/dimmer brightness.
func (s *Service) LightBrightness(ctx context.Context, identifier string, lightID, brightness int) error {
	return s.withGenAwareAction(ctx, identifier,
		func(conn *client.Gen1Client) error {
			light, err := conn.Light(lightID)
			if err != nil {
				return err
			}
			return light.SetBrightness(ctx, brightness)
		},
		func(conn *client.Client) error {
			return conn.Light(lightID).SetBrightness(ctx, brightness)
		},
	)
}

// LightStatus gets the status of a light component.
// For Gen1 devices, this returns light/dimmer status.
func (s *Service) LightStatus(ctx context.Context, identifier string, lightID int) (*model.LightStatus, error) {
	isGen1, _, err := s.IsGen1Device(ctx, identifier)
	if err != nil {
		return nil, err
	}

	if isGen1 {
		var result *model.LightStatus
		err := s.WithGen1Connection(ctx, identifier, func(conn *client.Gen1Client) error {
			light, err := conn.Light(lightID)
			if err != nil {
				return err
			}
			status, err := light.GetStatus(ctx)
			if err != nil {
				return err
			}
			result = gen1LightStatusToLight(lightID, status)
			return nil
		})
		return result, err
	}

	var result *model.LightStatus
	err = s.WithConnection(ctx, identifier, func(conn *client.Client) error {
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
// For Gen1 devices, use LightOn/LightOff/LightBrightness instead (Gen1 doesn't support combined set).
func (s *Service) LightSet(ctx context.Context, identifier string, lightID int, brightness *int, on *bool) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.Light(lightID).Set(ctx, brightness, on)
	})
}

// LightList lists all light components on a device with their status.
// Note: Gen1 devices don't have a component enumeration API, so this only works for Gen2+.
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

// gen1LightStatusToLight converts Gen1 light status to model.LightStatus.
func gen1LightStatusToLight(id int, status *gen1comp.LightStatus) *model.LightStatus {
	brightness := status.Brightness
	return &model.LightStatus{
		ID:         id,
		Output:     status.IsOn,
		Brightness: &brightness,
	}
}
