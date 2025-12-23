package kvs

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
		t.Error("loading should be false initially")
	}
	if m.focused {
		t.Error("focused should be false initially")
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
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})

	cmd := m.Init()

	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestModel_SetDevice(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})

	m, cmd := m.SetDevice(testDevice)

	if m.device != testDevice {
		t.Errorf("device = %q, want %q", m.device, testDevice)
	}
	if !m.loading {
		t.Error("loading should be true after SetDevice")
	}
	if cmd == nil {
		t.Error("SetDevice should return a command")
	}
}

func TestModel_SetDevice_Empty(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})
	m.device = "old-device"

	m, cmd := m.SetDevice("")

	if m.device != "" {
		t.Errorf("device = %q, want empty", m.device)
	}
	if m.loading {
		t.Error("loading should be false for empty device")
	}
	if cmd != nil {
		t.Error("SetDevice('') should return nil command")
	}
}

func TestModel_SetSize(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})

	m = m.SetSize(80, 24)

	if m.width != 80 {
		t.Errorf("width = %d, want 80", m.width)
	}
	if m.height != 24 {
		t.Errorf("height = %d, want 24", m.height)
	}
}

func TestModel_SetFocused(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})

	m = m.SetFocused(true)

	if !m.focused {
		t.Error("focused should be true")
	}

	m = m.SetFocused(false)

	if m.focused {
		t.Error("focused should be false")
	}
}

func TestModel_Update_LoadedMsg(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})
	m.loading = true

	items := []Item{
		{Key: "key1", Value: "value1", Etag: "etag1"},
		{Key: "key2", Value: 42.0, Etag: "etag2"},
	}
	msg := LoadedMsg{Items: items}

	m, cmd := m.Update(msg)

	if m.loading {
		t.Error("loading should be false after LoadedMsg")
	}
	if len(m.items) != 2 {
		t.Errorf("items count = %d, want 2", len(m.items))
	}
	if m.cursor != 0 {
		t.Error("cursor should be reset to 0")
	}
	if cmd != nil {
		t.Error("LoadedMsg should not return a command")
	}
}

func TestModel_Update_LoadedMsg_Error(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})
	m.loading = true

	testErr := errors.New("test error")
	msg := LoadedMsg{Err: testErr}

	m, _ = m.Update(msg)

	if m.loading {
		t.Error("loading should be false after error")
	}
	if !errors.Is(m.err, testErr) {
		t.Errorf("err = %v, want %v", m.err, testErr)
	}
}

func TestModel_Update_ActionMsg(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})
	m.device = testDevice
	m.items = []Item{{Key: "key1", Value: "value1"}}

	msg := ActionMsg{Action: "delete", Key: "key1"}

	m, cmd := m.Update(msg)

	if !m.loading {
		t.Error("loading should be true after successful action")
	}
	if cmd == nil {
		t.Error("successful action should trigger refresh")
	}
}

func TestModel_Update_ActionMsg_Error(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})

	testErr := errors.New("delete failed")
	msg := ActionMsg{Action: "delete", Key: "key1", Err: testErr}

	m, _ = m.Update(msg)

	if !errors.Is(m.err, testErr) {
		t.Errorf("err = %v, want %v", m.err, testErr)
	}
}

func TestModel_Update_KeyPressMsg_NotFocused(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})
	m.focused = false
	m.items = []Item{{Key: "key1"}, {Key: "key2"}}

	msg := tea.KeyPressMsg{Code: 106} // 'j'

	m, _ = m.Update(msg)

	if m.cursor != 0 {
		t.Error("cursor should not change when not focused")
	}
}

func TestModel_Update_KeyPressMsg_Navigation(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})
	m.focused = true
	m.items = []Item{{Key: "key1"}, {Key: "key2"}, {Key: "key3"}}
	m.height = 20

	// Test j/down
	msg := tea.KeyPressMsg{Code: 106} // 'j'
	m, _ = m.Update(msg)
	if m.cursor != 1 {
		t.Errorf("cursor after j = %d, want 1", m.cursor)
	}

	// Test k/up
	msg = tea.KeyPressMsg{Code: 107} // 'k'
	m, _ = m.Update(msg)
	if m.cursor != 0 {
		t.Errorf("cursor after k = %d, want 0", m.cursor)
	}

	// Test G (go to end)
	msg = tea.KeyPressMsg{Code: 71} // 'G'
	m, _ = m.Update(msg)
	if m.cursor != 2 {
		t.Errorf("cursor after G = %d, want 2", m.cursor)
	}

	// Test g (go to start)
	msg = tea.KeyPressMsg{Code: 103} // 'g'
	m, _ = m.Update(msg)
	if m.cursor != 0 {
		t.Errorf("cursor after g = %d, want 0", m.cursor)
	}
}

func TestModel_CursorBounds(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})
	m.items = []Item{{Key: "key1"}, {Key: "key2"}}
	m.height = 20

	// Cursor should not go below 0
	m.cursor = 0
	m = m.cursorUp()
	if m.cursor != 0 {
		t.Errorf("cursor should stay at 0, got %d", m.cursor)
	}

	// Cursor should not exceed items
	m.cursor = 1
	m = m.cursorDown()
	if m.cursor != 1 {
		t.Errorf("cursor should stay at 1, got %d", m.cursor)
	}
}

