// Package client provides device communication abstraction over shelly-go SDK.
package client

import (
	"context"

	"github.com/tj-smith47/shelly-go/gen2/components"
	"github.com/tj-smith47/shelly-go/rpc"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// InputComponent provides access to input component operations.
type InputComponent struct {
	input *components.Input
	rpc   *rpc.Client
	id    int
}

// GetStatus returns the current status of the input.
func (i *InputComponent) GetStatus(ctx context.Context) (*model.InputStatus, error) {
	status, err := i.input.GetStatus(ctx)
	if err != nil {
		return nil, err
	}

	result := &model.InputStatus{
		ID: status.ID,
	}

	if status.State != nil {
		result.State = *status.State
	}

	return result, nil
}

// GetConfig returns the configuration of the input.
func (i *InputComponent) GetConfig(ctx context.Context) (*model.InputConfig, error) {
	config, err := i.input.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	result := &model.InputConfig{
		ID:   config.ID,
		Name: config.Name,
		Type: config.Type,
	}

	if config.Invert != nil {
		result.Invert = *config.Invert
	}

	return result, nil
}

// Trigger triggers an input event.
func (i *InputComponent) Trigger(ctx context.Context, eventType string) error {
	return i.input.Trigger(ctx, eventType)
}

// ResetCounters resets the input counters.
func (i *InputComponent) ResetCounters(ctx context.Context, resetTypes []string) error {
	_, err := i.input.ResetCounters(ctx, resetTypes)
	return err
}
