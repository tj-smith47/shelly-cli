package smarthome

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/wireless"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
)

func TestNewZWaveEditModel(t *testing.T) {
	t.Parallel()
	m := NewZWaveEditModel()

	if m.Visible() {
		t.Error("should not be visible initially")
	}
	if len(m.modes) != 3 {
		t.Errorf("modes = %d, want 3", len(m.modes))
	}
}

func TestZWaveEditModel_Show(t *testing.T) {
	t.Parallel()
	m := newTestZWaveEditModel()

	zw := &shelly.TUIZWaveStatus{
		DeviceModel: "SPSW-001PE16ZW",
		DeviceName:  "Wave Pro 1PM",
		IsPro:       true,
		SupportsLR:  true,
	}

	m, _ = m.Show(testDevice, zw)

	if !m.Visible() {
		t.Error("should be visible after Show")
	}
	if m.Device != testDevice {
		t.Errorf("device = %q, want %q", m.Device, testDevice)
	}
	if m.deviceModel != "SPSW-001PE16ZW" {
		t.Errorf("deviceModel = %q, want SPSW-001PE16ZW", m.deviceModel)
	}
	if m.deviceName != "Wave Pro 1PM" {
		t.Errorf("deviceName = %q, want Wave Pro 1PM", m.deviceName)
	}
	if !m.isPro {
		t.Error("isPro should be true")
	}
	if !m.supportsLR {
		t.Error("supportsLR should be true")
	}
	if zwaveEditField(m.Cursor) != zwaveFieldInclusion {
		t.Errorf("Cursor = %d, want %d", m.Cursor, zwaveFieldInclusion)
	}
	if m.inclusionIdx != 0 {
		t.Errorf("inclusionIdx = %d, want 0", m.inclusionIdx)
	}
	if m.pendingReset {
		t.Error("pendingReset should be false")
	}
}

func TestZWaveEditModel_Show_Standard(t *testing.T) {
	t.Parallel()
	m := newTestZWaveEditModel()

	zw := &shelly.TUIZWaveStatus{
		DeviceModel: "SNSW-001P16ZW",
		DeviceName:  "Wave 1PM",
		IsPro:       false,
		SupportsLR:  true,
	}
	m, _ = m.Show(testDevice, zw)

	if m.isPro {
		t.Error("isPro should be false for standard Wave")
	}
}

func TestZWaveEditModel_Show_NilStatus(t *testing.T) {
	t.Parallel()
	m := newTestZWaveEditModel()

	m, _ = m.Show(testDevice, nil)

	if !m.Visible() {
		t.Error("should be visible even with nil status")
	}
	if m.deviceModel != "" {
		t.Errorf("deviceModel = %q, want empty", m.deviceModel)
	}
}

func TestZWaveEditModel_Hide(t *testing.T) {
	t.Parallel()
	m := newTestZWaveEditModel()
	zw := &shelly.TUIZWaveStatus{DeviceModel: "SNSW-001P16ZW"}
	m, _ = m.Show(testDevice, zw)
	m.pendingReset = true

	m = m.Hide()

	if m.Visible() {
		t.Error("should not be visible after Hide")
	}
	if m.pendingReset {
		t.Error("pendingReset should be cleared on Hide")
	}
}

func TestZWaveEditModel_Visible(t *testing.T) {
	t.Parallel()
	m := newTestZWaveEditModel()

	if m.Visible() {
		t.Error("should not be visible initially")
	}

	m, _ = m.Show(testDevice, nil)
	if !m.Visible() {
		t.Error("should be visible")
	}
}

func TestZWaveEditModel_SetSize(t *testing.T) {
	t.Parallel()
	m := newTestZWaveEditModel()

	m = m.SetSize(100, 50)

	if m.Width != 100 {
		t.Errorf("Width = %d, want 100", m.Width)
	}
	if m.Height != 50 {
		t.Errorf("Height = %d, want 50", m.Height)
	}
}

