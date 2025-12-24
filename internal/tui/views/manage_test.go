package views

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/tabs"
)

func TestNewManage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc := &shelly.Service{}
	deps := ManageDeps{Ctx: ctx, Svc: svc}

	m := NewManage(deps)

	if m.ctx != ctx {
		t.Error("ctx not set")
	}
	if m.svc != svc {
		t.Error("svc not set")
	}
	if m.focusedPanel != ManagePanelDiscovery {
		t.Errorf("focusedPanel = %v, want ManagePanelDiscovery", m.focusedPanel)
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

	deps := ManageDeps{Ctx: nil, Svc: &shelly.Service{}}
	NewManage(deps)
}

func TestNewManage_PanicOnNilSvc(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil svc")
		}
	}()

	deps := ManageDeps{Ctx: context.Background(), Svc: nil}
	NewManage(deps)
}

func TestManageDeps_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		deps    ManageDeps
		wantErr bool
	}{
		{
			name:    "valid",
			deps:    ManageDeps{Ctx: context.Background(), Svc: &shelly.Service{}},
			wantErr: false,
		},
		{
			name:    "nil ctx",
			deps:    ManageDeps{Ctx: nil, Svc: &shelly.Service{}},
			wantErr: true,
		},
		{
			name:    "nil svc",
			deps:    ManageDeps{Ctx: context.Background(), Svc: nil},
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
	m.focusedPanel = ManagePanelDiscovery

	msg := tea.KeyPressMsg{Code: tea.KeyTab}
	updated, _ := m.Update(msg)
	manage, ok := updated.(*Manage)
	if !ok {
		t.Fatal("Update should return *Manage")
	}

	if manage.focusedPanel != ManagePanelBatch {
		t.Errorf("focusedPanel after tab = %v, want ManagePanelBatch", manage.focusedPanel)
	}
}

func TestManage_Update_FocusPrev(t *testing.T) {
	t.Parallel()
	m := newTestManage()
	m.focusedPanel = ManagePanelBatch

	msg := tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	updated, _ := m.Update(msg)
	manage, ok := updated.(*Manage)
	if !ok {
		t.Fatal("Update should return *Manage")
	}

	if manage.focusedPanel != ManagePanelDiscovery {
		t.Errorf("focusedPanel after shift+tab = %v, want ManagePanelDiscovery", manage.focusedPanel)
	}
}

func TestManage_Update_NumberKeyFocus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		key      rune
		expected ManagePanel
	}{
		{'1', ManagePanelDiscovery},
		{'2', ManagePanelBatch},
		{'3', ManagePanelFirmware},
		{'4', ManagePanelBackup},
	}

	for _, tt := range tests {
		t.Run(string(tt.key), func(t *testing.T) {
			t.Parallel()
			m := newTestManage()

			msg := tea.KeyPressMsg{Code: tt.key}
			updated, _ := m.Update(msg)
			manage, ok := updated.(*Manage)
			if !ok {
				t.Fatal("Update should return *Manage")
			}

			if manage.focusedPanel != tt.expected {
				t.Errorf("focusedPanel after '%c' = %v, want %v", tt.key, manage.focusedPanel, tt.expected)
			}
		})
	}
}

func TestManage_FocusCycle(t *testing.T) {
	t.Parallel()
	m := newTestManage()

	// Start at discovery
	if m.focusedPanel != ManagePanelDiscovery {
		t.Fatal("should start at ManagePanelDiscovery")
	}

	// Cycle through all panels
	m.focusNext()
	if m.focusedPanel != ManagePanelBatch {
		t.Errorf("after 1 focusNext = %v, want ManagePanelBatch", m.focusedPanel)
	}

	m.focusNext()
	if m.focusedPanel != ManagePanelFirmware {
		t.Errorf("after 2 focusNext = %v, want ManagePanelFirmware", m.focusedPanel)
	}

	m.focusNext()
	if m.focusedPanel != ManagePanelBackup {
		t.Errorf("after 3 focusNext = %v, want ManagePanelBackup", m.focusedPanel)
	}

	m.focusNext()
	if m.focusedPanel != ManagePanelDiscovery {
		t.Errorf("after 4 focusNext = %v, want ManagePanelDiscovery (wrap)", m.focusedPanel)
	}
}

func TestManage_FocusPrevCycle(t *testing.T) {
	t.Parallel()
	m := newTestManage()

	// Start at discovery
	m.focusedPanel = ManagePanelDiscovery

	// Go backwards
	m.focusPrev()
	if m.focusedPanel != ManagePanelBackup {
		t.Errorf("after 1 focusPrev = %v, want ManagePanelBackup (wrap)", m.focusedPanel)
	}

	m.focusPrev()
	if m.focusedPanel != ManagePanelFirmware {
		t.Errorf("after 2 focusPrev = %v, want ManagePanelFirmware", m.focusedPanel)
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
	m.focusedPanel = ManagePanelFirmware

	if m.FocusedPanel() != ManagePanelFirmware {
		t.Errorf("FocusedPanel() = %v, want ManagePanelFirmware", m.FocusedPanel())
	}

	// Verify components are accessible
	_ = m.Discovery()
	_ = m.Batch()
	_ = m.Firmware()
	_ = m.Backup()
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

	// Set focus to batch
	m.focusedPanel = ManagePanelBatch
	m.updateFocusStates()

	// Verify focus states are updated (method doesn't panic)
	// Discovery should not be focused since we set focus to batch
	_ = m.discovery.Scanning() // Access model to verify it's accessible
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
	deps := ManageDeps{Ctx: ctx, Svc: svc}
	return NewManage(deps)
}
