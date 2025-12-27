package shelly

import (
	"context"
	"testing"

	"github.com/tj-smith47/shelly-go/backup"

	backuppkg "github.com/tj-smith47/shelly-cli/internal/shelly/backup"
	"github.com/tj-smith47/shelly-cli/internal/testutil"
)

func TestMigrationOptions_ToMigrationOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts MigrationOptions
		want backup.MigrationOptions
	}{
		{
			name: "default options",
			opts: MigrationOptions{},
			want: backup.MigrationOptions{
				IncludeWiFi:      false,
				IncludeCloud:     false,
				IncludeMQTT:      false,
				IncludeBLE:       false,
				IncludeSchedules: false,
				IncludeWebhooks:  false,
				IncludeScripts:   false,
				IncludeKVS:       false,
				RebootAfter:      false,
				DryRun:           false,
			},
		},
		{
			name: "include all",
			opts: MigrationOptions{
				IncludeWiFi:      true,
				IncludeCloud:     true,
				IncludeMQTT:      true,
				IncludeBLE:       true,
				IncludeSchedules: true,
				IncludeWebhooks:  true,
				IncludeScripts:   true,
				IncludeKVS:       true,
				RebootAfter:      true,
			},
			want: backup.MigrationOptions{
				IncludeWiFi:      true,
				IncludeCloud:     true,
				IncludeMQTT:      true,
				IncludeBLE:       true,
				IncludeSchedules: true,
				IncludeWebhooks:  true,
				IncludeScripts:   true,
				IncludeKVS:       true,
				RebootAfter:      true,
				DryRun:           false,
			},
		},
		{
			name: "dry run",
			opts: MigrationOptions{
				DryRun: true,
			},
			want: backup.MigrationOptions{
				IncludeWiFi:      false,
				IncludeCloud:     false,
				IncludeMQTT:      false,
				IncludeBLE:       false,
				IncludeSchedules: false,
				IncludeWebhooks:  false,
				IncludeScripts:   false,
				IncludeKVS:       false,
				RebootAfter:      false,
				DryRun:           true,
			},
		},
		{
			name: "scripts and schedules only",
			opts: MigrationOptions{
				IncludeScripts:   true,
				IncludeSchedules: true,
			},
			want: backup.MigrationOptions{
				IncludeWiFi:      false,
				IncludeCloud:     false,
				IncludeMQTT:      false,
				IncludeBLE:       false,
				IncludeSchedules: true,
				IncludeWebhooks:  false,
				IncludeScripts:   true,
				IncludeKVS:       false,
				RebootAfter:      false,
				DryRun:           false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.opts.toMigrationOptions()
			testutil.AssertEqual(t, got.IncludeWiFi, tt.want.IncludeWiFi)
			testutil.AssertEqual(t, got.IncludeCloud, tt.want.IncludeCloud)
			testutil.AssertEqual(t, got.IncludeMQTT, tt.want.IncludeMQTT)
			testutil.AssertEqual(t, got.IncludeBLE, tt.want.IncludeBLE)
			testutil.AssertEqual(t, got.IncludeSchedules, tt.want.IncludeSchedules)
			testutil.AssertEqual(t, got.IncludeWebhooks, tt.want.IncludeWebhooks)
			testutil.AssertEqual(t, got.IncludeScripts, tt.want.IncludeScripts)
			testutil.AssertEqual(t, got.IncludeKVS, tt.want.IncludeKVS)
			testutil.AssertEqual(t, got.RebootAfter, tt.want.RebootAfter)
			testutil.AssertEqual(t, got.DryRun, tt.want.DryRun)
		})
	}
}

func TestDefaultMigrationOptions(t *testing.T) {
	t.Parallel()

	opts := DefaultMigrationOptions()

	// Security: WiFi should default to false
	testutil.AssertFalse(t, opts.IncludeWiFi, "WiFi should be excluded by default for security")

	// Default includes
	testutil.AssertTrue(t, opts.IncludeCloud, "Cloud should be included by default")
	testutil.AssertTrue(t, opts.IncludeMQTT, "MQTT should be included by default")
	testutil.AssertTrue(t, opts.IncludeBLE, "BLE should be included by default")
	testutil.AssertTrue(t, opts.IncludeSchedules, "Schedules should be included by default")
	testutil.AssertTrue(t, opts.IncludeWebhooks, "Webhooks should be included by default")
	testutil.AssertTrue(t, opts.IncludeScripts, "Scripts should be included by default")
	testutil.AssertTrue(t, opts.IncludeKVS, "KVS should be included by default")
	testutil.AssertTrue(t, opts.RebootAfter, "RebootAfter should be true by default")

	// Default excludes
	testutil.AssertFalse(t, opts.DryRun, "DryRun should be false by default")
	testutil.AssertFalse(t, opts.AllowDifferentModels, "AllowDifferentModels should be false by default")
	testutil.AssertFalse(t, opts.AllowDifferentGenerations, "AllowDifferentGenerations should be false by default")
}

