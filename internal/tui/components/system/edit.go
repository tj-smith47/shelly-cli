// Package system provides TUI components for managing device system settings.
package system

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/editmodal"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/form"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
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

// EditSaveResultMsg is an alias for the shared save result message.
type EditSaveResultMsg = messages.SaveResultMsg

// EditOpenedMsg is an alias for the shared edit opened message.
type EditOpenedMsg = messages.EditOpenedMsg

// EditClosedMsg is an alias for the shared edit closed message.
type EditClosedMsg = messages.EditClosedMsg

// EditModel represents the system settings edit modal.
type EditModel struct {
	editmodal.Base

	original *shelly.SysConfig

	// Form inputs
	nameInput        textinput.Model
	timezoneDropdown form.Select
	latitudeInput    textinput.Model
	longitudeInput   textinput.Model

	// Alias management
	deviceName    string   // Registered device name for alias operations
	aliases       []string // Current device aliases
	aliasCursor   int      // Selected alias index for deletion
	newAliasInput textinput.Model
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

	timezoneDropdown := form.NewSelect(
		form.WithSelectOptions(CommonTimezones),
		form.WithSelectHelp("Type to filter, Enter to select"),
		form.WithSelectMaxVisible(10),
		form.WithSelectFiltering(true),
	)

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
		Base: editmodal.Base{
			Ctx:    ctx,
			Svc:    svc,
			Styles: editmodal.DefaultStyles().WithLabelWidth(12),
		},
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
	m.Base.Show(device, int(EditFieldCount))
	m.SetCursor(int(initialField))
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
	if dev, err := config.ResolveDevice(m.Device); err == nil {
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
	m.Base.Hide()
	m.nameInput.Blur()
	m.timezoneDropdown = m.timezoneDropdown.Blur()
	m.latitudeInput.Blur()
	m.longitudeInput.Blur()
	m.newAliasInput.Blur()
	return m
}

// Visible returns whether the modal is visible.
func (m EditModel) Visible() bool {
	return m.Base.Visible()
}

// SetSize sets the modal dimensions.
func (m EditModel) SetSize(width, height int) EditModel {
	m.Base.SetSize(width, height)
	return m
}

// Init returns the initial command.
func (m EditModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m EditModel) Update(msg tea.Msg) (EditModel, tea.Cmd) {
	if !m.Visible() {
		return m, nil
	}

	return m.handleMessage(msg)
}

func (m EditModel) handleMessage(msg tea.Msg) (EditModel, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.SaveResultMsg:
		_, cmd := m.HandleSaveResult(msg)
		return m, cmd

	// Action messages from context system
	case messages.NavigationMsg:
		return m.handleNavigation(msg)
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m EditModel) handleNavigation(msg messages.NavigationMsg) (EditModel, tea.Cmd) {
	action := m.HandleNavigation(msg)
	switch action {
	case editmodal.ActionNavUp:
		return m.handleUp()
	case editmodal.ActionNavDown:
		return m.handleDown()
	default:
		return m, nil
	}
}

func (m EditModel) handleKey(msg tea.KeyPressMsg) (EditModel, tea.Cmd) {
	if m.Saving {
		return m, nil
	}

	switch msg.String() {
	case keyconst.KeyEsc, "ctrl+[":
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }

	case keyconst.KeyCtrlS:
		return m.save()

	case keyconst.KeyEnter:
		return m.handleEnter()

	case keyconst.KeyTab:
		return m.nextField(), nil

	case keyconst.KeyShiftTab:
		return m.prevField(), nil

	case "d", "delete", "backspace":
		if EditField(m.Cursor) == EditFieldAliases {
			return m.deleteSelectedAlias()
		}
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m EditModel) handleEnter() (EditModel, tea.Cmd) {
	if m.Saving {
		return m, nil
	}
	if EditField(m.Cursor) == EditFieldNewAlias {
		return m.addAlias()
	}
	// Enter does NOT trigger save; use Ctrl+S for save
	return m.updateFocusedInput(tea.KeyPressMsg{Code: tea.KeyEnter})
}

func (m EditModel) handleDown() (EditModel, tea.Cmd) {
	// Navigate within alias list when focused
	if EditField(m.Cursor) == EditFieldAliases && len(m.aliases) > 0 {
		if m.aliasCursor < len(m.aliases)-1 {
			m.aliasCursor++
		}
		return m, nil
	}
	return m.nextField(), nil
}

func (m EditModel) handleUp() (EditModel, tea.Cmd) {
	// Navigate within alias list when focused
	if EditField(m.Cursor) == EditFieldAliases && len(m.aliases) > 0 {
		if m.aliasCursor > 0 {
			m.aliasCursor--
		}
		return m, nil
	}
	return m.prevField(), nil
}

func (m EditModel) updateFocusedInput(msg tea.Msg) (EditModel, tea.Cmd) {
	var cmd tea.Cmd

	switch EditField(m.Cursor) {
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

	// Move to next (custom navigation because of alias list)
	if m.Cursor < int(EditFieldCount)-1 {
		m.Cursor++
	}

	// Focus new field
	m = m.focusCurrentField()

	return m
}

func (m EditModel) prevField() EditModel {
	// Blur current field
	m = m.blurCurrentField()

	// Move to previous
	if m.Cursor > 0 {
		m.Cursor--
	}

	// Focus new field
	m = m.focusCurrentField()

	return m
}

func (m EditModel) blurCurrentField() EditModel {
	switch EditField(m.Cursor) {
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
	switch EditField(m.Cursor) {
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
		m.Err = err
		return m, nil
	}

	m.StartSave()

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
		m.Err = fmt.Errorf("device not registered, cannot add aliases")
		return m, nil
	}

	// Validate alias format
	if err := config.ValidateDeviceAlias(alias); err != nil {
		m.Err = err
		return m, nil
	}

	// Check for conflicts
	if err := config.CheckAliasConflict(alias, m.deviceName); err != nil {
		m.Err = err
		return m, nil
	}

	// Add the alias
	if err := config.AddDeviceAlias(m.deviceName, alias); err != nil {
		m.Err = err
		return m, nil
	}

	// Reload aliases and clear input
	m.aliases = append(m.aliases, alias)
	m.newAliasInput.SetValue("")
	m.Err = nil

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
		m.Err = err
		return m, nil
	}

	// Remove from local list
	m.aliases = append(m.aliases[:m.aliasCursor], m.aliases[m.aliasCursor+1:]...)

	// Adjust cursor if needed
	if m.aliasCursor >= len(m.aliases) && m.aliasCursor > 0 {
		m.aliasCursor--
	}
	m.Err = nil

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
	original := m.original
	device := m.Device
	svc := m.Svc

	return m.SaveCmd(func(ctx context.Context) error {
		var lastErr error

		// Update name if changed
		if original == nil || data.name != original.Name {
			if err := svc.SetSysName(ctx, device, data.name); err != nil {
				lastErr = err
			}
		}

		// Update timezone if changed
		if original == nil || data.timezone != original.Timezone {
			if err := svc.SetSysTimezone(ctx, device, data.timezone); err != nil {
				lastErr = err
			}
		}

		// Update location if changed
		if data.hasLocation && (original == nil || data.lat != original.Lat || data.lng != original.Lng) {
			if err := svc.SetSysLocation(ctx, device, data.lat, data.lng); err != nil {
				lastErr = err
			}
		}

		return lastErr
	})
}

// formatFloat formats a float64 for display in the text input.
func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func (m EditModel) initializeInputs(sysConfig *shelly.SysConfig) EditModel {
	if sysConfig == nil {
		m.nameInput.SetValue("")
		m.timezoneDropdown = m.timezoneDropdown.SetSelectedValue("")
		m.latitudeInput.SetValue("")
		m.longitudeInput.SetValue("")
		return m
	}

	m.nameInput.SetValue(sysConfig.Name)
	m.timezoneDropdown = m.timezoneDropdown.SetSelectedValue(sysConfig.Timezone)
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
	if !m.Visible() {
		return ""
	}

	// Build footer based on context
	var normalFooter string
	switch {
	case EditField(m.Cursor) == EditFieldAliases && len(m.aliases) > 0:
		normalFooter = "d: Delete | Tab: Next | Esc: Cancel"
	case EditField(m.Cursor) == EditFieldNewAlias:
		normalFooter = "Enter: Add alias | Tab: Next | Esc: Cancel"
	default:
		normalFooter = "Ctrl+S: Save | Esc: Cancel | Tab: Next field"
	}
	footer := m.RenderSavingFooter(normalFooter)

	// Build content
	var content strings.Builder

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
		content.WriteString(m.Styles.Title.Render("Device Aliases"))
		content.WriteString("\n")
		content.WriteString(m.renderAliasSection())
		content.WriteString("\n")
	}

	// Error display
	if errStr := m.RenderError(); errStr != "" {
		content.WriteString(errStr)
	}

	return m.RenderModal("Edit System Settings", content.String(), footer)
}

