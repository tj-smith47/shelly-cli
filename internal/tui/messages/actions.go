// Package messages provides shared message types for TUI components.
package messages

import (
	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/tui/focus"
)

// IsActionRequest returns true if the message is an action request that should
// be forwarded only to the focused component, not all components.
func IsActionRequest(msg tea.Msg) bool {
	switch msg.(type) {
	case EditRequestMsg, NewRequestMsg, DeleteRequestMsg,
		CopyRequestMsg, PauseRequestMsg, ClearRequestMsg,
		ExpandRequestMsg, FilterToggleRequestMsg, SortRequestMsg,
		RefreshRequestMsg, NavigationMsg,
		RunRequestMsg, StopRequestMsg, TestRequestMsg,
		ScanRequestMsg, ExportRequestMsg, ImportRequestMsg,
		ActivateRequestMsg, ViewRequestMsg, CaptureRequestMsg,
		ToggleEnableRequestMsg, SelectRequestMsg, SelectAllRequestMsg,
		BackupRequestMsg, RestoreRequestMsg, UpdateRequestMsg,
		UpdateAllRequestMsg, StagedUpdateRequestMsg, RollbackRequestMsg,
		CheckRequestMsg, TimezoneRequestMsg, ActionsRequestMsg,
		MQTTRequestMsg, AuthRequestMsg, CloudRequestMsg,
		ResetRequestMsg, DownloadRequestMsg, UploadRequestMsg,
		EvalRequestMsg, SnoozeRequestMsg, ModeSelectMsg:
		return true
	default:
		return false
	}
}

// Action request messages - sent from app.go to components when context system matches a key.
// Components should handle these instead of checking raw key strings.

// EditRequestMsg requests the component to edit the currently selected item.
type EditRequestMsg struct{}

// NewRequestMsg requests the component to create a new item.
type NewRequestMsg struct{}

// DeleteRequestMsg requests the component to delete the currently selected item.
type DeleteRequestMsg struct{}

// CopyRequestMsg requests the component to copy content (e.g., JSON viewer).
type CopyRequestMsg struct{}

// PauseRequestMsg requests the component to pause/resume (e.g., event stream).
type PauseRequestMsg struct{}

// ClearRequestMsg requests the component to clear its content (e.g., event list).
type ClearRequestMsg struct{}

// ExpandRequestMsg requests the component to expand/collapse or select all.
type ExpandRequestMsg struct{}

// FilterToggleRequestMsg requests the component to toggle its filter.
type FilterToggleRequestMsg struct{}

// SortRequestMsg requests the component to sort its content.
type SortRequestMsg struct{}

// RefreshRequestMsg requests the component to refresh its data.
type RefreshRequestMsg struct{}

// NavDirection represents a navigation direction.
type NavDirection int

// Navigation direction constants.
const (
	NavUp NavDirection = iota
	NavDown
	NavLeft
	NavRight
	NavPageUp
	NavPageDown
	NavHome
	NavEnd
)

// NavigationMsg is a unified navigation message for components.
// Components can handle this instead of checking individual navigation keys.
type NavigationMsg struct {
	Direction NavDirection
}

// Component-specific action messages for keys that don't map to generic actions.

// RunRequestMsg requests running/starting something (e.g., script).
type RunRequestMsg struct{}

// StopRequestMsg requests stopping something (e.g., script).
type StopRequestMsg struct{}

// TestRequestMsg requests testing something (e.g., webhook).
type TestRequestMsg struct{}

// ScanRequestMsg requests scanning (e.g., WiFi networks).
type ScanRequestMsg struct{}

// ExportRequestMsg requests exporting data.
type ExportRequestMsg struct{}

// ImportRequestMsg requests importing data.
type ImportRequestMsg struct{}

// ActivateRequestMsg requests activating something (e.g., scene).
type ActivateRequestMsg struct{}

// ViewRequestMsg requests viewing details of something.
type ViewRequestMsg struct{}

// CaptureRequestMsg requests capturing current state (e.g., scene from devices).
type CaptureRequestMsg struct{}

// ToggleEnableRequestMsg requests toggling enable/disable state.
type ToggleEnableRequestMsg struct{}

// SelectRequestMsg requests selecting/deselecting an item (space key).
type SelectRequestMsg struct{}

// SelectAllRequestMsg requests selecting all items.
type SelectAllRequestMsg struct{}

// BackupRequestMsg requests creating a backup.
type BackupRequestMsg struct{}

// RestoreRequestMsg requests restoring from backup.
type RestoreRequestMsg struct{}

// UpdateRequestMsg requests updating (e.g., firmware).
type UpdateRequestMsg struct{}

// UpdateAllRequestMsg requests updating all items.
type UpdateAllRequestMsg struct{}

// StagedUpdateRequestMsg requests a staged/gradual update.
type StagedUpdateRequestMsg struct{}

// RollbackRequestMsg requests rolling back (e.g., firmware).
type RollbackRequestMsg struct{}

// CheckRequestMsg requests checking for updates.
type CheckRequestMsg struct{}

// TimezoneRequestMsg requests opening timezone editor.
type TimezoneRequestMsg struct{}

// ActionsRequestMsg requests opening actions editor (e.g., input actions).
type ActionsRequestMsg struct{}

// MQTTRequestMsg requests opening MQTT configuration.
type MQTTRequestMsg struct{}

// AuthRequestMsg requests opening auth configuration.
type AuthRequestMsg struct{}

// CloudRequestMsg requests cloud toggle/configuration.
type CloudRequestMsg struct{}

// ResetRequestMsg requests resetting something.
type ResetRequestMsg struct{}

// DownloadRequestMsg requests downloading something.
type DownloadRequestMsg struct{}

// UploadRequestMsg requests uploading something.
type UploadRequestMsg struct{}

// EvalRequestMsg requests evaluating code.
type EvalRequestMsg struct{}

// SnoozeRequestMsg requests snoozing (e.g., alert).
type SnoozeRequestMsg struct {
	Duration string // "1h" or "24h"
}

// ModeSelectMsg requests selecting a mode (e.g., operation mode 1 or 2).
type ModeSelectMsg struct {
	Mode int
}

// Overlay/Modal coordination messages - used to synchronize focus state between
// app.go and views/components when modals are opened or closed.

// ModalOpenedMsg notifies the app that a modal/overlay was opened.
// Views/components emit this when they open edit modals, dialogs, etc.
type ModalOpenedMsg struct {
	ID   focus.OverlayID
	Mode focus.Mode
}

// ModalClosedMsg notifies the app that a modal/overlay was closed.
// Views/components emit this when their modals are dismissed.
type ModalClosedMsg struct {
	ID focus.OverlayID
}
