// Package cmdutil provides display helpers that print directly to IOStreams.
// Display* functions wrap pure formatters from output/ with printing and semantic output.
package cmdutil

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/export"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/version"
)

// DisplayPowerMetrics outputs power, voltage, and current with units.
// Nil values are skipped.
func DisplayPowerMetrics(ios *iostreams.IOStreams, power, voltage, current *float64) {
	displayPowerMetricsWithWidth(ios, power, voltage, current, 9)
}

// DisplayPowerMetricsWide outputs power metrics with wider alignment for cover status.
func DisplayPowerMetricsWide(ios *iostreams.IOStreams, power, voltage, current *float64) {
	displayPowerMetricsWithWidth(ios, power, voltage, current, 10)
}

func displayPowerMetricsWithWidth(ios *iostreams.IOStreams, power, voltage, current *float64, width int) {
	if power != nil {
		ios.Printf("  %-*s%.1f W\n", width, "Power:", *power)
	}
	if voltage != nil {
		ios.Printf("  %-*s%.1f V\n", width, "Voltage:", *voltage)
	}
	if current != nil {
		ios.Printf("  %-*s%.3f A\n", width, "Current:", *current)
	}
}

// printTable prints a table to ios.Out with standard error handling.
func printTable(ios *iostreams.IOStreams, table *output.Table) {
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
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
	printTable(ios, table)
	ios.Count("device", len(devices))
}

