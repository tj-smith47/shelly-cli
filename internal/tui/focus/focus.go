// Package focus provides a focus management system for TUI panels.
// It tracks which panel has focus and provides methods for cycling focus.
package focus

import (
	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/tui/tabs"
)

const unknownStr = "unknown"

// Mode represents the current keyboard input routing mode.
type Mode int

const (
	// ModeNormal routes keys to the active panel.
	ModeNormal Mode = iota
	// ModeOverlay routes keys to a non-blocking overlay (help, JSON viewer).
	ModeOverlay
	// ModeModal routes keys to a blocking modal (edit forms).
	ModeModal
	// ModeInput routes keys to text input capture (search, command mode).
	ModeInput
)

// String returns the mode name.
func (m Mode) String() string {
	switch m {
	case ModeNormal:
		return "normal"
	case ModeOverlay:
		return "overlay"
	case ModeModal:
		return "modal"
	case ModeInput:
		return "input"
	default:
		return unknownStr
	}
}

// OverlayID identifies a specific overlay or modal.
type OverlayID int

const (
	// OverlayNone indicates no overlay is active.
	OverlayNone OverlayID = iota
	// OverlayHelp is the help overlay.
	OverlayHelp
	// OverlayJSONViewer is the JSON viewer overlay.
	OverlayJSONViewer
	// OverlayConfirm is the confirmation dialog.
	OverlayConfirm
	// OverlayDeviceDetail is the device detail overlay.
	OverlayDeviceDetail
	// OverlayControlPanel is the device control panel overlay.
	OverlayControlPanel
	// OverlayEditModal is a generic view edit modal.
	OverlayEditModal
	// OverlayCmdMode is command mode input.
	OverlayCmdMode
	// OverlaySearch is search input mode.
	OverlaySearch
	// OverlayScriptEditor is the script editor modal.
	OverlayScriptEditor
	// OverlayScriptCreate is the script creation modal.
	OverlayScriptCreate
	// OverlayScriptConsole is the script console modal.
	OverlayScriptConsole
	// OverlayScriptEval is the script eval modal.
	OverlayScriptEval
	// OverlayScheduleCreate is the schedule creation modal.
	OverlayScheduleCreate
	// OverlayAlertForm is the alert form modal.
	OverlayAlertForm
	// OverlayTemplateSelect is the template selection modal.
	OverlayTemplateSelect
	// OverlayProvisioning is the provisioning wizard.
	OverlayProvisioning
	// OverlayMigration is the migration wizard.
	OverlayMigration
)

// String returns the overlay name.
func (o OverlayID) String() string {
	names := map[OverlayID]string{
		OverlayNone:           "none",
		OverlayHelp:           "help",
		OverlayJSONViewer:     "json",
		OverlayConfirm:        "confirm",
		OverlayDeviceDetail:   "deviceDetail",
		OverlayControlPanel:   "controlPanel",
		OverlayEditModal:      "editModal",
		OverlayCmdMode:        "cmdMode",
		OverlaySearch:         "search",
		OverlayScriptEditor:   "scriptEditor",
		OverlayScriptCreate:   "scriptCreate",
		OverlayScriptConsole:  "scriptConsole",
		OverlayScriptEval:     "scriptEval",
		OverlayScheduleCreate: "scheduleCreate",
		OverlayAlertForm:      "alertForm",
		OverlayTemplateSelect: "templateSelect",
		OverlayProvisioning:   "provisioning",
		OverlayMigration:      "migration",
	}
	if name, ok := names[o]; ok {
		return name
	}
	return unknownStr
}

// OverlayEntry represents an entry in the overlay stack.
type OverlayEntry struct {
	ID        OverlayID
	FocusMode Mode
}

