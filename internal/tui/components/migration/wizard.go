// Package migration provides TUI components for device configuration migration.
package migration

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/generics"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// WizardStep represents a step in the migration wizard.
type WizardStep int

// Wizard step constants.
const (
	StepSourceSelect WizardStep = iota
	StepTargetSelect
	StepPreview
	StepApply
	StepComplete
)

// Deps holds the dependencies for the migration wizard.
type Deps struct {
	Ctx context.Context
	Svc *shelly.Service
}

// Validate ensures all required dependencies are set.
func (d Deps) Validate() error {
	if d.Ctx == nil {
		return fmt.Errorf("context is required")
	}
	if d.Svc == nil {
		return fmt.Errorf("service is required")
	}
	return nil
}

// DeviceInfo holds device information for selection.
type DeviceInfo struct {
	Name    string
	Model   string
	Address string
}

// CaptureCompleteMsg signals that config capture is complete.
type CaptureCompleteMsg struct {
	SourceConfig map[string]any
	SourceModel  string
	Err          error
}

// CompareCompleteMsg signals that config comparison is complete.
type CompareCompleteMsg struct {
	Diffs       []model.ConfigDiff
	TargetModel string
	Err         error
}

// ApplyCompleteMsg signals that migration is complete.
type ApplyCompleteMsg struct {
	Success bool
	Err     error
}

// Wizard is the migration wizard model.
type Wizard struct {
	panel.Sizable
	ctx           context.Context
	svc           *shelly.Service
	step          WizardStep
	devices       []DeviceInfo
	sourceIdx     int
	targetIdx     int
	sourceConfig  map[string]any
	sourceModel   string
	targetModel   string
	diffs         []model.ConfigDiff
	selectedDiffs []bool // Tracks which diffs to apply
	diffScroller  *panel.Scroller
	loading       bool
	applying      bool
	err           error
	focused       bool
	panelIndex    int
	includeWiFi   bool
	styles        Styles
}

// Styles holds styles for the migration wizard.
type Styles struct {
	Title       lipgloss.Style
	Step        lipgloss.Style
	Selected    lipgloss.Style
	Cursor      lipgloss.Style
	Error       lipgloss.Style
	Muted       lipgloss.Style
	Success     lipgloss.Style
	Warning     lipgloss.Style
	DiffAdd     lipgloss.Style
	DiffRemove  lipgloss.Style
	DiffChange  lipgloss.Style
	DeviceName  lipgloss.Style
	DeviceModel lipgloss.Style
}

// DefaultStyles returns the default styles.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Step: lipgloss.NewStyle().
			Foreground(colors.Info),
		Selected: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Foreground(colors.Highlight).
			Bold(true),
		Cursor: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Success: lipgloss.NewStyle().
			Foreground(colors.Online),
		Warning: lipgloss.NewStyle().
			Foreground(colors.Warning),
		DiffAdd: lipgloss.NewStyle().
			Foreground(colors.Online),
		DiffRemove: lipgloss.NewStyle().
			Foreground(colors.Error),
		DiffChange: lipgloss.NewStyle().
			Foreground(colors.Warning),
		DeviceName: lipgloss.NewStyle().
			Foreground(colors.Text).
			Bold(true),
		DeviceModel: lipgloss.NewStyle().
			Foreground(colors.Info),
	}
}

// New creates a new migration wizard.
func New(deps Deps) Wizard {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("migration wizard init", err)
		panic(fmt.Sprintf("migration: invalid deps: %v", err))
	}

	w := Wizard{
		Sizable:      panel.NewSizable(6, panel.NewScroller(0, 10)),
		ctx:          deps.Ctx,
		svc:          deps.Svc,
		step:         StepSourceSelect,
		diffScroller: panel.NewScroller(0, 10),
		styles:       DefaultStyles(),
	}
	w.Loader = w.Loader.SetMessage("Loading...")
	return w
}

// Init returns the initial command.
func (w Wizard) Init() tea.Cmd {
	return w.loadDevices()
}

