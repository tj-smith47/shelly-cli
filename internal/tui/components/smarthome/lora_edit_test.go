package smarthome

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
)

func newTestLoRaEditModel() LoRaEditModel {
	ctx := context.Background()
	svc := &shelly.Service{}
	return NewLoRaEditModel(ctx, svc)
}

func testLoRaStatus() *shelly.TUILoRaStatus {
	return &shelly.TUILoRaStatus{
		Enabled:   true,
		Frequency: 868000000,
		Bandwidth: 125,
		DataRate:  7,
		TxPower:   14,
		RSSI:      -70,
		SNR:       8.5,
	}
}

func TestNewLoRaEditModel(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()

	if m.Visible() {
		t.Error("should not be visible initially")
	}
}

func TestLoRaEditModel_Show(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	lora := testLoRaStatus()

	shown, cmd := m.Show(testDevice, lora)
	if !shown.Visible() {
		t.Error("should be visible after Show")
	}
	if shown.device != testDevice {
		t.Errorf("device = %q, want %q", shown.device, testDevice)
	}
	if shown.freq != 868000000 {
		t.Errorf("freq = %d, want 868000000", shown.freq)
	}
	if shown.bw != 125 {
		t.Errorf("bw = %d, want 125", shown.bw)
	}
	if shown.dr != 7 {
		t.Errorf("dr = %d, want 7", shown.dr)
	}
	if shown.txp != 14 {
		t.Errorf("txp = %d, want 14", shown.txp)
	}
	if shown.rssi != -70 {
		t.Errorf("rssi = %d, want -70", shown.rssi)
	}
	if shown.snr != 8.5 {
		t.Errorf("snr = %f, want 8.5", shown.snr)
	}
	if cmd != nil {
		t.Error("should not return command (no async loading)")
	}
	if shown.field != loraFieldFreq {
		t.Errorf("field = %d, want %d", shown.field, loraFieldFreq)
	}
}

func TestLoRaEditModel_ShowSetsPending(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	lora := testLoRaStatus()

	shown, _ := m.Show(testDevice, lora)
	if shown.pendingFreq != 868000000 {
		t.Error("pendingFreq should match freq")
	}
	if shown.pendingBW != 125 {
		t.Error("pendingBW should match bw")
	}
	if shown.pendingDR != 7 {
		t.Error("pendingDR should match dr")
	}
	if shown.pendingTxP != 14 {
		t.Error("pendingTxP should match txp")
	}
}

func TestLoRaEditModel_Hide(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	lora := testLoRaStatus()

	shown, _ := m.Show(testDevice, lora)
	hidden := shown.Hide()
	if hidden.Visible() {
		t.Error("should not be visible after Hide")
	}
}

func TestLoRaEditModel_ShowNilLoRa(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	shown, cmd := m.Show(testDevice, nil)

	if !shown.Visible() {
		t.Error("should be visible even with nil lora")
	}
	if shown.freq != 0 {
		t.Error("freq should be 0 with nil lora")
	}
	if shown.bw != 0 {
		t.Error("bw should be 0 with nil lora")
	}
	if cmd != nil {
		t.Error("should not return command with nil lora")
	}
}

func TestLoRaEditModel_ShowResetsState(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m.saving = true
	m.sending = true
	m.err = errors.New("old error")

	shown, _ := m.Show(testDevice, testLoRaStatus())

	if shown.saving {
		t.Error("saving should be reset")
	}
	if shown.sending {
		t.Error("sending should be reset")
	}
	if shown.err != nil {
		t.Error("err should be reset")
	}
}

func TestLoRaEditModel_SetSize(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m = m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}

func TestLoRaEditModel_UpdateNotVisible(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	keyMsg := tea.KeyPressMsg{Code: tea.KeyEscape}
	updated, cmd := m.Update(keyMsg)

	if updated.Visible() {
		t.Error("should remain hidden")
	}
	if cmd != nil {
		t.Error("cmd should be nil when not visible")
	}
}

// --- Close tests ---

func TestLoRaEditModel_EscClose(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())

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

// --- Navigation tests ---

