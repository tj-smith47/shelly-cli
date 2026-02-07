package smarthome

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
)

func newTestMatterEditModel() MatterEditModel {
	ctx := context.Background()
	svc := &shelly.Service{}
	return NewMatterEditModel(ctx, svc)
}

func TestNewMatterEditModel(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()

	if m.Visible() {
		t.Error("should not be visible initially")
	}
}

func TestMatterEditModel_ShowHide(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	matter := &shelly.TUIMatterStatus{
		Enabled:        true,
		Commissionable: false,
		FabricsCount:   2,
	}

	shown, _ := m.Show("192.168.1.100", matter)
	if !shown.Visible() {
		t.Error("should be visible after Show")
	}
	if shown.Device != "192.168.1.100" {
		t.Errorf("Device = %q, want %q", shown.Device, "192.168.1.100")
	}
	if !shown.enabled {
		t.Error("enabled should be true")
	}
	if !shown.pendingEnabled {
		t.Error("pendingEnabled should match enabled")
	}
	if shown.fabricsCount != 2 {
		t.Errorf("fabricsCount = %d, want 2", shown.fabricsCount)
	}

	hidden := shown.Hide()
	if hidden.Visible() {
		t.Error("should not be visible after Hide")
	}
}

func TestMatterEditModel_ShowCommissionable(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	matter := &shelly.TUIMatterStatus{
		Enabled:        true,
		Commissionable: true,
		FabricsCount:   0,
	}

	shown, cmd := m.Show("192.168.1.100", matter)
	if !shown.loadingCodes {
		t.Error("should be loading codes when commissionable")
	}
	if cmd == nil {
		t.Error("should return fetch codes command when commissionable")
	}
}

func TestMatterEditModel_ShowNotCommissionable(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	matter := &shelly.TUIMatterStatus{
		Enabled:        true,
		Commissionable: false,
		FabricsCount:   1,
	}

	shown, cmd := m.Show("192.168.1.100", matter)
	if shown.loadingCodes {
		t.Error("should not be loading codes when not commissionable")
	}
	if cmd != nil {
		t.Error("should not return fetch codes command when not commissionable")
	}
}

func TestMatterEditModel_ShowDisabled(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	matter := &shelly.TUIMatterStatus{
		Enabled: false,
	}

	shown, cmd := m.Show("192.168.1.100", matter)
	if shown.enabled {
		t.Error("enabled should be false")
	}
	if shown.FieldCount != 1 {
		t.Errorf("FieldCount = %d, want 1 (no reset button when disabled)", shown.FieldCount)
	}
	if cmd != nil {
		t.Error("should not fetch codes when disabled")
	}
}

func TestMatterEditModel_SetSize(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	m = m.SetSize(100, 50)

	if m.Width != 100 {
		t.Errorf("Width = %d, want 100", m.Width)
	}
	if m.Height != 50 {
		t.Errorf("Height = %d, want 50", m.Height)
	}
}

func TestMatterEditModel_UpdateNotVisible(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	keyMsg := tea.KeyPressMsg{Code: tea.KeyEscape}
	updated, cmd := m.Update(keyMsg)

	if updated.Visible() {
		t.Error("should remain hidden")
	}
	if cmd != nil {
		t.Error("cmd should be nil when not visible")
	}
}

