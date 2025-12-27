package virtuals

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

const testDeviceAddr = "192.168.1.100"

func TestVirtual(t *testing.T) {
	t.Parallel()
	boolVal := true
	numVal := 22.5

	v := Virtual{
		Key:       "boolean:200",
		Type:      shelly.VirtualBoolean,
		ID:        200,
		Name:      "Test",
		BoolValue: &boolVal,
		NumValue:  &numVal,
	}

	if v.Key != "boolean:200" {
		t.Errorf("Key = %q, want %q", v.Key, "boolean:200")
	}
	if v.Type != shelly.VirtualBoolean {
		t.Errorf("Type = %q, want %q", v.Type, shelly.VirtualBoolean)
	}
	if v.ID != 200 {
		t.Errorf("ID = %d, want 200", v.ID)
	}
}

func TestDeps_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		deps    Deps
		wantErr bool
	}{
		{
			name:    "nil context",
			deps:    Deps{Ctx: nil, Svc: nil},
			wantErr: true,
		},
		{
			name:    "nil service",
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

func TestModel_SetSize(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m = m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}

func TestModel_SetFocused(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	m = m.SetFocused(true)
	if !m.focused {
		t.Error("focused = false, want true")
	}

	m = m.SetFocused(false)
	if m.focused {
		t.Error("focused = true, want false")
	}
}

func TestModel_ScrollerVisibleRows(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.virtuals = make([]Virtual, 20)
	m.scroller.SetItemCount(20)

	// SetSize configures visible rows (height - 4 overhead)
	m = m.SetSize(80, 20)
	if m.scroller.VisibleRows() != 16 {
		t.Errorf("visibleRows = %d, want 16", m.scroller.VisibleRows())
	}

	// Small height
	m = m.SetSize(80, 5)
	if m.scroller.VisibleRows() < 1 {
		t.Errorf("visibleRows with small height = %d, want >= 1", m.scroller.VisibleRows())
	}
}

func TestModel_ScrollerCursorNavigation(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.virtuals = []Virtual{
		{Key: "boolean:200", Type: shelly.VirtualBoolean},
		{Key: "number:201", Type: shelly.VirtualNumber},
		{Key: "text:202", Type: shelly.VirtualText},
	}
	m.scroller.SetItemCount(len(m.virtuals))
	m = m.SetSize(80, 20)

	// Cursor down
	m.scroller.CursorDown()
	if m.Cursor() != 1 {
		t.Errorf("after CursorDown: cursor = %d, want 1", m.Cursor())
	}

	m.scroller.CursorDown()
	if m.Cursor() != 2 {
		t.Errorf("after 2nd CursorDown: cursor = %d, want 2", m.Cursor())
	}

	// Don't go past end
	m.scroller.CursorDown()
	if m.Cursor() != 2 {
		t.Errorf("cursor at end: cursor = %d, want 2", m.Cursor())
	}

	// Cursor up
	m.scroller.CursorUp()
	if m.Cursor() != 1 {
		t.Errorf("after CursorUp: cursor = %d, want 1", m.Cursor())
	}

	// Don't go before start
	m.scroller.CursorToStart()
	m.scroller.CursorUp()
	if m.Cursor() != 0 {
		t.Errorf("cursor at start: cursor = %d, want 0", m.Cursor())
	}

	// Cursor to end
	m.scroller.CursorToEnd()
	if m.Cursor() != 2 {
		t.Errorf("after CursorToEnd: cursor = %d, want 2", m.Cursor())
	}
}

func TestModel_ScrollerEnsureVisible(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.virtuals = make([]Virtual, 20)
	for i := range m.virtuals {
		m.virtuals[i] = Virtual{Key: "boolean:200"}
	}
	m.scroller.SetItemCount(20)
	m = m.SetSize(80, 10) // Sets visibleRows = 10 - 4 = 6

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

func TestModel_SelectedVirtual(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.virtuals = []Virtual{
		{Key: "boolean:200", ID: 200},
		{Key: "number:201", ID: 201},
	}
	m.scroller.SetItemCount(len(m.virtuals))
	m.scroller.SetCursor(1)

	selected := m.SelectedVirtual()
	if selected == nil {
		t.Fatal("SelectedVirtual() = nil, want virtual")
	}
	if selected.ID != 201 {
		t.Errorf("SelectedVirtual().ID = %d, want 201", selected.ID)
	}

	// Empty list
	m.virtuals = nil
	m.scroller.SetItemCount(0)
	selected = m.SelectedVirtual()
	if selected != nil {
		t.Error("SelectedVirtual() on empty list should return nil")
	}
}

func TestModel_VirtualCount(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.virtuals = []Virtual{{Key: "a"}, {Key: "b"}, {Key: "c"}}

	if got := m.VirtualCount(); got != 3 {
		t.Errorf("VirtualCount() = %d, want 3", got)
	}

	m.virtuals = nil
	if got := m.VirtualCount(); got != 0 {
		t.Errorf("VirtualCount() on nil = %d, want 0", got)
	}
}

func TestModel_Device(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDeviceAddr
	if got := m.Device(); got != testDeviceAddr {
		t.Errorf("Device() = %q, want %q", got, testDeviceAddr)
	}
}

func TestModel_Loading(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.loading = true
	if !m.Loading() {
		t.Error("Loading() = false, want true")
	}

	m.loading = false
	if m.Loading() {
		t.Error("Loading() = true, want false")
	}
}

func TestModel_Error(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	if err := m.Error(); err != nil {
		t.Errorf("Error() = %v, want nil", err)
	}

	m.err = context.DeadlineExceeded
	if err := m.Error(); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Error() = %v, want %v", err, context.DeadlineExceeded)
	}
}

func TestModel_View_NoDevice(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = ""
	m = m.SetSize(40, 10)

	view := m.View()
	if !strings.Contains(view, "No device selected") {
		t.Errorf("View() should show 'No device selected', got:\n%s", view)
	}
}

func TestModel_View_Loading(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDeviceAddr
	m.loading = true
	m = m.SetSize(40, 10)

	view := m.View()
	if !strings.Contains(view, "Loading") {
		t.Errorf("View() should show 'Loading', got:\n%s", view)
	}
}

func TestModel_View_NoVirtuals(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDeviceAddr
	m.loading = false
	m.virtuals = []Virtual{}
	m = m.SetSize(40, 10)

	view := m.View()
	if !strings.Contains(view, "No virtual") {
		t.Errorf("View() should show 'No virtual', got:\n%s", view)
	}
}

func TestModel_View_WithVirtuals(t *testing.T) {
	t.Parallel()
	boolVal := true
	numVal := 22.5

	m := newTestModel()
	m.device = testDeviceAddr
	m.loading = false
	m.virtuals = []Virtual{
		{
			Key:       "boolean:200",
			Type:      shelly.VirtualBoolean,
			ID:        200,
			Name:      "Light",
			BoolValue: &boolVal,
		},
		{
			Key:      "number:201",
			Type:     shelly.VirtualNumber,
			ID:       201,
			Name:     "Temp",
			NumValue: &numVal,
		},
	}
	m.scroller.SetItemCount(len(m.virtuals))
	m = m.SetSize(60, 15)

	view := m.View()
	if !strings.Contains(view, "BOOL") {
		t.Errorf("View() should show type indicator, got:\n%s", view)
	}
}

func TestModel_RenderVirtualLine(t *testing.T) {
	t.Parallel()
	m := Model{styles: DefaultStyles()}

	boolVal := true
	numVal := 42.0
	strVal := "hello"

	tests := []struct {
		name       string
		virtual    Virtual
		isSelected bool
	}{
		{
			name: "boolean selected",
			virtual: Virtual{
				Key:       "boolean:200",
				Type:      shelly.VirtualBoolean,
				ID:        200,
				Name:      "Light",
				BoolValue: &boolVal,
			},
			isSelected: true,
		},
		{
			name: "number",
			virtual: Virtual{
				Key:      "number:201",
				Type:     shelly.VirtualNumber,
				ID:       201,
				Name:     "Temperature",
				NumValue: &numVal,
			},
			isSelected: false,
		},
		{
			name: "text",
			virtual: Virtual{
				Key:      "text:202",
				Type:     shelly.VirtualText,
				ID:       202,
				Name:     "Note",
				StrValue: &strVal,
			},
			isSelected: false,
		},
		{
			name: "enum",
			virtual: Virtual{
				Key:      "enum:203",
				Type:     shelly.VirtualEnum,
				ID:       203,
				Name:     "Mode",
				Options:  []string{"off", "low", "high"},
				StrValue: &strVal,
			},
			isSelected: false,
		},
		{
			name: "button",
			virtual: Virtual{
				Key:  "button:204",
				Type: shelly.VirtualButton,
				ID:   204,
				Name: "Action",
			},
			isSelected: false,
		},
		{
			name: "group",
			virtual: Virtual{
				Key:  "group:205",
				Type: shelly.VirtualGroup,
				ID:   205,
				Name: "Group1",
			},
			isSelected: false,
		},
		{
			name: "no name",
			virtual: Virtual{
				Key:  "boolean:206",
				Type: shelly.VirtualBoolean,
				ID:   206,
			},
			isSelected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			line := m.renderVirtualLine(tt.virtual, tt.isSelected)
			if line == "" {
				t.Error("renderVirtualLine() returned empty string")
			}
			if tt.isSelected && !strings.Contains(line, "▶") {
				t.Errorf("selected line should contain ▶, got: %s", line)
			}
		})
	}
}

func TestModel_FormatValue(t *testing.T) {
	t.Parallel()
	m := Model{styles: DefaultStyles()}

	boolTrue := true
	boolFalse := false
	numVal := 42.5
	strVal := "test"
	unit := "°C"

	tests := []struct {
		name    string
		virtual Virtual
		want    string
	}{
		{
			name:    "boolean true",
			virtual: Virtual{Type: shelly.VirtualBoolean, BoolValue: &boolTrue},
			want:    "ON",
		},
		{
			name:    "boolean false",
			virtual: Virtual{Type: shelly.VirtualBoolean, BoolValue: &boolFalse},
			want:    "OFF",
		},
		{
			name:    "number with unit",
			virtual: Virtual{Type: shelly.VirtualNumber, NumValue: &numVal, Unit: &unit},
			want:    "42.5°C",
		},
		{
			name:    "button",
			virtual: Virtual{Type: shelly.VirtualButton},
			want:    "Press",
		},
		{
			name:    "text",
			virtual: Virtual{Type: shelly.VirtualText, StrValue: &strVal},
			want:    "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := m.formatValue(tt.virtual)
			if !strings.Contains(got, tt.want) {
				t.Errorf("formatValue() = %q, should contain %q", got, tt.want)
			}
		})
	}
}

