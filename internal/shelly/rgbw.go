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

// RGBWInfo holds RGBW information for list operations.
type RGBWInfo struct {
	ID         int
	Name       string
	Output     bool
	Brightness int
	Red        int
	Green      int
	Blue       int
	White      int
	Power      float64
}

// ListHeaders returns the column headers for the table.
func (r RGBWInfo) ListHeaders() []string {
	return []string{"ID", "Name", "State", "Color", "White", "Brightness", "Power"}
}

// ListRow returns the formatted row values for the table.
func (r RGBWInfo) ListRow() []string {
	name := output.FormatComponentName(r.Name, "rgbw", r.ID)
	state := output.RenderOnOff(r.Output, output.CaseUpper, theme.FalseError)
	color := fmt.Sprintf("R:%d G:%d B:%d", r.Red, r.Green, r.Blue)

	white := "-"
	if r.White >= 0 {
		white = fmt.Sprintf("%d", r.White)
	}

	brightness := "-"
	if r.Brightness >= 0 {
		brightness = fmt.Sprintf("%d%%", r.Brightness)
	}

	power := output.FormatPowerTableValue(r.Power)
	return []string{fmt.Sprintf("%d", r.ID), name, state, color, white, brightness, power}
}

// RGBWSetParams holds parameters for RGBWSet operation.
type RGBWSetParams struct {
	Red        *int
	Green      *int
	Blue       *int
	White      *int
	Brightness *int
	On         *bool
}

// BuildRGBWSetParams creates RGBWSetParams from flag values.
// Uses -1 as sentinel for "not set" for color/brightness/white values.
// Only sets On if explicitly true.
func BuildRGBWSetParams(red, green, blue, white, brightness int, on bool) RGBWSetParams {
	params := RGBWSetParams{}

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

	// White is valid from 0-100 (or 0-255 depending on device)
	if white >= 0 && white <= 255 {
		params.White = &white
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

// RGBWOn turns on an RGBW component.
// For Gen1 devices, this controls the color light.
func (s *Service) RGBWOn(ctx context.Context, identifier string, rgbwID int) error {
	return s.withGenAwareAction(ctx, identifier,
		func(conn *client.Gen1Client) error {
			color, err := conn.Color(rgbwID)
			if err != nil {
				return err
			}
			return color.TurnOn(ctx)
		},
		func(conn *client.Client) error {
			return conn.RGBW(rgbwID).On(ctx)
		},
	)
}

// RGBWOff turns off an RGBW component.
// For Gen1 devices, this controls the color light.
func (s *Service) RGBWOff(ctx context.Context, identifier string, rgbwID int) error {
	return s.withGenAwareAction(ctx, identifier,
		func(conn *client.Gen1Client) error {
			color, err := conn.Color(rgbwID)
			if err != nil {
				return err
			}
			return color.TurnOff(ctx)
		},
		func(conn *client.Client) error {
			return conn.RGBW(rgbwID).Off(ctx)
		},
	)
}

// RGBWToggle toggles an RGBW component and returns the new status.
// For Gen1 devices, this controls the color light.
func (s *Service) RGBWToggle(ctx context.Context, identifier string, rgbwID int) (*model.RGBWStatus, error) {
	var result *model.RGBWStatus
	err := s.WithDevice(ctx, identifier, func(dev *DeviceClient) error {
		if dev.IsGen1() {
			color, err := dev.Gen1().Color(rgbwID)
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
			result = gen1ColorStatusToRGBW(rgbwID, status)
			return nil
		}

		// Gen2+
		var err error
		result, err = dev.Gen2().RGBW(rgbwID).Toggle(ctx)
		return err
	})
	return result, err
}

// RGBWBrightness sets the brightness of an RGBW component (0-100).
// For Gen1 devices, this sets the gain (brightness) of the color light.
func (s *Service) RGBWBrightness(ctx context.Context, identifier string, rgbwID, brightness int) error {
	return s.withGenAwareAction(ctx, identifier,
		func(conn *client.Gen1Client) error {
			color, err := conn.Color(rgbwID)
			if err != nil {
				return err
			}
			return color.SetGain(ctx, brightness)
		},
		func(conn *client.Client) error {
			return conn.RGBW(rgbwID).SetBrightness(ctx, brightness)
		},
	)
}

// RGBWWhite sets the white channel of an RGBW component (0-100).
// For Gen1 devices, this sets the white channel of the color light.
func (s *Service) RGBWWhite(ctx context.Context, identifier string, rgbwID, white int) error {
	return s.withGenAwareAction(ctx, identifier,
		func(conn *client.Gen1Client) error {
			color, err := conn.Color(rgbwID)
			if err != nil {
				return err
			}
			// Gen1 doesn't have a dedicated SetWhite, use SetRGBW with current RGB
			status, err := color.GetStatus(ctx)
			if err != nil {
				return err
			}
			return color.SetRGBW(ctx, status.Red, status.Green, status.Blue, white)
		},
		func(conn *client.Client) error {
			return conn.RGBW(rgbwID).SetWhite(ctx, white)
		},
	)
}

// RGBWColor sets the color of an RGBW component.
// For Gen1 devices, this sets the RGB color of the color light.
func (s *Service) RGBWColor(ctx context.Context, identifier string, rgbwID, r, g, b int) error {
	return s.withGenAwareAction(ctx, identifier,
		func(conn *client.Gen1Client) error {
			color, err := conn.Color(rgbwID)
			if err != nil {
				return err
			}
			return color.SetRGB(ctx, r, g, b)
		},
		func(conn *client.Client) error {
			return conn.RGBW(rgbwID).SetColor(ctx, r, g, b)
		},
	)
}

// RGBWColorAndWhite sets both color and white channel of an RGBW component.
// For Gen1 devices, this sets RGBW on the color light.
func (s *Service) RGBWColorAndWhite(ctx context.Context, identifier string, rgbwID, r, g, b, white int) error {
	return s.withGenAwareAction(ctx, identifier,
		func(conn *client.Gen1Client) error {
			color, err := conn.Color(rgbwID)
			if err != nil {
				return err
			}
			return color.SetRGBW(ctx, r, g, b, white)
		},
		func(conn *client.Client) error {
			w := white
			return conn.RGBW(rgbwID).SetColorAndBrightness(ctx, r, g, b, 100, &w)
		},
	)
}

// RGBWStatus gets the status of an RGBW component.
// For Gen1 devices, this returns color light status.
func (s *Service) RGBWStatus(ctx context.Context, identifier string, rgbwID int) (*model.RGBWStatus, error) {
	var result *model.RGBWStatus
	err := s.WithDevice(ctx, identifier, func(dev *DeviceClient) error {
		if dev.IsGen1() {
			color, err := dev.Gen1().Color(rgbwID)
			if err != nil {
				return err
			}
			status, err := color.GetStatus(ctx)
			if err != nil {
				return err
			}
			result = gen1ColorStatusToRGBW(rgbwID, status)
			return nil
		}

		// Gen2+
		status, err := dev.Gen2().RGBW(rgbwID).GetStatus(ctx)
		if err != nil {
			return err
		}
		result = status
		return nil
	})
	return result, err
}

// RGBWSet sets parameters of an RGBW component.
// For Gen1 devices, use RGBWOn/RGBWOff/RGBWColor/RGBWBrightness instead (Gen1 doesn't support combined set).
func (s *Service) RGBWSet(ctx context.Context, identifier string, rgbwID int, params RGBWSetParams) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.RGBW(rgbwID).Set(ctx, params.Red, params.Green, params.Blue, params.Brightness, params.White, params.On)
	})
}

