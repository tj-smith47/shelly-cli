// Package backup provides backup and restore operations for Shelly devices.
package backup

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/spf13/afero"
	shellybackup "github.com/tj-smith47/shelly-go/backup"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

const testBackupFilePath = "/test/backup.json"

func TestDeviceBackup_Device(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		deviceInfo *shellybackup.DeviceInfo
		want       DeviceInfo
	}{
		{
			name: "with device info",
			deviceInfo: &shellybackup.DeviceInfo{
				ID:         "shellyplus2pm-123456",
				Name:       "Living Room",
				Model:      "SNSW-002P16EU",
				Generation: 2,
				Version:    "1.0.0",
				MAC:        "AA:BB:CC:DD:EE:FF",
			},
			want: DeviceInfo{
				ID:         "shellyplus2pm-123456",
				Name:       "Living Room",
				Model:      "SNSW-002P16EU",
				Generation: 2,
				FWVersion:  "1.0.0",
				MAC:        "AA:BB:CC:DD:EE:FF",
			},
		},
		{
			name:       "nil device info",
			deviceInfo: nil,
			want:       DeviceInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			bkp := &DeviceBackup{
				Backup: &shellybackup.Backup{
					DeviceInfo: tt.deviceInfo,
				},
			}

			got := bkp.Device()

			if got.ID != tt.want.ID {
				t.Errorf("got ID=%q, want %q", got.ID, tt.want.ID)
			}
			if got.Name != tt.want.Name {
				t.Errorf("got Name=%q, want %q", got.Name, tt.want.Name)
			}
			if got.Model != tt.want.Model {
				t.Errorf("got Model=%q, want %q", got.Model, tt.want.Model)
			}
			if got.Generation != tt.want.Generation {
				t.Errorf("got Generation=%d, want %d", got.Generation, tt.want.Generation)
			}
			if got.FWVersion != tt.want.FWVersion {
				t.Errorf("got FWVersion=%q, want %q", got.FWVersion, tt.want.FWVersion)
			}
			if got.MAC != tt.want.MAC {
				t.Errorf("got MAC=%q, want %q", got.MAC, tt.want.MAC)
			}
		})
	}
}

func TestDeviceBackup_Encrypted(t *testing.T) {
	t.Parallel()

	bkp := &DeviceBackup{
		Backup: &shellybackup.Backup{},
	}

	if bkp.Encrypted() {
		t.Error("expected Encrypted() to return false")
	}
}

func TestOptions_ToExportOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts Options
		want *shellybackup.ExportOptions
	}{
		{
			name: "default options",
			opts: Options{},
			want: &shellybackup.ExportOptions{
				IncludeWiFi:       true,
				IncludeCloud:      true,
				IncludeAuth:       true,
				IncludeBLE:        true,
				IncludeMQTT:       true,
				IncludeWebhooks:   true,
				IncludeSchedules:  true,
				IncludeScripts:    true,
				IncludeKVS:        true,
				IncludeComponents: true,
			},
		},
		{
			name: "skip all",
			opts: Options{
				SkipScripts:   true,
				SkipSchedules: true,
				SkipWebhooks:  true,
				SkipKVS:       true,
				SkipWiFi:      true,
			},
			want: &shellybackup.ExportOptions{
				IncludeWiFi:       false,
				IncludeCloud:      true,
				IncludeAuth:       true,
				IncludeBLE:        true,
				IncludeMQTT:       true,
				IncludeWebhooks:   false,
				IncludeSchedules:  false,
				IncludeScripts:    false,
				IncludeKVS:        false,
				IncludeComponents: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.opts.ToExportOptions()

			if got.IncludeWiFi != tt.want.IncludeWiFi {
				t.Errorf("IncludeWiFi: got %v, want %v", got.IncludeWiFi, tt.want.IncludeWiFi)
			}
			if got.IncludeWebhooks != tt.want.IncludeWebhooks {
				t.Errorf("IncludeWebhooks: got %v, want %v", got.IncludeWebhooks, tt.want.IncludeWebhooks)
			}
			if got.IncludeSchedules != tt.want.IncludeSchedules {
				t.Errorf("IncludeSchedules: got %v, want %v", got.IncludeSchedules, tt.want.IncludeSchedules)
			}
			if got.IncludeScripts != tt.want.IncludeScripts {
				t.Errorf("IncludeScripts: got %v, want %v", got.IncludeScripts, tt.want.IncludeScripts)
			}
			if got.IncludeKVS != tt.want.IncludeKVS {
				t.Errorf("IncludeKVS: got %v, want %v", got.IncludeKVS, tt.want.IncludeKVS)
			}
		})
	}
}