func (w Wizard) loadDevices() tea.Cmd {
	return func() tea.Msg {
		cfg := config.Get()
		if cfg == nil {
			return devicesLoadedMsg{Err: fmt.Errorf("config not loaded")}
		}

		devices := make([]DeviceInfo, 0, len(cfg.Devices))
		for name, dev := range cfg.Devices {
			devices = append(devices, DeviceInfo{
				Name:    name,
				Model:   dev.Model,
				Address: dev.Address,
			})
		}

		return devicesLoadedMsg{Devices: devices}
	}
}

type devicesLoadedMsg struct {
	Devices []DeviceInfo
	Err     error
}

// SetSize sets the component dimensions.
func (w Wizard) SetSize(width, height int) Wizard {
	w.ApplySize(width, height)
	if w.diffScroller != nil {
		// Use 6 as overhead for step indicator, headers, etc.
		w.diffScroller.SetVisibleRows(max(height-6, 1))
	}
	return w
}

// SetFocused sets the focus state.
func (w Wizard) SetFocused(focused bool) Wizard {
	w.focused = focused
	return w
}

// SetPanelIndex sets the panel index for Shift+N hint.
func (w Wizard) SetPanelIndex(index int) Wizard {
	w.panelIndex = index
	return w
}

// Update handles messages.
func (w Wizard) Update(msg tea.Msg) (Wizard, tea.Cmd) {
	// Forward tick messages to loader when loading
	if w.loading || w.applying {
		result := generics.UpdateLoader(w.Loader, msg, func(msg tea.Msg) bool {
			switch msg.(type) {
			case CaptureCompleteMsg, CompareCompleteMsg, ApplyCompleteMsg, devicesLoadedMsg:
				return true
			}
			return false
		})
		w.Loader = result.Loader
		if result.Consumed {
			return w, result.Cmd
		}
	}

	return w.handleMessage(msg)
}

func (w Wizard) handleMessage(msg tea.Msg) (Wizard, tea.Cmd) {
	switch msg := msg.(type) {
	case devicesLoadedMsg:
		return w.handleDevicesLoaded(msg)
	case CaptureCompleteMsg:
		return w.handleCaptureComplete(msg)
	case CompareCompleteMsg:
		return w.handleCompareComplete(msg)
	case ApplyCompleteMsg:
		return w.handleApplyComplete(msg)

	// Action messages from context-based keybindings
	case messages.NavigationMsg:
		if !w.focused {
			return w, nil
		}
		return w.handleNavigation(msg)
	case messages.ToggleEnableRequestMsg:
		if !w.focused {
			return w, nil
		}
		if w.step == StepPreview {
			w = w.withDiffToggled()
		}
		return w, nil
	case messages.RefreshRequestMsg:
		if !w.focused {
			return w, nil
		}
		if w.step == StepComplete {
			return w.reset()
		}
		return w, nil

	case tea.KeyPressMsg:
		if !w.focused {
			return w, nil
		}
		return w.handleKey(msg)
	}
	return w, nil
}

func (w Wizard) handleDevicesLoaded(msg devicesLoadedMsg) (Wizard, tea.Cmd) {
	w.loading = false
	if msg.Err != nil {
		w.err = msg.Err
		return w, nil
	}
	w.devices = msg.Devices
	w.Scroller.SetItemCount(len(w.devices))
	return w, nil
}

func (w Wizard) handleCaptureComplete(msg CaptureCompleteMsg) (Wizard, tea.Cmd) {
	w.loading = false
	if msg.Err != nil {
		w.err = msg.Err
		return w, nil
	}
	w.sourceConfig = msg.SourceConfig
	w.sourceModel = msg.SourceModel
	w.step = StepTargetSelect
	w.Scroller.SetCursor(0)
	return w, nil
}

func (w Wizard) handleCompareComplete(msg CompareCompleteMsg) (Wizard, tea.Cmd) {
	w.loading = false
	if msg.Err != nil {
		w.err = msg.Err
		return w, nil
	}
	w.diffs = msg.Diffs
	w.targetModel = msg.TargetModel
	// Initialize all diffs as selected by default
	w.selectedDiffs = make([]bool, len(w.diffs))
	for i := range w.selectedDiffs {
		w.selectedDiffs[i] = true
	}
	w.diffScroller.SetItemCount(len(w.diffs))
	w.step = StepPreview
	return w, nil
}

