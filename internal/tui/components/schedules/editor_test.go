package schedules

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
)

func TestNewEditor(t *testing.T) {
	t.Parallel()
	m := NewEditor()
	if m.schedule != nil {
		t.Error("schedule should be nil initially")
	}
}

func TestEditorModel_SetSize(t *testing.T) {
	t.Parallel()
	m := EditorModel{}
	m = m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}

func TestEditorModel_SetFocused(t *testing.T) {
	t.Parallel()
	m := EditorModel{}

	m = m.SetFocused(true)
	if !m.focused {
		t.Error("focused = false, want true")
	}

	m = m.SetFocused(false)
	if m.focused {
		t.Error("focused = true, want false")
	}
}

func TestEditorModel_SetSchedule(t *testing.T) {
	t.Parallel()
	m := NewEditor()
	schedule := &Schedule{
		ID:       1,
		Enable:   true,
		Timespec: "0 0 9 * * *",
		Calls:    []automation.ScheduleCall{{Method: "Switch.Set"}},
	}

	m = m.SetSchedule(schedule)

	if m.schedule == nil {
		t.Fatal("schedule should not be nil")
	}
	if m.schedule.ID != 1 {
		t.Errorf("schedule.ID = %d, want 1", m.schedule.ID)
	}
	if m.scroll != 0 {
		t.Errorf("scroll = %d, want 0", m.scroll)
	}
}

func TestEditorModel_Clear(t *testing.T) {
	t.Parallel()
	m := NewEditor()
	m = m.SetSchedule(&Schedule{ID: 1})
	m.scroll = 5

	m = m.Clear()

	if m.schedule != nil {
		t.Error("schedule should be nil after Clear")
	}
	if m.scroll != 0 {
		t.Errorf("scroll = %d, want 0", m.scroll)
	}
}

func TestEditorModel_HasSchedule(t *testing.T) {
	t.Parallel()
	m := NewEditor()

	if m.HasSchedule() {
		t.Error("HasSchedule() = true, want false")
	}

	m = m.SetSchedule(&Schedule{ID: 1})
	if !m.HasSchedule() {
		t.Error("HasSchedule() = false, want true")
	}
}

func TestEditorModel_Schedule(t *testing.T) {
	t.Parallel()
	m := NewEditor()

	if m.Schedule() != nil {
		t.Error("Schedule() should return nil initially")
	}

	schedule := &Schedule{ID: 5}
	m = m.SetSchedule(schedule)

	if m.Schedule() == nil {
		t.Fatal("Schedule() should not return nil")
	}
	if m.Schedule().ID != 5 {
		t.Errorf("Schedule().ID = %d, want 5", m.Schedule().ID)
	}
}

