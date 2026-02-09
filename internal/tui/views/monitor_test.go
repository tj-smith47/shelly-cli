package views

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/alerts"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/monitor"
	"github.com/tj-smith47/shelly-cli/internal/tui/focus"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
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

func TestMonitor_AlertsPanel_View(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	view := m.View()

	// Alerts panel should be rendered (not the old "Coming in next session" stub)
	if strings.Contains(view, "Coming in next session") {
		t.Error("expected alerts panel to replace stub")
	}
	if !strings.Contains(view, "Alerts") {
		t.Error("expected 'Alerts' title in view")
	}
}

func TestMonitor_AlertsPanel_Focus(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	// Focus on alerts panel (index 3)
	m.focusState.JumpToPanel(3)
	m.updateFocusStates()

	activePanel := m.focusState.ActivePanel()
	if activePanel != focus.PanelMonitorAlerts {
		t.Errorf("expected PanelMonitorAlerts, got %v", activePanel)
	}
}

func TestMonitor_AlertsPanel_NarrowView(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(60, 30)

	// Focus on alerts in narrow mode
	m.focusState.JumpToPanel(3)
	m.updateFocusStates()

	view := m.View()
	if !strings.Contains(view, "Alerts") {
		t.Error("narrow mode should show Alerts panel when focused")
	}
}

func TestMonitor_EventFeedPanel_View(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	view := m.View()

	if !strings.Contains(view, "Event Feed") {
		t.Error("expected 'Event Feed' title in view")
	}
}

func TestMonitor_EventFeedPanel_Focus(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	// Focus on event feed panel (index 4)
	m.focusState.JumpToPanel(4)
	m.updateFocusStates()

	activePanel := m.focusState.ActivePanel()
	if activePanel != focus.PanelMonitorEventFeed {
		t.Errorf("expected PanelMonitorEventFeed, got %v", activePanel)
	}
}

func TestMonitor_EventFeedPanel_NarrowView(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(60, 30)

	// Focus on event feed in narrow mode
	m.focusState.JumpToPanel(4)
	m.updateFocusStates()

	view := m.View()
	if !strings.Contains(view, "Event Feed") {
		t.Error("narrow mode should show Event Feed panel when focused")
	}
}

func TestMonitor_HasActiveModal_WithAlertForm(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	if m.HasActiveModal() {
		t.Error("expected no active modal initially")
	}

	// Simulate opening alert form
	m.alertFormOpen = true
	if !m.HasActiveModal() {
		t.Error("expected active modal when alert form is open")
	}
}

func TestMonitor_RenderModal_WithAlertForm(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	if m.RenderModal() != "" {
		t.Error("expected empty modal render initially")
	}

	// Simulate opening alert create form
	m.alertFormOpen = true
	m.alertForm = alerts.NewAlertForm(alerts.FormModeCreate, nil)
	m.alertForm = m.alertForm.SetSize(80, 30)

	modal := m.RenderModal()
	if modal == "" {
		t.Error("expected non-empty modal render when alert form is open")
	}
	if !strings.Contains(modal, "Create Alert") {
		t.Error("expected 'Create Alert' in modal render")
	}
}

func TestMonitor_HandleAlertMessages_Create(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	// Handle AlertCreateMsg should open form
	m.handleAlertMessages(alerts.AlertCreateMsg{})

	if !m.alertFormOpen {
		t.Error("expected alert form to be open after AlertCreateMsg")
	}
}

func TestMonitor_HandleAlertMessages_Delete(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	// DeleteMsg should return a command (the delete operation)
	cmd := m.handleAlertMessages(alerts.AlertDeleteMsg{Name: "test-alert"})
	if cmd == nil {
		t.Error("expected non-nil command for AlertDeleteMsg")
	}
}

func TestMonitor_HandleAlertFormMessages_Cancel(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	m.alertFormOpen = true
	m.handleAlertFormMessages(alerts.AlertFormCancelMsg{})

	if m.alertFormOpen {
		t.Error("expected alert form to be closed after cancel")
	}
}

func TestHandleMonitorAlertActionResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		msg     alerts.AlertActionResultMsg
		wantNil bool
	}{
		{
			name:    "toggle success",
			msg:     alerts.AlertActionResultMsg{Action: actionToggle, Name: "test"},
			wantNil: false,
		},
		{
			name:    "delete success",
			msg:     alerts.AlertActionResultMsg{Action: actionDelete, Name: "test"},
			wantNil: false,
		},
		{
			name:    "snooze success",
			msg:     alerts.AlertActionResultMsg{Action: actionSnooze, Name: "test"},
			wantNil: false,
		},
		{
			name:    "save success",
			msg:     alerts.AlertActionResultMsg{Action: actionSave, Name: "test"},
			wantNil: false,
		},
		{
			name:    "error",
			msg:     alerts.AlertActionResultMsg{Action: actionToggle, Name: "test", Err: errors.New("fail")},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := handleMonitorAlertActionResult(tt.msg)
			if (cmd == nil) != tt.wantNil {
				t.Errorf("handleMonitorAlertActionResult() returned nil=%v, want nil=%v", cmd == nil, tt.wantNil)
			}
		})
	}
}

func TestHandleMonitorAlertTestResult(t *testing.T) {
	t.Parallel()

	t.Run("triggered", func(t *testing.T) {
		t.Parallel()
		cmd := handleMonitorAlertTestResult(alerts.AlertTestResultMsg{Name: "test", Triggered: true, Value: "500W"})
		if cmd == nil {
			t.Error("expected non-nil command for triggered test")
		}
	})

	t.Run("ok", func(t *testing.T) {
		t.Parallel()
		cmd := handleMonitorAlertTestResult(alerts.AlertTestResultMsg{Name: "test", Triggered: false, Value: "50W"})
		if cmd == nil {
			t.Error("expected non-nil command for ok test")
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		cmd := handleMonitorAlertTestResult(alerts.AlertTestResultMsg{Name: "test", Err: errors.New("fail")})
		if cmd == nil {
			t.Error("expected non-nil command for test error")
		}
	})
}

func TestMonitor_TriggeredAlertCount(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()

	// Just verify it doesn't panic and returns a count
	count := m.TriggeredAlertCount()
	if count < 0 {
		t.Errorf("TriggeredAlertCount() = %d, expected >= 0", count)
	}
}

func TestMonitor_UpdateFocusedComponent_Alerts(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	// Focus alerts panel
	m.focusState.JumpToPanel(3)
	m.updateFocusStates()

	// Sending a message to the focused component should not panic
	cmd := m.updateFocusedComponent(tea.KeyPressMsg{Code: 'j'})
	_ = cmd
}

func TestMonitor_UpdateFocusedComponent_EventFeed(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	// Focus event feed panel
	m.focusState.JumpToPanel(4)
	m.updateFocusStates()

	// Sending a message to the focused component should not panic
	cmd := m.updateFocusedComponent(tea.KeyPressMsg{Code: 'j'})
	_ = cmd
}

// --- Energy History Overlay Tests ---

func TestMonitor_EnergyHistoryRequestMsg_OpensOverlay(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	msg := EnergyHistoryRequestMsg{DeviceName: "kitchen", Address: "192.168.1.10", Type: "SPEM-003CEBEU"}

	cmd := m.handleOverlayMessages(msg)

	if !m.energyHistoryOpen {
		t.Error("expected energyHistoryOpen to be true after request msg")
	}
	if m.energyHistory == nil {
		t.Fatal("expected energyHistory overlay to be non-nil")
	}
	if m.energyHistory.deviceName != "kitchen" {
		t.Errorf("deviceName = %q, want %q", m.energyHistory.deviceName, "kitchen")
	}
	if !m.energyHistory.loading {
		t.Error("expected overlay to be in loading state")
	}
	if cmd == nil {
		t.Error("expected non-nil command for data fetch")
	}
}

func TestMonitor_EnergyHistoryDataMsg_PopulatesOverlay(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	// First open the overlay
	m.energyHistory = &energyHistoryOverlay{overlayBase: overlayBase{deviceName: "kitchen", loading: true}}
	m.energyHistoryOpen = true

	// Then deliver data
	dataMsg := energyHistoryDataMsg{
		DeviceName: "kitchen",
		Energy:     1.234,
		AvgPower:   150.5,
		PeakPower:  340.2,
		DataPoints: 288,
		PowerData:  []float64{100, 200, 150, 300, 250},
	}

	m.handleOverlayDataMessages(dataMsg)

	if m.energyHistory.loading {
		t.Error("expected loading to be false after data msg")
	}
	if m.energyHistory.energy != 1.234 {
		t.Errorf("energy = %f, want 1.234", m.energyHistory.energy)
	}
	if m.energyHistory.avgPower != 150.5 {
		t.Errorf("avgPower = %f, want 150.5", m.energyHistory.avgPower)
	}
	if m.energyHistory.peakPower != 340.2 {
		t.Errorf("peakPower = %f, want 340.2", m.energyHistory.peakPower)
	}
	if m.energyHistory.dataPoints != 288 {
		t.Errorf("dataPoints = %d, want 288", m.energyHistory.dataPoints)
	}
	if len(m.energyHistory.powerData) != 5 {
		t.Errorf("powerData len = %d, want 5", len(m.energyHistory.powerData))
	}
	if m.energyHistory.err != nil {
		t.Errorf("unexpected error: %v", m.energyHistory.err)
	}
}

func TestMonitor_EnergyHistoryDataMsg_Error(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	m.energyHistory = &energyHistoryOverlay{overlayBase: overlayBase{deviceName: "kitchen", loading: true}}
	m.energyHistoryOpen = true

	testErr := fmt.Errorf("no historical data — only EM/EM1 devices store history")
	dataMsg := energyHistoryDataMsg{
		DeviceName: "kitchen",
		Err:        testErr,
	}

	m.handleOverlayDataMessages(dataMsg)

	if m.energyHistory.loading {
		t.Error("expected loading to be false after error msg")
	}
	if m.energyHistory.err == nil {
		t.Error("expected error to be set")
	}
}

func TestMonitor_EnergyHistoryDataMsg_WrongDevice(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	m.energyHistory = &energyHistoryOverlay{overlayBase: overlayBase{deviceName: "kitchen", loading: true}}
	m.energyHistoryOpen = true

	// Send data for a different device
	dataMsg := energyHistoryDataMsg{
		DeviceName: "office",
		Energy:     5.0,
		AvgPower:   200,
	}

	m.handleOverlayDataMessages(dataMsg)

	// Should NOT update since device name doesn't match
	if !m.energyHistory.loading {
		t.Error("should still be loading — data was for different device")
	}
	if m.energyHistory.energy != 0 {
		t.Errorf("energy should be 0, got %f", m.energyHistory.energy)
	}
}

func TestMonitor_RenderEnergyHistoryOverlay_Loading(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	m.energyHistory = &energyHistoryOverlay{overlayBase: overlayBase{deviceName: "kitchen", loading: true}}
	m.energyHistoryOpen = true

	rendered := m.renderEnergyHistoryOverlay()

	if !strings.Contains(rendered, "Energy History") {
		t.Error("expected 'Energy History' title")
	}
	if !strings.Contains(rendered, "Loading") {
		t.Error("expected loading indicator")
	}
}

func TestMonitor_RenderEnergyHistoryOverlay_WithData(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	m.energyHistory = &energyHistoryOverlay{
		overlayBase: overlayBase{deviceName: "kitchen"},
		energy:      1.234,
		avgPower:    150.5,
		peakPower:   340.2,
		dataPoints:  288,
		powerData:   []float64{100, 200, 150, 300, 250, 180, 220},
	}
	m.energyHistoryOpen = true

	rendered := m.renderEnergyHistoryOverlay()

	if !strings.Contains(rendered, "Energy History") {
		t.Error("expected 'Energy History' title")
	}
	if !strings.Contains(rendered, "kitchen") {
		t.Error("expected device name in overlay")
	}
	if !strings.Contains(rendered, "150.5 W") {
		t.Error("expected avg power in overlay")
	}
	if !strings.Contains(rendered, "340.2 W") {
		t.Error("expected peak power in overlay")
	}
	if !strings.Contains(rendered, "1.234 kWh") {
		t.Error("expected energy in overlay")
	}
	if !strings.Contains(rendered, "288") {
		t.Error("expected data points count in overlay")
	}
}

func TestMonitor_RenderEnergyHistoryOverlay_WithError(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	m.energyHistory = &energyHistoryOverlay{
		overlayBase: overlayBase{deviceName: "kitchen", err: fmt.Errorf("no historical data")},
	}
	m.energyHistoryOpen = true

	rendered := m.renderEnergyHistoryOverlay()

	if !strings.Contains(rendered, "no historical data") {
		t.Error("expected error message in overlay")
	}
}

func TestMonitor_RenderEnergyHistoryOverlay_Nil(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	m.energyHistory = nil
	m.energyHistoryOpen = true

	rendered := m.renderEnergyHistoryOverlay()

	if !strings.Contains(rendered, "No data") {
		t.Error("expected 'No data' when overlay is nil")
	}
}

func TestMonitor_HasActiveModal_EnergyHistory(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()

	m.energyHistoryOpen = true
	if !m.HasActiveModal() {
		t.Error("expected active modal when energy history is open")
	}
}

func TestMonitor_RenderModal_EnergyHistory(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	m.energyHistory = &energyHistoryOverlay{overlayBase: overlayBase{deviceName: "test", loading: true}}
	m.energyHistoryOpen = true

	modal := m.RenderModal()
	if modal == "" {
		t.Error("expected non-empty modal for energy history")
	}
	if !strings.Contains(modal, "Energy History") {
		t.Error("expected 'Energy History' in modal")
	}
}

// --- 3-Phase Detail Overlay Tests ---

func TestMonitor_PhaseDetailRequestMsg_OpensOverlay(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	msg := PhaseDetailRequestMsg{DeviceName: "meter", Address: "192.168.1.20"}

	cmd := m.handleOverlayMessages(msg)

	if !m.phaseDetailOpen {
		t.Error("expected phaseDetailOpen to be true after request msg")
	}
	if m.phaseDetail == nil {
		t.Fatal("expected phaseDetail overlay to be non-nil")
	}
	if m.phaseDetail.deviceName != "meter" {
		t.Errorf("deviceName = %q, want %q", m.phaseDetail.deviceName, "meter")
	}
	if !m.phaseDetail.loading {
		t.Error("expected overlay to be in loading state")
	}
	if cmd == nil {
		t.Error("expected non-nil command for data fetch")
	}
}

func TestMonitor_PhaseDetailDataMsg_PopulatesOverlay(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	m.phaseDetail = &phaseDetailOverlay{overlayBase: overlayBase{deviceName: "meter", loading: true}}
	m.phaseDetailOpen = true

	pf := 0.95
	freq := 50.0
	nCurrent := 0.5
	em := &model.EMStatus{
		AVoltage:         230.1,
		ACurrent:         1.5,
		AActivePower:     345.15,
		AApparentPower:   363.3,
		APowerFactor:     &pf,
		AFreq:            &freq,
		BVoltage:         231.2,
		BCurrent:         2.1,
		BActivePower:     485.5,
		BApparentPower:   510.0,
		BPowerFactor:     &pf,
		BFreq:            &freq,
		CVoltage:         229.8,
		CCurrent:         0.8,
		CActivePower:     183.8,
		CApparentPower:   193.5,
		CPowerFactor:     &pf,
		CFreq:            &freq,
		NCurrent:         &nCurrent,
		TotalCurrent:     4.4,
		TotalActivePower: 1014.45,
		TotalAprtPower:   1066.8,
	}

	dataMsg := phaseDetailDataMsg{
		DeviceName: "meter",
		EM:         em,
	}

	m.handleOverlayDataMessages(dataMsg)

	if m.phaseDetail.loading {
		t.Error("expected loading to be false after data msg")
	}
	if m.phaseDetail.em == nil {
		t.Fatal("expected EM data to be set")
	}
	if m.phaseDetail.em.AVoltage != 230.1 {
		t.Errorf("AVoltage = %f, want 230.1", m.phaseDetail.em.AVoltage)
	}
	if m.phaseDetail.em.TotalActivePower != 1014.45 {
		t.Errorf("TotalActivePower = %f, want 1014.45", m.phaseDetail.em.TotalActivePower)
	}
	if m.phaseDetail.err != nil {
		t.Errorf("unexpected error: %v", m.phaseDetail.err)
	}
}

func TestMonitor_PhaseDetailDataMsg_Error(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	m.phaseDetail = &phaseDetailOverlay{overlayBase: overlayBase{deviceName: "meter", loading: true}}
	m.phaseDetailOpen = true

	dataMsg := phaseDetailDataMsg{
		DeviceName: "meter",
		Err:        fmt.Errorf("device is not a 3-phase energy meter"),
	}

	m.handleOverlayDataMessages(dataMsg)

	if m.phaseDetail.loading {
		t.Error("expected loading to be false after error msg")
	}
	if m.phaseDetail.err == nil {
		t.Error("expected error to be set")
	}
}

func TestMonitor_PhaseDetailDataMsg_WrongDevice(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	m.phaseDetail = &phaseDetailOverlay{overlayBase: overlayBase{deviceName: "meter", loading: true}}
	m.phaseDetailOpen = true

	dataMsg := phaseDetailDataMsg{
		DeviceName: "other",
		EM:         &model.EMStatus{AVoltage: 230},
	}

	m.handleOverlayDataMessages(dataMsg)

	if !m.phaseDetail.loading {
		t.Error("should still be loading — data was for different device")
	}
	if m.phaseDetail.em != nil {
		t.Error("EM data should not be set for wrong device")
	}
}

func TestMonitor_RenderPhaseDetailOverlay_Loading(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	m.phaseDetail = &phaseDetailOverlay{overlayBase: overlayBase{deviceName: "meter", loading: true}}
	m.phaseDetailOpen = true

	rendered := m.renderPhaseDetailOverlay()

	if !strings.Contains(rendered, "3-Phase Detail") {
		t.Error("expected '3-Phase Detail' title")
	}
	if !strings.Contains(rendered, "Loading") {
		t.Error("expected loading indicator")
	}
}

func TestMonitor_RenderPhaseDetailOverlay_WithData(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	pf := 0.95
	freq := 50.0
	nCurrent := 0.5
	m.phaseDetail = &phaseDetailOverlay{
		overlayBase: overlayBase{deviceName: "meter"},
		em: &model.EMStatus{
			AVoltage:         230.1,
			ACurrent:         1.5,
			AActivePower:     345.15,
			AApparentPower:   363.3,
			APowerFactor:     &pf,
			AFreq:            &freq,
			BVoltage:         231.2,
			BCurrent:         2.1,
			BActivePower:     485.5,
			BApparentPower:   510.0,
			BPowerFactor:     &pf,
			BFreq:            &freq,
			CVoltage:         229.8,
			CCurrent:         0.8,
			CActivePower:     183.8,
			CApparentPower:   193.5,
			CPowerFactor:     &pf,
			CFreq:            &freq,
			NCurrent:         &nCurrent,
			TotalCurrent:     4.4,
			TotalActivePower: 1014.45,
			TotalAprtPower:   1066.8,
		},
	}
	m.phaseDetailOpen = true

	rendered := m.renderPhaseDetailOverlay()

	if !strings.Contains(rendered, "3-Phase Detail") {
		t.Error("expected '3-Phase Detail' title")
	}
	if !strings.Contains(rendered, "meter") {
		t.Error("expected device name")
	}
	if !strings.Contains(rendered, "Phase A") {
		t.Error("expected 'Phase A' header")
	}
	if !strings.Contains(rendered, "Phase B") {
		t.Error("expected 'Phase B' header")
	}
	if !strings.Contains(rendered, "Phase C") {
		t.Error("expected 'Phase C' header")
	}
	if !strings.Contains(rendered, "230.1 V") {
		t.Error("expected Phase A voltage")
	}
	if !strings.Contains(rendered, "Neutral Current") {
		t.Error("expected neutral current")
	}
	if !strings.Contains(rendered, "0.500 A") {
		t.Error("expected neutral current value")
	}
	if !strings.Contains(rendered, "1.01 kW") {
		t.Error("expected total active power in kW")
	}
}

func TestMonitor_RenderPhaseDetailOverlay_WithError(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	m.phaseDetail = &phaseDetailOverlay{
		overlayBase: overlayBase{deviceName: "meter", err: fmt.Errorf("not a 3-phase meter")},
	}
	m.phaseDetailOpen = true

	rendered := m.renderPhaseDetailOverlay()

	if !strings.Contains(rendered, "not a 3-phase meter") {
		t.Error("expected error message in overlay")
	}
}

func TestMonitor_RenderPhaseDetailOverlay_Nil(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	m.phaseDetail = nil
	m.phaseDetailOpen = true

	rendered := m.renderPhaseDetailOverlay()

	if !strings.Contains(rendered, "No data") {
		t.Error("expected 'No data' when overlay is nil")
	}
}

func TestMonitor_HasActiveModal_PhaseDetail(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()

	m.phaseDetailOpen = true
	if !m.HasActiveModal() {
		t.Error("expected active modal when phase detail is open")
	}
}

func TestMonitor_RenderModal_PhaseDetail(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	m.phaseDetail = &phaseDetailOverlay{overlayBase: overlayBase{deviceName: "test", loading: true}}
	m.phaseDetailOpen = true

	modal := m.RenderModal()
	if modal == "" {
		t.Error("expected non-empty modal for phase detail")
	}
	if !strings.Contains(modal, "3-Phase Detail") {
		t.Error("expected '3-Phase Detail' in modal")
	}
}

// --- Overlay Dismiss Tests ---

func TestMonitor_UpdateOverlay_EscClosesEnergyHistory(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	m.energyHistory = &energyHistoryOverlay{overlayBase: overlayBase{deviceName: "kitchen", loading: false}}
	m.energyHistoryOpen = true

	m.updateOverlay(tea.KeyPressMsg{Code: 0, Text: keyconst.KeyEsc})

	if m.energyHistoryOpen {
		t.Error("expected energy history to be closed after Esc")
	}
}

func TestMonitor_UpdateOverlay_QClosesPhaseDetail(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	m.phaseDetail = &phaseDetailOverlay{overlayBase: overlayBase{deviceName: "meter", loading: false}}
	m.phaseDetailOpen = true

	m.updateOverlay(tea.KeyPressMsg{Code: 'q'})

	if m.phaseDetailOpen {
		t.Error("expected phase detail to be closed after q")
	}
}

func TestMonitor_UpdateOverlay_ProcessesDataWhileOpen(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	m.energyHistory = &energyHistoryOverlay{overlayBase: overlayBase{deviceName: "kitchen", loading: true}}
	m.energyHistoryOpen = true

	// Send data msg while overlay is open (non-key msg)
	dataMsg := energyHistoryDataMsg{
		DeviceName: "kitchen",
		Energy:     2.5,
		AvgPower:   200,
	}

	m.updateOverlay(dataMsg)

	// Data should be received
	if m.energyHistory.loading {
		t.Error("expected data to be processed while overlay is open")
	}
	if m.energyHistory.energy != 2.5 {
		t.Errorf("energy = %f, want 2.5", m.energyHistory.energy)
	}
}

// --- Panel-Specific Keys Tests ---

func TestMonitor_HandlePanelSpecificKeys_AlertsFocused(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  rune
	}{
		{"delete", 'd'},
		{"test", 't'},
		{"snooze", 's'},
		{"snooze 24h", 'S'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := newTestMonitor()
			m.SetSize(120, 40)

			// Focus alerts panel
			m.focusState.JumpToPanel(3)
			m.updateFocusStates()

			cmd := m.handlePanelSpecificKeys(tea.KeyPressMsg{Code: tt.key})
			// Commands are generated for all alert panel keys when focused
			_ = cmd // Not nil check since updateFocusedComponent may or may not produce cmd
		})
	}
}

