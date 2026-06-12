// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"fmt"

	gen1comp "github.com/tj-smith47/shelly-go/gen1/components"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// LightInfo holds light information for list operations.
type LightInfo struct {
	ID         int
	Name       string
	Output     bool
	Brightness int
	Power      float64
}

// ListHeaders returns the column headers for the table.
func (l LightInfo) ListHeaders() []string {
	return []string{"ID", headerName, headerState, headerBrightness, headerPower}
}

// ListRow returns the formatted row values for the table.
func (l LightInfo) ListRow() []string {
	name := output.FormatComponentName(l.Name, "light", l.ID)
	state := output.RenderOnOff(l.Output, output.CaseUpper, theme.FalseError)

	brightness := "-"
	if l.Brightness >= 0 {
		brightness = fmt.Sprintf("%d%%", l.Brightness)
	}

	power := output.FormatPowerTableValue(l.Power)
	return []string{fmt.Sprintf("%d", l.ID), name, state, brightness, power}
}

// LightOn turns on a light component.
// For Gen1 devices, this controls the light/dimmer.
func (s *Service) LightOn(ctx context.Context, identifier string, lightID int) error {
	return s.withComponentAction(ctx, identifier,
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
	return s.withComponentAction(ctx, identifier,
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
	var result *model.LightStatus
	err := s.WithDevice(ctx, identifier, func(dev *DeviceClient) error {
		if dev.IsGen1() {
			light, err := dev.Gen1().Light(lightID)
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
		}

		// Gen2+
		var err error
		if _, err = dev.Gen2().Light(lightID).Toggle(ctx); err != nil {
			return err
		}
		// Get current status after toggle.
		result, err = dev.Gen2().Light(lightID).GetStatus(ctx)
		return err
	})
	if err == nil {
		s.invalidateCache(identifier, cache.TypeComponents)
	}
	return result, err
}

// LightBrightness sets the brightness of a light component (0-100).
// For Gen1 devices, this controls the light/dimmer brightness.
func (s *Service) LightBrightness(ctx context.Context, identifier string, lightID, brightness int) error {
	return s.withComponentAction(ctx, identifier,
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
	var result *model.LightStatus
	err := s.WithDevice(ctx, identifier, func(dev *DeviceClient) error {
		if dev.IsGen1() {
			light, err := dev.Gen1().Light(lightID)
			if err != nil {
				return err
			}
			status, err := light.GetStatus(ctx)
			if err != nil {
				return err
			}
			result = gen1LightStatusToLight(lightID, status)
			return nil
		}

		// Gen2+
		status, err := dev.Gen2().Light(lightID).GetStatus(ctx)
		if err != nil {
			return err
		}
		result = status
		return nil
	})
	return result, err
}

// LightSet sets a light's brightness, white color temperature, and on/off state,
// auto-detecting the device generation. Color temperature is applied for Gen1
// white-temp bulbs (e.g. the Duo); Gen2+ tunable white is a separate component,
// so a non-nil temp on a Gen2+ device is reported as unsupported rather than
// silently ignored.
func (s *Service) LightSet(ctx context.Context, identifier string, lightID int, brightness, temp *int, on *bool) error {
	isGen1, _, err := s.IsGen1Device(ctx, identifier)
	if err != nil {
		return err
	}

	var setErr error
	switch {
	case isGen1:
		setErr = s.lightSetGen1(ctx, identifier, lightID, brightness, temp, on)
	case temp != nil:
		return fmt.Errorf("setting color temperature is not supported for Gen2+ lights via this command")
	default:
		setErr = s.WithConnection(ctx, identifier, func(conn *client.Client) error {
			return conn.Light(lightID).Set(ctx, brightness, on)
		})
	}

	// A successful write makes any cached component status stale; drop it so the
	// next `light status` reflects the change instead of waiting out the TTL.
	if setErr == nil {
		s.invalidateCache(identifier, cache.TypeComponents)
	}
	return setErr
}

// lightSetGen1 applies brightness, color temperature, and on/off to a Gen1 light.
// Temperature is applied before brightness so the bulb lands on its final colour
// and level together.
func (s *Service) lightSetGen1(ctx context.Context, identifier string, lightID int, brightness, temp *int, on *bool) error {
	return s.WithGen1Connection(ctx, identifier, func(conn *client.Gen1Client) error {
		light, err := conn.Light(lightID)
		if err != nil {
			return err
		}
		if temp != nil {
			if tErr := light.SetColorTemp(ctx, *temp); tErr != nil {
				return tErr
			}
		}
		if brightness != nil {
			if bErr := light.SetBrightness(ctx, *brightness); bErr != nil {
				return bErr
			}
		}
		if on != nil {
			if *on {
				return light.TurnOn(ctx)
			}
			return light.TurnOff(ctx)
		}
		return nil
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
	light := &model.LightStatus{
		ID:         id,
		Output:     status.IsOn,
		Brightness: &brightness,
	}
	// White-temp bulbs (Duo) report a color temperature; plain dimmers report 0.
	if status.Temp > 0 {
		temp := status.Temp
		light.Temp = &temp
	}
	return light
}
