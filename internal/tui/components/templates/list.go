// Package templates provides TUI components for managing device templates.
package templates

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/generics"
	"github.com/tj-smith47/shelly-cli/internal/tui/helpers"
	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/styles"
	"github.com/tj-smith47/shelly-cli/internal/tui/tuierrors"
)

const footerKeybindings = "c:create a:apply d:diff D:del x:export i:import r:refresh"

// ListDeps holds the dependencies for the templates list component.
type ListDeps struct {
	Ctx context.Context
	Svc *shelly.Service
}

// Validate ensures all required dependencies are set.
func (d ListDeps) Validate() error {
	if d.Ctx == nil {
		return fmt.Errorf("context is required")
	}
	if d.Svc == nil {
		return fmt.Errorf("service is required")
	}
	return nil
}

// LoadedMsg signals that templates were loaded.
type LoadedMsg struct {
	Templates []config.DeviceTemplate
	Err       error
}

// ActionMsg signals a template action result.
type ActionMsg struct {
	Action       string // "delete", "apply"
	TemplateName string
	Err          error
}

// CreateTemplateMsg signals that a new template should be created from a device.
type CreateTemplateMsg struct {
	Device string // Device to create template from
}

// EditTemplateMsg signals that a template should be edited.
type EditTemplateMsg struct {
	Template config.DeviceTemplate
}

// ApplyTemplateMsg signals that a template should be applied to a device.
type ApplyTemplateMsg struct {
	Template config.DeviceTemplate
}

// ExportTemplateMsg signals that a template was exported.
type ExportTemplateMsg struct {
	TemplateName string
	FilePath     string
	Err          error
}

// ImportTemplateMsg signals that a template was imported.
type ImportTemplateMsg struct {
	TemplateName string
	Err          error
}

// DiffTemplateMsg signals that a template diff should be shown.
type DiffTemplateMsg struct {
	Template config.DeviceTemplate
}

// ListModel displays and manages templates.
type ListModel struct {
	helpers.Sizable
	ctx           context.Context
	svc           *shelly.Service
	templates     []config.DeviceTemplate
	loading       bool
	applying      bool
	err           error
	focused       bool
	panelIndex    int
	pendingDelete string // Template name pending delete confirmation
	statusMsg     string // Temporary status message (export/import success)
	styles        ListStyles
}

// ListStyles holds styles for the list component.
type ListStyles struct {
	Name        lipgloss.Style
	Description lipgloss.Style
	Model       lipgloss.Style
	Selected    lipgloss.Style
	Error       lipgloss.Style
	Success     lipgloss.Style
	Muted       lipgloss.Style
	Cursor      lipgloss.Style
}

// DefaultListStyles returns the default styles for the template list.
func DefaultListStyles() ListStyles {
	colors := theme.GetSemanticColors()
	return ListStyles{
		Name: lipgloss.NewStyle().
			Foreground(colors.Text).
			Bold(true),
		Description: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Model: lipgloss.NewStyle().
			Foreground(colors.Info),
		Selected: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Foreground(colors.Highlight).
			Bold(true),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Success: lipgloss.NewStyle().
			Foreground(colors.Success),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Cursor: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
	}
}

// NewList creates a new templates list model.
func NewList(deps ListDeps) ListModel {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("templates list component init", err)
		panic(fmt.Sprintf("templates: invalid deps: %v", err))
	}

	m := ListModel{
		Sizable: helpers.NewSizable(4, panel.NewScroller(0, 10)),
		ctx:     deps.Ctx,
		svc:     deps.Svc,
		styles:  DefaultListStyles(),
	}
	m.Loader = m.Loader.SetMessage("Loading templates...")
	return m
}

// Init returns the initial command.
func (m ListModel) Init() tea.Cmd {
	return m.loadTemplates()
}

