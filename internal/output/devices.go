// Package output provides formatters for CLI output.
package output

import (
	"fmt"

	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// GetLightLevel returns a human-readable description of a lux value.
func GetLightLevel(lux float64) string {
	switch {
	case lux < 1:
		return "Very dark"
	case lux < 50:
		return "Dark"
	case lux < 200:
		return "Dim"
	case lux < 500:
		return "Indoor light"
	case lux < 1000:
		return "Bright indoor"
	case lux < 10000:
		return "Overcast daylight"
	case lux < 25000:
		return "Daylight"
	default:
		return "Direct sunlight"
	}
}

// MeterReading provides uniform access to power meter values.
type MeterReading interface {
	GetID() int
	GetPower() float64
	GetVoltage() float64
	GetCurrent() float64
}

// FormatDiscoveredDevices builds a table of discovered devices with themed formatting.
// Returns nil if devices slice is empty.
func FormatDiscoveredDevices(devices []discovery.DiscoveredDevice) *Table {
	if len(devices) == 0 {
		return nil
	}

	table := NewTable("ID", "Address", "Model", "Generation", "Protocol", "Auth")

	for _, d := range devices {
		gen := fmt.Sprintf("Gen%d", d.Generation)

		// Theme the auth status
		auth := theme.StatusOK().Render("No")
		if d.AuthRequired {
			auth = theme.StatusWarn().Render("Yes")
		}

		table.AddRow(
			d.ID,
			d.Address.String(),
			d.Model,
			gen,
			string(d.Protocol),
			auth,
		)
	}

	return table
}

// FormatAlarmSensor formats a single alarm sensor reading as a styled string.
// sensorType is "Flood" or "Smoke", alarmMsg is the alarm text like "WATER DETECTED!".
func FormatAlarmSensor(s model.AlarmSensorReading, sensorType, alarmMsg string, okStyle, errorStyle, dimStyle theme.StyleFunc) string {
	status := okStyle("Clear")
	if s.Alarm {
		status = errorStyle(alarmMsg)
	}
	muteStr := ""
	if s.Mute {
		muteStr = " " + dimStyle("(muted)")
	}
	return fmt.Sprintf("    %s Sensor %d: %s%s", sensorType, s.ID, status, muteStr)
}

// FormatAlarmSensors formats multiple alarm sensor readings as styled strings.
// Returns nil if sensors slice is empty.
func FormatAlarmSensors(sensors []model.AlarmSensorReading, sensorType, alarmMsg string, okStyle, errorStyle, dimStyle theme.StyleFunc) []string {
	if len(sensors) == 0 {
		return nil
	}
	lines := make([]string, len(sensors))
	for i, s := range sensors {
		lines[i] = FormatAlarmSensor(s, sensorType, alarmMsg, okStyle, errorStyle, dimStyle)
	}
	return lines
}

// FormatComponentName returns the component name or a fallback "{type}:{id}".
func FormatComponentName(name, componentType string, id int) string {
	if name != "" {
		return name
	}
	return fmt.Sprintf("%s:%d", componentType, id)
}

// FormatPower returns a human-readable power string (W or kW).
func FormatPower(watts float64) string {
	if watts >= 1000 {
		return fmt.Sprintf("%.2f kW", watts/1000)
	}
	return fmt.Sprintf("%.1f W", watts)
}

// FormatEnergy returns a human-readable energy string (Wh or kWh).
func FormatEnergy(wh float64) string {
	if wh >= 1000 {
		return fmt.Sprintf("%.2f kWh", wh/1000)
	}
	return fmt.Sprintf("%.0f Wh", wh)
}

// FormatPowerColored returns a themed power string based on usage level.
// Red for >=1000W, yellow for >=100W, green otherwise.
func FormatPowerColored(watts float64) string {
	s := FormatPower(watts)
	if watts >= 1000 {
		return theme.StatusError().Render(s)
	} else if watts >= 100 {
		return theme.StatusWarn().Render(s)
	}
	return theme.StatusOK().Render(s)
}

// FormatPowerTableValue returns formatted power string or "-" if zero.
// Use for table cell display where zero values should show placeholder.
func FormatPowerTableValue(power float64) string {
	if power > 0 {
		return fmt.Sprintf("%.1f W", power)
	}
	return "-"
}

// FormatPowerWithChange formats power value with directional change indicator.
// Shows ↑ (warning) for increases, ↓ (ok) for decreases.
func FormatPowerWithChange(power float64, prevPower *float64) string {
	formatted := FormatPower(power)
	if prevPower == nil {
		return formatted
	}
	if power > *prevPower {
		return theme.StatusWarn().Render(formatted + " ↑")
	} else if power < *prevPower {
		return theme.StatusOK().Render(formatted + " ↓")
	}
	return formatted
}

// FormatMeterLine formats a single meter reading with optional change indicator.
// Optional pf (power factor) is displayed if non-nil.
func FormatMeterLine(label string, id int, power, voltage, current float64, pf, prevPower *float64) string {
	powerStr := FormatPowerWithChange(power, prevPower)
	pfStr := ""
	if pf != nil {
		pfStr = fmt.Sprintf("  PF:%.2f", *pf)
	}
	return fmt.Sprintf("  %s %d: %s  %.1fV  %.2fA%s", label, id, powerStr, voltage, current, pfStr)
}

// FormatMeterLineWithEnergy formats a meter line including energy total.
// Optional pf (power factor) is displayed if non-nil.
func FormatMeterLineWithEnergy(label string, id int, power, voltage, current float64, pf, energy, prevPower *float64) string {
	base := FormatMeterLine(label, id, power, voltage, current, pf, prevPower)
	if energy != nil {
		return fmt.Sprintf("%s  %.2f Wh", base, *energy)
	}
	return base
}

// FormatEMPhase formats a single phase line for 3-phase EM.
// Optional pf (power factor) is displayed if non-nil.
func FormatEMPhase(label string, power, voltage, current float64, pf, prevPower *float64) string {
	powerStr := FormatPowerWithChange(power, prevPower)
	pfStr := ""
	if pf != nil {
		pfStr = fmt.Sprintf("  PF:%.2f", *pf)
	}
	return fmt.Sprintf("    %s: %s  %.1fV  %.2fA%s", label, powerStr, voltage, current, pfStr)
}

// FormatEMLines returns formatted lines for a 3-phase energy meter.
func FormatEMLines(em, prev *shelly.EMStatus) []string {
	var prevA, prevB, prevC *float64
	if prev != nil {
		prevA = &prev.AActivePower
		prevB = &prev.BActivePower
		prevC = &prev.CActivePower
	}

	return []string{
		fmt.Sprintf("  EM %d:", em.ID),
		FormatEMPhase("Phase A", em.AActivePower, em.AVoltage, em.ACurrent, em.APowerFactor, prevA),
		FormatEMPhase("Phase B", em.BActivePower, em.BVoltage, em.BCurrent, em.BPowerFactor, prevB),
		FormatEMPhase("Phase C", em.CActivePower, em.CVoltage, em.CCurrent, em.CPowerFactor, prevC),
		fmt.Sprintf("    Total:   %.1f W", em.TotalActivePower),
	}
}

// FormatEM1Line returns formatted line for a single-phase energy meter.
func FormatEM1Line(em1, prev *shelly.EM1Status) string {
	var prevPower *float64
	if prev != nil {
		prevPower = &prev.ActPower
	}
	return FormatMeterLine("EM1", em1.ID, em1.ActPower, em1.Voltage, em1.Current, em1.PF, prevPower)
}

// FormatPMLine returns formatted line for a power meter.
func FormatPMLine(pm, prev *shelly.PMStatus) string {
	var prevPower *float64
	if prev != nil {
		prevPower = &prev.APower
	}
	var energy *float64
	if pm.AEnergy != nil {
		energy = &pm.AEnergy.Total
	}
	return FormatMeterLineWithEnergy("PM", pm.ID, pm.APower, pm.Voltage, pm.Current, nil, energy, prevPower)
}

// FindPrevious is a generic finder for previous status by ID.
func FindPrevious[T any](id int, items []T, getID func(*T) int) *T {
	for i := range items {
		if getID(&items[i]) == id {
			return &items[i]
		}
	}
	return nil
}

// FindPreviousEM finds matching EM status by ID from previous snapshot.
func FindPreviousEM(id int, prev *shelly.MonitoringSnapshot) *shelly.EMStatus {
	if prev == nil {
		return nil
	}
	return FindPrevious(id, prev.EM, func(e *shelly.EMStatus) int { return e.ID })
}

// FindPreviousEM1 finds matching EM1 status by ID from previous snapshot.
func FindPreviousEM1(id int, prev *shelly.MonitoringSnapshot) *shelly.EM1Status {
	if prev == nil {
		return nil
	}
	return FindPrevious(id, prev.EM1, func(e *shelly.EM1Status) int { return e.ID })
}

// FindPreviousPM finds matching PM status by ID from previous snapshot.
func FindPreviousPM(id int, prev *shelly.MonitoringSnapshot) *shelly.PMStatus {
	if prev == nil {
		return nil
	}
	return FindPrevious(id, prev.PM, func(e *shelly.PMStatus) int { return e.ID })
}

// GetPrevEMPhasePower returns previous phase power values for an EM by ID.
// Returns nil pointers if previous snapshot is nil or EM not found.
func GetPrevEMPhasePower(id int, prev *shelly.MonitoringSnapshot) (prevA, prevB, prevC *float64) {
	em := FindPreviousEM(id, prev)
	if em == nil {
		return nil, nil, nil
	}
	return &em.AActivePower, &em.BActivePower, &em.CActivePower
}

// GetPrevEM1Power returns previous power value for an EM1 by ID.
func GetPrevEM1Power(id int, prev *shelly.MonitoringSnapshot) *float64 {
	em1 := FindPreviousEM1(id, prev)
	if em1 == nil {
		return nil
	}
	return &em1.ActPower
}

// GetPrevPMPower returns previous power value for a PM by ID.
func GetPrevPMPower(id int, prev *shelly.MonitoringSnapshot) *float64 {
	pm := FindPreviousPM(id, prev)
	if pm == nil {
		return nil
	}
	return &pm.APower
}

// CalculateSnapshotTotals calculates total power and energy from a monitoring snapshot.
func CalculateSnapshotTotals(snapshot *shelly.MonitoringSnapshot) (totalPower, totalEnergy float64) {
	if snapshot == nil {
		return 0, 0
	}

	for i := range snapshot.EM {
		totalPower += snapshot.EM[i].TotalActivePower
	}
	for i := range snapshot.EM1 {
		totalPower += snapshot.EM1[i].ActPower
	}
	for i := range snapshot.PM {
		pm := &snapshot.PM[i]
		totalPower += pm.APower
		if pm.AEnergy != nil {
			totalEnergy += pm.AEnergy.Total
		}
	}

	return totalPower, totalEnergy
}
