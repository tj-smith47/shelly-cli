package term

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestOutputReport_JSON(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	report := model.DeviceReport{
		Timestamp:  time.Now(),
		ReportType: "status",
		Devices: []model.DeviceReportInfo{
			{Name: "kitchen-light", IP: "192.168.1.100"},
		},
	}
	err := OutputReport(ios, report, "json", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "kitchen-light") {
		t.Error("expected device name in JSON")
	}
	if !strings.Contains(output, "192.168.1.100") {
		t.Error("expected IP in JSON")
	}
}

func TestOutputReport_Text(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	report := model.DeviceReport{
		Timestamp:  time.Now(),
		ReportType: "status",
		Devices: []model.DeviceReportInfo{
			{Name: "living-room", IP: "192.168.1.101"},
		},
	}
	err := OutputReport(ios, report, "text", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if output == "" {
		t.Error("expected text output")
	}
}

func TestOutputReport_InvalidFormat(t *testing.T) {
	t.Parallel()

	ios, _, _ := testIOStreams()
	report := model.DeviceReport{}
	err := OutputReport(ios, report, "invalid", "")
	if err == nil {
		t.Error("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "unknown format") {
		t.Errorf("expected unknown format error, got: %v", err)
	}
}

func TestOutputReport_ToFile(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	report := model.DeviceReport{
		Timestamp:  time.Now(),
		ReportType: "status",
		Devices: []model.DeviceReportInfo{
			{Name: "test-device", IP: "192.168.1.50"},
		},
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "report.json")

	err := OutputReport(ios, report, "json", outputPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check success message
	if !strings.Contains(out.String(), "Report saved to") {
		t.Error("expected save success message")
	}

	// Verify file was created
	content, err := os.ReadFile(outputPath) //nolint:gosec // G304: outputPath is from t.TempDir(), safe in test
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if !strings.Contains(string(content), "test-device") {
		t.Error("expected device name in file")
	}
}

func TestOutputReport_ToFileText(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	report := model.DeviceReport{
		Timestamp:  time.Now(),
		ReportType: "status",
		Devices: []model.DeviceReportInfo{
			{Name: "text-device", IP: "192.168.1.60"},
		},
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "report.txt")

	err := OutputReport(ios, report, "text", outputPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "Report saved to") {
		t.Error("expected save success message")
	}

	// Verify file exists
	_, err = os.Stat(outputPath)
	if err != nil {
		t.Error("expected output file to exist")
	}
}
