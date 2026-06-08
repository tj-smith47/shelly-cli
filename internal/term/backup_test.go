package term

import (
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

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
