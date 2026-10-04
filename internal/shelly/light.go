// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"errors"
	"fmt"
	"time"

	gen1comp "github.com/tj-smith47/shelly-go/gen1/components"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// LightInfo holds light information for list operations.
type LightInfo struct {
	ID         int     `json:"id" yaml:"id"`
	Name       string  `json:"name" yaml:"name"`
	Output     bool    `json:"output" yaml:"output"`
	Brightness int     `json:"brightness" yaml:"brightness"`
	Power      float64 `json:"power" yaml:"power"`
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

// LightIDs returns the IDs of the dimmable light components on a device: the
// channels of a Gen1 light in white mode, or the Light components of a Gen2+
// device. A device with only relays, switches, covers, or a Gen1 light in
// color mode returns an empty list.
func (s *Service) LightIDs(ctx context.Context, identifier string) ([]int, error) {
	var ids []int
	err := s.WithDevice(ctx, identifier, func(dev *DeviceClient) error {
		if dev.IsGen1() {
			settings, err := dev.Gen1().GetSettings(ctx)
			if err != nil {
				return err
			}
			// A Gen1 RGBW in color mode lists its channel under "lights" too,
			// but it has no /light endpoint to set brightness on.
			if settings.Mode == "color" {
				return nil
			}
			for i := range settings.Lights {
				ids = append(ids, i)
			}
			return nil
		}
		comps, err := dev.Gen2().FilterComponents(ctx, model.ComponentLight)
		if err != nil {
			return err
		}
		for _, c := range comps {
			ids = append(ids, c.ID)
		}
		return nil
	})
	return ids, err
}

// LightRamp turns the given lights on at brightness from and raises them to
// brightness to in even steps spread over duration. Steps are at least minStep
// apart, so a short ramp takes fewer, larger steps. It returns the brightness
// every light last reached. Cancelling ctx stops the ramp at that level and
// returns ctx.Err().
func (s *Service) LightRamp(ctx context.Context, identifier string, lightIDs []int, from, to int, duration, minStep time.Duration) (int, error) {
	if duration <= 0 {
		return 0, errors.New("ramp duration must be greater than zero")
	}
	on := true
	set := func(level int) error {
		for _, id := range lightIDs {
			if err := s.LightSet(ctx, identifier, id, &level, nil, &on); err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				return fmt.Errorf("set light %d to %d%%: %w", id, level, err)
			}
		}
		return nil
	}

	if err := set(from); err != nil {
		return 0, err
	}
	level := from

	steps := to - from
	if minStep > 0 {
		steps = min(steps, int(duration/minStep))
	}
	steps = max(steps, 1)
	ticker := time.NewTicker(duration / time.Duration(steps))
	defer ticker.Stop()

	for i := 1; i <= steps; i++ {
		select {
		case <-ctx.Done():
			return level, ctx.Err()
		case <-ticker.C:
		}
		next := from + (to-from)*i/steps
		if err := set(next); err != nil {
			return level, err
		}
		level = next
	}
	return level, nil
}

// LightList lists all light components on a device with their status.
// Gen2+ devices enumerate their components; a Gen1 device lists its white
// light channels from /settings.
func (s *Service) LightList(ctx context.Context, identifier string) ([]LightInfo, error) {
	var result []LightInfo
	err := s.withGenAwareAction(ctx, identifier, func(conn *client.Gen1Client) error {
		names, err := gen1ComponentNames(ctx, conn, model.ComponentLight)
		if err != nil {
			return err
		}
		result = make([]LightInfo, 0, len(names))
		for id, name := range names {
			light, err := conn.Light(id)
			if err != nil {
				continue
			}
			status, err := light.GetStatus(ctx)
			if err != nil {
				continue
			}
			result = append(result, lightInfo(id, name, gen1LightStatusToLight(id, status)))
		}
		return nil
	}, func(conn *client.Client) error {
		components, err := conn.FilterComponents(ctx, model.ComponentLight)
		if err != nil {
			return err
		}

		result = make([]LightInfo, 0, len(components))
		for _, comp := range components {
			status, err := conn.Light(comp.ID).GetStatus(ctx)
			if err != nil {
				continue
			}
			name := ""
			config, err := conn.Light(comp.ID).GetConfig(ctx)
			if err == nil && config.Name != nil {
				name = *config.Name
			}
			result = append(result, lightInfo(comp.ID, name, status))
		}

		return nil
	})
	return result, err
}

// lightInfo builds a list row from a light status. Brightness is -1 when the
// device does not report one.
func lightInfo(id int, name string, status *model.LightStatus) LightInfo {
	info := LightInfo{ID: id, Name: name, Output: status.Output, Brightness: -1}
	if status.Brightness != nil {
		info.Brightness = *status.Brightness
	}
	if status.Power != nil {
		info.Power = *status.Power
	}
	return info
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
