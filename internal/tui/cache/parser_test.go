package cache

import (
	"encoding/json"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// --- Test Data: Gen2 Device Responses ---

// mockPlugSStatus simulates a Shelly Plug S response (Gen2 switch with power monitoring).
var mockPlugSStatus = map[string]json.RawMessage{
	"switch:0": json.RawMessage(`{
		"id": 0,
		"source": "init",
		"output": true,
		"apower": 45.2,
		"voltage": 230.5,
		"current": 0.196,
		"aenergy": {"total": 12345.67},
		"temperature": {"tC": 42.3}
	}`),
	"sys": json.RawMessage(`{
		"mac": "AABBCCDDEEFF",
		"uptime": 86400,
		"ram_free": 32768,
		"ram_size": 65536,
		"fs_free": 102400,
		"fs_size": 204800,
		"restart_required": false
	}`),
	"wifi": json.RawMessage(`{
		"sta_ip": "192.168.1.100",
		"status": "got ip",
		"ssid": "HomeNetwork",
		"rssi": -55
	}`),
}

// mockPlus2PMStatus simulates a Shelly Plus 2PM response (2 switches with PM).
var mockPlus2PMStatus = map[string]json.RawMessage{
	"switch:0": json.RawMessage(`{
		"id": 0,
		"output": true,
		"apower": 100.0,
		"voltage": 230.0,
		"current": 0.435
	}`),
	"switch:1": json.RawMessage(`{
		"id": 1,
		"output": false,
		"apower": 0.0,
		"voltage": 230.0,
		"current": 0.0
	}`),
	"sys": json.RawMessage(`{"mac": "112233445566", "uptime": 3600}`),
}

// mockPro3EMStatus simulates a Shelly Pro 3EM response (3-phase energy meter).
var mockPro3EMStatus = map[string]json.RawMessage{
	"em:0": json.RawMessage(`{
		"id": 0,
		"a_voltage": 230.1,
		"a_current": 5.5,
		"a_act_power": 1200.0,
		"b_voltage": 231.2,
		"b_current": 3.2,
		"b_act_power": 700.0,
		"c_voltage": 229.8,
		"c_current": 2.1,
		"c_act_power": 450.0,
		"total_current": 10.8,
		"total_act_power": 2350.0
	}`),
	"sys": json.RawMessage(`{"mac": "FFEEDDCCBBAA", "uptime": 7200}`),
}

// mockDimmer2Status simulates a Shelly Dimmer 2 response (light).
var mockDimmer2Status = map[string]json.RawMessage{
	"light:0": json.RawMessage(`{
		"id": 0,
		"output": true,
		"brightness": 75
	}`),
	"temperature:0": json.RawMessage(`{"tC": 38.5, "tF": 101.3}`),
	"sys":           json.RawMessage(`{"mac": "AABBCC112233", "uptime": 1800}`),
}

// mockCoverStatus simulates a Shelly 2.5 in cover mode response.
var mockCoverStatus = map[string]json.RawMessage{
	"cover:0": json.RawMessage(`{
		"id": 0,
		"state": "stopped",
		"apower": 0.0,
		"current_pos": 50,
		"target_pos": 50
	}`),
	"sys": json.RawMessage(`{"mac": "DDEEFF001122", "uptime": 600}`),
}

// mockPMOnlyStatus simulates a device with standalone PM components.
var mockPMOnlyStatus = map[string]json.RawMessage{
	"pm:0": json.RawMessage(`{
		"id": 0,
		"voltage": 231.5,
		"current": 2.5,
		"apower": 575.0,
		"aenergy": {"total": 99999.99}
	}`),
	"pm:1": json.RawMessage(`{
		"id": 1,
		"voltage": 232.0,
		"current": 1.2,
		"apower": 270.0,
		"aenergy": {"total": 55555.55}
	}`),
}

// mockEM1Status simulates a device with EM1 (single-phase energy meter) components.
var mockEM1Status = map[string]json.RawMessage{
	"em1:0": json.RawMessage(`{
		"id": 0,
		"voltage": 229.5,
		"current": 8.5,
		"act_power": 1900.0,
		"aprt_power": 2000.0
	}`),
}

// --- Test Data: Gen1 Device Responses ---

// mockGen1PlugStatus simulates a Gen1 Shelly Plug response.
var mockGen1PlugStatus = map[string]json.RawMessage{
	"relays":    json.RawMessage(`[{"ison": true, "source": "http"}]`),
	"meters":    json.RawMessage(`[{"power": 55.5, "total": 1234.56}]`),
	"wifi_sta":  json.RawMessage(`{"connected": true, "ssid": "Gen1Network", "ip": "192.168.1.50", "rssi": -60}`),
	"uptime":    json.RawMessage(`7200`),
	"ram_total": json.RawMessage(`51200`),
	"ram_free":  json.RawMessage(`25600`),
	"fs_size":   json.RawMessage(`233681`),
	"fs_free":   json.RawMessage(`156648`),
	"mac":       json.RawMessage(`"AABBCC001122"`),
}

// mockGen1EMStatus simulates a Gen1 Shelly EM response.
var mockGen1EMStatus = map[string]json.RawMessage{
	"relays":  json.RawMessage(`[{"ison": false}]`),
	"emeters": json.RawMessage(`[{"power": 500.0, "voltage": 230.0, "current": 2.17, "total": 5000.0}, {"power": 300.0, "voltage": 231.0, "current": 1.3, "total": 3000.0}]`),
	"uptime":  json.RawMessage(`3600`),
	"mac":     json.RawMessage(`"DDEEFF334455"`),
}

// mockGen1RGBWStatus simulates a Gen1 Shelly RGBW2 response.
var mockGen1RGBWStatus = map[string]json.RawMessage{
	"lights": json.RawMessage(`[{"ison": true, "mode": "color", "brightness": 100}]`),
	"uptime": json.RawMessage(`1800`),
	"mac":    json.RawMessage(`"112233AABBCC"`),
}

// mockGen1RollerStatus simulates a Gen1 Shelly 2.5 in roller mode response.
var mockGen1RollerStatus = map[string]json.RawMessage{
	"rollers": json.RawMessage(`[{"state": "stop", "power": 0.0, "current_pos": 75}]`),
	"uptime":  json.RawMessage(`900`),
	"mac":     json.RawMessage(`"FFEE00112233"`),
}

// mockGen1TempStatus simulates a Gen1 device with external temperature sensors.
var mockGen1TempStatus = map[string]json.RawMessage{
	"relays":          json.RawMessage(`[{"ison": false}]`),
	"temperature":     json.RawMessage(`45.6`),
	"ext_temperature": json.RawMessage(`{"0": {"tC": 23.5, "tF": 74.3, "is_valid": true}, "1": {"tC": 25.0, "tF": 77.0, "is_valid": true}}`),
	"uptime":          json.RawMessage(`600`),
	"mac":             json.RawMessage(`"AABBCCDDEE00"`),
}

// --- ParseFullStatus Tests ---

func TestParseFullStatus_PlugS(t *testing.T) {
	t.Parallel()
	parsed, err := ParseFullStatus("plug-s", mockPlugSStatus)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check switches
	if len(parsed.Switches) != 1 {
		t.Errorf("expected 1 switch, got %d", len(parsed.Switches))
	}
	if !parsed.Switches[0].On {
		t.Error("expected switch to be on")
	}
	if parsed.Switches[0].ID != 0 {
		t.Errorf("expected switch ID 0, got %d", parsed.Switches[0].ID)
	}

	// Check power metrics
	assertFloat(t, "power", 45.2, parsed.Power)
	assertFloat(t, "voltage", 230.5, parsed.Voltage)
	assertFloat(t, "current", 0.196, parsed.Current)
	assertFloat(t, "totalEnergy", 12345.67, parsed.TotalEnergy)
	assertFloat(t, "temperature", 42.3, parsed.Temperature)

	// Check WiFi
	if parsed.WiFi == nil {
		t.Fatal("expected WiFi to be parsed")
	}
	if parsed.WiFi.SSID != "HomeNetwork" {
		t.Errorf("expected SSID 'HomeNetwork', got '%s'", parsed.WiFi.SSID)
	}
	if parsed.WiFi.RSSI != -55 {
		t.Errorf("expected RSSI -55, got %d", parsed.WiFi.RSSI)
	}

	// Check Sys
	if parsed.Sys == nil {
		t.Fatal("expected Sys to be parsed")
	}
	if parsed.Sys.Uptime != 86400 {
		t.Errorf("expected uptime 86400, got %d", parsed.Sys.Uptime)
	}
	if parsed.MAC != "AABBCCDDEEFF" {
		t.Errorf("expected MAC 'AABBCCDDEEFF', got '%s'", parsed.MAC)
	}
}

func TestParseFullStatus_Plus2PM(t *testing.T) {
	t.Parallel()
	parsed, err := ParseFullStatus("plus-2pm", mockPlus2PMStatus)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(parsed.Switches) != 2 {
		t.Errorf("expected 2 switches, got %d", len(parsed.Switches))
	}

	// Check switch states (order may vary)
	var sw0, sw1 *SwitchState
	for i := range parsed.Switches {
		switch parsed.Switches[i].ID {
		case 0:
			sw0 = &parsed.Switches[i]
		case 1:
			sw1 = &parsed.Switches[i]
		}
	}

	if sw0 == nil || !sw0.On {
		t.Error("expected switch 0 to be on")
	}
	if sw1 == nil || sw1.On {
		t.Error("expected switch 1 to be off")
	}

	// Power should aggregate from both switches
	assertFloat(t, "power", 100.0, parsed.Power)
}

func TestParseFullStatus_Pro3EM(t *testing.T) {
	t.Parallel()
	parsed, err := ParseFullStatus("pro-3em", mockPro3EMStatus)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(parsed.EM) != 1 {
		t.Errorf("expected 1 EM component, got %d", len(parsed.EM))
	}

	assertFloat(t, "power", 2350.0, parsed.Power)
	assertFloat(t, "current", 10.8, parsed.Current)
	assertFloat(t, "voltage", 230.1, parsed.Voltage)
}

func TestParseFullStatus_Dimmer(t *testing.T) {
	t.Parallel()
	parsed, err := ParseFullStatus("dimmer-2", mockDimmer2Status)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(parsed.Lights) != 1 {
		t.Errorf("expected 1 light, got %d", len(parsed.Lights))
	}
	if !parsed.Lights[0].On {
		t.Error("expected light to be on")
	}

	assertFloat(t, "temperature", 38.5, parsed.Temperature)
}

func TestParseFullStatus_Cover(t *testing.T) {
	t.Parallel()
	parsed, err := ParseFullStatus("cover", mockCoverStatus)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(parsed.Covers) != 1 {
		t.Errorf("expected 1 cover, got %d", len(parsed.Covers))
	}
	if parsed.Covers[0].State != "stopped" {
		t.Errorf("expected state 'stopped', got '%s'", parsed.Covers[0].State)
	}
}

func TestParseFullStatus_MultiplePM(t *testing.T) {
	t.Parallel()
	parsed, err := ParseFullStatus("multi-pm", mockPMOnlyStatus)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(parsed.PM) != 2 {
		t.Errorf("expected 2 PM components, got %d", len(parsed.PM))
	}

	// Power should aggregate
	assertFloat(t, "power", 845.0, parsed.Power)
	assertFloat(t, "totalEnergy", 155555.54, parsed.TotalEnergy)
}

func TestParseFullStatus_EM1(t *testing.T) {
	t.Parallel()
	parsed, err := ParseFullStatus("em1-device", mockEM1Status)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(parsed.EM1) != 1 {
		t.Errorf("expected 1 EM1 component, got %d", len(parsed.EM1))
	}

	assertFloat(t, "power", 1900.0, parsed.Power)
	assertFloat(t, "voltage", 229.5, parsed.Voltage)
	assertFloat(t, "current", 8.5, parsed.Current)
}

// --- ParseGen1Status Tests ---

func TestParseGen1Status_Plug(t *testing.T) {
	t.Parallel()
	parsed, err := ParseGen1Status("gen1-plug", mockGen1PlugStatus)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsed.Generation != 1 {
		t.Errorf("expected generation 1, got %d", parsed.Generation)
	}

	if len(parsed.Switches) != 1 {
		t.Errorf("expected 1 switch, got %d", len(parsed.Switches))
	}
	if !parsed.Switches[0].On {
		t.Error("expected switch to be on")
	}

	assertFloat(t, "power", 55.5, parsed.Power)
	assertFloat(t, "totalEnergy", 1234.56, parsed.TotalEnergy)

	// Check WiFi
	if parsed.WiFi == nil {
		t.Fatal("expected WiFi to be parsed")
	}
	if parsed.WiFi.SSID != "Gen1Network" {
		t.Errorf("expected SSID 'Gen1Network', got '%s'", parsed.WiFi.SSID)
	}
}

func TestParseGen1Status_EM(t *testing.T) {
	t.Parallel()
	parsed, err := ParseGen1Status("gen1-em", mockGen1EMStatus)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(parsed.Switches) != 1 {
		t.Errorf("expected 1 switch (relay), got %d", len(parsed.Switches))
	}

	// Power aggregates from both emeters
	assertFloat(t, "power", 800.0, parsed.Power)
	assertFloat(t, "current", 3.47, parsed.Current)
	assertFloat(t, "voltage", 230.0, parsed.Voltage)
	assertFloat(t, "totalEnergy", 8000.0, parsed.TotalEnergy)
}

func TestParseGen1Status_RGBW(t *testing.T) {
	t.Parallel()
	parsed, err := ParseGen1Status("gen1-rgbw", mockGen1RGBWStatus)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(parsed.Lights) != 1 {
		t.Errorf("expected 1 light, got %d", len(parsed.Lights))
	}
	if !parsed.Lights[0].On {
		t.Error("expected light to be on")
	}
}

func TestParseGen1Status_Roller(t *testing.T) {
	t.Parallel()
	parsed, err := ParseGen1Status("gen1-roller", mockGen1RollerStatus)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(parsed.Covers) != 1 {
		t.Errorf("expected 1 cover, got %d", len(parsed.Covers))
	}
	// "stop" should be normalized to "stopped"
	if parsed.Covers[0].State != CoverStateStopped {
		t.Errorf("expected state '%s', got '%s'", CoverStateStopped, parsed.Covers[0].State)
	}
}

func TestParseGen1Status_Temperature(t *testing.T) {
	t.Parallel()
	parsed, err := ParseGen1Status("gen1-temp", mockGen1TempStatus)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should get temperature from the first source
	if parsed.Temperature == 0 {
		t.Error("expected temperature to be set")
	}
}

// --- ApplyParsedStatus Tests ---

func TestApplyParsedStatus_Basic(t *testing.T) {
	t.Parallel()
	data := &DeviceData{}
	parsed := &ParsedStatus{
		DeviceID: "test-device",
		Switches: []SwitchState{{ID: 0, On: true}},
		Power:    100.0,
		Voltage:  230.0,
		MAC:      "AABBCCDDEEFF",
	}

	ApplyParsedStatus(data, parsed)

	if len(data.Switches) != 1 {
		t.Errorf("expected 1 switch, got %d", len(data.Switches))
	}
	if data.Power != 100.0 {
		t.Errorf("expected power 100.0, got %f", data.Power)
	}
	if data.Device.MAC != "AABBCCDDEEFF" {
		t.Errorf("expected MAC 'AABBCCDDEEFF', got '%s'", data.Device.MAC)
	}
}

func TestApplyParsedStatus_PreservesExisting(t *testing.T) {
	t.Parallel()
	data := &DeviceData{
		Switches: []SwitchState{{ID: 0, On: false}},
	}
	data.Device.Model = "ExistingModel"

	parsed := &ParsedStatus{
		DeviceID: "test-device",
		// Empty switches - should preserve existing
		Power: 50.0,
		Model: "NewModel", // Should NOT override existing
	}

	ApplyParsedStatus(data, parsed)

	// Switches should be preserved (parsed has none)
	if len(data.Switches) != 1 {
		t.Errorf("expected 1 switch (preserved), got %d", len(data.Switches))
	}

	// Model should be preserved
	if data.Device.Model != "ExistingModel" {
		t.Errorf("expected model 'ExistingModel', got '%s'", data.Device.Model)
	}

	// Power should be updated
	if data.Power != 50.0 {
		t.Errorf("expected power 50.0, got %f", data.Power)
	}
}

func TestApplyParsedStatus_WiFiAndSys(t *testing.T) {
	t.Parallel()
	data := &DeviceData{}
	parsed := &ParsedStatus{
		WiFi: &WiFiInfo{SSID: "TestNetwork", RSSI: -50},
		Sys:  &SysInfo{Uptime: 3600, RAMFree: 16384},
	}

	ApplyParsedStatus(data, parsed)

	if data.WiFi == nil {
		t.Fatal("expected WiFi to be set")
	}
	if data.WiFi.SSID != "TestNetwork" {
		t.Errorf("expected SSID 'TestNetwork', got '%s'", data.WiFi.SSID)
	}

	if data.Sys == nil {
		t.Fatal("expected Sys to be set")
	}
	if data.Sys.Uptime != 3600 {
		t.Errorf("expected uptime 3600, got %d", data.Sys.Uptime)
	}
}

func TestApplyParsedStatus_Snapshot(t *testing.T) {
	t.Parallel()
	data := &DeviceData{}
	parsed := &ParsedStatus{
		PM: []model.PMStatus{{ID: 0, APower: 100.0}},
		EM: []model.EMStatus{{ID: 0, TotalActivePower: 500.0}},
	}

	ApplyParsedStatus(data, parsed)

	if data.Snapshot == nil {
		t.Fatal("expected Snapshot to be created")
	}
	if len(data.Snapshot.PM) != 1 {
		t.Errorf("expected 1 PM in snapshot, got %d", len(data.Snapshot.PM))
	}
	if len(data.Snapshot.EM) != 1 {
		t.Errorf("expected 1 EM in snapshot, got %d", len(data.Snapshot.EM))
	}
}

// --- ApplyIncrementalUpdate Tests ---

func TestApplyIncrementalUpdate_Switch(t *testing.T) {
	t.Parallel()
	data := &DeviceData{
		Switches: []SwitchState{{ID: 0, On: false}, {ID: 1, On: false}},
	}

	// Update switch 0 to on
	status := json.RawMessage(`{"output": true, "apower": 50.0}`)
	ApplyIncrementalUpdate(data, ComponentSwitch, 0, status)

	if !data.Switches[0].On {
		t.Error("expected switch 0 to be on")
	}
	if data.Switches[1].On {
		t.Error("expected switch 1 to still be off")
	}
	if data.Power != 50.0 {
		t.Errorf("expected power 50.0, got %f", data.Power)
	}
}

func TestApplyIncrementalUpdate_NewSwitch(t *testing.T) {
	t.Parallel()
	data := &DeviceData{}

	// Add new switch that doesn't exist yet
	status := json.RawMessage(`{"output": true}`)
	ApplyIncrementalUpdate(data, ComponentSwitch, 2, status)

	if len(data.Switches) != 1 {
		t.Errorf("expected 1 switch, got %d", len(data.Switches))
	}
	if data.Switches[0].ID != 2 || !data.Switches[0].On {
		t.Error("expected switch 2 to be added and on")
	}
}

func TestApplyIncrementalUpdate_Light(t *testing.T) {
	t.Parallel()
	data := &DeviceData{
		Lights: []LightState{{ID: 0, On: true}},
	}

	status := json.RawMessage(`{"output": false}`)
	ApplyIncrementalUpdate(data, ComponentLight, 0, status)

	if data.Lights[0].On {
		t.Error("expected light 0 to be off")
	}
}

func TestApplyIncrementalUpdate_Cover(t *testing.T) {
	t.Parallel()
	data := &DeviceData{
		Covers: []CoverState{{ID: 0, State: CoverStateStopped}},
	}

	status := json.RawMessage(`{"state": "opening", "apower": 25.0}`)
	ApplyIncrementalUpdate(data, ComponentCover, 0, status)

	if data.Covers[0].State != "opening" {
		t.Errorf("expected state 'opening', got '%s'", data.Covers[0].State)
	}
	if data.Power != 25.0 {
		t.Errorf("expected power 25.0, got %f", data.Power)
	}
}

func TestApplyIncrementalUpdate_PM(t *testing.T) {
	t.Parallel()
	data := &DeviceData{}

	status := json.RawMessage(`{"apower": 150.0, "voltage": 231.0, "current": 0.65}`)
	ApplyIncrementalUpdate(data, ComponentPM, 0, status)

	assertFloat(t, "power", 150.0, data.Power)
	assertFloat(t, "voltage", 231.0, data.Voltage)
	assertFloat(t, "current", 0.65, data.Current)
}

func TestApplyIncrementalUpdate_EM(t *testing.T) {
	t.Parallel()
	data := &DeviceData{}

	status := json.RawMessage(`{"total_act_power": 2500.0, "total_current": 11.0, "a_voltage": 230.0}`)
	ApplyIncrementalUpdate(data, ComponentEM, 0, status)

	assertFloat(t, "power", 2500.0, data.Power)
	assertFloat(t, "current", 11.0, data.Current)
	assertFloat(t, "voltage", 230.0, data.Voltage)
}

// --- ParseComponentName Tests ---

func TestParseComponentName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		wantType string
		wantID   int
	}{
		{"switch:0", ComponentSwitch, 0},
		{"switch:1", ComponentSwitch, 1},
		{"light:2", ComponentLight, 2},
		{"cover:0", ComponentCover, 0},
		{"pm:0", ComponentPM, 0},
		{"em:1", ComponentEM, 1},
		{"sys", ComponentSys, 0},
		{"wifi", ComponentWiFi, 0},
		{"temperature:3", ComponentTemperature, 3},
		{"invalid", "invalid", 0},
		{"switch:notanumber", ComponentSwitch, 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			gotType, gotID := ParseComponentName(tt.input)
			if gotType != tt.wantType {
				t.Errorf("ParseComponentName(%q) type = %q, want %q", tt.input, gotType, tt.wantType)
			}
			if gotID != tt.wantID {
				t.Errorf("ParseComponentName(%q) id = %d, want %d", tt.input, gotID, tt.wantID)
			}
		})
	}
}

