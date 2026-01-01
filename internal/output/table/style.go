// Package table provides table rendering with styled output.
package table

import (
	"charm.land/lipgloss/v2"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// ModeChecker is an interface for checking output mode and color state.
// This allows decoupling from the iostreams package.
type ModeChecker interface {
	IsPlainMode() bool
	ColorEnabled() bool
}

// BorderStyle defines the border style for tables.
type BorderStyle int

// BorderStyle constants define available border styles.
const (
	BorderNone    BorderStyle = iota // No visible borders
	BorderRounded                    // Modern rounded corners (default for TTY)
	BorderSquare                     // Square corners
	BorderDouble                     // Double-line borders
	BorderHeavy                      // Heavy/bold borders
	BorderASCII                      // ASCII-only for --plain mode
)

// borderStyles maps border style constants to lipgloss border definitions.
var borderStyles = map[BorderStyle]lipgloss.Border{
	BorderNone:    lipgloss.HiddenBorder(),
	BorderRounded: lipgloss.RoundedBorder(),
	BorderSquare:  lipgloss.NormalBorder(),
	BorderDouble:  lipgloss.DoubleBorder(),
	BorderHeavy:   lipgloss.BlockBorder(),
	BorderASCII:   lipgloss.ASCIIBorder(),
}

// Style defines the visual style for a table.
type Style struct {
	Header           lipgloss.Style
	Cell             lipgloss.Style
	AltCell          lipgloss.Style // Alternating row color
	PrimaryCell      lipgloss.Style // First column styling (e.g., Name column)
	Border           lipgloss.Style // Border character styling
	BorderStyle      BorderStyle    // Border style (rounded, square, etc.)
	Padding          int
	ShowBorder       bool
	UppercaseHeaders bool // Make headers ALL CAPS
	PlainMode        bool // True for --plain: no borders, tab-separated
}

// DefaultStyle returns the default table style using semantic colors.
// Uses rounded borders with themed colors for modern terminal appearance.
// Colors are drawn from the semantic theme system for consistency.
func DefaultStyle() Style {
	colors := theme.GetSemanticColors()
	return Style{
		Header:           lipgloss.NewStyle().Bold(true).Foreground(colors.TableHeader),
		Cell:             lipgloss.NewStyle().Foreground(colors.TableCell),
		AltCell:          lipgloss.NewStyle().Foreground(colors.TableAltCell),
		PrimaryCell:      lipgloss.NewStyle().Foreground(colors.Primary),
		Border:           lipgloss.NewStyle().Foreground(colors.TableBorder),
		BorderStyle:      BorderRounded,
		Padding:          1,
		ShowBorder:       true,
		UppercaseHeaders: true, // Table headers are ALL CAPS by default
		PlainMode:        false,
	}
}

// NoColorStyle returns a table style for --no-color output.
// Uses ASCII borders (|-+) with no ANSI codes (no colors, no bold).
func NoColorStyle() Style {
	return Style{
		Header:           lipgloss.NewStyle(),
		Cell:             lipgloss.NewStyle(),
		AltCell:          lipgloss.NewStyle(),
		PrimaryCell:      lipgloss.NewStyle(),
		Border:           lipgloss.NewStyle(),
		BorderStyle:      BorderASCII,
		Padding:          1,
		ShowBorder:       true,
		UppercaseHeaders: true,
		PlainMode:        false,
	}
}

// PlainStyle returns a table style for --plain output.
// No borders, tab-separated values for machine-readable/scriptable output.
func PlainStyle() Style {
	return Style{
		Header:           lipgloss.NewStyle(),
		Cell:             lipgloss.NewStyle(),
		AltCell:          lipgloss.NewStyle(),
		PrimaryCell:      lipgloss.NewStyle(),
		Border:           lipgloss.NewStyle(),
		BorderStyle:      BorderNone,
		Padding:          0,
		ShowBorder:       false,
		UppercaseHeaders: true,
		PlainMode:        true, // Enables tab-separated output
	}
}

// GetStyle returns the appropriate table style based on output mode.
// Priority: --plain > --no-color > default (with colors).
func GetStyle(ios ModeChecker) Style {
	if ios == nil {
		return DefaultStyle()
	}
	// --plain: aligned columns, no borders
	if ios.IsPlainMode() {
		return PlainStyle()
	}
	// --no-color or non-TTY: ASCII borders, no colors
	if !ios.ColorEnabled() {
		return NoColorStyle()
	}
	return DefaultStyle()
}

// ShouldHideHeaders returns true if --no-headers flag is set.
func ShouldHideHeaders() bool {
	return viper.GetBool("no-headers")
}
