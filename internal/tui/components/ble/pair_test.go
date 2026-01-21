package ble

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
)

func TestNewPairModel(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc := &shelly.Service{}

	m := NewPairModel(ctx, svc)

	if m.ctx != ctx {
		t.Error("ctx not set")
	}
	if m.svc != svc {
		t.Error("svc not set")
	}
	if m.visible {
		t.Error("should not be visible initially")
	}
}

func TestPairModel_Show(t *testing.T) {
	t.Parallel()
	m := newTestPairModel()

	updated, cmd := m.Show(testDevice)

	if !updated.visible {
		t.Error("should be visible after Show")
	}
	if updated.device != testDevice {
		t.Errorf("device = %q, want %q", updated.device, testDevice)
	}
	if updated.cursor != PairFieldAddr {
		t.Error("cursor should be on address field")
	}
	if cmd == nil {
		t.Error("Show should return a command")
	}
}

func TestPairModel_Hide(t *testing.T) {
	t.Parallel()
	m := newTestPairModel()
	m.visible = true

	updated := m.Hide()

	if updated.visible {
		t.Error("should not be visible after Hide")
	}
}

func TestPairModel_Visible(t *testing.T) {
	t.Parallel()
	m := newTestPairModel()

	if m.Visible() {
		t.Error("should not be visible initially")
	}

	m.visible = true

	if !m.Visible() {
		t.Error("should be visible")
	}
}

func TestPairModel_SetSize(t *testing.T) {
	t.Parallel()
	m := newTestPairModel()

	updated := m.SetSize(100, 50)

	if updated.width != 100 {
		t.Errorf("width = %d, want 100", updated.width)
	}
	if updated.height != 50 {
		t.Errorf("height = %d, want 50", updated.height)
	}
}

func TestPairModel_Init(t *testing.T) {
	t.Parallel()
	m := newTestPairModel()

	cmd := m.Init()

	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestPairModel_Update_NotVisible(t *testing.T) {
	t.Parallel()
	m := newTestPairModel()
	m.visible = false

	updated, cmd := m.Update(tea.KeyPressMsg{})

	if cmd != nil {
		t.Error("should return nil command when not visible")
	}
	if updated.visible {
		t.Error("should remain not visible")
	}
}

func TestPairModel_Update_DeviceAdded(t *testing.T) {
	t.Parallel()
	m := newTestPairModel()
	m.visible = true
	m.saving = true
	msg := DeviceAddedMsg{Key: "bthomesensor:200"}

	updated, cmd := m.Update(msg)

	if updated.saving {
		t.Error("should not be saving after DeviceAddedMsg")
	}
	if updated.visible {
		t.Error("should be hidden after successful add")
	}
	if cmd == nil {
		t.Error("should return PairClosedMsg command")
	}
}

func TestPairModel_Update_DeviceAddedError(t *testing.T) {
	t.Parallel()
	m := newTestPairModel()
	m.visible = true
	m.saving = true
	testErr := errors.New("add failed")
	msg := DeviceAddedMsg{Err: testErr}

	updated, cmd := m.Update(msg)

	if updated.saving {
		t.Error("should not be saving after error")
	}
	if updated.err == nil {
		t.Error("err should be set")
	}
	if !updated.visible {
		t.Error("should remain visible after error")
	}
	if cmd != nil {
		t.Error("should not return command after error")
	}
}

func TestPairModel_HandleKey_Escape(t *testing.T) {
	t.Parallel()
	m := newTestPairModel()
	m.visible = true
	msg := tea.KeyPressMsg{Code: tea.KeyEscape}

	updated, cmd := m.Update(msg)

	if updated.visible {
		t.Error("should be hidden after escape")
	}
	if cmd == nil {
		t.Error("should return PairClosedMsg command")
	}
}

func TestPairModel_HandleKey_CtrlOpenBracket(t *testing.T) {
	t.Parallel()
	m := newTestPairModel()
	m.visible = true
	// Create key press for ctrl+[
	msg := tea.KeyPressMsg{Code: '[', Mod: tea.ModCtrl}

	updated, cmd := m.Update(msg)

	if updated.visible {
		t.Error("should be hidden after ctrl+[")
	}
	if cmd == nil {
		t.Error("should return PairClosedMsg command")
	}
}

func TestPairModel_HandleKey_Tab(t *testing.T) {
	t.Parallel()
	m := newTestPairModel()
	m.visible = true
	m.cursor = PairFieldAddr
	msg := tea.KeyPressMsg{Code: tea.KeyTab}

	updated, _ := m.Update(msg)

	if updated.cursor != PairFieldName {
		t.Errorf("cursor = %d, want %d", updated.cursor, PairFieldName)
	}
}

func TestPairModel_HandleKey_ShiftTab(t *testing.T) {
	t.Parallel()
	m := newTestPairModel()
	m.visible = true
	m.cursor = PairFieldName
	msg := tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}

	updated, _ := m.Update(msg)

	if updated.cursor != PairFieldAddr {
		t.Errorf("cursor = %d, want %d", updated.cursor, PairFieldAddr)
	}
}

