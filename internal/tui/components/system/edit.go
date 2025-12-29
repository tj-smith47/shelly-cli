// Package system provides TUI components for managing device system settings.
package system

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/form"
)

// EditField represents a field in the edit form.
type EditField int

// Edit field constants.
const (
	EditFieldName EditField = iota
	EditFieldTimezone
	EditFieldLatitude
	EditFieldLongitude
	EditFieldAliases
	EditFieldNewAlias
	EditFieldCount
)

// EditSaveResultMsg signals a save operation completed.
type EditSaveResultMsg struct {
	Err error
}

// EditOpenedMsg signals the edit modal was opened.
type EditOpenedMsg struct{}

// EditClosedMsg signals the edit modal was closed.
type EditClosedMsg struct {
	Saved bool
}

// EditModel represents the system settings edit modal.
type EditModel struct {
	ctx      context.Context
	svc      *shelly.Service
	device   string
	visible  bool
	cursor   EditField
	saving   bool
	err      error
	width    int
	height   int
	styles   EditStyles
	original *shelly.SysConfig

	// Form inputs
	nameInput        textinput.Model
	timezoneDropdown form.SearchableDropdown
	latitudeInput    textinput.Model
	longitudeInput   textinput.Model

	// Alias management
	deviceName    string   // Registered device name for alias operations
	aliases       []string // Current device aliases
	aliasCursor   int      // Selected alias index for deletion
	newAliasInput textinput.Model
}

// EditStyles holds styles for the edit modal.
type EditStyles struct {
	Overlay     lipgloss.Style
	Modal       lipgloss.Style
	Title       lipgloss.Style
	Label       lipgloss.Style
	LabelFocus  lipgloss.Style
	Input       lipgloss.Style
	InputFocus  lipgloss.Style
	Button      lipgloss.Style
	ButtonFocus lipgloss.Style
	Error       lipgloss.Style
	Help        lipgloss.Style
	Selector    lipgloss.Style
	AliasItem   lipgloss.Style
	AliasSelect lipgloss.Style
}

// DefaultEditStyles returns the default edit modal styles.
func DefaultEditStyles() EditStyles {
	colors := theme.GetSemanticColors()
	return EditStyles{
		Overlay: lipgloss.NewStyle().
			Background(lipgloss.Color("#000000")),
		Modal: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colors.TableBorder).
			Background(colors.Background).
			Padding(1, 2),
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true).
			MarginBottom(1),
		Label: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Width(12),
		LabelFocus: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true).
			Width(12),
		Input: lipgloss.NewStyle().
			Foreground(colors.Text),
		InputFocus: lipgloss.NewStyle().
			Foreground(colors.Highlight),
		Button: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Padding(0, 2),
		ButtonFocus: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true).
			Padding(0, 2),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Help: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Selector: lipgloss.NewStyle().
			Foreground(colors.Highlight),
		AliasItem: lipgloss.NewStyle().
			Foreground(colors.Text),
		AliasSelect: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
	}
}

// NewEditModel creates a new system settings edit modal.
func NewEditModel(ctx context.Context, svc *shelly.Service) EditModel {
	colors := theme.GetSemanticColors()

	// Create input styles
	inputStyles := textinput.Styles{}
	inputStyles.Focused.Text = inputStyles.Focused.Text.Foreground(colors.Highlight)
	inputStyles.Focused.Placeholder = inputStyles.Focused.Placeholder.Foreground(colors.Muted)
	inputStyles.Blurred.Text = inputStyles.Blurred.Text.Foreground(colors.Text)
	inputStyles.Blurred.Placeholder = inputStyles.Blurred.Placeholder.Foreground(colors.Muted)

	nameInput := textinput.New()
	nameInput.Placeholder = "Device name"
	nameInput.CharLimit = 64
	nameInput.SetWidth(35)
	nameInput.SetStyles(inputStyles)

	timezoneDropdown := form.NewSearchableDropdown().
		SetOptions(CommonTimezones).
		SetHelp("Type to filter, Enter to select").
		SetMaxVisible(10)

	latitudeInput := textinput.New()
	latitudeInput.Placeholder = "e.g., 40.7128"
	latitudeInput.CharLimit = 20
	latitudeInput.SetWidth(15)
	latitudeInput.SetStyles(inputStyles)

	longitudeInput := textinput.New()
	longitudeInput.Placeholder = "e.g., -74.0060"
	longitudeInput.CharLimit = 20
	longitudeInput.SetWidth(15)
	longitudeInput.SetStyles(inputStyles)

	newAliasInput := textinput.New()
	newAliasInput.Placeholder = "New alias"
	newAliasInput.CharLimit = 32
	newAliasInput.SetWidth(25)
	newAliasInput.SetStyles(inputStyles)

	return EditModel{
		ctx:              ctx,
		svc:              svc,
		styles:           DefaultEditStyles(),
		nameInput:        nameInput,
		timezoneDropdown: timezoneDropdown,
		latitudeInput:    latitudeInput,
		longitudeInput:   longitudeInput,
		newAliasInput:    newAliasInput,
	}
}