func TestLoRaEditModel_NavigationUpDown(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())

	// Should start at freq field
	if m.field != loraFieldFreq {
		t.Errorf("field = %d, want %d", m.field, loraFieldFreq)
	}

	// Navigate up at top - should stay at top
	updated, _ := m.Update(messages.NavigationMsg{Direction: messages.NavUp})
	if updated.field != loraFieldFreq {
		t.Errorf("field should stay at freq, got %d", updated.field)
	}

	// Navigate down through all fields
	updated, _ = m.Update(messages.NavigationMsg{Direction: messages.NavDown})
	if updated.field != loraFieldBW {
		t.Errorf("field = %d, want %d (bw)", updated.field, loraFieldBW)
	}

	updated, _ = updated.Update(messages.NavigationMsg{Direction: messages.NavDown})
	if updated.field != loraFieldDR {
		t.Errorf("field = %d, want %d (dr)", updated.field, loraFieldDR)
	}

	updated, _ = updated.Update(messages.NavigationMsg{Direction: messages.NavDown})
	if updated.field != loraFieldTxP {
		t.Errorf("field = %d, want %d (txp)", updated.field, loraFieldTxP)
	}

	updated, _ = updated.Update(messages.NavigationMsg{Direction: messages.NavDown})
	if updated.field != loraFieldTestSend {
		t.Errorf("field = %d, want %d (testSend)", updated.field, loraFieldTestSend)
	}

	// Navigate down at bottom - should stay
	updated, _ = updated.Update(messages.NavigationMsg{Direction: messages.NavDown})
	if updated.field != loraFieldTestSend {
		t.Errorf("field should stay at testSend, got %d", updated.field)
	}

	// Navigate up from bottom
	updated, _ = updated.Update(messages.NavigationMsg{Direction: messages.NavUp})
	if updated.field != loraFieldTxP {
		t.Errorf("field = %d, want %d (txp)", updated.field, loraFieldTxP)
	}
}

func TestLoRaEditModel_KeyJKNavigation(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())

	// Navigate down with j
	updated, _ := m.Update(tea.KeyPressMsg{Code: 'j'})
	if updated.field != loraFieldBW {
		t.Errorf("field = %d, want %d after 'j'", updated.field, loraFieldBW)
	}

	// Navigate up with k
	updated, _ = updated.Update(tea.KeyPressMsg{Code: 'k'})
	if updated.field != loraFieldFreq {
		t.Errorf("field = %d, want %d after 'k'", updated.field, loraFieldFreq)
	}

	// k at top should stay
	updated, _ = updated.Update(tea.KeyPressMsg{Code: 'k'})
	if updated.field != loraFieldFreq {
		t.Error("should stay at top after k at top")
	}
}

func TestLoRaEditModel_NavigationNonApplicable(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())

	// PageUp/PageDown/Home/End should be no-ops for navigation
	directions := []messages.NavDirection{
		messages.NavPageUp,
		messages.NavPageDown,
		messages.NavHome,
		messages.NavEnd,
	}

	for _, dir := range directions {
		updated, _ := m.Update(messages.NavigationMsg{Direction: dir})
		if updated.field != loraFieldFreq {
			t.Errorf("field changed for non-applicable direction %d", dir)
		}
	}
}

// --- Value adjustment tests ---

func TestLoRaEditModel_AdjustFrequency(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())
	m.field = loraFieldFreq

	// Increase frequency by 100 kHz
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	if updated.pendingFreq != 868100000 {
		t.Errorf("pendingFreq = %d, want 868100000", updated.pendingFreq)
	}

	// Decrease frequency by 100 kHz
	updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	if updated.pendingFreq != 868000000 {
		t.Errorf("pendingFreq = %d, want 868000000", updated.pendingFreq)
	}
}

func TestLoRaEditModel_AdjustFrequencyHLKeys(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())
	m.field = loraFieldFreq

	// Increase with 'l' key
	updated, _ := m.Update(tea.KeyPressMsg{Code: 'l'})
	if updated.pendingFreq != 868100000 {
		t.Errorf("pendingFreq = %d, want 868100000 after 'l'", updated.pendingFreq)
	}

	// Decrease with 'h' key
	updated, _ = updated.Update(tea.KeyPressMsg{Code: 'h'})
	if updated.pendingFreq != 868000000 {
		t.Errorf("pendingFreq = %d, want 868000000 after 'h'", updated.pendingFreq)
	}
}

