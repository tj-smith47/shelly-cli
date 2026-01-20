package focus

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/tui/tabs"
)

func TestNewState(t *testing.T) {
	t.Parallel()
	s := NewState()
	if s == nil {
		t.Fatal("NewState() returned nil")
	}
	if s.ActiveTab() != tabs.TabDashboard {
		t.Errorf("ActiveTab() = %v, want %v", s.ActiveTab(), tabs.TabDashboard)
	}
	if s.ActivePanel() != PanelDeviceList {
		t.Errorf("ActivePanel() = %v, want %v", s.ActivePanel(), PanelDeviceList)
	}
	if s.ViewFocused() {
		t.Error("ViewFocused() should be false initially")
	}
}

func TestState_SetActiveTab(t *testing.T) {
	t.Parallel()
	s := NewState()

	// Change to automation tab
	changed := s.SetActiveTab(tabs.TabAutomation)
	if !changed {
		t.Error("SetActiveTab() should return true when tab changes")
	}
	if s.ActiveTab() != tabs.TabAutomation {
		t.Errorf("ActiveTab() = %v, want %v", s.ActiveTab(), tabs.TabAutomation)
	}
	// Panel should reset to first panel of new tab
	if s.ActivePanel() != PanelDeviceList {
		t.Errorf("ActivePanel() = %v, want %v (first panel of Automation)", s.ActivePanel(), PanelDeviceList)
	}

	// Setting same tab should return false
	changed = s.SetActiveTab(tabs.TabAutomation)
	if changed {
		t.Error("SetActiveTab() should return false when tab doesn't change")
	}
}

func TestState_SetActivePanel(t *testing.T) {
	t.Parallel()
	s := NewState()

	// Change panel
	changed := s.SetActivePanel(PanelDashboardInfo)
	if !changed {
		t.Error("SetActivePanel() should return true when panel changes")
	}
	if s.ActivePanel() != PanelDashboardInfo {
		t.Errorf("ActivePanel() = %v, want %v", s.ActivePanel(), PanelDashboardInfo)
	}
	// ViewFocused should be true when not on device list
	if !s.ViewFocused() {
		t.Error("ViewFocused() should be true when not on device list")
	}

	// Setting same panel should return false
	changed = s.SetActivePanel(PanelDashboardInfo)
	if changed {
		t.Error("SetActivePanel() should return false when panel doesn't change")
	}
}

func TestState_NextPanel(t *testing.T) {
	t.Parallel()
	s := NewState()

	// Dashboard panels: DeviceList -> Info -> Events -> EnergyBars -> EnergyHistory -> DeviceList
	s.NextPanel()
	if s.ActivePanel() != PanelDashboardInfo {
		t.Errorf("After first NextPanel(), ActivePanel() = %v, want %v", s.ActivePanel(), PanelDashboardInfo)
	}

	s.NextPanel()
	if s.ActivePanel() != PanelDashboardEvents {
		t.Errorf("After second NextPanel(), ActivePanel() = %v, want %v", s.ActivePanel(), PanelDashboardEvents)
	}

	s.NextPanel()
	if s.ActivePanel() != PanelDashboardEnergyBars {
		t.Errorf("After third NextPanel(), ActivePanel() = %v, want %v", s.ActivePanel(), PanelDashboardEnergyBars)
	}

	s.NextPanel()
	if s.ActivePanel() != PanelDashboardEnergyHistory {
		t.Errorf("After fourth NextPanel(), ActivePanel() = %v, want %v", s.ActivePanel(), PanelDashboardEnergyHistory)
	}

	// Wrap around
	s.NextPanel()
	if s.ActivePanel() != PanelDeviceList {
		t.Errorf("After wrap NextPanel(), ActivePanel() = %v, want %v", s.ActivePanel(), PanelDeviceList)
	}
}

