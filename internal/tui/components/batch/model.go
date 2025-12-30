// Package batch provides TUI components for batch device operations.
package batch

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/tuierrors"
)

// Deps holds the dependencies for the Batch component.
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

// Operation represents a batch operation type.
type Operation int

// Operation constants.
const (
	OpToggle Operation = iota
	OpOn
	OpOff
	OpReboot
	OpCheckFirmware
)

// String returns the operation name.
func (o Operation) String() string {
	switch o {
	case OpToggle:
		return "Toggle"
	case OpOn:
		return "On"
	case OpOff:
		return "Off"
	case OpReboot:
		return "Reboot"
	case OpCheckFirmware:
		return "Check Firmware"
	default:
		return "Unknown"
	}
}

// DeviceSelection represents a device that can be selected for batch operations.
type DeviceSelection struct {
	Name     string
	Address  string
	Selected bool
}

// OperationResult represents the result of an operation on a single device.
type OperationResult struct {
	Name    string
	Success bool
	Err     error
}

// CompleteMsg signals that a batch operation completed.
type CompleteMsg struct {
	Results []OperationResult
}

// Model displays batch operations.
type Model struct {
	ctx        context.Context
	svc        *shelly.Service
	devices    []DeviceSelection
	scroller   *panel.Scroller
	operation  Operation
	executing  bool
	results    []OperationResult
	err        error
	width      int
	height     int
	focused    bool
	panelIndex int
	styles     Styles
	loader     loading.Model
}

// Styles holds styles for the Batch component.
type Styles struct {
	Selected   lipgloss.Style
	Unselected lipgloss.Style
	Cursor     lipgloss.Style
	Operation  lipgloss.Style
	Success    lipgloss.Style
	Failure    lipgloss.Style
	Label      lipgloss.Style
	Error      lipgloss.Style
	Muted      lipgloss.Style
}

// DefaultStyles returns the default styles for the Batch component.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Selected: lipgloss.NewStyle().
			Foreground(colors.Online),
		Unselected: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Cursor: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Foreground(colors.Highlight).
			Bold(true),
		Operation: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Success: lipgloss.NewStyle().
			Foreground(colors.Online),
		Failure: lipgloss.NewStyle().
			Foreground(colors.Error),
		Label: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
	}
}

