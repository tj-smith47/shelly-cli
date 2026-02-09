package monitor

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func ptr[T any](v T) *T {
	return &v
}

func TestNewEnvironment(t *testing.T) {
	t.Parallel()

	m := NewEnvironment()
	if len(m.devices) != 0 {
		t.Errorf("expected 0 devices initially, got %d", len(m.devices))
	}
	if m.focused {
		t.Error("expected not focused initially")
	}
}

func TestEnvironmentModel_SetDevices(t *testing.T) {
	t.Parallel()

	m := NewEnvironment()
	m = m.SetSize(80, 30)

	statuses := []DeviceStatus{
		{
			Name:   "kitchen",
			Online: true,
			Sensors: &model.SensorData{
				Temperature: []model.TemperatureReading{
					{ID: 0, TC: ptr(22.5)},
				},
				Humidity: []model.HumidityReading{
					{ID: 0, RH: ptr(55.0)},
				},
			},
		},
		{
			Name:   "office",
			Online: true,
			Sensors: &model.SensorData{
				Temperature: []model.TemperatureReading{
					{ID: 0, TC: ptr(24.0)},
				},
			},
		},
		{Name: "offline-device", Online: false},
	}

	m = m.SetDevices(statuses)

	// Should only include online devices with sensors
	if len(m.devices) != 2 {
		t.Errorf("expected 2 devices (online only), got %d", len(m.devices))
	}
}

func TestEnvironmentModel_SetDevices_IgnoresNilSensors(t *testing.T) {
	t.Parallel()

	m := NewEnvironment()
	m = m.SetSize(80, 30)

	statuses := []DeviceStatus{
		{Name: "no-sensors", Online: true, Sensors: nil},
	}

	m = m.SetDevices(statuses)

	if len(m.devices) != 0 {
		t.Errorf("expected 0 devices with nil sensors, got %d", len(m.devices))
	}
}

func TestEnvironmentModel_SetFocused(t *testing.T) {
	t.Parallel()

	m := NewEnvironment()
	if m.IsFocused() {
		t.Error("expected not focused initially")
	}

	m = m.SetFocused(true)
	if !m.IsFocused() {
		t.Error("expected focused after SetFocused(true)")
	}

	m = m.SetFocused(false)
	if m.IsFocused() {
		t.Error("expected not focused after SetFocused(false)")
	}
}

func TestEnvironmentModel_SetSize(t *testing.T) {
	t.Parallel()

	m := NewEnvironment()
	m = m.SetSize(100, 50)

	if m.Width != 100 {
		t.Errorf("expected width 100, got %d", m.Width)
	}
	if m.Height != 50 {
		t.Errorf("expected height 50, got %d", m.Height)
	}
}

func TestEnvironmentModel_View_TooSmall(t *testing.T) {
	t.Parallel()

	m := NewEnvironment()
	m = m.SetSize(5, 2)
	if m.View() != "" {
		t.Error("expected empty view for tiny dimensions")
	}
}

func TestEnvironmentModel_View_NoDevices(t *testing.T) {
	t.Parallel()

	m := NewEnvironment()
	m = m.SetSize(80, 20)
	m = m.SetDevices([]DeviceStatus{})

	view := m.View()
	if !strings.Contains(view, "Environment") {
		t.Error("expected view to contain 'Environment' title")
	}
	if !strings.Contains(view, "Safety") {
		t.Error("expected view to contain 'Safety' section")
	}
	if !strings.Contains(view, "No safety sensors configured") {
		t.Error("expected 'No safety sensors configured' message")
	}
}

func TestEnvironmentModel_View_WithTemperature(t *testing.T) {
	t.Parallel()

	m := NewEnvironment()
	m = m.SetSize(80, 30)
	m = m.SetDevices([]DeviceStatus{
		{
			Name:   "kitchen",
			Online: true,
			Sensors: &model.SensorData{
				Temperature: []model.TemperatureReading{
					{ID: 0, TC: ptr(22.5)},
				},
			},
		},
	})

	view := m.View()
	if !strings.Contains(view, "Temperature") {
		t.Error("expected view to contain 'Temperature' section")
	}
	if !strings.Contains(view, "kitchen") {
		t.Error("expected view to contain device name 'kitchen'")
	}
	if !strings.Contains(view, "22.5°C") {
		t.Error("expected view to contain temperature value")
	}
}

