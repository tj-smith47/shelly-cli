// Package cmdutil provides display helpers that print directly to IOStreams.
// Display* functions wrap pure formatters from output/ with printing and semantic output.
package cmdutil

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/export"
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
// Works with any type implementing model.AlarmSensor (smoke, flood, etc.).
func DisplayAlarmSensorList[T model.AlarmSensor](ios *iostreams.IOStreams, sensors []T, sensorName, alarmMsg string) {
	ios.Println(theme.Bold().Render(sensorName + " Sensors:"))
	ios.Println()
	for _, s := range sensors {
		status := output.RenderAlarmState(s.IsAlarm(), alarmMsg)
		muteStr := output.RenderMuteAnnotation(s.IsMuted())
		ios.Printf("  Sensor %d: %s%s\n", s.GetID(), status, muteStr)
	}
}

// DisplayAlarmSensorStatus displays a single alarm-type sensor status using Go generics.
// Works with any type implementing model.AlarmSensor (smoke, flood, etc.).
func DisplayAlarmSensorStatus[T model.AlarmSensor](ios *iostreams.IOStreams, status T, id int, sensorName, alarmMsg string) {
	ios.Println(theme.Bold().Render(fmt.Sprintf("%s Sensor %d:", sensorName, id)))
	ios.Println()
	ios.Printf("  Status: %s\n", output.RenderAlarmState(status.IsAlarm(), alarmMsg))
	ios.Printf("  Mute: %s\n", output.RenderMuteState(status.IsMuted()))
	displaySensorErrors(ios, status.GetErrors())
}

// DisplayTemperatureList displays a list of temperature sensors.
func DisplayTemperatureList(ios *iostreams.IOStreams, sensors []model.TemperatureReading) {
	ios.Println(theme.Bold().Render("Temperature Sensors:"))
	ios.Println()
	for _, s := range sensors {
		ios.Printf("  Sensor %d:\n", s.ID)
		if s.TC != nil {
			ios.Printf("    Temperature: %.1f°C", *s.TC)
			if s.TF != nil {
				ios.Printf(" (%.1f°F)", *s.TF)
			}
			ios.Println()
		}
	}
}

// DisplayTemperatureStatus displays a single temperature sensor status.
func DisplayTemperatureStatus(ios *iostreams.IOStreams, status model.TemperatureReading, id int) {
	ios.Println(theme.Bold().Render(fmt.Sprintf("Temperature Sensor %d:", id)))
	ios.Println()
	if status.TC != nil {
		ios.Printf("  Temperature: %s", theme.Highlight().Render(fmt.Sprintf("%.1f°C", *status.TC)))
		if status.TF != nil {
			ios.Printf(" (%s)", theme.Dim().Render(fmt.Sprintf("%.1f°F", *status.TF)))
		}
		ios.Println()
	} else {
		ios.Warning("No temperature reading available.")
	}
	displaySensorErrors(ios, status.Errors)
}

// DisplayHumidityList displays a list of humidity sensors.
func DisplayHumidityList(ios *iostreams.IOStreams, sensors []model.HumidityReading) {
	ios.Println(theme.Bold().Render("Humidity Sensors:"))
	ios.Println()
	for _, s := range sensors {
		ios.Printf("  Sensor %d:\n", s.ID)
		if s.RH != nil {
			ios.Printf("    Humidity: %.1f%%\n", *s.RH)
		}
	}
}

// DisplayHumidityStatus displays a single humidity sensor status.
func DisplayHumidityStatus(ios *iostreams.IOStreams, status model.HumidityReading, id int) {
	ios.Println(theme.Bold().Render(fmt.Sprintf("Humidity Sensor %d:", id)))
	ios.Println()
	if status.RH != nil {
		ios.Printf("  Humidity: %s\n", theme.Highlight().Render(fmt.Sprintf("%.1f%%", *status.RH)))
	} else {
		ios.Warning("No humidity reading available.")
	}
	displaySensorErrors(ios, status.Errors)
}

// DisplayIlluminanceList displays a list of illuminance sensors.
func DisplayIlluminanceList(ios *iostreams.IOStreams, sensors []model.IlluminanceReading) {
	ios.Println(theme.Bold().Render("Illuminance Sensors:"))
	ios.Println()
	for _, s := range sensors {
		ios.Printf("  Sensor %d:\n", s.ID)
		if s.Lux != nil {
			ios.Printf("    Light Level: %.0f lux\n", *s.Lux)
		}
	}
}