func TestRestoreOptions_ToRestoreOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts RestoreOptions
		want *shellybackup.RestoreOptions
	}{
		{
			name: "default options",
			opts: RestoreOptions{},
			want: &shellybackup.RestoreOptions{
				RestoreWiFi:       true,
				RestoreCloud:      true,
				RestoreAuth:       true,
				RestoreBLE:        true,
				RestoreMQTT:       true,
				RestoreWebhooks:   true,
				RestoreSchedules:  true,
				RestoreScripts:    true,
				RestoreKVS:        true,
				RestoreComponents: true,
				DryRun:            false,
				StopScripts:       true,
			},
		},
		{
			name: "skip all with dry run",
			opts: RestoreOptions{
				DryRun:        true,
				SkipAuth:      true,
				SkipNetwork:   true,
				SkipScripts:   true,
				SkipSchedules: true,
				SkipWebhooks:  true,
				SkipKVS:       true,
			},
			want: &shellybackup.RestoreOptions{
				RestoreWiFi:       false,
				RestoreCloud:      true,
				RestoreAuth:       false,
				RestoreBLE:        true,
				RestoreMQTT:       true,
				RestoreWebhooks:   false,
				RestoreSchedules:  false,
				RestoreScripts:    false,
				RestoreKVS:        false,
				RestoreComponents: true,
				DryRun:            true,
				StopScripts:       true,
			},
		},
		{
			name: "skip auth only",
			opts: RestoreOptions{
				SkipAuth: true,
			},
			want: &shellybackup.RestoreOptions{
				RestoreWiFi:       true,
				RestoreCloud:      true,
				RestoreAuth:       false,
				RestoreBLE:        true,
				RestoreMQTT:       true,
				RestoreWebhooks:   true,
				RestoreSchedules:  true,
				RestoreScripts:    true,
				RestoreKVS:        true,
				RestoreComponents: true,
				DryRun:            false,
				StopScripts:       true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.opts.ToRestoreOptions()

			if got.RestoreWiFi != tt.want.RestoreWiFi {
				t.Errorf("RestoreWiFi: got %v, want %v", got.RestoreWiFi, tt.want.RestoreWiFi)
			}
			if got.RestoreAuth != tt.want.RestoreAuth {
				t.Errorf("RestoreAuth: got %v, want %v", got.RestoreAuth, tt.want.RestoreAuth)
			}
			if got.DryRun != tt.want.DryRun {
				t.Errorf("DryRun: got %v, want %v", got.DryRun, tt.want.DryRun)
			}
		})
	}
}

func TestCompatibilityError_Error(t *testing.T) {
	t.Parallel()

	err := &CompatibilityError{
		SourceModel: "SHSW-1",
		TargetModel: "SNSW-002P16EU",
	}

	if err.Error() != "device type mismatch" {
		t.Errorf("got %q, want %q", err.Error(), "device type mismatch")
	}
}

