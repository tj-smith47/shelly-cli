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

const testDevice = "192.168.1.100"

func TestNew(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc := &shelly.Service{}
	deps := Deps{Ctx: ctx, Svc: svc}

	m := New(deps)

	if m.ctx != ctx {
		t.Error("ctx not set")
	}
	if m.svc != svc {
		t.Error("svc not set")
	}
	if m.loading {
		t.Error("should not be loading initially")
	}
	if m.activeProtocol != ProtocolMatter {
		t.Error("should start with Matter protocol selected")
	}
}

func TestNew_PanicOnNilCtx(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil ctx")
		}
	}()

	deps := Deps{Ctx: nil, Svc: &shelly.Service{}}
	New(deps)
}

func TestNew_PanicOnNilSvc(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil svc")
		}
	}()

	deps := Deps{Ctx: context.Background(), Svc: nil}
	New(deps)
}

func TestDeps_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		deps    Deps
		wantErr bool
	}{
		{
			name:    "valid",
			deps:    Deps{Ctx: context.Background(), Svc: &shelly.Service{}},
			wantErr: false,
		},
		{
			name:    "nil ctx",
			deps:    Deps{Ctx: nil, Svc: &shelly.Service{}},
			wantErr: true,
		},
		{
			name:    "nil svc",
			deps:    Deps{Ctx: context.Background(), Svc: nil},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.deps.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestModel_Init(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	cmd := m.Init()

	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestModel_SetDevice(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	updated, cmd := m.SetDevice(testDevice)

	if updated.device != testDevice {
		t.Errorf("device = %q, want %q", updated.device, testDevice)
	}
	if cmd == nil {
		t.Error("SetDevice should return a command")
	}
	if !updated.loading {
		t.Error("should be loading after SetDevice")
	}
}

func TestModel_SetDevice_Empty(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice

	updated, cmd := m.SetDevice("")

	if updated.device != "" {
		t.Errorf("device = %q, want empty", updated.device)
	}
	if cmd != nil {
		t.Error("SetDevice with empty should return nil")
	}
}

func TestModel_SetDevice_ClearsState(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.matter = &shelly.TUIMatterStatus{Enabled: true}
	m.zigbee = &shelly.TUIZigbeeStatus{Enabled: true}
	m.lora = &shelly.TUILoRaStatus{Enabled: true}
	m.activeProtocol = ProtocolLoRa
	m.err = errors.New("previous error")

	updated, _ := m.SetDevice(testDevice)

	if updated.matter != nil {
		t.Error("matter should be nil after SetDevice")
	}
	if updated.zigbee != nil {
		t.Error("zigbee should be nil after SetDevice")
	}
	if updated.lora != nil {
		t.Error("lora should be nil after SetDevice")
	}
	if updated.activeProtocol != ProtocolMatter {
		t.Error("activeProtocol should reset to Matter")
	}
	if updated.err != nil {
		t.Error("err should be nil after SetDevice")
	}
}

func TestModel_SetSize(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	updated := m.SetSize(100, 50)

	if updated.Width != 100 {
		t.Errorf("width = %d, want 100", updated.Width)
	}
	if updated.Height != 50 {
		t.Errorf("height = %d, want 50", updated.Height)
	}
}

func TestModel_SetFocused(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	if m.focused {
		t.Error("should not be focused initially")
	}

	updated := m.SetFocused(true)

	if !updated.focused {
		t.Error("should be focused after SetFocused(true)")
	}

	updated = updated.SetFocused(false)

	if updated.focused {
		t.Error("should not be focused after SetFocused(false)")
	}
}

func TestModel_Update_StatusLoaded(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.loading = true
	msg := StatusLoadedMsg{
		Matter: &shelly.TUIMatterStatus{
			Enabled:        true,
			Commissionable: false,
			FabricsCount:   2,
		},
		Zigbee: &shelly.TUIZigbeeStatus{
			Enabled:      true,
			NetworkState: zigbeeStateJoined,
			Channel:      15,
			PANID:        0x1234,
		},
		LoRa: &shelly.TUILoRaStatus{
			Enabled:   true,
			Frequency: 868000000,
			TxPower:   14,
		},
	}

	updated, _ := m.Update(msg)

	if updated.loading {
		t.Error("should not be loading after StatusLoadedMsg")
	}
	if updated.matter == nil {
		t.Error("matter should be set")
	}
	if !updated.matter.Enabled {
		t.Error("matter.Enabled should be true")
	}
	if updated.zigbee == nil {
		t.Error("zigbee should be set")
	}
	if updated.zigbee.NetworkState != zigbeeStateJoined {
		t.Errorf("zigbee.NetworkState = %q, want %s", updated.zigbee.NetworkState, zigbeeStateJoined)
	}
	if updated.lora == nil {
		t.Error("lora should be set")
	}
	if updated.lora.Frequency != 868000000 {
		t.Errorf("lora.Frequency = %d, want 868000000", updated.lora.Frequency)
	}
}

func TestModel_Update_StatusLoadedError(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.loading = true
	testErr := errors.New("no protocols available")
	msg := StatusLoadedMsg{Err: testErr}

	updated, _ := m.Update(msg)

	if updated.loading {
		t.Error("should not be loading after error")
	}
	if updated.err == nil {
		t.Error("err should be set")
	}
	if !errors.Is(updated.err, testErr) {
		t.Errorf("err = %v, want %v", updated.err, testErr)
	}
}

func TestModel_Update_StatusLoadedPartial(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.loading = true
	msg := StatusLoadedMsg{
		Matter: &shelly.TUIMatterStatus{Enabled: true},
		// Zigbee and LoRa are nil (not supported)
	}

	updated, _ := m.Update(msg)

	if updated.loading {
		t.Error("should not be loading")
	}
	if updated.err != nil {
		t.Errorf("err should be nil, got %v", updated.err)
	}
	if updated.matter == nil {
		t.Error("matter should be set")
	}
	if updated.zigbee != nil {
		t.Error("zigbee should be nil")
	}
	if updated.lora != nil {
		t.Error("lora should be nil")
	}
}

func TestModel_HandleAction_Refresh(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice

	updated, cmd := m.Update(messages.RefreshRequestMsg{})

	if !updated.loading {
		t.Error("should be loading after RefreshRequestMsg")
	}
	if cmd == nil {
		t.Error("should return refresh command")
	}
}

func TestModel_HandleAction_RefreshWhileLoading(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.loading = true

	_, cmd := m.Update(messages.RefreshRequestMsg{})

	if cmd != nil {
		t.Error("should not return command when already loading")
	}
}

func TestModel_HandleAction_Navigate(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.activeProtocol = ProtocolMatter

	// Navigate down
	updated, _ := m.Update(messages.NavigationMsg{Direction: messages.NavDown})
	if updated.activeProtocol != ProtocolZigbee {
		t.Errorf("after NavDown, activeProtocol = %d, want %d", updated.activeProtocol, ProtocolZigbee)
	}

	// Navigate down again
	updated, _ = updated.Update(messages.NavigationMsg{Direction: messages.NavDown})
	if updated.activeProtocol != ProtocolLoRa {
		t.Errorf("after second NavDown, activeProtocol = %d, want %d", updated.activeProtocol, ProtocolLoRa)
	}

	// Navigate down to Z-Wave
	updated, _ = updated.Update(messages.NavigationMsg{Direction: messages.NavDown})
	if updated.activeProtocol != ProtocolZWave {
		t.Errorf("after third NavDown, activeProtocol = %d, want %d", updated.activeProtocol, ProtocolZWave)
	}

	// Navigate down to Modbus
	updated, _ = updated.Update(messages.NavigationMsg{Direction: messages.NavDown})
	if updated.activeProtocol != ProtocolModbus {
		t.Errorf("after fourth NavDown, activeProtocol = %d, want %d", updated.activeProtocol, ProtocolModbus)
	}

	// Navigate down wraps to Matter
	updated, _ = updated.Update(messages.NavigationMsg{Direction: messages.NavDown})
	if updated.activeProtocol != ProtocolMatter {
		t.Errorf("after wrap, activeProtocol = %d, want %d", updated.activeProtocol, ProtocolMatter)
	}

	// Navigate up wraps to Modbus
	updated, _ = updated.Update(messages.NavigationMsg{Direction: messages.NavUp})
	if updated.activeProtocol != ProtocolModbus {
		t.Errorf("after NavUp from Matter, activeProtocol = %d, want %d", updated.activeProtocol, ProtocolModbus)
	}

	// Navigate up to Z-Wave
	updated, _ = updated.Update(messages.NavigationMsg{Direction: messages.NavUp})
	if updated.activeProtocol != ProtocolZWave {
		t.Errorf("after NavUp, activeProtocol = %d, want %d", updated.activeProtocol, ProtocolZWave)
	}

	// Navigate up to LoRa
	updated, _ = updated.Update(messages.NavigationMsg{Direction: messages.NavUp})
	if updated.activeProtocol != ProtocolLoRa {
		t.Errorf("after NavUp, activeProtocol = %d, want %d", updated.activeProtocol, ProtocolLoRa)
	}

	// Navigate up to Zigbee
	updated, _ = updated.Update(messages.NavigationMsg{Direction: messages.NavUp})
	if updated.activeProtocol != ProtocolZigbee {
		t.Errorf("after NavUp, activeProtocol = %d, want %d", updated.activeProtocol, ProtocolZigbee)
	}
}

func TestModel_HandleKey_NumberSelect(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true

	// Select Matter (1)
	updated, _ := m.Update(tea.KeyPressMsg{Code: '1'})
	if updated.activeProtocol != ProtocolMatter {
		t.Errorf("after 1, activeProtocol = %d, want %d", updated.activeProtocol, ProtocolMatter)
	}

	// Select Zigbee (2)
	updated, _ = m.Update(tea.KeyPressMsg{Code: '2'})
	if updated.activeProtocol != ProtocolZigbee {
		t.Errorf("after 2, activeProtocol = %d, want %d", updated.activeProtocol, ProtocolZigbee)
	}

	// Select LoRa (3)
	updated, _ = m.Update(tea.KeyPressMsg{Code: '3'})
	if updated.activeProtocol != ProtocolLoRa {
		t.Errorf("after 3, activeProtocol = %d, want %d", updated.activeProtocol, ProtocolLoRa)
	}
}

func TestModel_HandleAction_NotFocused(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = false
	m.device = testDevice

	updated, _ := m.Update(messages.RefreshRequestMsg{})

	if updated.loading {
		t.Error("should not respond to actions when not focused")
	}
}

func TestModel_View_NoDevice(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_Loading(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.loading = true
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_Error(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.err = errors.New("no protocols available")
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_MatterEnabled(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.matter = &shelly.TUIMatterStatus{
		Enabled:        true,
		Commissionable: true,
		FabricsCount:   0,
	}
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_MatterDisabled(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.matter = &shelly.TUIMatterStatus{
		Enabled: false,
	}
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_MatterPaired(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.matter = &shelly.TUIMatterStatus{
		Enabled:        true,
		Commissionable: false,
		FabricsCount:   3,
	}
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_ZigbeeJoined(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.zigbee = &shelly.TUIZigbeeStatus{
		Enabled:      true,
		NetworkState: zigbeeStateJoined,
		Channel:      15,
		PANID:        0xABCD,
	}
	m.activeProtocol = ProtocolZigbee
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_ZigbeeSteering(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.zigbee = &shelly.TUIZigbeeStatus{
		Enabled:      true,
		NetworkState: zigbeeStateSteering,
	}
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_ZigbeeReady(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.zigbee = &shelly.TUIZigbeeStatus{
		Enabled:      true,
		NetworkState: zigbeeStateReady,
	}
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_ZigbeeDisabled(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.zigbee = &shelly.TUIZigbeeStatus{
		Enabled: false,
	}
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_LoRaEnabled(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.lora = &shelly.TUILoRaStatus{
		Enabled:   true,
		Frequency: 868000000,
		TxPower:   14,
		RSSI:      -70,
		SNR:       8.5,
	}
	m.activeProtocol = ProtocolLoRa
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_LoRaNoRSSI(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.lora = &shelly.TUILoRaStatus{
		Enabled:   true,
		Frequency: 915000000,
		TxPower:   20,
		RSSI:      0, // No signal yet
	}
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_LoRaDisabled(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.lora = &shelly.TUILoRaStatus{
		Enabled: false,
	}
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_NilProtocols(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	// All protocols nil = not supported
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_AllProtocols(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.matter = &shelly.TUIMatterStatus{
		Enabled:        true,
		Commissionable: false,
		FabricsCount:   1,
	}
	m.zigbee = &shelly.TUIZigbeeStatus{
		Enabled:      true,
		NetworkState: zigbeeStateJoined,
		Channel:      20,
		PANID:        0x5678,
	}
	m.lora = &shelly.TUILoRaStatus{
		Enabled:   true,
		Frequency: 433000000,
		TxPower:   10,
	}
	m = m.SetSize(80, 50)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_Accessors(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.matter = &shelly.TUIMatterStatus{Enabled: true}
	m.zigbee = &shelly.TUIZigbeeStatus{Enabled: true}
	m.lora = &shelly.TUILoRaStatus{Enabled: true}
	m.zwave = &shelly.TUIZWaveStatus{DeviceModel: "SNSW-102ZW", IsPro: false, SupportsLR: true}
	m.loading = true
	m.activeProtocol = ProtocolZigbee
	m.err = errors.New("test error")

	if m.Device() != testDevice {
		t.Errorf("Device() = %q, want %q", m.Device(), testDevice)
	}
	if m.Matter() == nil || !m.Matter().Enabled {
		t.Error("Matter() incorrect")
	}
	if m.Zigbee() == nil || !m.Zigbee().Enabled {
		t.Error("Zigbee() incorrect")
	}
	if m.LoRa() == nil || !m.LoRa().Enabled {
		t.Error("LoRa() incorrect")
	}
	if m.ZWave() == nil || m.ZWave().DeviceModel != "SNSW-102ZW" {
		t.Error("ZWave() incorrect")
	}
	if !m.Loading() {
		t.Error("Loading() should be true")
	}
	if m.ActiveProtocol() != ProtocolZigbee {
		t.Errorf("ActiveProtocol() = %d, want %d", m.ActiveProtocol(), ProtocolZigbee)
	}
	if m.Error() == nil {
		t.Error("Error() should not be nil")
	}
}

func TestModel_Refresh(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice

	updated, cmd := m.Refresh()

	if !updated.loading {
		t.Error("should be loading after Refresh")
	}
	if cmd == nil {
		t.Error("Refresh should return a command")
	}
}

func TestModel_Refresh_NoDevice(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	updated, cmd := m.Refresh()

	if updated.loading {
		t.Error("should not be loading without device")
	}
	if cmd != nil {
		t.Error("Refresh without device should return nil")
	}
}

func TestDefaultStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultStyles()

	// Verify styles are created without panic
	_ = styles.Enabled.Render("test")
	_ = styles.Disabled.Render("test")
	_ = styles.Label.Render("test")
	_ = styles.Value.Render("test")
	_ = styles.Highlight.Render("test")
	_ = styles.Error.Render("test")
	_ = styles.Muted.Render("test")
	_ = styles.Section.Render("test")
	_ = styles.Active.Render("test")
	_ = styles.Warning.Render("test")
}

func TestNextProtocol(t *testing.T) {
	t.Parallel()
	tests := []struct {
		from Protocol
		want Protocol
	}{
		{ProtocolMatter, ProtocolZigbee},
		{ProtocolZigbee, ProtocolLoRa},
		{ProtocolLoRa, ProtocolZWave},
		{ProtocolZWave, ProtocolModbus},
		{ProtocolModbus, ProtocolMatter},
	}

	for _, tt := range tests {
		m := newTestModel()
		m.activeProtocol = tt.from
		updated := m.nextProtocol()
		if updated.activeProtocol != tt.want {
			t.Errorf("nextProtocol() from %d = %d, want %d", tt.from, updated.activeProtocol, tt.want)
		}
	}
}

func TestPrevProtocol(t *testing.T) {
	t.Parallel()
	tests := []struct {
		from Protocol
		want Protocol
	}{
		{ProtocolMatter, ProtocolModbus},
		{ProtocolZigbee, ProtocolMatter},
		{ProtocolLoRa, ProtocolZigbee},
		{ProtocolZWave, ProtocolLoRa},
		{ProtocolModbus, ProtocolZWave},
	}

	for _, tt := range tests {
		m := newTestModel()
		m.activeProtocol = tt.from
		updated := m.prevProtocol()
		if updated.activeProtocol != tt.want {
			t.Errorf("prevProtocol() from %d = %d, want %d", tt.from, updated.activeProtocol, tt.want)
		}
	}
}

func TestModel_HandleAction_EditRequest_Matter(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolMatter
	m.matter = &shelly.TUIMatterStatus{Enabled: true, Commissionable: false, FabricsCount: 1}

	updated, cmd := m.Update(messages.EditRequestMsg{})

	if !updated.editing {
		t.Error("should be editing after EditRequestMsg with Matter selected")
	}
	if cmd == nil {
		t.Error("should return command for modal open")
	}
}

func TestModel_HandleAction_EditRequest_NoMatter(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolMatter
	m.matter = nil // Matter not supported

	updated, _ := m.Update(messages.EditRequestMsg{})

	if updated.editing {
		t.Error("should not open edit modal when matter is nil")
	}
}

func TestModel_HandleAction_EditRequest_NotFocused(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = false
	m.device = testDevice
	m.matter = &shelly.TUIMatterStatus{Enabled: true}

	updated, _ := m.Update(messages.EditRequestMsg{})

	if updated.editing {
		t.Error("should not open edit modal when not focused")
	}
}

func TestModel_HandleAction_EditRequest_WrongProtocol(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolZigbee
	m.matter = &shelly.TUIMatterStatus{Enabled: true}

	updated, _ := m.Update(messages.EditRequestMsg{})

	if updated.editing {
		t.Error("should not open edit modal when Zigbee is selected")
	}
}

func TestModel_HandleKey_MatterToggle(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolMatter
	m.matter = &shelly.TUIMatterStatus{Enabled: true}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 't'})

	if !updated.toggling {
		t.Error("should be toggling after 't' key")
	}
	if cmd == nil {
		t.Error("should return toggle command")
	}
}

func TestModel_HandleKey_MatterToggle_NotMatterProtocol(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolZigbee
	m.matter = &shelly.TUIMatterStatus{Enabled: true}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 't'})

	if updated.toggling {
		t.Error("should not toggle when Zigbee is selected")
	}
}

func TestModel_HandleKey_MatterToggle_NoDevice(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.activeProtocol = ProtocolMatter
	m.matter = &shelly.TUIMatterStatus{Enabled: true}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 't'})

	if updated.toggling {
		t.Error("should not toggle without device")
	}
	if cmd != nil {
		t.Error("should not return command without device")
	}
}

func TestModel_HandleKey_MatterCode(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolMatter
	m.matter = &shelly.TUIMatterStatus{Enabled: true, Commissionable: true}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'c'})

	if !updated.editing {
		t.Error("should open edit modal for 'c' key")
	}
	if cmd == nil {
		t.Error("should return command for modal open")
	}
}

func TestModel_HandleKey_MatterReset(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolMatter
	m.matter = &shelly.TUIMatterStatus{Enabled: true}

	// First press - should set pendingReset
	updated, _ := m.Update(tea.KeyPressMsg{Code: 'R'})

	if !updated.pendingReset {
		t.Error("should set pendingReset on first 'R' press")
	}
	if updated.toggling {
		t.Error("should not be toggling yet")
	}

	// Second press - should execute reset
	updated, cmd := updated.Update(tea.KeyPressMsg{Code: 'R'})

	if updated.pendingReset {
		t.Error("pendingReset should be false after confirmation")
	}
	if !updated.toggling {
		t.Error("should be toggling (busy) after confirmation")
	}
	if cmd == nil {
		t.Error("should return reset command")
	}
}

func TestModel_HandleKey_MatterReset_Disabled(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolMatter
	m.matter = &shelly.TUIMatterStatus{Enabled: false}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'R'})

	if updated.pendingReset {
		t.Error("should not allow reset when Matter is disabled")
	}
}

func TestModel_HandleKey_MatterReset_Cancel(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolMatter
	m.matter = &shelly.TUIMatterStatus{Enabled: true}
	m.pendingReset = true

	// Press Esc to cancel
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if updated.pendingReset {
		t.Error("pendingReset should be canceled on Esc")
	}
}

func TestModel_HandleKey_NumberClearsPendingReset(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolMatter
	m.pendingReset = true

	// Press 2 to switch to Zigbee
	updated, _ := m.Update(tea.KeyPressMsg{Code: '2'})

	if updated.pendingReset {
		t.Error("pendingReset should be cleared on protocol switch")
	}
	if updated.activeProtocol != ProtocolZigbee {
		t.Error("should switch to Zigbee")
	}
}

func TestModel_MatterToggleResult_Success(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.toggling = true

	updated, cmd := m.Update(MatterToggleResultMsg{Enabled: false, Err: nil})

	if updated.toggling {
		t.Error("toggling should be false after success")
	}
	if !updated.loading {
		t.Error("should be loading (refreshing) after toggle success")
	}
	if cmd == nil {
		t.Error("should return refresh command")
	}
}

func TestModel_MatterToggleResult_Error(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.toggling = true

	toggleErr := errors.New("toggle failed")
	updated, _ := m.Update(MatterToggleResultMsg{Err: toggleErr})

	if updated.toggling {
		t.Error("toggling should be false after error")
	}
	if updated.err == nil {
		t.Error("err should be set")
	}
}

func TestModel_MatterResetResult_Success(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.toggling = true
	m.pendingReset = true

	updated, cmd := m.Update(MatterResetResultMsg{Err: nil})

	if updated.toggling {
		t.Error("toggling should be false after success")
	}
	if updated.pendingReset {
		t.Error("pendingReset should be false after success")
	}
	if !updated.loading {
		t.Error("should be loading (refreshing) after reset success")
	}
	if cmd == nil {
		t.Error("should return refresh command")
	}
}

func TestModel_MatterResetResult_Error(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.toggling = true

	resetErr := errors.New("reset failed")
	updated, _ := m.Update(MatterResetResultMsg{Err: resetErr})

	if updated.toggling {
		t.Error("toggling should be false after error")
	}
	if updated.err == nil {
		t.Error("err should be set")
	}
}

func TestModel_IsEditing(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	if m.IsEditing() {
		t.Error("should not be editing initially")
	}

	m.editing = true

	if !m.IsEditing() {
		t.Error("should be editing when editing=true")
	}
}

func TestModel_RenderEditModal_NotEditing(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	view := m.RenderEditModal()

	if view != "" {
		t.Error("RenderEditModal should return empty string when not editing")
	}
}

func TestModel_SetEditModalSize(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.editing = true

	updated := m.SetEditModalSize(100, 50)

	if updated.editModal.Width != 100 {
		t.Errorf("editModal Width = %d, want 100", updated.editModal.Width)
	}
	if updated.editModal.Height != 50 {
		t.Errorf("editModal Height = %d, want 50", updated.editModal.Height)
	}
}

func TestModel_SetEditModalSize_NotEditing(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	updated := m.SetEditModalSize(100, 50)

	// Should not set size when not editing
	if updated.editModal.Width == 100 {
		t.Error("editModal Width should not be set when not editing")
	}
}

func TestModel_SetDevice_ClearsEditing(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.editing = true
	m.toggling = true
	m.pendingReset = true

	updated, _ := m.SetDevice(testDevice)

	if updated.editing {
		t.Error("editing should be cleared on SetDevice")
	}
	if updated.toggling {
		t.Error("toggling should be cleared on SetDevice")
	}
	if updated.pendingReset {
		t.Error("pendingReset should be cleared on SetDevice")
	}
}

func TestModel_BuildFooter_Matter(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.activeProtocol = ProtocolMatter
	m.matter = &shelly.TUIMatterStatus{Enabled: true}

	footer := m.buildFooter()

	if footer == "" {
		t.Error("footer should not be empty")
	}
}

func TestModel_BuildFooter_MatterDisabled(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.activeProtocol = ProtocolMatter
	m.matter = &shelly.TUIMatterStatus{Enabled: false}

	footer := m.buildFooter()

	if footer == "" {
		t.Error("footer should not be empty")
	}
}

func TestModel_BuildFooter_PendingReset(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.pendingReset = true

	footer := m.buildFooter()

	if footer == "" {
		t.Error("footer should not be empty")
	}
}

func TestModel_BuildFooter_Toggling(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.toggling = true

	footer := m.buildFooter()

	if footer == "" {
		t.Error("footer should not be empty")
	}
}

func TestModel_BuildFooter_NonMatterProtocol(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.activeProtocol = ProtocolZigbee

	footer := m.buildFooter()

	if footer == "" {
		t.Error("footer should not be empty")
	}
}

func TestModel_HandleAction_ViewRequest_OpensMatterEdit(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolMatter
	m.matter = &shelly.TUIMatterStatus{Enabled: true, Commissionable: true}

	updated, cmd := m.Update(messages.ViewRequestMsg{})

	if !updated.editing {
		t.Error("ViewRequestMsg should open edit modal for Matter")
	}
	if cmd == nil {
		t.Error("should return command for modal open")
	}
}

func TestModel_HandleAction_ResetRequest(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolMatter
	m.matter = &shelly.TUIMatterStatus{Enabled: true}

	updated, _ := m.Update(messages.ResetRequestMsg{})

	if !updated.pendingReset {
		t.Error("should set pendingReset on ResetRequestMsg")
	}
}

func TestModel_EditModal_ClosedRefreshes(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.editing = true
	m.matter = &shelly.TUIMatterStatus{Enabled: true}

	// Show the modal
	m.editModal, _ = m.editModal.Show(m.device, m.matter)

	// Simulate Esc to close modal
	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if updated.editing {
		t.Error("editing should be false after modal close")
	}
	if !updated.loading {
		t.Error("should be loading (refreshing) after modal close")
	}
	if cmd == nil {
		t.Error("should return refresh commands")
	}
}

// --- Zigbee-specific model tests ---

func TestModel_HandleAction_EditRequest_Zigbee(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolZigbee
	m.zigbee = &shelly.TUIZigbeeStatus{Enabled: true, NetworkState: zigbeeStateJoined}

	updated, cmd := m.Update(messages.EditRequestMsg{})

	if !updated.zigbeeEditing {
		t.Error("should open Zigbee edit modal when Zigbee is selected")
	}
	if cmd == nil {
		t.Error("should return command for modal open")
	}
}

func TestModel_HandleAction_EditRequest_ZigbeeNoData(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolZigbee
	m.zigbee = nil

	updated, _ := m.Update(messages.EditRequestMsg{})

	if updated.zigbeeEditing {
		t.Error("should not open edit modal when zigbee data is nil")
	}
}

func TestModel_HandleKey_ZigbeeToggle(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolZigbee
	m.zigbee = &shelly.TUIZigbeeStatus{Enabled: true}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 't'})

	if !updated.toggling {
		t.Error("should be toggling after 't' key with Zigbee")
	}
	if cmd == nil {
		t.Error("should return toggle command")
	}
}

func TestModel_HandleKey_ZigbeeToggle_NotZigbeeProtocol(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolLoRa
	m.zigbee = &shelly.TUIZigbeeStatus{Enabled: true}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 't'})

	if updated.toggling {
		t.Error("should not toggle Zigbee when LoRa is selected")
	}
}

func TestModel_HandleKey_ZigbeeToggle_NoDevice(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.activeProtocol = ProtocolZigbee
	m.zigbee = &shelly.TUIZigbeeStatus{Enabled: true}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 't'})

	if updated.toggling {
		t.Error("should not toggle without device")
	}
	if cmd != nil {
		t.Error("should not return command without device")
	}
}

func TestModel_HandleKey_ZigbeePair(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolZigbee
	m.zigbee = &shelly.TUIZigbeeStatus{Enabled: true, NetworkState: zigbeeStateReady}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'p'})

	if !updated.toggling {
		t.Error("should be toggling (busy) after 'p' key for pair")
	}
	if cmd == nil {
		t.Error("should return steering command")
	}
}

