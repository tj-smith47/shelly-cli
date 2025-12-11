// Package client provides device communication abstraction over shelly-go SDK.
package client

import (
	"context"

	"github.com/tj-smith47/shelly-go/gen2/components"
	"github.com/tj-smith47/shelly-go/rpc"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// CoverComponent provides access to cover component operations.
type CoverComponent struct {
	cv  *components.Cover
	rpc *rpc.Client
	id  int
}

// GetStatus returns the current status of the cover.
func (c *CoverComponent) GetStatus(ctx context.Context) (*model.CoverStatus, error) {
	status, err := c.cv.GetStatus(ctx)
	if err != nil {
		return nil, err
	}

	result := &model.CoverStatus{
		ID:              status.ID,
		State:           status.State,
		Source:          status.Source,
		CurrentPosition: status.CurrentPos,
		TargetPosition:  status.TargetPos,
		Power:           status.APower,
		Voltage:         status.Voltage,
		Current:         status.Current,
	}

	if status.MoveTimeout != nil {
		result.MoveTimeout = *status.MoveTimeout
	}

	return result, nil
}

// GetConfig returns the configuration of the cover.
func (c *CoverComponent) GetConfig(ctx context.Context) (*model.CoverConfig, error) {
	config, err := c.cv.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	result := &model.CoverConfig{
		ID:   config.ID,
		Name: config.Name,
	}

	if config.InitialState != nil {
		result.InitialState = *config.InitialState
	}
	if config.InvertDirections != nil {
		result.InvertDirections = *config.InvertDirections
	}
	if config.SwapInputs != nil {
		result.SwapInputs = *config.SwapInputs
	}

	return result, nil
}

// Open opens the cover with optional duration in seconds.
func (c *CoverComponent) Open(ctx context.Context, duration *int) error {
	var dur *float64
	if duration != nil {
		d := float64(*duration)
		dur = &d
	}
	return c.cv.Open(ctx, dur)
}

// Close closes the cover with optional duration in seconds.
func (c *CoverComponent) Close(ctx context.Context, duration *int) error {
	var dur *float64
	if duration != nil {
		d := float64(*duration)
		dur = &d
	}
	return c.cv.Close(ctx, dur)
}

// Stop stops the cover movement.
func (c *CoverComponent) Stop(ctx context.Context) error {
	return c.cv.Stop(ctx)
}

// GoToPosition moves the cover to a specific position (0-100).
func (c *CoverComponent) GoToPosition(ctx context.Context, pos int) error {
	return c.cv.GoToPosition(ctx, pos)
}

// Calibrate starts cover calibration.
func (c *CoverComponent) Calibrate(ctx context.Context) error {
	return c.cv.Calibrate(ctx)
}
