// Package term - sensor.go provides generic sensor display functions.
package term

import (
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/output/table"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// SensorValueFormatter formats a sensor's value(s) for display.
// The highlight parameter controls whether to apply theme styling.
type SensorValueFormatter[T model.Sensor] func(s T, highlight bool) string

// SensorValueChecker returns true if the sensor has a valid value.
type SensorValueChecker[T model.Sensor] func(s T) bool

// SensorOpts configures generic sensor display functions.
type SensorOpts[T model.Sensor] struct {
	Title      string                  // e.g., "Temperature Sensors"
	SingleName string                  // e.g., "Temperature Sensor"
	NoValueMsg string                  // e.g., "No temperature reading available."
	HasValue   SensorValueChecker[T]   // Checks if sensor has valid value
	Format     SensorValueFormatter[T] // Formats sensor value
}

// DisplaySensorList displays a list of sensors using generic options.
func DisplaySensorList[T model.Sensor](ios *iostreams.IOStreams, sensors []T, opts SensorOpts[T]) {
	builder := table.NewBuilder("ID", "Reading")
	for _, s := range sensors {
		if formatted := opts.Format(s, false); formatted != "" {
			builder.AddRow(fmt.Sprintf("%d", s.GetID()), formatted)
		}
	}
	ios.Println(theme.Bold().Render(opts.Title + ":"))
	ios.Println()
	tbl := builder.WithModeStyle(ios).Build()
	if err := tbl.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print sensor list table", err)
	}
}

// DisplaySensorStatus displays a single sensor status using generic options.
func DisplaySensorStatus[T model.Sensor](ios *iostreams.IOStreams, status T, id int, opts SensorOpts[T]) {
	ios.Println(theme.Bold().Render(fmt.Sprintf("%s %d:", opts.SingleName, id)))
	ios.Println()
	if !opts.HasValue(status) {
		ios.Warning("%s", opts.NoValueMsg)
		return
	}
	builder := table.NewBuilder("Metric", "Value")
	builder.AddRow("Reading", opts.Format(status, true))
	tbl := builder.WithModeStyle(ios).Build()
	if err := tbl.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print sensor status table", err)
	}
	displaySensorErrors(ios, status.GetErrors())
}

// displaySensorErrors prints sensor errors if present.
func displaySensorErrors(ios *iostreams.IOStreams, errors []string) {
	if len(errors) > 0 {
		ios.Println()
		ios.Warning("Errors: %v", errors)
	}
}

// displayAllSection displays a section in the all-sensors view.
// Returns true if any items were displayed.
func displayAllSection[T any](ios *iostreams.IOStreams, items []T, title string, format func(T) string) bool {
	if len(items) == 0 {
		return false
	}
	ios.Println("  " + theme.Highlight().Render(title+":"))
	for _, item := range items {
		if s := format(item); s != "" {
			ios.Println(s)
		}
	}
	ios.Println()
	return true
}

// Pre-configured sensor options for each type.

// TemperatureOpts provides configuration for temperature sensor displays.
var TemperatureOpts = SensorOpts[model.TemperatureReading]{
	Title:      "Temperature Sensors",
	SingleName: "Temperature Sensor",
	NoValueMsg: "No temperature reading available.",
	HasValue:   func(s model.TemperatureReading) bool { return s.TC != nil },
	Format: func(s model.TemperatureReading, highlight bool) string {
		if s.TC == nil {
			return ""
		}
		value := fmt.Sprintf("%.1f°C", *s.TC)
		if highlight {
			value = theme.Highlight().Render(value)
		}
		result := "Temperature: " + value
		if s.TF != nil {
			extra := fmt.Sprintf("%.1f°F", *s.TF)
			if highlight {
				extra = theme.Dim().Render(extra)
			}
			result += " (" + extra + ")"
		}
		return result
	},
}

// HumidityOpts provides configuration for humidity sensor displays.
var HumidityOpts = SensorOpts[model.HumidityReading]{
	Title:      "Humidity Sensors",
	SingleName: "Humidity Sensor",
	NoValueMsg: "No humidity reading available.",
	HasValue:   func(s model.HumidityReading) bool { return s.RH != nil },
	Format: func(s model.HumidityReading, highlight bool) string {
		if s.RH == nil {
			return ""
		}
		value := fmt.Sprintf("%.1f%%", *s.RH)
		if highlight {
			value = theme.Highlight().Render(value)
		}
		return "Humidity: " + value
	},
}

