package term

import (
	"errors"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

func TestFirmwareStageStable(t *testing.T) {
	t.Parallel()
	if FirmwareStageStable != "stable" {
		t.Errorf("FirmwareStageStable = %q, want %q", FirmwareStageStable, "stable")
	}
}

func TestUpdateTarget_Fields(t *testing.T) {
	t.Parallel()

	target := UpdateTarget{
		DeviceID:    testDeviceIDPro1,
		DeviceModel: testModel1PM,
		Current:     testFWVersion,
		Available:   testFWVersionNew,
		Beta:        testFWVersionBeta,
		CustomURL:   "",
		UseBeta:     false,
	}

	if target.DeviceID != testDeviceIDPro1 {
		t.Errorf("DeviceID = %q, want shellypro1pm-123456", target.DeviceID)
	}
	if target.DeviceModel != testModel1PM {
		t.Errorf("DeviceModel = %q, want SNSW-001P16EU", target.DeviceModel)
	}
	if target.Current != testFWVersion {
		t.Errorf("Current = %q, want 1.0.0", target.Current)
	}
	if target.Available != testFWVersionNew {
		t.Errorf("Available = %q, want 1.1.0", target.Available)
	}
}

func TestUpdateResult_Fields(t *testing.T) {
	t.Parallel()

	t.Run("success result", func(t *testing.T) {
		t.Parallel()
		result := UpdateResult{
			Name:    testDevice1,
			Success: true,
			Err:     nil,
		}
		if result.Name != testDevice1 {
			t.Errorf("Name = %q, want device1", result.Name)
		}
		if !result.Success {
			t.Error("Success = false, want true")
		}
	})

	t.Run("failure result", func(t *testing.T) {
		t.Parallel()
		err := errors.New("update failed")
		result := UpdateResult{
			Name:    testDevice2,
			Success: false,
			Err:     err,
		}
		if result.Success {
			t.Error("Success = true, want false")
		}
		if !errors.Is(result.Err, err) {
			t.Errorf("Err = %v, want %v", result.Err, err)
		}
	})
}

func TestConvertToTermResults(t *testing.T) {
	t.Parallel()

	shellyResults := []shelly.UpdateResult{
		{Name: testDevice1, Success: true, Err: nil},
		{Name: testDevice2, Success: false, Err: errors.New("failed")},
	}

	termResults := ConvertToTermResults(shellyResults)

	if len(termResults) != 2 {
		t.Fatalf("len(termResults) = %d, want 2", len(termResults))
	}

	if termResults[0].Name != testDevice1 || !termResults[0].Success {
		t.Errorf("termResults[0] = %+v, want Name=device1 Success=true", termResults[0])
	}

	if termResults[1].Name != testDevice2 || termResults[1].Success {
		t.Errorf("termResults[1] = %+v, want Name=device2 Success=false", termResults[1])
	}
}

func TestFirmwareUpdateEntry_Fields(t *testing.T) {
	t.Parallel()

	entry := FirmwareUpdateEntry{
		Name: testDevice1,
		FwInfo: &shelly.FirmwareInfo{
			Current:   testFWVersion,
			Available: testFWVersionNew,
			Beta:      testFWVersionBeta,
			HasUpdate: true,
		},
		HasUpdate: true,
		HasBeta:   true,
	}

	if entry.Name != testDevice1 {
		t.Errorf("Name = %q, want device1", entry.Name)
	}
	if !entry.HasUpdate {
		t.Error("HasUpdate = false, want true")
	}
	if !entry.HasBeta {
		t.Error("HasBeta = false, want true")
	}
}

func TestConvertToTermEntries(t *testing.T) {
	t.Parallel()

	shellyEntries := []shelly.FirmwareUpdateEntry{
		{Name: testDevice1, HasUpdate: true, HasBeta: false},
		{Name: testDevice2, HasUpdate: true, HasBeta: true},
	}

	termEntries := ConvertToTermEntries(shellyEntries)

	if len(termEntries) != 2 {
		t.Fatalf("len(termEntries) = %d, want 2", len(termEntries))
	}

	if termEntries[0].Name != testDevice1 {
		t.Errorf("termEntries[0].Name = %q, want device1", termEntries[0].Name)
	}

	if termEntries[1].HasBeta != true {
		t.Error("termEntries[1].HasBeta = false, want true")
	}
}

func TestBuildFirmwareCheckRow_Success(t *testing.T) {
	t.Parallel()

	result := shelly.FirmwareCheckResult{
		Name: testDevice1,
		Info: &shelly.FirmwareInfo{
			Current:   testFWVersion,
			Available: testFWVersionNew,
			Beta:      testFWVersionBeta,
			HasUpdate: true,
			Platform:  testValueShelly,
		},
		Err: nil,
	}

	row := buildFirmwareCheckRow(result)

	if row.name != testDevice1 {
		t.Errorf("name = %q, want device1", row.name)
	}
	if row.platform != testValueShelly {
		t.Errorf("platform = %q, want shelly", row.platform)
	}
	if row.current != testFWVersion {
		t.Errorf("current = %q, want 1.0.0", row.current)
	}
	if row.stable != testFWVersionNew {
		t.Errorf("stable = %q, want 1.1.0", row.stable)
	}
	if row.beta != testFWVersionBeta {
		t.Errorf("beta = %q, want 1.2.0-beta1", row.beta)
	}
	if !row.hasUpdate {
		t.Error("hasUpdate = false, want true")
	}
}

func TestBuildFirmwareCheckRow_Error(t *testing.T) {
	t.Parallel()

	result := shelly.FirmwareCheckResult{
		Name: testDevice1,
		Info: nil,
		Err:  errors.New("connection failed"),
	}

	row := buildFirmwareCheckRow(result)

	if row.name != testDevice1 {
		t.Errorf("name = %q, want device1", row.name)
	}
	if row.hasUpdate {
		t.Error("hasUpdate = true, want false")
	}
	if !strings.Contains(row.stable, "connection failed") {
		t.Errorf("stable should contain error message, got %q", row.stable)
	}
}

func TestBuildFirmwareCheckRow_EmptyPlatform(t *testing.T) {
	t.Parallel()

	result := shelly.FirmwareCheckResult{
		Name: testDevice1,
		Info: &shelly.FirmwareInfo{
			Current:  testFWVersion,
			Platform: "",
		},
		Err: nil,
	}

	row := buildFirmwareCheckRow(result)

	if row.platform != testValueShelly {
		t.Errorf("platform = %q, want shelly (default)", row.platform)
	}
}

func TestDisplayFirmwareStatus(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	status := &shelly.FirmwareStatus{
		Status:      "idle",
		HasUpdate:   true,
		NewVersion:  testFWVersionNew,
		CanRollback: true,
		Progress:    0,
	}

	DisplayFirmwareStatus(ios, status)

	output := out.String()
	if output == "" {
		t.Error("DisplayFirmwareStatus should produce output")
	}
	if !strings.Contains(output, "Firmware Status") {
		t.Error("output should contain 'Firmware Status'")
	}
}

func TestDisplayFirmwareInfo(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	info := &shelly.FirmwareInfo{
		DeviceID:    testDeviceIDPro1,
		DeviceModel: testModel1PM,
		Generation:  2,
		Current:     testFWVersion,
		Available:   testFWVersionNew,
		Beta:        testFWVersionBeta,
		HasUpdate:   true,
	}

	DisplayFirmwareInfo(ios, info)

	output := out.String()
	if output == "" {
		t.Error("DisplayFirmwareInfo should produce output")
	}
	if !strings.Contains(output, "Firmware Information") {
		t.Error("output should contain 'Firmware Information'")
	}
	if !strings.Contains(output, "Gen2") {
		t.Error("output should contain 'Gen2'")
	}
}

func TestDisplayFirmwareCheckAll(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	results := []shelly.FirmwareCheckResult{
		{
			Name: testDevice1,
			Info: &shelly.FirmwareInfo{
				Current:   testFWVersion,
				Available: testFWVersionNew,
				HasUpdate: true,
			},
		},
		{
			Name: testDevice2,
			Info: &shelly.FirmwareInfo{
				Current:   testFWVersionNew,
				Available: testFWVersionNew,
				HasUpdate: false,
			},
		},
	}

	DisplayFirmwareCheckAll(ios, results)

	output := out.String()
	if output == "" {
		t.Error("DisplayFirmwareCheckAll should produce output")
	}
	if !strings.Contains(output, testDevice1) {
		t.Error("output should contain 'device1'")
	}
}

func TestDisplayUpdateTarget(t *testing.T) {
	t.Parallel()

	t.Run("stable update", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		target := UpdateTarget{
			DeviceID:    testDeviceIDPro1,
			DeviceModel: testModel1PM,
			Current:     testFWVersion,
			Available:   testFWVersionNew,
		}

		DisplayUpdateTarget(ios, target)

		output := out.String()
		if !strings.Contains(output, testFWVersion) {
			t.Error("output should contain current version")
		}
		if !strings.Contains(output, testFWVersionNew) {
			t.Error("output should contain target version")
		}
	})

	t.Run("beta update", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		target := UpdateTarget{
			DeviceID:    testDeviceIDPro1,
			DeviceModel: testModel1PM,
			Current:     testFWVersion,
			Beta:        testFWVersionBeta,
			UseBeta:     true,
		}

		DisplayUpdateTarget(ios, target)

		output := out.String()
		if !strings.Contains(output, "beta") {
			t.Error("output should contain 'beta'")
		}
	})

	t.Run("custom URL update", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		target := UpdateTarget{
			DeviceID:    testDeviceIDPro1,
			DeviceModel: testModel1PM,
			Current:     testFWVersion,
			CustomURL:   "http://example.com/firmware.bin",
		}

		DisplayUpdateTarget(ios, target)

		output := out.String()
		if !strings.Contains(output, "custom URL") {
			t.Error("output should contain 'custom URL'")
		}
	})
}

