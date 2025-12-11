// Package client provides device communication abstraction over shelly-go SDK.
package client

import (
	"context"

	"github.com/tj-smith47/shelly-go/gen2/components"
	"github.com/tj-smith47/shelly-go/rpc"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// LightComponent provides access to light component operations.
type LightComponent struct {
	lt  *components.Light
	rpc *rpc.Client
	id  int
}

// GetStatus returns the current status of the light.
func (l *LightComponent) GetStatus(ctx context.Context) (*model.LightStatus, error) {
	status, err := l.lt.GetStatus(ctx)
	if err != nil {
		return nil, err
	}

	result := &model.LightStatus{
		ID:         status.ID,
		Output:     status.Output,
		Source:     status.Source,
		Brightness: status.Brightness,
		Power:      status.APower,
		Voltage:    status.Voltage,
		Current:    status.Current,
	}

	return result, nil
}

// GetConfig returns the configuration of the light.
func (l *LightComponent) GetConfig(ctx context.Context) (*model.LightConfig, error) {
	config, err := l.lt.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	result := &model.LightConfig{
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

// On turns the light on.
func (l *LightComponent) On(ctx context.Context) error {
	on := true
	_, err := l.lt.Set(ctx, &components.LightSetParams{On: &on})
	return err
}

// Off turns the light off.
func (l *LightComponent) Off(ctx context.Context) error {
	on := false
	_, err := l.lt.Set(ctx, &components.LightSetParams{On: &on})
	return err
}

// Toggle toggles the light state.
func (l *LightComponent) Toggle(ctx context.Context) (*model.LightStatus, error) {
	result, err := l.lt.Toggle(ctx)
	if err != nil {
		return nil, err
	}

	output := true
	if result.WasOn != nil {
		output = !*result.WasOn // Toggle returns was_on, so current is inverse
	}

	return &model.LightStatus{
		ID:     l.id,
		Output: output,
	}, nil
}

// SetBrightness sets the light brightness (0-100).
func (l *LightComponent) SetBrightness(ctx context.Context, brightness int) error {
	on := true
	_, err := l.lt.Set(ctx, &components.LightSetParams{On: &on, Brightness: &brightness})
	return err
}

// Set sets light parameters.
func (l *LightComponent) Set(ctx context.Context, brightness *int, on *bool) error {
	params := &components.LightSetParams{
		On:         on,
		Brightness: brightness,
	}
	_, err := l.lt.Set(ctx, params)
	return err
}