func (m ListModel) loadTemplates() tea.Cmd {
	return func() tea.Msg {
		templates := config.ListDeviceTemplates()

		// Convert map to sorted slice
		result := make([]config.DeviceTemplate, 0, len(templates))
		for _, tpl := range templates {
			result = append(result, tpl)
		}

		// Sort by name
		sort.Slice(result, func(i, j int) bool {
			return result[i].Name < result[j].Name
		})

		return LoadedMsg{Templates: result}
	}
}

// SetSize sets the component dimensions.
func (m ListModel) SetSize(width, height int) ListModel {
	m.ApplySize(width, height)
	return m
}

// SetFocused sets the focus state.
func (m ListModel) SetFocused(focused bool) ListModel {
	m.focused = focused
	return m
}

// SetPanelIndex sets the 1-based panel index for Shift+N hotkey hint.
func (m ListModel) SetPanelIndex(index int) ListModel {
	m.panelIndex = index
	return m
}

// Update handles messages.
func (m ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
	// Forward tick messages to loader when loading
	if m.loading {
		result := generics.UpdateLoader(m.Loader, msg, func(msg tea.Msg) bool {
			switch msg.(type) {
			case LoadedMsg, ActionMsg:
				return true
			}
			return false
		})
		m.Loader = result.Loader
		if result.Consumed {
			return m, result.Cmd
		}
	}

	return m.handleMessage(msg)
}

func (m ListModel) handleMessage(msg tea.Msg) (ListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case LoadedMsg:
		return m.handleLoaded(msg)
	case ActionMsg:
		return m.handleAction(msg)
	case ExportTemplateMsg:
		return m.handleExport(msg)
	case ImportTemplateMsg:
		return m.handleImport(msg)
	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m ListModel) handleLoaded(msg LoadedMsg) (ListModel, tea.Cmd) {
	m.loading = false
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.templates = msg.Templates
	m.Scroller.SetItemCount(len(m.templates))
	return m, nil
}

func (m ListModel) handleAction(msg ActionMsg) (ListModel, tea.Cmd) {
	m.applying = false
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	// Refresh list after action
	return m, m.loadTemplates()
}

func (m ListModel) handleExport(msg ExportTemplateMsg) (ListModel, tea.Cmd) {
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.statusMsg = fmt.Sprintf("Exported %q to %s", msg.TemplateName, msg.FilePath)
	return m, nil
}

func (m ListModel) handleImport(msg ImportTemplateMsg) (ListModel, tea.Cmd) {
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.statusMsg = fmt.Sprintf("Imported template %q", msg.TemplateName)
	// Refresh list after import
	return m, m.loadTemplates()
}

func (m ListModel) handleKey(msg tea.KeyPressMsg) (ListModel, tea.Cmd) {
	// Handle navigation keys first
	if keys.HandleScrollNavigation(msg.String(), m.Scroller) {
		m.pendingDelete = "" // Clear pending delete on navigation
		return m, nil
	}

	// Handle action keys
	switch msg.String() {
	case "c":
		// Create template (signals to parent to show device selector)
		m.pendingDelete = ""
		return m, func() tea.Msg { return CreateTemplateMsg{} }
	case "enter", "a":
		// Apply template to device
		m.pendingDelete = ""
		return m.applyTemplate()
	case "e":
		// Edit template
		m.pendingDelete = ""
		return m.editTemplate()
	case "d":
		// Diff template vs device
		m.pendingDelete = ""
		return m.diffTemplate()
	case "D":
		// Delete template - requires double press for confirmation
		tpl := m.selectedTemplate()
		if tpl == nil {
			return m, nil
		}
		if m.pendingDelete == tpl.Name {
			// Second press - confirm delete
			m.pendingDelete = ""
			return m.deleteTemplate()
		}
		// First press - mark pending
		m.pendingDelete = tpl.Name
		return m, nil
	case "x":
		// Export template
		m.pendingDelete = ""
		m.statusMsg = ""
		return m.exportTemplate()
	case "i":
		// Import template
		m.pendingDelete = ""
		m.statusMsg = ""
		return m.importTemplate()
	case "r", "R":
		// Refresh list
		m.pendingDelete = ""
		m.statusMsg = ""
		m.loading = true
		return m, tea.Batch(m.Loader.Tick(), m.loadTemplates())
	case "esc":
		// Cancel pending delete
		if m.pendingDelete != "" {
			m.pendingDelete = ""
			return m, nil
		}
	}

	return m, nil
}

