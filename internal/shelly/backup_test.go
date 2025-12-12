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
		want backup.ExportOptions
	}{
		{
			name: "default options",
			opts: BackupOptions{},
			want: backup.ExportOptions{
				IncludeScripts:   true,
				IncludeSchedules: true,
				IncludeWebhooks:  true,
				IncludeKVS:       true,
			},
		},
		{
			name: "skip scripts",
			opts: BackupOptions{
				SkipScripts: true,
			},
			want: backup.ExportOptions{
				IncludeScripts:   false,
				IncludeSchedules: true,
				IncludeWebhooks:  true,
				IncludeKVS:       true,
			},
		},
		{
			name: "skip schedules",
			opts: BackupOptions{
				SkipSchedules: true,
			},
			want: backup.ExportOptions{
				IncludeScripts:   true,
				IncludeSchedules: false,
				IncludeWebhooks:  true,
				IncludeKVS:       true,
			},
		},
		{
			name: "skip webhooks",
			opts: BackupOptions{
				SkipWebhooks: true,
			},
			want: backup.ExportOptions{
				IncludeScripts:   true,
				IncludeSchedules: true,
				IncludeWebhooks:  false,
				IncludeKVS:       true,
			},
		},
		{
			name: "skip all",
			opts: BackupOptions{
				SkipScripts:   true,
				SkipSchedules: true,
				SkipWebhooks:  true,
			},
			want: backup.ExportOptions{
				IncludeScripts:   false,
				IncludeSchedules: false,
				IncludeWebhooks:  false,
				IncludeKVS:       true,
			},
		},
		{
			name: "with password",
			opts: BackupOptions{
				Password: "secret123",
			},
			want: backup.ExportOptions{
				IncludeScripts:   true,
				IncludeSchedules: true,
				IncludeWebhooks:  true,
				IncludeKVS:       true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.opts.toExportOptions()
			testutil.AssertEqual(t, got.IncludeScripts, tt.want.IncludeScripts)
			testutil.AssertEqual(t, got.IncludeSchedules, tt.want.IncludeSchedules)
			testutil.AssertEqual(t, got.IncludeWebhooks, tt.want.IncludeWebhooks)
			testutil.AssertEqual(t, got.IncludeKVS, tt.want.IncludeKVS)
		})
	}
}

func TestRestoreOptions_ToRestoreOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts RestoreOptions
		want backup.RestoreOptions
	}{
		{
			name: "default options",
			opts: RestoreOptions{},
			want: backup.RestoreOptions{
				IncludeScripts:   true,
				IncludeSchedules: true,
				IncludeWebhooks:  true,
				IncludeKVS:       true,
			},
		},
		{
			name: "dry run",
			opts: RestoreOptions{
				DryRun: true,
			},
			want: backup.RestoreOptions{
				DryRun:           true,
				IncludeScripts:   true,
				IncludeSchedules: true,
				IncludeWebhooks:  true,
				IncludeKVS:       true,
			},
		},
		{
			name: "skip network",
			opts: RestoreOptions{
				SkipNetwork: true,
			},
			want: backup.RestoreOptions{
				SkipNetwork:      true,
				IncludeScripts:   true,
				IncludeSchedules: true,
				IncludeWebhooks:  true,
				IncludeKVS:       true,
			},
		},
		{
			name: "skip scripts",
			opts: RestoreOptions{
				SkipScripts: true,
			},
			want: backup.RestoreOptions{
				IncludeScripts:   false,
				IncludeSchedules: true,
				IncludeWebhooks:  true,
				IncludeKVS:       true,
			},
		},
		{
			name: "skip all",
			opts: RestoreOptions{
				SkipNetwork:   true,
				SkipScripts:   true,
				SkipSchedules: true,
				SkipWebhooks:  true,
			},
			want: backup.RestoreOptions{
				SkipNetwork:      true,
				IncludeScripts:   false,
				IncludeSchedules: false,
				IncludeWebhooks:  false,
				IncludeKVS:       true,
			},
		},
		{
			name: "with password",
			opts: RestoreOptions{
				Password: "secret123",
			},
			want: backup.RestoreOptions{
				IncludeScripts:   true,
				IncludeSchedules: true,
				IncludeWebhooks:  true,
				IncludeKVS:       true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.opts.toRestoreOptions()
			testutil.AssertEqual(t, got.DryRun, tt.want.DryRun)
			testutil.AssertEqual(t, got.SkipNetwork, tt.want.SkipNetwork)
			testutil.AssertEqual(t, got.IncludeScripts, tt.want.IncludeScripts)
			testutil.AssertEqual(t, got.IncludeSchedules, tt.want.IncludeSchedules)
			testutil.AssertEqual(t, got.IncludeWebhooks, tt.want.IncludeWebhooks)
			testutil.AssertEqual(t, got.IncludeKVS, tt.want.IncludeKVS)
		})
	}
}