// DisplayIlluminanceStatus displays a single illuminance sensor status.
func DisplayIlluminanceStatus(ios *iostreams.IOStreams, status model.IlluminanceReading, id int) {
	ios.Println(theme.Bold().Render(fmt.Sprintf("Illuminance Sensor %d:", id)))
	ios.Println()
	if status.Lux != nil {
		level := output.GetLightLevel(*status.Lux)
		ios.Printf("  Light Level: %s (%s)\n",
			theme.Highlight().Render(fmt.Sprintf("%.0f lux", *status.Lux)),
			theme.Dim().Render(level))
	} else {
		ios.Warning("No illuminance reading available.")
	}
	displaySensorErrors(ios, status.Errors)
}

// DisplayVoltmeterList displays a list of voltmeter sensors.
func DisplayVoltmeterList(ios *iostreams.IOStreams, sensors []model.VoltmeterReading) {
	ios.Println(theme.Bold().Render("Voltmeter Sensors:"))
	ios.Println()
	for _, s := range sensors {
		ios.Printf("  Sensor %d:\n", s.ID)
		if s.Voltage != nil {
			ios.Printf("    Voltage: %.3f V\n", *s.Voltage)
		}
	}
}

// DisplayVoltmeterStatus displays a single voltmeter sensor status.
func DisplayVoltmeterStatus(ios *iostreams.IOStreams, status model.VoltmeterReading, id int) {
	ios.Println(theme.Bold().Render(fmt.Sprintf("Voltmeter Sensor %d:", id)))
	ios.Println()
	if status.Voltage != nil {
		ios.Printf("  Voltage: %s\n", theme.Highlight().Render(fmt.Sprintf("%.3f V", *status.Voltage)))
	} else {
		ios.Warning("No voltage reading available.")
	}
	displaySensorErrors(ios, status.Errors)
}

// DisplayDevicePowerList displays a list of device power (battery) sensors.
func DisplayDevicePowerList(ios *iostreams.IOStreams, sensors []model.DevicePowerReading) {
	ios.Println(theme.Bold().Render("Device Power Sensors:"))
	ios.Println()
	for _, s := range sensors {
		ios.Printf("  Sensor %d:\n", s.ID)
		ios.Printf("    Battery: %d%% (%.2fV)\n", s.Battery.Percent, s.Battery.V)
		ios.Printf("    External Power: %s\n", output.RenderYesNo(s.External.Present, output.CaseTitle, theme.FalseDim))
	}
}

// DisplayDevicePowerStatus displays a single device power sensor status.
func DisplayDevicePowerStatus(ios *iostreams.IOStreams, status model.DevicePowerReading, id int) {
	ios.Println(theme.Bold().Render(fmt.Sprintf("Device Power Sensor %d:", id)))
	ios.Println()

	// Battery percentage with color coding
	percentStr := fmt.Sprintf("%d%%", status.Battery.Percent)
	switch {
	case status.Battery.Percent <= 20:
		percentStr = theme.StatusError().Render(percentStr)
	case status.Battery.Percent <= 50:
		percentStr = theme.StatusWarn().Render(percentStr)
	default:
		percentStr = theme.StatusOK().Render(percentStr)
	}

	ios.Printf("  Battery: %s (%s)\n", percentStr, theme.Dim().Render(fmt.Sprintf("%.2fV", status.Battery.V)))
	ios.Printf("  External Power: %s\n", output.RenderYesNo(status.External.Present, output.CaseTitle, theme.FalseDim))
	displaySensorErrors(ios, status.Errors)
}

// displaySensorErrors prints sensor errors if present.
func displaySensorErrors(ios *iostreams.IOStreams, errors []string) {
	if len(errors) > 0 {
		ios.Println()
		ios.Warning("Errors: %v", errors)
	}
}

