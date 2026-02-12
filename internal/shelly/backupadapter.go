// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
)

// BackupConnector wraps a Service to implement backup.ShellyConnector.
type BackupConnector struct {
	svc     *Service
	autoSvc *automation.Service
}

// NewBackupConnector creates a BackupConnector from a Service.
// Note: cache and IOStreams are nil since backup only uses read operations.
func NewBackupConnector(svc *Service) *BackupConnector {
	return &BackupConnector{
		svc:     svc,
		autoSvc: automation.New(svc, nil, nil),
	}
}

// WithConnection implements backup.ShellyConnector.
func (c *BackupConnector) WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error {
	return c.svc.WithConnection(ctx, identifier, fn)
}

// WithGen1Connection implements backup.ShellyConnector.
func (c *BackupConnector) WithGen1Connection(ctx context.Context, identifier string, fn func(*client.Gen1Client) error) error {
	return c.svc.WithGen1Connection(ctx, identifier, fn)
}

// IsGen1Device implements backup.ShellyConnector.
func (c *BackupConnector) IsGen1Device(ctx context.Context, identifier string) (bool, error) {
	isGen1, _, err := c.svc.IsGen1Device(ctx, identifier)
	return isGen1, err
}

// DeviceInfo implements backup.ShellyConnector.
func (c *BackupConnector) DeviceInfo(ctx context.Context, identifier string) (*backup.DeviceInfoResult, error) {
	info, err := c.svc.DeviceInfo(ctx, identifier)
	if err != nil {
		return nil, err
	}
	return &backup.DeviceInfoResult{
		ID:         info.ID,
		MAC:        info.MAC,
		Model:      info.Type, // Raw SKU for backup compatibility checks
		Generation: info.Generation,
		Firmware:   info.Firmware,
		App:        info.App,
		AuthEn:     info.AuthEn,
	}, nil
}

// GetConfig implements backup.ShellyConnector.
func (c *BackupConnector) GetConfig(ctx context.Context, identifier string) (map[string]any, error) {
	return c.svc.GetConfig(ctx, identifier)
}

// ListScripts implements backup.ShellyConnector.
func (c *BackupConnector) ListScripts(ctx context.Context, identifier string) ([]backup.ScriptInfoResult, error) {
	scripts, err := c.autoSvc.ListScripts(ctx, identifier)
	if err != nil {
		return nil, err
	}
	result := make([]backup.ScriptInfoResult, len(scripts))
	for i, s := range scripts {
		result[i] = backup.ScriptInfoResult{
			ID:      s.ID,
			Name:    s.Name,
			Enable:  s.Enable,
			Running: s.Running,
		}
	}
	return result, nil
}

// ListSchedules implements backup.ShellyConnector.
func (c *BackupConnector) ListSchedules(ctx context.Context, identifier string) ([]backup.ScheduleJobResult, error) {
	schedules, err := c.autoSvc.ListSchedules(ctx, identifier)
	if err != nil {
		return nil, err
	}
	result := make([]backup.ScheduleJobResult, len(schedules))
	for i, s := range schedules {
		calls := make([]backup.ScheduleCallResult, len(s.Calls))
		for j, call := range s.Calls {
			calls[j] = backup.ScheduleCallResult{
				Method: call.Method,
				Params: call.Params,
			}
		}
		result[i] = backup.ScheduleJobResult{
			ID:       s.ID,
			Enable:   s.Enable,
			Timespec: s.Timespec,
			Calls:    calls,
		}
	}
	return result, nil
}

// ListWebhooks implements backup.ShellyConnector.
func (c *BackupConnector) ListWebhooks(ctx context.Context, identifier string) ([]backup.WebhookInfoResult, error) {
	webhooks, err := c.svc.ListWebhooks(ctx, identifier)
	if err != nil {
		return nil, err
	}
	result := make([]backup.WebhookInfoResult, len(webhooks))
	for i, w := range webhooks {
		result[i] = backup.WebhookInfoResult{
			ID:     w.ID,
			Cid:    w.Cid,
			Enable: w.Enable,
			Event:  w.Event,
			Name:   w.Name,
			URLs:   w.URLs,
		}
	}
	return result, nil
}

// BackupService returns a backup.Service that can perform backup operations.
// This bridges the shelly.Service with the backup package.
func (s *Service) BackupService() *backup.Service {
	return backup.NewService(NewBackupConnector(s))
}

// CreateBackup creates a complete backup of a device.
// This is a convenience method that delegates to BackupService.
func (s *Service) CreateBackup(ctx context.Context, identifier string, opts backup.Options) (*backup.DeviceBackup, error) {
	return s.BackupService().CreateBackup(ctx, identifier, opts)
}

// RestoreBackup restores a backup to a device.
// This is a convenience method that delegates to BackupService.
func (s *Service) RestoreBackup(ctx context.Context, identifier string, deviceBackup *backup.DeviceBackup, opts backup.RestoreOptions) (*backup.RestoreResult, error) {
	return s.BackupService().RestoreBackup(ctx, identifier, deviceBackup, opts)
}

// CompareBackup compares a backup with a device's current state.
// This is a convenience method that delegates to BackupService.
func (s *Service) CompareBackup(ctx context.Context, identifier string, deviceBackup *backup.DeviceBackup) (*model.BackupDiff, error) {
	return s.BackupService().CompareBackup(ctx, identifier, deviceBackup)
}

// LoadMigrationSource loads a backup from either a file or device.
// This is a convenience method that delegates to BackupService.
func (s *Service) LoadMigrationSource(ctx context.Context, source string) (*backup.DeviceBackup, backup.MigrationSource, error) {
	return s.BackupService().LoadMigrationSource(ctx, source)
}

// CheckMigrationCompatibility checks if the backup is compatible with the target device.
// This is a convenience method that delegates to BackupService.
func (s *Service) CheckMigrationCompatibility(ctx context.Context, bkp *backup.DeviceBackup, target string, force bool) error {
	return s.BackupService().CheckMigrationCompatibility(ctx, bkp, target, force)
}
