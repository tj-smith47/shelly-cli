package wifi

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/network"
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
		Status: &network.WiFiStatusFull{
			Status: "got ip",
			StaIP:  "192.168.1.50",
			SSID:   "MyNetwork",
			RSSI:   -45,
		},
		Config: &network.WiFiConfigFull{
			STA: &network.WiFiStationFull{
				SSID:    "MyNetwork",
				Enabled: true,
			},
		},
	}

	updated, _ := m.Update(msg)

	if updated.loading {
		t.Error("should not be loading after StatusLoadedMsg")
	}
	if updated.status == nil {
		t.Error("status should be set")
	}
	if updated.status.SSID != "MyNetwork" {
		t.Errorf("status.SSID = %q, want MyNetwork", updated.status.SSID)
	}
	if updated.config == nil {
		t.Error("config should be set")
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

func TestModel_Update_ScanResult(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.scanning = true
	networks := []network.WiFiNetworkFull{
		{SSID: "Network1", RSSI: -50},
		{SSID: "Network2", RSSI: -70},
	}
	msg := ScanResultMsg{Networks: networks}

	updated, _ := m.Update(msg)

	if updated.scanning {
		t.Error("should not be scanning after ScanResultMsg")
	}
	if len(updated.networks) != 2 {
		t.Errorf("networks len = %d, want 2", len(updated.networks))
	}
}

func TestModel_Update_ScanResultError(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.scanning = true
	testErr := errors.New("scan failed")
	msg := ScanResultMsg{Err: testErr}

	updated, _ := m.Update(msg)

	if updated.scanning {
		t.Error("should not be scanning after error")
	}
	if updated.err == nil {
		t.Error("err should be set")
	}
}

func TestModel_HandleKey_Navigation(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.networks = []network.WiFiNetworkFull{
		{SSID: "Network1"},
		{SSID: "Network2"},
		{SSID: "Network3"},
	}
	m.scroller.SetItemCount(len(m.networks))

	// Move down
	updated, _ := m.Update(tea.KeyPressMsg{Code: 'j'})
	if updated.Cursor() != 1 {
		t.Errorf("cursor after j = %d, want 1", updated.Cursor())
	}

	// Move down again
	updated, _ = updated.Update(tea.KeyPressMsg{Code: 'j'})
	if updated.Cursor() != 2 {
		t.Errorf("cursor after second j = %d, want 2", updated.Cursor())
	}

	// Move up
	updated, _ = updated.Update(tea.KeyPressMsg{Code: 'k'})
	if updated.Cursor() != 1 {
		t.Errorf("cursor after k = %d, want 1", updated.Cursor())
	}
}

func TestModel_HandleKey_Scan(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 's'})

	if !updated.scanning {
		t.Error("should be scanning after 's' key")
	}
	if cmd == nil {
		t.Error("should return scan command")
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
	m.networks = []network.WiFiNetworkFull{{SSID: "Network1"}}
	m.scroller.SetItemCount(len(m.networks))

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'j'})

	if updated.Cursor() != 0 {
		t.Error("cursor should not change when not focused")
	}
}

func TestModel_ScrollerCursorBounds(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.networks = []network.WiFiNetworkFull{
		{SSID: "Network1"},
		{SSID: "Network2"},
	}
	m.scroller.SetItemCount(len(m.networks))

	// Can't go below 0
	m.scroller.CursorUp()
	if m.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0 (can't go below)", m.Cursor())
	}

	// Can't exceed list length
	m.scroller.SetCursor(1)
	m.scroller.CursorDown()
	if m.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1 (can't exceed list)", m.Cursor())
	}
}