func TestMigrationResult_Fields(t *testing.T) {
	t.Parallel()

	result := MigrationResult{
		SourceDevice: &backuppkg.DeviceInfo{
			ID:         "source-123",
			Name:       "Source Device",
			Model:      "SNSW-001P16EU",
			Generation: 2,
			FWVersion:  "1.0.0",
			MAC:        "AA:BB:CC:DD:EE:01",
		},
		TargetDevice: &backuppkg.DeviceInfo{
			ID:         "target-456",
			Name:       "Target Device",
			Model:      "SNSW-001P16EU",
			Generation: 2,
			FWVersion:  "1.0.1",
			MAC:        "AA:BB:CC:DD:EE:02",
		},
		ComponentsMigrated: []string{"scripts", "schedules", "webhooks"},
		Warnings:           []string{"Warning 1"},
		Errors:             []error{},
		Success:            true,
		RestartRequired:    true,
		DurationSeconds:    45.5,
	}

	testutil.AssertEqual(t, result.SourceDevice.ID, "source-123")
	testutil.AssertEqual(t, result.TargetDevice.ID, "target-456")
	testutil.AssertEqual(t, len(result.ComponentsMigrated), 3)
	testutil.AssertEqual(t, result.ComponentsMigrated[0], "scripts")
	testutil.AssertEqual(t, len(result.Warnings), 1)
	testutil.AssertEqual(t, len(result.Errors), 0)
	testutil.AssertTrue(t, result.Success, "expected migration to succeed")
	testutil.AssertTrue(t, result.RestartRequired, "expected restart to be required")
	testutil.AssertEqual(t, result.DurationSeconds, 45.5)
}

func TestMigrationValidation_Fields(t *testing.T) {
	t.Parallel()

	validation := MigrationValidation{
		SourceDevice: &backuppkg.DeviceInfo{
			ID:         "source-123",
			Model:      "SNSW-001P16EU",
			Generation: 2,
		},
		TargetDevice: &backuppkg.DeviceInfo{
			ID:         "target-456",
			Model:      "SNSW-001P16EU",
			Generation: 2,
		},
		Warnings: []string{"Minor version mismatch"},
		Errors:   []string{},
		Valid:    true,
	}

	testutil.AssertEqual(t, validation.SourceDevice.ID, "source-123")
	testutil.AssertEqual(t, validation.TargetDevice.ID, "target-456")
	testutil.AssertEqual(t, len(validation.Warnings), 1)
	testutil.AssertEqual(t, len(validation.Errors), 0)
	testutil.AssertTrue(t, validation.Valid, "expected validation to pass")
}

func TestMigrationValidation_InvalidCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		validation MigrationValidation
		wantValid  bool
	}{
		{
			name: "different models without allow flag",
			validation: MigrationValidation{
				SourceDevice: &backuppkg.DeviceInfo{Model: "SNSW-001P16EU"},
				TargetDevice: &backuppkg.DeviceInfo{Model: "SNSW-002P16EU"},
				Errors:       []string{"Model mismatch"},
				Valid:        false,
			},
			wantValid: false,
		},
		{
			name: "different generations without allow flag",
			validation: MigrationValidation{
				SourceDevice: &backuppkg.DeviceInfo{Generation: 1},
				TargetDevice: &backuppkg.DeviceInfo{Generation: 2},
				Errors:       []string{"Generation mismatch"},
				Valid:        false,
			},
			wantValid: false,
		},
		{
			name: "valid same model and generation",
			validation: MigrationValidation{
				SourceDevice: &backuppkg.DeviceInfo{Model: "SNSW-001P16EU", Generation: 2},
				TargetDevice: &backuppkg.DeviceInfo{Model: "SNSW-001P16EU", Generation: 2},
				Errors:       []string{},
				Valid:        true,
			},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.wantValid {
				testutil.AssertTrue(t, tt.validation.Valid, "expected validation to pass")
			} else {
				testutil.AssertFalse(t, tt.validation.Valid, "expected validation to fail")
			}
		})
	}
}

// TestService_Migrate_ContextCancellation tests that Migrate respects context cancellation.
func TestService_Migrate_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	svc := NewService()
	_, err := svc.Migrate(ctx, "source", "target", nil, nil)

	testutil.AssertError(t, err)
	testutil.AssertErrorContains(t, err, "context canceled")
}

// TestService_ValidateMigration_ContextCancellation tests that ValidateMigration respects context cancellation.
func TestService_ValidateMigration_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	svc := NewService()
	_, err := svc.ValidateMigration(ctx, "source", "target", nil)

	testutil.AssertError(t, err)
	testutil.AssertErrorContains(t, err, "context canceled")
}