// DisplayConfigTable prints a configuration map as formatted tables.
// Each top-level key becomes a titled section with a settings table.
func DisplayConfigTable(ios *iostreams.IOStreams, configData any) error {
	configMap, ok := configData.(map[string]any)
	if !ok {
		return output.PrintJSON(configData)
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
func DisplayBackupsTable(ios *iostreams.IOStreams, backups []model.BackupFileInfo) {
	table := output.FormatBackupsTable(backups)
	printTable(ios, table)
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

// Migration diff display uses model.DiffAdded, model.DiffRemoved, model.DiffChanged constants.

// displayDiffSection is a generic helper for printing diff sections.
// It handles the common pattern of: empty check, header, loop with switch on diff type.
func displayDiffSection[T any](
	ios *iostreams.IOStreams,
	diffs []T,
	verbose bool,
	header, addedMsg string,
	getLabel func(T) string,
	getDiffType func(T) string,
) {
	if len(diffs) == 0 {
		return
	}
	if verbose {
		ios.Printf("%s:\n", header)
	} else {
		ios.Printf("%s changes:\n", header)
	}
	for _, d := range diffs {
		label := getLabel(d)
		switch getDiffType(d) {
		case model.DiffAdded:
			if verbose {
				ios.Printf("  + %s (%s)\n", label, addedMsg)
			} else {
				ios.Printf("  + %s\n", label)
			}
		case model.DiffRemoved:
			if verbose {
				ios.Printf("  - %s (exists on device, not in backup)\n", label)
			} else {
				ios.Printf("  - %s\n", label)
			}
		case model.DiffChanged:
			if verbose {
				ios.Printf("  ~ %s (will be updated)\n", label)
			} else {
				ios.Printf("  ~ %s\n", label)
			}
		}
	}
	ios.Println()
}

// DisplayConfigDiffs prints configuration differences.
// When verbose is true, adds descriptive context to each line.
func DisplayConfigDiffs(ios *iostreams.IOStreams, diffs []model.ConfigDiff, verbose bool) {
	displayDiffSection(ios, diffs, verbose, "Configuration", "will be added from backup",
		func(d model.ConfigDiff) string { return d.Path },
		func(d model.ConfigDiff) string { return d.DiffType })
}

// DisplayScriptDiffs prints script differences.
// When verbose is true, adds descriptive context to each line.
func DisplayScriptDiffs(ios *iostreams.IOStreams, diffs []model.ScriptDiff, verbose bool) {
	displayDiffSection(ios, diffs, verbose, "Script", "will be created",
		func(d model.ScriptDiff) string { return d.Name },
		func(d model.ScriptDiff) string { return d.DiffType })
}

// DisplayScheduleDiffs prints schedule differences.
// When verbose is true, adds descriptive context to each line.
func DisplayScheduleDiffs(ios *iostreams.IOStreams, diffs []model.ScheduleDiff, verbose bool) {
	displayDiffSection(ios, diffs, verbose, "Schedule", "will be created",
		func(d model.ScheduleDiff) string { return d.Timespec },
		func(d model.ScheduleDiff) string { return d.DiffType })
}

// DisplayWebhookDiffs prints webhook differences.
// When verbose is true, adds descriptive context to each line.
func DisplayWebhookDiffs(ios *iostreams.IOStreams, diffs []model.WebhookDiff, verbose bool) {
	displayDiffSection(ios, diffs, verbose, "Webhook", "will be created",
		func(d model.WebhookDiff) string {
			if d.Name != "" {
				return d.Name
			}
			return d.Event
		},
		func(d model.WebhookDiff) string { return d.DiffType })
}

// DisplayConfigDiffsSummary prints configuration differences with grouped sections and summary.
func DisplayConfigDiffsSummary(ios *iostreams.IOStreams, diffs []model.ConfigDiff) {
	var added, removed, changed []model.ConfigDiff
	for _, d := range diffs {
		switch d.DiffType {
		case model.DiffAdded:
			added = append(added, d)
		case model.DiffRemoved:
			removed = append(removed, d)
		case model.DiffChanged:
			changed = append(changed, d)
		}
	}

	if len(removed) > 0 {
		ios.Println(output.RenderDiffRemoved())
		for _, d := range removed {
			ios.Printf("  - %s: %v\n", d.Path, output.FormatDisplayValue(d.OldValue))
		}
		ios.Println("")
	}

	if len(added) > 0 {
		ios.Println(output.RenderDiffAdded())
		for _, d := range added {
			ios.Printf("  + %s: %v\n", d.Path, output.FormatDisplayValue(d.NewValue))
		}
		ios.Println("")
	}

	if len(changed) > 0 {
		ios.Println(output.RenderDiffChanged())
		for _, d := range changed {
			ios.Printf("  ~ %s:\n", d.Path)
			ios.Printf("    - %v\n", output.FormatDisplayValue(d.OldValue))
			ios.Printf("    + %v\n", output.FormatDisplayValue(d.NewValue))
		}
		ios.Println("")
	}

	ios.Printf("Summary: %d added, %d removed, %d changed\n", len(added), len(removed), len(changed))
}

// DisplayUpdateAvailable prints an update notification with current and available versions.
func DisplayUpdateAvailable(ios *iostreams.IOStreams, currentVersion, availableVersion string) {
	ios.Printf("\n")
	ios.Warning("Update available: %s -> %s", currentVersion, availableVersion)
	ios.Printf("  Run 'shelly update' to install the latest version\n")
}

// DisplayUpToDate prints a success message indicating the CLI is up to date.
func DisplayUpToDate(ios *iostreams.IOStreams) {
	ios.Printf("\n")
	ios.Success("You are using the latest version")
}

// DisplayUpdateResult displays the result of an update check.
// Logs any cache write errors via ios.DebugErr.
func DisplayUpdateResult(ios *iostreams.IOStreams, currentVersion, latestVersion string, updateAvailable bool, cacheErr error) {
	if updateAvailable {
		DisplayUpdateAvailable(ios, currentVersion, latestVersion)
	} else {
		DisplayUpToDate(ios)
	}
	if cacheErr != nil {
		ios.DebugErr("writing version cache", cacheErr)
	}
}

// DisplayVersionInfo prints version information to the console.
func DisplayVersionInfo(ios *iostreams.IOStreams, ver, commit, date, builtBy, goVer, osName, arch string) {
	const unknownValue = "unknown"
	ios.Printf("shelly version %s\n", ver)
	if commit != "" && commit != unknownValue {
		ios.Printf("  commit: %s\n", commit)
	}
	if date != "" && date != unknownValue {
		ios.Printf("  built: %s\n", date)
	}
	if builtBy != "" && builtBy != unknownValue {
		ios.Printf("  by: %s\n", builtBy)
	}
	ios.Printf("  go: %s\n", goVer)
	ios.Printf("  os/arch: %s/%s\n", osName, arch)
}

// UpdateChecker is a function that checks for updates and returns the result.
type UpdateChecker func(ctx context.Context) (*version.UpdateResult, error)

// RunUpdateCheck runs an update check with spinner and displays results.
func RunUpdateCheck(ctx context.Context, ios *iostreams.IOStreams, checker UpdateChecker) {
	ios.StartProgress("Checking for updates...")
	checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	result, err := checker(checkCtx)
	cancel()
	ios.StopProgress()
	if err != nil {
		ios.DebugErr("checking for updates", err)
	} else if !result.SkippedDevBuild {
		DisplayUpdateResult(ios, result.CurrentVersion, result.LatestVersion, result.UpdateAvailable, result.CacheWriteErr)
	}
}

// DisplayDashboard prints the energy dashboard with summary and device breakdown.
func DisplayDashboard(ios *iostreams.IOStreams, data model.DashboardData) {
	ios.Printf("%s\n", theme.Bold().Render("Energy Dashboard"))
	ios.Printf("Timestamp: %s\n\n", data.Timestamp.Format(time.RFC3339))

	ios.Printf("%s\n", theme.Bold().Render("Summary"))
	ios.Printf("  Devices:     %d total (%d online, %d offline)\n",
		data.DeviceCount, data.OnlineCount, data.OfflineCount)
	ios.Printf("  Total Power: %s\n", theme.StyledPower(data.TotalPower))

	if data.TotalEnergy > 0 {
		ios.Printf("  Total Energy: %s\n", theme.StyledEnergy(data.TotalEnergy))
	}

	if data.EstimatedCost != nil {
		ios.Printf("  Est. Cost:   %s %.2f/kWh = %s %.4f\n",
			data.CostCurrency, data.CostPerKwh,
			data.CostCurrency, *data.EstimatedCost)
	}

	ios.Printf("\n%s\n", theme.Bold().Render("Device Breakdown"))

	table := output.NewTable("Device", "Status", "Power", "Components")
	for _, dev := range data.Devices {
		statusStr := theme.StatusOK().Render("online")
		if !dev.Online {
			statusStr = theme.StatusError().Render("offline")
		}

		powerStr := output.FormatPower(dev.TotalPower)
		if !dev.Online {
			powerStr = "-"
		}

		table.AddRow(dev.Device, statusStr, powerStr, formatComponentSummary(dev.Components))
	}

	printTable(ios, table)
}

func formatComponentSummary(components []model.ComponentPower) string {
	if len(components) == 0 {
		return "-"
	}

	counts := make(map[string]int)
	for _, c := range components {
		counts[c.Type]++
	}

	parts := make([]string, 0, len(counts))
	for typ, count := range counts {
		parts = append(parts, fmt.Sprintf("%d %s", count, typ))
	}

	return fmt.Sprintf("%d (%s)", len(components), joinStrings(parts, ", "))
}

func joinStrings(parts []string, sep string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += sep
		}
		result += p
	}
	return result
}

// DisplayComparison prints energy comparison results with summary and bar chart.
func DisplayComparison(ios *iostreams.IOStreams, data model.ComparisonData) {
	ios.Printf("%s\n", theme.Bold().Render("Energy Comparison"))
	ios.Printf("Period: %s\n", data.Period)
	if !data.From.IsZero() {
		ios.Printf("From:   %s\n", data.From.Format("2006-01-02 15:04:05"))
	}
	if !data.To.IsZero() {
		ios.Printf("To:     %s\n", data.To.Format("2006-01-02 15:04:05"))
	}
	ios.Printf("\n")

	ios.Printf("%s\n", theme.Bold().Render("Summary"))
	ios.Printf("  Total Energy: %s\n", theme.StyledEnergy(data.TotalEnergy*1000))
	ios.Printf("  Max Device:   %s\n", theme.StyledEnergy(data.MaxEnergy*1000))
	ios.Printf("  Min Device:   %s\n", theme.StyledEnergy(data.MinEnergy*1000))
	ios.Printf("\n")

	// Sort by energy consumption (descending) for display
	sorted := sortDevicesByEnergy(data.Devices)

	ios.Printf("%s\n", theme.Bold().Render("Device Breakdown"))

	table := output.NewTable("Rank", "Device", "Energy", "Avg Power", "Peak Power", "Share", "Status")
	for i, dev := range sorted {
		rank := fmt.Sprintf("#%d", i+1)

		statusStr := theme.StatusOK().Render("ok")
		if !dev.Online {
			statusStr = theme.StatusError().Render("offline")
		} else if dev.Error != "" {
			statusStr = theme.StatusWarn().Render(output.Truncate(dev.Error, 15))
		}

		energyStr, avgStr, peakStr, shareStr := "-", "-", "-", "-"
		if dev.Online {
			energyStr = output.FormatEnergy(dev.Energy * 1000)
			avgStr = output.FormatPower(dev.AvgPower)
			peakStr = output.FormatPower(dev.PeakPower)
			if dev.Percentage > 0 {
				shareStr = fmt.Sprintf("%.1f%%", dev.Percentage)
			}
		}

		table.AddRow(rank, dev.Device, energyStr, avgStr, peakStr, shareStr, statusStr)
	}

	printTable(ios, table)

	ios.Printf("\n%s\n", theme.Bold().Render("Energy Distribution"))
	displayBarChart(ios, sorted, data.MaxEnergy)
}

func sortDevicesByEnergy(devices []model.DeviceEnergy) []model.DeviceEnergy {
	sorted := make([]model.DeviceEnergy, len(devices))
	copy(sorted, devices)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Energy > sorted[j].Energy
	})
	return sorted
}

