package term

import (
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
)

func TestDisplayRestoreResult_SurfacesErrorsAndDestabilizedStep(t *testing.T) {
	t.Parallel()

	t.Run("errors and destabilized step are shown", func(t *testing.T) {
		t.Parallel()
		ios, out, errOut := testIOStreams()
		DisplayRestoreResult(ios, &backup.RestoreResult{
			Success:          false,
			Warnings:         []string{"light schedule rejected (no clock)"},
			Errors:           []string{"device became unstable after writing coiot"},
			DestabilizedStep: "coiot",
		})
		got := out.String() + errOut.String()
		for _, want := range []string{"Warnings", "Errors", "coiot", "reboot loop"} {
			if !strings.Contains(got, want) {
				t.Errorf("output missing %q; got:\n%s", want, got)
			}
		}
	})

	t.Run("clean result shows neither errors nor halt", func(t *testing.T) {
		t.Parallel()
		ios, out, errOut := testIOStreams()
		DisplayRestoreResult(ios, &backup.RestoreResult{Success: true, ConfigRestored: true})
		got := out.String() + errOut.String()
		if strings.Contains(got, "Errors") || strings.Contains(got, "reboot loop") {
			t.Errorf("clean restore should not mention errors or a halt; got:\n%s", got)
		}
	})
}

func TestRestoreResultError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		result  *backup.RestoreResult
		wantErr bool
		wantSub string
	}{
		{name: "nil result is success", result: nil, wantErr: false},
		{name: "clean success", result: &backup.RestoreResult{Success: true}, wantErr: false},
		{
			name:    "rejected sections",
			result:  &backup.RestoreResult{Success: false, Errors: []string{"wifi rejected", "mqtt rejected"}},
			wantErr: true,
			wantSub: "2 section(s) rejected",
		},
		{
			name:    "destabilized step wins over section count",
			result:  &backup.RestoreResult{Success: false, Errors: []string{"x"}, DestabilizedStep: "coiot"},
			wantErr: true,
			wantSub: "reboot loop",
		},
		{
			name:    "failure with no detail still errors",
			result:  &backup.RestoreResult{Success: false},
			wantErr: true,
			wantSub: "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := RestoreResultError("dev1", tt.result)
			if (err != nil) != tt.wantErr {
				t.Fatalf("RestoreResultError() err = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if !strings.Contains(err.Error(), "dev1") {
					t.Errorf("error should name the target; got %q", err)
				}
				if !strings.Contains(err.Error(), tt.wantSub) {
					t.Errorf("error %q missing %q", err, tt.wantSub)
				}
			}
		})
	}
}

func TestReportRestoreResult(t *testing.T) {
	t.Parallel()

	t.Run("partial failure prints no success line and returns an error", func(t *testing.T) {
		t.Parallel()
		ios, out, errOut := testIOStreams()
		err := ReportRestoreResult(ios, "dev1", &backup.RestoreResult{
			Success: false,
			Errors:  []string{"wifi rejected"},
		})
		if err == nil {
			t.Fatal("a rejected restore must return a non-nil error so the command exits non-zero")
		}
		got := out.String() + errOut.String()
		if strings.Contains(got, "Backup restored to") {
			t.Errorf("must not print a success line on failure; got:\n%s", got)
		}
		if !strings.Contains(got, "wifi rejected") {
			t.Errorf("the rejected section must be surfaced; got:\n%s", got)
		}
	})

	t.Run("clean restore prints success and returns nil", func(t *testing.T) {
		t.Parallel()
		ios, out, errOut := testIOStreams()
		err := ReportRestoreResult(ios, "dev1", &backup.RestoreResult{Success: true, ConfigRestored: true})
		if err != nil {
			t.Fatalf("clean restore should return nil; got %v", err)
		}
		got := out.String() + errOut.String()
		if !strings.Contains(got, "Backup restored to dev1") {
			t.Errorf("clean restore should print the success line; got:\n%s", got)
		}
	})
}

