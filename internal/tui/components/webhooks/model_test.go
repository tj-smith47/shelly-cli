package webhooks

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestWebhook(t *testing.T) {
	t.Parallel()
	w := Webhook{
		ID:     1,
		Name:   "Test Hook",
		Event:  "switch.on",
		Enable: true,
		URLs:   []string{"http://example.com/webhook"},
		Cid:    0,
	}

	if w.ID != 1 {
		t.Errorf("ID = %d, want 1", w.ID)
	}
	if w.Name != "Test Hook" {
		t.Errorf("Name = %q, want %q", w.Name, "Test Hook")
	}
	if !w.Enable {
		t.Error("Enable = false, want true")
	}
	if len(w.URLs) != 1 {
		t.Errorf("URLs length = %d, want 1", len(w.URLs))
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
	webhooks := []Webhook{
		{ID: 1, Event: "switch.on"},
		{ID: 2, Event: "switch.off"},
		{ID: 3, Event: "button.press"},
	}

	m := Model{
		webhooks: webhooks,
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
	webhooks := make([]Webhook, 20)
	for i := range webhooks {
		webhooks[i] = Webhook{ID: i, Event: "event"}
	}

	m := Model{
		webhooks: webhooks,
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

func TestModel_SelectedWebhook(t *testing.T) {
	t.Parallel()
	webhooks := []Webhook{
		{ID: 1, Event: "switch.on"},
		{ID: 2, Event: "switch.off"},
	}

	m := Model{
		webhooks: webhooks,
		cursor:   1,
	}

	selected := m.SelectedWebhook()
	if selected == nil {
		t.Fatal("SelectedWebhook() = nil, want webhook")
	}
	if selected.ID != 2 {
		t.Errorf("SelectedWebhook().ID = %d, want 2", selected.ID)
	}

	// Empty list
	m.webhooks = nil
	selected = m.SelectedWebhook()
	if selected != nil {
		t.Error("SelectedWebhook() on empty list should return nil")
	}
}

func TestModel_WebhookCount(t *testing.T) {
	t.Parallel()
	m := Model{
		webhooks: []Webhook{{ID: 1}, {ID: 2}, {ID: 3}},
	}

	if got := m.WebhookCount(); got != 3 {
		t.Errorf("WebhookCount() = %d, want 3", got)
	}

	m.webhooks = nil
	if got := m.WebhookCount(); got != 0 {
		t.Errorf("WebhookCount() on nil = %d, want 0", got)
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

func TestModel_View_NoWebhooks(t *testing.T) {
	t.Parallel()
	m := Model{
		device:   "192.168.1.100",
		loading:  false,
		webhooks: []Webhook{},
		width:    40,
		height:   10,
		styles:   DefaultStyles(),
	}

	view := m.View()
	if !strings.Contains(view, "No webhooks") {
		t.Errorf("View() should show 'No webhooks', got:\n%s", view)
	}
}

func TestModel_View_WithWebhooks(t *testing.T) {
	t.Parallel()
	m := Model{
		device:  "192.168.1.100",
		loading: false,
		webhooks: []Webhook{
			{
				ID:     1,
				Event:  "switch.on",
				Enable: true,
				URLs:   []string{"http://example.com/hook1"},
			},
			{
				ID:     2,
				Event:  "switch.off",
				Enable: false,
				URLs:   []string{"http://example.com/hook2"},
			},
		},
		cursor: 0,
		width:  60,
		height: 15,
		styles: DefaultStyles(),
	}

	view := m.View()
	if !strings.Contains(view, "switch.on") {
		t.Errorf("View() should show event name, got:\n%s", view)
	}
}

func TestModel_RenderWebhookLine(t *testing.T) {
	t.Parallel()
	m := Model{styles: DefaultStyles()}

	tests := []struct {
		name       string
		webhook    Webhook
		isSelected bool
	}{
		{
			name: "enabled webhook selected",
			webhook: Webhook{
				ID:     1,
				Event:  "switch.on",
				Enable: true,
				URLs:   []string{"http://example.com"},
			},
			isSelected: true,
		},
		{
			name: "disabled webhook",
			webhook: Webhook{
				ID:     2,
				Event:  "switch.off",
				Enable: false,
				URLs:   []string{"http://example.com"},
			},
			isSelected: false,
		},
		{
			name: "multiple URLs",
			webhook: Webhook{
				ID:     3,
				Event:  "button.press",
				Enable: true,
				URLs: []string{
					"http://example.com/hook1",
					"http://example.com/hook2",
				},
			},
			isSelected: false,
		},
		{
			name: "no URLs",
			webhook: Webhook{
				ID:     4,
				Event:  "input.toggle",
				Enable: true,
				URLs:   []string{},
			},
			isSelected: false,
		},
		{
			name: "long event name",
			webhook: Webhook{
				ID:     5,
				Event:  "very_long_event_name_that_should_be_truncated_eventually",
				Enable: true,
				URLs:   []string{"http://example.com"},
			},
			isSelected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			line := m.renderWebhookLine(tt.webhook, tt.isSelected)
			if line == "" {
				t.Error("renderWebhookLine() returned empty string")
			}
			if tt.isSelected && !strings.Contains(line, "▶") {
				t.Errorf("selected line should contain ▶, got: %s", line)
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

	webhooks := []Webhook{
		{ID: 1, Event: "switch.on"},
		{ID: 2, Event: "switch.off"},
	}

	m, _ = m.Update(LoadedMsg{Webhooks: webhooks})

	if m.loading {
		t.Error("loading should be false after LoadedMsg")
	}
	if len(m.webhooks) != 2 {
		t.Errorf("webhooks length = %d, want 2", len(m.webhooks))
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
		webhooks: []Webhook{{ID: 1}},
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

	_ = styles.Enabled.Render("test")
	_ = styles.Disabled.Render("test")
	_ = styles.Event.Render("test")
	_ = styles.URL.Render("test")
	_ = styles.Name.Render("test")
	_ = styles.Selected.Render("test")
	_ = styles.Error.Render("test")
	_ = styles.Muted.Render("test")
}
