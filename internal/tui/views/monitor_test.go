package views

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/monitor"
	"github.com/tj-smith47/shelly-cli/internal/tui/focus"
	"github.com/tj-smith47/shelly-cli/internal/tui/tabs"
)

func TestNewMonitor(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()

	if m.ctx == nil {
		t.Error("ctx not set")
	}
	if m.ID() != tabs.TabMonitor {
		t.Errorf("ID() = %v, want tabs.TabMonitor", m.ID())
	}
	if m.focusState == nil {
		t.Error("focusState not set")
	}
	if m.layout == nil {
		t.Error("layout not set")
	}
}

func TestNewMonitor_PanicOnNilCtx(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil ctx")
		}
	}()

	svc := &shelly.Service{}
	ios := iostreams.Test(nil, &bytes.Buffer{}, &bytes.Buffer{})
	es := automation.NewEventStream(svc)
	focusState := focus.NewState()
	deps := MonitorDeps{Ctx: nil, Svc: svc, IOS: ios, EventStream: es, FocusState: focusState}
	NewMonitor(deps)
}

func TestNewMonitor_PanicOnNilSvc(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil svc")
		}
	}()

	ios := iostreams.Test(nil, &bytes.Buffer{}, &bytes.Buffer{})
	svc := &shelly.Service{}
	es := automation.NewEventStream(svc)
	focusState := focus.NewState()
	deps := MonitorDeps{Ctx: context.Background(), Svc: nil, IOS: ios, EventStream: es, FocusState: focusState}
	NewMonitor(deps)
}

func TestNewMonitor_PanicOnNilFocusState(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil focus state")
		}
	}()

	svc := &shelly.Service{}
	ios := iostreams.Test(nil, &bytes.Buffer{}, &bytes.Buffer{})
	es := automation.NewEventStream(svc)
	deps := MonitorDeps{Ctx: context.Background(), Svc: svc, IOS: ios, EventStream: es, FocusState: nil}
	NewMonitor(deps)
}