func TestMatterEditModel_EscClose(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	matter := &shelly.TUIMatterStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", matter)

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

func TestMatterEditModel_ToggleEnabled(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	matter := &shelly.TUIMatterStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", matter)

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

func TestMatterEditModel_ToggleViaMessage(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	matter := &shelly.TUIMatterStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", matter)

	updated, _ := m.Update(messages.ToggleEnableRequestMsg{})

	if updated.pendingEnabled {
		t.Error("pendingEnabled should be false after ToggleEnableRequestMsg")
	}
}

func TestMatterEditModel_NavigationUpDown(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	matter := &shelly.TUIMatterStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", matter)

	// Should start at enable field
	if matterEditField(m.Cursor) != matterFieldEnable {
		t.Errorf("Cursor = %d, want %d", m.Cursor, matterFieldEnable)
	}

	// Navigate down to reset button
	downMsg := messages.NavigationMsg{Direction: messages.NavDown}
	updated, _ := m.Update(downMsg)

	if matterEditField(updated.Cursor) != matterFieldReset {
		t.Errorf("Cursor = %d, want %d", updated.Cursor, matterFieldReset)
	}

	// Navigate up back to enable
	upMsg := messages.NavigationMsg{Direction: messages.NavUp}
	updated, _ = updated.Update(upMsg)

	if matterEditField(updated.Cursor) != matterFieldEnable {
		t.Errorf("Cursor = %d, want %d", updated.Cursor, matterFieldEnable)
	}
}

func TestMatterEditModel_NavigationBounds(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	matter := &shelly.TUIMatterStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", matter)

	// Navigate up at top - should stay at top
	upMsg := messages.NavigationMsg{Direction: messages.NavUp}
	updated, _ := m.Update(upMsg)

	if matterEditField(updated.Cursor) != matterFieldEnable {
		t.Errorf("Cursor should stay at %d, got %d", matterFieldEnable, updated.Cursor)
	}

	// Navigate down to reset
	downMsg := messages.NavigationMsg{Direction: messages.NavDown}
	updated, _ = updated.Update(downMsg)

	// Navigate down past bottom - should stay at bottom
	updated, _ = updated.Update(downMsg)

	if matterEditField(updated.Cursor) != matterFieldReset {
		t.Errorf("Cursor should stay at %d, got %d", matterFieldReset, updated.Cursor)
	}
}

func TestMatterEditModel_SaveNoChanges(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	matter := &shelly.TUIMatterStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", matter)

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

func TestMatterEditModel_SaveWithChanges(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	matter := &shelly.TUIMatterStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", matter)

	// Toggle to disable
	m.pendingEnabled = false

	// Save
	enterMsg := tea.KeyPressMsg{Code: tea.KeyEnter}
	updated, cmd := m.Update(enterMsg)

	if !updated.Saving {
		t.Error("should be saving after Enter with changes")
	}
	if cmd == nil {
		t.Error("should return save command")
	}
}

func TestMatterEditModel_SaveResultSuccess(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	matter := &shelly.TUIMatterStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", matter)
	m.Saving = true
	m.pendingEnabled = false

	saveMsg := messages.SaveResultMsg{Success: true}
	updated, cmd := m.Update(saveMsg)

	if updated.Saving {
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

func TestMatterEditModel_SaveResultError(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	matter := &shelly.TUIMatterStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", matter)
	m.Saving = true

	saveErr := errors.New("connection failed")
	saveMsg := messages.SaveResultMsg{Err: saveErr}
	updated, _ := m.Update(saveMsg)

	if updated.Saving {
		t.Error("saving should be false after error")
	}
	if !updated.Visible() {
		t.Error("should remain visible after error")
	}
	if updated.Err == nil {
		t.Error("err should be set")
	}
}

func TestMatterEditModel_ResetConfirmation(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	matter := &shelly.TUIMatterStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", matter)

	// Navigate to reset button
	m.Cursor = int(matterFieldReset)

	// First Enter - should set pendingReset
	enterMsg := tea.KeyPressMsg{Code: tea.KeyEnter}
	updated, _ := m.Update(enterMsg)

	if !updated.pendingReset {
		t.Error("pendingReset should be true after first Enter on reset")
	}
	if updated.resetting {
		t.Error("should not be resetting yet")
	}

	// Second Enter - should execute reset
	updated, cmd := updated.Update(enterMsg)

	if updated.pendingReset {
		t.Error("pendingReset should be false after confirmation")
	}
	if !updated.resetting {
		t.Error("should be resetting after confirmation")
	}
	if cmd == nil {
		t.Error("should return reset command")
	}
}

func TestMatterEditModel_ResetCancelOnFieldChange(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	matter := &shelly.TUIMatterStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", matter)

	// Navigate to reset and initiate confirmation
	m.Cursor = int(matterFieldReset)
	m.pendingReset = true

	// Navigate up to cancel confirmation
	upMsg := tea.KeyPressMsg{Code: 'k'}
	updated, _ := m.Update(upMsg)

	if updated.pendingReset {
		t.Error("pendingReset should be canceled on field change")
	}
}

func TestMatterEditModel_ResetResultSuccess(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	matter := &shelly.TUIMatterStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", matter)
	m.resetting = true

	resetMsg := MatterResetResultMsg{Err: nil}
	updated, cmd := m.Update(resetMsg)

	if updated.resetting {
		t.Error("resetting should be false after success")
	}
	if updated.Visible() {
		t.Error("should be hidden after successful reset")
	}
	if cmd == nil {
		t.Fatal("should return close cmd")
	}
}

func TestMatterEditModel_ResetResultError(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	matter := &shelly.TUIMatterStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", matter)
	m.resetting = true

	resetErr := errors.New("reset failed")
	resetMsg := MatterResetResultMsg{Err: resetErr}
	updated, _ := m.Update(resetMsg)

	if updated.resetting {
		t.Error("resetting should be false after error")
	}
	if !updated.Visible() {
		t.Error("should remain visible after error")
	}
	if updated.Err == nil {
		t.Error("err should be set")
	}
}

func TestMatterEditModel_CodesLoaded(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	matter := &shelly.TUIMatterStatus{Enabled: true, Commissionable: true}
	m, _ = m.Show("192.168.1.100", matter)

	codesMsg := MatterCodesLoadedMsg{
		Codes: model.CommissioningInfo{
			ManualCode:    "34970112332",
			QRCode:        "MT:Y.K9042C00KA0648G00",
			SetupPINCode:  20202021,
			Discriminator: 3840,
			Available:     true,
		},
	}
	updated, _ := m.Update(codesMsg)

	if updated.loadingCodes {
		t.Error("loadingCodes should be false after codes loaded")
	}
	if updated.codes == nil {
		t.Fatal("codes should be set")
	}
	if updated.codes.ManualCode != "34970112332" {
		t.Errorf("ManualCode = %q, want %q", updated.codes.ManualCode, "34970112332")
	}
}

func TestMatterEditModel_CodesLoadedError(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	matter := &shelly.TUIMatterStatus{Enabled: true, Commissionable: true}
	m, _ = m.Show("192.168.1.100", matter)

	codesMsg := MatterCodesLoadedMsg{
		Err: errors.New("codes not available"),
	}
	updated, _ := m.Update(codesMsg)

	if updated.loadingCodes {
		t.Error("loadingCodes should be false")
	}
	if updated.codes != nil {
		t.Error("codes should be nil after error")
	}
	// Non-fatal error - should not set m.Err
	if updated.Err != nil {
		t.Error("err should not be set for codes load failure")
	}
}

func TestMatterEditModel_View_NotVisible(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	view := m.View()

	if view != "" {
		t.Error("View should be empty when not visible")
	}
}

func TestMatterEditModel_View_Enabled(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	m = m.SetSize(80, 40)
	matter := &shelly.TUIMatterStatus{
		Enabled:        true,
		Commissionable: false,
		FabricsCount:   2,
	}
	m, _ = m.Show("192.168.1.100", matter)

	view := m.View()

	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestMatterEditModel_View_Disabled(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	m = m.SetSize(80, 40)
	matter := &shelly.TUIMatterStatus{Enabled: false}
	m, _ = m.Show("192.168.1.100", matter)

	view := m.View()

	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestMatterEditModel_View_WithCodes(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	m = m.SetSize(80, 40)
	matter := &shelly.TUIMatterStatus{
		Enabled:        true,
		Commissionable: true,
	}
	m, _ = m.Show("192.168.1.100", matter)
	m.loadingCodes = false
	m.codes = &model.CommissioningInfo{
		ManualCode:    "34970112332",
		QRCode:        "MT:Y.K9042C00KA0648G00",
		SetupPINCode:  20202021,
		Discriminator: 3840,
		Available:     true,
	}

	view := m.View()

	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestMatterEditModel_View_Saving(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	m = m.SetSize(80, 40)
	matter := &shelly.TUIMatterStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", matter)
	m.Saving = true

	view := m.View()

	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestMatterEditModel_View_Resetting(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	m = m.SetSize(80, 40)
	matter := &shelly.TUIMatterStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", matter)
	m.resetting = true

	view := m.View()

	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestMatterEditModel_View_PendingReset(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	m = m.SetSize(80, 40)
	matter := &shelly.TUIMatterStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", matter)
	m.pendingReset = true
	m.Cursor = int(matterFieldReset)

	view := m.View()

	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestMatterEditModel_View_WithError(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	m = m.SetSize(80, 40)
	matter := &shelly.TUIMatterStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", matter)
	m.Err = errors.New("test error")

	view := m.View()

	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestMatterEditModel_ToggleBlockedWhileSaving(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	matter := &shelly.TUIMatterStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", matter)
	m.Saving = true

	// Toggle should be blocked
	spaceMsg := tea.KeyPressMsg{Code: tea.KeySpace}
	updated, _ := m.Update(spaceMsg)

	if !updated.pendingEnabled {
		t.Error("pendingEnabled should not change while saving")
	}
}

func TestMatterEditModel_KeyJNavigation(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	matter := &shelly.TUIMatterStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", matter)

	// Navigate down with j
	jMsg := tea.KeyPressMsg{Code: 'j'}
	updated, _ := m.Update(jMsg)

	if matterEditField(updated.Cursor) != matterFieldReset {
		t.Errorf("Cursor = %d, want %d after 'j'", updated.Cursor, matterFieldReset)
	}

	// Navigate up with k
	kMsg := tea.KeyPressMsg{Code: 'k'}
	updated, _ = updated.Update(kMsg)

	if matterEditField(updated.Cursor) != matterFieldEnable {
		t.Errorf("Cursor = %d, want %d after 'k'", updated.Cursor, matterFieldEnable)
	}
}

func TestMatterEditModel_ShowNilMatter(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	shown, cmd := m.Show("192.168.1.100", nil)

	if !shown.Visible() {
		t.Error("should be visible even with nil matter")
	}
	if shown.enabled {
		t.Error("enabled should be false with nil matter")
	}
	if cmd != nil {
		t.Error("should not fetch codes with nil matter")
	}
}

func TestMatterEditModel_CtrlSClose(t *testing.T) {
	t.Parallel()

	m := newTestMatterEditModel()
	matter := &shelly.TUIMatterStatus{Enabled: true}
	m, _ = m.Show("192.168.1.100", matter)

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
