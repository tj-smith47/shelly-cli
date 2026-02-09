package monitor

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
)

// createTestModel creates a model with test statuses for navigation testing.
func createTestModel(statusCount int) Model {
	m := Model{
		Sizable:  panel.NewSizable(11, panel.NewScroller(0, 10)),
		statuses: make([]DeviceStatus, statusCount),
	}
	m = m.SetSize(100, 100)
	for i := range statusCount {
		m.statuses[i] = DeviceStatus{
			Name:   "test-device",
			Online: true,
		}
	}
	m.Scroller.SetItemCount(statusCount)
	return m
}

func TestScrollerCursorDown(t *testing.T) {
	t.Parallel()

	t.Run("moves cursor down", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(10)

		m.Scroller.CursorDown()

		if m.Cursor() != 1 {
			t.Errorf("expected cursor 1, got %d", m.Cursor())
		}
	})

	t.Run("stops at last item", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(5)
		m.Scroller.SetCursor(4)

		m.Scroller.CursorDown()

		if m.Cursor() != 4 {
			t.Errorf("expected cursor to stay at 4, got %d", m.Cursor())
		}
	})

	t.Run("handles empty list", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)

		m.Scroller.CursorDown()

		if m.Cursor() != 0 {
			t.Errorf("expected cursor to stay at 0, got %d", m.Cursor())
		}
	})
}

func TestScrollerCursorUp(t *testing.T) {
	t.Parallel()

	t.Run("moves cursor up", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(10)
		m.Scroller.SetCursor(5)

		m.Scroller.CursorUp()

		if m.Cursor() != 4 {
			t.Errorf("expected cursor 4, got %d", m.Cursor())
		}
	})

	t.Run("stops at first item", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(5)

		m.Scroller.CursorUp()

		if m.Cursor() != 0 {
			t.Errorf("expected cursor to stay at 0, got %d", m.Cursor())
		}
	})
}

func TestScrollerCursorToEnd(t *testing.T) {
	t.Parallel()

	t.Run("moves cursor to last item", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(10)

		m.Scroller.CursorToEnd()

		if m.Cursor() != 9 {
			t.Errorf("expected cursor 9, got %d", m.Cursor())
		}
	})

	t.Run("handles empty list", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)

		m.Scroller.CursorToEnd()

		if m.Cursor() != 0 {
			t.Errorf("expected cursor to stay at 0, got %d", m.Cursor())
		}
	})
}

func TestScrollerPageDown(t *testing.T) {
	t.Parallel()

	t.Run("moves cursor by visible rows", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(50)
		m = m.SetSize(100, 30)

		m.Scroller.PageDown()

		if m.Cursor() <= 0 {
			t.Errorf("expected cursor to move forward, got %d", m.Cursor())
		}
	})

	t.Run("stops at last item", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(5)
		m.Scroller.SetCursor(3)

		m.Scroller.PageDown()

		if m.Cursor() != 4 {
			t.Errorf("expected cursor 4, got %d", m.Cursor())
		}
	})

	t.Run("handles empty list", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)

		m.Scroller.PageDown()

		if m.Cursor() != 0 {
			t.Errorf("expected cursor to stay at 0, got %d", m.Cursor())
		}
	})
}

func TestScrollerPageUp(t *testing.T) {
	t.Parallel()

	t.Run("moves cursor backward", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(50)
		m = m.SetSize(100, 30)
		m.Scroller.SetCursor(20)

		m.Scroller.PageUp()

		if m.Cursor() >= 20 {
			t.Errorf("expected cursor to move backward from 20, got %d", m.Cursor())
		}
	})

	t.Run("stops at first item", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(10)
		m.Scroller.SetCursor(2)

		m.Scroller.PageUp()

		if m.Cursor() != 0 {
			t.Errorf("expected cursor 0, got %d", m.Cursor())
		}
	})
}