func displayBarChart(ios *iostreams.IOStreams, devices []model.DeviceEnergy, maxEnergy float64) {
	if maxEnergy <= 0 {
		return
	}

	maxNameLen := 0
	for _, dev := range devices {
		if len(dev.Device) > maxNameLen {
			maxNameLen = len(dev.Device)
		}
	}

	for _, dev := range devices {
		if !dev.Online || dev.Energy <= 0 {
			continue
		}

		name := output.PadRight(dev.Device, maxNameLen)
		barLen := int((dev.Energy / maxEnergy) * 40)
		if barLen < 1 {
			barLen = 1
		}

		bar := repeatChar('█', barLen)
		ios.Printf("  %s │ %s %.2f kWh\n", name, theme.Highlight().Render(bar), dev.Energy)
	}
}

func repeatChar(c rune, n int) string {
	result := make([]rune, n)
	for i := range result {
		result[i] = c
	}
	return string(result)
}

// DisplaySwitchStatus prints switch component status.
func DisplaySwitchStatus(ios *iostreams.IOStreams, status *model.SwitchStatus) {
	ios.Title("Switch %d Status", status.ID)
	ios.Println()

	ios.Printf("  State:   %s\n", output.RenderOnOff(status.Output, output.CaseUpper, theme.FalseError))
	DisplayPowerMetrics(ios, status.Power, status.Voltage, status.Current)
	if status.Energy != nil {
		ios.Printf("  Energy:  %.2f Wh\n", status.Energy.Total)
	}
}