func (w Wizard) handleApplyComplete(msg ApplyCompleteMsg) (Wizard, tea.Cmd) {
	w.applying = false
	if msg.Err != nil {
		w.err = msg.Err
		return w, nil
	}
	w.step = StepComplete
	return w, nil
}

func (w Wizard) handleKey(msg tea.KeyPressMsg) (Wizard, tea.Cmd) {
	// Handle action keys
	return w.handleActionKey(msg.String())
}

func (w Wizard) handleNavigation(msg messages.NavigationMsg) (Wizard, tea.Cmd) {
	// Handle step-specific navigation
	switch w.step {
	case StepSourceSelect, StepTargetSelect:
		w.applyNavToScroller(msg, w.Scroller)
	case StepPreview:
		w.applyNavToScroller(msg, w.diffScroller)
	case StepApply, StepComplete:
		// No navigation in these steps
	}
	return w, nil
}

func (w Wizard) applyNavToScroller(msg messages.NavigationMsg, scroller *panel.Scroller) {
	switch msg.Direction {
	case messages.NavUp:
		scroller.CursorUp()
	case messages.NavDown:
		scroller.CursorDown()
	case messages.NavPageUp:
		scroller.PageUp()
	case messages.NavPageDown:
		scroller.PageDown()
	case messages.NavHome:
		scroller.CursorToStart()
	case messages.NavEnd:
		scroller.CursorToEnd()
	case messages.NavLeft, messages.NavRight:
		// Not applicable for migration wizard
	}
}

func (w Wizard) handleActionKey(key string) (Wizard, tea.Cmd) {
	switch key {
	case keyconst.KeyEnter:
		return w.handleEnter()
	case keyconst.KeyEsc:
		return w.handleEscape()
	case "w":
		if w.step == StepSourceSelect {
			w.includeWiFi = !w.includeWiFi
		}
	case "r", "R":
		if w.step == StepComplete {
			return w.reset()
		}
	case " ":
		if w.step == StepPreview {
			w = w.withDiffToggled()
		}
	case "a":
		if w.step == StepPreview {
			w = w.withAllDiffsSelected()
		}
	case "n":
		if w.step == StepPreview {
			w = w.withNoDiffsSelected()
		}
	}

	return w, nil
}

func (w Wizard) withDiffToggled() Wizard {
	cursor := w.diffScroller.Cursor()
	if cursor < len(w.selectedDiffs) {
		w.selectedDiffs[cursor] = !w.selectedDiffs[cursor]
	}
	return w
}

func (w Wizard) withAllDiffsSelected() Wizard {
	for i := range w.selectedDiffs {
		w.selectedDiffs[i] = true
	}
	return w
}

func (w Wizard) withNoDiffsSelected() Wizard {
	for i := range w.selectedDiffs {
		w.selectedDiffs[i] = false
	}
	return w
}

func (w Wizard) handleEnter() (Wizard, tea.Cmd) {
	switch w.step {
	case StepSourceSelect:
		if len(w.devices) == 0 {
			return w, nil
		}
		w.sourceIdx = w.Scroller.Cursor()
		w.loading = true
		w.Loader = w.Loader.SetMessage("Capturing source config...")
		return w, tea.Batch(w.Loader.Tick(), w.captureSource())

	case StepTargetSelect:
		if len(w.devices) == 0 {
			return w, nil
		}
		targetIdx := w.Scroller.Cursor()
		if targetIdx == w.sourceIdx {
			w.err = fmt.Errorf("target must be different from source")
			return w, nil
		}
		w.targetIdx = targetIdx
		w.loading = true
		w.Loader = w.Loader.SetMessage("Comparing configurations...")
		return w, tea.Batch(w.Loader.Tick(), w.compareConfigs())

	case StepPreview:
		if len(w.diffs) == 0 {
			w.step = StepComplete
			return w, nil
		}
		// Check if any diffs are selected
		selectedCount := w.selectedDiffCount()
		if selectedCount == 0 {
			w.err = fmt.Errorf("no changes selected - use space to select items or 'a' to select all")
			return w, nil
		}
		w.applying = true
		w.err = nil
		w.Loader = w.Loader.SetMessage(fmt.Sprintf("Applying %d changes...", selectedCount))
		return w, tea.Batch(w.Loader.Tick(), w.applyMigration())

	case StepApply, StepComplete:
		// No enter action in these steps
	}

	return w, nil
}