// Show displays the edit modal with the given device and config.
func (m EditModel) Show(device string, sysConfig *shelly.SysConfig) EditModel {
	return m.showAt(device, sysConfig, EditFieldName)
}

// ShowAtTimezone displays the edit modal focused on the timezone field.
func (m EditModel) ShowAtTimezone(device string, sysConfig *shelly.SysConfig) EditModel {
	return m.showAt(device, sysConfig, EditFieldTimezone)
}

func (m EditModel) showAt(device string, sysConfig *shelly.SysConfig, initialField EditField) EditModel {
	m.device = device
	m.visible = true
	m.cursor = initialField
	m.saving = false
	m.err = nil
	m.original = sysConfig

	// Initialize inputs with current values
	m = m.initializeInputs(sysConfig)

	// Load aliases for the device
	m = m.loadAliases()

	// Blur all fields first
	m.nameInput.Blur()
	m.timezoneDropdown = m.timezoneDropdown.Blur()
	m.newAliasInput.Blur()

	// Focus the initial field
	m = m.focusCurrentField()

	return m
}

// loadAliases loads the device's aliases from config.
func (m EditModel) loadAliases() EditModel {
	// Try to resolve the device to get its registered name
	if dev, err := config.ResolveDevice(m.device); err == nil {
		m.deviceName = dev.Name
		m.aliases = dev.Aliases
	} else {
		// Device not registered, no aliases available
		m.deviceName = ""
		m.aliases = nil
	}
	m.aliasCursor = 0
	m.newAliasInput.SetValue("")
	return m
}

// Hide hides the edit modal.
func (m EditModel) Hide() EditModel {
	m.visible = false
	m.nameInput.Blur()
	m.timezoneDropdown = m.timezoneDropdown.Blur()
	m.latitudeInput.Blur()
	m.longitudeInput.Blur()
	m.newAliasInput.Blur()
	return m
}

// Visible returns whether the modal is visible.
func (m EditModel) Visible() bool {
	return m.visible
}

// SetSize sets the modal dimensions.
func (m EditModel) SetSize(width, height int) EditModel {
	m.width = width
	m.height = height
	return m
}

// Init returns the initial command.
func (m EditModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m EditModel) Update(msg tea.Msg) (EditModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case EditSaveResultMsg:
		m.saving = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		// Success - close modal
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: true} }

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m EditModel) handleKey(msg tea.KeyPressMsg) (EditModel, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc":
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }

	case "enter":
		return m.handleEnter()

	case "tab":
		return m.nextField(), nil

	case "shift+tab":
		return m.prevField(), nil

	case "down", "j":
		return m.handleDown()

	case "up", "k":
		return m.handleUp()

	case "d", "delete", "backspace":
		if m.cursor == EditFieldAliases {
			return m.deleteSelectedAlias()
		}
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m EditModel) handleEnter() (EditModel, tea.Cmd) {
	if m.saving {
		return m, nil
	}
	if m.cursor == EditFieldNewAlias {
		return m.addAlias()
	}
	return m.save()
}

func (m EditModel) handleDown() (EditModel, tea.Cmd) {
	// Navigate within alias list when focused
	if m.cursor == EditFieldAliases && len(m.aliases) > 0 {
		if m.aliasCursor < len(m.aliases)-1 {
			m.aliasCursor++
		}
		return m, nil
	}
	return m.nextField(), nil
}