func TestModel_HandleKey_ZigbeePair_NotEnabled(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolZigbee
	m.zigbee = &shelly.TUIZigbeeStatus{Enabled: false}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'p'})

	if updated.toggling {
		t.Error("should not pair when Zigbee is disabled")
	}
	if cmd != nil {
		t.Error("should not return command when disabled")
	}
}

func TestModel_HandleKey_ZigbeePair_NotZigbeeProtocol(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolMatter

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'p'})

	if updated.toggling {
		t.Error("should not pair when Matter is selected")
	}
}

func TestModel_HandleKey_ZigbeeLeave(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolZigbee
	m.zigbee = &shelly.TUIZigbeeStatus{Enabled: true, NetworkState: zigbeeStateJoined}

	// First press - should set pendingLeave
	updated, _ := m.Update(tea.KeyPressMsg{Code: 'R'})

	if !updated.pendingLeave {
		t.Error("should set pendingLeave on first 'R' press")
	}
	if updated.toggling {
		t.Error("should not be toggling yet")
	}

	// Second press - should execute leave
	updated, cmd := updated.Update(tea.KeyPressMsg{Code: 'R'})

	if updated.pendingLeave {
		t.Error("pendingLeave should be false after confirmation")
	}
	if !updated.toggling {
		t.Error("should be toggling (busy) after confirmation")
	}
	if cmd == nil {
		t.Error("should return leave command")
	}
}