// DisplayPowerSnapshot displays a power monitoring snapshot with all energy meters.
// Handles 3-phase EM, single-phase EM1, and power meters with change indicators.
func DisplayPowerSnapshot(ios *iostreams.IOStreams, current, previous *shelly.MonitoringSnapshot) {
	// Clear screen for non-first updates
	if previous != nil {
		ios.ClearScreen()
	}

	ios.Title("Power Consumption")
	ios.Printf("  Updated: %s\n\n", current.Timestamp.Format("15:04:05"))

	// Display 3-phase energy meters
	for i := range current.EM {
		em := &current.EM[i]
		prevA, prevB, prevC := output.GetPrevEMPhasePower(em.ID, previous)
		ios.Printf("EM %d (3-phase):\n", em.ID)
		displayPhase(ios, "A", em.AActivePower, em.AVoltage, em.ACurrent, em.APowerFactor, prevA)
		displayPhase(ios, "B", em.BActivePower, em.BVoltage, em.BCurrent, em.BPowerFactor, prevB)
		displayPhase(ios, "C", em.CActivePower, em.CVoltage, em.CCurrent, em.CPowerFactor, prevC)
		ios.Printf("  Total: %s\n\n", output.FormatPowerColored(em.TotalActivePower))
	}

	// Display single-phase energy meters
	for i := range current.EM1 {
		em1 := &current.EM1[i]
		ios.Printf("EM1 %d:\n", em1.ID)
		prevPower := output.GetPrevEM1Power(em1.ID, previous)
		displayMeter(ios, em1.ActPower, em1.Voltage, em1.Current, em1.PF, prevPower)
	}

	// Display power meters
	for i := range current.PM {
		pm := &current.PM[i]
		ios.Printf("PM %d:\n", pm.ID)
		prevPower := output.GetPrevPMPower(pm.ID, previous)
		displayMeter(ios, pm.APower, pm.Voltage, pm.Current, nil, prevPower)
	}

	// Calculate and display totals
	totalPower, totalEnergy := output.CalculateSnapshotTotals(current)
	ios.Println()
	ios.Printf("═══════════════════════════════════════\n")
	ios.Printf("  Total Power:  %s\n", theme.StatusOK().Render(output.FormatPower(totalPower)))
	if totalEnergy > 0 {
		ios.Printf("  Total Energy: %.2f Wh\n", totalEnergy)
	}
}

// displayPhase prints a single phase line for power monitoring.
func displayPhase(ios *iostreams.IOStreams, phase string, power, voltage, current float64, pf, prevPower *float64) {
	powerStr := output.FormatPowerWithChange(power, prevPower)
	pfStr := ""
	if pf != nil {
		pfStr = fmt.Sprintf("  PF:%.2f", *pf)
	}
	ios.Printf("  Phase %s: %s  %.1fV  %.2fA%s\n", phase, powerStr, voltage, current, pfStr)
}

// displayMeter prints a meter line for power monitoring.
func displayMeter(ios *iostreams.IOStreams, power, voltage, current float64, pf, prevPower *float64) {
	powerStr := output.FormatPowerWithChange(power, prevPower)
	pfStr := ""
	if pf != nil {
		pfStr = fmt.Sprintf("  PF:%.2f", *pf)
	}
	ios.Printf("  Power: %s  Voltage: %.1fV  Current: %.2fA%s\n", powerStr, voltage, current, pfStr)
}

// DisplayStatusSnapshot displays a device status monitoring snapshot.
// Shows energy meters in a simpler format than power monitoring.
func DisplayStatusSnapshot(ios *iostreams.IOStreams, current, previous *shelly.MonitoringSnapshot) {
	// Clear screen for non-first updates
	if previous != nil {
		ios.ClearScreen()
	}

	ios.Title("Device Status")
	ios.Printf("  Timestamp: %s\n\n", current.Timestamp.Format("2006-01-02T15:04:05Z07:00"))

	// Display energy meters
	displayEMStatus(ios, current.EM, previous)
	displayEM1Status(ios, current.EM1, previous)

	// Display power meters
	displayPMStatus(ios, current.PM, previous)

	ios.Println()
}

func displayEMStatus(ios *iostreams.IOStreams, statuses []shelly.EMStatus, previous *shelly.MonitoringSnapshot) {
	if len(statuses) == 0 {
		return
	}

	ios.Printf("Energy Meters (3-phase):\n")
	for i := range statuses {
		em := &statuses[i]
		prev := output.FindPreviousEM(em.ID, previous)
		for _, line := range output.FormatEMLines(em, prev) {
			ios.Println(line)
		}
	}
	ios.Println()
}

func displayEM1Status(ios *iostreams.IOStreams, statuses []shelly.EM1Status, previous *shelly.MonitoringSnapshot) {
	if len(statuses) == 0 {
		return
	}

	ios.Printf("Energy Meters (single-phase):\n")
	for i := range statuses {
		em1 := &statuses[i]
		prev := output.FindPreviousEM1(em1.ID, previous)
		ios.Println(output.FormatEM1Line(em1, prev))
	}
	ios.Println()
}