// RGBWList lists all RGBW components on a device with their status.
// Note: Gen1 devices don't have a component enumeration API, so this only works for Gen2+.
func (s *Service) RGBWList(ctx context.Context, identifier string) ([]RGBWInfo, error) {
	var result []RGBWInfo
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		components, err := conn.FilterComponents(ctx, model.ComponentRGBW)
		if err != nil {
			return err
		}

		result = make([]RGBWInfo, 0, len(components))
		for _, comp := range components {
			info := RGBWInfo{ID: comp.ID, Brightness: -1, White: -1}

			status, err := conn.RGBW(comp.ID).GetStatus(ctx)
			if err != nil {
				continue
			}
			info.Output = status.Output
			if status.Brightness != nil {
				info.Brightness = *status.Brightness
			}
			if status.White != nil {
				info.White = *status.White
			}
			if status.RGB != nil {
				info.Red = status.RGB.Red
				info.Green = status.RGB.Green
				info.Blue = status.RGB.Blue
			}
			if status.Power != nil {
				info.Power = *status.Power
			}

			config, err := conn.RGBW(comp.ID).GetConfig(ctx)
			if err == nil && config.Name != nil {
				info.Name = *config.Name
			}

			result = append(result, info)
		}

		return nil
	})
	return result, err
}

// gen1ColorStatusToRGBW converts Gen1 color status to model.RGBWStatus.
func gen1ColorStatusToRGBW(id int, status *gen1comp.ColorStatus) *model.RGBWStatus {
	gain := status.Gain
	white := status.White
	return &model.RGBWStatus{
		ID:         id,
		Output:     status.IsOn,
		Brightness: &gain,
		White:      &white,
		RGB: &model.RGBColor{
			Red:   status.Red,
			Green: status.Green,
			Blue:  status.Blue,
		},
	}
}
