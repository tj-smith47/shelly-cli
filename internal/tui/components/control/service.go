package control

import (
	"context"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/component"
)

// ServiceAdapter adapts the shelly service to the Service interface.
type ServiceAdapter struct {
	svc     *shelly.Service
	compSvc *component.Service
}

// NewServiceAdapter creates a new service adapter.
func NewServiceAdapter(svc *shelly.Service) *ServiceAdapter {
	return &ServiceAdapter{
		svc:     svc,
		compSvc: svc.ComponentService(),
	}
}

// SwitchOn turns on a switch component.
func (s *ServiceAdapter) SwitchOn(ctx context.Context, device string, switchID int) error {
	return s.compSvc.SwitchOn(ctx, device, switchID)
}

// SwitchOff turns off a switch component.
func (s *ServiceAdapter) SwitchOff(ctx context.Context, device string, switchID int) error {
	return s.compSvc.SwitchOff(ctx, device, switchID)
}

// SwitchToggle toggles a switch component.
func (s *ServiceAdapter) SwitchToggle(ctx context.Context, device string, switchID int) error {
	_, err := s.compSvc.SwitchToggle(ctx, device, switchID)
	return err
}

// LightOn turns on a light component.
func (s *ServiceAdapter) LightOn(ctx context.Context, device string, lightID int) error {
	return s.compSvc.LightOn(ctx, device, lightID)
}

// LightOff turns off a light component.
func (s *ServiceAdapter) LightOff(ctx context.Context, device string, lightID int) error {
	return s.compSvc.LightOff(ctx, device, lightID)
}

// LightToggle toggles a light component.
func (s *ServiceAdapter) LightToggle(ctx context.Context, device string, lightID int) error {
	_, err := s.compSvc.LightToggle(ctx, device, lightID)
	return err
}

// LightBrightness sets the brightness of a light component.
func (s *ServiceAdapter) LightBrightness(ctx context.Context, device string, lightID, brightness int) error {
	return s.compSvc.LightBrightness(ctx, device, lightID, brightness)
}

// RGBOn turns on an RGB component.
func (s *ServiceAdapter) RGBOn(ctx context.Context, device string, rgbID int) error {
	return s.compSvc.RGBOn(ctx, device, rgbID)
}

// RGBOff turns off an RGB component.
func (s *ServiceAdapter) RGBOff(ctx context.Context, device string, rgbID int) error {
	return s.compSvc.RGBOff(ctx, device, rgbID)
}

// RGBToggle toggles an RGB component.
func (s *ServiceAdapter) RGBToggle(ctx context.Context, device string, rgbID int) error {
	_, err := s.compSvc.RGBToggle(ctx, device, rgbID)
	return err
}

// RGBBrightness sets the brightness of an RGB component.
func (s *ServiceAdapter) RGBBrightness(ctx context.Context, device string, rgbID, brightness int) error {
	return s.compSvc.RGBBrightness(ctx, device, rgbID, brightness)
}

// RGBColor sets the color of an RGB component.
func (s *ServiceAdapter) RGBColor(ctx context.Context, device string, rgbID, r, g, b int) error {
	return s.compSvc.RGBColor(ctx, device, rgbID, r, g, b)
}

// RGBColorAndBrightness sets both color and brightness of an RGB component.
func (s *ServiceAdapter) RGBColorAndBrightness(ctx context.Context, device string, rgbID, r, g, b, brightness int) error {
	return s.compSvc.RGBColorAndBrightness(ctx, device, rgbID, r, g, b, brightness)
}

// CoverOpen opens a cover component.
func (s *ServiceAdapter) CoverOpen(ctx context.Context, device string, coverID int, duration *int) error {
	return s.compSvc.CoverOpen(ctx, device, coverID, duration)
}

// CoverClose closes a cover component.
func (s *ServiceAdapter) CoverClose(ctx context.Context, device string, coverID int, duration *int) error {
	return s.compSvc.CoverClose(ctx, device, coverID, duration)
}

// CoverStop stops a cover component.
func (s *ServiceAdapter) CoverStop(ctx context.Context, device string, coverID int) error {
	return s.compSvc.CoverStop(ctx, device, coverID)
}

// CoverPosition moves a cover to a specific position.
func (s *ServiceAdapter) CoverPosition(ctx context.Context, device string, coverID, position int) error {
	return s.compSvc.CoverPosition(ctx, device, coverID, position)
}

// CoverCalibrate starts cover calibration.
func (s *ServiceAdapter) CoverCalibrate(ctx context.Context, device string, coverID int) error {
	return s.compSvc.CoverCalibrate(ctx, device, coverID)
}

// ThermostatSetTarget sets the target temperature.
func (s *ServiceAdapter) ThermostatSetTarget(ctx context.Context, device string, thermostatID int, targetC float64) error {
	return s.svc.WithConnection(ctx, device, func(conn *client.Client) error {
		return conn.Thermostat(thermostatID).SetTarget(ctx, targetC)
	})
}

// ThermostatSetMode sets the thermostat operating mode.
func (s *ServiceAdapter) ThermostatSetMode(ctx context.Context, device string, thermostatID int, mode string) error {
	return s.svc.WithConnection(ctx, device, func(conn *client.Client) error {
		return conn.Thermostat(thermostatID).SetMode(ctx, mode)
	})
}

// ThermostatBoost activates boost mode.
func (s *ServiceAdapter) ThermostatBoost(ctx context.Context, device string, thermostatID, durationSec int) error {
	return s.svc.WithConnection(ctx, device, func(conn *client.Client) error {
		return conn.Thermostat(thermostatID).Boost(ctx, durationSec)
	})
}

// ThermostatCancelBoost cancels boost mode.
func (s *ServiceAdapter) ThermostatCancelBoost(ctx context.Context, device string, thermostatID int) error {
	return s.svc.WithConnection(ctx, device, func(conn *client.Client) error {
		return conn.Thermostat(thermostatID).CancelBoost(ctx)
	})
}

// Ensure ServiceAdapter implements Service.
var _ Service = (*ServiceAdapter)(nil)
