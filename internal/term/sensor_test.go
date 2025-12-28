package term

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestDisplaySensorList_Temperature(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	tc := 22.5
	tf := 72.5
	sensors := []model.TemperatureReading{
		{ID: 0, TC: &tc, TF: &tf, Errors: nil},
		{ID: 1, TC: &tc, TF: nil, Errors: nil},
	}

	DisplayTemperatureList(ios, sensors)

	output := out.String()
	if !strings.Contains(output, "Temperature") {
		t.Error("output should contain 'Temperature'")
	}
	if !strings.Contains(output, "22.5") {
		t.Error("output should contain temperature value")
	}
}

func TestDisplaySensorStatus_Temperature(t *testing.T) {
	t.Parallel()

	t.Run("with valid reading", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		tc := 25.0
		tf := 77.0
		status := model.TemperatureReading{ID: 0, TC: &tc, TF: &tf, Errors: nil}

		DisplayTemperatureStatus(ios, status, 0)

		output := out.String()
		if !strings.Contains(output, "25.0") {
			t.Errorf("output should contain temperature value, got %q", output)
		}
	})

	t.Run("without reading", func(t *testing.T) {
		t.Parallel()

		ios, out, errOut := testIOStreams()
		status := model.TemperatureReading{ID: 0, TC: nil, TF: nil, Errors: nil}

		DisplayTemperatureStatus(ios, status, 0)

		allOutput := out.String() + errOut.String()
		if !strings.Contains(allOutput, "No temperature reading") {
			t.Errorf("output should contain warning message, got %q", allOutput)
		}
	})
}

func TestDisplaySensorList_Humidity(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	rh := 55.5
	sensors := []model.HumidityReading{
		{ID: 0, RH: &rh, Errors: nil},
	}

	DisplayHumidityList(ios, sensors)

	output := out.String()
	if !strings.Contains(output, "Humidity") {
		t.Error("output should contain 'Humidity'")
	}
	if !strings.Contains(output, "55.5") {
		t.Error("output should contain humidity value")
	}
}

func TestDisplaySensorStatus_Humidity(t *testing.T) {
	t.Parallel()

	t.Run("with valid reading", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		rh := 60.0
		status := model.HumidityReading{ID: 0, RH: &rh, Errors: nil}

		DisplayHumidityStatus(ios, status, 0)

		output := out.String()
		if !strings.Contains(output, "60.0") {
			t.Errorf("output should contain humidity value, got %q", output)
		}
	})

	t.Run("without reading", func(t *testing.T) {
		t.Parallel()

		ios, out, errOut := testIOStreams()
		status := model.HumidityReading{ID: 0, RH: nil, Errors: nil}

		DisplayHumidityStatus(ios, status, 0)

		allOutput := out.String() + errOut.String()
		if !strings.Contains(allOutput, "No humidity reading") {
			t.Errorf("output should contain warning message, got %q", allOutput)
		}
	})
}

func TestDisplaySensorList_Illuminance(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	lux := 500.0
	sensors := []model.IlluminanceReading{
		{ID: 0, Lux: &lux, Errors: nil},
	}

	DisplayIlluminanceList(ios, sensors)

	output := out.String()
	if !strings.Contains(output, "Illuminance") {
		t.Error("output should contain 'Illuminance'")
	}
	if !strings.Contains(output, "500") {
		t.Error("output should contain lux value")
	}
}

func TestDisplaySensorStatus_Illuminance(t *testing.T) {
	t.Parallel()

	t.Run("with valid reading", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		lux := 750.0
		status := model.IlluminanceReading{ID: 0, Lux: &lux, Errors: nil}

		DisplayIlluminanceStatus(ios, status, 0)

		output := out.String()
		if !strings.Contains(output, "750") {
			t.Errorf("output should contain lux value, got %q", output)
		}
	})

	t.Run("without reading", func(t *testing.T) {
		t.Parallel()

		ios, out, errOut := testIOStreams()
		status := model.IlluminanceReading{ID: 0, Lux: nil, Errors: nil}

		DisplayIlluminanceStatus(ios, status, 0)

		allOutput := out.String() + errOut.String()
		if !strings.Contains(allOutput, "No illuminance reading") {
			t.Errorf("output should contain warning message, got %q", allOutput)
		}
	})
}

