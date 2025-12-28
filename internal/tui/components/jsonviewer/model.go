// Package jsonviewer provides a syntax-highlighted JSON viewer overlay.
package jsonviewer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// CloseMsg is sent when the viewer is closed.
type CloseMsg struct{}

// FetchedMsg is sent when JSON data is fetched.
type FetchedMsg struct {
	Data  map[string]any
	Error error
}

// Model displays JSON responses with syntax highlighting.
type Model struct {
	ctx           context.Context
	svc           *shelly.Service
	deviceAddress string
	endpoint      string
	endpoints     []string
	endpointIdx   int
	data          map[string]any
	viewport      viewport.Model
	isLoading     bool
	error         error
	visible       bool
	width         int
	height        int
	styles        Styles
	loader        loading.Model // Loading spinner
}

// Styles for the JSON viewer.
type Styles struct {
	Container lipgloss.Style
	Title     lipgloss.Style
	Key       lipgloss.Style
	String    lipgloss.Style
	Number    lipgloss.Style
	Bool      lipgloss.Style
	Null      lipgloss.Style
	Bracket   lipgloss.Style
	Error     lipgloss.Style
	Loading   lipgloss.Style
	Nav       lipgloss.Style
	Footer    lipgloss.Style
}

// DefaultStyles returns default styles for the JSON viewer.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Container: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colors.Highlight).
			Padding(1, 2),
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Key: lipgloss.NewStyle().
			Foreground(colors.Info),
		String: lipgloss.NewStyle().
			Foreground(colors.Online),
		Number: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Bool: lipgloss.NewStyle().
			Foreground(colors.Info),
		Null: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Bracket: lipgloss.NewStyle().
			Foreground(colors.Text),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Loading: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
		Nav: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Footer: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
	}
}

// New creates a new JSON viewer model.
func New(ctx context.Context, svc *shelly.Service) Model {
	vp := viewport.New(viewport.WithWidth(60), viewport.WithHeight(20))

	return Model{
		ctx:      ctx,
		svc:      svc,
		viewport: vp,
		styles:   DefaultStyles(),
		loader: loading.New(
			loading.WithMessage("Loading..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(true, true),
		),
	}
}

// Init initializes the JSON viewer.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the JSON viewer.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	// Handle FetchedMsg first (regardless of loading state)
	if fetchedMsg, ok := msg.(FetchedMsg); ok {
		return m.handleFetchedMsg(fetchedMsg)
	}

	// Update loader for spinner animation when loading
	if m.isLoading {
		var cmd tea.Cmd
		m.loader, cmd = m.loader.Update(msg)
		return m, cmd
	}

	// Handle key presses
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		return m.handleKeyPress(keyMsg)
	}

	return m, nil
}

// handleFetchedMsg processes the FetchedMsg when data arrives.
func (m Model) handleFetchedMsg(msg FetchedMsg) (Model, tea.Cmd) {
	m.isLoading = false
	if msg.Error != nil {
		m.error = msg.Error
	} else {
		m.data = msg.Data
		m.error = nil
		m.viewport.SetContent(m.formatJSON())
	}
	return m, nil
}

// handleKeyPress processes key press messages.
func (m Model) handleKeyPress(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("escape", "q"))):
		m.visible = false
		return m, func() tea.Msg { return CloseMsg{} }

	case key.Matches(msg, key.NewBinding(key.WithKeys("h", "left"))):
		return m.prevEndpoint()

	case key.Matches(msg, key.NewBinding(key.WithKeys("l", "right"))):
		return m.nextEndpoint()

	case key.Matches(msg, key.NewBinding(key.WithKeys("g"))):
		m.viewport.GotoTop()

	case key.Matches(msg, key.NewBinding(key.WithKeys("G"))):
		m.viewport.GotoBottom()

	case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
		m.isLoading = true
		return m, tea.Batch(m.loader.Tick(), m.fetchEndpoint())
	}

	// Forward to viewport for scrolling (handles j/k/up/down/pgup/pgdn)
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// prevEndpoint navigates to the previous endpoint.
func (m Model) prevEndpoint() (Model, tea.Cmd) {
	if len(m.endpoints) > 1 && m.endpointIdx > 0 {
		m.endpointIdx--
		m.endpoint = m.endpoints[m.endpointIdx]
		m.isLoading = true
		return m, tea.Batch(m.loader.Tick(), m.fetchEndpoint())
	}
	return m, nil
}

