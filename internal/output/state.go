// Package output provides state rendering helpers for consistent output formatting.
package output

import (
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// isTTY is a package-level function for TTY detection, overridable in tests.
var isTTY = func() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

// colorEnabled returns true if colored output should be used.
// Returns false for --no-color, --plain, NO_COLOR env, or non-TTY.
func colorEnabled() bool {
	if !isTTY() {
		return false
	}
	if viper.GetBool("plain") || viper.GetBool("no-color") {
		return false
	}
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	if _, ok := os.LookupEnv("SHELLY_NO_COLOR"); ok {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	return true
}

// Case defines text casing for boolean labels.
type Case int

const (
	// CaseLower uses lowercase labels (on/off, yes/no, active/inactive).
	CaseLower Case = iota
	// CaseTitle uses title case labels (On/Off, Yes/No, Active/Inactive).
	CaseTitle
	// CaseUpper uses uppercase labels (ON/OFF). Only valid for OnOff family.
	CaseUpper
)

// State label constants for consistent display text.
const (
	// On/Off states (all casings).
	LabelOnUpper  = "ON"
	LabelOffUpper = "OFF"
	LabelOnTitle  = "On"
	LabelOffTitle = "Off"
	LabelOnLower  = "on"
	LabelOffLower = "off"

	// Active/Inactive states (all casings).
	LabelActiveTitle   = "Active"
	LabelInactiveTitle = "Inactive"
	LabelActiveLower   = "active"
	LabelInactiveLower = "inactive"

	// Enabled/Disabled states.
	LabelEnabled  = "Enabled"
	LabelDisabled = "Disabled"

	// Yes/No states (all casings).
	LabelYesTitle = "Yes"
	LabelNoTitle  = "No"
	LabelYesLower = "yes"
	LabelNoLower  = "no"

	// Online/Offline states (all casings).
	LabelOnlineTitle  = "Online"
	LabelOfflineTitle = "Offline"
	LabelOnlineLower  = "online"
	LabelOfflineLower = "offline"

	// Running/Stopped states.
	LabelRunning = "running"
	LabelStopped = "stopped"

	// Available state.
	LabelAvailable = "available"

	// Input states.
	LabelTriggered = "triggered"
	LabelIdle      = "idle"

	// Error state.
	LabelError = "error"

	// Valve states.
	LabelValveOpen   = "Open (heating)"
	LabelValveClosed = "Closed"

	// Network states.
	LabelJoined = "joined"

	// Cover states.
	LabelCoverOpen   = "open"
	LabelCoverClosed = "closed"

	// Action count.
	LabelEmpty   = "0 (empty)"
	LabelAction  = "1 action"
	LabelActions = "%d actions"

	// Placeholder.
	LabelPlaceholder = "-"

	// Boolean values.
	LabelTrue  = "true"
	LabelFalse = "false"

	// Cloud/Auth states.
	LabelLoggedIn    = "Logged in"
	LabelNotLoggedIn = "Not logged in"
	LabelValid       = "Valid"
	LabelInvalid     = "Invalid"
	LabelExpired     = "Expired"

	// Firmware update status.
	LabelUpdateAvailable = "(update available)"
	LabelUpToDate        = "(up to date)"

	// Alarm states.
	LabelClear   = "Clear"
	LabelMuted   = "Muted"
	LabelMutedLC = "(muted)"

	// Diff section labels.
	LabelRemoved = "Removed (only in source):"
	LabelAdded   = "Added (only in target):"
	LabelChanged = "Changed:"
)

// =============================================================================
// Consolidated Boolean Renderers
// =============================================================================

// RenderOnOff returns a themed on/off string with configurable case and false style.
// Respects --no-color, --plain, and NO_COLOR environment variable.
func RenderOnOff(on bool, c Case, fs theme.FalseStyle) string {
	var trueLabel, falseLabel string
	switch c {
	case CaseUpper:
		trueLabel, falseLabel = LabelOnUpper, LabelOffUpper
	case CaseTitle:
		trueLabel, falseLabel = LabelOnTitle, LabelOffTitle
	default:
		trueLabel, falseLabel = LabelOnLower, LabelOffLower
	}
	if !colorEnabled() {
		if on {
			return trueLabel
		}
		return falseLabel
	}
	if on {
		return theme.StatusOK().Render(trueLabel)
	}
	return fs.Render(falseLabel)
}

// RenderYesNo returns a themed yes/no string with configurable case and false style.
func RenderYesNo(value bool, c Case, fs theme.FalseStyle) string {
	var trueLabel, falseLabel string
	switch c {
	case CaseTitle, CaseUpper: // Yes/No only has Title and Lower
		trueLabel, falseLabel = LabelYesTitle, LabelNoTitle
	default:
		trueLabel, falseLabel = LabelYesLower, LabelNoLower
	}
	if value {
		return theme.StatusOK().Render(trueLabel)
	}
	return fs.Render(falseLabel)
}

// RenderOnline returns a themed online/offline string with configurable case.
// Uses dedicated Online/Offline semantic colors for device state consistency.
// Respects --no-color, --plain, and NO_COLOR environment variable.
func RenderOnline(online bool, c Case) string {
	var trueLabel, falseLabel string
	switch c {
	case CaseTitle, CaseUpper: // Online only has Title and Lower
		trueLabel, falseLabel = LabelOnlineTitle, LabelOfflineTitle
	default:
		trueLabel, falseLabel = LabelOnlineLower, LabelOfflineLower
	}
	if !colorEnabled() {
		if online {
			return trueLabel
		}
		return falseLabel
	}
	if online {
		return theme.StatusOnline().Render(trueLabel)
	}
	return theme.StatusOffline().Render(falseLabel)
}

// RenderActive returns a themed active/inactive string with configurable case and false style.
func RenderActive(active bool, c Case, fs theme.FalseStyle) string {
	var trueLabel, falseLabel string
	switch c {
	case CaseTitle, CaseUpper: // Active only has Title and Lower
		trueLabel, falseLabel = LabelActiveTitle, LabelInactiveTitle
	default:
		trueLabel, falseLabel = LabelActiveLower, LabelInactiveLower
	}
	if active {
		return theme.StatusOK().Render(trueLabel)
	}
	return fs.Render(falseLabel)
}

// =============================================================================
// Other Renderers (not part of the consolidated families)
// =============================================================================

// RenderBoolState returns a themed string with custom labels.
func RenderBoolState(value bool, trueLabel, falseLabel string) string {
	if value {
		return theme.StatusOK().Render(trueLabel)
	}
	return theme.StatusError().Render(falseLabel)
}

// RenderEnabledState returns themed "Enabled" or "Disabled" string.
func RenderEnabledState(enabled bool) string {
	if enabled {
		return theme.StatusOK().Render(LabelEnabled)
	}
	return theme.Dim().Render(LabelDisabled)
}

// RenderAvailableState returns themed "available" or custom unavailable text.
func RenderAvailableState(available bool, unavailableText string) string {
	if available {
		return theme.StatusOK().Render(LabelAvailable)
	}
	return theme.Dim().Render(unavailableText)
}

// RenderCoverState returns a themed cover state string.
func RenderCoverState(state string) string {
	switch state {
	case LabelCoverOpen:
		return theme.StatusOK().Render(state)
	case LabelCoverClosed:
		return theme.StatusError().Render(state)
	default:
		return theme.StatusWarn().Render(state)
	}
}

// RenderRunningState returns themed "running" or dimmed "stopped" string.
func RenderRunningState(running bool) string {
	if running {
		return theme.StatusOK().Render(LabelRunning)
	}
	return theme.Dim().Render(LabelStopped)
}

// RenderNetworkState returns themed network state (joined = green, others = dim).
func RenderNetworkState(state string) string {
	if state == LabelJoined {
		return theme.StatusOK().Render(state)
	}
	return theme.Dim().Render(state)
}

// RenderInputTriggeredState returns themed "triggered" or "idle" string.
func RenderInputTriggeredState(triggered bool) string {
	if triggered {
		return theme.StatusWarn().Render(LabelTriggered)
	}
	return theme.Dim().Render(LabelIdle)
}

// RenderErrorState returns themed "error" string.
func RenderErrorState() string {
	return theme.StatusError().Render(LabelError)
}

// RenderOnOffStateWithBrightness returns themed "ON (X%)" or dimmed "off" string.
func RenderOnOffStateWithBrightness(on bool, brightness *int) string {
	if on {
		if brightness != nil && *brightness > 0 {
			return theme.StatusOK().Render(fmt.Sprintf("%s (%d%%)", LabelOnUpper, *brightness))
		}
		return theme.StatusOK().Render(LabelOnUpper)
	}
	return theme.Dim().Render(LabelOffLower)
}

// RenderValveState returns themed "Open (heating)" or dimmed "Closed" string.
func RenderValveState(open bool) string {
	if open {
		return theme.StatusOK().Render(LabelValveOpen)
	}
	return theme.Dim().Render(LabelValveClosed)
}

// RenderLoggedInState returns themed "Logged in" or "Not logged in" string.
func RenderLoggedInState(loggedIn bool) string {
	if loggedIn {
		return theme.StatusOK().Render(LabelLoggedIn)
	}
	return theme.StatusError().Render(LabelNotLoggedIn)
}

// RenderTokenValidity returns themed token validity state.
// Returns "Valid", "Expired", or "Invalid".
func RenderTokenValidity(valid, expired bool) string {
	switch {
	case !valid:
		return theme.StatusError().Render(LabelInvalid)
	case expired:
		return theme.StatusError().Render(LabelExpired)
	default:
		return theme.StatusOK().Render(LabelValid)
	}
}

// RenderUpdateStatus returns themed firmware update status string.
func RenderUpdateStatus(hasUpdate bool) string {
	if hasUpdate {
		return theme.StatusOK().Render(LabelUpdateAvailable)
	}
	return theme.Dim().Render(LabelUpToDate)
}

// RenderAlarmState returns themed "Clear" or custom alarm text.
func RenderAlarmState(alarm bool, alarmText string) string {
	if alarm {
		return theme.StatusError().Render(alarmText)
	}
	return theme.StatusOK().Render(LabelClear)
}

// RenderMuteState returns themed "Muted" or "Active" string.
func RenderMuteState(muted bool) string {
	if muted {
		return theme.Dim().Render(LabelMuted)
	}
	return theme.Highlight().Render("Active")
}

// RenderMuteAnnotation returns dimmed "(muted)" annotation or empty string.
func RenderMuteAnnotation(muted bool) string {
	if muted {
		return " " + theme.Dim().Render(LabelMutedLC)
	}
	return ""
}

// RenderDiffRemoved returns themed diff "Removed" section header.
func RenderDiffRemoved() string {
	return theme.StatusError().Render(LabelRemoved)
}

// RenderDiffAdded returns themed diff "Added" section header.
func RenderDiffAdded() string {
	return theme.StatusOK().Render(LabelAdded)
}

// RenderDiffChanged returns themed diff "Changed" section header.
func RenderDiffChanged() string {
	return theme.StatusWarn().Render(LabelChanged)
}

// RenderAuthRequired returns themed "Yes" (warning) or "No" (ok) for auth status.
// Respects --no-color, --plain, and NO_COLOR environment variable.
func RenderAuthRequired(required bool) string {
	if required {
		if colorEnabled() {
			return theme.StatusWarn().Render(LabelYesTitle)
		}
		return LabelYesTitle
	}
	if colorEnabled() {
		return theme.StatusOK().Render(LabelNoTitle)
	}
	return LabelNoTitle
}

// RenderGeneration returns generation string (e.g., "Gen2") or "unknown".
// Returns plain text to allow table styling to take effect.
func RenderGeneration(gen int) string {
	if gen == 0 {
		return "unknown"
	}
	return fmt.Sprintf("Gen%d", gen)
}

// Component state formatters

// RenderSwitchState returns the state string for a switch component.
func RenderSwitchState(status *model.SwitchStatus) string {
	return RenderOnOff(status.Output, CaseUpper, theme.FalseDim)
}

// RenderLightState returns the state string for a light component.
func RenderLightState(status *model.LightStatus) string {
	return RenderOnOffStateWithBrightness(status.Output, status.Brightness)
}

// RenderRGBState returns the state string for an RGB component.
func RenderRGBState(status *model.RGBStatus) string {
	return RenderOnOffStateWithBrightness(status.Output, status.Brightness)
}

// RenderRGBWState returns the state string for an RGBW component.
func RenderRGBWState(status *model.RGBWStatus) string {
	return RenderOnOffStateWithBrightness(status.Output, status.Brightness)
}

// RenderCoverStatusState returns the state string for a cover component.
// Note: This is different from RenderCoverState which takes a state string.
func RenderCoverStatusState(status *model.CoverStatus) string {
	state := status.State
	if status.CurrentPosition != nil && *status.CurrentPosition >= 0 {
		state = fmt.Sprintf("%s (%d%%)", status.State, *status.CurrentPosition)
	}
	return state
}

// RenderInputState returns the state string for an input component.
func RenderInputState(status *model.InputStatus) string {
	return RenderInputTriggeredState(status.State)
}

// StatusField represents a field in a status display with label and formatted value.
type StatusField struct {
	Label string
	Value string
}

// FormatSwitchStatusFields returns the status fields for a switch component.
func FormatSwitchStatusFields(status *model.SwitchStatus) []StatusField {
	fields := []StatusField{
		{Label: "State", Value: RenderOnOff(status.Output, CaseUpper, theme.FalseError)},
	}
	fields = appendPowerFields(fields, status.Power, status.Voltage, status.Current)
	if status.Energy != nil {
		fields = append(fields, StatusField{Label: "Energy", Value: fmt.Sprintf("%.2f Wh", status.Energy.Total)})
	}
	return fields
}

// FormatLightStatusFields returns the status fields for a light component.
func FormatLightStatusFields(status *model.LightStatus) []StatusField {
	fields := []StatusField{
		{Label: "State", Value: RenderOnOff(status.Output, CaseUpper, theme.FalseError)},
	}
	if status.Brightness != nil {
		fields = append(fields, StatusField{Label: "Brightness", Value: fmt.Sprintf("%d%%", *status.Brightness)})
	}
	fields = appendPowerFields(fields, status.Power, status.Voltage, status.Current)
	return fields
}

// FormatRGBStatusFields returns the status fields for an RGB component.
func FormatRGBStatusFields(status *model.RGBStatus) []StatusField {
	fields := []StatusField{
		{Label: "State", Value: RenderOnOff(status.Output, CaseUpper, theme.FalseError)},
	}
	if status.RGB != nil {
		fields = append(fields, StatusField{Label: "Color", Value: fmt.Sprintf("R:%d G:%d B:%d", status.RGB.Red, status.RGB.Green, status.RGB.Blue)})
	}
	if status.Brightness != nil {
		fields = append(fields, StatusField{Label: "Brightness", Value: fmt.Sprintf("%d%%", *status.Brightness)})
	}
	fields = appendPowerFields(fields, status.Power, status.Voltage, status.Current)
	return fields
}

// FormatRGBWStatusFields returns the status fields for an RGBW component.
func FormatRGBWStatusFields(status *model.RGBWStatus) []StatusField {
	fields := []StatusField{
		{Label: "State", Value: RenderOnOff(status.Output, CaseUpper, theme.FalseError)},
	}
	if status.RGB != nil {
		fields = append(fields, StatusField{Label: "Color", Value: fmt.Sprintf("R:%d G:%d B:%d", status.RGB.Red, status.RGB.Green, status.RGB.Blue)})
	}
	if status.White != nil {
		fields = append(fields, StatusField{Label: "White", Value: fmt.Sprintf("%d", *status.White)})
	}
	if status.Brightness != nil {
		fields = append(fields, StatusField{Label: "Brightness", Value: fmt.Sprintf("%d%%", *status.Brightness)})
	}
	fields = appendPowerFields(fields, status.Power, status.Voltage, status.Current)
	return fields
}

// FormatCoverStatusFields returns the status fields for a cover component.
func FormatCoverStatusFields(status *model.CoverStatus) []StatusField {
	fields := []StatusField{
		{Label: "State", Value: RenderCoverState(status.State)},
	}
	if status.CurrentPosition != nil {
		fields = append(fields, StatusField{Label: "Position", Value: fmt.Sprintf("%d%%", *status.CurrentPosition)})
	}
	fields = appendPowerFields(fields, status.Power, status.Voltage, status.Current)
	return fields
}

// FormatInputStatusFields returns the status fields for an input component.
func FormatInputStatusFields(status *model.InputStatus) []StatusField {
	fields := []StatusField{
		{Label: "State", Value: RenderActive(status.State, CaseLower, theme.FalseError)},
	}
	if status.Type != "" {
		fields = append(fields, StatusField{Label: "Type", Value: status.Type})
	}
	return fields
}

// appendPowerFields appends power, voltage, and current fields if present.
func appendPowerFields(fields []StatusField, power, voltage, current *float64) []StatusField {
	if power != nil {
		fields = append(fields, StatusField{Label: "Power", Value: fmt.Sprintf("%.1f W", *power)})
	}
	if voltage != nil {
		fields = append(fields, StatusField{Label: "Voltage", Value: fmt.Sprintf("%.1f V", *voltage)})
	}
	if current != nil {
		fields = append(fields, StatusField{Label: "Current", Value: fmt.Sprintf("%.3f A", *current)})
	}
	return fields
}