func TestMonitor_HandlePanelSpecificKeys_NotAlerts(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	// Focus power ranking (not alerts)
	m.focusState.JumpToPanel(1)
	m.updateFocusStates()

	cmd := m.handlePanelSpecificKeys(tea.KeyPressMsg{Code: 'd'})
	if cmd != nil {
		t.Error("expected nil cmd when not on alerts panel")
	}
}

// --- Helper Function Tests ---

func TestExtractPowerSeries_EM(t *testing.T) {
	t.Parallel()

	blocks := []components.EMDataBlock{
		{
			Values: []components.EMDataValues{
				{TotalActivePower: 100.5},
				{TotalActivePower: 200.3},
			},
		},
		{
			Values: []components.EMDataValues{
				{TotalActivePower: 150.0},
			},
		},
	}

	powers := extractPowerSeries(blocks,
		func(b components.EMDataBlock) []components.EMDataValues { return b.Values },
		func(v components.EMDataValues) float64 { return v.TotalActivePower },
	)

	if len(powers) != 3 {
		t.Fatalf("expected 3 power values, got %d", len(powers))
	}
	if powers[0] != 100.5 {
		t.Errorf("powers[0] = %f, want 100.5", powers[0])
	}
	if powers[1] != 200.3 {
		t.Errorf("powers[1] = %f, want 200.3", powers[1])
	}
	if powers[2] != 150.0 {
		t.Errorf("powers[2] = %f, want 150.0", powers[2])
	}
}