func TestValidateBackup_ValidJSON(t *testing.T) {
	t.Parallel()

	now := time.Now()
	bkup := &backup.Backup{
		Version:   "1.0",
		CreatedAt: now,
		DeviceInfo: backup.DeviceInfo{
			ID:         "shellyplus1-123456",
			Name:       "Test Device",
			Model:      "SNSW-001P16EU",
			Generation: 2,
			Version:    "1.0.0",
			MAC:        "AA:BB:CC:DD:EE:FF",
		},
		Config:    json.RawMessage(`{"sys":{"device":{"name":"Test"}}}`),
		Scripts:   json.RawMessage(`[]`),
		Schedules: json.RawMessage(`[]`),
		Webhooks:  json.RawMessage(`[]`),
		KVS:       json.RawMessage(`{}`),
	}

	data, err := json.Marshal(bkup)
	testutil.AssertNil(t, err)

	// Validate the backup
	validated, err := ValidateBackup(data)
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, validated.Version, "1.0")
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
		ID:     1,
		Name:   "test_script",
		Code:   "console.log('test');",
		Enable: true,
	}

	testutil.AssertEqual(t, script.ID, 1)
	testutil.AssertEqual(t, script.Name, "test_script")
	testutil.AssertEqual(t, script.Code, "console.log('test');")
	testutil.AssertTrue(t, script.Enable, "expected script to be enabled")
}

func TestBackupSchedule_Fields(t *testing.T) {
	t.Parallel()

	schedule := BackupSchedule{
		ID:       1,
		Enable:   true,
		Timespec: "0 0 * * *",
		Calls: []ScheduleCall{
			{
				Method: "Switch.Set",
				Params: map[string]any{"id": 0, "on": true},
			},
		},
	}

	testutil.AssertEqual(t, schedule.ID, 1)
	testutil.AssertTrue(t, schedule.Enable, "expected schedule to be enabled")
	testutil.AssertEqual(t, schedule.Timespec, "0 0 * * *")
	testutil.AssertEqual(t, len(schedule.Calls), 1)
	testutil.AssertEqual(t, schedule.Calls[0].Method, "Switch.Set")
}

func TestWebhookInfo_Fields(t *testing.T) {
	t.Parallel()

	webhook := WebhookInfo{
		ID:      1,
		CID:     1,
		Enable:  true,
		Event:   "switch.toggle",
		Name:    "test_webhook",
		URLs:    []string{"http://example.com/webhook"},
		Condition: map[string]any{
			"input": 0,
		},
		Repeat: 3,
	}

	testutil.AssertEqual(t, webhook.ID, 1)
	testutil.AssertEqual(t, webhook.CID, 1)
	testutil.AssertTrue(t, webhook.Enable, "expected webhook to be enabled")
	testutil.AssertEqual(t, webhook.Event, "switch.toggle")
	testutil.AssertEqual(t, webhook.Name, "test_webhook")
	testutil.AssertEqual(t, len(webhook.URLs), 1)
	testutil.AssertEqual(t, webhook.URLs[0], "http://example.com/webhook")
	testutil.AssertEqual(t, webhook.Repeat, 3)
}

func TestRestoreResult_Fields(t *testing.T) {
	t.Parallel()

	result := RestoreResult{
		ConfigRestored:    true,
		ScriptsRestored:   5,
		SchedulesRestored: 3,
		WebhooksRestored:  2,
		KVSRestored:       10,
		Warnings:          []string{"Warning 1", "Warning 2"},
	}

	testutil.AssertTrue(t, result.ConfigRestored, "expected config to be restored")
	testutil.AssertEqual(t, result.ScriptsRestored, 5)
	testutil.AssertEqual(t, result.SchedulesRestored, 3)
	testutil.AssertEqual(t, result.WebhooksRestored, 2)
	testutil.AssertEqual(t, result.KVSRestored, 10)
	testutil.AssertEqual(t, len(result.Warnings), 2)
}