func TestDisplayUpdateResults(t *testing.T) {
	t.Parallel()

	t.Run("all success", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		results := []UpdateResult{
			{Name: testDevice1, Success: true},
			{Name: testDevice2, Success: true},
		}

		DisplayUpdateResults(ios, results)

		output := out.String()
		if !strings.Contains(output, "Updated") {
			t.Error("output should contain 'Updated'")
		}
	})

	t.Run("with failures", func(t *testing.T) {
		t.Parallel()
		ios, _, errOut := testIOStreams()
		results := []UpdateResult{
			{Name: testDevice1, Success: true},
			{Name: testDevice2, Success: false, Err: errors.New("connection failed")},
		}

		DisplayUpdateResults(ios, results)

		// Error/Warning messages go to errOut
		errOutput := errOut.String()
		if !strings.Contains(errOutput, "Failed") {
			t.Errorf("errOut should contain 'Failed', got %q", errOutput)
		}
	})
}

func TestDisplayDevicesToUpdate(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	devices := []shelly.DeviceUpdateStatus{
		{
			Name: testDevice1,
			Info: &shelly.FirmwareInfo{
				Current:   testFWVersion,
				Available: testFWVersionNew,
			},
		},
	}

	DisplayDevicesToUpdate(ios, devices)

	output := out.String()
	if output == "" {
		t.Error("DisplayDevicesToUpdate should produce output")
	}
	if !strings.Contains(output, testDevice1) {
		t.Error("output should contain 'device1'")
	}
}

