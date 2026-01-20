package kvs

import (
	"context"
	"errors"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/shelly/kvs"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
)

const testDevice = "192.168.1.100"

func TestNew(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc := &kvs.Service{}
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

	deps := Deps{Ctx: nil, Svc: &kvs.Service{}}
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
			deps:    Deps{Ctx: context.Background(), Svc: &kvs.Service{}},
			wantErr: false,
		},
		{
			name:    "nil ctx",
			deps:    Deps{Ctx: nil, Svc: &kvs.Service{}},
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
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})

	cmd := m.Init()

	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestModel_SetDevice(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})

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
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})
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
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})

	m = m.SetSize(80, 24)

	if m.Width != 80 {
		t.Errorf("Width = %d, want 80", m.Width)
	}
	if m.Height != 24 {
		t.Errorf("Height = %d, want 24", m.Height)
	}
}

func TestModel_SetFocused(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})

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
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})
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
	if m.Cursor() != 0 {
		t.Error("cursor should be reset to 0")
	}
	if cmd != nil {
		t.Error("LoadedMsg should not return a command")
	}
}

func TestModel_Update_LoadedMsg_Error(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})
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
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})
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
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})

	testErr := errors.New("delete failed")
	msg := ActionMsg{Action: "delete", Key: "key1", Err: testErr}

	m, _ = m.Update(msg)

	if !errors.Is(m.err, testErr) {
		t.Errorf("err = %v, want %v", m.err, testErr)
	}
}

func TestModel_HandleAction_NotFocused(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})
	m.focused = false
	m.items = []Item{{Key: "key1"}, {Key: "key2"}}
	m.Scroller.SetItemCount(len(m.items))

	msg := messages.NavigationMsg{Direction: messages.NavDown}

	m, _ = m.Update(msg)

	if m.Cursor() != 0 {
		t.Error("cursor should not change when not focused")
	}
}

func TestModel_HandleAction_Navigation(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})
	m.focused = true
	m.items = []Item{{Key: "key1"}, {Key: "key2"}, {Key: "key3"}}
	m.Scroller.SetItemCount(len(m.items))
	m = m.SetSize(80, 20)

	// Test NavDown
	msg := messages.NavigationMsg{Direction: messages.NavDown}
	m, _ = m.Update(msg)
	if m.Cursor() != 1 {
		t.Errorf("cursor after NavDown = %d, want 1", m.Cursor())
	}

	// Test NavUp
	msg = messages.NavigationMsg{Direction: messages.NavUp}
	m, _ = m.Update(msg)
	if m.Cursor() != 0 {
		t.Errorf("cursor after NavUp = %d, want 0", m.Cursor())
	}

	// Test NavEnd (go to end)
	msg = messages.NavigationMsg{Direction: messages.NavEnd}
	m, _ = m.Update(msg)
	if m.Cursor() != 2 {
		t.Errorf("cursor after NavEnd = %d, want 2", m.Cursor())
	}

	// Test NavHome (go to start)
	msg = messages.NavigationMsg{Direction: messages.NavHome}
	m, _ = m.Update(msg)
	if m.Cursor() != 0 {
		t.Errorf("cursor after NavHome = %d, want 0", m.Cursor())
	}
}

func TestModel_ScrollerCursorBounds(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})
	m.items = []Item{{Key: "key1"}, {Key: "key2"}}
	m.Scroller.SetItemCount(len(m.items))
	m = m.SetSize(80, 20)

	// Cursor should not go below 0
	m.Scroller.CursorUp()
	if m.Cursor() != 0 {
		t.Errorf("cursor should stay at 0, got %d", m.Cursor())
	}

	// Cursor should not exceed items
	m.Scroller.SetCursor(1)
	m.Scroller.CursorDown()
	if m.Cursor() != 1 {
		t.Errorf("cursor should stay at 1, got %d", m.Cursor())
	}
}

func TestModel_View_NoDevice(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})
	m = m.SetSize(40, 10)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_Loading(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})
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
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})
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
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})
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
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})
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
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})

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
			result := m.formatValueWithWidth(tt.value, 30)
			if result == "" {
				t.Errorf("formatValueWithWidth(%v) returned empty string", tt.value)
			}
		})
	}
}

func TestModel_SelectedItem(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})

	// No items
	if m.SelectedItem() != nil {
		t.Error("SelectedItem should return nil when no items")
	}

	// With items
	m.items = []Item{
		{Key: "key1", Value: "value1"},
		{Key: "key2", Value: "value2"},
	}
	m.Scroller.SetItemCount(len(m.items))
	m.Scroller.SetCursor(1)

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
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})

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
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})

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
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})

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
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})

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
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})

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

func TestModel_ScrollerVisibleRows(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})
	m.items = make([]Item, 20)
	m.Scroller.SetItemCount(20)

	// SetSize configures visible rows (height - 4 overhead)
	m = m.SetSize(80, 20)
	if m.Scroller.VisibleRows() != 16 {
		t.Errorf("visibleRows = %d, want 16", m.Scroller.VisibleRows())
	}

	// Small height
	m = m.SetSize(80, 5)
	if m.Scroller.VisibleRows() < 1 {
		t.Errorf("visibleRows with small height = %d, want >= 1", m.Scroller.VisibleRows())
	}
}

func TestModel_ScrollerEnsureVisible(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})
	m.items = make([]Item, 20)
	for i := range m.items {
		m.items[i] = Item{Key: string(rune('a' + i))}
	}
	m.Scroller.SetItemCount(20)
	m = m.SetSize(80, 10) // Sets visibleRows = 10 - 4 = 6

	// Cursor at end should scroll
	m.Scroller.CursorToEnd()
	start, _ := m.Scroller.VisibleRange()
	if start == 0 {
		t.Error("scroll should increase when cursor at end of long list")
	}

	// Cursor back to start
	m.Scroller.CursorToStart()
	start, _ = m.Scroller.VisibleRange()
	if start != 0 {
		t.Errorf("scroll = %d, want 0 when cursor at beginning", start)
	}
}

func TestModel_SelectItem_Empty(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})
	m.focused = true

	cmd := m.selectItem()
	if cmd != nil {
		t.Error("selectItem() should return nil when no items")
	}
}

func TestModel_DeleteItem_Empty(t *testing.T) {
	t.Parallel()
	m := New(Deps{Ctx: context.Background(), Svc: &kvs.Service{}})
	m.focused = true

	// With no items, handleDeleteKey returns nil and doesn't start confirmation
	m, _ = m.handleDeleteKey()
	if m.confirmingDelete {
		t.Error("confirmingDelete should be false when no items")
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