func TestLoRaEditModel_AdjustFrequencyBounds(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	lora := testLoRaStatus()
	lora.Frequency = loraFreqMin
	m, _ = m.Show(testDevice, lora)
	m.field = loraFieldFreq

	// Decrease below min - should not change
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	if updated.pendingFreq != loraFreqMin {
		t.Errorf("pendingFreq = %d, should not go below %d", updated.pendingFreq, loraFreqMin)
	}

	// Set to max and increase - should not change
	m.pendingFreq = loraFreqMax
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	if updated.pendingFreq != loraFreqMax {
		t.Errorf("pendingFreq = %d, should not go above %d", updated.pendingFreq, loraFreqMax)
	}
}

func TestLoRaEditModel_AdjustFrequencyNavMsg(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())
	m.field = loraFieldFreq

	// NavLeft/NavRight should also adjust values
	updated, _ := m.Update(messages.NavigationMsg{Direction: messages.NavRight})
	if updated.pendingFreq != 868100000 {
		t.Errorf("pendingFreq = %d, want 868100000 after NavRight", updated.pendingFreq)
	}

	updated, _ = updated.Update(messages.NavigationMsg{Direction: messages.NavLeft})
	if updated.pendingFreq != 868000000 {
		t.Errorf("pendingFreq = %d, want 868000000 after NavLeft", updated.pendingFreq)
	}
}

func TestLoRaEditModel_AdjustBandwidth(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())
	m.field = loraFieldBW

	// Toggle from 125 to 250
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	if updated.pendingBW != 250 {
		t.Errorf("pendingBW = %d, want 250", updated.pendingBW)
	}

	// Toggle back to 125
	updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	if updated.pendingBW != 125 {
		t.Errorf("pendingBW = %d, want 125", updated.pendingBW)
	}

	// Left also toggles
	updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	if updated.pendingBW != 250 {
		t.Errorf("pendingBW = %d, want 250 after left", updated.pendingBW)
	}
}

func TestLoRaEditModel_AdjustDataRate(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())
	m.field = loraFieldDR

	// Increase from SF7 to SF8
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	if updated.pendingDR != 8 {
		t.Errorf("pendingDR = %d, want 8", updated.pendingDR)
	}

	// Decrease back to SF7
	updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	if updated.pendingDR != 7 {
		t.Errorf("pendingDR = %d, want 7", updated.pendingDR)
	}
}

func TestLoRaEditModel_AdjustDataRateBounds(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())
	m.field = loraFieldDR

	// At SF7, decrease should not change
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	if updated.pendingDR != 7 {
		t.Errorf("pendingDR = %d, should not go below 7", updated.pendingDR)
	}

	// Set to SF12 and increase - should not change
	m.pendingDR = 12
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	if updated.pendingDR != 12 {
		t.Errorf("pendingDR = %d, should not go above 12", updated.pendingDR)
	}
}

func TestLoRaEditModel_AdjustTxPower(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())
	m.field = loraFieldTxP

	// Decrease from 14 to 13
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	if updated.pendingTxP != 13 {
		t.Errorf("pendingTxP = %d, want 13", updated.pendingTxP)
	}

	// Increase back to 14
	updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	if updated.pendingTxP != 14 {
		t.Errorf("pendingTxP = %d, want 14", updated.pendingTxP)
	}
}

func TestLoRaEditModel_AdjustTxPowerBounds(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	lora := testLoRaStatus()
	lora.TxPower = 0
	m, _ = m.Show(testDevice, lora)
	m.field = loraFieldTxP

	// At 0, decrease should not change
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	if updated.pendingTxP != 0 {
		t.Errorf("pendingTxP = %d, should not go below 0", updated.pendingTxP)
	}

	// Set to 14 and increase - should not change
	m.pendingTxP = 14
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	if updated.pendingTxP != 14 {
		t.Errorf("pendingTxP = %d, should not go above 14", updated.pendingTxP)
	}
}

func TestLoRaEditModel_AdjustBlockedWhileSaving(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())
	m.field = loraFieldFreq
	m.saving = true

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	if updated.pendingFreq != 868000000 {
		t.Error("frequency should not change while saving")
	}
}

