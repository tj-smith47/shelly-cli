// Package syntax provides syntax highlighting for CLI output.
package syntax

import (
	"bytes"
	"os"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/mattn/go-isatty"
	"github.com/spf13/viper"
)

// IsTTY is a package-level function for TTY detection, overridable in tests.
var IsTTY = func() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

// ShouldHighlight returns true if syntax highlighting should be enabled.
// Returns false for --no-color, --plain, non-TTY output, or test environments.
func ShouldHighlight() bool {
	// Require stdout to be a TTY (disables in tests and pipes)
	if !IsTTY() {
		return false
	}
	// Disabled for --plain or --no-color
	if viper.GetBool("plain") || viper.GetBool("no-color") {
		return false
	}
	// Check NO_COLOR environment variable
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	// Check SHELLY_NO_COLOR environment variable
	if _, ok := os.LookupEnv("SHELLY_NO_COLOR"); ok {
		return false
	}
	// Check TERM=dumb
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	return true
}

// HighlightCode applies syntax highlighting to code using chroma.
// Falls back to plain text if highlighting fails.
func HighlightCode(code, language string) string {
	lexer := lexers.Get(language)
	if lexer == nil {
		return code
	}
	lexer = chroma.Coalesce(lexer)

	// Use a theme that works well with the current terminal theme
	style := GetChromaStyle()

	// Use terminal256 formatter for broad compatibility
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return code
	}

	return buf.String()
}

// GetChromaStyle returns a chroma style that matches the current theme.
func GetChromaStyle() *chroma.Style {
	// Try to match the current theme name to a chroma style
	currentTheme := viper.GetString("theme.name")
	if currentTheme == "" {
		currentTheme = "dracula"
	}

	// Map theme names to chroma styles
	styleMap := map[string]string{
		"dracula":      "dracula",
		"nord":         "nord",
		"gruvbox":      "gruvbox",
		"gruvbox-dark": "gruvbox",
		"tokyo-night":  "tokyonight-night",
		"catppuccin":   "catppuccin-mocha",
	}

	if chromaStyle, ok := styleMap[strings.ToLower(currentTheme)]; ok {
		if style := styles.Get(chromaStyle); style != nil {
			return style
		}
	}

	// Default to dracula which works well on dark terminals
	if style := styles.Get("dracula"); style != nil {
		return style
	}

	return styles.Fallback
}
