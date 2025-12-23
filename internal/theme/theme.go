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

// StyleFunc represents a styling function compatible with lipgloss Style.Render.
// Used to pass theme styling functions to formatters.
type StyleFunc func(...string) string

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

// CurrentThemeName returns the name of the current theme.
func CurrentThemeName() string {
	t := Current()
	if t == nil {
		return DefaultTheme
	}
	return t.ID
}

// SetTheme sets the current theme by name.
// Also updates the semantic color mapping for the theme.
func SetTheme(name string) bool {
	if !tint.SetTintID(name) {
		return false
	}
	// Apply semantic color mapping for this theme
	setSemanticColors(GetThemeMapping(name))
	return true
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

// ApplyConfig applies a theme configuration with optional color and semantic overrides.
// It supports setting a base theme by name, loading from file, or applying color overrides.
func ApplyConfig(name string, colors map[string]string, semantics *SemanticOverrides, filePath string) error {
	// Load external file if specified
	if filePath != "" {
		if err := applyFromFile(filePath); err != nil {
			return err
		}
		// Apply semantic overrides on top of file config
		if semantics != nil {
			ApplySemanticOverrides(semantics)
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

	// Apply semantic overrides if specified
	if semantics != nil {
		ApplySemanticOverrides(semantics)
	}

	return nil
}

// SaveTheme saves the theme name to configuration file.
func SaveTheme(themeName string) error {
	viper.Set("theme", themeName)

	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		// Create default config path
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configDir := filepath.Join(home, ".config", "shelly")
		if err := os.MkdirAll(configDir, 0o700); err != nil {
			return err
		}
		configFile = filepath.Join(configDir, "config.yaml")
	}

	return viper.WriteConfigAs(configFile)
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
	draculaPink      = &tint.Color{R: 255, G: 121, B: 198, A: 255}
	draculaOrange    = &tint.Color{R: 255, G: 184, B: 108, A: 255}
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
// Note: bubbletint's Dracula theme incorrectly defines Purple as #ff79c6 (pink),
// so we always use draculaPurple (#bd93f9) for consistency with the actual Dracula palette.
func Purple() color.Color {
	if c := getCustomColor("purple"); c != nil {
		return c
	}
	// Always use draculaPurple because bubbletint's Dracula has incorrect Purple value
	return draculaPurple
}

// Magenta is an alias for Purple for compatibility.
func Magenta() color.Color {
	return Purple()
}

// Pink returns the current theme pink color (Dracula's bright magenta).
// Falls back to Dracula pink (#ff79c6) if not available in theme.
func Pink() color.Color {
	// Pink is Dracula-specific (#ff79c6) - not a standard terminal color
	// Always return the Dracula pink constant
	return draculaPink
}

// Orange returns the Dracula orange color (#ffb86c).
// Orange is Dracula-specific and not a standard terminal color.
func Orange() color.Color {
	return draculaOrange
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
// Uses the semantic Success color for consistent theming.
func StatusOK() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().Success)
}

// StatusWarn returns a style for warning status.
// Uses the semantic Warning color for consistent theming.
func StatusWarn() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().Warning)
}

// StatusError returns a style for error status.
// Uses the semantic Error color for consistent theming.
func StatusError() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().Error)
}

// StatusInfo returns a style for info status.
// Uses the semantic Info color for consistent theming.
func StatusInfo() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().Info)
}

// StatusOnline returns a style for online status.
// Uses the semantic Online color for consistent theming.
func StatusOnline() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().Online)
}

// StatusOffline returns a style for offline status.
// Uses the semantic Offline color for consistent theming.
func StatusOffline() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().Offline)
}

// StatusUpdating returns a style for updating status.
// Uses the semantic Updating color for consistent theming.
func StatusUpdating() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().Updating)
}

// Bold returns a bold style with the semantic text color.
func Bold() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(GetSemanticColors().Text)
}

// Dim returns a dimmed style using the semantic muted color.
func Dim() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().Muted)
}

// Highlight returns a highlighted style using the semantic highlight color.
func Highlight() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().Highlight)
}

// Title returns a style for titles using the semantic highlight color (cyan).
// The Shelly logo is blue, so we use cyan to complement it.
func Title() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(GetSemanticColors().Highlight)
}

// Subtitle returns a style for subtitles using the semantic alt text color.
func Subtitle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().AltText)
}

// Link returns a style for links/URLs using the semantic secondary color.
func Link() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().Secondary).Underline(true)
}

// Code returns a style for code/commands using the semantic success color.
func Code() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().Success)
}

// =============================================================================
// FalseStyle - Styling for false case in boolean renderers
// =============================================================================

// FalseStyle defines how to style the false case in boolean renderers.
type FalseStyle int

const (
	// FalseError uses StatusError (red) for the false case.
	// Use when false represents a bad/error state (offline, failed, etc).
	FalseError FalseStyle = iota
	// FalseDim uses Dim (gray) for the false case.
	// Use when false is the normal/expected state (disabled, stopped, etc).
	FalseDim
)

// Render applies the false style to the given text.
func (s FalseStyle) Render(text string) string {
	switch s {
	case FalseDim:
		return Dim().Render(text)
	default:
		return StatusError().Render(text)
	}
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

// StyledPower formats power value with cyan styling for dashboard display.
func StyledPower(watts float64) string {
	style := lipgloss.NewStyle().Foreground(Cyan())
	return style.Render(formatFloat(watts) + "W")
}

// StyledEnergy formats energy value with blue styling for dashboard display.
func StyledEnergy(wh float64) string {
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

// ColorToHex converts a color.Color to a hex string.
func ColorToHex(c interface{ RGBA() (r, g, b, a uint32) }) string {
	if c == nil {
		return ""
	}
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
}

// BuildColorOverrides creates a map of custom color overrides from CustomColors.
func BuildColorOverrides(custom *CustomColors) map[string]string {
	if custom == nil {
		return nil
	}

	colors := make(map[string]string)
	addIfSet := func(key, value string) {
		if value != "" {
			colors[key] = value
		}
	}

	addIfSet("foreground", custom.Foreground)
	addIfSet("background", custom.Background)
	addIfSet("green", custom.Green)
	addIfSet("red", custom.Red)
	addIfSet("yellow", custom.Yellow)
	addIfSet("blue", custom.Blue)
	addIfSet("cyan", custom.Cyan)
	addIfSet("purple", custom.Purple)
	addIfSet("bright_black", custom.BrightBlack)

	if len(colors) == 0 {
		return nil
	}
	return colors
}