func TestLoRaEditModel_AdjustBlockedWhileSending(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())
	m.field = loraFieldFreq
	m.sending = true

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	if updated.pendingFreq != 868000000 {
		t.Error("frequency should not change while sending")
	}
}

func TestLoRaEditModel_AdjustTestSendField(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())
	m.field = loraFieldTestSend

	// Arrow keys on test send button should not change any values
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	if updated.pendingFreq != 868000000 || updated.pendingBW != 125 ||
		updated.pendingDR != 7 || updated.pendingTxP != 14 {
		t.Error("values should not change when adjusting test send button field")
	}
}

// --- hasChanges tests ---

func TestLoRaEditModel_HasChanges(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())

	if m.hasChanges() {
		t.Error("should have no changes initially")
	}

	m.pendingFreq = 869000000
	if !m.hasChanges() {
		t.Error("should have changes after modifying frequency")
	}

	m.pendingFreq = m.freq
	m.pendingBW = 250
	if !m.hasChanges() {
		t.Error("should have changes after modifying bandwidth")
	}

	m.pendingBW = m.bw
	m.pendingDR = 10
	if !m.hasChanges() {
		t.Error("should have changes after modifying data rate")
	}

	m.pendingDR = m.dr
	m.pendingTxP = 5
	if !m.hasChanges() {
		t.Error("should have changes after modifying tx power")
	}
}

// --- Save flow tests ---

func TestLoRaEditModel_SaveNoChanges(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())

	// Save without changes
	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

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

func TestLoRaEditModel_SaveWithChanges(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())
	m.pendingFreq = 869000000

	// Save
	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if !updated.saving {
		t.Error("should be saving after Enter with changes")
	}
	if cmd == nil {
		t.Error("should return save command")
	}
}

func TestLoRaEditModel_SaveResultSuccess(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())
	m.saving = true
	m.pendingFreq = 869000000

	saveMsg := messages.SaveResultMsg{Success: true}
	updated, cmd := m.Update(saveMsg)

	if updated.saving {
		t.Error("saving should be false after success")
	}
	if updated.Visible() {
		t.Error("should be hidden after successful save")
	}
	if updated.freq != 869000000 {
		t.Errorf("freq should be updated to pending value, got %d", updated.freq)
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

func TestLoRaEditModel_SaveResultError(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())
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

func TestLoRaEditModel_SaveBlockedWhileSaving(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())
	m.saving = true
	m.pendingFreq = 869000000

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if !updated.saving {
		t.Error("saving should remain true")
	}
	if cmd != nil {
		t.Error("should not return command when already saving")
	}
}

func TestLoRaEditModel_SaveBlockedWhileSending(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())
	m.sending = true
	m.pendingFreq = 869000000

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if updated.saving {
		t.Error("should not start saving while sending")
	}
	if cmd != nil {
		t.Error("should not return command when sending")
	}
}

func TestLoRaEditModel_CtrlSSave(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())

	// Ctrl+S with no changes should close
	updated, cmd := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})

	if updated.Visible() {
		t.Error("should be hidden after Ctrl+S with no changes")
	}
	if cmd == nil {
		t.Fatal("should return close cmd")
	}
}

// --- Test send tests ---

func TestLoRaEditModel_TestSendViaKey(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())

	// T key triggers test send
	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'T'})

	if !updated.sending {
		t.Error("should be sending after T key")
	}
	if cmd == nil {
		t.Error("should return test send command")
	}
}

func TestLoRaEditModel_TestSendViaButton(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())
	m.field = loraFieldTestSend

	// Enter on test send button
	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if !updated.sending {
		t.Error("should be sending after Enter on test send button")
	}
	if cmd == nil {
		t.Error("should return test send command")
	}
}

func TestLoRaEditModel_TestSendBlockedWhileSaving(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())
	m.saving = true

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'T'})

	if updated.sending {
		t.Error("should not send while saving")
	}
	if cmd != nil {
		t.Error("should not return command while saving")
	}
}

func TestLoRaEditModel_TestSendBlockedWhileSending(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())
	m.sending = true

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'T'})

	if cmd != nil {
		t.Error("should not return command when already sending")
	}
	_ = updated
}