func displayPMStatus(ios *iostreams.IOStreams, statuses []shelly.PMStatus, previous *shelly.MonitoringSnapshot) {
	if len(statuses) == 0 {
		return
	}

	ios.Printf("Power Meters:\n")
	for i := range statuses {
		pm := &statuses[i]
		prev := output.FindPreviousPM(pm.ID, previous)
		ios.Println(output.FormatPMLine(pm, prev))
	}
}

// DisplayEvent displays a single device event with color-coded type.
func DisplayEvent(ios *iostreams.IOStreams, event shelly.DeviceEvent) error {
	timestamp := event.Timestamp.Format("15:04:05.000")

	// Color code by event type
	eventStyle := theme.StatusOK()
	switch event.Event {
	case "state_changed":
		eventStyle = theme.StatusWarn()
	case "error":
		eventStyle = theme.StatusError()
	case "notification":
		eventStyle = theme.StatusInfo()
	}

	ios.Printf("[%s] %s %s:%d %s\n",
		theme.Dim().Render(timestamp),
		eventStyle.Render(event.Event),
		event.Component,
		event.ComponentID,
		formatEventData(event.Data))

	return nil
}

// OutputEventJSON outputs a device event as JSON.
func OutputEventJSON(ios *iostreams.IOStreams, event shelly.DeviceEvent) error {
	enc := json.NewEncoder(ios.Out)
	return enc.Encode(event)
}

func formatEventData(data map[string]any) string {
	if len(data) == 0 {
		return ""
	}

	// Format key fields
	var parts []string

	// Common fields
	if outputState, ok := data["output"].(bool); ok {
		if outputState {
			parts = append(parts, theme.StatusOK().Render("ON"))
		} else {
			parts = append(parts, theme.StatusError().Render("OFF"))
		}
	}

	if power, ok := data["apower"].(float64); ok {
		parts = append(parts, output.FormatPowerColored(power))
	}

	if temp, ok := data["temperature"].(map[string]any); ok {
		if tc, ok := temp["tC"].(float64); ok {
			parts = append(parts, formatTemp(tc))
		}
	}

	if len(parts) == 0 {
		// Fallback to JSON for unknown data
		bytes, err := json.Marshal(data)
		if err != nil {
			return ""
		}
		return string(bytes)
	}

	result := ""
	for i, p := range parts {
		if i > 0 {
			result += " "
		}
		result += p
	}
	return result
}

func formatTemp(c float64) string {
	s := fmt.Sprintf("%.1f°C", c)
	if c >= 70 {
		return theme.StatusError().Render(s)
	} else if c >= 50 {
		return theme.StatusWarn().Render(s)
	}
	return theme.StatusOK().Render(s)
}

// DisplayAllSnapshots displays a summary of all device snapshots.
// Requires a mutex lock to be held by the caller when accessing the snapshots map.
func DisplayAllSnapshots(ios *iostreams.IOStreams, snapshots map[string]*shelly.DeviceSnapshot) {
	// Clear screen
	ios.ClearScreen()

	ios.Title("Device Status Summary")
	ios.Printf("  Updated: %s\n\n", time.Now().Format("15:04:05"))

	totalPower := 0.0
	totalEnergy := 0.0
	onlineCount := 0
	offlineCount := 0

	// Display each device
	for name, snap := range snapshots {
		if snap.Error != nil {
			ios.Printf("%s %s: %s\n",
				theme.StatusError().Render("●"),
				name,
				theme.Dim().Render(snap.Error.Error()))
			offlineCount++
			continue
		}

		onlineCount++

		// Calculate device power using shared helper
		devicePower, deviceEnergy := output.CalculateSnapshotTotals(snap.Snapshot)
		totalPower += devicePower
		totalEnergy += deviceEnergy

		// Display device line
		statusIcon := theme.StatusOK().Render("●")
		deviceModel := "Unknown"
		if snap.Info != nil {
			deviceModel = snap.Info.Model
		}

		powerStr := output.FormatPowerColored(devicePower)
		energyStr := ""
		if deviceEnergy > 0 {
			energyStr = fmt.Sprintf("  %.2f Wh", deviceEnergy)
		}
		ios.Printf("%s %s (%s): %s%s\n",
			statusIcon, name, deviceModel, powerStr, energyStr)
	}

	// Display summary
	ios.Println()
	ios.Printf("═══════════════════════════════════════\n")
	ios.Printf("  Online:       %s / %d devices\n",
		theme.StatusOK().Render(fmt.Sprintf("%d", onlineCount)),
		onlineCount+offlineCount)
	ios.Printf("  Total Power:  %s\n", theme.StatusOK().Render(output.FormatPower(totalPower)))
	if totalEnergy > 0 {
		ios.Printf("  Total Energy: %.2f Wh\n", totalEnergy)
	}
}

