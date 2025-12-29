// Package kvs provides TUI components for browsing device key-value store.
package kvs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	shellykvs "github.com/tj-smith47/shelly-cli/internal/shelly/kvs"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// Item represents a key-value pair in the device KVS.
type Item struct {
	Key   string
	Value any
	Etag  string
}

// Deps holds the dependencies for the KVS browser component.
type Deps struct {
	Ctx context.Context
	Svc *shellykvs.Service
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

// LoadedMsg signals that KVS items were loaded.
type LoadedMsg struct {
	Items []Item
	Err   error
}

// ActionMsg signals a KVS action result.
type ActionMsg struct {
	Action string // "set", "delete"
	Key    string
	Err    error
}

// SelectMsg signals that an item was selected for viewing.
type SelectMsg struct {
	Item Item
}

// Model displays KVS items for a device.
type Model struct {
	ctx              context.Context
	svc              *shellykvs.Service
	device           string
	items            []Item
	scroller         *panel.Scroller
	loading          bool
	editing          bool
	confirmingDelete bool
	deleteKey        string
	err              error
	width            int
	height           int
	focused          bool
	panelIndex       int // 1-based panel index for Shift+N hotkey hint
	styles           Styles
	loader           loading.Model
	editModal        EditModel
}

// Styles holds styles for the KVS browser component.
type Styles struct {
	Key      lipgloss.Style
	Value    lipgloss.Style
	String   lipgloss.Style
	Number   lipgloss.Style
	Bool     lipgloss.Style
	Null     lipgloss.Style
	Object   lipgloss.Style
	Selected lipgloss.Style
	Error    lipgloss.Style
	Muted    lipgloss.Style
	Warning  lipgloss.Style
	Confirm  lipgloss.Style
}

// DefaultStyles returns the default styles for the KVS browser.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Key: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Value: lipgloss.NewStyle().
			Foreground(colors.Text),
		String: lipgloss.NewStyle().
			Foreground(colors.Online),
		Number: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Bool: lipgloss.NewStyle().
			Foreground(colors.Info),
		Null: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Object: lipgloss.NewStyle().
			Foreground(colors.Error),
		Selected: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Foreground(colors.Highlight).
			Bold(true),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Warning: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Confirm: lipgloss.NewStyle().
			Foreground(colors.Error).
			Bold(true),
	}
}

// New creates a new KVS browser model.
func New(deps Deps) Model {
	if err := deps.Validate(); err != nil {
		panic(fmt.Sprintf("kvs: invalid deps: %v", err))
	}

	return Model{
		ctx:      deps.Ctx,
		svc:      deps.Svc,
		scroller: panel.NewScroller(0, 10),
		loading:  false,
		styles:   DefaultStyles(),
		loader: loading.New(
			loading.WithMessage("Loading KVS..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(true, true),
		),
		editModal: NewEditModel(deps.Ctx, deps.Svc),
	}
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// SetDevice sets the device to browse KVS for and triggers a fetch.
func (m Model) SetDevice(device string) (Model, tea.Cmd) {
	m.device = device
	m.items = nil
	m.scroller.SetItemCount(0)
	m.scroller.CursorToStart()
	m.err = nil

	if device == "" {
		return m, nil
	}

	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchItems())
}

// fetchItems creates a command to fetch KVS items from the device.
func (m Model) fetchItems() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		kvsItems, err := m.svc.GetAll(ctx, m.device)
		if err != nil {
			return LoadedMsg{Err: err}
		}

		result := make([]Item, len(kvsItems))
		for i, item := range kvsItems {
			result[i] = Item{
				Key:   item.Key,
				Value: item.Value,
				Etag:  item.Etag,
			}
		}

		return LoadedMsg{Items: result}
	}
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	visibleRows := height - 4
	if visibleRows < 1 {
		visibleRows = 1
	}
	m.scroller.SetVisibleRows(visibleRows)
	return m
}

// SetFocused sets the focus state.
func (m Model) SetFocused(focused bool) Model {
	m.focused = focused
	return m
}

