// Package theme provides theme-to-semantic color mappings.
package theme

import "charm.land/lipgloss/v2"

// themeMappings maps theme names to their semantic color mapping functions.
var themeMappings = map[string]func() SemanticColors{
	"dracula":          DraculaSemanticMapping,
	"nord":             NordSemanticMapping,
	"tokyo-night":      TokyoNightSemanticMapping,
	"tokyonight":       TokyoNightSemanticMapping,
	"tokyonight-night": TokyoNightSemanticMapping,
	"gruvbox":          GruvboxSemanticMapping,
	"gruvbox-dark":     GruvboxSemanticMapping,
	"catppuccin":       CatppuccinSemanticMapping,
	"catppuccin-mocha": CatppuccinSemanticMapping,
}

// GetThemeMapping returns the semantic colors for a named theme.
// Falls back to MappingFromTheme() for themes without custom mappings.
func GetThemeMapping(themeName string) SemanticColors {
	if fn, ok := themeMappings[themeName]; ok {
		return fn()
	}
	return MappingFromTheme()
}

// MappingFromTheme creates semantic colors from the current bubbletint theme.
// This is the fallback for themes without custom semantic mappings.
// Colors are read directly from the active theme - no conversion needed.
func MappingFromTheme() SemanticColors {
	return SemanticColors{
		Primary:       Purple(),
		Secondary:     Blue(),
		Highlight:     Cyan(),
		Muted:         BrightBlack(),
		Text:          Fg(),
		AltText:       BrightBlack(),
		Success:       Green(),
		Warning:       Yellow(),
		Error:         Red(),
		Info:          Cyan(),
		Background:    Bg(),
		AltBackground: BrightBlack(),
		Online:        Green(),
		Offline:       Red(),
		Updating:      Yellow(),
		Idle:          BrightBlack(),
		TableHeader:   Cyan(),
		TableCell:     Yellow(),
		TableAltCell:  Orange(),
		TableBorder:   Pink(), // Pink border
	}
}

// DraculaSemanticMapping returns semantic colors for the Dracula theme.
func DraculaSemanticMapping() SemanticColors {
	return SemanticColors{
		Primary:       lipgloss.Color("#bd93f9"), // Purple
		Secondary:     lipgloss.Color("#6272a4"), // Comment blue
		Highlight:     lipgloss.Color("#8be9fd"), // Cyan
		Muted:         lipgloss.Color("#6272a4"), // Comment
		Text:          lipgloss.Color("#f8f8f2"), // Foreground
		AltText:       lipgloss.Color("#6272a4"), // Dimmed
		Success:       lipgloss.Color("#50fa7b"), // Green
		Warning:       lipgloss.Color("#f1fa8c"), // Yellow
		Error:         lipgloss.Color("#ff5555"), // Red
		Info:          lipgloss.Color("#8be9fd"), // Cyan
		Background:    lipgloss.Color("#282a36"), // Background
		AltBackground: lipgloss.Color("#44475a"), // Current line
		Online:        lipgloss.Color("#50fa7b"), // Green
		Offline:       lipgloss.Color("#ff5555"), // Red
		Updating:      lipgloss.Color("#f1fa8c"), // Yellow
		Idle:          lipgloss.Color("#6272a4"), // Comment
		TableHeader:   lipgloss.Color("#8be9fd"), // Cyan
		TableCell:     lipgloss.Color("#f1fa8c"), // Yellow
		TableAltCell:  lipgloss.Color("#ffb86c"), // Orange
		TableBorder:   lipgloss.Color("#ff79c6"), // Pink
	}
}

// NordSemanticMapping returns semantic colors for the Nord theme.
func NordSemanticMapping() SemanticColors {
	return SemanticColors{
		Primary:       lipgloss.Color("#88c0d0"), // Nord8 - frost cyan
		Secondary:     lipgloss.Color("#81a1c1"), // Nord9 - frost blue
		Highlight:     lipgloss.Color("#8fbcbb"), // Nord7 - frost teal
		Muted:         lipgloss.Color("#4c566a"), // Nord3 - polar night
		Text:          lipgloss.Color("#eceff4"), // Nord6 - snow storm
		AltText:       lipgloss.Color("#d8dee9"), // Nord4 - snow storm dim
		Success:       lipgloss.Color("#a3be8c"), // Nord14 - aurora green
		Warning:       lipgloss.Color("#ebcb8b"), // Nord13 - aurora yellow
		Error:         lipgloss.Color("#bf616a"), // Nord11 - aurora red
		Info:          lipgloss.Color("#5e81ac"), // Nord10 - frost deep blue
		Background:    lipgloss.Color("#2e3440"), // Nord0 - polar night
		AltBackground: lipgloss.Color("#3b4252"), // Nord1 - polar night
		Online:        lipgloss.Color("#a3be8c"), // Nord14 - aurora green
		Offline:       lipgloss.Color("#bf616a"), // Nord11 - aurora red
		Updating:      lipgloss.Color("#ebcb8b"), // Nord13 - aurora yellow
		Idle:          lipgloss.Color("#4c566a"), // Nord3 - polar night
		TableHeader:   lipgloss.Color("#88c0d0"), // Nord8 - frost cyan
		TableCell:     lipgloss.Color("#ebcb8b"), // Nord13 - aurora yellow
		TableAltCell:  lipgloss.Color("#d08770"), // Nord12 - aurora orange
		TableBorder:   lipgloss.Color("#81a1c1"), // Nord9 - frost blue
	}
}

