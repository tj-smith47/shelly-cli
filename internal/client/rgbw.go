// Package client provides device communication abstraction over shelly-go SDK.
package client

import (
	"context"

	"github.com/tj-smith47/shelly-go/gen2/components"
	"github.com/tj-smith47/shelly-go/rpc"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// RGBWComponent provides access to RGBW component operations.
type RGBWComponent struct {
	rgbw *components.RGBW
	rpc  *rpc.Client
	id   int
}

// GetStatus returns the current status of the RGBW component.
func (r *RGBWComponent) GetStatus(ctx context.Context) (*model.RGBWStatus, error) {
	status, err := r.rgbw.GetStatus(ctx)
	if err != nil {
		return nil, err
	}

	result := &model.RGBWStatus{
		ID:         status.ID,
		Output:     status.Output,
		Source:     status.Source,
		Brightness: status.Brightness,
		White:      status.White,
		Power:      status.APower,
		Voltage:    status.Voltage,
		Current:    status.Current,
	}

	if len(status.RGB) >= 3 {
		result.RGB = &model.RGBColor{
			Red:   status.RGB[0],
			Green: status.RGB[1],
			Blue:  status.RGB[2],
		}
	}

	return result, nil
}

// GetConfig returns the configuration of the RGBW component.
func (r *RGBWComponent) GetConfig(ctx context.Context) (*model.RGBWConfig, error) {
	config, err := r.rgbw.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	result := &model.RGBWConfig{
		ID:   config.ID,
		Name: config.Name,
	}

	if config.InitialState != nil {
		result.InitialState = *config.InitialState
	}
	if config.AutoOn != nil {
		result.AutoOn = *config.AutoOn
	}
	if config.AutoOnDelay != nil {
		result.AutoOnDelay = *config.AutoOnDelay
	}
	if config.AutoOff != nil {
		result.AutoOff = *config.AutoOff
	}
	if config.AutoOffDelay != nil {
		result.AutoOffDelay = *config.AutoOffDelay
	}
	if config.DefaultBrightness != nil {
		result.DefaultBright = *config.DefaultBrightness
	}
	if config.DefaultWhite != nil {
		result.DefaultWhite = *config.DefaultWhite
	}
	if config.NightMode != nil {
		if config.NightMode.Enable != nil {
			result.NightModeEnable = *config.NightMode.Enable
		}
		if config.NightMode.Brightness != nil {
			result.NightModeBright = *config.NightMode.Brightness
		}
		if config.NightMode.White != nil {
			result.NightModeWhite = *config.NightMode.White
		}
	}

	return result, nil
}

// On turns the RGBW on.
func (r *RGBWComponent) On(ctx context.Context) error {
	on := true
	_, err := r.rgbw.Set(ctx, &components.RGBWSetParams{On: &on})
	return err
}

// Off turns the RGBW off.
func (r *RGBWComponent) Off(ctx context.Context) error {
	on := false
	_, err := r.rgbw.Set(ctx, &components.RGBWSetParams{On: &on})
	return err
}

// Toggle toggles the RGBW state.
func (r *RGBWComponent) Toggle(ctx context.Context) (*model.RGBWStatus, error) {
	result, err := r.rgbw.Toggle(ctx)
	if err != nil {
		return nil, err
	}

	return &model.RGBWStatus{
		ID:     r.id,
		Output: !result.WasOn,
	}, nil
}

// SetBrightness sets the RGBW brightness (0-100).
func (r *RGBWComponent) SetBrightness(ctx context.Context, brightness int) error {
	on := true
	_, err := r.rgbw.Set(ctx, &components.RGBWSetParams{On: &on, Brightness: &brightness})
	return err
}

// SetWhite sets the RGBW white channel (0-100).
func (r *RGBWComponent) SetWhite(ctx context.Context, white int) error {
	on := true
	_, err := r.rgbw.Set(ctx, &components.RGBWSetParams{On: &on, White: &white})
	return err
}

// SetColor sets the RGBW color.
func (r *RGBWComponent) SetColor(ctx context.Context, red, green, blue int) error {
	on := true
	_, err := r.rgbw.Set(ctx, &components.RGBWSetParams{On: &on, RGB: []int{red, green, blue}})
	return err
}

// SetColorAndBrightness sets color, brightness, and optionally white channel.
func (r *RGBWComponent) SetColorAndBrightness(ctx context.Context, red, green, blue, brightness int, white *int) error {
	on := true
	params := &components.RGBWSetParams{
		On:         &on,
		Brightness: &brightness,
		RGB:        []int{red, green, blue},
		White:      white,
	}
	_, err := r.rgbw.Set(ctx, params)
	return err
}

// Set sets RGBW parameters.
func (r *RGBWComponent) Set(ctx context.Context, red, green, blue, brightness, white *int, on *bool) error {
	params := &components.RGBWSetParams{
		On:         on,
		Brightness: brightness,
		White:      white,
	}
	if red != nil && green != nil && blue != nil {
		params.RGB = []int{*red, *green, *blue}
	}
	_, err := r.rgbw.Set(ctx, params)
	return err
}