func TestMonitorDeps_Validate(t *testing.T) {
	t.Parallel()
	svc := &shelly.Service{}
	ios := iostreams.Test(nil, &bytes.Buffer{}, &bytes.Buffer{})
	es := automation.NewEventStream(svc)
	focusState := focus.NewState()

	tests := []struct {
		name    string
		deps    MonitorDeps
		wantErr bool
	}{
		{
			name:    "valid",
			deps:    MonitorDeps{Ctx: context.Background(), Svc: svc, IOS: ios, EventStream: es, FocusState: focusState},
			wantErr: false,
		},
		{
			name:    "nil ctx",
			deps:    MonitorDeps{Ctx: nil, Svc: svc, IOS: ios, EventStream: es, FocusState: focusState},
			wantErr: true,
		},
		{
			name:    "nil svc",
			deps:    MonitorDeps{Ctx: context.Background(), Svc: nil, IOS: ios, EventStream: es, FocusState: focusState},
			wantErr: true,
		},
		{
			name:    "nil ios",
			deps:    MonitorDeps{Ctx: context.Background(), Svc: svc, IOS: nil, EventStream: es, FocusState: focusState},
			wantErr: true,
		},
		{
			name:    "nil event stream",
			deps:    MonitorDeps{Ctx: context.Background(), Svc: svc, IOS: ios, EventStream: nil, FocusState: focusState},
			wantErr: true,
		},
		{
			name:    "nil focus state",
			deps:    MonitorDeps{Ctx: context.Background(), Svc: svc, IOS: ios, EventStream: es, FocusState: nil},
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

func TestMonitorDeps_Errors(t *testing.T) {
	t.Parallel()
	svc := &shelly.Service{}
	ios := iostreams.Test(nil, &bytes.Buffer{}, &bytes.Buffer{})
	es := automation.NewEventStream(svc)
	focusState := focus.NewState()

	t.Run("nil context error", func(t *testing.T) {
		t.Parallel()
		deps := MonitorDeps{Ctx: nil, Svc: svc, IOS: ios, EventStream: es, FocusState: focusState}
		err := deps.Validate()
		if !errors.Is(err, errNilContext) {
			t.Errorf("Validate() error = %v, want errNilContext", err)
		}
	})

	t.Run("nil service error", func(t *testing.T) {
		t.Parallel()
		deps := MonitorDeps{Ctx: context.Background(), Svc: nil, IOS: ios, EventStream: es, FocusState: focusState}
		err := deps.Validate()
		if !errors.Is(err, errNilService) {
			t.Errorf("Validate() error = %v, want errNilService", err)
		}
	})

	t.Run("nil iostreams error", func(t *testing.T) {
		t.Parallel()
		deps := MonitorDeps{Ctx: context.Background(), Svc: svc, IOS: nil, EventStream: es, FocusState: focusState}
		err := deps.Validate()
		if !errors.Is(err, errNilIOStreams) {
			t.Errorf("Validate() error = %v, want errNilIOStreams", err)
		}
	})

	t.Run("nil event stream error", func(t *testing.T) {
		t.Parallel()
		deps := MonitorDeps{Ctx: context.Background(), Svc: svc, IOS: ios, EventStream: nil, FocusState: focusState}
		err := deps.Validate()
		if !errors.Is(err, errNilEventStream) {
			t.Errorf("Validate() error = %v, want errNilEventStream", err)
		}
	})

	t.Run("nil focus state error", func(t *testing.T) {
		t.Parallel()
		deps := MonitorDeps{Ctx: context.Background(), Svc: svc, IOS: ios, EventStream: es, FocusState: nil}
		err := deps.Validate()
		if !errors.Is(err, errNilFocusState) {
			t.Errorf("Validate() error = %v, want errNilFocusState", err)
		}
	})
}

func TestMonitor_Init(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()

	// Init returns a batch of sub-component init commands.
	_ = m.Init()

	// Just verify it doesn't panic.
}

func TestMonitor_SetSize(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()

	result := m.SetSize(120, 40)

	updated, ok := result.(*Monitor)
	if !ok {
		t.Fatal("SetSize should return *Monitor")
	}
	if updated.width != 120 {
		t.Errorf("width = %d, want 120", updated.width)
	}
	if updated.height != 40 {
		t.Errorf("height = %d, want 40", updated.height)
	}
}

func TestMonitor_SetSize_Narrow(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()

	result := m.SetSize(60, 30)

	updated, ok := result.(*Monitor)
	if !ok {
		t.Fatal("SetSize should return *Monitor")
	}
	if updated.width != 60 {
		t.Errorf("width = %d, want 60", updated.width)
	}
	if !updated.isNarrow() {
		t.Error("expected narrow mode for width < 80")
	}
}

func TestMonitor_FocusCycling(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()

	initialPanel := m.focusState.ActivePanel()

	// Tab/Shift+Tab are handled at the app level, not in the view.
	// Test focus cycling via the focusState directly.
	m.focusState.NextPanel()

	newPanel := m.focusState.ActivePanel()
	if newPanel == initialPanel {
		t.Error("NextPanel should change focused panel")
	}
}

func TestMonitor_FocusCycling_PrevPanel(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()

	// Move to second panel first
	m.focusState.NextPanel()
	panelAfterNext := m.focusState.ActivePanel()

	// Go back
	m.focusState.PrevPanel()

	panelAfterPrev := m.focusState.ActivePanel()
	if panelAfterPrev == panelAfterNext {
		t.Error("PrevPanel should change focused panel")
	}
}

func TestMonitor_HandleKeyPress_ShiftJump(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key     string
		keyMsg  tea.KeyPressMsg
		wantIdx int
	}{
		{"Shift+1", tea.KeyPressMsg{Code: '!'}, 1},
		{"Shift+2", tea.KeyPressMsg{Code: '@'}, 2},
		{"Shift+3", tea.KeyPressMsg{Code: '#'}, 3},
		{"Shift+4", tea.KeyPressMsg{Code: '$'}, 4},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()
			m := newTestMonitor()
			m.handleKeyPress(tt.keyMsg)
			activePanel := m.focusState.ActivePanel()
			idx := activePanel.PanelIndex()
			if idx != tt.wantIdx {
				t.Errorf("after %s: panel index = %d, want %d", tt.key, idx, tt.wantIdx)
			}
		})
	}
}