// DisplayLightStatus prints light component status.
func DisplayLightStatus(ios *iostreams.IOStreams, status *model.LightStatus) {
	ios.Title("Light %d Status", status.ID)
	ios.Println()

	ios.Printf("  State:      %s\n", output.RenderOnOff(status.Output, output.CaseUpper, theme.FalseError))
	if status.Brightness != nil {
		ios.Printf("  Brightness: %d%%\n", *status.Brightness)
	}
	if status.Power != nil {
		ios.Printf("  Power:      %.1f W\n", *status.Power)
	}
	if status.Voltage != nil {
		ios.Printf("  Voltage:    %.1f V\n", *status.Voltage)
	}
	if status.Current != nil {
		ios.Printf("  Current:    %.3f A\n", *status.Current)
	}
}

// DisplayRGBStatus prints RGB component status.
func DisplayRGBStatus(ios *iostreams.IOStreams, status *model.RGBStatus) {
	ios.Title("RGB %d Status", status.ID)
	ios.Println()

	ios.Printf("  State:      %s\n", output.RenderOnOff(status.Output, output.CaseUpper, theme.FalseError))
	if status.RGB != nil {
		ios.Printf("  Color:      R:%d G:%d B:%d\n",
			status.RGB.Red, status.RGB.Green, status.RGB.Blue)
	}
	if status.Brightness != nil {
		ios.Printf("  Brightness: %d%%\n", *status.Brightness)
	}
	if status.Power != nil {
		ios.Printf("  Power:      %.1f W\n", *status.Power)
	}
	if status.Voltage != nil {
		ios.Printf("  Voltage:    %.1f V\n", *status.Voltage)
	}
	if status.Current != nil {
		ios.Printf("  Current:    %.3f A\n", *status.Current)
	}
}

// DisplayCoverStatus prints cover component status.
func DisplayCoverStatus(ios *iostreams.IOStreams, status *model.CoverStatus) {
	ios.Title("Cover %d Status", status.ID)
	ios.Println()

	ios.Printf("  State:    %s\n", output.RenderCoverState(status.State))
	if status.CurrentPosition != nil {
		ios.Printf("  Position: %d%%\n", *status.CurrentPosition)
	}
	DisplayPowerMetricsWide(ios, status.Power, status.Voltage, status.Current)
}

// DisplayInputStatus prints input component status.
func DisplayInputStatus(ios *iostreams.IOStreams, status *model.InputStatus) {
	ios.Title("Input %d Status", status.ID)
	ios.Println()

	ios.Printf("  State: %s\n", output.RenderActive(status.State, output.CaseLower, theme.FalseError))
	if status.Type != "" {
		ios.Printf("  Type:  %s\n", status.Type)
	}
}

// DisplaySwitchList prints a table of switches.
func DisplaySwitchList(ios *iostreams.IOStreams, switches []shelly.SwitchInfo) {
	t := output.NewTable("ID", "Name", "State", "Power")
	for _, sw := range switches {
		name := output.FormatComponentName(sw.Name, "switch", sw.ID)
		state := output.RenderOnOff(sw.Output, output.CaseUpper, theme.FalseError)
		power := output.FormatPowerTableValue(sw.Power)
		t.AddRow(fmt.Sprintf("%d", sw.ID), name, state, power)
	}
	if err := t.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}
}

// DisplayLightList prints a table of lights.
func DisplayLightList(ios *iostreams.IOStreams, lights []shelly.LightInfo) {
	t := output.NewTable("ID", "Name", "State", "Brightness", "Power")
	for _, lt := range lights {
		name := output.FormatComponentName(lt.Name, "light", lt.ID)
		state := output.RenderOnOff(lt.Output, output.CaseUpper, theme.FalseError)

		brightness := "-"
		if lt.Brightness >= 0 {
			brightness = fmt.Sprintf("%d%%", lt.Brightness)
		}

		power := output.FormatPowerTableValue(lt.Power)
		t.AddRow(fmt.Sprintf("%d", lt.ID), name, state, brightness, power)
	}
	if err := t.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}
}

// DisplayRGBList prints a table of RGB components.
func DisplayRGBList(ios *iostreams.IOStreams, rgbs []shelly.RGBInfo) {
	t := output.NewTable("ID", "Name", "State", "Color", "Brightness", "Power")
	for _, rgb := range rgbs {
		name := output.FormatComponentName(rgb.Name, "rgb", rgb.ID)
		state := output.RenderOnOff(rgb.Output, output.CaseUpper, theme.FalseError)
		color := fmt.Sprintf("R:%d G:%d B:%d", rgb.Red, rgb.Green, rgb.Blue)

		brightness := "-"
		if rgb.Brightness >= 0 {
			brightness = fmt.Sprintf("%d%%", rgb.Brightness)
		}

		power := output.FormatPowerTableValue(rgb.Power)
		t.AddRow(fmt.Sprintf("%d", rgb.ID), name, state, color, brightness, power)
	}
	if err := t.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}
}