// DisplayBackupSummary prints a summary of a created backup.
func DisplayBackupSummary(ios *iostreams.IOStreams, backup *shelly.DeviceBackup) {
	ios.Println()
	ios.Printf("  Device:    %s (%s)\n", backup.Device().ID, backup.Device().Model)
	ios.Printf("  Firmware:  %s\n", backup.Device().FWVersion)
	ios.Printf("  Config:    %d keys\n", len(backup.Config))
	if len(backup.Scripts) > 0 {
		ios.Printf("  Scripts:   %d\n", len(backup.Scripts))
	}
	if len(backup.Schedules) > 0 {
		ios.Printf("  Schedules: %d\n", len(backup.Schedules))
	}
	if len(backup.Webhooks) > 0 {
		ios.Printf("  Webhooks:  %d\n", len(backup.Webhooks))
	}
	if backup.Encrypted() {
		ios.Printf("  Encrypted: yes\n")
	}
}

// DisplayRestorePreview prints a preview of what would be restored.
func DisplayRestorePreview(ios *iostreams.IOStreams, backup *shelly.DeviceBackup, opts shelly.RestoreOptions) {
	DisplayBackupSource(ios, backup)
	ios.Printf("Will restore:\n")
	displayConfigPreview(ios, backup, opts)
	displayScriptsPreview(ios, backup, opts)
	displaySchedulesPreview(ios, backup, opts)
	displayWebhooksPreview(ios, backup, opts)
}

// DisplayBackupSource prints information about the backup source device.
func DisplayBackupSource(ios *iostreams.IOStreams, backup *shelly.DeviceBackup) {
	device := backup.Device()
	ios.Printf("Backup source:\n")
	ios.Printf("  Device:    %s (%s)\n", device.ID, device.Model)
	ios.Printf("  Firmware:  %s\n", device.FWVersion)
	ios.Printf("  Created:   %s\n", backup.CreatedAt.Format("2006-01-02 15:04:05"))
	ios.Println()
}

func displayConfigPreview(ios *iostreams.IOStreams, backup *shelly.DeviceBackup, opts shelly.RestoreOptions) {
	if len(backup.Config) > 0 {
		if opts.SkipNetwork {
			ios.Printf("  Config:    %d keys (network config excluded)\n", len(backup.Config))
		} else {
			ios.Printf("  Config:    %d keys\n", len(backup.Config))
		}
	}
}

func displayScriptsPreview(ios *iostreams.IOStreams, backup *shelly.DeviceBackup, opts shelly.RestoreOptions) {
	if len(backup.Scripts) > 0 {
		if opts.SkipScripts {
			ios.Printf("  Scripts:   %d (skipped)\n", len(backup.Scripts))
		} else {
			ios.Printf("  Scripts:   %d\n", len(backup.Scripts))
		}
	}
}

func displaySchedulesPreview(ios *iostreams.IOStreams, backup *shelly.DeviceBackup, opts shelly.RestoreOptions) {
	if len(backup.Schedules) > 0 {
		if opts.SkipSchedules {
			ios.Printf("  Schedules: %d (skipped)\n", len(backup.Schedules))
		} else {
			ios.Printf("  Schedules: %d\n", len(backup.Schedules))
		}
	}
}

func displayWebhooksPreview(ios *iostreams.IOStreams, backup *shelly.DeviceBackup, opts shelly.RestoreOptions) {
	if len(backup.Webhooks) > 0 {
		if opts.SkipWebhooks {
			ios.Printf("  Webhooks:  %d (skipped)\n", len(backup.Webhooks))
		} else {
			ios.Printf("  Webhooks:  %d\n", len(backup.Webhooks))
		}
	}
}

// DisplayRestoreResult prints the results of a restore operation.
func DisplayRestoreResult(ios *iostreams.IOStreams, result *shelly.RestoreResult) {
	ios.Println()
	if result.ConfigRestored {
		ios.Printf("  Config:    restored\n")
	}
	if result.ScriptsRestored > 0 {
		ios.Printf("  Scripts:   %d restored\n", result.ScriptsRestored)
	}
	if result.SchedulesRestored > 0 {
		ios.Printf("  Schedules: %d restored\n", result.SchedulesRestored)
	}
	if result.WebhooksRestored > 0 {
		ios.Printf("  Webhooks:  %d restored\n", result.WebhooksRestored)
	}

	if len(result.Warnings) > 0 {
		ios.Println()
		ios.Warning("Warnings:")
		for _, w := range result.Warnings {
			ios.Printf("  - %s\n", w)
		}
	}
}

