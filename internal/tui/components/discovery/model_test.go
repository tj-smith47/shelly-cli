package discovery

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
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
	if m.scanning {
		t.Error("should not be scanning initially")
	}
	if m.method != shelly.DiscoveryMDNS {
		t.Errorf("method = %v, want DiscoveryMDNS", m.method)
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

func TestModel_SetMethod(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	updated := m.SetMethod(shelly.DiscoveryHTTP)

	if updated.method != shelly.DiscoveryHTTP {
		t.Errorf("method = %v, want DiscoveryHTTP", updated.method)
	}
}

func TestModel_StartScan(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	updated, cmd := m.StartScan()

	if !updated.scanning {
		t.Error("should be scanning after StartScan")
	}
	if cmd == nil {
		t.Error("StartScan should return a command")
	}
}

func TestModel_StartScan_AlreadyScanning(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.scanning = true

	updated, cmd := m.StartScan()

	if cmd != nil {
		t.Error("should not return command when already scanning")
	}
	if !updated.scanning {
		t.Error("should still be scanning")
	}
}

func TestModel_Update_ScanComplete(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.scanning = true
	devices := []shelly.DiscoveredDevice{
		{ID: "device1", Address: "192.168.1.100", Model: "SHSW-1"},
		{ID: "device2", Address: "192.168.1.101", Model: "SHPLG-S"},
	}
	msg := ScanCompleteMsg{Devices: devices}

	updated, _ := m.Update(msg)

	if updated.scanning {
		t.Error("should not be scanning after ScanCompleteMsg")
	}
	if len(updated.devices) != 2 {
		t.Errorf("devices len = %d, want 2", len(updated.devices))
	}
	if updated.devices[0].Address != "192.168.1.100" {
		t.Errorf("devices[0].Address = %q, want 192.168.1.100", updated.devices[0].Address)
	}
}

func TestModel_Update_ScanCompleteError(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.scanning = true
	testErr := errors.New("scan failed")
	msg := ScanCompleteMsg{Err: testErr}

	updated, _ := m.Update(msg)

	if updated.scanning {
		t.Error("should not be scanning after error")
	}
	if updated.err == nil {
		t.Error("err should be set")
	}
	if !errors.Is(updated.err, testErr) {
		t.Errorf("err = %v, want %v", updated.err, testErr)
	}
}

func TestModel_Update_DeviceAdded(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.devices = []shelly.DiscoveredDevice{
		{ID: "device1", Address: "192.168.1.100", Added: false},
		{ID: "device2", Address: "192.168.1.101", Added: false},
	}
	msg := DeviceAddedMsg{Address: "192.168.1.100"}

	updated, _ := m.Update(msg)

	if !updated.devices[0].Added {
		t.Error("device should be marked as added")
	}
	if updated.devices[1].Added {
		t.Error("other device should not be marked as added")
	}
}

func TestModel_Update_DeviceAddedError(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	testErr := errors.New("add failed")
	msg := DeviceAddedMsg{Address: "192.168.1.100", Err: testErr}

	updated, _ := m.Update(msg)

	if updated.err == nil {
		t.Error("err should be set")
	}
}

func TestModel_HandleKey_Navigation(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.devices = []shelly.DiscoveredDevice{
		{ID: "device0"},
		{ID: "device1"},
		{ID: "device2"},
	}
	m.scroller.SetItemCount(len(m.devices))

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

func TestModel_HandleKey_MethodSwitch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		key  string
		want shelly.DiscoveryMethod
	}{
		{"m", shelly.DiscoveryMDNS},
		{"h", shelly.DiscoveryHTTP},
		{"c", shelly.DiscoveryCoIoT},
		{"b", shelly.DiscoveryBLE},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()
			m := newTestModel()
			m.focused = true

			msg := tea.KeyPressMsg{Code: rune(tt.key[0])}

			updated, _ := m.Update(msg)

			if updated.method != tt.want {
				t.Errorf("method = %v, want %v", updated.method, tt.want)
			}
		})
	}
}