// DisplayCoverList prints a table of covers.
func DisplayCoverList(ios *iostreams.IOStreams, covers []shelly.CoverInfo) {
	t := output.NewTable("ID", "Name", "State", "Position", "Power")
	for _, cover := range covers {
		name := output.FormatComponentName(cover.Name, "cover", cover.ID)
		state := output.RenderCoverState(cover.State)

		position := "-"
		if cover.Position >= 0 {
			position = fmt.Sprintf("%d%%", cover.Position)
		}

		power := output.FormatPowerTableValue(cover.Power)
		t.AddRow(fmt.Sprintf("%d", cover.ID), name, state, position, power)
	}
	if err := t.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}
}

// DisplayInputList prints a table of inputs.
func DisplayInputList(ios *iostreams.IOStreams, inputs []shelly.InputInfo) {
	table := output.NewTable("ID", "Name", "Type", "State")

	for _, input := range inputs {
		name := input.Name
		if name == "" {
			name = theme.Dim().Render("-")
		}

		state := output.RenderActive(input.State, output.CaseLower, theme.FalseError)

		table.AddRow(
			fmt.Sprintf("%d", input.ID),
			name,
			input.Type,
			state,
		)
	}

	printTable(ios, table)
	ios.Println()
	ios.Count("input", len(inputs))
}

// DisplayDeviceStatus prints the device status information.
func DisplayDeviceStatus(ios *iostreams.IOStreams, status *shelly.DeviceStatus) {
	ios.Info("Device: %s", theme.Bold().Render(status.Info.ID))
	ios.Info("Model: %s (Gen%d)", status.Info.Model, status.Info.Generation)
	ios.Info("Firmware: %s", status.Info.Firmware)
	ios.Println()

	table := output.NewTable("Component", "Value")
	for key, value := range status.Status {
		table.AddRow(key, output.FormatDisplayValue(value))
	}
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print device status table", err)
	}
}

// DisplayAuthStatus prints the authentication status.
func DisplayAuthStatus(ios *iostreams.IOStreams, status *shelly.AuthStatus) {
	ios.Title("Authentication Status")
	ios.Println()

	ios.Printf("  Status: %s\n", output.RenderBoolState(status.Enabled, "Enabled", "Disabled"))
}

// DisplayFirmwareStatus prints the firmware status.
func DisplayFirmwareStatus(ios *iostreams.IOStreams, status *shelly.FirmwareStatus) {
	ios.Println(theme.Bold().Render("Firmware Status"))
	ios.Println("")

	// Status
	statusStr := status.Status
	if statusStr == "" {
		statusStr = "idle"
	}
	ios.Printf("  Status:      %s\n", statusStr)

	// Update available
	ios.Printf("  Update:      %s\n", output.RenderAvailableState(status.HasUpdate, "up to date"))
	if status.HasUpdate && status.NewVersion != "" {
		ios.Printf("  New Version: %s\n", status.NewVersion)
	}

	// Progress (if updating)
	if status.Progress > 0 {
		ios.Printf("  Progress:    %d%%\n", status.Progress)
	}

	// Rollback
	ios.Printf("  Rollback:    %s\n", output.RenderAvailableState(status.CanRollback, "not available"))
}

// DisplayFirmwareInfo prints firmware check information.
func DisplayFirmwareInfo(ios *iostreams.IOStreams, info *shelly.FirmwareInfo) {
	ios.Println(theme.Bold().Render("Firmware Information"))
	ios.Println("")

	// Device info
	ios.Printf("  Device:     %s (%s)\n", info.DeviceID, info.DeviceModel)
	ios.Printf("  Generation: Gen%d\n", info.Generation)
	ios.Println("")

	// Version info
	ios.Printf("  Current:    %s\n", info.Current)

	switch {
	case info.HasUpdate:
		ios.Printf("  Available:  %s %s\n",
			info.Available,
			output.RenderUpdateStatus(true))
	case info.Available != "":
		ios.Printf("  Available:  %s %s\n",
			info.Available,
			output.RenderUpdateStatus(false))
	default:
		ios.Printf("  Available:  %s\n", output.RenderUpdateStatus(false))
	}

	if info.Beta != "" {
		ios.Printf("  Beta:       %s\n", info.Beta)
	}
}

// DisplayWiFiStatus prints WiFi status information.
func DisplayWiFiStatus(ios *iostreams.IOStreams, status *shelly.WiFiStatus) {
	ios.Title("WiFi Status")
	ios.Println()

	ios.Printf("  Status:      %s\n", status.Status)
	ios.Printf("  SSID:        %s\n", valueOrEmpty(status.SSID))
	ios.Printf("  IP Address:  %s\n", valueOrEmpty(status.StaIP))
	ios.Printf("  Signal:      %d dBm\n", status.RSSI)
	if status.APCount > 0 {
		ios.Printf("  AP Clients:  %d\n", status.APCount)
	}
}

func valueOrEmpty(s string) string {
	if s == "" {
		return "<not connected>"
	}
	return s
}