func TestDisplayFirmwareUpdateInfo(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	info := &shelly.FirmwareInfo{
		DeviceModel: testModel1PM,
		Current:     testFWVersion,
		Available:   testFWVersionNew,
		Beta:        testFWVersionBeta,
		HasUpdate:   true,
	}

	DisplayFirmwareUpdateInfo(ios, info, testDevice1, testValueShelly)

	output := out.String()
	if output == "" {
		t.Error("DisplayFirmwareUpdateInfo should produce output")
	}
	if !strings.Contains(output, testDevice1) {
		t.Error("output should contain device name")
	}
	if !strings.Contains(output, testValueShelly) {
		t.Error("output should contain platform")
	}
}

func TestDisplayFirmwareUpdatesTable(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	devices := []FirmwareUpdateEntry{
		{
			Name: testDevice1,
			FwInfo: &shelly.FirmwareInfo{
				Current:   testFWVersion,
				Available: testFWVersionNew,
				Platform:  testValueShelly,
			},
			HasUpdate: true,
		},
	}

	DisplayFirmwareUpdatesTable(ios, devices)

	output := out.String()
	if output == "" {
		t.Error("DisplayFirmwareUpdatesTable should produce output")
	}
	if !strings.Contains(output, testDevice1) {
		t.Error("output should contain 'device1'")
	}
}