// SetPanelIndex sets the 1-based panel index for Shift+N hotkey hint.
func (m Model) SetPanelIndex(index int) Model {
	m.panelIndex = index
	return m
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// Handle edit modal if visible
	if m.editing {
		return m.handleEditModalUpdate(msg)
	}

	// Handle delete confirmation
	if m.confirmingDelete {
		return m.handleDeleteConfirmation(msg)
	}

	// Forward tick messages to loader when loading
	if m.loading {
		var cmd tea.Cmd
		m.loader, cmd = m.loader.Update(msg)
		// Continue processing LoadedMsg even during loading
		if _, ok := msg.(LoadedMsg); !ok {
			if cmd != nil {
				return m, cmd
			}
		}
	}

	switch msg := msg.(type) {
	case LoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.items = msg.Items
		m.scroller.SetItemCount(len(m.items))
		m.scroller.CursorToStart()
		return m, nil

	case ActionMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.loading = true
		return m, tea.Batch(m.loader.Tick(), m.fetchItems(), func() tea.Msg {
			return EditClosedMsg{Saved: true}
		})

	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleEditModalUpdate(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.editModal, cmd = m.editModal.Update(msg)

	// Check if modal was closed
	if !m.editModal.IsVisible() {
		m.editing = false
		// Refresh data after edit
		m.loading = true
		return m, tea.Batch(cmd, m.loader.Tick(), m.fetchItems())
	}

	// Handle save result message
	if saveMsg, ok := msg.(EditSaveResultMsg); ok {
		if saveMsg.Err == nil {
			m.editing = false
			m.editModal = m.editModal.Hide()
			// Refresh data after successful save
			m.loading = true
			return m, tea.Batch(m.loader.Tick(), m.fetchItems(), func() tea.Msg {
				return EditClosedMsg{Saved: true}
			})
		}
	}

	return m, cmd
}

func (m Model) handleDeleteConfirmation(msg tea.Msg) (Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch keyMsg.String() {
		case "y", "Y":
			m.confirmingDelete = false
			return m, m.executeDelete(m.deleteKey)
		case "n", "N", "esc":
			m.confirmingDelete = false
			m.deleteKey = ""
			return m, nil
		}
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
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
	case "enter":
		return m, m.selectItem()
	case "e":
		return m.handleEditKey()
	case "n":
		return m.handleNewKey()
	case "d":
		return m.handleDeleteKey()
	case "r":
		m.loading = true
		return m, tea.Batch(m.loader.Tick(), m.fetchItems())
	}

	return m, nil
}

func (m Model) handleEditKey() (Model, tea.Cmd) {
	if m.device == "" || m.loading || len(m.items) == 0 {
		return m, nil
	}
	cursor := m.scroller.Cursor()
	if cursor >= len(m.items) {
		return m, nil
	}
	item := m.items[cursor]
	m.editing = true
	m.editModal = m.editModal.SetSize(m.width, m.height)
	m.editModal = m.editModal.ShowEdit(m.device, &item)
	return m, func() tea.Msg { return EditOpenedMsg{} }
}

func (m Model) handleNewKey() (Model, tea.Cmd) {
	if m.device == "" || m.loading {
		return m, nil
	}
	m.editing = true
	m.editModal = m.editModal.SetSize(m.width, m.height)
	m.editModal = m.editModal.ShowNew(m.device)
	return m, func() tea.Msg { return EditOpenedMsg{} }
}

func (m Model) handleDeleteKey() (Model, tea.Cmd) {
	if m.device == "" || m.loading || len(m.items) == 0 {
		return m, nil
	}
	cursor := m.scroller.Cursor()
	if cursor >= len(m.items) {
		return m, nil
	}
	// Start delete confirmation
	m.confirmingDelete = true
	m.deleteKey = m.items[cursor].Key
	return m, nil
}

func (m Model) executeDelete(key string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.Delete(ctx, m.device, key)
		return ActionMsg{Action: "delete", Key: key, Err: err}
	}
}

func (m Model) selectItem() tea.Cmd {
	cursor := m.scroller.Cursor()
	if len(m.items) == 0 || cursor >= len(m.items) {
		return nil
	}
	item := m.items[cursor]
	return func() tea.Msg {
		return SelectMsg{Item: item}
	}
}