func TestExtractPowerSeries_Empty(t *testing.T) {
	t.Parallel()

	powers := extractPowerSeries([]components.EMDataBlock(nil),
		func(b components.EMDataBlock) []components.EMDataValues { return b.Values },
		func(v components.EMDataValues) float64 { return v.TotalActivePower },
	)

	if len(powers) != 0 {
		t.Errorf("expected 0 powers, got %d", len(powers))
	}
}

func TestExtractPowerSeries_EM1(t *testing.T) {
	t.Parallel()

	blocks := []components.EM1DataBlock{
		{
			Values: []components.EM1DataValues{
				{ActivePower: 50.0},
				{ActivePower: 75.5},
			},
		},
	}

	powers := extractPowerSeries(blocks,
		func(b components.EM1DataBlock) []components.EM1DataValues { return b.Values },
		func(v components.EM1DataValues) float64 { return v.ActivePower },
	)

	if len(powers) != 2 {
		t.Fatalf("expected 2 power values, got %d", len(powers))
	}
	if powers[0] != 50.0 {
		t.Errorf("powers[0] = %f, want 50.0", powers[0])
	}
	if powers[1] != 75.5 {
		t.Errorf("powers[1] = %f, want 75.5", powers[1])
	}
}

func TestScaleFloatData(t *testing.T) {
	t.Parallel()

	t.Run("same length", func(t *testing.T) {
		t.Parallel()
		data := []float64{1, 2, 3, 4, 5}
		result := scaleFloatData(data, 5)
		if len(result) != 5 {
			t.Errorf("expected len 5, got %d", len(result))
		}
		for i, v := range data {
			if result[i] != v {
				t.Errorf("result[%d] = %f, want %f", i, result[i], v)
			}
		}
	})

	t.Run("compress", func(t *testing.T) {
		t.Parallel()
		data := []float64{10, 20, 30, 40}
		result := scaleFloatData(data, 2)
		if len(result) != 2 {
			t.Fatalf("expected len 2, got %d", len(result))
		}
		// First bucket averages [10,20]=15, second averages [30,40]=35
		if result[0] != 15.0 {
			t.Errorf("result[0] = %f, want 15.0", result[0])
		}
		if result[1] != 35.0 {
			t.Errorf("result[1] = %f, want 35.0", result[1])
		}
	})

	t.Run("stretch", func(t *testing.T) {
		t.Parallel()
		data := []float64{0, 100}
		result := scaleFloatData(data, 3)
		if len(result) != 3 {
			t.Fatalf("expected len 3, got %d", len(result))
		}
		// Interpolation: [0, 50, 100]
		if result[0] != 0 {
			t.Errorf("result[0] = %f, want 0", result[0])
		}
		if result[1] != 50.0 {
			t.Errorf("result[1] = %f, want 50.0", result[1])
		}
		if result[2] != 100.0 {
			t.Errorf("result[2] = %f, want 100.0", result[2])
		}
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		result := scaleFloatData(nil, 5)
		if len(result) != 0 {
			t.Errorf("expected empty result, got %d", len(result))
		}
	})

	t.Run("zero width", func(t *testing.T) {
		t.Parallel()
		result := scaleFloatData([]float64{1, 2, 3}, 0)
		if len(result) != 3 {
			t.Errorf("expected original data, got %d", len(result))
		}
	})
}