func TestZWaveEditModel_Navigation(t *testing.T) {
	t.Parallel()
	m := showTestZWaveEditModel()

	// Start on inclusion field
	if zwaveEditField(m.Cursor) != zwaveFieldInclusion {
		t.Errorf("initial Cursor = %d, want %d", m.Cursor, zwaveFieldInclusion)
	}

	// Navigate down through all fields
	m, _ = m.Update(tea.KeyPressMsg{Code: 'j'})
	if zwaveEditField(m.Cursor) != zwaveFieldExclusion {
		t.Errorf("after j, Cursor = %d, want %d", m.Cursor, zwaveFieldExclusion)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: 'j'})
	if zwaveEditField(m.Cursor) != zwaveFieldConfig {
		t.Errorf("after j, Cursor = %d, want %d", m.Cursor, zwaveFieldConfig)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: 'j'})
	if zwaveEditField(m.Cursor) != zwaveFieldReset {
		t.Errorf("after j, Cursor = %d, want %d", m.Cursor, zwaveFieldReset)
	}

	// Can't go past last field
	m, _ = m.Update(tea.KeyPressMsg{Code: 'j'})
	if zwaveEditField(m.Cursor) != zwaveFieldReset {
		t.Errorf("past bottom, Cursor = %d, want %d", m.Cursor, zwaveFieldReset)
	}

	// Navigate up
	m, _ = m.Update(tea.KeyPressMsg{Code: 'k'})
	if zwaveEditField(m.Cursor) != zwaveFieldConfig {
		t.Errorf("after k, Cursor = %d, want %d", m.Cursor, zwaveFieldConfig)
	}

	// Navigate up to top
	m, _ = m.Update(tea.KeyPressMsg{Code: 'k'})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'k'})
	if zwaveEditField(m.Cursor) != zwaveFieldInclusion {
		t.Errorf("at top, Cursor = %d, want %d", m.Cursor, zwaveFieldInclusion)
	}

	// Can't go past first field
	m, _ = m.Update(tea.KeyPressMsg{Code: 'k'})
	if zwaveEditField(m.Cursor) != zwaveFieldInclusion {
		t.Errorf("past top, Cursor = %d, want %d", m.Cursor, zwaveFieldInclusion)
	}
}

func TestZWaveEditModel_NavigationMsg(t *testing.T) {
	t.Parallel()
	m := showTestZWaveEditModel()

	m, _ = m.Update(messages.NavigationMsg{Direction: messages.NavDown})
	if zwaveEditField(m.Cursor) != zwaveFieldExclusion {
		t.Errorf("after NavDown, Cursor = %d, want %d", m.Cursor, zwaveFieldExclusion)
	}

	m, _ = m.Update(messages.NavigationMsg{Direction: messages.NavUp})
	if zwaveEditField(m.Cursor) != zwaveFieldInclusion {
		t.Errorf("after NavUp, Cursor = %d, want %d", m.Cursor, zwaveFieldInclusion)
	}
}

func TestZWaveEditModel_NavigationClearsPendingReset(t *testing.T) {
	t.Parallel()
	m := showTestZWaveEditModel()
	m.Cursor = int(zwaveFieldReset)
	m.pendingReset = true

	m, _ = m.Update(tea.KeyPressMsg{Code: 'k'})

	if m.pendingReset {
		t.Error("navigation should clear pendingReset")
	}
}

func TestZWaveEditModel_NavigationMsgClearsPendingReset(t *testing.T) {
	t.Parallel()
	m := showTestZWaveEditModel()
	m.Cursor = int(zwaveFieldReset)
	m.pendingReset = true

	m, _ = m.Update(messages.NavigationMsg{Direction: messages.NavUp})

	if m.pendingReset {
		t.Error("NavigationMsg should clear pendingReset")
	}
}