// View renders the KVS browser.
func (m Model) View() string {
	// Render edit modal if editing
	if m.editing {
		return m.editModal.View()
	}

	r := rendering.New(m.width, m.height).
		SetTitle("Key-Value Store").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	if m.device == "" {
		r.SetContent(m.styles.Muted.Render("No device selected"))
		return r.Render()
	}

	if m.loading {
		r.SetContent(m.loader.View())
		return r.Render()
	}

	if m.err != nil {
		errMsg := m.err.Error()
		// Detect Gen1 or unsupported device errors and show a friendly message
		if strings.Contains(errMsg, "404") || strings.Contains(errMsg, "unknown method") ||
			strings.Contains(errMsg, "not found") {
			r.SetContent(m.styles.Muted.Render("KVS not supported on this device"))
		} else {
			r.SetContent(m.styles.Error.Render("Error: " + errMsg))
		}
		return r.Render()
	}

	// Show delete confirmation
	if m.confirmingDelete {
		var content strings.Builder
		content.WriteString(m.styles.Confirm.Render("Delete key: " + m.deleteKey + "?"))
		content.WriteString("\n\n")
		content.WriteString(m.styles.Warning.Render("Press Y to confirm, N or Esc to cancel"))
		r.SetContent(content.String())
		return r.Render()
	}

	if len(m.items) == 0 {
		r.SetContent(m.styles.Muted.Render("No KVS entries"))
		return r.Render()
	}

	var content strings.Builder

	start, end := m.scroller.VisibleRange()
	for i := start; i < end; i++ {
		item := m.items[i]
		isSelected := m.scroller.IsCursorAt(i)

		line := m.renderItemLine(item, isSelected)
		content.WriteString(line)
		if i < end-1 {
			content.WriteString("\n")
		}
	}

	// Scroll indicator
	content.WriteString("\n")
	content.WriteString(m.styles.Muted.Render(m.scroller.ScrollInfo()))

	r.SetContent(content.String())
	return r.Render()
}

func (m Model) renderItemLine(item Item, isSelected bool) string {
	// Selection indicator
	selector := "  "
	if isSelected {
		selector = "▶ "
	}

	// Key (truncate if too long)
	key := item.Key
	if len(key) > 20 {
		key = key[:17] + "..."
	}
	keyStr := m.styles.Key.Render(fmt.Sprintf("%-20s", key))

	// Value display
	valueStr := m.formatValue(item.Value)

	line := fmt.Sprintf("%s%s = %s", selector, keyStr, valueStr)

	if isSelected {
		return m.styles.Selected.Render(line)
	}
	return line
}

func (m Model) formatValue(value any) string {
	if value == nil {
		return m.styles.Null.Render("null")
	}

	switch v := value.(type) {
	case string:
		display := v
		if len(display) > 30 {
			display = display[:27] + "..."
		}
		return m.styles.String.Render(fmt.Sprintf("%q", display))
	case float64:
		if v == float64(int64(v)) {
			return m.styles.Number.Render(fmt.Sprintf("%d", int64(v)))
		}
		return m.styles.Number.Render(fmt.Sprintf("%g", v))
	case bool:
		return m.styles.Bool.Render(fmt.Sprintf("%v", v))
	case map[string]any, []any:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return m.styles.Object.Render("{...}")
		}
		display := string(jsonBytes)
		if len(display) > 30 {
			display = display[:27] + "..."
		}
		return m.styles.Object.Render(display)
	default:
		return m.styles.Value.Render(fmt.Sprintf("%v", v))
	}
}

// SelectedItem returns the currently selected item, if any.
func (m Model) SelectedItem() *Item {
	cursor := m.scroller.Cursor()
	if len(m.items) == 0 || cursor >= len(m.items) {
		return nil
	}
	return &m.items[cursor]
}

// Cursor returns the current cursor position.
func (m Model) Cursor() int {
	return m.scroller.Cursor()
}

// ItemCount returns the number of items.
func (m Model) ItemCount() int {
	return len(m.items)
}

// Device returns the current device address.
func (m Model) Device() string {
	return m.device
}

// Loading returns whether the component is loading.
func (m Model) Loading() bool {
	return m.loading
}

// Error returns any error that occurred.
func (m Model) Error() error {
	return m.err
}

// Refresh triggers a refresh of the KVS items.
func (m Model) Refresh() (Model, tea.Cmd) {
	if m.device == "" {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchItems())
}

// FooterText returns keybinding hints for the footer.
func (m Model) FooterText() string {
	return "j/k:scroll e:edit n:new d:delete r:refresh"
}

// IsEditing returns whether the edit modal is currently open.
func (m Model) IsEditing() bool {
	return m.editing
}