func TestModel_HandleKey_ZigbeeLeave_NotJoined(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolZigbee
	m.zigbee = &shelly.TUIZigbeeStatus{Enabled: true, NetworkState: zigbeeStateReady}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'R'})

	if updated.pendingLeave {
		t.Error("should not allow leave when not joined")
	}
}

func TestModel_HandleKey_ZigbeeLeave_Cancel(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolZigbee
	m.zigbee = &shelly.TUIZigbeeStatus{Enabled: true, NetworkState: zigbeeStateJoined}
	m.pendingLeave = true

	// Press Esc to cancel
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if updated.pendingLeave {
		t.Error("pendingLeave should be canceled on Esc")
	}
}

func TestModel_HandleKey_NumberClearsPendingLeave(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolZigbee
	m.pendingLeave = true

	// Press 1 to switch to Matter
	updated, _ := m.Update(tea.KeyPressMsg{Code: '1'})

	if updated.pendingLeave {
		t.Error("pendingLeave should be cleared on protocol switch")
	}
	if updated.activeProtocol != ProtocolMatter {
		t.Error("should switch to Matter")
	}
}

func TestModel_ZigbeeToggleResult_Success(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.toggling = true

	updated, cmd := m.Update(ZigbeeToggleResultMsg{Enabled: false, Err: nil})

	if updated.toggling {
		t.Error("toggling should be false after success")
	}
	if !updated.loading {
		t.Error("should be loading (refreshing) after toggle success")
	}
	if cmd == nil {
		t.Error("should return refresh command")
	}
}

