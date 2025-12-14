// Package theme provides theming support for the CLI using bubbletint v2.
// Themes apply to ALL CLI output, not just the TUI dashboard.
//
// This package wraps bubbletint v2 to provide:
// - 280+ terminal color themes
// - Theme-aware lipgloss styles for rich text formatting
// - Pre-rendered styled strings for common UI elements
package theme

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"charm.land/lipgloss/v2"
	tint "github.com/lrstanley/bubbletint/v2"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// DefaultTheme is the default color theme for the CLI.
const DefaultTheme = "dracula"

// CustomColors holds user-defined color overrides.
// Colors are specified as hex strings (e.g., "#50fa7b").
type CustomColors struct {
	Foreground  string `yaml:"foreground" json:"foreground,omitempty"`
	Background  string `yaml:"background" json:"background,omitempty"`
	Green       string `yaml:"green" json:"green,omitempty"`
	Red         string `yaml:"red" json:"red,omitempty"`
	Yellow      string `yaml:"yellow" json:"yellow,omitempty"`
	Blue        string `yaml:"blue" json:"blue,omitempty"`
	Cyan        string `yaml:"cyan" json:"cyan,omitempty"`
	Purple      string `yaml:"purple" json:"purple,omitempty"`
	BrightBlack string `yaml:"bright_black" json:"bright_black,omitempty"`
}

// File represents the structure of an external theme file.
type File struct {
	Name   string       `yaml:"name" json:"name,omitempty"`
	Colors CustomColors `yaml:"colors" json:"colors,omitempty"`
}

var (
	initOnce        sync.Once
	customOverrides *CustomColors
	customMu        sync.RWMutex
)

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
// Custom Theme Configuration
// =============================================================================

// ApplyConfig applies a theme configuration with optional color overrides.
// It supports setting a base theme by name, loading from file, or applying color overrides.
func ApplyConfig(name string, colors map[string]string, filePath string) error {
	// Load external file if specified
	if filePath != "" {
		if err := applyFromFile(filePath); err != nil {
			return err
		}
		return nil
	}

	// Set base theme if specified
	if name != "" {
		if !SetTheme(name) {
			return fmt.Errorf("unknown theme: %s", name)
		}
	}

	// Apply color overrides if specified
	if len(colors) > 0 {
		cc := parseColorsMap(colors)
		SetCustomColors(cc)
	}

	return nil
}

// applyFromFile loads and applies theme configuration from a file.
func applyFromFile(filePath string) error {
	expanded := expandPath(filePath)
	//nolint:gosec // G304: File path is user-configured, intentional
	data, err := os.ReadFile(expanded)
	if err != nil {
		return fmt.Errorf("failed to read theme file: %w", err)
	}
	var fileTheme File
	if err := yaml.Unmarshal(data, &fileTheme); err != nil {
		return fmt.Errorf("invalid theme file: %w", err)
	}
	// Apply base theme from file if specified
	if fileTheme.Name != "" {
		if !SetTheme(fileTheme.Name) {
			return fmt.Errorf("unknown theme in file: %s", fileTheme.Name)
		}
	}
	// Apply colors from file
	SetCustomColors(&fileTheme.Colors)
	return nil
}

// SetCustomColors sets custom color overrides.
func SetCustomColors(colors *CustomColors) {
	customMu.Lock()
	defer customMu.Unlock()
	customOverrides = colors
}

// ClearCustomColors removes all custom color overrides.
func ClearCustomColors() {
	customMu.Lock()
	defer customMu.Unlock()
	customOverrides = nil
}

// GetCustomColors returns the current custom color overrides.
func GetCustomColors() *CustomColors {
	customMu.RLock()
	defer customMu.RUnlock()
	return customOverrides
}

