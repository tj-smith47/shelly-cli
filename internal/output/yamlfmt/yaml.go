// Package yamlfmt provides YAML formatting with optional syntax highlighting.
package yamlfmt

import (
	"bytes"
	"io"

	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/output/synfmt"
)

// Formatter formats output as YAML with optional syntax highlighting.
type Formatter struct {
	Highlight bool // Enable syntax highlighting (disabled in --no-color/--plain)
}

// New creates a new YAML formatter with syntax highlighting.
func New() *Formatter {
	return &Formatter{
		Highlight: synfmt.ShouldHighlight(),
	}
}

// Format outputs data as YAML with optional syntax highlighting.
func (f *Formatter) Format(w io.Writer, data any) error {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(data); err != nil {
		return err
	}
	if err := encoder.Close(); err != nil {
		return err
	}

	output := buf.String()
	if f.Highlight {
		output = synfmt.HighlightCode(output, "yaml")
	}
	_, err := io.WriteString(w, output)
	return err
}