func TestEditorModel_VisibleLines(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		height int
		want   int
	}{
		{"zero height", 0, 1},
		{"small height", 5, 1},
		{"normal height", 20, 16},
		{"large height", 50, 46},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := EditorModel{height: tt.height}
			got := m.visibleLines()
			if got != tt.want {
				t.Errorf("visibleLines() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestEditorModel_MaxScroll(t *testing.T) {
	t.Parallel()
	m := EditorModel{height: 20}

	// No schedule
	if got := m.maxScroll(); got != 0 {
		t.Errorf("maxScroll() with no schedule = %d, want 0", got)
	}

	// With schedule and calls
	m.schedule = &Schedule{
		ID: 1,
		Calls: []automation.ScheduleCall{
			{Method: "Switch.Set"},
			{Method: "Light.Set"},
		},
	}
	got := m.maxScroll()
	// Should be > 0 since we have content
	if got < 0 {
		t.Errorf("maxScroll() should be >= 0, got %d", got)
	}
}

func TestEditorModel_ScrollNavigation(t *testing.T) {
	t.Parallel()
	m := EditorModel{
		height: 10,
		schedule: &Schedule{
			ID:    1,
			Calls: make([]automation.ScheduleCall, 10), // Lots of calls for scrolling
		},
	}

	// Scroll down
	m = m.scrollDown()
	if m.scroll != 1 {
		t.Errorf("after scrollDown: scroll = %d, want 1", m.scroll)
	}

	// Scroll up
	m = m.scrollUp()
	if m.scroll != 0 {
		t.Errorf("after scrollUp: scroll = %d, want 0", m.scroll)
	}

	// Don't go below 0
	m = m.scrollUp()
	if m.scroll != 0 {
		t.Errorf("scroll below 0: scroll = %d, want 0", m.scroll)
	}

	// Scroll to end
	m = m.scrollToEnd()
	maxScroll := m.maxScroll()
	if m.scroll != maxScroll {
		t.Errorf("scrollToEnd: scroll = %d, want %d", m.scroll, maxScroll)
	}
}

func TestEditorModel_View_NoSchedule(t *testing.T) {
	t.Parallel()
	m := NewEditor().SetSize(50, 20)

	view := m.View()
	if !strings.Contains(view, "No schedule selected") {
		t.Errorf("View() should show 'No schedule selected', got:\n%s", view)
	}
}

func TestEditorModel_View_WithSchedule(t *testing.T) {
	t.Parallel()
	m := NewEditor().SetSize(60, 30)
	m = m.SetSchedule(&Schedule{
		ID:       1,
		Enable:   true,
		Timespec: "0 0 9 * * MON-FRI",
		Calls: []automation.ScheduleCall{
			{Method: "Switch.Set", Params: map[string]any{"on": true}},
		},
	})

	view := m.View()
	if !strings.Contains(view, "Schedule #1") {
		t.Errorf("View() should show schedule ID, got:\n%s", view)
	}
	if !strings.Contains(view, "Switch.Set") {
		t.Errorf("View() should show method name, got:\n%s", view)
	}
}

func TestEditorModel_View_DisabledSchedule(t *testing.T) {
	t.Parallel()
	m := NewEditor().SetSize(60, 30)
	m = m.SetSchedule(&Schedule{
		ID:       2,
		Enable:   false,
		Timespec: "0 0 18 * * *",
		Calls:    []automation.ScheduleCall{},
	})

	view := m.View()
	if !strings.Contains(view, "Disabled") {
		t.Errorf("View() should show 'Disabled', got:\n%s", view)
	}
}

func TestEditorModel_ExplainTimespec(t *testing.T) {
	t.Parallel()
	m := NewEditor()

	tests := []struct {
		name     string
		spec     string
		contains string
	}{
		{"sunrise", "@sunrise", "sunrise"},
		{"sunrise offset", "@sunrise+30", "After sunrise"},
		{"sunset", "@sunset", "sunset"},
		{"sunset before", "@sunset-15", "Before sunset"},
		{"daily", "0 30 9 * * *", "Daily at 9:30"},
		{"weekdays", "0 0 8 * * MON-FRI", "Weekdays"},
		{"weekends", "0 0 10 * * SAT,SUN", "Weekends"},
		{"complex", "0 */15 * * * *", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := m.explainTimespec(tt.spec)
			if tt.contains == "" {
				if got != "" {
					t.Errorf("explainTimespec(%q) = %q, want empty", tt.spec, got)
				}
			} else {
				if !strings.Contains(got, tt.contains) {
					t.Errorf("explainTimespec(%q) = %q, should contain %q", tt.spec, got, tt.contains)
				}
			}
		})
	}
}

func TestEditorModel_RenderCall(t *testing.T) {
	t.Parallel()
	m := NewEditor()

	tests := []struct {
		name string
		call automation.ScheduleCall
	}{
		{
			name: "simple method",
			call: automation.ScheduleCall{Method: "Switch.Set"},
		},
		{
			name: "with params",
			call: automation.ScheduleCall{
				Method: "Light.Set",
				Params: map[string]any{"brightness": 50},
			},
		},
		{
			name: "no params",
			call: automation.ScheduleCall{
				Method: "Sys.Reboot",
				Params: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := m.renderCall(1, tt.call)
			if result == "" {
				t.Error("renderCall() returned empty string")
			}
			if !strings.Contains(result, tt.call.Method) {
				t.Errorf("renderCall() should contain method name, got: %s", result)
			}
		})
	}
}

func TestEditorModel_Init(t *testing.T) {
	t.Parallel()
	m := NewEditor()
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestDefaultEditorStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultEditorStyles()

	_ = styles.Header.Render("test")
	_ = styles.Label.Render("test")
	_ = styles.Value.Render("test")
	_ = styles.Enabled.Render("test")
	_ = styles.Disabled.Render("test")
	_ = styles.Method.Render("test")
	_ = styles.Params.Render("test")
	_ = styles.Separator.Render("test")
	_ = styles.Muted.Render("test")
}