func TestFormatOverlayPower(t *testing.T) {
	t.Parallel()

	tests := []struct {
		watts float64
		want  string
	}{
		{0, "0.0 W"},
		{150.5, "150.5 W"},
		{999.9, "999.9 W"},
		{1000.0, "1.00 kW"},
		{1500.0, "1.50 kW"},
		{-1500.0, "-1.50 kW"},
		{-500.0, "-500.0 W"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%.1f", tt.watts), func(t *testing.T) {
			t.Parallel()
			got := formatOverlayPower(tt.watts)
			if got != tt.want {
				t.Errorf("formatOverlayPower(%f) = %q, want %q", tt.watts, got, tt.want)
			}
		})
	}
}

func TestFormatOptionalFloat(t *testing.T) {
	t.Parallel()

	t.Run("nil", func(t *testing.T) {
		t.Parallel()
		got := formatOptionalFloat(nil, "%.3f")
		if got != "—" {
			t.Errorf("formatOptionalFloat(nil) = %q, want %q", got, "—")
		}
	})

	t.Run("value", func(t *testing.T) {
		t.Parallel()
		v := 0.95
		got := formatOptionalFloat(&v, "%.3f")
		if got != "0.950" {
			t.Errorf("formatOptionalFloat(0.95) = %q, want %q", got, "0.950")
		}
	})

	t.Run("with unit", func(t *testing.T) {
		t.Parallel()
		v := 50.0
		got := formatOptionalFloat(&v, "%.1f Hz")
		if got != "50.0 Hz" {
			t.Errorf("formatOptionalFloat(50.0) = %q, want %q", got, "50.0 Hz")
		}
	})
}

