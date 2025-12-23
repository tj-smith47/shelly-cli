// Package jsonviewer provides a syntax-highlighted JSON viewer overlay.
package jsonviewer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
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
	loading       bool
	error         error
	visible       bool
	width         int
	height        int
	styles        Styles
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

	switch msg := msg.(type) {
	case FetchedMsg:
		m.loading = false
		if msg.Error != nil {
			m.error = msg.Error
		} else {
			m.data = msg.Data
			m.error = nil
			m.viewport.SetContent(m.formatJSON())
		}
		return m, nil

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("escape", "q"))):
			m.visible = false
			return m, func() tea.Msg { return CloseMsg{} }

		case key.Matches(msg, key.NewBinding(key.WithKeys("h", "left"))):
			// Previous endpoint
			if len(m.endpoints) > 1 && m.endpointIdx > 0 {
				m.endpointIdx--
				m.endpoint = m.endpoints[m.endpointIdx]
				return m, m.fetchEndpoint()
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("l", "right"))):
			// Next endpoint
			if len(m.endpoints) > 1 && m.endpointIdx < len(m.endpoints)-1 {
				m.endpointIdx++
				m.endpoint = m.endpoints[m.endpointIdx]
				return m, m.fetchEndpoint()
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("g"))):
			// Go to top
			m.viewport.GotoTop()

		case key.Matches(msg, key.NewBinding(key.WithKeys("G"))):
			// Go to bottom
			m.viewport.GotoBottom()

		case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
			// Refresh
			return m, m.fetchEndpoint()
		}

		// Forward to viewport for scrolling (handles j/k/up/down/pgup/pgdn)
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
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
	case m.loading:
		content.WriteString(m.styles.Loading.Render("Loading..."))
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

func (m Model) highlightJSON(jsonStr string) string {
	var result strings.Builder
	inString := false
	inKey := false
	afterColon := false
	escapeNext := false

	for i := 0; i < len(jsonStr); i++ {
		c := jsonStr[i]

		if escapeNext {
			result.WriteByte(c)
			escapeNext = false
			continue
		}

		if c == '\\' && inString {
			result.WriteByte(c)
			escapeNext = true
			continue
		}

		switch c {
		case '"':
			if !inString {
				// Opening quote - entering string
				inString = true
				inKey = isJSONKey(jsonStr, i)
				result.WriteString(m.renderStringChar(c, inKey))
			} else {
				// Closing quote - exiting string
				result.WriteString(m.renderStringChar(c, inKey))
				inString = false
				inKey = false
			}
		case ':':
			result.WriteString(m.renderPunctuation(c, inString, inKey, &afterColon, true))
		case '{', '}', '[', ']':
			result.WriteString(m.renderPunctuation(c, inString, inKey, &afterColon, false))
		case ',':
			result.WriteString(m.renderPunctuation(c, inString, inKey, &afterColon, false))
			if !inString {
				afterColon = false
			}
		default:
			i = m.renderDefault(&result, jsonStr, i, c, inString, inKey, afterColon)
		}
	}

	return result.String()
}

func (m Model) renderStringChar(c byte, inKey bool) string {
	if inKey {
		return m.styles.Key.Render(string(c))
	}
	return m.styles.String.Render(string(c))
}

func (m Model) renderPunctuation(c byte, inString, inKey bool, afterColon *bool, setAfterColon bool) string {
	switch {
	case !inString:
		if setAfterColon {
			*afterColon = true
		}
		return m.styles.Bracket.Render(string(c))
	case inKey:
		return m.styles.Key.Render(string(c))
	default:
		return m.styles.String.Render(string(c))
	}
}

func (m Model) renderDefault(result *strings.Builder, jsonStr string, i int, c byte, inString, inKey, afterColon bool) int {
	switch {
	case inString && inKey:
		result.WriteString(m.styles.Key.Render(string(c)))
	case inString:
		result.WriteString(m.styles.String.Render(string(c)))
	case afterColon:
		remaining := jsonStr[i:]
		switch {
		case strings.HasPrefix(remaining, "true"):
			result.WriteString(m.styles.Bool.Render("true"))
			return i + 3
		case strings.HasPrefix(remaining, "false"):
			result.WriteString(m.styles.Bool.Render("false"))
			return i + 4
		case strings.HasPrefix(remaining, "null"):
			result.WriteString(m.styles.Null.Render("null"))
			return i + 3
		case c == '-' || (c >= '0' && c <= '9'):
			numEnd := i
			for numEnd < len(jsonStr) && isNumberChar(jsonStr[numEnd]) {
				numEnd++
			}
			result.WriteString(m.styles.Number.Render(jsonStr[i:numEnd]))
			return numEnd - 1
		default:
			result.WriteByte(c)
		}
	default:
		result.WriteByte(c)
	}
	return i
}

func isJSONKey(s string, quotePos int) bool {
	// Look for `:` after the closing quote
	depth := 0
	for i := quotePos + 1; i < len(s); i++ {
		c := s[i]
		if c == '\\' {
			i++ // Skip escaped char
			continue
		}
		if c == '"' {
			depth++
			if depth == 1 {
				// Found closing quote, look for :
				for j := i + 1; j < len(s); j++ {
					if s[j] == ' ' || s[j] == '\t' || s[j] == '\n' {
						continue
					}
					return s[j] == ':'
				}
			}
		}
	}
	return false
}

func isNumberChar(c byte) bool {
	return (c >= '0' && c <= '9') || c == '.' || c == '-' || c == '+' || c == 'e' || c == 'E'
}

// fetchEndpoint fetches JSON from device.
func (m Model) fetchEndpoint() tea.Cmd {
	return func() tea.Msg {
		m.loading = true

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
	m.loading = true
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
	contentWidth := m.width - 6
	contentHeight := m.height - 10
	if contentWidth > 10 {
		m.viewport.SetWidth(contentWidth)
	}
	if contentHeight > 5 {
		m.viewport.SetHeight(contentHeight)
	}

	return m, m.fetchEndpoint()
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

	contentWidth := width - 6
	contentHeight := height - 10
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
	return m.loading
}

// Error returns any error that occurred.
func (m Model) Error() error {
	return m.error
}
