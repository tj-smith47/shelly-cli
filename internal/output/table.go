// Package output provides output formatting utilities for the CLI.
package output

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// TableBorderStyle defines the border style for tables.
type TableBorderStyle int

// TableBorderStyle constants define available border styles.
const (
	BorderNone    TableBorderStyle = iota // No visible borders
	BorderRounded                         // Modern rounded corners (default for TTY)
	BorderSquare                          // Square corners
	BorderDouble                          // Double-line borders
	BorderHeavy                           // Heavy/bold borders
	BorderASCII                           // ASCII-only for --plain mode
)

// borderStyles maps border style constants to lipgloss border definitions.
var borderStyles = map[TableBorderStyle]lipgloss.Border{
	BorderNone:    lipgloss.HiddenBorder(),
	BorderRounded: lipgloss.RoundedBorder(),
	BorderSquare:  lipgloss.NormalBorder(),
	BorderDouble:  lipgloss.DoubleBorder(),
	BorderHeavy:   lipgloss.BlockBorder(),
	BorderASCII:   lipgloss.ASCIIBorder(),
}

// Table represents a formatted table.
type Table struct {
	headers []string
	rows    [][]string
	style   TableStyle
}

// TableStyle defines the visual style for a table.
type TableStyle struct {
	Header      lipgloss.Style
	Cell        lipgloss.Style
	AltCell     lipgloss.Style   // Alternating row color
	PrimaryCell lipgloss.Style   // First column styling (e.g., Name column)
	Border      lipgloss.Style   // Border character styling
	BorderStyle TableBorderStyle // Border style (rounded, square, etc.)
	Padding     int
	ShowBorder  bool
}

// NewTable creates a new table with the given headers.
func NewTable(headers ...string) *Table {
	return &Table{
		headers: headers,
		rows:    [][]string{},
		style:   DefaultTableStyle(),
	}
}

// DefaultTableStyle returns the default table style using semantic colors.
// Uses rounded borders with themed colors for modern terminal appearance.
// Colors are drawn from the semantic theme system for consistency.
func DefaultTableStyle() TableStyle {
	colors := theme.GetSemanticColors()
	return TableStyle{
		Header:      lipgloss.NewStyle().Bold(true).Foreground(colors.TableHeader),
		Cell:        lipgloss.NewStyle().Foreground(colors.TableCell),
		AltCell:     lipgloss.NewStyle().Foreground(colors.TableAltCell),
		PrimaryCell: lipgloss.NewStyle().Foreground(colors.Primary),
		Border:      lipgloss.NewStyle().Foreground(colors.TableBorder),
		BorderStyle: BorderRounded,
		Padding:     1,
		ShowBorder:  true,
	}
}

// PlainTableStyle returns a table style for plain (non-TTY or --plain) output.
// Uses ASCII borders with no colors for machine-readable output.
func PlainTableStyle() TableStyle {
	return TableStyle{
		Header:      lipgloss.NewStyle(),
		Cell:        lipgloss.NewStyle(),
		AltCell:     lipgloss.NewStyle(),
		PrimaryCell: lipgloss.NewStyle(),
		Border:      lipgloss.NewStyle(),
		BorderStyle: BorderASCII,
		Padding:     1,
		ShowBorder:  true,
	}
}

// GetTableStyle returns the appropriate table style based on plain mode.
// Returns PlainTableStyle() if ios.IsPlainMode() is true, otherwise DefaultTableStyle().
func GetTableStyle(ios PlainModeChecker) TableStyle {
	if ios != nil && ios.IsPlainMode() {
		return PlainTableStyle()
	}
	return DefaultTableStyle()
}

// PlainModeChecker is an interface for checking plain mode.
// This allows decoupling from the iostreams package.
type PlainModeChecker interface {
	IsPlainMode() bool
}

// SetStyle sets a custom table style.
func (t *Table) SetStyle(style TableStyle) *Table {
	t.style = style
	return t
}

// SetBorderStyle sets the border style for the table.
func (t *Table) SetBorderStyle(style TableBorderStyle) *Table {
	t.style.BorderStyle = style
	return t
}

// HideBorders hides all table borders.
func (t *Table) HideBorders() *Table {
	t.style.ShowBorder = false
	t.style.BorderStyle = BorderNone
	return t
}

// AddRow adds a row to the table.
func (t *Table) AddRow(cells ...string) *Table {
	// Ensure row has same number of columns as headers
	row := make([]string, len(t.headers))
	for i := range row {
		if i < len(cells) {
			row[i] = cells[i]
		}
	}
	t.rows = append(t.rows, row)
	return t
}

// AddRows adds multiple rows to the table.
func (t *Table) AddRows(rows [][]string) *Table {
	for _, row := range rows {
		t.AddRow(row...)
	}
	return t
}