func TestModel_HandleKey_Scan(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 's'})

	if !updated.scanning {
		t.Error("should be scanning after 's' key")
	}
	if cmd == nil {
		t.Error("should return scan command")
	}
}

func TestModel_HandleKey_NotFocused(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = false
	m.devices = []shelly.DiscoveredDevice{{ID: "device0"}}
	m.scroller.SetItemCount(len(m.devices))

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'j'})

	if updated.Cursor() != 0 {
		t.Error("cursor should not change when not focused")
	}
}

func TestModel_ScrollerCursorBounds(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.devices = []shelly.DiscoveredDevice{
		{ID: "device0"},
		{ID: "device1"},
	}
	m.scroller.SetItemCount(len(m.devices))

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
	m.devices = make([]shelly.DiscoveredDevice, 20)
	m.scroller.SetItemCount(20)

	// SetSize configures visible rows (height - 8 overhead)
	m = m.SetSize(80, 20)
	if m.scroller.VisibleRows() != 12 {
		t.Errorf("visibleRows = %d, want 12", m.scroller.VisibleRows())
	}

	m = m.SetSize(80, 5)
	if m.scroller.VisibleRows() < 1 {
		t.Errorf("visibleRows with small height = %d, want >= 1", m.scroller.VisibleRows())
	}
}

func TestModel_View_NoDevices(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_Scanning(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.scanning = true
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_Error(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.err = errors.New("scan failed")
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_WithDevices(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.devices = []shelly.DiscoveredDevice{
		{ID: "device1", Address: "192.168.1.100", Model: "SHSW-1", Name: "Kitchen", Generation: 2, Added: true},
		{ID: "device2", Address: "192.168.1.101", Model: "SHPLG-S", Name: "", Generation: 1, Added: false},
	}
	m = m.SetSize(80, 30)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_Accessors(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.devices = []shelly.DiscoveredDevice{
		{ID: "device0"},
		{ID: "device1"},
		{ID: "device2"},
	}
	m.scroller.SetItemCount(len(m.devices))
	m.scanning = true
	m.method = shelly.DiscoveryHTTP
	m.err = errors.New("test error")
	m.scroller.SetCursor(2)

	if len(m.Devices()) != 3 {
		t.Errorf("Devices() len = %d, want 3", len(m.Devices()))
	}
	if !m.Scanning() {
		t.Error("Scanning() should be true")
	}
	if m.Method() != shelly.DiscoveryHTTP {
		t.Errorf("Method() = %v, want DiscoveryHTTP", m.Method())
	}
	if m.Error() == nil {
		t.Error("Error() should not be nil")
	}
	if m.Cursor() != 2 {
		t.Errorf("Cursor() = %d, want 2", m.Cursor())
	}
}

func TestModel_Refresh(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	updated, cmd := m.Refresh()

	if !updated.scanning {
		t.Error("should be scanning after Refresh")
	}
	if cmd == nil {
		t.Error("Refresh should return a command")
	}
}

func TestModel_ScrollerEnsureVisible(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.devices = make([]shelly.DiscoveredDevice, 20)
	for i := range m.devices {
		m.devices[i] = shelly.DiscoveredDevice{ID: string(rune('a' + i))}
	}
	m.scroller.SetItemCount(20)
	m = m.SetSize(80, 15) // Sets visibleRows = 15 - 8 = 7

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

func TestDefaultStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultStyles()

	// Verify styles are created without panic
	_ = styles.Added.Render("test")
	_ = styles.NotAdded.Render("test")
	_ = styles.Model.Render("test")
	_ = styles.Address.Render("test")
	_ = styles.Generation.Render("test")
	_ = styles.Selected.Render("test")
	_ = styles.Label.Render("test")
	_ = styles.Error.Render("test")
	_ = styles.Muted.Render("test")
	_ = styles.ScanButton.Render("test")
}

func newTestModel() Model {
	ctx := context.Background()
	svc := &shelly.Service{}
	deps := Deps{Ctx: ctx, Svc: svc}
	return New(deps)
}