func TestModel_ZigbeeToggleResult_Error(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.toggling = true

	toggleErr := errors.New("toggle failed")
	updated, _ := m.Update(ZigbeeToggleResultMsg{Err: toggleErr})

	if updated.toggling {
		t.Error("toggling should be false after error")
	}
	if updated.err == nil {
		t.Error("err should be set")
	}
}

func TestModel_ZigbeeSteeringResult_Success(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.toggling = true

	updated, cmd := m.Update(ZigbeeSteeringResultMsg{Err: nil})

	if updated.toggling {
		t.Error("toggling should be false after success")
	}
	if !updated.loading {
		t.Error("should be loading (refreshing) after steering success")
	}
	if cmd == nil {
		t.Error("should return refresh command")
	}
}

func TestModel_ZigbeeSteeringResult_Error(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.toggling = true

	steerErr := errors.New("steering failed")
	updated, _ := m.Update(ZigbeeSteeringResultMsg{Err: steerErr})

	if updated.toggling {
		t.Error("toggling should be false after error")
	}
	if updated.err == nil {
		t.Error("err should be set")
	}
}

func TestModel_ZigbeeLeaveResult_Success(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.toggling = true
	m.pendingLeave = true

	updated, cmd := m.Update(ZigbeeLeaveResultMsg{Err: nil})

	if updated.toggling {
		t.Error("toggling should be false after success")
	}
	if updated.pendingLeave {
		t.Error("pendingLeave should be false after success")
	}
	if !updated.loading {
		t.Error("should be loading (refreshing) after leave success")
	}
	if cmd == nil {
		t.Error("should return refresh command")
	}
}