func TestSelectedDevice(t *testing.T) {
	t.Parallel()

	t.Run("returns nil for empty list", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(0)

		if m.SelectedDevice() != nil {
			t.Error("expected nil for empty list")
		}
	})

	t.Run("returns selected device", func(t *testing.T) {
		t.Parallel()
		m := createTestModel(5)
		m.statuses[2].Name = "selected-device"
		m.Scroller.SetCursor(2)

		selected := m.SelectedDevice()
		if selected == nil {
			t.Fatal("expected non-nil selected device")
		}
		if selected.Name != "selected-device" {
			t.Errorf("expected selected-device, got %s", selected.Name)
		}
	})
}

func TestSetSize(t *testing.T) {
	t.Parallel()

	m := createTestModel(0)
	m = m.SetSize(100, 50)

	if m.Width != 100 {
		t.Errorf("expected width 100, got %d", m.Width)
	}
	if m.Height != 50 {
		t.Errorf("expected height 50, got %d", m.Height)
	}
}

func TestAggregateMetrics_PM(t *testing.T) {
	t.Parallel()

	freq := 50.0
	energy := &model.PMEnergyCounters{Total: 1234.5}

	pms := []model.PMStatus{
		{APower: 100, Voltage: 230, Current: 0.43, Freq: &freq, AEnergy: energy},
		{APower: 50, Voltage: 231, Current: 0.22},
	}

	status := &DeviceStatus{}
	aggregateMetrics(status, pms, false)

	if status.Power != 150 {
		t.Errorf("expected power 150, got %f", status.Power)
	}
	if status.Voltage != 230 {
		t.Errorf("expected voltage 230 (first non-zero), got %f", status.Voltage)
	}
	if status.Current != 0.43 {
		t.Errorf("expected current 0.43 (first non-zero), got %f", status.Current)
	}
	if status.Frequency != 50 {
		t.Errorf("expected frequency 50, got %f", status.Frequency)
	}
	if status.TotalEnergy != 1234.5 {
		t.Errorf("expected total energy 1234.5, got %f", status.TotalEnergy)
	}
}

func TestAggregateMetrics_EM(t *testing.T) {
	t.Parallel()

	freq := 60.0
	ems := []model.EMStatus{
		{TotalActivePower: 500, AVoltage: 120, TotalCurrent: 4.2, AFreq: &freq},
		{TotalActivePower: 300, AVoltage: 121, TotalCurrent: 2.5},
	}

	status := &DeviceStatus{}
	aggregateMetrics(status, ems, true)

	if status.Power != 800 {
		t.Errorf("expected power 800, got %f", status.Power)
	}
	if status.Voltage != 120 {
		t.Errorf("expected voltage 120 (first non-zero), got %f", status.Voltage)
	}
	// EM accumulates current
	if status.Current != 6.7 {
		t.Errorf("expected current 6.7 (accumulated), got %f", status.Current)
	}
	if status.Frequency != 60 {
		t.Errorf("expected frequency 60, got %f", status.Frequency)
	}
	if status.TotalEnergy != 0 {
		t.Errorf("expected total energy 0 (EM has no energy), got %f", status.TotalEnergy)
	}
}

func TestAggregateMetrics_EM1(t *testing.T) {
	t.Parallel()

	freq := 50.0
	em1s := []model.EM1Status{
		{ActPower: 200, Voltage: 240, Current: 0.83, Freq: &freq},
		{ActPower: 100, Voltage: 241, Current: 0.42},
	}

	status := &DeviceStatus{}
	aggregateMetrics(status, em1s, false)

	if status.Power != 300 {
		t.Errorf("expected power 300, got %f", status.Power)
	}
	if status.Voltage != 240 {
		t.Errorf("expected voltage 240 (first non-zero), got %f", status.Voltage)
	}
	if status.Current != 0.83 {
		t.Errorf("expected current 0.83 (first non-zero), got %f", status.Current)
	}
	if status.Frequency != 50 {
		t.Errorf("expected frequency 50, got %f", status.Frequency)
	}
}

func TestAggregateMetrics_EmptySlice(t *testing.T) {
	t.Parallel()

	status := &DeviceStatus{}
	aggregateMetrics(status, []model.PMStatus{}, false)

	if status.Power != 0 {
		t.Errorf("expected power 0, got %f", status.Power)
	}
	if status.Voltage != 0 {
		t.Errorf("expected voltage 0, got %f", status.Voltage)
	}
}

