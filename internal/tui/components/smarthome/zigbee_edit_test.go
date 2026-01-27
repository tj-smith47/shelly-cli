package smarthome

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
)

var errTest = errors.New("test error")

func TestNewZigbeeEditModel(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc := &shelly.Service{}

	m := NewZigbeeEditModel(ctx, svc)

	if m.visible {
		t.Error("should not be visible initially")
	}
	if m.ctx != ctx {
		t.Error("ctx not set")
	}
	if m.svc != svc {
		t.Error("svc not set")
	}
}

func TestZigbeeEditModel_Show(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()

	zigbee := &shelly.TUIZigbeeStatus{
		Enabled:          true,
		NetworkState:     zigbeeStateJoined,
		Channel:          15,
		PANID:            0x1234,
		EUI64:            "00:11:22:33:44:55:66:77",
		CoordinatorEUI64: "AA:BB:CC:DD:EE:FF:00:11",
	}

	m, _ = m.Show(testDevice, zigbee)

	if !m.visible {
		t.Error("should be visible after Show")
	}
	if m.device != testDevice {
		t.Errorf("device = %q, want %q", m.device, testDevice)
	}
	if !m.enabled {
		t.Error("enabled should be true")
	}
	if m.networkState != zigbeeStateJoined {
		t.Errorf("networkState = %q, want %q", m.networkState, zigbeeStateJoined)
	}
	if m.channel != 15 {
		t.Errorf("channel = %d, want 15", m.channel)
	}
	if m.panID != 0x1234 {
		t.Errorf("panID = 0x%04X, want 0x1234", m.panID)
	}
	if m.eui64 != "00:11:22:33:44:55:66:77" {
		t.Errorf("eui64 = %q, want expected value", m.eui64)
	}
	if m.coordinator != "AA:BB:CC:DD:EE:FF:00:11" {
		t.Errorf("coordinator = %q, want expected value", m.coordinator)
	}
	if !m.pendingEnabled {
		t.Error("pendingEnabled should match enabled")
	}
	if m.saving || m.steering || m.leaving {
		t.Error("operation flags should be false")
	}
}

func TestZigbeeEditModel_Show_Disabled(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()

	zigbee := &shelly.TUIZigbeeStatus{Enabled: false}
	m, _ = m.Show(testDevice, zigbee)

	if m.enabled {
		t.Error("enabled should be false")
	}
	if m.pendingEnabled {
		t.Error("pendingEnabled should be false")
	}
	if m.fieldCount != 1 {
		t.Errorf("fieldCount = %d, want 1 (only enable toggle)", m.fieldCount)
	}
}

func TestZigbeeEditModel_Show_EnabledNotJoined(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()

	zigbee := &shelly.TUIZigbeeStatus{
		Enabled:      true,
		NetworkState: zigbeeStateReady,
	}
	m, _ = m.Show(testDevice, zigbee)

	// Should have 2 fields: enable + pair
	if m.fieldCount != 2 {
		t.Errorf("fieldCount = %d, want 2 (enable + pair)", m.fieldCount)
	}
}

func TestZigbeeEditModel_Show_Joined(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()

	zigbee := &shelly.TUIZigbeeStatus{
		Enabled:      true,
		NetworkState: zigbeeStateJoined,
		Channel:      15,
		PANID:        0x1234,
	}
	m, _ = m.Show(testDevice, zigbee)

	// Should have 3 fields: enable + pair + leave
	if m.fieldCount != 3 {
		t.Errorf("fieldCount = %d, want 3 (enable + pair + leave)", m.fieldCount)
	}
}

func TestZigbeeEditModel_Hide(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()
	zigbee := &shelly.TUIZigbeeStatus{Enabled: true, NetworkState: zigbeeStateJoined}
	m, _ = m.Show(testDevice, zigbee)
	m.pendingLeave = true

	m = m.Hide()

	if m.visible {
		t.Error("should not be visible after Hide")
	}
	if m.pendingLeave {
		t.Error("pendingLeave should be cleared on Hide")
	}
}