func TestModel_ZigbeeLeaveResult_Error(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.toggling = true

	leaveErr := errors.New("leave failed")
	updated, _ := m.Update(ZigbeeLeaveResultMsg{Err: leaveErr})

	if updated.toggling {
		t.Error("toggling should be false after error")
	}
	if updated.err == nil {
		t.Error("err should be set")
	}
}

func TestModel_BuildFooter_Zigbee(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.activeProtocol = ProtocolZigbee
	m.zigbee = &shelly.TUIZigbeeStatus{Enabled: true, NetworkState: zigbeeStateReady}

	footer := m.buildFooter()

	if footer == "" {
		t.Error("footer should not be empty")
	}
}

func TestModel_BuildFooter_ZigbeeDisabled(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.activeProtocol = ProtocolZigbee
	m.zigbee = &shelly.TUIZigbeeStatus{Enabled: false}

	footer := m.buildFooter()

	if footer == "" {
		t.Error("footer should not be empty")
	}
}

func TestModel_BuildFooter_ZigbeeJoined(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.activeProtocol = ProtocolZigbee
	m.zigbee = &shelly.TUIZigbeeStatus{Enabled: true, NetworkState: zigbeeStateJoined}

	footer := m.buildFooter()

	if footer == "" {
		t.Error("footer should not be empty")
	}
}

func TestModel_BuildFooter_PendingLeave(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.pendingLeave = true

	footer := m.buildFooter()

	if footer == "" {
		t.Error("footer should not be empty")
	}
}