// IlluminanceOpts provides configuration for illuminance sensor displays.
var IlluminanceOpts = SensorOpts[model.IlluminanceReading]{
	Title:      "Illuminance Sensors",
	SingleName: "Illuminance Sensor",
	NoValueMsg: "No illuminance reading available.",
	HasValue:   func(s model.IlluminanceReading) bool { return s.Lux != nil },
	Format: func(s model.IlluminanceReading, highlight bool) string {
		if s.Lux == nil {
			return ""
		}
		value := fmt.Sprintf("%.0f lux", *s.Lux)
		if highlight {
			level := output.GetLightLevel(*s.Lux)
			return "Light Level: " + theme.Highlight().Render(value) + " (" + theme.Dim().Render(level) + ")"
		}
		return "Light Level: " + value
	},
}

// VoltmeterOpts provides configuration for voltmeter sensor displays.
var VoltmeterOpts = SensorOpts[model.VoltmeterReading]{
	Title:      "Voltmeter Sensors",
	SingleName: "Voltmeter Sensor",
	NoValueMsg: "No voltage reading available.",
	HasValue:   func(s model.VoltmeterReading) bool { return s.Voltage != nil },
	Format: func(s model.VoltmeterReading, highlight bool) string {
		if s.Voltage == nil {
			return ""
		}
		value := fmt.Sprintf("%.3f V", *s.Voltage)
		if highlight {
			value = theme.Highlight().Render(value)
		}
		return "Voltage: " + value
	},
}

// DevicePowerOpts provides configuration for device power sensor displays.
var DevicePowerOpts = SensorOpts[model.DevicePowerReading]{
	Title:      "Device Power Sensors",
	SingleName: "Device Power Sensor",
	NoValueMsg: "", // Always has value
	HasValue:   func(s model.DevicePowerReading) bool { return true },
	Format: func(s model.DevicePowerReading, highlight bool) string {
		percentStr := fmt.Sprintf("%d%%", s.Battery.Percent)
		voltStr := fmt.Sprintf("%.2fV", s.Battery.V)
		extStr := output.RenderYesNo(s.External.Present, output.CaseTitle, theme.FalseDim)

		if highlight {
			switch {
			case s.Battery.Percent <= 20:
				percentStr = theme.StatusError().Render(percentStr)
			case s.Battery.Percent <= 50:
				percentStr = theme.StatusWarn().Render(percentStr)
			default:
				percentStr = theme.StatusOK().Render(percentStr)
			}
			voltStr = theme.Dim().Render(voltStr)
		}

		return fmt.Sprintf("Battery: %s (%s)\n    External Power: %s", percentStr, voltStr, extStr)
	},
}

// Partial application: create concrete functions from generic + opts.

func sensorListFunc[T model.Sensor](opts SensorOpts[T]) func(*iostreams.IOStreams, []T) {
	return func(ios *iostreams.IOStreams, sensors []T) {
		DisplaySensorList(ios, sensors, opts)
	}
}

func sensorStatusFunc[T model.Sensor](opts SensorOpts[T]) func(*iostreams.IOStreams, T, int) {
	return func(ios *iostreams.IOStreams, status T, id int) {
		DisplaySensorStatus(ios, status, id, opts)
	}
}

// Concrete sensor display functions.
var (
	DisplayTemperatureList   = sensorListFunc(TemperatureOpts)
	DisplayTemperatureStatus = sensorStatusFunc(TemperatureOpts)
	DisplayHumidityList      = sensorListFunc(HumidityOpts)
	DisplayHumidityStatus    = sensorStatusFunc(HumidityOpts)
	DisplayIlluminanceList   = sensorListFunc(IlluminanceOpts)
	DisplayIlluminanceStatus = sensorStatusFunc(IlluminanceOpts)
	DisplayVoltmeterList     = sensorListFunc(VoltmeterOpts)
	DisplayVoltmeterStatus   = sensorStatusFunc(VoltmeterOpts)
	DisplayDevicePowerList   = sensorListFunc(DevicePowerOpts)
	DisplayDevicePowerStatus = sensorStatusFunc(DevicePowerOpts)
)

// Alarm sensor display functions (different pattern - alarm/mute state instead of value).