func (w Wizard) handleEscape() (Wizard, tea.Cmd) {
	switch w.step {
	case StepTargetSelect:
		w.step = StepSourceSelect
		w.sourceConfig = nil
		w.sourceModel = ""
		w.err = nil
	case StepPreview:
		w.step = StepTargetSelect
		w.diffs = nil
		w.err = nil
	case StepSourceSelect, StepApply, StepComplete:
		// No escape action in these steps
	}
	return w, nil
}

func (w Wizard) reset() (Wizard, tea.Cmd) {
	w.step = StepSourceSelect
	w.sourceConfig = nil
	w.sourceModel = ""
	w.targetModel = ""
	w.diffs = nil
	w.selectedDiffs = nil
	w.err = nil
	w.Scroller.SetCursor(0)
	return w, nil
}

func (w Wizard) captureSource() tea.Cmd {
	if w.sourceIdx >= len(w.devices) {
		return func() tea.Msg {
			return CaptureCompleteMsg{Err: fmt.Errorf("invalid source device")}
		}
	}

	source := w.devices[w.sourceIdx]
	includeWiFi := w.includeWiFi

	return func() tea.Msg {
		tpl, err := w.svc.CaptureTemplate(w.ctx, source.Name, includeWiFi)
		if err != nil {
			return CaptureCompleteMsg{Err: fmt.Errorf("failed to capture config: %w", err)}
		}
		return CaptureCompleteMsg{
			SourceConfig: tpl.Config,
			SourceModel:  tpl.Model,
		}
	}
}

func (w Wizard) compareConfigs() tea.Cmd {
	if w.targetIdx >= len(w.devices) {
		return func() tea.Msg {
			return CompareCompleteMsg{Err: fmt.Errorf("invalid target device")}
		}
	}

	target := w.devices[w.targetIdx]
	sourceConfig := w.sourceConfig
	sourceModel := w.sourceModel

	return func() tea.Msg {
		// First get target device info to validate model compatibility
		targetInfo, err := w.svc.GetDeviceInfo(w.ctx, target.Name)
		if err != nil {
			return CompareCompleteMsg{Err: fmt.Errorf("failed to get target info: %w", err)}
		}

		// Validate device type compatibility
		if !areModelsCompatible(sourceModel, targetInfo.Model) {
			return CompareCompleteMsg{
				Err: fmt.Errorf("incompatible device types: %s → %s (migration requires same model family)", sourceModel, targetInfo.Model),
			}
		}

		diffs, err := w.svc.CompareTemplate(w.ctx, target.Name, sourceConfig)
		if err != nil {
			return CompareCompleteMsg{Err: fmt.Errorf("failed to compare configs: %w", err)}
		}
		return CompareCompleteMsg{Diffs: diffs, TargetModel: targetInfo.Model}
	}
}

// areModelsCompatible checks if two device models are compatible for migration.
// Compatible means they are the same model or in the same product family.
func areModelsCompatible(source, target string) bool {
	// Exact match is always compatible
	if source == target {
		return true
	}

	// Extract model family (e.g., "SHSW-25" from both Plus and non-Plus variants)
	sourceFamily := extractModelFamily(source)
	targetFamily := extractModelFamily(target)

	return sourceFamily == targetFamily && sourceFamily != ""
}

// extractModelFamily returns the model family for compatibility checking.
func extractModelFamily(modelName string) string {
	// Shelly model naming conventions:
	// Gen1: SHSW-25, SHPLG-S, etc.
	// Gen2+: Plus 1PM, Pro 2PM, etc.
	// The first part before space/dash is typically the family

	// Common Plus/Pro prefix handling - strip "Shelly " prefix if present
	name := modelName
	if len(name) > 7 && name[:7] == "Shelly " {
		name = name[7:]
	}

	// For Gen1 models (contain hyphen), use the full model
	for i, c := range name {
		if c == ' ' || (c == '-' && i > 0) {
			return name[:i]
		}
	}

	return name
}

