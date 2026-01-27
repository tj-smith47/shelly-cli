package smarthome

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
)

func newTestModbusEditModel() ModbusEditModel {
	ctx := context.Background()
	svc := &shelly.Service{}
	return NewModbusEditModel(ctx, svc)
}

func TestNewModbusEditModel(t *testing.T) {
	t.Parallel()

	m := newTestModbusEditModel()

	if m.Visible() {
		t.Error("should not be visible initially")
	}
}

func TestModbusEditModel_ShowHide(t *testing.T) {
	t.Parallel()

	m := newTestModbusEditModel()
	modbus := &shelly.TUIModbusStatus{Enabled: true}

	shown, cmd := m.Show("192.168.1.100", modbus)
	if !shown.Visible() {
		t.Error("should be visible after Show")
	}
	if shown.device != "192.168.1.100" {
		t.Errorf("device = %q, want %q", shown.device, "192.168.1.100")
	}
	if !shown.enabled {
		t.Error("enabled should be true")
	}
	if !shown.pendingEnabled {
		t.Error("pendingEnabled should match enabled")
	}
	if cmd != nil {
		t.Error("should not return command (no async loading)")
	}

	hidden := shown.Hide()
	if hidden.Visible() {
		t.Error("should not be visible after Hide")
	}
}

func TestModbusEditModel_ShowDisabled(t *testing.T) {
	t.Parallel()

	m := newTestModbusEditModel()
	modbus := &shelly.TUIModbusStatus{Enabled: false}

	shown, cmd := m.Show("192.168.1.100", modbus)
	if shown.enabled {
		t.Error("enabled should be false")
	}
	if shown.pendingEnabled {
		t.Error("pendingEnabled should be false")
	}
	if cmd != nil {
		t.Error("should not return command")
	}
}

func TestModbusEditModel_ShowNilModbus(t *testing.T) {
	t.Parallel()

	m := newTestModbusEditModel()
	shown, cmd := m.Show("192.168.1.100", nil)

	if !shown.Visible() {
		t.Error("should be visible even with nil modbus")
	}
	if shown.enabled {
		t.Error("enabled should be false with nil modbus")
	}
	if cmd != nil {
		t.Error("should not return command with nil modbus")
	}
}

func TestModbusEditModel_SetSize(t *testing.T) {
	t.Parallel()

	m := newTestModbusEditModel()
	m = m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}

func TestModbusEditModel_UpdateNotVisible(t *testing.T) {
	t.Parallel()

	m := newTestModbusEditModel()
	keyMsg := tea.KeyPressMsg{Code: tea.KeyEscape}
	updated, cmd := m.Update(keyMsg)

	if updated.Visible() {
		t.Error("should remain hidden")
	}
	if cmd != nil {
		t.Error("cmd should be nil when not visible")
	}
}

func TestModbusEditModel_EscClose(t *testing.T) {
	t.Parallel()

	m := newTestModbusEditModel()
	modbus := &shelly.TUIModbusStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", modbus)

	keyMsg := tea.KeyPressMsg{Code: tea.KeyEscape}
	updated, cmd := m.Update(keyMsg)

	if updated.Visible() {
		t.Error("should be hidden after Esc")
	}
	if cmd == nil {
		t.Fatal("should return close message cmd")
	}

	msg := cmd()
	closedMsg, ok := msg.(EditClosedMsg)
	if !ok {
		t.Fatal("should return EditClosedMsg")
	}
	if closedMsg.Saved {
		t.Error("Saved should be false for cancel")
	}
}

func TestModbusEditModel_ToggleEnabled(t *testing.T) {
	t.Parallel()

	m := newTestModbusEditModel()
	modbus := &shelly.TUIModbusStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", modbus)

	// Toggle with space key
	spaceMsg := tea.KeyPressMsg{Code: tea.KeySpace}
	updated, _ := m.Update(spaceMsg)

	if updated.pendingEnabled {
		t.Error("pendingEnabled should be false after toggling from true")
	}

	// Toggle back with 't' key
	tMsg := tea.KeyPressMsg{Code: 't'}
	updated, _ = updated.Update(tMsg)

	if !updated.pendingEnabled {
		t.Error("pendingEnabled should be true after toggling back")
	}
}

func TestModbusEditModel_ToggleViaMessage(t *testing.T) {
	t.Parallel()

	m := newTestModbusEditModel()
	modbus := &shelly.TUIModbusStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", modbus)

	updated, _ := m.Update(messages.ToggleEnableRequestMsg{})

	if updated.pendingEnabled {
		t.Error("pendingEnabled should be false after ToggleEnableRequestMsg")
	}
}