func TestFindIndex(t *testing.T) {
	t.Parallel()
	slice := []string{"a", "b", "c"}

	tests := []struct {
		name string
		val  string
		want int
	}{
		{"first", "a", 0},
		{"middle", "b", 1},
		{"last", "c", 2},
		{"not found", "d", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := findIndex(slice, tt.val); got != tt.want {
				t.Errorf("findIndex(%q) = %d, want %d", tt.val, got, tt.want)
			}
		})
	}
}

func TestModel_Update_LoadedMsg(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.loading = true

	virtuals := []Virtual{
		{Key: "boolean:200"},
		{Key: "number:201"},
	}

	m, _ = m.Update(LoadedMsg{Virtuals: virtuals})

	if m.loading {
		t.Error("loading should be false after LoadedMsg")
	}
	if len(m.virtuals) != 2 {
		t.Errorf("virtuals length = %d, want 2", len(m.virtuals))
	}
}

func TestModel_Update_LoadedMsg_Error(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.loading = true

	m, _ = m.Update(LoadedMsg{Err: context.DeadlineExceeded})

	if m.loading {
		t.Error("loading should be false after error")
	}
	if !errors.Is(m.err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want %v", m.err, context.DeadlineExceeded)
	}
}

func TestModel_Update_KeyPress_NotFocused(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = false
	m.virtuals = []Virtual{{Key: "a"}}
	m.scroller.SetItemCount(len(m.virtuals))

	m, _ = m.Update(tea.KeyPressMsg{Code: 106}) // 'j'

	if m.Cursor() != 0 {
		t.Errorf("cursor changed when not focused: %d", m.Cursor())
	}
}

func TestModel_Init(t *testing.T) {
	t.Parallel()
	m := Model{}
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestDefaultStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultStyles()

	_ = styles.TypeBoolean.Render("test")
	_ = styles.TypeNumber.Render("test")
	_ = styles.TypeText.Render("test")
	_ = styles.TypeEnum.Render("test")
	_ = styles.TypeButton.Render("test")
	_ = styles.TypeGroup.Render("test")
	_ = styles.Value.Render("test")
	_ = styles.Name.Render("test")
	_ = styles.Selected.Render("test")
	_ = styles.Error.Render("test")
	_ = styles.Muted.Render("test")
}

func newTestModel() Model {
	ctx := context.Background()
	svc := &shelly.Service{}
	deps := Deps{Ctx: ctx, Svc: svc}
	return New(deps)
}