func (w Wizard) applyMigration() tea.Cmd {
	if w.targetIdx >= len(w.devices) {
		return func() tea.Msg {
			return ApplyCompleteMsg{Err: fmt.Errorf("invalid target device")}
		}
	}

	target := w.devices[w.targetIdx]
	// Filter config to only include selected paths
	filteredConfig := w.filterSelectedConfig()

	return func() tea.Msg {
		_, err := w.svc.ApplyTemplate(w.ctx, target.Name, filteredConfig, false)
		if err != nil {
			return ApplyCompleteMsg{Err: err}
		}
		return ApplyCompleteMsg{Success: true}
	}
}

// filterSelectedConfig returns a config map with only the selected diff paths.
func (w Wizard) filterSelectedConfig() map[string]any {
	if len(w.diffs) == 0 || len(w.selectedDiffs) == 0 {
		return w.sourceConfig
	}

	// Collect the top-level keys from selected diffs
	selectedKeys := make(map[string]bool)
	for i, diff := range w.diffs {
		if i < len(w.selectedDiffs) && w.selectedDiffs[i] {
			// Extract top-level key from path (e.g., "sys" from "sys.device.name")
			topKey := extractTopLevelKey(diff.Path)
			if topKey != "" {
				selectedKeys[topKey] = true
			}
		}
	}

	// If no keys selected, return empty config
	if len(selectedKeys) == 0 {
		return make(map[string]any)
	}

	// Build filtered config with only selected top-level keys
	filtered := make(map[string]any)
	for key, val := range w.sourceConfig {
		if selectedKeys[key] {
			filtered[key] = val
		}
	}

	return filtered
}

// extractTopLevelKey extracts the top-level key from a dot-notation path.
func extractTopLevelKey(path string) string {
	for i, c := range path {
		if c == '.' {
			return path[:i]
		}
	}
	return path
}

// selectedDiffCount returns the number of selected diffs.
func (w Wizard) selectedDiffCount() int {
	count := 0
	for _, selected := range w.selectedDiffs {
		if selected {
			count++
		}
	}
	return count
}

// View renders the wizard.
func (w Wizard) View() string {
	r := rendering.New(w.Width, w.Height).
		SetTitle("Migration Wizard").
		SetFocused(w.focused).
		SetPanelIndex(w.panelIndex)

	// Footer based on step
	if w.focused {
		r.SetFooter(w.footerForStep())
	}

	var content strings.Builder

	// Step indicator
	content.WriteString(w.renderStepIndicator())
	content.WriteString("\n\n")

	// Loading state
	if w.loading || w.applying {
		content.WriteString(w.Loader.View())
		r.SetContent(content.String())
		return r.Render()
	}

	// Error display
	if w.err != nil {
		content.WriteString(w.styles.Error.Render(w.err.Error()))
		content.WriteString("\n")
	}

	// Step content
	switch w.step {
	case StepSourceSelect:
		content.WriteString(w.renderSourceSelect())
	case StepTargetSelect:
		content.WriteString(w.renderTargetSelect())
	case StepPreview:
		content.WriteString(w.renderPreview())
	case StepApply:
		// Apply step is handled by loading state
		content.WriteString(w.styles.Muted.Render("Applying configuration..."))
	case StepComplete:
		content.WriteString(w.renderComplete())
	}

	r.SetContent(content.String())
	return r.Render()
}

func (w Wizard) renderStepIndicator() string {
	steps := []string{"Source", "Target", "Preview", "Apply"}
	parts := make([]string, 0, len(steps))

	for i, name := range steps {
		step := WizardStep(i)
		style := w.styles.Muted
		if step == w.step {
			style = w.styles.Title
		} else if step < w.step {
			style = w.styles.Success
		}
		parts = append(parts, style.Render(fmt.Sprintf("%d. %s", i+1, name)))
	}

	return strings.Join(parts, " → ")
}

func (w Wizard) renderSourceSelect() string {
	var content strings.Builder
	content.WriteString(w.styles.Step.Render("Select source device:"))
	content.WriteString("\n")

	if w.includeWiFi {
		content.WriteString(w.styles.Warning.Render("[w] WiFi: INCLUDED"))
	} else {
		content.WriteString(w.styles.Muted.Render("[w] WiFi: excluded"))
	}
	content.WriteString("\n\n")

	if len(w.devices) == 0 {
		content.WriteString(w.styles.Muted.Render("No devices registered"))
		return content.String()
	}

	content.WriteString(w.renderDeviceList(w.Scroller, -1))
	return content.String()
}