// nextEndpoint navigates to the next endpoint.
func (m Model) nextEndpoint() (Model, tea.Cmd) {
	if len(m.endpoints) > 1 && m.endpointIdx < len(m.endpoints)-1 {
		m.endpointIdx++
		m.endpoint = m.endpoints[m.endpointIdx]
		m.isLoading = true
		return m, tea.Batch(m.loader.Tick(), m.fetchEndpoint())
	}
	return m, nil
}

// View renders the JSON viewer overlay.
func (m Model) View() string {
	if !m.visible {
		return ""
	}

	colors := theme.GetSemanticColors()
	r := rendering.New(m.width, m.height).
		SetTitle("JSON: " + m.endpoint).
		SetFocused(true).
		SetFocusColor(colors.Highlight)

	var content strings.Builder

	// Navigation bar if multiple endpoints
	if len(m.endpoints) > 1 {
		nav := m.renderNav()
		content.WriteString(nav + "\n\n")
	}

	// Content
	switch {
	case m.isLoading:
		content.WriteString(m.loader.View())
	case m.error != nil:
		content.WriteString(m.styles.Error.Render("Error: " + m.error.Error()))
	default:
		content.WriteString(m.viewport.View())
	}

	// Footer
	content.WriteString("\n\n")
	content.WriteString(m.renderFooter())

	return r.SetContent(content.String()).Render()
}

func (m Model) renderNav() string {
	var parts []string

	// Previous indicator
	if m.endpointIdx > 0 {
		parts = append(parts, m.styles.Nav.Render("← "+m.endpoints[m.endpointIdx-1]))
	} else {
		parts = append(parts, strings.Repeat(" ", 20))
	}

	// Current
	parts = append(parts, m.styles.Title.Render(m.endpoint))

	// Next indicator
	if m.endpointIdx < len(m.endpoints)-1 {
		parts = append(parts, m.styles.Nav.Render(m.endpoints[m.endpointIdx+1]+" →"))
	}

	return strings.Join(parts, " │ ")
}

func (m Model) renderFooter() string {
	return m.styles.Footer.Render("Esc close │ h/l endpoint │ j/k scroll │ r refresh")
}

func (m Model) formatJSON() string {
	if m.data == nil {
		return m.styles.Null.Render("null")
	}

	formatted, err := json.MarshalIndent(m.data, "", "  ")
	if err != nil {
		return m.styles.Error.Render("Error formatting JSON: " + err.Error())
	}

	return m.highlightJSON(string(formatted))
}

// highlightJSON applies chroma-based syntax highlighting to JSON.
func (m Model) highlightJSON(jsonStr string) string {
	lexer := lexers.Get("json")
	if lexer == nil {
		return jsonStr
	}
	lexer = chroma.Coalesce(lexer)

	// Get a style that matches the current theme
	style := m.getChromaStyle()

	// Use terminal256 formatter for broad compatibility
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	iterator, err := lexer.Tokenise(nil, jsonStr)
	if err != nil {
		return jsonStr
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return jsonStr
	}

	return buf.String()
}

