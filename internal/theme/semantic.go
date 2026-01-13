// Package theme provides semantic color support for consistent theming.
package theme

import (
	"image/color"
	"sync"

	"charm.land/lipgloss/v2"
)

// SemanticColors defines colors by their semantic role, not their hue.
// This allows themes to consistently apply meaning to UI elements.
type SemanticColors struct {
	// Primary actions and highlights
	Primary   color.Color
	Secondary color.Color

	// Emphasis levels
	Highlight color.Color
	Muted     color.Color

	// Text colors
	Text    color.Color
	AltText color.Color

	// Feedback colors
	Success color.Color
	Warning color.Color
	Error   color.Color
	Info    color.Color

	// Backgrounds
	Background    color.Color
	AltBackground color.Color

	// States
	Online   color.Color
	Offline  color.Color
	Updating color.Color
	Idle     color.Color

	// Table styling
	TableHeader  color.Color
	TableCell    color.Color
	TableAltCell color.Color
	TableBorder  color.Color
}

// SemanticOverrides allows users to override semantic colors via config.
// Colors are specified as hex strings (e.g., "#ff79c6").
type SemanticOverrides struct {
	Primary       string `yaml:"primary,omitempty" json:"primary,omitempty" mapstructure:"primary"`
	Secondary     string `yaml:"secondary,omitempty" json:"secondary,omitempty" mapstructure:"secondary"`
	Highlight     string `yaml:"highlight,omitempty" json:"highlight,omitempty" mapstructure:"highlight"`
	Muted         string `yaml:"muted,omitempty" json:"muted,omitempty" mapstructure:"muted"`
	Text          string `yaml:"text,omitempty" json:"text,omitempty" mapstructure:"text"`
	AltText       string `yaml:"alt_text,omitempty" json:"alt_text,omitempty" mapstructure:"alt_text"`
	Success       string `yaml:"success,omitempty" json:"success,omitempty" mapstructure:"success"`
	Warning       string `yaml:"warning,omitempty" json:"warning,omitempty" mapstructure:"warning"`
	Error         string `yaml:"error,omitempty" json:"error,omitempty" mapstructure:"error"`
	Info          string `yaml:"info,omitempty" json:"info,omitempty" mapstructure:"info"`
	Background    string `yaml:"background,omitempty" json:"background,omitempty" mapstructure:"background"`
	AltBackground string `yaml:"alt_background,omitempty" json:"alt_background,omitempty" mapstructure:"alt_background"`
	Online        string `yaml:"online,omitempty" json:"online,omitempty" mapstructure:"online"`
	Offline       string `yaml:"offline,omitempty" json:"offline,omitempty" mapstructure:"offline"`
	Updating      string `yaml:"updating,omitempty" json:"updating,omitempty" mapstructure:"updating"`
	Idle          string `yaml:"idle,omitempty" json:"idle,omitempty" mapstructure:"idle"`
	TableHeader   string `yaml:"table_header,omitempty" json:"table_header,omitempty" mapstructure:"table_header"`
	TableCell     string `yaml:"table_cell,omitempty" json:"table_cell,omitempty" mapstructure:"table_cell"`
	TableAltCell  string `yaml:"table_alt_cell,omitempty" json:"table_alt_cell,omitempty" mapstructure:"table_alt_cell"`
	TableBorder   string `yaml:"table_border,omitempty" json:"table_border,omitempty" mapstructure:"table_border"`
}

var (
	semanticColors SemanticColors
	semanticMu     sync.RWMutex
)

// GetSemanticColors returns the current semantic color configuration.
// Thread-safe for concurrent access.
func GetSemanticColors() SemanticColors {
	semanticMu.RLock()
	defer semanticMu.RUnlock()
	return semanticColors
}

// setSemanticColors sets the semantic colors. Internal use only.
func setSemanticColors(colors SemanticColors) {
	semanticMu.Lock()
	defer semanticMu.Unlock()
	semanticColors = colors
}

// ApplySemanticOverrides applies user overrides to the current semantic colors.
// Only non-empty override values are applied.
func ApplySemanticOverrides(overrides *SemanticOverrides) {
	if overrides == nil {
		return
	}
	semanticMu.Lock()
	defer semanticMu.Unlock()

	applyOverrides(overrides)
}

// applyOverrides applies individual override values (called with lock held).
func applyOverrides(o *SemanticOverrides) {
	applyPrimaryOverrides(o)
	applyFeedbackOverrides(o)
	applyStateOverrides(o)
	applyTableOverrides(o)
}