func (w Wizard) renderTargetSelect() string {
	var content strings.Builder

	// Show source info
	if w.sourceIdx < len(w.devices) {
		source := w.devices[w.sourceIdx]
		content.WriteString(w.styles.Muted.Render("Source: "))
		content.WriteString(w.styles.DeviceName.Render(source.Name))
		content.WriteString(" ")
		content.WriteString(w.styles.DeviceModel.Render(fmt.Sprintf("[%s]", w.sourceModel)))
		content.WriteString("\n\n")
	}

	content.WriteString(w.styles.Step.Render("Select target device:"))
	content.WriteString("\n\n")

	content.WriteString(w.renderDeviceList(w.Scroller, w.sourceIdx))
	return content.String()
}

func (w Wizard) renderDeviceList(scroller *panel.Scroller, excludeIdx int) string {
	var content strings.Builder

	start, end := scroller.VisibleRange()
	for i := start; i < end && i < len(w.devices); i++ {
		device := w.devices[i]
		isCursor := i == scroller.Cursor()

		cursor := "  "
		if isCursor && w.focused {
			cursor = w.styles.Cursor.Render("> ")
		}

		// Mark excluded device
		var marker string
		if i == excludeIdx {
			marker = w.styles.Muted.Render(" (source)")
		}

		line := fmt.Sprintf("%s%s %s%s",
			cursor,
			w.styles.DeviceName.Render(device.Name),
			w.styles.DeviceModel.Render(fmt.Sprintf("[%s]", device.Model)),
			marker,
		)

		if isCursor && w.focused {
			line = w.styles.Selected.Render(line)
		}

		content.WriteString(line)
		content.WriteString("\n")
	}

	// Scroll indicator
	if len(w.devices) > scroller.VisibleRows() {
		content.WriteString(w.styles.Muted.Render(
			fmt.Sprintf("\n[%d/%d]", scroller.Cursor()+1, len(w.devices)),
		))
	}

	return content.String()
}

func (w Wizard) renderPreview() string {
	var content strings.Builder

	// Show source and target with model info
	if w.sourceIdx < len(w.devices) && w.targetIdx < len(w.devices) {
		source := w.devices[w.sourceIdx]
		target := w.devices[w.targetIdx]
		content.WriteString(w.styles.Muted.Render("Migration: "))
		content.WriteString(w.styles.DeviceName.Render(source.Name))
		content.WriteString(w.styles.DeviceModel.Render(fmt.Sprintf(" [%s]", w.sourceModel)))
		content.WriteString(w.styles.Muted.Render(" → "))
		content.WriteString(w.styles.DeviceName.Render(target.Name))
		content.WriteString(w.styles.DeviceModel.Render(fmt.Sprintf(" [%s]", w.targetModel)))
		content.WriteString("\n\n")
	}

	// Show selection count
	selectedCount := w.selectedDiffCount()
	content.WriteString(w.styles.Step.Render(
		fmt.Sprintf("Configuration changes (%d/%d selected):", selectedCount, len(w.diffs)),
	))
	content.WriteString("\n\n")

	if len(w.diffs) == 0 {
		content.WriteString(w.styles.Success.Render("No changes detected - configurations are identical"))
		return content.String()
	}

	// Render diffs with checkboxes
	start, end := w.diffScroller.VisibleRange()
	for i := start; i < end && i < len(w.diffs); i++ {
		diff := w.diffs[i]
		isCursor := i == w.diffScroller.Cursor()

		// Checkbox
		var checkbox string
		isSelected := i < len(w.selectedDiffs) && w.selectedDiffs[i]
		if isSelected {
			checkbox = w.styles.Success.Render("[✓]")
		} else {
			checkbox = w.styles.Muted.Render("[ ]")
		}

		// Cursor indicator
		cursor := "  "
		if isCursor && w.focused {
			cursor = w.styles.Cursor.Render("> ")
		}

		var style lipgloss.Style
		var prefix string
		switch diff.DiffType {
		case model.DiffAdded:
			style = w.styles.DiffAdd
			prefix = "+"
		case model.DiffRemoved:
			style = w.styles.DiffRemove
			prefix = "-"
		default:
			style = w.styles.DiffChange
			prefix = "~"
		}

		line := w.renderDiffLine(cursor, checkbox, prefix, diff, style, isCursor, isSelected)

		content.WriteString(line)
		content.WriteString("\n")
	}

	// Scroll indicator
	if len(w.diffs) > w.diffScroller.VisibleRows() {
		content.WriteString(w.styles.Muted.Render(
			fmt.Sprintf("\n[%d/%d changes]", w.diffScroller.Cursor()+1, len(w.diffs)),
		))
	}

	return content.String()
}

