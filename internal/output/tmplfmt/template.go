// Package tmplfmt provides Go text/template output formatting.
package tmplfmt

import (
	"fmt"
	"io"
	"strings"
	"text/template"
)

// Formatter formats output using Go text/template.
type Formatter struct {
	Template string
}

// New creates a new template formatter with the given template string.
func New(tmpl string) *Formatter {
	return &Formatter{Template: tmpl}
}

// Format outputs data using the Go template.
func (f *Formatter) Format(w io.Writer, data any) error {
	if f.Template == "" {
		return fmt.Errorf("template string is required when using template output format (use --template flag)")
	}

	tmpl, err := template.New("output").Parse(f.Template)
	if err != nil {
		return fmt.Errorf("invalid template: %w", err)
	}

	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("template execution failed: %w", err)
	}

	// Add newline if template doesn't end with one
	if !strings.HasSuffix(f.Template, "\n") {
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}

	return nil
}
