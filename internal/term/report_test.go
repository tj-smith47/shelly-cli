package term

import (
	"strings"
	"testing"
	"time"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

func TestOutputReport_JSON(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	report := model.DevicesReport{
		Timestamp:  time.Now(),
		ReportType: cmdStatus,
		Devices: []model.DeviceReportRow{
			{Name: "kitchen-light", IP: testIP100},
		},
	}
	err := OutputReport(ios, report, output.FormatDevicesReportText, "json", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "kitchen-light") {
		t.Error("expected device name in JSON")
	}
	if !strings.Contains(got, testIP100) {
		t.Error("expected IP in JSON")
	}
}

func TestOutputReport_Text(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	report := model.DevicesReport{
		Timestamp:  time.Now(),
		ReportType: cmdStatus,
		Devices: []model.DeviceReportRow{
			{Name: "living-room", IP: testIP101},
		},
	}
	err := OutputReport(ios, report, output.FormatDevicesReportText, "text", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := out.String()
	if got == "" {
		t.Error("expected text output")
	}
}

func TestOutputReport_InvalidFormat(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()
	report := model.DevicesReport{}
	err := OutputReport(ios, report, output.FormatDevicesReportText, "invalid", "")
	if err == nil {
		t.Error("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "unknown format") {
		t.Errorf("expected unknown format error, got: %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestOutputReport_ToFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	ios, out, _ := testIOStreams()
	report := model.DevicesReport{
		Timestamp:  time.Now(),
		ReportType: cmdStatus,
		Devices: []model.DeviceReportRow{
			{Name: "test-device", IP: testIP50},
		},
	}

	outputPath := "/test/output/report.json"

	err := OutputReport(ios, report, output.FormatDevicesReportText, "json", outputPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check success message
	if !strings.Contains(out.String(), "Report saved to") {
		t.Error("expected save success message")
	}

	// Verify file was created in virtual filesystem
	content, err := afero.ReadFile(fs, outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if !strings.Contains(string(content), "test-device") {
		t.Error("expected device name in file")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestOutputReport_ToFileText(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	ios, out, _ := testIOStreams()
	report := model.DevicesReport{
		Timestamp:  time.Now(),
		ReportType: cmdStatus,
		Devices: []model.DeviceReportRow{
			{Name: "text-device", IP: "192.168.1.60"},
		},
	}

	outputPath := "/test/output/report.txt"

	err := OutputReport(ios, report, output.FormatDevicesReportText, "text", outputPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "Report saved to") {
		t.Error("expected save success message")
	}

	// Verify file exists in virtual filesystem
	_, err = fs.Stat(outputPath)
	if err != nil {
		t.Error("expected output file to exist")
	}
}

func TestOutputReport_YAMLKeysMatchJSON(t *testing.T) {
	t.Parallel()

	report := model.EnergyReport{
		ReportType: model.ReportTypeEnergy,
		Devices:    []model.EnergyReportRow{{Name: "plug", Online: true, Reporting: true, PowerW: 45.2}},
		Summary:    model.EnergyReportSummary{Total: 1, Online: 1, DevicesReporting: 1, TotalPowerW: 45.2},
	}
	for _, format := range []string{"yaml", "table"} {
		ios, out, _ := testIOStreams()
		if err := OutputReport(ios, report, output.FormatEnergyReportText, format, ""); err != nil {
			t.Fatalf("%s: %v", format, err)
		}
		for _, want := range []string{"plug", "45.2"} {
			if !strings.Contains(out.String(), want) {
				t.Errorf("%s output lacks %q:\n%s", format, want, out.String())
			}
		}
	}
	ios, out, _ := testIOStreams()
	if err := OutputReport(ios, report, output.FormatEnergyReportText, "yaml", ""); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"report_type:", "power_w:", "total_power_w:", "devices_reporting:"} {
		if !strings.Contains(out.String(), key) {
			t.Errorf("yaml lacks key %s:\n%s", key, out.String())
		}
	}
}
