// Package jsonfmt provides JSON formatting with optional syntax highlighting.
package jsonfmt

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/tj-smith47/shelly-cli/internal/output/synfmt"
)

// Formatter formats output as JSON with optional syntax highlighting.
type Formatter struct {
	Indent    bool
	Highlight bool // Enable syntax highlighting (disabled in --no-color/--plain)
}

// New creates a new JSON formatter with syntax highlighting.
func New() *Formatter {
	return &Formatter{
		Indent:    true,
		Highlight: synfmt.ShouldHighlight(),
	}
}

// Format outputs data as JSON with optional syntax highlighting.
func (f *Formatter) Format(w io.Writer, data any) error {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	if f.Indent {
		encoder.SetIndent("", "  ")
	}
	if err := encoder.Encode(data); err != nil {
		return err
	}

	output := buf.String()
	if f.Highlight {
		output = synfmt.HighlightCode(output, "json")
	}
	_, err := io.WriteString(w, output)
	return err
}