// DisplayAlarmSensors displays alarm-type sensors (flood, smoke) with a consistent format.
// Returns true if any sensors were displayed.
func DisplayAlarmSensors(ios *iostreams.IOStreams, sensors []model.AlarmSensorReading, sensorType, alarmMsg string) bool {
	lines := output.FormatAlarmSensors(sensors, sensorType, alarmMsg,
		theme.StatusOK().Render, theme.StatusError().Render, theme.Dim().Render)
	if lines == nil {
		return false
	}
	for _, line := range lines {
		ios.Println(line)
	}
	return true
}

// DisplayAlarmSensorList displays a list of alarm-type sensors using Go generics.
func DisplayAlarmSensorList[T model.AlarmSensor](ios *iostreams.IOStreams, sensors []T, sensorName, alarmMsg string) {
	builder := table.NewBuilder("ID", "Status", "Muted")
	for _, s := range sensors {
		status := output.RenderAlarmState(s.IsAlarm(), alarmMsg)
		muted := output.RenderMuteState(s.IsMuted())
		builder.AddRow(fmt.Sprintf("%d", s.GetID()), status, muted)
	}
	ios.Println(theme.Bold().Render(sensorName + " Sensors:"))
	ios.Println()
	tbl := builder.WithModeStyle(ios).Build()
	if err := tbl.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print alarm sensor list table", err)
	}
}

// DisplayAlarmSensorStatus displays a single alarm-type sensor status using Go generics.
func DisplayAlarmSensorStatus[T model.AlarmSensor](ios *iostreams.IOStreams, status T, id int, sensorName, alarmMsg string) {
	ios.Println(theme.Bold().Render(fmt.Sprintf("%s Sensor %d:", sensorName, id)))
	ios.Println()
	builder := table.NewBuilder("Metric", "Value")
	builder.AddRow("Status", output.RenderAlarmState(status.IsAlarm(), alarmMsg))
	builder.AddRow("Muted", output.RenderMuteState(status.IsMuted()))
	tbl := builder.WithModeStyle(ios).Build()
	if err := tbl.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print alarm sensor status table", err)
	}
	displaySensorErrors(ios, status.GetErrors())
}

// DisplayAllSensorData displays aggregated sensor data from a device.
func DisplayAllSensorData(ios *iostreams.IOStreams, data *model.SensorData, device string) {
	ios.Println(theme.Bold().Render(fmt.Sprintf("Sensor Readings for %s:", device)))
	ios.Println()

	hasData := displayAllSection(ios, data.Temperature, "Temperature", func(t model.TemperatureReading) string {
		if t.TC == nil {
			return ""
		}
		s := fmt.Sprintf("    Sensor %d: %.1f°C", t.ID, *t.TC)
		if t.TF != nil {
			s += fmt.Sprintf(" (%.1f°F)", *t.TF)
		}
		return s
	})

	hasData = displayAllSection(ios, data.Humidity, "Humidity", func(h model.HumidityReading) string {
		if h.RH == nil {
			return ""
		}
		return fmt.Sprintf("    Sensor %d: %.1f%%", h.ID, *h.RH)
	}) || hasData

	hasData = displayAllAlarmSection(ios, data.Flood, "Flood Detection", "Flood", "WATER DETECTED!") || hasData
	hasData = displayAllAlarmSection(ios, data.Smoke, "Smoke Detection", "Smoke", "SMOKE DETECTED!") || hasData

	hasData = displayAllSection(ios, data.Illuminance, "Illuminance", func(l model.IlluminanceReading) string {
		if l.Lux == nil {
			return ""
		}
		return fmt.Sprintf("    Sensor %d: %.0f lux", l.ID, *l.Lux)
	}) || hasData

	hasData = displayAllSection(ios, data.Voltmeter, "Voltage", func(v model.VoltmeterReading) string {
		if v.Voltage == nil {
			return ""
		}
		return fmt.Sprintf("    Sensor %d: %.2f V", v.ID, *v.Voltage)
	}) || hasData

	if !hasData {
		ios.Info("No sensors found on this device.")
	}
}

// displayAllAlarmSection displays alarm sensors in the all-sensors view.
func displayAllAlarmSection(ios *iostreams.IOStreams, sensors []model.AlarmSensorReading, title, sensorType, alarmMsg string) bool {
	if len(sensors) == 0 {
		return false
	}
	ios.Println("  " + theme.Highlight().Render(title+":"))
	DisplayAlarmSensors(ios, sensors, sensorType, alarmMsg)
	ios.Println()
	return true
}
