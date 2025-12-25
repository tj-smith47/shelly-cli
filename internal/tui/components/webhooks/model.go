// Package webhooks provides TUI components for managing device webhooks.
package webhooks

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// Webhook represents a webhook configuration on a device.
type Webhook struct {
	ID     int
	Name   string
	Event  string
	Enable bool
	URLs   []string
	Cid    int
}

// Deps holds the dependencies for the webhooks component.
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

// LoadedMsg signals that webhooks were loaded.
type LoadedMsg struct {
	Webhooks []Webhook
	Err      error
}

// ActionMsg signals a webhook action result.
type ActionMsg struct {
	Action    string // "enable", "disable", "delete"
	WebhookID int
	Err       error
}

// SelectMsg signals that a webhook was selected.
type SelectMsg struct {
	Webhook Webhook
}

// CreateMsg signals that a new webhook should be created.
type CreateMsg struct {
	Device string
}

// Model displays webhooks for a device.
type Model struct {
	ctx      context.Context
	svc      *shelly.Service
	device   string
	webhooks []Webhook
	cursor   int
	scroll   int
	loading  bool
	err      error
	width    int
	height   int
	focused  bool
	styles   Styles
}

// Styles holds styles for the webhook list component.
type Styles struct {
	Enabled  lipgloss.Style
	Disabled lipgloss.Style
	Event    lipgloss.Style
	URL      lipgloss.Style
	Name     lipgloss.Style
	Selected lipgloss.Style
	Error    lipgloss.Style
	Muted    lipgloss.Style
}

// DefaultStyles returns the default styles for the webhook list.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Enabled: lipgloss.NewStyle().
			Foreground(colors.Online),
		Disabled: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Event: lipgloss.NewStyle().
			Foreground(colors.Warning),
		URL: lipgloss.NewStyle().
			Foreground(colors.Info),
		Name: lipgloss.NewStyle().
			Foreground(colors.Text).
			Bold(true),
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

// New creates a new webhooks model.
func New(deps Deps) Model {
	if err := deps.Validate(); err != nil {
		panic(fmt.Sprintf("webhooks: invalid deps: %v", err))
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

// SetDevice sets the device to list webhooks for and triggers a fetch.
func (m Model) SetDevice(device string) (Model, tea.Cmd) {
	m.device = device
	m.webhooks = nil
	m.cursor = 0
	m.scroll = 0
	m.err = nil

	if device == "" {
		return m, nil
	}

	m.loading = true
	return m, m.fetchWebhooks()
}

// fetchWebhooks creates a command to fetch webhooks from the device.
func (m Model) fetchWebhooks() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		hooks, err := m.svc.ListWebhooks(ctx, m.device)
		if err != nil {
			return LoadedMsg{Err: err}
		}

		result := make([]Webhook, len(hooks))
		for i, h := range hooks {
			result[i] = Webhook{
				ID:     h.ID,
				Name:   h.Name,
				Event:  h.Event,
				Enable: h.Enable,
				URLs:   h.URLs,
				Cid:    h.Cid,
			}
		}

		return LoadedMsg{Webhooks: result}
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
		m.webhooks = msg.Webhooks
		m.cursor = 0
		m.scroll = 0
		return m, nil

	case ActionMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.loading = true
		return m, m.fetchWebhooks()

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
		return m, m.selectWebhook()
	case "t":
		return m, m.toggleWebhook()
	case "d":
		return m, m.deleteWebhook()
	case "n":
		return m, m.createWebhook()
	case "r":
		m.loading = true
		return m, m.fetchWebhooks()
	}

	return m, nil
}

func (m Model) createWebhook() tea.Cmd {
	if m.device == "" {
		return nil
	}
	return func() tea.Msg {
		return CreateMsg{Device: m.device}
	}
}