func TestZWaveEditModel_InclusionModeSelector(t *testing.T) {
	t.Parallel()
	m := showTestZWaveEditModel()

	// Start on SmartStart (index 0)
	if m.inclusionIdx != 0 {
		t.Errorf("initial inclusionIdx = %d, want 0", m.inclusionIdx)
	}

	// Cycle with Space
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeySpace})
	if m.inclusionIdx != 1 {
		t.Errorf("after Space, inclusionIdx = %d, want 1", m.inclusionIdx)
	}

	// Cycle again
	m, _ = m.Update(tea.KeyPressMsg{Code: 't'})
	if m.inclusionIdx != 2 {
		t.Errorf("after t, inclusionIdx = %d, want 2", m.inclusionIdx)
	}

	// Wrap around
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeySpace})
	if m.inclusionIdx != 0 {
		t.Errorf("after wrap, inclusionIdx = %d, want 0", m.inclusionIdx)
	}
}

func TestZWaveEditModel_InclusionModeLeftRight(t *testing.T) {
	t.Parallel()
	m := showTestZWaveEditModel()

	// Navigate right
	m, _ = m.Update(tea.KeyPressMsg{Code: 'l'})
	if m.inclusionIdx != 1 {
		t.Errorf("after l, inclusionIdx = %d, want 1", m.inclusionIdx)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: 'l'})
	if m.inclusionIdx != 2 {
		t.Errorf("after l, inclusionIdx = %d, want 2", m.inclusionIdx)
	}

	// Can't go past end
	m, _ = m.Update(tea.KeyPressMsg{Code: 'l'})
	if m.inclusionIdx != 2 {
		t.Errorf("past end, inclusionIdx = %d, want 2", m.inclusionIdx)
	}

	// Navigate left
	m, _ = m.Update(tea.KeyPressMsg{Code: 'h'})
	if m.inclusionIdx != 1 {
		t.Errorf("after h, inclusionIdx = %d, want 1", m.inclusionIdx)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: 'h'})
	if m.inclusionIdx != 0 {
		t.Errorf("after h, inclusionIdx = %d, want 0", m.inclusionIdx)
	}

	// Can't go past start
	m, _ = m.Update(tea.KeyPressMsg{Code: 'h'})
	if m.inclusionIdx != 0 {
		t.Errorf("past start, inclusionIdx = %d, want 0", m.inclusionIdx)
	}
}

func TestZWaveEditModel_InclusionModeNavLeftRight(t *testing.T) {
	t.Parallel()
	m := showTestZWaveEditModel()

	// NavigationMsg left/right on inclusion field
	m, _ = m.Update(messages.NavigationMsg{Direction: messages.NavRight})
	if m.inclusionIdx != 1 {
		t.Errorf("after NavRight, inclusionIdx = %d, want 1", m.inclusionIdx)
	}

	m, _ = m.Update(messages.NavigationMsg{Direction: messages.NavLeft})
	if m.inclusionIdx != 0 {
		t.Errorf("after NavLeft, inclusionIdx = %d, want 0", m.inclusionIdx)
	}
}

func TestZWaveEditModel_InclusionModeOnlyOnInclusionField(t *testing.T) {
	t.Parallel()
	m := showTestZWaveEditModel()
	m.Cursor = int(zwaveFieldExclusion) // Not on inclusion field

	originalIdx := m.inclusionIdx
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeySpace})

	if m.inclusionIdx != originalIdx {
		t.Error("Space should not change inclusionIdx when not on inclusion field")
	}
}

func TestZWaveEditModel_ResetConfirmation(t *testing.T) {
	t.Parallel()
	m := showTestZWaveEditModel()
	m.Cursor = int(zwaveFieldReset)

	// First press - request confirmation
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !m.pendingReset {
		t.Error("should set pendingReset on first press")
	}
	if !m.Visible() {
		t.Error("should still be visible")
	}

	// Second press - confirm and close
	m, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.pendingReset {
		t.Error("pendingReset should be false after confirmation")
	}
	if m.Visible() {
		t.Error("should close after second press")
	}
	if cmd == nil {
		t.Error("should return close command")
	}
}

func TestZWaveEditModel_ResetNonDestructiveOnOtherFields(t *testing.T) {
	t.Parallel()
	m := showTestZWaveEditModel()
	m.Cursor = int(zwaveFieldExclusion)

	// Enter on non-reset fields should not trigger reset
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.pendingReset {
		t.Error("pendingReset should not be set on non-reset fields")
	}
}

