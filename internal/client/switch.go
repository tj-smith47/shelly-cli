// Package client provides device communication abstraction over shelly-go SDK.
package client

import (
	"context"

	"github.com/tj-smith47/shelly-go/gen2/components"
	"github.com/tj-smith47/shelly-go/rpc"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// SwitchComponent provides access to switch component operations.
type SwitchComponent struct {
	sw  *components.Switch
	rpc *rpc.Client
	id  int
}

// GetStatus returns the current status of the switch.
func (s *SwitchComponent) GetStatus(ctx context.Context) (*model.SwitchStatus, error) {
	status, err := s.sw.GetStatus(ctx)
	if err != nil {
		return nil, err
	}

	result := &model.SwitchStatus{
		ID:     status.ID,
		Output: status.Output,
		Source: status.Source,
	}

	if status.APower != nil {
		result.Power = status.APower
	}
	if status.Voltage != nil {
		result.Voltage = status.Voltage
	}
	if status.Current != nil {
		result.Current = status.Current
	}
	if status.AEnergy != nil {
		energy := &model.EnergyCounter{
			Total:    status.AEnergy.Total,
			ByMinute: status.AEnergy.ByMinute,
		}
		if status.AEnergy.MinuteTs != nil {
			energy.MinuteTs = *status.AEnergy.MinuteTs
		}
		result.Energy = energy
	}

	return result, nil
}

// GetConfig returns the configuration of the switch.
func (s *SwitchComponent) GetConfig(ctx context.Context) (*model.SwitchConfig, error) {
	config, err := s.sw.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	result := &model.SwitchConfig{
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

	return result, nil
}

// On turns the switch on.
func (s *SwitchComponent) On(ctx context.Context) error {
	on := true
	_, err := s.sw.Set(ctx, &components.SwitchSetParams{On: &on})
	return err
}

// Off turns the switch off.
func (s *SwitchComponent) Off(ctx context.Context) error {
	off := false
	_, err := s.sw.Set(ctx, &components.SwitchSetParams{On: &off})
	return err
}

// Toggle toggles the switch state.
func (s *SwitchComponent) Toggle(ctx context.Context) (*model.SwitchStatus, error) {
	result, err := s.sw.Toggle(ctx)
	if err != nil {
		return nil, err
	}

	return &model.SwitchStatus{
		ID:     s.id,
		Output: !result.WasOn, // Toggle returns was_on, so current is inverse
	}, nil
}

// Set sets the switch to the specified state.
func (s *SwitchComponent) Set(ctx context.Context, on bool) error {
	_, err := s.sw.Set(ctx, &components.SwitchSetParams{On: &on})
	return err
}
