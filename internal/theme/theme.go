// Package theme provides theming support for the CLI using bubbletint v2.
// Themes apply to ALL CLI output, not just the TUI dashboard.
//
// This package wraps bubbletint v2 to provide:
// - 280+ terminal color themes
// - Theme-aware lipgloss styles for rich text formatting
// - Pre-rendered styled strings for common UI elements
package theme

import (
	"image/color"
	"sync"

	"charm.land/lipgloss/v2"
	tint "github.com/lrstanley/bubbletint/v2"
	"github.com/spf13/viper"
)

// DefaultTheme is the default color theme for the CLI.
const DefaultTheme = "dracula"

var initOnce sync.Once

// init initializes the theme system with all available themes.
func init() {
	initOnce.Do(func() {
		// Initialize global registry with all 280+ default themes
		tint.NewDefaultRegistry()
		// Set default theme
		SetTheme(DefaultTheme)
	})
}

// =============================================================================
// Theme Management
// =============================================================================

// Current returns the current theme tint.
func Current() *tint.Tint {
	return tint.DefaultRegistry.Current()
}

// SetTheme sets the current theme by name.
func SetTheme(name string) bool {
	return tint.SetTintID(name)
}

// SetThemeFromConfig sets the theme from viper configuration.
func SetThemeFromConfig() {
	name := viper.GetString("theme")
	if name == "" {
		name = DefaultTheme
	}
	SetTheme(name)
}

// ListThemes returns all available theme IDs (280+).
func ListThemes() []string {
	return tint.TintIDs()
}

// GetTheme returns a tint by ID.
func GetTheme(id string) (*tint.Tint, bool) {
	return tint.DefaultRegistry.GetTint(id)
}

// NextTheme cycles to the next theme.
func NextTheme() {
	tint.NextTint()
}

// PrevTheme cycles to the previous theme.
func PrevTheme() {
	tint.PreviousTint()
}

// =============================================================================
// Color Accessors - Return color.Color (compatible with lipgloss v2)
// =============================================================================

// defaultColor returns a fallback color when the theme color is nil.
func defaultColor(c, fallback *tint.Color) color.Color {
	if c != nil {
		return c
	}
	return fallback
}

// Dracula fallback colors.
var (
	draculaFg        = &tint.Color{R: 248, G: 248, B: 242, A: 255}
	draculaBg        = &tint.Color{R: 40, G: 42, B: 54, A: 255}
	draculaGreen     = &tint.Color{R: 80, G: 250, B: 123, A: 255}
	draculaRed       = &tint.Color{R: 255, G: 85, B: 85, A: 255}
	draculaYellow    = &tint.Color{R: 241, G: 250, B: 140, A: 255}
	draculaBlue      = &tint.Color{R: 98, G: 114, B: 164, A: 255}
	draculaCyan      = &tint.Color{R: 139, G: 233, B: 253, A: 255}
	draculaPurple    = &tint.Color{R: 189, G: 147, B: 249, A: 255}
	draculaBrightBlk = &tint.Color{R: 98, G: 114, B: 164, A: 255}
)

// Fg returns the current theme foreground color.
func Fg() color.Color {
	t := Current()
	if t == nil {
		return draculaFg
	}
	return defaultColor(t.Fg, draculaFg)
}

// Bg returns the current theme background color.
func Bg() color.Color {
	t := Current()
	if t == nil {
		return draculaBg
	}
	return defaultColor(t.Bg, draculaBg)
}

// Green returns the current theme green color.
func Green() color.Color {
	t := Current()
	if t == nil {
		return draculaGreen
	}
	return defaultColor(t.Green, draculaGreen)
}

// Red returns the current theme red color.
func Red() color.Color {
	t := Current()
	if t == nil {
		return draculaRed
	}
	return defaultColor(t.Red, draculaRed)
}

// Yellow returns the current theme yellow color.
func Yellow() color.Color {
	t := Current()
	if t == nil {
		return draculaYellow
	}
	return defaultColor(t.Yellow, draculaYellow)
}

