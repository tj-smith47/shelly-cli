// Package control provides interactive device control panels for the TUI.
package control

import (
	"context"
	"image/color"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// ComponentType identifies the type of component being controlled.
type ComponentType string

// Component type constants.
const (
	TypeSwitch     ComponentType = "switch"
	TypeLight      ComponentType = "light"
	TypeRGB        ComponentType = "rgb"
	TypeCover      ComponentType = "cover"
	TypeThermostat ComponentType = "thermostat"
	TypePlugin     ComponentType = "plugin"
)

// Cover state constants.
const (
	coverStateOpening = "opening"
	coverStateClosing = "closing"
	coverStateStopped = "stopped"
)

// Action name constants.
const (
	actionToggle     = "toggle"
	actionOn         = "on"
	actionOff        = "off"
	actionBrightness = "brightness"
	actionColor      = "color"
	actionPosition   = "position"
	actionOpen       = "open"
	actionClose      = "close"
	actionStop       = "stop"
	actionCalibrate  = "calibrate"
	actionTarget     = "target"
	actionMode       = "mode"
	actionBoost      = "boost"
	actionCancelBst  = "cancel_boost"
)

// ActionMsg is sent when a control action is executed.
type ActionMsg struct {
	Device    string
	Component ComponentType
	ID        int
	Action    string
	Value     any
	Err       error
}

// RefreshMsg requests a refresh of the control panel state.
type RefreshMsg struct {
	Device    string
	Component ComponentType
	ID        int
}

// Styles for control components.
type Styles struct {
	Container   lipgloss.Style
	Title       lipgloss.Style
	Label       lipgloss.Style
	Value       lipgloss.Style
	OnState     lipgloss.Style
	OffState    lipgloss.Style
	Action      lipgloss.Style
	ActionFocus lipgloss.Style
	Divider     lipgloss.Style
	Muted       lipgloss.Style
	Power       lipgloss.Style
	Error       lipgloss.Style
	Help        lipgloss.Style
	Focused     lipgloss.Style
	Selected    lipgloss.Style
}

// DefaultStyles returns default styles for control components.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Container: lipgloss.NewStyle().
			Padding(1, 2),
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true).
			MarginBottom(1),
		Label: lipgloss.NewStyle().
			Foreground(colors.Text).
			Width(12),
		Value: lipgloss.NewStyle().
			Foreground(colors.Highlight),
		OnState: lipgloss.NewStyle().
			Foreground(colors.Online).
			Bold(true),
		OffState: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Action: lipgloss.NewStyle().
			Foreground(colors.Text).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colors.TableBorder).
			Padding(0, 2),
		ActionFocus: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colors.Highlight).
			Padding(0, 2),
		Divider: lipgloss.NewStyle().
			Foreground(colors.TableBorder),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Power: lipgloss.NewStyle().
			Foreground(colors.Warning).
			Bold(true),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Help: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
		Focused: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colors.Highlight).
			Padding(1, 2),
		Selected: lipgloss.NewStyle().
			Background(colors.Highlight).
			Foreground(colors.Primary),
	}
}

// Service defines the interface for component control operations.
type Service interface {
	// Switch operations
	SwitchOn(ctx context.Context, device string, switchID int) error
	SwitchOff(ctx context.Context, device string, switchID int) error
	SwitchToggle(ctx context.Context, device string, switchID int) error

	// Light operations
	LightOn(ctx context.Context, device string, lightID int) error
	LightOff(ctx context.Context, device string, lightID int) error
	LightToggle(ctx context.Context, device string, lightID int) error
	LightBrightness(ctx context.Context, device string, lightID, brightness int) error

	// RGB operations
	RGBOn(ctx context.Context, device string, rgbID int) error
	RGBOff(ctx context.Context, device string, rgbID int) error
	RGBToggle(ctx context.Context, device string, rgbID int) error
	RGBBrightness(ctx context.Context, device string, rgbID, brightness int) error
	RGBColor(ctx context.Context, device string, rgbID, r, g, b int) error
	RGBColorAndBrightness(ctx context.Context, device string, rgbID, r, g, b, brightness int) error

	// Cover operations
	CoverOpen(ctx context.Context, device string, coverID int, duration *int) error
	CoverClose(ctx context.Context, device string, coverID int, duration *int) error
	CoverStop(ctx context.Context, device string, coverID int) error
	CoverPosition(ctx context.Context, device string, coverID, position int) error
	CoverCalibrate(ctx context.Context, device string, coverID int) error

	// Thermostat operations
	ThermostatSetTarget(ctx context.Context, device string, thermostatID int, targetC float64) error
	ThermostatSetMode(ctx context.Context, device string, thermostatID int, mode string) error
	ThermostatBoost(ctx context.Context, device string, thermostatID int, durationSec int) error
	ThermostatCancelBoost(ctx context.Context, device string, thermostatID int) error
}

// PluginService defines the interface for plugin device control operations.
type PluginService interface {
	// PluginControl dispatches a control action to a plugin-managed device.
	PluginControl(ctx context.Context, device, action, component string, id int) error
}

// PresetColor represents a preset color option.
type PresetColor struct {
	Name  string
	Color color.Color
	RGB   [3]int
}

// PresetColors are commonly used colors for RGB controls.
var PresetColors = []PresetColor{
	{Name: "Red", RGB: [3]int{255, 0, 0}},
	{Name: "Green", RGB: [3]int{0, 255, 0}},
	{Name: "Blue", RGB: [3]int{0, 0, 255}},
	{Name: "Yellow", RGB: [3]int{255, 255, 0}},
	{Name: "Cyan", RGB: [3]int{0, 255, 255}},
	{Name: "Magenta", RGB: [3]int{255, 0, 255}},
	{Name: "Orange", RGB: [3]int{255, 128, 0}},
	{Name: "White", RGB: [3]int{255, 255, 255}},
	{Name: "Warm", RGB: [3]int{255, 200, 100}},
}

// ThermostatMode represents a thermostat operating mode.
type ThermostatMode struct {
	ID    string
	Label string
}

// ThermostatModes are the available thermostat modes.
var ThermostatModes = []ThermostatMode{
	{ID: "heat", Label: "Heat"},
	{ID: "cool", Label: "Cool"},
	{ID: "auto", Label: "Auto"},
	{ID: "off", Label: "Off"},
}

// executeAction is a helper to run an action and return an ActionMsg.
func executeAction(device string, comp ComponentType, id int, action string, fn func() error) tea.Cmd {
	return func() tea.Msg {
		err := fn()
		return ActionMsg{
			Device:    device,
			Component: comp,
			ID:        id,
			Action:    action,
			Err:       err,
		}
	}
}