// New creates a new Batch model.
func New(deps Deps) Model {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("batch component init", err)
		panic(fmt.Sprintf("batch: invalid deps: %v", err))
	}

	return Model{
		ctx:       deps.Ctx,
		svc:       deps.Svc,
		scroller:  panel.NewScroller(0, 10),
		operation: OpToggle,
		styles:    DefaultStyles(),
		loader: loading.New(
			loading.WithMessage("Executing..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(false, false),
		),
	}
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// LoadDevices loads registered devices.
func (m Model) LoadDevices() Model {
	cfg := config.Get()
	if cfg == nil {
		return m
	}

	m.devices = make([]DeviceSelection, 0, len(cfg.Devices))
	for name, dev := range cfg.Devices {
		m.devices = append(m.devices, DeviceSelection{
			Name:     name,
			Address:  dev.Address,
			Selected: false,
		})
	}

	m.scroller.SetItemCount(len(m.devices))
	return m
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	// Reserve space for header, operation selector, and footer
	visibleRows := height - 10
	if visibleRows < 1 {
		visibleRows = 1
	}
	m.scroller.SetVisibleRows(visibleRows)
	// Update loader size
	m.loader = m.loader.SetSize(width-4, height-4)
	return m
}

// SetFocused sets the focus state.
func (m Model) SetFocused(focused bool) Model {
	m.focused = focused
	return m
}

// SetPanelIndex sets the panel index for Shift+N hint.
func (m Model) SetPanelIndex(index int) Model {
	m.panelIndex = index
	return m
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// Forward tick messages to loader when executing
	if m.executing {
		var cmd tea.Cmd
		m.loader, cmd = m.loader.Update(msg)
		switch msg.(type) {
		case CompleteMsg:
			// Pass through to main switch below
		default:
			if cmd != nil {
				return m, cmd
			}
		}
	}

	switch msg := msg.(type) {
	case CompleteMsg:
		m.executing = false
		m.results = msg.Results
		return m, nil

	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	// Handle navigation keys
	if m.handleNavKey(msg.String()) {
		return m, nil
	}

	// Handle action keys
	switch msg.String() {
	case "space":
		m = m.toggleSelection()
	case "a":
		m = m.selectAll()
	case "n":
		m = m.selectNone()
	case "enter", "x":
		return m.execute()
	case "1":
		m.operation = OpToggle
	case "2":
		m.operation = OpOn
	case "3":
		m.operation = OpOff
	case "4":
		m.operation = OpReboot
	case "5":
		m.operation = OpCheckFirmware
	case "r":
		m.results = nil
		m.err = nil
	}

	return m, nil
}

// handleNavKey handles navigation keys, returns true if handled.
func (m Model) handleNavKey(key string) bool {
	switch key {
	case "j", "down":
		m.scroller.CursorDown()
	case "k", "up":
		m.scroller.CursorUp()
	case "g":
		m.scroller.CursorToStart()
	case "G":
		m.scroller.CursorToEnd()
	case "ctrl+d", "pgdown":
		m.scroller.PageDown()
	case "ctrl+u", "pgup":
		m.scroller.PageUp()
	default:
		return false
	}
	return true
}

func (m Model) toggleSelection() Model {
	cursor := m.scroller.Cursor()
	if len(m.devices) > 0 && cursor < len(m.devices) {
		m.devices[cursor].Selected = !m.devices[cursor].Selected
	}
	return m
}

func (m Model) selectAll() Model {
	for i := range m.devices {
		m.devices[i].Selected = true
	}
	return m
}

func (m Model) selectNone() Model {
	for i := range m.devices {
		m.devices[i].Selected = false
	}
	return m
}

func (m Model) selectedDevices() []DeviceSelection {
	selected := make([]DeviceSelection, 0)
	for _, d := range m.devices {
		if d.Selected {
			selected = append(selected, d)
		}
	}
	return selected
}

func (m Model) execute() (Model, tea.Cmd) {
	if m.executing {
		return m, nil
	}

	selected := m.selectedDevices()
	if len(selected) == 0 {
		m.err = fmt.Errorf("no devices selected")
		return m, nil
	}

	m.executing = true
	m.results = nil
	m.err = nil

	return m, tea.Batch(m.loader.Tick(), m.executeOperation(selected))
}

func (m Model) executeOperation(devices []DeviceSelection) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 60*time.Second)
		defer cancel()

		var (
			results []OperationResult
			mu      sync.Mutex
		)

		// Rate limiting is handled at the service layer
		g, gctx := errgroup.WithContext(ctx)

		for _, dev := range devices {
			device := dev
			g.Go(func() error {
				var opErr error
				switch m.operation {
				case OpToggle:
					_, opErr = m.svc.QuickToggle(gctx, device.Address, nil)
				case OpOn:
					_, opErr = m.svc.QuickOn(gctx, device.Address, nil)
				case OpOff:
					_, opErr = m.svc.QuickOff(gctx, device.Address, nil)
				case OpReboot:
					opErr = m.svc.DeviceReboot(gctx, device.Address, 0)
				case OpCheckFirmware:
					_, opErr = m.svc.CheckFirmware(gctx, device.Address)
				}

				mu.Lock()
				results = append(results, OperationResult{
					Name:    device.Name,
					Success: opErr == nil,
					Err:     opErr,
				})
				mu.Unlock()

				return nil // Don't fail the group on individual errors
			})
		}

		if err := g.Wait(); err != nil {
			// This shouldn't happen since we return nil from each goroutine
			iostreams.DebugErr("batch operation", err)
		}

		return CompleteMsg{Results: results}
	}
}

// View renders the Batch component.
func (m Model) View() string {
	r := rendering.New(m.width, m.height).
		SetTitle("Batch Operations").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	// Add footer with keybindings when focused
	if m.focused {
		r.SetFooter("spc:sel a:all n:none x:exec 1-5:op")
	}

	var content strings.Builder

	// Operation selector
	content.WriteString(m.renderOperationSelector())
	content.WriteString("\n\n")

	// Device list
	if len(m.devices) == 0 {
		content.WriteString(m.styles.Muted.Render("No devices registered"))
	} else {
		content.WriteString(m.renderDeviceList())
	}

	// Results
	if len(m.results) > 0 {
		content.WriteString("\n\n")
		content.WriteString(m.renderResults())
	}

	// Error display with categorized messaging and retry hint
	if m.err != nil {
		msg, hint := tuierrors.FormatError(m.err)
		content.WriteString("\n")
		content.WriteString(m.styles.Error.Render(msg))
		content.WriteString("\n")
		content.WriteString(m.styles.Muted.Render("  " + hint))
		content.WriteString("\n")
		content.WriteString(m.styles.Muted.Render("  Press 'r' to reset and retry"))
	}

	// Executing indicator
	if m.executing {
		content.WriteString("\n")
		content.WriteString(m.loader.View())
	}

	r.SetContent(content.String())
	return r.Render()
}