func TestPadRight(t *testing.T) {
	t.Parallel()

	tests := []struct {
		s     string
		width int
		want  string
	}{
		{"abc", 6, "abc   "},
		{"abc", 3, "abc"},
		{"abc", 2, "abc"},
		{"", 4, "    "},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q/%d", tt.s, tt.width), func(t *testing.T) {
			t.Parallel()
			got := padRight(tt.s, tt.width)
			if got != tt.want {
				t.Errorf("padRight(%q, %d) = %q, want %q", tt.s, tt.width, got, tt.want)
			}
		})
	}
}

func TestRenderOverlaySparkline(t *testing.T) {
	t.Parallel()

	t.Run("basic data", func(t *testing.T) {
		t.Parallel()
		data := []float64{0, 50, 100, 150, 200, 250, 300, 350}
		result := renderOverlaySparkline(data, 8)
		if result == "" {
			t.Error("expected non-empty sparkline")
		}
	})

	t.Run("empty data", func(t *testing.T) {
		t.Parallel()
		result := renderOverlaySparkline(nil, 10)
		if result != "" {
			t.Errorf("expected empty string for nil data, got %q", result)
		}
	})

	t.Run("flat data", func(t *testing.T) {
		t.Parallel()
		data := []float64{100, 100, 100, 100}
		result := renderOverlaySparkline(data, 4)
		if result == "" {
			t.Error("expected non-empty sparkline for flat data")
		}
	})

	t.Run("flat near zero", func(t *testing.T) {
		t.Parallel()
		data := []float64{0.001, 0.001, 0.001}
		result := renderOverlaySparkline(data, 3)
		if result == "" {
			t.Error("expected non-empty sparkline for flat near-zero data")
		}
	})
}

