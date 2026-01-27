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

	if m.visible {
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

	if !m.visible {
		t.Error("should be visible after Show")
	}
	if m.device != testDevice {
		t.Errorf("device = %q, want %q", m.device, testDevice)
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
	if m.field != zwaveFieldInclusion {
		t.Errorf("field = %d, want %d", m.field, zwaveFieldInclusion)
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

	if !m.visible {
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

	if m.visible {
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

	m.visible = true
	if !m.Visible() {
		t.Error("should be visible")
	}
}

func TestZWaveEditModel_SetSize(t *testing.T) {
	t.Parallel()
	m := newTestZWaveEditModel()

	m = m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}

func TestZWaveEditModel_Navigation(t *testing.T) {
	t.Parallel()
	m := showTestZWaveEditModel()

	// Start on inclusion field
	if m.field != zwaveFieldInclusion {
		t.Errorf("initial field = %d, want %d", m.field, zwaveFieldInclusion)
	}

	// Navigate down through all fields
	m, _ = m.Update(tea.KeyPressMsg{Code: 'j'})
	if m.field != zwaveFieldExclusion {
		t.Errorf("after j, field = %d, want %d", m.field, zwaveFieldExclusion)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: 'j'})
	if m.field != zwaveFieldConfig {
		t.Errorf("after j, field = %d, want %d", m.field, zwaveFieldConfig)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: 'j'})
	if m.field != zwaveFieldReset {
		t.Errorf("after j, field = %d, want %d", m.field, zwaveFieldReset)
	}

	// Can't go past last field
	m, _ = m.Update(tea.KeyPressMsg{Code: 'j'})
	if m.field != zwaveFieldReset {
		t.Errorf("past bottom, field = %d, want %d", m.field, zwaveFieldReset)
	}

	// Navigate up
	m, _ = m.Update(tea.KeyPressMsg{Code: 'k'})
	if m.field != zwaveFieldConfig {
		t.Errorf("after k, field = %d, want %d", m.field, zwaveFieldConfig)
	}

	// Navigate up to top
	m, _ = m.Update(tea.KeyPressMsg{Code: 'k'})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'k'})
	if m.field != zwaveFieldInclusion {
		t.Errorf("at top, field = %d, want %d", m.field, zwaveFieldInclusion)
	}

	// Can't go past first field
	m, _ = m.Update(tea.KeyPressMsg{Code: 'k'})
	if m.field != zwaveFieldInclusion {
		t.Errorf("past top, field = %d, want %d", m.field, zwaveFieldInclusion)
	}
}

func TestZWaveEditModel_NavigationMsg(t *testing.T) {
	t.Parallel()
	m := showTestZWaveEditModel()

	m, _ = m.Update(messages.NavigationMsg{Direction: messages.NavDown})
	if m.field != zwaveFieldExclusion {
		t.Errorf("after NavDown, field = %d, want %d", m.field, zwaveFieldExclusion)
	}

	m, _ = m.Update(messages.NavigationMsg{Direction: messages.NavUp})
	if m.field != zwaveFieldInclusion {
		t.Errorf("after NavUp, field = %d, want %d", m.field, zwaveFieldInclusion)
	}
}

func TestZWaveEditModel_NavigationClearsPendingReset(t *testing.T) {
	t.Parallel()
	m := showTestZWaveEditModel()
	m.field = zwaveFieldReset
	m.pendingReset = true

	m, _ = m.Update(tea.KeyPressMsg{Code: 'k'})

	if m.pendingReset {
		t.Error("navigation should clear pendingReset")
	}
}

func TestZWaveEditModel_NavigationMsgClearsPendingReset(t *testing.T) {
	t.Parallel()
	m := showTestZWaveEditModel()
	m.field = zwaveFieldReset
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
	m.field = zwaveFieldExclusion // Not on inclusion field

	originalIdx := m.inclusionIdx
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeySpace})

	if m.inclusionIdx != originalIdx {
		t.Error("Space should not change inclusionIdx when not on inclusion field")
	}
}

func TestZWaveEditModel_ResetConfirmation(t *testing.T) {
	t.Parallel()
	m := showTestZWaveEditModel()
	m.field = zwaveFieldReset

	// First press - request confirmation
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !m.pendingReset {
		t.Error("should set pendingReset on first press")
	}
	if !m.visible {
		t.Error("should still be visible")
	}

	// Second press - confirm and close
	m, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.pendingReset {
		t.Error("pendingReset should be false after confirmation")
	}
	if m.visible {
		t.Error("should close after second press")
	}
	if cmd == nil {
		t.Error("should return close command")
	}
}

func TestZWaveEditModel_ResetNonDestructiveOnOtherFields(t *testing.T) {
	t.Parallel()
	m := showTestZWaveEditModel()
	m.field = zwaveFieldExclusion

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

	if m.visible {
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

	if m.visible {
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

	if m.visible {
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
	m.field = zwaveFieldReset
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
	m.field = zwaveFieldExclusion

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
	m.field = zwaveFieldConfig

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
			setup:    func(m *ZWaveEditModel) { m.field = zwaveFieldInclusion },
			contains: "mode",
		},
		{
			name:     "exclusion",
			setup:    func(m *ZWaveEditModel) { m.field = zwaveFieldExclusion },
			contains: "Navigate",
		},
		{
			name:     "config",
			setup:    func(m *ZWaveEditModel) { m.field = zwaveFieldConfig },
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
