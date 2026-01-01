package shelly

import (
	"context"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly/monitoring"
)

// monitoringAdapter adapts shelly.Service to implement monitoring.ShellyConnector.
type monitoringAdapter struct {
	*Service
}

// Ensure monitoringAdapter implements monitoring.ShellyConnector.
var _ monitoring.ShellyConnector = (*monitoringAdapter)(nil)

// WithConnection implements monitoring.ShellyConnector.
func (a *monitoringAdapter) WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error {
	return a.Service.WithConnection(ctx, identifier, fn)
}

// WithGen1Connection implements monitoring.ShellyConnector.
func (a *monitoringAdapter) WithGen1Connection(ctx context.Context, identifier string, fn func(*client.Gen1Client) error) error {
	return a.Service.WithGen1Connection(ctx, identifier, fn)
}

// Resolve implements monitoring.ShellyConnector.
func (a *monitoringAdapter) Resolve(identifier string) (model.Device, error) {
	return a.Service.resolver.Resolve(identifier)
}

// ResolveWithGeneration implements monitoring.ShellyConnector.
func (a *monitoringAdapter) ResolveWithGeneration(ctx context.Context, identifier string) (*monitoring.ResolvedDevice, error) {
	dev, err := a.Service.ResolveWithGeneration(ctx, identifier)
	if err != nil {
		return nil, err
	}
	return &monitoring.ResolvedDevice{
		Device:     dev,
		Generation: dev.Generation,
	}, nil
}

// DeviceInfo implements monitoring.ShellyConnector.
func (a *monitoringAdapter) DeviceInfo(ctx context.Context, identifier string) (*monitoring.DeviceInfo, error) {
	info, err := a.Service.DeviceInfo(ctx, identifier)
	if err != nil {
		return nil, err
	}
	return &monitoring.DeviceInfo{
		ID:         info.ID,
		MAC:        info.MAC,
		Model:      info.Model,
		Generation: info.Generation,
		Firmware:   info.Firmware,
		App:        info.App,
		AuthEn:     info.AuthEn,
	}, nil
}

// DeviceStatus implements monitoring.ShellyConnector.
func (a *monitoringAdapter) DeviceStatus(ctx context.Context, identifier string) (*monitoring.DeviceStatusResult, error) {
	status, err := a.Service.DeviceStatus(ctx, identifier)
	if err != nil {
		return nil, err
	}
	return &monitoring.DeviceStatusResult{
		Status: status.Status,
	}, nil
}

// Monitoring returns the monitoring service.
// The service is lazily initialized on first access.
func (s *Service) Monitoring() *monitoring.Service {
	if s.monitoringService == nil {
		s.monitoringService = monitoring.NewService(&monitoringAdapter{s})
	}
	return s.monitoringService
}