func TestModel_View_NoDevice(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})
	m = m.SetSize(40, 10)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_Loading(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})
	m.device = testDevice
	m.loading = true
	m = m.SetSize(40, 10)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_Error(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})
	m.device = testDevice
	m.err = errors.New("test error")
	m = m.SetSize(40, 10)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_Empty(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})
	m.device = testDevice
	m.items = []Item{}
	m = m.SetSize(40, 10)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_WithItems(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})
	m.device = testDevice
	m.items = []Item{
		{Key: "key1", Value: "string value", Etag: "etag1"},
		{Key: "key2", Value: 42.0, Etag: "etag2"},
		{Key: "key3", Value: true, Etag: "etag3"},
	}
	m = m.SetSize(60, 20)
	m = m.SetFocused(true)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_FormatValue(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})

	tests := []struct {
		name  string
		value any
	}{
		{"nil", nil},
		{"string", "test"},
		{"long string", "this is a very long string that should be truncated"},
		{"integer", 42.0},
		{"float", 3.14159},
		{"bool true", true},
		{"bool false", false},
		{"map", map[string]any{"nested": "value"}},
		{"array", []any{"a", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := m.formatValue(tt.value)
			if result == "" {
				t.Errorf("formatValue(%v) returned empty string", tt.value)
			}
		})
	}
}

func TestModel_SelectedItem(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})

	// No items
	if m.SelectedItem() != nil {
		t.Error("SelectedItem should return nil when no items")
	}

	// With items
	m.items = []Item{
		{Key: "key1", Value: "value1"},
		{Key: "key2", Value: "value2"},
	}
	m.cursor = 1

	item := m.SelectedItem()
	if item == nil {
		t.Fatal("SelectedItem should not be nil")
	}
	if item.Key != "key2" {
		t.Errorf("SelectedItem().Key = %q, want %q", item.Key, "key2")
	}
}

func TestModel_ItemCount(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})

	if m.ItemCount() != 0 {
		t.Errorf("ItemCount() = %d, want 0", m.ItemCount())
	}

	m.items = []Item{{Key: "key1"}, {Key: "key2"}}

	if m.ItemCount() != 2 {
		t.Errorf("ItemCount() = %d, want 2", m.ItemCount())
	}
}

func TestModel_Device(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})

	if m.Device() != "" {
		t.Errorf("Device() = %q, want empty", m.Device())
	}

	m.device = testDevice

	if m.Device() != testDevice {
		t.Errorf("Device() = %q, want %q", m.Device(), testDevice)
	}
}

func TestModel_Loading(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})

	if m.Loading() {
		t.Error("Loading() should be false initially")
	}

	m.loading = true

	if !m.Loading() {
		t.Error("Loading() should be true")
	}
}

func TestModel_Error(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})

	if m.Error() != nil {
		t.Error("Error() should be nil initially")
	}

	testErr := errors.New("test error")
	m.err = testErr

	if !errors.Is(m.Error(), testErr) {
		t.Errorf("Error() = %v, want %v", m.Error(), testErr)
	}
}

func TestModel_Refresh(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})

	// No device - should not refresh
	m, cmd := m.Refresh()
	if cmd != nil {
		t.Error("Refresh() should return nil when no device")
	}

	// With device
	m.device = testDevice
	m, cmd = m.Refresh()

	if !m.loading {
		t.Error("loading should be true after Refresh")
	}
	if cmd == nil {
		t.Error("Refresh() should return a command")
	}
}

func TestModel_VisibleRows(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})

	// Very small height
	m.height = 2
	if m.visibleRows() != 1 {
		t.Errorf("visibleRows() = %d, want 1 for small height", m.visibleRows())
	}

	// Normal height
	m.height = 20
	expected := 16 // 20 - 4 for borders/title
	if m.visibleRows() != expected {
		t.Errorf("visibleRows() = %d, want %d", m.visibleRows(), expected)
	}
}

func TestModel_EnsureVisible(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})
	m.height = 10 // 6 visible rows (10 - 4)
	m.items = make([]Item, 20)

	// Cursor above scroll
	m.cursor = 2
	m.scroll = 5
	m = m.ensureVisible()
	if m.scroll != 2 {
		t.Errorf("scroll = %d, want 2 (scroll to cursor)", m.scroll)
	}

	// Cursor below visible area
	m.cursor = 15
	m.scroll = 0
	m = m.ensureVisible()
	// cursor 15, visible 6, so scroll should be 15 - 6 + 1 = 10
	if m.scroll != 10 {
		t.Errorf("scroll = %d, want 10 (scroll up to show cursor)", m.scroll)
	}
}

func TestModel_SelectItem_Empty(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})
	m.focused = true

	cmd := m.selectItem()
	if cmd != nil {
		t.Error("selectItem() should return nil when no items")
	}
}

func TestModel_DeleteItem_Empty(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &shelly.Service{}})
	m.focused = true

	cmd := m.deleteItem()
	if cmd != nil {
		t.Error("deleteItem() should return nil when no items")
	}
}

func TestDefaultStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultStyles()

	// Just verify styles are created without panic
	_ = styles.Key.Render("test")
	_ = styles.Value.Render("test")
	_ = styles.String.Render("test")
	_ = styles.Number.Render("test")
	_ = styles.Bool.Render("test")
	_ = styles.Null.Render("test")
	_ = styles.Object.Render("test")
	_ = styles.Selected.Render("test")
	_ = styles.Error.Render("test")
	_ = styles.Muted.Render("test")
}