func TestEnvironmentModel_View_WithHumidity(t *testing.T) {
	t.Parallel()

	m := NewEnvironment()
	m = m.SetSize(80, 30)
	m = m.SetDevices([]DeviceStatus{
		{
			Name:   "bathroom",
			Online: true,
			Sensors: &model.SensorData{
				Humidity: []model.HumidityReading{
					{ID: 0, RH: ptr(75.0)},
				},
			},
		},
	})

	view := m.View()
	if !strings.Contains(view, "Humidity") {
		t.Error("expected view to contain 'Humidity' section")
	}
	if !strings.Contains(view, "75%") {
		t.Error("expected view to contain humidity value")
	}
}

func TestEnvironmentModel_View_WithIlluminance(t *testing.T) {
	t.Parallel()

	m := NewEnvironment()
	m = m.SetSize(80, 30)

	illumination := "bright"
	m = m.SetDevices([]DeviceStatus{
		{
			Name:   "hallway",
			Online: true,
			Sensors: &model.SensorData{
				Illuminance: []model.IlluminanceReading{
					{ID: 0, Lux: ptr(500.0), Illumination: &illumination},
				},
			},
		},
	})

	view := m.View()
	if !strings.Contains(view, "Illuminance") {
		t.Error("expected view to contain 'Illuminance' section")
	}
	if !strings.Contains(view, "500 lux") {
		t.Error("expected view to contain lux value")
	}
	if !strings.Contains(view, "bright") {
		t.Error("expected view to contain illumination level")
	}
}

func TestEnvironmentModel_View_WithIlluminanceNoLevel(t *testing.T) {
	t.Parallel()

	m := NewEnvironment()
	m = m.SetSize(80, 30)

	m = m.SetDevices([]DeviceStatus{
		{
			Name:   "hallway",
			Online: true,
			Sensors: &model.SensorData{
				Illuminance: []model.IlluminanceReading{
					{ID: 0, Lux: ptr(300.0)},
				},
			},
		},
	})

	view := m.View()
	if !strings.Contains(view, "300 lux") {
		t.Error("expected view to contain raw lux value")
	}
}

func TestEnvironmentModel_View_WithBattery(t *testing.T) {
	t.Parallel()

	m := NewEnvironment()
	m = m.SetSize(80, 30)
	m = m.SetDevices([]DeviceStatus{
		{
			Name:   "sensor-a",
			Online: true,
			Sensors: &model.SensorData{
				DevicePower: []model.DevicePowerReading{
					{ID: 0, Battery: model.DevicePowerBatteryStatus{V: 3.2, Percent: 85}},
				},
			},
		},
	})

	view := m.View()
	if !strings.Contains(view, "Battery") {
		t.Error("expected view to contain 'Battery' section")
	}
	if !strings.Contains(view, "85%") {
		t.Error("expected view to contain battery percentage")
	}
}

func TestEnvironmentModel_View_WithBatteryExtPower(t *testing.T) {
	t.Parallel()

	m := NewEnvironment()
	m = m.SetSize(80, 30)
	m = m.SetDevices([]DeviceStatus{
		{
			Name:   "plugged-device",
			Online: true,
			Sensors: &model.SensorData{
				DevicePower: []model.DevicePowerReading{
					{
						ID:       0,
						Battery:  model.DevicePowerBatteryStatus{V: 4.0, Percent: 100},
						External: model.DevicePowerExternalStatus{Present: true},
					},
				},
			},
		},
	})

	view := m.View()
	if !strings.Contains(view, "[ext]") {
		t.Error("expected view to contain '[ext]' for external power")
	}
}

func TestEnvironmentModel_View_WithVoltmeter(t *testing.T) {
	t.Parallel()

	m := NewEnvironment()
	m = m.SetSize(80, 30)
	m = m.SetDevices([]DeviceStatus{
		{
			Name:   "voltage-meter",
			Online: true,
			Sensors: &model.SensorData{
				Voltmeter: []model.VoltmeterReading{
					{ID: 0, Voltage: ptr(12.45)},
				},
			},
		},
	})

	view := m.View()
	if !strings.Contains(view, "Voltmeter") {
		t.Error("expected view to contain 'Voltmeter' section")
	}
	if !strings.Contains(view, "12.45V") {
		t.Error("expected view to contain voltage value")
	}
}

func TestEnvironmentModel_View_WithBTHome(t *testing.T) {
	t.Parallel()

	m := NewEnvironment()
	m = m.SetSize(80, 30)
	m = m.SetDevices([]DeviceStatus{
		{
			Name:   "bt-gateway",
			Online: true,
			Sensors: &model.SensorData{
				BTHome: []model.BTHomeSensorReading{
					{ID: 200, Value: 21.5},
				},
			},
		},
	})

	view := m.View()
	if !strings.Contains(view, "BTHome") {
		t.Error("expected view to contain 'BTHome' section")
	}
	if !strings.Contains(view, "21.5") {
		t.Error("expected view to contain BTHome value")
	}
}