func TestGenerateFilename(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		deviceName string
		deviceID   string
		encrypted  bool
		wantPrefix string
		wantSuffix string
	}{
		{
			name:       "regular backup",
			deviceName: "Living Room",
			deviceID:   "shellyplus2pm-123",
			encrypted:  false,
			wantPrefix: "backup-Living-Room-",
			wantSuffix: config.ExtJSON,
		},
		{
			name:       "encrypted backup",
			deviceName: "Kitchen",
			deviceID:   "shellyplus1-456",
			encrypted:  true,
			wantPrefix: "backup-Kitchen-",
			wantSuffix: ".enc.json",
		},
		{
			name:       "empty device name",
			deviceName: "",
			deviceID:   "shellyplus2pm-123",
			encrypted:  false,
			wantPrefix: "backup-shellyplus2pm-123-",
			wantSuffix: config.ExtJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := GenerateFilename(tt.deviceName, tt.deviceID, tt.encrypted)

			if len(got) < len(tt.wantPrefix)+len(tt.wantSuffix)+15 {
				t.Errorf("filename too short: %q", got)
			}
			if got[:len(tt.wantPrefix)] != tt.wantPrefix {
				t.Errorf("got prefix %q, want %q", got[:len(tt.wantPrefix)], tt.wantPrefix)
			}
			if got[len(got)-len(tt.wantSuffix):] != tt.wantSuffix {
				t.Errorf("got suffix %q, want %q", got[len(got)-len(tt.wantSuffix):], tt.wantSuffix)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name: "valid backup",
			data: []byte(`{
				"version": 1,
				"device_info": {"id": "test", "model": "test"},
				"config": {}
			}`),
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			data:    []byte(`not valid json`),
			wantErr: true,
		},
		{
			name:    "missing version",
			data:    []byte(`{"device_info": {}, "config": {}}`),
			wantErr: true,
		},
		{
			name:    "missing device_info",
			data:    []byte(`{"version": 1, "config": {}}`),
			wantErr: true,
		},
		{
			name:    "missing config",
			data:    []byte(`{"version": 1, "device_info": {}}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := Validate(tt.data)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestSaveToFile(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	filePath := "/test/subdir/backup.json"
	data := []byte(`{"test": "data"}`)

	err := SaveToFile(data, filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created
	content, err := afero.ReadFile(config.Fs(), filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if !bytes.Equal(content, data) {
		t.Errorf("got %q, want %q", string(content), string(data))
	}

	// Verify permissions
	info, err := config.Fs().Stat(filePath)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("got permissions %o, want %o", info.Mode().Perm(), 0o600)
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestLoadFromFile(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	filePath := testBackupFilePath
	expectedData := []byte(`{"test": "data"}`)

	if err := afero.WriteFile(config.Fs(), filePath, expectedData, 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	data, err := LoadFromFile(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(data, expectedData) {
		t.Errorf("got %q, want %q", string(data), string(expectedData))
	}
}

func TestLoadFromFile_NotExists(t *testing.T) {
	t.Parallel()

	_, err := LoadFromFile("/nonexistent/path/backup.json")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestIsFile(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	filePath := "/test/test.json"
	dirPath := "/test"

	// Before file exists
	if IsFile(filePath) {
		t.Error("expected false for non-existent file")
	}

	// Create file
	if err := afero.WriteFile(config.Fs(), filePath, []byte("test"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	if !IsFile(filePath) {
		t.Error("expected true for existing file")
	}

	// Check directory returns false
	if IsFile(dirPath) {
		t.Error("expected false for directory")
	}
}

func TestCompareConfigs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		current map[string]any
		backup  map[string]any
		want    []string
	}{
		{
			name:    "identical configs",
			current: map[string]any{"key": "value"},
			backup:  map[string]any{"key": "value"},
			want:    nil,
		},
		{
			name:    "added key",
			current: map[string]any{},
			backup:  map[string]any{"new_key": "value"},
			want:    []string{model.DiffAdded},
		},
		{
			name:    "removed key",
			current: map[string]any{"old_key": "value"},
			backup:  map[string]any{},
			want:    []string{model.DiffRemoved},
		},
		{
			name:    "changed value",
			current: map[string]any{"key": "old_value"},
			backup:  map[string]any{"key": "new_value"},
			want:    []string{model.DiffChanged},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			diffs := CompareConfigs(tt.current, tt.backup)

			if len(diffs) != len(tt.want) {
				t.Errorf("got %d diffs, want %d", len(diffs), len(tt.want))
				return
			}

			for i, wantType := range tt.want {
				if diffs[i].DiffType != wantType {
					t.Errorf("diff[%d]: got type %v, want %v", i, diffs[i].DiffType, wantType)
				}
			}
		})
	}
}

func TestConvertBackupScripts(t *testing.T) {
	t.Parallel()

	scripts := []*shellybackup.Script{
		{ID: 1, Name: "script1", Enable: true, Code: "// code 1"},
		{ID: 2, Name: "script2", Enable: false, Code: "// code 2"},
	}

	result := ConvertBackupScripts(scripts)

	if len(result) != 2 {
		t.Fatalf("got %d scripts, want 2", len(result))
	}

	if result[0].ID != 1 || result[0].Name != "script1" || !result[0].Enable || result[0].Code != "// code 1" {
		t.Error("first script conversion incorrect")
	}
	if result[1].ID != 2 || result[1].Name != "script2" || result[1].Enable || result[1].Code != "// code 2" {
		t.Error("second script conversion incorrect")
	}
}

func TestConvertBackupSchedules(t *testing.T) {
	t.Parallel()

	data := json.RawMessage(`{
		"jobs": [
			{"enable": true, "timespec": "0 0 8 * * *", "calls": [{"method": "Switch.Set", "params": {"on": true}}]},
			{"enable": false, "timespec": "0 0 22 * * *", "calls": []}
		]
	}`)

	result := ConvertBackupSchedules(data)

	if len(result) != 2 {
		t.Fatalf("got %d schedules, want 2", len(result))
	}

	if !result[0].Enable || result[0].Timespec != "0 0 8 * * *" || len(result[0].Calls) != 1 {
		t.Error("first schedule conversion incorrect")
	}
	if result[1].Enable || result[1].Timespec != "0 0 22 * * *" || len(result[1].Calls) != 0 {
		t.Error("second schedule conversion incorrect")
	}
}

func TestConvertBackupSchedules_InvalidJSON(t *testing.T) {
	t.Parallel()

	data := json.RawMessage(`invalid json`)

	result := ConvertBackupSchedules(data)

	if result != nil {
		t.Errorf("expected nil for invalid JSON, got %v", result)
	}
}

func TestConvertBackupWebhooks(t *testing.T) {
	t.Parallel()

	data := json.RawMessage(`{
		"hooks": [
			{"id": 1, "cid": 0, "enable": true, "event": "switch.on", "name": "webhook1", "urls": ["http://example.com"]},
			{"id": 2, "cid": 1, "enable": false, "event": "switch.off", "name": "webhook2", "urls": []}
		]
	}`)

	result := ConvertBackupWebhooks(data)

	if len(result) != 2 {
		t.Fatalf("got %d webhooks, want 2", len(result))
	}

	if result[0].ID != 1 || result[0].Cid != 0 || !result[0].Enable || result[0].Event != "switch.on" {
		t.Error("first webhook conversion incorrect")
	}
}

func TestNewService(t *testing.T) {
	t.Parallel()

	// Use nil connector for this test
	svc := NewService(nil)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestRestoreResult_Fields(t *testing.T) {
	t.Parallel()

	result := RestoreResult{
		Success:           true,
		ConfigRestored:    true,
		ScriptsRestored:   3,
		SchedulesRestored: 2,
		WebhooksRestored:  1,
		RestartRequired:   true,
		Warnings:          []string{"warning1", "warning2"},
	}

	if !result.Success {
		t.Error("expected Success to be true")
	}
	if !result.ConfigRestored {
		t.Error("expected ConfigRestored to be true")
	}
	if result.ScriptsRestored != 3 {
		t.Errorf("got ScriptsRestored=%d, want 3", result.ScriptsRestored)
	}
	if result.SchedulesRestored != 2 {
		t.Errorf("got SchedulesRestored=%d, want 2", result.SchedulesRestored)
	}
	if result.WebhooksRestored != 1 {
		t.Errorf("got WebhooksRestored=%d, want 1", result.WebhooksRestored)
	}
	if !result.RestartRequired {
		t.Error("expected RestartRequired to be true")
	}
	if len(result.Warnings) != 2 {
		t.Errorf("got %d warnings, want 2", len(result.Warnings))
	}
}

func TestUpdateResultCounts(t *testing.T) {
	t.Parallel()

	result := &RestoreResult{}
	backup := &shellybackup.Backup{
		Scripts: []*shellybackup.Script{
			{ID: 1, Name: "script1"},
			{ID: 2, Name: "script2"},
		},
		Schedules: json.RawMessage(`{"jobs": [{"id": 1}, {"id": 2}, {"id": 3}]}`),
		Webhooks:  json.RawMessage(`{"hooks": [{"id": 1}]}`),
	}

	UpdateResultCounts(result, backup)

	if !result.ConfigRestored {
		t.Error("expected ConfigRestored to be true")
	}
	if result.ScriptsRestored != 2 {
		t.Errorf("got ScriptsRestored=%d, want 2", result.ScriptsRestored)
	}
	if result.SchedulesRestored != 3 {
		t.Errorf("got SchedulesRestored=%d, want 3", result.SchedulesRestored)
	}
	if result.WebhooksRestored != 1 {
		t.Errorf("got WebhooksRestored=%d, want 1", result.WebhooksRestored)
	}
}

func TestMigrationSource_Constants(t *testing.T) {
	t.Parallel()

	if SourceFile != "file" {
		t.Errorf("got SourceFile=%q, want %q", SourceFile, "file")
	}
	if SourceDevice != "device" {
		t.Errorf("got SourceDevice=%q, want %q", SourceDevice, "device")
	}
}

func TestScript_Fields(t *testing.T) {
	t.Parallel()

	script := Script{
		ID:     1,
		Name:   "test-script",
		Enable: true,
		Code:   "// script code",
	}

	if script.ID != 1 {
		t.Errorf("got ID=%d, want 1", script.ID)
	}
	if script.Name != "test-script" {
		t.Errorf("got Name=%q, want %q", script.Name, "test-script")
	}
	if !script.Enable {
		t.Error("expected Enable to be true")
	}
	if script.Code != "// script code" {
		t.Errorf("got Code=%q, want %q", script.Code, "// script code")
	}
}

func TestSchedule_Fields(t *testing.T) {
	t.Parallel()

	schedule := Schedule{
		Enable:   true,
		Timespec: "0 0 8 * * *",
		Calls: []ScheduleCall{
			{Method: "Switch.Set", Params: map[string]any{"on": true}},
		},
	}

	if !schedule.Enable {
		t.Error("expected Enable to be true")
	}
	if schedule.Timespec != "0 0 8 * * *" {
		t.Errorf("got Timespec=%q, want %q", schedule.Timespec, "0 0 8 * * *")
	}
	if len(schedule.Calls) != 1 {
		t.Errorf("got %d calls, want 1", len(schedule.Calls))
	}
}

func TestDiffBackups(t *testing.T) {
	t.Parallel()

	backup1 := &DeviceBackup{
		Backup: &shellybackup.Backup{
			Config: json.RawMessage(`{"key1": "value1"}`),
		},
	}
	backup2 := &DeviceBackup{
		Backup: &shellybackup.Backup{
			Config: json.RawMessage(`{"key1": "value2"}`),
		},
	}

	diff, err := DiffBackups(backup1, backup2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if diff == nil {
		t.Fatal("expected non-nil diff")
	}
	if len(diff.ConfigDiffs) == 0 {
		t.Error("expected at least one config diff")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestLoadAndValidate(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	filePath := testBackupFilePath

	// Create a valid backup file
	backupData := `{
		"version": 1,
		"device_info": {"id": "test-device", "model": "test-model"},
		"config": {"key": "value"},
		"created_at": "` + time.Now().Format(time.RFC3339) + `"
	}`

	if err := afero.WriteFile(config.Fs(), filePath, []byte(backupData), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	bkp, err := LoadAndValidate(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if bkp == nil {
		t.Fatal("expected non-nil backup")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestAutoSavePath(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	tests := []struct {
		name         string
		identifier   string
		mac          string
		deviceID     string
		format       string
		wantContains []string
	}{
		{
			name:         "config name with mac",
			identifier:   "back-porch",
			mac:          "7C87CE557FA0",
			deviceID:     "shellyplus1-7c87ce557fa0",
			format:       "json",
			wantContains: []string{"backups", "back-porch-7c87ce557fa0-", config.ExtJSON},
		},
		{
			name:         "short name gen1",
			identifier:   "fl",
			mac:          "C82B961166C0",
			deviceID:     "C82B961166C0",
			format:       "json",
			wantContains: []string{"backups", "fl-c82b961166c0-", config.ExtJSON},
		},
		{
			name:         "falls back to device ID",
			identifier:   "",
			mac:          "AABBCCDD",
			deviceID:     "shellyplus1-123",
			format:       "yaml",
			wantContains: []string{"backups", "shellyplus1-123-aabbccdd-", ".yaml"},
		},
		{
			name:         "empty everything",
			identifier:   "",
			mac:          "",
			deviceID:     "",
			format:       "json",
			wantContains: []string{"backups", "backup-", config.ExtJSON},
		},
	}

	for _, tt := range tests {
		bkp := &DeviceBackup{
			Backup: &shellybackup.Backup{
				DeviceInfo: &shellybackup.DeviceInfo{
					ID:  tt.deviceID,
					MAC: tt.mac,
				},
			},
		}

		path, err := AutoSavePath(tt.identifier, bkp, tt.format)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tt.name, err)
		}

		for _, substr := range tt.wantContains {
			if !strings.Contains(path, substr) {
				t.Errorf("%s: path %q missing %q", tt.name, path, substr)
			}
		}
	}
}

func TestLoadAndValidate_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := LoadAndValidate("/nonexistent/backup.json")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestLoadAndValidate_InvalidBackup(t *testing.T) {
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	filePath := testBackupFilePath

	// Create an invalid backup file
	if err := afero.WriteFile(config.Fs(), filePath, []byte(`{"version": 0}`), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := LoadAndValidate(filePath)
	if err == nil {
		t.Error("expected error for invalid backup")
	}
}

// gen1RestoreFixture drives a real shelly-go Gen1 restore against an in-process
// HTTP server so restoreGen1Backup's result-shaping branches run without any host
// or device I/O. fw seeds the device's live firmware; uptime/unixtime control the
// stability and clock the restore observes.
type gen1RestoreFixture struct {
	fw       string
	uptime   int
	unixtime int64
}

// newGen1Server starts an httptest server emulating the Gen1 REST endpoints a
// restore touches (/shelly, /settings, /status) and 200-replies to every write.
func (f gen1RestoreFixture) newGen1Server(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/shelly", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, map[string]any{
			"type": "SHSW-1", "mac": "AABBCCDDEEFF", "fw": f.fw, "gen": 1, "model": "SHSW-1",
		})
	})
	mux.HandleFunc("/status", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, map[string]any{"uptime": f.uptime, "unixtime": f.unixtime})
	})
	// /settings is both the live-firmware read (no query) and the write target for
	// every paced step (with a query); a flat object satisfies both.
	mux.HandleFunc("/settings", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, map[string]any{"fw": f.fw, "name": "shelly1"})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, map[string]any{})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func writeJSON(t *testing.T, w http.ResponseWriter, body any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		t.Errorf("encode response: %v", err)
	}
}

// fakeGen1Connector is a ShellyConnector whose WithGen1Connection bridges the
// restore onto a real client.Gen1Client pointed at serverURL. Only the Gen1 hook
// is exercised by restoreGen1Backup; the rest satisfy the interface.
type fakeGen1Connector struct {
	t         *testing.T
	serverURL string
}

func (c fakeGen1Connector) WithGen1Connection(ctx context.Context, _ string, fn func(*client.Gen1Client) error) error {
	gc, err := client.ConnectGen1(ctx, model.Device{Address: c.serverURL})
	if err != nil {
		return err
	}
	defer func() {
		if cerr := gc.Close(); cerr != nil {
			c.t.Logf("warning: gen1 close error: %v", cerr)
		}
	}()
	return fn(gc)
}

func (c fakeGen1Connector) WithConnection(context.Context, string, func(*client.Client) error) error {
	return nil
}
func (c fakeGen1Connector) IsGen1Device(context.Context, string) (bool, error) { return true, nil }
func (c fakeGen1Connector) DeviceInfo(context.Context, string) (*DeviceInfoResult, error) {
	return &DeviceInfoResult{}, nil
}
func (c fakeGen1Connector) GetConfig(context.Context, string) (map[string]any, error) {
	return map[string]any{}, nil
}
func (c fakeGen1Connector) ListScripts(context.Context, string) ([]ScriptInfoResult, error) {
	return nil, nil
}
func (c fakeGen1Connector) ListSchedules(context.Context, string) ([]ScheduleJobResult, error) {
	return nil, nil
}
func (c fakeGen1Connector) ListWebhooks(context.Context, string) ([]WebhookInfoResult, error) {
	return nil, nil
}

// modernGen1Backup builds a minimal backup carrying modern firmware (fast 750ms
// pacing) and two webhook actions, with everything but webhooks skippable so the
// success path runs a short sequence.
func modernGen1Backup(fw string) *DeviceBackup {
	return &DeviceBackup{Backup: &shellybackup.Backup{
		DeviceInfo: &shellybackup.DeviceInfo{Model: "SHSW-1"},
		Config:     json.RawMessage(`{"fw":"` + fw + `","name":"shelly1"}`),
		Webhooks:   json.RawMessage(`{"actions":[{"index":0},{"index":1}]}`),
	}}
}

const modernGen1FW = "20230913-111821/v1.14.0-gcb84623"

func TestRestoreGen1Backup_Success(t *testing.T) {
	t.Parallel()
	srv := gen1RestoreFixture{fw: modernGen1FW, uptime: 120, unixtime: 1699300000}.newGen1Server(t)
	svc := NewService(fakeGen1Connector{t: t, serverURL: srv.URL})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Skip every step except webhooks so a stable device completes quickly while
	// still exercising the success result-shaping and the webhook-count branch.
	res, err := svc.restoreGen1Backup(ctx, "dev", modernGen1Backup(modernGen1FW), RestoreOptions{
		SkipNetwork: true, SkipAuth: true, SkipState: true, SkipMeters: true,
	})
	if err != nil {
		t.Fatalf("restoreGen1Backup() error = %v", err)
	}
	if res == nil || !res.Success {
		t.Fatalf("expected successful result, got %+v", res)
	}
	if !res.ConfigRestored {
		t.Error("ConfigRestored should mirror Success on a clean restore")
	}
	if res.WebhooksRestored != 2 {
		t.Errorf("WebhooksRestored = %d, want 2", res.WebhooksRestored)
	}
	if len(res.Errors) != 0 {
		t.Errorf("expected no errors, got %v", res.Errors)
	}
	if res.DestabilizedStep != "" {
		t.Errorf("expected no destabilized step, got %q", res.DestabilizedStep)
	}
}

func TestRestoreGen1Backup_SuccessSkipWebhooks(t *testing.T) {
	t.Parallel()
	srv := gen1RestoreFixture{fw: modernGen1FW, uptime: 120, unixtime: 1699300000}.newGen1Server(t)
	svc := NewService(fakeGen1Connector{t: t, serverURL: srv.URL})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	res, err := svc.restoreGen1Backup(ctx, "dev", modernGen1Backup(modernGen1FW), RestoreOptions{
		SkipNetwork: true, SkipAuth: true, SkipState: true, SkipMeters: true, SkipWebhooks: true,
	})
	if err != nil {
		t.Fatalf("restoreGen1Backup() error = %v", err)
	}
	if res == nil || !res.Success {
		t.Fatalf("expected successful result, got %+v", res)
	}
	// The webhook count must stay zero when the webhook restore is skipped.
	if res.WebhooksRestored != 0 {
		t.Errorf("WebhooksRestored = %d, want 0 when skipped", res.WebhooksRestored)
	}
}

func TestRestoreGen1Backup_FirmwareDowngradeRefused(t *testing.T) {
	t.Parallel()
	// Device runs older firmware than the backup; with no derivable image (the
	// backup carries no model) and downgrade not allowed, RestoreGen1 returns an
	// error before any write, so restoreGen1Backup returns a nil result + error.
	srv := gen1RestoreFixture{fw: "20200101-000000/v1.9.0", uptime: 120, unixtime: 1699300000}.newGen1Server(t)
	svc := NewService(fakeGen1Connector{t: t, serverURL: srv.URL})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	backup := &DeviceBackup{Backup: &shellybackup.Backup{
		Config: json.RawMessage(`{"fw":"` + modernGen1FW + `"}`),
	}}
	res, err := svc.restoreGen1Backup(ctx, "dev", backup, RestoreOptions{})
	if err == nil {
		t.Fatal("expected firmware-downgrade refusal error")
	}
	if res != nil {
		t.Errorf("expected nil result on restore error, got %+v", res)
	}
}

func TestRestoreGen1Backup_Destabilized(t *testing.T) {
	t.Parallel()
	// A device stuck at uptime 0 never restabilizes; a short context bounds the
	// recovery poll so the first step halts the restore with a destabilized step.
	srv := gen1RestoreFixture{fw: modernGen1FW, uptime: 0, unixtime: 1699300000}.newGen1Server(t)
	svc := NewService(fakeGen1Connector{t: t, serverURL: srv.URL})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	res, err := svc.restoreGen1Backup(ctx, "dev", modernGen1Backup(modernGen1FW), RestoreOptions{
		SkipNetwork: true, SkipAuth: true, SkipState: true, SkipMeters: true,
	})
	if err == nil {
		t.Fatal("expected restore-halted error on a destabilized device")
	}
	if !strings.Contains(err.Error(), "restore halted") {
		t.Errorf("error should name the halt, got %v", err)
	}
	if res == nil {
		t.Fatal("expected a populated result alongside the halt error")
	}
	if res.DestabilizedStep == "" {
		t.Error("DestabilizedStep should be set on a halted restore")
	}
	if res.Success {
		t.Error("Success must be false on a halted restore")
	}
	// The library's per-step error must surface through errorStrings.
	if len(res.Errors) == 0 {
		t.Error("expected the destabilization error to surface in Errors")
	}
}

func TestErrorStrings(t *testing.T) {
	t.Parallel()
	if got := errorStrings(nil); got != nil {
		t.Errorf("nil slice should map to nil, got %v", got)
	}
	errs := []error{context.Canceled, http.ErrServerClosed}
	got := errorStrings(errs)
	if len(got) != 2 || got[0] != context.Canceled.Error() || got[1] != http.ErrServerClosed.Error() {
		t.Errorf("errorStrings rendered errors wrong: %v", got)
	}
}

func TestRestoreResult_Err(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		result     *RestoreResult
		wantFailed bool
		wantErrSub string // "" means Err() must be nil
	}{
		{name: "nil result", result: nil, wantFailed: false},
		{name: "clean success", result: &RestoreResult{Success: true}, wantFailed: false},
		{
			name:       "rejected sections list every section",
			result:     &RestoreResult{Success: false, Errors: []string{"wifi rejected", "mqtt rejected"}},
			wantFailed: true,
			wantErrSub: "2 section(s) rejected: wifi rejected; mqtt rejected",
		},
		{
			name:       "destabilized step takes precedence",
			result:     &RestoreResult{Success: false, Errors: []string{"x"}, DestabilizedStep: "coiot"},
			wantFailed: true,
			wantErrSub: "reboot loop after the \"coiot\" step",
		},
		{
			name:       "failure with no detail",
			result:     &RestoreResult{Success: false},
			wantFailed: true,
			wantErrSub: "did not complete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.result.Failed(); got != tt.wantFailed {
				t.Errorf("Failed() = %v, want %v", got, tt.wantFailed)
			}
			err := tt.result.Err()
			if tt.wantErrSub == "" {
				if err != nil {
					t.Errorf("Err() = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErrSub) {
				t.Errorf("Err() = %v, want substring %q", err, tt.wantErrSub)
			}
		})
	}
}