func TestMonitor_View_Standard(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestMonitor_View_Narrow(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(60, 30)

	view := m.View()

	if view == "" {
		t.Error("View() with narrow width should not return empty string")
	}
}

func TestMonitor_HasActiveModal(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()

	if m.HasActiveModal() {
		t.Error("expected no active modal initially")
	}
}

func TestMonitor_RenderModal(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()

	if m.RenderModal() != "" {
		t.Error("expected empty modal render initially")
	}
}

func TestMonitor_SyncDataToComponents(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	// Initially no devices - syncDataToComponents should be safe to call
	m.syncDataToComponents()

	// Verify summary data is zero-valued
	data := m.summary.Data()
	if data.TotalPower != 0 {
		t.Errorf("expected 0 total power, got %f", data.TotalPower)
	}
	if data.OnlineCount != 0 {
		t.Errorf("expected 0 online count, got %d", data.OnlineCount)
	}
}

func TestMonitor_UpdateFocusStates(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	// Focus on power ranking
	m.focusState.JumpToPanel(1) // PowerRanking is index 1
	m.updateFocusStates()

	if !m.powerRanking.IsFocused() {
		t.Error("power ranking should be focused after JumpToPanel(1)")
	}

	// Focus on another panel
	m.focusState.JumpToPanel(2) // Environment is index 2
	m.updateFocusStates()

	if m.powerRanking.IsFocused() {
		t.Error("power ranking should not be focused when another panel is active")
	}
}

func TestMonitor_IsNarrow(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()

	m.width = 79
	if !m.isNarrow() {
		t.Error("width 79 should be narrow")
	}

	m.width = 80
	if m.isNarrow() {
		t.Error("width 80 should not be narrow")
	}

	m.width = 120
	if m.isNarrow() {
		t.Error("width 120 should not be narrow")
	}
}

func TestHandleMonitorExportResult(t *testing.T) {
	t.Parallel()

	t.Run("success CSV", func(t *testing.T) {
		t.Parallel()
		msg := monitor.ExportResultMsg{Path: "/tmp/export.csv", Format: monitor.ExportCSV}
		cmd := handleMonitorExportResult(msg)
		if cmd == nil {
			t.Error("expected non-nil command for successful export")
		}
	})

	t.Run("success JSON", func(t *testing.T) {
		t.Parallel()
		msg := monitor.ExportResultMsg{Path: "/tmp/export.json", Format: monitor.ExportJSON}
		cmd := handleMonitorExportResult(msg)
		if cmd == nil {
			t.Error("expected non-nil command for successful export")
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		msg := monitor.ExportResultMsg{Err: errors.New("disk full")}
		cmd := handleMonitorExportResult(msg)
		if cmd == nil {
			t.Error("expected non-nil command for export error")
		}
	})
}

func TestMonitor_SelectedDevice_Empty(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	// No devices loaded - should return nil
	if m.SelectedDevice() != nil {
		t.Error("expected nil selected device when no devices loaded")
	}
}

func TestMonitor_SelectedDevice_WithDevices(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	// Manually set devices on the power ranking (simulating data sync)
	statuses := []monitor.DeviceStatus{
		{Name: "kitchen", Address: "192.168.1.10", Online: true, Power: 342},
		{Name: "office", Address: "192.168.1.11", Online: true, Power: 180},
	}
	m.powerRanking = m.powerRanking.SetDevices(statuses)

	// Power ranking should be populated but SelectedDevice needs matching data source
	// Since data source is empty, SelectedDevice returns nil (no match)
	sel := m.SelectedDevice()
	if sel != nil {
		t.Error("expected nil when data source has no matching devices")
	}
}

func TestMonitor_EnvironmentFocus(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	// Focus on environment panel
	m.focusState.JumpToPanel(2) // Environment is index 2
	m.updateFocusStates()

	if !m.environment.IsFocused() {
		t.Error("environment should be focused after JumpToPanel(2)")
	}
	if m.powerRanking.IsFocused() {
		t.Error("power ranking should not be focused when environment is")
	}
}

func TestMonitor_EnvironmentView(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	view := m.View()

	// Environment panel should be rendered (not just pending stub)
	if !strings.Contains(view, "Environment") {
		t.Error("expected 'Environment' in view")
	}
	// Safety section should be visible even with no data
	if !strings.Contains(view, "Safety") {
		t.Error("expected 'Safety' section in view")
	}
}

func TestMonitor_EnvironmentNarrowView(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(60, 30)

	// Focus on environment in narrow mode
	m.focusState.JumpToPanel(2)
	m.updateFocusStates()

	view := m.View()
	if !strings.Contains(view, "Environment") {
		t.Error("narrow mode should show Environment panel when focused")
	}
}

// newTestMonitor creates a test monitor view.
func newTestMonitor() *Monitor {
	ctx := context.Background()
	svc := &shelly.Service{}
	ios := iostreams.Test(nil, &bytes.Buffer{}, &bytes.Buffer{})
	es := automation.NewEventStream(svc)
	focusState := focus.NewState()
	// Set to monitor tab so panel cycling works correctly
	focusState.SetActiveTab(tabs.TabMonitor)
	deps := MonitorDeps{Ctx: ctx, Svc: svc, IOS: ios, EventStream: es, FocusState: focusState}
	return NewMonitor(deps)
}