func TestEnvironmentModel_View_FloodSensorOK(t *testing.T) {
	t.Parallel()

	m := NewEnvironment()
	m = m.SetSize(80, 30)
	m = m.SetDevices([]DeviceStatus{
		{
			Name:   "basement",
			Online: true,
			Sensors: &model.SensorData{
				Flood: []model.AlarmSensorReading{
					{ID: 0, Alarm: false, Mute: false},
				},
			},
		},
	})

	view := m.View()
	if !strings.Contains(view, "Safety") {
		t.Error("expected view to contain 'Safety' section")
	}
	if !strings.Contains(view, "Flood") {
		t.Error("expected view to contain 'Flood' label")
	}
	if !strings.Contains(view, "OK") {
		t.Error("expected view to contain 'OK' status")
	}
}

func TestEnvironmentModel_View_FloodSensorAlarm(t *testing.T) {
	t.Parallel()

	m := NewEnvironment()
	m = m.SetSize(80, 30)
	m = m.SetDevices([]DeviceStatus{
		{
			Name:   "basement",
			Online: true,
			Sensors: &model.SensorData{
				Flood: []model.AlarmSensorReading{
					{ID: 0, Alarm: true, Mute: false},
				},
			},
		},
	})

	view := m.View()
	if !strings.Contains(view, "ALARM") {
		t.Error("expected view to contain 'ALARM' for active flood alarm")
	}
}

func TestEnvironmentModel_View_FloodSensorMuted(t *testing.T) {
	t.Parallel()

	m := NewEnvironment()
	m = m.SetSize(80, 30)
	m = m.SetDevices([]DeviceStatus{
		{
			Name:   "basement",
			Online: true,
			Sensors: &model.SensorData{
				Flood: []model.AlarmSensorReading{
					{ID: 0, Alarm: false, Mute: true},
				},
			},
		},
	})

	view := m.View()
	if !strings.Contains(view, "MUTED") {
		t.Error("expected view to contain 'MUTED' for muted flood sensor")
	}
}

func TestEnvironmentModel_View_SmokeSensor(t *testing.T) {
	t.Parallel()

	m := NewEnvironment()
	m = m.SetSize(80, 30)
	m = m.SetDevices([]DeviceStatus{
		{
			Name:   "living-room",
			Online: true,
			Sensors: &model.SensorData{
				Smoke: []model.AlarmSensorReading{
					{ID: 0, Alarm: true, Mute: false},
				},
			},
		},
	})

	view := m.View()
	if !strings.Contains(view, "Smoke") {
		t.Error("expected view to contain 'Smoke' label")
	}
	if !strings.Contains(view, "ALARM") {
		t.Error("expected view to contain 'ALARM' for active smoke alarm")
	}
}

func TestEnvironmentModel_View_MultiSensorDevice(t *testing.T) {
	t.Parallel()

	m := NewEnvironment()
	m = m.SetSize(80, 40)
	m = m.SetDevices([]DeviceStatus{
		{
			Name:   "multi-sensor",
			Online: true,
			Sensors: &model.SensorData{
				Temperature: []model.TemperatureReading{
					{ID: 0, TC: ptr(20.0)},
					{ID: 1, TC: ptr(22.0)},
				},
				Humidity: []model.HumidityReading{
					{ID: 0, RH: ptr(45.0)},
				},
			},
		},
	})

	view := m.View()
	// Should show BOTH temperature readings
	if !strings.Contains(view, "20.0°C") {
		t.Error("expected view to contain first temperature")
	}
	if !strings.Contains(view, "22.0°C") {
		t.Error("expected view to contain second temperature")
	}
}

func TestEnvironmentModel_View_SafetyAlwaysShown(t *testing.T) {
	t.Parallel()

	// Even with only temperature data, Safety section should be visible
	m := NewEnvironment()
	m = m.SetSize(80, 30)
	m = m.SetDevices([]DeviceStatus{
		{
			Name:   "temp-only",
			Online: true,
			Sensors: &model.SensorData{
				Temperature: []model.TemperatureReading{
					{ID: 0, TC: ptr(25.0)},
				},
			},
		},
	})

	view := m.View()
	if !strings.Contains(view, "Safety") {
		t.Error("Safety section should always be visible")
	}
	if !strings.Contains(view, "No safety sensors configured") {
		t.Error("expected 'No safety sensors configured' when no safety sensors")
	}
}