func TestDisplaySensorList_Voltmeter(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	voltage := 12.5
	sensors := []model.VoltmeterReading{
		{ID: 0, Voltage: &voltage, Errors: nil},
	}

	DisplayVoltmeterList(ios, sensors)

	output := out.String()
	if !strings.Contains(output, "Voltmeter") {
		t.Error("output should contain 'Voltmeter'")
	}
	if !strings.Contains(output, "12.5") {
		t.Error("output should contain voltage value")
	}
}

func TestDisplaySensorStatus_Voltmeter(t *testing.T) {
	t.Parallel()

	t.Run("with valid reading", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		voltage := 24.0
		status := model.VoltmeterReading{ID: 0, Voltage: &voltage, Errors: nil}

		DisplayVoltmeterStatus(ios, status, 0)

		output := out.String()
		if !strings.Contains(output, "24.0") {
			t.Errorf("output should contain voltage value, got %q", output)
		}
	})

	t.Run("without reading", func(t *testing.T) {
		t.Parallel()

		ios, out, errOut := testIOStreams()
		status := model.VoltmeterReading{ID: 0, Voltage: nil, Errors: nil}

		DisplayVoltmeterStatus(ios, status, 0)

		allOutput := out.String() + errOut.String()
		if !strings.Contains(allOutput, "No voltage reading") {
			t.Errorf("output should contain warning message, got %q", allOutput)
		}
	})
}

func TestDisplaySensorList_DevicePower(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	sensors := []model.DevicePowerReading{
		{
			ID:       0,
			Battery:  model.DevicePowerBatteryStatus{Percent: 75, V: 3.7},
			External: model.DevicePowerExternalStatus{Present: false},
			Errors:   nil,
		},
	}

	DisplayDevicePowerList(ios, sensors)

	output := out.String()
	if !strings.Contains(output, "75%") {
		t.Error("output should contain battery percentage")
	}
}

func TestDisplaySensorStatus_DevicePower(t *testing.T) {
	t.Parallel()

	t.Run("low battery", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		status := model.DevicePowerReading{
			ID:       0,
			Battery:  model.DevicePowerBatteryStatus{Percent: 15, V: 3.2},
			External: model.DevicePowerExternalStatus{Present: false},
			Errors:   nil,
		}

		DisplayDevicePowerStatus(ios, status, 0)

		output := out.String()
		if !strings.Contains(output, "15%") {
			t.Errorf("output should contain battery percentage, got %q", output)
		}
	})

	t.Run("medium battery", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		status := model.DevicePowerReading{
			ID:       0,
			Battery:  model.DevicePowerBatteryStatus{Percent: 45, V: 3.5},
			External: model.DevicePowerExternalStatus{Present: false},
			Errors:   nil,
		}

		DisplayDevicePowerStatus(ios, status, 0)

		output := out.String()
		if !strings.Contains(output, "45%") {
			t.Errorf("output should contain battery percentage, got %q", output)
		}
	})

	t.Run("full battery with external power", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		status := model.DevicePowerReading{
			ID:       0,
			Battery:  model.DevicePowerBatteryStatus{Percent: 100, V: 4.2},
			External: model.DevicePowerExternalStatus{Present: true},
			Errors:   nil,
		}

		DisplayDevicePowerStatus(ios, status, 0)

		output := out.String()
		if !strings.Contains(output, "100%") {
			t.Errorf("output should contain battery percentage, got %q", output)
		}
	})
}

