package views

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/focus"
	"github.com/tj-smith47/shelly-cli/internal/tui/tabs"
)

func TestNewManage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc := &shelly.Service{}
	focusState := focus.NewState()
	focusState.SetActiveTab(tabs.TabManage)
	deps := ManageDeps{Ctx: ctx, Svc: svc, FocusState: focusState}

	m := NewManage(deps)

	if m.ctx != ctx {
		t.Error("ctx not set")
	}
	if m.svc != svc {
		t.Error("svc not set")
	}
	if m.id != tabs.TabManage {
		t.Errorf("id = %v, want tabs.TabManage", m.id)
	}
}

func TestNewManage_PanicOnNilCtx(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil ctx")
		}
	}()

	focusState := focus.NewState()
	deps := ManageDeps{Ctx: nil, Svc: &shelly.Service{}, FocusState: focusState}
	NewManage(deps)
}

func TestNewManage_PanicOnNilSvc(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil svc")
		}
	}()

	focusState := focus.NewState()
	deps := ManageDeps{Ctx: context.Background(), Svc: nil, FocusState: focusState}
	NewManage(deps)
}

func TestManageDeps_Validate(t *testing.T) {
	t.Parallel()
	focusState := focus.NewState()
	tests := []struct {
		name    string
		deps    ManageDeps
		wantErr bool
	}{
		{
			name:    "valid",
			deps:    ManageDeps{Ctx: context.Background(), Svc: &shelly.Service{}, FocusState: focusState},
			wantErr: false,
		},
		{
			name:    "nil ctx",
			deps:    ManageDeps{Ctx: nil, Svc: &shelly.Service{}, FocusState: focusState},
			wantErr: true,
		},
		{
			name:    "nil svc",
			deps:    ManageDeps{Ctx: context.Background(), Svc: nil, FocusState: focusState},
			wantErr: true,
		},
		{
			name:    "nil focus state",
			deps:    ManageDeps{Ctx: context.Background(), Svc: &shelly.Service{}, FocusState: nil},
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

func TestManage_Init(t *testing.T) {
	t.Parallel()
	m := newTestManage()

	// Init may return nil or a batch command depending on component implementations
	_ = m.Init()
}

func TestManage_ID(t *testing.T) {
	t.Parallel()
	m := newTestManage()

	if m.ID() != tabs.TabManage {
		t.Errorf("ID() = %v, want tabs.TabManage", m.ID())
	}
}

func TestManage_SetSize(t *testing.T) {
	t.Parallel()
	m := newTestManage()

	result := m.SetSize(100, 50)
	updated, ok := result.(*Manage)
	if !ok {
		t.Fatal("SetSize should return *Manage")
	}

	if updated.width != 100 {
		t.Errorf("width = %d, want 100", updated.width)
	}
	if updated.height != 50 {
		t.Errorf("height = %d, want 50", updated.height)
	}
}

func TestManage_Update_FocusNext(t *testing.T) {
	t.Parallel()
	m := newTestManage()
	initialPanel := m.focusState.ActivePanel()

	msg := tea.KeyPressMsg{Code: tea.KeyTab}
	updated, _ := m.Update(msg)
	manage, ok := updated.(*Manage)
	if !ok {
		t.Fatal("Update should return *Manage")
	}

	newPanel := manage.focusState.ActivePanel()
	if newPanel == initialPanel {
		t.Error("Tab should change focused panel")
	}
}

func TestManage_Update_FocusPrev(t *testing.T) {
	t.Parallel()
	m := newTestManage()
	// Move to second panel first
	m.focusState.NextPanel()
	panelAfterNext := m.focusState.ActivePanel()

	msg := tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	updated, _ := m.Update(msg)
	manage, ok := updated.(*Manage)
	if !ok {
		t.Fatal("Update should return *Manage")
	}

	panelAfterPrev := manage.focusState.ActivePanel()
	if panelAfterPrev == panelAfterNext {
		t.Error("Shift+Tab should change focused panel")
	}
}

func TestManage_FocusCycle(t *testing.T) {
	t.Parallel()
	m := newTestManage()

	// Start at first panel
	initialPanel := m.focusState.ActivePanel()

	// Panel order: Discovery -> Firmware -> Backup -> Scenes -> Templates -> Batch (6 panels)
	m.focusState.NextPanel()
	p1 := m.focusState.ActivePanel()
	if p1 == initialPanel {
		t.Error("NextPanel should change panel")
	}

	m.focusState.NextPanel()
	p2 := m.focusState.ActivePanel()
	if p2 == p1 {
		t.Error("NextPanel should change panel again")
	}

	m.focusState.NextPanel()
	p3 := m.focusState.ActivePanel()
	if p3 == p2 {
		t.Error("NextPanel should change panel again")
	}

	m.focusState.NextPanel()
	p4 := m.focusState.ActivePanel()
	if p4 == p3 {
		t.Error("NextPanel should change panel again")
	}

	m.focusState.NextPanel()
	p5 := m.focusState.ActivePanel()
	if p5 == p4 {
		t.Error("NextPanel should change panel again")
	}

	// After cycling through all 6 panels, should wrap back
	m.focusState.NextPanel()
	p6 := m.focusState.ActivePanel()
	if p6 != initialPanel {
		t.Errorf("After cycling through all 6 panels, should wrap back to initial, got %v", p6)
	}
}

func TestManage_View_Empty(t *testing.T) {
	t.Parallel()
	m := newTestManage()

	// Without SetSize, should return empty
	view := m.View()
	if view != "" {
		t.Error("View() without SetSize should return empty string")
	}
}

func TestManage_View_WithSize(t *testing.T) {
	t.Parallel()
	m := newTestManage()
	result, ok := m.SetSize(100, 50).(*Manage)
	if !ok {
		t.Fatal("SetSize should return *Manage")
	}
	m = result

	view := m.View()

	if view == "" {
		t.Error("View() with SetSize should not return empty string")
	}
}

func TestManage_Refresh(t *testing.T) {
	t.Parallel()
	m := newTestManage()

	cmd := m.Refresh()

	if cmd == nil {
		t.Error("Refresh should return a command")
	}
}

func TestManage_Accessors(t *testing.T) {
	t.Parallel()
	m := newTestManage()

	// Verify components are accessible
	_ = m.Discovery()
	_ = m.Batch()
	_ = m.Firmware()
	_ = m.Backup()
	_ = m.Scenes()
	_ = m.Templates()
}

func TestManage_StatusSummary(t *testing.T) {
	t.Parallel()
	m := newTestManage()

	summary := m.StatusSummary()

	if summary == "" {
		t.Error("StatusSummary() should not return empty string")
	}
}

func TestManage_UpdateFocusStates(t *testing.T) {
	t.Parallel()
	m := newTestManage()

	// Set focus to a different panel
	m.focusState.NextPanel()
	m.updateFocusStates()

	// Verify focus states are updated (method doesn't panic)
	// Access model to verify it's accessible
	_ = m.discovery.Scanning()
}

func TestDefaultManageStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultManageStyles()

	// Verify styles are created without panic
	_ = styles.Panel.Render("test")
	_ = styles.PanelActive.Render("test")
	_ = styles.Title.Render("test")
	_ = styles.Muted.Render("test")
}

func newTestManage() *Manage {
	ctx := context.Background()
	svc := &shelly.Service{}
	focusState := focus.NewState()
	// Set to manage tab so panel cycling works correctly
	focusState.SetActiveTab(tabs.TabManage)
	deps := ManageDeps{Ctx: ctx, Svc: svc, FocusState: focusState}
	return NewManage(deps)
}