func TestModel_View_ZigbeeJoinedExtended(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.zigbee = &shelly.TUIZigbeeStatus{
		Enabled:          true,
		NetworkState:     zigbeeStateJoined,
		Channel:          15,
		PANID:            0xABCD,
		EUI64:            "00:11:22:33:44:55:66:77",
		CoordinatorEUI64: "AA:BB:CC:DD:EE:FF:00:11",
	}
	m.activeProtocol = ProtocolZigbee
	m = m.SetSize(80, 50)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_SetDevice_ClearsZigbeeEditing(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.zigbeeEditing = true
	m.pendingLeave = true

	updated, _ := m.SetDevice(testDevice)

	if updated.zigbeeEditing {
		t.Error("zigbeeEditing should be cleared on SetDevice")
	}
	if updated.pendingLeave {
		t.Error("pendingLeave should be cleared on SetDevice")
	}
}

func TestModel_IsEditing_Zigbee(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	if m.IsEditing() {
		t.Error("should not be editing initially")
	}

	m.zigbeeEditing = true

	if !m.IsEditing() {
		t.Error("should be editing when zigbeeEditing=true")
	}

	m.zigbeeEditing = false
	m.editing = true

	if !m.IsEditing() {
		t.Error("should be editing when editing=true")
	}
}

func TestModel_RenderEditModal_Zigbee(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.zigbeeEditing = true
	m.zigbeeModal = m.zigbeeModal.SetSize(80, 40)
	m.zigbeeModal, _ = m.zigbeeModal.Show(testDevice, &shelly.TUIZigbeeStatus{Enabled: true, NetworkState: zigbeeStateJoined})

	view := m.RenderEditModal()

	if view == "" {
		t.Error("RenderEditModal should return Zigbee modal view")
	}
}

func TestModel_SetEditModalSize_Zigbee(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.zigbeeEditing = true

	updated := m.SetEditModalSize(100, 50)

	if updated.zigbeeModal.Width != 100 {
		t.Errorf("zigbeeModal Width = %d, want 100", updated.zigbeeModal.Width)
	}
	if updated.zigbeeModal.Height != 50 {
		t.Errorf("zigbeeModal Height = %d, want 50", updated.zigbeeModal.Height)
	}
}

func TestModel_EditModal_ZigbeeClosedRefreshes(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.zigbeeEditing = true
	m.zigbee = &shelly.TUIZigbeeStatus{Enabled: true, NetworkState: zigbeeStateJoined}

	// Show the zigbee modal
	m.zigbeeModal, _ = m.zigbeeModal.Show(m.device, m.zigbee)

	// Simulate Esc to close modal
	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if updated.zigbeeEditing {
		t.Error("zigbeeEditing should be false after modal close")
	}
	if !updated.loading {
		t.Error("should be loading (refreshing) after modal close")
	}
	if cmd == nil {
		t.Error("should return refresh commands")
	}
}

// --- Z-Wave model tests ---

func TestModel_HandleAction_EditRequest_ZWave(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolZWave
	m.zwave = &shelly.TUIZWaveStatus{DeviceModel: "SNSW-102ZW", DeviceName: "Switch", IsPro: false, SupportsLR: true}

	updated, cmd := m.Update(messages.EditRequestMsg{})

	if !updated.zwaveEditing {
		t.Error("should open Z-Wave edit modal when Z-Wave is selected")
	}
	if cmd == nil {
		t.Error("should return command for modal open")
	}
}

func TestModel_HandleAction_EditRequest_ZWaveNoData(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolZWave
	m.zwave = nil

	updated, _ := m.Update(messages.EditRequestMsg{})

	if updated.zwaveEditing {
		t.Error("should not open edit modal when zwave data is nil")
	}
}

func TestModel_HandleKey_ProtocolSelect_ZWave(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolMatter

	updated, _ := m.Update(tea.KeyPressMsg{Code: '4'})

	if updated.activeProtocol != ProtocolZWave {
		t.Errorf("activeProtocol = %d, want %d (ProtocolZWave)", updated.activeProtocol, ProtocolZWave)
	}
}

func TestModel_BuildFooter_ZWave(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.Width = 80
	m.focused = true
	m.activeProtocol = ProtocolZWave
	m.zwave = &shelly.TUIZWaveStatus{DeviceModel: "SNSW-102ZW"}

	footer := m.buildFooter()

	if !strings.Contains(footer, "e:edit") {
		t.Errorf("Z-Wave footer should contain 'e:edit', got %q", footer)
	}
	if !strings.Contains(footer, "r:refresh") {
		t.Errorf("Z-Wave footer should contain 'r:refresh', got %q", footer)
	}
}

func TestModel_View_ZWave(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.zwave = &shelly.TUIZWaveStatus{
		DeviceModel: "SNSW-102ZW",
		DeviceName:  "Switch",
		IsPro:       false,
		SupportsLR:  true,
	}
	m = m.SetSize(80, 50)

	view := m.View()

	if !strings.Contains(view, "Z-Wave") {
		t.Error("View should contain Z-Wave section header")
	}
	if !strings.Contains(view, "SNSW-102ZW") {
		t.Error("View should contain device model")
	}
}

func TestModel_View_ZWavePro(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.activeProtocol = ProtocolZWave
	m.zwave = &shelly.TUIZWaveStatus{
		DeviceModel: "SPSW-201ZW",
		DeviceName:  "Shutter",
		IsPro:       true,
		SupportsLR:  true,
	}
	m = m.SetSize(80, 50)

	view := m.View()

	if !strings.Contains(view, "Wave Pro") {
		t.Error("View should contain 'Wave Pro' for Pro devices")
	}
}

func TestModel_View_ZWaveNil(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.activeProtocol = ProtocolZWave
	m = m.SetSize(80, 50)

	view := m.View()

	if !strings.Contains(view, "Not a Wave device") {
		t.Error("View should show 'Not a Wave device' when zwave is nil")
	}
}

func TestModel_IsEditing_ZWave(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	if m.IsEditing() {
		t.Error("should not be editing initially")
	}

	m.zwaveEditing = true

	if !m.IsEditing() {
		t.Error("should be editing when zwaveEditing=true")
	}
}

func TestModel_RenderEditModal_ZWave(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.zwaveEditing = true
	m.zwaveModal = m.zwaveModal.SetSize(80, 40)
	m.zwaveModal, _ = m.zwaveModal.Show(testDevice, nil)

	view := m.RenderEditModal()

	if view == "" {
		t.Error("RenderEditModal should return Z-Wave modal view")
	}
}

func TestModel_SetEditModalSize_ZWave(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.zwaveEditing = true

	updated := m.SetEditModalSize(100, 50)

	if updated.zwaveModal.Width != 100 {
		t.Errorf("zwaveModal Width = %d, want 100", updated.zwaveModal.Width)
	}
	if updated.zwaveModal.Height != 50 {
		t.Errorf("zwaveModal Height = %d, want 50", updated.zwaveModal.Height)
	}
}

func TestModel_SetDevice_ClearsZWaveEditing(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.zwaveEditing = true

	updated, _ := m.SetDevice(testDevice)

	if updated.zwaveEditing {
		t.Error("zwaveEditing should be cleared on SetDevice")
	}
}

func TestModel_EditModal_ZWaveClosedNoRefresh(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.zwaveEditing = true
	m.zwave = &shelly.TUIZWaveStatus{DeviceModel: "SNSW-102ZW", DeviceName: "Switch"}

	// Show the Z-Wave modal
	m.zwaveModal, _ = m.zwaveModal.Show(m.device, m.zwave)

	// Simulate Esc to close modal
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if updated.zwaveEditing {
		t.Error("zwaveEditing should be false after modal close")
	}
	// Z-Wave modal is informational, so no refresh needed
	if updated.loading {
		t.Error("should not trigger loading/refresh for informational Z-Wave modal")
	}
}

func TestModel_ZWaveToggle_NoOp(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolZWave
	m.zwave = &shelly.TUIZWaveStatus{DeviceModel: "SNSW-102ZW"}

	// Toggle should be no-op for Z-Wave
	updated, cmd := m.Update(tea.KeyPressMsg{Code: 't'})

	if updated.toggling {
		t.Error("should not toggle for Z-Wave (no RPC support)")
	}
	if cmd != nil {
		t.Error("should not return command for Z-Wave toggle")
	}
}

func TestModel_ZWaveDestructive_NoOp(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolZWave
	m.zwave = &shelly.TUIZWaveStatus{DeviceModel: "SNSW-102ZW"}

	// 'R' should be no-op for Z-Wave
	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'R'})

	if updated.toggling {
		t.Error("should not toggle for Z-Wave destructive action")
	}
	if cmd != nil {
		t.Error("should not return command for Z-Wave destructive")
	}
}

// --- Modbus model tests ---

func TestModel_HandleAction_EditRequest_Modbus(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolModbus
	m.modbus = &shelly.TUIModbusStatus{Enabled: true}

	updated, cmd := m.Update(messages.EditRequestMsg{})

	if !updated.modbusEditing {
		t.Error("should open Modbus edit modal when Modbus is selected")
	}
	if cmd == nil {
		t.Error("should return command for modal open")
	}
}

func TestModel_HandleAction_EditRequest_ModbusNoData(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolModbus
	m.modbus = nil

	updated, _ := m.Update(messages.EditRequestMsg{})

	if updated.modbusEditing {
		t.Error("should not open edit modal when modbus data is nil")
	}
}