func TestLoRaEditModel_TestSendResultSuccess(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())
	m.sending = true

	updated, _ := m.Update(LoRaTestSendResultMsg{Err: nil})

	if updated.sending {
		t.Error("sending should be false after success")
	}
	if updated.err != nil {
		t.Error("err should be nil after success")
	}
	if !updated.Visible() {
		t.Error("should remain visible after test send")
	}
}

func TestLoRaEditModel_TestSendResultError(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m, _ = m.Show(testDevice, testLoRaStatus())
	m.sending = true

	sendErr := errors.New("send failed")
	updated, _ := m.Update(LoRaTestSendResultMsg{Err: sendErr})

	if updated.sending {
		t.Error("sending should be false after error")
	}
	if updated.err == nil {
		t.Error("err should be set")
	}
	if !updated.Visible() {
		t.Error("should remain visible after error")
	}
}

// --- View rendering tests ---

func TestLoRaEditModel_View_NotVisible(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	view := m.View()

	if view != "" {
		t.Error("View should be empty when not visible")
	}
}

func TestLoRaEditModel_View_Active(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m = m.SetSize(80, 40)
	m, _ = m.Show(testDevice, testLoRaStatus())

	view := m.View()

	if view == "" {
		t.Error("View should not be empty")
	}
	if !strings.Contains(view, "LoRa Configuration") {
		t.Error("View should contain title 'LoRa Configuration'")
	}
	if !strings.Contains(view, "868.00 MHz") {
		t.Error("View should contain frequency")
	}
	if !strings.Contains(view, "125 kHz") {
		t.Error("View should contain bandwidth")
	}
	if !strings.Contains(view, "SF7") {
		t.Error("View should contain data rate")
	}
	if !strings.Contains(view, "14 dBm") {
		t.Error("View should contain TX power")
	}
	if !strings.Contains(view, "Send Test Packet") {
		t.Error("View should contain test send button")
	}
}

func TestLoRaEditModel_View_WithRSSI(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m = m.SetSize(80, 40)
	m, _ = m.Show(testDevice, testLoRaStatus())

	view := m.View()

	if !strings.Contains(view, "-70 dBm") {
		t.Error("View should contain RSSI value")
	}
	if !strings.Contains(view, "8.5 dB") {
		t.Error("View should contain SNR value")
	}
}

func TestLoRaEditModel_View_WithoutRSSI(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m = m.SetSize(80, 40)
	lora := testLoRaStatus()
	lora.RSSI = 0
	lora.SNR = 0
	m, _ = m.Show(testDevice, lora)

	view := m.View()

	if strings.Contains(view, "Last RSSI") {
		t.Error("View should not contain RSSI label when RSSI is 0")
	}
}

func TestLoRaEditModel_View_Saving(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m = m.SetSize(80, 40)
	m, _ = m.Show(testDevice, testLoRaStatus())
	m.saving = true

	view := m.View()

	if !strings.Contains(view, "Saving") {
		t.Error("View should contain 'Saving' when saving")
	}
}

func TestLoRaEditModel_View_Sending(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m = m.SetSize(80, 40)
	m, _ = m.Show(testDevice, testLoRaStatus())
	m.sending = true

	view := m.View()

	if !strings.Contains(view, "Sending") {
		t.Error("View should contain 'Sending' when sending test packet")
	}
}

func TestLoRaEditModel_View_WithError(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m = m.SetSize(80, 40)
	m, _ = m.Show(testDevice, testLoRaStatus())
	m.err = errors.New("test error")

	view := m.View()

	if view == "" {
		t.Error("View should not be empty with error")
	}
}

func TestLoRaEditModel_View_PendingChanges(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m = m.SetSize(80, 40)
	m, _ = m.Show(testDevice, testLoRaStatus())
	m.pendingFreq = 869000000

	view := m.View()

	if !strings.Contains(view, "changed") {
		t.Error("View should show change indicator when config has changed")
	}
}

func TestLoRaEditModel_View_FieldSelectors(t *testing.T) {
	t.Parallel()

	m := newTestLoRaEditModel()
	m = m.SetSize(80, 40)
	m, _ = m.Show(testDevice, testLoRaStatus())

	// Freq should have arrows when selected
	m.field = loraFieldFreq
	view := m.View()
	if !strings.Contains(view, "◀ ▶") {
		t.Error("Selected field should show arrow indicators")
	}
}