// State is the single source of truth for all focus management.
// All focus-related queries and mutations go through this struct.
type State struct {
	// Tab focus
	activeTab tabs.TabID

	// Panel focus within active tab
	activePanel GlobalPanelID

	// View focus mode: true = view content has focus, false = device list has focus
	// Only applicable for tabs that have a device list (Dashboard, Automation, Config)
	viewFocused bool

	// Overlay/modal stack (existing functionality)
	overlayStack []OverlayEntry
	mode         Mode

	// Panel cycle order per tab (initialized once)
	tabPanels map[tabs.TabID][]GlobalPanelID
}

// NewState creates a new unified focus state with default values.
func NewState() *State {
	s := &State{
		activeTab:    tabs.TabDashboard,
		activePanel:  PanelDeviceList,
		viewFocused:  false,
		overlayStack: make([]OverlayEntry, 0, 4),
		mode:         ModeNormal,
		tabPanels:    make(map[tabs.TabID][]GlobalPanelID),
	}
	s.initTabPanels()
	return s
}

// initTabPanels sets up the panel cycling order for each tab.
// This determines what Tab/Shift+Tab cycles through.
func (s *State) initTabPanels() {
	// Dashboard: DeviceList -> Info -> Events -> EnergyBars -> EnergyHistory
	s.tabPanels[tabs.TabDashboard] = []GlobalPanelID{
		PanelDeviceList,
		PanelDashboardInfo,
		PanelDashboardEvents,
		PanelDashboardEnergyBars,
		PanelDashboardEnergyHistory,
	}

	// Automation: DeviceList -> Scripts -> Schedules -> Webhooks -> Virtuals -> KVS -> Alerts
	s.tabPanels[tabs.TabAutomation] = []GlobalPanelID{
		PanelDeviceList,
		PanelAutoScripts,
		PanelAutoSchedules,
		PanelAutoWebhooks,
		PanelAutoVirtuals,
		PanelAutoKVS,
		PanelAutoAlerts,
	}

	// Config: DeviceList -> WiFi -> System -> Cloud -> Security -> BLE -> Inputs -> Protocols -> SmartHome
	s.tabPanels[tabs.TabConfig] = []GlobalPanelID{
		PanelDeviceList,
		PanelConfigWiFi,
		PanelConfigSystem,
		PanelConfigCloud,
		PanelConfigSecurity,
		PanelConfigBLE,
		PanelConfigInputs,
		PanelConfigProtocols,
		PanelConfigSmartHome,
	}

	// Manage: Discovery -> Firmware -> Backup -> Scenes -> Templates -> Batch
	// Note: No device list - manage view has its own panel layout
	s.tabPanels[tabs.TabManage] = []GlobalPanelID{
		PanelManageDiscovery,
		PanelManageFirmware,
		PanelManageBackup,
		PanelManageScenes,
		PanelManageTemplates,
		PanelManageBatch,
	}

	// Monitor: Just the main monitor panel
	s.tabPanels[tabs.TabMonitor] = []GlobalPanelID{
		PanelMonitorMain,
	}

	// Fleet: Devices -> Groups -> Health -> Operations
	s.tabPanels[tabs.TabFleet] = []GlobalPanelID{
		PanelFleetDevices,
		PanelFleetGroups,
		PanelFleetHealth,
		PanelFleetOperations,
	}
}

// Mode returns the current focus mode.
func (s *State) Mode() Mode {
	return s.mode
}

// ActiveTab returns the currently active tab.
func (s *State) ActiveTab() tabs.TabID {
	return s.activeTab
}

// SetActiveTab changes the active tab and resets panel focus.
// Returns true if tab actually changed.
func (s *State) SetActiveTab(tab tabs.TabID) bool {
	if s.activeTab == tab {
		return false
	}
	s.activeTab = tab
	// Reset to first panel in the new tab
	if panels, ok := s.tabPanels[tab]; ok && len(panels) > 0 {
		s.activePanel = panels[0]
	} else {
		s.activePanel = PanelNone
	}
	// Reset view focus when changing tabs
	s.viewFocused = false
	return true
}