// DisplayWiFiAPClients prints a table of connected WiFi AP clients.
func DisplayWiFiAPClients(ios *iostreams.IOStreams, clients []shelly.WiFiAPClient) {
	ios.Title("Connected Clients")
	ios.Println()

	table := output.NewTable("MAC Address", "IP Address")
	for _, c := range clients {
		ip := c.IP
		if ip == "" {
			ip = "<no IP>"
		}
		table.AddRow(c.MAC, ip)
	}
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print wifi ap clients table", err)
	}

	ios.Printf("\n%d client(s) connected\n", len(clients))
}

// DisplayWiFiScanResults prints a table of WiFi scan results.
func DisplayWiFiScanResults(ios *iostreams.IOStreams, results []shelly.WiFiScanResult) {
	ios.Title("Available WiFi Networks")
	ios.Println()

	table := output.NewTable("SSID", "Signal", "Channel", "Security")
	for _, r := range results {
		ssid := r.SSID
		if ssid == "" {
			ssid = "<hidden>"
		}
		signal := formatWiFiSignal(r.RSSI)
		table.AddRow(ssid, signal, fmt.Sprintf("%d", r.Channel), r.Auth)
	}
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print wifi scan table", err)
	}

	ios.Printf("\n%d network(s) found\n", len(results))
}

func formatWiFiSignal(rssi int) string {
	bars := "▁▃▅▇"
	var strength int
	switch {
	case rssi >= -50:
		strength = 4
	case rssi >= -60:
		strength = 3
	case rssi >= -70:
		strength = 2
	default:
		strength = 1
	}
	return fmt.Sprintf("%s %d dBm", bars[:strength], rssi)
}

// DisplayEthernetStatus prints Ethernet status information.
func DisplayEthernetStatus(ios *iostreams.IOStreams, status *shelly.EthernetStatus) {
	ios.Title("Ethernet Status")
	ios.Println()

	if status.IP != "" {
		ios.Printf("  Status:     Connected\n")
		ios.Printf("  IP Address: %s\n", status.IP)
	} else {
		ios.Printf("  Status:     Not connected\n")
	}
}

// DisplayMQTTStatus prints MQTT status information.
func DisplayMQTTStatus(ios *iostreams.IOStreams, status *shelly.MQTTStatus) {
	ios.Title("MQTT Status")
	ios.Println()

	ios.Printf("  Status: %s\n", output.RenderBoolState(status.Connected, "Connected", "Disconnected"))
}

// DisplayCloudConnectionStatus prints cloud connection status.
func DisplayCloudConnectionStatus(ios *iostreams.IOStreams, status *shelly.CloudStatus) {
	ios.Title("Cloud Status")
	ios.Println()

	ios.Printf("  Status: %s\n", output.RenderBoolState(status.Connected, "Connected", "Disconnected"))
}

// DisplayCloudDevice prints cloud device details.
func DisplayCloudDevice(ios *iostreams.IOStreams, device *shelly.CloudDevice, showStatus bool) {
	ios.Title("Cloud Device")
	ios.Println()

	ios.Printf("  ID:     %s\n", device.ID)

	if device.Model != "" {
		ios.Printf("  Model:  %s\n", device.Model)
	}

	if device.Generation > 0 {
		ios.Printf("  Gen:    %d\n", device.Generation)
	}

	if device.MAC != "" {
		ios.Printf("  MAC:    %s\n", device.MAC)
	}

	if device.FirmwareVersion != "" {
		ios.Printf("  FW:     %s\n", device.FirmwareVersion)
	}

	ios.Printf("  Status: %s\n", output.RenderOnline(device.Online, output.CaseTitle))

	// Show status JSON if requested and available
	if showStatus && len(device.Status) > 0 {
		ios.Println()
		ios.Title("Device Status")
		ios.Println()
		printCloudJSON(ios, device.Status)
	}

	// Show settings if available
	if showStatus && len(device.Settings) > 0 {
		ios.Println()
		ios.Title("Device Settings")
		ios.Println()
		printCloudJSON(ios, device.Settings)
	}
}

func printCloudJSON(ios *iostreams.IOStreams, data json.RawMessage) {
	var prettyJSON map[string]any
	if err := json.Unmarshal(data, &prettyJSON); err != nil {
		ios.Printf("  %s\n", string(data))
		return
	}

	formatted, err := json.MarshalIndent(prettyJSON, "  ", "  ")
	if err != nil {
		ios.Printf("  %s\n", string(data))
		return
	}

	ios.Printf("  %s\n", string(formatted))
}

// DisplayCloudDevices prints a table of cloud devices.
func DisplayCloudDevices(ios *iostreams.IOStreams, devices []shelly.CloudDevice) {
	if len(devices) == 0 {
		ios.Info("No devices found in your Shelly Cloud account")
		return
	}

	// Sort by ID for consistent display
	sort.Slice(devices, func(i, j int) bool {
		return devices[i].ID < devices[j].ID
	})

	table := output.NewTable("ID", "Model", "Gen", "Online")

	for _, d := range devices {
		devModel := d.Model
		if devModel == "" {
			devModel = output.FormatPlaceholder("unknown")
		}

		gen := output.FormatPlaceholder("-")
		if d.Generation > 0 {
			gen = fmt.Sprintf("%d", d.Generation)
		}

		table.AddRow(d.ID, devModel, gen, output.RenderYesNo(d.Online, output.CaseLower, theme.FalseError))
	}

	ios.Printf("Found %d device(s):\n\n", len(devices))
	printTable(ios, table)
}