// Render renders the table to a string using lipgloss table.
func (t *Table) Render() string {
	if len(t.headers) == 0 {
		return ""
	}

	// Get the appropriate border style
	border := borderStyles[t.style.BorderStyle]
	if !t.style.ShowBorder {
		border = lipgloss.HiddenBorder()
	}

	// Build lipgloss table with styling
	tbl := table.New().
		Border(border).
		BorderStyle(t.style.Border).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return t.style.Header.Padding(0, t.style.Padding)
			}
			// First column (Name) gets primary styling
			if col == 0 {
				return t.style.PrimaryCell.Padding(0, t.style.Padding)
			}
			if row%2 == 0 {
				return t.style.Cell.Padding(0, t.style.Padding)
			}
			return t.style.AltCell.Padding(0, t.style.Padding)
		}).
		Headers(t.headers...).
		Rows(t.rows...)

	return tbl.Render()
}

// Print prints the table to stdout.
func (t *Table) Print() {
	fmt.Print(t.Render())
}

// PrintTo prints the table to the specified writer.
func (t *Table) PrintTo(w io.Writer) error {
	_, err := fmt.Fprint(w, t.Render())
	return err
}

// String returns the rendered table as a string.
func (t *Table) String() string {
	return t.Render()
}

// PrintTableTo is a convenience function to create and print a table to a writer.
func PrintTableTo(w io.Writer, headers []string, rows [][]string) error {
	t := NewTable(headers...)
	t.AddRows(rows)
	return t.PrintTo(w)
}

// Empty returns true if the table has no rows.
func (t *Table) Empty() bool {
	return len(t.rows) == 0
}

// RowCount returns the number of rows in the table.
func (t *Table) RowCount() int {
	return len(t.rows)
}

// buildTableFromData uses reflection to build a table from structured data.
// Supports slices/arrays of structs. Returns nil for unsupported types.
func buildTableFromData(data any) *Table {
	v := reflect.ValueOf(data)

	// Dereference pointers
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	// Must be a slice or array
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return nil
	}

	if v.Len() == 0 {
		return nil
	}

	// Get the element type
	elemType := v.Type().Elem()
	for elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}

	// Must be a struct
	if elemType.Kind() != reflect.Struct {
		return nil
	}

	// Build headers from struct fields
	headers := buildHeadersFromType(elemType)
	if len(headers) == 0 {
		return nil
	}

	tbl := NewTable(headers...)

	// Build rows from slice elements
	for i := range v.Len() {
		elem := v.Index(i)
		for elem.Kind() == reflect.Ptr {
			if elem.IsNil() {
				continue
			}
			elem = elem.Elem()
		}
		row := buildRowFromStruct(elem, elemType)
		tbl.AddRow(row...)
	}

	return tbl
}

// buildHeadersFromType extracts column headers from struct field names/tags.
func buildHeadersFromType(t reflect.Type) []string {
	headers := make([]string, 0, t.NumField())
	for i := range t.NumField() {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Skip fields with json:"-"
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		// Use json tag name if available, otherwise field name
		name := field.Name
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				name = parts[0]
			}
		}

		headers = append(headers, name)
	}
	return headers
}

// buildRowFromStruct extracts cell values from a struct.
func buildRowFromStruct(v reflect.Value, t reflect.Type) []string {
	cells := make([]string, 0, t.NumField())
	for i := range t.NumField() {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Skip fields with json:"-"
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		fieldVal := v.Field(i)
		cells = append(cells, formatFieldValue(fieldVal))
	}
	return cells
}

// formatFieldValue converts a reflect.Value to a string for table display.
func formatFieldValue(v reflect.Value) string {
	// Handle pointers
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return LabelPlaceholder
		}
		v = v.Elem()
	}

	return formatByKind(v)
}

// formatByKind formats a value based on its reflect.Kind.
func formatByKind(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return formatString(v.String())
	case reflect.Bool:
		return formatBool(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%.2f", v.Float())
	case reflect.Slice, reflect.Array:
		return formatCollection(v.Len(), "items")
	case reflect.Map:
		return formatCollection(v.Len(), "keys")
	case reflect.Struct:
		return formatStruct(v)
	default:
		return formatDefault(v)
	}
}

func formatString(s string) string {
	if s == "" {
		return LabelPlaceholder
	}
	return s
}

func formatBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func formatCollection(length int, label string) string {
	if length == 0 {
		return LabelPlaceholder
	}
	return fmt.Sprintf("[%d %s]", length, label)
}

func formatStruct(v reflect.Value) string {
	if v.CanInterface() {
		if stringer, ok := v.Interface().(fmt.Stringer); ok {
			return stringer.String()
		}
		return fmt.Sprintf("%+v", v.Interface())
	}
	return LabelPlaceholder
}

func formatDefault(v reflect.Value) string {
	if v.CanInterface() {
		return fmt.Sprintf("%v", v.Interface())
	}
	return LabelPlaceholder
}
