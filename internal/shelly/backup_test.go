package shelly

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/testutil"
	"github.com/tj-smith47/shelly-go/backup"
)

func TestBackupDeviceInfo_Fields(t *testing.T) {
	t.Parallel()

	info := BackupDeviceInfo{
		ID:         "shellypro1pm-123456",
		Name:       "Living Room",
		Model:      "SNSW-001P16EU",
		Generation: 2,
		FWVersion:  "1.0.0",
		MAC:        "AA:BB:CC:DD:EE:FF",
	}

	testutil.AssertEqual(t, info.ID, "shellypro1pm-123456")
	testutil.AssertEqual(t, info.Name, "Living Room")
	testutil.AssertEqual(t, info.Model, "SNSW-001P16EU")
	testutil.AssertEqual(t, info.Generation, 2)
	testutil.AssertEqual(t, info.FWVersion, "1.0.0")
	testutil.AssertEqual(t, info.MAC, "AA:BB:CC:DD:EE:FF")
}

func TestDeviceBackup_WrapperMethods(t *testing.T) {
	t.Parallel()

	now := time.Now()
	bkup := &backup.Backup{
		Version:   1,
		CreatedAt: now,
		DeviceInfo: &backup.DeviceInfo{
			ID:         "shellyplus1-123456",
			Name:       "Test Device",
			Model:      "SNSW-001P16EU",
			Generation: 2,
			Version:    "1.0.0",
			MAC:        "AA:BB:CC:DD:EE:FF",
		},
		Config:    json.RawMessage(`{"sys":{"device":{"name":"Test"}}}`),
		Scripts:   []*backup.Script{},
		Schedules: json.RawMessage(`[]`),
		Webhooks:  json.RawMessage(`[]`),
		KVS:       map[string]json.RawMessage{},
	}

	wrapped := &DeviceBackup{Backup: bkup}

	// Test Device() method
	device := wrapped.Device()
	testutil.AssertEqual(t, device.ID, "shellyplus1-123456")
	testutil.AssertEqual(t, device.Name, "Test Device")
	testutil.AssertEqual(t, device.Model, "SNSW-001P16EU")
	testutil.AssertEqual(t, device.Generation, 2)
	testutil.AssertEqual(t, device.FWVersion, "1.0.0")
	testutil.AssertEqual(t, device.MAC, "AA:BB:CC:DD:EE:FF")

	// Test Encrypted() method - should return false for non-encrypted backups
	testutil.AssertFalse(t, wrapped.Encrypted(), "expected non-encrypted backup")
}

func TestBackupOptions_ToExportOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts BackupOptions
	}{
		{
			name: "default options",
			opts: BackupOptions{},
		},
		{
			name: "skip scripts",
			opts: BackupOptions{SkipScripts: true},
		},
		{
			name: "skip schedules",
			opts: BackupOptions{SkipSchedules: true},
		},
		{
			name: "skip webhooks",
			opts: BackupOptions{SkipWebhooks: true},
		},
		{
			name: "skip all",
			opts: BackupOptions{
				SkipScripts:   true,
				SkipSchedules: true,
				SkipWebhooks:  true,
			},
		},
		{
			name: "with password",
			opts: BackupOptions{Password: "secret123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.opts.toExportOptions()
			// Verify we got a non-nil result
			if got == nil {
				t.Error("expected non-nil ExportOptions")
			}
		})
	}
}

func TestRestoreOptions_ToRestoreOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts RestoreOptions
	}{
		{
			name: "default options",
			opts: RestoreOptions{},
		},
		{
			name: "dry run",
			opts: RestoreOptions{DryRun: true},
		},
		{
			name: "skip network",
			opts: RestoreOptions{SkipNetwork: true},
		},
		{
			name: "skip scripts",
			opts: RestoreOptions{SkipScripts: true},
		},
		{
			name: "skip all",
			opts: RestoreOptions{
				SkipNetwork:   true,
				SkipScripts:   true,
				SkipSchedules: true,
				SkipWebhooks:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.opts.toRestoreOptions()
			// Verify we got a non-nil result
			if got == nil {
				t.Error("expected non-nil RestoreOptions")
			}
		})
	}
}

func TestValidateBackup_ValidJSON(t *testing.T) {
	t.Parallel()

	now := time.Now()
	bkup := &backup.Backup{
		Version:   1,
		CreatedAt: now,
		DeviceInfo: &backup.DeviceInfo{
			ID:         "shellyplus1-123456",
			Name:       "Test Device",
			Model:      "SNSW-001P16EU",
			Generation: 2,
			Version:    "1.0.0",
			MAC:        "AA:BB:CC:DD:EE:FF",
		},
		Config:    json.RawMessage(`{"sys":{"device":{"name":"Test"}}}`),
		Scripts:   []*backup.Script{},
		Schedules: json.RawMessage(`[]`),
		Webhooks:  json.RawMessage(`[]`),
		KVS:       map[string]json.RawMessage{},
	}

	data, err := json.Marshal(bkup)
	testutil.AssertNil(t, err)

	// Validate the backup
	validated, err := ValidateBackup(data)
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, validated.Device().ID, "shellyplus1-123456")
}