func (m EditModel) handleUp() (EditModel, tea.Cmd) {
	// Navigate within alias list when focused
	if m.cursor == EditFieldAliases && len(m.aliases) > 0 {
		if m.aliasCursor > 0 {
			m.aliasCursor--
		}
		return m, nil
	}
	return m.prevField(), nil
}

func (m EditModel) updateFocusedInput(msg tea.Msg) (EditModel, tea.Cmd) {
	var cmd tea.Cmd

	switch m.cursor {
	case EditFieldName:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case EditFieldTimezone:
		m.timezoneDropdown, cmd = m.timezoneDropdown.Update(msg)
	case EditFieldLatitude:
		m.latitudeInput, cmd = m.latitudeInput.Update(msg)
	case EditFieldLongitude:
		m.longitudeInput, cmd = m.longitudeInput.Update(msg)
	case EditFieldAliases:
		// Alias list navigation handled in handleKey
	case EditFieldNewAlias:
		m.newAliasInput, cmd = m.newAliasInput.Update(msg)
	case EditFieldCount:
		// No input to update
	}

	return m, cmd
}

func (m EditModel) nextField() EditModel {
	// Blur current field
	m = m.blurCurrentField()

	// Move to next
	if m.cursor < EditFieldCount-1 {
		m.cursor++
	}

	// Focus new field
	m = m.focusCurrentField()

	return m
}

func (m EditModel) prevField() EditModel {
	// Blur current field
	m = m.blurCurrentField()

	// Move to previous
	if m.cursor > 0 {
		m.cursor--
	}

	// Focus new field
	m = m.focusCurrentField()

	return m
}

func (m EditModel) blurCurrentField() EditModel {
	switch m.cursor {
	case EditFieldName:
		m.nameInput.Blur()
	case EditFieldTimezone:
		m.timezoneDropdown = m.timezoneDropdown.Blur()
	case EditFieldLatitude:
		m.latitudeInput.Blur()
	case EditFieldLongitude:
		m.longitudeInput.Blur()
	case EditFieldAliases:
		// No blur needed for alias list
	case EditFieldNewAlias:
		m.newAliasInput.Blur()
	case EditFieldCount:
		// No input
	}
	return m
}

func (m EditModel) focusCurrentField() EditModel {
	switch m.cursor {
	case EditFieldName:
		m.nameInput.Focus()
	case EditFieldTimezone:
		m.timezoneDropdown = m.timezoneDropdown.Focus()
	case EditFieldLatitude:
		m.latitudeInput.Focus()
	case EditFieldLongitude:
		m.longitudeInput.Focus()
	case EditFieldAliases:
		// Focus indicator is on the alias list
	case EditFieldNewAlias:
		m.newAliasInput.Focus()
	case EditFieldCount:
		// No input
	}
	return m
}

// saveFormData holds the validated form data for saving.
type saveFormData struct {
	name        string
	timezone    string
	lat         float64
	lng         float64
	hasLocation bool
}

func (m EditModel) save() (EditModel, tea.Cmd) {
	data, err := m.validateAndCollectFormData()
	if err != nil {
		m.err = err
		return m, nil
	}

	m.saving = true
	m.err = nil

	return m, m.createSaveCmd(data)
}

// addAlias validates and adds a new alias to the device.
func (m EditModel) addAlias() (EditModel, tea.Cmd) {
	alias := strings.TrimSpace(m.newAliasInput.Value())
	if alias == "" {
		return m, nil
	}

	// Device must be registered for aliases
	if m.deviceName == "" {
		m.err = fmt.Errorf("device not registered, cannot add aliases")
		return m, nil
	}

	// Validate alias format
	if err := config.ValidateDeviceAlias(alias); err != nil {
		m.err = err
		return m, nil
	}

	// Check for conflicts
	if err := config.CheckAliasConflict(alias, m.deviceName); err != nil {
		m.err = err
		return m, nil
	}

	// Add the alias
	if err := config.AddDeviceAlias(m.deviceName, alias); err != nil {
		m.err = err
		return m, nil
	}

	// Reload aliases and clear input
	m.aliases = append(m.aliases, alias)
	m.newAliasInput.SetValue("")
	m.err = nil

	return m, nil
}