func TestZWaveEditModel_EscClose(t *testing.T) {
	t.Parallel()
	m := showTestZWaveEditModel()

	m, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if m.Visible() {
		t.Error("should not be visible after Esc")
	}
	if cmd == nil {
		t.Error("should return close command")
	}
}

func TestZWaveEditModel_CtrlBracketClose(t *testing.T) {
	t.Parallel()
	m := showTestZWaveEditModel()

	m, cmd := m.Update(tea.KeyPressMsg{Mod: tea.ModCtrl, Code: '['})

	if m.Visible() {
		t.Error("should not be visible after Ctrl+[")
	}
	if cmd == nil {
		t.Error("should return close command")
	}
}

func TestZWaveEditModel_UpdateNotVisible(t *testing.T) {
	t.Parallel()
	m := newTestZWaveEditModel()

	m, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if m.Visible() {
		t.Error("should not become visible")
	}
	if cmd != nil {
		t.Error("should return nil command when not visible")
	}
}

func TestZWaveEditModel_View(t *testing.T) {
	t.Parallel()
	m := newTestZWaveEditModel()
	m = m.SetSize(80, 40)
	zw := &shelly.TUIZWaveStatus{
		DeviceModel: "SPSW-001PE16ZW",
		DeviceName:  "Wave Pro 1PM",
		IsPro:       true,
		SupportsLR:  true,
	}
	m, _ = m.Show(testDevice, zw)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestZWaveEditModel_View_NotVisible(t *testing.T) {
	t.Parallel()
	m := newTestZWaveEditModel()

	view := m.View()

	if view != "" {
		t.Error("View() should return empty string when not visible")
	}
}

func TestZWaveEditModel_View_Standard(t *testing.T) {
	t.Parallel()
	m := newTestZWaveEditModel()
	m = m.SetSize(80, 40)
	zw := &shelly.TUIZWaveStatus{
		DeviceModel: "SNSW-001P16ZW",
		DeviceName:  "Wave 1",
		IsPro:       false,
		SupportsLR:  true,
	}
	m, _ = m.Show(testDevice, zw)

	view := m.View()
	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestZWaveEditModel_View_PendingReset(t *testing.T) {
	t.Parallel()
	m := newTestZWaveEditModel()
	m = m.SetSize(80, 40)
	zw := &shelly.TUIZWaveStatus{DeviceModel: "SNSW-001P16ZW"}
	m, _ = m.Show(testDevice, zw)
	m.Cursor = int(zwaveFieldReset)
	m.pendingReset = true

	view := m.View()
	if view == "" {
		t.Error("View() should not return empty string with pending reset")
	}
}

func TestZWaveEditModel_View_FocusedExclusion(t *testing.T) {
	t.Parallel()
	m := newTestZWaveEditModel()
	m = m.SetSize(80, 40)
	zw := &shelly.TUIZWaveStatus{DeviceModel: "SNSW-001P16ZW"}
	m, _ = m.Show(testDevice, zw)
	m.Cursor = int(zwaveFieldExclusion)

	view := m.View()
	if view == "" {
		t.Error("View() should not return empty string with exclusion focused")
	}
}

func TestZWaveEditModel_View_FocusedConfig(t *testing.T) {
	t.Parallel()
	m := newTestZWaveEditModel()
	m = m.SetSize(80, 40)
	zw := &shelly.TUIZWaveStatus{DeviceModel: "SNSW-001P16ZW"}
	m, _ = m.Show(testDevice, zw)
	m.Cursor = int(zwaveFieldConfig)

	view := m.View()
	if view == "" {
		t.Error("View() should not return empty string with config focused")
	}
}

func TestZWaveEditModel_Footer(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		setup    func(m *ZWaveEditModel)
		contains string
	}{
		{
			name:     "inclusion mode",
			setup:    func(m *ZWaveEditModel) { m.Cursor = int(zwaveFieldInclusion) },
			contains: "mode",
		},
		{
			name:     "exclusion",
			setup:    func(m *ZWaveEditModel) { m.Cursor = int(zwaveFieldExclusion) },
			contains: "Navigate",
		},
		{
			name:     "config",
			setup:    func(m *ZWaveEditModel) { m.Cursor = int(zwaveFieldConfig) },
			contains: "Navigate",
		},
		{
			name:     "pending reset",
			setup:    func(m *ZWaveEditModel) { m.pendingReset = true },
			contains: "Enter again",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := showTestZWaveEditModel()
			tt.setup(&m)

			footer := m.buildFooter()

			if footer == "" {
				t.Error("footer should not be empty")
			}
		})
	}
}