func (w Wizard) renderDiffLine(cursor, checkbox, prefix string, diff model.ConfigDiff, style lipgloss.Style, isCursor, isSelected bool) string {
	switch {
	case isCursor && w.focused:
		line := fmt.Sprintf("%s%s %s %s: %v → %v", cursor, checkbox, prefix, diff.Path, diff.OldValue, diff.NewValue)
		return w.styles.Selected.Render(line)
	case !isSelected:
		line := fmt.Sprintf("%s%s %s %s: %v → %v", cursor, checkbox, prefix, diff.Path, diff.OldValue, diff.NewValue)
		return w.styles.Muted.Render(line)
	default:
		// Render prefix with diff style, rest with normal style
		return fmt.Sprintf("%s%s %s %s: %v → %v",
			cursor, checkbox,
			style.Render(prefix),
			diff.Path, diff.OldValue, diff.NewValue)
	}
}

func (w Wizard) renderComplete() string {
	var content strings.Builder

	if w.sourceIdx < len(w.devices) && w.targetIdx < len(w.devices) {
		source := w.devices[w.sourceIdx]
		target := w.devices[w.targetIdx]
		content.WriteString(w.styles.Success.Render("✓ Migration complete!"))
		content.WriteString("\n\n")
		content.WriteString(w.styles.Muted.Render("Configuration from "))
		content.WriteString(w.styles.DeviceName.Render(source.Name))
		content.WriteString(w.styles.Muted.Render(" applied to "))
		content.WriteString(w.styles.DeviceName.Render(target.Name))
		content.WriteString("\n\n")
		content.WriteString(w.styles.Muted.Render("Press R to start a new migration"))
	}

	return content.String()
}

func (w Wizard) footerForStep() string {
	var hints []keys.Hint
	switch w.step {
	case StepSourceSelect:
		hints = []keys.Hint{
			{Key: "j/k", Desc: "nav"},
			{Key: "enter", Desc: "select"},
			{Key: "w", Desc: "toggle-wifi"},
		}
	case StepTargetSelect:
		hints = []keys.Hint{
			{Key: "j/k", Desc: "nav"},
			{Key: "enter", Desc: "select"},
			{Key: "esc", Desc: "back"},
		}
	case StepPreview:
		hints = []keys.Hint{
			{Key: "j/k", Desc: "scroll"},
			{Key: "spc", Desc: "toggle"},
			{Key: "a", Desc: "all"},
			{Key: "n", Desc: "none"},
			{Key: "enter", Desc: "apply"},
			{Key: "esc", Desc: "back"},
		}
	case StepComplete:
		hints = []keys.Hint{
			{Key: "R", Desc: "new migration"},
		}
	default:
		return ""
	}
	return theme.StyledKeybindings(keys.FormatHints(hints, keys.FooterHintWidth(w.Width)))
}

// Step returns the current wizard step.
func (w Wizard) Step() WizardStep {
	return w.step
}

// Loading returns whether the wizard is loading.
func (w Wizard) Loading() bool {
	return w.loading
}

// Applying returns whether the wizard is applying.
func (w Wizard) Applying() bool {
	return w.applying
}

// Error returns any error that occurred.
func (w Wizard) Error() error {
	return w.err
}

// FooterText returns keybinding hints for the footer.
func (w Wizard) FooterText() string {
	return w.footerForStep()
}

// Refresh reloads the device list.
func (w Wizard) Refresh() (Wizard, tea.Cmd) {
	w.loading = true
	w.Loader = w.Loader.SetMessage("Loading devices...")
	return w, tea.Batch(w.Loader.Tick(), w.loadDevices())
}