// Blue returns the current theme blue color.
func Blue() color.Color {
	t := Current()
	if t == nil {
		return draculaBlue
	}
	return defaultColor(t.Blue, draculaBlue)
}

// Cyan returns the current theme cyan color.
func Cyan() color.Color {
	t := Current()
	if t == nil {
		return draculaCyan
	}
	return defaultColor(t.Cyan, draculaCyan)
}

// Purple returns the current theme purple color.
func Purple() color.Color {
	t := Current()
	if t == nil {
		return draculaPurple
	}
	return defaultColor(t.Purple, draculaPurple)
}

// Magenta is an alias for Purple for compatibility.
func Magenta() color.Color {
	return Purple()
}

// BrightBlack returns the current theme bright black (gray) color.
func BrightBlack() color.Color {
	t := Current()
	if t == nil {
		return draculaBrightBlk
	}
	return defaultColor(t.BrightBlack, draculaBrightBlk)
}

// =============================================================================
// Style Functions - Return lipgloss.Style for rich text formatting
// =============================================================================

// StatusOK returns a style for success/ok status.
func StatusOK() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(Green())
}

// StatusWarn returns a style for warning status.
func StatusWarn() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(Yellow())
}

// StatusError returns a style for error status.
func StatusError() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(Red())
}

// StatusInfo returns a style for info status.
func StatusInfo() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(Blue())
}

// StatusOnline returns a style for online status.
func StatusOnline() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(Green())
}

// StatusOffline returns a style for offline status.
func StatusOffline() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(Red())
}

// StatusUpdating returns a style for updating status.
func StatusUpdating() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(Yellow())
}

// Bold returns a bold style with the theme foreground.
func Bold() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(Fg())
}

// Dim returns a dimmed style.
func Dim() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(BrightBlack())
}

// Highlight returns a highlighted style.
func Highlight() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(Cyan())
}

// Title returns a style for titles.
func Title() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(Purple())
}

// Subtitle returns a style for subtitles.
func Subtitle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(BrightBlack())
}

// Link returns a style for links/URLs.
func Link() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(Blue()).Underline(true)
}

// Code returns a style for code/commands.
func Code() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(Green())
}

// =============================================================================
// Rendered Functions - Return pre-styled strings for direct output
// =============================================================================

// DeviceOnline returns styled text for online devices.
func DeviceOnline() string {
	return StatusOnline().Render("● online")
}

// DeviceOffline returns styled text for offline devices.
func DeviceOffline() string {
	return StatusOffline().Render("○ offline")
}

// DeviceUpdating returns styled text for updating devices.
func DeviceUpdating() string {
	return StatusUpdating().Render("◐ updating")
}

// SwitchOn returns styled text for switch on state.
func SwitchOn() string {
	return StatusOK().Render("ON")
}

// SwitchOff returns styled text for switch off state.
func SwitchOff() string {
	return StatusOffline().Render("OFF")
}

// FormatPower formats power value with appropriate styling.
func FormatPower(watts float64) string {
	style := lipgloss.NewStyle().Foreground(Cyan())
	return style.Render(formatFloat(watts) + "W")
}

// FormatEnergy formats energy value with appropriate styling.
func FormatEnergy(wh float64) string {
	style := lipgloss.NewStyle().Foreground(Blue())
	if wh >= 1000 {
		return style.Render(formatFloat(wh/1000) + "kWh")
	}
	return style.Render(formatFloat(wh) + "Wh")
}

// =============================================================================
// Utility Functions
// =============================================================================

// formatFloat formats a float with appropriate precision.
func formatFloat(f float64) string {
	if f == float64(int(f)) {
		return intToStr(int(f))
	}
	i := int(f)
	frac := int((f - float64(i)) * 100)
	if frac == 0 {
		return intToStr(i)
	}
	return intToStr(i) + "." + intToStr(frac)
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + intToStr(-n)
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