// DisplaySceneList prints a table of scenes.
func DisplaySceneList(ios *iostreams.IOStreams, scenes []config.Scene) {
	table := output.NewTable("Name", "Actions", "Description")
	for _, scene := range scenes {
		actions := output.FormatActionCount(len(scene.Actions))
		description := scene.Description
		if description == "" {
			description = "-"
		}
		table.AddRow(scene.Name, actions, description)
	}

	printTable(ios, table)
	ios.Println()
	ios.Count("scene", len(scenes))
}

// DisplayAliasList prints a table of aliases.
func DisplayAliasList(ios *iostreams.IOStreams, aliases []config.Alias) {
	table := output.NewTable("Name", "Command", "Type")

	for _, alias := range aliases {
		aliasType := "command"
		if alias.Shell {
			aliasType = "shell"
		}
		table.AddRow(alias.Name, alias.Command, aliasType)
	}

	printTable(ios, table)
	ios.Println()
	ios.Count("alias", len(aliases))
}

// DisplayScriptList prints a table of scripts.
func DisplayScriptList(ios *iostreams.IOStreams, scripts []shelly.ScriptInfo) {
	table := output.NewTable("ID", "Name", "Enabled", "Running")
	for _, s := range scripts {
		name := s.Name
		if name == "" {
			name = output.FormatPlaceholder("(unnamed)")
		}
		table.AddRow(fmt.Sprintf("%d", s.ID), name, output.RenderYesNo(s.Enable, output.CaseLower, theme.FalseDim), output.RenderRunningState(s.Running))
	}
	printTable(ios, table)
}

// DisplayScheduleList prints a table of schedules.
func DisplayScheduleList(ios *iostreams.IOStreams, schedules []shelly.ScheduleJob) {
	table := output.NewTable("ID", "Enabled", "Timespec", "Calls")
	for _, s := range schedules {
		callsSummary := formatScheduleCallsSummary(s.Calls)
		table.AddRow(fmt.Sprintf("%d", s.ID), output.RenderYesNo(s.Enable, output.CaseLower, theme.FalseDim), s.Timespec, callsSummary)
	}
	printTable(ios, table)
}

func formatScheduleCallsSummary(calls []shelly.ScheduleCall) string {
	if len(calls) == 0 {
		return output.FormatPlaceholder("(none)")
	}

	if len(calls) == 1 {
		call := calls[0]
		if len(call.Params) == 0 {
			return call.Method
		}
		params, err := json.Marshal(call.Params)
		if err != nil {
			return call.Method
		}
		return fmt.Sprintf("%s %s", call.Method, string(params))
	}

	return fmt.Sprintf("%d calls", len(calls))
}

// DisplayWebhookList prints a table of webhooks.
func DisplayWebhookList(ios *iostreams.IOStreams, webhooks []shelly.WebhookInfo) {
	ios.Title("Webhooks")
	ios.Println()

	table := output.NewTable("ID", "Event", "URLs", "Enabled")
	for _, w := range webhooks {
		urls := joinStrings(w.URLs, ", ")
		if len(urls) > 40 {
			urls = urls[:37] + "..."
		}
		table.AddRow(fmt.Sprintf("%d", w.ID), w.Event, urls, output.RenderYesNo(w.Enable, output.CaseTitle, theme.FalseError))
	}
	printTable(ios, table)

	ios.Printf("\n%d webhook(s) configured\n", len(webhooks))
}

// DisplayKVSRaw prints just the raw value from a KVS result.
func DisplayKVSRaw(ios *iostreams.IOStreams, result *shelly.KVSGetResult) {
	switch v := result.Value.(type) {
	case string:
		ios.Println(v)
	case nil:
		ios.Println("null")
	default:
		// For other types (numbers, bools), use JSON encoding
		data, err := json.Marshal(v)
		if err != nil {
			ios.Printf("%v\n", v)
			return
		}
		ios.Println(string(data))
	}
}

// DisplayKVSResult prints a formatted KVS get result.
func DisplayKVSResult(ios *iostreams.IOStreams, key string, result *shelly.KVSGetResult) {
	label := theme.Highlight()
	ios.Printf("%s: %s\n", label.Render("Key"), key)
	ios.Printf("%s: %s\n", label.Render("Value"), output.FormatJSONValue(result.Value))
	ios.Printf("%s: %s\n", label.Render("Type"), output.ValueType(result.Value))
	ios.Printf("%s: %s\n", label.Render("Etag"), result.Etag)
}

