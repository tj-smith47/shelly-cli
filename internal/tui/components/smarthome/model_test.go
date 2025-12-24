package smarthome

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

const (
	testDevice          = "192.168.1.100"
	zigbeeStateJoined   = "joined"
	zigbeeStateSteering = "steering"
	zigbeeStateReady    = "ready"
)

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

	if updated.width != 100 {
		t.Errorf("width = %d, want 100", updated.width)
	}
	if updated.height != 50 {
		t.Errorf("height = %d, want 50", updated.height)
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

func TestModel_HandleKey_Refresh(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'r'})

	if !updated.loading {
		t.Error("should be loading after 'r' key")
	}
	if cmd == nil {
		t.Error("should return refresh command")
	}
}

func TestModel_HandleKey_RefreshWhileLoading(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.loading = true

	_, cmd := m.Update(tea.KeyPressMsg{Code: 'r'})

	if cmd != nil {
		t.Error("should not return command when already loading")
	}
}

func TestModel_HandleKey_Navigate(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.activeProtocol = ProtocolMatter

	// Navigate down
	updated, _ := m.Update(tea.KeyPressMsg{Code: 'j'})
	if updated.activeProtocol != ProtocolZigbee {
		t.Errorf("after j, activeProtocol = %d, want %d", updated.activeProtocol, ProtocolZigbee)
	}

	// Navigate down again
	updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if updated.activeProtocol != ProtocolLoRa {
		t.Errorf("after down, activeProtocol = %d, want %d", updated.activeProtocol, ProtocolLoRa)
	}

	// Navigate down wraps to Matter
	updated, _ = updated.Update(tea.KeyPressMsg{Code: 'j'})
	if updated.activeProtocol != ProtocolMatter {
		t.Errorf("after wrap, activeProtocol = %d, want %d", updated.activeProtocol, ProtocolMatter)
	}

	// Navigate up wraps to LoRa
	updated, _ = updated.Update(tea.KeyPressMsg{Code: 'k'})
	if updated.activeProtocol != ProtocolLoRa {
		t.Errorf("after k from Matter, activeProtocol = %d, want %d", updated.activeProtocol, ProtocolLoRa)
	}

	// Navigate up
	updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if updated.activeProtocol != ProtocolZigbee {
		t.Errorf("after up, activeProtocol = %d, want %d", updated.activeProtocol, ProtocolZigbee)
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

func TestModel_HandleKey_NotFocused(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = false
	m.device = testDevice

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'r'})

	if updated.loading {
		t.Error("should not respond to keys when not focused")
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
		{ProtocolLoRa, ProtocolMatter},
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
		{ProtocolMatter, ProtocolLoRa},
		{ProtocolZigbee, ProtocolMatter},
		{ProtocolLoRa, ProtocolZigbee},
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

func newTestModel() Model {
	ctx := context.Background()
	svc := &shelly.Service{}
	deps := Deps{Ctx: ctx, Svc: svc}
	return New(deps)
}
