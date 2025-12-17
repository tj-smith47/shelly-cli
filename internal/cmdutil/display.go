// Package cmdutil provides display helpers that print directly to IOStreams.
// Display* functions wrap pure formatters from output/ with printing and semantic output.
package cmdutil

import (
	"encoding/json"

	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DisplayPowerMetrics outputs power, voltage, and current with units.
// Nil values are skipped.
func DisplayPowerMetrics(ios *iostreams.IOStreams, power, voltage, current *float64) {
	if power != nil {
		ios.Printf("  Power:   %.1f W\n", *power)
	}
	if voltage != nil {
		ios.Printf("  Voltage: %.1f V\n", *voltage)
	}
	if current != nil {
		ios.Printf("  Current: %.3f A\n", *current)
	}
}

// DisplayPowerMetricsWide outputs power metrics with wider alignment for cover status.
func DisplayPowerMetricsWide(ios *iostreams.IOStreams, power, voltage, current *float64) {
	if power != nil {
		ios.Printf("  Power:    %.1f W\n", *power)
	}
	if voltage != nil {
		ios.Printf("  Voltage:  %.1f V\n", *voltage)
	}
	if current != nil {
		ios.Printf("  Current:  %.3f A\n", *current)
	}
}

// DisplayDiscoveredDevices prints a table of discovered devices.
// Handles empty case with ios.NoResults and prints count at the end.
func DisplayDiscoveredDevices(ios *iostreams.IOStreams, devices []discovery.DiscoveredDevice) {
	table := output.FormatDiscoveredDevices(devices)
	if table == nil {
		ios.NoResults("devices")
		return
	}
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}
	ios.Count("device", len(devices))
}

// DisplayConfigTable prints a configuration map as formatted tables.
// Each top-level key becomes a titled section with a settings table.
func DisplayConfigTable(ios *iostreams.IOStreams, config any) error {
	configMap, ok := config.(map[string]any)
	if !ok {
		return output.PrintJSON(config)
	}

	for component, cfg := range configMap {
		ios.Title("%s", component)

		table := output.FormatConfigTable(cfg)
		if table == nil {
			// If it's not a map, print as JSON
			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				ios.DebugErr("marshaling config component", err)
			} else {
				ios.Printf("%s\n", data)
			}
			ios.Printf("\n")
			continue
		}

		if err := table.PrintTo(ios.Out); err != nil {
			ios.DebugErr("print config table", err)
		}
		ios.Printf("\n")
	}

	return nil
}

// DisplayAlarmSensors displays alarm-type sensors (flood, smoke) with a consistent format.
// Returns true if any sensors were displayed.
func DisplayAlarmSensors(ios *iostreams.IOStreams, sensors []output.AlarmSensorReading, sensorType, alarmMsg string) bool {
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