func TestZigbeeEditModel_Visible(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()

	if m.Visible() {
		t.Error("should not be visible initially")
	}

	m.visible = true
	if !m.Visible() {
		t.Error("should be visible")
	}
}

func TestZigbeeEditModel_SetSize(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()

	m = m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}

func TestZigbeeEditModel_Toggle(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()
	zigbee := &shelly.TUIZigbeeStatus{Enabled: true, NetworkState: zigbeeStateReady}
	m, _ = m.Show(testDevice, zigbee)

	// Toggle via 't' key
	m, _ = m.Update(tea.KeyPressMsg{Code: 't'})

	if m.pendingEnabled {
		t.Error("pendingEnabled should be false after toggle")
	}

	// Toggle back via space
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeySpace})

	if !m.pendingEnabled {
		t.Error("pendingEnabled should be true after toggle back")
	}
}

func TestZigbeeEditModel_Toggle_NotOnEnableField(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()
	zigbee := &shelly.TUIZigbeeStatus{Enabled: true, NetworkState: zigbeeStateReady}
	m, _ = m.Show(testDevice, zigbee)

	// Move to pair field
	m.field = zigbeeFieldPair
	originalEnabled := m.pendingEnabled

	// Toggle should not work on pair field
	m, _ = m.Update(tea.KeyPressMsg{Code: 't'})

	if m.pendingEnabled != originalEnabled {
		t.Error("toggle should not work on non-enable field")
	}
}

func TestZigbeeEditModel_Navigation(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()
	zigbee := &shelly.TUIZigbeeStatus{
		Enabled:      true,
		NetworkState: zigbeeStateJoined,
	}
	m, _ = m.Show(testDevice, zigbee)

	// Start on enable field
	if m.field != zigbeeFieldEnable {
		t.Errorf("initial field = %d, want %d", m.field, zigbeeFieldEnable)
	}

	// Navigate down
	m, _ = m.Update(tea.KeyPressMsg{Code: 'j'})
	if m.field != zigbeeFieldPair {
		t.Errorf("after j, field = %d, want %d", m.field, zigbeeFieldPair)
	}

	// Navigate down again
	m, _ = m.Update(tea.KeyPressMsg{Code: 'j'})
	if m.field != zigbeeFieldLeave {
		t.Errorf("after second j, field = %d, want %d", m.field, zigbeeFieldLeave)
	}

	// Navigate down at bottom (should stay)
	m, _ = m.Update(tea.KeyPressMsg{Code: 'j'})
	if m.field != zigbeeFieldLeave {
		t.Errorf("at bottom, field = %d, want %d", m.field, zigbeeFieldLeave)
	}

	// Navigate up
	m, _ = m.Update(tea.KeyPressMsg{Code: 'k'})
	if m.field != zigbeeFieldPair {
		t.Errorf("after k, field = %d, want %d", m.field, zigbeeFieldPair)
	}

	// Navigate up to top
	m, _ = m.Update(tea.KeyPressMsg{Code: 'k'})
	if m.field != zigbeeFieldEnable {
		t.Errorf("at top, field = %d, want %d", m.field, zigbeeFieldEnable)
	}

	// Navigate up at top (should stay)
	m, _ = m.Update(tea.KeyPressMsg{Code: 'k'})
	if m.field != zigbeeFieldEnable {
		t.Errorf("past top, field = %d, want %d", m.field, zigbeeFieldEnable)
	}
}

func TestZigbeeEditModel_NavigationMsg(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()
	zigbee := &shelly.TUIZigbeeStatus{
		Enabled:      true,
		NetworkState: zigbeeStateJoined,
	}
	m, _ = m.Show(testDevice, zigbee)

	// Navigate via NavigationMsg
	m, _ = m.Update(messages.NavigationMsg{Direction: messages.NavDown})
	if m.field != zigbeeFieldPair {
		t.Errorf("after NavDown, field = %d, want %d", m.field, zigbeeFieldPair)
	}

	m, _ = m.Update(messages.NavigationMsg{Direction: messages.NavUp})
	if m.field != zigbeeFieldEnable {
		t.Errorf("after NavUp, field = %d, want %d", m.field, zigbeeFieldEnable)
	}
}