func TestCompareScripts2(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		scripts1 []backuppkg.Script
		scripts2 []backuppkg.Script
		wantLen  int
	}{
		{
			name:     "identical scripts",
			scripts1: []backuppkg.Script{{Name: "script1", Enable: true}},
			scripts2: []backuppkg.Script{{Name: "script1", Enable: true}},
			wantLen:  1, // Still shows as "changed"
		},
		{
			name:     "added script",
			scripts1: []backuppkg.Script{},
			scripts2: []backuppkg.Script{{Name: "script1", Enable: true}},
			wantLen:  1,
		},
		{
			name:     "removed script",
			scripts1: []backuppkg.Script{{Name: "script1", Enable: true}},
			scripts2: []backuppkg.Script{},
			wantLen:  1,
		},
		{
			name: "multiple changes",
			scripts1: []backuppkg.Script{
				{Name: "script1", Enable: true},
				{Name: "script2", Enable: false},
			},
			scripts2: []backuppkg.Script{
				{Name: "script1", Enable: false},
				{Name: "script3", Enable: true},
			},
			wantLen: 3, // script1 changed, script2 removed, script3 added
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			diffs := compareScripts2(tt.scripts1, tt.scripts2)
			testutil.AssertEqual(t, len(diffs), tt.wantLen)
		})
	}
}

func TestCompareSchedules2(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		schedules1 []backuppkg.Schedule
		schedules2 []backuppkg.Schedule
		wantLen    int
	}{
		{
			name:       "identical schedules",
			schedules1: []backuppkg.Schedule{{Timespec: "0 0 * * *", Enable: true}},
			schedules2: []backuppkg.Schedule{{Timespec: "0 0 * * *", Enable: true}},
			wantLen:    0,
		},
		{
			name:       "added schedule",
			schedules1: []backuppkg.Schedule{},
			schedules2: []backuppkg.Schedule{{Timespec: "0 0 * * *", Enable: true}},
			wantLen:    1,
		},
		{
			name:       "removed schedule",
			schedules1: []backuppkg.Schedule{{Timespec: "0 0 * * *", Enable: true}},
			schedules2: []backuppkg.Schedule{},
			wantLen:    1,
		},
		{
			name: "multiple changes",
			schedules1: []backuppkg.Schedule{
				{Timespec: "0 0 * * *", Enable: true},
				{Timespec: "0 12 * * *", Enable: false},
			},
			schedules2: []backuppkg.Schedule{
				{Timespec: "0 0 * * *", Enable: false},
				{Timespec: "0 6 * * *", Enable: true},
			},
			wantLen: 2, // 0 12 removed, 0 6 added
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			diffs := compareSchedules2(tt.schedules1, tt.schedules2)
			testutil.AssertEqual(t, len(diffs), tt.wantLen)
		})
	}
}

func TestCompareWebhooks2(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		webhooks1 []WebhookInfo
		webhooks2 []WebhookInfo
		wantLen   int
	}{
		{
			name:      "identical webhooks",
			webhooks1: []WebhookInfo{{Event: "switch.toggle", Name: "webhook1", Enable: true}},
			webhooks2: []WebhookInfo{{Event: "switch.toggle", Name: "webhook1", Enable: true}},
			wantLen:   0,
		},
		{
			name:      "added webhook",
			webhooks1: []WebhookInfo{},
			webhooks2: []WebhookInfo{{Event: "switch.toggle", Name: "webhook1", Enable: true}},
			wantLen:   1,
		},
		{
			name:      "removed webhook",
			webhooks1: []WebhookInfo{{Event: "switch.toggle", Name: "webhook1", Enable: true}},
			webhooks2: []WebhookInfo{},
			wantLen:   1,
		},
		{
			name: "multiple changes",
			webhooks1: []WebhookInfo{
				{Event: "switch.toggle", Name: "webhook1", Enable: true},
				{Event: "input.toggle", Name: "webhook2", Enable: false},
			},
			webhooks2: []WebhookInfo{
				{Event: "switch.toggle", Name: "webhook1", Enable: false},
				{Event: "cover.open", Name: "webhook3", Enable: true},
			},
			wantLen: 2, // webhook2 removed, webhook3 added
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			diffs := compareWebhooks2(tt.webhooks1, tt.webhooks2)
			testutil.AssertEqual(t, len(diffs), tt.wantLen)
		})
	}
}

func TestCompareConfigs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config1 map[string]any
		config2 map[string]any
		wantLen int
	}{
		{
			name:    "identical configs",
			config1: map[string]any{"key": "value"},
			config2: map[string]any{"key": "value"},
			wantLen: 0,
		},
		{
			name:    "added key",
			config1: map[string]any{},
			config2: map[string]any{"key": "value"},
			wantLen: 1,
		},
		{
			name:    "removed key",
			config1: map[string]any{"key": "value"},
			config2: map[string]any{},
			wantLen: 1,
		},
		{
			name:    "changed value",
			config1: map[string]any{"key": "value1"},
			config2: map[string]any{"key": "value2"},
			wantLen: 1,
		},
		{
			name: "nested changes",
			config1: map[string]any{
				"sys": map[string]any{
					"device": map[string]any{
						"name": "Old Name",
					},
				},
			},
			config2: map[string]any{
				"sys": map[string]any{
					"device": map[string]any{
						"name": "New Name",
					},
				},
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			diffs := compareConfigs(tt.config1, tt.config2)
			testutil.AssertEqual(t, len(diffs), tt.wantLen)
		})
	}
}