func TestZWaveEditModel_InclusionModes(t *testing.T) {
	t.Parallel()
	modes := wireless.ZWaveInclusionModes()

	if len(modes) != 3 {
		t.Fatalf("expected 3 inclusion modes, got %d", len(modes))
	}

	expected := []wireless.ZWaveInclusionMode{
		wireless.ZWaveInclusionSmartStart,
		wireless.ZWaveInclusionButton,
		wireless.ZWaveInclusionSwitch,
	}
	for i, mode := range modes {
		if mode != expected[i] {
			t.Errorf("mode[%d] = %q, want %q", i, mode, expected[i])
		}
	}
}

func TestZWaveEditModel_InclusionSteps(t *testing.T) {
	t.Parallel()

	for _, mode := range wireless.ZWaveInclusionModes() {
		steps := wireless.ZWaveInclusionSteps(mode)
		if len(steps) == 0 {
			t.Errorf("no steps for mode %q", mode)
		}
	}
}

func TestZWaveEditModel_ExclusionSteps(t *testing.T) {
	t.Parallel()

	steps := wireless.ZWaveExclusionSteps(wireless.ZWaveInclusionButton)
	if len(steps) == 0 {
		t.Error("no exclusion steps for button mode")
	}
}

func TestZWaveEditModel_FactoryResetInfo(t *testing.T) {
	t.Parallel()

	warning := wireless.ZWaveFactoryResetWarning()
	if warning == "" {
		t.Error("factory reset warning should not be empty")
	}

	steps := wireless.ZWaveFactoryResetSteps()
	if len(steps) == 0 {
		t.Error("factory reset steps should not be empty")
	}
}

func TestZWaveEditModel_ConfigParams(t *testing.T) {
	t.Parallel()

	params := wireless.ZWaveCommonConfigParams()
	if len(params) == 0 {
		t.Error("config params should not be empty")
	}

	for _, p := range params {
		if p.Name == "" {
			t.Errorf("param %d has empty name", p.Number)
		}
		if p.Description == "" {
			t.Errorf("param %d (%s) has empty description", p.Number, p.Name)
		}
	}
}

func TestZWaveEditModel_InclusionModeName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mode wireless.ZWaveInclusionMode
		want string
	}{
		{wireless.ZWaveInclusionSmartStart, "SmartStart (QR Code)"},
		{wireless.ZWaveInclusionButton, "S Button"},
		{wireless.ZWaveInclusionSwitch, "Connected Switch"},
		{wireless.ZWaveInclusionMode("unknown"), "unknown"},
	}

	for _, tt := range tests {
		name := wireless.ZWaveInclusionModeName(tt.mode)
		if name != tt.want {
			t.Errorf("ZWaveInclusionModeName(%q) = %q, want %q", tt.mode, name, tt.want)
		}
	}
}

func newTestZWaveEditModel() ZWaveEditModel {
	return NewZWaveEditModel()
}

func showTestZWaveEditModel() ZWaveEditModel {
	m := newTestZWaveEditModel()
	zw := &shelly.TUIZWaveStatus{
		DeviceModel: "SPSW-001PE16ZW",
		DeviceName:  "Wave Pro 1PM",
		IsPro:       true,
		SupportsLR:  true,
	}
	m, _ = m.Show(testDevice, zw)
	return m
}