func TestModel_ScrollerVisibleRows(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	// SetSize configures visible rows (height - 12 overhead)
	m = m.SetSize(80, 20)
	if m.scroller.VisibleRows() != 8 {
		t.Errorf("visibleRows = %d, want 8", m.scroller.VisibleRows())
	}

	m = m.SetSize(80, 5)
	if m.scroller.VisibleRows() < 1 {
		t.Errorf("visibleRows with small height = %d, want >= 1", m.scroller.VisibleRows())
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

func TestModel_View_WithStatus(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.status = &network.WiFiStatusFull{
		Status:        "got ip",
		StaIP:         "192.168.1.50",
		SSID:          "MyNetwork",
		RSSI:          -45,
		APClientCount: 2,
	}
	m.config = &network.WiFiConfigFull{
		STA: &network.WiFiStationFull{
			SSID:    "MyNetwork",
			Enabled: true,
		},
		AP: &network.WiFiAPFull{
			SSID:    "Shelly-AP",
			Enabled: false,
		},
	}
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_WithNetworks(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.status = &network.WiFiStatusFull{Status: "got ip"}
	m.networks = []network.WiFiNetworkFull{
		{SSID: "Network1", RSSI: -40, Auth: "wpa2"},
		{SSID: "Network2", RSSI: -55, Auth: "wpa2"},
		{SSID: "OpenNetwork", RSSI: -65, Auth: "open"},
		{SSID: "WeakNetwork", RSSI: -80, Auth: "wpa3"},
	}
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_Scanning(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.status = &network.WiFiStatusFull{Status: "got ip"}
	m.scanning = true
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_Accessors(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.status = &network.WiFiStatusFull{SSID: "Test"}
	m.config = &network.WiFiConfigFull{}
	m.networks = []network.WiFiNetworkFull{{SSID: "Net1"}}
	m.loading = true
	m.scanning = true
	m.err = errors.New("test error")

	if m.Device() != testDevice {
		t.Errorf("Device() = %q, want %q", m.Device(), testDevice)
	}
	if m.Status() == nil || m.Status().SSID != "Test" {
		t.Error("Status() incorrect")
	}
	if m.Config() == nil {
		t.Error("Config() should not be nil")
	}
	if len(m.Networks()) != 1 {
		t.Errorf("Networks() len = %d, want 1", len(m.Networks()))
	}
	if !m.Loading() {
		t.Error("Loading() should be true")
	}
	if !m.Scanning() {
		t.Error("Scanning() should be true")
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

func TestModel_RenderSignalStrength(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	tests := []struct {
		rssi float64
		want string
	}{
		{-40, "excellent"},
		{-55, "good"},
		{-65, "fair"},
		{-80, "weak"},
	}

	for _, tt := range tests {
		result := m.renderSignalStrength(tt.rssi)
		if result == "" {
			t.Errorf("renderSignalStrength(%v) returned empty", tt.rssi)
		}
	}
}

func TestModel_GetSignalIconAndStyle(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	tests := []struct {
		rssi     float64
		wantIcon string
	}{
		{-40, "████"},
		{-55, "███░"},
		{-65, "██░░"},
		{-80, "█░░░"},
	}

	for _, tt := range tests {
		icon, _ := m.getSignalIconAndStyle(tt.rssi)
		if icon != tt.wantIcon {
			t.Errorf("getSignalIconAndStyle(%v) icon = %q, want %q", tt.rssi, icon, tt.wantIcon)
		}
	}
}

func TestModel_RenderNetworkLine_LongSSID(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	netw := network.WiFiNetworkFull{
		SSID: "This Is A Very Long Network Name That Should Be Truncated",
		RSSI: -50,
		Auth: "wpa2",
	}

	line := m.renderNetworkLine(netw, false)

	if line == "" {
		t.Error("renderNetworkLine should not return empty")
	}
}

func TestModel_RenderNetworkLine_Selected(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	netw := network.WiFiNetworkFull{SSID: "TestNetwork", RSSI: -50, Auth: "wpa2"}

	line := m.renderNetworkLine(netw, true)

	if line == "" {
		t.Error("renderNetworkLine should not return empty for selected")
	}
}

func TestModel_RenderNetworkLine_OpenNetwork(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	netw := network.WiFiNetworkFull{SSID: "OpenWifi", RSSI: -60, Auth: "open"}

	line := m.renderNetworkLine(netw, false)

	if line == "" {
		t.Error("renderNetworkLine should not return empty for open network")
	}
}

func TestDefaultStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultStyles()

	// Verify styles are created without panic
	_ = styles.Connected.Render("test")
	_ = styles.Disconnected.Render("test")
	_ = styles.SSID.Render("test")
	_ = styles.Signal.Render("test")
	_ = styles.SignalWeak.Render("test")
	_ = styles.Label.Render("test")
	_ = styles.Value.Render("test")
	_ = styles.Selected.Render("test")
	_ = styles.Error.Render("test")
	_ = styles.Muted.Render("test")
}

func TestModel_ScrollerEnsureVisible(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.networks = make([]network.WiFiNetworkFull, 20)
	for i := range m.networks {
		m.networks[i] = network.WiFiNetworkFull{SSID: "Network"}
	}
	m.scroller.SetItemCount(20)
	m = m.SetSize(80, 20) // Sets visibleRows = 20 - 12 = 8

	// Cursor at end should scroll
	m.scroller.CursorToEnd()
	start, _ := m.scroller.VisibleRange()
	if start == 0 {
		t.Error("scroll should increase when cursor at end of long list")
	}

	// Cursor back to start
	m.scroller.CursorToStart()
	start, _ = m.scroller.VisibleRange()
	if start != 0 {
		t.Errorf("scroll = %d, want 0 when cursor at beginning", start)
	}
}

func newTestModel() Model {
	ctx := context.Background()
	svc := &shelly.Service{}
	deps := Deps{Ctx: ctx, Svc: svc}
	return New(deps)
}