// --- DetectComponents Tests ---

func TestDetectComponents(t *testing.T) {
	t.Parallel()
	tests := []struct {
		model      string
		wantSwitch bool
		wantLight  bool
		wantCover  bool
		wantPM     bool
		wantEM     bool
	}{
		{"SHPLG-S", true, false, false, true, false},        // Plug S (Gen1)
		{"SHSW-PM", true, false, false, true, false},        // 1PM (Gen1)
		{"SNPL-00116US", true, false, false, true, false},   // Plus Plug US
		{"SNSW-001X16EU", true, false, false, false, false}, // Plus 1 (no PM)
		{"SNSW-102PM16EU", true, false, false, true, false}, // Plus 2PM
		{"SHRGBW2", false, true, false, false, false},       // RGBW2
		{"SHDM-2", false, true, false, false, false},        // Dimmer 2
		{"SHSW-25", true, false, true, false, false},        // 2.5 (cover or switch mode)
		{"SPEM-003CEBEU", false, false, false, false, true}, // Pro 3EM
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			t.Parallel()
			caps := DetectComponents(tt.model)
			if caps.HasSwitches != tt.wantSwitch {
				t.Errorf("HasSwitches = %v, want %v", caps.HasSwitches, tt.wantSwitch)
			}
			if caps.HasLights != tt.wantLight {
				t.Errorf("HasLights = %v, want %v", caps.HasLights, tt.wantLight)
			}
			if caps.HasCovers != tt.wantCover {
				t.Errorf("HasCovers = %v, want %v", caps.HasCovers, tt.wantCover)
			}
			if caps.HasPM != tt.wantPM {
				t.Errorf("HasPM = %v, want %v", caps.HasPM, tt.wantPM)
			}
			if caps.HasEM != tt.wantEM {
				t.Errorf("HasEM = %v, want %v", caps.HasEM, tt.wantEM)
			}
		})
	}
}

// --- Helper Functions ---

func assertFloat(t *testing.T, name string, expected, actual float64) {
	t.Helper()
	const epsilon = 0.01
	diff := expected - actual
	if diff < 0 {
		diff = -diff
	}
	if diff > epsilon {
		t.Errorf("%s: expected %f, got %f", name, expected, actual)
	}
}