func TestState_PrevPanel(t *testing.T) {
	t.Parallel()
	s := NewState()

	// First Prev: DeviceList -> EnergyHistory (wrap around)
	s.PrevPanel()
	if s.ActivePanel() != PanelDashboardEnergyHistory {
		t.Errorf("After first PrevPanel(), ActivePanel() = %v, want %v", s.ActivePanel(), PanelDashboardEnergyHistory)
	}

	s.PrevPanel()
	if s.ActivePanel() != PanelDashboardEnergyBars {
		t.Errorf("After second PrevPanel(), ActivePanel() = %v, want %v", s.ActivePanel(), PanelDashboardEnergyBars)
	}
}

func TestState_JumpToPanel(t *testing.T) {
	t.Parallel()
	s := NewState()

	// Jump to panel 3 (Events) in Dashboard
	ok := s.JumpToPanel(3)
	if !ok {
		t.Error("JumpToPanel(3) should return true")
	}
	if s.ActivePanel() != PanelDashboardEvents {
		t.Errorf("ActivePanel() = %v, want %v", s.ActivePanel(), PanelDashboardEvents)
	}

	// Invalid index should return false
	ok = s.JumpToPanel(99)
	if ok {
		t.Error("JumpToPanel(99) should return false")
	}

	// Panel should not change
	if s.ActivePanel() != PanelDashboardEvents {
		t.Errorf("ActivePanel() = %v, want %v (unchanged)", s.ActivePanel(), PanelDashboardEvents)
	}
}

func TestState_IsPanelFocused(t *testing.T) {
	t.Parallel()
	s := NewState()

	if !s.IsPanelFocused(PanelDeviceList) {
		t.Error("IsPanelFocused(PanelDeviceList) should be true")
	}
	if s.IsPanelFocused(PanelDashboardInfo) {
		t.Error("IsPanelFocused(PanelDashboardInfo) should be false")
	}

	// When overlay is present, no panel should be focused
	s.PushOverlay(OverlayHelp, ModeModal)
	if s.IsPanelFocused(PanelDeviceList) {
		t.Error("IsPanelFocused() should be false when overlay is present")
	}
}

func TestState_ReturnToDeviceList(t *testing.T) {
	t.Parallel()
	s := NewState()

	// Move to different panel
	s.SetActivePanel(PanelDashboardInfo)
	if s.ActivePanel() != PanelDashboardInfo {
		t.Errorf("ActivePanel() = %v, want %v", s.ActivePanel(), PanelDashboardInfo)
	}

	// Return to device list
	changed := s.ReturnToDeviceList()
	if !changed {
		t.Error("ReturnToDeviceList() should return true when panel changes")
	}
	if s.ActivePanel() != PanelDeviceList {
		t.Errorf("ActivePanel() = %v, want %v", s.ActivePanel(), PanelDeviceList)
	}

	// Already on device list, should return false
	changed = s.ReturnToDeviceList()
	if changed {
		t.Error("ReturnToDeviceList() should return false when already on device list")
	}
}

func TestState_Overlays(t *testing.T) {
	t.Parallel()
	s := NewState()

	if s.HasOverlay() {
		t.Error("HasOverlay() should be false initially")
	}

	// Push overlay
	s.PushOverlay(OverlayHelp, ModeModal)
	if !s.HasOverlay() {
		t.Error("HasOverlay() should be true after push")
	}
	if s.OverlayCount() != 1 {
		t.Errorf("OverlayCount() = %d, want 1", s.OverlayCount())
	}
	if s.TopOverlay() != OverlayHelp {
		t.Errorf("TopOverlay() = %v, want %v", s.TopOverlay(), OverlayHelp)
	}
	if s.Mode() != ModeModal {
		t.Errorf("Mode() = %v, want %v", s.Mode(), ModeModal)
	}

	// Push another
	s.PushOverlay(OverlayConfirm, ModeModal)
	if s.OverlayCount() != 2 {
		t.Errorf("OverlayCount() = %d, want 2", s.OverlayCount())
	}
	if s.TopOverlay() != OverlayConfirm {
		t.Errorf("TopOverlay() = %v, want %v", s.TopOverlay(), OverlayConfirm)
	}

	// Pop
	s.PopOverlay()
	if s.OverlayCount() != 1 {
		t.Errorf("OverlayCount() = %d, want 1", s.OverlayCount())
	}
	if s.TopOverlay() != OverlayHelp {
		t.Errorf("TopOverlay() = %v, want %v", s.TopOverlay(), OverlayHelp)
	}

	// Pop again
	s.PopOverlay()
	if s.HasOverlay() {
		t.Error("HasOverlay() should be false after all pops")
	}
	if s.Mode() != ModeNormal {
		t.Errorf("Mode() = %v, want %v", s.Mode(), ModeNormal)
	}
}