func TestPairModel_HandleNavigation_Up(t *testing.T) {
	t.Parallel()
	m := newTestPairModel()
	m.visible = true
	m.cursor = PairFieldName
	msg := messages.NavigationMsg{Direction: messages.NavUp}

	updated, _ := m.Update(msg)

	if updated.cursor != PairFieldAddr {
		t.Errorf("cursor = %d, want %d", updated.cursor, PairFieldAddr)
	}
}

func TestPairModel_HandleNavigation_Down(t *testing.T) {
	t.Parallel()
	m := newTestPairModel()
	m.visible = true
	m.cursor = PairFieldAddr
	msg := messages.NavigationMsg{Direction: messages.NavDown}

	updated, _ := m.Update(msg)

	if updated.cursor != PairFieldName {
		t.Errorf("cursor = %d, want %d", updated.cursor, PairFieldName)
	}
}

func TestPairModel_HandleNavigation_Saving(t *testing.T) {
	t.Parallel()
	m := newTestPairModel()
	m.visible = true
	m.saving = true
	m.cursor = PairFieldAddr
	msg := messages.NavigationMsg{Direction: messages.NavDown}

	updated, _ := m.Update(msg)

	if updated.cursor != PairFieldAddr {
		t.Error("cursor should not change when saving")
	}
}

func TestPairModel_NextField(t *testing.T) {
	t.Parallel()
	m := newTestPairModel()
	m.cursor = PairFieldAddr

	updated := m.nextField()

	if updated.cursor != PairFieldName {
		t.Errorf("cursor = %d, want %d", updated.cursor, PairFieldName)
	}
}

func TestPairModel_NextField_Wrap(t *testing.T) {
	t.Parallel()
	m := newTestPairModel()
	m.cursor = PairFieldName

	updated := m.nextField()

	if updated.cursor != PairFieldAddr {
		t.Errorf("cursor = %d, want %d (wrap to first)", updated.cursor, PairFieldAddr)
	}
}

func TestPairModel_PrevField(t *testing.T) {
	t.Parallel()
	m := newTestPairModel()
	m.cursor = PairFieldName

	updated := m.prevField()

	if updated.cursor != PairFieldAddr {
		t.Errorf("cursor = %d, want %d", updated.cursor, PairFieldAddr)
	}
}

func TestPairModel_PrevField_Wrap(t *testing.T) {
	t.Parallel()
	m := newTestPairModel()
	m.cursor = PairFieldAddr

	updated := m.prevField()

	if updated.cursor != PairFieldName {
		t.Errorf("cursor = %d, want %d (wrap to last)", updated.cursor, PairFieldName)
	}
}

func TestPairModel_View_NotVisible(t *testing.T) {
	t.Parallel()
	m := newTestPairModel()
	m.visible = false

	view := m.View()

	if view != "" {
		t.Error("View should return empty string when not visible")
	}
}

func TestPairModel_View_Visible(t *testing.T) {
	t.Parallel()
	m := newTestPairModel()
	m.visible = true
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View should not return empty string when visible")
	}
}