// deleteSelectedAlias removes the currently selected alias.
func (m EditModel) deleteSelectedAlias() (EditModel, tea.Cmd) {
	if m.deviceName == "" || len(m.aliases) == 0 {
		return m, nil
	}

	if m.aliasCursor >= len(m.aliases) {
		return m, nil
	}

	alias := m.aliases[m.aliasCursor]
	if err := config.RemoveDeviceAlias(m.deviceName, alias); err != nil {
		m.err = err
		return m, nil
	}

	// Remove from local list
	m.aliases = append(m.aliases[:m.aliasCursor], m.aliases[m.aliasCursor+1:]...)

	// Adjust cursor if needed
	if m.aliasCursor >= len(m.aliases) && m.aliasCursor > 0 {
		m.aliasCursor--
	}
	m.err = nil

	return m, nil
}

func (m EditModel) validateAndCollectFormData() (saveFormData, error) {
	var data saveFormData

	data.name = strings.TrimSpace(m.nameInput.Value())
	data.timezone = m.timezoneDropdown.SelectedValue()
	latStr := strings.TrimSpace(m.latitudeInput.Value())
	lngStr := strings.TrimSpace(m.longitudeInput.Value())

	// Validate name is required
	if data.name == "" {
		return data, fmt.Errorf("device name is required")
	}

	// Parse and validate location if provided
	if latStr != "" || lngStr != "" {
		lat, lng, err := parseLocation(latStr, lngStr)
		if err != nil {
			return data, err
		}
		data.lat = lat
		data.lng = lng
		data.hasLocation = true
	}

	return data, nil
}

func parseLocation(latStr, lngStr string) (lat, lng float64, err error) {
	if latStr != "" {
		lat, err = strconv.ParseFloat(latStr, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid latitude: %w", err)
		}
		if lat < -90 || lat > 90 {
			return 0, 0, fmt.Errorf("latitude must be between -90 and 90")
		}
	}

	if lngStr != "" {
		lng, err = strconv.ParseFloat(lngStr, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid longitude: %w", err)
		}
		if lng < -180 || lng > 180 {
			return 0, 0, fmt.Errorf("longitude must be between -180 and 180")
		}
	}

	return lat, lng, nil
}

func (m EditModel) createSaveCmd(data saveFormData) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		var lastErr error

		// Update name if changed
		if m.original == nil || data.name != m.original.Name {
			if err := m.svc.SetSysName(ctx, m.device, data.name); err != nil {
				lastErr = err
			}
		}

		// Update timezone if changed
		if m.original == nil || data.timezone != m.original.Timezone {
			if err := m.svc.SetSysTimezone(ctx, m.device, data.timezone); err != nil {
				lastErr = err
			}
		}

		// Update location if changed
		if data.hasLocation && (m.original == nil || data.lat != m.original.Lat || data.lng != m.original.Lng) {
			if err := m.svc.SetSysLocation(ctx, m.device, data.lat, data.lng); err != nil {
				lastErr = err
			}
		}

		return EditSaveResultMsg{Err: lastErr}
	}
}

// formatFloat formats a float64 for display in the text input.
func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func (m EditModel) initializeInputs(sysConfig *shelly.SysConfig) EditModel {
	if sysConfig == nil {
		m.nameInput.SetValue("")
		m.timezoneDropdown = m.timezoneDropdown.SetSelected("")
		m.latitudeInput.SetValue("")
		m.longitudeInput.SetValue("")
		return m
	}

	m.nameInput.SetValue(sysConfig.Name)
	m.timezoneDropdown = m.timezoneDropdown.SetSelected(sysConfig.Timezone)
	m.latitudeInput.SetValue(formatCoord(sysConfig.Lat))
	m.longitudeInput.SetValue(formatCoord(sysConfig.Lng))
	return m
}

func formatCoord(val float64) string {
	if val == 0 {
		return ""
	}
	return formatFloat(val)
}

