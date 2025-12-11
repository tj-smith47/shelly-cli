// Package client provides device communication abstraction over shelly-go SDK.
package client

import (
	"context"

	"github.com/tj-smith47/shelly-go/gen2/components"
	"github.com/tj-smith47/shelly-go/rpc"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// RGBComponent provides access to RGB component operations.
type RGBComponent struct {
	rgb *components.RGB
	rpc *rpc.Client
	id  int
}

// GetStatus returns the current status of the RGB component.
func (r *RGBComponent) GetStatus(ctx context.Context) (*model.RGBStatus, error) {
	status, err := r.rgb.GetStatus(ctx)
	if err != nil {
		return nil, err
	}

	result := &model.RGBStatus{
		ID:         status.ID,
		Output:     status.Output,
		Source:     status.Source,
		Brightness: status.Brightness,
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

// GetConfig returns the configuration of the RGB component.
func (r *RGBComponent) GetConfig(ctx context.Context) (*model.RGBConfig, error) {
	config, err := r.rgb.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	result := &model.RGBConfig{
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

	return result, nil
}

// On turns the RGB on.
func (r *RGBComponent) On(ctx context.Context) error {
	on := true
	_, err := r.rgb.Set(ctx, &components.RGBSetParams{On: &on})
	return err
}

// Off turns the RGB off.
func (r *RGBComponent) Off(ctx context.Context) error {
	on := false
	_, err := r.rgb.Set(ctx, &components.RGBSetParams{On: &on})
	return err
}

// Toggle toggles the RGB state.
func (r *RGBComponent) Toggle(ctx context.Context) (*model.RGBStatus, error) {
	result, err := r.rgb.Toggle(ctx)
	if err != nil {
		return nil, err
	}

	return &model.RGBStatus{
		ID:     r.id,
		Output: !result.WasOn,
	}, nil
}

// SetBrightness sets the RGB brightness (0-100).
func (r *RGBComponent) SetBrightness(ctx context.Context, brightness int) error {
	on := true
	_, err := r.rgb.Set(ctx, &components.RGBSetParams{On: &on, Brightness: &brightness})
	return err
}

// SetColor sets the RGB color.
func (r *RGBComponent) SetColor(ctx context.Context, red, green, blue int) error {
	on := true
	_, err := r.rgb.Set(ctx, &components.RGBSetParams{On: &on, RGB: []int{red, green, blue}})
	return err
}

// SetColorAndBrightness sets both color and brightness.
func (r *RGBComponent) SetColorAndBrightness(ctx context.Context, red, green, blue, brightness int) error {
	on := true
	_, err := r.rgb.Set(ctx, &components.RGBSetParams{On: &on, Brightness: &brightness, RGB: []int{red, green, blue}})
	return err
}

// Set sets RGB parameters.
func (r *RGBComponent) Set(ctx context.Context, red, green, blue, brightness *int, on *bool) error {
	params := &components.RGBSetParams{
		On:         on,
		Brightness: brightness,
	}
	if red != nil && green != nil && blue != nil {
		params.RGB = []int{*red, *green, *blue}
	}
	_, err := r.rgb.Set(ctx, params)
	return err
}