// applyPrimaryOverrides applies primary, secondary, text, and background overrides.
func applyPrimaryOverrides(o *SemanticOverrides) {
	if o.Primary != "" {
		semanticColors.Primary = lipgloss.Color(o.Primary)
	}
	if o.Secondary != "" {
		semanticColors.Secondary = lipgloss.Color(o.Secondary)
	}
	if o.Highlight != "" {
		semanticColors.Highlight = lipgloss.Color(o.Highlight)
	}
	if o.Muted != "" {
		semanticColors.Muted = lipgloss.Color(o.Muted)
	}
	if o.Text != "" {
		semanticColors.Text = lipgloss.Color(o.Text)
	}
	if o.AltText != "" {
		semanticColors.AltText = lipgloss.Color(o.AltText)
	}
	if o.Background != "" {
		semanticColors.Background = lipgloss.Color(o.Background)
	}
	if o.AltBackground != "" {
		semanticColors.AltBackground = lipgloss.Color(o.AltBackground)
	}
}

// applyFeedbackOverrides applies success, warning, error, and info overrides.
func applyFeedbackOverrides(o *SemanticOverrides) {
	if o.Success != "" {
		semanticColors.Success = lipgloss.Color(o.Success)
	}
	if o.Warning != "" {
		semanticColors.Warning = lipgloss.Color(o.Warning)
	}
	if o.Error != "" {
		semanticColors.Error = lipgloss.Color(o.Error)
	}
	if o.Info != "" {
		semanticColors.Info = lipgloss.Color(o.Info)
	}
}

// applyStateOverrides applies online, offline, updating, and idle overrides.
func applyStateOverrides(o *SemanticOverrides) {
	if o.Online != "" {
		semanticColors.Online = lipgloss.Color(o.Online)
	}
	if o.Offline != "" {
		semanticColors.Offline = lipgloss.Color(o.Offline)
	}
	if o.Updating != "" {
		semanticColors.Updating = lipgloss.Color(o.Updating)
	}
	if o.Idle != "" {
		semanticColors.Idle = lipgloss.Color(o.Idle)
	}
}

// applyTableOverrides applies table styling overrides.
func applyTableOverrides(o *SemanticOverrides) {
	if o.TableHeader != "" {
		semanticColors.TableHeader = lipgloss.Color(o.TableHeader)
	}
	if o.TableCell != "" {
		semanticColors.TableCell = lipgloss.Color(o.TableCell)
	}
	if o.TableAltCell != "" {
		semanticColors.TableAltCell = lipgloss.Color(o.TableAltCell)
	}
	if o.TableBorder != "" {
		semanticColors.TableBorder = lipgloss.Color(o.TableBorder)
	}
}

// =============================================================================
// Semantic Style Accessors - Return lipgloss.Style for each semantic role
// =============================================================================

// SemanticPrimary returns a style for primary actions and highlights.
func SemanticPrimary() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().Primary)
}

// SemanticSecondary returns a style for secondary/supporting elements.
func SemanticSecondary() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().Secondary)
}

// SemanticHighlight returns a style for emphasis and selection.
func SemanticHighlight() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().Highlight)
}

// SemanticMuted returns a style for disabled or less important content.
func SemanticMuted() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().Muted)
}

// SemanticText returns a style for primary text.
func SemanticText() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().Text)
}

// SemanticAltText returns a style for secondary/dimmed text.
func SemanticAltText() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().AltText)
}

// SemanticSuccess returns a style for successful operations.
func SemanticSuccess() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().Success)
}

// SemanticWarning returns a style for warnings and cautions.
func SemanticWarning() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().Warning)
}

// SemanticError returns a style for errors and failures.
func SemanticError() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().Error)
}

// SemanticInfo returns a style for informational messages.
func SemanticInfo() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().Info)
}

// SemanticOnline returns a style for device online status.
func SemanticOnline() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().Online)
}

// SemanticOffline returns a style for device offline status.
func SemanticOffline() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().Offline)
}

// SemanticUpdating returns a style for device updating status.
func SemanticUpdating() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().Updating)
}

// SemanticIdle returns a style for device idle/inactive status.
func SemanticIdle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().Idle)
}

// SemanticTableHeader returns a style for table headers.
func SemanticTableHeader() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().TableHeader)
}

// SemanticTableCell returns a style for table cells.
func SemanticTableCell() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().TableCell)
}

// SemanticTableAltCell returns a style for alternating table cells.
func SemanticTableAltCell() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().TableAltCell)
}

// SemanticTableBorder returns a style for table borders.
func SemanticTableBorder() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(GetSemanticColors().TableBorder)
}

// StyledKeybindings applies Warning color to keybinding text (e.g., "j/k:nav").
func StyledKeybindings(text string) string {
	return SemanticWarning().Render(text)
}
