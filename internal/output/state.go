// Package output provides state rendering helpers for consistent output formatting.
package output

import (
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// State label constants for consistent display text.
const (
	// On/Off states.
	LabelOn    = "ON"
	LabelOff   = "OFF"
	LabelOnLC  = "on"  // lowercase variant for dim styling
	LabelOffLC = "off" // lowercase variant for dim styling

	// Active/Inactive states.
	LabelActive   = "Active"
	LabelInactive = "inactive"
	LabelActiveLC = "active" // lowercase
	LabelOffTitle = "Off"    // title case for thermostat inactive

	// Enabled/Disabled states.
	LabelEnabled  = "Enabled"
	LabelDisabled = "Disabled"

	// Yes/No states.
	LabelYes   = "Yes"
	LabelNo    = "No"
	LabelYesLC = "yes" // lowercase
	LabelNoLC  = "no"  // lowercase

	// Online/Offline states.
	LabelOnline  = "online"
	LabelOffline = "offline"

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

	// Cloud/Auth states.
	LabelLoggedIn    = "Logged in"
	LabelNotLoggedIn = "Not logged in"
	LabelValid       = "Valid"
	LabelInvalid     = "Invalid"
	LabelExpired     = "Expired"

	// Online/Offline title case variants.
	LabelOnlineTitle  = "Online"
	LabelOfflineTitle = "Offline"

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

// RenderOnOffState returns a themed "ON" or "OFF" string based on state.
func RenderOnOffState(on bool) string {
	if on {
		return theme.StatusOK().Render(LabelOn)
	}
	return theme.StatusError().Render(LabelOff)
}

// RenderActiveState returns a themed "active" or "inactive" string.
func RenderActiveState(active bool) string {
	if active {
		return theme.StatusOK().Render(LabelActiveLC)
	}
	return theme.StatusError().Render(LabelInactive)
}

// RenderBoolState returns a themed string with custom labels.
func RenderBoolState(value bool, trueLabel, falseLabel string) string {
	if value {
		return theme.StatusOK().Render(trueLabel)
	}
	return theme.StatusError().Render(falseLabel)
}

// RenderEnabledState returns themed "Enabled" or "Disabled" string.
// Uses Dim style for disabled (less alarming than error style).
func RenderEnabledState(enabled bool) string {
	if enabled {
		return theme.StatusOK().Render(LabelEnabled)
	}
	return theme.Dim().Render(LabelDisabled)
}

// RenderAvailableState returns themed "available" or custom unavailable text.
// Uses Dim style for unavailable (less alarming than error style).
func RenderAvailableState(available bool, unavailableText string) string {
	if available {
		return theme.StatusOK().Render(LabelAvailable)
	}
	return theme.Dim().Render(unavailableText)
}

// RenderCoverState returns a themed cover state string.
// States: open (green), closed (red), opening/closing/stopped (yellow).
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

// FormatComponentName returns the component name or a fallback "{type}:{id}".
func FormatComponentName(name, componentType string, id int) string {
	if name != "" {
		return name
	}
	return fmt.Sprintf("%s:%d", componentType, id)
}

// FormatPowerValue returns formatted power string or "-" if zero.
func FormatPowerValue(power float64) string {
	if power > 0 {
		return fmt.Sprintf("%.1f W", power)
	}
	return "-"
}

// RenderYesNo returns themed "Yes" or "No" string.
func RenderYesNo(value bool) string {
	if value {
		return theme.StatusOK().Render(LabelYes)
	}
	return theme.StatusError().Render(LabelNo)
}

// RenderYesNoDim returns themed "yes" or dimmed "no" string.
// Useful when "no" is the normal/default state.
func RenderYesNoDim(value bool) string {
	if value {
		return theme.StatusOK().Render(LabelYesLC)
	}
	return theme.Dim().Render(LabelNoLC)
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

// FormatPlaceholder returns dimmed placeholder text.
func FormatPlaceholder(text string) string {
	return theme.Dim().Render(text)
}

// FormatActionCount returns themed action count string.
func FormatActionCount(count int) string {
	if count == 0 {
		return theme.StatusWarn().Render(LabelEmpty)
	}
	if count == 1 {
		return theme.StatusOK().Render(LabelAction)
	}
	return theme.StatusOK().Render(fmt.Sprintf(LabelActions, count))
}

// RenderOnlineState returns themed "online" or "offline" string.
func RenderOnlineState(online bool) string {
	if online {
		return theme.StatusOK().Render(LabelOnline)
	}
	return theme.StatusError().Render(LabelOffline)
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

// RenderOnOffStateDim returns themed "ON" or dimmed "off" string.
// Useful when OFF is the normal/default state.
func RenderOnOffStateDim(on bool) string {
	if on {
		return theme.StatusOK().Render(LabelOn)
	}
	return theme.Dim().Render(LabelOffLC)
}

// RenderOnOffStateWithBrightness returns themed "ON (X%)" or dimmed "off" string.
func RenderOnOffStateWithBrightness(on bool, brightness *int) string {
	if on {
		if brightness != nil && *brightness > 0 {
			return theme.StatusOK().Render(fmt.Sprintf("%s (%d%%)", LabelOn, *brightness))
		}
		return theme.StatusOK().Render(LabelOn)
	}
	return theme.Dim().Render(LabelOffLC)
}

// RenderActiveStateDim returns themed "Active" or dimmed "Off" string.
// Useful for thermostat and similar active/inactive states.
func RenderActiveStateDim(active bool) string {
	if active {
		return theme.StatusOK().Render(LabelActive)
	}
	return theme.Dim().Render(LabelOffTitle)
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

// RenderOnlineStateTitle returns themed "Online" or "Offline" string (title case).
func RenderOnlineStateTitle(online bool) string {
	if online {
		return theme.StatusOK().Render(LabelOnlineTitle)
	}
	return theme.StatusError().Render(LabelOfflineTitle)
}

// RenderYesNoLC returns themed lowercase "yes" or "no" string with error style for no.
func RenderYesNoLC(value bool) string {
	if value {
		return theme.StatusOK().Render(LabelYesLC)
	}
	return theme.StatusError().Render(LabelNoLC)
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