func TestModel_HandleKey_ModbusToggle(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolModbus
	m.modbus = &shelly.TUIModbusStatus{Enabled: true}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 't'})

	if !updated.toggling {
		t.Error("should be toggling after 't' key with Modbus")
	}
	if cmd == nil {
		t.Error("should return toggle command")
	}
}

func TestModel_HandleKey_ModbusToggle_NotModbusProtocol(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolMatter
	m.modbus = &shelly.TUIModbusStatus{Enabled: true}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 't'})

	// Matter toggle may fire, but modbus should not be affected
	if updated.activeProtocol != ProtocolMatter {
		t.Error("activeProtocol should remain Matter")
	}
}

func TestModel_HandleKey_ModbusToggle_NoDevice(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.activeProtocol = ProtocolModbus
	m.modbus = &shelly.TUIModbusStatus{Enabled: true}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 't'})

	if updated.toggling {
		t.Error("should not toggle without device")
	}
	if cmd != nil {
		t.Error("should not return command without device")
	}
}

func TestModel_HandleKey_ModbusToggle_NilModbus(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolModbus
	m.modbus = nil

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 't'})

	if updated.toggling {
		t.Error("should not toggle when modbus is nil")
	}
	if cmd != nil {
		t.Error("should not return command when modbus is nil")
	}
}

func TestModel_HandleKey_ProtocolSelect_Modbus(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolMatter

	updated, _ := m.Update(tea.KeyPressMsg{Code: '5'})

	if updated.activeProtocol != ProtocolModbus {
		t.Errorf("activeProtocol = %d, want %d (ProtocolModbus)", updated.activeProtocol, ProtocolModbus)
	}
}

func TestModel_ModbusToggleResult_Success(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.toggling = true

	updated, cmd := m.Update(ModbusToggleResultMsg{Enabled: false, Err: nil})

	if updated.toggling {
		t.Error("toggling should be false after success")
	}
	if !updated.loading {
		t.Error("should be loading (refreshing) after toggle success")
	}
	if cmd == nil {
		t.Error("should return refresh command")
	}
}

func TestModel_ModbusToggleResult_Error(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.toggling = true

	toggleErr := errors.New("toggle failed")
	updated, _ := m.Update(ModbusToggleResultMsg{Err: toggleErr})

	if updated.toggling {
		t.Error("toggling should be false after error")
	}
	if updated.err == nil {
		t.Error("err should be set")
	}
}

func TestModel_BuildFooter_Modbus(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.Width = 80
	m.focused = true
	m.activeProtocol = ProtocolModbus
	m.modbus = &shelly.TUIModbusStatus{Enabled: true}

	footer := m.buildFooter()

	if !strings.Contains(footer, "e:edit") {
		t.Errorf("Modbus footer should contain 'e:edit', got %q", footer)
	}
	if !strings.Contains(footer, "t:toggle") {
		t.Errorf("Modbus footer should contain 't:toggle', got %q", footer)
	}
	if !strings.Contains(footer, "r:refresh") {
		t.Errorf("Modbus footer should contain 'r:refresh', got %q", footer)
	}
}

func TestModel_BuildFooter_ModbusNil(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.activeProtocol = ProtocolModbus
	m.modbus = nil

	footer := m.buildFooter()

	// Should fall through to default footer
	if !strings.Contains(footer, "1-5:sel") {
		t.Errorf("Modbus nil footer should contain '1-5:sel', got %q", footer)
	}
}

func TestModel_View_ModbusEnabled(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.activeProtocol = ProtocolModbus
	m.modbus = &shelly.TUIModbusStatus{Enabled: true}
	m = m.SetSize(80, 50)

	view := m.View()

	if !strings.Contains(view, "Modbus") {
		t.Error("View should contain Modbus section header")
	}
	if !strings.Contains(view, "502") {
		t.Error("View should contain port 502 when enabled")
	}
}

func TestModel_View_ModbusDisabled(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.activeProtocol = ProtocolModbus
	m.modbus = &shelly.TUIModbusStatus{Enabled: false}
	m = m.SetSize(80, 50)

	view := m.View()

	if !strings.Contains(view, "Modbus") {
		t.Error("View should contain Modbus section header")
	}
	if !strings.Contains(view, "Disabled") {
		t.Error("View should contain 'Disabled' when modbus is off")
	}
}

func TestModel_View_ModbusNil(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.activeProtocol = ProtocolModbus
	m = m.SetSize(80, 50)

	view := m.View()

	if !strings.Contains(view, "Not supported") {
		t.Error("View should show 'Not supported' when modbus is nil")
	}
}

func TestModel_IsEditing_Modbus(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	if m.IsEditing() {
		t.Error("should not be editing initially")
	}

	m.modbusEditing = true

	if !m.IsEditing() {
		t.Error("should be editing when modbusEditing=true")
	}
}

func TestModel_RenderEditModal_Modbus(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.modbusEditing = true
	m.modbusModal = m.modbusModal.SetSize(80, 40)
	m.modbusModal, _ = m.modbusModal.Show("test", &shelly.TUIModbusStatus{})

	view := m.RenderEditModal()

	if view == "" {
		t.Error("RenderEditModal should return Modbus modal view")
	}
}

func TestModel_SetEditModalSize_Modbus(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.modbusEditing = true

	updated := m.SetEditModalSize(100, 50)

	if updated.modbusModal.Width != 100 {
		t.Errorf("modbusModal Width = %d, want 100", updated.modbusModal.Width)
	}
	if updated.modbusModal.Height != 50 {
		t.Errorf("modbusModal Height = %d, want 50", updated.modbusModal.Height)
	}
}

func TestModel_SetDevice_ClearsModbusEditing(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.modbusEditing = true
	m.modbus = &shelly.TUIModbusStatus{Enabled: true}

	updated, _ := m.SetDevice(testDevice)

	if updated.modbusEditing {
		t.Error("modbusEditing should be cleared on SetDevice")
	}
	if updated.modbus != nil {
		t.Error("modbus should be nil after SetDevice")
	}
}

func TestModel_EditModal_ModbusClosedRefreshes(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.modbusEditing = true
	m.modbus = &shelly.TUIModbusStatus{Enabled: true}

	// Show the modbus modal
	m.modbusModal, _ = m.modbusModal.Show(m.device, m.modbus)

	// Simulate Esc to close modal
	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if updated.modbusEditing {
		t.Error("modbusEditing should be false after modal close")
	}
	if !updated.loading {
		t.Error("should be loading (refreshing) after modal close")
	}
	if cmd == nil {
		t.Error("should return refresh commands")
	}
}

func TestModel_ModbusDestructive_NoOp(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolModbus
	m.modbus = &shelly.TUIModbusStatus{Enabled: true}

	// 'R' should be no-op for Modbus
	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'R'})

	if updated.toggling {
		t.Error("should not toggle for Modbus destructive action")
	}
	if cmd != nil {
		t.Error("should not return command for Modbus destructive")
	}
}

func TestModel_Accessors_Modbus(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.modbus = &shelly.TUIModbusStatus{Enabled: true}

	if m.Modbus() == nil {
		t.Error("Modbus() should not be nil")
	}
	if !m.Modbus().Enabled {
		t.Error("Modbus().Enabled should be true")
	}
}