func TestValidateBackup_InvalidJSON(t *testing.T) {
	t.Parallel()

	data := []byte(`{invalid json`)

	_, err := ValidateBackup(data)
	testutil.AssertError(t, err)
}

func TestValidateBackup_EmptyData(t *testing.T) {
	t.Parallel()

	data := []byte(``)

	_, err := ValidateBackup(data)
	testutil.AssertError(t, err)
}

func TestBackupScript_Fields(t *testing.T) {
	t.Parallel()

	script := BackupScript{
		Name:   "test_script",
		Code:   "console.log('test');",
		Enable: true,
	}

	testutil.AssertEqual(t, script.Name, "test_script")
	testutil.AssertEqual(t, script.Code, "console.log('test');")
	testutil.AssertTrue(t, script.Enable, "expected script to be enabled")
}

func TestBackupSchedule_Fields(t *testing.T) {
	t.Parallel()

	schedule := BackupSchedule{
		Enable:   true,
		Timespec: "0 0 * * *",
		Calls: []ScheduleCall{
			{
				Method: "Switch.Set",
				Params: map[string]any{"id": 0, "on": true},
			},
		},
	}

	testutil.AssertTrue(t, schedule.Enable, "expected schedule to be enabled")
	testutil.AssertEqual(t, schedule.Timespec, "0 0 * * *")
	testutil.AssertEqual(t, len(schedule.Calls), 1)
	testutil.AssertEqual(t, schedule.Calls[0].Method, "Switch.Set")
}

func TestWebhookInfo_Fields(t *testing.T) {
	t.Parallel()

	webhook := WebhookInfo{
		ID:     1,
		Cid:    1,
		Enable: true,
		Event:  "switch.toggle",
		Name:   "test_webhook",
		URLs:   []string{"http://example.com/webhook"},
	}

	testutil.AssertEqual(t, webhook.ID, 1)
	testutil.AssertEqual(t, webhook.Cid, 1)
	testutil.AssertTrue(t, webhook.Enable, "expected webhook to be enabled")
	testutil.AssertEqual(t, webhook.Event, "switch.toggle")
	testutil.AssertEqual(t, webhook.Name, "test_webhook")
	testutil.AssertEqual(t, len(webhook.URLs), 1)
	testutil.AssertEqual(t, webhook.URLs[0], "http://example.com/webhook")
}

func TestRestoreResult_Fields(t *testing.T) {
	t.Parallel()

	result := RestoreResult{
		Success:           true,
		ConfigRestored:    true,
		ScriptsRestored:   5,
		SchedulesRestored: 3,
		WebhooksRestored:  2,
		RestartRequired:   true,
		Warnings:          []string{"Warning 1", "Warning 2"},
	}

	testutil.AssertTrue(t, result.Success, "expected restore to succeed")
	testutil.AssertTrue(t, result.ConfigRestored, "expected config to be restored")
	testutil.AssertEqual(t, result.ScriptsRestored, 5)
	testutil.AssertEqual(t, result.SchedulesRestored, 3)
	testutil.AssertEqual(t, result.WebhooksRestored, 2)
	testutil.AssertTrue(t, result.RestartRequired, "expected restart to be required")
	testutil.AssertEqual(t, len(result.Warnings), 2)
}

func TestBackupDiff_Fields(t *testing.T) {
	t.Parallel()

	diff := BackupDiff{
		ConfigDiffs: []ConfigDiff{
			{Key: "sys.device.name", DiffType: "changed", Current: "Old Name", Backup: "New Name"},
		},
		ScriptDiffs: []ScriptDiff{
			{Name: "script1", DiffType: "added", Details: "new script"},
		},
		ScheduleDiffs: []ScheduleDiff{
			{Timespec: "0 0 * * *", DiffType: "removed", Details: "removed schedule"},
		},
		WebhookDiffs: []WebhookDiff{
			{Event: "switch.toggle", Name: "webhook1", DiffType: "changed", Details: "modified"},
		},
	}

	testutil.AssertEqual(t, len(diff.ConfigDiffs), 1)
	testutil.AssertEqual(t, len(diff.ScriptDiffs), 1)
	testutil.AssertEqual(t, len(diff.ScheduleDiffs), 1)
	testutil.AssertEqual(t, len(diff.WebhookDiffs), 1)
}

// TestService_CreateBackup_ContextCancellation tests that CreateBackup respects context cancellation.
func TestService_CreateBackup_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	svc := NewService()
	_, err := svc.CreateBackup(ctx, "test-device", BackupOptions{})

	testutil.AssertError(t, err)
	testutil.AssertErrorContains(t, err, "context canceled")
}

// TestService_RestoreBackup_ContextCancellation tests that RestoreBackup respects context cancellation.
func TestService_RestoreBackup_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	svc := NewService()
	bkup := &DeviceBackup{
		Backup: &backup.Backup{
			Version: 1,
		},
	}
	_, err := svc.RestoreBackup(ctx, "test-device", bkup, RestoreOptions{})

	testutil.AssertError(t, err)
	testutil.AssertErrorContains(t, err, "context canceled")
}