func (m Model) cursorDown() Model {
	if m.cursor < len(m.webhooks)-1 {
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
	if len(m.webhooks) > 0 {
		m.cursor = len(m.webhooks) - 1
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

func (m Model) selectWebhook() tea.Cmd {
	if len(m.webhooks) == 0 || m.cursor >= len(m.webhooks) {
		return nil
	}
	webhook := m.webhooks[m.cursor]
	return func() tea.Msg {
		return SelectMsg{Webhook: webhook}
	}
}

func (m Model) toggleWebhook() tea.Cmd {
	if len(m.webhooks) == 0 || m.cursor >= len(m.webhooks) {
		return nil
	}
	webhook := m.webhooks[m.cursor]

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		newEnable := !webhook.Enable
		err := m.svc.UpdateWebhook(ctx, m.device, webhook.ID, shelly.UpdateWebhookParams{
			Event:  webhook.Event,
			URLs:   webhook.URLs,
			Name:   webhook.Name,
			Enable: &newEnable,
		})

		action := "enable"
		if !newEnable {
			action = "disable"
		}
		return ActionMsg{Action: action, WebhookID: webhook.ID, Err: err}
	}
}

func (m Model) deleteWebhook() tea.Cmd {
	if len(m.webhooks) == 0 || m.cursor >= len(m.webhooks) {
		return nil
	}
	webhook := m.webhooks[m.cursor]

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.DeleteWebhook(ctx, m.device, webhook.ID)
		return ActionMsg{Action: "delete", WebhookID: webhook.ID, Err: err}
	}
}

// View renders the webhooks list.
func (m Model) View() string {
	r := rendering.New(m.width, m.height).
		SetTitle("Webhooks").
		SetFocused(m.focused)

	// Add footer with keybindings when focused
	if m.focused && m.device != "" && len(m.webhooks) > 0 {
		r.SetFooter("t:toggle d:del n:new r:refresh")
	} else if m.focused && m.device != "" {
		r.SetFooter("n:new r:refresh")
	}

	if m.device == "" {
		r.SetContent(m.styles.Muted.Render("No device selected"))
		return r.Render()
	}

	if m.loading {
		r.SetContent(m.styles.Muted.Render("Loading webhooks..."))
		return r.Render()
	}

	if m.err != nil {
		r.SetContent(m.styles.Error.Render("Error: " + m.err.Error()))
		return r.Render()
	}

	if len(m.webhooks) == 0 {
		r.SetContent(m.styles.Muted.Render("No webhooks configured"))
		return r.Render()
	}

	var content strings.Builder
	visible := m.visibleRows()
	endIdx := m.scroll + visible
	if endIdx > len(m.webhooks) {
		endIdx = len(m.webhooks)
	}

	for i := m.scroll; i < endIdx; i++ {
		webhook := m.webhooks[i]
		isSelected := i == m.cursor

		line := m.renderWebhookLine(webhook, isSelected)
		content.WriteString(line)
		if i < endIdx-1 {
			content.WriteString("\n")
		}
	}

	if len(m.webhooks) > visible {
		content.WriteString(m.styles.Muted.Render(
			fmt.Sprintf("\n[%d/%d]", m.cursor+1, len(m.webhooks)),
		))
	}

	r.SetContent(content.String())
	return r.Render()
}

func (m Model) renderWebhookLine(webhook Webhook, isSelected bool) string {
	// Status icon
	var icon string
	if webhook.Enable {
		icon = m.styles.Enabled.Render("●")
	} else {
		icon = m.styles.Disabled.Render("○")
	}

	// Selection indicator
	selector := "  "
	if isSelected {
		selector = "▶ "
	}

	// Event type (truncate if too long)
	event := webhook.Event
	if len(event) > 25 {
		event = event[:22] + "..."
	}
	eventStr := m.styles.Event.Render(event)

	// URL count or first URL
	urlInfo := ""
	if len(webhook.URLs) > 0 {
		url := webhook.URLs[0]
		if len(url) > 30 {
			url = url[:27] + "..."
		}
		if len(webhook.URLs) > 1 {
			urlInfo = fmt.Sprintf("%s +%d", url, len(webhook.URLs)-1)
		} else {
			urlInfo = url
		}
		urlInfo = m.styles.URL.Render(urlInfo)
	}

	line := fmt.Sprintf("%s%s %s %s", selector, icon, eventStr, urlInfo)

	if isSelected {
		return m.styles.Selected.Render(line)
	}
	return line
}

// SelectedWebhook returns the currently selected webhook, if any.
func (m Model) SelectedWebhook() *Webhook {
	if len(m.webhooks) == 0 || m.cursor >= len(m.webhooks) {
		return nil
	}
	return &m.webhooks[m.cursor]
}

// WebhookCount returns the number of webhooks.
func (m Model) WebhookCount() int {
	return len(m.webhooks)
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

// Refresh triggers a refresh of the webhook list.
func (m Model) Refresh() (Model, tea.Cmd) {
	if m.device == "" {
		return m, nil
	}
	m.loading = true
	return m, m.fetchWebhooks()
}