// DisplayKVSKeys prints a table of KVS keys.
func DisplayKVSKeys(ios *iostreams.IOStreams, result *shelly.KVSListResult) {
	if len(result.Keys) == 0 {
		ios.NoResults("No keys stored")
		return
	}

	ios.Title("KVS Keys")
	ios.Println()

	table := output.NewTable("Key")
	for _, key := range result.Keys {
		table.AddRow(key)
	}
	printTable(ios, table)

	ios.Printf("\n%d key(s), revision %d\n", len(result.Keys), result.Rev)
}

// DisplayKVSItems prints a table of KVS key-value pairs.
func DisplayKVSItems(ios *iostreams.IOStreams, items []shelly.KVSItem) {
	ios.Title("KVS Data")
	ios.Println()

	table := output.NewTable("Key", "Value", "Type")
	for _, item := range items {
		table.AddRow(item.Key, output.FormatDisplayValue(item.Value), output.ValueType(item.Value))
	}
	printTable(ios, table)

	ios.Printf("\n%d key(s)\n", len(items))
}

// DisplayBLEDevices prints a table of BLE discovered devices.
func DisplayBLEDevices(ios *iostreams.IOStreams, devices []discovery.BLEDiscoveredDevice) {
	if len(devices) == 0 {
		return
	}

	table := output.NewTable("Name", "Address", "Model", "RSSI", "Connectable", "BTHome")

	for _, d := range devices {
		name := d.LocalName
		if name == "" {
			name = d.ID
		}

		// Theme RSSI value (stronger is better)
		rssiStr := fmt.Sprintf("%d dBm", d.RSSI)
		switch {
		case d.RSSI > -50:
			rssiStr = theme.StatusOK().Render(rssiStr)
		case d.RSSI > -70:
			rssiStr = theme.StatusWarn().Render(rssiStr)
		default:
			rssiStr = theme.StatusError().Render(rssiStr)
		}

		// Connectable status
		connStr := output.RenderYesNo(d.Connectable, output.CaseTitle, theme.FalseError)

		// BTHome data indicator
		btHomeStr := "-"
		if d.BTHomeData != nil {
			btHomeStr = theme.StatusInfo().Render("Yes")
		}

		table.AddRow(
			name,
			d.Address.String(),
			d.Model,
			rssiStr,
			connStr,
			btHomeStr,
		)
	}

	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print BLE devices table", err)
	}
	ios.Println("")
	ios.Count("BLE device", len(devices))
}

// DisplayThermostatSchedules displays thermostat schedules with optional details.
func DisplayThermostatSchedules(ios *iostreams.IOStreams, schedules []shelly.ThermostatSchedule, device string, showAll bool) {
	if len(schedules) == 0 {
		if showAll {
			ios.Info("No schedules found on %s", device)
		} else {
			ios.Info("No thermostat schedules found on %s", device)
			ios.Info("Use --all to see all device schedules")
		}
		return
	}

	title := "Thermostat Schedules"
	if showAll {
		title = "All Schedules"
	}
	ios.Println(theme.Bold().Render(fmt.Sprintf("%s on %s:", title, device)))
	ios.Println()

	for _, sched := range schedules {
		ios.Printf("  %s %d\n", theme.Highlight().Render("Schedule"), sched.ID)
		ios.Printf("    Status:   %s\n", output.RenderEnabledState(sched.Enabled))
		ios.Printf("    Timespec: %s\n", sched.Timespec)

		if sched.ThermostatID > 0 {
			ios.Printf("    Thermostat: %d\n", sched.ThermostatID)
		}
		if sched.TargetC != nil {
			ios.Printf("    Target: %.1f°C\n", *sched.TargetC)
		}
		if sched.Mode != "" {
			ios.Printf("    Mode: %s\n", sched.Mode)
		}
		if sched.Enable != nil {
			enableStr := "disable"
			if *sched.Enable {
				enableStr = "enable"
			}
			ios.Printf("    Action: %s thermostat\n", enableStr)
		}
		ios.Println()
	}

	ios.Success("Found %d schedule(s)", len(schedules))
}

// ThermostatScheduleCreateDisplay contains display parameters for schedule creation success.
type ThermostatScheduleCreateDisplay struct {
	Device     string
	ScheduleID int
	Timespec   string
	TargetC    *float64
	Mode       string
	Enable     bool
	Disable    bool
	Enabled    bool
}

// DisplayThermostatScheduleCreate displays the result of creating a thermostat schedule.
func DisplayThermostatScheduleCreate(ios *iostreams.IOStreams, d ThermostatScheduleCreateDisplay) {
	ios.Success("Created schedule %d", d.ScheduleID)
	ios.Printf("  Timespec: %s\n", d.Timespec)

	if d.TargetC != nil {
		ios.Printf("  Target: %.1f°C\n", *d.TargetC)
	}
	if d.Mode != "" {
		ios.Printf("  Mode: %s\n", d.Mode)
	}
	if d.Enable {
		ios.Printf("  Action: enable thermostat\n")
	}
	if d.Disable {
		ios.Printf("  Action: disable thermostat\n")
	}

	if !d.Enabled {
		ios.Info("Schedule is disabled. Enable with: shelly thermostat schedule enable %s --id %d", d.Device, d.ScheduleID)
	}
}