// DisplayBackupsTable prints a table of backup files.
func DisplayBackupsTable(ios *iostreams.IOStreams, backups []export.BackupFileInfo) {
	table := output.FormatBackupsTable(backups)
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}
}

// DisplayBackupExportResults prints the results of a backup export operation.
func DisplayBackupExportResults(ios *iostreams.IOStreams, results []export.BackupResult) {
	for _, r := range results {
		if r.Success {
			ios.Printf("  Backing up %s (%s)... OK\n", r.DeviceName, r.Address)
		} else {
			ios.Printf("  Backing up %s (%s)... FAILED\n", r.DeviceName, r.Address)
		}
	}
}

// DisplayAllSensorData displays aggregated sensor data from a device.
// Only displays sections for sensors that are present.
func DisplayAllSensorData(ios *iostreams.IOStreams, data *model.SensorData, device string) {
	ios.Println(theme.Bold().Render(fmt.Sprintf("Sensor Readings for %s:", device)))
	ios.Println()

	hasData := displayAllTemperature(ios, data.Temperature) ||
		displayAllHumidity(ios, data.Humidity) ||
		displayAllAlarmSensor(ios, data.Flood, "Flood Detection", "Flood", "WATER DETECTED!") ||
		displayAllAlarmSensor(ios, data.Smoke, "Smoke Detection", "Smoke", "SMOKE DETECTED!") ||
		displayAllIlluminance(ios, data.Illuminance) ||
		displayAllVoltmeter(ios, data.Voltmeter)

	if !hasData {
		ios.Info("No sensors found on this device.")
	}
}

// displayAllAlarmSensor displays alarm-type sensors (flood, smoke) in the all-sensors view.
func displayAllAlarmSensor(ios *iostreams.IOStreams, sensors []model.AlarmSensorReading, title, sensorType, alarmMsg string) bool {
	if len(sensors) == 0 {
		return false
	}
	ios.Println("  " + theme.Highlight().Render(title+":"))
	DisplayAlarmSensors(ios, sensors, sensorType, alarmMsg)
	ios.Println()
	return true
}

// displayAllHumidity displays humidity sensors in the all-sensors view.
func displayAllHumidity(ios *iostreams.IOStreams, humids []model.HumidityReading) bool {
	if len(humids) == 0 {
		return false
	}
	ios.Println("  " + theme.Highlight().Render("Humidity:"))
	for _, h := range humids {
		if h.RH != nil {
			ios.Printf("    Sensor %d: %.1f%%\n", h.ID, *h.RH)
		}
	}
	ios.Println()
	return true
}

// displayAllIlluminance displays illuminance sensors in the all-sensors view.
func displayAllIlluminance(ios *iostreams.IOStreams, luxes []model.IlluminanceReading) bool {
	if len(luxes) == 0 {
		return false
	}
	ios.Println("  " + theme.Highlight().Render("Illuminance:"))
	for _, l := range luxes {
		if l.Lux != nil {
			ios.Printf("    Sensor %d: %.0f lux\n", l.ID, *l.Lux)
		}
	}
	ios.Println()
	return true
}

// displayAllTemperature displays temperature sensors in the all-sensors view.
func displayAllTemperature(ios *iostreams.IOStreams, temps []model.TemperatureReading) bool {
	if len(temps) == 0 {
		return false
	}
	ios.Println("  " + theme.Highlight().Render("Temperature:"))
	for _, t := range temps {
		if t.TC != nil {
			ios.Printf("    Sensor %d: %.1f°C", t.ID, *t.TC)
			if t.TF != nil {
				ios.Printf(" (%.1f°F)", *t.TF)
			}
			ios.Println()
		}
	}
	ios.Println()
	return true
}

// displayAllVoltmeter displays voltmeter sensors in the all-sensors view.
func displayAllVoltmeter(ios *iostreams.IOStreams, volts []model.VoltmeterReading) bool {
	if len(volts) == 0 {
		return false
	}
	ios.Println("  " + theme.Highlight().Render("Voltage:"))
	for _, v := range volts {
		if v.Voltage != nil {
			ios.Printf("    Sensor %d: %.2f V\n", v.ID, *v.Voltage)
		}
	}
	ios.Println()
	return true
}