func (m Model) renderOperationSelector() string {
	ops := []struct {
		op   Operation
		key  string
		name string
	}{
		{OpToggle, "1", "Toggle"},
		{OpOn, "2", "On"},
		{OpOff, "3", "Off"},
		{OpReboot, "4", "Reboot"},
		{OpCheckFirmware, "5", "Firmware"},
	}

	parts := make([]string, 0, len(ops))
	for _, op := range ops {
		style := m.styles.Muted
		if op.op == m.operation {
			style = m.styles.Operation
		}
		parts = append(parts, style.Render(fmt.Sprintf("[%s] %s", op.key, op.name)))
	}

	return m.styles.Label.Render("Operation: ") + strings.Join(parts, " ")
}

func (m Model) renderDeviceList() string {
	var content strings.Builder

	selected := m.selectedDevices()
	content.WriteString(m.styles.Label.Render(
		fmt.Sprintf("Devices (%d selected):", len(selected)),
	))
	content.WriteString("\n\n")

	start, end := m.scroller.VisibleRange()
	for i := start; i < end; i++ {
		device := m.devices[i]
		isCursor := m.scroller.IsCursorAt(i)
		content.WriteString(m.renderDeviceLine(device, isCursor))
		if i < end-1 {
			content.WriteString("\n")
		}
	}

	// Scroll indicator
	content.WriteString("\n")
	content.WriteString(m.styles.Muted.Render(m.scroller.ScrollInfo()))

	return content.String()
}

func (m Model) renderDeviceLine(device DeviceSelection, isCursor bool) string {
	cursor := "  "
	if isCursor {
		cursor = "▶ "
	}

	// Checkbox
	var checkbox string
	if device.Selected {
		checkbox = m.styles.Selected.Render("[✓]")
	} else {
		checkbox = m.styles.Unselected.Render("[ ]")
	}

	// Name and address
	nameStr := device.Name
	addrStr := m.styles.Muted.Render(fmt.Sprintf(" (%s)", device.Address))

	line := fmt.Sprintf("%s%s %s%s", cursor, checkbox, nameStr, addrStr)

	if isCursor {
		return m.styles.Cursor.Render(line)
	}
	return line
}

func (m Model) renderResults() string {
	var content strings.Builder
	content.WriteString(m.styles.Label.Render("Results:\n"))

	successCount := 0
	failCount := 0
	for _, r := range m.results {
		if r.Success {
			successCount++
			content.WriteString(m.styles.Success.Render(fmt.Sprintf("  ✓ %s\n", r.Name)))
		} else {
			failCount++
			errMsg := "unknown error"
			if r.Err != nil {
				errMsg = r.Err.Error()
			}
			content.WriteString(m.styles.Failure.Render(fmt.Sprintf("  ✗ %s: %s\n", r.Name, errMsg)))
		}
	}

	content.WriteString(m.styles.Label.Render(
		fmt.Sprintf("\nSuccess: %d, Failed: %d", successCount, failCount),
	))

	return content.String()
}

// Devices returns the device list.
func (m Model) Devices() []DeviceSelection {
	return m.devices
}

// Operation returns the current operation.
func (m Model) Operation() Operation {
	return m.operation
}

// SetOperation sets the operation.
func (m Model) SetOperation(op Operation) Model {
	m.operation = op
	return m
}

// Executing returns whether an operation is in progress.
func (m Model) Executing() bool {
	return m.executing
}

// Results returns the operation results.
func (m Model) Results() []OperationResult {
	return m.results
}

// Error returns any error that occurred.
func (m Model) Error() error {
	return m.err
}

// Cursor returns the current cursor position.
func (m Model) Cursor() int {
	return m.scroller.Cursor()
}

// SelectedCount returns the number of selected devices.
func (m Model) SelectedCount() int {
	count := 0
	for _, d := range m.devices {
		if d.Selected {
			count++
		}
	}
	return count
}

// Refresh reloads devices and clears state.
func (m Model) Refresh() (Model, tea.Cmd) {
	m.results = nil
	m.err = nil
	m = m.LoadDevices()
	return m, nil
}

// FooterText returns keybinding hints for the footer.
func (m Model) FooterText() string {
	return "j/k:scroll g/G:top/bottom space:select enter:run a:all"
}
