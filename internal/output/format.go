// Package output provides output formatting utilities for the CLI.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Format represents an output format.
type Format string

// Output format constants.
const (
	FormatJSON     Format = "json"
	FormatYAML     Format = "yaml"
	FormatTable    Format = "table"
	FormatText     Format = "text"
	FormatTemplate Format = "template"
)

// Formatter defines the interface for output formatters.
type Formatter interface {
	Format(w io.Writer, data any) error
}

// GetFormat returns the current output format from config.
func GetFormat() Format {
	format := viper.GetString("output")
	switch strings.ToLower(format) {
	case "json":
		return FormatJSON
	case "yaml", "yml":
		return FormatYAML
	case "table":
		return FormatTable
	case "text", "plain":
		return FormatText
	case "template", "go-template":
		return FormatTemplate
	default:
		return FormatTable
	}
}

// GetTemplate returns the current template string from config.
func GetTemplate() string {
	return viper.GetString("template")
}

// Print outputs data in the configured format.
func Print(data any) error {
	return PrintTo(os.Stdout, data)
}

// PrintTo outputs data to the specified writer in the configured format.
func PrintTo(w io.Writer, data any) error {
	formatter := NewFormatter(GetFormat())
	return formatter.Format(w, data)
}

// PrintJSON outputs data as JSON.
func PrintJSON(data any) error {
	return NewJSONFormatter().Format(os.Stdout, data)
}

// PrintYAML outputs data as YAML.
func PrintYAML(data any) error {
	return NewYAMLFormatter().Format(os.Stdout, data)
}

// JSON outputs data as JSON to the specified writer.
func JSON(w io.Writer, data any) error {
	return NewJSONFormatter().Format(w, data)
}

// YAML outputs data as YAML to the specified writer.
func YAML(w io.Writer, data any) error {
	return NewYAMLFormatter().Format(w, data)
}

// NewFormatter creates a formatter for the given format.
func NewFormatter(format Format) Formatter {
	switch format {
	case FormatJSON:
		return NewJSONFormatter()
	case FormatYAML:
		return NewYAMLFormatter()
	case FormatTable:
		return NewTableFormatter()
	case FormatText:
		return NewTextFormatter()
	case FormatTemplate:
		return NewTemplateFormatter(GetTemplate())
	default:
		return NewTableFormatter()
	}
}

// JSONFormatter formats output as JSON.
type JSONFormatter struct {
	Indent bool
}

// NewJSONFormatter creates a new JSON formatter.
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{Indent: true}
}

// Format outputs data as JSON.
func (f *JSONFormatter) Format(w io.Writer, data any) error {
	encoder := json.NewEncoder(w)
	if f.Indent {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(data)
}

// YAMLFormatter formats output as YAML.
type YAMLFormatter struct{}

// NewYAMLFormatter creates a new YAML formatter.
func NewYAMLFormatter() *YAMLFormatter {
	return &YAMLFormatter{}
}

// Format outputs data as YAML.
func (f *YAMLFormatter) Format(w io.Writer, data any) (err error) {
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	defer func() {
		if cerr := encoder.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	return encoder.Encode(data)
}

// TextFormatter formats output as plain text.
type TextFormatter struct{}

// NewTextFormatter creates a new text formatter.
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{}
}

// Format outputs data as plain text.
func (f *TextFormatter) Format(w io.Writer, data any) error {
	// Handle different types
	switch v := data.(type) {
	case string:
		_, err := fmt.Fprintln(w, v)
		return err
	case []string:
		for _, s := range v {
			if _, err := fmt.Fprintln(w, s); err != nil {
				return err
			}
		}
		return nil
	case fmt.Stringer:
		_, err := fmt.Fprintln(w, v.String())
		return err
	default:
		_, err := fmt.Fprintf(w, "%+v\n", v)
		return err
	}
}

// TableFormatter formats output as a table.
// This is a placeholder - the actual implementation is in table.go.
type TableFormatter struct{}

// NewTableFormatter creates a new table formatter.
func NewTableFormatter() *TableFormatter {
	return &TableFormatter{}
}

// Format outputs data as a table.
func (f *TableFormatter) Format(w io.Writer, data any) error {
	// For non-tabular data, fall back to text format
	// Actual table formatting is done via the Table type in table.go
	return NewTextFormatter().Format(w, data)
}

// ParseFormat parses a format string into a Format.
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(s) {
	case "json":
		return FormatJSON, nil
	case "yaml", "yml":
		return FormatYAML, nil
	case "table":
		return FormatTable, nil
	case "text", "plain":
		return FormatText, nil
	case "template", "go-template":
		return FormatTemplate, nil
	default:
		return "", fmt.Errorf("unknown format: %s", s)
	}
}

// ValidFormats returns a list of valid format strings.
func ValidFormats() []string {
	return []string{"json", "yaml", "table", "text", "template"}
}

// TemplateFormatter formats output using Go text/template.
type TemplateFormatter struct {
	Template string
}

// NewTemplateFormatter creates a new template formatter with the given template string.
func NewTemplateFormatter(tmpl string) *TemplateFormatter {
	return &TemplateFormatter{Template: tmpl}
}

// Format outputs data using the Go template.
func (f *TemplateFormatter) Format(w io.Writer, data any) error {
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

// Template outputs data using the specified template to the given writer.
func Template(w io.Writer, tmpl string, data any) error {
	return NewTemplateFormatter(tmpl).Format(w, data)
}

// PrintTemplate outputs data using the specified template to stdout.
func PrintTemplate(tmpl string, data any) error {
	return Template(os.Stdout, tmpl, data)
}