func TestPairModel_View_Saving(t *testing.T) {
	t.Parallel()
	m := newTestPairModel()
	m.visible = true
	m.saving = true
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View should not return empty string when saving")
	}
}

func TestPairModel_View_Error(t *testing.T) {
	t.Parallel()
	m := newTestPairModel()
	m.visible = true
	m.err = errors.New("test error")
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View should not return empty string with error")
	}
}

func TestPairModel_Device(t *testing.T) {
	t.Parallel()
	m := newTestPairModel()
	m.device = testDevice

	if m.Device() != testDevice {
		t.Errorf("Device() = %q, want %q", m.Device(), testDevice)
	}
}

func TestIsValidMAC(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		addr  string
		valid bool
	}{
		{"valid colon", "AA:BB:CC:DD:EE:FF", true},
		{"valid lowercase", "aa:bb:cc:dd:ee:ff", true},
		{"valid mixed", "Aa:Bb:Cc:Dd:Ee:Ff", true},
		{"valid hyphen", "AA-BB-CC-DD-EE-FF", true},
		{"invalid short", "AA:BB:CC:DD:EE", false},
		{"invalid long", "AA:BB:CC:DD:EE:FF:00", false},
		{"invalid chars", "GG:HH:II:JJ:KK:LL", false},
		{"empty", "", false},
		{"spaces", "AA: BB:CC:DD:EE:FF", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isValidMAC(tt.addr)
			if got != tt.valid {
				t.Errorf("isValidMAC(%q) = %v, want %v", tt.addr, got, tt.valid)
			}
		})
	}
}

func TestPairModel_Save_EmptyAddress(t *testing.T) {
	t.Parallel()
	m := newTestPairModel()
	m.visible = true
	m.addrInput = m.addrInput.SetValue("")

	updated, cmd := m.save()

	if updated.err == nil {
		t.Error("err should be set for empty address")
	}
	if !errors.Is(updated.err, errMACRequired) {
		t.Errorf("err = %v, want %v", updated.err, errMACRequired)
	}
	if cmd != nil {
		t.Error("should not return command for validation error")
	}
}

func TestPairModel_Save_InvalidAddress(t *testing.T) {
	t.Parallel()
	m := newTestPairModel()
	m.visible = true
	m.addrInput = m.addrInput.SetValue("invalid")

	updated, cmd := m.save()

	if updated.err == nil {
		t.Error("err should be set for invalid address")
	}
	if !errors.Is(updated.err, errMACInvalid) {
		t.Errorf("err = %v, want %v", updated.err, errMACInvalid)
	}
	if cmd != nil {
		t.Error("should not return command for validation error")
	}
}

func TestPairModel_Save_ValidAddress(t *testing.T) {
	t.Parallel()
	m := newTestPairModel()
	m.visible = true
	m.device = testDevice
	m.addrInput = m.addrInput.SetValue("AA:BB:CC:DD:EE:FF")
	m.nameInput = m.nameInput.SetValue("Test Sensor")

	updated, cmd := m.save()

	if updated.err != nil {
		t.Errorf("unexpected err: %v", updated.err)
	}
	if !updated.saving {
		t.Error("should be saving")
	}
	if cmd == nil {
		t.Error("should return command for valid save")
	}
}

func TestPairModel_KeyConst_Enter(t *testing.T) {
	t.Parallel()
	// Verify keyconst.KeyEnter is used correctly
	if keyconst.KeyEnter != "enter" {
		t.Errorf("keyconst.KeyEnter = %q, want %q", keyconst.KeyEnter, "enter")
	}
}

func TestPairModel_KeyConst_CtrlS(t *testing.T) {
	t.Parallel()
	// Verify keyconst.KeyCtrlS is used correctly
	if keyconst.KeyCtrlS != "ctrl+s" {
		t.Errorf("keyconst.KeyCtrlS = %q, want %q", keyconst.KeyCtrlS, "ctrl+s")
	}
}

func newTestPairModel() PairModel {
	ctx := context.Background()
	svc := &shelly.Service{}
	return NewPairModel(ctx, svc)
}