func TestModel_Update_StatusLoaded_WithModbus(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.loading = true
	msg := StatusLoadedMsg{
		Modbus: &shelly.TUIModbusStatus{Enabled: true},
	}

	updated, _ := m.Update(msg)

	if updated.loading {
		t.Error("should not be loading after StatusLoadedMsg")
	}
	if updated.modbus == nil {
		t.Error("modbus should be set")
	}
	if !updated.modbus.Enabled {
		t.Error("modbus.Enabled should be true")
	}
}

// --- LoRa model tests ---

func TestModel_HandleAction_EditRequest_LoRa(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolLoRa
	m.lora = &shelly.TUILoRaStatus{Enabled: true, Frequency: 868000000, Bandwidth: 125, DataRate: 7}

	updated, cmd := m.Update(messages.EditRequestMsg{})

	if !updated.loraEditing {
		t.Error("should open LoRa edit modal when LoRa is selected")
	}
	if cmd == nil {
		t.Error("should return command for modal open")
	}
}

func TestModel_HandleAction_EditRequest_LoRaNoData(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolLoRa
	m.lora = nil

	updated, _ := m.Update(messages.EditRequestMsg{})

	if updated.loraEditing {
		t.Error("should not open edit modal when lora data is nil")
	}
}

func TestModel_HandleKey_LoRaTestSend(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolLoRa
	m.lora = &shelly.TUILoRaStatus{Enabled: true}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'T'})

	if !updated.toggling {
		t.Error("should be toggling (busy) after 'T' key with LoRa")
	}
	if cmd == nil {
		t.Error("should return test send command")
	}
}

func TestModel_HandleKey_LoRaTestSend_NotLoRaProtocol(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolMatter
	m.lora = &shelly.TUILoRaStatus{Enabled: true}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'T'})

	if updated.toggling {
		t.Error("should not toggle when not on LoRa protocol")
	}
	if cmd != nil {
		t.Error("should not return command when not on LoRa protocol")
	}
}

func TestModel_HandleKey_LoRaTestSend_NoDevice(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.activeProtocol = ProtocolLoRa
	m.lora = &shelly.TUILoRaStatus{Enabled: true}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'T'})

	if updated.toggling {
		t.Error("should not test send without device")
	}
	if cmd != nil {
		t.Error("should not return command without device")
	}
}

func TestModel_HandleKey_LoRaTestSend_NilLoRa(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolLoRa
	m.lora = nil

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'T'})

	if updated.toggling {
		t.Error("should not test send when lora is nil")
	}
	if cmd != nil {
		t.Error("should not return command when lora is nil")
	}
}

func TestModel_HandleKey_LoRaTestSend_DisabledLoRa(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolLoRa
	m.lora = &shelly.TUILoRaStatus{Enabled: false}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'T'})

	if updated.toggling {
		t.Error("should not test send when lora is disabled")
	}
	if cmd != nil {
		t.Error("should not return command when lora is disabled")
	}
}

func TestModel_LoRaTestSendResult_Success(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.toggling = true

	updated, _ := m.Update(LoRaTestSendResultMsg{Err: nil})

	if updated.toggling {
		t.Error("toggling should be false after success")
	}
	if updated.err != nil {
		t.Error("err should be nil after successful test send")
	}
}

func TestModel_LoRaTestSendResult_Error(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.toggling = true

	sendErr := errors.New("test send failed")
	updated, _ := m.Update(LoRaTestSendResultMsg{Err: sendErr})

	if updated.toggling {
		t.Error("toggling should be false after error")
	}
	if updated.err == nil {
		t.Error("err should be set")
	}
}

func TestModel_BuildFooter_LoRa(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.Width = 80
	m.focused = true
	m.activeProtocol = ProtocolLoRa
	m.lora = &shelly.TUILoRaStatus{Enabled: true}

	footer := m.buildFooter()

	if !strings.Contains(footer, "e:edit") {
		t.Errorf("LoRa footer should contain 'e:edit', got %q", footer)
	}
	if !strings.Contains(footer, "T:test") {
		t.Errorf("LoRa footer should contain 'T:test', got %q", footer)
	}
	if !strings.Contains(footer, "r:refresh") {
		t.Errorf("LoRa footer should contain 'r:refresh', got %q", footer)
	}
}

func TestModel_BuildFooter_LoRaNil(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.activeProtocol = ProtocolLoRa
	m.lora = nil

	footer := m.buildFooter()

	// Should fall through to default footer
	if !strings.Contains(footer, "1-5:sel") {
		t.Errorf("LoRa nil footer should contain '1-5:sel', got %q", footer)
	}
}

func TestModel_IsEditing_LoRa(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	if m.IsEditing() {
		t.Error("should not be editing initially")
	}

	m.loraEditing = true

	if !m.IsEditing() {
		t.Error("should be editing when loraEditing=true")
	}
}

func TestModel_RenderEditModal_LoRa(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.loraEditing = true
	m.loraModal = m.loraModal.SetSize(80, 40)
	m.loraModal, _ = m.loraModal.Show(testDevice, nil)

	view := m.RenderEditModal()

	if view == "" {
		t.Error("RenderEditModal should return LoRa modal view")
	}
}

func TestModel_SetEditModalSize_LoRa(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.loraEditing = true

	updated := m.SetEditModalSize(100, 50)

	if updated.loraModal.Width != 100 {
		t.Errorf("loraModal Width = %d, want 100", updated.loraModal.Width)
	}
	if updated.loraModal.Height != 50 {
		t.Errorf("loraModal Height = %d, want 50", updated.loraModal.Height)
	}
}

func TestModel_SetDevice_ClearsLoRaEditing(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.loraEditing = true
	m.lora = &shelly.TUILoRaStatus{Enabled: true}

	updated, _ := m.SetDevice(testDevice)

	if updated.loraEditing {
		t.Error("loraEditing should be cleared on SetDevice")
	}
	if updated.lora != nil {
		t.Error("lora should be nil after SetDevice")
	}
}

func TestModel_EditModal_LoRaClosedRefreshes(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.loraEditing = true
	m.lora = &shelly.TUILoRaStatus{Enabled: true, Frequency: 868000000, Bandwidth: 125, DataRate: 7}

	// Show the LoRa modal
	m.loraModal, _ = m.loraModal.Show(m.device, m.lora)

	// Simulate Esc to close modal
	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if updated.loraEditing {
		t.Error("loraEditing should be false after modal close")
	}
	if !updated.loading {
		t.Error("should be loading (refreshing) after modal close")
	}
	if cmd == nil {
		t.Error("should return refresh commands")
	}
}

func TestModel_LoRaToggle_NoOp(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolLoRa
	m.lora = &shelly.TUILoRaStatus{Enabled: true}

	// Lowercase 't' toggle should be no-op for LoRa
	updated, cmd := m.Update(tea.KeyPressMsg{Code: 't'})

	if updated.toggling {
		t.Error("should not toggle for LoRa (no toggle support)")
	}
	if cmd != nil {
		t.Error("should not return command for LoRa toggle")
	}
}

func TestModel_LoRaDestructive_NoOp(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.activeProtocol = ProtocolLoRa
	m.lora = &shelly.TUILoRaStatus{Enabled: true}

	// 'R' should be no-op for LoRa
	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'R'})

	if updated.toggling {
		t.Error("should not toggle for LoRa destructive action")
	}
	if cmd != nil {
		t.Error("should not return command for LoRa destructive")
	}
}

func newTestModel() Model {
	ctx := context.Background()
	svc := &shelly.Service{}
	deps := Deps{Ctx: ctx, Svc: svc}
	return New(deps)
}