func (m EditModel) renderField(field EditField, label, input string) string {
	var selector, labelStr string

	if EditField(m.Cursor) == field {
		selector = m.Styles.Selector.Render("▶ ")
		labelStr = m.Styles.LabelFocus.Render(label)
	} else {
		selector = "  "
		labelStr = m.Styles.Label.Render(label)
	}

	prefix := selector + labelStr + " "

	// Handle multi-line inputs by indenting subsequent lines
	lines := strings.Split(input, "\n")
	if len(lines) <= 1 {
		return prefix + input
	}

	// Calculate indent for subsequent lines (selector width + label width + space)
	indent := strings.Repeat(" ", 2+12+1) // 2 for selector, 12 for label width, 1 for space

	var result strings.Builder
	result.WriteString(prefix + lines[0])
	for i := 1; i < len(lines); i++ {
		result.WriteString("\n")
		result.WriteString(indent + lines[i])
	}
	return result.String()
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
		if EditField(m.Cursor) == EditFieldAliases {
			prefix = m.Styles.Selector.Render("▶ ")
		}
		content.WriteString(prefix)
		content.WriteString(m.Styles.Help.Render("(no aliases)"))
		return content.String()
	}

	for i, alias := range m.aliases {
		isSelected := EditField(m.Cursor) == EditFieldAliases && i == m.aliasCursor
		prefix := "  "
		if isSelected {
			prefix = m.Styles.Selector.Render("▶ ")
		}
		content.WriteString(prefix)
		if isSelected {
			content.WriteString(m.Styles.Selected.Render(alias))
		} else {
			content.WriteString(m.Styles.Value.Render(alias))
		}
		if i < len(m.aliases)-1 {
			content.WriteString("\n")
		}
	}

	return content.String()
}
