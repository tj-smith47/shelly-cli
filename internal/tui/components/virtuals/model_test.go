package virtuals

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

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
	m := Model{}
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
	m := Model{}

	m = m.SetFocused(true)
	if !m.focused {
		t.Error("focused = false, want true")
	}

	m = m.SetFocused(false)
	if m.focused {
		t.Error("focused = true, want false")
	}
}

func TestModel_VisibleRows(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		height int
		want   int
	}{
		{"zero height", 0, 1},
		{"small height", 5, 1},
		{"normal height", 20, 16},
		{"large height", 100, 96},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := Model{height: tt.height}
			got := m.visibleRows()
			if got != tt.want {
				t.Errorf("visibleRows() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestModel_CursorNavigation(t *testing.T) {
	t.Parallel()
	virtuals := []Virtual{
		{Key: "boolean:200", Type: shelly.VirtualBoolean},
		{Key: "number:201", Type: shelly.VirtualNumber},
		{Key: "text:202", Type: shelly.VirtualText},
	}

	m := Model{
		virtuals: virtuals,
		cursor:   0,
		height:   20,
	}

	// Cursor down
	m = m.cursorDown()
	if m.cursor != 1 {
		t.Errorf("after cursorDown: cursor = %d, want 1", m.cursor)
	}

	m = m.cursorDown()
	if m.cursor != 2 {
		t.Errorf("after 2nd cursorDown: cursor = %d, want 2", m.cursor)
	}

	// Don't go past end
	m = m.cursorDown()
	if m.cursor != 2 {
		t.Errorf("cursor at end: cursor = %d, want 2", m.cursor)
	}

	// Cursor up
	m = m.cursorUp()
	if m.cursor != 1 {
		t.Errorf("after cursorUp: cursor = %d, want 1", m.cursor)
	}

	// Don't go before start
	m.cursor = 0
	m = m.cursorUp()
	if m.cursor != 0 {
		t.Errorf("cursor at start: cursor = %d, want 0", m.cursor)
	}

	// Cursor to end
	m = m.cursorToEnd()
	if m.cursor != 2 {
		t.Errorf("after cursorToEnd: cursor = %d, want 2", m.cursor)
	}
}

func TestModel_EnsureVisible(t *testing.T) {
	t.Parallel()
	virtuals := make([]Virtual, 20)
	for i := range virtuals {
		virtuals[i] = Virtual{Key: "boolean:200"}
	}

	m := Model{
		virtuals: virtuals,
		height:   10,
		cursor:   0,
		scroll:   0,
	}

	// Move cursor past visible
	m.cursor = 10
	m = m.ensureVisible()
	if m.scroll < 5 {
		t.Errorf("scroll should adjust for cursor=10, got scroll=%d", m.scroll)
	}

	// Move cursor before scroll
	m.cursor = 2
	m = m.ensureVisible()
	if m.scroll != 2 {
		t.Errorf("scroll should adjust to cursor=2, got scroll=%d", m.scroll)
	}
}

func TestModel_SelectedVirtual(t *testing.T) {
	t.Parallel()
	virtuals := []Virtual{
		{Key: "boolean:200", ID: 200},
		{Key: "number:201", ID: 201},
	}

	m := Model{
		virtuals: virtuals,
		cursor:   1,
	}

	selected := m.SelectedVirtual()
	if selected == nil {
		t.Fatal("SelectedVirtual() = nil, want virtual")
	}
	if selected.ID != 201 {
		t.Errorf("SelectedVirtual().ID = %d, want 201", selected.ID)
	}

	// Empty list
	m.virtuals = nil
	selected = m.SelectedVirtual()
	if selected != nil {
		t.Error("SelectedVirtual() on empty list should return nil")
	}
}

func TestModel_VirtualCount(t *testing.T) {
	t.Parallel()
	m := Model{
		virtuals: []Virtual{{Key: "a"}, {Key: "b"}, {Key: "c"}},
	}

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
	m := Model{device: "192.168.1.100"}
	if got := m.Device(); got != "192.168.1.100" {
		t.Errorf("Device() = %q, want %q", got, "192.168.1.100")
	}
}

func TestModel_Loading(t *testing.T) {
	t.Parallel()
	m := Model{loading: true}
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
	m := Model{}
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
	m := Model{
		device: "",
		width:  40,
		height: 10,
		styles: DefaultStyles(),
	}

	view := m.View()
	if !strings.Contains(view, "No device selected") {
		t.Errorf("View() should show 'No device selected', got:\n%s", view)
	}
}

func TestModel_View_Loading(t *testing.T) {
	t.Parallel()
	m := Model{
		device:  "192.168.1.100",
		loading: true,
		width:   40,
		height:  10,
		styles:  DefaultStyles(),
	}

	view := m.View()
	if !strings.Contains(view, "Loading") {
		t.Errorf("View() should show 'Loading', got:\n%s", view)
	}
}

func TestModel_View_NoVirtuals(t *testing.T) {
	t.Parallel()
	m := Model{
		device:   "192.168.1.100",
		loading:  false,
		virtuals: []Virtual{},
		width:    40,
		height:   10,
		styles:   DefaultStyles(),
	}

	view := m.View()
	if !strings.Contains(view, "No virtual") {
		t.Errorf("View() should show 'No virtual', got:\n%s", view)
	}
}

func TestModel_View_WithVirtuals(t *testing.T) {
	t.Parallel()
	boolVal := true
	numVal := 22.5

	m := Model{
		device:  "192.168.1.100",
		loading: false,
		virtuals: []Virtual{
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
		},
		cursor: 0,
		width:  60,
		height: 15,
		styles: DefaultStyles(),
	}

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
	m := Model{
		loading: true,
		styles:  DefaultStyles(),
	}

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
	m := Model{
		loading: true,
		styles:  DefaultStyles(),
	}

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
	m := Model{
		focused:  false,
		virtuals: []Virtual{{Key: "a"}},
		cursor:   0,
		styles:   DefaultStyles(),
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: 106}) // 'j'

	if m.cursor != 0 {
		t.Errorf("cursor changed when not focused: %d", m.cursor)
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