// ActivePanel returns the currently focused panel.
func (s *State) ActivePanel() GlobalPanelID {
	return s.activePanel
}

// SetActivePanel changes the focused panel.
// Returns true if panel actually changed.
func (s *State) SetActivePanel(panel GlobalPanelID) bool {
	if s.activePanel == panel {
		return false
	}
	s.activePanel = panel
	// Update viewFocused based on whether device list is focused
	s.viewFocused = (panel != PanelDeviceList)
	return true
}

// NextPanel cycles to the next panel in the current tab.
// Returns the new panel (or same if at end and no wrap).
func (s *State) NextPanel() GlobalPanelID {
	panels := s.tabPanels[s.activeTab]
	if len(panels) == 0 {
		return s.activePanel
	}

	// Find current index
	idx := s.findPanelIndex(s.activePanel, panels)

	// Move to next (with wrap)
	nextIdx := (idx + 1) % len(panels)
	s.activePanel = panels[nextIdx]
	s.viewFocused = (s.activePanel != PanelDeviceList)
	return s.activePanel
}

// PrevPanel cycles to the previous panel in the current tab.
// Returns the new panel (or same if at start and no wrap).
func (s *State) PrevPanel() GlobalPanelID {
	panels := s.tabPanels[s.activeTab]
	if len(panels) == 0 {
		return s.activePanel
	}

	// Find current index
	idx := s.findPanelIndex(s.activePanel, panels)

	// Move to previous (with wrap)
	prevIdx := (idx - 1 + len(panels)) % len(panels)
	s.activePanel = panels[prevIdx]
	s.viewFocused = (s.activePanel != PanelDeviceList)
	return s.activePanel
}

// JumpToPanel sets focus to a specific panel by its 1-based index (Shift+N).
// Returns true if the jump was successful.
func (s *State) JumpToPanel(index int) bool {
	panels := s.tabPanels[s.activeTab]
	if index < 1 || index > len(panels) {
		return false
	}
	s.activePanel = panels[index-1]
	s.viewFocused = (s.activePanel != PanelDeviceList)
	return true
}

// ReturnToDeviceList sets focus back to the device list.
// Returns true if there was a change.
func (s *State) ReturnToDeviceList() bool {
	// Find device list in current tab's panels
	panels := s.tabPanels[s.activeTab]
	for _, p := range panels {
		if p == PanelDeviceList {
			if s.activePanel == PanelDeviceList {
				return false
			}
			s.activePanel = PanelDeviceList
			s.viewFocused = false
			return true
		}
	}
	// Tab doesn't have device list - just set first panel
	if len(panels) > 0 {
		s.activePanel = panels[0]
		s.viewFocused = false
		return true
	}
	return false
}

// findPanelIndex returns the index of panel in the slice, or 0 if not found.
func (s *State) findPanelIndex(panel GlobalPanelID, panels []GlobalPanelID) int {
	for i, p := range panels {
		if p == panel {
			return i
		}
	}
	return 0
}

// ViewFocused returns whether the view content has focus (vs device list).
func (s *State) ViewFocused() bool {
	return s.viewFocused
}

// SetViewFocused sets whether view content has focus.
// This is derived from activePanel but can be set explicitly for edge cases.
func (s *State) SetViewFocused(focused bool) {
	s.viewFocused = focused
}

// IsPanelFocused returns true if the given panel is currently focused.
// Components call this to determine their visual focus state.
func (s *State) IsPanelFocused(panel GlobalPanelID) bool {
	// If there's an overlay, no panel is focused
	if s.HasOverlay() {
		return false
	}
	return s.activePanel == panel
}

// IsPanelInActiveTab returns true if the panel belongs to the active tab.
func (s *State) IsPanelInActiveTab(panel GlobalPanelID) bool {
	return panel.TabFor() == s.activeTab
}

// CurrentPanelIndex returns the 1-based index of the current panel.
func (s *State) CurrentPanelIndex() int {
	return s.activePanel.PanelIndex()
}