// --- Update with Overlay Captures Input ---

func TestMonitor_Update_OverlayCapturesInput(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	m.energyHistory = &energyHistoryOverlay{overlayBase: overlayBase{deviceName: "kitchen", loading: false}}
	m.energyHistoryOpen = true

	// Regular keys should not pass through to components
	result, _ := m.Update(tea.KeyPressMsg{Code: 'j'})
	updated, ok := result.(*Monitor)
	if !ok {
		t.Fatal("expected *Monitor from Update")
	}

	// Overlay should still be open (j does NOT close it)
	if !updated.energyHistoryOpen {
		t.Error("overlay should still be open — j is not a dismiss key")
	}
}

func TestMonitor_Update_OverlayEscDismisses(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	m.energyHistory = &energyHistoryOverlay{overlayBase: overlayBase{deviceName: "kitchen", loading: false}}
	m.energyHistoryOpen = true

	result, _ := m.Update(tea.KeyPressMsg{Code: 0, Text: keyconst.KeyEsc})
	updated, ok := result.(*Monitor)
	if !ok {
		t.Fatal("expected *Monitor from Update")
	}

	if updated.energyHistoryOpen {
		t.Error("overlay should be closed after Esc")
	}
}

// --- Integration: handleTypedMessages dispatches overlays ---

