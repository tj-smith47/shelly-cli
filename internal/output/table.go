// Package output provides output formatting utilities for the CLI.
package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Table represents a formatted table.
type Table struct {
	headers []string
	rows    [][]string
	style   TableStyle
}

// TableStyle defines the visual style for a table.
type TableStyle struct {
	Header     lipgloss.Style
	Cell       lipgloss.Style
	AltCell    lipgloss.Style // Alternating row color
	Border     lipgloss.Style
	Padding    int
	ShowBorder bool
}

// NewTable creates a new table with the given headers.
func NewTable(headers ...string) *Table {
	return &Table{
		headers: headers,
		rows:    [][]string{},
		style:   DefaultTableStyle(),
	}
}

// DefaultTableStyle returns the default table style using the current theme.
func DefaultTableStyle() TableStyle {
	return TableStyle{
		Header:     lipgloss.NewStyle().Bold(true).Foreground(theme.Fg()),
		Cell:       lipgloss.NewStyle().Foreground(theme.Fg()),
		AltCell:    lipgloss.NewStyle().Foreground(theme.BrightBlack()),
		Border:     lipgloss.NewStyle().Foreground(theme.BrightBlack()),
		Padding:    2,
		ShowBorder: false,
	}
}

// SetStyle sets a custom table style.
func (t *Table) SetStyle(style TableStyle) *Table {
	t.style = style
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

// Render renders the table to a string.
func (t *Table) Render() string {
	if len(t.headers) == 0 {
		return ""
	}

	// Calculate column widths
	widths := make([]int, len(t.headers))
	for i, h := range t.headers {
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

	// Render header
	headerCells := make([]string, len(t.headers))
	for i, h := range t.headers {
		headerCells[i] = t.style.Header.
			Width(widths[i] + t.style.Padding).
			Render(h)
	}
	sb.WriteString(strings.Join(headerCells, ""))
	sb.WriteString("\n")

	// Render separator
	if t.style.ShowBorder {
		for i, w := range widths {
			sb.WriteString(t.style.Border.Render(strings.Repeat("─", w+t.style.Padding)))
			if i < len(widths)-1 {
				sb.WriteString(t.style.Border.Render("┼"))
			}
		}
		sb.WriteString("\n")
	}

	// Render rows
	for rowIdx, row := range t.rows {
		style := t.style.Cell
		if rowIdx%2 == 1 {
			style = t.style.AltCell
		}

		rowCells := make([]string, len(t.headers))
		for i := range t.headers {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			rowCells[i] = style.
				Width(widths[i] + t.style.Padding).
				Render(cell)
		}
		sb.WriteString(strings.Join(rowCells, ""))
		sb.WriteString("\n")
	}

	return sb.String()
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

// PrintTable is a convenience function to create and print a table.
func PrintTable(headers []string, rows [][]string) {
	t := NewTable(headers...)
	t.AddRows(rows)
	t.Print()
}

// PrintTableTo is a convenience function to create and print a table to a writer.
func PrintTableTo(w io.Writer, headers []string, rows [][]string) error {
	t := NewTable(headers...)
	t.AddRows(rows)
	return t.PrintTo(w)
}

// KeyValueTable creates a table from key-value pairs.
func KeyValueTable(pairs map[string]string) *Table {
	t := NewTable("Key", "Value")
	for k, v := range pairs {
		t.AddRow(k, v)
	}
	return t
}

// StatusTable creates a table suitable for device status display.
func StatusTable() *Table {
	return NewTable("Name", "Address", "Type", "Status", "Power")
}

// DeviceTable creates a table for device listings.
func DeviceTable() *Table {
	return NewTable("Name", "Address", "Generation", "Type", "Model")
}

// PrintTableToWriter outputs the table to the given writer.
func PrintTableToWriter(w io.Writer, t *Table) error {
	_, err := io.WriteString(w, t.Render())
	return err
}

// MustPrintTable prints a table and panics on error.
func MustPrintTable(w io.Writer, t *Table) {
	if err := PrintTableToWriter(w, t); err != nil {
		panic(err)
	}
}

// Empty returns true if the table has no rows.
func (t *Table) Empty() bool {
	return len(t.rows) == 0
}

// RowCount returns the number of rows in the table.
func (t *Table) RowCount() int {
	return len(t.rows)
}
