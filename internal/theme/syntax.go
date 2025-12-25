// Package theme provides theming support including syntax highlighting.
package theme

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// SyntaxColors holds the colors for syntax highlighting tokens.
// These are derived from the current bubbletint theme.
type SyntaxColors struct {
	Background color.Color
	Text       color.Color
	Keyword    color.Color
	String     color.Color
	Number     color.Color
	Comment    color.Color
	Function   color.Color
	Operator   color.Color
	Type       color.Color
	Constant   color.Color
}

// GetSyntaxColors returns syntax highlighting colors derived from the current theme.
// Maps bubbletint ANSI colors to syntax token types.
func GetSyntaxColors() SyntaxColors {
	return SyntaxColors{
		Background: Bg(),
		Text:       Fg(),
		Keyword:    Purple(),      // Keywords: if, else, return, function, etc.
		String:     Yellow(),      // String literals
		Number:     Cyan(),        // Numeric literals
		Comment:    BrightBlack(), // Comments (muted)
		Function:   Green(),       // Function names
		Operator:   Red(),         // Operators: +, -, =, etc.
		Type:       Cyan(),        // Type names
		Constant:   Purple(),      // Constants: true, false, null
	}
}

// tokenTypeToColor maps a chroma token type to a syntax color.
func tokenTypeToColor(tt chroma.TokenType, colors SyntaxColors) color.Color {
	switch tt {
	// Comments
	case chroma.Comment,
		chroma.CommentHashbang,
		chroma.CommentMultiline,
		chroma.CommentPreproc,
		chroma.CommentPreprocFile,
		chroma.CommentSingle,
		chroma.CommentSpecial:
		return colors.Comment

	// Keywords
	case chroma.Keyword,
		chroma.KeywordDeclaration,
		chroma.KeywordNamespace,
		chroma.KeywordPseudo,
		chroma.KeywordReserved,
		chroma.KeywordType:
		return colors.Keyword

	// Strings
	case chroma.String,
		chroma.StringAffix,
		chroma.StringBacktick,
		chroma.StringChar,
		chroma.StringDelimiter,
		chroma.StringDoc,
		chroma.StringDouble,
		chroma.StringEscape,
		chroma.StringHeredoc,
		chroma.StringInterpol,
		chroma.StringOther,
		chroma.StringRegex,
		chroma.StringSingle,
		chroma.StringSymbol:
		return colors.String

	// Numbers
	case chroma.Number,
		chroma.NumberBin,
		chroma.NumberFloat,
		chroma.NumberHex,
		chroma.NumberInteger,
		chroma.NumberIntegerLong,
		chroma.NumberOct:
		return colors.Number

	// Operators
	case chroma.Operator,
		chroma.OperatorWord:
		return colors.Operator

	// Functions/Methods
	case chroma.NameFunction,
		chroma.NameFunctionMagic,
		chroma.NameBuiltin,
		chroma.NameBuiltinPseudo:
		return colors.Function

	// Types/Classes
	case chroma.NameClass,
		chroma.NameException:
		return colors.Type

	// Constants (true, false, null, etc.)
	case chroma.NameConstant,
		chroma.KeywordConstant:
		return colors.Constant

	default:
		return colors.Text
	}
}

// HighlightCode returns syntax-highlighted code using the current theme colors.
// The language parameter specifies the programming language (e.g., "javascript", "json").
func HighlightCode(code, language string) string {
	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code
	}

	colors := GetSyntaxColors()
	var result strings.Builder

	for token := iterator(); token != chroma.EOF; token = iterator() {
		tokenColor := tokenTypeToColor(token.Type, colors)
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(colorToHex(tokenColor)))

		// Handle newlines separately to avoid styling issues
		value := token.Value
		if strings.Contains(value, "\n") {
			parts := strings.Split(value, "\n")
			for i, part := range parts {
				if part != "" {
					result.WriteString(style.Render(part))
				}
				if i < len(parts)-1 {
					result.WriteString("\n")
				}
			}
		} else {
			result.WriteString(style.Render(value))
		}
	}

	return result.String()
}

// HighlightJavaScript is a convenience function for JavaScript code.
func HighlightJavaScript(code string) string {
	return HighlightCode(code, "javascript")
}

// HighlightJSON is a convenience function for JSON code.
func HighlightJSON(code string) string {
	return HighlightCode(code, "json")
}

// colorToHex converts a color.Color to a hex string for lipgloss.
func colorToHex(c color.Color) string {
	if c == nil {
		return "#ffffff"
	}
	r, g, b, _ := c.RGBA()
	// RGBA returns 16-bit values (0-65535), convert to 8-bit (0-255)
	return fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
}