// PanelCount returns the number of panels in the current tab.
func (s *State) PanelCount() int {
	return len(s.tabPanels[s.activeTab])
}

// GetTabPanels returns the panel list for a given tab.
func (s *State) GetTabPanels(tab tabs.TabID) []GlobalPanelID {
	return s.tabPanels[tab]
}

// PushOverlay adds an overlay to the stack and updates the mode.
func (s *State) PushOverlay(id OverlayID, mode Mode) {
	s.overlayStack = append(s.overlayStack, OverlayEntry{
		ID:        id,
		FocusMode: mode,
	})
	s.mode = mode
}

// PopOverlay removes the top overlay from the stack and returns its ID.
// Returns OverlayNone if the stack is empty.
func (s *State) PopOverlay() OverlayID {
	if len(s.overlayStack) == 0 {
		return OverlayNone
	}

	// Get the top entry
	lastIdx := len(s.overlayStack) - 1
	entry := s.overlayStack[lastIdx]

	// Remove from stack
	s.overlayStack = s.overlayStack[:lastIdx]

	// Update mode based on new top of stack (or normal if empty)
	if len(s.overlayStack) > 0 {
		s.mode = s.overlayStack[len(s.overlayStack)-1].FocusMode
	} else {
		s.mode = ModeNormal
	}

	return entry.ID
}

// TopOverlay returns the ID of the top overlay without removing it.
// Returns OverlayNone if the stack is empty.
func (s *State) TopOverlay() OverlayID {
	if len(s.overlayStack) == 0 {
		return OverlayNone
	}
	return s.overlayStack[len(s.overlayStack)-1].ID
}

// HasOverlay returns true if there are any overlays on the stack.
func (s *State) HasOverlay() bool {
	return len(s.overlayStack) > 0
}

// OverlayCount returns the number of overlays on the stack.
func (s *State) OverlayCount() int {
	return len(s.overlayStack)
}

// Clear removes all overlays and resets to normal mode.
func (s *State) Clear() {
	s.overlayStack = s.overlayStack[:0]
	s.mode = ModeNormal
}

// ContainsOverlay returns true if the specified overlay is anywhere in the stack.
func (s *State) ContainsOverlay(id OverlayID) bool {
	for _, entry := range s.overlayStack {
		if entry.ID == id {
			return true
		}
	}
	return false
}

// ChangedMsg is emitted when focus state changes.
// Components and views listen for this to update their visual state.
type ChangedMsg struct {
	// What changed
	TabChanged     bool
	PanelChanged   bool
	OverlayChanged bool

	// Current state
	ActiveTab   tabs.TabID
	ActivePanel GlobalPanelID
	ViewFocused bool
	Mode        Mode

	// Previous state
	PrevTab   tabs.TabID
	PrevPanel GlobalPanelID
}

// NewChangedMsg creates a ChangedMsg from the current state.
func (s *State) NewChangedMsg(prevTab tabs.TabID, prevPanel GlobalPanelID, tabChanged, panelChanged, overlayChanged bool) ChangedMsg {
	return ChangedMsg{
		TabChanged:     tabChanged,
		PanelChanged:   panelChanged,
		OverlayChanged: overlayChanged,
		ActiveTab:      s.activeTab,
		ActivePanel:    s.activePanel,
		ViewFocused:    s.viewFocused,
		Mode:           s.mode,
		PrevTab:        prevTab,
		PrevPanel:      prevPanel,
	}
}

// EmitChanged returns a tea.Cmd that emits a ChangedMsg.
func (s *State) EmitChanged(prevTab tabs.TabID, prevPanel GlobalPanelID, tabChanged, panelChanged, overlayChanged bool) tea.Cmd {
	msg := s.NewChangedMsg(prevTab, prevPanel, tabChanged, panelChanged, overlayChanged)
	return func() tea.Msg { return msg }
}