// View renders the edit modal.
func (m EditModel) View() string {
	if !m.visible {
		return ""
	}

	var content strings.Builder

	// Title
	content.WriteString(m.styles.Title.Render("Edit System Settings"))
	content.WriteString("\n\n")

	// Name field
	content.WriteString(m.renderField(EditFieldName, "Name:", m.nameInput.View()))
	content.WriteString("\n")

	// Timezone field
	content.WriteString(m.renderField(EditFieldTimezone, "Timezone:", m.timezoneDropdown.View()))
	content.WriteString("\n")

	// Location fields
	content.WriteString(m.renderField(EditFieldLatitude, "Latitude:", m.latitudeInput.View()))
	content.WriteString("\n")
	content.WriteString(m.renderField(EditFieldLongitude, "Longitude:", m.longitudeInput.View()))
	content.WriteString("\n\n")

	// Alias section (only show if device is registered)
	if m.deviceName != "" {
		content.WriteString(m.styles.Title.Render("Device Aliases"))
		content.WriteString("\n")
		content.WriteString(m.renderAliasSection())
		content.WriteString("\n")
	}

	// Error display
	if m.err != nil {
		content.WriteString(m.styles.Error.Render("Error: " + m.err.Error()))
		content.WriteString("\n\n")
	}

	// Status/buttons
	if m.saving {
		content.WriteString(m.styles.Help.Render("Saving..."))
	} else {
		helpText := "Enter: Save | Esc: Cancel | Tab: Next field"
		if m.cursor == EditFieldAliases && len(m.aliases) > 0 {
			helpText = "d: Delete | Tab: Next | Esc: Cancel"
		} else if m.cursor == EditFieldNewAlias {
			helpText = "Enter: Add alias | Tab: Next | Esc: Cancel"
		}
		content.WriteString(m.styles.Help.Render(helpText))
	}

	// Render modal box
	modalContent := content.String()
	modalWidth := min(60, m.width-4)
	modal := m.styles.Modal.Width(modalWidth).Render(modalContent)

	// Center the modal
	return m.centerModal(modal)
}

func (m EditModel) renderField(field EditField, label, input string) string {
	var selector, labelStr string

	if m.cursor == field {
		selector = m.styles.Selector.Render("▶ ")
		labelStr = m.styles.LabelFocus.Render(label)
	} else {
		selector = "  "
		labelStr = m.styles.Label.Render(label)
	}

	return selector + labelStr + " " + input
}

func (m EditModel) renderAliasSection() string {
	var content strings.Builder

	// Current aliases
	content.WriteString(m.renderAliasesList())
	content.WriteString("\n")

	// New alias input
	content.WriteString(m.renderField(EditFieldNewAlias, "Add alias:", m.newAliasInput.View()))

	return content.String()
}

func (m EditModel) renderAliasesList() string {
	var content strings.Builder

	if len(m.aliases) == 0 {
		prefix := "  "
		if m.cursor == EditFieldAliases {
			prefix = m.styles.Selector.Render("▶ ")
		}
		content.WriteString(prefix)
		content.WriteString(m.styles.Help.Render("(no aliases)"))
		return content.String()
	}

	for i, alias := range m.aliases {
		isSelected := m.cursor == EditFieldAliases && i == m.aliasCursor
		prefix := "  "
		if isSelected {
			prefix = m.styles.Selector.Render("▶ ")
		}
		content.WriteString(prefix)
		if isSelected {
			content.WriteString(m.styles.AliasSelect.Render(alias))
		} else {
			content.WriteString(m.styles.AliasItem.Render(alias))
		}
		if i < len(m.aliases)-1 {
			content.WriteString("\n")
		}
	}

	return content.String()
}

func (m EditModel) centerModal(modal string) string {
	lines := strings.Split(modal, "\n")
	modalHeight := len(lines)
	modalWidth := 0
	for _, line := range lines {
		if lipgloss.Width(line) > modalWidth {
			modalWidth = lipgloss.Width(line)
		}
	}

	// Calculate centering
	topPad := (m.height - modalHeight) / 2
	leftPad := (m.width - modalWidth) / 2

	if topPad < 0 {
		topPad = 0
	}
	if leftPad < 0 {
		leftPad = 0
	}

	// Build centered output
	var result strings.Builder
	for range topPad {
		result.WriteString("\n")
	}

	padding := strings.Repeat(" ", leftPad)
	for _, line := range lines {
		result.WriteString(padding)
		result.WriteString(line)
		result.WriteString("\n")
	}

	return result.String()
}