func TestDisplayAlarmSensors(t *testing.T) {
	t.Parallel()

	t.Run("with sensors", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		sensors := []model.AlarmSensorReading{
			{ID: 0, Alarm: false, Mute: false},
			{ID: 1, Alarm: true, Mute: false},
		}

		result := DisplayAlarmSensors(ios, sensors, "Flood", "WATER DETECTED!")

		if !result {
			t.Error("DisplayAlarmSensors should return true when sensors exist")
		}
		output := out.String()
		if output == "" {
			t.Error("output should not be empty")
		}
	})

	t.Run("empty sensors", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()

		result := DisplayAlarmSensors(ios, []model.AlarmSensorReading{}, "Flood", "WATER DETECTED!")

		if result {
			t.Error("DisplayAlarmSensors should return false when no sensors")
		}
		output := out.String()
		if output != "" {
			t.Errorf("output should be empty, got %q", output)
		}
	})
}

func TestDisplayAlarmSensorList(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	sensors := []model.AlarmSensorReading{
		{ID: 0, Alarm: false, Mute: false},
		{ID: 1, Alarm: true, Mute: true},
	}

	DisplayAlarmSensorList(ios, sensors, "Flood", "WATER DETECTED!")

	output := out.String()
	if !strings.Contains(output, "Flood") {
		t.Error("output should contain sensor name")
	}
}

func TestDisplayAlarmSensorStatus(t *testing.T) {
	t.Parallel()

	t.Run("no alarm", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		status := model.AlarmSensorReading{ID: 0, Alarm: false, Mute: false}

		DisplayAlarmSensorStatus(ios, status, 0, "Flood", "WATER DETECTED!")

		output := out.String()
		if !strings.Contains(output, "Flood") {
			t.Error("output should contain sensor name")
		}
	})

	t.Run("with alarm and errors", func(t *testing.T) {
		t.Parallel()

		ios, out, errOut := testIOStreams()
		status := model.AlarmSensorReading{ID: 0, Alarm: true, Mute: false, Errors: []string{"sensor error"}}

		DisplayAlarmSensorStatus(ios, status, 0, "Smoke", "SMOKE DETECTED!")

		allOutput := out.String() + errOut.String()
		if !strings.Contains(allOutput, "Smoke") {
			t.Error("output should contain sensor name")
		}
	})
}

func TestDisplayAllSensorData(t *testing.T) {
	t.Parallel()

	t.Run("with all sensor types", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		tc := 22.5
		tf := 72.5
		rh := 55.0
		lux := 500.0
		voltage := 12.0
		data := &model.SensorData{
			Temperature: []model.TemperatureReading{
				{ID: 0, TC: &tc, TF: &tf},
			},
			Humidity: []model.HumidityReading{
				{ID: 0, RH: &rh},
			},
			Flood: []model.AlarmSensorReading{
				{ID: 0, Alarm: false, Mute: false},
			},
			Smoke: []model.AlarmSensorReading{
				{ID: 0, Alarm: false, Mute: false},
			},
			Illuminance: []model.IlluminanceReading{
				{ID: 0, Lux: &lux},
			},
			Voltmeter: []model.VoltmeterReading{
				{ID: 0, Voltage: &voltage},
			},
		}

		DisplayAllSensorData(ios, data, "test-device")

		output := out.String()
		if !strings.Contains(output, "test-device") {
			t.Error("output should contain device name")
		}
		if !strings.Contains(output, "22.5") {
			t.Error("output should contain temperature")
		}
		if !strings.Contains(output, "55.0") {
			t.Error("output should contain humidity")
		}
	})

	t.Run("empty sensor data", func(t *testing.T) {
		t.Parallel()

		ios, out, errOut := testIOStreams()
		data := &model.SensorData{}

		DisplayAllSensorData(ios, data, "empty-device")

		allOutput := out.String() + errOut.String()
		if !strings.Contains(allOutput, "No sensors found") {
			t.Errorf("output should contain 'No sensors found', got %q", allOutput)
		}
	})
}