func TestTempStyle(t *testing.T) {
	t.Parallel()

	m := NewEnvironment()

	// Test that different temperatures get different styles
	cold := m.tempStyle(10.0)
	ok := m.tempStyle(20.0)
	warm := m.tempStyle(30.0)
	hot := m.tempStyle(40.0)

	// Verify styles are different by checking they render differently
	// (they have different foreground colors)
	coldStr := cold.Render("test")
	okStr := ok.Render("test")
	warmStr := warm.Render("test")
	hotStr := hot.Render("test")

	if coldStr == okStr {
		t.Error("cold and ok styles should be different")
	}
	if okStr == warmStr {
		t.Error("ok and warm styles should be different")
	}
	if warmStr == hotStr {
		t.Error("warm and hot styles should be different")
	}
}

func TestHumidStyle(t *testing.T) {
	t.Parallel()

	m := NewEnvironment()

	// Bad low
	bad1 := m.humidStyle(15.0)
	// Caution low
	caution1 := m.humidStyle(25.0)
	// OK
	ok := m.humidStyle(45.0)
	// Caution high
	caution2 := m.humidStyle(70.0)
	// Bad high
	bad2 := m.humidStyle(85.0)

	bad1Str := bad1.Render("test")
	caution1Str := caution1.Render("test")
	okStr := ok.Render("test")
	caution2Str := caution2.Render("test")
	bad2Str := bad2.Render("test")

	// Bad ranges should render the same
	if bad1Str != bad2Str {
		t.Error("both bad humidity ranges should use same style")
	}
	// Caution ranges should render the same
	if caution1Str != caution2Str {
		t.Error("both caution humidity ranges should use same style")
	}
	// OK should differ from caution
	if okStr == caution1Str {
		t.Error("ok and caution humidity styles should be different")
	}
}

func TestBatteryStyle(t *testing.T) {
	t.Parallel()

	m := NewEnvironment()

	crit := m.batteryStyle(10)
	low := m.batteryStyle(30)
	good := m.batteryStyle(60)

	critStr := crit.Render("test")
	lowStr := low.Render("test")
	goodStr := good.Render("test")

	if critStr == lowStr {
		t.Error("critical and low battery styles should be different")
	}
	if lowStr == goodStr {
		t.Error("low and good battery styles should be different")
	}
}

func TestFormatBTHomeValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
		want  string
	}{
		{name: "float", value: 21.5, want: "21.5"},
		{name: "bool true", value: true, want: "true"},
		{name: "bool false", value: false, want: "false"},
		{name: "string", value: "open", want: "open"},
		{name: "nil", value: nil, want: "<nil>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatBTHomeValue(tt.value)
			if got != tt.want {
				t.Errorf("formatBTHomeValue(%v) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestEnvironmentModel_BatterySortOrder(t *testing.T) {
	t.Parallel()

	m := NewEnvironment()
	m = m.SetSize(80, 30)
	m = m.SetDevices([]DeviceStatus{
		{
			Name:   "high-batt",
			Online: true,
			Sensors: &model.SensorData{
				DevicePower: []model.DevicePowerReading{
					{ID: 0, Battery: model.DevicePowerBatteryStatus{V: 3.7, Percent: 90}},
				},
			},
		},
		{
			Name:   "low-batt",
			Online: true,
			Sensors: &model.SensorData{
				DevicePower: []model.DevicePowerReading{
					{ID: 0, Battery: model.DevicePowerBatteryStatus{V: 2.5, Percent: 10}},
				},
			},
		},
		{
			Name:   "med-batt",
			Online: true,
			Sensors: &model.SensorData{
				DevicePower: []model.DevicePowerReading{
					{ID: 0, Battery: model.DevicePowerBatteryStatus{V: 3.0, Percent: 35}},
				},
			},
		},
	})

	// Collect batteries and verify sort order (lowest first)
	batts := m.collectBatteries()
	if len(batts) != 3 {
		t.Fatalf("expected 3 batteries, got %d", len(batts))
	}
	if batts[0].DeviceName != "low-batt" {
		t.Errorf("expected lowest battery first, got %q", batts[0].DeviceName)
	}
	if batts[1].DeviceName != "med-batt" {
		t.Errorf("expected medium battery second, got %q", batts[1].DeviceName)
	}
	if batts[2].DeviceName != "high-batt" {
		t.Errorf("expected highest battery last, got %q", batts[2].DeviceName)
	}
}

func TestEnvironmentModel_AlarmSortOrder(t *testing.T) {
	t.Parallel()

	m := NewEnvironment()
	m = m.SetSize(80, 30)
	m = m.SetDevices([]DeviceStatus{
		{
			Name:   "ok-device",
			Online: true,
			Sensors: &model.SensorData{
				Flood: []model.AlarmSensorReading{
					{ID: 0, Alarm: false},
				},
			},
		},
		{
			Name:   "alarm-device",
			Online: true,
			Sensors: &model.SensorData{
				Flood: []model.AlarmSensorReading{
					{ID: 0, Alarm: true},
				},
			},
		},
	})

	// Collect flood sensors and verify alarms sort first
	floods := m.collectFloodSensors()
	if len(floods) != 2 {
		t.Fatalf("expected 2 flood sensors, got %d", len(floods))
	}
	if floods[0].DeviceName != "alarm-device" {
		t.Errorf("expected alarm device first, got %q", floods[0].DeviceName)
	}
	if floods[1].DeviceName != "ok-device" {
		t.Errorf("expected ok device second, got %q", floods[1].DeviceName)
	}
}

func TestEnvironmentModel_CountDisplayLines(t *testing.T) {
	t.Parallel()

	t.Run("empty devices", func(t *testing.T) {
		t.Parallel()
		m := NewEnvironment()
		// Even with no devices, Safety header + "No safety sensors" = 2 lines
		count := m.countDisplayLines()
		if count != 2 {
			t.Errorf("expected 2 lines for empty (Safety header + no sensors msg), got %d", count)
		}
	})

	t.Run("with temperature and flood", func(t *testing.T) {
		t.Parallel()
		m := NewEnvironment()
		m = m.SetDevices([]DeviceStatus{
			{
				Name:   "dev",
				Online: true,
				Sensors: &model.SensorData{
					Temperature: []model.TemperatureReading{
						{ID: 0, TC: ptr(20.0)},
						{ID: 1, TC: ptr(22.0)},
					},
					Flood: []model.AlarmSensorReading{
						{ID: 0, Alarm: false},
					},
				},
			},
		})
		// Temperature: 1 header + 2 entries = 3
		// Safety: 1 header + 1 flood = 2
		// Total = 5
		count := m.countDisplayLines()
		if count != 5 {
			t.Errorf("expected 5 display lines, got %d", count)
		}
	})
}

func TestEnvironmentModel_SetPanelIndex(t *testing.T) {
	t.Parallel()

	m := NewEnvironment()
	m = m.SetPanelIndex(2)

	if m.panelIdx != 2 {
		t.Errorf("expected panel index 2, got %d", m.panelIdx)
	}
}

func TestEnvironmentModel_View_FullContent(t *testing.T) {
	t.Parallel()

	// Test a full environment panel with all sensor types
	m := NewEnvironment()
	m = m.SetSize(100, 50)
	m = m.SetDevices([]DeviceStatus{
		{
			Name:   "all-sensors",
			Online: true,
			Sensors: &model.SensorData{
				Temperature: []model.TemperatureReading{
					{ID: 0, TC: ptr(22.5)},
				},
				Humidity: []model.HumidityReading{
					{ID: 0, RH: ptr(55.0)},
				},
				Illuminance: []model.IlluminanceReading{
					{ID: 0, Lux: ptr(400.0)},
				},
				DevicePower: []model.DevicePowerReading{
					{ID: 0, Battery: model.DevicePowerBatteryStatus{V: 3.5, Percent: 75}},
				},
				Voltmeter: []model.VoltmeterReading{
					{ID: 0, Voltage: ptr(12.0)},
				},
				Flood: []model.AlarmSensorReading{
					{ID: 0, Alarm: false, Mute: false},
				},
				Smoke: []model.AlarmSensorReading{
					{ID: 0, Alarm: false, Mute: false},
				},
				BTHome: []model.BTHomeSensorReading{
					{ID: 200, Value: 42.0},
				},
			},
		},
	})

	view := m.View()

	// Verify all sections are present
	sections := []string{"Temperature", "Humidity", "Illuminance", "Battery", "Voltmeter", "BTHome", "Safety"}
	for _, sec := range sections {
		if !strings.Contains(view, sec) {
			t.Errorf("expected view to contain '%s' section", sec)
		}
	}

	// Should NOT contain "No safety sensors configured" since we have flood/smoke
	if strings.Contains(view, "No safety sensors configured") {
		t.Error("should not show 'No safety sensors' when flood/smoke sensors exist")
	}
}