func (m ListModel) selectedTemplate() *config.DeviceTemplate {
	cursor := m.Scroller.Cursor()
	if len(m.templates) == 0 || cursor >= len(m.templates) {
		return nil
	}
	return &m.templates[cursor]
}

func (m ListModel) applyTemplate() (ListModel, tea.Cmd) {
	tpl := m.selectedTemplate()
	if tpl == nil {
		return m, nil
	}
	tplCopy := *tpl
	return m, func() tea.Msg {
		return ApplyTemplateMsg{Template: tplCopy}
	}
}

func (m ListModel) editTemplate() (ListModel, tea.Cmd) {
	tpl := m.selectedTemplate()
	if tpl == nil {
		return m, nil
	}
	tplCopy := *tpl
	return m, func() tea.Msg {
		return EditTemplateMsg{Template: tplCopy}
	}
}

func (m ListModel) deleteTemplate() (ListModel, tea.Cmd) {
	tpl := m.selectedTemplate()
	if tpl == nil {
		return m, nil
	}
	tplName := tpl.Name

	return m, func() tea.Msg {
		err := config.DeleteDeviceTemplate(tplName)
		return ActionMsg{Action: "delete", TemplateName: tplName, Err: err}
	}
}

func (m ListModel) diffTemplate() (ListModel, tea.Cmd) {
	tpl := m.selectedTemplate()
	if tpl == nil {
		return m, nil
	}
	tplCopy := *tpl
	return m, func() tea.Msg {
		return DiffTemplateMsg{Template: tplCopy}
	}
}

func (m ListModel) exportTemplate() (ListModel, tea.Cmd) {
	tpl := m.selectedTemplate()
	if tpl == nil {
		return m, nil
	}
	tplName := tpl.Name

	return m, func() tea.Msg {
		// Get templates export directory
		configDir, err := config.Dir()
		if err != nil {
			return ExportTemplateMsg{TemplateName: tplName, Err: err}
		}

		templatesDir := filepath.Join(configDir, "templates")
		if err := config.Fs().MkdirAll(templatesDir, 0o755); err != nil {
			return ExportTemplateMsg{TemplateName: tplName, Err: err}
		}

		// Export to file
		outputPath := filepath.Join(templatesDir, tplName+".json")
		filePath, err := config.ExportDeviceTemplateToFile(tplName, outputPath)
		return ExportTemplateMsg{TemplateName: tplName, FilePath: filePath, Err: err}
	}
}

func (m ListModel) importTemplate() (ListModel, tea.Cmd) {
	return m, func() tea.Msg {
		// Get templates import directory
		configDir, err := config.Dir()
		if err != nil {
			return ImportTemplateMsg{Err: err}
		}

		templatesDir := filepath.Join(configDir, "templates")

		// List available template files
		entries, err := afero.ReadDir(config.Fs(), templatesDir)
		if err != nil {
			return ImportTemplateMsg{Err: fmt.Errorf("no templates directory: %s", templatesDir)}
		}

		// Find first JSON/YAML file that's not already imported
		existingTemplates := config.ListDeviceTemplates()
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			ext := filepath.Ext(name)
			if ext != ".json" && ext != ".yaml" && ext != ".yml" {
				continue
			}
			// Check if template already exists
			baseName := strings.TrimSuffix(name, ext)
			if _, exists := existingTemplates[baseName]; exists {
				continue
			}
			// Import this template
			filePath := filepath.Join(templatesDir, name)
			msg, err := config.ImportTemplateFromFile(filePath, "", false)
			if err != nil {
				return ImportTemplateMsg{Err: err}
			}
			// Parse template name from message
			return ImportTemplateMsg{TemplateName: msg}
		}

		return ImportTemplateMsg{Err: fmt.Errorf("no new template files found in %s", templatesDir)}
	}
}

