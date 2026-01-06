package protocols

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
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
		MQTT: &MQTTData{
			Connected:   true,
			Enable:      true,
			Server:      "mqtt://broker.example.com",
			User:        "shelly",
			TopicPrefix: "shelly/kitchen",
		},
		Modbus: &ModbusData{
			Enabled: true,
			Enable:  true,
		},
		Ethernet: &EthernetData{
			IP:       "192.168.1.50",
			Enable:   true,
			IPv4Mode: "dhcp",
		},
	}

	updated, _ := m.Update(msg)

	if updated.loading {
		t.Error("should not be loading after StatusLoadedMsg")
	}
	if updated.mqtt == nil {
		t.Error("mqtt should be set")
	}
	if updated.mqtt.Server != "mqtt://broker.example.com" {
		t.Errorf("mqtt.Server = %q, want mqtt://broker.example.com", updated.mqtt.Server)
	}
	if updated.modbus == nil {
		t.Error("modbus should be set")
	}
	if !updated.modbus.Enabled {
		t.Error("modbus.Enabled should be true")
	}
	if updated.ethernet == nil {
		t.Error("ethernet should be set")
	}
	if updated.ethernet.IP != "192.168.1.50" {
		t.Errorf("ethernet.IP = %q, want 192.168.1.50", updated.ethernet.IP)
	}
}

func TestModel_Update_StatusLoadedError(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.loading = true
	testErr := errors.New("connection failed")
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

func TestModel_HandleKey_Navigation(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.mqtt = &MQTTData{}
	m.modbus = &ModbusData{}
	m.ethernet = &EthernetData{}

	// Start at MQTT
	if m.activeProtocol != ProtocolMQTT {
		t.Error("should start at MQTT")
	}

	// Move down to Modbus
	updated, _ := m.Update(tea.KeyPressMsg{Code: 'j'})
	if updated.activeProtocol != ProtocolModbus {
		t.Errorf("activeProtocol after j = %d, want Modbus", updated.activeProtocol)
	}

	// Move down to Ethernet
	updated, _ = updated.Update(tea.KeyPressMsg{Code: 'j'})
	if updated.activeProtocol != ProtocolEthernet {
		t.Errorf("activeProtocol after second j = %d, want Ethernet", updated.activeProtocol)
	}

	// Move down wraps to MQTT
	updated, _ = updated.Update(tea.KeyPressMsg{Code: 'j'})
	if updated.activeProtocol != ProtocolMQTT {
		t.Errorf("activeProtocol after third j = %d, want MQTT (wrap)", updated.activeProtocol)
	}

	// Move up wraps to Ethernet
	updated, _ = updated.Update(tea.KeyPressMsg{Code: 'k'})
	if updated.activeProtocol != ProtocolEthernet {
		t.Errorf("activeProtocol after k = %d, want Ethernet (wrap)", updated.activeProtocol)
	}
}

func TestModel_HandleKey_NumberSelection(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true

	// Press 2 for Modbus
	updated, _ := m.Update(tea.KeyPressMsg{Code: '2'})
	if updated.activeProtocol != ProtocolModbus {
		t.Errorf("activeProtocol after '2' = %d, want Modbus", updated.activeProtocol)
	}

	// Press 3 for Ethernet
	updated, _ = updated.Update(tea.KeyPressMsg{Code: '3'})
	if updated.activeProtocol != ProtocolEthernet {
		t.Errorf("activeProtocol after '3' = %d, want Ethernet", updated.activeProtocol)
	}

	// Press 1 for MQTT
	updated, _ = updated.Update(tea.KeyPressMsg{Code: '1'})
	if updated.activeProtocol != ProtocolMQTT {
		t.Errorf("activeProtocol after '1' = %d, want MQTT", updated.activeProtocol)
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

func TestModel_HandleKey_NotFocused(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = false
	m.device = testDevice

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'j'})

	if updated.activeProtocol != ProtocolMQTT {
		t.Error("activeProtocol should not change when not focused")
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
	m.err = errors.New("connection failed")
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_WithData(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.mqtt = &MQTTData{
		Connected:   true,
		Enable:      true,
		Server:      "mqtt://broker.example.com",
		User:        "shelly",
		TopicPrefix: "home/kitchen",
	}
	m.modbus = &ModbusData{
		Enabled: true,
		Enable:  true,
	}
	m.ethernet = &EthernetData{
		IP:       "192.168.1.50",
		Enable:   true,
		IPv4Mode: "dhcp",
	}
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_MQTTDisabled(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.mqtt = &MQTTData{
		Enable: false,
	}
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_EthernetStatic(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.ethernet = &EthernetData{
		IP:         "192.168.1.50",
		Enable:     true,
		IPv4Mode:   "static",
		StaticIP:   "192.168.1.50",
		Netmask:    "255.255.255.0",
		Gateway:    "192.168.1.1",
		Nameserver: "8.8.8.8",
	}
	m.activeProtocol = ProtocolEthernet
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_EthernetNoLink(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.ethernet = &EthernetData{
		Enable:   true,
		IPv4Mode: "dhcp",
		IP:       "", // No link
	}
	m.activeProtocol = ProtocolEthernet
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
	// All protocols nil = "Not supported"
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_Accessors(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.mqtt = &MQTTData{Server: "test"}
	m.modbus = &ModbusData{Enabled: true}
	m.ethernet = &EthernetData{IP: "1.2.3.4"}
	m.loading = true
	m.err = errors.New("test error")
	m.activeProtocol = ProtocolEthernet

	if m.Device() != testDevice {
		t.Errorf("Device() = %q, want %q", m.Device(), testDevice)
	}
	if m.MQTT() == nil || m.MQTT().Server != "test" {
		t.Error("MQTT() incorrect")
	}
	if m.Modbus() == nil || !m.Modbus().Enabled {
		t.Error("Modbus() incorrect")
	}
	if m.Ethernet() == nil || m.Ethernet().IP != "1.2.3.4" {
		t.Error("Ethernet() incorrect")
	}
	if !m.Loading() {
		t.Error("Loading() should be true")
	}
	if m.Error() == nil {
		t.Error("Error() should not be nil")
	}
	if m.ActiveProtocol() != ProtocolEthernet {
		t.Errorf("ActiveProtocol() = %d, want Ethernet", m.ActiveProtocol())
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
}

func TestProtocol_Constants(t *testing.T) {
	t.Parallel()

	// Verify protocol constants are distinct
	protocols := []Protocol{ProtocolMQTT, ProtocolModbus, ProtocolEthernet}
	seen := make(map[Protocol]bool)
	for _, p := range protocols {
		if seen[p] {
			t.Errorf("Protocol %d is duplicated", p)
		}
		seen[p] = true
	}
}

func newTestModel() Model {
	ctx := context.Background()
	svc := &shelly.Service{}
	deps := Deps{Ctx: ctx, Svc: svc}
	return New(deps)
}