func TestState_ContainsOverlay(t *testing.T) {
	t.Parallel()
	s := NewState()

	s.PushOverlay(OverlayHelp, ModeModal)
	s.PushOverlay(OverlayConfirm, ModeModal)

	if !s.ContainsOverlay(OverlayHelp) {
		t.Error("ContainsOverlay(OverlayHelp) should be true")
	}
	if !s.ContainsOverlay(OverlayConfirm) {
		t.Error("ContainsOverlay(OverlayConfirm) should be true")
	}
	if s.ContainsOverlay(OverlayJSONViewer) {
		t.Error("ContainsOverlay(OverlayJSONViewer) should be false")
	}
}

func TestState_Clear(t *testing.T) {
	t.Parallel()
	s := NewState()

	s.PushOverlay(OverlayHelp, ModeModal)
	s.PushOverlay(OverlayConfirm, ModeModal)

	s.Clear()
	if s.HasOverlay() {
		t.Error("HasOverlay() should be false after Clear()")
	}
	if s.Mode() != ModeNormal {
		t.Errorf("Mode() = %v, want %v", s.Mode(), ModeNormal)
	}
}

func TestGlobalPanelID_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		panel GlobalPanelID
		want  string
	}{
		{PanelNone, "none"},
		{PanelDeviceList, "deviceList"},
		{PanelDashboardInfo, "dashboard.info"},
		{PanelDashboardEvents, "dashboard.events"},
		{PanelAutoScripts, "auto.scripts"},
		{PanelConfigWiFi, "config.wifi"},
		{PanelManageDiscovery, "manage.discovery"},
		{PanelFleetDevices, "fleet.devices"},
		{GlobalPanelID(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			if got := tt.panel.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGlobalPanelID_TabFor(t *testing.T) {
	t.Parallel()
	tests := []struct {
		panel GlobalPanelID
		want  tabs.TabID
	}{
		{PanelDeviceList, tabs.TabDashboard},
		{PanelDashboardInfo, tabs.TabDashboard},
		{PanelAutoScripts, tabs.TabAutomation},
		{PanelConfigWiFi, tabs.TabConfig},
		{PanelManageDiscovery, tabs.TabManage},
		{PanelMonitorMain, tabs.TabMonitor},
		{PanelFleetDevices, tabs.TabFleet},
	}

	for _, tt := range tests {
		t.Run(tt.panel.String(), func(t *testing.T) {
			t.Parallel()
			if got := tt.panel.TabFor(); got != tt.want {
				t.Errorf("TabFor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGlobalPanelID_PanelIndex(t *testing.T) {
	t.Parallel()
	tests := []struct {
		panel GlobalPanelID
		want  int
	}{
		{PanelDeviceList, 1},
		{PanelDashboardInfo, 2},
		{PanelDashboardEvents, 3},
		{PanelAutoScripts, 2},
		{PanelConfigWiFi, 2},
		{PanelManageDiscovery, 1},
		{PanelFleetDevices, 1},
		{PanelNone, 0},
	}

	for _, tt := range tests {
		t.Run(tt.panel.String(), func(t *testing.T) {
			t.Parallel()
			if got := tt.panel.PanelIndex(); got != tt.want {
				t.Errorf("PanelIndex() = %d, want %d", got, tt.want)
			}
		})
	}
}