// parseColorsMap converts a string map to CustomColors.
func parseColorsMap(colors map[string]string) *CustomColors {
	return &CustomColors{
		Foreground:  colors["foreground"],
		Background:  colors["background"],
		Green:       colors["green"],
		Red:         colors["red"],
		Yellow:      colors["yellow"],
		Blue:        colors["blue"],
		Cyan:        colors["cyan"],
		Purple:      colors["purple"],
		BrightBlack: colors["bright_black"],
	}
}

// expandPath expands ~ to home directory.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

// parseHexColor parses a hex color string (e.g., "#50fa7b") to color.Color.
func parseHexColor(hex string) color.Color {
	if hex == "" {
		return nil
	}
	// Remove # prefix if present
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return nil
	}

	var r, g, b uint8
	n, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	if err != nil || n != 3 {
		return nil
	}
	return &tint.Color{R: r, G: g, B: b, A: 255}
}

// getCustomColor returns the custom override for a color if set.
func getCustomColor(colorName string) color.Color {
	customMu.RLock()
	defer customMu.RUnlock()

	if customOverrides == nil {
		return nil
	}

	var hex string
	switch colorName {
	case "foreground":
		hex = customOverrides.Foreground
	case "background":
		hex = customOverrides.Background
	case "green":
		hex = customOverrides.Green
	case "red":
		hex = customOverrides.Red
	case "yellow":
		hex = customOverrides.Yellow
	case "blue":
		hex = customOverrides.Blue
	case "cyan":
		hex = customOverrides.Cyan
	case "purple":
		hex = customOverrides.Purple
	case "bright_black":
		hex = customOverrides.BrightBlack
	}

	return parseHexColor(hex)
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
// Custom overrides take precedence over the base theme.
func Fg() color.Color {
	if c := getCustomColor("foreground"); c != nil {
		return c
	}
	t := Current()
	if t == nil {
		return draculaFg
	}
	return defaultColor(t.Fg, draculaFg)
}

// Bg returns the current theme background color.
// Custom overrides take precedence over the base theme.
func Bg() color.Color {
	if c := getCustomColor("background"); c != nil {
		return c
	}
	t := Current()
	if t == nil {
		return draculaBg
	}
	return defaultColor(t.Bg, draculaBg)
}

// Green returns the current theme green color.
// Custom overrides take precedence over the base theme.
func Green() color.Color {
	if c := getCustomColor("green"); c != nil {
		return c
	}
	t := Current()
	if t == nil {
		return draculaGreen
	}
	return defaultColor(t.Green, draculaGreen)
}

// Red returns the current theme red color.
// Custom overrides take precedence over the base theme.
func Red() color.Color {
	if c := getCustomColor("red"); c != nil {
		return c
	}
	t := Current()
	if t == nil {
		return draculaRed
	}
	return defaultColor(t.Red, draculaRed)
}

// Yellow returns the current theme yellow color.
// Custom overrides take precedence over the base theme.
func Yellow() color.Color {
	if c := getCustomColor("yellow"); c != nil {
		return c
	}
	t := Current()
	if t == nil {
		return draculaYellow
	}
	return defaultColor(t.Yellow, draculaYellow)
}

// Blue returns the current theme blue color.
// Custom overrides take precedence over the base theme.
func Blue() color.Color {
	if c := getCustomColor("blue"); c != nil {
		return c
	}
	t := Current()
	if t == nil {
		return draculaBlue
	}
	return defaultColor(t.Blue, draculaBlue)
}

// Cyan returns the current theme cyan color.
// Custom overrides take precedence over the base theme.
func Cyan() color.Color {
	if c := getCustomColor("cyan"); c != nil {
		return c
	}
	t := Current()
	if t == nil {
		return draculaCyan
	}
	return defaultColor(t.Cyan, draculaCyan)
}

// Purple returns the current theme purple color.
// Custom overrides take precedence over the base theme.
func Purple() color.Color {
	if c := getCustomColor("purple"); c != nil {
		return c
	}
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
// Custom overrides take precedence over the base theme.
func BrightBlack() color.Color {
	if c := getCustomColor("bright_black"); c != nil {
		return c
	}
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
