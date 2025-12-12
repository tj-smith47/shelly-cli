// Package theme provides theming support for the CLI using bubbletint.
// Themes apply to ALL CLI output, not just the TUI dashboard.
package theme

import (
	"sync"

	"github.com/charmbracelet/lipgloss"
	tint "github.com/lrstanley/bubbletint"
	"github.com/spf13/viper"
)

const DefaultTheme = "dracula"

var (
	initOnce sync.Once
)

// init initializes the theme system.
func init() {
	initOnce.Do(func() {
		// Initialize default registry
		tint.NewDefaultRegistry()
		// Set default theme
		SetTheme(DefaultTheme)
	})
}

// Current returns the current theme tint.
func Current() tint.Tint {
	return tint.GetCurrentTint()
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

// ListThemes returns all available theme IDs.
func ListThemes() []string {
	return tint.TintIDs()
}

// GetTheme returns a tint by ID.
func GetTheme(id string) (tint.Tint, bool) {
	return tint.GetTint(id)
}

// NextTheme cycles to the next theme.
func NextTheme() {
	tint.NextTint()
}

// PrevTheme cycles to the previous theme.
func PrevTheme() {
	tint.PreviousTint()
}

// Styled Components using current theme

// StatusOK returns a style for success/ok status.
func StatusOK() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(tint.Green())
}

// StatusWarn returns a style for warning status.
func StatusWarn() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(tint.Yellow())
}

// StatusError returns a style for error status.
func StatusError() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(tint.Red())
}

// StatusInfo returns a style for info status.
func StatusInfo() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(tint.Blue())
}

// StatusOnline returns a style for online status.
func StatusOnline() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(tint.Green())
}

// StatusOffline returns a style for offline status.
func StatusOffline() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(tint.Red())
}

// StatusUpdating returns a style for updating status.
func StatusUpdating() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(tint.Yellow())
}

// Bold returns a bold style with the theme foreground.
func Bold() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(tint.Fg())
}

// Dim returns a dimmed style.
func Dim() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(tint.BrightBlack())
}

// Highlight returns a highlighted style.
func Highlight() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(tint.Cyan())
}

// Title returns a style for titles.
func Title() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(tint.Purple())
}

// Subtitle returns a style for subtitles.
func Subtitle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(tint.BrightBlack())
}

// Link returns a style for links/URLs.
func Link() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(tint.Blue()).
		Underline(true)
}

// Code returns a style for code/commands.
func Code() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(tint.Green())
}

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
	style := lipgloss.NewStyle().Foreground(tint.Cyan())
	return style.Render(formatFloat(watts) + "W")
}

// FormatEnergy formats energy value with appropriate styling.
func FormatEnergy(wh float64) string {
	style := lipgloss.NewStyle().Foreground(tint.Blue())
	if wh >= 1000 {
		return style.Render(formatFloat(wh/1000) + "kWh")
	}
	return style.Render(formatFloat(wh) + "Wh")
}

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

// Fg returns the current theme foreground color.
func Fg() lipgloss.TerminalColor {
	return tint.Fg()
}

// Bg returns the current theme background color.
func Bg() lipgloss.TerminalColor {
	return tint.Bg()
}

// Green returns the current theme green color.
func Green() lipgloss.TerminalColor {
	return tint.Green()
}

// Red returns the current theme red color.
func Red() lipgloss.TerminalColor {
	return tint.Red()
}

// Yellow returns the current theme yellow color.
func Yellow() lipgloss.TerminalColor {
	return tint.Yellow()
}

// Blue returns the current theme blue color.
func Blue() lipgloss.TerminalColor {
	return tint.Blue()
}

// Cyan returns the current theme cyan color.
func Cyan() lipgloss.TerminalColor {
	return tint.Cyan()
}

// Purple returns the current theme purple color.
func Purple() lipgloss.TerminalColor {
	return tint.Purple()
}

// BrightBlack returns the current theme bright black color.
func BrightBlack() lipgloss.TerminalColor {
	return tint.BrightBlack()
}