func TestZigbeeEditModel_NavigationClearsPendingLeave(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()
	zigbee := &shelly.TUIZigbeeStatus{
		Enabled:      true,
		NetworkState: zigbeeStateJoined,
	}
	m, _ = m.Show(testDevice, zigbee)
	m.field = zigbeeFieldLeave
	m.pendingLeave = true

	m, _ = m.Update(tea.KeyPressMsg{Code: 'k'})

	if m.pendingLeave {
		t.Error("navigation should clear pendingLeave")
	}
}

func TestZigbeeEditModel_SaveNoChange(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()
	zigbee := &shelly.TUIZigbeeStatus{Enabled: true, NetworkState: zigbeeStateReady}
	m, _ = m.Show(testDevice, zigbee)

	// Press Enter with no changes
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if m.visible {
		t.Error("should close on save with no changes")
	}
}

func TestZigbeeEditModel_SaveToggle(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()
	zigbee := &shelly.TUIZigbeeStatus{Enabled: true, NetworkState: zigbeeStateReady}
	m, _ = m.Show(testDevice, zigbee)

	// Toggle enable
	m, _ = m.Update(tea.KeyPressMsg{Code: 't'})

	if m.pendingEnabled {
		t.Error("pendingEnabled should be false after toggle")
	}

	// Save
	m, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if !m.saving {
		t.Error("should be saving")
	}
	if cmd == nil {
		t.Error("should return save command")
	}
}

func TestZigbeeEditModel_SaveResult_Success(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()
	zigbee := &shelly.TUIZigbeeStatus{Enabled: true, NetworkState: zigbeeStateReady}
	m, _ = m.Show(testDevice, zigbee)
	m.saving = true

	m, cmd := m.Update(messages.NewSaveResult(nil))

	if m.saving {
		t.Error("should not be saving after success")
	}
	if m.visible {
		t.Error("should close modal after successful save")
	}
	if cmd == nil {
		t.Error("should return close command")
	}
}

func TestZigbeeEditModel_SaveResult_Error(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()
	zigbee := &shelly.TUIZigbeeStatus{Enabled: true, NetworkState: zigbeeStateReady}
	m, _ = m.Show(testDevice, zigbee)
	m.saving = true

	m, _ = m.Update(messages.NewSaveError(nil, errTest))

	if m.saving {
		t.Error("should not be saving after error")
	}
	if m.err == nil {
		t.Error("error should be set")
	}
	if !m.visible {
		t.Error("should still be visible after error")
	}
}

func TestZigbeeEditModel_StartSteering(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()
	zigbee := &shelly.TUIZigbeeStatus{
		Enabled:      true,
		NetworkState: zigbeeStateReady,
	}
	m, _ = m.Show(testDevice, zigbee)

	// Navigate to pair button
	m, _ = m.Update(tea.KeyPressMsg{Code: 'j'})
	if m.field != zigbeeFieldPair {
		t.Errorf("field = %d, want %d", m.field, zigbeeFieldPair)
	}

	// Press Enter to start steering
	m, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if !m.steering {
		t.Error("should be steering")
	}
	if cmd == nil {
		t.Error("should return steering command")
	}
}

func TestZigbeeEditModel_SteeringResult_Success(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()
	zigbee := &shelly.TUIZigbeeStatus{Enabled: true, NetworkState: zigbeeStateReady}
	m, _ = m.Show(testDevice, zigbee)
	m.steering = true

	m, cmd := m.Update(ZigbeeSteeringResultMsg{Err: nil})

	if m.steering {
		t.Error("should not be steering after success")
	}
	if m.visible {
		t.Error("should close modal after steering success")
	}
	if cmd == nil {
		t.Error("should return close command")
	}
}

