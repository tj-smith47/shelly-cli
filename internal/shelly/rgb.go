// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"

	gen1comp "github.com/tj-smith47/shelly-go/gen1/components"

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

// BuildRGBSetParams creates RGBSetParams from flag values.
// Uses -1 as sentinel for "not set" for color/brightness values.
// Only sets On if explicitly true.
func BuildRGBSetParams(red, green, blue, brightness int, on bool) RGBSetParams {
	params := RGBSetParams{}

	// Color values are valid from 0-255, -1 means not set
	if red >= 0 && red <= 255 {
		params.Red = &red
	}
	if green >= 0 && green <= 255 {
		params.Green = &green
	}
	if blue >= 0 && blue <= 255 {
		params.Blue = &blue
	}

	// Brightness is valid from 0-100
	if brightness >= 0 && brightness <= 100 {
		params.Brightness = &brightness
	}

	// Only set on if explicitly requested
	if on {
		params.On = &on
	}

	return params
}

// RGBOn turns on an RGB component.
// For Gen1 devices, this controls the color light.
func (s *Service) RGBOn(ctx context.Context, identifier string, rgbID int) error {
	return s.withGenAwareAction(ctx, identifier,
		func(conn *client.Gen1Client) error {
			color, err := conn.Color(rgbID)
			if err != nil {
				return err
			}
			return color.TurnOn(ctx)
		},
		func(conn *client.Client) error {
			return conn.RGB(rgbID).On(ctx)
		},
	)
}

// RGBOff turns off an RGB component.
// For Gen1 devices, this controls the color light.
func (s *Service) RGBOff(ctx context.Context, identifier string, rgbID int) error {
	return s.withGenAwareAction(ctx, identifier,
		func(conn *client.Gen1Client) error {
			color, err := conn.Color(rgbID)
			if err != nil {
				return err
			}
			return color.TurnOff(ctx)
		},
		func(conn *client.Client) error {
			return conn.RGB(rgbID).Off(ctx)
		},
	)
}

// RGBToggle toggles an RGB component and returns the new status.
// For Gen1 devices, this controls the color light.
func (s *Service) RGBToggle(ctx context.Context, identifier string, rgbID int) (*model.RGBStatus, error) {
	var result *model.RGBStatus
	err := s.WithDevice(ctx, identifier, func(dev *DeviceClient) error {
		if dev.IsGen1() {
			color, err := dev.Gen1().Color(rgbID)
			if err != nil {
				return err
			}
			if err := color.Toggle(ctx); err != nil {
				return err
			}
			status, err := color.GetStatus(ctx)
			if err != nil {
				return err
			}
			result = gen1ColorStatusToRGB(rgbID, status)
			return nil
		}

		// Gen2+
		var err error
		if _, err = dev.Gen2().RGB(rgbID).Toggle(ctx); err != nil {
			return err
		}
		// Get current status after toggle.
		result, err = dev.Gen2().RGB(rgbID).GetStatus(ctx)
		return err
	})
	return result, err
}

// RGBBrightness sets the brightness of an RGB component (0-100).
// For Gen1 devices, this sets the gain (brightness) of the color light.
func (s *Service) RGBBrightness(ctx context.Context, identifier string, rgbID, brightness int) error {
	return s.withGenAwareAction(ctx, identifier,
		func(conn *client.Gen1Client) error {
			color, err := conn.Color(rgbID)
			if err != nil {
				return err
			}
			return color.SetGain(ctx, brightness)
		},
		func(conn *client.Client) error {
			return conn.RGB(rgbID).SetBrightness(ctx, brightness)
		},
	)
}

// RGBColor sets the color of an RGB component.
// For Gen1 devices, this sets the RGB color of the color light.
func (s *Service) RGBColor(ctx context.Context, identifier string, rgbID, r, g, b int) error {
	return s.withGenAwareAction(ctx, identifier,
		func(conn *client.Gen1Client) error {
			color, err := conn.Color(rgbID)
			if err != nil {
				return err
			}
			return color.SetRGB(ctx, r, g, b)
		},
		func(conn *client.Client) error {
			return conn.RGB(rgbID).SetColor(ctx, r, g, b)
		},
	)
}

// RGBColorAndBrightness sets both color and brightness of an RGB component.
// For Gen1 devices, this sets RGB and gain on the color light.
func (s *Service) RGBColorAndBrightness(ctx context.Context, identifier string, rgbID, r, g, b, brightness int) error {
	return s.withGenAwareAction(ctx, identifier,
		func(conn *client.Gen1Client) error {
			color, err := conn.Color(rgbID)
			if err != nil {
				return err
			}
			// Gen1 requires setting them together via TurnOnWithRGB
			return color.TurnOnWithRGB(ctx, r, g, b, brightness)
		},
		func(conn *client.Client) error {
			return conn.RGB(rgbID).SetColorAndBrightness(ctx, r, g, b, brightness)
		},
	)
}

// RGBStatus gets the status of an RGB component.
// For Gen1 devices, this returns color light status.
func (s *Service) RGBStatus(ctx context.Context, identifier string, rgbID int) (*model.RGBStatus, error) {
	var result *model.RGBStatus
	err := s.WithDevice(ctx, identifier, func(dev *DeviceClient) error {
		if dev.IsGen1() {
			color, err := dev.Gen1().Color(rgbID)
			if err != nil {
				return err
			}
			status, err := color.GetStatus(ctx)
			if err != nil {
				return err
			}
			result = gen1ColorStatusToRGB(rgbID, status)
			return nil
		}

		// Gen2+
		status, err := dev.Gen2().RGB(rgbID).GetStatus(ctx)
		if err != nil {
			return err
		}
		result = status
		return nil
	})
	return result, err
}

// RGBSet sets parameters of an RGB component.
// For Gen1 devices, use RGBOn/RGBOff/RGBColor/RGBBrightness instead (Gen1 doesn't support combined set).
func (s *Service) RGBSet(ctx context.Context, identifier string, rgbID int, params RGBSetParams) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.RGB(rgbID).Set(ctx, params.Red, params.Green, params.Blue, params.Brightness, params.On)
	})
}

// RGBList lists all RGB components on a device with their status.
// Note: Gen1 devices don't have a component enumeration API, so this only works for Gen2+.
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

// gen1ColorStatusToRGB converts Gen1 color status to model.RGBStatus.
func gen1ColorStatusToRGB(id int, status *gen1comp.ColorStatus) *model.RGBStatus {
	gain := status.Gain
	return &model.RGBStatus{
		ID:         id,
		Output:     status.IsOn,
		Brightness: &gain,
		RGB: &model.RGBColor{
			Red:   status.Red,
			Green: status.Green,
			Blue:  status.Blue,
		},
	}
}
