// Package table provides table rendering with styled output.
package table

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"charm.land/lipgloss/v2"
	lgtable "charm.land/lipgloss/v2/table"
)

// LabelPlaceholder is the placeholder for empty/missing values.
const LabelPlaceholder = "-"

// LabelTrue and LabelFalse are boolean display labels.
const (
	LabelTrue  = "true"
	LabelFalse = "false"
)

// Table represents a formatted table.
type Table struct {
	headers     []string
	rows        [][]string
	style       Style
	hideHeaders bool // Hide header row entirely (for --no-headers)
}

// New creates a new table with the given headers.
func New(headers ...string) *Table {
	return &Table{
		headers: headers,
		rows:    [][]string{},
		style:   DefaultStyle(),
	}
}

// NewStyled creates a table with style and headers settings from flags.
// This is a convenience function that applies GetStyle and --no-headers.
func NewStyled(ios ModeChecker, headers ...string) *Table {
	t := New(headers...)
	t.SetStyle(GetStyle(ios))
	if ShouldHideHeaders() {
		t.HideHeaders()
	}
	return t
}

// SetStyle sets a custom table style.
func (t *Table) SetStyle(style Style) *Table {
	t.style = style
	return t
}

// SetBorderStyle sets the border style for the table.
func (t *Table) SetBorderStyle(style BorderStyle) *Table {
	t.style.BorderStyle = style
	return t
}

// HideBorders hides all table borders.
func (t *Table) HideBorders() *Table {
	t.style.ShowBorder = false
	t.style.BorderStyle = BorderNone
	return t
}

// HideHeaders hides the header row (for --no-headers flag).
func (t *Table) HideHeaders() *Table {
	t.hideHeaders = true
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

	// Plain mode: tab-separated values, no borders
	if t.style.PlainMode {
		return t.renderPlain()
	}

	// Prepare headers (uppercase if configured)
	headers := t.prepareHeaders()

	// Get the appropriate border style
	border := borderStyles[t.style.BorderStyle]
	if !t.style.ShowBorder {
		border = lipgloss.HiddenBorder()
	}

	// Build lipgloss table with styling
	tbl := lgtable.New().
		Border(border).
		BorderStyle(t.style.Border).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == lgtable.HeaderRow {
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
		})

	// Add headers unless hidden
	if !t.hideHeaders {
		tbl = tbl.Headers(headers...)
	}

	tbl = tbl.Rows(t.rows...)

	// Ensure output ends with newline (lipgloss doesn't add one)
	rendered := tbl.Render()
	if rendered != "" && rendered[len(rendered)-1] != '\n' {
		rendered += "\n"
	}
	return rendered
}

// prepareHeaders returns headers, uppercased if configured.
func (t *Table) prepareHeaders() []string {
	if !t.style.UppercaseHeaders {
		return t.headers
	}
	headers := make([]string, len(t.headers))
	for i, h := range t.headers {
		headers[i] = strings.ToUpper(h)
	}
	return headers
}

// renderPlain renders the table with aligned columns but no borders for --plain mode.
func (t *Table) renderPlain() string {
	headers := t.prepareHeaders()

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range t.rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	var sb strings.Builder

	// Headers (unless hidden)
	if !t.hideHeaders {
		for i, h := range headers {
			if i > 0 {
				sb.WriteString("  ") // Column separator
			}
			sb.WriteString(padRight(h, widths[i]))
		}
		sb.WriteString("\n")
	}

	// Rows
	for _, row := range t.rows {
		for i := range widths {
			if i > 0 {
				sb.WriteString("  ") // Column separator
			}
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			sb.WriteString(padRight(cell, widths[i]))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// padRight pads a string with spaces to reach the specified width.
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
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

// PrintTo is a convenience function to create and print a table to a writer.
func PrintTo(w io.Writer, headers []string, rows [][]string) error {
	t := New(headers...)
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

// BuildFromData uses reflection to build a table from structured data.
// Supports slices/arrays of structs. Returns nil for unsupported types.
func BuildFromData(data any) *Table {
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

	tbl := New(headers...)

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
		return LabelTrue
	}
	return LabelFalse
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
