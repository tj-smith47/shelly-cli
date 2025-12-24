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

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
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
	ctx     context.Context
	svc     *shelly.Service
	device  string
	items   []Item
	cursor  int
	scroll  int
	loading bool
	err     error
	width   int
	height  int
	focused bool
	styles  Styles
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
	}
}

// New creates a new KVS browser model.
func New(deps Deps) Model {
	if err := deps.Validate(); err != nil {
		panic(fmt.Sprintf("kvs: invalid deps: %v", err))
	}

	return Model{
		ctx:     deps.Ctx,
		svc:     deps.Svc,
		loading: false,
		styles:  DefaultStyles(),
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
	m.cursor = 0
	m.scroll = 0
	m.err = nil

	if device == "" {
		return m, nil
	}

	m.loading = true
	return m, m.fetchItems()
}

// fetchItems creates a command to fetch KVS items from the device.
func (m Model) fetchItems() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		kvsItems, err := m.svc.GetAllKVS(ctx, m.device)
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
	return m
}

// SetFocused sets the focus state.
func (m Model) SetFocused(focused bool) Model {
	m.focused = focused
	return m
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case LoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.items = msg.Items
		m.cursor = 0
		m.scroll = 0
		return m, nil

	case ActionMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.loading = true
		return m, m.fetchItems()

	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		m = m.cursorDown()
	case "k", "up":
		m = m.cursorUp()
	case "g":
		m.cursor = 0
		m.scroll = 0
	case "G":
		m = m.cursorToEnd()
	case "enter":
		return m, m.selectItem()
	case "d":
		return m, m.deleteItem()
	case "r":
		m.loading = true
		return m, m.fetchItems()
	}

	return m, nil
}

func (m Model) cursorDown() Model {
	if m.cursor < len(m.items)-1 {
		m.cursor++
		m = m.ensureVisible()
	}
	return m
}

func (m Model) cursorUp() Model {
	if m.cursor > 0 {
		m.cursor--
		m = m.ensureVisible()
	}
	return m
}

func (m Model) cursorToEnd() Model {
	if len(m.items) > 0 {
		m.cursor = len(m.items) - 1
		m = m.ensureVisible()
	}
	return m
}

func (m Model) ensureVisible() Model {
	visible := m.visibleRows()
	if m.cursor < m.scroll {
		m.scroll = m.cursor
	} else if m.cursor >= m.scroll+visible {
		m.scroll = m.cursor - visible + 1
	}
	return m
}

func (m Model) visibleRows() int {
	rows := m.height - 4
	if rows < 1 {
		return 1
	}
	return rows
}

func (m Model) selectItem() tea.Cmd {
	if len(m.items) == 0 || m.cursor >= len(m.items) {
		return nil
	}
	item := m.items[m.cursor]
	return func() tea.Msg {
		return SelectMsg{Item: item}
	}
}

func (m Model) deleteItem() tea.Cmd {
	if len(m.items) == 0 || m.cursor >= len(m.items) {
		return nil
	}
	item := m.items[m.cursor]

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.DeleteKVS(ctx, m.device, item.Key)
		return ActionMsg{Action: "delete", Key: item.Key, Err: err}
	}
}

// View renders the KVS browser.
func (m Model) View() string {
	r := rendering.New(m.width, m.height).
		SetTitle("Key-Value Store").
		SetFocused(m.focused)

	if m.device == "" {
		r.SetContent(m.styles.Muted.Render("No device selected"))
		return r.Render()
	}

	if m.loading {
		r.SetContent(m.styles.Muted.Render("Loading KVS..."))
		return r.Render()
	}

	if m.err != nil {
		r.SetContent(m.styles.Error.Render("Error: " + m.err.Error()))
		return r.Render()
	}

	if len(m.items) == 0 {
		r.SetContent(m.styles.Muted.Render("No KVS entries"))
		return r.Render()
	}

	var content strings.Builder
	visible := m.visibleRows()
	endIdx := m.scroll + visible
	if endIdx > len(m.items) {
		endIdx = len(m.items)
	}

	for i := m.scroll; i < endIdx; i++ {
		item := m.items[i]
		isSelected := i == m.cursor

		line := m.renderItemLine(item, isSelected)
		content.WriteString(line)
		if i < endIdx-1 {
			content.WriteString("\n")
		}
	}

	if len(m.items) > visible {
		content.WriteString(m.styles.Muted.Render(
			fmt.Sprintf("\n[%d/%d]", m.cursor+1, len(m.items)),
		))
	}

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
	if len(m.items) == 0 || m.cursor >= len(m.items) {
		return nil
	}
	return &m.items[m.cursor]
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
	return m, m.fetchItems()
}