func TestZigbeeEditModel_SteeringResult_Error(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()
	zigbee := &shelly.TUIZigbeeStatus{Enabled: true, NetworkState: zigbeeStateReady}
	m, _ = m.Show(testDevice, zigbee)
	m.steering = true

	m, _ = m.Update(ZigbeeSteeringResultMsg{Err: errTest})

	if m.steering {
		t.Error("should not be steering after error")
	}
	if m.err == nil {
		t.Error("error should be set")
	}
	if !m.visible {
		t.Error("should still be visible after error")
	}
}

func TestZigbeeEditModel_LeaveConfirmation(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()
	zigbee := &shelly.TUIZigbeeStatus{
		Enabled:      true,
		NetworkState: zigbeeStateJoined,
		Channel:      15,
	}
	m, _ = m.Show(testDevice, zigbee)

	// Navigate to leave button
	m.field = zigbeeFieldLeave

	// First press - request confirmation
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if !m.pendingLeave {
		t.Error("should set pendingLeave on first press")
	}
	if m.leaving {
		t.Error("should not be leaving yet")
	}

	// Second press - confirm
	m, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if m.pendingLeave {
		t.Error("pendingLeave should be false after confirmation")
	}
	if !m.leaving {
		t.Error("should be leaving after confirmation")
	}
	if cmd == nil {
		t.Error("should return leave command")
	}
}

func TestZigbeeEditModel_LeaveResult_Success(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()
	zigbee := &shelly.TUIZigbeeStatus{
		Enabled:      true,
		NetworkState: zigbeeStateJoined,
	}
	m, _ = m.Show(testDevice, zigbee)
	m.leaving = true

	m, cmd := m.Update(ZigbeeLeaveResultMsg{Err: nil})

	if m.leaving {
		t.Error("should not be leaving after success")
	}
	if m.visible {
		t.Error("should close modal after leave success")
	}
	if cmd == nil {
		t.Error("should return close command")
	}
}

func TestZigbeeEditModel_LeaveResult_Error(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()
	zigbee := &shelly.TUIZigbeeStatus{
		Enabled:      true,
		NetworkState: zigbeeStateJoined,
	}
	m, _ = m.Show(testDevice, zigbee)
	m.leaving = true

	m, _ = m.Update(ZigbeeLeaveResultMsg{Err: errTest})

	if m.leaving {
		t.Error("should not be leaving after error")
	}
	if m.err == nil {
		t.Error("error should be set")
	}
	if !m.visible {
		t.Error("should still be visible after error")
	}
}

func TestZigbeeEditModel_EscClose(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()
	zigbee := &shelly.TUIZigbeeStatus{Enabled: true, NetworkState: zigbeeStateReady}
	m, _ = m.Show(testDevice, zigbee)

	m, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if m.visible {
		t.Error("should not be visible after Esc")
	}
	if cmd == nil {
		t.Error("should return close command")
	}
}

func TestZigbeeEditModel_CtrlSave(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()
	zigbee := &shelly.TUIZigbeeStatus{Enabled: true, NetworkState: zigbeeStateReady}
	m, _ = m.Show(testDevice, zigbee)

	// Toggle and save with Ctrl+S
	m, _ = m.Update(tea.KeyPressMsg{Code: 't'})
	m, cmd := m.Update(tea.KeyPressMsg{Mod: tea.ModCtrl, Code: 's'})

	if !m.saving {
		t.Error("should be saving after Ctrl+S")
	}
	if cmd == nil {
		t.Error("should return save command")
	}
}

func TestZigbeeEditModel_UpdateNotVisible(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()

	// Should be no-op when not visible
	m, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if m.visible {
		t.Error("should not become visible")
	}
	if cmd != nil {
		t.Error("should return nil command when not visible")
	}
}