func TestModbusEditModel_NavigationBounds(t *testing.T) {
	t.Parallel()

	m := newTestModbusEditModel()
	modbus := &shelly.TUIModbusStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", modbus)

	// Should start at enable field
	if m.field != modbusFieldEnable {
		t.Errorf("field = %d, want %d", m.field, modbusFieldEnable)
	}

	// Navigate up at top - should stay at top
	upMsg := messages.NavigationMsg{Direction: messages.NavUp}
	updated, _ := m.Update(upMsg)

	if updated.field != modbusFieldEnable {
		t.Errorf("field should stay at %d, got %d", modbusFieldEnable, updated.field)
	}

	// Navigate down at bottom (only 1 field) - should stay at bottom
	downMsg := messages.NavigationMsg{Direction: messages.NavDown}
	updated, _ = updated.Update(downMsg)

	if updated.field != modbusFieldEnable {
		t.Errorf("field should stay at %d, got %d", modbusFieldEnable, updated.field)
	}
}

func TestModbusEditModel_KeyJKNavigation(t *testing.T) {
	t.Parallel()

	m := newTestModbusEditModel()
	modbus := &shelly.TUIModbusStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", modbus)

	// Navigate down with j (only 1 field, should stay)
	jMsg := tea.KeyPressMsg{Code: 'j'}
	updated, _ := m.Update(jMsg)

	if updated.field != modbusFieldEnable {
		t.Errorf("field = %d, want %d after 'j' (single field)", updated.field, modbusFieldEnable)
	}

	// Navigate up with k (at top, should stay)
	kMsg := tea.KeyPressMsg{Code: 'k'}
	updated, _ = updated.Update(kMsg)

	if updated.field != modbusFieldEnable {
		t.Errorf("field = %d, want %d after 'k'", updated.field, modbusFieldEnable)
	}
}

func TestModbusEditModel_SaveNoChanges(t *testing.T) {
	t.Parallel()

	m := newTestModbusEditModel()
	modbus := &shelly.TUIModbusStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", modbus)

	// Save without changes (Enter while on enable field)
	enterMsg := tea.KeyPressMsg{Code: tea.KeyEnter}
	updated, cmd := m.Update(enterMsg)

	// Should close without saving
	if updated.Visible() {
		t.Error("should be hidden when no changes")
	}
	if cmd == nil {
		t.Fatal("should return close cmd")
	}

	msg := cmd()
	closedMsg, ok := msg.(EditClosedMsg)
	if !ok {
		t.Fatal("should return EditClosedMsg")
	}
	if closedMsg.Saved {
		t.Error("Saved should be false for no-changes close")
	}
}

func TestModbusEditModel_SaveWithChanges(t *testing.T) {
	t.Parallel()

	m := newTestModbusEditModel()
	modbus := &shelly.TUIModbusStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", modbus)

	// Toggle to disable
	m.pendingEnabled = false

	// Save
	enterMsg := tea.KeyPressMsg{Code: tea.KeyEnter}
	updated, cmd := m.Update(enterMsg)

	if !updated.saving {
		t.Error("should be saving after Enter with changes")
	}
	if cmd == nil {
		t.Error("should return save command")
	}
}

func TestModbusEditModel_SaveResultSuccess(t *testing.T) {
	t.Parallel()

	m := newTestModbusEditModel()
	modbus := &shelly.TUIModbusStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", modbus)
	m.saving = true
	m.pendingEnabled = false

	saveMsg := messages.SaveResultMsg{Success: true}
	updated, cmd := m.Update(saveMsg)

	if updated.saving {
		t.Error("saving should be false after success")
	}
	if updated.Visible() {
		t.Error("should be hidden after successful save")
	}
	if cmd == nil {
		t.Fatal("should return close cmd")
	}

	msg := cmd()
	closedMsg, ok := msg.(EditClosedMsg)
	if !ok {
		t.Fatal("should return EditClosedMsg")
	}
	if !closedMsg.Saved {
		t.Error("Saved should be true for successful save")
	}
}

func TestModbusEditModel_SaveResultError(t *testing.T) {
	t.Parallel()

	m := newTestModbusEditModel()
	modbus := &shelly.TUIModbusStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", modbus)
	m.saving = true

	saveErr := errors.New("connection failed")
	saveMsg := messages.SaveResultMsg{Err: saveErr}
	updated, _ := m.Update(saveMsg)

	if updated.saving {
		t.Error("saving should be false after error")
	}
	if !updated.Visible() {
		t.Error("should remain visible after error")
	}
	if updated.err == nil {
		t.Error("err should be set")
	}
}