// getChromaStyle returns a chroma style that matches the current theme.
func (m Model) getChromaStyle() *chroma.Style {
	// Try to match the current theme name to a chroma style
	currentTheme := viper.GetString("theme.name")
	if currentTheme == "" {
		currentTheme = "dracula"
	}

	// Map theme names to chroma styles
	styleMap := map[string]string{
		"dracula":      "dracula",
		"nord":         "nord",
		"gruvbox":      "gruvbox",
		"gruvbox-dark": "gruvbox",
		"tokyo-night":  "tokyonight-night",
		"catppuccin":   "catppuccin-mocha",
	}

	if chromaStyle, ok := styleMap[strings.ToLower(currentTheme)]; ok {
		if style := styles.Get(chromaStyle); style != nil {
			return style
		}
	}

	// Default to dracula which works well on dark terminals
	if style := styles.Get("dracula"); style != nil {
		return style
	}

	return styles.Fallback
}

// fetchEndpoint fetches JSON from device.
func (m Model) fetchEndpoint() tea.Cmd {
	return func() tea.Msg {
		// Note: isLoading is set by caller, this just performs the fetch

		// Parse endpoint to method call
		// Format: "Switch.GetStatus?id=0" or "Shelly.GetStatus"
		method := m.endpoint
		params := make(map[string]any)

		if idx := strings.Index(m.endpoint, "?"); idx != -1 {
			method = m.endpoint[:idx]
			queryStr := m.endpoint[idx+1:]
			for _, part := range strings.Split(queryStr, "&") {
				kv := strings.SplitN(part, "=", 2)
				if len(kv) == 2 {
					params[kv[0]] = kv[1]
				}
			}
		}

		// Call the device
		ctx, cancel := context.WithTimeout(m.ctx, 5000000000) // 5 seconds
		defer cancel()

		result, err := m.svc.RawRPC(ctx, m.deviceAddress, method, params)
		if err != nil {
			return FetchedMsg{Error: fmt.Errorf("failed to call %s: %w", method, err)}
		}

		// Convert result to map
		data, ok := result.(map[string]any)
		if !ok {
			// Wrap non-map results
			data = map[string]any{"result": result}
		}

		return FetchedMsg{Data: data}
	}
}

// Open opens the JSON viewer for a device endpoint.
func (m Model) Open(deviceAddress, endpoint string, endpoints []string) (Model, tea.Cmd) {
	m.visible = true
	m.isLoading = true
	m.error = nil
	m.data = nil
	m.deviceAddress = deviceAddress
	m.endpoint = endpoint
	m.endpoints = endpoints

	// Find current endpoint index
	m.endpointIdx = 0
	for i, ep := range endpoints {
		if ep == endpoint {
			m.endpointIdx = i
			break
		}
	}

	// Update viewport size
	// Subtract: 2 for borders, 2 for horizontal padding
	contentWidth := m.width - 4
	// Subtract: 2 for borders, 3 for footer (\n\n + line), 1 buffer for nav
	contentHeight := m.height - 6
	if contentWidth > 10 {
		m.viewport.SetWidth(contentWidth)
	}
	if contentHeight > 5 {
		m.viewport.SetHeight(contentHeight)
	}

	return m, tea.Batch(m.loader.Tick(), m.fetchEndpoint())
}

// Close closes the JSON viewer.
func (m Model) Close() Model {
	m.visible = false
	m.data = nil
	m.error = nil
	return m
}

// Visible returns whether the viewer is open.
func (m Model) Visible() bool {
	return m.visible
}

// SetSize sets the viewer dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height

	// Subtract: 2 for borders, 2 for horizontal padding
	contentWidth := width - 4
	// Subtract: 2 for borders, 3 for footer (\n\n + line), 1 buffer for nav
	contentHeight := height - 6
	if contentWidth > 10 {
		m.viewport.SetWidth(contentWidth)
	}
	if contentHeight > 5 {
		m.viewport.SetHeight(contentHeight)
	}

	return m
}

// Loading returns whether the viewer is loading.
func (m Model) Loading() bool {
	return m.isLoading
}

// Error returns any error that occurred.
func (m Model) Error() error {
	return m.error
}