func TestConvertBackupScripts(t *testing.T) {
	t.Parallel()

	scripts := []BackupScript{
		{ID: 1, Name: "script1", Enable: true},
		{ID: 2, Name: "script2", Enable: false},
	}

	scriptsJSON, err := json.Marshal(scripts)
	testutil.AssertNil(t, err)

	converted := convertBackupScripts(scriptsJSON)
	testutil.AssertEqual(t, len(converted), 2)
	testutil.AssertEqual(t, converted[0].ID, 1)
	testutil.AssertEqual(t, converted[0].Name, "script1")
	testutil.AssertTrue(t, converted[0].Enable, "expected script1 to be enabled")
	testutil.AssertEqual(t, converted[1].ID, 2)
	testutil.AssertEqual(t, converted[1].Name, "script2")
	testutil.AssertFalse(t, converted[1].Enable, "expected script2 to be disabled")
}

func TestConvertBackupSchedules(t *testing.T) {
	t.Parallel()

	schedules := []BackupSchedule{
		{ID: 1, Enable: true, Timespec: "0 0 * * *"},
		{ID: 2, Enable: false, Timespec: "0 12 * * *"},
	}

	schedulesJSON, err := json.Marshal(schedules)
	testutil.AssertNil(t, err)

	converted := convertBackupSchedules(schedulesJSON)
	testutil.AssertEqual(t, len(converted), 2)
	testutil.AssertEqual(t, converted[0].ID, 1)
	testutil.AssertTrue(t, converted[0].Enable, "expected schedule1 to be enabled")
	testutil.AssertEqual(t, converted[0].Timespec, "0 0 * * *")
	testutil.AssertEqual(t, converted[1].ID, 2)
	testutil.AssertFalse(t, converted[1].Enable, "expected schedule2 to be disabled")
	testutil.AssertEqual(t, converted[1].Timespec, "0 12 * * *")
}

func TestConvertBackupWebhooks(t *testing.T) {
	t.Parallel()

	webhooks := []WebhookInfo{
		{ID: 1, Enable: true, Event: "switch.toggle", Name: "webhook1"},
		{ID: 2, Enable: false, Event: "input.toggle", Name: "webhook2"},
	}

	webhooksJSON, err := json.Marshal(webhooks)
	testutil.AssertNil(t, err)

	converted := convertBackupWebhooks(webhooksJSON)
	testutil.AssertEqual(t, len(converted), 2)
	testutil.AssertEqual(t, converted[0].ID, 1)
	testutil.AssertTrue(t, converted[0].Enable, "expected webhook1 to be enabled")
	testutil.AssertEqual(t, converted[0].Event, "switch.toggle")
	testutil.AssertEqual(t, converted[0].Name, "webhook1")
	testutil.AssertEqual(t, converted[1].ID, 2)
	testutil.AssertFalse(t, converted[1].Enable, "expected webhook2 to be disabled")
	testutil.AssertEqual(t, converted[1].Event, "input.toggle")
	testutil.AssertEqual(t, converted[1].Name, "webhook2")
}

func TestConvertBackupScripts_InvalidJSON(t *testing.T) {
	t.Parallel()

	invalidJSON := json.RawMessage(`{invalid}`)
	converted := convertBackupScripts(invalidJSON)
	testutil.AssertEqual(t, len(converted), 0)
}

func TestConvertBackupSchedules_InvalidJSON(t *testing.T) {
	t.Parallel()

	invalidJSON := json.RawMessage(`{invalid}`)
	converted := convertBackupSchedules(invalidJSON)
	testutil.AssertEqual(t, len(converted), 0)
}

func TestConvertBackupWebhooks_InvalidJSON(t *testing.T) {
	t.Parallel()

	invalidJSON := json.RawMessage(`{invalid}`)
	converted := convertBackupWebhooks(invalidJSON)
	testutil.AssertEqual(t, len(converted), 0)
}

func TestBackupDiff_Fields(t *testing.T) {
	t.Parallel()

	diff := BackupDiff{
		ConfigDiffs: []ConfigDiff{
			{Path: "sys.device.name", Type: "changed", Old: "Old Name", New: "New Name"},
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
			Version: "1.0",
		},
	}
	_, err := svc.RestoreBackup(ctx, "test-device", bkup, RestoreOptions{})

	testutil.AssertError(t, err)
	testutil.AssertErrorContains(t, err, "context canceled")
}