func TestAggregateMetrics_MultipleTypes(t *testing.T) {
	t.Parallel()

	// Test aggregating across all three types into same status (like checkDeviceStatus does)
	freq := 50.0
	energy := &model.PMEnergyCounters{Total: 500}

	status := &DeviceStatus{}
	aggregateMetrics(status, []model.PMStatus{
		{APower: 100, Voltage: 230, Current: 0.43, Freq: &freq, AEnergy: energy},
	}, false)
	aggregateMetrics(status, []model.EMStatus{
		{TotalActivePower: 200, AVoltage: 231, TotalCurrent: 1.5},
	}, true)
	aggregateMetrics(status, []model.EM1Status{
		{ActPower: 50, Voltage: 232, Current: 0.21},
	}, false)

	// Power accumulated from all
	if status.Power != 350 {
		t.Errorf("expected power 350, got %f", status.Power)
	}
	// Voltage set from first PM (first non-zero)
	if status.Voltage != 230 {
		t.Errorf("expected voltage 230, got %f", status.Voltage)
	}
	// Current: PM sets 0.43, EM accumulates +1.5, EM1 skipped (already non-zero, not accumulate)
	expectedCurrent := 0.43 + 1.5
	if status.Current != expectedCurrent {
		t.Errorf("expected current %f, got %f", expectedCurrent, status.Current)
	}
	// Frequency from PM
	if status.Frequency != 50 {
		t.Errorf("expected frequency 50, got %f", status.Frequency)
	}
	// Energy from PM only
	if status.TotalEnergy != 500 {
		t.Errorf("expected total energy 500, got %f", status.TotalEnergy)
	}
}

func TestExtractHealthData(t *testing.T) {
	t.Parallel()

	t.Run("extracts sys data", func(t *testing.T) {
		t.Parallel()
		statusMap := map[string]json.RawMessage{
			"sys": json.RawMessage(`{
				"mac": "AABBCCDDEEFF",
				"fs_size": 458752,
				"fs_free": 229376,
				"ram_size": 262144,
				"ram_free": 131072,
				"uptime": 1000,
				"restart_required": false,
				"available_updates": {"stable": {"version": "1.2.3"}}
			}`),
		}
		var status DeviceStatus
		extractHealthData(statusMap, &status)

		if status.FSSize != 458752 {
			t.Errorf("FSSize = %d, want 458752", status.FSSize)
		}
		if status.FSFree != 229376 {
			t.Errorf("FSFree = %d, want 229376", status.FSFree)
		}
		if !status.HasUpdate {
			t.Error("expected HasUpdate true")
		}
	})

	t.Run("extracts wifi rssi", func(t *testing.T) {
		t.Parallel()
		statusMap := map[string]json.RawMessage{
			"wifi": json.RawMessage(`{"rssi": -65.0, "ssid": "MyNetwork", "status": "got ip"}`),
		}
		var status DeviceStatus
		extractHealthData(statusMap, &status)

		if status.WiFiRSSI == nil || *status.WiFiRSSI != -65.0 {
			t.Errorf("WiFiRSSI = %v, want -65.0", status.WiFiRSSI)
		}
	})

	t.Run("extracts chip temp from switch", func(t *testing.T) {
		t.Parallel()
		statusMap := map[string]json.RawMessage{
			"switch:0": json.RawMessage(`{"id": 0, "output": true, "temperature": {"tC": 72.5, "tF": 162.5}}`),
		}
		var status DeviceStatus
		extractHealthData(statusMap, &status)

		if status.ChipTemp == nil || *status.ChipTemp != 72.5 {
			t.Errorf("ChipTemp = %v, want 72.5", status.ChipTemp)
		}
	})

	t.Run("takes highest chip temp across components", func(t *testing.T) {
		t.Parallel()
		statusMap := map[string]json.RawMessage{
			"switch:0": json.RawMessage(`{"id": 0, "temperature": {"tC": 60.0}}`),
			"switch:1": json.RawMessage(`{"id": 1, "temperature": {"tC": 85.0}}`),
			"cover:0":  json.RawMessage(`{"id": 0, "temperature": {"tC": 70.0}}`),
		}
		var status DeviceStatus
		extractHealthData(statusMap, &status)

		if status.ChipTemp == nil || *status.ChipTemp != 85.0 {
			t.Errorf("ChipTemp = %v, want 85.0 (highest)", status.ChipTemp)
		}
	})

	t.Run("no update when stable missing", func(t *testing.T) {
		t.Parallel()
		statusMap := map[string]json.RawMessage{
			"sys": json.RawMessage(`{"fs_size": 100, "fs_free": 50}`),
		}
		var status DeviceStatus
		extractHealthData(statusMap, &status)

		if status.HasUpdate {
			t.Error("expected HasUpdate false when no updates available")
		}
	})

	t.Run("empty status map", func(t *testing.T) {
		t.Parallel()
		var status DeviceStatus
		extractHealthData(map[string]json.RawMessage{}, &status)

		if status.ChipTemp != nil {
			t.Error("expected nil ChipTemp")
		}
		if status.WiFiRSSI != nil {
			t.Error("expected nil WiFiRSSI")
		}
		if status.FSSize != 0 {
			t.Error("expected zero FSSize")
		}
	})

	t.Run("ignores non-component keys", func(t *testing.T) {
		t.Parallel()
		statusMap := map[string]json.RawMessage{
			"cloud":         json.RawMessage(`{"connected": true}`),
			"temperature:0": json.RawMessage(`{"tC": 25.0}`),
			"input:0":       json.RawMessage(`{"id": 0}`),
		}
		var status DeviceStatus
		extractHealthData(statusMap, &status)

		if status.ChipTemp != nil {
			t.Error("should not extract chip temp from non-component keys")
		}
	})
}

func TestIsComponentKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key  string
		want bool
	}{
		{"switch:0", true},
		{"switch:1", true},
		{"cover:0", true},
		{"light:0", true},
		{"rgb:0", true},
		{"rgbw:0", true},
		{"temperature:0", false},
		{"humidity:0", false},
		{"cloud", false},
		{"sys", false},
		{"wifi", false},
		{"em:0", false},
		{"input:0", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()
			got := isComponentKey(tt.key)
			if got != tt.want {
				t.Errorf("isComponentKey(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestOptionalFloat(t *testing.T) {
	t.Parallel()

	t.Run("nil returns empty", func(t *testing.T) {
		t.Parallel()
		if got := optionalFloat(nil, "%.1f"); got != "" {
			t.Errorf("optionalFloat(nil) = %q, want empty", got)
		}
	})

	t.Run("formats value", func(t *testing.T) {
		t.Parallel()
		v := 72.5
		if got := optionalFloat(&v, "%.1f"); got != "72.5" {
			t.Errorf("optionalFloat(72.5) = %q, want %q", got, "72.5")
		}
	})

	t.Run("respects format", func(t *testing.T) {
		t.Parallel()
		v := 42.0
		if got := optionalFloat(&v, "%.0f"); got != "42" {
			t.Errorf("optionalFloat(42.0, %%.0f) = %q, want %q", got, "42")
		}
	})
}

func TestOptionalInt(t *testing.T) {
	t.Parallel()

	t.Run("nil returns empty", func(t *testing.T) {
		t.Parallel()
		if got := optionalInt(nil); got != "" {
			t.Errorf("optionalInt(nil) = %q, want empty", got)
		}
	})

	t.Run("formats value", func(t *testing.T) {
		t.Parallel()
		v := 42
		if got := optionalInt(&v); got != "42" {
			t.Errorf("optionalInt(42) = %q, want %q", got, "42")
		}
	})
}

func TestExportCSV_IncludesHealthData(t *testing.T) {
	t.Parallel()

	memFs := afero.NewMemMapFs()
	path := "/tmp/test.csv"

	chipTemp := 85.0
	rssi := -80.0
	temp := 22.5
	m := Model{
		statuses: []DeviceStatus{
			{
				Name:        "test-device",
				Address:     "192.168.1.100",
				Type:        "switch",
				Online:      true,
				Power:       100.5,
				Temperature: &temp,
				ChipTemp:    &chipTemp,
				WiFiRSSI:    &rssi,
				FSFree:      10000,
				FSSize:      100000,
				HasUpdate:   true,
				UpdatedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	err := m.exportCSV(memFs, path)
	if err != nil {
		t.Fatalf("exportCSV failed: %v", err)
	}

	data, err := afero.ReadFile(memFs, path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	content := string(data)

	// Check header includes health columns
	if !strings.Contains(content, "chip_temp_c") {
		t.Error("CSV header missing chip_temp_c")
	}
	if !strings.Contains(content, "wifi_rssi_dbm") {
		t.Error("CSV header missing wifi_rssi_dbm")
	}
	if !strings.Contains(content, "fs_used_pct") {
		t.Error("CSV header missing fs_used_pct")
	}
	if !strings.Contains(content, "has_update") {
		t.Error("CSV header missing has_update")
	}

	// Check data row includes health values
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) < 2 {
		t.Fatal("expected at least 2 lines (header + data)")
	}
	dataLine := lines[1]
	if !strings.Contains(dataLine, "85.0") {
		t.Error("CSV data missing chip temp value")
	}
	if !strings.Contains(dataLine, "-80") {
		t.Error("CSV data missing WiFi RSSI value")
	}
	if !strings.Contains(dataLine, "true") {
		t.Error("CSV data missing has_update flag")
	}
}

func TestExportJSON_IncludesHealthData(t *testing.T) {
	t.Parallel()

	memFs := afero.NewMemMapFs()
	path := "/tmp/test.json"

	chipTemp := 85.0
	rssi := -80.0
	m := Model{
		statuses: []DeviceStatus{
			{
				Name:     "test-device",
				Address:  "192.168.1.100",
				Online:   true,
				Power:    100.5,
				ChipTemp: &chipTemp,
				WiFiRSSI: &rssi,
				FSFree:   10000,
				FSSize:   100000,
			},
		},
	}

	err := m.exportJSON(memFs, path)
	if err != nil {
		t.Fatalf("exportJSON failed: %v", err)
	}

	data, err := afero.ReadFile(memFs, path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	content := string(data)

	// Check JSON includes health section
	if !strings.Contains(content, `"health"`) {
		t.Error("JSON missing health section")
	}
	if !strings.Contains(content, `"chip_temp_c"`) {
		t.Error("JSON missing chip_temp_c field")
	}
	if !strings.Contains(content, `"wifi_rssi_dbm"`) {
		t.Error("JSON missing wifi_rssi_dbm field")
	}
	if !strings.Contains(content, `"fs_used_pct"`) {
		t.Error("JSON missing fs_used_pct field")
	}

	// Verify JSON is valid and can be decoded
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	devices, ok := result["devices"].([]any)
	if !ok || len(devices) == 0 {
		t.Fatal("expected devices array")
	}

	device, ok := devices[0].(map[string]any)
	if !ok {
		t.Fatal("expected device map")
	}
	health, ok := device["health"].(map[string]any)
	if !ok {
		t.Fatal("expected health object in device")
	}
	chipTemp, ctOK := health["chip_temp_c"].(float64)
	if !ctOK || chipTemp != 85.0 {
		t.Errorf("chip_temp_c = %v, want 85.0", health["chip_temp_c"])
	}
}

func TestExportJSON_NoHealthWhenEmpty(t *testing.T) {
	t.Parallel()

	memFs := afero.NewMemMapFs()
	path := "/tmp/test.json"

	m := Model{
		statuses: []DeviceStatus{
			{
				Name:    "basic-device",
				Address: "192.168.1.100",
				Online:  true,
				Power:   50,
			},
		},
	}

	err := m.exportJSON(memFs, path)
	if err != nil {
		t.Fatalf("exportJSON failed: %v", err)
	}

	data, err := afero.ReadFile(memFs, path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	// Health section should be omitted for devices without health data
	if strings.Contains(string(data), `"health"`) {
		t.Error("JSON should not include health section when no health data")
	}
}