func TestMonitor_HandleTypedMessages_OverlayRequest(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	cmds := m.handleTypedMessages(EnergyHistoryRequestMsg{DeviceName: "dev1", Address: "10.0.0.1"})

	if m.energyHistoryOpen != true {
		t.Error("expected overlay to open via handleTypedMessages")
	}
	// Should have at least one command (the fetch command)
	if len(cmds) == 0 {
		t.Error("expected at least one command from overlay request")
	}
}

func TestMonitor_HandleTypedMessages_PanelSpecificKey(t *testing.T) {
	t.Parallel()
	m := newTestMonitor()
	m.SetSize(120, 40)

	// Focus alerts panel
	m.focusState.JumpToPanel(3)
	m.updateFocusStates()

	// handleTypedMessages should process key events for panel-specific actions
	cmds := m.handleTypedMessages(tea.KeyPressMsg{Code: 'd'})
	// The command may or may not be nil depending on alerts state, but it should not panic
	_ = cmds
}

// --- Messages package types used by panel-specific keys ---

func TestMessages_IsActionRequest(t *testing.T) {
	t.Parallel()

	if !messages.IsActionRequest(messages.DeleteRequestMsg{}) {
		t.Error("DeleteRequestMsg should be an action request")
	}
	if !messages.IsActionRequest(messages.TestRequestMsg{}) {
		t.Error("TestRequestMsg should be an action request")
	}
	if !messages.IsActionRequest(messages.SnoozeRequestMsg{}) {
		t.Error("SnoozeRequestMsg should be an action request")
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