// View renders the templates list.
func (m ListModel) View() string {
	r := rendering.New(m.Width, m.Height).
		SetTitle("Templates").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	// Add footer with keybindings when focused
	if m.focused && len(m.templates) > 0 {
		footer := m.buildFooter()
		r.SetFooter(footer)
	}

	if m.loading {
		r.SetContent(m.Loader.View())
		return r.Render()
	}

	if m.applying {
		r.SetContent(m.styles.Muted.Render("Applying template..."))
		return r.Render()
	}

	if m.err != nil {
		if tuierrors.IsUnsupportedFeature(m.err) {
			r.SetContent(styles.EmptyStateWithBorder(tuierrors.UnsupportedMessage("Templates"), m.Width, m.Height))
		} else {
			msg, _ := tuierrors.FormatError(m.err)
			r.SetContent(m.styles.Error.Render(msg))
		}
		return r.Render()
	}

	if len(m.templates) == 0 {
		r.SetContent(styles.EmptyStateWithBorder("No templates defined\nPress c to create one", m.Width, m.Height))
		return r.Render()
	}

	content := generics.RenderScrollableList(generics.ListRenderConfig[config.DeviceTemplate]{
		Items:    m.templates,
		Scroller: m.Scroller,
		RenderItem: func(tpl config.DeviceTemplate, _ int, isCursor bool) string {
			return m.renderTemplateLine(tpl, isCursor)
		},
		ScrollStyle:    m.styles.Muted,
		ScrollInfoMode: generics.ScrollInfoAlways,
	})

	r.SetContent(content)
	return r.Render()
}

func (m ListModel) renderTemplateLine(tpl config.DeviceTemplate, isSelected bool) string {
	// Selection indicator
	selector := "  "
	if isSelected && m.focused {
		selector = m.styles.Cursor.Render("> ")
	}

	// Model info
	modelInfo := fmt.Sprintf("[%s]", tpl.Model)

	// Calculate available width for name
	// Fixed: selector(2) + space(1) + modelInfo length
	available := output.ContentWidth(m.Width, 4+3+len(modelInfo))
	name := output.Truncate(tpl.Name, max(available, 10))

	line := fmt.Sprintf("%s%s %s",
		selector,
		m.styles.Name.Render(name),
		m.styles.Model.Render(modelInfo),
	)

	if isSelected && m.focused {
		return m.styles.Selected.Render(line)
	}
	return line
}

func (m ListModel) buildFooter() string {
	// Show delete confirmation message if pending
	if m.pendingDelete != "" {
		return m.styles.Error.Render("Press D again to delete, Esc to cancel")
	}

	// Show status message if set
	if m.statusMsg != "" {
		return m.styles.Success.Render(m.statusMsg)
	}

	return footerKeybindings
}

// SelectedTemplate returns the currently selected template, if any.
func (m ListModel) SelectedTemplate() *config.DeviceTemplate {
	return m.selectedTemplate()
}

// Cursor returns the current cursor position.
func (m ListModel) Cursor() int {
	return m.Scroller.Cursor()
}

// TemplateCount returns the number of templates.
func (m ListModel) TemplateCount() int {
	return len(m.templates)
}

// Loading returns whether the component is loading.
func (m ListModel) Loading() bool {
	return m.loading
}

// Applying returns whether a template is being applied.
func (m ListModel) Applying() bool {
	return m.applying
}

// Error returns any error that occurred.
func (m ListModel) Error() error {
	return m.err
}

// Refresh triggers a refresh of the template list.
func (m ListModel) Refresh() (ListModel, tea.Cmd) {
	m.loading = true
	return m, tea.Batch(m.Loader.Tick(), m.loadTemplates())
}

// FooterText returns keybinding hints for the footer.
func (m ListModel) FooterText() string {
	return footerKeybindings
}
