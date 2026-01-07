// Package scripts provides TUI components for managing device scripts.
package scripts

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// TemplateSelectedMsg signals that a template was selected for insertion.
type TemplateSelectedMsg struct {
	Device   string
	ScriptID int
	Name     string
	Code     string
	Append   bool // If true, append to existing code; if false, replace
}

// TemplateModel displays a list of templates for selection.
type TemplateModel struct {
	device    string
	scriptID  int
	visible   bool
	templates []config.ScriptTemplate
	cursor    int
	scroll    int
	width     int
	height    int
	append    bool // Append mode (vs replace)
	styles    templateStyles
}

type templateStyles struct {
	Title       lipgloss.Style
	Selected    lipgloss.Style
	Normal      lipgloss.Style
	Description lipgloss.Style
	Muted       lipgloss.Style
	ModeActive  lipgloss.Style
}

func defaultTemplateStyles() templateStyles {
	colors := theme.GetSemanticColors()
	return templateStyles{
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Selected: lipgloss.NewStyle().
			Background(colors.Highlight).
			Foreground(colors.Background).
			Bold(true),
		Normal: lipgloss.NewStyle().
			Foreground(colors.Text),
		Description: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		ModeActive: lipgloss.NewStyle().
			Foreground(colors.Success).
			Bold(true),
	}
}

// NewTemplateModel creates a new template selector model.
func NewTemplateModel() TemplateModel {
	// Load all available templates
	templateMap := automation.ListAllScriptTemplates()
	templates := make([]config.ScriptTemplate, 0, len(templateMap))
	for _, tpl := range templateMap {
		templates = append(templates, tpl)
	}

	return TemplateModel{
		templates: templates,
		append:    true, // Default to append mode
		styles:    defaultTemplateStyles(),
	}
}

// Show displays the template selector.
func (m TemplateModel) Show(device string, scriptID int) TemplateModel {
	m.device = device
	m.scriptID = scriptID
	m.visible = true
	m.cursor = 0
	m.scroll = 0
	return m
}

// Hide hides the template selector.
func (m TemplateModel) Hide() TemplateModel {
	m.visible = false
	return m
}

// IsVisible returns whether the selector is visible.
func (m TemplateModel) IsVisible() bool {
	return m.visible
}

// SetSize sets the modal dimensions.
func (m TemplateModel) SetSize(width, height int) TemplateModel {
	m.width = width
	m.height = height
	return m
}

// Init returns the initial command.
func (m TemplateModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m TemplateModel) Update(msg tea.Msg) (TemplateModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		return m.handleKey(keyMsg)
	}

	return m, nil
}

func (m TemplateModel) handleKey(msg tea.KeyPressMsg) (TemplateModel, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m = m.Hide()
		return m, func() tea.Msg { return messages.EditClosedMsg{Saved: false} }

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			m = m.adjustScroll()
		}

	case "down", "j":
		if m.cursor < len(m.templates)-1 {
			m.cursor++
			m = m.adjustScroll()
		}

	case "enter":
		if m.cursor >= 0 && m.cursor < len(m.templates) {
			tpl := m.templates[m.cursor]
			m = m.Hide()
			return m, func() tea.Msg {
				return TemplateSelectedMsg{
					Device:   m.device,
					ScriptID: m.scriptID,
					Name:     tpl.Name,
					Code:     tpl.Code,
					Append:   m.append,
				}
			}
		}

	case "tab":
		// Toggle append/replace mode
		m.append = !m.append

	case "home", "g":
		m.cursor = 0
		m.scroll = 0

	case "end", "G":
		m.cursor = len(m.templates) - 1
		m = m.adjustScroll()
	}

	return m, nil
}

func (m TemplateModel) adjustScroll() TemplateModel {
	visible := m.visibleItems()
	if m.cursor < m.scroll {
		m.scroll = m.cursor
	} else if m.cursor >= m.scroll+visible {
		m.scroll = m.cursor - visible + 1
	}
	return m
}

func (m TemplateModel) visibleItems() int {
	// Account for borders, title, mode indicator, footer
	return max(1, m.height-10)
}

// View renders the template selector.
func (m TemplateModel) View() string {
	if !m.visible {
		return ""
	}

	modeStr := "APPEND"
	if !m.append {
		modeStr = "REPLACE"
	}
	footer := "↑/↓:Select | Tab:" + modeStr + " | Enter:Insert | Esc:Cancel"

	r := rendering.NewModal(m.width, m.height, "Insert Template", footer)
	return r.SetContent(m.renderContent()).Render()
}

func (m TemplateModel) renderContent() string {
	if len(m.templates) == 0 {
		return m.styles.Muted.Render("No templates available")
	}

	var content strings.Builder

	// Mode indicator
	modeLabel := "Mode: "
	if m.append {
		content.WriteString(modeLabel + m.styles.ModeActive.Render("Append") + m.styles.Muted.Render(" (add to existing code)"))
	} else {
		content.WriteString(modeLabel + m.styles.ModeActive.Render("Replace") + m.styles.Muted.Render(" (replace all code)"))
	}
	content.WriteString("\n\n")

	// Template list
	visible := m.visibleItems()
	endIdx := min(m.scroll+visible, len(m.templates))

	for i := m.scroll; i < endIdx; i++ {
		tpl := m.templates[i]

		var name string
		if i == m.cursor {
			name = m.styles.Selected.Render("▶ " + tpl.Name)
		} else {
			name = m.styles.Normal.Render("  " + tpl.Name)
		}
		content.WriteString(name + "\n")

		// Show description for selected item
		if i == m.cursor && tpl.Description != "" {
			desc := "    " + tpl.Description
			content.WriteString(m.styles.Description.Render(desc) + "\n")
		}
	}

	// Scroll indicator
	if len(m.templates) > visible {
		content.WriteString(m.styles.Muted.Render(
			strings.Repeat("\n", 1) +
				"[" + string(rune('0'+m.cursor+1)) + "/" + string(rune('0'+len(m.templates))) + "]",
		))
	}

	return content.String()
}