// TokyoNightSemanticMapping returns semantic colors for the Tokyo Night theme.
func TokyoNightSemanticMapping() SemanticColors {
	return SemanticColors{
		Primary:       lipgloss.Color("#7aa2f7"), // Blue
		Secondary:     lipgloss.Color("#bb9af7"), // Purple
		Highlight:     lipgloss.Color("#7dcfff"), // Cyan
		Muted:         lipgloss.Color("#565f89"), // Comment
		Text:          lipgloss.Color("#c0caf5"), // Foreground
		AltText:       lipgloss.Color("#a9b1d6"), // Foreground dim
		Success:       lipgloss.Color("#9ece6a"), // Green
		Warning:       lipgloss.Color("#e0af68"), // Yellow
		Error:         lipgloss.Color("#f7768e"), // Red
		Info:          lipgloss.Color("#7dcfff"), // Cyan
		Background:    lipgloss.Color("#1a1b26"), // Background
		AltBackground: lipgloss.Color("#24283b"), // Background alt
		Online:        lipgloss.Color("#9ece6a"), // Green
		Offline:       lipgloss.Color("#f7768e"), // Red
		Updating:      lipgloss.Color("#e0af68"), // Yellow
		Idle:          lipgloss.Color("#565f89"), // Comment
		TableHeader:   lipgloss.Color("#7dcfff"), // Cyan
		TableCell:     lipgloss.Color("#e0af68"), // Yellow
		TableAltCell:  lipgloss.Color("#ff9e64"), // Orange
		TableBorder:   lipgloss.Color("#bb9af7"), // Purple
	}
}

// GruvboxSemanticMapping returns semantic colors for the Gruvbox theme.
func GruvboxSemanticMapping() SemanticColors {
	return SemanticColors{
		Primary:       lipgloss.Color("#d79921"), // Yellow/gold
		Secondary:     lipgloss.Color("#458588"), // Blue
		Highlight:     lipgloss.Color("#689d6a"), // Aqua
		Muted:         lipgloss.Color("#928374"), // Gray
		Text:          lipgloss.Color("#ebdbb2"), // Foreground
		AltText:       lipgloss.Color("#a89984"), // Gray light
		Success:       lipgloss.Color("#b8bb26"), // Green
		Warning:       lipgloss.Color("#fabd2f"), // Yellow bright
		Error:         lipgloss.Color("#fb4934"), // Red bright
		Info:          lipgloss.Color("#83a598"), // Blue light
		Background:    lipgloss.Color("#282828"), // Background
		AltBackground: lipgloss.Color("#3c3836"), // Background alt
		Online:        lipgloss.Color("#b8bb26"), // Green
		Offline:       lipgloss.Color("#fb4934"), // Red bright
		Updating:      lipgloss.Color("#fabd2f"), // Yellow bright
		Idle:          lipgloss.Color("#928374"), // Gray
		TableHeader:   lipgloss.Color("#83a598"), // Blue light
		TableCell:     lipgloss.Color("#fabd2f"), // Yellow bright
		TableAltCell:  lipgloss.Color("#fe8019"), // Orange
		TableBorder:   lipgloss.Color("#d3869b"), // Purple
	}
}

// CatppuccinSemanticMapping returns semantic colors for the Catppuccin Mocha theme.
func CatppuccinSemanticMapping() SemanticColors {
	return SemanticColors{
		Primary:       lipgloss.Color("#cba6f7"), // Mauve
		Secondary:     lipgloss.Color("#89b4fa"), // Blue
		Highlight:     lipgloss.Color("#94e2d5"), // Teal
		Muted:         lipgloss.Color("#6c7086"), // Overlay0
		Text:          lipgloss.Color("#cdd6f4"), // Text
		AltText:       lipgloss.Color("#a6adc8"), // Subtext0
		Success:       lipgloss.Color("#a6e3a1"), // Green
		Warning:       lipgloss.Color("#f9e2af"), // Yellow
		Error:         lipgloss.Color("#f38ba8"), // Red
		Info:          lipgloss.Color("#89dceb"), // Sky
		Background:    lipgloss.Color("#1e1e2e"), // Base
		AltBackground: lipgloss.Color("#313244"), // Surface0
		Online:        lipgloss.Color("#a6e3a1"), // Green
		Offline:       lipgloss.Color("#f38ba8"), // Red
		Updating:      lipgloss.Color("#f9e2af"), // Yellow
		Idle:          lipgloss.Color("#6c7086"), // Overlay0
		TableHeader:   lipgloss.Color("#94e2d5"), // Teal
		TableCell:     lipgloss.Color("#f9e2af"), // Yellow
		TableAltCell:  lipgloss.Color("#fab387"), // Peach
		TableBorder:   lipgloss.Color("#cba6f7"), // Mauve
	}
}