func TestDisplaySensorErrors(t *testing.T) {
	t.Parallel()

	t.Run("with errors", func(t *testing.T) {
		t.Parallel()

		ios, out, errOut := testIOStreams()
		displaySensorErrors(ios, []string{"error1", "error2"})

		allOutput := out.String() + errOut.String()
		if !strings.Contains(allOutput, "Errors") {
			t.Errorf("output should contain 'Errors', got %q", allOutput)
		}
	})

	t.Run("no errors", func(t *testing.T) {
		t.Parallel()

		ios, out, errOut := testIOStreams()
		displaySensorErrors(ios, []string{})

		allOutput := out.String() + errOut.String()
		if allOutput != "" {
			t.Errorf("output should be empty when no errors, got %q", allOutput)
		}
	})
}

func TestDisplayAllSection(t *testing.T) {
	t.Parallel()

	t.Run("with items", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		items := []string{"item1", "item2"}

		result := displayAllSection(ios, items, "Test Section", func(s string) string {
			return "    " + s
		})

		if !result {
			t.Error("displayAllSection should return true when items exist")
		}
		output := out.String()
		if !strings.Contains(output, "Test Section") {
			t.Error("output should contain section title")
		}
	})

	t.Run("empty items", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()

		result := displayAllSection(ios, []string{}, "Empty Section", func(s string) string {
			return s
		})

		if result {
			t.Error("displayAllSection should return false when no items")
		}
		output := out.String()
		if output != "" {
			t.Errorf("output should be empty, got %q", output)
		}
	})

	t.Run("with empty formatted item", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		items := []string{"", "item"}

		result := displayAllSection(ios, items, "Test", func(s string) string {
			return s // Returns empty for first item
		})

		if !result {
			t.Error("displayAllSection should return true when items exist")
		}
		output := out.String()
		if !strings.Contains(output, "item") {
			t.Error("output should contain non-empty item")
		}
	})
}

func TestSensorOpts_Formatters(t *testing.T) {
	t.Parallel()

	t.Run("temperature format with highlight", func(t *testing.T) {
		t.Parallel()

		tc := 20.0
		tf := 68.0
		sensor := model.TemperatureReading{ID: 0, TC: &tc, TF: &tf}

		result := TemperatureOpts.Format(sensor, true)
		if !strings.Contains(result, "20.0") {
			t.Errorf("format should contain temperature, got %q", result)
		}
	})

	t.Run("temperature format without value", func(t *testing.T) {
		t.Parallel()

		sensor := model.TemperatureReading{ID: 0, TC: nil, TF: nil}

		result := TemperatureOpts.Format(sensor, false)
		if result != "" {
			t.Errorf("format should return empty string for nil value, got %q", result)
		}
	})

	t.Run("humidity format with highlight", func(t *testing.T) {
		t.Parallel()

		rh := 50.0
		sensor := model.HumidityReading{ID: 0, RH: &rh}

		result := HumidityOpts.Format(sensor, true)
		if !strings.Contains(result, "50.0") {
			t.Errorf("format should contain humidity, got %q", result)
		}
	})

	t.Run("illuminance format with highlight", func(t *testing.T) {
		t.Parallel()

		lux := 1000.0
		sensor := model.IlluminanceReading{ID: 0, Lux: &lux}

		result := IlluminanceOpts.Format(sensor, true)
		if !strings.Contains(result, "1000") {
			t.Errorf("format should contain lux, got %q", result)
		}
	})

	t.Run("voltmeter format with highlight", func(t *testing.T) {
		t.Parallel()

		voltage := 5.5
		sensor := model.VoltmeterReading{ID: 0, Voltage: &voltage}

		result := VoltmeterOpts.Format(sensor, true)
		if !strings.Contains(result, "5.5") {
			t.Errorf("format should contain voltage, got %q", result)
		}
	})

	t.Run("device power format with low battery", func(t *testing.T) {
		t.Parallel()

		sensor := model.DevicePowerReading{
			ID:       0,
			Battery:  model.DevicePowerBatteryStatus{Percent: 10, V: 3.0},
			External: model.DevicePowerExternalStatus{Present: false},
		}

		result := DevicePowerOpts.Format(sensor, true)
		if !strings.Contains(result, "10%") {
			t.Errorf("format should contain battery percentage, got %q", result)
		}
	})
}