func TestZigbeeEditModel_NoOpWhileBusy(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()
	zigbee := &shelly.TUIZigbeeStatus{Enabled: true, NetworkState: zigbeeStateReady}
	m, _ = m.Show(testDevice, zigbee)
	m.saving = true

	// Toggle should be ignored while saving
	original := m.pendingEnabled
	m, _ = m.Update(tea.KeyPressMsg{Code: 't'})

	if m.pendingEnabled != original {
		t.Error("toggle should be ignored while saving")
	}
}

func TestZigbeeEditModel_ToggleEnableRequestMsg(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()
	zigbee := &shelly.TUIZigbeeStatus{Enabled: true, NetworkState: zigbeeStateReady}
	m, _ = m.Show(testDevice, zigbee)

	m, _ = m.Update(messages.ToggleEnableRequestMsg{})

	if m.pendingEnabled {
		t.Error("pendingEnabled should be toggled by ToggleEnableRequestMsg")
	}
}

func TestZigbeeEditModel_View(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()
	m = m.SetSize(80, 40)
	zigbee := &shelly.TUIZigbeeStatus{
		Enabled:      true,
		NetworkState: zigbeeStateJoined,
		Channel:      15,
		PANID:        0xABCD,
	}
	m, _ = m.Show(testDevice, zigbee)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestZigbeeEditModel_View_NotVisible(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()

	view := m.View()

	if view != "" {
		t.Error("View() should return empty string when not visible")
	}
}

func TestZigbeeEditModel_View_WithError(t *testing.T) {
	t.Parallel()
	m := newTestZigbeeEditModel()
	m = m.SetSize(80, 40)
	zigbee := &shelly.TUIZigbeeStatus{Enabled: true, NetworkState: zigbeeStateReady}
	m, _ = m.Show(testDevice, zigbee)
	m.err = errTest

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string with error")
	}
}

func TestZigbeeEditModel_Footer(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		setup    func(m *ZigbeeEditModel)
		wantNon  string // footer should NOT be empty
		contains string // optional substring check
	}{
		{
			name:    "default",
			setup:   func(_ *ZigbeeEditModel) {},
			wantNon: "non-empty",
		},
		{
			name:     "saving",
			setup:    func(m *ZigbeeEditModel) { m.saving = true },
			contains: "Saving",
		},
		{
			name:     "steering",
			setup:    func(m *ZigbeeEditModel) { m.steering = true },
			contains: "steering",
		},
		{
			name:     "leaving",
			setup:    func(m *ZigbeeEditModel) { m.leaving = true },
			contains: "Leaving",
		},
		{
			name:     "pending leave",
			setup:    func(m *ZigbeeEditModel) { m.pendingLeave = true },
			contains: "confirm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := newTestZigbeeEditModel()
			zigbee := &shelly.TUIZigbeeStatus{Enabled: true, NetworkState: zigbeeStateReady}
			m, _ = m.Show(testDevice, zigbee)
			tt.setup(&m)

			footer := m.buildFooter()

			if footer == "" {
				t.Error("footer should not be empty")
			}
		})
	}
}

func TestZigbeeEditModel_CalcFieldCount(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		setup func(m *ZigbeeEditModel)
		want  int
	}{
		{
			name:  "disabled",
			setup: func(m *ZigbeeEditModel) { m.enabled = false },
			want:  1,
		},
		{
			name: "enabled not joined",
			setup: func(m *ZigbeeEditModel) {
				m.enabled = true
				m.networkState = zigbeeStateReady
			},
			want: 2,
		},
		{
			name: "enabled and joined",
			setup: func(m *ZigbeeEditModel) {
				m.enabled = true
				m.networkState = zigbeeStateJoined
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := newTestZigbeeEditModel()
			tt.setup(&m)

			count := m.calcFieldCount()
			if count != tt.want {
				t.Errorf("calcFieldCount() = %d, want %d", count, tt.want)
			}
		})
	}
}

func newTestZigbeeEditModel() ZigbeeEditModel {
	return NewZigbeeEditModel(context.Background(), &shelly.Service{})
}