func TestDisplayBackupsTable(t *testing.T) {
	t.Parallel()

	t.Run("with backups", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		backups := []model.BackupFileInfo{
			{
				Filename:    "backup1.json",
				DeviceID:    testDeviceIDPlus1,
				DeviceModel: testModel1PM,
				FWVersion:   testFWVersion,
				CreatedAt:   time.Now(),
			},
			{
				Filename:    "backup2.json",
				DeviceID:    "shellyplus2pm-654321",
				DeviceModel: testModel2PM,
				FWVersion:   testFWVersionNew,
				CreatedAt:   time.Now(),
			},
		}

		DisplayBackupsTable(ios, backups)

		output := out.String()
		if !strings.Contains(output, "backup1.json") {
			t.Error("output should contain 'backup1.json'")
		}
		if !strings.Contains(output, "backup2.json") {
			t.Error("output should contain 'backup2.json'")
		}
	})

	t.Run("empty list", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		DisplayBackupsTable(ios, []model.BackupFileInfo{})

		output := out.String()
		// Should still produce output (table header or message)
		if output == "" {
			t.Error("output should not be empty")
		}
	})
}

func TestDisplayBackupExportResults(t *testing.T) {
	t.Parallel()

	t.Run("all success", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		results := []shelly.BackupResult{
			{DeviceName: testDevice1, Address: testIP101, Success: true},
			{DeviceName: testDevice2, Address: "192.168.1.102", Success: true},
		}

		DisplayBackupExportResults(ios, results)

		output := out.String()
		if !strings.Contains(output, testDevice1) {
			t.Error("output should contain 'device1'")
		}
		if !strings.Contains(output, "OK") {
			t.Error("output should contain 'OK' for successful backups")
		}
	})

	t.Run("with failures", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		results := []shelly.BackupResult{
			{DeviceName: testDevice1, Address: testIP101, Success: true},
			{DeviceName: testDevice2, Address: "192.168.1.102", Success: false},
		}

		DisplayBackupExportResults(ios, results)

		output := out.String()
		if !strings.Contains(output, "FAILED") {
			t.Error("output should contain 'FAILED' for failed backups")
		}
	})

	t.Run("empty results", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		DisplayBackupExportResults(ios, []shelly.BackupResult{})

		output := out.String()
		// Empty results should produce no output
		if output != "" {
			t.Errorf("output should be empty for no results, got %q", output)
		}
	})
}

func TestBackupFileInfo_Fields(t *testing.T) {
	t.Parallel()

	info := model.BackupFileInfo{
		Filename:    "backup.json",
		DeviceID:    testDeviceIDPlus1,
		DeviceModel: testModel1PM,
		FWVersion:   testFWVersion,
		Encrypted:   true,
		Size:        1024,
	}

	if info.Filename != "backup.json" {
		t.Errorf("got Filename=%q, want backup.json", info.Filename)
	}
	if info.DeviceID != testDeviceIDPlus1 {
		t.Errorf("got DeviceID=%q, want shellyplus1-123456", info.DeviceID)
	}
	if info.DeviceModel != testModel1PM {
		t.Errorf("got DeviceModel=%q, want SNSW-001P16EU", info.DeviceModel)
	}
	if info.FWVersion != testFWVersion {
		t.Errorf("got FWVersion=%q, want 1.0.0", info.FWVersion)
	}
	if !info.Encrypted {
		t.Error("expected Encrypted to be true")
	}
	if info.Size != 1024 {
		t.Errorf("got Size=%d, want 1024", info.Size)
	}
}

func TestBackupResult_Fields(t *testing.T) {
	t.Parallel()

	t.Run("success result", func(t *testing.T) {
		t.Parallel()

		result := shelly.BackupResult{
			DeviceName: testDevice1,
			Address:    testIP100,
			Success:    true,
		}

		if result.DeviceName != testDevice1 {
			t.Errorf("got DeviceName=%q, want device1", result.DeviceName)
		}
		if result.Address != testIP100 {
			t.Errorf("got Address=%q, want 192.168.1.100", result.Address)
		}
		if !result.Success {
			t.Error("expected Success to be true")
		}
	})

	t.Run("failure result", func(t *testing.T) {
		t.Parallel()

		result := shelly.BackupResult{
			DeviceName: testDevice2,
			Address:    testIP101,
			Success:    false,
		}

		if result.Success {
			t.Error("expected Success to be false")
		}
	})
}
