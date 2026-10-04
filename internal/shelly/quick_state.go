package shelly

import (
	"context"
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly/connection"
)

// OutputStates reports whether each switch, light and RGB output of a device is
// on, in component order. If componentID is set, only that component is read.
// Covers have no on/off state and are left out. A device with no such output,
// or a plugin-managed device, returns an error. Supports Gen1 and Gen2+.
func (s *Service) OutputStates(ctx context.Context, identifier string, componentID *int) ([]bool, error) {
	device, err := s.resolver.Resolve(identifier)
	if err != nil {
		return nil, err
	}
	if device.IsPluginManaged() {
		return nil, fmt.Errorf("reading the on/off state of %s devices is not supported", device.Platform)
	}

	var states []bool
	err = s.WithDevice(ctx, identifier, func(dev *connection.DeviceClient) error {
		var readErr error
		if dev.IsGen1() {
			states, readErr = outputStatesGen1(ctx, dev.Gen1(), componentID)
		} else {
			states, readErr = outputStatesGen2(ctx, dev.Gen2(), componentID)
		}
		return readErr
	})
	if err != nil {
		return nil, err
	}
	if len(states) == 0 {
		if componentID != nil {
			return nil, fmt.Errorf("component ID %d not found on device", *componentID)
		}
		return nil, fmt.Errorf("device %s has no switch or light output", identifier)
	}
	return states, nil
}

func outputStatesGen2(ctx context.Context, conn *client.Client, componentID *int) ([]bool, error) {
	controllable, err := findControllable(ctx, conn)
	if err != nil {
		return nil, err
	}

	var states []bool
	for _, comp := range selectComponents(controllable, componentID) {
		switch comp.Type {
		case model.ComponentSwitch:
			status, statusErr := conn.Switch(comp.ID).GetStatus(ctx)
			if statusErr != nil {
				return nil, statusErr
			}
			states = append(states, status.Output)
		case model.ComponentLight:
			status, statusErr := conn.Light(comp.ID).GetStatus(ctx)
			if statusErr != nil {
				return nil, statusErr
			}
			states = append(states, status.Output)
		case model.ComponentRGB:
			status, statusErr := conn.RGB(comp.ID).GetStatus(ctx)
			if statusErr != nil {
				return nil, statusErr
			}
			states = append(states, status.Output)
		default:
		}
	}
	return states, nil
}

func outputStatesGen1(ctx context.Context, conn *client.Gen1Client, componentID *int) ([]bool, error) {
	controllable, err := findControllableGen1(ctx, conn)
	if err != nil {
		return nil, err
	}

	var states []bool
	for _, comp := range selectComponents(controllable, componentID) {
		switch comp.Type {
		case model.ComponentSwitch:
			relay, relayErr := conn.Relay(comp.ID)
			if relayErr != nil {
				return nil, relayErr
			}
			status, statusErr := relay.GetStatus(ctx)
			if statusErr != nil {
				return nil, statusErr
			}
			states = append(states, status.IsOn)
		case model.ComponentLight:
			light, lightErr := conn.Light(comp.ID)
			if lightErr != nil {
				return nil, lightErr
			}
			status, statusErr := light.GetStatus(ctx)
			if statusErr != nil {
				return nil, statusErr
			}
			states = append(states, status.IsOn)
		case model.ComponentRGB:
			color, colorErr := conn.Color(comp.ID)
			if colorErr != nil {
				return nil, colorErr
			}
			status, statusErr := color.GetStatus(ctx)
			if statusErr != nil {
				return nil, statusErr
			}
			states = append(states, status.IsOn)
		default:
		}
	}
	return states, nil
}