func TestModbusEditModel_ToggleBlockedWhileSaving(t *testing.T) {
	t.Parallel()

	m := newTestModbusEditModel()
	modbus := &shelly.TUIModbusStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", modbus)
	m.saving = true

	// Toggle should be blocked
	spaceMsg := tea.KeyPressMsg{Code: tea.KeySpace}
	updated, _ := m.Update(spaceMsg)

	if !updated.pendingEnabled {
		t.Error("pendingEnabled should not change while saving")
	}
}

func TestModbusEditModel_ToggleViaMessageBlockedWhileSaving(t *testing.T) {
	t.Parallel()

	m := newTestModbusEditModel()
	modbus := &shelly.TUIModbusStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", modbus)
	m.saving = true

	updated, _ := m.Update(messages.ToggleEnableRequestMsg{})

	if !updated.pendingEnabled {
		t.Error("pendingEnabled should not change while saving")
	}
}

func TestModbusEditModel_SaveBlockedWhileSaving(t *testing.T) {
	t.Parallel()

	m := newTestModbusEditModel()
	modbus := &shelly.TUIModbusStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", modbus)
	m.saving = true
	m.pendingEnabled = false

	// Save should be blocked while already saving
	enterMsg := tea.KeyPressMsg{Code: tea.KeyEnter}
	updated, cmd := m.Update(enterMsg)

	if !updated.saving {
		t.Error("saving should remain true")
	}
	if cmd != nil {
		t.Error("should not return command when already saving")
	}
}

func TestModbusEditModel_CtrlSClose(t *testing.T) {
	t.Parallel()

	m := newTestModbusEditModel()
	modbus := &shelly.TUIModbusStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", modbus)

	// Ctrl+S with no changes should close
	updated, cmd := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})

	// No changes, should close
	if updated.Visible() {
		t.Error("should be hidden after Ctrl+S with no changes")
	}
	if cmd == nil {
		t.Fatal("should return close cmd")
	}
}

func TestModbusEditModel_View_NotVisible(t *testing.T) {
	t.Parallel()

	m := newTestModbusEditModel()
	view := m.View()

	if view != "" {
		t.Error("View should be empty when not visible")
	}
}

func TestModbusEditModel_View_Enabled(t *testing.T) {
	t.Parallel()

	m := newTestModbusEditModel()
	m = m.SetSize(80, 40)
	modbus := &shelly.TUIModbusStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", modbus)

	view := m.View()

	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestModbusEditModel_View_Disabled(t *testing.T) {
	t.Parallel()

	m := newTestModbusEditModel()
	m = m.SetSize(80, 40)
	modbus := &shelly.TUIModbusStatus{Enabled: false}
	m, _ = m.Show("192.168.1.100", modbus)

	view := m.View()

	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestModbusEditModel_View_Saving(t *testing.T) {
	t.Parallel()

	m := newTestModbusEditModel()
	m = m.SetSize(80, 40)
	modbus := &shelly.TUIModbusStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", modbus)
	m.saving = true

	view := m.View()

	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestModbusEditModel_View_WithError(t *testing.T) {
	t.Parallel()

	m := newTestModbusEditModel()
	m = m.SetSize(80, 40)
	modbus := &shelly.TUIModbusStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", modbus)
	m.err = errors.New("test error")

	view := m.View()

	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestModbusEditModel_View_PendingChange(t *testing.T) {
	t.Parallel()

	m := newTestModbusEditModel()
	m = m.SetSize(80, 40)
	modbus := &shelly.TUIModbusStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", modbus)
	m.pendingEnabled = false // Change from enabled to disabled

	view := m.View()

	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestModbusEditModel_NavigationNonApplicable(t *testing.T) {
	t.Parallel()

	m := newTestModbusEditModel()
	modbus := &shelly.TUIModbusStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", modbus)

	// Non-applicable navigation directions should be no-ops
	directions := []messages.NavDirection{
		messages.NavLeft,
		messages.NavRight,
		messages.NavPageUp,
		messages.NavPageDown,
		messages.NavHome,
		messages.NavEnd,
	}

	for _, dir := range directions {
		updated, _ := m.Update(messages.NavigationMsg{Direction: dir})
		if updated.field != modbusFieldEnable {
			t.Errorf("field changed for non-applicable direction %d", dir)
		}
	}
}
